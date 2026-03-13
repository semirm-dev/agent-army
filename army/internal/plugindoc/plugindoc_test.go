package plugindoc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGithubSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://github.com/obra/superpowers", "obra/superpowers"},
		{"https://github.com/obra/superpowers.git", "obra/superpowers"},
		{"https://github.com/coderabbitai/claude-plugin", "coderabbitai/claude-plugin"},
		{"https://example.com/not-github", ""},
		{"", ""},
		{"github.com/foo/bar/baz", "foo/bar"},
	}
	for _, tt := range tests {
		got := githubSlug(tt.input)
		if got != tt.want {
			t.Errorf("githubSlug(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExtractDescription(t *testing.T) {
	dir := t.TempDir()

	// Single-line description
	singleLine := filepath.Join(dir, "single.md")
	os.WriteFile(singleLine, []byte("---\nname: test\ndescription: A simple skill\n---\n\nBody here.\n"), 0644)
	if got := extractDescription(singleLine); got != "A simple skill" {
		t.Errorf("single-line: got %q, want %q", got, "A simple skill")
	}

	// Multiline description with |
	multiLine := filepath.Join(dir, "multi.md")
	os.WriteFile(multiLine, []byte("---\nname: test\ndescription: |\n  This is the real description\n---\n\nBody.\n"), 0644)
	if got := extractDescription(multiLine); got != "This is the real description" {
		t.Errorf("multiline: got %q, want %q", got, "This is the real description")
	}

	// Pipe character in description gets escaped
	pipeDesc := filepath.Join(dir, "pipe.md")
	os.WriteFile(pipeDesc, []byte("---\ndescription: Has | pipe char\n---\n"), 0644)
	got := extractDescription(pipeDesc)
	if got != "Has \u2014 pipe char" {
		t.Errorf("pipe escaping: got %q", got)
	}

	// Missing file
	if got := extractDescription(filepath.Join(dir, "nope.md")); got != "" {
		t.Errorf("missing file: got %q, want empty", got)
	}
}

func TestShortDescription(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"First sentence. Second sentence.", "First sentence."},
		{"No period here", "No period here"},
		{"<tag>Wrapped</tag> text. More.", "Wrapped text."},
	}
	for _, tt := range tests {
		got := shortDescription(tt.input)
		if got != tt.want {
			t.Errorf("shortDescription(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}

	// XML tag stripping
	got := shortDescription("<b>Bold</b> and <i>italic</i> text.")
	if got != "Bold and italic text." {
		t.Errorf("XML stripping: got %q", got)
	}

	// Truncation at 200 chars
	long := make([]byte, 250)
	for i := range long {
		long[i] = 'a'
	}
	got = shortDescription(string(long))
	if len(got) != 200 {
		t.Errorf("truncation: got len %d, want 200", len(got))
	}
}

func TestGenerate_empty(t *testing.T) {
	// With a fake home that has no files, Generate should still produce valid markdown
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	output, err := Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if !contains(output, "## Plugins (0)") {
		t.Error("expected Plugins (0) in output")
	}
	if !contains(output, "## Skills (0)") {
		t.Error("expected Skills (0) in output")
	}
	if !contains(output, "## Custom Agents (0)") {
		t.Error("expected Custom Agents (0) in output")
	}
	if !contains(output, "## Plugin Marketplaces (0)") {
		t.Error("expected Plugin Marketplaces (0) in output")
	}
	if !contains(output, "## MCP Servers (0)") {
		t.Error("expected MCP Servers (0) in output")
	}
}

func TestBuildPluginSkillNames(t *testing.T) {
	dir := t.TempDir()

	// Create a fake plugin with skills/ and commands/
	pluginDir := filepath.Join(dir, "my-plugin")
	skillDir := filepath.Join(pluginDir, "skills", "foo-skill")
	cmdsDir := filepath.Join(pluginDir, "commands")
	os.MkdirAll(skillDir, 0755)
	os.MkdirAll(cmdsDir, 0755)

	// Plugin metadata
	metaDir := filepath.Join(pluginDir, ".claude-plugin")
	os.MkdirAll(metaDir, 0755)
	os.WriteFile(filepath.Join(metaDir, "plugin.json"), []byte(`{"name":"my-plugin","description":"test"}`), 0644)

	// A non-deprecated command
	os.WriteFile(filepath.Join(cmdsDir, "bar-cmd.md"), []byte("---\ndescription: A useful command\n---\n"), 0644)

	// A deprecated command (should be excluded)
	os.WriteFile(filepath.Join(cmdsDir, "old-cmd.md"), []byte("---\ndescription: Deprecated - use bar-cmd instead\n---\n"), 0644)

	plugins := installedPluginsFile{
		Plugins: map[string][]pluginInstance{
			"my-plugin@mkt": {{InstallPath: pluginDir, Version: "1.0.0"}},
		},
	}

	result := buildPluginSkillNames(plugins)

	if result["foo-skill"] != "my-plugin" {
		t.Errorf("expected foo-skill → my-plugin, got %q", result["foo-skill"])
	}
	if result["bar-cmd"] != "my-plugin" {
		t.Errorf("expected bar-cmd → my-plugin, got %q", result["bar-cmd"])
	}
	if _, ok := result["old-cmd"]; ok {
		t.Error("deprecated command old-cmd should not be in plugin skill names")
	}
}

func TestDuplicateSkillExclusion(t *testing.T) {
	plugins := installedPluginsFile{
		Plugins: map[string][]pluginInstance{},
	}
	skillLock := skillLockFile{
		Skills: map[string]skillEntry{
			"my-skill": {Source: "owner/repo", SourceURL: "https://github.com/owner/repo"},
		},
	}
	pluginRepoMap := map[string]string{}
	pluginSkillNames := map[string]string{
		"my-skill": "cool-plugin",
	}

	var b strings.Builder
	generateSkillsSection(&b, plugins, skillLock, pluginRepoMap, pluginSkillNames)
	output := b.String()

	// Duplicate standalone skills should be excluded from skill tables
	// but listed in the redundant blockquote for sync to remove
	if contains(output, "| `my-skill`") {
		t.Error("expected duplicate skill to be excluded from skill tables")
	}
	if !contains(output, "Redundant standalone skills") {
		t.Error("expected redundant standalone skills blockquote")
	}
	if !contains(output, "npx skills remove my-skill") {
		t.Error("expected removal command for duplicate skill")
	}
	// Total should be 0: 0 plugin-provided + 0 standalone (1 excluded as duplicate)
	if !contains(output, "## Skills (0)") {
		t.Errorf("expected Skills (0) in output, got: %s", output[:200])
	}
}

func TestDetectDrift(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create .agents directory with lock file
	agentsDir := filepath.Join(tmpHome, ".agents")
	os.MkdirAll(agentsDir, 0755)

	// Create lock file with two skills
	lockData := `{"skills":{"existing-skill":{"source":"owner/repo","sourceUrl":"https://github.com/owner/repo"},"missing-skill":{"source":"other/repo","sourceUrl":"https://github.com/other/repo"}}}`
	os.WriteFile(filepath.Join(agentsDir, ".skill-lock.json"), []byte(lockData), 0644)

	// Create directory for existing-skill only
	skillDir := filepath.Join(agentsDir, "skills", "existing-skill")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: existing-skill\n---\n"), 0644)

	entries, err := DetectDrift()
	if err != nil {
		t.Fatalf("DetectDrift() error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 drift entry, got %d", len(entries))
	}
	if entries[0].Name != "missing-skill" {
		t.Errorf("expected drift entry name 'missing-skill', got %q", entries[0].Name)
	}
	if entries[0].Source != "other/repo" {
		t.Errorf("expected drift entry source 'other/repo', got %q", entries[0].Source)
	}
}

func TestDetectDrift_noDrift(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// No lock file at all
	entries, err := DetectDrift()
	if err != nil {
		t.Fatalf("DetectDrift() error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 drift entries, got %d", len(entries))
	}
}

func TestRemoveDriftEntries(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	agentsDir := filepath.Join(tmpHome, ".agents")
	os.MkdirAll(agentsDir, 0755)

	lockData := `{
  "skills": {
    "keep-me": {"source": "a/b", "sourceUrl": "https://github.com/a/b"},
    "remove-me": {"source": "c/d", "sourceUrl": "https://github.com/c/d"},
    "also-remove": {"source": "e/f", "sourceUrl": "https://github.com/e/f"}
  }
}`
	lockPath := filepath.Join(agentsDir, ".skill-lock.json")
	os.WriteFile(lockPath, []byte(lockData), 0644)

	err := RemoveDriftEntries([]DriftEntry{
		{Name: "remove-me", Source: "c/d"},
		{Name: "also-remove", Source: "e/f"},
	})
	if err != nil {
		t.Fatalf("RemoveDriftEntries() error: %v", err)
	}

	// Read back and verify
	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatalf("reading lock file: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "remove-me") {
		t.Error("expected remove-me to be removed from lock file")
	}
	if strings.Contains(content, "also-remove") {
		t.Error("expected also-remove to be removed from lock file")
	}
	if !strings.Contains(content, "keep-me") {
		t.Error("expected keep-me to remain in lock file")
	}
}

func TestDetectOrphans(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create .agents directory with lock file and skill directories
	agentsDir := filepath.Join(tmpHome, ".agents")
	os.MkdirAll(agentsDir, 0755)

	// Lock file has only "locked-skill"
	lockData := `{"skills":{"locked-skill":{"source":"owner/repo","sourceUrl":"https://github.com/owner/repo"}}}`
	os.WriteFile(filepath.Join(agentsDir, ".skill-lock.json"), []byte(lockData), 0644)

	// Create directories for both locked-skill and orphan-skill
	skillsDir := filepath.Join(agentsDir, "skills")
	os.MkdirAll(filepath.Join(skillsDir, "locked-skill"), 0755)
	os.MkdirAll(filepath.Join(skillsDir, "orphan-skill"), 0755)

	entries, err := DetectOrphans()
	if err != nil {
		t.Fatalf("DetectOrphans() error: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 orphan entry, got %d", len(entries))
	}
	if entries[0].Name != "orphan-skill" {
		t.Errorf("expected orphan entry name 'orphan-skill', got %q", entries[0].Name)
	}
}

func TestDetectOrphans_noOrphans(t *testing.T) {
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// No skills directory at all
	entries, err := DetectOrphans()
	if err != nil {
		t.Fatalf("DetectOrphans() error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 orphan entries, got %d", len(entries))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
