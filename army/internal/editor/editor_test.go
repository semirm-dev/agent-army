package editor

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semir/agent-army/internal/tui"
)

func TestEditFlow_AddSkillDep(t *testing.T) {
	root := t.TempDir()

	// Create two skills
	skillsDir := filepath.Join(root, "spec", "skills")
	os.MkdirAll(skillsDir, 0755)
	os.WriteFile(filepath.Join(skillsDir, "security.md"),
		[]byte("---\nscope: universal\nuses_skills: []\n---\n\n# Security\n"), 0644)
	os.WriteFile(filepath.Join(skillsDir, "api-design.md"),
		[]byte("---\nscope: universal\nuses_skills: []\n---\n\n# API Design\n"), 0644)

	// Flow: choose skill(1) -> choose security.md(2) -> auto-select uses_skills -> add(1) -> select api-design(1) -> confirm(y)
	p := tui.NewFakePrompter(
		"1",          // entity type: skill
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
	content, _ := os.ReadFile(filepath.Join(skillsDir, "security.md"))
	if !strings.Contains(string(content), "api-design") {
		t.Error("expected security.md to contain api-design in uses_skills")
	}
}

func TestEditFlow_RemoveSkillDep(t *testing.T) {
	root := t.TempDir()

	skillsDir := filepath.Join(root, "spec", "skills")
	os.MkdirAll(skillsDir, 0755)
	os.WriteFile(filepath.Join(skillsDir, "security.md"),
		[]byte("---\nscope: universal\nuses_skills: [api-design, cross-cutting]\n---\n\n# Security\n"), 0644)
	os.WriteFile(filepath.Join(skillsDir, "api-design.md"),
		[]byte("---\nscope: universal\n---\n\n# API Design\n"), 0644)
	os.WriteFile(filepath.Join(skillsDir, "cross-cutting.md"),
		[]byte("---\nscope: universal\n---\n\n# Cross-Cutting\n"), 0644)

	// Flow: skill(1) -> security.md(3) -> remove(2) -> select api-design(1) -> confirm(y)
	p := tui.NewFakePrompter(
		"1",  // entity type: skill
		"3",  // file: security.md (sorted: api-design=1, cross-cutting=2, security=3)
		"2",  // action: remove
		"1",  // select api-design to remove
		"y",  // confirm
	)

	err := EditFlow(root, p, io.Discard)
	if err != nil {
		t.Fatal(err)
	}

	content, _ := os.ReadFile(filepath.Join(skillsDir, "security.md"))
	s := string(content)
	if strings.Contains(s, "api-design") {
		t.Error("api-design should have been removed")
	}
	if !strings.Contains(s, "cross-cutting") {
		t.Error("cross-cutting should still be present")
	}
}
