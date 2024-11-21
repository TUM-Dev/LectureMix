package main

import (
	"fmt"
	"os"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
)

// daemon is the main service of captured
type daemon struct {
	//mu sync.RWMutex
	daemonState
}

// unrefs all unmanaged objects
// daemon must not be used after calling unref()
func (d *daemon) unref() {
}

// daemonState contains all the state of the daemon
// A copy of it may be requested for consumers
type daemonState struct {
}

func videoSourceTestBin() (*gst.Bin, error) {
	bin, err := gst.NewBinFromString("videotestsrc ! capsfilter name=capsfilter caps=video/x-raw,width=1920,height=1080,framerate=30/1", true)
	if err != nil {
		return nil, err
	}

	return bin, nil
}

func runPipeline(mainLoop *glib.MainLoop) error {
	gst.Init(&os.Args)

	videoSourceBin, err := videoSourceTestBin()
	if err != nil {
		return err
	}

	pipeline, err := gst.NewPipeline("pipeline")
	if err != nil {
		return err
	}

	sink, err := gst.NewElement("autovideosink")
	if err != nil {
		return err
	}

	pipeline.Add(videoSourceBin.Element)
	pipeline.Add(sink)

	videoSourceBin.Link(sink)

	pipeline.GetBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageEOS: // When end-of-stream is received stop the main loop
			pipeline.BlockSetState(gst.StateNull)
			mainLoop.Quit()
		case gst.MessageError: // Error messages are always fatal
			err := msg.ParseError()
			fmt.Println("ERROR:", err.Error())
			if debug := err.DebugString(); debug != "" {
				fmt.Println("DEBUG:", debug)
			}
			mainLoop.Quit()
		default:
			// All messages implement a Stringer. However, this is
			// typically an expensive thing to do and should be avoided.
			fmt.Println(msg)
		}
		return true
	})

	// Start the pipeline
	pipeline.SetState(gst.StatePlaying)

	// Block on the main loop
	return mainLoop.RunError()
}

func main() {
	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)

	if err := runPipeline(mainLoop); err != nil {
		fmt.Println("ERROR!", err)
	}
}
