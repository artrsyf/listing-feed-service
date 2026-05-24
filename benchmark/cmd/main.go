package main

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"benchmark/benchmark"
	"benchmark/executors"

	_ "github.com/lib/pq"
)

func main() {
	// flags
	generate := flag.Bool("generate", false, "Generate test data only")
	initElasticIndex := flag.Bool("init-elastic", false, "Create Elasticsearch index")
	load := flag.Bool("load", false, "Load data into databases")
	loadPostgresOnly := flag.Bool("load-postgres", false, "Load generated order data into PostgreSQL only")
	runBenchmark := flag.Bool("benchmark", false, "Run benchmarks")
	generateListings := flag.Bool("generate-listings", false, "Generate marketplace listings dataset")
	initListings := flag.Bool("init-listings", false, "Create marketplace PostgreSQL schema and Elasticsearch index")
	loadListings := flag.Bool("load-listings", false, "Load marketplace listings dataset")
	benchmarkListings := flag.Bool("benchmark-listings", false, "Run marketplace listings read benchmark")
	listingsConfig := flag.String("listings-config", "config.listing.yaml", "Marketplace listings config")
	initPhotoStorage := flag.Bool("init-photos", false, "Create photo storage schemas/buckets")
	loadPhotoStorage := flag.Bool("load-photos", false, "Load photo payloads into PostgreSQL and MinIO")
	benchmarkPhotoStorage := flag.Bool("benchmark-photos", false, "Run photo storage benchmark")
	photosConfig := flag.String("photos-config", "config.photos.yaml", "Photo storage benchmark config")
	benchmarkIndexJoins := flag.Bool("benchmark-index-joins", false, "Run isolated PostgreSQL index and join benchmark")

	postgresDSN := flag.String("postgres", envOrDefault("POSTGRES_DSN", "postgres://benchmark:benchmark@localhost:5432/benchmark?sslmode=disable"), "PostgreSQL DSN")
	elasticURL := flag.String("elastic", envOrDefault("ELASTIC_URL", "http://localhost:9200"), "Elasticsearch URL")

	workers := flag.Int("workers", 8, "Number of concurrent workers")
	iterations := flag.Int("iterations", 1000, "Number of query iterations per scenario")

	flag.Parse()

	ctx := context.Background()

	if *generate {
		fmt.Println("📊 Data generation should be run via generator/cmd/main.go")
		fmt.Println("   Usage: go run generator/cmd/main.go -config config.yaml -out ./output")
	}

	if *initElasticIndex {
		if err := initElastic(ctx, *elasticURL); err != nil {
			log.Fatalf("Failed to initialize Elasticsearch: %v", err)
		}
	}

	if *load {
		if err := loadData(ctx, *postgresDSN, *elasticURL); err != nil {
			log.Fatalf("Failed to load data: %v", err)
		}
	}

	if *loadPostgresOnly {
		if err := loadPostgres(ctx, *postgresDSN); err != nil {
			log.Fatalf("Failed to load PostgreSQL data: %v", err)
		}
	}

	if *runBenchmark {
		if err := runBenchmarks(*postgresDSN, *elasticURL, *workers, *iterations); err != nil {
			log.Fatalf("Benchmark failed: %v", err)
		}
	}

	if *generateListings {
		if err := generateListingData(*listingsConfig, "./output"); err != nil {
			log.Fatalf("Listing generation failed: %v", err)
		}
	}

	if *initListings {
		if err := initListingElastic(ctx, *elasticURL); err != nil {
			log.Fatalf("Listing Elasticsearch init failed: %v", err)
		}
	}

	if *loadListings {
		if err := loadListingsData(ctx, *postgresDSN, *elasticURL); err != nil {
			log.Fatalf("Listing load failed: %v", err)
		}
	}

	if *benchmarkListings {
		if err := runListingBenchmarks(ctx, *postgresDSN, *elasticURL, *workers, *iterations); err != nil {
			log.Fatalf("Listing benchmark failed: %v", err)
		}
	}

	if *initPhotoStorage {
		if err := initPhotos(ctx, *postgresDSN); err != nil {
			log.Fatalf("Photo init failed: %v", err)
		}
	}

	if *loadPhotoStorage {
		if err := loadPhotos(ctx, *postgresDSN, *photosConfig); err != nil {
			log.Fatalf("Photo load failed: %v", err)
		}
	}

	if *benchmarkPhotoStorage {
		if err := runPhotoBenchmarks(ctx, *postgresDSN, *photosConfig, *workers, *iterations); err != nil {
			log.Fatalf("Photo benchmark failed: %v", err)
		}
	}

	if *benchmarkIndexJoins {
		if err := runIndexJoinBenchmarks(ctx, *postgresDSN, *workers, *iterations); err != nil {
			log.Fatalf("Index/join benchmark failed: %v", err)
		}
	}

	if !*generate && !*initElasticIndex && !*load && !*loadPostgresOnly && !*runBenchmark && !*generateListings && !*initListings && !*loadListings && !*benchmarkListings && !*initPhotoStorage && !*loadPhotoStorage && !*benchmarkPhotoStorage && !*benchmarkIndexJoins {
		fmt.Println("Usage: benchmark [-generate] [-load] [-benchmark] [flags]")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Workflow:")
		fmt.Println("  1. Generate data: go run generator/cmd/main.go -config config.yaml -out ./output")
		fmt.Println("  2. Load data:     benchmark -load")
		fmt.Println("  3. Run benchmarks: benchmark -benchmark")
	}
}

