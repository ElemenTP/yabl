package lib

import (
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

type WsServer struct {
	listener net.Listener
	address  string
	lnet     string
	upgrade  *websocket.Upgrader
}

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

func (w *WsServer) Start() {
	listener, err := net.Listen(w.lnet, w.address)
	if err != nil {
		log.Fatalln(err)
	}
	w.listener = listener
	log.Println("Listening address ", listener.Addr().String())
	if err := http.Serve(w.listener, w); err != nil {
		log.Fatalln("Shutting server ", err)
	}
}

func (w *WsServer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/ws" {
		httpCode := http.StatusNotFound
		reasePhrase := http.StatusText(httpCode)
		log.Println("Path error ", reasePhrase)
		http.Error(rw, reasePhrase, httpCode)
		return
	}

	conn, err := w.upgrade.Upgrade(rw, r, nil)
	if err != nil {
		log.Println("Websocket error ", err)
		return
	}
	log.Println("Client connect ", conn.RemoteAddr())
	go w.connHandle(conn)
}

func (w *WsServer) connHandle(conn *websocket.Conn) {
	defer func() {
		conn.Close()
	}()
}
