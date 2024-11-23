package main

import (
	"fmt"
	"time"

	"github.com/go-gst/go-gst/gst"
	"k8s.io/klog"
)

type pipelineStats struct {
	warnings   uint64
	qosEvents  map[string]uint64 // key is the name of the element
	minLatency time.Duration
}

func newPipelineStats() pipelineStats {
	return pipelineStats{
		qosEvents: make(map[string]uint64),
	}
}

func (d *daemon) registerBusWatch() bool {
	p := d.pipeline.pipeline

	return p.GetBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageEOS: // When end-of-stream is received stop the main loop
			p.BlockSetState(gst.StateNull)
			d.mainloop.Quit()
		case gst.MessageError: // Error messages are always fatal
			err := msg.ParseError()
			fmt.Println("ERROR:", err.Error())
			if debug := err.DebugString(); debug != "" {
				klog.Info("DEBUG:", debug)
			}
			d.mainloop.Quit()
		case gst.MessageWarning:
			d.mu.Lock()
			d.metrics.pipelineStats.warnings += 1
			d.mu.Unlock()
		// When buffers arrive late in the sink, i.e. when their running-time is
		// smaller than that of the clock, we have a QoS problem
		// https://gstreamer.freedesktop.org/documentation/plugin-development/advanced/qos.html?gi-language=c
		//
		// A useful statistic to have when monitoring the pipeline
		case gst.MessageQoS:
			source := msg.Source()

			d.mu.Lock()
			d.metrics.pipelineStats.qosEvents[source] += 1
			d.mu.Unlock()

			klog.Warning(msg)
		default:
			// All messages implement a Stringer. However, this is
			// typically an expensive thing to do and should be avoided.
			klog.Info(msg)
		}
		return true
	})
}
