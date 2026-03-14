package catalog

import (
	"encoding/json"
	"testing"

	"github.com/smahovkic/agent-army/armyv2/internal/core/types"
)

func validCatalogJSON(plugins []types.CatalogPlugin, skills []types.CatalogSkill, profiles map[string]types.TechProfile) []byte {
	cat := types.Catalog{
		Version:      1,
		UpdatedAt:    "2026-01-01",
		Plugins:      plugins,
		Skills:       skills,
		TechProfiles: profiles,
	}
	data, _ := json.Marshal(cat)
	return data
}

func minimalCatalogJSON() []byte {
	return validCatalogJSON(
		[]types.CatalogPlugin{
			{Name: "plugin-a", Marketplace: "mkt", Description: "A", Tags: []string{"t1"}},
			{Name: "plugin-b", Marketplace: "mkt", Description: "B", Tags: []string{"t2"}},
		},
		[]types.CatalogSkill{
			{Name: "skill-x", Source: "src", Description: "X", Tags: []string{"s1"}},
		},
		map[string]types.TechProfile{
			"go": {Detect: []string{"go.mod"}, Plugins: []string{"plugin-a"}, Skills: []string{"skill-x"}},
		},
	)
}

func TestNewFromBytes(t *testing.T) {
	t.Run("valid catalog", func(t *testing.T) {
		svc, err := NewFromBytes(minimalCatalogJSON())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(svc.AllPlugins()) != 2 {
			t.Errorf("got %d plugins, want 2", len(svc.AllPlugins()))
		}
		if len(svc.AllSkills()) != 1 {
			t.Errorf("got %d skills, want 1", len(svc.AllSkills()))
		}
	})

	t.Run("empty plugins and skills", func(t *testing.T) {
		data := validCatalogJSON(nil, nil, nil)
		svc, err := NewFromBytes(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if svc.AllPlugins() == nil {
			t.Error("AllPlugins should return non-nil slice")
		}
		if svc.AllSkills() == nil {
			t.Error("AllSkills should return non-nil slice")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := NewFromBytes([]byte("not json"))
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})

	t.Run("missing version", func(t *testing.T) {
		_, err := NewFromBytes([]byte(`{"plugins":[]}`))
		if err == nil {
			t.Fatal("expected error for missing version")
		}
	})

	t.Run("version zero", func(t *testing.T) {
		_, err := NewFromBytes([]byte(`{"version":0}`))
		if err == nil {
			t.Fatal("expected error for version 0")
		}
	})
}

func TestFindPlugin(t *testing.T) {
	svc, _ := NewFromBytes(minimalCatalogJSON())

	tests := []struct {
		name  string
		query string
		found bool
	}{
		{"exact match", "plugin-a", true},
		{"case insensitive", "Plugin-A", true},
		{"not found", "nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, ok := svc.FindPlugin(tt.query)
			if ok != tt.found {
				t.Errorf("FindPlugin(%q) found=%v, want %v", tt.query, ok, tt.found)
			}
			if tt.found && p.Name != "plugin-a" {
				t.Errorf("got Name=%q, want plugin-a", p.Name)
			}
		})
	}
}

func TestFindSkill(t *testing.T) {
	svc, _ := NewFromBytes(minimalCatalogJSON())

	tests := []struct {
		name  string
		query string
		found bool
	}{
		{"exact match", "skill-x", true},
		{"case insensitive", "SKILL-X", true},
		{"not found", "nope", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sk, ok := svc.FindSkill(tt.query)
			if ok != tt.found {
				t.Errorf("FindSkill(%q) found=%v, want %v", tt.query, ok, tt.found)
			}
			if tt.found && sk.Name != "skill-x" {
				t.Errorf("got Name=%q, want skill-x", sk.Name)
			}
		})
	}
}

func TestAllPluginsReturnsCopy(t *testing.T) {
	svc, _ := NewFromBytes(minimalCatalogJSON())
	plugins := svc.AllPlugins()
	plugins[0].Name = "mutated"

	original := svc.AllPlugins()
	if original[0].Name == "mutated" {
		t.Error("AllPlugins should return a copy, not a reference to internal state")
	}
}

