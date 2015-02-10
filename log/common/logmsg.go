package common

import (
	"time"
)

type LogMsg struct {
	File  string
	Line  int
	Level int
	Msg   string
	Date  time.Time
}
