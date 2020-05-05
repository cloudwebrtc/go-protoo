package peer

import (
	"encoding/json"

	"github.com/cloudwebrtc/go-protoo/logger"
	"github.com/cloudwebrtc/go-protoo/transport"
)

type Transcation struct {
	id     int
	accept AcceptFunc
	reject RejectFunc
	close  func()
}

type RequestData struct {
	Accept  RespondFunc
	Reject  RejectFunc
	Request Request
}

type PeerChans struct {
	OnRequest      chan RequestData
	OnNotification chan transport.TransportErr
	OnClose        chan transport.TransportErr
	OnError        chan transport.TransportErr
}

type Peer struct {
	// emission.Emitter
	PeerChans
	id           string
	transport    *transport.WebSocketTransport
	transcations map[int]*Transcation
}

func NewPeer(id string, con *transport.WebSocketTransport) *Peer {
	var peer Peer
	// peer.Emitter = *emission.NewEmitter()
	peer.id = id
	peer.transport = con
	// peer.transport.On("message", peer.handleMessage)/
	// peer.transport.On("close", func(code int, err string) {
	// 	logger.Infof("Transport closed [%d] %s", code, err)
	// 	peer.Emit("close", code, err)
	// })
	// peer.transport.On("error", func(code int, err string) {
	// 	logger.Warnf("Transport got error (%d, %s)", code, err)
	// 	peer.Emit("error", code, err)
	// })
	peer.PeerChans = PeerChans{
		OnRequest:      make(chan RequestData),
		OnNotification: make(chan transport.TransportErr),
	}
	peer.transcations = make(map[int]*Transcation)
	return &peer
}

func (peer *Peer) Run() {
	for {
		select {
		case msg := <-peer.transport.OnMsg:
			peer.handleMessage(msg)

		case err := <-peer.transport.OnErr:
			peer.handleErr(err)

		case err := <-peer.transport.OnClose:
			peer.handleClose(err)
		}
	}
}

func (peer *Peer) handleClose(err transport.TransportErr) {
	logger.Infof("Transport closed [%d] %s", err.Code, err.Text)
	// peer.Emit("close", code, err)
}

func (peer *Peer) handleErr(err transport.TransportErr) {
	logger.Warnf("Transport got error (%d, %s)", err.Code, err.Text)
	// peer.Emit("error", code, err)
}

func (peer *Peer) Close() {
	peer.transport.Close()
	// peer.Emit("close", 1000, "")
}

func (peer *Peer) ID() string {
	return peer.id
}

func (peer *Peer) Request(method string, data interface{}, success AcceptFunc, reject RejectFunc) {
	id := GenerateRandomNumber()
	dataStr, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("Marshal data %v", err)
		return
	}
	request := &Request{
		Request: true,
		Id:      id,
		Method:  method,
		Data:    dataStr,
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
	peer.transport.SendCh <- str
}

func (peer *Peer) Notify(method string, data interface{}) {
	dataStr, err := json.Marshal(data)
	if err != nil {
		logger.Errorf("Marshal data %v", err)
		return
	}
	notification := &Notification{
		Notification: true,
		Method:       method,
		Data:         dataStr,
	}
	str, err := json.Marshal(notification)
	if err != nil {
		logger.Errorf("Marshal %v", err)
		return
	}
	logger.Infof("Send notification [%s]", method)
	peer.transport.SendCh <- str
}

func (peer *Peer) handleMessage(message []byte) {
	var msg PeerMsg
	if err := json.Unmarshal(message, &msg); err != nil {
		logger.Errorf("Marshal %v", err)
		return
	}
	if msg.Request {
		var data Request
		if err := json.Unmarshal(message, &data); err != nil {
			logger.Errorf("Request Marshal %v", err)
			return
		}
		peer.handleRequest(data)
	} else if msg.Response {
		if msg.Ok {
			var data Response
			if err := json.Unmarshal(message, &data); err != nil {
				logger.Errorf("Response Marshal %v", err)
				return
			}
			peer.handleResponse(data)
		} else {
			var data ResponseError
			if err := json.Unmarshal(message, &data); err != nil {
				logger.Errorf("ResponseError Marshal %v", err)
				return
			}
			peer.handleResponseError(data)
		}
	} else if msg.Notification {
		var data Notification
		if err := json.Unmarshal(message, &data); err != nil {
			logger.Errorf("Notification Marshal %v", err)
			return
		}
		peer.handleNotification(data)
	}
}

func (peer *Peer) handleRequest(request Request) {

	logger.Infof("Handle request [%s]", request.Method)

	accept := func(data interface{}) {
		dataStr, err := json.Marshal(data)
		if err != nil {
			logger.Errorf("Marshal %v", err)
			return
		}
		response := &Response{
			Response: true,
			Ok:       true,
			Id:       request.Id,
			Data:     dataStr,
		}
		str, err := json.Marshal(response)
		if err != nil {
			logger.Errorf("Marshal %v", err)
			return
		}
		//send accept
		logger.Infof("Accept [%s] => (%s)", request.Method, str)
		peer.transport.SendCh <- str
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
		peer.transport.SendCh <- str
	}
	_, _ = accept, reject
	peer.OnRequest <- RequestData{
		Accept:  accept,
		Reject:  reject,
		Request: request,
	}
	// peer.Emit("request", request, accept, reject)
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
	// peer.Emit("notification", notification)
}
