# streamd - A streaming daemon for streaming multiple live sources.

This daemon handles the processing of audio and video streams from multiple
sources, enabling real-time streaming and recording. It is designed to run on
edge servers, typically in environments such as lecture halls, and uses
off-the-shelf hardware for cost efficiency and flexibility.

## Design

'streamd' uses GStreamer for capturing from live sources, multiplexing,
processing, and streaming via
[SRT](https://www.haivision.com/products/srt-secure-reliable-transport/).

The overall design off the daemon focuses on simplicity and clearity, avoiding
leaky abstractions on top of GStreamer. The user should know the basic concepts
of GStreamer and be able to construct simple pipelines via `gst-launch-1.0`.

The daemon provides a Prometheus metrics endpoint which includes CPU, Memory,
I/O, SRT Connection, and pipeline metrics.

The pipeline's filter graph only contains functionality required for basic
live-streaming. Processing elements are packed into reusable bins (see `gstreamer_bins.go`).
Bins usually resemble physical components such as sources or splitters.

Features like scheduled recordings are out of scope to reduce
complexity and increase resilience. 

## Example Filter graph
![pipeline](../resources/pipeline.svg)
This is a filter graph from a `streamd` instance with hardware acceleration enabled.

### An overly simplified diagram of the software stack

```
    ┌───────────────────────────────────────────────────────┐
    │                        streamd                        │
    └───────────────────────────────────────────────────────┘
    ┌───────────────────────────┐┌──────────────────────────┐
    │    GStreamer Framework    ││                          │
    └───────────────────────────┘│     Decklink Library     │
    ┌───────────────────────────┐│                          │
    │          VA API           ││                          │
    └───────────────────────────┘└──────────────────────────┘
    ┌───────────────────────────────────────────────────────┐
    │  ┌───────────┐┌───────────┐┌────────────┐┌─────────┐  │
    │  │   V4L2    ││ Decklink  ││    ALSA    ││   DRM   │  │
    │  │ Subsystem ││           ││ Subsystem  ││         │  │
    │  └───────────┘└───────────┘└────────────┘└─────────┘  │
    │                        Kernel                         │
    └───────────────────────────────────────────────────────┘
    ┌───────────────────────────────────────────────────────┐
    │                       Hardware                        │
    └───────────────────────────────────────────────────────┘
```

## Building

`streamd` only runs on a Linux distribution and architecture supported by Go. To
compile the daemon, install all dependencies or enter the nix development shell
and run `go build`.

### Dependencies

If you have [nix/nixos](https://nixos.org/) just enter the development shell with `nix-shell` in the repository root.

- Go
- A working C compiler
- pkg-config
- Glib
- GStreamer
- GStreamer Plugins Ugly
- GStreamer Plugins Bad
- GStreamer Plugins Base
- GStreamer Plugins Good
- GStreamer libav

Take a look at `nix.shell` in the repository root for a detailed list.


## Usage

The following flags configure the streamd daemon:

	-listen-http string
	    Address to listen for HTTP requests. Defaults to ":8080".

	-listen-comb-srt string
	    SRT URI for receiving the combined stream. Defaults to "srt://[::]:7000?mode=listener".

	-listen-present-srt string
	    SRT URI for receiving the presentation stream. Defaults to "srt://[::]:7001?mode=listener".

	-listen-cam-srt string
	    SRT URI for receiving the camera stream. Defaults to "srt://[::]:7002?mode=listener".

	-source-present string
	    GStreamer factory name for the presentation source. Defaults to "videotestsrc".

	-source-present-opts string
	    Properties for configuring the presentation source. Defaults to an empty string.

	-source-cam string
	    GStreamer factory name for the camera source. Defaults to "videotestsrc".

	-source-cam-opts string
	    Properties for configuring the camera source. Defaults to an empty string.

	-source-audio string
	    GStreamer factory name for the audio source. Defaults to "audiotestsrc".

	-source-audio-opts string
	    Properties for configuring the audio source. Defaults to an empty string.

	-hw-accel
	    Enables hardware acceleration, offloading processing tasks to the GPU using VA-API. Defaults to false.

For details on SRT URIs, see: https://github.com/hwangsaeul/libsrt/blob/master/docs/srt-live-transmit.md.

## HTTP API

- **`HTTP GET /metrics`**  
  Prometheus metrics endpoint.

- **`HTTP GET /graph?details=<OPTIONAL_DETAILS_QUERY>`**  
  Retrieve the current filter graph as `text/vnd.graphviz`.  

  `OPTIONAL_DETAILS_QUERY` options:  
  - `media-type`  
  - `caps`  
  - `non-default-params`  
  - `states`  
  - `full-params`  
  - `all`  
  - `verbose`

## Examples

### V4L2 and ALSA stream with hardware acceleration:
```
streamd -hw-accel \
    -source-cam v4l2src -source-cam-opts "device=/dev/video0" \
    -source-present v4l2src -source-present-opts "device=/dev/video2" \
    -source-audio alsasrc -source-audio-opts "device=hw:2,0"
```