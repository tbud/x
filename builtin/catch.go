package builtin

import (
	"runtime"
)

func Catch(fun func(interface{})) {
	if r := recover(); r != nil {
		if _, ok := r.(runtime.Error); ok {
			panic(r)
		}
		if fun != nil {
			fun(r)
		}
	}
}
