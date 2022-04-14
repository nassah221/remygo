package message

import (
	"encoding/json"
	"log"
)

type sessionType uint8

type JoinAnswer uint8

const (
	Allow JoinAnswer = iota
	Deny
)

// Variants of the 'session' websocket message type
const (
	JoinRequest sessionType = iota
	JoinResponse
	Leave
)

// Underlying message type for the 'session' websocket message
type SessionMessage struct {
	Type     sessionType `json:"event"` // Type of session message
	Token    string      `json:"token"` // Token of the session
	Response JoinAnswer  `json:"response,omitempty"`
}

// Returns a new 'session' message wrapped in a Message struct
func NewSession(t sessionType, token string, response *string) *Message {
	if response != nil {
		switch *response {
		case "allow":
			msg, err := json.Marshal(&SessionMessage{Type: t, Token: token, Response: Allow})
			if err != nil {
				log.Panicf("error marshalling message. %v", err)
			}
			return &Message{Type: Session, Data: msg}
		case "deny":
			msg, err := json.Marshal(&SessionMessage{Type: t, Token: token, Response: Deny})
			if err != nil {
				log.Panicf("error marshalling message. %v", err)
			}
			return &Message{Type: Session, Data: msg}
		}
	}
	msg, err := json.Marshal(&SessionMessage{Type: t, Token: token})
	if err != nil {
		log.Panicf("error marshalling message. %v", err)
	}
	return &Message{Type: Session, Data: msg}
}

func (s SessionMessage) String() string {
	switch s.Type {
	case JoinRequest:
		return "Join Request"
	case Leave:
		return "Leave"
	default:
		return Unsupported
	}
}
