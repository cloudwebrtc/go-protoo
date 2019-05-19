package server

import (
	"net/http"
	"protoo/logger"
	"protoo/transport"

	"github.com/gorilla/websocket"
)

type WebSocketServer struct {
	handleWebSocket func(ws *transport.WebSocketTransport, request *http.Request)
	// Websocket upgrader
	upgrader websocket.Upgrader
}

func NewWebSocketServer(handler func(ws *transport.WebSocketTransport, request *http.Request)) *WebSocketServer {
	var server = &WebSocketServer{
		handleWebSocket: handler,
	}
	server.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return server
}

func (server *WebSocketServer) handleWebSocketRequest(writer http.ResponseWriter, request *http.Request) {
	responseHeader := http.Header{}
	responseHeader.Add("Sec-WebSocket-Protocol", "protoo")
	socket, err := server.upgrader.Upgrade(writer, request, responseHeader)
	if err != nil {
		panic(err)
	}
	wsTransport := transport.NewWebSocketTransport(socket)
	server.handleWebSocket(wsTransport, request)
	wsTransport.ReadMessage()
}

func (server *WebSocketServer) Bind(host string, port string) {
	// Websocket handle func
	http.HandleFunc("/ws", server.handleWebSocketRequest)
	http.Handle("/", http.FileServer(http.Dir(".")))
	logger.Infof("WebSocketServer listening on: %s:%s", host, port)
	panic(http.ListenAndServeTLS(host+":"+port, "certs/cert.pem", "certs/key.pem", nil))
}
