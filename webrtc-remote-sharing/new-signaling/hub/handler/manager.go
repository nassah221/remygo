package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/remygo/pkg/message"

	"nhooyr.io/websocket"
)

type RequestID string

type JoinRequest struct {
	ID        RequestID
	Sender    string
	Recipient string
	Status    string
}

type Manager struct {
	peers       map[string]*Peer
	rooms       map[string]*Room
	sessions    map[string]*Peer
	apiCallChan chan APICall
	requests    map[RequestID]*JoinRequest
	// recvChan chan *message.Message
	mux sync.RWMutex
}

func NewManager(apiChan chan APICall) *Manager {
	return &Manager{
		peers:       make(map[string]*Peer),
		rooms:       make(map[string]*Room),
		sessions:    make(map[string]*Peer),
		requests:    make(map[RequestID]*JoinRequest),
		apiCallChan: apiChan,
		// recvChan: make(chan *message.Message),
		mux: sync.RWMutex{},
	}
}

func (m *Manager) ServeWs(w http.ResponseWriter, r *http.Request) {
	fmt.Println()
	log.Printf("[WS] Incoming connection: %s", r.RemoteAddr)

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{Subprotocols: []string{"signaling"}})
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	// if c.Subprotocol() != "signaling" {
	// 	c.Close(websocket.StatusPolicyViolation, "status policy not supported")
	// }

	p, err := m.registerPeer(c, r.RemoteAddr)
	if err != nil {
		log.Printf("[HUB] Error registering peer %s. %v", p.id, err)
		return
	}
	log.Printf("[HUB] Peer registered: %s", p.id)

	// ctx := r.Context()
	var wg sync.WaitGroup

	// Blocking call - wait for the readPump and writePump to return then remove the peer
	p.Start(context.TODO(), &wg)
	wg.Wait()

	log.Printf("[HUB] Peer %s disconnected", p.id)

	if err = p.removePeer(context.TODO()); err != nil {
		log.Printf("[HUB] Error removing peer %s. %v", p.id, err)
		return
	}
}

// Runs in a goroutine, routes peer messages and services sessions
func (p *Peer) handleIncomingMessage(ctx context.Context, msg message.Message) {
	// Log the incoming message information for debugging purposes
	logMessage(&msg)

	switch msg.Type {
	case message.Signal:
		// Broadcast message to the other peer in the room
		if err := p.handleSignaling(ctx, &msg); err != nil {
			log.Printf("%v", err)
		}

	case message.Session:
		if err := p.handleSession(ctx, &msg); err != nil {
			log.Printf("%v", err)
		}

	case message.Command:
		// TODO: Implement command handling on the server side

	case message.Info:
		if err := p.handleInfo(ctx, &msg); err != nil {
			log.Printf("%v", err)
		}
	}
}

