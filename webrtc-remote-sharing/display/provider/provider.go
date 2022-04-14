package provider

import (
	"github.com/remygo/display/capture"
	"github.com/remygo/display/play"
	"github.com/remygo/internal/types"
	"github.com/remygo/internal/utils"
)

// type PlayBackProvider interface {
// 	Start(width, height, payloadType int, codecName string) error
// 	Stop() error
// 	Push(buffer []byte)
// }

func GetPlaybackProvider() *play.GstPlayback {
	return &play.GstPlayback{}
}

func GetCaptureProvider() *capture.GstCapture {
	return &capture.GstCapture{}
}

func GetWindowProvider() *Window {
	return newWindow(utils.GetDisplaySize())
}

func (w *Window) GetEventQueue() chan *types.RemoteEvent {
	return w.eventQueue
}
