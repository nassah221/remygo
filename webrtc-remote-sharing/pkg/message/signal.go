package message

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/pion/webrtc/v3"
)

type signalType uint8

// Variants of the 'signal' websocket message type
const (
	ICE signalType = iota
	Offer
	Answer
)

// Underlying message type for the 'signal' websocket message
type SignalMessage struct {
	Type signalType `json:"event"` // Type of signal message
	Data []byte     `json:"data"`  // Payload of the underlying 'signal' message sum types
}

// Returns a new 'signal' message wrapped in a Message struct
func NewSignal(t signalType, data []byte) *Message {
	msg, err := json.Marshal(&SignalMessage{Type: t, Data: data})
	if err != nil {
		log.Panicf("error marshalling message. %v", err)
	}
	return &Message{Type: Signal, Data: msg}
}

// Unmarshals the message into a webrtc.ICECandidate
func (msg *SignalMessage) IntoICE(candidate *webrtc.ICECandidateInit) error {
	if err := json.Unmarshal([]byte(msg.Data), candidate); err != nil {
		return fmt.Errorf("could not unmarshal: %v", err)
	}

	if candidate.Candidate == "" {
		log.Printf("[WARN] Empty ICECandidate: %v", candidate)
	}
	return nil
}

// Unmarshals the message into a webrtc.SessionDescription
func (msg *SignalMessage) IntoSDP(sdp *webrtc.SessionDescription) error {
	if err := json.Unmarshal([]byte(msg.Data), sdp); err != nil {
		return fmt.Errorf("could not unmarshal: %v", err)
	}

	if sdp.SDP == "" {
		log.Printf("[WARN] Empty SDP: %v", sdp)
	}
	return nil
}

func (msg SignalMessage) String() string {
	switch msg.Type {
	case ICE:
		return "ICECandidate"
	case Offer:
		return "Offer"
	case Answer:
		return "Answer"
	default:
		return Unsupported
	}
}

// package message

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"

// 	"github.com/pion/webrtc/v3"
// )

// type signalType uint8

// // Variants of the 'signal' websocket message type
// const (
// 	ICE signalType = iota
// 	Offer
// 	Answer
// )

// // Underlying message type for the 'signal' websocket message
// type SignalMessage struct {
// 	Type signalType `json:"event"` // Type of signal message
// 	Data []byte     `json:"data"`  // Payload of the underlying 'signal' message sum types
// }

// // Returns a new 'signal' message wrapped in a Message struct
// func NewSignal(t signalType, data []byte) (*Message, error) {
// 	msg, err := json.Marshal(&SignalMessage{Type: t, Data: data})
// 	if err != nil {
// 		return nil, fmt.Errorf("error marshalling message. %v", err)
// 	}
// 	return &Message{Type: Signal, Data: msg}, nil
// }

// // Unmarshals the message into a webrtc.ICECandidate
// func (msg *SignalMessage) IntoICE(candidate *webrtc.ICECandidateInit) error {
// 	if err := json.Unmarshal([]byte(msg.Data), candidate); err != nil {
// 		return fmt.Errorf("could not unmarshal: %v", err)
// 	}

// 	if candidate.Candidate == "" {
// 		log.Printf("[WARN] Empty ICECandidate: %v", candidate)
// 	}
// 	return nil
// }

// // Unmarshals the message into a webrtc.SessionDescription
// func (msg *SignalMessage) IntoSDP(sdp *webrtc.SessionDescription) error {
// 	if err := json.Unmarshal([]byte(msg.Data), sdp); err != nil {
// 		return fmt.Errorf("could not unmarshal: %v", err)
// 	}

// 	if sdp.SDP == "" {
// 		log.Printf("[WARN] Empty SDP: %v", sdp)
// 	}
// 	return nil
// }

// func (msg SignalMessage) String() string {
// 	switch msg.Type {
// 	case ICE:
// 		return "ICECandidate"
// 	case Offer:
// 		return "Offer"
// 	case Answer:
// 		return "Answer"
// 	default:
// 		return Unsupported
// 	}
// }
