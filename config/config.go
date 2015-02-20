package config

import (
	"errors"
	"path/filepath"
	"runtime"
	"strings"
)

type Config map[string]interface{}

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

	err = scan.setOptions(config)
	return
}

func (c Config) Int(key string) (result int, found bool) {
	result, found = 0, false
	value := c.getValue(key)
	if value == nil {
		return
	}

	if retFloat, ok := value.(float64); ok {
		return int(retFloat), ok
	}
	result, found = value.(int)
	return
}

func (c Config) IntDefault(key string, defaultValue int) int {
	result, found := c.Int(key)
	if !found {
		result = defaultValue
	}
	return result
}

func (c Config) Float(key string) (result float64, found bool) {
	result, found = 0.0, false
	value := c.getValue(key)
	if value == nil {
		return
	}

	result, found = value.(float64)
	return
}

func (c Config) FloatDefault(key string, defaultValue float64) float64 {
	result, found := c.Float(key)
	if !found {
		result = defaultValue
	}
	return result
}

func (c Config) String(key string) (result string, found bool) {
	result, found = "", false
	value := c.getValue(key)
	if value == nil {
		return
	}

	result, found = value.(string)
	return
}

func (c Config) StringDefault(key, defaultValue string) string {
	result, found := c.String(key)
	if !found {
		result = defaultValue
	}
	return result
}

func (c Config) Bool(key string) (result, found bool) {
	result, found = false, false
	value := c.getValue(key)
	if value == nil {
		return
	}

	result, found = value.(bool)
	return
}

func (c Config) BoolDefault(key string, defaultValue bool) bool {
	result, found := c.Bool(key)
	if !found {
		result = defaultValue
	}
	return result
}

func (c Config) Strings(key string) (result []string, found bool) {
	result, found = []string{}, false
	value := c.getValue(key)
	if value == nil {
		return
	}

	if infs, f := value.([]interface{}); f {
		for _, inf := range infs {
			if v, ok := inf.(string); ok {
				found = true
				result = append(result, v)
			}
		}
		return
	}

	result, found = value.([]string)
	return
}

func (c Config) StringsDefault(key string, defaultValue []string) []string {
	result, found := c.Strings(key)
	if !found {
		result = defaultValue
	}
	return result
}

func subConfig(value interface{}) Config {
	if value != nil {
		if v, ok := value.(map[string]interface{}); ok {
			return Config(v)
		}
		if v, ok := value.(Config); ok {
			return v
		}
	}
	return nil
}

func (c Config) SubConfig(key string) Config {
	result := c.getValue(key)
	return subConfig(result)
}

func (c Config) EachSubConfig(fun func(key string, conf Config) error) error {
	if c == nil {
		return errors.New("Config is nil")
	}
	for key, value := range c {
		err := fun(key, subConfig(value))
		if err != nil {
			return err
		}
	}
	return nil
}

func (c Config) EachKey(fun func(key string) error) error {
	if c == nil {
		return errors.New("Config is nil")
	}
	for key, _ := range c {
		err := fun(key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c Config) KeyLen() int {
	if c == nil {
		return 0
	}
	return len(c)
}

func (c Config) Merge(key string, value interface{}) error {
	if len(key) == 0 || c == nil {
		return nil
	}
	ops := c

	keys := strings.Split(key, ".")
	lastKeyIndex := len(keys) - 1

	for i, k := range keys {
		if i == lastKeyIndex {
			switch val := value.(type) {
			case int, string, bool, float64, []string:
				ops[k] = val
			case map[string]interface{}, Config:
				tmpOps := ops.SubConfig(k)
				if tmpOps == nil {
					ops[k] = Config{}
					ops, _ = ops[k].(Config)
				} else {
					ops = tmpOps
				}

				var conf Config
				var ok bool
				if conf, ok = value.(Config); !ok {
					conf, _ = value.(map[string]interface{})
				}

				return conf.EachKey(func(key string) error {
					return ops.Merge(key, conf.getValue(key))
				})
			}
		} else {
			tmpOps := ops.SubConfig(k)
			if tmpOps == nil {
				ops[k] = Config{}
				ops, _ = ops[k].(Config)
			} else {
				ops = tmpOps
			}
		}
	}
	return nil
}

func (c Config) getValue(key string) interface{} {
	if len(key) == 0 || c == nil {
		return nil
	}
	ops := c

	keys := strings.Split(key, ".")
	lastkeyIndex := len(keys) - 1
	for i, key := range keys {
		if value, ok := ops[key]; ok {
			if i == lastkeyIndex {
				return value
			} else {
				if ops, ok = value.(map[string]interface{}); ok {
					continue
				}
				if ops, ok = value.(Config); ok {
					continue
				}
				return nil
			}
		} else {
			return nil
		}
	}
	return nil
}
