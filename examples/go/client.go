package main

import (
	"encoding/json"

	"github.com/cloudwebrtc/go-protoo/client"
	"github.com/cloudwebrtc/go-protoo/logger"
	"github.com/cloudwebrtc/go-protoo/peer"
	"github.com/cloudwebrtc/go-protoo/transport"
)

var peerId = "go-client-id-xxxx"

func handleWebSocketOpen(transport *transport.WebSocketTransport) {
	logger.Infof("handleWebSocketOpen")

	pr := peer.NewPeer(peerId, transport)
	pr.On("close", func(code int, err string) {
		logger.Infof("peer close [%d] %s", code, err)
	})

	handleRequest := func(request peer.Request, accept peer.RespondFunc, reject peer.RejectFunc) {
		method := request.Method
		logger.Infof("handleRequest =>  (%s) ", method)
		if method == "kick" {
			reject(486, "Busy Here")
		} else {
			accept(nil)
		}
	}

	handleNotification := func(notification map[string]interface{}) {
		logger.Infof("handleNotification => %s", notification["method"])
	}

	handleClose := func(code int, err string) {
		logger.Infof("handleClose => peer (%s) [%d] %s", pr.ID(), code, err)
	}

	pr.On("request", handleRequest)
	pr.On("notification", handleNotification)
	pr.On("close", handleClose)

	pr.Request("login", json.RawMessage(`{"username":"alice","password":"alicespass"}`),
		func(result json.RawMessage) {
			logger.Infof("login success: =>  %s", result)
		},
		func(code int, err string) {
			logger.Infof("login reject: %d => %s", code, err)
		})
	pr.Request("offer", json.RawMessage(`{"sdp":"empty"}`),
		func(result json.RawMessage) {
			logger.Infof("offer success: =>  %s", result)
		},
		func(code int, err string) {
			logger.Infof("offer reject: %d => %s", code, err)
		})
	/*
		pr.Request("join", JsonEncode(`{"client":"aaa", "type":"sender"}`),
			func(result map[string]interface{}) {
				logger.Infof("join success: =>  %s", result)
			},
			func(code int, err string) {
				logger.Infof("join reject: %d => %s", code, err)
			})
		pr.Request("publish", JsonEncode(`{"type":"sender", "jsep":{"type":"offer", "sdp":"111111111111111"}}`),
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
