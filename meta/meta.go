package meta

import (
	"reflect"
	"runtime"
	"strings"
	"sync"
	"unicode"
)

type MetaInfo struct {
	Name          string
	Min           int
	Max           int
	Skip          bool
	OmitEmpty     bool
	Quote         bool
	MatchRegExp   string
	SerializeType string
}

const (
	metaTag     = "@"
	jsonTag     = "json"
	ormTag      = "orm"
	validateTag = "validate"
)

func JsonMeta(t reflect.Type) ([]MetaInfo, error) {
	return meta(t, jsonTag)
}

func OrmMeta(t reflect.Type) ([]MetaInfo, error) {
	return meta(t, ormTag)
}

func ValidateMeta(t reflect.Type) ([]MetaInfo, error) {
	return meta(t, validateTag)
}

type metaCache struct {
	sync.RWMutex
	m map[reflect.Type][]MetaInfo
}

func (m metaCache) getOrElse(key reflect.Type, f func() []MetaInfo) []MetaInfo {
	m.RLock()
	v, ok := m.m[key]
	m.RUnlock()
	if ok {
		return v
	}

	if f == nil {
		return nil
	}

	v = f()
	if v == nil {
		v = []MetaInfo{}
	}

	m.Lock()
	if m.m == nil {
		m.m = map[reflect.Type][]MetaInfo{}
	}
	m.m[key] = v
	m.Unlock()

	return v
}

var metaCaches = map[string]metaCache{
	metaTag:     metaCache{},
	jsonTag:     metaCache{},
	ormTag:      metaCache{},
	validateTag: metaCache{},
}

func originName(name string) string {
	return name
}

func firstLittleName(name string) string {
	n := []rune(name)
	n[0] = unicode.ToLower(n[0])
	return string(n)
}

func underscoreLittleName(name string) string {
	r := []rune{}
	for _, s := range name {
		if unicode.IsUpper(s) {
			if len(r) > 0 {
				r = append(r, []rune("_")[0])
			}

			r = append(r, unicode.ToLower(s))
		} else {
			r = append(r, s)
		}
	}

	return string(r)
}

var nameStrategy = map[string]func(string) string{
	jsonTag:     firstLittleName,
	ormTag:      underscoreLittleName,
	validateTag: originName,
}

func meta(t reflect.Type, tagName string) (retMeta []MetaInfo, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	retMeta = metaCaches[tagName].getOrElse(t, func() []MetaInfo {
		ms := metaCaches[metaTag].getOrElse(t, func() []MetaInfo {
			metaInfos := make([]MetaInfo, metaTp.NumField())
			metaFromTag(t, metaTag, metaInfos)
			return metaInfos
		})

		metaFromTag(t, tagName, ms)

		for i := 0; i < t.NumField(); i++ {
			if len(ms[i].Name) == 0 {
				ms[i].Name = nameStrategy[tagName](t.Field(i).Name)
			}
		}

		return ms
	})

	return
}

func metaFromTag(t reflect.Type, tagName string, metaInfos []MetaInfo) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(tagName)
		meta := &metaInfos[i]

		if len(tag) > 0 {
			for _, v := range strings.Split(tag, ",") {
				switch {
				case v == "-":
					meta.Skip = true
				case v == "~" || v == "omitempty":
					meta.OmitEmpty = true
				case v == "string" || v == "%q":
					meta.Quote = true
				case unicode.IsLetter([]rune(v)[0]):
					meta.Name = v
				}
			}
		}
	}
}
