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
	stopLock sync.RWMutex
	shutdown bool
	wg sync.WaitGroup
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
		transport.Stop()

		return nil
	})

	transport.TransportChans = TransportChans{
		OnMsg: make(chan []byte, 100),
		OnErr: make(chan TransportErr, 1),
		OnClose: make(chan TransportErr, 1),
		SendCh: make(chan []byte, 100),
	}
	return &transport
}

func (transport *WebSocketTransport) Start() {
	go transport.ReadLoop()
	go transport.WriteLoop()
	transport.wg.Add(2)

	// Cleanup after exit
	go func() {
		transport.wg.Wait()
		transport.close()
	}()
}

func (transport *WebSocketTransport) ReadLoop() {
	defer func() {
		transport.wg.Done()
		logger.Debugf("EXITING READ LOOP")
		close(transport.stop)
	}()

	transport.socket.SetReadLimit(maxMessageSize)
	transport.socket.SetReadDeadline(time.Now().Add(pongWait))
	transport.socket.SetPongHandler(func(string) error {
		transport.socket.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := transport.socket.ReadMessage()
		if err != nil {
			logger.Warnf("Got error: %v", err)
			if c, k := err.(*websocket.CloseError); k {
				transport.OnClose <- TransportErr{c.Code, c.Text}
				logger.Debugf("Send close msg")

				// transport.Emit("error", c.Code, c.Text)
			} else {
				if c, k := err.(*net.OpError); k {
					logger.Debugf("Send err msg")
					transport.OnErr <- TransportErr{1008, c.Error()}
					// transport.Emit("error", 1008, c.Error())
				}
			}
			// // Probably drop
			// if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			// 	logger.Errorf("Error reading message: %v\n", err)
			// }
			break
		}


		logger.Infof("Received: %s", message)
		transport.OnMsg <- []byte(message)

		// Check stop
		transport.stopLock.RLock()
		stop := transport.shutdown
		transport.stopLock.RUnlock()
		if stop {
			return
		}
	}


}

func (transport *WebSocketTransport) WriteLoop() {
	defer func() {
		logger.Debugf("exited ws write loop")
		transport.wg.Done()
		// Make sure reader goes down if it has not alreday
		transport.Stop()
	}()

	pingTicker := time.NewTicker(pingPeriod)
	doneC := false

	for !doneC {
		select {
		case _ = <-pingTicker.C:
			logger.Debugf("Send keepalive !!!")
			if err := transport.socket.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Warnf("Keepalive has failed")
				pingTicker.Stop()
				return
			}
		case message := <-transport.SendCh:
			{
				logger.Infof("Send data: %s", message)
				// if transport.closed {
				// 	logger.Infof("Transport closed. Exiting write loop")
				// 	break
				// 	// TODO shutdown transport
				// 	// errors.New("websocket: write closed")
				// }

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

// Start shutdown
// First shutdown read loop, then write loop, then finally close the socket

func (transport *WebSocketTransport) Close() {
 transport.Stop()
}

func (transport *WebSocketTransport) Stop() {
 transport.stopLock.Lock()
 transport.shutdown = true
 transport.stopLock.Unlock()
}


/*
* Close connection.
 */

func (transport *WebSocketTransport) close() {
	transport.mutex.Lock()
	defer transport.mutex.Unlock()
	if transport.closed == false {
		logger.Debugf("Close ws transport now")
		transport.socket.Close()
		transport.closed = true
	} else {
		logger.Warnf("Transport already closed")
	}
}
