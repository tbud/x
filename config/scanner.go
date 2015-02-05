package config

// HOCON value parser state machine.
// Just about at the limit of what is reasonable to write by hand.
// Some parts are a bit tedious, but overall it nicely factors out the
// otherwise common code from the multiple scanning functions
// in this package (Compact, Indent, checkValid, nextValue, etc).
//
// This file starts with two simple examples using the scanner
// before diving into the scanner itself.

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// A SyntaxError is a description of a HOCON syntax error.
type SyntaxError struct {
	msg    string // description of error
	buf    []byte // save file scan parse buf
	Offset int64  // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string {
	return e.msg + strconv.Quote(string(e.buf)) + " Offset " + strconv.FormatInt(e.Offset, 10)
}

// A scanner is a HOCON scanning state machine.
// Callers call scan.reset() and then pass bytes in one at a time
// by calling scan.step(&scan, c) for each byte.
// The return value, referred to as an opcode, tells the
// caller about significant parsing events like beginning
// and ending literals, objects, and arrays, so that the
// caller can follow along if it wishes.
// The return value scanEnd indicates that a single top-level
// HOCON value has been completed, *before* the byte that
// just got passed in.  (The indication must be delayed in order
// to recognize the end of numbers: is 123 a whole value or
// the beginning of 12345e+6?).
type fileScanner struct {
	// The step is a func to be called to execute the next transition.
	// Also tried using an integer constant and a single func
	// with a switch, but using the func directly was 10% faster
	// on a 64-bit Mac Mini, and it's nicer to read.
	step func(*fileScanner, int) int

	// Error that happened, if any.
	err error

	// total bytes consumed, updated by decoder.Decode
	bytes int64

	file         string   // file name
	dir          string   // dir of the file
	data         []byte   // store data load from file
	baseKeys     []string // save base key
	keyStack     []int    // stack for key
	parseBuf     []byte   // save parsed key or value
	bufType      int      // buf type
	currentState int      // save current state
	kvs          []kvPair //save key value
}

type kvPair struct {
	keys  []string
	value interface{} // value or filescanner
}

// These values are returned by the state transition functions
// assigned to scanner.state and the method scanner.eof.
// They give details about the current state of the scan that
// callers might be interested to know about.
// It is okay to ignore the return value of any particular
// call to scanner.state: if one call returns scanError,
// every subsequent call will return scanError too.
const (
	// Continue.
	scanContinue  = iota // uninteresting byte
	scanSkipSpace        // space byte; can skip; known to be last "continue" result
	scanAppendBuf        // byte need to append buf

	// Stop.
	scanEnd   // top-level value ended *before* this byte; known to be first "stop" result
	scanError // hit an error, scanner.err.
)

// These values are stored in the parseState stack.
// They give the current state of a composite value
// being scanned.  If the parser is inside a nested value
// the parseState describes the nested state, outermost at entry 0.
const (
	parseKey        = iota // parsing object key (before colon)
	parseValue             // parsing object value (after colon)
	parseArrayValue        // parsing array value
	parseError
)

const (
	bufTypeNull = iota
	bufTypeString
	bufTypeNoQuoteString
	bufTypeNumber
	bufTypeBoolTrue
	bufTypeBoolFalse
	bufTypeInclude
)

// reset prepares the scanner for use.
// It must be called before calling s.step.
func (s *fileScanner) init() {
	s.step = stateBeginKey
	s.err = nil
	s.bytes = 0
	s.currentState = parseKey
}

