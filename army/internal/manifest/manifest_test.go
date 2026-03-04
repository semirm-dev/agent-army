package manifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatEntry(t *testing.T) {
	e := Entry{}
	e.Add("name", "test")
	e.Add("scope", "universal")
	e.AddList("uses_rules", []string{"a", "b"})

	got := formatEntry(e)
	if !strings.Contains(got, `"name": "test"`) {
		t.Errorf("missing name field in: %s", got)
	}
	if !strings.Contains(got, `"uses_rules": ["a", "b"]`) {
		t.Errorf("missing uses_rules in: %s", got)
	}
	if !strings.HasPrefix(got, "{ ") || !strings.HasSuffix(got, " }") {
		t.Errorf("wrong format: %s", got)
	}
}

func TestFormatEntry_EmptyList(t *testing.T) {
	e := Entry{}
	e.Add("name", "test")
	e.AddList("uses_rules", nil)

	got := formatEntry(e)
	if !strings.Contains(got, `"uses_rules": []`) {
		t.Errorf("expected empty array in: %s", got)
	}
}

func TestFormatManifestJSON(t *testing.T) {
	e := Entry{}
	e.Add("name", "test-rule")
	e.Add("scope", "universal")
	e.Add("path", "spec/rules/test.md")

	m := OrderedMap{
		Keys: []string{"rules"},
		Sections: map[string][]Entry{
			"rules": {e},
		},
	}

	got := formatManifestJSON(m)

	if !strings.HasPrefix(got, "{\n") {
		t.Error("should start with {")
	}
	if !strings.HasSuffix(got, "}\n") {
		t.Error("should end with }\\n")
	}
	if !strings.Contains(got, `  "rules": [`) {
		t.Error("missing section header")
	}
	if !strings.Contains(got, `    { "name": "test-rule"`) {
		t.Error("missing entry")
	}
}

func TestGenerateManifest(t *testing.T) {
	root := t.TempDir()

	// Create rules
	rulesDir := filepath.Join(root, "spec", "rules")
	os.MkdirAll(rulesDir, 0755)
	os.WriteFile(filepath.Join(rulesDir, "security.md"),
		[]byte("---\nscope: universal\n---\n\n# Security\n"), 0644)

	// Create skills
	skillsDir := filepath.Join(root, "spec", "skills")
	os.MkdirAll(skillsDir, 0755)
	os.WriteFile(filepath.Join(skillsDir, "api-designer.md"),
		[]byte("---\nname: api-designer\nscope: universal\nuses_rules: [security]\n---\n\n# API Designer\n"), 0644)

	// Create agents
	agentsDir := filepath.Join(root, "spec", "agents")
	os.MkdirAll(agentsDir, 0755)
	os.WriteFile(filepath.Join(agentsDir, "coder.md"),
		[]byte("---\nname: go-coder\ndescription: Go coder\nrole: coder\nscope: universal\naccess: read-write\nuses_skills: [api-designer]\nuses_rules: []\nuses_plugins: []\ndelegates_to: []\n---\n\n# Coder\n"), 0644)

	m, err := GenerateManifest(root)
	if err != nil {
		t.Fatal(err)
	}

	if len(m.Sections["rules"]) != 1 {
		t.Errorf("rules count = %d, want 1", len(m.Sections["rules"]))
	}
	if len(m.Sections["skills"]) != 1 {
		t.Errorf("skills count = %d, want 1", len(m.Sections["skills"]))
	}
	if len(m.Sections["agents"]) != 1 {
		t.Errorf("agents count = %d, want 1", len(m.Sections["agents"]))
	}

	// Check agent has transitive rules from skill
	agentE := m.Sections["agents"][0]
	usesRules, ok := agentE.Values["uses_rules"].([]string)
	if !ok {
		t.Fatal("uses_rules not a string slice")
	}
	if len(usesRules) != 1 || usesRules[0] != "security" {
		t.Errorf("agent uses_rules = %v, want [security]", usesRules)
	}
}

