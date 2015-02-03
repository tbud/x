package config

import (
	"github.com/tbud/x/encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
)

func getJsonMap(file string) (m interface{}, err error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	err = json.Unmarshal(buf, &m)
	return
}

func TestLoadSingleFile(t *testing.T) {
	json, err := getJsonMap("testdata/singlefile.json")
	if err != nil {
		t.Error(err)
	}

	conf, err := Load("testdata/singlefile.conf")
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(json, conf.options) {
		t.Errorf("\nwant %v\n got %v", json, conf.options)
	}
}
