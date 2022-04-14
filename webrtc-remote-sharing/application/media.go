package application

import "github.com/remygo/display"

type MediaComponents struct {
	Playback *PlaybackComponent
	Capture  *CaptureComponent
	Done     chan struct{}
}

func newMediaComponents() *MediaComponents {
	return &MediaComponents{
		getPlaybackComponent(),
		getCaptureComponent(),
		make(chan struct{}, 1),
	}
}

type PlaybackComponent struct {
	*display.Playback
	*display.UI
}

type CaptureComponent struct {
	*display.Capture
}

func getPlaybackComponent() *PlaybackComponent {
	playback, window := display.NewPlaybackProvider()
	return &PlaybackComponent{
		playback,
		window,
	}
}

func getCaptureComponent() *CaptureComponent {
	return &CaptureComponent{
		display.NewCaptureProvider(),
	}
}
