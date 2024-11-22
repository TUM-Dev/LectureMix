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
	srtCompSinkStats() (*srtStats, error)
}

func (d *daemon) metricsSnapshot() metrics {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.metrics
}

func (d *daemon) srtCompSinkStats() (*srtStats, error) {
	d.mu.Lock()
	sink := d.pipeline.srtCompositorSink
	d.mu.Unlock()

	// GStreamer elements are thread-safe
	val, err := sink.GetProperty("stats")
	if err != nil {
		return nil, err
	}

	if s, ok := val.(*gst.Structure); ok != false {
		return nil, errors.New("'stats' value is not '*gst.Structure'")
	} else {
		return newSRTStatsFromStructure(s)
	}
}

func (d *daemon) runPipeline() error {
	gst.Init(&os.Args)

	var err error
	d.pipeline, err = newPipeline()
	if err != nil {
		return err
	}

	p := d.pipeline.pipeline

	p.GetBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageEOS: // When end-of-stream is received stop the main loop
			p.BlockSetState(gst.StateNull)
			d.mainloop.Quit()
		case gst.MessageError: // Error messages are always fatal
			err := msg.ParseError()
			fmt.Println("ERROR:", err.Error())
			if debug := err.DebugString(); debug != "" {
				fmt.Println("DEBUG:", debug)
			}
			d.mainloop.Quit()
		default:
			// All messages implement a Stringer. However, this is
			// typically an expensive thing to do and should be avoided.
			fmt.Println(msg)
		}
		return true
	})

	// Start the pipeline
	p.SetState(gst.StatePlaying)

	// Block on the main loop
	return d.mainloop.RunError()
}

func main() {
	d := &daemon{}

	d.mainloop = glib.NewMainLoop(glib.MainContextDefault(), false)

	// TODO(hugo): ctx is currently useless as we have the pesky gmainloop
	// floating around
	go d.metricsProcess(context.TODO())

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
