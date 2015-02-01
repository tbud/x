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
	"io/ioutil"
	"path/filepath"
	"strconv"
	"unicode"
)

// nextValue splits data after the next whole HOCON value,
// returning that value and the bytes that follow it as separate slices.
// scan is passed in for use by nextValue to avoid an allocation.
func nextValue(data []byte, scan *fileScanner) (value, rest []byte, err error) {
	scan.reset()
	for i, c := range data {
		v := scan.step(scan, int(c))
		if v >= scanEnd {
			switch v {
			case scanError:
				return nil, nil, scan.err
			case scanEnd:
				return data[0:i], data[i:], nil
			}
		}
	}
	if scan.eof() == scanError {
		return nil, nil, scan.err
	}
	return data, nil, nil
}

// A SyntaxError is a description of a HOCON syntax error.
type SyntaxError struct {
	msg    string // description of error
	Offset int64  // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { return e.msg + " Offset " + strconv.FormatInt(e.Offset, 10) }

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

	// Reached end of top-level value.
	endTop bool

	// Stack of what we're in the middle of - array values, object keys, object values.
	parseState []int

	// Error that happened, if any.
	err error

	// 1-byte redo (see undo method)
	redo      bool
	redoCode  int
	redoState func(*fileScanner, int) int

	// total bytes consumed, updated by decoder.Decode
	bytes int64

	file            string         // file name
	dir             string         // dir of the file
	data            []byte         // store data load from file
	includesScanner []*fileScanner // store scanner include form current file
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
	scanContinue     = iota // uninteresting byte
	scanBeginLiteral        // end implied by next result != scanContinue
	scanBeginObject         // begin object
	scanObjectKey           // just finished object key (string)
	scanObjectValue         // just finished non-last object value
	scanEndObject           // end object (implies scanObjectValue if possible)
	scanBeginArray          // begin array
	scanArrayValue          // just finished array value
	scanEndArray            // end array (implies scanArrayValue if possible)
	scanSkipSpace           // space byte; can skip; known to be last "continue" result

	// Stop.
	scanEnd   // top-level value ended *before* this byte; known to be first "stop" result
	scanError // hit an error, scanner.err.
)

// These values are stored in the parseState stack.
// They give the current state of a composite value
// being scanned.  If the parser is inside a nested value
// the parseState describes the nested state, outermost at entry 0.
const (
	parseObjectKey   = iota // parsing object key (before colon)
	parseObjectValue        // parsing object value (after colon)
	parseArrayValue         // parsing array value
)

// reset prepares the scanner for use.
// It must be called before calling s.step.
func (s *fileScanner) reset() {
	s.step = stateBegin
	s.parseState = s.parseState[0:0]
	s.err = nil
	s.redo = false
	s.endTop = false
	s.bytes = 0
}

// checkValid verifies that data is valid HOCON-encoded data.
// scan is passed in for use by checkValid to avoid an allocation.
func (s *fileScanner) checkValid(fileName string) error {
	s.reset()

	if !filepath.IsAbs(fileName) {
		return &SyntaxError{"file '" + fileName + "' is not absolute path", s.bytes}
	}
	var err error
	s.data, err = ioutil.ReadFile(fileName)
	if err != nil {
		return &SyntaxError{err.Error(), s.bytes}
	}

	s.file = filepath.Base(fileName)
	s.dir = filepath.Dir(fileName)

	for _, c := range s.data {
		s.bytes++
		if s.step(s, int(c)) == scanError {
			return s.err
		}
	}
	if s.eof() == scanError {
		return s.err
	}

	for i := range s.includesScanner {
		scan := s.includesScanner[i]
		var checkFile string
		if filepath.IsAbs(scan.file) {
			checkFile = scan.file
		} else {
			checkFile = filepath.Join(s.dir, scan.file)
		}
		err := scan.checkValid(checkFile)
		if err != nil {
			return err
		}
	}
	return nil
}

// eof tells the scanner that the end of input has been reached.
// It returns a scan status just as s.step does.
func (s *fileScanner) eof() int {
	if s.err != nil {
		return scanError
	}
	if s.endTop {
		return scanEnd
	}
	s.step(s, ' ')
	if s.endTop {
		return scanEnd
	}
	if s.err == nil {
		s.err = &SyntaxError{"unexpected end of HOCON input", s.bytes}
	}
	return scanError
}

// pushParseState pushes a new parse state p onto the parse stack.
func (s *fileScanner) pushParseState(p int) {
	s.parseState = append(s.parseState, p)
}

// popParseState pops a parse state (already obtained) off the stack
// and updates s.step accordingly.
func (s *fileScanner) popParseState() {
	n := len(s.parseState) - 1
	s.parseState = s.parseState[0:n]
	s.redo = false
	if n == 0 {
		s.step = stateEndTop
		s.endTop = true
	} else {
		s.step = stateEndValue
	}
}

