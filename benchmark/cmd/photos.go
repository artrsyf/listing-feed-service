package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gopkg.in/yaml.v3"
)

type photoConfig struct {
	Photos        int   `yaml:"photos"`
	AvgSizeKB     int   `yaml:"avg_size_kb"`
	SizeJitterPct int   `yaml:"size_jitter_pct"`
	BatchSize     int   `yaml:"batch_size"`
	Seed          int64 `yaml:"seed"`
}

type photoMetric struct {
	Backend  string
	Scenario string
	Requests int
	Duration time.Duration
	Avg      time.Duration
	P50      time.Duration
	P95      time.Duration
	P99      time.Duration
	Max      time.Duration
	RPS      float64
	MBps     float64
}

func initPhotos(ctx context.Context, postgresDSN string) error {
	db, err := sql.Open("postgres", postgresDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	ddl, err := os.ReadFile("postgres/init/photos_ddl.sql")
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, string(ddl))
	return err
}

func loadPhotos(ctx context.Context, postgresDSN string, cfgPath string) error {
	cfg, err := loadPhotoConfig(cfgPath)
	if err != nil {
		return err
	}

	db, err := sql.Open("postgres", postgresDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	minioClient, bucket, err := newMinioClient()
	if err != nil {
		return err
	}
	if err := ensurePhotoBucket(ctx, minioClient, bucket); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, "TRUNCATE photo_blobs"); err != nil {
		return err
	}

	fmt.Printf("Loading %d photos, avg_size=%dKB\n", cfg.Photos, cfg.AvgSizeKB)
	start := time.Now()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO photo_blobs(id, listing_id, content_type, data, size_bytes)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for id := 1; id <= cfg.Photos; id++ {
		payload := photoPayload(cfg, id)
		objectName := photoObjectName(id)

		if _, err := stmt.ExecContext(ctx, id, (id-1)/5+1, "image/jpeg", payload, len(payload)); err != nil {
			tx.Rollback()
			return fmt.Errorf("insert postgres photo %d: %w", id, err)
		}

		_, err := minioClient.PutObject(ctx, bucket, objectName, bytes.NewReader(payload), int64(len(payload)), minio.PutObjectOptions{
			ContentType: "image/jpeg",
		})
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("put minio photo %d: %w", id, err)
		}

		if id%cfg.BatchSize == 0 {
			if err := tx.Commit(); err != nil {
				return err
			}
			tx, err = db.BeginTx(ctx, nil)
			if err != nil {
				return err
			}
			stmt, err = tx.PrepareContext(ctx, `
				INSERT INTO photo_blobs(id, listing_id, content_type, data, size_bytes)
				VALUES ($1, $2, $3, $4, $5)
			`)
			if err != nil {
				tx.Rollback()
				return err
			}
		}

		if id%1000 == 0 {
			fmt.Printf("  Loaded %d photos...\n", id)
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, "ANALYZE photo_blobs"); err != nil {
		return err
	}

	fmt.Printf("Photos loaded in %s\n", time.Since(start))
	return nil
}

