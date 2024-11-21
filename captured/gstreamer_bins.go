package main

import (
	"fmt"

	"github.com/go-gst/go-gst/gst"
)

type rational struct {
	Nominator   int
	Denominator int
}

type videoPattern int

// Video test patterns for the VideoTestSourceBin
// Maps one-to-one to the GStreamer mappings
const (
	videoPatternSMPTE           videoPattern = iota // SMPTE 100% color bars
	videoPatternSnow                                // Random (television snow)
	videoPatternBlack                               // 100% Black
	videoPatternWhite                               // 100% White
	videoPatternRed                                 // Red
	videoPatternGreen                               // Green
	videoPatternBlue                                // Blue
	videoPatternCheckers1                           // Checkers 1px
	videoPatternCheckers2                           // Checkers 2px
	videoPatternCheckers4                           // Checkers 4px
	videoPatternCheckers8                           // Checkers 8px
	videoPatternCircular                            // Circular
	videoPatternBlink                               // Blink
	videoPatternSMPTE75                             // SMPTE 75% color bars
	videoPatternZonePlate                           // Zone plate
	videoPatternGamut                               // Gamut checkers
	videoPatternChromaZonePlate                     // Chroma zone plate
	videoPatternSolidColor                          // Solid color
	videoPatternBall                                // Moving ball
	videoPatternSMPTE100                            // SMPTE 100% color bars
	videoPatternBar                                 //  Bar
	videoPatternPinwheel                            //  Pinwheel
	videoPatternSpokes                              // Spokes
	videoPatternGradient                            //  Gradient
	videoPatternColors                              // Colors
	videoPatternSMPTERp219                          // SMPTE test pattern, RP 219 conformant
)

// A VideoCapsFilter enforces limitation of formats in the process of linking pads.
type videoCapsFilter struct {
	Mimetype  string
	Width     int
	Height    int
	Framerate rational
}

// Returns a description of the VideoCapsFilter instance that can be used in a
// pipeline description.
func (c *videoCapsFilter) string() string {
	return fmt.Sprintf("%s,width=%d,height=%d,framerate=%d/%d", c.Mimetype, c.Width, c.Height, c.Framerate.Nominator, c.Framerate.Denominator)
}

// An audioCapsFilter enforces limitation of formats in the process of linking pads.
type audioCapsFilter struct {
	Mimetype string
	Channels int
	Rate     int
}

// Returns a description of the AudioCapsfilter instance that can be used in a
// pipeline description.
func (c *audioCapsFilter) string() string {
	return fmt.Sprintf("%s,channels=%d,rate=%d", c.Mimetype, c.Channels, c.Rate)
}

// Creates a VideoTestSourceBin with a single sink ghost-pad
func newVideoTestSourceBin(pattern videoPattern, caps videoCapsFilter) (*gst.Bin, error) {
	desc := fmt.Sprintf("videotestsrc pattern=%d ! capsfilter caps=%s", pattern, caps.string())

	// Automatically create ghost-pads for all unlinked pads. In this case this
	// is the capsfilter sink pad.
	bin, err := gst.NewBinFromString(desc, true)
	return bin, err
}

// Creates a V4L2SourceBin with a single sink ghost-pad
func newV4L2SourceBin(device string, caps videoCapsFilter) (*gst.Bin, error) {
	desc := fmt.Sprintf("v4l2src device=%s ! videoconvertscale ! capsfilter caps=%s", device, caps.string())
	bin, err := gst.NewBinFromString(desc, true)
	return bin, err
}

// Creates an AudioTestSourceBin with a single sink ghost-pad
func newAudioTestSourceBin(caps audioCapsFilter) (*gst.Bin, error) {
	desc := fmt.Sprintf("audiotestsrc ! capsfilter caps=%s", caps.string())

	// Automatically create ghost-pads for all unlinked pads. In this case this
	// is the capsfilter sink pad.
	bin, err := gst.NewBinFromString(desc, true)
	return bin, err
}

