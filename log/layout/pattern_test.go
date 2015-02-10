package layout

import (
	"github.com/tbud/x/log/common"
	"testing"
	"time"
)

func TestPatternFormat(t *testing.T) {
	pattern, err := patternLayout(nil)
	if err != nil {
		t.Error(err)
	}

	buf := []byte{}
	pattern.Format(&buf, &common.LogMsg{Date: time.Now(), Msg: "py test 123"})

	t.Log(string(buf))
}