func runPhotoBenchmarks(ctx context.Context, postgresDSN string, cfgPath string, workers, iterations int) error {
	cfg, err := loadPhotoConfig(cfgPath)
	if err != nil {
		return err
	}

	db, err := sql.Open("postgres", postgresDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	minioClient, bucket, err := newMinioClient()
	if err != nil {
		return err
	}

	storage, err := collectPhotoStorage(ctx, db, minioClient, bucket)
	if err != nil {
		return err
	}
	fmt.Println("\n=== PHOTO STORAGE SIZE ===")
	fmt.Printf("PostgreSQL total: %.2f MB\n", float64(storage.pgTotalBytes)/1024/1024)
	fmt.Printf("PostgreSQL table: %.2f MB\n", float64(storage.pgTableBytes)/1024/1024)
	fmt.Printf("PostgreSQL indexes: %.2f MB\n", float64(storage.pgIndexBytes)/1024/1024)
	fmt.Printf("PostgreSQL raw payload: %.2f MB\n", float64(storage.pgPayloadBytes)/1024/1024)
	fmt.Printf("MinIO objects total: %.2f MB\n", float64(storage.minioPayloadBytes)/1024/1024)
	fmt.Printf("MinIO object count: %d\n", storage.minioObjects)

	scenarios := []struct {
		backend  string
		scenario string
		read     func(context.Context, int) (int64, error)
	}{
		{
			"postgres",
			"photo_pg_random_read",
			func(ctx context.Context, id int) (int64, error) {
				var data []byte
				if err := db.QueryRowContext(ctx, "SELECT data FROM photo_blobs WHERE id = $1", id).Scan(&data); err != nil {
					return 0, err
				}
				return int64(len(data)), nil
			},
		},
		{
			"minio",
			"photo_minio_random_read",
			func(ctx context.Context, id int) (int64, error) {
				obj, err := minioClient.GetObject(ctx, bucket, photoObjectName(id), minio.GetObjectOptions{})
				if err != nil {
					return 0, err
				}
				defer obj.Close()
				n, err := io.Copy(io.Discard, obj)
				return n, err
			},
		},
	}

	results := make([]photoMetric, 0, len(scenarios))
	for _, scenario := range scenarios {
		metric := measurePhotoScenario(ctx, scenario.backend, scenario.scenario, cfg.Photos, workers, iterations, scenario.read)
		results = append(results, metric)
		printPhotoMetric(metric)
	}

	sort.Slice(results, func(i, j int) bool { return results[i].MBps > results[j].MBps })
	fmt.Println("\n=== PHOTO TOP BY THROUGHPUT ===")
	for i, result := range results {
		fmt.Printf("%d. %s/%s %.2f MB/s %.2f RPS avg=%s p95=%s\n", i+1, result.Backend, result.Scenario, result.MBps, result.RPS, result.Avg, result.P95)
	}

	return nil
}

type photoStorage struct {
	pgTotalBytes      int64
	pgTableBytes      int64
	pgIndexBytes      int64
	pgPayloadBytes    int64
	minioPayloadBytes int64
	minioObjects      int
}

func collectPhotoStorage(ctx context.Context, db *sql.DB, client *minio.Client, bucket string) (*photoStorage, error) {
	var s photoStorage
	err := db.QueryRowContext(ctx, `
		SELECT
			pg_total_relation_size('photo_blobs'),
			pg_relation_size('photo_blobs'),
			pg_indexes_size('photo_blobs'),
			COALESCE(SUM(size_bytes), 0)
		FROM photo_blobs
	`).Scan(&s.pgTotalBytes, &s.pgTableBytes, &s.pgIndexBytes, &s.pgPayloadBytes)
	if err != nil {
		return nil, err
	}

	for object := range client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Recursive: true}) {
		if object.Err != nil {
			return nil, object.Err
		}
		s.minioPayloadBytes += object.Size
		s.minioObjects++
	}
	return &s, nil
}

func measurePhotoScenario(ctx context.Context, backend, scenario string, maxID, workers, iterations int, read func(context.Context, int) (int64, error)) photoMetric {
	if workers < 1 {
		workers = 1
	}
	if iterations < 1 {
		iterations = 1
	}

	jobs := make(chan int, iterations)
	type result struct {
		latency time.Duration
		bytes   int64
	}
	results := make(chan result, iterations)

	start := time.Now()
	for w := 0; w < workers; w++ {
		go func(workerID int) {
			for id := range jobs {
				readStart := time.Now()
				n, err := read(ctx, id)
				if err != nil {
					fmt.Printf("Read error: %v\n", err)
					continue
				}
				results <- result{latency: time.Since(readStart), bytes: n}
			}
		}(w)
	}

	r := rand.New(rand.NewSource(99))
	for i := 0; i < iterations; i++ {
		jobs <- r.Intn(maxID) + 1
	}
	close(jobs)

	latencies := make([]time.Duration, 0, iterations)
	var totalBytes int64
	for len(latencies) < iterations {
		select {
		case result := <-results:
			latencies = append(latencies, result.latency)
			totalBytes += result.bytes
		case <-time.After(120 * time.Second):
			goto done
		}
	}

done:
	duration := time.Since(start)
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	var sum time.Duration
	var max time.Duration
	for _, latency := range latencies {
		sum += latency
		if latency > max {
			max = latency
		}
	}
	avg := time.Duration(0)
	if len(latencies) > 0 {
		avg = sum / time.Duration(len(latencies))
	}

	return photoMetric{
		Backend:  backend,
		Scenario: scenario,
		Requests: len(latencies),
		Duration: duration,
		Avg:      avg,
		P50:      durationPercentile(latencies, 50),
		P95:      durationPercentile(latencies, 95),
		P99:      durationPercentile(latencies, 99),
		Max:      max,
		RPS:      float64(len(latencies)) / duration.Seconds(),
		MBps:     float64(totalBytes) / 1024 / 1024 / duration.Seconds(),
	}
}

