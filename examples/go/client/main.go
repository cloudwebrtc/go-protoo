package main

import (
	"encoding/json"
	"log"

	"github.com/cloudwebrtc/go-protoo/client"
	"github.com/cloudwebrtc/go-protoo/logger"
	"github.com/cloudwebrtc/go-protoo/peer"
	"github.com/cloudwebrtc/go-protoo/transport"
)

var peerId = "go-client-id-xxxx"

func handleWebSocketOpen(con *transport.WebSocketTransport) {
	logger.Infof("handleWebSocketOpen")

	pr := peer.NewPeer(peerId, con)

	handleRequest := func(request peer.Request, accept peer.RespondFunc, reject peer.RejectFunc) {
		method := request.Method
		logger.Infof("handleRequest =>  (%s) ", method)
		if method == "kick" {
			reject(486, "Busy Here")
		} else {
			accept(nil)
		}
	}

	handleNotification := func(notification peer.Notification) {
		logger.Infof("handleNotification => %s", notification.Method)
	}

	handleClose := func(err transport.TransportErr) {
		logger.Infof("handleClose => peer (%s) [%d] %s", pr.ID(), err.Code, err.Text)
	}

	go func() {
		for {
			select {
			case msg := <-pr.OnNotification:
				log.Println(msg)
				handleNotification(msg)
			case msg := <-pr.OnRequest:
				handleRequest(msg.Request, msg.Accept, msg.Reject)
			case msg := <-pr.OnClose:
				handleClose(msg)
			}
		}
	}()

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
		pr.Request("join", json.RawMessage(`{"client":"aaa", "type":"sender"}`),
			func(result json.RawMessage) {
				logger.Infof("join success: =>  %s", result)
			},
			func(code int, err string) {
				logger.Infof("join reject: %d => %s", code, err)
			})
		pr.Request("publish", json.RawMessage(`{"type":"sender", "jsep":{"type":"offer", "sdp":"111111111111111"}}`),
			func(result json.RawMessage) {
				logger.Infof("publish success: =>  %s", result)
			},
			func(code int, err string) {
				logger.Infof("publish reject: %d => %s", code, err)
			})
	*/
}

func main() {
	client.NewClient("ws://127.0.0.1:9090/ws?peer="+peerId, handleWebSocketOpen)
	for {
	}
}
