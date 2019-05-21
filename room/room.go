package room

import (
	"github.com/cloudwebrtc/go-protoo/logger"
	"github.com/cloudwebrtc/go-protoo/peer"
	"github.com/cloudwebrtc/go-protoo/transport"
)

type Room struct {
	peers  map[string]*peer.Peer
	closed bool
	id     string
}

func NewRoom(roomId string) *Room {
	return &Room{
		peers:  make(map[string]*peer.Peer),
		closed: false,
		id:     roomId,
	}
}

func (room *Room) CreatePeer(peerId string, transport *transport.WebSocketTransport) *peer.Peer {
	newPeer := peer.NewPeer(peerId, transport)
	newPeer.On("close", func() {
		delete(room.peers, peerId)
	})
	room.peers[peerId] = newPeer
	return newPeer
}

func (room *Room) GetPeer(peerId string) *peer.Peer {
	if peer, ok := room.peers[peerId]; ok {
		return peer
	}
	return nil
}

func (room *Room) GetPeers() map[string]*peer.Peer {
	return room.peers
}

func (room *Room) ID() string {
	return room.id
}

func (room *Room) HasPeer(peerId string) bool {
	_, ok := room.peers[peerId]
	return ok
}

func (room *Room) Notify(from *peer.Peer, method string, data map[string]interface{}) {
	for id, peer := range room.peers {
		//send to other peers
		if id != from.ID() {
			peer.Notify(method, data)
		}
	}
}

func (room *Room) Close() {
	logger.Warnf("Close all peers !")
	for id, peer := range room.peers {
		logger.Warnf("Close => peer(%s).", id)
		peer.Close()
	}
	room.closed = true
}
