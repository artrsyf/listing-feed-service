package main

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

type listingConfig struct {
	Listings      int   `yaml:"listings"`
	Sellers       int   `yaml:"sellers"`
	Categories    int   `yaml:"categories"`
	BatchSize     int   `yaml:"batch_size"`
	TimeRangeDays int   `yaml:"time_range_days"`
	Seed          int64 `yaml:"seed"`
}

type listingMetric struct {
	Name     string
	Backend  string
	Requests int
	Duration time.Duration
	Avg      time.Duration
	P50      time.Duration
	P95      time.Duration
	P99      time.Duration
	Max      time.Duration
	RPS      float64
}

func generateListingData(configPath, outDir string) error {
	cfg, err := loadListingConfig(configPath)
	if err != nil {
		return err
	}
	rand.Seed(cfg.Seed)

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	start := time.Now()
	fmt.Printf("Generating listing dataset: listings=%d sellers=%d categories=%d\n", cfg.Listings, cfg.Sellers, cfg.Categories)

	if err := writeListingCategories(outDir, cfg); err != nil {
		return err
	}
	if err := writeListingSellers(outDir, cfg); err != nil {
		return err
	}
	if err := writeListings(outDir, cfg); err != nil {
		return err
	}

	fmt.Printf("Listing dataset generated in %s\n", time.Since(start))
	return nil
}

