package play

import (
	"fmt"
	"image"
	"log"
	"strings"

	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/app"
)

type GstPlayback struct {
	pipeline *gst.Pipeline
}

func (g *GstPlayback) createPipeline(width, height, payloadType int, codecName string) *gst.Pipeline {
	gst.Init(nil)

	pipelineStr := "appsrc format=time is-live=true do-timestamp=true name=src ! application/x-rtp"

	switch strings.ToLower(codecName) {
	case "vp8":
		pipelineStr += fmt.Sprintf(", payload=%d,encoding-name=VP8 ! rtpvp8depay ! queue ! avdec_vp8  ! queue ! videoconvert"+
			" ! queue ! videoscale ! video/x-raw,width=%d,height=%d,format=RGBA,pixel-aspect-ratio=1/1,framrate=10/1"+
			" ! queue ! appsink name=sink sync=false", payloadType, width, height)

	case "opus":
		pipelineStr += fmt.Sprintf(", payload=%d,encoding-name=OPUS ! rtpopusdepay ! decodebin ! autoaudiosink", payloadType)

	case "vp9":
		pipelineStr += fmt.Sprintf(",payload=%d,encoding-name=VP9 ! rtpvp9depay ! avdec_vp9 ! queue"+
			" ! videoconvert ! videoscale ! video/x-raw,width=%d,height=%d,format=RGBA,pixel-aspect-ratio=1/1,framrate=10/1"+
			" ! queue ! appsink name=sink sync=false", payloadType, width, height)

	case "h264":
		pipelineStr += fmt.Sprintf(", payload=%d,encoding-name=H264,media=video,profile=high,clock-rate=90000 ! rtph264depay"+
			" ! avdec_h264 output-corrupt=false ! queue ! videoconvert ! video/x-raw,format=AYUV ! gaussianblur sigma=-0.5"+
			" ! queue ! videoconvert ! queue ! videoscale sharpen=1 sharpness=1.5 method=7 n-threads=6 add-borders=false"+
			" ! video/x-raw,width=%d,height=%d,format=RGBA,pixel-aspect-ratio=1/1 ! queue"+
			" ! appsink name=sink sync=false", payloadType, width, height)

	default:
		log.Fatal("Unhandled codec " + codecName)
	}

	log.Printf("[GST] Creating pipeline: %s", pipelineStr)

	pipeline, err := gst.NewPipelineFromString(pipelineStr)
	if err != nil {
		log.Fatal("[ERR] Cannot parse launch from pipeline string")
	}
	return pipeline
}

func (g *GstPlayback) Start(width, height, payloadType int, codecName string, imageChan chan<- *image.NRGBA) error {
	g.pipeline = g.createPipeline(width, height, payloadType, codecName)

	log.Println("[GST] Starting pipeline")

	g.pipeline.GetBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageEOS:
			_ = g.pipeline.BlockSetState(gst.StateNull)
		case gst.MessageError:
			err := msg.ParseError()
			fmt.Println("[ERR]:", err.Error())
		}
		return true
	})

	sink, err := g.pipeline.GetElementByName("sink")
	if err != nil {
		log.Fatal("[ERR] Cannot find appsink in pipeline")
	}
	appSink := app.SinkFromElement(sink)

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
			img.Stride = width * 4
			img.Rect = image.Rect(0, 0, width, height)

			imageChan <- img

			return gst.FlowOK
		},
	})

	return g.pipeline.SetState(gst.StatePlaying)
}

func (g *GstPlayback) Stop() error {
	log.Println("[GST] Stopping pipeline")

	if g.pipeline == nil {
		log.Println("[GST] Pipeline is not initialized, nothing to stop")
		return nil
	}
	return g.pipeline.SetState(gst.StateNull)
}

func (g *GstPlayback) Push(buffer []byte) {
	src, err := g.pipeline.GetElementByName("src")
	if err != nil {
		log.Fatal("[ERR] Cannot find appsrc in pipeline")
	}
	gbuf := gst.NewBufferFromBytes(buffer)

	appSrc := app.SrcFromElement(src)
	_ = appSrc.PushBuffer(gbuf)
}
