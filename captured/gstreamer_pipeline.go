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

	camSrc            *gst.Bin
	presentSrc        *gst.Bin
	audioSrc          *gst.Bin
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

func newPipeline() (*pipeline, error) {
	p := &pipeline{}

	p.outputCaps = caps1920x1080p30
	p.presentSrcCaps = caps1920x1080p30
	p.camSrcCaps = caps1920x1080p30
	p.presentCompCaps = caps1440x810p30
	p.camCompCaps = caps480x270p30
	p.audioCaps = capsStereo48Khz

	var err error

	p.camSrc, err = newVideoTestSourceBin(videoPatternSMPTE, p.camSrcCaps)
	if err != nil {
		return nil, err
	}
	p.presentSrc, err = newVideoTestSourceBin(videoPatternSMPTE, p.presentSrcCaps)
	if err != nil {
		return nil, err
	}
	p.audioSrc, err = newAudioTestSourceBin(p.audioCaps)
	if err != nil {
		return nil, err
	}

	p.compositor, err = newCombinedViewBin(combinedViewConfig{
		OutputCaps:       p.outputCaps,
		PresentationCaps: p.presentCompCaps,
		CameraCaps:       p.camCompCaps,
	})
	if err != nil {
		return nil, err
	}

	p.muxerCompositor, err = newMPEGTSMuxerBin()
	if err != nil {
		return nil, err
	}

	// TODO(hugo): Move into custom bin constructor function with config struct
	p.srtCompositorSink, err = gst.NewBinFromString("srtsink name=srtsink uri=srt://:8888 wait-for-connection=false", true)
	if err != nil {
		return nil, err
	}

	// Create main pipelines and link bins

	p.pipeline, err = gst.NewPipeline("Pipeline")
	if err != nil {
		return nil, err
	}

	err = p.pipeline.AddMany(
		p.camSrc.Element,
		p.presentSrc.Element,
		p.audioSrc.Element,
		p.compositor.Element,
		p.muxerCompositor.Element,
		p.srtCompositorSink.Element,
	)
	if err != nil {
		return nil, err
	}

	// Link all bins
	p.presentSrc.Link(p.compositor.Element)
	p.camSrc.Link(p.compositor.Element)
	p.compositor.Link(p.muxerCompositor.Element)
	p.audioSrc.Link(p.muxerCompositor.Element)
	p.muxerCompositor.Link(p.srtCompositorSink.Element)

	p.constructed = true

	return p, nil
}
