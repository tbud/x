package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

// RFC5424 log message levels.
const (
	LevelEmergency = iota
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInformational
	LevelDebug
)

type BudLogger struct {
	*log.Logger
}

func New(out io.Writer, prefix string) *BudLogger {
	budLog := &BudLogger{}
	budLog.Logger = log.New(out, prefix, log.Ldate|log.Ltime|log.Lshortfile)
	return budLog
}

func (l *BudLogger) EFatal(err error, v ...interface{}) {
	if err != nil {
		l.Output(2, fmt.Sprint(err)+fmt.Sprint(v...))
		os.Exit(1)
	}
}

func (l *BudLogger) EFatalf(err error, format string, v ...interface{}) {
	if err != nil {
		l.Output(2, fmt.Sprint(err)+fmt.Sprintf(format, v...))
		os.Exit(1)
	}
}

func (l *BudLogger) EFatalln(err error, v ...interface{}) {
	if err != nil {
		l.Output(2, fmt.Sprint(err)+fmt.Sprintln(v...))
		os.Exit(1)
	}
}

func (l *BudLogger) EPanic(err error, v ...interface{}) {
	if err != nil {
		s := fmt.Sprint(err) + fmt.Sprint(v...)
		l.Output(2, s)
		panic(s)
	}
}

func (l *BudLogger) EPanicf(err error, format string, v ...interface{}) {
	if err != nil {
		s := fmt.Sprint(err) + fmt.Sprintf(format, v...)
		l.Output(2, s)
		panic(s)
	}
}

func (l *BudLogger) EPanicln(err error, v ...interface{}) {
	if err != nil {
		s := fmt.Sprint(err) + fmt.Sprintln(v...)
		l.Output(2, s)
		panic(s)
	}
}
