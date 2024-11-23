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
	Other     string
}

// Returns a description of the VideoCapsFilter instance that can be used in a
// pipeline description.
func (c *videoCapsFilter) string() string {
	str := fmt.Sprintf("\"%s,width=%d,height=%d,framerate=%d/%d", c.Mimetype, c.Width, c.Height, c.Framerate.Nominator, c.Framerate.Denominator)
	if c.Other != "" {
		str = str + "," + c.Other
	}

	return str + "\""
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
func newVideoTestSourceBin(name string, pattern videoPattern, caps videoCapsFilter) (*gst.Bin, error) {
	desc := fmt.Sprintf("videotestsrc name=videotestsrc_%s pattern=%d ! capsfilter name=capsfilter_%s caps=%s", name, pattern, name, caps.string())

	// Automatically create ghost-pads for all unlinked pads. In this case this
	// is the capsfilter sink pad.
	bin, err := gst.NewBinFromString(desc, true)
	if err != nil {
		return nil, err
	}
	bin.Element.SetProperty("name", name)

	return bin, nil
}

// Creates a V4L2SourceBin with a single sink ghost-pad
func newV4L2SourceBin(name string, device string, caps videoCapsFilter) (*gst.Bin, error) {
	desc := fmt.Sprintf(
		"v4l2src name=v4l2src_%s device=%s ! videoconvertscale name=videoconvertscale_%s ! capsfilter name=capsfilter_%s caps=%s",
		name,
		device,
		name,
		name,
		caps.string(),
	)
	bin, err := gst.NewBinFromString(desc, true)
	bin.Element.SetProperty("name", name)
	return bin, err
}

// Creates an AudioTestSourceBin with a single sink ghost-pad
func newAudioTestSourceBin(name string, caps audioCapsFilter) (*gst.Bin, error) {
	desc := fmt.Sprintf("audiotestsrc name=audiotestsrc_%s ! capsfilter name=capsfilter_%s caps=%s",
		name,
		name,
		caps.string(),
	)

	// Automatically create ghost-pads for all unlinked pads. In this case this
	// is the capsfilter sink pad.
	bin, err := gst.NewBinFromString(desc, true)
	if err != nil {
		return nil, err
	}
	bin.Element.SetProperty("name", name)
	return bin, err
}

func newALSASourceBin(name string, device string, caps audioCapsFilter) (*gst.Bin, error) {
	alsasrcName := "alsasrc_" + name
	queue0Name := "queue0_" + name
	audioconvertName := "audioconvert_" + name
	audioresampleName := "audioresample_" + name
	audiorateName := "audiorate_" + name
	capsfilterName := "capsfilter_" + name
	queue1Name := "queue1_" + name

	// Isolating conversion, resampling, and timestamping to a new thread is necessary.
	// Leaving out one queue results in clock problems.
	desc := fmt.Sprintf("alsasrc name=%s device=%s ! queue name=%s ! audioconvert name=%s ! audioresample name=%s ! audiorate name=%s ! capsfilter name=%s caps=%s ! queue name=%s",
		alsasrcName,
		device,
		queue0Name,
		audioconvertName,
		audioresampleName,
		audiorateName,
		capsfilterName,
		caps.string(),
		queue1Name,
	)
	bin, err := gst.NewBinFromString(desc, true)
	if err != nil {
		return nil, err
	}
	bin.Element.SetProperty("name", name)
	return bin, err
}

type combinedViewConfig struct {
	OutputCaps       videoCapsFilter
	CameraCaps       videoCapsFilter
	PresentationCaps videoCapsFilter
	HwAccel          bool
}

func createGhostPad(elementName string, elementPad string, ghostPad string, bin *gst.Bin) error {
	// Retrieve the two sink pads from the queue elements
	element, err := bin.GetElementByName(elementName)
	if err != nil {
		return err
	}

	return createGhostPadWithElement(element, elementPad, ghostPad, bin)
}

func createGhostPadWithElement(element *gst.Element, elementPad string, ghostPad string, bin *gst.Bin) error {
	static_pad := element.GetStaticPad(elementPad)
	if static_pad == nil {
		return fmt.Errorf("failed to get static pad '%s' from '%s' element", elementPad, element.GetName())
	}

	return createGhostPadWithPad(static_pad, ghostPad, bin)
}

func createGhostPadWithPad(pad *gst.Pad, ghostPad string, bin *gst.Bin) error {
	ghost_pad := gst.NewGhostPad(ghostPad, pad)
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

func newCompositorBin(config combinedViewConfig) (*gst.Bin, error) {
	sink1_xpos := config.OutputCaps.Width - config.CameraCaps.Width

	comp := "compositor background=black"
	scaler := "videoconvertscale"
	if config.HwAccel {
		comp = "vacompositor"
		scaler = "vapostproc"
	}

	// TODO(hugo): add autoincrementing names to avoid a naming conflict in case we have multiple combined views
	comp_desc := fmt.Sprintf("%s name=comp sink_1::xpos=%d ! capsfilter name=capsfilter_compositor caps=%s", comp, sink1_xpos, config.OutputCaps.string())
	sink0_desc := fmt.Sprintf("queue name=sink0_queue_compositor ! %s add-borders=1 ! capsfilter caps=%s ! comp.sink_0", scaler, config.PresentationCaps.string())
	sink1_desc := fmt.Sprintf("queue name=sink1_queue_compositor ! %s add-borders=1 ! capsfilter caps=%s ! comp.sink_1", scaler, config.CameraCaps.string())

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

func new1x2SplitterBin(name string) (*gst.Bin, error) {
	bin := gst.NewBin(name)
	if bin == nil {
		return nil, fmt.Errorf("cannot create bin '%s'", name)
	}

	teeName := "tee_" + name
	tee, err := gst.NewElementWithName("tee", teeName)
	if err != nil {
		return nil, err
	}

	if err = bin.Add(tee); err != nil {
		return nil, err
	}

	src0 := tee.GetRequestPad("src_0")
	src1 := tee.GetRequestPad("src_1")
	if src0 == nil || src1 == nil {
		return nil, fmt.Errorf("failed to request 'src_0' or 'src_1' pad from '%s'", teeName)
	}

	err = createGhostPadWithElement(tee, "sink", "sink", bin)
	if err != nil {
		return nil, err
	}
	err = createGhostPadWithPad(src0, "src_0", bin)
	if err != nil {
		return nil, err
	}
	err = createGhostPadWithPad(src1, "src_1", bin)
	if err != nil {
		return nil, err
	}

	return bin, nil
}

// TODO(hugo): make splitter constructor generic or argument based
func new1x3SplitterBin(name string) (*gst.Bin, error) {
	bin := gst.NewBin(name)
	if bin == nil {
		return nil, fmt.Errorf("cannot create bin '%s'", name)
	}

	teeName := "tee_" + name
	tee, err := gst.NewElementWithName("tee", teeName)
	if err != nil {
		return nil, err
	}

	if err = bin.Add(tee); err != nil {
		return nil, err
	}

	src0 := tee.GetRequestPad("src_0")
	src1 := tee.GetRequestPad("src_1")
	src2 := tee.GetRequestPad("src_2")
	if src0 == nil || src1 == nil || src2 == nil {
		return nil, fmt.Errorf("failed to request 'src_0', 'src_1', 'src_2' pad from '%s'", teeName)
	}

	err = createGhostPadWithElement(tee, "sink", "sink", bin)
	if err != nil {
		return nil, err
	}
	err = createGhostPadWithPad(src0, "src_0", bin)
	if err != nil {
		return nil, err
	}
	err = createGhostPadWithPad(src1, "src_1", bin)
	if err != nil {
		return nil, err
	}
	err = createGhostPadWithPad(src2, "src_2", bin)
	if err != nil {
		return nil, err
	}

	return bin, nil
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
