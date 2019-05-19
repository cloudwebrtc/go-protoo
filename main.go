package main

import (
	"encoding/json"
	"net/http"

	"protoo/logger"
	"protoo/room"
	"protoo/server"
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

var testRoom *room.Room

func handleNewWebSocket(transport *transport.WebSocketTransport, request *http.Request) {

	//https://127.0.0.1:8443/ws?peer-id=xxxxx&room-id=room1
	vars := request.URL.Query()
	peerId := vars["peer-id"][0]
	//roomId := vars["room-id"][0]

	peer := testRoom.CreatePeer(peerId, transport)

	handleRequest := func(request map[string]interface{}, accept AcceptFunc, reject RejectFunc) {

		method := request["method"]

		/*handle login and offer reequest*/
		if method == "login" {
			accept(JsonEncode(`{"name":"xxxx","status":"login"}`))
		} else if method == "offer" {
			reject(500, "sdp error!")
		}

		/*send `kick` request to peer*/
		peer.Request("kick", JsonEncode(`{"name":"xxxx","why":"I don't like you"}`),
			func(result map[string]interface{}) {
				logger.Infof("kick success: =>  %s", result)
				// close transport
				peer.Close()
			},
			func(code int, err string) {
				logger.Infof("kick reject: %d => %s", code, err)
			})
	}

	handleNotification := func(notification map[string]interface{}) {
		logger.Infof("handleNotification => %s", notification["method"])

		method := notification["method"].(string)
		data := notification["data"].(map[string]interface{})

		//Forward notification to testRoom.
		testRoom.Notify(peer, method, data)
	}

	handleClose := func() {
		logger.Infof("handleClose => peer (%s) ", peer.ID())
	}

	peer.On("request", handleRequest)
	peer.On("notification", handleNotification)
	peer.On("close", handleClose)
}

func main() {
	testRoom = room.NewRoom("room1")
	protooServer := server.NewWebSocketServer(handleNewWebSocket)
	protooServer.Bind("0.0.0.0", "8443", "examples/certs/cert.pem", "examples/certs/key.pem")
}
