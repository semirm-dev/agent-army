package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semir/agent-army/internal/model"
)

func TestExtractBody(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "test.md")
	os.WriteFile(fp, []byte("---\nname: foo\n---\n\n# Title\n\nBody here.\n"), 0644)

	got, err := extractBody(fp)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(got, "# Title") {
		t.Errorf("expected body to start with # Title, got: %q", got[:20])
	}
}

func TestExtractBody_NoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "test.md")
	content := "# Just Content\n\nNo frontmatter here.\n"
	os.WriteFile(fp, []byte(content), 0644)

	got, err := extractBody(fp)
	if err != nil {
		t.Fatal(err)
	}
	if got != content {
		t.Errorf("expected full content, got: %q", got)
	}
}

func TestFlattenName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"go/patterns", "go-patterns"},
		{"security", "security"},
		{"go/testing", "go-testing"},
	}
	for _, tt := range tests {
		got := flattenName(tt.input)
		if got != tt.want {
			t.Errorf("flattenName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestAgentToCursor_Substitutions(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "spec", "agents")
	os.MkdirAll(agentsDir, 0755)
	os.WriteFile(filepath.Join(agentsDir, "coder.md"),
		[]byte("---\nname: go-coder\ndescription: Go coder\n---\n\nUse `Edit` and `Bash` tools.\nCheck ~/.claude/config.\n"), 0644)

	got, err := agentToCursor(dir, model.Agent{
		Name:        "go-coder",
		Description: "Go coder",
		Path:        "spec/agents/coder.md",
	}, model.ResolvedDeps{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "`Edit`") {
		t.Error("Edit should be replaced with StrReplace")
	}
	if !strings.Contains(got, "`StrReplace`") {
		t.Error("missing StrReplace")
	}
	if strings.Contains(got, "`Bash`") {
		t.Error("Bash should be replaced with Shell")
	}
	if strings.Contains(got, "~/.claude/") {
		t.Error("~/.claude/ should be replaced with ~/.cursor/")
	}
}

func TestAgentToClaude_Frontmatter(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, "spec", "agents")
	os.MkdirAll(agentsDir, 0755)
	os.WriteFile(filepath.Join(agentsDir, "coder.md"),
		[]byte("---\nname: go/coder\ndescription: Go coder\n---\n\n# Go Coder\n\n## Workflow\n1. Do stuff\n"), 0644)

	got, err := agentToClaude(dir, model.Agent{
		Name:        "go/coder",
		Description: "Go coder",
		Access:      "read-write",
		Path:        "spec/agents/coder.md",
	}, model.ResolvedDeps{})
	if err != nil {
		t.Fatal(err)
	}

	// Check frontmatter has only native fields
	if !strings.Contains(got, "name: go-coder") {
		t.Error("missing name in frontmatter")
	}
	if !strings.Contains(got, "tools: "+claudeToolsRW) {
		t.Error("missing tools in frontmatter")
	}
	if !strings.Contains(got, "model: inherit") {
		t.Error("missing model in frontmatter")
	}
	// Should NOT contain non-native fields
	if strings.Contains(got, "skills:") {
		t.Error("frontmatter should not contain skills field")
	}
	// Should contain enriched body
	if !strings.Contains(got, "## Resources Available") {
		t.Error("missing Resources Available section")
	}
}

func TestSkillToClaude(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, "spec", "skills")
	os.MkdirAll(skillsDir, 0755)
	os.WriteFile(filepath.Join(skillsDir, "error-handling.md"),
		[]byte("---\nname: error-handling\nscope: universal\nuses_skills: [cross-cutting]\n---\n\n# Error Handling\n\nBody content.\n"), 0644)

	got, err := skillToClaude(dir, model.Skill{
		Name: "error-handling",
		Path: "spec/skills/error-handling.md",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Should NOT contain frontmatter
	if strings.Contains(got, "---") {
		t.Error("Claude skill should not contain frontmatter")
	}
	if strings.Contains(got, "scope:") {
		t.Error("should not contain spec-only fields")
	}
	if strings.Contains(got, "uses_skills:") {
		t.Error("should not contain uses_skills")
	}
	if !strings.Contains(got, "# Error Handling") {
		t.Error("missing body content")
	}
}

func TestSkillToCursor(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, "spec", "skills")
	os.MkdirAll(skillsDir, 0755)
	os.WriteFile(filepath.Join(skillsDir, "test.md"),
		[]byte("---\nname: test\n---\n\n# Test\n\nUse `Edit` and `Bash` at ~/.claude/path.\n"), 0644)

	got, err := skillToCursor(dir, model.Skill{
		Name:    "test",
		Summary: "A test skill for testing",
		Path:    "spec/skills/test.md",
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(got, "---") {
		t.Error("Cursor skill should contain frontmatter")
	}
	if !strings.Contains(got, "name: test") {
		t.Error("Cursor skill frontmatter should contain name")
	}
	if !strings.Contains(got, "description: A test skill for testing") {
		t.Error("Cursor skill frontmatter should contain description")
	}
	if strings.Contains(got, "`Edit`") {
		t.Error("Edit should be replaced with StrReplace")
	}
	if !strings.Contains(got, "`StrReplace`") {
		t.Error("missing StrReplace")
	}
	if strings.Contains(got, "~/.claude/") {
		t.Error("~/.claude/ should be replaced with ~/.cursor/")
	}
}
