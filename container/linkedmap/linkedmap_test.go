package linkedmap

import (
	"testing"
)

func TestLinkedmapGetAndIsExist(t *testing.T) {
	m := New()

	m.Append("abc", "123")
	m.Append("bbb", "345")
	m.Append("ccc", "987")

	if v, ok := m.Get("abc"); !ok || v != "123" {
		t.Errorf("want 123, got %s", v)
	}

	if v, ok := m.Get("py"); ok || v != nil {
		t.Errorf("want nil, got %s", v)
	}

	if m.IsExist("bbb") != true {
		t.Error("bbb must exist")
	}

	if m.IsExist("ttt") != false {
		t.Error("ttt must not exist")
	}

	if m.Remove("ddd") != false {
		t.Error("remove ddd must be false")
	}
}
