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
	errors := filterBySeverity(issues, "error")
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errors), errors)
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

func TestCheck_SkillOrphanDir(t *testing.T) {
	home, _ := os.UserHomeDir()
	orphanName := "orphan-dir-test-xyz-99999"
	orphanDir := filepath.Join(home, ".agents", "skills", orphanName)
	os.MkdirAll(orphanDir, 0o755)
	defer os.RemoveAll(orphanDir)

	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{},
		Skills:  []types.ManifestSkill{},
	}

	// No installed skills — the directory should be flagged as orphan.
	issues := Check(manifest, nil, nil)

	found := findIssue(issues, "drift", orphanName)
	if found == nil {
		t.Error("expected drift warning for orphan skill directory")
	}
	if found != nil && found.Severity != "warning" {
		t.Errorf("orphan dir severity: got %q, want warning", found.Severity)
	}
}

func TestCheck_SkillOrphanDir_NoOrphans(t *testing.T) {
	home, _ := os.UserHomeDir()
	skillName := "tracked-skill-test-xyz-99999"
	skillDir := filepath.Join(home, ".agents", "skills", skillName)
	os.MkdirAll(skillDir, 0o755)
	defer os.RemoveAll(skillDir)

	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{},
		Skills:  []types.ManifestSkill{},
	}
	// Skill is in the installed list — should not be flagged.
	skills := []types.InstalledSkill{{Name: skillName}}

	issues := Check(manifest, nil, skills)

	found := findIssue(issues, "drift", skillName)
	if found != nil {
		t.Error("should not report drift when skill dir is tracked in lock file")
	}
}

func TestCheck_PluginDiskDrift(t *testing.T) {
	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{},
		Skills:  []types.ManifestSkill{},
	}
	plugins := []types.InstalledPlugin{{
		Name:        "ghost-plugin",
		InstallPath: "/tmp/nonexistent-plugin-path-xyz-99999",
	}}

	issues := Check(manifest, plugins, nil)

	found := findIssue(issues, "drift", "ghost-plugin")
	if found == nil {
		t.Error("expected drift warning for plugin with missing installPath")
	}
	if found != nil && found.Severity != "warning" {
		t.Errorf("plugin drift severity: got %q, want warning", found.Severity)
	}
}

func TestCheck_PluginDiskDrift_Exists(t *testing.T) {
	tmpDir := t.TempDir()

	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{},
		Skills:  []types.ManifestSkill{},
	}
	plugins := []types.InstalledPlugin{{
		Name:        "real-plugin",
		InstallPath: tmpDir,
	}}

	issues := Check(manifest, plugins, nil)

	found := findIssue(issues, "drift", "real-plugin")
	if found != nil {
		t.Error("should not report drift when plugin installPath exists")
	}
}

func TestCheck_PluginDiskDrift_EmptyPath(t *testing.T) {
	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{},
		Skills:  []types.ManifestSkill{},
	}
	plugins := []types.InstalledPlugin{{
		Name:        "no-path-plugin",
		InstallPath: "",
	}}

	issues := Check(manifest, plugins, nil)

	found := findIssue(issues, "drift", "no-path-plugin")
	if found != nil {
		t.Error("should not report drift when plugin has no installPath")
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
	errors := filterBySeverity(issues, "error")
	if len(errors) != 0 {
		t.Errorf("expected no errors for empty state, got %d", len(errors))
	}
}

// filterBySeverity returns issues matching the given severity.
func filterBySeverity(issues []types.DoctorIssue, severity string) []types.DoctorIssue {
	var filtered []types.DoctorIssue
	for _, issue := range issues {
		if issue.Severity == severity {
			filtered = append(filtered, issue)
		}
	}
	return filtered
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
