package config

import (
	"path/filepath"
	"runtime"
)

type Config struct {
	options map[string]interface{}
}

// type ConfigError struct {
// 	file   string
// 	offset int
// 	err    error
// }

// func (e *ConfigError) Error() string {
// 	return "config file " + e.err.Error()
// }

func Load(fileName string) (config Config, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	config = Config{}
	scan := fileScanner{}

	if !filepath.IsAbs(fileName) {
		fileName, err = filepath.Abs(fileName)
		if err != nil {
			return
		}
	}

	err = scan.checkValid(fileName)
	if err != nil {
		return
	}

	// file := filepath.Base(fileName)

	// config.include(config.top, file)
	return
}

func (c *Config) Int(key string) (result int, found bool) {
	return 0, false
}

func (c *Config) IntDefault(key string, defaultValue int) int {
	return defaultValue
}

func (c *Config) String(key string) (result string, found bool) {
	return "", false
}

func (c *Config) StringDefault(key, defaultValue string) string {
	return defaultValue
}

func (c *Config) Bool(key string) (result, found bool) {
	return false, false
}

func (c *Config) BoolDefault(key string, defaultValue bool) bool {
	return defaultValue
}

func (c *Config) SubOptions(key string) *Config {
	return nil
}

// func (c *Config) include(p map[string]interface{}, fileName string) {
// 	includeFile := filepath.Join(c.fileDir, fileName)
// 	buf, err := ioutil.ReadFile(includeFile)
// 	if err != nil {
// 		panic(&ConfigError{err})
// 	}

// 	scan := scanner{data: buf}
// }
