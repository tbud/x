package set

import (
	"testing"
)

func BenchmarkIntSetAdd(b *testing.B) {
	s := NewIntSet()
	for i := 0; i < b.N; i++ {
		s.Add(i)
	}
}
