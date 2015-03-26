package linkedmap_test

import (
	"fmt"
	"github.com/tbud/x/container/linkedmap"
)

func Example() {
	m := linkedmap.New()

	m.Append("ppp", "123")
	m.Append("uya", "444")
	m.Append("ttt", "999")
	m.Append("abc", "001")
	m.Append("ttt", "333")

	m.Remove("uya")

	m.Each(func(key, value interface{}) error {
		fmt.Printf("%s:%s\n", key, value)
		return nil
	})

	// Output:
	// ppp:123
	// ttt:333
	// abc:001
}