func loadData(ctx context.Context, postgresDSN, elasticURL string) error {
	fmt.Println("📥 Loading data into databases...")

	// PostgreSQL
	fmt.Println("▶ Loading PostgreSQL data...")
	if err := loadPostgres(ctx, postgresDSN); err != nil {
		return fmt.Errorf("postgres load: %w", err)
	}

	// Elasticsearch
	fmt.Println("▶ Loading Elasticsearch data...")
	if err := loadElastic(ctx, elasticURL); err != nil {
		return fmt.Errorf("elastic load: %w", err)
	}

	fmt.Println("✅ Data loaded successfully")
	return nil
}

func initElastic(ctx context.Context, url string) error {
	fmt.Println("Initializing Elasticsearch index...")

	settings, err := os.ReadFile("elasticsearch/index-settings.json")
	if err != nil {
		return fmt.Errorf("read index settings: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	deleteReq, err := http.NewRequestWithContext(ctx, "DELETE", url+"/orders", nil)
	if err != nil {
		return err
	}
	deleteResp, err := client.Do(deleteReq)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, deleteResp.Body)
	deleteResp.Body.Close()

	putReq, err := http.NewRequestWithContext(ctx, "PUT", url+"/orders", bytes.NewReader(settings))
	if err != nil {
		return err
	}
	putReq.Header.Set("Content-Type", "application/json")

	putResp, err := client.Do(putReq)
	if err != nil {
		return err
	}
	defer putResp.Body.Close()

	body, err := io.ReadAll(putResp.Body)
	if err != nil {
		return err
	}
	if putResp.StatusCode < 200 || putResp.StatusCode >= 300 {
		return fmt.Errorf("create index failed: %s - %s", putResp.Status, string(body))
	}

	fmt.Println("Elasticsearch index is ready")
	return nil
}

func loadPostgres(ctx context.Context, dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, "TRUNCATE order_items, orders, products, categories, users RESTART IDENTITY CASCADE"); err != nil {
		return fmt.Errorf("truncate postgres tables: %w", err)
	}

	loads := []struct {
		table   string
		columns string
		file    string
	}{
		{"categories", "id, parent_id, name, created_at", "/data/categories.csv"},
		{"users", "id, email, country, status, created_at", "/data/users.csv"},
		{"products", "id, category_id, sku, name, price, rating, created_at", "/data/products.csv"},
		{"orders", "id, user_id, status, total_amount, created_at", "/data/orders.csv"},
		{"order_items", "id, order_id, product_id, quantity, price, created_at", "/data/order_items.csv"},
	}

	for _, load := range loads {
		fmt.Printf("  Loading %s...\n", load.table)

		query := fmt.Sprintf("COPY %s (%s) FROM '%s' WITH (FORMAT csv, NULL '\\N')", load.table, load.columns, load.file)
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("copy %s: %w", load.table, err)
		}
	}

	if _, err := db.ExecContext(ctx, `
		UPDATE orders o
		SET total_amount = totals.total_amount
		FROM (
			SELECT order_id, SUM(quantity * price)::numeric(14,2) AS total_amount
			FROM order_items
			GROUP BY order_id
		) totals
		WHERE o.id = totals.order_id
	`); err != nil {
		return fmt.Errorf("update order totals: %w", err)
	}

	if _, err := db.ExecContext(ctx, `
		SELECT setval('users_id_seq', COALESCE((SELECT MAX(id) FROM users), 1), true);
		SELECT setval('categories_id_seq', COALESCE((SELECT MAX(id) FROM categories), 1), true);
		SELECT setval('products_id_seq', COALESCE((SELECT MAX(id) FROM products), 1), true);
		SELECT setval('orders_id_seq', COALESCE((SELECT MAX(id) FROM orders), 1), true);
		SELECT setval('order_items_id_seq', COALESCE((SELECT MAX(id) FROM order_items), 1), true);
	`); err != nil {
		return fmt.Errorf("reset sequences: %w", err)
	}

	// Analyze tables
	fmt.Println("  Running ANALYZE...")
	if _, err := db.ExecContext(ctx, "ANALYZE"); err != nil {
		return err
	}

	return nil
}

