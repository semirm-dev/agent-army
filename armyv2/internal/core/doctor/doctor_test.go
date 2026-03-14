package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/smahovkic/agent-army/armyv2/internal/core/types"
)

func TestCheck_NoIssues(t *testing.T) {
	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{{Name: "p1"}},
		Skills:  []types.ManifestSkill{{Name: "s1"}},
	}
	plugins := []types.InstalledPlugin{{Name: "p1"}}

	// Create the skill directory on disk so checkSkillDiskDrift passes.
	home, _ := os.UserHomeDir()
	skillDir := filepath.Join(home, ".agents", "skills", "s1")
	os.MkdirAll(skillDir, 0o755)
	defer os.RemoveAll(skillDir)

	skills := []types.InstalledSkill{{Name: "s1"}}

	issues := Check(manifest, plugins, skills)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d: %v", len(issues), issues)
	}
}

func TestCheck_MissingPlugin(t *testing.T) {
	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{{Name: "missing-plugin"}},
		Skills:  []types.ManifestSkill{},
	}

	issues := Check(manifest, nil, nil)

	found := findIssue(issues, "missing", "missing-plugin")
	if found == nil {
		t.Error("expected missing plugin issue")
	}
	if found != nil && found.Severity != "error" {
		t.Errorf("missing plugin severity: got %q, want error", found.Severity)
	}
}

func TestCheck_MissingSkill(t *testing.T) {
	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{},
		Skills:  []types.ManifestSkill{{Name: "missing-skill"}},
	}

	issues := Check(manifest, nil, nil)

	found := findIssue(issues, "missing", "missing-skill")
	if found == nil {
		t.Error("expected missing skill issue")
	}
}

func TestCheck_OrphanPlugin(t *testing.T) {
	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{},
		Skills:  []types.ManifestSkill{},
	}
	plugins := []types.InstalledPlugin{{Name: "orphan-plugin"}}

	issues := Check(manifest, plugins, nil)

	found := findIssue(issues, "orphan", "orphan-plugin")
	if found == nil {
		t.Error("expected orphan plugin issue")
	}
	if found != nil && found.Severity != "warning" {
		t.Errorf("orphan plugin severity: got %q, want warning", found.Severity)
	}
}

func TestCheck_OrphanSkill(t *testing.T) {
	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{},
		Skills:  []types.ManifestSkill{},
	}
	skills := []types.InstalledSkill{{Name: "orphan-skill"}}

	issues := Check(manifest, nil, skills)

	found := findIssue(issues, "orphan", "orphan-skill")
	if found == nil {
		t.Error("expected orphan skill issue")
	}
	if found != nil && found.Severity != "warning" {
		t.Errorf("orphan skill severity: got %q, want warning", found.Severity)
	}
}

func TestCheck_SkillDiskDrift(t *testing.T) {
	// Use a skill name that definitely won't exist on disk.
	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{},
		Skills:  []types.ManifestSkill{},
	}
	skills := []types.InstalledSkill{{Name: "nonexistent-skill-xyz-test-12345"}}

	issues := Check(manifest, nil, skills)

	found := findIssue(issues, "drift", "nonexistent-skill-xyz-test-12345")
	if found == nil {
		t.Error("expected drift issue for skill missing from disk")
	}
	if found != nil && found.Severity != "error" {
		t.Errorf("drift severity: got %q, want error", found.Severity)
	}
}

func TestCheck_SkillDiskDrift_Exists(t *testing.T) {
	// Create the skill directory so there's no drift.
	home, _ := os.UserHomeDir()
	skillName := "test-drift-skill-exists"
	skillDir := filepath.Join(home, ".agents", "skills", skillName)
	os.MkdirAll(skillDir, 0o755)
	defer os.RemoveAll(skillDir)

	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{},
		Skills:  []types.ManifestSkill{},
	}
	skills := []types.InstalledSkill{{Name: skillName}}

	issues := Check(manifest, nil, skills)

	driftIssue := findIssue(issues, "drift", skillName)
	if driftIssue != nil {
		t.Error("should not report drift when skill dir exists on disk")
	}
}

func TestCheck_MultipleIssues(t *testing.T) {
	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{{Name: "wanted-p"}},
		Skills:  []types.ManifestSkill{{Name: "wanted-s"}},
	}
	plugins := []types.InstalledPlugin{{Name: "orphan-p"}}
	skills := []types.InstalledSkill{{Name: "orphan-s"}}

	issues := Check(manifest, plugins, skills)

	// Should have: missing plugin, missing skill, orphan plugin, orphan skill
	// Plus potential drift issues for orphan-s
	if len(issues) < 4 {
		t.Errorf("expected at least 4 issues, got %d", len(issues))
	}

	if findIssue(issues, "missing", "wanted-p") == nil {
		t.Error("missing wanted-p issue")
	}
	if findIssue(issues, "missing", "wanted-s") == nil {
		t.Error("missing wanted-s issue")
	}
	if findIssue(issues, "orphan", "orphan-p") == nil {
		t.Error("missing orphan-p issue")
	}
	if findIssue(issues, "orphan", "orphan-s") == nil {
		t.Error("missing orphan-s issue")
	}
}

func TestCheck_EmptyEverything(t *testing.T) {
	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{},
		Skills:  []types.ManifestSkill{},
	}

	issues := Check(manifest, nil, nil)
	if len(issues) != 0 {
		t.Errorf("expected no issues for empty state, got %d", len(issues))
	}
}

// findIssue returns the first issue matching the given category and item, or nil.
func findIssue(issues []types.DoctorIssue, category, item string) *types.DoctorIssue {
	for i, issue := range issues {
		if issue.Category == category && issue.Item == item {
			return &issues[i]
		}
	}
	return nil
}
