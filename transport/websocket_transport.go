package transport

import (
	"errors"
	"sync"
	"time"
	"net"

	"github.com/cloudwebrtc/go-protoo/logger"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 5 * time.Second //(pongWait * 8) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4096
)

type TransportErr struct {
	Code int
	Text string
}

type TransportChans struct {
	OnMsg chan []byte
	OnErr chan TransportErr
	OnClose chan TransportErr
	SendCh chan []byte
}

type WebSocketTransport struct {
	TransportChans
	socket *websocket.Conn
	mutex  *sync.Mutex
	closed bool
	stop chan struct{}
}



func NewWebSocketTransport(socket *websocket.Conn) *WebSocketTransport {
	var transport WebSocketTransport
	transport.socket = socket
	transport.mutex = new(sync.Mutex)
	transport.closed = false
	transport.stop = make(chan struct{})

	transport.socket.SetCloseHandler(func(code int, text string) error {
		logger.Warnf("On transport close %s [%d]", text, code)
		// transport.OnClose <- TransportErr{code, text}
		// transport.Close()
		if !transport.closed {
			close(transport.stop)
		}
		return nil
	})

	transport.TransportChans = TransportChans{
		OnMsg: make(chan []byte, 1),
		OnErr: make(chan TransportErr, 1),
		OnClose: make(chan TransportErr, 1),
		SendCh: make(chan []byte, 1),
	}
	return &transport
}
//
// func (transport *WebSocketTransport) ReadMessage() {
// 	in := make(chan []byte)
// 	stop := make(chan struct{})
// 	// pingTicker := time.NewTicker(pingPeriod)
//
// 	var c = transport.socket
// 	go func() {
// 		for {
// 			_, message, err := c.ReadMessage()
// 			if err != nil {
// 				logger.Warnf("Got error: %v", err)
// 				if c, k := err.(*websocket.CloseError); k {
// 					transport.OnClose <- TransportErr{c.Code, c.Text}
// 					// transport.Emit("error", c.Code, c.Text)
// 				} else {
// 					if c, k := err.(*net.OpError); k {
// 						transport.OnClose <- TransportErr{1008, c.Error()}
// 						// transport.Emit("error", 1008, c.Error())
// 					}
// 				}
// 				close(stop)
// 				break
// 			}
// 			in <- message
// 		}
// 	}()
//
// 	// for {
// 	// 	select {
// 	// 	case _ = <-pingTicker.C:
// 	// 		logger.Debugf("Send keepalive !!!")
// 	// 		if err := transport.Send(websocket.PingMessage); err != nil {
// 	// 			logger.Errorf("Keepalive has failed")
// 	// 			pingTicker.Stop()
// 	// 			return
// 	// 		}
// 	// 	case message := <-in:
// 	// 		{
// 	// 			logger.Infof("Recivied data: %s", message)
// 	// 			// transport.Emit("message", []byte(message))
// 	// 			transport.OnMsg <- []byte(message)
// 	// 		}
// 	// 	case <-stop:
// 	// 		return
// 	// 	}
// 	// }
// }

func (transport *WebSocketTransport) Start() {
	go transport.ReadLoop()
	go transport.WriteLoop()
}

func (transport *WebSocketTransport) ReadLoop() {
	transport.socket.SetReadLimit(maxMessageSize)
	transport.socket.SetReadDeadline(time.Now().Add(pongWait))
	transport.socket.SetPongHandler(func(string) error {
		transport.socket.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for !transport.closed {
		_, message, err := transport.socket.ReadMessage()
		if err != nil {
			logger.Warnf("Got error: %v", err)
			if c, k := err.(*websocket.CloseError); k {
				transport.OnClose <- TransportErr{c.Code, c.Text}
				// transport.Emit("error", c.Code, c.Text)
			} else {
				if c, k := err.(*net.OpError); k {
					transport.OnClose <- TransportErr{1008, c.Error()}
					// transport.Emit("error", 1008, c.Error())
				}
			}
			// Probably drop
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Errorf("Error reading message: %v\n", err)
			}
			break
		}


		logger.Infof("Received: %s", message)
		transport.OnMsg <- []byte(message)
	}
}

func (transport *WebSocketTransport) WriteLoop() {
	defer transport.close()

	pingTicker := time.NewTicker(pingPeriod)
	doneC := false

	for !doneC {
		select {
		case _ = <-pingTicker.C:
			logger.Debugf("Send keepalive !!!")
			if err := transport.socket.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Errorf("Keepalive has failed")
				pingTicker.Stop()
				// TODO Trigger close
				doneC = true
				break
			}
		case message := <-transport.SendCh:
			{
				logger.Infof("Send data: %s", message)
				if transport.closed {
					logger.Infof("Transport closed. Exiting write loop")
					break
					// TODO shutdown transport
					// errors.New("websocket: write closed")
				}

				err := transport.socket.WriteMessage(websocket.TextMessage, message)
				// TODO handle send error
				if err != nil {
					logger.Warnf("Socker send Error %v", err)
				}
			}
		case <-transport.stop:
			doneC = true
			break
		}
	}
	logger.Infof("exited write loop")

}

/*
* Send |message| to the connection.
 */
func (transport *WebSocketTransport) Send(message string) error {
	logger.Infof("Send data: %s", message)
	transport.mutex.Lock()
	defer transport.mutex.Unlock()
	if transport.closed {
		return errors.New("websocket: write closed")
	}
	return transport.socket.WriteMessage(websocket.TextMessage, []byte(message))
}

/*
* Close connection.
 */

func (transport *WebSocketTransport) Close() {
 close(transport.stop)
}

func (transport *WebSocketTransport) close() {
	transport.mutex.Lock()
	defer transport.mutex.Unlock()
	if transport.closed == false {
		logger.Infof("Close ws transport now : ")
		transport.socket.Close()
		transport.closed = true
	} else {
		logger.Warnf("Transport already closed :")
	}
}
