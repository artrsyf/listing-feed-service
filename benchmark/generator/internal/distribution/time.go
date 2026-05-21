package distribution

import (
	"math"
	"math/rand"
	"time"
)

// TimeGenerator produces skewed timestamps
type TimeGenerator struct {
	start time.Time
	end   time.Time
	skew  float64
	rng   *rand.Rand
}

func NewTimeGenerator(start, end time.Time, skew float64, seed int64) *TimeGenerator {
	return &TimeGenerator{
		start: start,
		end:   end,
		skew:  skew,
		rng:   rand.New(rand.NewSource(seed)),
	}
}

// Next returns skewed timestamp (recent-heavy if skew > 1)
func (t *TimeGenerator) Next() time.Time {
	delta := t.end.Sub(t.start).Seconds()

	r := math.Pow(t.rng.Float64(), t.skew)

	sec := r * delta

	return t.start.Add(time.Duration(sec) * time.Second)
}
