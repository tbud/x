package config

import (
	"errors"
	"github.com/tbud/x/encoding/json"
	"io/ioutil"
	"reflect"
	"strings"
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

	if !reflect.DeepEqual(Config(json.(map[string]interface{})), conf) {
		t.Errorf("\nwant %v\n got %v", json, conf)
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

func TestLoadNotExistFile(t *testing.T) {
	_, err := Load("testdata/includenoexistfile.conf")
	if err != nil && !strings.HasPrefix(err.Error(), "error in load include:") {
		t.Error(err)
	}
}

func TestScannerError(t *testing.T) {
	_, err := Load("testdata/scanerr.conf")
	if err != nil && !strings.HasPrefix(err.Error(), "invalid character '\"'") {
		t.Error(err)
	}

	_, err = Load("testdata/scanerr1.conf")
	if err != nil && !strings.HasPrefix(err.Error(), "invalid character '+'") {
		t.Error(err)
	}
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

func TestConfigGetFloat(t *testing.T) {
	conf, err := Load("testdata/multifile.conf")
	if err != nil {
		t.Error(err)
	}

	// test get ok
	if get, ok := conf.Float("test1.cover.fnum"); !ok || get != 12.58 {
		t.Errorf("get test1.cover.fnum float value, want 12.58 get %d, %v", get, ok)
	}

	// test get error
	if get, ok := conf.Float("test1.num1"); ok || get != 0 {
		t.Error("get test1.num1 float value, not error")
	}
}

func TestConfigGetFloatDefault(t *testing.T) {
	conf, err := Load("testdata/multifile.conf")
	if err != nil {
		t.Error(err)
	}

	if conf.FloatDefault("test1.cover.fnum", 5.5) != 12.58 {
		t.Error("get int default error")
	}

	if conf.FloatDefault("test1.num1", 5.5) != 5.5 {
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
		t.Errorf("get test1.comment string value, want '#' get %s, %v", get, ok)
	}

	// test get error
	if get, ok := conf.String("test1.comment1"); ok || get != "" {
		t.Error("get test1.comment1 string value, not error")
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

func TestConfigGetStrings(t *testing.T) {
	conf, err := Load("testdata/multifile.conf")
	if err != nil {
		t.Error(err)
	}

	// test get ok
	if get, ok := conf.Strings("test2.mylist"); !ok || !reflect.DeepEqual(get, []string{"1", "2", "3"}) {
		t.Errorf("get test2.mylist strings value, want [1,2,3] get %s, %v", get, ok)
	}

	// test get error
	if get, ok := conf.Strings("test2.mylist1"); ok || !reflect.DeepEqual(get, []string{}) {
		t.Error("get test2.mylist1 strings value, not error")
	}
}

func TestConfigGetStringsDefault(t *testing.T) {
	conf, err := Load("testdata/multifile.conf")
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(conf.StringsDefault("test2.mylist", []string{"1", "2"}), []string{"1", "2", "3"}) {
		t.Error("get strings default error")
	}

	if !reflect.DeepEqual(conf.StringsDefault("test2.mylist1", []string{"1", "2", "3"}), []string{"1", "2", "3"}) {
		t.Error("get strings default error")
	}
}

func TestConfigGetBool(t *testing.T) {
	conf, err := Load("testdata/multifile.conf")
	if err != nil {
		t.Error(err)
	}

	// test get ok
	if get, ok := conf.Bool("test1.ok"); !ok || get != true {
		t.Errorf("get test1.ok bool value, want true get %v, %v", get, ok)
	}

	// test get error
	if get, ok := conf.Bool("test1.ok1"); ok || get != false {
		t.Error("get test1.ok1 bool value, not error")
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

func TestConfigSubConfig(t *testing.T) {
	conf, err := Load("testdata/multifile.conf")
	if err != nil {
		t.Error(err)
	}

	subConf := conf.SubConfig("test1")
	if subConf == nil {
		t.Error("get sub option test1 error")
	}

	if subConf.BoolDefault("ok", false) != true {
		t.Error("get bool default error")
	}

	if subConf.BoolDefault(".ok1", false) != false {
		t.Error("get bool default error")
	}
}

func TestConfigSubConfigNotExist(t *testing.T) {
	conf, err := Load("testdata/multifile.conf")
	if err != nil {
		t.Error(err)
	}

	subConf := conf.SubConfig("test1.tttt")
	if subConf != nil {
		t.Error("get sub option test1.tttt error")
	}

	subConf = conf.SubConfig("test1.num")
	if subConf != nil {
		t.Error("get sub option test1.tttt error")
	}
}

func TestConfigEachAndKeyLen(t *testing.T) {
	conf, err := Load("testdata/multifile.conf")
	if err != nil {
		t.Error(err)
	}

	subConf := conf.SubConfig("test3.nums")

	if subConf.KeyLen() > 0 {
		err := subConf.EachSubConfig(func(key string, conf Config) error {
			if !strings.HasPrefix(key, "num") {
				return errors.New("unaccept key in each config test: " + key)
			}
			return nil
		})

		if err != nil {
			t.Error(err)
		}
	} else {
		t.Error("load file error")
	}
}
