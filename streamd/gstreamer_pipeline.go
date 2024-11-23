package main

import "github.com/go-gst/go-gst/gst"

var hz30 = rational{30, 1}
var caps1920x1080p30 = videoCapsFilter{Mimetype: "video/x-raw", Width: 1920, Height: 1080, Framerate: hz30}
var caps1440x810p30 = videoCapsFilter{Mimetype: "video/x-raw", Width: 1440, Height: 810, Framerate: hz30}
var caps480x270p30 = videoCapsFilter{Mimetype: "video/x-raw", Width: 480, Height: 270, Framerate: hz30}

var capsStereo48Khz = audioCapsFilter{Mimetype: "audio/x-raw", Channels: 2, Rate: 48000}

// pipeline is the main AV processing pipeline
type pipeline struct {
	constructed bool
	pipeline    *gst.Pipeline

	camSrc     *gst.Bin
	presentSrc *gst.Bin
	audioSrc   *gst.Bin

	// a 1x2 splitter for splitting the presentation stream into two streams.
	// One that is muxed directly, the other is fed into the compositor.
	splitterPresent *gst.Bin
	splitterCam     *gst.Bin
	splitterAudio   *gst.Bin

	muxerPresent   *gst.Bin
	srtPresentSink *gst.Bin

	muxerCam   *gst.Bin
	srtCamSink *gst.Bin

	compositor        *gst.Bin
	muxerCompositor   *gst.Bin
	srtCompositorSink *gst.Bin

	camSrcCaps     videoCapsFilter
	presentSrcCaps videoCapsFilter
	outputCaps     videoCapsFilter

	// Caps for compositor
	presentCompCaps videoCapsFilter
	camCompCaps     videoCapsFilter

	audioCaps audioCapsFilter
}

func newPipeline(hwAccel bool) (*pipeline, error) {
	p := &pipeline{}

	p.outputCaps = caps1920x1080p30
	p.presentSrcCaps = caps1920x1080p30
	p.camSrcCaps = caps1920x1080p30
	p.audioCaps = capsStereo48Khz

	p.camCompCaps = caps480x270p30
	p.presentCompCaps = caps1440x810p30

	var err error

	// TODO(hugo): to much redundancy here. find a way to reduce err checking boilerplate

	p.camSrc, err = newVideoTestSourceBin("present", videoPatternSMPTE, p.camSrcCaps)
	if err != nil {
		return nil, err
	}
	p.presentSrc, err = newVideoTestSourceBin("cam", videoPatternSMPTE, p.presentSrcCaps)
	if err != nil {
		return nil, err
	}
	p.audioSrc, err = newAudioTestSourceBin("master", p.audioCaps)
	if err != nil {
		return nil, err
	}

	p.splitterAudio, err = new1x3SplitterBin("splitter_audio")
	if err != nil {
		return nil, err
	}
	p.splitterPresent, err = new1x2SplitterBin("splitter_present")
	if err != nil {
		return nil, err
	}
	p.splitterCam, err = new1x2SplitterBin("splitter_cam")
	if err != nil {
		return nil, err
	}

	p.muxerPresent, err = newMPEGTSMuxerBin("muxer_present")
	if err != nil {
		return nil, err
	}
	// TODO(hugo): Move into custom bin constructor function with config struct
	p.srtPresentSink, err = gst.NewBinFromString("srtsink uri=srt://:8886 wait-for-connection=false", true)
	if err != nil {
		return nil, err
	}
	p.srtPresentSink.Element.SetProperty("name", "srt_present")

	p.muxerCam, err = newMPEGTSMuxerBin("muxer_cam")
	if err != nil {
		return nil, err
	}
	// TODO(hugo): Move into custom bin constructor function with config struct
	p.srtCamSink, err = gst.NewBinFromString("srtsink uri=srt://:8884 wait-for-connection=false", true)
	if err != nil {
		return nil, err
	}
	p.srtCamSink.Element.SetProperty("name", "srt_cam")

	// Scaling and compositng on GPU results in a big load reduction
	// on the CPU.
	// Keep buffers in VRAM between postproc and compositor
	if hwAccel {
		p.presentCompCaps.Mimetype = "video/x-raw(memory:VAMemory)"
		p.camCompCaps.Mimetype = "video/x-raw(memory:VAMemory)"
	}
	p.compositor, err = newCompositorBin("compositor", combinedViewConfig{
		OutputCaps:       p.outputCaps,
		PresentationCaps: p.presentCompCaps,
		CameraCaps:       p.camCompCaps,
		HwAccel:          hwAccel,
	})
	if err != nil {
		return nil, err
	}

	p.muxerCompositor, err = newMPEGTSMuxerBin("muxer_comp")
	if err != nil {
		return nil, err
	}

	// TODO(hugo): Move into custom bin constructor function with config struct
	p.srtCompositorSink, err = gst.NewBinFromString("srtsink name=srtsink uri=srt://:8888 wait-for-connection=false", true)
	if err != nil {
		return nil, err
	}
	p.srtCompositorSink.Element.SetProperty("name", "srt_combined")

	// Create main pipelines and link bins

	p.pipeline, err = gst.NewPipeline("Pipeline")
	if err != nil {
		return nil, err
	}

	err = p.pipeline.AddMany(
		// Sources
		p.camSrc.Element,
		p.presentSrc.Element,
		p.audioSrc.Element,
		// Splitters
		p.splitterPresent.Element,
		p.splitterCam.Element,
		p.splitterAudio.Element,
		// Processors
		p.compositor.Element,
		// Muxers
		p.muxerCompositor.Element,
		p.muxerPresent.Element,
		p.muxerCam.Element,
		// Sinks
		p.srtCompositorSink.Element,
		p.srtPresentSink.Element,
		p.srtCamSink.Element,
	)
	if err != nil {
		return nil, err
	}

	// Link all bins
	p.presentSrc.Link(p.splitterPresent.Element)
	p.camSrc.Link(p.splitterCam.Element)
	p.audioSrc.Link(p.splitterAudio.Element)

	p.splitterPresent.Link(p.compositor.Element)
	p.splitterPresent.Link(p.muxerPresent.Element)
	p.splitterCam.Link(p.compositor.Element)
	p.splitterCam.Link(p.muxerCam.Element)
	p.splitterAudio.Link(p.muxerCompositor.Element)
	p.splitterAudio.Link(p.muxerPresent.Element)
	p.splitterAudio.Link(p.muxerCam.Element)

	p.compositor.Link(p.muxerCompositor.Element)

	p.muxerCompositor.Link(p.srtCompositorSink.Element)
	p.muxerPresent.Link(p.srtPresentSink.Element)
	p.muxerCam.Link(p.srtCamSink.Element)

	p.constructed = true

	return p, nil
}
