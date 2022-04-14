package gstappsink

import (
	"fmt"
	"image"
	"log"
	"strings"

	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/app"
)

const (
	VideoWidth  = 1920
	VideoHeight = 1080
)

type Pipeline struct {
	Pipeline *gst.Pipeline
}

func StartMainLoop() {
	log.Println("[INFO] Starting main loop")
	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)
	mainLoop.Run()
}

func CreatePipeline(payloadType int, codecName string) *Pipeline {
	gst.Init(nil)
	pipelineStr := "appsrc format=time is-live=true do-timestamp=true name=src ! application/x-rtp"

	switch strings.ToLower(codecName) {
	case "vp8":
		pipelineStr += fmt.Sprintf(", payload=%d,encoding-name=VP8 ! rtpvp8depay ! queue ! avdec_vp8  ! queue ! videoconvert"+
			" ! queue ! videoscale ! video/x-raw,width=%d,height=%d,format=RGBA,pixel-aspect-ratio=1/1,framrate=10/1"+
			" ! queue ! appsink name=sink sync=false", payloadType, VideoWidth, VideoHeight)

	case "opus":
		pipelineStr += fmt.Sprintf(", payload=%d,encoding-name=OPUS ! rtpopusdepay ! decodebin ! autoaudiosink", payloadType)

	case "vp9":
		pipelineStr += fmt.Sprintf(",payload=%d,encoding-name=VP9 ! rtpvp9depay ! avdec_vp9 ! queue"+
			" ! videoconvert ! videoscale ! video/x-raw,width=%d,height=%d,format=RGBA,pixel-aspect-ratio=1/1,framrate=10/1"+
			" ! queue ! appsink name=sink sync=false", payloadType, VideoWidth, VideoHeight)

	case "h264":
		pipelineStr += fmt.Sprintf(", payload=%d,encoding-name=H264,media=video,profile=high,clock-rate=90000 ! rtph264depay"+
			" ! avdec_h264 output-corrupt=false ! queue ! videoconvert ! video/x-raw,format=AYUV ! gaussianblur sigma=-0.5"+
			" ! queue ! videoconvert ! queue ! videoscale sharpen=1 sharpness=1.5 method=7 n-threads=6 add-borders=false"+
			" ! video/x-raw,width=%d,height=%d,format=RGBA,pixel-aspect-ratio=1/1 ! queue"+
			" ! appsink name=sink sync=false", payloadType, VideoWidth, VideoHeight)

	default:
		log.Fatal("Unhandled codec " + codecName)
	}

	log.Printf("[INFO] Creating pipeline: %s", pipelineStr)

	pipeline, err := gst.NewPipelineFromString(pipelineStr)
	if err != nil {
		log.Fatal("[ERR] Cannot parse launch from pipeline string")
	}
	return &Pipeline{pipeline}
}

func (p *Pipeline) Start(imageChan chan<- *image.NRGBA) {
	log.Println("[INFO] Starting pipeline")
	p.Pipeline.GetBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageEOS:
			_ = p.Pipeline.BlockSetState(gst.StateNull)
		case gst.MessageError:
			err := msg.ParseError()
			fmt.Println("ERROR:", err.Error())
		}
		return true
	})
	_ = p.Pipeline.SetState(gst.StatePlaying)

	log.Println("[INFO] Getting appsink from pipeline")
	sink, err := p.Pipeline.GetElementByName("sink")
	if err != nil {
		log.Fatal("[ERR] Cannot find appsink in pipeline")
	}
	appSink := app.SinkFromElement(sink)

	log.Println("[INFO] Connecting appsink sample handler")
	appSink.SetCallbacks(&app.SinkCallbacks{

		NewSampleFunc: func(appSink *app.Sink) gst.FlowReturn {
			sample := appSink.PullSample()
			if sample == nil {
				return gst.FlowEOS
			}

			buffer := sample.GetBuffer()
			if buffer == nil {
				return gst.FlowError
			}

			samples := buffer.Map(gst.MapRead).Bytes()
			defer buffer.Unmap()

			img := new(image.NRGBA)
			img.Pix = samples
			img.Stride = VideoWidth * 4
			img.Rect = image.Rect(0, 0, VideoWidth, VideoHeight)

			imageChan <- img

			return gst.FlowOK
		},
	})
}

func (p *Pipeline) Stop() {
	fmt.Println("Stopping pipeline")
	_ = p.Pipeline.SetState(gst.StateNull)
}

func (p *Pipeline) Push(buffer []byte) {
	src, err := p.Pipeline.GetElementByName("src")
	if err != nil {
		log.Fatal("[ERR] Cannot find appsrc in pipeline")
	}
	gbuf := gst.NewBufferFromBytes(buffer)

	appSrc := app.SrcFromElement(src)
	_ = appSrc.PushBuffer(gbuf)
}
