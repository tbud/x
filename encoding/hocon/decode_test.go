package hocon

import (
	"fmt"
	"github.com/tbud/bud/encoding/json"
	"io/ioutil"
	"testing"
)

func getJsonMap(file string) (m interface{}, err error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	err = json.Unmarshal(buf, &m)
	return
}

func TestDecode(t *testing.T) {
	r, err := getJsonMap("testdata/decode.json")
	if err != nil {
		t.Fatalf("Get json from file %s error:%s", "testdata/decode.json", err)
	}

	m := r.(map[string]interface{})

	v, ok := m["hocon"]
	if ok {
		fmt.Printf("%v\n", v)
		m := v.(map[string]interface{})
		if v, ok := m["user"]; ok {
			fmt.Printf("%v\n", v)
			m := v.(map[string]interface{})
			if v, ok := m["name"]; ok {
				fmt.Printf("%v\n", v)
			}
		}
	}
}
