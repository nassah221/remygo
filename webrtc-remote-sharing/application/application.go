package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/remygo/conn/wrtc"
	"github.com/remygo/conn/ws"
	"github.com/remygo/display"
	"github.com/remygo/pkg/logger"
	"github.com/remygo/pkg/message"

	"github.com/pion/webrtc/v3"
)

type Resolution struct {
	Width, Height int
}

type Args struct {
	URL, TurnCreds, Codec, Addr, ConfigPath, UserCreds string
}

func NewArgs() *Args {
	return &Args{}
}

type Request struct {
	Token  string // Session token
	Status string // pending, complete
	Next   string // Which call should come next
}

// Main application which bundles all the necessary components
type App struct {
	*ws.Socket
	*Args
	*wrtc.PeerConn
	*webrtc.DataChannel
	Resolution
	*MediaComponents
	candidatesTXQueue chan *webrtc.ICECandidate
	candidatesRXQueue chan webrtc.ICECandidateInit
	ticker            *time.Ticker
	HostTrack         *webrtc.TrackLocalStaticSample
	mode              Mode
	UserID, DeviceID  string
	SessionToken      string
	registerRequest   *Request
	callRequest       *Request
	renewRequest      *Request
	From              <-chan message.Message
	done              chan struct{} // Signal to close the application when signaling server rejects any client request
	sessionEvents     chan SessionEvent
	reset             bool
	Ctx               context.Context // Context to cancel the ICE service
	CtxCancel         context.CancelFunc
}

// Returns a new instance of the application
func New(socket *ws.Socket, fromSocket <-chan message.Message, events chan SessionEvent,
	cfg *Args, userID, deviceID string) *App {
	width, height := display.Dimensions()

	// Configure the application with the provided configuration
	return &App{
		Socket:            socket,
		Args:              cfg,
		Resolution:        Resolution{width, height},
		UserID:            userID,
		DeviceID:          deviceID,
		ticker:            time.NewTicker(time.Millisecond * 100),
		MediaComponents:   newMediaComponents(),
		candidatesTXQueue: make(chan *webrtc.ICECandidate, 32),
		candidatesRXQueue: make(chan webrtc.ICECandidateInit, 32),
		done:              make(chan struct{}),
		From:              fromSocket,
		sessionEvents:     events,
	}
}

func (app *App) ICEService(ctx context.Context) {
	// TODO: Handle proper context cancel propagation from the app main loop
	// defer ctx.Done()

	// <-app.ticker.C

	defer func() {
		close(app.candidatesRXQueue)
		close(app.candidatesTXQueue)
		log.Println("[INFO] Closing ICE service")
	}()
outer:
	for {
		select {
		case c, ok := <-app.candidatesTXQueue:
			if !ok {
				continue
			}
			cs, err := json.Marshal(c.ToJSON())
			if err != nil {
				log.Panicf("[ERR] Creating candidate string: %v", err)
			}
			msgICE := message.NewSignal(message.ICE, cs)

			log.Println("[ICE] Sending ICE candidate")

			if err = app.Socket.Write(*msgICE); err != nil {
				log.Panicf("[ERR] Cannot send ICE signal message. %v", err)
			}
			<-app.ticker.C
		case c, ok := <-app.candidatesRXQueue:
			if !ok {
				continue
			}
			log.Println("[ICE] Received ICE candidate")
			if app.PeerConn.RemoteDescription() != nil {
				if err := app.PeerConn.AddICECandidate(c); err != nil {
					log.Panicf("[ERR] Adding ICE candidate: %v", err)
				}
			}
		case <-app.ticker.C:
			if app.PeerConn.ICEGatheringState() == webrtc.ICEGatheringStateComplete {
				log.Println("[ICE] Local gathering complete")
				break outer
			}
			// case <-ctx.Done():
			// 	break outer
		}
	}
}

func (app *App) configureAsHost() {
	peerConnection, videoTrack, err := wrtc.NewHost(app.Args.URL, app.Args.TurnCreds, app.Args.Codec)
	if err != nil {
		log.Panicf("[ERR] Creating peer connection: %v", err)
	}
	app.PeerConn, app.HostTrack = peerConnection, videoTrack

	if err := app.ConnectCallbacks(app.Ctx, Host); err != nil {
		log.Panicf("[ERR] Connecting callbacks: %v", err)
	}
	log.Println("[INFO] Host callbacks connected")

	if err := app.Capture.Start(app.Resolution.Height, app.Resolution.Height,
		app.Args.Codec, app.HostTrack); err != nil {
		log.Panicf("[ERR] Starting capture: %v", err)
	}
	app.mode = Host

	go app.ICEService(context.TODO())
}

func (app *App) configureAsRemote() {
	// On clicking the join button, we need to start the remote application mode
	peerConnection, err := wrtc.NewRemote(app.Args.URL, app.Args.TurnCreds)
	if err != nil {
		log.Panicf("[ERR] Creating peer connection: %v", err)
	}
	dataChannel, err := peerConnection.CreateDataChannel("data", nil)
	if err != nil {
		log.Panicf("[ERR] Creating data channel: %v", err)
	}
	app.PeerConn, app.DataChannel = peerConnection, dataChannel
	if err := app.ConnectCallbacks(app.Ctx, Remote); err != nil {
		log.Panicf("[ERR] Connecting callbacks: %v", err)
	}
	log.Println("[INFO] Remote callbacks connected")
	app.mode = Remote

	go app.ICEService(context.TODO())
}