func newALSASourceBin(device string, caps audioCapsFilter) (*gst.Bin, error) {
	// Isolating conversion, resampling, and timestamping to a new thread is necessary.
	// Leaving out one queue results in clock problems.
	desc := fmt.Sprintf("alsasrc device=%s ! queue ! audioconvert ! audioresample ! audiorate ! capsfilter caps=%s ! queue ", device, caps.string())
	bin, err := gst.NewBinFromString(desc, true)
	return bin, err
}

type combinedViewConfig struct {
	OutputCaps       videoCapsFilter
	CameraCaps       videoCapsFilter
	PresentationCaps videoCapsFilter
}

func createGhostPad(elementName string, elementPad string, ghostPad string, bin *gst.Bin) error {
	// Retrieve the two sink pads from the queue elements
	element, err := bin.GetElementByName(elementName)
	if err != nil {
		return err
	}

	static_pad := element.GetStaticPad(elementPad)
	if static_pad == nil {
		return fmt.Errorf("failed to get static pad '%s' from '%s' element", elementPad, elementName)
	}

	ghost_pad := gst.NewGhostPad(ghostPad, static_pad)
	if ghost_pad == nil {
		return fmt.Errorf("unable to create ghost pad '%s'", ghostPad)
	}

	if ghost_pad.SetActive(true) == false {
		return fmt.Errorf("failed to activate ghost pad '%s'", ghostPad)
	}

	if bin.AddPad(ghost_pad.Pad) == false {
		return fmt.Errorf("failed to add pad to bin '%s'", bin.GetName())
	}

	return nil
}

func newCombinedViewBin(config combinedViewConfig) (*gst.Bin, error) {
	sink1_xpos := config.OutputCaps.Width - config.CameraCaps.Width

	// TODO(hugo): add autoincrementing names to avoid a naming conflict in case we have multiple combined views
	comp_desc := fmt.Sprintf("compositor name=comp sink_1::xpos=%d background=black ! capsfilter name=capsfilter_compositor caps=%s", sink1_xpos, config.OutputCaps.string())
	sink0_desc := fmt.Sprintf("queue name=sink0_queue_compositor ! videoconvertscale add-borders=1 ! capsfilter caps=%s ! comp.sink_0", config.PresentationCaps.string())
	sink1_desc := fmt.Sprintf("queue name=sink1_queue_compositor ! videoconvertscale add-borders=1 ! capsfilter caps=%s ! comp.sink_1", config.CameraCaps.string())

	// Do not automatically create Ghostpads, as sink ghost-pads are not configured correctly.
	bin, err := gst.NewBinFromString(comp_desc+" "+sink0_desc+" "+sink1_desc, false)
	if err != nil {
		return nil, err
	}

	err = createGhostPad("sink0_queue_compositor", "sink", "sink0", bin)
	if err != nil {
		return nil, err
	}
	err = createGhostPad("sink1_queue_compositor", "sink", "sink1", bin)
	if err != nil {
		return nil, err
	}
	err = createGhostPad("capsfilter_compositor", "src", "src", bin)
	if err != nil {
		return nil, err
	}

	return bin, err
}

func newMPEGTSMuxerBin() (*gst.Bin, error) {
	mux_desc := "mpegtsmux name=mux"
	// TODO(hugo): we might want to move the encoders into seperate bins
	// TODO(hugo): check if we need to create a thread boundry with a queue between encoders and muxer
	audio_queue_desc := "queue name=audio_queue_mux ! avenc_aac ! mux."
	video_queue_desc := "queue name=video_queue_mux ! x264enc ! mux."

	bin, err := gst.NewBinFromString(mux_desc+" "+audio_queue_desc+" "+video_queue_desc, false)
	if err != nil {
		return nil, err
	}

	err = createGhostPad("audio_queue_mux", "sink", "audio_sink", bin)
	if err != nil {
		return nil, err
	}
	err = createGhostPad("video_queue_mux", "sink", "video_sink", bin)
	if err != nil {
		return nil, err
	}
	err = createGhostPad("mux", "src", "src", bin)
	if err != nil {
		return nil, err
	}

	return bin, err
}
