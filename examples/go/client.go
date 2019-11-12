package main

import (
	"encoding/json"

	"github.com/cloudwebrtc/go-protoo/client"
	"github.com/cloudwebrtc/go-protoo/logger"
	"github.com/cloudwebrtc/go-protoo/peer"
	"github.com/cloudwebrtc/go-protoo/transport"
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

var peerId = "go-client-id-xxxx"

func handleWebSocketOpen(transport *transport.WebSocketTransport) {
	logger.Infof("handleWebSocketOpen")

	peer := peer.NewPeer(peerId, transport)
	peer.On("close", func(code int, err string) {
		logger.Infof("peer close [%d] %s", code, err)
	})

	handleRequest := func(request map[string]interface{}, accept AcceptFunc, reject RejectFunc) {
		method := request["method"]
		logger.Infof("handleRequest =>  (%s) ", method)
		if method == "kick" {
			reject(486, "Busy Here")
		} else {
			accept(JsonEncode(`{}`))
		}
	}

	handleNotification := func(notification map[string]interface{}) {
		logger.Infof("handleNotification => %s", notification["method"])
	}

	handleClose := func(code int, err string) {
		logger.Infof("handleClose => peer (%s) [%d] %s", peer.ID(), code, err)
	}

	peer.On("request", handleRequest)
	peer.On("notification", handleNotification)
	peer.On("close", handleClose)

	peer.Request("login", JsonEncode(`{"username":"alice","password":"alicespass"}`),
		func(result map[string]interface{}) {
			logger.Infof("login success: =>  %s", result)
		},
		func(code int, err string) {
			logger.Infof("login reject: %d => %s", code, err)
		})
	peer.Request("offer", JsonEncode(`{"sdp":"empty"}`),
		func(result map[string]interface{}) {
			logger.Infof("offer success: =>  %s", result)
		},
		func(code int, err string) {
			logger.Infof("offer reject: %d => %s", code, err)
		})
	/*
		peer.Request("join", JsonEncode(`{"client":"aaa", "type":"sender"}`),
			func(result map[string]interface{}) {
				logger.Infof("join success: =>  %s", result)
			},
			func(code int, err string) {
				logger.Infof("join reject: %d => %s", code, err)
			})
		peer.Request("publish", JsonEncode(`{"type":"sender", "jsep":{"type":"offer", "sdp":"111111111111111"}}`),
			func(result map[string]interface{}) {
				logger.Infof("publish success: =>  %s", result)
			},
			func(code int, err string) {
				logger.Infof("publish reject: %d => %s", code, err)
			})
	*/
}

func main() {
	var ws_client = client.NewClient("wss://127.0.0.1:8443/ws?peer="+peerId, handleWebSocketOpen)
	ws_client.ReadMessage()
}
