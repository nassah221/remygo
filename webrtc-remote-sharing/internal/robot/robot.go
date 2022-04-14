package robot

import (
	"github.com/remygo/internal/types"
	"github.com/remygo/internal/utils"

	"github.com/go-vgo/robotgo"
)

var (
	remoteW, remoteH float32

	userW, userH int = utils.GetDisplaySize()
	mouseDown    bool
)

func DragEvent(event *types.MouseDrag) {
	//TODO: implement Secondary and Tertiary drags
	nX, nY := utils.NormalizedPos(event.X, event.Y, remoteW, remoteH, userW, userH)
	switch event.Button {
	case "ButtonPrimary":
		if !mouseDown {
			mouseDown = true
			robotgo.Toggle("left", "down")
		}

		robotgo.Move(nX, nY)
	case "ButtonSecondary":
	}
}

func ScrollEvent(event *types.MouseScroll) {
	if event.Scroll > 0 {
		robotgo.Scroll(0, -2)
	} else {
		robotgo.Scroll(0, 2)
	}
}

func MoveEvent(event *types.MouseMove) {
	// On the first move event received, set the remote window size
	// if there is a change during runtime, scale accordingly
	if event.W != remoteW || event.H != remoteH {
		remoteW = event.W
		remoteH = event.H
	}

	// Make sure that mouse move events are not confused for drag events
	// on the host side
	// TODO: implement for Secondary and Tertiary buttons
	if mouseDown {
		robotgo.Toggle("left", "up")
		mouseDown = false
	}

	nX, nY := utils.NormalizedPos(event.X, event.Y, remoteW, remoteH, userW, userH)
	robotgo.Move(nX, nY)
}

func ClickEvent(event *types.MouseClick) {
	// TODO: 1) add Secondary and Tertiary clicks
	// TODO: 2) implement multiple simultaneous button presses
	// TODO: 3) add modifier key behavior
	switch event.Button {
	case "ButtonPrimary":
		if event.Action == "press" {
			robotgo.Toggle("left", "down")
			mouseDown = true
		} else if event.Action == "release" {
			robotgo.Toggle("left", "up")
			mouseDown = false
		}
	case "ButtonSecondary":
		if event.Action == "press" {
			robotgo.Click("right", false)
		} else if event.Action == "release" {
		}
	}
}

func KeyEvent(event *types.KeyEvent) {
	// TODO: implement KeyToggle instead of KeyTap for press/release event semantics
	if event.Action == "press" {
		if len(event.Modifiers) != 0 {
			robotgo.KeyTap(event.Key, event.Modifiers)
			return
		}
		robotgo.KeyTap(event.Key)
	} else if event.Action == "release" {
		// robotgo.KeyToggle(event.Key, "up", event.Modifiers)
	}
}
