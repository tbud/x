package selector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Selector struct {
	negGlobs []*Glob
	globs    []*Glob
}

func New(patterns ...string) (selector *Selector, err error) {
	if len(patterns) == 0 {
		panic(fmt.Errorf("pattern is empty"))
	}

	selector = &Selector{}
	var g *Glob
	for _, pattern := range patterns {
		if g, err = Parse(pattern); err != nil {
			return nil, err
		}

		if g.isNeg {
			selector.negGlobs = append(selector.negGlobs, g)
		} else {
			selector.globs = append(selector.globs, g)
		}
	}

	return selector, nil
}

func (s *Selector) Matches(root string) (matches []string, err error) {
	err = s.Walk(root, func(path string, info os.FileInfo, err error) error {
		matches = append(matches, path)
		return nil
	})
	return
}

func (s *Selector) Walk(root string, walkFn filepath.WalkFunc) (err error) {
	if !filepath.IsAbs(root) {
		if root, err = filepath.Abs(root); err != nil {
			return err
		}
	}

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return err
		}

		shortPath := strings.TrimPrefix(path, root)
		if len(shortPath) > 0 {
			shortPath = shortPath[1:]
		}

		for _, negGlob := range s.negGlobs {
			if negGlob.Match(shortPath) && negGlob.checkType(info.IsDir()) {
				if info.IsDir() {
					return filepath.SkipDir
				} else {
					return nil
				}
			}
		}

		for _, glob := range s.globs {
			if glob.Match(shortPath) && glob.checkType(info.IsDir()) {
				return walkFn(path, info, err)
			}
		}

		return nil
	})
}

func (s *Selector) String() (ret string) {
	for _, negGlob := range s.negGlobs {
		ret += fmt.Sprintf("neg glob: %s, debug: %s\n", negGlob.pattern, negGlob.debug)
	}

	for _, glob := range s.globs {
		ret += fmt.Sprintf("glob: %s, debug: %s\n", glob.pattern, glob.debug)
	}

	return
}
