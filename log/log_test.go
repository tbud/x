package log

import (
	"github.com/tbud/x/config"
	"runtime"
	"testing"
)

func BenchmarkRuntimeCallerTest(b *testing.B) {
	var pcs [2]uintptr
	// var pc uintptr
	for i := 0; i < b.N; i++ {
		runtime.Callers(0, pcs[:])
		// pc = pcs[1]
	}

	// b.Log(pc)
}

func BenchmarkFastmodeDebug(b *testing.B) {
	conf, err := config.Load("log.conf")
	if err != nil {
		b.Error(err)
	}

	logger, err := New(&conf)
	if err != nil {
		b.Error(err)
		return
	}

	for i := 0; i < b.N; i++ {
		logger.Debug("py test benchmark")
	}
}

func TestDebug(t *testing.T) {
	conf, err := config.Load("log.conf")
	if err != nil {
		t.Error(err)
	}

	logger, err1 := New(&conf)
	if err1 != nil {
		t.Error(err1)
		return
	}

	logger.Debug("py test console")
}
