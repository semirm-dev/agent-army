package runner

import (
	"fmt"
	"strings"
)

// DryRunner prints commands that would be executed without running them.
type DryRunner struct{}

// NewDry creates a DryRunner.
func NewDry() *DryRunner {
	return &DryRunner{}
}

func (d *DryRunner) Run(cmd string, args ...string) (string, error) {
	parts := append([]string{cmd}, args...)
	fmt.Printf("[dry-run] %s\n", strings.Join(parts, " "))
	return "", nil
}
