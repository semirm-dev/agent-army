package loader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindMDFiles(t *testing.T) {
	dir := t.TempDir()
	// Create nested structure
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "a.md"), []byte("# A"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("not md"), 0644)
	os.WriteFile(filepath.Join(dir, "sub", "c.md"), []byte("# C"), 0644)

	files, err := FindMDFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("got %d files, want 2", len(files))
	}
}

func TestLoadRules(t *testing.T) {
	root := t.TempDir()
	rulesDir := filepath.Join(root, "spec", "rules", "go")
	os.MkdirAll(rulesDir, 0755)

	content := "---\nscope: language-specific\nlanguages: [go]\nuses_rules: [cross-cutting]\n---\n\n# Go Patterns\n"
	os.WriteFile(filepath.Join(rulesDir, "patterns.md"), []byte(content), 0644)

	rules, err := LoadRules(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 {
		t.Fatalf("got %d rules, want 1", len(rules))
	}

	r := rules[0]
	if r.Name != filepath.Join("go", "patterns") {
		t.Errorf("name = %q", r.Name)
	}
	if r.Description != "Go Patterns" {
		t.Errorf("description = %q", r.Description)
	}
	if r.Scope != "language-specific" {
		t.Errorf("scope = %q", r.Scope)
	}
	if len(r.Languages) != 1 || r.Languages[0] != "go" {
		t.Errorf("languages = %v", r.Languages)
	}
	if len(r.UsesRules) != 1 || r.UsesRules[0] != "cross-cutting" {
		t.Errorf("uses_rules = %v", r.UsesRules)
	}
}

func TestLoadSkills(t *testing.T) {
	root := t.TempDir()
	skillsDir := filepath.Join(root, "spec", "skills")
	os.MkdirAll(skillsDir, 0755)

	content := "---\nname: api-designer\nscope: universal\nuses_rules: [api-design]\n---\n\n# API Designer\n"
	os.WriteFile(filepath.Join(skillsDir, "api-designer.md"), []byte(content), 0644)

	skills, err := LoadSkills(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 1 {
		t.Fatalf("got %d skills, want 1", len(skills))
	}

	s := skills[0]
	if s.Name != "api-designer" {
		t.Errorf("name = %q", s.Name)
	}
	if s.Description != "API Designer" {
		t.Errorf("description = %q", s.Description)
	}
}

func TestLoadAgents(t *testing.T) {
	root := t.TempDir()
	agentsDir := filepath.Join(root, "spec", "agents", "go")
	os.MkdirAll(agentsDir, 0755)

	content := "---\nname: go-coder\ndescription: Go code writer\nrole: coder\nscope: language-specific\naccess: read-write\nlanguages: [go]\nuses_skills: [golang-pro]\nuses_rules: [go/patterns]\nuses_plugins: [context7]\ndelegates_to: []\n---\n\n# Go Coder\n"
	os.WriteFile(filepath.Join(agentsDir, "coder.md"), []byte(content), 0644)

	agents, err := LoadAgents(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(agents) != 1 {
		t.Fatalf("got %d agents, want 1", len(agents))
	}

	a := agents[0]
	if a.Name != "go-coder" {
		t.Errorf("name = %q", a.Name)
	}
	if a.Description != "Go code writer" {
		t.Errorf("description = %q", a.Description)
	}
	if a.Role != "coder" {
		t.Errorf("role = %q", a.Role)
	}
	if a.Access != "read-write" {
		t.Errorf("access = %q", a.Access)
	}
}

func TestLoadPlugins(t *testing.T) {
	root := t.TempDir()
	config := `{"public_plugins": [{"name": "context7"}, {"name": "superpowers"}]}`
	os.WriteFile(filepath.Join(root, "config.json"), []byte(config), 0644)

	plugins, err := LoadPlugins(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(plugins) != 2 {
		t.Fatalf("got %d plugins, want 2", len(plugins))
	}
	if plugins[0] != "context7" || plugins[1] != "superpowers" {
		t.Errorf("plugins = %v", plugins)
	}
}

func TestLoadPlugins_NoFile(t *testing.T) {
	root := t.TempDir()
	plugins, err := LoadPlugins(root)
	if err != nil {
		t.Fatal(err)
	}
	if plugins != nil {
		t.Errorf("got %v, want nil", plugins)
	}
}

func TestLoadRules_NoDir(t *testing.T) {
	root := t.TempDir()
	rules, err := LoadRules(root)
	if err != nil {
		t.Fatal(err)
	}
	if rules != nil {
		t.Errorf("got %v, want nil", rules)
	}
}
