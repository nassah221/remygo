package login

import (
	"context"
	"log"
	"time"

	uievents "github.com/remygo/gui/events"
	page "github.com/remygo/gui/pages"
	"github.com/remygo/swagger"

	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

var (
	textFieldWidth = unit.Dp(200)
)

type userInfo struct {
	email    string
	password string
}

type Page struct {
	*page.Router
	inputEmail, inputPwd component.TextField
	loginBtn             widget.Clickable
	user                 userInfo                                                                         // User email and password
	fields               map[int]*component.TextField                                                     // Focusable fields
	currentfield         int                                                                              // Current focused field
	eventsTX             chan<- uievents.Event                                                            // Login error channel
	loginUser            func(context.Context, swagger.AddCredentials) (swagger.InlineResponse200, error) // Function for logging in with user email and password
	btnDisabled          bool
	redraw               chan struct{}
}

func New(router *page.Router, loginErr chan<- uievents.Event,
	loginUser func(context.Context, swagger.AddCredentials) (swagger.InlineResponse200, error), redraw chan struct{}) *Page {
	return &Page{Router: router, fields: make(map[int]*component.TextField),
		eventsTX: loginErr, loginUser: loginUser, redraw: redraw}
}

func (p *Page) TabHandler() {
	if v, ok := p.fields[p.currentfield+1]; ok {
		v.Focus()
	} else {
		p.fields[0].Focus()
	}
}

func (p *Page) Layout(gtx C, th *material.Theme) D {
	margin := layout.Inset{Top: unit.Dp(10), Left: textFieldWidth, Right: textFieldWidth}

	if p.loginBtn.Clicked() {
		if p.inputEmail.Text() == "" {
			p.inputEmail.SetError("Email is required")
		}
		if p.inputPwd.Text() == "" {
			p.inputPwd.SetError("Password is required")
		}

		if p.inputEmail.Text() != "" && p.inputPwd.Text() != "" {
			p.user.email = p.inputEmail.Text()
			p.user.password = p.inputPwd.Text()

			log.Printf("%+v\n", p.user)

			// Create new context every time login button is pressed
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)

			// Call the login function
			go func() {
				p.btnDisabled = true
				defer cancel()

				loginResp, err := p.loginUser(ctx, swagger.AddCredentials{Email: p.user.email, Password: p.user.password})

				p.btnDisabled = false

				if err != nil {
					log.Printf("%q", err)

					p.inputEmail.SetError(err.Error())
					p.inputPwd.SetError(err.Error())

					return
				}
				p.Router.SwitchTo(1)

				log.Println("[INFO] Login successful")
				p.inputEmail.ClearError()
				p.inputPwd.ClearError()

				p.eventsTX <- uievents.Event{Type: uievents.LoginSuccess, Error: nil, Payload: loginResp}
				p.redraw <- struct{}{}
			}()

			// if err == nil {
			// } else {
			// 	log.Printf("%q", err)

			// 	p.inputEmail.SetError(err.Error())
			// 	p.inputPwd.SetError(err.Error())
			// }
		}
	}

	for k, v := range p.fields {
		if v.Focused() {
			p.currentfield = k
		}
	}

	p.inputEmail.Helper = "foobar@bazqux.com"

	// gtx.Constraints.Max.X = gtx.Px(unit.Dp(800))
	// gtx.Constraints.Max.Y = gtx.Px(unit.Dp(300))
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return margin.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				header := material.H4(th, "LOGIN")
				header.Font.Weight = text.UltraBlack
				header.Color = th.ContrastBg
				return header.Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return margin.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				for _, evt := range p.inputEmail.Events() {
					switch evt.(type) {
					case widget.ChangeEvent:
						p.inputEmail.ClearError()
					}
				}
				p.inputEmail.Submit = true
				p.inputEmail.SingleLine = true
				p.fields[0] = &p.inputEmail
				return p.inputEmail.Layout(gtx, th, "Email")
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return margin.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				for _, evt := range p.inputPwd.Events() {
					switch evt.(type) {
					case widget.ChangeEvent:
						p.inputPwd.ClearError()
					case widget.SubmitEvent:
						p.loginBtn.Click()
					}
				}
				p.inputPwd.Submit = true
				p.inputPwd.SingleLine = true
				p.inputPwd.Mask = 42
				p.fields[1] = &p.inputPwd
				return p.inputPwd.Layout(gtx, th, "Password")
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return margin.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				btn := material.Button(th, &p.loginBtn, "Login")
				if p.btnDisabled {
					return btn.Layout(gtx.Disabled())
				}
				return btn.Layout(gtx)
			})
		}),
	)
}
