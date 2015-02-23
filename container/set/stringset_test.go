package set

import (
	"strconv"
	"testing"
)

func BenchmarkStringSetAdd(b *testing.B) {
	s := NewStringSet()
	for i := 0; i < b.N; i++ {
		s.Add(strconv.Itoa(i))
	}
}
