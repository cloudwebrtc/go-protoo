package transport

import (
	"protoo/logger"
	"strconv"
	"time"

	"github.com/chuckpreslar/emission"
	"github.com/gorilla/websocket"
)

const pingPeriod = 5 * time.Second

type WebSocketTransport struct {
	emission.Emitter
	socket  *websocket.Conn
	msgType int
}

func NewWebSocketTransport(socket *websocket.Conn) *WebSocketTransport {
	var transport WebSocketTransport
	transport.Emitter = *emission.NewEmitter()
	transport.socket = socket
	transport.socket.SetCloseHandler(func(code int, text string) error {
		logger.Warnf("%s [%d]", text, strconv.Itoa(code))
		transport.Emit("close")
		return nil
	})
	return &transport
}

func (transport *WebSocketTransport) ReadMessage() {
	in := make(chan []byte)
	stop := make(chan struct{})
	pingTicker := time.NewTicker(pingPeriod)

	var c = transport.socket
	go func() {
		for {
			mt, message, err := c.ReadMessage()
			transport.msgType = mt
			if err != nil {
				logger.Warnf("Got error:", err)
				if c, k := err.(*websocket.CloseError); k {
					transport.Emit("error", c.Code, c.Text)
				}
				close(stop)
				break
			}
			in <- message
		}
	}()

	for {
		select {
		case _ = <-pingTicker.C:
			logger.Infof("Send keepalive !!!")
			if err := transport.Send("{}"); err != nil {
				logger.Errorf("Keepalive has failed")
				pingTicker.Stop()
				return
			}
		case message := <-in:
			{
				logger.Infof("Recivied data: %s", message)
				transport.Emit("message", []byte(message))
			}
		case <-stop:
			return
		}
	}
}

/*
* Send |message| to the connection.
 */
func (transport *WebSocketTransport) Send(message string) error {
	logger.Infof("Send data: %s", message)
	return transport.socket.WriteMessage(transport.msgType, []byte(message))
}

/*
* Close connection.
 */
func (transport *WebSocketTransport) Close() {
	logger.Infof("Close transport: ", transport)
	transport.socket.Close()
}
