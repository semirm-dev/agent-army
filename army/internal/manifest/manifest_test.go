package manifest

import (
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
