package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// RealRunner executes real shell commands, streaming output in real-time
// while capturing stdout for the return value.
type RealRunner struct {
	ctx context.Context
	out io.Writer // live output destination (os.Stdout or io.Discard)
}

// NewReal creates a RealRunner. If ctx is nil, context.Background() is used.
// If out is nil, os.Stdout is used for live output streaming.
func NewReal(ctx context.Context, out ...io.Writer) *RealRunner {
	if ctx == nil {
		ctx = context.Background()
	}
	w := io.Writer(os.Stdout)
	if len(out) > 0 && out[0] != nil {
		w = out[0]
	}
	return &RealRunner{ctx: ctx, out: w}
}

func (r *RealRunner) Run(cmd string, args ...string) (string, error) {
	c := exec.CommandContext(r.ctx, cmd, args...)

	var stdoutBuf bytes.Buffer
	c.Stdout = io.MultiWriter(r.out, &stdoutBuf)
	c.Stderr = os.Stderr
	c.Stdin = nil // prevent interactive prompts from consuming input

	if err := c.Run(); err != nil {
		return stdoutBuf.String(), fmt.Errorf("running %s: %w", cmd, err)
	}
	return stdoutBuf.String(), nil
}
