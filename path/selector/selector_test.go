package selector

import (
	"strings"
	"testing"
)

func TestSelect(t *testing.T) {
	selector, err := New("f`x/**", "!**/.git", "!**/*.(go|conf)")
	if err != nil {
		t.Errorf("%v, error: %v", selector, err)
	}

	matches, err := selector.Matches("../../..")
	if err != nil {
		t.Errorf("%v, error: %v", selector, err)
	}

	if len(matches) > 0 {
		hasReadMe := false
		for _, match := range matches {
			if strings.HasSuffix(match, "README.md") {
				hasReadMe = true
			}
		}
		if !hasReadMe {
			t.Errorf("missing readme file. %v", matches)
		}
	} else {
		t.Error("parse error, matches nothing.")
	}
}
