package handler

import (
	"fmt"

	"github.com/google/uuid"
)

func newSessionToken() string {
	return uuid.New().String()
}

// Get the peer with the given peer id
// func (m *Manager) getPeerWithID(pid string) (*Peer, error) {
// 	if p, ok := m.peers[pid]; ok {
// 		return p, nil
// 	}
// 	// Sanity check for non-existent peer
// 	return nil, fmt.Errorf("non-existent peer %s", pid)
// }

// Get the room with the given session token
func (m *Manager) getRoomByToken(token string) (*Room, error) {
	if r, ok := m.rooms[token]; ok {
		return r, nil
	}
	// Sanity check for non-existent room
	return nil, fmt.Errorf("non-existent room %s", token)
}

// Get the peer with the given session token
func (m *Manager) getPeerByToken(sessionToken string) (*Peer, error) {
	if p, ok := m.sessions[sessionToken]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("non-exitent session token %s", sessionToken)
}

// Get the peer's own room - by its own session token
func (p *Peer) getOwnRoom() (*Room, error) {
	if r, ok := p.m.rooms[p.sessionToken]; ok {
		return r, nil
	}
	// Sanity check for non-existent room
	return nil, fmt.Errorf("non-existent room %s", p.sessionToken)
}

// Get the room the peer has joined - by its status
func (p *Peer) getJoinedRoom() (*Room, error) {
	if r, ok := p.m.rooms[p.status]; ok {
		return r, nil
	}
	// Sanity check for non-existent room
	return nil, fmt.Errorf("non-existent room %s", p.status)
}
