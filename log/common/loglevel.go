package common

import (
	"strings"
)

const (
	LevelFatal = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
	LevelTrace
)

var logStringToLevels = map[string]int{
	"fatal": LevelFatal,
	"error": LevelError,
	"warn":  LevelWarn,
	"info":  LevelInfo,
	"debug": LevelDebug,
	"trace": LevelTrace,
}

var logLevelToStrings = map[int]string{}

var logLevelToShortStrings = map[int]string{}

func LogStringToLevel(name string) int {
	if ret, ok := logStringToLevels[strings.ToLower(name)]; ok {
		return ret
	}
	return LevelError
}

func LogLevelToString(level int) string {
	if ret, ok := logLevelToStrings[level]; ok {
		return ret
	}
	return "unknown"
}

func LogLevelToShortString(level int) string {
	if ret, ok := logLevelToShortStrings[level]; ok {
		return ret
	}
	return "U"
}

func init() {
	for key, value := range logStringToLevels {
		if len(key) == 4 {
			logLevelToStrings[value] = strings.ToUpper(key) + " "
		} else {
			logLevelToStrings[value] = strings.ToUpper(key)
		}

		logLevelToShortStrings[value] = strings.ToUpper(string(key[0]))
	}
}