func loadElastic(ctx context.Context, url string) error {
	filePath := "./output/orders.ndjson"

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open orders.ndjson: %w", err)
	}
	defer file.Close()

	client := &http.Client{Timeout: 5 * time.Minute}

	// Read and bulk index in chunks
	scanner := bufio.NewScanner(file)
	var bulkBody bytes.Buffer
	lineCount := 0
	totalIndexed := 0

	for scanner.Scan() {
		bulkBody.WriteString(scanner.Text())
		bulkBody.WriteString("\n")
		lineCount++

		// Each document is 2 lines (action + doc)
		if lineCount%2000 == 0 {
			if err := sendBulk(ctx, client, url, bulkBody.Bytes()); err != nil {
				return err
			}
			totalIndexed += lineCount / 2
			if totalIndexed%50000 == 0 {
				fmt.Printf("  Indexed %d documents...\n", totalIndexed)
			}
			bulkBody.Reset()
			lineCount = 0
		}
	}

	// Send remaining
	if bulkBody.Len() > 0 {
		if err := sendBulk(ctx, client, url, bulkBody.Bytes()); err != nil {
			return err
		}
		totalIndexed += lineCount / 2
	}

	// Refresh index
	refreshReq, _ := http.NewRequestWithContext(ctx, "POST", url+"/orders/_refresh", nil)
	refreshResp, err := client.Do(refreshReq)
	if err != nil {
		return err
	}
	refreshResp.Body.Close()

	fmt.Printf("  Elasticsearch: %d documents indexed\n", totalIndexed)
	return nil
}

func sendBulk(ctx context.Context, client *http.Client, url string, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, "POST", url+"/orders/_bulk", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bulk index failed: %s - %s", resp.Status, string(respBody))
	}

	var result struct {
		Errors bool `json:"errors"`
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return err
	}
	if result.Errors {
		return fmt.Errorf("bulk index completed with item errors")
	}

	return nil
}

func runBenchmarks(postgresDSN, elasticURL string, workers, iterations int) error {
	fmt.Println("🏃 Running benchmarks...")

	// Setup PostgreSQL
	db, err := sql.Open("postgres", postgresDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	pgExecutor := executors.NewPostgresExecutor(db)
	elasticExecutor := executors.NewElasticExecutor(elasticURL)

	runner := benchmark.NewRunner(pgExecutor, elasticExecutor, workers, iterations)

	// =======================
	// JOIN SCENARIOS (PostgreSQL)
	// =======================
	fmt.Println("\n=== JOIN SCENARIOS ===")

	// Hash Join
	fmt.Println("\n--- Hash Join ---")
	benchmark.SetHashJoin(pgExecutor)
	runner.RunScenario("hash_join", benchmark.JoinScenario, "postgres")
	runner.Report()
	runner.Reset()

	// Merge Join
	fmt.Println("\n--- Merge Join ---")
	benchmark.SetMergeJoin(pgExecutor)
	runner.RunScenario("merge_join", benchmark.JoinScenario, "postgres")
	runner.Report()
	runner.Reset()

	// Nested Loop
	fmt.Println("\n--- Nested Loop ---")
	benchmark.SetNestedLoop(pgExecutor)
	runner.RunScenario("nested_loop", benchmark.JoinScenario, "postgres")
	runner.Report()
	runner.Reset()

	// Default (optimizer choice)
	fmt.Println("\n--- Default (optimizer choice) ---")
	benchmark.SetDefaultJoin(pgExecutor)
	runner.RunScenario("default_join", benchmark.JoinScenario, "postgres")
	runner.Report()
	runner.Reset()

	// =======================
	// POINT LOOKUP (PostgreSQL)
	// =======================
	fmt.Println("\n=== POINT LOOKUP SCENARIOS ===")
	runner.RunScenario("point_lookup", benchmark.PointLookupScenario, "postgres")
	runner.Report()
	runner.Reset()

	// =======================
	// RANGE SCAN (PostgreSQL)
	// =======================
	fmt.Println("\n=== RANGE SCAN SCENARIOS ===")
	runner.RunScenario("range_scan", benchmark.RangeScanScenario, "postgres")
	runner.Report()
	runner.Reset()

	// =======================
	// AGGREGATION (PostgreSQL)
	// =======================
	fmt.Println("\n=== AGGREGATION SCENARIOS ===")
	runner.RunScenario("aggregation", benchmark.AggregationScenario, "postgres")
	runner.Report()
	runner.Reset()

	// =======================
	// ELASTICSEARCH SCENARIOS
	// =======================
	fmt.Println("\n=== ELASTICSEARCH SCENARIOS ===")

	fmt.Println("\n--- Elastic Aggregation ---")
	runner.RunScenario("elastic_agg", benchmark.ElasticAggScenario, "elastic")
	runner.Report()
	runner.Reset()

	fmt.Println("\n--- Elastic Range Query ---")
	runner.RunScenario("elastic_range", benchmark.ElasticRangeScenario, "elastic")
	runner.Report()
	runner.Reset()

	fmt.Println("\n--- Elastic Search ---")
	runner.RunScenario("elastic_search", benchmark.ElasticSearchScenario, "elastic")
	runner.Report()

	return nil
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
