package client

import (
	"crypto/tls"
	"net"
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
	client.handleWebSocket(client.transport)
	return &client
}

func (client *WebSocketClient) ReadMessage() {
	if client == nil {
		logger.Errorf("Client is nil")
		return
	}
	in := make(chan []byte)
	stop := make(chan struct{})
	pingTicker := time.NewTicker(pingPeriod)
	var c = client.socket
	var transport = client.transport
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				logger.Warnf("Got error: %v", err)
				if c, k := err.(*websocket.CloseError); k {
					transport.Emit("error", c.Code, c.Text)
				} else {
					if c, k := err.(*net.OpError); k {
						transport.Emit("error", 1008, c.Error())
					}
				}
				close(stop)
				break
			}
			in <- message
		}
	}()

	for {
		select {
		case <-stop:
			return
		case message := <-in:
			{
				logger.Infof("Recivied data: %s", message)
				transport.Emit("message", []byte(message))
			}
		case _ = <-pingTicker.C:
			logger.Infof("Send keepalive !!!")
			if err := transport.Send("{}"); err != nil {
				logger.Errorf("Keepalive has failed")
				pingTicker.Stop()
				return
			}
		}
	}
}

func (client *WebSocketClient) GetTransport() *transport.WebSocketTransport {
	return client.transport
}

func (client *WebSocketClient) Close() {
	client.transport.Close()
}
