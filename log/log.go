package log

import (
	"errors"
	"fmt"
	"github.com/tbud/x/config"
	"github.com/tbud/x/log/appender"
	. "github.com/tbud/x/log/common"
	"runtime"
	"sync"
	"time"
)

type Logger struct {
	fastMode      bool
	level         int
	rootAppenders []appender.Appender
	appenders     map[string]appender.Appender
	needFile      bool
	needTime      bool
}

func New(conf config.Config) (*Logger, error) {
	logger := Logger{appenders: map[string]appender.Appender{}}

	err := logger.loadAppenders(conf.SubConfig("appender"))
	if err != nil {
		return nil, err
	}

	err = logger.initRoot(conf.SubConfig("root"))
	if err != nil {
		return nil, err
	}

	return &logger, nil
}

func (l *Logger) SetLevel(level int) {
	if level >= LevelFatal && level <= LevelTrace {
		l.level = level
	}
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	if l.level >= LevelFatal {
		l.output(LevelFatal, format, v...)
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.level >= LevelError {
		l.output(LevelError, format, v...)
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level >= LevelWarn {
		l.output(LevelWarn, format, v...)
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level >= LevelInfo {
		l.output(LevelInfo, format, v...)
	}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level >= LevelDebug {
		l.output(LevelDebug, format, v...)
	}
}

func (l *Logger) Trace(format string, v ...interface{}) {
	if l.level >= LevelTrace {
		l.output(LevelTrace, format, v...)
	}
}

func (l *Logger) initRoot(conf config.Config) error {
	l.fastMode = conf.BoolDefault("fastmode", true)
	l.level = LogStringToLevel(conf.StringDefault("level", "info"))

	appenderRefs := conf.StringsDefault("appendrefs", []string{"console"})
	for _, appenderRef := range appenderRefs {
		if appender, ok := l.appenders[appenderRef]; ok {
			l.rootAppenders = append(l.rootAppenders, appender)
		} else {
			return errors.New("appender " + appenderRef + " not exist for root init.")
		}
	}

	for _, appender := range l.rootAppenders {
		if appender.NeedFile() {
			l.needFile = true
		}
		if appender.NeedTime() {
			l.needTime = true
		}
	}
	return nil
}

func (l *Logger) loadAppenders(conf config.Config) error {
	if conf == nil || conf.KeyLen() == 0 {
		appender, err := appender.New(nil)
		if err != nil {
			panic("Load default appender error: " + err.Error())
		}

		l.appenders["console"] = appender
	} else {
		return conf.EachSubConfig(func(key string, subConf config.Config) error {
			appender, err := appender.New(subConf)
			if err != nil {
				return errors.New("Load appender " + key + " error: " + err.Error())
			}

			l.appenders[key] = appender
			return nil
		})
	}
	return nil
}

func (l *Logger) output(level int, format string, v ...interface{}) {
	if l.fastMode {
		msg := LogMsg{Level: level, Msg: fmt.Sprintf(format, v...)}
		if l.needTime {
			msg.Date = time.Now()
		}
		if l.needFile {
			msg.File, msg.Line = pcFileLineMaps.getFileLine()
		}
		for i := range l.rootAppenders {
			l.rootAppenders[i].Append(&msg)
		}
	} else {
		// TODO detail mode
	}
}

var pcFileLineMaps = pcFileLineMap{m: map[uintptr]fileLine{}}

type fileLine struct {
	file string
	line int
}

type pcFileLineMap struct {
	sync.RWMutex
	m map[uintptr]fileLine
}

func (p *pcFileLineMap) getFileLine() (file string, line int) {
	var rpc [2]uintptr
	runtime.Callers(3, rpc[:])

	p.RLock()
	v, ok := p.m[rpc[1]]
	p.RUnlock()

	if ok {
		return v.file, v.line
	}

	var pc uintptr
	pc, file, line, ok = runtime.Caller(3)
	if !ok {
		file = "???"
		line = 0
	}

	fileLine := fileLine{file, line}
	p.Lock()
	p.m[pc] = fileLine
	p.Unlock()
	return
}