func isSpace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}

// stateBeginValueOrEmpty is the state after reading `[`.
func stateBeginValueOrEmpty(s *fileScanner, c int) int {
	if c <= ' ' && isSpace(rune(c)) {
		return scanSkipSpace
	}
	if c == ']' {
		return stateEndValue(s, c)
	}
	return stateBeginValue(s, c)
}

func stateBegin(s *fileScanner, c int) int {
	if c <= ' ' && isSpace(rune(c)) {
		return scanSkipSpace
	}
	switch c {
	case '{':
		s.step = stateBeginStringOrEmpty
		s.pushParseState(parseObjectKey)
		return scanBeginObject
	case '[':
		s.step = stateBeginValueOrEmpty
		s.pushParseState(parseArrayValue)
		return scanBeginArray
	case '"':
		s.step = stateInString
		return scanBeginLiteral
	case '#':
		s.step = stateComment
		return scanBeginLiteral
	case '/':
		s.step = stateSlashComment
		return scanBeginLiteral
	case 'i':
		s.step = stateI
	}
	if unicode.IsLetter(rune(c)) {
		s.step = stateNoQuoteString
		s.pushParseState(parseObjectKey)
		return scanBeginLiteral
	}
	return s.error(c, "looking for beginning")
}

// stateBeginValue is the state at the beginning of the input.
func stateBeginValue(s *fileScanner, c int) int {
	if c <= ' ' && isSpace(rune(c)) {
		return scanSkipSpace
	}
	switch c {
	case '{':
		s.step = stateBeginStringOrEmpty
		s.pushParseState(parseObjectKey)
		return scanBeginObject
	case '[':
		s.step = stateBeginValueOrEmpty
		s.pushParseState(parseArrayValue)
		return scanBeginArray
	case '"':
		s.step = stateInString
		return scanBeginLiteral
	case '-':
		s.step = stateNeg
		return scanBeginLiteral
	case '0': // beginning of 0.123
		s.step = state0
		return scanBeginLiteral
	case 't': // beginning of true
		s.step = stateT
		return scanBeginLiteral
	case 'f': // beginning of false
		s.step = stateF
		return scanBeginLiteral
	case 'n': // beginning of null
		s.step = stateN
		return scanBeginLiteral
	case '#':
		s.step = stateComment
		return scanBeginLiteral
	case '/':
		s.step = stateSlashComment
		return scanBeginLiteral
	}
	if '1' <= c && c <= '9' { // beginning of 1234.5
		s.step = state1
		return scanBeginLiteral
	}
	return s.error(c, "looking for beginning of value")
}

// stateBeginStringOrEmpty is the state after reading `{`.
func stateBeginStringOrEmpty(s *fileScanner, c int) int {
	if c <= ' ' && isSpace(rune(c)) {
		return scanSkipSpace
	}
	if c == '}' {
		n := len(s.parseState)
		s.parseState[n-1] = parseObjectValue
		return stateEndValue(s, c)
	}
	return stateBeginString(s, c)
}

// stateBeginString is the state after reading `{"key": value,`.
func stateBeginString(s *fileScanner, c int) int {
	if c <= ' ' && isSpace(rune(c)) {
		return scanSkipSpace
	}
	if c == '"' {
		s.step = stateInString
		return scanBeginLiteral
	}
	return s.error(c, "looking for beginning of object key string")
}

// stateEndValue is the state after completing a value,
// such as after reading `{}` or `true` or `["x"`.
func stateEndValue(s *fileScanner, c int) int {
	n := len(s.parseState)
	if n == 0 {
		// Completed top-level before the current byte.
		s.step = stateEndTop
		s.endTop = true
		return stateEndTop(s, c)
	}
	if c <= ' ' && isSpace(rune(c)) {
		s.step = stateEndValue
		return scanSkipSpace
	}
	ps := s.parseState[n-1]
	switch ps {
	case parseObjectKey:
		switch c {
		case ':':
			s.parseState[n-1] = parseObjectValue
			s.step = stateBeginValue
			return scanObjectKey
		case '.':
			s.parseState[n-1] = parseObjectValue
			s.step = stateBegin
			return scanObjectKey
		}
		return s.error(c, "after object key")
	case parseObjectValue:
		if c == ',' {
			s.parseState[n-1] = parseObjectKey
			s.step = stateBeginString
			return scanObjectValue
		}
		if c == '}' {
			s.popParseState()
			return scanEndObject
		}
		return s.error(c, "after object key:value pair")
	case parseArrayValue:
		if c == ',' {
			s.step = stateBeginValue
			return scanArrayValue
		}
		if c == ']' {
			s.popParseState()
			return scanEndArray
		}
		return s.error(c, "after array element")
	}
	return s.error(c, "")
}

