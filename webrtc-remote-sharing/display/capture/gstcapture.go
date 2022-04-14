package capture

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/go-gst/gst/app"
)

type GstCapture struct {
	pipeline *gst.Pipeline
	track    *webrtc.TrackLocalStaticSample
}

func (g *GstCapture) createPipeline(width, height int, codecName string) *gst.Pipeline {
	gst.Init(nil)

	pipelineSrc := fmt.Sprintf("gdiscreencapsrc do-timestamp=true cursor=true ! queue ! videoconvert ! "+
		"queue ! videoscale ! video/x-raw,format=I420,framerate=10/1,width=%d,height=%d ! queue", width, height)

	pipelineStr := "appsink name=appsink sync=false qos=true drop=true"

	switch strings.ToLower(codecName) {
	case "video/h264":
		pipelineStr = pipelineSrc + " ! x264enc speed-preset=ultrafast tune=zerolatency psy-tune=film vbv-buf-capacity=50 key-int-max=60 subme=10 ! queue ! " + pipelineStr
	case "video/vp8":
		pipelineStr = pipelineSrc + " ! vp8enc error-resilient=partitions keyframe-max-dist=10 auto-alt-ref=true cpu-used=5 deadline=1 ! queue ! " + pipelineStr
	default:
		log.Fatal("[GST] Unhandled codec " + codecName)
	}

	log.Printf("[GST] Creating pipeline: %s", pipelineStr)

	pipeline, err := gst.NewPipelineFromString(pipelineStr)
	if err != nil {
		log.Fatal("[ERR] Cannot parse launch from pipeline string")
	}
	return pipeline
}

func (g *GstCapture) Start(width, height int, codecName string, track *webrtc.TrackLocalStaticSample) error {
	pipeline := g.createPipeline(width, height, codecName)
	g.pipeline = pipeline
	g.track = track

	log.Println("[GST] Starting pipeline")

	pipeline.GetBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageEOS:
			_ = g.pipeline.BlockSetState(gst.StateNull)
		case gst.MessageError:
			err := msg.ParseError()
			fmt.Println("[ERR]:", err.Error())
		}
		return true
	})
	appsink, err := pipeline.GetElementByName("appsink")
	if err != nil {
		panic("[GST] Cannot get appsink")
	}
	sink := app.SinkFromElement(appsink)

	sink.SetCallbacks(&app.SinkCallbacks{
		NewSampleFunc: func(appSink *app.Sink) gst.FlowReturn {
			sample := appSink.PullSample()
			if sample == nil {
				return gst.FlowEOS
			}
			buffer := sample.GetBuffer()
			if buffer == nil {
				return gst.FlowError
			}

			buf := buffer.Extract(0, buffer.GetSize())

			g.writeToTrack(buf, buffer.Duration())

			return gst.FlowOK
		}})

	return pipeline.SetState(gst.StatePlaying)
}

func (g *GstCapture) Stop() error {
	fmt.Println("[GST] Stopping pipeline")
	return g.pipeline.SetState(gst.StateNull)
}

func (g *GstCapture) writeToTrack(buffer []byte, bufferDuration time.Duration) {
	if err := g.track.WriteSample(media.Sample{Data: buffer, Duration: bufferDuration}); err != nil {
		panic(err)
	}
}