// If app.debugToken is non-empty, send it to create a session of the same name
// to make debugging easier otherwise the signaling server will generate and assign a uuid to the session
func (app *App) RegisterSession() error {
	log.Println("[INFO] Registering session")
	app.registerRequest = &Request{Status: "pending", Next: message.Token.String()}

	return app.Socket.Write(*message.NewInfo(message.Register, "", app.UserID, app.DeviceID))
}

func (app *App) GetSessionToken() string {
	for range app.ticker.C {
		if app.SessionToken == "" {
			continue
		}
		break
	}
	return app.SessionToken
}

func (app *App) JoinSession(token string) error {
	app.callRequest = &Request{Token: token, Status: "pending", Next: message.Ack.String()}

	return app.Socket.Write(*message.NewSession(message.JoinRequest, token, nil))
}

func (app *App) Close() error {
	// Stop the ICE service if it's running
	app.CtxCancel()
	switch app.mode {
	case Host:
		if err := app.Capture.Stop(); err != nil {
			return fmt.Errorf("stopping capture: %q", err)
		}
	case Remote:
		if err := app.Playback.Stop(); err != nil {
			return fmt.Errorf("stopping playback: %q", err)
		}
	}

	if err := app.PeerConn.Close(); err != nil {
		log.Panicf("[ERR] Closing peer connection: %v", err)
	}
	app.Socket = nil
	app.PeerConn = nil
	app.DataChannel = nil
	app.MediaComponents = nil
	app.callRequest = nil
	app.registerRequest = nil
	app.HostTrack = nil
	app.Ctx, app.CtxCancel = nil, nil
	app.ticker.Stop()

	return nil
}

func (app *App) Reset() {
	app.renewRequest = &Request{Status: "pending", Next: message.Renew.String()}
	switch app.mode {
	case Remote:
		if err := app.Playback.Stop(); err != nil {
			log.Panicf("stopping playback: %q", err)
		}

		log.Println("[APP] Remote mode. Client will leave the current session and renew the token")
		app.Socket.Write(*message.NewSession(message.Leave, "", nil))
	case Host:
		if err := app.Capture.Stop(); err != nil {
			log.Panicf("stopping capture: %q", err)
		}

		log.Println("[APP] Host mode. Current session terminated and token will be renewed automatically")
	}
	app.reset = false

	if err := app.PeerConn.Close(); err != nil {
		log.Panicf("[ERR] Closing peer connection: %v", err)
	}
	if app.DataChannel != nil {
		if err := app.DataChannel.Close(); err != nil {
			log.Panicf("[ERR] Closing data channel: %v", err)
		}
		app.DataChannel = nil
	}
	app.PeerConn = nil
	app.callRequest = nil
	app.MediaComponents = newMediaComponents()
	app.candidatesTXQueue = make(chan *webrtc.ICECandidate, 32)
	app.candidatesRXQueue = make(chan webrtc.ICECandidateInit, 32)
	app.done = make(chan struct{})
	app.SessionToken = ""
	app.mode = 0
}

// Message loop that blocks on receiver channel of the websocket type and handles the messages
func (app *App) Start(ctx context.Context) {
	defer func() {
		app.Close()
	}()
	log.Println("[APP] Starting main loop")

	if app.CtxCancel == nil {
		app.Ctx, app.CtxCancel = context.WithCancel(ctx)
	}
	// // TODO: Interrupt handling should happen inside the main package since this loop is non-blocking
loop:
	for {
		select {
		case m, ok := <-app.From:
			if !ok {
				log.Print("[APP] Websocket connection closed")
				break loop
			}

			logger.LogMessage(&m, "[WS] Received message. ")

			switch m.Type {
			case message.Signal:
				var msg message.SignalMessage

				if err := json.Unmarshal([]byte(m.Data), &msg); err != nil {
					log.Panicf("[ERR] Unmarshalling session message. %v", err)
				}

				app.handleSignaling(&msg)
			case message.Session:
				var msg message.SessionMessage

				if err := json.Unmarshal([]byte(m.Data), &msg); err != nil {
					log.Panicf("[ERR] Unmarshalling session message. %v", err)
				}
				app.handleSession(&msg)
			case message.Command:
				var msg message.CommandMessage

				if err := json.Unmarshal([]byte(m.Data), &msg); err != nil {
					log.Panicf("[ERR] Unmarshalling session message. %v", err)
				}

				app.handleCommand(&msg)
			case message.Info:
				var msg message.InfoMessage

				if err := json.Unmarshal([]byte(m.Data), &msg); err != nil {
					log.Panicf("[ERR] Unmarshalling info message. %v", err)
				}

				app.handleInfo(&msg)
			}
		case <-app.done:
			log.Println("[INFO] Done received, closing")
			break loop
		case <-app.MediaComponents.Done:
			log.Println("[APP] Media component close signal received. Restarting")
			// mode := app.mode
			if app.reset {
				app.Reset()
			}

			// case <-ctx.Done():
			// 	fmt.Println("[APP] Context canceled")
			// 	break loop
		}
	}
	log.Println("[APP] Exiting main loop")
}
