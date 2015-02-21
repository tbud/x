package config

import (
	"reflect"
	"testing"
)

var mergeConf = Config{
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

func TestMergeInt(t *testing.T) {
	if get, ok := mergeConf.Int("test1.num"); !ok || get != 1 {
		t.Errorf("get test1.num int value, want 1 get %d, %v", get, ok)
	}

	mergeConf.Merge("test1.num", 2)

	if get, ok := mergeConf.Int("test1.num"); !ok || get != 2 {
		t.Errorf("get test1.num int value, want 2 get %d, %v", get, ok)
	}
}

func TestMergeIntToString(t *testing.T) {
	if get, ok := mergeConf.Float("test3.nums.num1"); !ok || get != -0.123 {
		t.Errorf("get test3.nums.num1 float value, want -0.123 get %v, %v", get, ok)
	}

	mergeConf.Merge("test3.nums.num1", "two")

	if get, ok := mergeConf.String("test3.nums.num1"); !ok || get != "two" {
		t.Errorf("get test3.nums.num1 string value, want two get %v, %v", get, ok)
	}
}

func TestMergeBool(t *testing.T) {
	mergeConf.Merge("test4.abc.ok", true)

	if get, ok := mergeConf.Bool("test4.abc.ok"); !ok || get != true {
		t.Errorf("get test4.abc.ok bool value, want true get %v, %v", get, ok)
	}
}

func TestMergeConfig(t *testing.T) {
	if get, ok := mergeConf.Float("test1.cover.fnum"); !ok || get != 12.58 {
		t.Errorf("get test1.cover.fnum float value, want 12.58 get %v, %v", get, ok)
	}

	mergeConf.Merge("test1", Config{
		"num": 1,
		"ok":  false,
		"cover": map[string]interface{}{
			"fnum": 0.001,
		},
		"list": Config{
			"strlist": []string{"11", "22", "33"},
		},
	})

	if get, ok := mergeConf.Int("test1.num"); !ok || get != 1 {
		t.Errorf("get test1.num int value, want 1 get %v, %v", get, ok)
	}

	if get, ok := mergeConf.Float("test1.cover.fnum"); !ok || get != 0.001 {
		t.Errorf("get test1.cover.fnum float value, want 12.58 get %v, %v", get, ok)
	}

	if get, ok := mergeConf.Bool("test1.ok"); !ok || get != false {
		t.Errorf("get test1.ok bool value, want 12.58 get %v, %v", get, ok)
	}

	if get, ok := mergeConf.Strings("test1.list.strlist"); !ok || !reflect.DeepEqual(get, []string{"11", "22", "33"}) {
		t.Errorf("get test1.list.strlist string list value, want [11, 22, 33] get %v, %v", get, ok)
	}
}

func TestMergeRootConfig(t *testing.T) {
	mergeConf.Merge("", Config{
		"test2": Config{
			"mylist1": []string{"11", "12", "13"},
		},
	})

	if get, ok := mergeConf.Strings("test2.mylist1"); !ok || !reflect.DeepEqual(get, []string{"11", "12", "13"}) {
		t.Errorf("get test2.mylist1 string list value, want [11, 12, 13] get %v, %v", get, ok)
	}
}

func TestMergeCoverage(t *testing.T) {
	err := mergeConf.Merge("", true)
	if err != nil {
		t.Error(err)
	}
}
