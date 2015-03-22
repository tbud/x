package config

import (
	"errors"
	"io"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"unicode"
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

func Read(reader io.Reader) (config Config, err error) {
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

	err = scan.checkReaderValid(reader)
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

func (c Config) Bools(key string) (result []bool, found bool) {
	result, found = []bool{}, false
	value := c.getValue(key)
	if value == nil {
		return
	}

	if infs, f := value.([]interface{}); f {
		for _, inf := range infs {
			if v, ok := inf.(bool); ok {
				found = true
				result = append(result, v)
			}
		}
		return
	}

	result, found = value.([]bool)
	return
}

func (c Config) BoolsDefault(key string, defaultValue []bool) []bool {
	result, found := c.Bools(key)
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
	if c == nil {
		return errors.New("Config is nil")
	}
	if len(key) == 0 {
		return c.mergeConf(value)
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
				ops = ops.subAndCreateConfig(k)

				return ops.mergeConf(value)
			}
		} else {
			ops = ops.subAndCreateConfig(k)
		}
	}
	return nil
}

func (c Config) SetStruct(v interface{}) error {
	if c == nil || v == nil {
		return nil
	}

	ev := reflect.ValueOf(v)
	if ev.Kind() == reflect.Ptr {
		ev = ev.Elem()
	} else {
		return errors.New("Struct must be a point.")
	}

	return c.EachKey(func(key string) error {
		value := ev.FieldByName(firstRuneToUpper(key))
		if value.IsValid() {
			switch value.Kind() {
			case reflect.Slice:
				if value.Type() == reflect.TypeOf([]string{}) {
					if ssv, ok := c.Strings(key); ok {
						value.Set(reflect.ValueOf(ssv))
					}
				}
			case reflect.Bool:
				if bv, ok := c.Bool(key); ok {
					value.SetBool(bv)
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if iv, ok := c.Int(key); ok {
					value.SetInt(int64(iv))
				}
			case reflect.String:
				if sv, ok := c.String(key); ok {
					value.SetString(sv)
				}
			case reflect.Float32, reflect.Float64:
				if fv, ok := c.Float(key); ok {
					value.SetFloat(fv)
				}
			}
		}
		return nil
	})
}

func firstRuneToUpper(key string) string {
	rkey := []rune(key)
	rkey[0] = unicode.ToUpper(rkey[0])
	return string(rkey)
}

func (c Config) subAndCreateConfig(key string) (ops Config) {
	tmpOps := c.SubConfig(key)
	if tmpOps == nil {
		c[key] = Config{}
		ops, _ = c[key].(Config)
	} else {
		ops = tmpOps
	}
	return
}

func (c Config) mergeConf(src interface{}) error {
	var conf Config
	var ok bool
	if conf, ok = src.(Config); !ok {
		if conf, ok = src.(map[string]interface{}); !ok {
			return nil
		}
	}

	return conf.EachKey(func(key string) error {
		return c.Merge(key, conf.getValue(key))
	})
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
