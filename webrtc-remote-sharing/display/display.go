package display

import (
	"image"

	"github.com/remygo/display/capture"
	"github.com/remygo/display/play"
	"github.com/remygo/display/provider"
	"github.com/remygo/internal/types"
	"github.com/remygo/internal/utils"

	"github.com/pion/webrtc/v3"
)

type Capture struct {
	Provider *capture.GstCapture
}

type UI struct {
	Window *provider.Window
	frames <-chan *image.NRGBA
}

type Playback struct {
	Provider *play.GstPlayback
	frames   chan<- *image.NRGBA
}

func NewCaptureProvider() *Capture {
	return &Capture{
		Provider: provider.GetCaptureProvider(),
	}
}

func NewPlaybackProvider() (playback *Playback, window *UI) {
	frames := make(chan *image.NRGBA)

	playback = &Playback{
		Provider: provider.GetPlaybackProvider(),
		frames:   frames,
	}
	window = &UI{
		Window: provider.GetWindowProvider(),
		frames: frames,
	}

	return
}

func (c *Capture) Start(width, height int, codecName string, track *webrtc.TrackLocalStaticSample) error {
	return c.Provider.Start(width, height, codecName, track)
}

func (c *Capture) Stop() error {
	return c.Provider.Stop()
}

func (u *UI) Loop() error {
	if err := u.Window.Loop(u.frames); err != nil {
		return err
	}
	//& In case the GIO window is closed, I want the application to exit
	//& therefore, returning a non nil error is a lazy workaround however,
	//& it is also triggered when the host exits, which is unintended
	//& I need to find a better way to handle this
	// return errors.New("GIO window closed")
	return nil
}

// func (u *UI) Stop() {
// 	u.Window.Close()
// }

func (u *UI) DispatchInputEvents(ev *types.RemoteEvent) {
	u.Window.GetEventQueue() <- ev
}

func (u *UI) ReceiveInputEvents() <-chan *types.RemoteEvent {
	return u.Window.GetEventQueue()
}

func Dimensions() (width, height int) {
	return utils.GetDisplaySize()
}

func (d *Playback) HandleFrameBuffer(frame []byte) {
	d.Provider.Push(frame)
}

func (d *Playback) Start(width, height, payloadType int, codecName string) error {
	return d.Provider.Start(width, height, payloadType, codecName, d.frames)
}

func (d *Playback) Stop() error {
	return d.Provider.Stop()
}
