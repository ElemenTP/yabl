package main

import (
	"net/http"
	"testing"
	"time"
	"yabl/lib"

	"github.com/gorilla/websocket"
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

func Test_Websockethandle(t *testing.T) {
	url := "ws://127.0.0.1:12580/ws"
	scriptstr := `
func main:
- postmsg "你好"`
	genScript([]byte(scriptstr))
	lib.Compile()
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Logf(err.Error())
		t.FailNow()
	}
	time.Now()
	recvmsg := new(lib.MsgStruct)
	err = conn.ReadJSON(&recvmsg)
	if err != nil {
		t.Logf(err.Error())
		t.FailNow()
	}
	t.Logf(time.Unix(recvmsg.Timestamp, 0).Format("2006-01-02 15:04:05"))
	t.Logf(recvmsg.Content)
	conn.Close()
}