// checkValid verifies that data is valid HOCON-encoded data.
// scan is passed in for use by checkValid to avoid an allocation.
func (s *fileScanner) checkValid(fileName string) error {
	s.init()

	if !filepath.IsAbs(fileName) {
		return s.errorSyntax("file '" + fileName + "' is not absolute path")
	}
	var err error
	s.data, err = ioutil.ReadFile(fileName)
	if err != nil {
		return s.errorSyntax(err.Error())
	}

	s.file = filepath.Base(fileName)
	s.dir = filepath.Dir(fileName)

	for _, c := range s.data {
		s.bytes++
		switch s.step(s, int(c)) {
		case scanError:
			return s.err
		case scanAppendBuf:
			s.parseBuf = append(s.parseBuf, c)
		}
	}

	// stateEndValue(s, '\n')

	// for i := range s.includesScanner {
	// 	scan := s.includesScanner[i]
	// 	var checkFile string
	// 	if filepath.IsAbs(scan.file) {
	// 		checkFile = scan.file
	// 	} else {
	// 		checkFile = filepath.Join(s.dir, scan.file)
	// 	}
	// 	err := scan.checkValid(checkFile)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

func (s *fileScanner) setOptions(options map[string]interface{}) error {
	for i := range s.kvs {
		kv := s.kvs[i]
		if _, ok := kv.value.(*fileScanner); ok {
			if err := setIncludeValue(kv, options); err != nil {
				return err
			}
		} else {
			if err := setOptionValue(kv, options); err != nil {
				return err
			}
		}
	}
	return nil
}

func setOptionValue(kv kvPair, options map[string]interface{}) error {
	for i := 0; i < len(kv.keys)-1; i++ {
		key := kv.keys[i]
		if optMap, ok := options[key]; ok {
			if v, ok := optMap.(map[string]interface{}); ok {
				options = v
			} else {
				fmt.Println(kv.keys, reflect.TypeOf(optMap))
				return &SyntaxError{"set options value error, key " + key + " is not a map[string]interface{}", []byte{}, int64(i)}
			}
		} else {
			options[key] = map[string]interface{}{}
			options = options[key].(map[string]interface{})
		}
	}

	options[kv.keys[len(kv.keys)-1]] = kv.value
	return nil
}

func setIncludeValue(kv kvPair, options map[string]interface{}) error {
	for i := 0; i < len(kv.keys); i++ {
		key := kv.keys[i]
		if optMap, ok := options[key]; ok {
			if v, ok := optMap.(map[string]interface{}); ok {
				options = v
			} else {
				return &SyntaxError{"set options value error, key " + key + " is not a map[string]interface{}", []byte{}, int64(i)}
			}
		} else {
			options[key] = map[string]interface{}{}
			options = options[key].(map[string]interface{})
		}
	}

	kv.value.(*fileScanner).setOptions(options)
	return nil
}

func (s *fileScanner) pushKeyStack() {
	s.keyStack = append(s.keyStack, len(s.baseKeys))
}

func (s *fileScanner) popKeyStack() {
	switch keylen := len(s.keyStack); {
	case keylen > 1:
		stack := s.keyStack[keylen-2]
		s.baseKeys = s.baseKeys[0:stack]
		s.keyStack = s.keyStack[0 : keylen-1]
	case keylen == 1:
		s.baseKeys = s.baseKeys[0:0]
		s.keyStack = s.keyStack[0:0]
	case keylen <= 0:
		panic(s.errorSyntax("error pop key stack operate"))
	}
}

func (s *fileScanner) pushKey() {
	s.baseKeys = append(s.baseKeys, string(s.parseBuf))
	s.parseBuf = s.parseBuf[0:0]
}

func (s *fileScanner) pushValue() {
	baseKeys := make([]string, len(s.baseKeys))
	copy(baseKeys, s.baseKeys)
	if len(s.parseBuf) == 0 {
		s.kvs = append(s.kvs, kvPair{baseKeys, nil})
	}

	s.kvs = append(s.kvs, kvPair{baseKeys, s.parseBufValue()})

	// pop base key
	if stackLen := len(s.keyStack); stackLen > 0 {
		stack := s.keyStack[stackLen-1]
		s.baseKeys = s.baseKeys[0:stack]
	} else {
		s.baseKeys = s.baseKeys[0:0]
	}

	// reinit parse buf
	s.parseBuf = s.parseBuf[0:0]
	s.bufType = bufTypeNull
}

func (s *fileScanner) pushArrayKey() {
	baseKeys := make([]string, len(s.baseKeys))
	copy(baseKeys, s.baseKeys)

	s.kvs = append(s.kvs, kvPair{baseKeys, []interface{}{}})

	if stackLen := len(s.keyStack); stackLen > 0 {
		stack := s.keyStack[stackLen-1]
		s.baseKeys = s.baseKeys[0:stack]
	} else {
		s.baseKeys = s.baseKeys[0:0]
	}
}

func (s *fileScanner) pushArrayValue() {
	kv := &s.kvs[len(s.kvs)-1]
	kv.value = append(kv.value.([]interface{}), s.parseBufValue())

	s.parseBuf = s.parseBuf[0:0]
	s.bufType = bufTypeNull
}

// parse buf to type value
func (s *fileScanner) parseBufValue() (value interface{}) {
	var err error

	if s.bufType == bufTypeNoQuoteString {
		valueKeyword.checkKeyword(s, s.parseBuf)
	}

	switch s.bufType {
	case bufTypeString, bufTypeNoQuoteString:
		if b, ok := stringBytes(s.parseBuf); ok {
			value = string(b)
		}
		// value = string(s.parseBuf)
	case bufTypeNumber:
		if value, err = strconv.ParseFloat(string(s.parseBuf), 64); err != nil {
			panic(s.errorSyntax("number " + string(s.parseBuf) + " parse error: " + err.Error()))
		}
	case bufTypeBoolTrue, bufTypeBoolFalse:
		value = s.bufType == bufTypeBoolTrue
	case bufTypeNull:
		value = nil
	}
	return
}

func (s *fileScanner) trimParseBuf() {
	s.parseBuf = []byte(strings.TrimRight(string(s.parseBuf), " "))
}

func isSpace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}

