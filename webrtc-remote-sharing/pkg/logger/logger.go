package logger

import (
	"encoding/json"
	"log"

	"github.com/remygo/pkg/message"
)

// Log the message received from a peer
func LogMessage(msg *message.Message, prefix string) {
	switch msg.Type {
	case message.Signal:
		var signal message.SignalMessage
		if err := json.Unmarshal(msg.Data, &signal); err != nil {
			log.Printf("[ERR] Unmarshalling websocket message: %v", err)
			return
		}
		// if signal.Type == message.Answer {
		// 	log.Printf(prefix+"\n[SDP] Answer: %s\n", signal.Data)
		// }
		log.Printf(prefix+"Type: %s", signal.String())
	case message.Session:
		var session message.SessionMessage
		if err := json.Unmarshal(msg.Data, &session); err != nil {
			log.Printf("[ERR] Unmarshalling websocket message: %v", err)
			return
		}
		log.Printf(prefix+"Type: %s", session.String())
	case message.Command:
		var command message.CommandMessage
		if err := json.Unmarshal(msg.Data, &command); err != nil {
			log.Printf("[ERR] Unmarshalling websocket message: %v", err)
			return
		}
		log.Printf(prefix+"Type: %s", command.String())
	case message.Info:
		var info message.InfoMessage
		if err := json.Unmarshal(msg.Data, &info); err != nil {
			log.Printf("[ERR] Unmarshalling websocket message: %v", err)
			return
		}
		log.Printf(prefix+"Type: %s", info.String())
	default:
		log.Printf("[HUB] Unknown message from %s: %v", msg.From, msg.Type)
	}
}
