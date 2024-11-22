package main

import (
	"fmt"
	"sync/atomic"

	"github.com/go-gst/go-gst/gst"
)

type pipelineStats struct {
	warnings  atomic.Uint64
	qosEvents atomic.Uint64
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
				fmt.Println("DEBUG:", debug)
			}
			d.mainloop.Quit()
		case gst.MessageWarning:
			d.metrics.pipelineStats.warnings.Add(1)
		// When buffers arrive late in the sink, i.e. when their running-time is
		// smaller than that of the clock, we have a QoS problem
		// https://gstreamer.freedesktop.org/documentation/plugin-development/advanced/qos.html?gi-language=c
		//
		// A useful statistic to have when monitoring the pipeline
		case gst.MessageQoS:
			d.metrics.pipelineStats.qosEvents.Add(1)
			_ = msg.ParseQoS()
		default:
			// All messages implement a Stringer. However, this is
			// typically an expensive thing to do and should be avoided.
			fmt.Println(msg)
		}
		return true
	})
}