func stateBeginKey(s *fileScanner, c int) int {
	if c <= ' ' && isSpace(rune(c)) {
		return scanSkipSpace
	}
	switch c {
	case '{':
		s.pushKeyStack()
		s.step = stateBeginKey
		return scanContinue
	case '"':
		s.step = stateInString
		return scanContinue
	case '#':
		s.step = stateComment
		return scanContinue
	case '}':
		s.popKeyStack()
		s.step = stateBeginKey
		return scanContinue
	}

	if unicode.IsLetter(rune(c)) {
		s.step = stateNoQuoteString
		return stateNoQuoteString(s, c)
	}

	if len(s.keyStack) > 0 && c == ',' {
		s.step = stateBeginKey
		return scanContinue
	}

	if len(s.baseKeys) > 0 && c == 'i' {
		s.step = stateI
		return scanContinue
	}

	return s.error(c, "looking for beginning")
}

// stateBeginValue is the state at the beginning of the input.
func stateBeginValue(s *fileScanner, c int) int {
	if c <= ' ' {
		if s.currentState == parseArrayValue {
			if isSpace(rune(c)) {
				return scanSkipSpace
			}
		} else {
			if c == ' ' || c == '\t' {
				return scanSkipSpace
			}
		}
	}

	switch c {
	case '{':
		s.pushKeyStack()
		s.step = stateBeginKey
		s.currentState = parseKey
		return scanContinue
	case '[':
		s.pushArrayKey()
		s.step = stateBeginValue
		s.currentState = parseArrayValue
		return scanContinue
	case '"':
		s.step = stateInString
		s.bufType = bufTypeString
		return scanContinue
	case '-':
		s.step = stateNeg
		s.bufType = bufTypeNumber
		return scanAppendBuf
	case '0': // beginning of 0.123
		s.step = state0
		s.bufType = bufTypeNumber
		return scanAppendBuf
	case '#':
		s.step = stateComment
		return stateEndValue(s, c)
	case '\r', '\n':
		s.step = stateEndValue
		return stateEndValue(s, c)
	}
	if '1' <= c && c <= '9' { // beginning of 1234.5
		s.step = state1
		s.bufType = bufTypeNumber
		return scanAppendBuf
	}
	if unicode.IsLetter(rune(c)) {
		s.step = stateNoQuoteString
		s.bufType = bufTypeNoQuoteString
		return stateNoQuoteString(s, c)
	}
	return s.error(c, "looking for beginning of value")
}

