package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
	"k8s.io/klog"
)

// daemonConfig contains all configurable parameters
type daemonConfig struct {
	listenHTTP string

	// srt listening URI for combined stream
	listenCombSRT string
	// srt listening URI for presentation stream
	listenPresentSRT string
	// srt listening URI for camera stream
	listenCamSRT string

	// the GStreamer factory name for the presentation source element
	sourcePresent string
	// the GStreamer properties for the presentation source element
	sourcePresentOpts string
	// the GStreamer factory name for the camera source element
	sourceCam string
	// the GStreamer properties for the camera source element
	sourceCamOpts string
	// the GStreamer factory name for the master audio source element
	sourceAudio string
	// the GStreamer properties for the master audio source element
	sourceAudioOpts string

	videoEncBitrateKbps int
	audioEncBitrateKbps int

	// whether to enable hardware acceleration in the filter graph
	hwAccel bool
}

// daemon is the main service of streamd
type daemon struct {
	daemonConfig
	// mu guards the state below.
	mu sync.RWMutex
	daemonState
}

// daemonState contains all the state of the daemon
type daemonState struct {
	pipeline *pipeline
	mainloop *glib.MainLoop
	metrics  metrics
}

// daemonController provides a MT-safe interface for other
// parts of the application (e.g. HTTP server or metrics collector)
type daemonController interface {
	metricsSnapshot() metrics
	graph(details gst.DebugGraphDetails) string
	srtStatistics() ([]*srtStats, error)
}

func (d *daemon) srtStatistics() ([]*srtStats, error) {
	d.mu.Lock()
	combBin := d.pipeline.srtCompositorSink
	presentBin := d.pipeline.srtPresentSink
	camBin := d.pipeline.srtCamSink
	d.mu.Unlock()

	combStats, err := getSRTStatistics(combBin)
	if err != nil {
		return nil, err
	}
	presentStats, err := getSRTStatistics(presentBin)
	if err != nil {
		return nil, err
	}
	camStats, err := getSRTStatistics(camBin)
	if err != nil {
		return nil, err
	}

	return []*srtStats{combStats, presentStats, camStats}, nil
}

// get a snapshot of the current metrics
func (d *daemon) metricsSnapshot() metrics {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.metrics
}

// get the current filter graph as 'text/vnd.graphviz'
func (d *daemon) graph(details gst.DebugGraphDetails) string {
	d.mu.Lock()
	p := d.pipeline.pipeline
	d.mu.Unlock()

	return p.DebugBinToDotData(details)
}

func (d *daemon) runPipeline() error {
	gst.Init(&os.Args)

	var err error
	d.pipeline, err = newPipeline(&d.daemonConfig)
	if err != nil {
		return err
	}

	p := d.pipeline.pipeline

	d.metrics.pipelineStats = newPipelineStats()
	d.registerBusWatch()

	// Start the pipeline
	p.SetState(gst.StatePlaying)

	return nil
}

func main() {
	d := &daemon{}

	flag.StringVar(&d.listenHTTP, "listen-http", ":8080", "Address at which to listen for HTTP requests")
	// See https://github.com/hwangsaeul/libsrt/blob/master/docs/srt-live-transmit.md for more information on SRT URIs
	flag.StringVar(&d.listenCombSRT, "listen-comb-srt", "srt://[::]:7000?mode=listener", "SRT listing address for combined stream")
	flag.StringVar(&d.listenPresentSRT, "listen-present-srt", "srt://[::]:7001?mode=listener", "SRT listing address for presentation stream")
	flag.StringVar(&d.listenCamSRT, "listen-cam-srt", "srt://[::]:7002?mode=listener", "SRT listing address for camera stream")
	flag.StringVar(&d.sourcePresent, "source-present", "videotestsrc", "GStreamer element factory name for the presentation source")
	flag.StringVar(&d.sourcePresentOpts, "source-present-opts", "", "GStreamer element properties for presentation source")
	flag.StringVar(&d.sourceCam, "source-cam", "videotestsrc", "GStreamer element factory name for the camera source")
	flag.StringVar(&d.sourceCamOpts, "source-cam-opts", "", "GStreamer element properties for camera source")
	flag.StringVar(&d.sourceAudio, "source-audio", "audiotestsrc", "GStreamer element factory name for the audio source")
	flag.StringVar(&d.sourceAudioOpts, "source-audio-opts", "", "GStreamer element properties for audio source")
	flag.IntVar(&d.videoEncBitrateKbps, "video-enc-bitrate", 8000, "Video encoding bitrate in Kbps")
	flag.IntVar(&d.audioEncBitrateKbps, "audio-enc-bitrate", 96, "Video encoding bitrate in Kbps")
	flag.BoolVar(&d.hwAccel, "hw-accel", false, "Enable hardware acceleration and offload processing tasks onto the GPU or a DSP")
	flag.Parse()

	d.mainloop = glib.NewMainLoop(glib.MainContextDefault(), false)
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	// Create and start HTTP server
	h := &httpServer{d}
	h.setupHTTPHandlers()

	klog.Infof("listening for HTTP at %s", d.listenHTTP)
	go func() {
		if err := http.ListenAndServe(d.listenHTTP, nil); err != nil {
			klog.Errorf("HTTP listen failed: %v", err)
		}
	}()

	if err := d.runPipeline(); err != nil {
		klog.Errorf("Failed to start pipeline: %v", err)
	}

	// floating around and move outside runPipeline
	go d.metricsProcess(ctx)

	// bridge the mainloop with our go context
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// this is essentially what g_main_loop_run does with some locking overhead
			d.mainloop.GetContext().Iteration(false)
			time.Sleep(time.Millisecond * 50)
		}
	}
}
