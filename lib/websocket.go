package lib

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WsServer struct {
	listener net.Listener
	address  string
	lnet     string
	upgrade  *websocket.Upgrader
}

//construct a new websocket server struct
func NewWsServer(address string, lnet string) *WsServer {
	ws := new(WsServer)
	ws.address = address
	ws.lnet = lnet
	ws.upgrade = &websocket.Upgrader{
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

//websocket server begin listen and serve.
func (w *WsServer) Start() {
	listener, err := net.Listen(w.lnet, w.address)
	if err != nil {
		log.Fatalln("[Server] Error : shutting server", err)
	}
	w.listener = listener
	log.Println("[Server] Info : listening address", listener.Addr().String())
	if err := http.Serve(w.listener, w); err != nil {
		log.Fatalln("[Server] Error : shutting server", err)
	}
}

//judge if a connection is valid and upgrade http to websocket
func (w *WsServer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/ws" {
		httpCode := http.StatusNotFound
		reasePhrase := http.StatusText(httpCode)
		log.Println("[Server] Error : path error", reasePhrase)
		http.Error(rw, reasePhrase, httpCode)
		return
	}

	conn, err := w.upgrade.Upgrade(rw, r, nil)
	if err != nil {
		log.Println("[Server] Error", err)
		return
	}
	log.Println("[Server] Info : client connect", conn.RemoteAddr())
	go w.connHandle(conn)
}

//handle a websocket connection
func (w *WsServer) connHandle(conn *websocket.Conn) {
	defer func() {
		recover()
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		<-time.After(time.Second)
		log.Println("[Server] Info : client disconnect", conn.RemoteAddr())
	}()

	params := make([]string, 0)
	funcInvoker("main", &params, conn)
}
