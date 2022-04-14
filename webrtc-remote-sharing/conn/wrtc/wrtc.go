package wrtc

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/pion/webrtc/v3"
)

type PeerConn struct {
	*webrtc.PeerConnection
}

// Creates a new peerConnection with STUN config
func newPeerConnection(url, creds string) (peerConnection *webrtc.PeerConnection, err error) {
	if creds != "" {
		split := strings.Split(creds, ":")
		user, pwd := split[0], split[1]
		peerConnection, err = webrtc.NewPeerConnection(webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs:           []string{url},
					Username:       user,
					Credential:     pwd,
					CredentialType: webrtc.ICECredentialTypePassword,
				},
			},
		})
	} else {
		peerConnection, err = webrtc.NewPeerConnection(webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{url},
				},
			},
		})
	}

	return
}

// Returns a peerConnection with receive only transceiver
func NewRemote(url, creds string) (*PeerConn, error) {
	log.Println("[PC] Creating remote connection")

	peerConnection, err := newPeerConnection(url, creds)
	if err != nil {
		return nil, err
	}

	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo,
		webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly}); err != nil {
		log.Printf("[ERR] Adding transceiver: %v", err)
		return nil, err
	}
	return &PeerConn{peerConnection}, err
}

// Returns a peerConnection and video track with the given codec
func NewHost(url, creds, codecName string) (*PeerConn, *webrtc.TrackLocalStaticSample, error) {
	log.Println("[PC] Creating host connection")

	peerConnection, err := newPeerConnection(url, creds)
	if err != nil {
		return nil, nil, err
	}

	log.Printf("[PC] Creating video track with %s", codecName)
	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: codecName}, "video", "host")
	if err != nil {
		log.Printf("[ERR] Creating video track: %v", err)
		return nil, nil, err
	}

	_, err = peerConnection.AddTransceiverFromTrack(videoTrack,
		webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionSendonly})
	if err != nil {
		log.Panicf("[ERR] Adding video track: %v", err)
	}

	return &PeerConn{peerConnection}, videoTrack, err
}

// Connects the OnConnectionStateChange callback to the PeerConnection
func (pc *PeerConn) ConnectionStateCallback() {
	pc.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
		switch p {
		case webrtc.PeerConnectionStateFailed:
			if err := pc.Close(); err != nil {
				log.Printf("[ERR] Connecting to remote: %v", err)
				os.Exit(0)
			}
		case webrtc.PeerConnectionStateClosed:
			log.Println("[PC] Connection closed")
			return
		}
	})
}

// Creates an offer, sets it as the local description and returns the json marshaled bytes
func (pc *PeerConn) NewOffer() ([]byte, error) {
	log.Println("[SDP] Creating offer")
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		return nil, err
	}
	log.Println("[PC] Setting local description")
	if err = pc.SetLocalDescription(offer); err != nil {
		return nil, err
	}
	offerString, err := json.Marshal(offer)
	if err != nil {
		return nil, err
	}

	return offerString, nil
}

// Creates an answer, sets it as the local description and returns the json marshaled bytes
func (pc *PeerConn) NewAnswer() ([]byte, error) {
	log.Println("[SDP] Creating Answer")
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}
	log.Println("[PC] Setting local description")
	err = pc.SetLocalDescription(answer)
	if err != nil {
		return nil, err
	}
	// Marshal the answer and send it back to SFU
	answerString, err := json.Marshal(answer)
	if err != nil {
		return nil, err
	}

	return answerString, nil
}

// func (pc *PeerConn) HandleICE(data string) {
// 	log.Println("[ICE] Candidate received")
// 	candidate := webrtc.ICECandidateInit{}
// 	if err := json.Unmarshal([]byte(data), &candidate); err != nil {
// 		log.Println("[ERR] Unmarshalling ICECandidate: ", err)
// 		return
// 	}
// 	log.Println("[ICE] Adding Candidate")
// 	if err := pc.AddICECandidate(candidate); err != nil {
// 		log.Println("[ERR] Adding ICECandidate: ", err)
// 		return
// 	}
// }
