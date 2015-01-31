package config

import (
	"testing"
)

func TestLoadSingleFile(t *testing.T) {
	conf, err := Load("testdata/singlefile.conf")
	if err != nil {
		t.Logf("%v", conf)
	}
}
