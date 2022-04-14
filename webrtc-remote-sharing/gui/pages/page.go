package page

import (
	"fmt"
	"log"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type Tab interface {
	TabHandler()
}

type Setter interface {
	IsInfoSet() bool
	SetTokenInfo(token, pwd string)
}

type ButtonDisabler interface {
	Disable(bool)
}

type Page interface {
	Layout(gtx layout.Context, th *material.Theme) layout.Dimensions
}

type Router struct {
	pages   map[interface{}]Page
	current interface{}
	prevBtn widget.Clickable
}

func NewRouter() Router {
	return Router{
		pages: make(map[interface{}]Page),
	}
}

func (r *Router) Add(tag interface{}, p Page) {
	r.pages[tag] = p
	if r.current == interface{}(nil) {
		r.current = tag
	}
}

func (r *Router) TabSwitch() {
	curPage := r.current.(int)
	if ft, ok := r.pages[curPage].(Tab); ok {
		ft.TabHandler()
	} else {
		fmt.Println("Cant handle tabs ")
	}
}

func (r *Router) SetToken(token string) {
	log.Printf("[INFO] Setting token. Current page: %v, Token: %s\n", r.current, token)
	if pg, ok := r.pages[r.current].(Setter); ok {
		pg.SetTokenInfo(token, "session password not yet implemented")
		// if !pg.IsInfoSet() {
		// 	log.Println("[INFO] Valid page")
		// 	pg.SetTokenInfo(token, "session password not yet implemented")
		// }
	}
}

func (r *Router) DisableJoinButton(disable bool) {
	if pg, ok := r.pages[r.current].(ButtonDisabler); ok {
		pg.Disable(disable)
		return
	}
	log.Printf("[WARN] Current page %d does not implement ButtonDisabler", r.current)
}

func (r *Router) SwitchTo(tag interface{}) {
	_, ok := r.pages[tag]
	if !ok {
		return
	}
	r.current = tag
}

func (r *Router) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if r.prevBtn.Clicked() {
		prevPage := r.current.(int) - 1
		r.SwitchTo(prevPage)
	}

	content := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return r.pages[r.current].Layout(gtx, th)
	})

	// if r.current.(int) == 1 {
	// 	go func() {
	// 		curPage := r.current.(int)
	// 		if ft, ok := r.pages[curPage].(Setter); ok {
	// 			if !ft.IsInfoSet() {
	// 				r.SetInfo("FOOBAR-BAZQUX", "bar")
	// 			}
	// 		}
	// 	}()
	// }

	icPrev, err := widget.NewIcon(icons.NavigationArrowBack)
	if err != nil {
		panic(err)
	}

	return layout.Flex{
		Axis:      layout.Vertical,
		Alignment: layout.Middle,
		Spacing:   layout.SpaceBetween,
	}.Layout(gtx, content, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: unit.Dp(10),
					Bottom: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return material.IconButton(th, &r.prevBtn, icPrev, "Back").Layout(gtx)
				})
			}),
		)
	}))
}
