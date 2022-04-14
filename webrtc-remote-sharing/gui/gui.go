package gui

import (
	"context"
	"errors"
	"log"

	uievents "github.com/remygo/gui/events"
	"github.com/remygo/gui/landing"
	"github.com/remygo/gui/login"
	page "github.com/remygo/gui/pages"
	"github.com/remygo/swagger"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
)

type State uint8

const (
	LoginPage State = iota
	LandingPage
	InSession
)

func (s State) String() string {
	switch s {
	case LoginPage:
		return "Login Page"
	case LandingPage:
		return "Landing Page"
	case InSession:
		return "InSession"
	default:
		return "Unknown"
	}
}

type GUI struct {
	w            *app.Window
	router       page.Router
	redraw       chan struct{}
	EventsTX     chan uievents.Event
	EventsRX     chan uievents.Event
	CurrentState State
	PrevState    State
	// ctx        context.Context
	// cancelFunc context.CancelFunc
}

func NewGUI(w *app.Window) *GUI {
	g := &GUI{
		w:            w,
		router:       page.NewRouter(),
		EventsTX:     make(chan uievents.Event, 1),
		EventsRX:     make(chan uievents.Event, 1),
		redraw:       make(chan struct{}, 60),
		CurrentState: LoginPage,
	}

	loginFunc := func(ctx context.Context, u swagger.AddCredentials) (swagger.InlineResponse200, error) {
		return g.login(ctx, u)
	}

	g.router.Add(0, login.New(&g.router, g.EventsTX, loginFunc, g.redraw))
	g.router.Add(1, landing.New(&g.router, g.EventsTX))

	return g
}

func apiClient() *swagger.APIClient {
	return swagger.NewAPIClient(swagger.NewConfiguration())
}

func (g *GUI) login(ctx context.Context, user swagger.AddCredentials) (swagger.InlineResponse200, error) {
	apiClient := apiClient()

	var loginRes swagger.InlineResponse200
	var err error

	if user.Email == "admin" && user.Password == "admin" {
		return swagger.InlineResponse200{User: &swagger.User{Id: "69"}}, nil
	}

	loginRes, _, err = apiClient.UserApi.Login(ctx,
		swagger.AddCredentials{Email: user.Email, Password: user.Password})

	log.Printf("Login error: %q", err)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Printf("[ERR] API Deadline Exceeded: %q", err)
			err = errors.New("service timeout")
			return loginRes, err
		}
	}

	return loginRes, err
}

func (g *GUI) Loop() error {
	th := material.NewTheme(gofont.Collection())

	var ops op.Ops

	for {
		select {
		case e := <-g.w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				g.router.Layout(gtx, th)
				e.Frame(gtx.Ops)
			case key.Event:
				if e.Name == key.NameTab && e.State == key.Press {
					g.router.TabSwitch()
				}
			}
		case <-g.redraw:
			g.w.Invalidate()
		case ev := <-g.EventsRX:
			log.Println("[INFO] Received event: ", ev.Payload)
			switch ev.Type {
			case uievents.SetToken:
				log.Println("[INFO] Received set token event: ", ev.Payload)
				if token, ok := ev.Payload.(string); ok {
					g.router.SetToken(token)
				}
			case uievents.RenewToken:
				log.Println("[INFO] Received renew token event: ", ev.Payload)
				if token, ok := ev.Payload.(string); ok {
					g.router.SetToken(token)
				}
			case uievents.SessionStarted:
				log.Println("[INFO] Received session started event: ", ev.Payload)
				g.PrevState = g.CurrentState
				g.CurrentState = InSession
				log.Printf("[GUI] Prev state: %s Current state: %s", g.PrevState.String(), g.CurrentState.String())

				g.router.DisableJoinButton(true)
			}
		}
	}
}
