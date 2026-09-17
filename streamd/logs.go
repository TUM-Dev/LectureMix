package main

import (
	"net/http"
	"strings"
	"sync"
)

const logBufferSize = 10000

type logBuffer struct {
	mu    sync.Mutex
	lines []string
	pos   int
	full  bool
}

func newLogBuffer() *logBuffer {
	return &logBuffer{lines: make([]string, logBufferSize)}
}

// Write implements io.Writer so logBuffer can be passed to klog.SetOutput.
// klog writes one complete formatted log line per Write call.
func (lb *logBuffer) Write(p []byte) (n int, err error) {
	s := string(p)
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	if s == "" {
		return len(p), nil
	}
	lb.mu.Lock()
	lb.lines[lb.pos] = s
	lb.pos = (lb.pos + 1) % logBufferSize
	if lb.pos == 0 {
		lb.full = true
	}
	lb.mu.Unlock()
	return len(p), nil
}

// snapshot returns buffered lines in chronological order.
func (lb *logBuffer) snapshot() []string {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	if !lb.full {
		out := make([]string, lb.pos)
		copy(out, lb.lines[:lb.pos])
		return out
	}
	out := make([]string, logBufferSize)
	copy(out, lb.lines[lb.pos:])
	copy(out[logBufferSize-lb.pos:], lb.lines[:lb.pos])
	return out
}

func (h *httpServer) handleLogs(w http.ResponseWriter, r *http.Request) {
	lines := h.lb.snapshot()
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(strings.Join(lines, "\n")))
}