func TestAllSkillsReturnsCopy(t *testing.T) {
	svc, _ := NewFromBytes(minimalCatalogJSON())
	skills := svc.AllSkills()
	skills[0].Name = "mutated"

	original := svc.AllSkills()
	if original[0].Name == "mutated" {
		t.Error("AllSkills should return a copy, not a reference to internal state")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid minimal", `{"version":1}`, false},
		{"valid full", string(minimalCatalogJSON()), false},
		{"invalid JSON", `{bad`, true},
		{"missing version", `{"plugins":[]}`, true},
		{"string version", `{"version":"1"}`, true},
		{"zero version", `{"version":0}`, true},
		{"negative version", `{"version":-1}`, true},
		{"plugins not array", `{"version":1,"plugins":"bad"}`, true},
		{"skills not array", `{"version":1,"skills":"bad"}`, true},
		{"tech_profiles not object", `{"version":1,"tech_profiles":"bad"}`, true},
		{"plugin missing name", `{"version":1,"plugins":[{}]}`, true},
		{"skill missing name", `{"version":1,"skills":[{}]}`, true},
		{"plugin with name", `{"version":1,"plugins":[{"name":"ok"}]}`, false},
		{"skill with name", `{"version":1,"skills":[{"name":"ok"}]}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestMergeCatalogs(t *testing.T) {
	t.Run("updated wins when version >= base", func(t *testing.T) {
		base := types.Catalog{
			Version: 1,
			Plugins: []types.CatalogPlugin{{Name: "base-plugin"}},
			Skills:  []types.CatalogSkill{{Name: "base-skill"}},
			TechProfiles: map[string]types.TechProfile{
				"go":     {Detect: []string{"go.mod"}},
				"python": {Detect: []string{"*.py"}},
			},
		}
		updated := types.Catalog{
			Version:   2,
			UpdatedAt: "2026-02-01",
			Plugins:   []types.CatalogPlugin{{Name: "updated-plugin"}},
			Skills:    []types.CatalogSkill{{Name: "updated-skill"}},
			TechProfiles: map[string]types.TechProfile{
				"go": {Detect: []string{"go.mod", "go.sum"}},
			},
		}

		merged := mergeCatalogs(base, updated)

		if merged.Version != 2 {
			t.Errorf("got version %d, want 2", merged.Version)
		}
		if len(merged.Plugins) != 1 || merged.Plugins[0].Name != "updated-plugin" {
			t.Error("plugins should come from updated catalog")
		}
		if len(merged.Skills) != 1 || merged.Skills[0].Name != "updated-skill" {
			t.Error("skills should come from updated catalog")
		}
		// Tech profiles are merged: base "python" + updated "go"
		if len(merged.TechProfiles) != 2 {
			t.Errorf("got %d tech profiles, want 2", len(merged.TechProfiles))
		}
		goProfile := merged.TechProfiles["go"]
		if len(goProfile.Detect) != 2 {
			t.Error("go profile should use updated detect list")
		}
		if _, ok := merged.TechProfiles["python"]; !ok {
			t.Error("python profile from base should be preserved")
		}
	})

	t.Run("base wins when updated version < base", func(t *testing.T) {
		base := types.Catalog{
			Version: 3,
			Plugins: []types.CatalogPlugin{{Name: "base-plugin"}},
		}
		updated := types.Catalog{
			Version: 2,
			Plugins: []types.CatalogPlugin{{Name: "old-plugin"}},
		}

		merged := mergeCatalogs(base, updated)
		if merged.Version != 3 {
			t.Errorf("got version %d, want 3", merged.Version)
		}
		if len(merged.Plugins) != 1 || merged.Plugins[0].Name != "base-plugin" {
			t.Error("should use base plugins when updated is older")
		}
	})

	t.Run("nil collections are initialized", func(t *testing.T) {
		base := types.Catalog{Version: 1, TechProfiles: map[string]types.TechProfile{}}
		updated := types.Catalog{Version: 1}

		merged := mergeCatalogs(base, updated)
		if merged.Plugins == nil {
			t.Error("plugins should be non-nil")
		}
		if merged.Skills == nil {
			t.Error("skills should be non-nil")
		}
	})
}

func TestNew(t *testing.T) {
	// New() loads the embedded catalog.json, which should always succeed.
	svc, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if len(svc.AllPlugins()) == 0 {
		t.Error("embedded catalog should have plugins")
	}
	if len(svc.AllSkills()) == 0 {
		t.Error("embedded catalog should have skills")
	}
}

func TestGetTechProfile(t *testing.T) {
	svc, _ := NewFromBytes(minimalCatalogJSON())

	t.Run("found", func(t *testing.T) {
		p, ok := svc.GetTechProfile("go")
		if !ok {
			t.Fatal("expected to find 'go' profile")
		}
		if len(p.Detect) == 0 {
			t.Error("expected non-empty detect list")
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		_, ok := svc.GetTechProfile("GO")
		if !ok {
			t.Error("GetTechProfile should be case-insensitive")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, ok := svc.GetTechProfile("rust")
		if ok {
			t.Error("expected not found for 'rust'")
		}
	})
}
