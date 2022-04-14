package ws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/remygo/pkg/message"

	"golang.org/x/time/rate"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type Socket struct {
	*websocket.Conn
	*rate.Limiter
	From chan<- message.Message
	// To   <-chan message.Message
}

func NewPeer(c *websocket.Conn, from chan message.Message) *Socket {
	return &Socket{
		Conn: c,
		// To:      make(chan message.Message, 2),
		// From:    make(chan message.Message, 2),
		From: from,
		// To:      to,
		Limiter: rate.NewLimiter(rate.Every(time.Millisecond*100), 1),
	}
}

// func (s *Socket) Receive() <-chan message.Message {
// 	return s.From
// }

func (s *Socket) Close() {
	log.Println("[WS] Closing connection")
	if err := s.Conn.Close(websocket.StatusNormalClosure, ""); err != nil {
		log.Fatalf("[ERR] Closing connection: %v", err)
	}
}

func (s *Socket) ReadPump(ctx context.Context) error {
	var err error

	log.Println("[WS] Starting read pump")
loop:
	for {
		select {
		default:
			if err = s.Limiter.Wait(ctx); err != nil {
				log.Panicf("[ERR] Waiting for read: %v", err)
			}
			var msg message.Message

			if readErr := wsjson.Read(ctx, s.Conn, &msg); readErr != nil {
				log.Println("[ERR] Reading message from peer:", websocket.CloseStatus(err))
				if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
					websocket.CloseStatus(err) == websocket.StatusAbnormalClosure {
					return fmt.Errorf("peer closed the connection")
				}
				err = readErr
				break loop
			}
			s.From <- msg
		}
	}
	return err
	// close(s.from)
}

// func (s *Socket) WritePump() error {
// 	var err error
// 	log.Println("[WS] Starting write pump")

// 	for msg := range s.To {
// 		log.Println("[WS] Sending message:", msg.Type.String())
// 		if writeErr := s.write(msg); writeErr != nil {
// 			err = writeErr
// 			break
// 		}
// 	}
// 	return err
// }

func (s *Socket) Write(msg message.Message) error {
	return wsjson.Write(context.TODO(), s.Conn, msg)
}