func TestGenerateManifest_WithPlugins(t *testing.T) {
	root := t.TempDir()

	// Create minimal spec dirs so loader doesn't error
	for _, dir := range []string{"spec/rules", "spec/skills", "spec/agents", "spec/claude"} {
		os.MkdirAll(filepath.Join(root, dir), 0755)
	}

	// Create spec/claude/settings.json with plugin data
	settingsJSON := `{
  "permissions": {"defaultMode": "plan"},
  "external_plugins": [
    { "name": "superpowers", "marketplace": "claude-plugins-official" },
    { "name": "context7", "marketplace": "claude-plugins-official" }
  ],
  "external_skills": [
    { "name": "skill-creator", "repo": "https://github.com/anthropics/skills", "tool": "npm" }
  ]
}`
	os.WriteFile(filepath.Join(root, "spec", "claude", "settings.json"), []byte(settingsJSON), 0644)

	m, err := GenerateManifest(root)
	if err != nil {
		t.Fatal(err)
	}

	// Verify external_plugins section exists
	if _, ok := m.RawSections["external_plugins"]; !ok {
		t.Error("missing external_plugins in RawSections")
	}

	// Verify external_skills section exists
	if _, ok := m.RawSections["external_skills"]; !ok {
		t.Error("missing external_skills in RawSections")
	}

	// Verify key ordering includes the new sections
	found := map[string]bool{}
	for _, k := range m.Keys {
		found[k] = true
	}
	if !found["external_plugins"] {
		t.Error("external_plugins not in Keys")
	}
	if !found["external_skills"] {
		t.Error("external_skills not in Keys")
	}

	// Verify raw content parses correctly
	var plugins []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(m.RawSections["external_plugins"], &plugins); err != nil {
		t.Fatalf("failed to parse external_plugins: %v", err)
	}
	if len(plugins) != 2 {
		t.Errorf("external_plugins count = %d, want 2", len(plugins))
	}
	if plugins[0].Name != "superpowers" {
		t.Errorf("first plugin = %q, want superpowers", plugins[0].Name)
	}
}

func TestGenerateManifest_NoPluginsFile(t *testing.T) {
	root := t.TempDir()

	// Create minimal spec dirs, no settings.json with plugin data
	for _, dir := range []string{"spec/rules", "spec/skills", "spec/agents"} {
		os.MkdirAll(filepath.Join(root, dir), 0755)
	}

	m, err := GenerateManifest(root)
	if err != nil {
		t.Fatal(err)
	}

	// Should still have the base 3 keys
	if len(m.Keys) != 3 {
		t.Errorf("keys count = %d, want 3 (no plugin sections)", len(m.Keys))
	}
	if len(m.RawSections) != 0 {
		t.Errorf("RawSections should be empty, got %d entries", len(m.RawSections))
	}
}

func TestFormatManifestJSON_WithRawSections(t *testing.T) {
	e := Entry{}
	e.Add("name", "test-rule")

	raw := json.RawMessage(`[{"name":"superpowers","marketplace":"official"}]`)

	m := OrderedMap{
		Keys: []string{"rules", "external_plugins"},
		Sections: map[string][]Entry{
			"rules": {e},
		},
		RawSections: map[string]json.RawMessage{
			"external_plugins": raw,
		},
	}

	got := formatManifestJSON(m)

	if !strings.Contains(got, `"rules": [`) {
		t.Error("missing rules section")
	}
	if !strings.Contains(got, `"external_plugins":`) {
		t.Error("missing external_plugins section")
	}
	if !strings.Contains(got, `"superpowers"`) {
		t.Error("missing plugin name in raw section")
	}
	// Verify it ends properly
	if !strings.HasSuffix(got, "}\n") {
		t.Error("should end with }\\n")
	}
}

func TestWriteManifest(t *testing.T) {
	root := t.TempDir()

	rulesDir := filepath.Join(root, "spec", "rules")
	os.MkdirAll(rulesDir, 0755)
	os.WriteFile(filepath.Join(rulesDir, "test.md"),
		[]byte("---\nscope: universal\n---\n\n# Test\n"), 0644)

	if err := WriteManifest(root); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), `"rules"`) {
		t.Error("manifest should contain rules section")
	}
}

func TestWriteManifest_WithPlugins(t *testing.T) {
	root := t.TempDir()

	for _, dir := range []string{"spec/rules", "spec/skills", "spec/agents", "spec/claude"} {
		os.MkdirAll(filepath.Join(root, dir), 0755)
	}

	os.WriteFile(filepath.Join(root, "spec", "claude", "settings.json"),
		[]byte(`{"permissions":{"defaultMode":"plan"},"external_plugins":[{"name":"context7"}],"external_skills":[{"name":"skill-creator"}]}`), 0644)

	if err := WriteManifest(root); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(content)

	if !strings.Contains(s, `"external_plugins"`) {
		t.Error("manifest should contain external_plugins section")
	}
	if !strings.Contains(s, `"external_skills"`) {
		t.Error("manifest should contain external_skills section")
	}
	if !strings.Contains(s, `"context7"`) {
		t.Error("manifest should contain context7 plugin")
	}
}
