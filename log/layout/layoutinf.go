package layout

import (
	"errors"
	"github.com/tbud/x/config"
	"github.com/tbud/x/log/common"
)

type Layout interface {
	Format(buf *[]byte, m *common.LogMsg) error
	NeedFile() bool
	NeedTime() bool
}

type LayoutMaker func(conf config.Config) (Layout, error)

var layoutMakers = make(map[string]LayoutMaker)

// Register makes a log layout maker available by the layout name.
// If Register is called twice with the same name or if layout maker is nil,
// it panics.
func Register(name string, layoutMaker LayoutMaker) {
	if layoutMaker == nil {
		panic("log: Register layout maker is nil")
	}
	if _, dup := layoutMakers[name]; dup {
		panic("log: Register called twice for layout maker " + name)
	}
	layoutMakers[name] = layoutMaker
}

func New(conf config.Config) (layout Layout, err error) {
	name := conf.StringDefault("type", "Pattern")
	if layoutMaker, ok := layoutMakers[name]; ok {
		layout, err = layoutMaker(conf)
		return
	}

	return nil, errors.New("Layout maker not exist.: " + name)
}
