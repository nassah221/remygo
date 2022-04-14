package events

import (
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/remygo/internal/robot"
	"github.com/remygo/internal/types"
	"github.com/remygo/internal/utils"

	"gioui.org/io/key"
	"gioui.org/io/pointer"
)

// Gio and robotgo maps special and modifier keys differently so this map pairs
// each gio key with the corresponding robotgo key for translation
var (
	specialKeys = map[string]string{"←": "left", "→": "right", "↑": "up", "↓": "down", "⏎": "enter", "⌤": "enter",
		"⎋": "escape", "⇱": "home", "⇲": "end", "⌫": "backspace", "⌦": "delete", "⇞": "pageup", "⇟": "pagedown",
		"⇥": "tab", "Space": "space"}

	modifiers = map[string]string{"Ctrl": "ctrl", "Alt": "alt", "Shift": "shift", "ModSuper": "cmd"}
)

// Deserializes event messages arriving on the host datachannel
// into local events for robotgo to consume
func ParseEvent(payload []byte) {
	event := types.RemoteEvent{}
	if err := json.Unmarshal(payload, &event); err != nil {
		log.Println("Failed to parse event", err)
		return
	}
	switch event.Type {
	case "move":
		ev := types.MouseMove{}
		if err := json.Unmarshal(event.Event, &ev); err != nil {
			log.Println(err)
			return
		}
		robot.MoveEvent(&ev)
	case "click":
		ev := types.MouseClick{}
		if err := json.Unmarshal(event.Event, &ev); err != nil {
			log.Println(err)
			return
		}
		robot.ClickEvent(&ev)
	case "scroll":
		ev := types.MouseScroll{}
		if err := json.Unmarshal(event.Event, &ev); err != nil {
			log.Println(err)
			return
		}
		robot.ScrollEvent(&ev)
	case "drag":
		ev := types.MouseDrag{}
		if err := json.Unmarshal(event.Event, &ev); err != nil {
			log.Println(err)
			return
		}
		robot.DragEvent(&ev)
	case "key":
		ev := types.KeyEvent{}
		if err := json.Unmarshal(event.Event, &ev); err != nil {
			log.Println(err)
			return
		}
		//? Just print the key events for now
		// log.Println("Key event: ", ev.Key, ev.Action, ev.Modifiers)
		robot.KeyEvent(&ev)
	}
}

// Constructs and returns JSON serialized pointer+key events from gio window
// to send across the remote datachannel to the host
func RemoteEvent(ev interface{}, cts ...int) (*types.RemoteEvent, error) {
	switch ev := ev.(type) {
	case *pointer.Event:
		switch ev.Type {
		case pointer.Press:
			event := &types.MouseClick{Button: ev.Buttons.String(), Action: strings.ToLower(ev.Type.String())}
			eventJSON := utils.MarshalEvent(event)

			return &types.RemoteEvent{Type: "click", Event: eventJSON}, nil

		case pointer.Release:
			event := &types.MouseClick{Action: strings.ToLower(ev.Type.String())}
			eventJSON := utils.MarshalEvent(event)

			return &types.RemoteEvent{Type: "click", Event: eventJSON}, nil

		case pointer.Move:
			// cts is used to pass in the gio window size on the remote 	//^ *(in a move only event)
			// so that the host can scale the mouse coordinates accordingly
			// this is done on the host side because of the following reason:
			// remote does not know the screen resolution of the host
			// therefore, mouse "move" events coming from the remote contain
			// the gio window size for the host to scale normalized coords from

			//^ *(the extra information [cts param] is only sent with a move event as 'activity' from the remote is demonstrated
			//^ when the remote mouse is moved. Therefore, any gio window size changes would be reflected when the remote
			//^ is active, hence the host would only need to check for scaling the coords when the remote is moving the mouse)

			// Currently, the gio window size is fixed to take the max available space on a 1080p display
			// The above is a note to self for future consideration

			// We may as well skip all of this realtime behavior and just send a config message to the host
			// with the screen resolution and the gio window size when initiating the datachannel since
			// the gio window size on the remote is going to be of a fixed size anyway

			// (only options are going to be fullscreen, not fullscreen i.e. windowed with just the taskbar visible)

			event := &types.MouseMove{
				X: ev.Position.X, Y: ev.Position.Y, W: float32(cts[0]),
				H: float32(cts[1])}
			eventJSON := utils.MarshalEvent(event)

			return &types.RemoteEvent{Type: "move", Event: eventJSON}, nil

		case pointer.Drag:
			event := &types.MouseDrag{Button: ev.Buttons.String(), X: ev.Position.X, Y: ev.Position.Y}
			eventJSON := utils.MarshalEvent(event)

			return &types.RemoteEvent{Type: "drag", Event: eventJSON}, nil

		case pointer.Scroll:
			event := &types.MouseScroll{Scroll: int(ev.Scroll.Y)}
			eventJSON, err := json.Marshal(event)
			if err != nil {
				log.Println(err)
			}

			return &types.RemoteEvent{Type: "scroll", Event: eventJSON}, nil
		case pointer.Cancel:
			// Ignore cancel event
			return nil, nil
		}
	case *key.Event:
		var event *types.KeyEvent

		specialKey := parseSpecialKey(ev.Name)
		modKey := parseModkey(ev.Modifiers.String())

		if specialKey == "" {
			event = &types.KeyEvent{Key: strings.ToLower(ev.Name), Action: strings.ToLower(ev.State.String()), Modifiers: modKey}
		} else {
			event = &types.KeyEvent{Key: specialKey, Action: strings.ToLower(ev.State.String()), Modifiers: modKey}
		}
		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Println(err)
		}
		return &types.RemoteEvent{Type: "key", Event: eventJSON}, nil
	}
	return nil, errors.New("unexpected event type")
}

//
//
//? ----------------The following functions may belong in a separate package-------------------|
//
//
// Returns the special key mapping if present, otherwise returns empty string
func parseSpecialKey(ev string) string {
	var specialKey string
	if val, ok := specialKeys[ev]; ok {
		specialKey = val
	}
	return specialKey
}

// Returns the modifier key(s) mapping if present, otherwise returns empty string slice
func parseModkey(mod string) []string {
	modKeys := strings.Split(mod, "|")

	// strings.Split() returns an empty string of length 1
	// in which case none of the modifiers are pressed
	// so we need to check for that and
	// return early with an empty slice of length 0
	if modKeys[0] == "" {
		return []string{}
	}

	switch len(modKeys) {
	case 1:
		modKeys[0] = modifiers[modKeys[0]]
	default:
		for i, v := range modKeys {
			modKeys[i] = modifiers[v]
		}
	}
	return modKeys
}

//
//? -------------------------------------------------------------------------------------------|
