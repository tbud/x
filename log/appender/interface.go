package appender

import (
	"errors"
	"github.com/tbud/x/config"
	"time"
)

type LogMsg struct {
	File string
	Line int
	Msg  string
	Date time.Time
}

type Appender interface {
	Append(m *LogMsg) error
}

type AppenderMaker func(conf *config.Config) (Appender, error)

var appenderMakers = make(map[string]AppenderMaker)

// Register makes a log appender maker available by the appender name.
// If Register is called twice with the same name or if appender maker is nil,
// it panics.
func Register(name string, appenderMaker AppenderMaker) {
	if appenderMaker == nil {
		panic("log: Register appender maker is nil")
	}
	if _, dup := appenderMakers[name]; dup {
		panic("log: Register called twice for appender maker " + name)
	}
	appenderMakers[name] = appenderMaker
}

func MakeAppender(name string, conf *config.Config) (appender Appender, err error) {
	if appenderMaker, ok := appenderMakers[name]; ok {
		appender, err = appenderMaker(conf)
		return
	}

	return nil, errors.New("Appender maker not exist.")
}
