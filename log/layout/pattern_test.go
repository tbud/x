package layout

import (
	"github.com/tbud/x/config"
	"github.com/tbud/x/log/common"
	"testing"
	"time"
)

func TestPatternFormat(t *testing.T) {
	conf, err := config.Load("pattern.conf")
	if err != nil {
		t.Error(err)
	}

	pattern, err := patternLayout(conf)
	if err != nil {
		t.Error(err)
	}

	buf := []byte{}
	pattern.Format(&buf, &common.LogMsg{Date: time.Now(), Msg: "py test 123"})

	t.Error(string(buf))
}