// stateEndValue is the state after completing a value,
// such as after reading `{}` or `true` or `["x"`.
func stateEndValue(s *fileScanner, c int) int {
	if c <= ' ' && (c == '\t' || c == ' ') {
		s.step = stateEndValue
		return scanSkipSpace
	}
	switch s.currentState {
	case parseKey:
		s.pushKey()
		switch c {
		case ':', '=':
			s.currentState = parseValue
			s.step = stateBeginValue
			return scanContinue
		case '.':
			s.step = stateBeginKey
			return scanContinue
		case '{':
			s.pushKeyStack()
			s.step = stateBeginKey
			return scanContinue
		case '\r', '\n', '#':
			s.currentState = parseValue
			s.step = stateEndValue
			return stateEndValue(s, c)
		}
		return s.error(c, "after object key")
	case parseValue:
		s.pushValue()
		s.currentState = parseKey
		switch c {
		case ',', '\r', '\n':
			s.step = stateBeginKey
			return scanContinue
		case '}':
			s.popKeyStack()
			s.step = stateBeginKey
			return scanContinue
		case '#':
			s.step = stateComment
			return scanContinue
		}
		return s.error(c, "after object key:value pair")
	case parseArrayValue:
		s.pushArrayValue()
		switch c {
		case ',', '\r', '\n':
			s.step = stateBeginValue
			return scanContinue
		case ']':
			s.step = stateBeginKey
			s.currentState = parseKey
			return scanContinue
		case '#':
			s.step = stateComment
			return scanContinue
		}
		return s.error(c, "after array element")
	}
	return s.error(c, "")
}

// stateInString is the state after reading `"`.
func stateInString(s *fileScanner, c int) int {
	switch {
	case c == '"':
		s.step = stateEndValue
		return scanContinue
	case c == '\\':
		s.step = stateInStringEsc
		return scanAppendBuf
	case c < 0x20:
		return s.error(c, "in string literal")
	}

	return scanAppendBuf
}

func stateNoQuoteString(s *fileScanner, c int) int {
	if s.currentState == parseKey {
		switch c {
		case '.', '=', ':', '{', '\r', '\n', '#':
			s.trimParseBuf()
			return stateEndValue(s, c)
		}
	}
	if s.currentState == parseValue {
		switch c {
		case ',', '\r', '\n', '}', '#':
			s.trimParseBuf()
			return stateEndValue(s, c)
		}
	}
	if s.currentState == parseArrayValue {
		switch c {
		case ',', '\r', '\n', ']', '#':
			s.trimParseBuf()
			return stateEndValue(s, c)
		}
	}

	if c < 0x20 || c == '\\' {
		return s.error(c, "in no quote string literal")
	}
	return scanAppendBuf
}

func stateComment(s *fileScanner, c int) int {
	if c == '\n' || c == '\r' {
		if s.currentState == parseArrayValue {
			s.step = stateBeginValue
		} else {
			s.step = stateBeginKey
		}
		return scanContinue
	}
	return scanContinue
}

// stateInStringEsc is the state after reading `"\` during a quoted string.
func stateInStringEsc(s *fileScanner, c int) int {
	switch c {
	case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
		s.step = stateInString
		return scanAppendBuf
	}
	if c == 'u' {
		s.step = stateInStringEscU
		return scanAppendBuf
	}
	return s.error(c, "in string escape code")
}

