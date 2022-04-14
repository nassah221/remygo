package types

import (
	"encoding/json"
)

type RemoteEvent struct {
	Type  string `json:"type"`
	Event json.RawMessage
}

type MouseMove struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	W float32 `json:"width"`
	H float32 `json:"height"`
}

//TODO: Add modifiers
type MouseClick struct {
	Button string `json:"button"`
	Action string `json:"action"`
}

type MouseScroll struct {
	Scroll int `json:"scroll"`
}

type MouseDrag struct {
	Button string  `json:"button"`
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
}

type KeyEvent struct {
	Key       string   `json:"key"`
	Action    string   `json:"action"`
	Modifiers []string `json:"modifiers"`
}
