package benchmark

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

type Metrics struct {
	mu sync.Mutex

	latencies []time.Duration
	startTime time.Time
	endTime   time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{
		latencies: make([]time.Duration, 0, 1000000),
		startTime: time.Now(),
	}
}

func (m *Metrics) Add(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.latencies = append(m.latencies, d)
}

func (m *Metrics) Finish() {
	m.endTime = time.Now()
}

func (m *Metrics) Percentile(p float64) time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()

	return percentile(m.latencies, p)
}

func (m *Metrics) Report() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.latencies) == 0 {
		fmt.Println("========== BENCHMARK REPORT ==========")
		fmt.Println("Total requests: 0")
		fmt.Println("No successful requests recorded")
		return
	}

	sort.Slice(m.latencies, func(i, j int) bool {
		return m.latencies[i] < m.latencies[j]
	})

	duration := m.endTime.Sub(m.startTime).Seconds()
	if duration <= 0 {
		duration = time.Since(m.startTime).Seconds()
	}

	var sum time.Duration
	min := time.Hour
	max := time.Duration(0)

	for _, l := range m.latencies {
		sum += l
		if l < min {
			min = l
		}
		if l > max {
			max = l
		}
	}

	avg := sum / time.Duration(len(m.latencies))
	rps := float64(len(m.latencies)) / duration

	fmt.Println("========== BENCHMARK REPORT ==========")
	fmt.Printf("Total requests: %d\n", len(m.latencies))
	fmt.Printf("Duration (sec): %.3f\n", duration)
	fmt.Printf("Throughput (req/sec): %.2f\n", rps)

	fmt.Printf("Avg latency: %s\n", avg)
	fmt.Printf("Min latency: %s\n", min)
	fmt.Printf("Max latency: %s\n", max)

	fmt.Printf("p50: %s\n", percentile(m.latencies, 50))
	fmt.Printf("p95: %s\n", percentile(m.latencies, 95))
	fmt.Printf("p99: %s\n", percentile(m.latencies, 99))
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}

	index := int(math.Ceil((p/100.0)*float64(len(sorted)))) - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}
