package sync

import (
	"sync"
)

type Map struct {
	sync.RWMutex
	m map[interface{}]interface{}
}

func (m *Map) Get(key interface{}) (v interface{}) {
	m.RLock()
	v = m.m[key]
	m.RUnlock()
	return
}

func (m *Map) GetOrElse(key interface{}, f func() (interface{}, error), defaultValue interface{}) interface{} {
	m.RLock()
	v, ok := m.m[key]
	m.RUnlock()
	if ok {
		return v
	}

	if f == nil {
		return nil
	}

	r, err := f()
	if err != nil {
		r = defaultValue
	}

	m.Lock()
	if m.m == nil {
		m.m = map[interface{}]interface{}{}
	}
	m.m[key] = r
	m.Unlock()

	return r
}

func (m *Map) Set(key, value interface{}) bool {
	m.RLock()
	v, ok := m.m[key]
	m.RUnlock()

	if !ok || value != v {
		m.Lock()
		if m.m == nil {
			m.m = map[interface{}]interface{}{}
		}
		m.m[key] = value
		m.Unlock()
	} else {
		return false
	}

	return true
}

func (m *Map) Delete(key interface{}) {
	m.Lock()
	delete(m.m, key)
	m.Unlock()
}

func (m *Map) Has(key interface{}) (ok bool) {
	m.RLock()
	_, ok = m.m[key]
	m.RUnlock()
	return
}
