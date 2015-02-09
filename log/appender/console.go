package appender

import (
	"github.com/tbud/x/config"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"
)

type ConsoleAppender struct {
	sync.Mutex
	out   io.Writer
	templ *template.Template
}

func (c *ConsoleAppender) Append(m *LogMsg) error {
	// return c.templ.Execute(c.out, m)
	msg := m.Date.Format("060102 15:04:05.000") + m.Msg
	c.Lock()
	c.out.Write([]byte(msg))
	c.Unlock()
	return nil
}

func fdate(layout string, date *time.Time) string {
	if date == nil {
		return ""
	}
	return date.Format(layout)
}

func consoleAppender(conf *config.Config) (app Appender, err error) {
	if conf == nil {
		conf = &config.Config{}
	}

	appender := &ConsoleAppender{}
	switch strings.ToLower(conf.StringDefault("target", "stdout")) {
	default:
		appender.out = os.Stdout
	case "stderr":
		appender.out = os.Stderr
	case "discard":
		appender.out = ioutil.Discard
	}

	funcMap := template.FuncMap{
		"fdate": fdate,
	}

	templ := conf.StringDefault("pattern", "{{.Date|fdate \"060102 15:04:05.000000\"}} - {{.Msg}}") + "\n"
	if appender.templ, err = template.New("Console").Funcs(funcMap).Parse(templ); err != nil {
		return nil, err
	}

	return appender, nil
}

func init() {
	Register("Console", consoleAppender)
}
