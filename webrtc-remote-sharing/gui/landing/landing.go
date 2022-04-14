package landing

import (
	"fmt"

	uievents "github.com/remygo/gui/events"
	page "github.com/remygo/gui/pages"

	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

var (
	textFieldWidth = unit.Dp(200)
)

var CopyIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ContentContentCopy)
	return icon
}()

var PasteIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ContentContentPaste)
	return icon
}()

type (
	C = layout.Context
	D = layout.Dimensions
)

type RichEditor struct {
	tag int
	component.TextField
	copyButton, pasteButton widget.Clickable
}

func (re *RichEditor) Update(gtx C) {
	if re.copyButton.Clicked() {
		fmt.Println(re.Text())
		clipboard.WriteOp{Text: re.Text()}.Add(gtx.Ops)
	}

	if re.pasteButton.Clicked() {
		clipboard.ReadOp{Tag: &re.tag}.Add(gtx.Ops)
	}

	for _, e := range gtx.Events(&re.tag) {
		switch e := e.(type) {
		case clipboard.Event:
			fmt.Println("event", e)
			re.Editor.Insert(e.Text)
		}
	}
}

func (re *RichEditor) Layout(gtx C, th *material.Theme, label string) D {
	// Update the internal state of the text field widget
	re.Update(gtx)

	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			if re.tag <= 1 {
				return layout.Inset{Right: unit.Dp(10), Top: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) D {
					cpy := material.IconButton(th, &re.copyButton, CopyIcon, "Copy")
					cpy.Size = unit.Dp(20)
					return cpy.Layout(gtx)
				})
			}
			return layout.Inset{Right: unit.Dp(10), Top: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) D {
				paste := material.IconButton(th, &re.pasteButton, PasteIcon, "Paste")
				paste.Size = unit.Dp(20)
				return paste.Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx C) D {
			// Make the host field non-editable
			if re.tag <= 1 {
				return re.TextField.Layout(gtx.Disabled(), th, label)
			}
			return re.TextField.Layout(gtx, th, label)
		}),
	)
}

type Page struct {
	*page.Router
	hostToken, hostPwd     RichEditor
	remoteToken, remotePwd RichEditor
	joinBtn                widget.Clickable
	joinBtnDisabled        bool
	unattendedCheck        widget.Bool
	promptPwd              bool
	eventsTX               chan<- uievents.Event
}

func (p *Page) IsInfoSet() bool {
	return p.hostToken.Text() != "" && p.hostPwd.Text() != ""
}

func (p *Page) SetTokenInfo(token, pwd string) {
	p.hostToken.SetText(token)
	p.hostPwd.SetText(pwd)
}

func (p *Page) Disable(b bool) {
	p.joinBtnDisabled = b
}

func New(router *page.Router, joinSignal chan<- uievents.Event) *Page {
	p := Page{Router: router, promptPwd: false, eventsTX: joinSignal}

	// Assign each of the text fields a unique tag
	// which will help us tag the clipboard events
	p.hostToken = RichEditor{tag: 0}
	p.hostPwd = RichEditor{tag: 1}
	p.remoteToken = RichEditor{tag: 2}
	p.remotePwd = RichEditor{tag: 3}

	return &p
}

func heightSpacer(gtx C, units float32) D {
	return layout.Spacer{Height: unit.Dp(units)}.Layout(gtx)
}

func (p *Page) Layout(gtx C, th *material.Theme) D {
	margin := layout.Inset{Left: textFieldWidth, Right: textFieldWidth}

	if p.remoteToken.Len() >= 4 {
		p.promptPwd = true
	} else {
		p.promptPwd = false
	}

	if p.joinBtn.Clicked() {
		if p.remoteToken.Len() == 0 {
			p.remoteToken.SetError("Please enter a session token you want to join as remote")
		} else if p.remoteToken.Len() < 4 {
			p.remoteToken.SetError("Invalid token")
		}
		p.eventsTX <- uievents.Event{Type: uievents.JoinSession, Payload: p.remoteToken.Text()}
	}

	for _, e := range p.remoteToken.Events() {
		switch e.(type) {
		case widget.ChangeEvent:
			if p.remoteToken.IsErrored() {
				p.remoteToken.ClearError()
			}
		}
	}

	// gtx.Constraints.Max.Y = gtx.Px(unit.Dp(300))
	gtx.Constraints.Max.X = gtx.Px(unit.Dp(800))

	return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) D {
			margin.Left = unit.Dp(180)
			return margin.Layout(gtx, func(gtx layout.Context) D {
				header := material.H4(th, "REMYGO")
				header.Font.Weight = text.UltraBlack
				header.Color = th.ContrastBg
				return header.Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx C) D {
			margin.Left = unit.Dp(150)
			return margin.Layout(gtx, func(gtx C) D {
				return p.hostToken.Layout(gtx, th, "Your Session Token")
			})
		}),
		layout.Rigid(func(gtx C) D {
			return heightSpacer(gtx, 10)
		}),
		layout.Rigid(func(gtx C) D {
			margin.Left = unit.Dp(150)
			return margin.Layout(gtx, func(gtx C) D {
				return p.hostPwd.Layout(gtx, th, "Your Session Password")
			})
		}),
		layout.Rigid(func(gtx C) D {
			return layout.Spacer{Height: unit.Dp(40)}.Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			p.remoteToken.SingleLine = true
			// p.remoteToken.Submit = true
			// margin.Left, margin.Right = unit.Dp(180), unit.Dp(170)
			margin.Left = unit.Dp(150)
			return margin.Layout(gtx, func(gtx C) D {
				return p.remoteToken.Layout(gtx, th, "Host Session Token")
			})
		}),
		layout.Rigid(func(gtx C) D {
			return heightSpacer(gtx, 20)
		}),
		layout.Rigid(func(gtx C) D {
			p.remotePwd.SingleLine = true
			// p.remotePwd.Submit = true
			if p.promptPwd {
				// margin.Left, margin.Right = unit.Dp(180), unit.Dp(170)
				margin.Left = unit.Dp(150)
				return margin.Layout(gtx, func(gtx C) D {
					return p.remotePwd.Layout(gtx, th, "Host Session Password")
				})
			}
			return D{}
		}),
		layout.Rigid(func(gtx C) D {
			return heightSpacer(gtx, 20)
		}),
		layout.Rigid(func(gtx C) D {
			margin.Left, margin.Right = unit.Dp(180), unit.Dp(170)
			return margin.Layout(gtx, func(gtx C) D {
				return material.CheckBox(th, &p.unattendedCheck, "Unattended Access").Layout(gtx)
			})
		}),
		layout.Rigid(func(gtx C) D {
			return heightSpacer(gtx, 10)
		}),
		layout.Rigid(func(gtx C) D {
			margin.Left, margin.Right = unit.Dp(180), unit.Dp(170)
			return margin.Layout(gtx, func(gtx C) D {
				btn := material.Button(th, &p.joinBtn, "Join Session")
				if p.joinBtnDisabled {
					return btn.Layout(gtx.Disabled())
				}
				return btn.Layout(gtx)
			})
		}),
	)
}
