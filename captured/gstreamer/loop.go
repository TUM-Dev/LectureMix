package gstreamer

// #cgo pkg-config: glib-2.0
// #include <glib.h>
import "C"

type Loop struct {
	handle *C.GMainLoop
}

func NewLoop() *Loop {
	var loop Loop

	loop.handle = C.g_main_loop_new(nil, 0)
	return &loop
}

func (l *Loop) Run() {
	C.g_main_loop_run(l.handle)
}

func (l *Loop) Unref() {
	C.g_main_loop_unref(l.handle)
}
