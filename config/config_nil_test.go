package config

import (
	"fmt"
	"strings"
	"testing"
)

func TestNil(t *testing.T) {
	var conf Config = nil

	err := conf.EachSubConfig(func(key string, c Config) error {
		return nil
	})
	if err == nil || !strings.Contains(fmt.Sprintf("%v", err), "Config is nil") {
		t.Error(err)
		return
	}

	err = conf.EachKey(func(key string) error {
		return nil
	})
	if err == nil || !strings.Contains(fmt.Sprintf("%v", err), "Config is nil") {
		t.Error(err)
		return
	}

	err = conf.Merge("str", "new string")
	if err == nil || !strings.Contains(fmt.Sprintf("%v", err), "Config is nil") {
		t.Error(err)
		return
	}

	if conf.KeyLen() != 0 {
		t.Error("nil config len must be 0")
	}

	if got := conf.IntDefault("intkey", 1980); got != 1980 {
		t.Errorf("want 1980 , get %d", got)
	}
}
