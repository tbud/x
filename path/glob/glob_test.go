package glob

import (
	"regexp"
	"testing"
)

func TestGlobOptions(t *testing.T) {
	glob, err := Parse("dh`abc")
	if err != nil {
		t.Error(err)
	}

	if glob.incHide != true {
		t.Error("glob must include hide file or dir")
	}

	if glob.incDir != true {
		t.Error("glob must include dir")
	}

	if glob.incFile != false {
		t.Error("glob must not include file")
	}
}

func TestGlobNoOptionSet(t *testing.T) {
	glob, err := Parse("*.go")
	if err != nil {
		t.Error(err)
	}

	if glob.incHide != false ||
		glob.incDir != true ||
		glob.incFile != true {
		t.Error("default option error.")
	}
}

func TestRegexp(t *testing.T) {
	// regex := regexp.MustCompile("^ab[[^/]&&[c]]$")
	regex := regexp.MustCompile("^ab[c]$")
	if regex.MatchString("abc") != true {
		t.Errorf("abc must be matched.")
	}
}

type MatchTest struct {
	pattern, s string
	match      bool
	err        error
}

var matchTests = []MatchTest{
	{"abc", "abc", true, nil},
	{"*", "abc", true, nil},
	{"*c", "abc", true, nil},
	{"a*", "a", true, nil},
	{"a*", "abc", true, nil},
	{"a*", "ab/c", false, nil},
	{"a*/b", "abc/b", true, nil},
	{"a*/b", "a/c/b", false, nil},
	{"a*b*c*d*e*/f", "axbxcxdxe/f", true, nil},
	{"a*b*c*d*e*/f", "axbxcxdxexxx/f", true, nil},
	{"a*b*c*d*e*/f", "axbxcxdxe/xxx/f", false, nil},
	{"a*b*c*d*e*/f", "axbxcxdxexxx/fff", false, nil},
	{"a*b?c*x", "abxbbxdbxebxczzx", true, nil},
	{"a*b?c*x", "abxbbxdbxebxczzy", false, nil},
	{"ab[c]", "abc", true, nil},
	{"ab[b-d]", "abc", true, nil},
	{"ab[e-g]", "abc", false, nil},
	{"ab[^c]", "abc", false, nil},
	{"ab[^b-d]", "abc", false, nil},
	{"ab[^e-g]", "abc", true, nil},
	{"a\\*b", "a*b", true, nil},
	{"a\\*b", "ab", false, nil},
	{"a?b", "a☺b", true, nil},
	{"a[^a]b", "a☺b", true, nil},
	{"a???b", "a☺b", false, nil},
	{"a[^a][^a][^a]b", "a☺b", false, nil},
	{"[a-ζ]*", "α", true, nil},
	{"*[a-ζ]", "A", false, nil},
	{"a?b", "a/b", false, nil},
	{"a*b", "a/b", false, nil},
	{"[\\]a]", "]", true, nil},
	{"[\\-]", "-", true, nil},
	{"[x\\-]", "x", true, nil},
	{"[x\\-]", "-", true, nil},
	{"[x\\-]", "z", false, nil},
	{"[\\-x]", "x", true, nil},
	{"[\\-x]", "-", true, nil},
	{"[\\-x]", "a", false, nil},
	{"[]a]", "]", false, nil},
	{"[-]", "-", false, nil},
	{"[x-]", "x", false, nil},
	{"[x-]", "-", false, nil},
	{"[x-]", "z", false, nil},
	{"[-x]", "x", false, nil},
	{"[-x]", "-", false, nil},
	{"[-x]", "a", false, nil},
	{"\\", "a", false, nil},
	{"[a-b-c]", "a", false, nil},
	{"[", "a", false, nil},
	{"[^", "a", false, nil},
	{"[^bc", "a", false, nil},
	{"a[", "a", false, nil},
	{"a[", "ab", false, nil},
	{"*x", "xxx", true, nil},
}

func TestMatch(t *testing.T) {
	for _, tt := range matchTests {
		if glob, err := Parse(tt.pattern); err != nil {
			t.Errorf("Parse %s err: %v, ", tt.pattern, err)
		} else {
			if glob.Match(tt.s) != tt.match {
				t.Errorf("Match %s and pattern is %s. want %v got %v: debug: %s", tt.s, tt.pattern, tt.match, !tt.match, glob.debug)
			}
		}
	}
}
