package selector

import (
	"errors"
	"regexp"
	"testing"
)

func TestGlobOptions(t *testing.T) {
	glob, err := Parse("d`abc")
	if err != nil {
		t.Error(err)
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

	if glob.incDir != true ||
		glob.incFile != true {
		t.Error("default option error.")
	}
}

func TestRegexp(t *testing.T) {
	// regex := regexp.MustCompile("^ab[[^/]&&[c]]$")
	regex := regexp.MustCompile("^[]a]$")
	if regex.MatchString("]") != true {
		t.Errorf("abc must be matched.")
	}
}

type MatchTest struct {
	pattern, s string
	match      bool
	err        error
}

var match1Tests = []MatchTest{
	{"[\\]a]", "]", true, nil},
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
	{"[]a]", "]", true, nil},
	{"[-]", "-", false, errors.New("error parsing regexp: missing closing ]: `[^]$`")},
	{"[x-]", "x", true, nil},
	{"[x-]", "-", true, nil},
	{"[x-]", "z", false, nil},
	{"[-x]", "x", false, nil},
	{"[-x]", "-", true, nil},
	{"[-x]", "a", true, nil},
	{"\\", "a", false, errors.New("No character to escape at index 1")},
	{"[a-b-c]", "a", false, errors.New("Invalid range: 5")},
	{"[", "a", false, errors.New("Missing ']': 1")},
	{"[^", "a", false, errors.New("Missing ']': 2")},
	{"[^bc", "a", false, errors.New("Missing ']': 4")},
	{"a[", "a", false, errors.New("Missing ']': 2")},
	{"a[", "ab", false, errors.New("Missing ']': 2")},
	{"*x", "xxx", true, nil},
	// double star
	{"**", "xx/bb/cc", true, nil},
	{"xx/**", "xx/bb/cc", true, nil},
	{"xx/**", "xx/", true, nil},
	{"test/a/*/(c|g)/./d", "test/a/b/c/./d", true, nil},

	{"test/a/**/[cg]/../[cg]", "test/a/abcdef/g/../g", true, nil},
	{"test/a/**/[cg]/../[cg]", "test/a/abcfed/g/../g", true, nil},
	{"test/a/**/[cg]/../[cg]", "test/a/b/c/../c", true, nil},
	{"test/a/**/[cg]/../[cg]", "test/a/c/../c", true, nil},
	{"test/a/**/[cg]/../[cg]", "test/a/c/d/c/../c", true, nil},
	{"test/a/**/[cg]/../[cg]", "test/a/symlink/a/b/c/../c", true, nil},

	{"test/**/g", "test/a/abcdef/g", true, nil},
	{"test/**/g", "test/a/abcfed/g", true, nil},

	{"test/a/abc{fed,def}/g/h", "test/a/abcdef/g/h", true, nil},
	{"test/a/abc{fed,def}/g/h", "test/a/abcfed/g/h", true, nil},

	{"test/a/abc{fed/g,def}/**/", "test/a/abcdef/", true, nil},
	{"test/a/abc{fed/g,def}/**/", "test/a/abcdef/g", true, nil},
	{"test/a/abc{fed/g,def}/**/", "test/a/abcfed/g/", true, nil},
}

func TestMatch(t *testing.T) {
	for _, tt := range matchTests {
		if glob, err := Parse(tt.pattern); err != nil {
			if tt.err == nil {
				t.Errorf("Parse %s err: %v", tt.pattern, err)
			} else if err.Error() != tt.err.Error() {
				t.Errorf("Parse %s. want err: %v, got err: %v", tt.pattern, tt.err, err)
			}

		} else {
			if glob.Match(tt.s) != tt.match {
				t.Errorf("Match %s and pattern is %s. want %v got %v: debug: %s", tt.s, tt.pattern, tt.match, !tt.match, glob.debug)
			}
		}
	}
}
