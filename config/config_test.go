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

func compareJsonAndConfig(t *testing.T, jsonFile string, confFile string) {
	json, err := getJsonMap(jsonFile)
	if err != nil {
		t.Error(err)
	}

	conf, err := Load(confFile)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(json, conf.options) {
		t.Errorf("\nwant %v\n got %v", json, conf.options)
	}
}

func TestLoadSingleFile(t *testing.T) {
	compareJsonAndConfig(t, "testdata/singlefile.json", "testdata/singlefile.conf")
}

func TestParseNormalJson(t *testing.T) {
	compareJsonAndConfig(t, "testdata/singlefile.json", "testdata/singlefile.json")
}

// func TestLoadMultiFile(t *testing.T) {
// 	compareJsonAndConfig(t, "testdata/singlefile.json", "testdata/multifile.conf")
// }
