package config_test

import (
	"fmt"
	"github.com/tbud/x/config"
)

func Example() {
	conf, err := config.Load("testdata/test2.conf")
	if err != nil {
		return
	}

	fmt.Println(conf.StringDefault("test2.comment", ""))

	if name, found := conf.String("test2.user.name"); found {
		fmt.Println(name)
	}

	fmt.Println(conf.BoolsDefault("test2.user.testbool.list", []bool{}))

	// Output:
	// #
	// 彭毅
	// [true false]
}
