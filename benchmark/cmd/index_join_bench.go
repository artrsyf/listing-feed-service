package main

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"
)

type indexJoinResult struct {
	Profile  string
	Scenario string
	Requests int
	Duration time.Duration
	Avg      time.Duration
	P50      time.Duration
	P95      time.Duration
	P99      time.Duration
	Max      time.Duration
	RPS      float64
}

type indexProfile struct {
	Name string
	SQL  []string
}

func runIndexJoinBenchmarks(ctx context.Context, postgresDSN string, workers, iterations int) error {
	db, err := sql.Open("postgres", postgresDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	if workers < 1 {
		workers = 1
	}
	if iterations < 1 {
		iterations = 1
	}

	results := make([]indexJoinResult, 0)
	for _, profile := range indexProfiles() {
		fmt.Printf("\n=== INDEX PROFILE: %s ===\n", profile.Name)
		if err := applyIndexProfile(ctx, db, profile); err != nil {
			return fmt.Errorf("apply profile %s: %w", profile.Name, err)
		}

		scenarios := []struct {
			name string
			mode string
			sql  string
		}{
			{"join_hash", "hash", isolatedJoinQuery()},
			{"join_merge", "merge", isolatedJoinQuery()},
			{"join_nested_loop", "nested_loop", isolatedJoinQuery()},
			{"join_default", "", isolatedJoinQuery()},
			{"point_lookup", "", "SELECT * FROM orders WHERE id = 100000"},
			{"range_scan_30d", "", "SELECT * FROM orders WHERE created_at > NOW() - interval '30 days'"},
			{"aggregation_30d", "", isolatedAggregationQuery()},
		}

		for _, scenario := range scenarios {
			result := measurePostgresScenario(ctx, db, profile.Name, scenario.name, scenario.mode, scenario.sql, workers, iterations)
			results = append(results, result)
			printIndexJoinResult(result)
		}
	}

	fmt.Println("\n=== INDEX/JOIN TOP BY RPS ===")
	sort.Slice(results, func(i, j int) bool { return results[i].RPS > results[j].RPS })
	for i, result := range results {
		fmt.Printf("%d. %s/%s RPS=%.2f avg=%s p95=%s\n", i+1, result.Profile, result.Scenario, result.RPS, result.Avg, result.P95)
	}

	return nil
}

func applyIndexProfile(ctx context.Context, db *sql.DB, profile indexProfile) error {
	fmt.Println("Dropping secondary indexes...")
	if _, err := db.ExecContext(ctx, dropSecondaryOrderIndexesSQL()); err != nil {
		return err
	}

	start := time.Now()
	for _, statement := range profile.SQL {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	if _, err := db.ExecContext(ctx, "ANALYZE users; ANALYZE categories; ANALYZE products; ANALYZE orders; ANALYZE order_items"); err != nil {
		return err
	}
	fmt.Printf("Profile %s ready in %s\n", profile.Name, time.Since(start))
	return nil
}

func measurePostgresScenario(ctx context.Context, db *sql.DB, profile, scenario, joinMode, query string, workers, iterations int) indexJoinResult {
	jobs := make(chan struct{}, iterations)
	latencies := make(chan time.Duration, iterations)
	start := time.Now()

	for worker := 0; worker < workers; worker++ {
		go func() {
			for range jobs {
				queryStart := time.Now()
				if err := runProfileQuery(ctx, db, joinMode, query); err != nil {
					fmt.Printf("Query error: %v\n", err)
					continue
				}
				latencies <- time.Since(queryStart)
			}
		}()
	}

	for i := 0; i < iterations; i++ {
		jobs <- struct{}{}
	}
	close(jobs)

	values := make([]time.Duration, 0, iterations)
	for len(values) < iterations {
		select {
		case latency := <-latencies:
			values = append(values, latency)
		case <-time.After(300 * time.Second):
			goto done
		}
	}

done:
	duration := time.Since(start)
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })

	var sum time.Duration
	var max time.Duration
	for _, value := range values {
		sum += value
		if value > max {
			max = value
		}
	}

	avg := time.Duration(0)
	if len(values) > 0 {
		avg = sum / time.Duration(len(values))
	}

	return indexJoinResult{
		Profile:  profile,
		Scenario: scenario,
		Requests: len(values),
		Duration: duration,
		Avg:      avg,
		P50:      durationPercentile(values, 50),
		P95:      durationPercentile(values, 95),
		P99:      durationPercentile(values, 99),
		Max:      max,
		RPS:      float64(len(values)) / duration.Seconds(),
	}
}

