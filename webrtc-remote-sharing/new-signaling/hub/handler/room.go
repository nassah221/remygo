package handler

import (
	"fmt"
)

type Room struct {
	id    string  // Room id with the same value of the peer session token
	peers []*Peer // Slice of other peers that have joined the room
}

// Create a new room with the peer's session token and add the peer to the room
func newRoom(token string) *Room {
	return &Room{
		id:    token,
		peers: []*Peer{},
	}
}

func (r *Room) addPeer(peer *Peer) {
	r.peers = append(r.peers, peer)
}

// Removes the peer from the room and sets the peer status to empty
func (r *Room) removePeer(p *Peer) error {
	for i, peer := range r.peers {
		if peer.id == p.id {
			p.leaveSession()
			r.peers = append(r.peers[:i], r.peers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("unable to remove peer %s. Not found in room %s", p.id, r.id)
}

// Returns the host peer of the session
func (r *Room) getHost() (*Peer, error) {
	var err error
	for _, peer := range r.peers {
		if peer.host() {
			return peer, err
		}
	}
	err = fmt.Errorf("unable to find session host for room %s", r.id)
	return nil, err
}

func (r *Room) String() string {
	result := fmt.Sprintf("\nRoom %s\n", r.id)
	for _, peer := range r.peers {
		result += fmt.Sprintf("\t%s\n", peer.id)
	}
	result += "\n"

	return result
}
