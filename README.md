<p align="center">
	An open-source live-streaming stack for recording and streaming lectures.
</p>

## Introduction

LectureMix aims to democratize the streaming and recording of lectures, by using
off-the-shelf hardware and the [GStreamer](https://gstreamer.freedesktop.org/)
framework.

The idea is simple, build a server with capture and sound card(s) and run the
LectureMix stack on a Linux distribution of your choice.

## Daemons

### streamd

`streamd` is the streaming daemon which does all the heavy lifting. It is typically running on
an edge server in a lecture hall, processing audio and video from one or more capture cards.

See [streamd/README.md](streamd/README.md) for more information.

### captured

`captured` synchronises with an iCalendar to get the recording schedule. When an
event starts, `captured` connects to one more more SRT streams and writes the
MPEG transport stream (TS) to disk.

Note that work on `captured` has not yet started.

### transcoded

The processing servers are quite powerful and idle when nobody connects to the SRT stream. One can utilize the processing power to transcode existing VoDs.

Work on this daemon has not yet started.
