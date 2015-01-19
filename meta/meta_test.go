package meta

import (
	"reflect"
	"testing"
)

type MetaInfoTest struct {
	MetaTag  int    `@:"metaSymbol,123,441"`
	JsonTag  string `@:"fdsa" json:"tt"`
	EmptyTag string
}

var metaTp = reflect.TypeOf(MetaInfoTest{})

func TestJsonMeta(t *testing.T) {
	mi, ok := JsonMeta(metaTp)
	if ok != nil {
		t.Errorf("test json meta get a error: %v", ok)
	}

	want := []MetaInfo{
		MetaInfo{Name: "metaSymbol", Tagged: true},
		MetaInfo{Name: "tt", Tagged: true},
		MetaInfo{Name: "emptyTag"},
	}

	for i := 0; i < len(mi); i++ {
		if r, w := mi[i], want[i]; r != w {
			t.Errorf("want %v, get %v", w, r)
		}
	}
}

func TestOrmMeta(t *testing.T) {
	mi, ok := OrmMeta(metaTp)
	if ok != nil {
		t.Errorf("test json meta get a error: %v", ok)
	}

	want := []MetaInfo{
		MetaInfo{Name: "metaSymbol", Tagged: true},
		MetaInfo{Name: "fdsa", Tagged: true},
		MetaInfo{Name: "empty_tag"},
	}

	for i := 0; i < len(mi); i++ {
		if r, w := mi[i], want[i]; r != w {
			t.Errorf("want %v, get %v", w, r)
		}
	}
}
