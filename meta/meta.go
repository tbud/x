package meta

import (
	"fmt"
	"reflect"
	"runtime"
)

type MetaInfo struct {
	Name          string
	Min           int
	Max           int
	Skip          bool
	OmitEmpty     bool
	WithQuote     bool
	MatchRegExp   string
	SerializeType string
}

const (
	metaTag     = "@"
	jsonTag     = "json"
	ormTag      = "orm"
	validateTag = "validate"
)

const (
	minFlag   = "("
	minEqual  = "["
	sizeSplit = ".."
	maxFlag   = ")"
	maxEqual  = "]"

	nonSkipFlag = "+"

	omitEmpty    = "~"
	nonOmitEmpty = "="
)

const (
	skipFlag = iota
	omitEmptyFlag
)

var MetaSymbols = map[string]int{
	"-":         skipFlag,
	"~":         omitEmptyFlag,
	"omitempty": omitEmptyFlag,
}

func Meta(v reflect.Value, tagName string) (m []MetaInfo, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	m = meta(v, tagName)
	return
}

func JsonMeta(v reflect.Value) ([]MetaInfo, error) {
	return Meta(v, ormTag)
}

func OrmMeta(v reflect.Value) ([]MetaInfo, error) {
	return Meta(v, ormTag)
}

func ValidateMeta(v reflect.Value) ([]MetaInfo, error) {
	return Meta(v, validateTag)
}

func meta(v reflect.Value, tagName string) []MetaInfo {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		fmt.Println(t.Field(i).Tag.Get(tagName))
	}

	return []MetaInfo{}
}