// stateInStringEscU is the state after reading `"\u` during a quoted string.
func stateInStringEscU(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInStringEscU1
		return scanAppendBuf
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU1 is the state after reading `"\u1` during a quoted string.
func stateInStringEscU1(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInStringEscU12
		return scanAppendBuf
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU12 is the state after reading `"\u12` during a quoted string.
func stateInStringEscU12(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInStringEscU123
		return scanAppendBuf
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU123 is the state after reading `"\u123` during a quoted string.
func stateInStringEscU123(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInString
		return scanAppendBuf
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateNeg is the state after reading `-` during a number.
func stateNeg(s *fileScanner, c int) int {
	if c == '0' {
		s.step = state0
		return scanAppendBuf
	}
	if '1' <= c && c <= '9' {
		s.step = state1
		return scanAppendBuf
	}
	return s.error(c, "in numeric literal")
}

// state1 is the state after reading a non-zero integer during a number,
// such as after reading `1` or `100` but not `0`.
func state1(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' {
		s.step = state1
		return scanAppendBuf
	}
	return state0(s, c)
}

// state0 is the state after reading `0` during a number.
func state0(s *fileScanner, c int) int {
	if c == '.' {
		s.step = stateDot
		return scanAppendBuf
	}
	if c == 'e' || c == 'E' {
		s.step = stateE
		return scanAppendBuf
	}
	return stateEndValue(s, c)
}

// stateDot is the state after reading the integer and decimal point in a number,
// such as after reading `1.`.
func stateDot(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' {
		s.step = stateDot0
		return scanAppendBuf
	}
	return s.error(c, "after decimal point in numeric literal")
}

// stateDot0 is the state after reading the integer, decimal point, and subsequent
// digits of a number, such as after reading `3.14`.
func stateDot0(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' {
		s.step = stateDot0
		return scanAppendBuf
	}
	if c == 'e' || c == 'E' {
		s.step = stateE
		return scanAppendBuf
	}
	return stateEndValue(s, c)
}

// stateE is the state after reading the mantissa and e in a number,
// such as after reading `314e` or `0.314e`.
func stateE(s *fileScanner, c int) int {
	if c == '+' {
		s.step = stateESign
		return scanAppendBuf
	}
	if c == '-' {
		s.step = stateESign
		return scanAppendBuf
	}
	return stateESign(s, c)
}

// stateESign is the state after reading the mantissa, e, and sign in a number,
// such as after reading `314e-` or `0.314e+`.
func stateESign(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' {
		s.step = stateE0
		return scanAppendBuf
	}
	return s.error(c, "in exponent of numeric literal")
}

// stateE0 is the state after reading the mantissa, e, optional sign,
// and at least one digit of the exponent in a number,
// such as after reading `314e-2` or `0.314e+1` or `3.14e0`.
func stateE0(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' {
		s.step = stateE0
		return scanAppendBuf
	}
	return stateEndValue(s, c)
}

func stateI(s *fileScanner, c int) int {
	if c == 'n' {
		s.step = stateIn
		return scanAppendBuf
	}
	return s.error(c, "in literal include (expecting 'n')")
}

func stateIn(s *fileScanner, c int) int {
	if c == 'c' {
		s.step = stateInc
		return scanAppendBuf
	}
	return s.error(c, "in literal include (expecting 'c')")
}

func stateInc(s *fileScanner, c int) int {
	if c == 'l' {
		s.step = stateIncl
		return scanAppendBuf
	}
	return s.error(c, "in literal include (expecting 'l')")
}

func stateIncl(s *fileScanner, c int) int {
	if c == 'u' {
		s.step = stateInclu
		return scanAppendBuf
	}
	return s.error(c, "in literal include (expecting 'u')")
}

func stateInclu(s *fileScanner, c int) int {
	if c == 'd' {
		s.step = stateInclud
		return scanAppendBuf
	}
	return s.error(c, "in literal include (expecting 'd')")
}

func stateInclud(s *fileScanner, c int) int {
	if c == 'e' {
		s.step = stateEndValue
		return scanAppendBuf
	}
	return s.error(c, "in literal include (expecting 'e')")
}

// stateError is the state after reaching a syntax error,
// such as after reading `[1}` or `5.1.2`.
func stateError(s *fileScanner, c int) int {
	return scanError
}

// error records an error and switches to the error state.
func (s *fileScanner) error(c int, context string) int {
	s.step = stateError
	s.err = &SyntaxError{"invalid character " + quoteChar(c) + " " + context, s.parseBuf, s.bytes}
	return scanError
}

func (s *fileScanner) errorSyntax(context string) *SyntaxError {
	return &SyntaxError{context, s.parseBuf, s.bytes}
}

