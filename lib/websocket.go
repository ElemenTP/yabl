package lib

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WsServer struct {
	Listener net.Listener        //addr listener
	Address  string              //listen address
	Lnet     string              //listen type
	Upgrade  *websocket.Upgrader //websocket upgrader
}

//Initialize a new websocket server struct.
func NewWsServer(address string, lnet string) *WsServer {
	ws := new(WsServer)
	ws.Address = address
	ws.Lnet = lnet
	ws.Upgrade = &websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			if r.Method != "GET" {
				return false
			}
			if r.URL.Path != "/ws" {
				return false
			}
			return true
		},
	}
	return ws
}

//Websocket server begin listen and serve.
func (w *WsServer) Start() {
	listener, err := net.Listen(w.Lnet, w.Address)
	if err != nil {
		log.Fatalln("[Server] Error : shutting server", err)
	}
	w.Listener = listener
	log.Println("[Server] Info : listening address", listener.Addr().String())
	if err := http.Serve(w.Listener, w); err != nil {
		log.Fatalln("[Server] Error : shutting server", err)
	}
}

//Judge if a connection is valid and upgrade http to websocket.
func (w *WsServer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/ws" {
		httpCode := http.StatusNotFound
		reasePhrase := http.StatusText(httpCode)
		log.Println("[Server] Error : path error", reasePhrase)
		http.Error(rw, reasePhrase, httpCode)
		return
	}

	conn, err := w.Upgrade.Upgrade(rw, r, nil)
	if err != nil {
		log.Println("[Server] Error", err)
		return
	}
	log.Println("[Server] Info : client connect", conn.RemoteAddr())
	go w.ConnHandle(conn)
}

//Handle a websocket connection.
func (w *WsServer) ConnHandle(conn *websocket.Conn) {
	defer func() {
		recover()
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		<-time.After(time.Second)
		log.Println("[Server] Info : client disconnect", conn.RemoteAddr())
	}()

	params := make([]string, 0)
	FuncInvoker("main", &params, conn)
}