func (p *Peer) handleSession(ctx context.Context, msg *message.Message) error {
	p.m.mux.Lock()
	defer p.m.mux.Unlock()

	var sessionMessage *message.SessionMessage
	if err := json.Unmarshal(msg.Data, &sessionMessage); err != nil {
		return fmt.Errorf("[HUB] Failed to parse session message: %v", err)
	}
	sessionToken := sessionMessage.Token

	switch sessionMessage.Type {
	case message.JoinResponse:
		switch sessionMessage.Response {
		case message.Allow:
			if !p.inRoom() {
				room, err := p.m.getRoomByToken(sessionToken)
				if err != nil {
					return fmt.Errorf("[HUB] Error fetching room. %v", err)
				}

				log.Printf("[HUB] Host %s has allowed session join request %s from remote peer", p.id, sessionToken)

				if req, ok := p.m.requests[RequestID(sessionToken)]; ok {
					p.joinSession(sessionToken)
					room.addPeer(p)

					if remote, ok := p.m.peers[req.Sender]; ok {
						remote.joinSession(sessionToken)
						room.addPeer(remote)

						remote.send(ctx, message.NewInfo(message.Ack, fmt.Sprintf("Session Join Request %s ALLOWED", sessionToken)))
						remote.send(ctx, message.NewCommand(message.InitiateSession))

						delete(p.m.requests, RequestID(sessionToken))

						// p.m.apiCallChan <- APICall{Type: JoinSession, UserID: p.id, DeviceID: p.deviceID, SessionToken: sessionToken}
						return nil
					}
					return fmt.Errorf("[HUB] Error fetching remote peer %s", req.Sender)
				}
				return fmt.Errorf("[HUB] Request ID not found for session join response sent by peer %s", p.id)
			}
			log.Printf("[WARN] Should not happen. Peer %s sent session message and already in session: %v", p.id, sessionMessage)
		case message.Deny:
			if !p.inRoom() {
				if req, ok := p.m.requests[RequestID(sessionToken)]; ok {
					if recipient, ok := p.m.sessions[req.Sender]; ok {
						recipient.send(ctx, message.NewInfo(message.Error, fmt.Sprintf("Session Join Request %s Denied", sessionToken)))
						delete(p.m.requests, RequestID(sessionToken))
						return nil
					}
					return fmt.Errorf("[HUB] Unable to fetch peer %s", req.Sender)
				}
				return fmt.Errorf("[HUB] Request ID not found for session join response sent by peer %s", p.id)
			}
		}
		log.Printf("[WARN] Should not happen. Peer %s sent session message and already in session: %v", p.id, sessionMessage)
	case message.JoinRequest:
		// A peer is intending to join another peer's room through the session token (room id)

		// Get the session token(room id) that the peer wants to join with
		// Check for empty string
		if sessionToken == "" {
			return fmt.Errorf("[HUB] No room specified in session message")
		}

		// If the session with the supplied token exits
		if _, ok := p.m.sessions[sessionToken]; ok {
			// Check peer status Proceed if empty
			if !p.inRoom() {
				// r, err := p.m.getRoomByToken(sessionToken)
				// if err != nil {
				// 	return fmt.Errorf("[HUB] Error fetching room. %v", err)
				// }

				// Get the host peer
				host, err := p.m.getPeerByToken(sessionToken)
				if err != nil {
					return fmt.Errorf("[HUB] Error fetching peer with session token. %v", err)
				}

				// Early return if the host peer is already in a session
				//? This ensures that the host peer can only join a room once per session
				//? and any requests to join a room which is already in a session will be end in the sender being disconnected
				if host.inRoom() {
					log.Printf("\n\n[HUB] Host %s peer is already in a session. Terminating peer %s\n\n", host.id, p.id)

					errorMsg := message.NewInfo(message.Error, "Peer already in room")
					p.send(ctx, errorMsg)
					return nil
				}

				// Check pending requests. If it's a new request, add it to the pending requests
				if _, ok := p.m.requests[RequestID(sessionToken)]; !ok {
					p.m.requests[RequestID(sessionToken)] = &JoinRequest{
						ID:        RequestID(sessionToken),
						Sender:    p.id,
						Recipient: host.id,
						Status:    "pending",
					}
					log.Printf("[HUB] Peer %s sent session join request to peer %s: %s\n", p.id, host.id, sessionToken)
					// Send the request to the host peer
					joinRequest := message.NewSession(message.JoinRequest, sessionToken, nil)
					host.send(ctx, joinRequest)

					return nil
				}
				// host.joinSession(sessionToken)
				// r.addPeer(host)

				// // Sanity check that the peer is indeed in its own room
				// if host.status != host.sessionToken {
				// 	log.Panicf("[HUB] Error asserting recipient peer status and session token")
				// }

				// p.joinSession(sessionToken)
				// r.addPeer(p)

				// //^ In this instance, the 'InitiateSession' message indicates that the peer should send an offer
				// //^ As soon as the peer receives this message, it will send an offer which is to be forwarded
				// //^ to the other peer in the session
				// initMsg := message.NewCommand(message.InitiateSession)
				// log.Printf("\n\nROOM: %v\n\n", r.String())

				// ackMsg := message.NewInfo(message.Ack, fmt.Sprintf("SESSION %s OK", sessionToken))
				// log.Printf("\n\nROOM: %v\n\n", r.String())

				// // Send acknowledge message to the peer that sent the join session message
				// p.send(ctx, ackMsg)

				// // The joining peer is always going to be the remote client which will send the offer
				// // to initate the session
				// p.send(ctx, initMsg)

				// p.m.apiCallChan <- APICall{Type: JoinSession, UserID: p.id, DeviceID: p.deviceID, SessionToken: sessionToken}
			}
		} else {
			errorMsg := message.NewInfo(message.Error, "Invalid session token")
			p.send(ctx, errorMsg)
		}
	case message.Leave:
		// if sessionToken == "" {
		// 	return fmt.Errorf("[HUB] No room specified in session message")
		// }
		if p.inRoom() {
			if !p.host() {
				log.Printf("[HUB] Remote peer %s is leaving the session\n", p.id)
				if err := p.sessionCleanup(ctx); err != nil {
					return fmt.Errorf("error removing peer %s from joined session %s. %v", p.id, p.status, err)
				}
				p.renewSessionToken(ctx)
			}
		}

		// Get the peer
		// p, err := m.getPeerWithID(msg.From)
		// if err != nil {
		// 	return fmt.Errorf("[HUB] Error fetching peer. %v", err)
		// }
	}

	return nil
}

