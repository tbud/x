package sync

import (
	"sync"
	"testing"
)

const k = "ttt"

func TestMapBehavior(t *testing.T) {
	m := Map{}

	if r := m.m["ttt"]; r != nil {
		t.Errorf("want nil but %v", r)
	}

	if r, ok := m.m["ttt"]; r != nil || ok {
		t.Errorf("want nil and false but: %v, %v", r, ok)
	}

	m.m = map[interface{}]interface{}{}

	m.m["ttt"] = nil

	if r := m.m["ttt"]; r != nil {
		t.Errorf("want nil but %v", r)
	}

	if r, ok := m.m["ttt"]; r != nil || !ok {
		t.Errorf("want nil and true but: %v, %v", r, ok)
	}
}

func TestNewMap(t *testing.T) {
	m := Map{}

	if r := m.Get(k); r != nil {
		t.Errorf("want nil get: %v ", r)
	}

	if !m.Set(k, 1) {
		t.Error("want true,but false.")
	}

	if m.Set(k, 1) {
		t.Error("want false, but true.")
	}

	if !m.Set(k, "string") {
		t.Error("want true,but false.")
	}

	if !m.Has(k) {
		t.Error("want true, but false.")
	}

	if v := m.Get(k); v != "string" {
		t.Errorf("want string get: %v", v)
	}

	m.Delete(k)

	if m.Has(k) {
		t.Error("want false, but true.")
	}
}

func sum(i, j int) int {
	return i + j
}

func TestGetOrElse(t *testing.T) {
	m := Map{}

	if v := m.Get(k); v != nil {
		t.Errorf("want 0, but %v", v)
	}

	if v := m.GetOrElse(k, func() (interface{}, error) { return sum(1, 2), nil }, 100); v != 3 {
		t.Errorf("want 3, but get: %v", v)
	}

	if v := m.Get(k); v != 3 {
		t.Errorf("want 3, but %v", v)
	}
}

/****************************************************/
func BenchmarkOriginMap(b *testing.B) {
	m := map[string]int{}
	m[k] = 1
	var sum int
	for i := 0; i < b.N; i++ {
		sum += m[k]
	}
}

type hardCodeMap struct {
	sync.RWMutex
	m map[string]int
}

func (h *hardCodeMap) Get(key string) (v int) {
	h.RLock()
	v = h.m[key]
	h.RUnlock()
	return
}

func BenchmarkHardCodeMap(b *testing.B) {
	m := hardCodeMap{}
	m.m = map[string]int{}
	m.m[k] = 1

	var sum int
	for i := 0; i < b.N; i++ {
		m.RLock()
		v := m.m[k]
		m.RUnlock()
		sum += v
	}
}

func BenchmarkHardCodeGet(b *testing.B) {
	h := hardCodeMap{}
	h.m = map[string]int{}
	h.m[k] = 1

	var sum int
	for i := 0; i < b.N; i++ {
		sum += h.Get(k)
	}
}

func BenchmarkSyncMap(b *testing.B) {
	m := Map{}
	m.m = map[interface{}]interface{}{}
	m.m[k] = 1

	var sum int
	for i := 0; i < b.N; i++ {
		sum += m.Get(k).(int)
	}
}

type hardCodeInfKeyMap struct {
	sync.RWMutex
	m map[interface{}]int
}

func (h *hardCodeInfKeyMap) Get(key interface{}) (v int) {
	h.RLock()
	v = h.m[key]
	h.RUnlock()
	return
}

func BenchmarkHardCodeInfKey(b *testing.B) {
	m := hardCodeInfKeyMap{}
	m.m = map[interface{}]int{}
	m.m[k] = 1

	var sum int
	for i := 0; i < b.N; i++ {
		sum += m.Get(k)
	}
}

type hardCodeInfValMap struct {
	sync.RWMutex
	m map[string]interface{}
}

func (h *hardCodeInfValMap) Get(key string) (v interface{}) {
	h.RLock()
	v = h.m[key]
	h.RUnlock()
	return
}

func BenchmarkHardCodeInfVal(b *testing.B) {
	m := hardCodeInfValMap{}
	m.m = map[string]interface{}{}
	m.m[k] = 1

	var sum int
	for i := 0; i < b.N; i++ {
		sum += m.Get(k).(int)
	}
}

type hardCodeMapUseDefer struct {
	sync.RWMutex
	m map[interface{}]interface{}
}

func (h *hardCodeMapUseDefer) Get(key interface{}) (v interface{}) {
	h.RLock()
	v = h.m[key]
	defer h.RUnlock()
	return
}

func BenchmarkHardCodeDefer(b *testing.B) {
	m := hardCodeMapUseDefer{}
	m.m = map[interface{}]interface{}{}
	m.m[k] = 1

	var sum int
	for i := 0; i < b.N; i++ {
		sum += m.Get(k).(int)
	}
}

/**********************************************/
type testObject struct {
	Id   int
	Name string
}

type hardCodeObjectMap struct {
	sync.RWMutex
	m map[testObject]interface{}
}

func (h *hardCodeObjectMap) Get(key testObject) (v interface{}) {
	h.RLock()
	v = h.m[key]
	defer h.RUnlock()
	return
}

func BenchmarkHardCodeObjectMap(b *testing.B) {
	m := hardCodeObjectMap{}
	m.m = map[testObject]interface{}{}
	kk := testObject{1, "name"}
	m.m[kk] = 1

	var sum int
	for i := 0; i < b.N; i++ {
		sum += m.Get(kk).(int)
	}
}

func BenchmarkHardCodeObjectMap1(b *testing.B) {
	m := Map{}
	kk := testObject{1, "name"}
	m.Set(kk, 1)

	var sum int
	for i := 0; i < b.N; i++ {
		sum += m.Get(kk).(int)
	}
}