func runProfileQuery(ctx context.Context, db *sql.DB, joinMode, query string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if joinMode != "" {
		if _, err := tx.ExecContext(ctx, joinModeSQLForProfile(joinMode)); err != nil {
			return err
		}
	}

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return tx.Commit()
}

func printIndexJoinResult(result indexJoinResult) {
	fmt.Println("========== INDEX/JOIN BENCHMARK REPORT ==========")
	fmt.Printf("Profile: %s\n", result.Profile)
	fmt.Printf("Scenario: %s\n", result.Scenario)
	fmt.Printf("Total requests: %d\n", result.Requests)
	fmt.Printf("Duration (sec): %.3f\n", result.Duration.Seconds())
	fmt.Printf("Throughput (req/sec): %.2f\n", result.RPS)
	fmt.Printf("Avg latency: %s\n", result.Avg)
	fmt.Printf("p50: %s\n", result.P50)
	fmt.Printf("p95: %s\n", result.P95)
	fmt.Printf("p99: %s\n", result.P99)
	fmt.Printf("Max latency: %s\n", result.Max)
}

func joinModeSQLForProfile(mode string) string {
	switch mode {
	case "hash":
		return "SET LOCAL enable_hashjoin = on; SET LOCAL enable_mergejoin = off; SET LOCAL enable_nestloop = off"
	case "merge":
		return "SET LOCAL enable_mergejoin = on; SET LOCAL enable_hashjoin = off; SET LOCAL enable_nestloop = off"
	case "nested_loop":
		return "SET LOCAL enable_nestloop = on; SET LOCAL enable_hashjoin = off; SET LOCAL enable_mergejoin = off"
	default:
		return "SET LOCAL enable_hashjoin = on; SET LOCAL enable_mergejoin = on; SET LOCAL enable_nestloop = on"
	}
}

func isolatedJoinQuery() string {
	return `SELECT u.country, c.name, SUM(oi.quantity * oi.price) as revenue
	FROM users u
	JOIN orders o ON u.id = o.user_id
	JOIN order_items oi ON o.id = oi.order_id
	JOIN products p ON oi.product_id = p.id
	JOIN categories c ON p.category_id = c.id
	WHERE o.created_at > NOW() - interval '30 days'
	GROUP BY u.country, c.name
	ORDER BY revenue DESC
	LIMIT 100`
}

func isolatedAggregationQuery() string {
	return `SELECT DATE_TRUNC('day', created_at) as day, COUNT(*) as orders, SUM(total_amount) as revenue
	FROM orders
	WHERE created_at > NOW() - interval '30 days'
	GROUP BY 1
	ORDER BY 1`
}

