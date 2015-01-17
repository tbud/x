package sync

import (
	"testing"
)

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
	k := "ttt"

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

	if v := m.GetOrElse("222", func() (interface{}, error) { return sum(1, 2), nil }, 100); v != 3 {
		t.Errorf("want 3, but get: %v", v)
	}
}