// stateEndTop is the state after finishing the top-level value,
// such as after reading `{}` or `[1,2,3]`.
// Only space characters should be seen now.
func stateEndTop(s *fileScanner, c int) int {
	if c != ' ' && c != '\t' && c != '\r' && c != '\n' {
		// Complain about non-space byte on next call.
		s.error(c, "after top-level value")
	}
	return scanEnd
}

// stateInString is the state after reading `"`.
func stateInString(s *fileScanner, c int) int {
	switch {
	case c == '"':
		s.step = stateEndValue
		return scanContinue
	case c == '\\':
		s.step = stateInStringEsc
		return scanContinue
	case c < 0x20:
		return s.error(c, "in string literal")
	}

	return scanContinue
}

func stateNoQuoteString(s *fileScanner, c int) int {
	switch c {
	case '.':
		s.pushParseState(parseObjectKey)
		return scanBeginObject
	case '=':
		s.step = stateEndValue
		return stateEndValue(s, c)
	case ':':
		s.step = stateEndValue
		return stateEndValue(s, c)
	}
	if c < 0x20 {
		return s.error(c, "in no quote string literal")
	}
	return scanContinue
}

func stateComment(s *fileScanner, c int) int {
	if c == '\n' || c == '\r' {
		s.step = stateBeginValue
		return scanContinue
	}
	return scanContinue
}

func stateSlashComment(s *fileScanner, c int) int {
	if c == '/' {
		s.step = stateComment
		return scanContinue
	}
	s.err = &SyntaxError{"invalid comment char " + quoteChar(c), s.bytes}
	return scanError
}

// stateInStringEsc is the state after reading `"\` during a quoted string.
func stateInStringEsc(s *fileScanner, c int) int {
	switch c {
	case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
		s.step = stateInString
		return scanContinue
	}
	if c == 'u' {
		s.step = stateInStringEscU
		return scanContinue
	}
	return s.error(c, "in string escape code")
}

// stateInStringEscU is the state after reading `"\u` during a quoted string.
func stateInStringEscU(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInStringEscU1
		return scanContinue
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU1 is the state after reading `"\u1` during a quoted string.
func stateInStringEscU1(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInStringEscU12
		return scanContinue
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU12 is the state after reading `"\u12` during a quoted string.
func stateInStringEscU12(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInStringEscU123
		return scanContinue
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateInStringEscU123 is the state after reading `"\u123` during a quoted string.
func stateInStringEscU123(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
		s.step = stateInString
		return scanContinue
	}
	// numbers
	return s.error(c, "in \\u hexadecimal character escape")
}

// stateNeg is the state after reading `-` during a number.
func stateNeg(s *fileScanner, c int) int {
	if c == '0' {
		s.step = state0
		return scanContinue
	}
	if '1' <= c && c <= '9' {
		s.step = state1
		return scanContinue
	}
	return s.error(c, "in numeric literal")
}

// state1 is the state after reading a non-zero integer during a number,
// such as after reading `1` or `100` but not `0`.
func state1(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' {
		s.step = state1
		return scanContinue
	}
	return state0(s, c)
}

// state0 is the state after reading `0` during a number.
func state0(s *fileScanner, c int) int {
	if c == '.' {
		s.step = stateDot
		return scanContinue
	}
	if c == 'e' || c == 'E' {
		s.step = stateE
		return scanContinue
	}
	return stateEndValue(s, c)
}

// stateDot is the state after reading the integer and decimal point in a number,
// such as after reading `1.`.
func stateDot(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' {
		s.step = stateDot0
		return scanContinue
	}
	return s.error(c, "after decimal point in numeric literal")
}

// stateDot0 is the state after reading the integer, decimal point, and subsequent
// digits of a number, such as after reading `3.14`.
func stateDot0(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' {
		s.step = stateDot0
		return scanContinue
	}
	if c == 'e' || c == 'E' {
		s.step = stateE
		return scanContinue
	}
	return stateEndValue(s, c)
}

// stateE is the state after reading the mantissa and e in a number,
// such as after reading `314e` or `0.314e`.
func stateE(s *fileScanner, c int) int {
	if c == '+' {
		s.step = stateESign
		return scanContinue
	}
	if c == '-' {
		s.step = stateESign
		return scanContinue
	}
	return stateESign(s, c)
}

// stateESign is the state after reading the mantissa, e, and sign in a number,
// such as after reading `314e-` or `0.314e+`.
func stateESign(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' {
		s.step = stateE0
		return scanContinue
	}
	return s.error(c, "in exponent of numeric literal")
}

