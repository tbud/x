package config

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

var scriptConf = Config{
	"test1": Config{
		"num":     1,
		"comment": "#",
		"ok":      true,
		"cover": Config{
			"fnum": 12.58,
		},
	},
	"test2": Config{
		"mylist": []string{"1", "2", "3"},
	},
	"test3": Config{
		"nums": Config{
			"num1": -0.123,
			"num2": 3.14e+2,
			"num3": 3.14e-3,
			"num4": 3.14e4,
			"num5": 3.14e+10,
			"num6": 0e6,
		},
	},
}

func TestScriptConfigGetInt(t *testing.T) {
	// test get ok
	if get, ok := scriptConf.Int("test1.num"); !ok || get != 1 {
		t.Errorf("get test1.num int value, want 1 get %d, %v", get, ok)
	}

	// test get error
	if get, ok := scriptConf.Int("test1.num1"); ok || get != 0 {
		t.Error("get test1.num1 int value, not error")
	}
}

func TestScriptConfigGetIntDefault(t *testing.T) {
	if scriptConf.IntDefault("test1.num", 5) != 1 {
		t.Error("get int default error")
	}

	if scriptConf.IntDefault("test1.num1", 5) != 5 {
		t.Error("get int default error")
	}
}

func TestScriptConfigGetFloat(t *testing.T) {
	// test get ok
	if get, ok := scriptConf.Float("test1.cover.fnum"); !ok || get != 12.58 {
		t.Errorf("get test1.cover.fnum float value, want 12.58 get %d, %v", get, ok)
	}

	// test get error
	if get, ok := scriptConf.Float("test1.num1"); ok || get != 0 {
		t.Error("get test1.num1 float value, not error")
	}
}

func TestScriptConfigGetFloatDefault(t *testing.T) {
	if scriptConf.FloatDefault("test1.cover.fnum", 5.5) != 12.58 {
		t.Error("get int default error")
	}

	if scriptConf.FloatDefault("test1.num1", 5.5) != 5.5 {
		t.Error("get int default error")
	}
}

func TestScriptConfigGetString(t *testing.T) {
	// test get ok
	if get, ok := scriptConf.String("test1.comment"); !ok || get != "#" {
		t.Errorf("get test1.comment string value, want '#' get %s, %v", get, ok)
	}

	// test get error
	if get, ok := scriptConf.String("test1.comment1"); ok || get != "" {
		t.Error("get test1.comment1 string value, not error")
	}
}

func TestScriptConfigGetStringDefault(t *testing.T) {
	if scriptConf.StringDefault("test1.comment", "##") != "#" {
		t.Error("get string default error")
	}

	if scriptConf.StringDefault("test1.comment1", "##") != "##" {
		t.Error("get string default error")
	}
}

func TestScriptConfigGetStrings(t *testing.T) {
	// test get ok
	if get, ok := scriptConf.Strings("test2.mylist"); !ok || !reflect.DeepEqual(get, []string{"1", "2", "3"}) {
		t.Errorf("get test2.mylist strings value, want [1,2,3] get %s, %v", get, ok)
	}

	// test get error
	if get, ok := scriptConf.Strings("test2.mylist1"); ok || !reflect.DeepEqual(get, []string{}) {
		t.Error("get test2.mylist1 strings value, not error")
	}
}

func TestScriptConfigGetStringsDefault(t *testing.T) {
	if !reflect.DeepEqual(scriptConf.StringsDefault("test2.mylist", []string{"1", "2"}), []string{"1", "2", "3"}) {
		t.Error("get strings default error")
	}

	if !reflect.DeepEqual(scriptConf.StringsDefault("test2.mylist1", []string{"1", "2", "3"}), []string{"1", "2", "3"}) {
		t.Error("get strings default error")
	}
}

func TestScriptConfigGetBool(t *testing.T) {
	// test get ok
	if get, ok := scriptConf.Bool("test1.ok"); !ok || get != true {
		t.Errorf("get test1.ok bool value, want true get %v, %v", get, ok)
	}

	// test get error
	if get, ok := scriptConf.Bool("test1.ok1"); ok || get != false {
		t.Error("get test1.ok1 bool value, not error")
	}
}

func TestScriptConfigGetBoolDefault(t *testing.T) {
	if scriptConf.BoolDefault("test1.ok", false) != true {
		t.Error("get bool default error")
	}

	if scriptConf.BoolDefault("test1.ok1", false) != false {
		t.Error("get bool default error")
	}
}

func TestScriptConfigSubConfig(t *testing.T) {
	subConf := scriptConf.SubConfig("test1")
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

func TestScriptConfigSubConfigNotExist(t *testing.T) {
	subConf := scriptConf.SubConfig("test1.tttt")
	if subConf != nil {
		t.Error("get sub option test1.tttt error")
	}

	subConf = scriptConf.SubConfig("test1.num")
	if subConf != nil {
		t.Error("get sub option test1.tttt error")
	}
}

func TestScriptConfigEachAndKeyLen(t *testing.T) {
	subConf := scriptConf.SubConfig("test3.nums")

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
