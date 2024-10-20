package main

import (
	"flag"
	"fmt"

	"github.com/TUM-Dev/captureagent/captured/gstreamer"

	"k8s.io/klog"
)

// daemon is the main service of captured
type daemon struct {
	//mu sync.RWMutex
	daemonState
}

// unrefs all unmanaged objects
// daemon must not be used after calling unref()
func (d *daemon) unref() {
	d.daemonState.pipeline.Unref()
	d.daemonState.loop.Unref()
}

// daemonState contains all the state of the daemon
// A copy of it may be requested for consumers
type daemonState struct {
	// glib main loop for installing signal handlers
	loop *gstreamer.Loop
	// the main AV pipeline for capturing, processing, and
	// distributing audio and video streams
	pipeline *gstreamer.Pipeline
	// the GStreamer pipeline description
	description string
}

func main() {
	d := daemon{}

	flag.StringVar(&d.daemonState.description, "pipeline_description", "",
		"The complete GStreamer pipeline description for the media pipeline")
	flag.Parse()

	if d.description == "" {
		klog.Exitf("Please supply a pipeline description of the AV pipeline")
	}

	// setup up daemon
	var err error
	d.daemonState.pipeline, err = gstreamer.NewPipeline("videotestsrc ! autovideosink")
	if err != nil {
		fmt.Printf("Error creating pipeline %v\n", err)
		return
	}

	d.daemonState.loop = gstreamer.NewLoop()

	// Run the glib main loop
	d.loop.Run()

	d.unref()
}
