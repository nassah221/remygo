package provider

import (
	"image"
	"image/color"
	"log"

	"github.com/remygo/internal/events"
	"github.com/remygo/internal/types"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

var done = make(chan struct{})
var cts layout.Constraints

type Window struct {
	window      *app.Window
	eventQueue  chan *types.RemoteEvent
	handlerChan chan event.Event
}

func (w *Window) Close() {
	done <- struct{}{}
}

func newWindow(width, height int) (w *Window) {
	gio := app.NewWindow(app.Title("Remote"),
		app.MaxSize(unit.Dp(float32(width)),
			unit.Dp(float32(height))))

	w = &Window{
		window:      gio,
		eventQueue:  make(chan *types.RemoteEvent),
		handlerChan: make(chan event.Event),
	}

	return
}

// Handle UI events here based on the event type
func (w *Window) handleEvents() {
	defer close(w.eventQueue)

	for ev := range w.handlerChan {
		switch ev := ev.(type) {
		case key.FocusEvent:
			log.Printf("[GIO] Focus: %v", ev.Focus)
		case key.Event:
			e, err := events.RemoteEvent(&ev)
			if err != nil {
				log.Println("[ERR] Creating remote event: ", err)
				return
			}

			w.eventQueue <- e

		case pointer.Event:
			e, err := events.RemoteEvent(&ev, cts.Max.X, cts.Max.Y)
			if err != nil {
				log.Println("[ERR] Creating remote event: ", err)
				return
			}

			w.eventQueue <- e
		}
	}
}

// windowLoop is responsible for handling the ui event loop
func (w *Window) Loop(frameQueue <-chan *image.NRGBA) error {
	var err error

	// Operation buffer
	var ops op.Ops
	var gtx layout.Context

	// Image buffer, reused on every frame
	var videoFrame *image.NRGBA

	// Image widget
	var img widget.Image
	img.Fit = widget.Contain
	img.Position = layout.Center

	// TODO: Handle clean shutdown on socket exit
	go w.handleEvents()
outer:
	for {
		select {
		case e := <-w.window.Events():
			switch e := e.(type) {
			// Triggers when window is maximized|minimized
			case system.StageEvent:
				log.Printf("[GIO] Stage: %v", e.Stage.String())
			// Triggers when window is closed
			case system.DestroyEvent:
				log.Println("[GIO] Destroy event")
				err = e.Err

				break outer
			// Triggers when window requests frame redraw
			case system.FrameEvent:
				gtx = layout.NewContext(&ops, e)
				cts = gtx.Constraints

				// Set the area for receiving pointer events
				// in our case, the entire window
				clip.Rect{Max: gtx.Constraints.Constrain(img.Src.Size())}.Push(gtx.Ops)

				// Handle events
				for _, ev := range gtx.Events(w.window) {
					select {
					case w.handlerChan <- ev:
					default:
					}
				}

				// Make our mouse cursor invisible
				// when hovering over the window
				pointer.CursorNameOp{Name: pointer.CursorNone}.Add(gtx.Ops)

				// Register pointer events we are interested in
				pointer.InputOp{
					Tag: w.window, Types: pointer.Press | pointer.Release | pointer.Drag |
						pointer.Move | pointer.Scroll, ScrollBounds: image.Rect(0, 0, 0, 100)}.Add(gtx.Ops)

				// Register keyboard events we are interested in when the handler is focused
				// in our case, the handler is the entire window
				key.FocusOp{Tag: w.window}.Add(gtx.Ops)
				key.InputOp{Tag: w.window}.Add(gtx.Ops)

				// Draw the image
				paint.Fill(gtx.Ops, color.NRGBA{A: 255})
				img.Layout(gtx)
				e.Frame(gtx.Ops)
			}

		case frame, ok := <-frameQueue:
			if !ok {
				log.Println("[GIO] Frame queue closed")
				break outer
			}

			videoFrame = frame
			img.Src = paint.NewImageOp(videoFrame)
			w.window.Invalidate()
		case _, ok := <-done:
			if !ok {
				log.Println("Done channel blocked")
				continue
			}
			// Close the handler channel and stop the event queue
			break outer
		}
	}

	close(w.handlerChan)
	log.Println("[GIO] Closing window")
	w.window.Perform(system.ActionClose)
	log.Println("[GIO] Window loop exited")
	return err
}
