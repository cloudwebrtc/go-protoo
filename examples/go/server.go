package main

import (
	"encoding/json"
	"net/http"

	"github.com/cloudwebrtc/go-protoo/logger"
	"github.com/cloudwebrtc/go-protoo/room"
	"github.com/cloudwebrtc/go-protoo/server"
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

var testRoom *room.Room

func handleNewWebSocket(transport *transport.WebSocketTransport, request *http.Request) {

	//https://127.0.0.1:8443/ws?peer=alice
	vars := request.URL.Query()
	peerId := vars["peer"][0]

	logger.Infof("handleNewWebSocket peerId => (%s)", peerId)

	peer := testRoom.CreatePeer(peerId, transport)

	handleRequest := func(request map[string]interface{}, accept AcceptFunc, reject RejectFunc) {

		method := request["method"].(string)
		data := request["data"].(map[string]interface{})

		/*handle login and offer request*/
		if method == "login" {
			username := data["username"].(string)
			password := data["password"].(string)
			logger.Infof("Handle login username => %s, password => %s", username, password)
			accept(JsonEncode(`{"status":"login success!"}`))
		} else if method == "offer" {
			sdp := data["sdp"].(string)
			logger.Infof("Handle offer sdp => %s", sdp)
			if sdp == "empty" {
				reject(500, "sdp error!")
			} else {
				accept(JsonEncode(`{}`))
			}
		}

		/*send `kick` request to peer*/
		peer.Request("kick", JsonEncode(`{"reason":"go away!"}`),
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

	handleClose := func(code int, err string) {
		logger.Infof("handleClose => peer (%s) [%d] %s", peer.ID(), code, err)
	}

	peer.On("request", handleRequest)
	peer.On("notification", handleNotification)
	peer.On("close", handleClose)
}

func main() {
	testRoom = room.NewRoom("room1")
	protooServer := server.NewWebSocketServer(handleNewWebSocket)
	config := server.DefaultConfig()
	config.CertFile = "../certs/cert.pem"
	config.KeyFile = "../certs/key.pem"
	protooServer.Bind(config)
}