func (p *Peer) handleInfo(ctx context.Context, msg *message.Message) error {
	p.m.mux.Lock()
	defer p.m.mux.Unlock()

	var tokenMsg message.InfoMessage
	if err := json.Unmarshal([]byte(msg.Data), &tokenMsg); err != nil {
		log.Panicf("[ERR] Unmarshalling info message. %v", err)
	}
	switch tokenMsg.Type {
	case message.Register:
		// Store the user & device ids for logging events with the rest api
		p.userID = tokenMsg.UserID
		p.deviceID = tokenMsg.DeviceID

		if tokenMsg.Data == "" {
			p.sessionToken = newSessionToken()
			log.Printf("[HUB] Peer supplied an empty session token. Assigning a new session token %s", p.sessionToken)
		} else {
			p.sessionToken = tokenMsg.Data
		}

		// Add peer session to the sessions map
		p.m.sessions[p.sessionToken] = p

		// Create a room for the peer session
		p.m.rooms[p.sessionToken] = newRoom(p.sessionToken)

		// Send the peer their assigned session token
		tokenMsg := message.NewInfo(message.Token, p.sessionToken)
		fmt.Printf("\n\n[HUB] -> Peer %s: Session: %s\nMSG:%#+v\n", p.id, p.sessionToken, tokenMsg)
		if err := p.send(ctx, tokenMsg); err != nil {
			log.Panicf("[ERR] sending message to socket: %q", err)
		}

		// p.m.apiCallChan <- APICall{Type: CreateSession, UserID: p.userID, DeviceID: p.deviceID, SessionToken: p.sessionToken}
	}

	return nil
}

// Route signaling message to the recipient in the session to commence webRTC connection
func (p *Peer) handleSignaling(ctx context.Context, msg *message.Message) error {
	// For now, there will only be a pair of peers namely "remote" & "host"
	// however, this may not be a hard-constrain and might be extended to multiple peers
	// in which case the webrtc communication happens in a mesh topology i.e. this won't work for
	// the SFU | MCU case
	p.m.mux.RLock()
	defer p.m.mux.RUnlock()

	// Check peer status and send the message if the peer is in a room
	if p.inRoom() {
		if r, err := p.getJoinedRoom(); err == nil {
			for _, recipient := range r.peers {
				if recipient.id != p.id {
					log.Printf("[HUB] Sending message to peer %s", recipient.id)
					recipient.send(ctx, msg)
				}
			}
			return nil
		}
		return fmt.Errorf("[HUB] Room %s not found. Ignoring message", p.status)
	}
	return fmt.Errorf("[HUB] Peer %s not in a room. Ignoring message", p.id)
}

func (m *Manager) registerPeer(conn *websocket.Conn, pid string) (*Peer, error) {
	m.mux.Lock()
	defer m.mux.Unlock()

	// Sanity check for duplicate peer
	if _, ok := m.peers[pid]; ok {
		return nil, fmt.Errorf("duplicate peer %s", pid)
	}

	// Create a new peer and generate a guid for it's session token
	p := newPeer(pid, conn, m)

	// Insert peer into the peers map
	m.peers[pid] = p

	return p, nil
}

func (p *Peer) renewSessionToken(ctx context.Context) {
	log.Println("[HUB] Renewing session token for peer", p.id)
	newToken := newSessionToken()

	p.m.rooms[newToken] = p.m.rooms[p.sessionToken]
	delete(p.m.rooms, p.sessionToken)
	log.Printf("[HUB] Updated rooms record. Old token: %s -> New token: %s", p.sessionToken, newToken)
	p.m.sessions[newToken] = p.m.sessions[p.sessionToken]
	delete(p.m.sessions, p.sessionToken)
	log.Printf("[HUB] Updated sessions record. Old token: %s -> New token: %s", p.sessionToken, newToken)
	p.sessionToken = newToken

	p.send(context.Background(), message.NewInfo(message.Renew, newToken))
}

