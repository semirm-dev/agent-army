package scaffold

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semir/agent-army/internal/tui"
)

func TestScaffoldSkill(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "spec", "skills"), 0755)

	p := tui.NewFakePrompter(
		"api-designer", // name
		"",             // description (default)
		"",             // scope (default=universal)
		"y",            // confirm
	)

	err := ScaffoldFlow(root, "skill", p, io.Discard)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(root, "spec", "skills", "api-designer.md"))
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if !strings.Contains(s, "name: api-designer") {
		t.Error("missing name")
	}
	if !strings.Contains(s, "## When to Use") {
		t.Error("missing skill template section")
	}
}

func TestScaffoldAgent(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "spec", "agents"), 0755)

	// name -> description -> role(enter=coder) -> scope(enter=universal) -> access(enter=read-write) -> confirm
	p := tui.NewFakePrompter(
		"go-coder",     // name
		"Go coder",     // description
		"",             // role (default=coder)
		"",             // scope (default=universal)
		"",             // access (default=read-write)
		"y",            // confirm
	)

	err := ScaffoldFlow(root, "agent", p, io.Discard)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(root, "spec", "agents", "go-coder.md"))
	if err != nil {
		t.Fatal(err)
	}

	s := string(content)
	if !strings.Contains(s, "name: go-coder") {
		t.Error("missing name")
	}
	if !strings.Contains(s, "role: coder") {
		t.Error("missing role")
	}
	if !strings.Contains(s, "## Capabilities") {
		t.Error("missing agent template section")
	}
}

func TestDefaultDescription(t *testing.T) {
	tests := []struct {
		entityType string
		name       string
		want       string
	}{
		{"skill", "api-designer", "Api Designer workflow and decision tree"},
		{"agent", "go-coder", "Go Coder specialist agent"},
	}

	for _, tt := range tests {
		got := defaultDescription(tt.entityType, tt.name)
		if got != tt.want {
			t.Errorf("defaultDescription(%q, %q) = %q, want %q", tt.entityType, tt.name, got, tt.want)
		}
	}
}

func TestNameToTitle(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"security", "Security"},
		{"go/testing", "Go Testing"},
		{"api-designer", "Api Designer"},
	}

	for _, tt := range tests {
		got := nameToTitle(tt.name)
		if got != tt.want {
			t.Errorf("nameToTitle(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
