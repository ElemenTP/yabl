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
var url string

func init() {
	testserver = lib.NewWsServer("127.0.0.1:34599", "tcp")
	url = "ws://127.0.0.1:34599/ws"
	go testserver.Start()
}

func Test_Httphandle(t *testing.T) {
	resp, err := http.Get("http://127.0.0.1:34599/ws")
	if err != nil {
		t.FailNow()
	}
	resp.Body.Close()
}

func Benchmark_Httphandle(b *testing.B) {
	for i := 0; i < b.N; i++ { //use b.N for looping
		resp, err := http.Get("http://127.0.0.1:34599/ws")
		if err != nil {
			b.FailNow()
		}
		resp.Body.Close()
	}
}

func Test_Websockethandle(t *testing.T) {
	scriptstr := `
func main:
- postmsg "你好"`
	genScript([]byte(scriptstr))
	lib.Compile()

	<-time.After(time.Second)
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
	fmt.Println(time.Unix(recvmsg.Timestamp, 0).Format(time.RFC3339), recvmsg.Content)
}

func overalltest(script *string, input *string, t *testing.T) {
	genScript([]byte(*script))
	lib.Compile()

	<-time.After(time.Second)
	strslice := strings.Split(*input, "\n")

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

	done := make(chan int)

	go func() {
		defer func() {
			done <- 1
		}()
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
		defer func() {
			done <- 1
		}()
		for _, s := range strslice {
			<-time.After(time.Second)
			conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(600))) //close connection after 600s
			sendmsg := new(lib.MsgStruct)
			sendmsg.Timestamp = time.Now().Unix()
			sendmsg.Content = s
			fmt.Println(s)
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

func Test_Websockethandles1(t *testing.T) {
	scriptstr := `
address: 127.0.0.1
port: 8080
func main:
  - hello = "亲亲，我是疼殉客服机器人小美，有什么问题尽管问我吧！"
  - postmsg hello
  - flag1 = ""
  - loop
  - loop
  - answer = getmsg
  - flag2 = answer contain "跳楼"
  - if flag2
  - postmsg "亲亲，消消气，消消气。"
  - break
  - fi
  - postmsg "亲亲，您不要生气呢，这边正在尝试解决，可以多等待几天看看呢。"
  - pool
  - flag1 = flag1 join "0"
  - flag3 = flag1 equal "000"
  - if flag3
  - break
  - fi
  - pool
  - postmsg "亲亲，正在为您接入人工客服呢。"`

	inputstr := `你好
我的微信号为什么被封了
我就是正常使用而已
你复读个什么呢
无语
跳楼～
跳楼～
跳楼～
。。。`

	overalltest(&scriptstr, &inputstr, t)
}

func Test_Websockethandles2(t *testing.T) {
	scriptstr := `
address: 127.0.0.1
port: 8080
func main:
  - postmsg "您好，这里是自动服务机器人，请问您要办理什么业务呢？\n本机器人可以办理开户、查询、咨询、注销等业务。"
  - loop
  - answer = getmsg
  - flag1 = answer contain "开户"
  - flag2 = answer contain "查询"
  - flag3 = answer contain "咨询"
  - flag4 = answer contain "注销"
  - if flag1
  - postmsg "正在为您转到开户业务，请稍等。"
  - break
  - elif flag2
  - postmsg "正在为您转到查询业务，请稍等。"
  - break
  - elif flag3
  - postmsg "正在为您转到咨询业务，请稍等。"
  - break
  - elif flag4
  - postmsg "正在为您转到注销业务，请稍等。"
  - break
  - else
  - postmsg "对不起，没有听懂。\n本机器人可以办理开户、查询、咨询、注销等业务。"
  - fi
  - pool
  - postmsg "感谢您的使用，再见。"`
	inputstr := `你好
我要查询
。。。
`

	overalltest(&scriptstr, &inputstr, t)
}

func Test_Websockethandles3(t *testing.T) {
	script := `#test script 3
address: 127.0.0.1
port: 8080
func main:
  - hello = invoke test "param1" "param2"
  - postmsg hello

func test param1:
  - return "hello"`
	input := ""
	overalltest(&script, &input, t)

}
