package peer

import (
	"encoding/json"

	"github.com/cloudwebrtc/go-protoo/logger"
	"github.com/cloudwebrtc/go-protoo/transport"

	"github.com/chuckpreslar/emission"
)

type Transcation struct {
	id     int
	accept AcceptFunc
	reject RejectFunc
	close  func()
}

type Peer struct {
	emission.Emitter
	id           string
	transport    *transport.WebSocketTransport
	transcations map[int]*Transcation
}

func NewPeer(id string, transport *transport.WebSocketTransport) *Peer {
	var peer Peer
	peer.Emitter = *emission.NewEmitter()
	peer.id = id
	peer.transport = transport
	peer.transport.On("message", peer.handleMessage)
	peer.transport.On("close", func(code int, err string) {
		logger.Infof("Transport closed [%d] %s", code, err)
		peer.Emit("close", code, err)
	})
	peer.transport.On("error", func(code int, err string) {
		logger.Warnf("Transport got error (%d, %s)", code, err)
		peer.Emit("error", code, err)
	})
	peer.transcations = make(map[int]*Transcation)
	return &peer
}

func (peer *Peer) Close() {
	peer.transport.Close()
	peer.Emit("close", 1000, "")
}

func (peer *Peer) ID() string {
	return peer.id
}

func (peer *Peer) Request(method string, data map[string]interface{}, success AcceptFunc, reject RejectFunc) {
	id := GenerateRandomNumber()
	request := &Request{
		Request: true,
		Id:      id,
		Method:  method,
		Data:    data,
	}
	str, err := json.Marshal(request)
	if err != nil {
		logger.Errorf("Marshal %v", err)
		return
	}

	transcation := &Transcation{
		id:     id,
		accept: success,
		reject: reject,
		close: func() {
			logger.Infof("Transport closed !")
		},
	}

	peer.transcations[id] = transcation
	logger.Infof("Send request [%s]", method)
	peer.transport.Send(string(str))
}

func (peer *Peer) Notify(method string, data map[string]interface{}) {
	notification := &Notification{
		Notification: true,
		Method:       method,
		Data:         data,
	}
	str, err := json.Marshal(notification)
	if err != nil {
		logger.Errorf("Marshal %v", err)
		return
	}
	logger.Infof("Send notification [%s]", method)
	peer.transport.Send(string(str))
}

func (peer *Peer) handleMessage(message []byte) {
	var msg PeerMsg
	if err := json.Unmarshal(message, &msg); err != nil {
		panic(err)
	}
	if msg.Request {
		var data Request
		if err := json.Unmarshal(message, &data); err != nil {
			panic(err)
		}
		peer.handleRequest(data)
	} else if msg.Response {
		if msg.Ok {
			var data Response
			if err := json.Unmarshal(message, &data); err != nil {
				panic(err)
			}
			peer.handleResponse(data)
		} else {
			var data ResponseError
			if err := json.Unmarshal(message, &data); err != nil {
				panic(err)
			}
			peer.handleResponseError(data)
		}
	} else if msg.Notification {
		var data Notification
		if err := json.Unmarshal(message, &data); err != nil {
			panic(err)
		}
		peer.handleNotification(data)
	}
	return
}

func (peer *Peer) handleRequest(request Request) {

	logger.Infof("Handle request [%s]", request.Method)

	accept := func(data map[string]interface{}) {
		response := &Response{
			Response: true,
			Ok:       true,
			Id:       request.Id,
			Data:     data,
		}
		str, err := json.Marshal(response)
		if err != nil {
			logger.Errorf("Marshal %v", err)
			return
		}
		//send accept
		logger.Infof("Accept [%s] => (%s)", request.Method, str)
		peer.transport.Send(string(str))
	}

	reject := func(errorCode int, errorReason string) {
		response := &ResponseError{
			Response:    true,
			Ok:          false,
			Id:          request.Id,
			ErrorCode:   errorCode,
			ErrorReason: errorReason,
		}
		str, err := json.Marshal(response)
		if err != nil {
			logger.Errorf("Marshal %v", err)
			return
		}
		//send reject
		logger.Infof("Reject [%s] => (errorCode:%d, errorReason:%s)", request.Method, errorCode, errorReason)
		peer.transport.Send(string(str))
	}

	peer.Emit("request", request, accept, reject)
}

func (peer *Peer) handleResponse(response Response) {
	id := response.Id
	transcation := peer.transcations[id]
	if transcation == nil {
		logger.Errorf("received response does not match any sent request [id:%d]", id)
		return
	}

	transcation.accept(response.Data)
	delete(peer.transcations, id)
}

func (peer *Peer) handleResponseError(response ResponseError) {
	id := response.Id
	transcation := peer.transcations[id]
	if transcation == nil {
		logger.Errorf("received response does not match any sent request [id:%d]", id)
		return
	}

	transcation.reject(response.ErrorCode, response.ErrorReason)
	delete(peer.transcations, id)
}

func (peer *Peer) handleNotification(notification Notification) {
	peer.Emit("notification", notification)
}
