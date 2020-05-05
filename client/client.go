package client

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/chuckpreslar/emission"
	"github.com/cloudwebrtc/go-protoo/logger"
	"github.com/cloudwebrtc/go-protoo/transport"
	"github.com/gorilla/websocket"
)

const pingPeriod = 5 * time.Second

type WebSocketClient struct {
	emission.Emitter
	socket          *websocket.Conn
	transport       *transport.WebSocketTransport
	handleWebSocket func(ws *transport.WebSocketTransport)
}

func NewClient(url string, handleWebSocket func(ws *transport.WebSocketTransport)) *WebSocketClient {
	var client WebSocketClient
	client.Emitter = *emission.NewEmitter()
	logger.Infof("Connecting to %s", url)

	responseHeader := http.Header{}
	responseHeader.Add("Sec-WebSocket-Protocol", "protoo")

	// only for testing
	tls_cfg := &tls.Config{
		InsecureSkipVerify: true,
	}

	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		TLSClientConfig:  tls_cfg,
	}

	socket, _, err := dialer.Dial(url, responseHeader)
	if err != nil {
		logger.Errorf("Dial failed: %v", err)
		return nil
	}
	client.socket = socket
	client.handleWebSocket = handleWebSocket
	client.transport = transport.NewWebSocketTransport(socket)
	client.transport.Start()
	client.handleWebSocket(client.transport)
	return &client
}

func (client *WebSocketClient) GetTransport() *transport.WebSocketTransport {
	return client.transport
}

func (client *WebSocketClient) Close() {
	client.transport.Close()
}
