package application

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/remygo/display"
	"github.com/remygo/internal/events"

	"github.com/pion/webrtc/v3"
)

type Mode uint8

const (
	Host Mode = iota + 1
	Remote
)

func (app *App) ConnectCallbacks(ctx context.Context, mode Mode) error {
	// Connect common callbacks between host and remote operating modes
	app.PeerConn.OnICECandidate(func(i *webrtc.ICECandidate) {
		log.Println("[ICE] Found candidate")
		if i == nil {
			return
		}

		if app.PeerConn.ICEGatheringState() != webrtc.ICEGatheringStateComplete {
			log.Println("[ICE] Queueing ICE candidate")
			app.candidatesTXQueue <- i
			return
		}
	})

	// Log ICE connection state
	app.PeerConn.OnICEConnectionStateChange(func(is webrtc.ICEConnectionState) {
		log.Printf("[ICE] Connection state changed: %s", is.String())
	})

	// Log peer connection state and exit if failed
	app.PeerConn.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
		switch p {
		case webrtc.PeerConnectionStateFailed:
			if err := app.PeerConn.Close(); err != nil {
				log.Panicf("[ERR] Connecting to remote: %v", err)
			}
		case webrtc.PeerConnectionStateClosed:
			log.Println("[PC] Connection closed")
			return
		case webrtc.PeerConnectionStateConnected:
			log.Println("[PC] Connection established")
			return
		case webrtc.PeerConnectionStateConnecting:
			log.Println("[PC] Connecting to remote")
			return
		case webrtc.PeerConnectionStateNew:
		}
	})

	// Connect mode specific callbacks
	switch mode {
	case Host:
		return app.connectHostCallbacks()
	case Remote:
		return app.connectRemoteCallbacks()
	}
	return nil
}

func (app *App) connectHostCallbacks() error {
	app.PeerConn.OnDataChannel(func(dc *webrtc.DataChannel) {
		log.Println("[PC] Remote peer opened DataChannel")

		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			events.ParseEvent(msg.Data)
		})
	})

	return nil
}

func (app *App) connectRemoteCallbacks() error {
	// On receiving a track, write to the created output track
	app.PeerConn.OnTrack(func(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		log.Println("[PC] Host video track received")

		cancelRead := make(chan struct{})

		log.Printf("[PC] Track has started, of type %d: %s \n", tr.PayloadType(), tr.Codec().MimeType)
		wg := &sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := app.Playback.Loop()
			if err != nil {
				log.Panicf("[ERR] Stopping window provider: %v", err)
			}
			// Stop the RTP read loop
			log.Println("[DEBUG] Canceling RTP read loop")
			cancelRead <- struct{}{}

			// Cancel the app context
			// app.CtxCancel()
			log.Println("[DEBUG] Canceling app media component")

			app.reset = true
			app.MediaComponents.Done <- struct{}{}
		}()

		// go func() {
		// 	ticker := time.NewTicker(time.Second * 3)
		// 	for {
		// 		select {
		// 		case <-ticker.C:
		// 			rtcpSendErr := app.PeerConn.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(tr.SSRC())}})
		// 			if rtcpSendErr != nil {
		// 				log.Printf("[ERR] Writing RTCP: %v", rtcpSendErr)
		// 				return
		// 			}
		// 		case <-rtpReadCtx.Done():
		// 			// Stop if the playback window is closed
		// 			return
		// 		}
		// 	}
		// }()

		width, height := display.Dimensions()
		codecName := strings.Split(tr.Codec().RTPCodecCapability.MimeType, "/")[1]

		if err := app.Playback.Start(width, height, int(tr.PayloadType()), strings.ToLower(codecName)); err != nil {
			log.Panicf("[ERR] Initiating display pipeline: %v", err)
		}
		// for {
		// _, _, readErr := tr.ReadRTP()
		// if readErr != nil {
		// 	log.Println("[ERR] Reading received track: ", readErr)
		// 	return
		// }

		buf := make([]byte, 1500)
	loop:
		for {
			select {
			case <-cancelRead:
				log.Println("[APP] Stopping RTP track read loop")
				break loop
			default:
				i, _, readErr := tr.Read(buf)
				if readErr != nil {
					log.Printf("[ERR] Reading remote track: %v", readErr)
					return
				}
				app.Playback.HandleFrameBuffer(buf[:i])
			}
		}

		wg.Wait()
	})

	app.DataChannel.OnOpen(func() {
		log.Println("[PC] Data channel open")
		for msg := range app.Playback.UI.ReceiveInputEvents() {
			msgToSend, dcErr := json.Marshal(msg)
			if dcErr != nil {
				log.Println("[ERR] Marshalling mouse event:", dcErr)
			}
			sendErr := app.DataChannel.Send(msgToSend)
			if sendErr != nil {
				log.Printf("[ERR] Sending on DataChannel: %v", sendErr)
				return
			}
		}
		log.Println("[GIO PLAYBACK] Data channel event queue closed")
	})
	return nil
}
