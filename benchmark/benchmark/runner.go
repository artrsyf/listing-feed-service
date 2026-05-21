package benchmark

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Runner struct {
	postgres Executor
	elastic  Executor

	workers int

	metrics *Metrics
}

type Executor interface {
	RunQuery(ctx context.Context, query string) (time.Duration, error)
}

func NewRunner(pg Executor, elastic Executor, workers int) *Runner {
	return &Runner{
		postgres: pg,
		elastic:  elastic,
		workers:  workers,
		metrics:  NewMetrics(),
	}
}

func (r *Runner) RunScenario(name string, queries []string, executor string) {
	fmt.Printf("▶ Running scenario: %s (executor: %s)\n", name, executor)

	var wg sync.WaitGroup
	jobs := make(chan string, len(queries))

	startTime := time.Now()

	// workers
	for i := 0; i < r.workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for q := range jobs {
				start := time.Now()

				var err error
				if executor == "elastic" {
					_, err = r.elastic.RunQuery(context.Background(), q)
				} else {
					_, err = r.postgres.RunQuery(context.Background(), q)
				}

				if err != nil {
					fmt.Printf("Query error: %v\n", err)
					continue
				}

				r.metrics.Add(time.Since(start))
			}
		}()
	}

	// send jobs
	for _, q := range queries {
		jobs <- q
	}

	close(jobs)
	wg.Wait()

	r.metrics.Finish()

	fmt.Printf("✅ Scenario finished in %s\n", time.Since(startTime))
}

func (r *Runner) Reset() {
	r.metrics = NewMetrics()
}

func (r *Runner) Report() {
	r.metrics.Report()
}

// =======================
// SCENARIOS
// =======================

var JoinScenario = []string{
	`SELECT u.country, c.name, SUM(oi.quantity * oi.price) as revenue
	FROM users u
	JOIN orders o ON u.id = o.user_id
	JOIN order_items oi ON o.id = oi.order_id
	JOIN products p ON oi.product_id = p.id
	JOIN categories c ON p.category_id = c.id
	WHERE o.created_at > NOW() - interval '30 days'
	GROUP BY u.country, c.name
	ORDER BY revenue DESC
	LIMIT 100`,
}

var JoinScenarioWithIndex = []string{
	`SELECT u.country, COUNT(*) as order_count, SUM(o.total_amount) as total_revenue
	FROM users u
	JOIN orders o ON u.id = o.user_id
	WHERE o.created_at > NOW() - interval '90 days'
	GROUP BY u.country
	ORDER BY total_revenue DESC
	LIMIT 50`,
}

var PointLookupScenario = []string{
	`SELECT * FROM orders WHERE id = 1`,
	`SELECT * FROM orders WHERE id = 100`,
	`SELECT * FROM orders WHERE id = 1000`,
	`SELECT * FROM orders WHERE id = 10000`,
	`SELECT * FROM orders WHERE id = 100000`,
	`SELECT * FROM orders WHERE id = 1000000`,
}

var RangeScanScenario = []string{
	`SELECT * FROM orders WHERE created_at > NOW() - interval '7 days'`,
	`SELECT * FROM orders WHERE created_at > NOW() - interval '30 days'`,
	`SELECT * FROM orders WHERE created_at > NOW() - interval '90 days'`,
}

var AggregationScenario = []string{
	`SELECT DATE_TRUNC('day', created_at) as day, COUNT(*) as orders, SUM(total_amount) as revenue
	FROM orders
	WHERE created_at > NOW() - interval '30 days'
	GROUP BY 1
	ORDER BY 1`,
}

var ElasticAggScenario = []string{
	`{"size":0,"aggs":{"by_country":{"terms":{"field":"user.country.keyword","size":100}}}}`,
}

var ElasticRangeScenario = []string{
	`{"query":{"range":{"created_at":{"gte":"now-30d/d"}}}}`,
}

var ElasticSearchScenario = []string{
	`{"query":{"match":{"user.country":"US"}}}`,
}

// =======================
// POSTGRES JOIN HELPERS
// =======================

func SetHashJoin(pg Executor) {
	pg.RunQuery(context.Background(),
		"SET enable_hashjoin = on; SET enable_mergejoin = off; SET enable_nestloop = off;")
}

func SetMergeJoin(pg Executor) {
	pg.RunQuery(context.Background(),
		"SET enable_mergejoin = on; SET enable_hashjoin = off; SET enable_nestloop = off;")
}

func SetNestedLoop(pg Executor) {
	pg.RunQuery(context.Background(),
		"SET enable_nestloop = on; SET enable_hashjoin = off; SET enable_mergejoin = off;")
}

func SetDefaultJoin(pg Executor) {
	pg.RunQuery(context.Background(),
		"SET enable_hashjoin = on; SET enable_mergejoin = on; SET enable_nestloop = on;")
}
