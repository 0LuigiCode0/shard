package idmpt

import (
	"math/rand"
	"time"
)

func calculateTTL(ttl time.Duration) time.Duration {
	return time.Duration(float64(ttl) * (0.99 + (rand.Float64() / 50)))
}
