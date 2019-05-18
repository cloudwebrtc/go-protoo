package main

import (
	"encoding/json"
	"net/http"

	"protoo/server"
	"protoo/logger"
	"protoo/transport"
	"protoo/room"
)

func JsonEncode(str string) map[string]interface{}{
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

	vars := request.URL.Query()
	peerId := vars["peer-id"][0]

	peer := testRoom.CreatePeer(peerId, transport)

	handleRequest := func(request map[string]interface{}, accept AcceptFunc, reject RejectFunc) {

		method := request["method"]

		/*handle login and offer reequest*/
		if(method == "login") {
			accept(JsonEncode(`{"name":"xxxx","status":"login"}`))
		} else if(method == "offer") {
			reject(500,"sdp error!");
		}

		/*send `kick` request to peer*/
		peer.Request("kick", JsonEncode(`{"name":"xxxx","why":"I don't like you"}`),
		func(result map[string]interface{}) {
			logger.Infof("kick success: =>  %s", result);
		},
		func(code int, err string) {
			logger.Infof("kick reject: %d => %s", code, err);
		})
	}


	handleNotification := func(notification map[string]interface{}) {
		logger.Infof("handleNotification => %s", notification["method"]);
	}

	peer.On("request", handleRequest)
	peer.On("notification", handleNotification)
}

func main() {
	testRoom = room.NewRoom()
	protooServer := server.NewWebSocketServer(handleNewWebSocket)
	protooServer.Bind("0.0.0.0", "8443")
}