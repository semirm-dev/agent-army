package tui

import (
	"bufio"
	"fmt"
	"io"
)

// Prompter abstracts user input for testability.
type Prompter interface {
	Prompt(msg string) (string, error)
}

// StdinPrompter reads from a reader and writes prompts to a writer.
type StdinPrompter struct {
	scanner *bufio.Scanner
	w       io.Writer
}

// NewStdinPrompter creates a prompter from reader/writer pair.
func NewStdinPrompter(r io.Reader, w io.Writer) *StdinPrompter {
	return &StdinPrompter{scanner: bufio.NewScanner(r), w: w}
}

// Prompt displays msg and reads a line of input.
func (p *StdinPrompter) Prompt(msg string) (string, error) {
	fmt.Fprint(p.w, msg)
	if p.scanner.Scan() {
		return p.scanner.Text(), nil
	}
	if err := p.scanner.Err(); err != nil {
		return "", err
	}
	return "", io.EOF
}

// FakePrompter returns pre-canned responses for testing.
type FakePrompter struct {
	Responses []string
	idx       int
}

// NewFakePrompter creates a FakePrompter from response strings.
func NewFakePrompter(responses ...string) *FakePrompter {
	return &FakePrompter{Responses: responses}
}

// Prompt returns the next pre-canned response.
func (p *FakePrompter) Prompt(_ string) (string, error) {
	if p.idx >= len(p.Responses) {
		return "", io.EOF
	}
	resp := p.Responses[p.idx]
	p.idx++
	return resp, nil
}
