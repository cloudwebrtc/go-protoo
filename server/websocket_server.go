package server

import (
	"net/http"
	"strconv"

	"github.com/cloudwebrtc/go-protoo/logger"
	"github.com/cloudwebrtc/go-protoo/transport"
	"github.com/gorilla/websocket"
)

type WebSocketServerConfig struct {
	Host          string
	Port          int
	CertFile      string
	KeyFile       string
	HTMLRoot      string
	WebSocketPath string
}

func DefaultConfig() WebSocketServerConfig {
	return WebSocketServerConfig{
		Host:          "0.0.0.0",
		Port:          8443,
		HTMLRoot:      ".",
		WebSocketPath: "/ws",
	}
}

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

func (server *WebSocketServer) Bind(cfg WebSocketServerConfig) {
	// Websocket handle func
	http.HandleFunc(cfg.WebSocketPath, server.handleWebSocketRequest)
	http.Handle("/", http.FileServer(http.Dir(cfg.HTMLRoot)))

	if cfg.CertFile == "" || cfg.KeyFile == "" {
		logger.Infof("non-TLS WebSocketServer listening on: %s:%d", cfg.Host, cfg.Port)
		panic(http.ListenAndServe(cfg.Host+":"+strconv.Itoa(cfg.Port), nil))
	} else {
		logger.Infof("TLS WebSocketServer listening on: %s:%d", cfg.Host, cfg.Port)
		panic(http.ListenAndServeTLS(cfg.Host+":"+strconv.Itoa(cfg.Port), cfg.CertFile, cfg.KeyFile, nil))
	}
}
