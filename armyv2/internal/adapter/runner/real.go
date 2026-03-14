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
}

// NewReal creates a RealRunner. If ctx is nil, context.Background() is used.
func NewReal(ctx context.Context) *RealRunner {
	if ctx == nil {
		ctx = context.Background()
	}
	return &RealRunner{ctx: ctx}
}

func (r *RealRunner) Run(cmd string, args ...string) (string, error) {
	c := exec.CommandContext(r.ctx, cmd, args...)

	var stdoutBuf bytes.Buffer
	c.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	c.Stderr = os.Stderr
	c.Stdin = nil // prevent interactive prompts from consuming input

	if err := c.Run(); err != nil {
		return stdoutBuf.String(), fmt.Errorf("running %s: %w", cmd, err)
	}
	return stdoutBuf.String(), nil
}
