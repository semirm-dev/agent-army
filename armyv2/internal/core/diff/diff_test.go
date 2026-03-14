package diff

import (
	"testing"

	"github.com/smahovkic/agent-army/armyv2/internal/core/types"
)

func TestCompare_NoDrift(t *testing.T) {
	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{{Name: "p1"}, {Name: "p2"}},
		Skills:  []types.ManifestSkill{{Name: "s1"}},
	}
	plugins := []types.InstalledPlugin{{Name: "p1"}, {Name: "p2"}}
	skills := []types.InstalledSkill{{Name: "s1"}}

	result := Compare(manifest, plugins, skills)

	if HasDrift(result) {
		t.Error("expected no drift")
	}
	if len(result.MissingPlugins) != 0 {
		t.Errorf("MissingPlugins: got %d, want 0", len(result.MissingPlugins))
	}
	if len(result.ExtraPlugins) != 0 {
		t.Errorf("ExtraPlugins: got %d, want 0", len(result.ExtraPlugins))
	}
	if len(result.MissingSkills) != 0 {
		t.Errorf("MissingSkills: got %d, want 0", len(result.MissingSkills))
	}
	if len(result.ExtraSkills) != 0 {
		t.Errorf("ExtraSkills: got %d, want 0", len(result.ExtraSkills))
	}
}

func TestCompare_MissingPlugins(t *testing.T) {
	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{{Name: "p1"}, {Name: "p2"}},
		Skills:  []types.ManifestSkill{},
	}
	plugins := []types.InstalledPlugin{{Name: "p1"}}

	result := Compare(manifest, plugins, nil)

	if !HasDrift(result) {
		t.Error("expected drift")
	}
	if len(result.MissingPlugins) != 1 || result.MissingPlugins[0].Name != "p2" {
		t.Errorf("MissingPlugins: got %v", result.MissingPlugins)
	}
}

func TestCompare_ExtraPlugins(t *testing.T) {
	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{{Name: "p1"}},
		Skills:  []types.ManifestSkill{},
	}
	plugins := []types.InstalledPlugin{{Name: "p1"}, {Name: "extra"}}

	result := Compare(manifest, plugins, nil)

	if !HasDrift(result) {
		t.Error("expected drift")
	}
	if len(result.ExtraPlugins) != 1 || result.ExtraPlugins[0].Name != "extra" {
		t.Errorf("ExtraPlugins: got %v", result.ExtraPlugins)
	}
}

func TestCompare_MissingSkills(t *testing.T) {
	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{},
		Skills:  []types.ManifestSkill{{Name: "s1"}, {Name: "s2"}},
	}
	skills := []types.InstalledSkill{{Name: "s1"}}

	result := Compare(manifest, nil, skills)

	if len(result.MissingSkills) != 1 || result.MissingSkills[0].Name != "s2" {
		t.Errorf("MissingSkills: got %v", result.MissingSkills)
	}
}

func TestCompare_ExtraSkills(t *testing.T) {
	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{},
		Skills:  []types.ManifestSkill{},
	}
	skills := []types.InstalledSkill{{Name: "orphan"}}

	result := Compare(manifest, nil, skills)

	if len(result.ExtraSkills) != 1 || result.ExtraSkills[0].Name != "orphan" {
		t.Errorf("ExtraSkills: got %v", result.ExtraSkills)
	}
}

func TestCompare_MixedDrift(t *testing.T) {
	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{{Name: "wanted-p"}, {Name: "missing-p"}},
		Skills:  []types.ManifestSkill{{Name: "wanted-s"}, {Name: "missing-s"}},
	}
	plugins := []types.InstalledPlugin{{Name: "wanted-p"}, {Name: "extra-p"}}
	skills := []types.InstalledSkill{{Name: "wanted-s"}, {Name: "extra-s"}}

	result := Compare(manifest, plugins, skills)

	if !HasDrift(result) {
		t.Error("expected drift")
	}
	if len(result.MissingPlugins) != 1 {
		t.Errorf("MissingPlugins: got %d, want 1", len(result.MissingPlugins))
	}
	if len(result.ExtraPlugins) != 1 {
		t.Errorf("ExtraPlugins: got %d, want 1", len(result.ExtraPlugins))
	}
	if len(result.MissingSkills) != 1 {
		t.Errorf("MissingSkills: got %d, want 1", len(result.MissingSkills))
	}
	if len(result.ExtraSkills) != 1 {
		t.Errorf("ExtraSkills: got %d, want 1", len(result.ExtraSkills))
	}
}

func TestCompare_EmptyManifestAndInstalled(t *testing.T) {
	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{},
		Skills:  []types.ManifestSkill{},
	}

	result := Compare(manifest, nil, nil)

	if HasDrift(result) {
		t.Error("empty manifest and no installed items should have no drift")
	}
}

func TestHasDrift(t *testing.T) {
	tests := []struct {
		name string
		diff types.DiffResult
		want bool
	}{
		{"empty", types.DiffResult{}, false},
		{"missing plugin", types.DiffResult{MissingPlugins: []types.ManifestPlugin{{Name: "p"}}}, true},
		{"extra plugin", types.DiffResult{ExtraPlugins: []types.InstalledPlugin{{Name: "p"}}}, true},
		{"missing skill", types.DiffResult{MissingSkills: []types.ManifestSkill{{Name: "s"}}}, true},
		{"extra skill", types.DiffResult{ExtraSkills: []types.InstalledSkill{{Name: "s"}}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasDrift(tt.diff); got != tt.want {
				t.Errorf("HasDrift() = %v, want %v", got, tt.want)
			}
		})
	}
}