// quoteChar formats c as a quoted character literal
func quoteChar(c int) string {
	// special cases - different from quoted strings
	if c == '\'' {
		return `'\''`
	}
	if c == '"' {
		return `'"'`
	}

	// use quoted string with different quotation marks
	s := strconv.Quote(string(c))
	return "'" + s[1:len(s)-1] + "'"
}

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	r, err := strconv.ParseUint(string(s[2:6]), 16, 64)
	if err != nil {
		return -1
	}
	return rune(r)
}

func stringBytes(s []byte) (t []byte, ok bool) {
	// Check for unusual characters. If there are none,
	// then no unquoting is needed, so return a slice of the
	// original bytes.
	r := 0
	for r < len(s) {
		c := s[r]
		if c == '\\' || c == '"' || c < ' ' {
			break
		}
		if c < utf8.RuneSelf {
			r++
			continue
		}
		rr, size := utf8.DecodeRune(s[r:])
		if rr == utf8.RuneError && size == 1 {
			break
		}
		r += size
	}
	if r == len(s) {
		return s, true
	}

	b := make([]byte, len(s)+2*utf8.UTFMax)
	w := copy(b, s[0:r])
	for r < len(s) {
		// Out of room?  Can only happen if s is full of
		// malformed UTF-8 and we're replacing each
		// byte with RuneError.
		if w >= len(b)-2*utf8.UTFMax {
			nb := make([]byte, (len(b)+utf8.UTFMax)*2)
			copy(nb, b[0:w])
			b = nb
		}
		switch c := s[r]; {
		case c == '\\':
			r++
			if r >= len(s) {
				return
			}
			switch s[r] {
			default:
				return
			case '"', '\\', '/', '\'':
				b[w] = s[r]
				r++
				w++
			case 'b':
				b[w] = '\b'
				r++
				w++
			case 'f':
				b[w] = '\f'
				r++
				w++
			case 'n':
				b[w] = '\n'
				r++
				w++
			case 'r':
				b[w] = '\r'
				r++
				w++
			case 't':
				b[w] = '\t'
				r++
				w++
			case 'u':
				r--
				rr := getu4(s[r:])
				if rr < 0 {
					return
				}
				r += 6
				if utf16.IsSurrogate(rr) {
					rr1 := getu4(s[r:])
					if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
						// A valid pair; consume.
						r += 6
						w += utf8.EncodeRune(b[w:], dec)
						break
					}
					// Invalid surrogate; fall back to replacement rune.
					rr = unicode.ReplacementChar
				}
				w += utf8.EncodeRune(b[w:], rr)
			}

		// Quote, control characters are invalid.
		case c == '"', c < ' ':
			return

		// ASCII
		case c < utf8.RuneSelf:
			b[w] = c
			r++
			w++

		// Coerce to well-formed UTF-8.
		default:
			rr, size := utf8.DecodeRune(s[r:])
			r += size
			w += utf8.EncodeRune(b[w:], rr)
		}
	}
	return b[0:w], true
}

type keywordScanner struct {
	maxWordLen   int
	keywords     []string
	keywordsType []int
}

var keyKeyword = keywordScanner{
	0,
	[]string{"include"},
	[]int{bufTypeInclude},
}

var valueKeyword = keywordScanner{
	0,
	[]string{"true", "false", "null", "include"},
	[]int{bufTypeBoolTrue, bufTypeBoolFalse, bufTypeNull, bufTypeInclude},
}

func (k *keywordScanner) init() {
	for i := range k.keywords {
		wordLen := len(k.keywords[i])
		if wordLen > k.maxWordLen {
			k.maxWordLen = wordLen
		}
	}
}

func (k *keywordScanner) checkKeyword(s *fileScanner, buf []byte) {
	if len(buf) <= k.maxWordLen {
		value := string(buf)
		for i := range k.keywords {
			if value == k.keywords[i] {
				s.bufType = k.keywordsType[i]
				return
			}
		}
	}
}

func init() {
	keyKeyword.init()
	valueKeyword.init()
}
