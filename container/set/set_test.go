package set

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestSetAddHas(t *testing.T) {
	s := New()
	if s.Has("abc") != false {
		t.Error("empty set has a item must return false")
	}

	s.Add("abc")
	if s.Has("abc") != true {
		t.Error("abc must in set.")
	}
}

func TestSetRemove(t *testing.T) {
	s := New()

	s.Add("abc")
	if s.Has("abc") != true {
		t.Error("set must has item abc")
	}

	s.Remove("abc")
	if s.Has("abc") != false {
		t.Error("set mustn't has item abc")
	}
}

func TestNilSet(t *testing.T) {
	var s Set = nil
	s.Add("abc")

	if s.Has("abc") != false {
		t.Error("nil set must not has item abc")
	}
}

func TestSetForEach(t *testing.T) {
	s := New()
	var num = 0

	s.Add("abc").Add("abd").Add("abe")
	err := s.ForEach(func(value interface{}) error {
		v, _ := value.(string)
		if !strings.HasPrefix(v, "ab") {
			return errors.New("there is a item without prefix ab")
		}
		num += 1
		return nil
	})
	if err != nil {
		t.Error(err)
	}

	if num != s.Len() {
		t.Errorf("count item error, want %d, got %d", s.Len(), num)
	}
}

func TestSetUnion(t *testing.T) {
	// s := New("1", "2", "3")
	// s1 := New("1", "3", "5")
	// s.Union(s1...)
}

func BenchmarkSetAddInt(b *testing.B) {
	s := New()
	for i := 0; i < b.N; i++ {
		s.Add(i)
	}
}

func BenchmarkSetAddString(b *testing.B) {
	s := New()
	for i := 0; i < b.N; i++ {
		s.Add(strconv.Itoa(i))
	}
}