func indexProfiles() []indexProfile {
	return []indexProfile{
		{
			Name: "pk_only",
			SQL:  nil,
		},
		{
			Name: "btree_fk_time",
			SQL: []string{
				"CREATE INDEX idx_users_country ON users(country)",
				"CREATE INDEX idx_products_category_id ON products(category_id)",
				"CREATE INDEX idx_orders_user_id ON orders(user_id)",
				"CREATE INDEX idx_orders_created_at ON orders(created_at)",
				"CREATE INDEX idx_order_items_order_id ON order_items(order_id)",
				"CREATE INDEX idx_order_items_product_id ON order_items(product_id)",
			},
		},
		{
			Name: "brin_time_btree_fk",
			SQL: []string{
				"CREATE INDEX idx_products_category_id ON products(category_id)",
				"CREATE INDEX idx_orders_user_id ON orders(user_id)",
				"CREATE INDEX idx_orders_created_at_brin ON orders USING BRIN(created_at)",
				"CREATE INDEX idx_order_items_order_id ON order_items(order_id)",
				"CREATE INDEX idx_order_items_product_id ON order_items(product_id)",
			},
		},
		{
			Name: "hash_lookup_btree_fk",
			SQL: []string{
				"CREATE INDEX idx_orders_id_hash ON orders USING HASH(id)",
				"CREATE INDEX idx_users_id_hash ON users USING HASH(id)",
				"CREATE INDEX idx_products_category_id ON products(category_id)",
				"CREATE INDEX idx_orders_user_id ON orders(user_id)",
				"CREATE INDEX idx_orders_created_at ON orders(created_at)",
				"CREATE INDEX idx_order_items_order_id ON order_items(order_id)",
				"CREATE INDEX idx_order_items_product_id ON order_items(product_id)",
			},
		},
		{
			Name: "covering_composite",
			SQL: []string{
				"CREATE INDEX idx_users_country ON users(country)",
				"CREATE INDEX idx_products_category_id ON products(category_id)",
				"CREATE INDEX idx_orders_created_user_covering ON orders(created_at, user_id) INCLUDE(total_amount, status)",
				"CREATE INDEX idx_orders_user_created ON orders(user_id, created_at)",
				"CREATE INDEX idx_order_items_order_includes ON order_items(order_id) INCLUDE(product_id, quantity, price)",
				"CREATE INDEX idx_order_items_product_id ON order_items(product_id)",
				"CREATE INDEX idx_products_active ON products(id, category_id, price) WHERE is_active = true",
			},
		},
		{
			Name: "mixed_full",
			SQL: []string{
				"CREATE INDEX idx_users_country ON users(country)",
				"CREATE INDEX idx_users_created_at ON users(created_at)",
				"CREATE INDEX idx_users_status ON users(status)",
				"CREATE INDEX idx_products_category_id ON products(category_id)",
				"CREATE INDEX idx_products_price ON products(price)",
				"CREATE INDEX idx_products_rating ON products(rating)",
				"CREATE UNIQUE INDEX idx_products_sku ON products(sku)",
				"CREATE INDEX idx_orders_user_id ON orders(user_id)",
				"CREATE INDEX idx_orders_created_at ON orders(created_at)",
				"CREATE INDEX idx_orders_status ON orders(status)",
				"CREATE INDEX idx_orders_user_created ON orders(user_id, created_at)",
				"CREATE INDEX idx_orders_created_status ON orders(created_at, status)",
				"CREATE INDEX idx_order_items_order_id ON order_items(order_id)",
				"CREATE INDEX idx_order_items_product_id ON order_items(product_id)",
				"CREATE INDEX idx_order_items_order_product ON order_items(order_id, product_id)",
				"CREATE INDEX idx_orders_id_hash ON orders USING HASH(id)",
				"CREATE INDEX idx_users_id_hash ON users USING HASH(id)",
				"CREATE INDEX idx_orders_created_user_covering ON orders(created_at, user_id) INCLUDE(total_amount, status)",
				"CREATE INDEX idx_order_items_order_includes ON order_items(order_id) INCLUDE(product_id, quantity, price)",
				"CREATE INDEX idx_orders_created_at_brin ON orders USING BRIN(created_at)",
				"CREATE INDEX idx_order_items_created_at_brin ON order_items USING BRIN(created_at)",
				"CREATE INDEX idx_orders_paid_recent_path ON orders(created_at) WHERE status IN (1, 2, 3)",
				"CREATE INDEX idx_orders_date_trunc_day ON orders(DATE_TRUNC('day', created_at))",
				"CREATE INDEX idx_users_country_upper ON users(UPPER(country))",
			},
		},
	}
}

func dropSecondaryOrderIndexesSQL() string {
	return `
DROP INDEX IF EXISTS idx_users_country;
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_country_upper;
DROP INDEX IF EXISTS idx_users_id_hash;
DROP INDEX IF EXISTS idx_products_category_id;
DROP INDEX IF EXISTS idx_products_price;
DROP INDEX IF EXISTS idx_products_rating;
DROP INDEX IF EXISTS idx_products_sku;
DROP INDEX IF EXISTS idx_products_active;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP INDEX IF EXISTS idx_orders_created_at;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_user_created;
DROP INDEX IF EXISTS idx_orders_created_status;
DROP INDEX IF EXISTS idx_orders_created_user_covering;
DROP INDEX IF EXISTS idx_orders_paid_recent_path;
DROP INDEX IF EXISTS idx_orders_date_trunc_day;
DROP INDEX IF EXISTS idx_orders_created_at_brin;
DROP INDEX IF EXISTS idx_orders_id_hash;
DROP INDEX IF EXISTS idx_order_items_order_id;
DROP INDEX IF EXISTS idx_order_items_product_id;
DROP INDEX IF EXISTS idx_order_items_order_product;
DROP INDEX IF EXISTS idx_order_items_order_includes;
DROP INDEX IF EXISTS idx_order_items_created_at_brin;
`
}
