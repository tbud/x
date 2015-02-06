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

func TestLoadMultiFile(t *testing.T) {
	compareJsonAndConfig(t, "testdata/singlefile.json", "testdata/multifile.conf")
}

func TestConfigGetInt(t *testing.T) {
	conf, err := Load("testdata/multifile.conf")
	if err != nil {
		t.Error(err)
	}

	// test get ok
	if get, ok := conf.Int("test1.num"); !ok || get != 1 {
		t.Errorf("get test1.num int value, want 1 get %d, %v", get, ok)
	}

	// test get error
	if get, ok := conf.Int("test1.num1"); ok || get != 0 {
		t.Error("get test1.num1 int value, not error")
	}
}

func TestConfigGetIntDefault(t *testing.T) {
	conf, err := Load("testdata/multifile.conf")
	if err != nil {
		t.Error(err)
	}

	if conf.IntDefault("test1.num", 5) != 1 {
		t.Error("get int default error")
	}

	if conf.IntDefault("test1.num1", 5) != 5 {
		t.Error("get int default error")
	}
}

func TestConfigGetString(t *testing.T) {
	conf, err := Load("testdata/multifile.conf")
	if err != nil {
		t.Error(err)
	}

	// test get ok
	if get, ok := conf.String("test1.comment"); !ok || get != "#" {
		t.Errorf("get test1.num string value, want 1 get %s, %v", get, ok)
	}

	// test get error
	if get, ok := conf.String("test1.comment1"); ok || get != "" {
		t.Error("get test1.num1 string value, not error")
	}
}

func TestConfigGetStringDefault(t *testing.T) {
	conf, err := Load("testdata/multifile.conf")
	if err != nil {
		t.Error(err)
	}

	if conf.StringDefault("test1.comment", "##") != "#" {
		t.Error("get string default error")
	}

	if conf.StringDefault("test1.comment1", "##") != "##" {
		t.Error("get string default error")
	}
}

func TestConfigGetBool(t *testing.T) {
	conf, err := Load("testdata/multifile.conf")
	if err != nil {
		t.Error(err)
	}

	// test get ok
	if get, ok := conf.Bool("test1.ok"); !ok || get != true {
		t.Errorf("get test1.num bool value, want 1 get %v, %v", get, ok)
	}

	// test get error
	if get, ok := conf.Bool("test1.ok1"); ok || get != false {
		t.Error("get test1.num1 bool value, not error")
	}
}

func TestConfigGetBoolDefault(t *testing.T) {
	conf, err := Load("testdata/multifile.conf")
	if err != nil {
		t.Error(err)
	}

	if conf.BoolDefault("test1.ok", false) != true {
		t.Error("get bool default error")
	}

	if conf.BoolDefault("test1.ok1", false) != false {
		t.Error("get bool default error")
	}
}
