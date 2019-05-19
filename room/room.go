package room

import (
	"fmt"
	"protoo/peer"
	"protoo/transport"
)

type Room struct {
	peers  map[string]*peer.Peer
	closed bool
}

func NewRoom() *Room {
	return &Room{
		peers:  make(map[string]*peer.Peer),
		closed: false,
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

func (room *Room) GetPeers() map[string]*peer.Peer {
	return room.peers
}

func (room *Room) HasPeer(peerId string) bool {
	_, ok := room.peers[peerId]
	return ok
}

func (room *Room) Close() {
	for id, peer := range room.peers {
		fmt.Println("Close peer :" + id)
		peer.Close()
	}
}
