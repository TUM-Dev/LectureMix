package main

import (
	"fmt"
	"os"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
)

// daemon is the main service of captured
type daemon struct {
	daemonState
}

// daemonState contains all the state of the daemon
type daemonState struct {
	pipeline *pipeline
	mainloop *glib.MainLoop
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

	if err := d.runPipeline(); err != nil {
		fmt.Println("ERROR!", err)
	}
}
