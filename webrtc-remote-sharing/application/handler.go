package application

import (
	"fmt"
	"log"

	"github.com/pion/webrtc/v3"
	"github.com/remygo/pkg/message"
)

func (app *App) handleSession(msg *message.SessionMessage) {
	fmt.Printf("[TYPE]: %s\n\n", msg.String())
	switch msg.Type {
	case message.JoinRequest:
		// For the host, there is no explicit flow. However, the host is the controller of the transaction
		// i.e. it can either allow or deny the call request. Before the host allows the call however,
		// it must first setup the correct state with appropriate paramenters which is what's happening here

		// // TODO: This state initiation should be refactored. I'm being lazy right now
		if app.callRequest == nil {
			log.Println("[INFO] Call request received. Configuring as host")
			app.configureAsHost()
		}

		// TODO: There should be a popup/dialog on the GUI to confirm the request
		// TODO: The host should be able to accept/reject the request
		// For now, allow the host to accept the session directly

		// FIXME: This is a really bad way to do this. I shouldn't be passing a pointer to a string
		// FIXME: just so it can be an optional parameter.
		allow := "allow"
		log.Println("[SESSION] Join request received. Accepting request")

		app.Socket.Write(*message.NewSession(message.JoinResponse, app.SessionToken, &allow))
	}
}

func (app *App) handleCommand(msg *message.CommandMessage) {
	fmt.Printf("\n[MESSAGE TYPE]: %s\n", msg.String())
	switch msg.Type {
	case message.InitiateSession:
		// Create an offer to send to the host peer
		offerString, err := app.PeerConn.NewOffer()
		if err != nil {
			log.Panicf("[ERR] Creating offer: %v", err)
		}
		offer := message.NewSignal(message.Offer, offerString)

		app.Socket.Write(*offer)

	case message.TerminateSession:
		fmt.Println("[WS] Session terminated")

		// Set the reset flag
		app.reset = true
		app.MediaComponents.Done <- struct{}{}
	}
}

// Handles the signaling messages
func (app *App) handleSignaling(msg *message.SignalMessage) {
	fmt.Printf("[TYPE]: %s\n\n", msg.String())
	switch msg.Type {
	case message.ICE:
		log.Println("[ICE] Candidate received")
		candidate := webrtc.ICECandidateInit{}
		if err := msg.IntoICE(&candidate); err != nil {
			log.Panicf("[ERR] Deserializing into ICE Candidate %v", err)
		}

		if app.PeerConn.RemoteDescription() == nil {
			log.Println("[ICE] Remote description is nil. Queueing received candidate")
			app.candidatesRXQueue <- candidate
			return
		}
		log.Println("[ICE] Adding candidate")
		if err := app.PeerConn.AddICECandidate(candidate); err != nil {
			log.Println("[ERR] Adding ICE candidate:", err)
		}

	case message.Offer:
		log.Println("[SDP] Offer received")
		offer := webrtc.SessionDescription{}
		if err := msg.IntoSDP(&offer); err != nil {
			log.Panicf("[ERR] Deserializing into SDP Offer: %v", err)
		}

		log.Println("[PC] Setting remote description")
		if err := app.PeerConn.SetRemoteDescription(offer); err != nil {
			log.Println("[ERR] Setting remote description: ", err)
			return
		}
		// Create an answer to send to the remote peer
		answerString, err := app.PeerConn.NewAnswer()
		if err != nil {
			log.Panicf("[ERR] Creating answer: %v", err)
		}
		msg := message.NewSignal(message.Answer, answerString)

		if err = app.Socket.Write(*msg); err != nil {
			log.Panicf("[ERR] Cannot send Answer signal message. %v", err)
		}

	case message.Answer:
		log.Println("[SDP] Answer received")
		answer := webrtc.SessionDescription{}
		if err := msg.IntoSDP(&answer); err != nil {
			log.Panicf("[ERR] Deserializing into SDP Answer: %v", err)
		}

		log.Println("[PC] Setting remote description")
		if err := app.PeerConn.SetRemoteDescription(answer); err != nil {
			log.Println("[ERR] Setting remote description: ", err)
			return
		}
	}
}

func (app *App) handleInfo(msg *message.InfoMessage) {
	fmt.Printf("\n[MESSAGE TYPE]: %s\n", msg.String())
	switch msg.Type {
	case message.Token:
		if app.registerRequest.Status == "pending" && app.registerRequest.Next == message.Token.String() {
			log.Printf("[INFO] Received register response: %+#v\n", msg)
			app.SessionToken = msg.Data
			log.Println("[INFO] Session token:", app.SessionToken)
			app.registerRequest = nil
			return
		}
		log.Panic("[WARN] Token received but no pending register call")
	case message.Error:
		fmt.Printf("\n(ERROR): %s\n\n", msg.Data)

		app.done <- struct{}{}
	case message.Ack:
		if app.callRequest != nil && app.callRequest.Status == "pending" {
			log.Println("[INFO] Call approval received. Configuring as remote")
			app.callRequest.Status = "active"
			app.callRequest.Next = ""

			fmt.Printf("\n(ACK): %s\n\n", msg.Data)
			app.configureAsRemote()
			return
		}
		log.Panic("[WARN] Acknowledge received but no pending join call")
	case message.Renew:
		if app.reset {
			app.Reset()
		}
		if app.renewRequest != nil && app.renewRequest.Status == "pending" {
			app.SessionToken = msg.Data
			log.Printf("[APP] Session token renewed. New token %s\n", app.SessionToken)
		} else if app.renewRequest == nil {
			log.Panic("[WARN] Renew request received but no pending renew call")
		}
		app.renewRequest = nil

		app.sessionEvents <- SessionEvent{Type: Renew}
	}
}
