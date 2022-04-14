package message

import (
	"encoding/json"
)

type Type uint8

// Sum type for websocket messages
const (
	Session Type = iota
	Signal
	Command
	Info
	API
)

const Unsupported = "Unsupported"

// Parent message type which is transmitted over the socket connection
// Server-Clients are expected to deserialize messages into this type and inspect the 'Type'
// upon which they can further determine which type to deserialize 'Data' into
type Message struct {
	Type Type            `json:"event"`          // Type of message
	From string          `json:"from,omitempty"` // Signaling server annotates this field with the sender id
	Data json.RawMessage `json:"data"`           // Payload of the underlying message sum types
}

func (t Type) String() string {
	switch t {
	case Session:
		return "Session"
	case Signal:
		return "Signal"
	case Command:
		return "Command"
	case Info:
		return "Info"
	case API:
		return "API"
	default:
		return Unsupported
	}
}
