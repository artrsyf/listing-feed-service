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
	"strings"
	"time"

	"benchmark/benchmark"
	"benchmark/executors"

	_ "github.com/lib/pq"
)

func main() {
	// flags
	generate := flag.Bool("generate", false, "Generate test data only")
	load := flag.Bool("load", false, "Load data into databases")
	runBenchmark := flag.Bool("benchmark", false, "Run benchmarks")

	postgresDSN := flag.String("postgres", "postgres://benchmark:benchmark@localhost:5432/benchmark?sslmode=disable", "PostgreSQL DSN")
	elasticURL := flag.String("elastic", "http://localhost:9200", "Elasticsearch URL")

	workers := flag.Int("workers", 8, "Number of concurrent workers")
	iterations := flag.Int("iterations", 1000, "Number of query iterations per scenario")

	flag.Parse()

	ctx := context.Background()

	if *generate {
		fmt.Println("📊 Data generation should be run via generator/cmd/main.go")
		fmt.Println("   Usage: go run generator/cmd/main.go -config config.yaml -out ./output")
	}

	if *load {
		if err := loadData(ctx, *postgresDSN, *elasticURL); err != nil {
			log.Fatalf("Failed to load data: %v", err)
		}
	}

	if *runBenchmark {
		if err := runBenchmarks(*postgresDSN, *elasticURL, *workers, *iterations); err != nil {
			log.Fatalf("Benchmark failed: %v", err)
		}
	}

	if !*generate && !*load && !*runBenchmark {
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

func loadPostgres(ctx context.Context, dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return err
	}

	// Load data using COPY
	tables := []string{"categories", "users", "products", "orders", "order_items"}

	for _, table := range tables {
		fmt.Printf("  Loading %s...\n", table)

		filePath := fmt.Sprintf("./output/%s.csv", table)

		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("open %s: %w", table, err)
		}

		// Use COPY FROM STDIN
		txn, err := db.Begin()
		if err != nil {
			file.Close()
			return err
		}

		stmt, err := txn.PrepareContext(ctx, fmt.Sprintf("COPY %s FROM STDIN WITH (FORMAT csv)", table))
		if err != nil {
			txn.Rollback()
			file.Close()
			return err
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				continue
			}

			fields := strings.Split(line, ",")
			args := make([]interface{}, len(fields))
			for i, f := range fields {
				if f == "\\N" {
					args[i] = nil
				} else {
					args[i] = f
				}
			}

			if _, err := stmt.Exec(args...); err != nil {
				stmt.Close()
				txn.Rollback()
				file.Close()
				return err
			}
		}

		if err := scanner.Err(); err != nil {
			stmt.Close()
			txn.Rollback()
			file.Close()
			return err
		}

		if _, err := stmt.Exec(); err != nil {
			stmt.Close()
			txn.Rollback()
			file.Close()
			return err
		}

		stmt.Close()

		if err := txn.Commit(); err != nil {
			file.Close()
			return err
		}

		file.Close()
	}

	// Analyze tables
	fmt.Println("  Running ANALYZE...")
	if _, err := db.Exec("ANALYZE"); err != nil {
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
			fmt.Printf("  Indexed %d documents...\n", totalIndexed)
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
	req, err := http.NewRequestWithContext(ctx, "POST", url+"/_bulk", bytes.NewReader(body))
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

	runner := benchmark.NewRunner(pgExecutor, elasticExecutor, workers)

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