func loadListingConfig(path string) (*listingConfig, error) {
	cfg := &listingConfig{
		Listings:      300000,
		Sellers:       50000,
		Categories:    2000,
		BatchSize:     25000,
		TimeRangeDays: 180,
		Seed:          42,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Listings <= 0 || cfg.Sellers <= 0 || cfg.Categories <= 0 || cfg.BatchSize <= 0 {
		return nil, fmt.Errorf("listing config values must be positive")
	}
	return cfg, nil
}

func writeListingCategories(outDir string, cfg *listingConfig) error {
	file, writer, err := newCSV(filepath.Join(outDir, "listing_categories.csv"))
	if err != nil {
		return err
	}
	defer file.Close()
	defer writer.Flush()

	rootCount := cfg.Categories / 20
	if rootCount < 1 {
		rootCount = 1
	}

	for i := 1; i <= cfg.Categories; i++ {
		parent := `\N`
		if i > rootCount {
			parent = fmt.Sprint(rand.Intn(i-1) + 1)
		}
		if err := writer.Write([]string{fmt.Sprint(i), parent, fmt.Sprintf("category_%d", i)}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func writeListingSellers(outDir string, cfg *listingConfig) error {
	file, writer, err := newCSV(filepath.Join(outDir, "listing_sellers.csv"))
	if err != nil {
		return err
	}
	defer file.Close()
	defer writer.Flush()

	for i := 1; i <= cfg.Sellers; i++ {
		createdAt := randomListingTime(cfg).Format(time.RFC3339)
		if err := writer.Write([]string{
			fmt.Sprint(i),
			randomListingCity(),
			fmt.Sprintf("%.2f", 2.5+rand.Float64()*2.5),
			createdAt,
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func writeListings(outDir string, cfg *listingConfig) error {
	listingsFile, listingsCSV, err := newCSV(filepath.Join(outDir, "listings.csv"))
	if err != nil {
		return err
	}
	defer listingsFile.Close()
	defer listingsCSV.Flush()

	attrsFile, attrsCSV, err := newCSV(filepath.Join(outDir, "listing_attribute_values.csv"))
	if err != nil {
		return err
	}
	defer attrsFile.Close()
	defer attrsCSV.Flush()

	ndjsonFile, err := os.Create(filepath.Join(outDir, "listings.ndjson"))
	if err != nil {
		return err
	}
	defer ndjsonFile.Close()
	ndjson := bufio.NewWriterSize(ndjsonFile, 4*1024*1024)
	defer ndjson.Flush()

	brands := []string{"apple", "samsung", "xiaomi", "sony", "bosch", "lg", "lenovo", "nike", "adidas", "ikea"}
	colors := []string{"black", "white", "gray", "blue", "red", "green"}
	materials := []string{"metal", "plastic", "wood", "textile", "glass"}

	for i := 1; i <= cfg.Listings; i++ {
		sellerID := rand.Intn(cfg.Sellers) + 1
		categoryID := rand.Intn(cfg.Categories) + 1
		city := randomListingCity()
		price := 100 + rand.ExpFloat64()*5000
		if price > 500000 {
			price = 500000
		}
		condition := rand.Intn(4) + 1
		delivery := rand.Float64() < 0.72
		promoted := rand.Float64() < 0.08
		status := 1
		if rand.Float64() < 0.04 {
			status = 0
		}
		createdAt := randomListingTime(cfg)
		title := fmt.Sprintf("%s listing %d in %s", brands[rand.Intn(len(brands))], i, city)
		description := fmt.Sprintf("marketplace searchable item %d with delivery filters attributes and seller rating", i)

		if err := listingsCSV.Write([]string{
			fmt.Sprint(i),
			fmt.Sprint(sellerID),
			fmt.Sprint(categoryID),
			city,
			title,
			description,
			fmt.Sprintf("%.2f", price),
			fmt.Sprint(condition),
			fmt.Sprint(delivery),
			fmt.Sprint(promoted),
			fmt.Sprint(status),
			createdAt.Format(time.RFC3339),
			createdAt.Add(time.Duration(rand.Intn(96)) * time.Hour).Format(time.RFC3339),
		}); err != nil {
			return err
		}

		attrs := map[string]string{
			"brand":    brands[rand.Intn(len(brands))],
			"color":    colors[rand.Intn(len(colors))],
			"material": materials[rand.Intn(len(materials))],
			"size":     fmt.Sprint(rand.Intn(8) + 1),
			"year":     fmt.Sprint(2017 + rand.Intn(10)),
		}
		for key, value := range attrs {
			if err := attrsCSV.Write([]string{fmt.Sprint(i), key, value}); err != nil {
				return err
			}
		}

		doc := map[string]interface{}{
			"listing_id":  i,
			"seller":      map[string]interface{}{"id": sellerID, "rating": 2.5 + rand.Float64()*2.5},
			"category_id": categoryID,
			"city":        city,
			"title":       title,
			"description": description,
			"price":       price,
			"condition":   condition,
			"delivery":    delivery,
			"promoted":    promoted,
			"status":      status,
			"created_at":  createdAt,
			"attrs":       attrs,
		}
		docBytes, err := json.Marshal(doc)
		if err != nil {
			return err
		}
		if _, err := ndjson.WriteString(`{"index":{}}` + "\n" + string(docBytes) + "\n"); err != nil {
			return err
		}

		if i%100000 == 0 {
			fmt.Printf("  Generated %d listings...\n", i)
		}
	}

	return firstErr(listingsCSV.Error(), attrsCSV.Error(), ndjson.Flush())
}

func initListingElastic(ctx context.Context, url string) error {
	settings, err := os.ReadFile("elasticsearch/listing-index-settings.json")
	if err != nil {
		return err
	}
	return recreateElasticIndex(ctx, url, "listings", settings)
}

func loadListingsData(ctx context.Context, postgresDSN, elasticURL string) error {
	db, err := sql.Open("postgres", postgresDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.ExecContext(ctx, "TRUNCATE listing_attribute_values, listings, listing_sellers, listing_categories"); err != nil {
		return err
	}

	loads := []struct {
		table   string
		columns string
		file    string
	}{
		{"listing_categories", "id, parent_id, name", "/data/listing_categories.csv"},
		{"listing_sellers", "id, city, rating, created_at", "/data/listing_sellers.csv"},
		{"listings", "id, seller_id, category_id, city, title, description, price, condition, delivery, promoted, status, created_at, updated_at", "/data/listings.csv"},
		{"listing_attribute_values", "listing_id, attr_key, attr_value", "/data/listing_attribute_values.csv"},
	}
	for _, load := range loads {
		fmt.Printf("  Loading %s...\n", load.table)
		query := fmt.Sprintf("COPY %s (%s) FROM '%s' WITH (FORMAT csv, NULL '\\N')", load.table, load.columns, load.file)
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("copy %s: %w", load.table, err)
		}
	}
	if _, err := db.ExecContext(ctx, "ANALYZE listing_categories; ANALYZE listing_sellers; ANALYZE listings; ANALYZE listing_attribute_values"); err != nil {
		return err
	}

	return bulkLoadElastic(ctx, elasticURL, "listings", "./output/listings.ndjson", 50000)
}

func runListingBenchmarks(ctx context.Context, postgresDSN, elasticURL string, workers, iterations int) error {
	db, err := sql.Open("postgres", postgresDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	scenarios := []struct {
		name    string
		backend string
		query   string
	}{
		{"listing_pg_search_page", "postgres", listingPostgresSearchQuery()},
		{"listing_pg_facets", "postgres", listingPostgresFacetQuery()},
		{"listing_es_search_page", "elastic", listingElasticSearchQuery()},
		{"listing_es_facets", "elastic", listingElasticFacetQuery()},
	}

	results := make([]listingMetric, 0, len(scenarios))
	for _, scenario := range scenarios {
		fmt.Printf("\n--- %s ---\n", scenario.name)
		var metric listingMetric
		if scenario.backend == "postgres" {
			metric = measureListingScenario(ctx, scenario.backend, scenario.name, workers, iterations, func(ctx context.Context) error {
				rows, err := db.QueryContext(ctx, scenario.query)
				if err != nil {
					return err
				}
				defer rows.Close()
				for rows.Next() {
				}
				return rows.Err()
			})
		} else {
			client := &http.Client{Timeout: 30 * time.Second}
			metric = measureListingScenario(ctx, scenario.backend, scenario.name, workers, iterations, func(ctx context.Context) error {
				req, err := http.NewRequestWithContext(ctx, "POST", elasticURL+"/listings/_search", bytes.NewReader([]byte(scenario.query)))
				if err != nil {
					return err
				}
				req.Header.Set("Content-Type", "application/json")
				resp, err := client.Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				io.Copy(io.Discard, resp.Body)
				if resp.StatusCode != 200 {
					return fmt.Errorf("elastic status: %s", resp.Status)
				}
				return nil
			})
		}
		results = append(results, metric)
		printListingMetric(metric)
	}

	sort.Slice(results, func(i, j int) bool { return results[i].RPS > results[j].RPS })
	fmt.Println("\n=== LISTING TOP BY RPS ===")
	for i, result := range results {
		fmt.Printf("%d. %s/%s RPS=%.2f avg=%s p95=%s\n", i+1, result.Backend, result.Name, result.RPS, result.Avg, result.P95)
	}

	return nil
}

func measureListingScenario(ctx context.Context, backend, name string, workers, iterations int, fn func(context.Context) error) listingMetric {
	if workers < 1 {
		workers = 1
	}
	if iterations < 1 {
		iterations = 1
	}

	jobs := make(chan struct{}, iterations)
	latencies := make(chan time.Duration, iterations)
	start := time.Now()

	for w := 0; w < workers; w++ {
		go func() {
			for range jobs {
				queryStart := time.Now()
				if err := fn(ctx); err != nil {
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
		case <-time.After(60 * time.Second):
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

	return listingMetric{
		Name:     name,
		Backend:  backend,
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

func listingPostgresSearchQuery() string {
	return `
SELECT l.id, l.title, l.price, l.city, l.created_at, s.rating
FROM listings l
JOIN listing_sellers s ON s.id = l.seller_id
JOIN listing_attribute_values brand ON brand.listing_id = l.id AND brand.attr_key = 'brand'
JOIN listing_attribute_values color ON color.listing_id = l.id AND color.attr_key = 'color'
JOIN listing_attribute_values material ON material.listing_id = l.id AND material.attr_key = 'material'
WHERE l.status = 1
  AND l.category_id BETWEEN 100 AND 900
  AND l.city IN ('moscow', 'spb', 'kazan')
  AND l.price BETWEEN 500 AND 75000
  AND l.delivery = true
  AND l.condition IN (2, 3, 4)
  AND s.rating >= 4.0
  AND brand.attr_value IN ('apple', 'samsung', 'xiaomi', 'sony')
  AND color.attr_value IN ('black', 'white', 'gray')
  AND material.attr_value IN ('metal', 'plastic', 'glass')
  AND (l.title ILIKE '%listing%' OR l.description ILIKE '%delivery%')
ORDER BY l.promoted DESC, l.created_at DESC
LIMIT 50`
}

func listingPostgresFacetQuery() string {
	return `
SELECT l.city, brand.attr_value AS brand, COUNT(*) AS listings, percentile_disc(0.5) WITHIN GROUP (ORDER BY l.price) AS median_price
FROM listings l
JOIN listing_sellers s ON s.id = l.seller_id
JOIN listing_attribute_values brand ON brand.listing_id = l.id AND brand.attr_key = 'brand'
JOIN listing_attribute_values color ON color.listing_id = l.id AND color.attr_key = 'color'
WHERE l.status = 1
  AND l.category_id BETWEEN 100 AND 900
  AND l.price BETWEEN 500 AND 75000
  AND l.delivery = true
  AND s.rating >= 4.0
  AND color.attr_value IN ('black', 'white', 'gray')
GROUP BY l.city, brand.attr_value
ORDER BY listings DESC
LIMIT 40`
}

func listingElasticSearchQuery() string {
	return `{
  "size": 50,
  "query": {
    "bool": {
      "must": [
        {"multi_match": {"query": "listing delivery", "fields": ["title", "description"]}}
      ],
      "filter": [
        {"term": {"status": 1}},
        {"range": {"category_id": {"gte": 100, "lte": 900}}},
        {"terms": {"city": ["moscow", "spb", "kazan"]}},
        {"range": {"price": {"gte": 500, "lte": 75000}}},
        {"term": {"delivery": true}},
        {"terms": {"condition": [2, 3, 4]}},
        {"range": {"seller.rating": {"gte": 4.0}}},
        {"terms": {"attrs.brand": ["apple", "samsung", "xiaomi", "sony"]}},
        {"terms": {"attrs.color": ["black", "white", "gray"]}},
        {"terms": {"attrs.material": ["metal", "plastic", "glass"]}}
      ]
    }
  },
  "sort": [
    {"promoted": "desc"},
    {"created_at": "desc"}
  ]
}`
}

func listingElasticFacetQuery() string {
	return `{
  "size": 0,
  "query": {
    "bool": {
      "filter": [
        {"term": {"status": 1}},
        {"range": {"category_id": {"gte": 100, "lte": 900}}},
        {"range": {"price": {"gte": 500, "lte": 75000}}},
        {"term": {"delivery": true}},
        {"range": {"seller.rating": {"gte": 4.0}}},
        {"terms": {"attrs.color": ["black", "white", "gray"]}}
      ]
    }
  },
  "aggs": {
    "by_city": {"terms": {"field": "city", "size": 20}},
    "by_brand": {"terms": {"field": "attrs.brand.keyword", "size": 20}},
    "price_stats": {"percentiles": {"field": "price", "percents": [50]}}
  }
}`
}

func recreateElasticIndex(ctx context.Context, url, index string, settings []byte) error {
	client := &http.Client{Timeout: 30 * time.Second}
	deleteReq, err := http.NewRequestWithContext(ctx, "DELETE", url+"/"+index, nil)
	if err != nil {
		return err
	}
	deleteResp, err := client.Do(deleteReq)
	if err != nil {
		return err
	}
	io.Copy(io.Discard, deleteResp.Body)
	deleteResp.Body.Close()

	putReq, err := http.NewRequestWithContext(ctx, "PUT", url+"/"+index, bytes.NewReader(settings))
	if err != nil {
		return err
	}
	putReq.Header.Set("Content-Type", "application/json")
	putResp, err := client.Do(putReq)
	if err != nil {
		return err
	}
	defer putResp.Body.Close()
	body, _ := io.ReadAll(putResp.Body)
	if putResp.StatusCode < 200 || putResp.StatusCode >= 300 {
		return fmt.Errorf("create index %s failed: %s - %s", index, putResp.Status, string(body))
	}
	return nil
}

func bulkLoadElastic(ctx context.Context, url, index, filePath string, reportEvery int) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	client := &http.Client{Timeout: 5 * time.Minute}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	var bulk bytes.Buffer
	lineCount := 0
	total := 0

	for scanner.Scan() {
		bulk.WriteString(scanner.Text())
		bulk.WriteByte('\n')
		lineCount++
		if lineCount%2000 == 0 {
			if err := sendBulkToIndex(ctx, client, url, index, bulk.Bytes()); err != nil {
				return err
			}
			total += lineCount / 2
			if total%reportEvery == 0 {
				fmt.Printf("  Indexed %d listings...\n", total)
			}
			bulk.Reset()
			lineCount = 0
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if bulk.Len() > 0 {
		if err := sendBulkToIndex(ctx, client, url, index, bulk.Bytes()); err != nil {
			return err
		}
		total += lineCount / 2
	}

	refreshReq, _ := http.NewRequestWithContext(ctx, "POST", url+"/"+index+"/_refresh", nil)
	refreshResp, err := client.Do(refreshReq)
	if err != nil {
		return err
	}
	refreshResp.Body.Close()
	fmt.Printf("  Elasticsearch %s: %d documents indexed\n", index, total)
	return nil
}

func sendBulkToIndex(ctx context.Context, client *http.Client, url, index string, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, "POST", url+"/"+index+"/_bulk", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-ndjson")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("bulk index failed: %s - %s", resp.Status, string(bodyBytes))
	}
	var result struct {
		Errors bool `json:"errors"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return err
	}
	if result.Errors {
		return fmt.Errorf("bulk index completed with item errors")
	}
	return nil
}

func newCSV(path string) (*os.File, *csv.Writer, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	writer := csv.NewWriter(file)
	return file, writer, nil
}

func randomListingCity() string {
	cities := []string{"moscow", "spb", "kazan", "novosibirsk", "ekb", "samara", "sochi", "perm", "ufa", "voronezh"}
	return cities[rand.Intn(len(cities))]
}

func randomListingTime(cfg *listingConfig) time.Time {
	return time.Now().Add(-time.Duration(rand.Intn(cfg.TimeRangeDays*24)) * time.Hour)
}

func firstErr(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func durationPercentile(values []time.Duration, p float64) time.Duration {
	if len(values) == 0 {
		return 0
	}
	index := int(float64(len(values))*p/100.0+0.999999) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(values) {
		index = len(values) - 1
	}
	return values[index]
}

func printListingMetric(metric listingMetric) {
	fmt.Println("========== LISTING BENCHMARK REPORT ==========")
	fmt.Printf("Backend: %s\n", metric.Backend)
	fmt.Printf("Scenario: %s\n", metric.Name)
	fmt.Printf("Total requests: %d\n", metric.Requests)
	fmt.Printf("Duration (sec): %.3f\n", metric.Duration.Seconds())
	fmt.Printf("Throughput (req/sec): %.2f\n", metric.RPS)
	fmt.Printf("Avg latency: %s\n", metric.Avg)
	fmt.Printf("p50: %s\n", metric.P50)
	fmt.Printf("p95: %s\n", metric.P95)
	fmt.Printf("p99: %s\n", metric.P99)
	fmt.Printf("Max latency: %s\n", metric.Max)
}
