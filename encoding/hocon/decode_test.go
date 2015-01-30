package hocon

import (
	// "fmt"
	"github.com/tbud/x/encoding/json"
	"io/ioutil"
	"reflect"
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

func getHoconMap(file string) (m interface{}, err error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	err = Unmarshal(buf, &m)
	return
}

func compareHoconAndJson(t *testing.T, hoconFile, jsonFile string) {
	j, err := getJsonMap(jsonFile)
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	h, err := getHoconMap(hoconFile)
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	if !reflect.DeepEqual(h, j) {
		t.Errorf("\n got %v\nwant %v", h, j)
	}
}

func TestHoconSupportJson(t *testing.T) {
	compareHoconAndJson(t, "testdata/decode.json", "testdata/decode.json")
}

func TestDecode(t *testing.T) {
	compareHoconAndJson(t, "testdata/decode_comments.hocon", "testdata/decode.json")
}
