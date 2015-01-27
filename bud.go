package bud

import (
	"github.com/tbud/bud/log"
	"os"
)

const (
	BUD_SEED_PATH = "github.com/tbud/seed"
)

var (
	ErrLog = log.New(os.Stderr, "[E]")
)
