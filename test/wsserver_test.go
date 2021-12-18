package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"
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
		fmt.Println(err)
		t.FailNow()
	}
	defer func() {
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		<-time.After(time.Second)
		fmt.Println("Disconnect ", conn.RemoteAddr())
	}()

	recvmsg := new(lib.MsgStruct)
	err = conn.ReadJSON(&recvmsg)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	fmt.Println(time.Unix(recvmsg.Timestamp, 0).Format("2006-01-02 15:04:05"), recvmsg.Content)
}

func Test_Websockethandles(t *testing.T) {
	url := "ws://127.0.0.1:12580/ws"
	scriptstr := `
func main:
- postmsg "你好"`
	genScript([]byte(scriptstr))
	lib.Compile()

	inputstr := ``
	strslice := strings.Split(inputstr, "\n")

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	defer func() {
		recover()
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		<-time.After(time.Second)
		fmt.Println("Disconnect ", conn.RemoteAddr())
	}()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(600))) //close connection after 600s
			recvmsg := new(lib.MsgStruct)
			err := conn.ReadJSON(&recvmsg)
			if err != nil {
				if netErr, ok := err.(net.Error); ok {
					if netErr.Timeout() {
						fmt.Print("websocket receive message from " + conn.RemoteAddr().String() + " timeout")
						return
					}
				}
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					fmt.Printf("websocket receive message from %v error: %v \n", conn.RemoteAddr(), err)
					return
				}
				return
			}
			fmt.Println(time.Unix(recvmsg.Timestamp, 0).Format(time.RFC3339), recvmsg.Content)
		}
	}()

	go func() {
		defer close(done)
		for _, s := range strslice {
			<-time.After(time.Second)
			conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(600))) //close connection after 600s
			sendmsg := new(lib.MsgStruct)
			sendmsg.Timestamp = time.Now().Unix()
			sendmsg.Content = s
			err := conn.WriteJSON(&sendmsg)
			if err != nil {
				if netErr, ok := err.(net.Error); ok {
					if netErr.Timeout() {
						fmt.Print("websocket send message to " + conn.RemoteAddr().String() + " timeout")
						return
					}
				}
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					fmt.Printf("websocket send message to %v error: %v \n", conn.RemoteAddr(), err)
					return
				}
				return
			}
		}
	}()

	for range done {
		return
	}
}
