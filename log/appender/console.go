package appender

import (
	"github.com/tbud/x/config"
	"github.com/tbud/x/log/common"
	"github.com/tbud/x/log/layout"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

type ConsoleAppender struct {
	sync.Mutex
	out      io.Writer
	layout   layout.Layout
	buf      []byte
	needFile bool
	needTime bool
}

func (c *ConsoleAppender) Append(m *common.LogMsg) (err error) {
	err = nil
	c.Lock()
	c.buf = c.buf[:0]
	err = c.layout.Format(&c.buf, m)
	switch m.Level {
	case common.LevelError, common.LevelFatal:
		c.out.Write([]byte("\x1B[31m"))
	case common.LevelWarn:
		c.out.Write([]byte("\x1B[35m"))
	case common.LevelInfo:
		c.out.Write([]byte("\x1B[32m"))
	}

	c.out.Write(c.buf)

	switch m.Level {
	default:
		c.out.Write([]byte("\x1B[39m"))
	case common.LevelDebug, common.LevelTrace:
		// ignore
	}

	c.Unlock()
	return
}

func (c *ConsoleAppender) NeedFile() bool {
	return c.needFile
}

func (c *ConsoleAppender) NeedTime() bool {
	return c.needTime
}

func consoleAppender(conf config.Config) (app Appender, err error) {
	appender := &ConsoleAppender{}
	switch strings.ToLower(conf.StringDefault("target", "stdout")) {
	default:
		appender.out = os.Stdout
	case "stderr":
		appender.out = os.Stderr
	case "discard":
		appender.out = ioutil.Discard
	}

	if appender.layout, err = layout.New(conf.SubConfig("layout")); err != nil {
		return nil, err
	}

	appender.needFile = appender.layout.NeedFile()
	appender.needTime = appender.layout.NeedTime()

	return appender, nil
}

func init() {
	Register("Console", consoleAppender)
}
