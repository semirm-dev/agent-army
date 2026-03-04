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

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		rule model.Rule
		want string
	}{
		{model.Rule{Name: "go/patterns"}, "go"},
		{model.Rule{Name: "typescript/testing"}, "typescript"},
		{model.Rule{Name: "security", Languages: []string{"python"}}, "python"},
		{model.Rule{Name: "security"}, ""},
	}
	for _, tt := range tests {
		got := detectLanguage(tt.rule)
		if got != tt.want {
			t.Errorf("detectLanguage(%q) = %q, want %q", tt.rule.Name, got, tt.want)
		}
	}
}

func TestRuleToClaude(t *testing.T) {
	dir := t.TempDir()
	rulesDir := filepath.Join(dir, "spec", "rules")
	os.MkdirAll(rulesDir, 0755)
	os.WriteFile(filepath.Join(rulesDir, "security.md"),
		[]byte("---\nscope: universal\n---\n\n# Security\n\nBody.\n"), 0644)

	got, err := ruleToClaude(dir, model.Rule{Name: "security", Path: "spec/rules/security.md"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "---") {
		t.Error("Claude rule should not contain frontmatter")
	}
	if !strings.Contains(got, "# Security") {
		t.Error("missing body content")
	}
}

func TestRuleToCursor(t *testing.T) {
	dir := t.TempDir()
	rulesDir := filepath.Join(dir, "spec", "rules", "go")
	os.MkdirAll(rulesDir, 0755)
	os.WriteFile(filepath.Join(rulesDir, "patterns.md"),
		[]byte("---\nscope: language-specific\nlanguages: [go]\n---\n\n# Go Patterns\n"), 0644)

	got, err := ruleToCursor(dir, model.Rule{
		Name:        "go/patterns",
		Description: "Go Patterns",
		Languages:   []string{"go"},
		Path:        "spec/rules/go/patterns.md",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "globs:") {
		t.Error("expected globs for language-specific rule")
	}
	if !strings.Contains(got, "# Go Patterns") {
		t.Error("missing body")
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
	})
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

func TestAssignCursorNumbers(t *testing.T) {
	rules := []model.Rule{
		{Name: "go/patterns"},
		{Name: "security"},
		{Name: "git-workflow"},
		{Name: "api-design"},
	}
	assignments := assignCursorNumbers(rules)
	if len(assignments) != 4 {
		t.Fatalf("got %d assignments, want 4", len(assignments))
	}
	// go/patterns → language (100)
	if assignments[0].Number != 100 {
		t.Errorf("go/patterns number = %d, want 100", assignments[0].Number)
	}
	// security → infrastructure (500)
	if assignments[1].Number != 500 {
		t.Errorf("security number = %d, want 500", assignments[1].Number)
	}
	// git-workflow → git (300)
	if assignments[2].Number != 300 {
		t.Errorf("git-workflow number = %d, want 300", assignments[2].Number)
	}
	// api-design → api-db (400)
	if assignments[3].Number != 400 {
		t.Errorf("api-design number = %d, want 400", assignments[3].Number)
	}
}

func TestCategorizeRule(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"go/patterns", "language"},
		{"typescript/testing", "language"},
		{"git-workflow", "git"},
		{"api-design", "api-db"},
		{"database", "api-db"},
		{"security", "infrastructure"},
		{"observability", "infrastructure"},
	}
	for _, tt := range tests {
		got := categorizeRule(model.Rule{Name: tt.name})
		if got != tt.want {
			t.Errorf("categorizeRule(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