func (p *Peer) sessionCleanup(ctx context.Context) error {
	log.Printf("[HUB] Peer %s in session %s is leaving the room", p.id, p.sessionToken)

	if p.host() {
		r, err := p.getOwnRoom()
		if err != nil {
			return fmt.Errorf("error fetching peer's %s room %s. %v", p.id, p.sessionToken, err)
		}
		log.Printf("[HUB] Peer %s is the session host. Removing other peer from the room\n", p.id)
		for _, recipient := range r.peers {
			// Send terminate session command to all peers in the session except the host peer
			if recipient.id != p.id {
				log.Printf("[HUB] Removing peer %s from the room", recipient.id)
				if err = recipient.send(ctx, message.NewCommand(message.TerminateSession)); err != nil {
					// // TODO: Error sending a message to a peer should terminate their websocket connection through ctx.Cancel
					return fmt.Errorf("error sending terminate session command to peer %s. %v", recipient.id, err)
				}
				// recipient.m.apiCallChan <- APICall{Type: EndSession, UserID: recipient.userID,
				// 	DeviceID: recipient.deviceID, SessionToken: recipient.sessionToken}

				// Remove the peer from the room map
				if err = r.removePeer(recipient); err != nil {
					return fmt.Errorf("error removing peer %s from room %s after host %s left the session. %v",
						recipient.id, r.id, p.id, err)
				}
			}
		}
		// Remove the peer from its own room
		log.Printf("[HUB] Removing peer %s from its own room\n", p.id)
		if err = r.removePeer(p); err != nil {
			return fmt.Errorf("error removing peer %s from room %s after host %s left the session. %v",
				p.id, r.id, p.id, err)
		}
		p.renewSessionToken(ctx)

		return nil
	}

	// If the peer is not the host, remove the peer from the room map and
	// send session terminate command to the host peer

	// Send terminate session command to the host peer
	r, err := p.getJoinedRoom()
	log.Printf("[HUB] Session cleanup for room %s\n", r.String())
	if err != nil {
		return fmt.Errorf("error fetching host session room %s. %v", p.status, err)
	}

	log.Printf("[HUB] Remote peer %s is leaving the room. Informing host to terminate session", p.id)
	host, err := r.getHost()
	if err != nil {
		log.Panicf("[HUB] Error fetching host peer from room %s. %v", r.id, err)
	}
	host.send(ctx, message.NewCommand(message.TerminateSession))
	defer host.sessionCleanup(ctx)

	// End session device
	// p.m.apiCallChan <- APICall{Type: LeaveSession, UserID: p.userID,
	// 	DeviceID: p.deviceID, SessionToken: p.status}

	// // End session
	// p.m.apiCallChan <- APICall{Type: EndSession, UserID: p.userID,
	// 	DeviceID: p.deviceID, SessionToken: p.sessionToken}

	// Remove the remote peer from the room map
	log.Printf("[HUB] Removing peer %s from the room", p.id)
	return r.removePeer(p)
}

// Removes the peer when the peer is disconnected i.e. socket connection is closed
func (p *Peer) removePeer(ctx context.Context) error {
	p.m.mux.Lock()
	defer p.m.mux.Unlock()

	var err error

	log.Printf("[HUB] Removing peer: %s", p.id)

	if p.inRoom() {
		if err = p.sessionCleanup(ctx); err != nil {
			err = fmt.Errorf("error cleaning up peer %s session. %v", p.id, err)
		}
	}

	// Remove the peer's room from the rooms map
	delete(p.m.rooms, p.sessionToken)

	// Delete the session token from the sessions map
	delete(p.m.sessions, p.sessionToken)

	// Delete peer from the peers map. Every record of the peer should've been removed at this point
	delete(p.m.peers, p.id)

	return err
}

// Log the message received from a peer
func logMessage(msg *message.Message) {
	switch msg.Type {
	case message.Signal:
		var signal message.SignalMessage
		if err := json.Unmarshal(msg.Data, &signal); err != nil {
			log.Printf("[ERR] Unmarshalling websocket message: %v", err)
			return
		}
		log.Printf("[HUB] Peer message from %s. Type: %s", msg.From, signal.String())
	case message.Session:
		var session message.SessionMessage
		if err := json.Unmarshal(msg.Data, &session); err != nil {
			log.Printf("[ERR] Unmarshalling websocket message: %v", err)
			return
		}
		log.Printf("[HUB] Peer message from %s. Type: %s", msg.From, session.String())
	case message.Command:
		var command message.CommandMessage
		if err := json.Unmarshal(msg.Data, &command); err != nil {
			log.Printf("[ERR] Unmarshalling websocket message: %v", err)
			return
		}
		log.Printf("[HUB] Peer message from %s. Type: %s", msg.From, command.String())
	case message.Info:
		var info message.InfoMessage
		if err := json.Unmarshal(msg.Data, &info); err != nil {
			log.Printf("[ERR] Unmarshalling websocket message: %v", err)
			return
		}
		log.Printf("[HUB] Peer message from %s. Type: %s", msg.From, info.String())
	default:
		log.Printf("[HUB] Unknown message from %s: %v", msg.From, msg.Type)
	}
}
