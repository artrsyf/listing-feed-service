package distribution

import (
	"math"
	"math/rand"
)

// ParetoGenerator produces 80/20-like distribution
type ParetoGenerator struct {
	alpha float64
	rng   *rand.Rand
}

func NewPareto(alpha float64, seed int64) *ParetoGenerator {
	return &ParetoGenerator{
		alpha: alpha,
		rng:   rand.New(rand.NewSource(seed)),
	}
}

// Next returns skewed value in range [0..1]
func (p *ParetoGenerator) Next() float64 {
	u := p.rng.Float64()
	return math.Pow(1.0-u, -1.0/p.alpha)
}
