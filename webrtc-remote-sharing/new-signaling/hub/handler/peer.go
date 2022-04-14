package handler

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/remygo/pkg/message"

	"golang.org/x/time/rate"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// // TODO: Implement read-write deadline and ping-pong handlers
// // TODO: Implement sessionToken handling - should be assigned by the hub or sent by the peer?

type Peer struct {
	id           string          // Peer id
	conn         *websocket.Conn // Websocket connection
	status       string          // Current session status - manager uses RWMutex to protect mutation
	sessionToken string          // Session token the peer joins with - interchangeably used with 'room id'
	rateLimiter  *rate.Limiter   // Rate limiter for writing to the peer
	// recvCh       chan *message.Message // Channel for the peer to pass on messages to the hub
	// sendCh       chan *message.Message // Channel for the hub to pass messages to for writing to socket
	// stopCh chan struct{} // Channel to signal that peer's serveWs() should return
	m        *Manager // Pointer to the manager
	userID   string
	deviceID string
}

func newPeer(id string, conn *websocket.Conn, m *Manager) *Peer {
	return &Peer{
		id:     id,
		conn:   conn,
		status: "",
		// stopCh:      make(chan struct{}),
		rateLimiter: rate.NewLimiter(rate.Every(time.Millisecond*100), 1),
		m:           m,
	}
}

func (p *Peer) Start(ctx context.Context, wg *sync.WaitGroup) {
	// Caller waits(wg.Wait) after this function call for return
	wg.Add(1)
	go p.readPump(ctx, wg)
}

// // TODO: Implement clean connection break & graceful shutdown in case of interruption
func (p *Peer) readPump(ctx context.Context, wg *sync.WaitGroup) {
	log.Println("[WS] Starting reader for peer:", p.id)
	defer func() {
		log.Printf("[PEER] Closing reader for peer: %s", p.id)
		wg.Done()
	}()

	for {
		var msg message.Message
		err := p.rateLimiter.Wait(ctx)
		if err != nil {
			log.Panicf("[ERR] Rate limiter error: %q", err)
		}

		if err := wsjson.Read(ctx, p.conn, &msg); err != nil {
			log.Println("[ERR] Reading message from peer:", p.id, websocket.CloseStatus(err))
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusAbnormalClosure {
				log.Printf("[PEER] Peer %s closed the connection", p.id)
				break
			}
			break
		}

		// Annotate the message with the sender id
		msg.From = p.id
		p.handleIncomingMessage(ctx, msg)
	}
	log.Println("[WS] Closing reader for peer:", p.id)
}

func (p *Peer) send(ctx context.Context, msg *message.Message) error {
	return wsjson.Write(ctx, p.conn, msg)
}

// Returns true if the peer status is not an empty string i.e. peer is in a session
func (p *Peer) inRoom() bool {
	return p.status != ""
}

// Sets the peer's status to the given session token
func (p *Peer) joinSession(sessionToken string) {
	p.status = sessionToken
}

// Sets the peer's status to an empty string to reflect peer is not in a session
func (p *Peer) leaveSession() {
	p.status = ""
}

// Returns true if the session is hosted by the peer i.e. peer has joined own session peer.status == p.sessionToken
func (p *Peer) host() bool {
	return p.status == p.sessionToken
}
