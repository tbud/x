package appender

import (
	"github.com/tbud/x/config"
	"github.com/tbud/x/log/common"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"
)

var (
	confInited     config.Config
	appenderInited Appender
	msgInited      = common.LogMsg{Msg: "hello py", Date: time.Now()}
)

func init() {
	conf, err := config.Load("console.conf")
	if err != nil {
		return
	}

	confInited = conf

	appender, err := New(confInited)
	if err != nil {
		return
	}

	appenderInited = appender
}

func BenchmarkConsoleInited(b *testing.B) {
	for i := 0; i < b.N; i++ {
		appenderInited.Append(&msgInited)
	}
}

func BenchmarkConsole(b *testing.B) {
	conf, err := config.Load("console.conf")
	if err != nil {
		b.Error(err)
	}

	appender, err := consoleAppender(conf)
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		m := common.LogMsg{Msg: "hello py", Date: time.Now()}
		appender.Append(&m)
	}
}

func BenchmarkLog(b *testing.B) {
	trace := log.New(ioutil.Discard, "TRACE1 ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
	for i := 0; i < b.N; i++ {
		trace.Println("hello py")
	}
}

func TestLog(t *testing.T) {
	trace := log.New(os.Stdout, "TRACE ", log.Ldate|log.Ltime|log.Lmicroseconds)
	trace.Println("hello py")
}

func BenchmarkNewTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := common.LogMsg{Msg: "hello py", Date: time.Now()}
		ioutil.Discard.Write([]byte(m.Msg))
	}
}

func BenchmarkTimeFormat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		time.Now().Format("2006-01-02 15:04:05")
	}
}

func TestConsole(t *testing.T) {
	appender, err := New(nil)
	if err != nil {
		t.Error(err)
	}

	m := common.LogMsg{Msg: "hello py: test console", Date: time.Now()}
	appender.Append(&m)
}
