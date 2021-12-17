package main

import (
	"net/http"
	"testing"
	"yabl/lib"
)

var testserver *lib.WsServer

func init() {
	testserver = lib.NewWsServer("127.0.0.1:12580", "tcp")
	go testserver.Start()
}

func Test_Httphandle(t *testing.T) {
	resp, err := http.Get("http://127.0.0.1:12580/ws")
	if err != nil {
		t.FailNow()
	}
	resp.Body.Close()
}

func Benchmark_Httphandle(b *testing.B) {
	for i := 0; i < b.N; i++ { //use b.N for looping
		resp, err := http.Get("http://127.0.0.1:12580/ws")
		if err != nil {
			b.FailNow()
		}
		resp.Body.Close()
	}
}
