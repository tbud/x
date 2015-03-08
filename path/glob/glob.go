package glob

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
)

const glob_option_split_char = "`"

type Glob struct {
	pattern string         // origin pattern
	regexp  *regexp.Regexp // save parsed match regexp
	incDir  bool           // include dir
	incFile bool           // include file
	incHide bool           // include hiden file/dir
	isNeg   bool           // negative regexp
	debug   string
}

func Parse(pattern string) (g Glob, err error) {
	g = Glob{pattern: pattern}
	return g, g.Parse()
}

func (g *Glob) Parse() (err error) {
	pattern := g.pattern
	if strings.Contains(pattern, glob_option_split_char) {
		splitedPattern := strings.Split(pattern, glob_option_split_char)
		if len(splitedPattern) != 2 {
			return fmt.Errorf("parse glob option err: %s", pattern)
		}

		if err = g.parseOptions(splitedPattern[0]); err != nil {
			return
		}

		pattern = splitedPattern[1]
	}

	if !g.incFile && !g.incDir {
		g.incFile = true
		g.incDir = true
	}

	return g.parsePattern(pattern)
}

func (g *Glob) Match(path string) bool {
	return g.regexp.MatchString(path)
}

func (g *Glob) String() string {
	return g.pattern
}

func (g *Glob) parseOptions(options string) error {
	for _, char := range options {
		switch char {
		default:
			return fmt.Errorf("Unkown option char %c", char)
		case 'd', 'D':
			g.incDir = true
		case 'f', 'F':
			g.incFile = true
		case 'h', 'H':
			g.incHide = true
		}
	}

	return nil
}

func (g *Glob) parsePattern(pattern string) (err error) {
	if len(pattern) == 0 {
		return fmt.Errorf("pattern is empty")
	}

	if pattern[0] == '!' {
		g.isNeg = true

		pattern = pattern[1:]
		if len(pattern) == 0 {
			return fmt.Errorf("neg pattern is empty")
		}
	}

	regex := []rune("^")
	isWindows := runtime.GOOS == "windows"

	patternRune := []rune(pattern)
	patternLen := len(patternRune)

	inGroup := false
	var i int
	for i = 0; i < patternLen; i++ {
		c := patternRune[i]
		switch c {
		case '\\':
			if i+1 == patternLen {
				return fmt.Errorf("No character to escape at index %d", i+1)
			}
			next := next(patternRune, i, patternLen)
			i += 1
			if isGlobMeta(next) || isRegexMeta(next) {
				regex = append(regex, '\\')
			}
			regex = append(regex, next)
		case '/':
			if isWindows {
				regex = append(regex, []rune("\\\\")...)
			} else {
				regex = append(regex, c)
			}
		case '[':
			regex = append(regex, c)
			switch next(patternRune, i, patternLen) {
			case '!', '-', '^':
				regex = append(regex, '^')
				i += 1
			}

			hasRangeStart := false
			var last rune = 0
			for i+1 < patternLen {
				c = next(patternRune, i, patternLen)
				i += 1
				if c == ']' {
					break
				}
				if c == '/' || (isWindows && c == '\\') {
					return fmt.Errorf("Explicit 'name separator' in class: %d", i+1)
				}
				// if c == '\\' || c == '[' || c == '&' && next(patternRune, i, patternLen) == '&' {
				// 	regex = append(regex, '\\')
				// }

				regex = append(regex, c)
				if c == '-' {
					if !hasRangeStart {
						return fmt.Errorf("Invalid range: %d", i+1)
					}
					c = next(patternRune, i, patternLen)
					i += 1
					if c == 0 || c == ']' {
						break
					}
					if c < last {
						fmt.Errorf("Invalid range: %d", i-2)
					}
					regex = append(regex, c)
					hasRangeStart = false
				} else {
					hasRangeStart = true
					last = c
				}
			}
			if c != ']' {
				return fmt.Errorf("Missing ']': %d", i+1)
			}
			regex = append(regex, ']')

		case '{':
			if inGroup {
				return fmt.Errorf("Cannot nest groups :%d", i)
			}
			regex = append(regex, []rune("(?:(?:")...)
			inGroup = true
		case '}':
			if inGroup {
				regex = append(regex, []rune("))")...)
				inGroup = false
			} else {
				regex = append(regex, c)
			}
		case ',':
			if inGroup {
				regex = append(regex, []rune(")|(?:")...)
			} else {
				regex = append(regex, c)
			}
		case '*':
			if next(patternRune, i, patternLen) == '*' {
				regex = append(regex, []rune(".*")...)
				i += 1
				if next(patternRune, i, patternLen) == '/' {
					i += 1
				}
			} else {
				if isWindows {
					regex = append(regex, []rune("[^\\\\]*")...)
				} else {
					regex = append(regex, []rune("[^/]*")...)
				}
			}
		case '?':
			if isWindows {
				regex = append(regex, []rune("[^\\\\]")...)
			} else {
				regex = append(regex, []rune("[^/]")...)
			}
		default:
			// if isRegexMeta(c) {
			// 	regex = append(regex, '\\')
			// }
			regex = append(regex, c)
		}
	}

	if inGroup {
		return fmt.Errorf("Missing }: ", i-1)
	}

	regex = append(regex, '$')

	g.debug = string(regex)
	g.regexp, err = regexp.Compile(string(regex))
	return
}

func next(patternRune []rune, i int, patternLen int) rune {
	if i+1 < patternLen {
		return patternRune[i+1]
	}
	return 0
}

func isRegexMeta(r rune) bool {
	return strings.Contains(".^$+{[]|()", string(r))
}

func isGlobMeta(r rune) bool {
	return strings.Contains("\\*?[{", string(r))
}