// stateE0 is the state after reading the mantissa, e, optional sign,
// and at least one digit of the exponent in a number,
// such as after reading `314e-2` or `0.314e+1` or `3.14e0`.
func stateE0(s *fileScanner, c int) int {
	if '0' <= c && c <= '9' {
		s.step = stateE0
		return scanContinue
	}
	return stateEndValue(s, c)
}

func stateI(s *fileScanner, c int) int {
	if c == 'n' {
		s.step = stateIn
		return scanContinue
	}
	return s.error(c, "in literal include (expecting 'n')")
}

func stateIn(s *fileScanner, c int) int {
	if c == 'c' {
		s.step = stateInc
		return scanContinue
	}
	return s.error(c, "in literal include (expecting 'c')")
}

func stateInc(s *fileScanner, c int) int {
	if c == 'l' {
		s.step = stateIncl
		return scanContinue
	}
	return s.error(c, "in literal include (expecting 'l')")
}

func stateIncl(s *fileScanner, c int) int {
	if c == 'u' {
		s.step = stateInclu
		return scanContinue
	}
	return s.error(c, "in literal include (expecting 'u')")
}

func stateInclu(s *fileScanner, c int) int {
	if c == 'd' {
		s.step = stateInclud
		return scanContinue
	}
	return s.error(c, "in literal include (expecting 'd')")
}

func stateInclud(s *fileScanner, c int) int {
	if c == 'e' {
		s.step = stateEndValue
		return scanContinue
	}
	return s.error(c, "in literal include (expecting 'e')")
}

// stateT is the state after reading `t`.
func stateT(s *fileScanner, c int) int {
	if c == 'r' {
		s.step = stateTr
		return scanContinue
	}
	return s.error(c, "in literal true (expecting 'r')")
}

// stateTr is the state after reading `tr`.
func stateTr(s *fileScanner, c int) int {
	if c == 'u' {
		s.step = stateTru
		return scanContinue
	}
	return s.error(c, "in literal true (expecting 'u')")
}

// stateTru is the state after reading `tru`.
func stateTru(s *fileScanner, c int) int {
	if c == 'e' {
		s.step = stateEndValue
		return scanContinue
	}
	return s.error(c, "in literal true (expecting 'e')")
}

// stateF is the state after reading `f`.
func stateF(s *fileScanner, c int) int {
	if c == 'a' {
		s.step = stateFa
		return scanContinue
	}
	return s.error(c, "in literal false (expecting 'a')")
}

// stateFa is the state after reading `fa`.
func stateFa(s *fileScanner, c int) int {
	if c == 'l' {
		s.step = stateFal
		return scanContinue
	}
	return s.error(c, "in literal false (expecting 'l')")
}

// stateFal is the state after reading `fal`.
func stateFal(s *fileScanner, c int) int {
	if c == 's' {
		s.step = stateFals
		return scanContinue
	}
	return s.error(c, "in literal false (expecting 's')")
}

// stateFals is the state after reading `fals`.
func stateFals(s *fileScanner, c int) int {
	if c == 'e' {
		s.step = stateEndValue
		return scanContinue
	}
	return s.error(c, "in literal false (expecting 'e')")
}

// stateN is the state after reading `n`.
func stateN(s *fileScanner, c int) int {
	if c == 'u' {
		s.step = stateNu
		return scanContinue
	}
	return s.error(c, "in literal null (expecting 'u')")
}

// stateNu is the state after reading `nu`.
func stateNu(s *fileScanner, c int) int {
	if c == 'l' {
		s.step = stateNul
		return scanContinue
	}
	return s.error(c, "in literal null (expecting 'l')")
}

// stateNul is the state after reading `nul`.
func stateNul(s *fileScanner, c int) int {
	if c == 'l' {
		s.step = stateEndValue
		return scanContinue
	}
	return s.error(c, "in literal null (expecting 'l')")
}

// stateError is the state after reaching a syntax error,
// such as after reading `[1}` or `5.1.2`.
func stateError(s *fileScanner, c int) int {
	return scanError
}

// error records an error and switches to the error state.
func (s *fileScanner) error(c int, context string) int {
	s.step = stateError
	s.err = &SyntaxError{"invalid character " + quoteChar(c) + " " + context, s.bytes}
	return scanError
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

// undo causes the scanner to return scanCode from the next state transition.
// This gives callers a simple 1-byte undo mechanism.
func (s *fileScanner) undo(scanCode int) {
	if s.redo {
		panic("hocon: invalid use of scanner")
	}
	s.redoCode = scanCode
	s.redoState = s.step
	s.step = stateRedo
	s.redo = true
}

// stateRedo helps implement the scanner's 1-byte undo.
func stateRedo(s *fileScanner, c int) int {
	s.redo = false
	s.step = s.redoState
	return s.redoCode
}
