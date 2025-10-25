package spinner

import (
	"testing"
)

func BenchmarkSpin(b *testing.B) {
	for b.Loop() {
		spin(2)
	}
}
