package set

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestStringSetAddHas(t *testing.T) {
	s := NewStringSet()
	if s.Has("abc") != false {
		t.Error("empty set has a item must return false")
	}

	s.Add("abc")
	if s.Has("abc") != true {
		t.Error("abc must in set.")
	}
}

func TestStringSetRemove(t *testing.T) {
	s := NewStringSet()

	s.Add("abc")
	if s.Has("abc") != true {
		t.Error("set must has item abc")
	}

	s.Remove("abc")
	if s.Has("abc") != false {
		t.Error("set mustn't has item abc")
	}
}

func TestNilStringSet(t *testing.T) {
	var s StringSet = nil
	s.Add("abc")

	if s.Has("abc") != false {
		t.Error("nil set must not has item abc")
	}
}

func TestStringSetForEach(t *testing.T) {
	s := NewStringSet()
	var num = 0

	s.Add("abc").Add("abd").Add("abe")
	err := s.ForEach(func(value string) error {
		if !strings.HasPrefix(value, "ab") {
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

func TestStringSetUnion(t *testing.T) {
	s := NewStringSet("1", "2", "3")
	s1 := NewStringSet("1", "3", "5")

	s.Union(s1.ToSeq()...)
	want := NewStringSet("1", "2", "3", "5")
	if !reflect.DeepEqual(s, want) {
		t.Errorf("string set error, want:%v, got:%v", want, s)
	}
}

func TestStringSetSubtract(t *testing.T) {
	s := NewStringSet("1", "2", "3")
	s1 := NewStringSet("1", "3", "5")

	s.Subtract(s1.ToSeq()...)
	want := NewStringSet("2")
	if !reflect.DeepEqual(s, want) {
		t.Errorf("string set error, want:%v, got:%v", want, s)
	}
}

func BenchmarkStringSetAdd(b *testing.B) {
	s := NewStringSet()
	for i := 0; i < b.N; i++ {
		s.Add(strconv.Itoa(i))
	}
}
