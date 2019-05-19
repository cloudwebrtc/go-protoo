package main

import (
	"encoding/json"
	"protoo/client"
	"protoo/logger"
	"protoo/peer"
	"protoo/transport"
)

func JsonEncode(str string) map[string]interface{} {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(str), &data); err != nil {
		panic(err)
	}
	return data
}

type AcceptFunc func(data map[string]interface{})
type RejectFunc func(errorCode int, errorReason string)

func handleWebSocketOpen(transport *transport.WebSocketTransport) {
	logger.Infof("handleWebSocketOpen")

	peer := peer.NewPeer("client-xxxx", transport)
	peer.On("close", func() {
	})

	handleRequest := func(request map[string]interface{}, accept AcceptFunc, reject RejectFunc) {
		method := request["method"]
		logger.Infof("handleRequest =>  (%s) ", method)
		if method == "kick" {
			reject(486, "Busy Here")
		} else if method == "offer" {
			reject(500, "sdp error!")
		}
	}

	handleNotification := func(notification map[string]interface{}) {
		logger.Infof("handleNotification => %s", notification["method"])
	}

	handleClose := func() {
		logger.Infof("handleClose => peer (%s) ", peer.ID())
	}

	peer.On("request", handleRequest)
	peer.On("notification", handleNotification)
	peer.On("close", handleClose)

	peer.Request("login", JsonEncode(`{"username":"xxxx","password":"XXXX"}`),
		func(result map[string]interface{}) {
			logger.Infof("login success: =>  %s", result)
		},
		func(code int, err string) {
			logger.Infof("login reject: %d => %s", code, err)
		})
}

func main() {
	var ws_client = client.NewClient("wss://127.0.0.1:8443/ws?peer-id=xxxx&room-id=bbbb", handleWebSocketOpen)
	ws_client.ReadMessage()
}
