package message

import (
	"encoding/json"
	"log"
)

type commandType uint8

const (
	InitiateSession commandType = iota
	TerminateSession
)

type CommandMessage struct {
	Type commandType `json:"event"`
	// Token string `json:"token"`
}

func NewCommand(t commandType) *Message {
	cmd, err := json.Marshal(&CommandMessage{Type: t})
	if err != nil {
		log.Panicf("error marshalling message. %v", err)
	}
	return &Message{Type: Command, Data: cmd}
}

func (c CommandMessage) String() string {
	switch c.Type {
	case InitiateSession:
		return "InitiateSession"
	case TerminateSession:
		return "TerminateSession"
	default:
		return Unsupported
	}
}
