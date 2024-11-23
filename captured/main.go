package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
	"k8s.io/klog"
)

// daemon is the main service of captured
type daemon struct {
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

// daemonController provides a small interface for the HTTP server
type daemonController interface {
	metricsSnapshot() metrics
	graph(details gst.DebugGraphDetails) string
	srtCompSinkStats() (*srtStats, error) // TODO: rename
}

func (d *daemon) metricsSnapshot() metrics {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.metrics
}

func (d *daemon) graph(details gst.DebugGraphDetails) string {
	d.mu.Lock()
	p := d.pipeline.pipeline
	d.mu.Unlock()

	return p.DebugBinToDotData(details)
}

func (d *daemon) srtCompSinkStats() (*srtStats, error) {
	d.mu.Lock()
	sink := d.pipeline.srtCompositorSink
	constructed := d.pipeline.constructed
	d.mu.Unlock()

	if constructed != true {
		return nil, errors.New("pipeline not constructed")
	}

	// GStreamer elements are thread-safe
	elem, err := sink.GetElementByName("srtsink")
	if err != nil {
		return nil, err
	}
	val, err := elem.GetProperty("stats")
	if err != nil {
		return nil, err
	}

	s, ok := val.(*gst.Structure)
	if ok != true {
		return nil, errors.New("'stats' value is not '*gst.Structure'")
	}

	return newSRTStatsFromStructure(s)
}

func (d *daemon) runPipeline() error {
	gst.Init(&os.Args)

	var err error
	d.pipeline, err = newPipeline()
	if err != nil {
		return err
	}

	p := d.pipeline.pipeline

	d.metrics.pipelineStats = newPipelineStats()
	d.registerBusWatch()

	// Start the pipeline
	p.SetState(gst.StatePlaying)

	// TODO(hugo): ctx is currently useless as we have the pesky gmainloop
	// floating around and move outside runPipeline
	go d.metricsProcess(context.TODO())

	// Block on the main loop
	return d.mainloop.RunError()
}

func main() {
	d := &daemon{}

	d.mainloop = glib.NewMainLoop(glib.MainContextDefault(), false)

	// Create and start HTTP server
	h := &httpServer{d}
	h.setupHTTPHandlers()

	addr := ":8080"
	klog.Infof("listening for HTTP at %s", addr)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			klog.Errorf("HTTP listen failed: %v", err)
		}
	}()

	if err := d.runPipeline(); err != nil {
		fmt.Println("ERROR!", err)
	}

}
