package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateAntigravityMD(t *testing.T) {
	dir := t.TempDir()
	tmplDir := filepath.Join(dir, "spec", "antigravity")
	os.MkdirAll(tmplDir, 0755)

	template := `# Orchestrator

Agent references in ` + "`{{BASE}}/workflows/`" + `.

End of doc.
`
	tmplPath := filepath.Join(tmplDir, "GEMINI.md")
	os.WriteFile(tmplPath, []byte(template), 0644)

	destDir := filepath.Join(dir, "project", ".agents")

	err := generateAntigravityMD(dir, destDir, tmplPath)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "GEMINI.md"))
	if err != nil {
		t.Fatal(err)
	}

	result := string(content)

	// Verify {{BASE}} was replaced with .agents (project-local, no artifact prefix)
	if strings.Contains(result, "{{BASE}}") {
		t.Error("{{BASE}} placeholder should be replaced")
	}
	if !strings.Contains(result, ".agents/workflows/") {
		t.Errorf("expected .agents/workflows/ path for project-local dest, got:\n%s", result)
	}

	// Verify no excessive blank lines
	if strings.Contains(result, "\n\n\n") {
		t.Error("should not have 3+ consecutive newlines after blank line collapse")
	}

	// Static content preserved
	if !strings.Contains(result, "# Orchestrator") {
		t.Error("missing static content")
	}
}

func TestGenerateAntigravityMD_GlobalWritesToGeminiLevel(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	dir := t.TempDir()
	tmplDir := filepath.Join(dir, "spec", "antigravity")
	os.MkdirAll(tmplDir, 0755)

	template := `# Orchestrator
`
	tmplPath := filepath.Join(tmplDir, "GEMINI.md")
	os.WriteFile(tmplPath, []byte(template), 0644)

	// Simulate global antigravity dest
	fakeHome := filepath.Join(dir, "fakehome")
	globalAntigravity := filepath.Join(fakeHome, ".gemini", "antigravity")
	globalGemini := filepath.Join(fakeHome, ".gemini")
	os.MkdirAll(globalAntigravity, 0755)

	// We can't easily override os.UserHomeDir in the test, so test the helper directly
	geminiDir, artifactPrefix := antigravityOutputPaths(globalAntigravity)

	// For non-real-home paths, antigravityOutputPaths won't match.
	// Test the function contract: real home path would produce parent + "antigravity/" prefix
	_ = geminiDir
	_ = artifactPrefix

	// Instead, verify the functions with real home path
	realGlobalAntigravity := filepath.Join(home, ".gemini", "antigravity")
	geminiDir, artifactPrefix = antigravityOutputPaths(realGlobalAntigravity)

	expectedGeminiDir := filepath.Join(home, ".gemini")
	if filepath.Clean(geminiDir) != filepath.Clean(expectedGeminiDir) {
		t.Errorf("global antigravity geminiDir = %q, want %q", geminiDir, expectedGeminiDir)
	}
	if artifactPrefix != "antigravity/" {
		t.Errorf("global antigravity artifactPrefix = %q, want %q", artifactPrefix, "antigravity/")
	}

	// Verify AntigravityGeminiMDPath for global
	mdPath := AntigravityGeminiMDPath(realGlobalAntigravity)
	expectedMDPath := filepath.Join(home, ".gemini", "GEMINI.md")
	if filepath.Clean(mdPath) != filepath.Clean(expectedMDPath) {
		t.Errorf("AntigravityGeminiMDPath(%q) = %q, want %q", realGlobalAntigravity, mdPath, expectedMDPath)
	}

	// Verify prefix for gemini-level dir
	prefix := antigravityDestPrefix(expectedGeminiDir)
	if prefix != "~/.gemini" {
		t.Errorf("antigravityDestPrefix(%q) = %q, want %q", expectedGeminiDir, prefix, "~/.gemini")
	}

	_ = globalGemini
}

func TestGenerateAntigravityMD_EmptyBuilders(t *testing.T) {
	dir := t.TempDir()
	tmplDir := filepath.Join(dir, "spec", "antigravity")
	os.MkdirAll(tmplDir, 0755)

	template := `# Header
- **Role:** Delegate tasks.
- **Verification:** Run build.
`
	tmplPath := filepath.Join(tmplDir, "GEMINI.md")
	os.WriteFile(tmplPath, []byte(template), 0644)

	destDir := filepath.Join(dir, "project", ".agents")

	err := generateAntigravityMD(dir, destDir, tmplPath)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "GEMINI.md"))
	if err != nil {
		t.Fatal(err)
	}

	result := string(content)

	// Static content preserved
	if !strings.Contains(result, "Role:") {
		t.Error("missing static content")
	}
	if !strings.Contains(result, "Verification:") {
		t.Error("missing static content")
	}
}

func TestAntigravityOutputPaths(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name               string
		dest               string
		wantGeminiDir      string
		wantArtifactPrefix string
	}{
		{
			name:               "global antigravity",
			dest:               filepath.Join(home, ".gemini", "antigravity"),
			wantGeminiDir:      filepath.Join(home, ".gemini"),
			wantArtifactPrefix: "antigravity/",
		},
		{
			name:               "project local .agents",
			dest:               "/some/project/.agents",
			wantGeminiDir:      "/some/project/.agents",
			wantArtifactPrefix: "",
		},
		{
			name:               "custom path",
			dest:               "/opt/custom/output",
			wantGeminiDir:      "/opt/custom/output",
			wantArtifactPrefix: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			geminiDir, artifactPrefix := antigravityOutputPaths(tt.dest)
			if filepath.Clean(geminiDir) != filepath.Clean(tt.wantGeminiDir) {
				t.Errorf("geminiDir = %q, want %q", geminiDir, tt.wantGeminiDir)
			}
			if artifactPrefix != tt.wantArtifactPrefix {
				t.Errorf("artifactPrefix = %q, want %q", artifactPrefix, tt.wantArtifactPrefix)
			}
		})
	}
}

func TestAntigravityDestPrefix(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		geminiDir string
		want     string
	}{
		{
			name:     "global gemini home",
			geminiDir: filepath.Join(home, ".gemini"),
			want:     "~/.gemini",
		},
		{
			name:     "project local .agents",
			geminiDir: "/some/project/.agents",
			want:     ".agents",
		},
		{
			name:     "custom path",
			geminiDir: "/opt/custom/output",
			want:     "/opt/custom/output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := antigravityDestPrefix(tt.geminiDir)
			if got != tt.want {
				t.Errorf("antigravityDestPrefix(%q) = %q, want %q", tt.geminiDir, got, tt.want)
			}
		})
	}
}

func TestAntigravityGeminiMDPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name string
		dest string
		want string
	}{
		{
			name: "global writes to gemini level",
			dest: filepath.Join(home, ".gemini", "antigravity"),
			want: filepath.Join(home, ".gemini", "GEMINI.md"),
		},
		{
			name: "local stays in dest",
			dest: "/some/project/.agents",
			want: "/some/project/.agents/GEMINI.md",
		},
		{
			name: "custom stays in dest",
			dest: "/opt/custom",
			want: "/opt/custom/GEMINI.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AntigravityGeminiMDPath(tt.dest)
			if filepath.Clean(got) != filepath.Clean(tt.want) {
				t.Errorf("AntigravityGeminiMDPath(%q) = %q, want %q", tt.dest, got, tt.want)
			}
		})
	}
}

