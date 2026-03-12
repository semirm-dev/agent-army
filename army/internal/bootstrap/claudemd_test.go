package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDestPrefix(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name string
		dest string
		want string
	}{
		{
			name: "global home",
			dest: filepath.Join(home, ".claude"),
			want: "~/.claude",
		},
		{
			name: "project local",
			dest: "/some/project/.claude",
			want: ".claude",
		},
		{
			name: "custom path",
			dest: "/opt/custom/output",
			want: "/opt/custom/output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := destPrefix(tt.dest)
			if got != tt.want {
				t.Errorf("destPrefix(%q) = %q, want %q", tt.dest, got, tt.want)
			}
		})
	}
}

func TestGenerateClaudeMD(t *testing.T) {
	dir := t.TempDir()
	tmplDir := filepath.Join(dir, "spec", "claude")
	os.MkdirAll(tmplDir, 0755)

	template := "# Orchestrator\n\nAgent prompts live in `{{BASE}}/agents/`.\n\n- **Verification:** Run build.\n"
	tmplPath := filepath.Join(tmplDir, "CLAUDE.md")
	os.WriteFile(tmplPath, []byte(template), 0644)

	destDir := filepath.Join(dir, "project", ".claude")

	err := generateClaudeMD(destDir, tmplPath)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}

	result := string(content)

	// Verify {{BASE}} was replaced
	if strings.Contains(result, "{{BASE}}") {
		t.Error("{{BASE}} placeholder should be replaced")
	}
	if !strings.Contains(result, ".claude/agents/") {
		t.Errorf("expected .claude/agents/ path, got:\n%s", result)
	}

	// Verify static content preserved
	if !strings.Contains(result, "# Orchestrator") {
		t.Error("missing static header")
	}
	if !strings.Contains(result, "Verification:") {
		t.Error("missing static content")
	}
}

func TestGenerateClaudeMD_StaticPassthrough(t *testing.T) {
	dir := t.TempDir()
	tmplDir := filepath.Join(dir, "spec", "claude")
	os.MkdirAll(tmplDir, 0755)

	template := "# Header\n- **Parallelism:** Split work.\n- **Verification:** Run build.\n"
	tmplPath := filepath.Join(tmplDir, "CLAUDE.md")
	os.WriteFile(tmplPath, []byte(template), 0644)

	destDir := filepath.Join(dir, "project", ".claude")

	err := generateClaudeMD(destDir, tmplPath)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}

	result := string(content)

	if !strings.Contains(result, "Parallelism:") {
		t.Error("missing static content")
	}
	if !strings.Contains(result, "Verification:") {
		t.Error("missing static content")
	}
}

func TestGenerateClaudeMD_GlobalDest(t *testing.T) {
	dir := t.TempDir()

	// Override HOME so os.UserHomeDir() returns our temp dir
	t.Setenv("HOME", dir)

	tmplDir := filepath.Join(dir, "spec", "claude")
	os.MkdirAll(tmplDir, 0755)

	template := "Paths: `{{BASE}}/agents/`, `{{BASE}}/skills/`\n"
	tmplPath := filepath.Join(tmplDir, "CLAUDE.md")
	os.WriteFile(tmplPath, []byte(template), 0644)

	destDir := filepath.Join(dir, ".claude")

	err := generateClaudeMD(destDir, tmplPath)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}

	result := string(content)
	if !strings.Contains(result, "`~/.claude/agents/`") {
		t.Errorf("expected ~/.claude prefix for global dest, got:\n%s", result)
	}
}
