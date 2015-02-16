package appender

import (
	"errors"
	"github.com/tbud/x/config"
	"github.com/tbud/x/log/common"
)

type Appender interface {
	Append(m *common.LogMsg) error
	NeedFile() bool
	NeedTime() bool
}

type AppenderMaker func(conf config.Config) (Appender, error)

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

func New(conf config.Config) (appender Appender, err error) {
	name := conf.StringDefault("type", "Console")
	if appenderMaker, ok := appenderMakers[name]; ok {
		appender, err = appenderMaker(conf)
		return
	}

	return nil, errors.New("Appender maker not exist.")
}
