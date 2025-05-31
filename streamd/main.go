package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
	"k8s.io/klog"
)

// daemonConfig contains all configurable parameters
type daemonConfig struct {
	listenHTTP string

	// cidr containing ip to listen on
	listenCidr string
	// srt listening port for combined stream
	combPort string
	// srt listening port for presentation stream
	presPort string
	// srt listening port for camera stream
	camPort string

	// ip to listen on
	listenAddr string

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

	flag.StringVar(&d.listenHTTP, "http-port", "8080", "Port at which to listen for HTTP requests")
	flag.StringVar(&d.listenCidr, "listen-cidr", "", "CIDR containing Address to listen for all srt requests. E.g. 100.64.0.0/10 for tailnets. If unset, [::] will be listened on.")
	flag.StringVar(&d.combPort, "port-comb-srt", "7000", "SRT listing port for combined stream")
	flag.StringVar(&d.presPort, "port-present-srt", "7001", "SRT listing port for presentation stream")
	flag.StringVar(&d.camPort, "port-cam-srt", "7002", "SRT listing port for camera stream")
	flag.StringVar(&d.sourcePresent, "source-present", "videotestsrc", "GStreamer element factory name for the presentation source")
	flag.StringVar(&d.sourcePresentOpts, "source-present-opts", "", "GStreamer element properties for presentation source")
	flag.StringVar(&d.sourceCam, "source-cam", "videotestsrc", "GStreamer element factory name for the camera source")
	flag.StringVar(&d.sourceCamOpts, "source-cam-opts", "", "GStreamer element properties for camera source")
	flag.StringVar(&d.sourceAudio, "source-audio", "audiotestsrc", "GStreamer element factory name for the audio source")
	flag.StringVar(&d.sourceAudioOpts, "source-audio-opts", "", "GStreamer element properties for audio source")
	flag.IntVar(&d.videoEncBitrateKbps, "video-enc-bitrate", 6000, "Video encoding bitrate in Kbps")
	flag.IntVar(&d.audioEncBitrateKbps, "audio-enc-bitrate", 96, "Video encoding bitrate in Kbps")
	flag.BoolVar(&d.hwAccel, "hw-accel", false, "Enable hardware acceleration and offload processing tasks onto the GPU or a DSP")
	flag.Parse()

	if d.listenCidr != "" {
		_, cidr, err := net.ParseCIDR(d.listenCidr)
		if err != nil {
			klog.Fatalf("cannot parse cidr %s: %v", d.listenCidr, err)
		}
		ip, err := getIfaceIP(cidr)
		if err != nil {
			klog.Fatalf("unable to obtain ip to listen on matching prefix: %v", err)
		}
		d.listenAddr = ip.String()
		if strings.Count(d.listenAddr, ":") >= 2 { // ipv6
			d.listenAddr = "[" + d.listenAddr + "]"
		}
	} else {
		d.listenAddr = "[::]"
	}

	d.mainloop = glib.NewMainLoop(glib.MainContextDefault(), false)
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)

	// Create and start HTTP server
	h := &httpServer{d}
	h.setupHTTPHandlers()

	klog.Infof("listening for HTTP at %s:%s", d.listenAddr, d.listenHTTP)
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf("%s:%s", d.listenAddr, d.listenHTTP), nil); err != nil {
			klog.Errorf("HTTP listen failed: %v", err)
		}
	}()

	if err := d.runPipeline(); err != nil {
		klog.Errorf("Failed to start pipeline: %v", err)
	}

	// floating around and move outside runPipeline
	go d.metricsProcess(ctx)

	go func() {
		<-ctx.Done() // Wait until the context is cancelled
		// When the context is cancelled, break out of the main loop
		d.mainloop.Quit()
	}()
	d.mainloop.Run()
}

// getIfaceIP returns the first IP address available on the system that is within cidr or an error if none is found.
func getIfaceIP(cidr *net.IPNet) (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if cidr.Contains(ip) {
				return ip, nil
			}
		}
	}
	return nil, fmt.Errorf("no interface in CIDR %s found", cidr)
}
