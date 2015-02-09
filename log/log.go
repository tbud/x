package log

import (
	"errors"
	"fmt"
	"github.com/tbud/x/config"
	"github.com/tbud/x/log/appender"
	"strings"
	"time"
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

var logConfLevels = map[string]int{
	"emergency": LevelEmergency,
	"alert":     LevelAlert,
	"critical":  LevelCritical,
	"error":     LevelError,
	"warn":      LevelWarning,
	"notice":    LevelNotice,
	"info":      LevelInformational,
	"debug":     LevelDebug,
}

type Logger struct {
	fastMode      bool
	level         int
	rootAppenders []appender.Appender
	appenders     map[string]appender.Appender
}

func stringToLogLevel(name string) int {
	if ret, ok := logConfLevels[strings.ToLower(name)]; ok {
		return ret
	}
	return LevelError
}

func New(conf *config.Config) (*Logger, error) {
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
	if level >= LevelEmergency && level <= LevelDebug {
		l.level = level
	}
}

func (l *Logger) Emergency(format string, v ...interface{}) {
	if l.level >= LevelEmergency {
		l.output(LevelEmergency, format, v...)
	}
}

func (l *Logger) Alert(format string, v ...interface{}) {
	if l.level >= LevelAlert {
		l.output(LevelAlert, format, v...)
	}
}

func (l *Logger) Critical(format string, v ...interface{}) {
	if l.level >= LevelCritical {
		l.output(LevelCritical, format, v...)
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.level >= LevelError {
		l.output(LevelError, format, v...)
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level >= LevelWarning {
		l.output(LevelWarning, format, v...)
	}
}

func (l *Logger) Notice(format string, v ...interface{}) {
	if l.level >= LevelNotice {
		l.output(LevelNotice, format, v...)
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level >= LevelInformational {
		l.output(LevelInformational, format, v...)
	}
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level >= LevelDebug {
		l.output(LevelDebug, format, v...)
	}
}

func (l *Logger) initRoot(conf *config.Config) error {
	l.fastMode = conf.BoolDefault("fastmode", true)
	l.level = stringToLogLevel(conf.StringDefault("level", "error"))

	appenderRefs := conf.StringsDefault("appendrefs", []string{"console"})
	for i := range appenderRefs {
		if appender, ok := l.appenders[appenderRefs[i]]; ok {
			l.rootAppenders = append(l.rootAppenders, appender)
		} else {
			return errors.New("appender " + appenderRefs[i] + " not exist for root init.")
		}
	}
	return nil
}

func (l *Logger) loadAppenders(conf *config.Config) error {
	if conf == nil || conf.KeyLen() == 0 {
		appender, err := appender.New("Console", nil)
		if err != nil {
			panic("Load default appender error: " + err.Error())
		}

		l.appenders["console"] = appender
	} else {
		return conf.EachSubConfig(func(key string, subConf *config.Config) error {
			if appenderType, ok := subConf.String("type"); ok {
				appender, err := appender.New(appenderType, subConf)
				if err != nil {
					return err
				}

				l.appenders[key] = appender
			} else {
				return errors.New("appender without type can't init. Appender name: " + key)
			}
			return nil
		})
	}
	return nil
}

func (l *Logger) output(level int, format string, v ...interface{}) {
	if l.fastMode {
		for i := range l.rootAppenders {
			msg := appender.LogMsg{Date: time.Now()}
			msg.Msg = fmt.Sprintf(format, v...)
			l.rootAppenders[i].Append(&msg)
		}
	} else {
		// TODO detail mode
	}
}
