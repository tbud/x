package log

import (
	"runtime"
	"testing"
)

func BenchmarkRuntimeCallerTest(b *testing.B) {
	var pcs [2]uintptr
	var pc uintptr
	for i := 0; i < b.N; i++ {
		runtime.Callers(0, pcs[:])
		pc = pcs[1]
	}

	b.Log(pc)
}

func TestA(t *testing.T) {
	t.Log("TestA")
}
