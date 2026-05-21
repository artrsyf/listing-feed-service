package benchmark

import (
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

	if len(m.latencies) == 0 {
		return 0
	}

	sort.Slice(m.latencies, func(i, j int) bool {
		return m.latencies[i] < m.latencies[j]
	})

	index := int(math.Ceil((p / 100.0) * float64(len(m.latencies))))
	if index >= len(m.latencies) {
		index = len(m.latencies) - 1
	}

	return m.latencies[index]
}

func (m *Metrics) Report() {

	m.mu.Lock()
	defer m.mu.Unlock()

	duration := m.endTime.Sub(m.startTime).Seconds()

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

	println("========== BENCHMARK REPORT ==========")
	println("Total requests:", len(m.latencies))
	println("Duration (sec):", duration)
	println("Throughput (req/sec):", rps)

	println("Avg latency:", avg.String())
	println("Min latency:", min.String())
	println("Max latency:", max.String())

	println("p50:", m.Percentile(50).String())
	println("p95:", m.Percentile(95).String())
	println("p99:", m.Percentile(99).String())
}
