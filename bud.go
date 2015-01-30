package x

import (
	"github.com/tbud/x/log"
	"os"
)

var (
	ErrLog = log.New(os.Stderr, "[E]")
)
