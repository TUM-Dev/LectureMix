<p align="center">
	A capture agent for recording and streaming events.
</p>

## Introduction

Note that this project is currently in alpha stage and not ready for production.

### streamd

`streamd` is the streaming daemon which does all the heavy lifting. It is typically running on
an edge server in a lecture hall, processing audio and video from one or more capture cards.

The edge server is built from off the shelf hardware to reduce cost and increase flexibilty.
