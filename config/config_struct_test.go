package config

import (
	"reflect"
	"strings"
	"testing"
)

type testSetStruct struct {
	AppName string

	HttpPort    int
	HttpAddr    string
	HttpSsl     bool
	HttpSslCert string
	HttpSslKey  string

	Seeds    []string
	FloatNum float64
}

func TestSetStruct(t *testing.T) {
	test := testSetStruct{
		AppName:  "sample",
		HttpAddr: "localhost",
	}

	conf := Config{
		"appName":    "sample",
		"HttpAddr":   "127.0.0.1",
		"httpSslKey": "test",
		"httpSsl":    true,
		"httpPort":   9000,
		"seeds":      []string{"1", "2", "3"},
		"floatNum":   1.18,
	}

	if err := conf.SetStruct(&test); err != nil {
		t.Error(err)
	}

	if test.AppName != "sample" {
		t.Error("appname must be sample")
	}

	if test.HttpAddr != "127.0.0.1" {
		t.Error("http addr must be 127.0.0.1")
	}

	if test.HttpSslKey != "test" {
		t.Error("http ssl key must be test")
	}

	if len(test.HttpSslCert) != 0 {
		t.Error("http ssl cert must be empty")
	}

	if test.HttpSsl != true {
		t.Error("http ssl must be true")
	}

	if test.HttpPort != 9000 {
		t.Error("http port must be 9000")
	}

	if !reflect.DeepEqual(test.Seeds, []string{"1", "2", "3"}) {
		t.Error("seeds set config err")
	}

	if test.FloatNum != 1.18 {
		t.Error("float num must be 1.18.")
	}

	if err := conf.SetStruct(test); err != nil {
		if !strings.Contains(err.Error(), "Struct must be a point.") {
			t.Error(err)
		}
	}
}
