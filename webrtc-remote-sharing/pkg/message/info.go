package message

import (
	"encoding/json"
	"log"
)

type infoType uint8

const (
	Token infoType = iota
	Register
	Renew
	Ack
	Error
)

type InfoMessage struct {
	Type     infoType `json:"event"`
	Data     string   `json:"data"`
	UserID   string   `json:"userID,omitempty"`
	DeviceID string   `json:"deviceID,omitempty"`
}

// Returns a new message of the 'Info' type. Info messages are used to communicate
// auxiliary information to and from the signaling server. In case of the 'Register'
// message, the first argument is the userID and the second argument is the deviceID.
func NewInfo(t infoType, data string, args ...string) *Message {
	if t == Register {
		if len(args) < 2 || len(args) > 2 {
			log.Panicf("register message requires a userID and deviceID")
		}
		registerMsg, err := json.Marshal(&InfoMessage{Type: t, Data: data, UserID: args[0], DeviceID: args[1]})
		if err != nil {
			log.Panicf("error marshalling info message. %v", err)
		}
		return &Message{Type: Info, Data: registerMsg}
	}

	msg, err := json.Marshal(&InfoMessage{Type: t, Data: data})
	if err != nil {
		log.Panicf("error marshalling message. %v", err)
	}
	return &Message{Type: Info, Data: msg}
}

func (i InfoMessage) String() string {
	switch i.Type {
	case Error:
		return "Error"
	case Register:
		return "Register"
	case Token:
		return "Token"
	case Ack:
		return "Ack"
	case Renew:
		return "Renew"
	default:
		return Unsupported
	}
}

func (i infoType) String() string {
	switch i {
	case Token:
		return "Token"
	case Register:
		return "Register"
	case Ack:
		return "Ack"
	case Error:
		return "Error"
	case Renew:
		return "Renew"
	default:
		return Unsupported
	}
}
