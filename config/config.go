package config

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
)

type Config struct {
	fileDir string // load config file path
	top     map[string]interface{}
}

type ConfigError struct {
	file   string
	offset int
	err    error
}

func (e *ConfigError) Error() string {
	return "config error: " + e.err.Error()
}

func Load(fileName string) (config Config, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	config = Config{fileDir: filepath.Dir(fileName)}
	config.top = map[string]interface{}{}

	file := filepath.Base(fileName)

	config.include(config.top, file)
	return
}

func (c *Config) include(p map[string]interface{}, fileName string) {
	includeFile := filepath.Join(c.fileDir, fileName)
	buf, err := ioutil.ReadFile(includeFile)
	if err != nil {
		panic(&ConfigError{err})
	}

	scan := scanner{data: buf}
}

type scanner struct {
	data   []byte
	offset int
	step   func(*scanner, int) int
	value  []byte
}

func (s *scanner) next() {

}
