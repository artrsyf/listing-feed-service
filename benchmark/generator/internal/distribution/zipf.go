package distribution

import (
	"math"
	"math/rand"
)

// ZipfGenerator generates skewed distribution
// based on Zipf's law (heavy tail distribution)
//
// Useful for:
// - user activity skew
// - product popularity
// - order concentration
type ZipfGenerator struct {
	n   float64
	s   float64
	v   float64
	rng *rand.Rand
	cdf []float64
}

// NewZipf creates precomputed CDF-based Zipf generator
func NewZipf(n int, s float64, seed int64) *ZipfGenerator {
	z := &ZipfGenerator{
		n:   float64(n),
		s:   s,
		v:   1.0,
		rng: rand.New(rand.NewSource(seed)),
		cdf: make([]float64, n),
	}

	var sum float64

	// compute normalization constant
	for i := 1; i <= n; i++ {
		sum += 1.0 / math.Pow(float64(i), s)
	}

	// build CDF
	var cumulative float64
	for i := 1; i <= n; i++ {
		p := (1.0 / math.Pow(float64(i), s)) / sum
		cumulative += p
		z.cdf[i-1] = cumulative
	}

	return z
}

// Next returns index [0..n)
func (z *ZipfGenerator) Next() int {
	r := z.rng.Float64()

	// binary search in CDF
	low, high := 0, len(z.cdf)-1

	for low < high {
		mid := (low + high) / 2
		if z.cdf[mid] < r {
			low = mid + 1
		} else {
			high = mid
		}
	}

	return low
}
