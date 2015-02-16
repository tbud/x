package layout

import (
	"errors"
	"github.com/tbud/x/config"
	. "github.com/tbud/x/log/common"
)

type PatternLayout struct {
	year, month, day int
	hour, min, sec   int
	nanoSec          int
	pattern          []byte
	segments         []patternSegment
	err              error
	step             func(*PatternLayout, int) int
	needFile         bool
	needTime         bool
	tempBuf          []byte
}

const (
	patternYear = iota
	patternMonth
	patternDay
	patternHour
	patternMin
	patternSec
	patternNanoSec

	patternLongFile
	patternShortFile
	patternLine

	patternLevel
	patternShortLevel

	patternMsg

	patternString
)

const (
	scanContinue = iota
	scanError
)

type patternSegment struct {
	patternType int
	segLen      int // save year, nanoSec len
	seg         []byte
}

func stateString(p *PatternLayout, c int) int {
	if c == '%' {
		p.step = stateKeyword
	} else {
		sLen := len(p.segments)
		currentSegment := &p.segments[sLen-1]
		if currentSegment.patternType == patternString {
			currentSegment.seg = append(currentSegment.seg, byte(c))
		} else {
			p.segments = append(p.segments, patternSegment{patternType: patternString, seg: []byte{byte(c)}})
		}
	}

	return scanContinue
}

var keyCharMap = map[int]int{
	'L': patternLevel,
	'l': patternShortLevel,
	'F': patternLongFile,
	'f': patternShortFile,
	'm': patternMsg,
	'n': patternLine,
}

func stateKeyword(p *PatternLayout, c int) int {
	if c == 'd' {
		p.step = stateDate
		return scanContinue
	}
	if v, ok := keyCharMap[c]; ok {
		p.segments = append(p.segments, patternSegment{patternType: v})
	} else {
		p.err = errors.New("keyword '" + string(c) + "' not support.")
		return scanError
	}

	p.step = stateString
	return scanContinue
}

var dateKeyCharMap = map[int]int{
	'y': patternYear,
	'M': patternMonth,
	'd': patternDay,
	'H': patternHour,
	'm': patternMin,
	's': patternSec,
	'S': patternNanoSec,
}

func stateDate(p *PatternLayout, c int) int {
	switch c {
	case '{':
		return scanContinue
	case '}':
		p.step = stateString
		return scanContinue
	}

	sLen := len(p.segments)
	currentSegment := &p.segments[sLen-1]
	if v, ok := dateKeyCharMap[c]; ok {
		if currentSegment.patternType == v {
			currentSegment.segLen += 1
		} else {
			p.segments = append(p.segments, patternSegment{patternType: v, segLen: 1})
		}
	} else {
		if currentSegment.patternType == patternString {
			currentSegment.seg = append(currentSegment.seg, byte(c))
		} else {
			p.segments = append(p.segments, patternSegment{patternType: patternString, seg: []byte{byte(c)}})
		}
	}
	return scanContinue
}

func (p *PatternLayout) parse() error {
	p.segments = append(p.segments, patternSegment{patternType: patternString})
	p.step = stateString
	for _, c := range p.pattern {
		if p.step(p, int(c)) == scanError {
			return p.err
		}
	}
	p.segments = append(p.segments, patternSegment{patternType: patternString, seg: []byte("\n")})

	for _, segment := range p.segments {
		switch segment.patternType {
		case patternYear:
			if !(segment.segLen == 2 || segment.segLen == 4) {
				return errors.New("year must 2 or 4 len.")
			}
			p.needTime = true
		case patternNanoSec:
			if segment.segLen < 3 || segment.segLen > 9 {
				return errors.New("nano second must 3 to 9 len.")
			}
			p.needTime = true
		case patternMonth, patternDay, patternHour,
			patternMin, patternSec:
			if segment.segLen != 2 {
				return errors.New("month, day, hour, min, sec must 2 len.")
			}
			p.needTime = true
		case patternLongFile, patternShortFile, patternLine:
			p.needFile = true
		}
	}
	return nil
}

func (p *PatternLayout) Format(buf *[]byte, m *LogMsg) error {
	if p.needTime {
		year, month, day := m.Date.Date()
		p.year = year
		p.month = int(month)
		p.day = day

		hour, min, sec := m.Date.Clock()
		p.hour = hour
		p.min = min
		p.sec = sec

		p.nanoSec = m.Date.Nanosecond()
	}

	for _, segment := range p.segments {
		switch segment.patternType {
		case patternYear:
			p.tempBuf = p.tempBuf[:0]
			itoa(&p.tempBuf, p.year, 4)
			*buf = append(*buf, p.tempBuf[4-segment.segLen:]...)
		case patternMonth:
			itoa(buf, p.month, 2)
		case patternDay:
			itoa(buf, p.day, 2)
		case patternHour:
			itoa(buf, p.hour, 2)
		case patternMin:
			itoa(buf, p.min, 2)
		case patternSec:
			itoa(buf, p.sec, 2)
		case patternNanoSec:
			p.tempBuf = p.tempBuf[:0]
			itoa(&p.tempBuf, p.nanoSec, 9)
			*buf = append(*buf, p.tempBuf[0:segment.segLen]...)
		case patternLongFile:
			*buf = append(*buf, m.File...)
		case patternShortFile:
			file := m.File
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			*buf = append(*buf, short...)
		case patternLine:
			itoa(buf, m.Line, -1)
		case patternLevel:
			*buf = append(*buf, LogLevelToString(m.Level)...)
		case patternShortLevel:
			*buf = append(*buf, LogLevelToShortString(m.Level)...)
		case patternMsg:
			*buf = append(*buf, m.Msg...)
		case patternString:
			*buf = append(*buf, segment.seg...)
		}
	}

	return nil
}

func (p *PatternLayout) NeedFile() bool {
	return p.needFile
}

func (p *PatternLayout) NeedTime() bool {
	return p.needTime
}

// Cheap integer to fixed-width decimal ASCII.  Give a negative width to avoid zero-padding.
// Knows the buffer has capacity.
func itoa(buf *[]byte, i int, wid int) {
	var u uint = uint(i)
	if u == 0 && wid <= 1 {
		*buf = append(*buf, '0')
		return
	}

	// Assemble decimal in reverse order.
	var b [32]byte
	bp := len(b)
	for ; u > 0 || wid > 0; u /= 10 {
		bp--
		wid--
		b[bp] = byte(u%10) + '0'
	}
	*buf = append(*buf, b[bp:]...)
}

func patternLayout(conf config.Config) (lay Layout, err error) {
	layout := &PatternLayout{}
	layout.pattern = []byte(conf.StringDefault("pattern", "[%l]%m"))
	err = layout.parse()
	return layout, err
}

func init() {
	Register("Pattern", patternLayout)
}
