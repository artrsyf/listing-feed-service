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

type Metrics struct {
	mu sync.Mutex

	Latencies []time.Duration

	Count int
}

func (m *Metrics) Add(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Latencies = append(m.Latencies, d)
	m.Count++
}

func (r *Runner) RunScenario(name string, queries []string) {

	fmt.Printf("▶ Running scenario: %s\n", name)

	var wg sync.WaitGroup
	jobs := make(chan string, len(queries))

	// workers
	for i := 0; i < r.workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for q := range jobs {
				start := time.Now()

				_, err := r.postgres.RunQuery(context.Background(), q)
				if err != nil {
					continue
				}

				r.metrics.Add(time.Since(start))
			}
		}()
	}

	for _, q := range queries {
		jobs <- q
	}

	close(jobs)
	wg.Wait()

	fmt.Println("✅ Scenario finished")
}

func setHashJoin(pg Executor) {
	pg.RunQuery(context.Background(),
		"SET enable_hashjoin = on; SET enable_mergejoin = off; SET enable_nestloop = off;")
}

func setMergeJoin(pg Executor) {
	pg.RunQuery(context.Background(),
		"SET enable_mergejoin = on; SET enable_hashjoin = off; SET enable_nestloop = off;")
}

func setNestedLoop(pg Executor) {
	pg.RunQuery(context.Background(),
		"SET enable_nestloop = on; SET enable_hashjoin = off; SET enable_mergejoin = off;")
}

var JoinScenario = []string{
	`SELECT u.country, c.name, SUM(oi.quantity * oi.price)
	FROM users u
	JOIN orders o ON u.id = o.user_id
	JOIN order_items oi ON o.id = oi.order_id
	JOIN products p ON oi.product_id = p.id
	JOIN categories c ON p.category_id = c.id
	WHERE o.created_at > NOW() - interval '30 days'
	GROUP BY u.country, c.name`,
}

var PointLookupScenario = []string{
	`SELECT * FROM orders WHERE id = $1`,
}

var ElasticAggScenario = []string{
	`GET orders/_search
	{
	  "size": 0,
	  "aggs": {
	    "by_country": {
	      "terms": { "field": "user.country" }
	    }
	  }
	}`,
}

func (r *Runner) Report() {

	var total time.Duration

	for _, l := range r.metrics.Latencies {
		total += l
	}

	avg := total / time.Duration(len(r.metrics.Latencies))

	fmt.Println("📊 RESULTS")
	fmt.Println("Count:", len(r.metrics.Latencies))
	fmt.Println("Avg latency:", avg)
}
