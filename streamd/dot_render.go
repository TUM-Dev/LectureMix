package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// RenderDOTToSVG takes a DOT graph definition and returns the generated SVG bytes.
func RenderDOTToSVG(dotGraph string) ([]byte, error) {
	// Use Context with a timeout to prevent hanging processes
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "dot", "-Tsvg")
	cmd.Stdin = strings.NewReader(dotGraph)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("dot process timed out")
		}
		return nil, fmt.Errorf("dot command failed: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.Bytes(), nil
}
