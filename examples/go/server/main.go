package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/cloudwebrtc/go-protoo/logger"
	"github.com/cloudwebrtc/go-protoo/peer"
	"github.com/cloudwebrtc/go-protoo/room"
	"github.com/cloudwebrtc/go-protoo/server"
	"github.com/cloudwebrtc/go-protoo/transport"
)

var testRoom *room.Room

type KickMsg struct {
	Reason string `json:"reason"`
}

type LoginMsg struct {
	Status string `json:"status"`
}

type Request struct {
	Method string
	Data   json.RawMessage
}

type LoginData struct {
	Username string
	Password string
}

type OfferData struct {
	Sdp string `json:"sdp"`
}

func handleNewWebSocket(transport *transport.WebSocketTransport, request *http.Request) {
	log.Println("Handle socket")

	//https://127.0.0.1:8443/ws?peer=alice
	vars := request.URL.Query()
	peerId := vars["peer"][0]

	logger.Infof("handleNewWebSocket peerId => (%s)", peerId)

	pr := testRoom.CreatePeer(peerId, transport)

	handleRequest := func(request peer.Request, accept peer.RespondFunc, reject peer.RejectFunc) {
		method := request.Method

		/*handle login and offer request*/
		if method == "login" {
			var data LoginData
			if err := json.Unmarshal(request.Data, &data); err != nil {
				log.Fatal("Marshal error")
			}
			username := data.Username
			password := data.Password
			logger.Infof("Handle login username => %s, password => %s", username, password)
			accept(LoginMsg{Status: "login success!"})
		} else if method == "offer" {
			var data OfferData
			if err := json.Unmarshal(request.Data, &data); err != nil {
				log.Fatal("Marshal error")
			}
			sdp := data.Sdp
			logger.Infof("Handle offer sdp => %s", sdp)
			if sdp == "empty" {
				reject(500, "sdp error!")
			} else {
				accept(nil)
			}
		}

		/*send `kick` request to peer*/
		pr.Request("kick", KickMsg{Reason: "go away!"},
			func(result json.RawMessage) {
				logger.Infof("kick success: =>  %s", result)
				// close transport
				pr.Close()
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
		testRoom.Notify(pr, method, data)
	}

	handleClose := func(code int, err string) {
		logger.Infof("handleClose => peer (%s) [%d] %s", pr.ID(), code, err)
	}

	_, _, _ = handleRequest, handleNotification, handleClose

	for {
		select {
		case msg := <-pr.OnNotification:
			log.Println("OnNotification msg", msg)
			// handleNotification
		case msg := <-pr.OnRequest:
			handleRequest(msg.Request, msg.Accept, msg.Reject)
			// log.Println(msg)
		case msg := <-pr.OnClose:
			log.Println("Close msg", msg)
		}
	}
}

func main() {
	testRoom = room.NewRoom("room1")
	protooServer := server.NewWebSocketServer(handleNewWebSocket)
	config := server.DefaultConfig()
	config.Port = 9090
	// config.CertFile = "../certs/cert.pem"
	// config.KeyFile = "../certs/key.pem"
	protooServer.Bind(config)
}