func printPhotoMetric(metric photoMetric) {
	fmt.Println("\n========== PHOTO BENCHMARK REPORT ==========")
	fmt.Printf("Backend: %s\n", metric.Backend)
	fmt.Printf("Scenario: %s\n", metric.Scenario)
	fmt.Printf("Total requests: %d\n", metric.Requests)
	fmt.Printf("Duration (sec): %.3f\n", metric.Duration.Seconds())
	fmt.Printf("Throughput (req/sec): %.2f\n", metric.RPS)
	fmt.Printf("Throughput (MB/sec): %.2f\n", metric.MBps)
	fmt.Printf("Avg latency: %s\n", metric.Avg)
	fmt.Printf("p50: %s\n", metric.P50)
	fmt.Printf("p95: %s\n", metric.P95)
	fmt.Printf("p99: %s\n", metric.P99)
	fmt.Printf("Max latency: %s\n", metric.Max)
}

func loadPhotoConfig(path string) (*photoConfig, error) {
	cfg := &photoConfig{
		Photos:        5000,
		AvgSizeKB:     128,
		SizeJitterPct: 25,
		BatchSize:     100,
		Seed:          42,
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Photos <= 0 || cfg.AvgSizeKB <= 0 || cfg.BatchSize <= 0 {
		return nil, fmt.Errorf("photo config values must be positive")
	}
	return cfg, nil
}

func photoPayload(cfg *photoConfig, id int) []byte {
	size := photoSize(cfg, id)
	payload := make([]byte, size)
	seed := cfg.Seed + int64(id)*7919
	var offset int
	for offset < len(payload) {
		sum := sha256.Sum256([]byte(fmt.Sprintf("%d:%d", seed, offset)))
		offset += copy(payload[offset:], sum[:])
	}
	return payload
}

func photoSize(cfg *photoConfig, id int) int {
	base := cfg.AvgSizeKB * 1024
	if cfg.SizeJitterPct <= 0 {
		return base
	}
	r := rand.New(rand.NewSource(cfg.Seed + int64(id)*104729))
	spread := base * cfg.SizeJitterPct / 100
	return base - spread + r.Intn(spread*2+1)
}

func photoObjectName(id int) string {
	return fmt.Sprintf("photos/%06d/%012d.jpg", id/1000, id)
}

func newMinioClient() (*minio.Client, string, error) {
	endpoint := envOrDefault("MINIO_ENDPOINT", "localhost:9000")
	accessKey := envOrDefault("MINIO_ACCESS_KEY", "benchmark")
	secretKey := envOrDefault("MINIO_SECRET_KEY", "benchmark123")
	bucket := envOrDefault("MINIO_BUCKET", "photos")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, "", err
	}
	return client, bucket, nil
}

func ensurePhotoBucket(ctx context.Context, client *minio.Client, bucket string) error {
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}

	for object := range client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Recursive: true}) {
		if object.Err != nil {
			return object.Err
		}
		if err := client.RemoveObject(ctx, bucket, object.Key, minio.RemoveObjectOptions{}); err != nil {
			return err
		}
	}
	return nil
}
