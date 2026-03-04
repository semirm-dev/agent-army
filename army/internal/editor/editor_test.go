package editor

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semir/agent-army/internal/tui"
)

func TestEditFlow_AddRule(t *testing.T) {
	root := t.TempDir()

	// Create two rules
	rulesDir := filepath.Join(root, "rules")
	os.MkdirAll(rulesDir, 0755)
	os.WriteFile(filepath.Join(rulesDir, "security.md"),
		[]byte("---\nscope: universal\nuses_rules: []\n---\n\n# Security\n"), 0644)
	os.WriteFile(filepath.Join(rulesDir, "api-design.md"),
		[]byte("---\nscope: universal\nuses_rules: []\n---\n\n# API Design\n"), 0644)

	// Flow: choose rule(1) → choose security.md(2) → auto-select uses_rules → add(1) → select api-design(1) → confirm(y)
	p := tui.NewFakePrompter(
		"1",          // entity type: rule
		"2",          // file: security.md (sorted: api-design=1, security=2)
		"1",          // action: add
		"1",          // select api-design to add
		"y",          // confirm
	)

	err := EditFlow(root, p, io.Discard)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the file was updated
	content, _ := os.ReadFile(filepath.Join(rulesDir, "security.md"))
	if !strings.Contains(string(content), "api-design") {
		t.Error("expected security.md to contain api-design in uses_rules")
	}
}

func TestEditFlow_RemoveRule(t *testing.T) {
	root := t.TempDir()

	rulesDir := filepath.Join(root, "rules")
	os.MkdirAll(rulesDir, 0755)
	os.WriteFile(filepath.Join(rulesDir, "security.md"),
		[]byte("---\nscope: universal\nuses_rules: [api-design, cross-cutting]\n---\n\n# Security\n"), 0644)
	os.WriteFile(filepath.Join(rulesDir, "api-design.md"),
		[]byte("---\nscope: universal\n---\n\n# API Design\n"), 0644)
	os.WriteFile(filepath.Join(rulesDir, "cross-cutting.md"),
		[]byte("---\nscope: universal\n---\n\n# Cross-Cutting\n"), 0644)

	// Flow: rule(1) → security.md(3) → remove(2) → select api-design(1) → confirm(y)
	p := tui.NewFakePrompter(
		"1",  // entity type: rule
		"3",  // file: security.md (sorted: api-design=1, cross-cutting=2, security=3)
		"2",  // action: remove
		"1",  // select api-design to remove
		"y",  // confirm
	)

	err := EditFlow(root, p, io.Discard)
	if err != nil {
		t.Fatal(err)
	}

	content, _ := os.ReadFile(filepath.Join(rulesDir, "security.md"))
	s := string(content)
	if strings.Contains(s, "api-design") {
		t.Error("api-design should have been removed")
	}
	if !strings.Contains(s, "cross-cutting") {
		t.Error("cross-cutting should still be present")
	}
}
