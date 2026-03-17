package orchestrator

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/smahovkic/agent-army/army/internal/core/types"
)

// --- Test doubles ---

type mockPluginInstaller struct {
	installed []string
	removed   []string
	failOn    map[string]bool
}

func (m *mockPluginInstaller) Install(name string) error {
	if m.failOn[name] {
		return fmt.Errorf("mock install fail: %s", name)
	}
	m.installed = append(m.installed, name)
	return nil
}

func (m *mockPluginInstaller) Remove(name string) error {
	if m.failOn[name] {
		return fmt.Errorf("mock remove fail: %s", name)
	}
	m.removed = append(m.removed, name)
	return nil
}

type mockSkillInstaller struct {
	installed []string
	removed   []string
	failOn    map[string]bool
}

func (m *mockSkillInstaller) Install(name, source string) error {
	if m.failOn[name] {
		return fmt.Errorf("mock install fail: %s", name)
	}
	m.installed = append(m.installed, name)
	return nil
}

func (m *mockSkillInstaller) Remove(name string) error {
	if m.failOn[name] {
		return fmt.Errorf("mock remove fail: %s", name)
	}
	m.removed = append(m.removed, name)
	return nil
}

type mockSystemReader struct {
	plugins []types.InstalledPlugin
	skills  []types.InstalledSkill
}

func (m *mockSystemReader) InstalledPlugins() ([]types.InstalledPlugin, error) {
	return m.plugins, nil
}

func (m *mockSystemReader) InstalledSkills() ([]types.InstalledSkill, error) {
	return m.skills, nil
}

// --- Tests ---

func TestExecute_RemoveMultiplePlugins(t *testing.T) {
	pi := &mockPluginInstaller{}
	si := &mockSkillInstaller{}
	orch := New(pi, si, &mockSystemReader{}, &bytes.Buffer{})

	actions := []types.Action{
		{Type: "remove", ItemType: "plugin", Name: "p1"},
		{Type: "remove", ItemType: "plugin", Name: "p2"},
		{Type: "remove", ItemType: "plugin", Name: "p3"},
	}

	result := orch.Execute(actions)

	if result.Succeeded != 3 {
		t.Errorf("Succeeded: got %d, want 3", result.Succeeded)
	}
	if result.Failed != 0 {
		t.Errorf("Failed: got %d, want 0", result.Failed)
	}
	if len(pi.removed) != 3 {
		t.Fatalf("removed count: got %d, want 3", len(pi.removed))
	}
	// Verify sequential order matches input order
	for i, name := range []string{"p1", "p2", "p3"} {
		if pi.removed[i] != name {
			t.Errorf("removed[%d]: got %q, want %q", i, pi.removed[i], name)
		}
	}
}

func TestExecute_RemoveMultipleSkills(t *testing.T) {
	pi := &mockPluginInstaller{}
	si := &mockSkillInstaller{}
	orch := New(pi, si, &mockSystemReader{}, &bytes.Buffer{})

	actions := []types.Action{
		{Type: "remove", ItemType: "skill", Name: "s1"},
		{Type: "remove", ItemType: "skill", Name: "s2"},
		{Type: "remove", ItemType: "skill", Name: "s3"},
	}

	result := orch.Execute(actions)

	if result.Succeeded != 3 {
		t.Errorf("Succeeded: got %d, want 3", result.Succeeded)
	}
	if len(si.removed) != 3 {
		t.Fatalf("removed count: got %d, want 3", len(si.removed))
	}
	for i, name := range []string{"s1", "s2", "s3"} {
		if si.removed[i] != name {
			t.Errorf("removed[%d]: got %q, want %q", i, si.removed[i], name)
		}
	}
}

func TestExecute_MixedActions(t *testing.T) {
	pi := &mockPluginInstaller{}
	si := &mockSkillInstaller{}
	orch := New(pi, si, &mockSystemReader{}, &bytes.Buffer{})

	actions := []types.Action{
		{Type: "install", ItemType: "plugin", Name: "p1"},
		{Type: "remove", ItemType: "plugin", Name: "p2"},
		{Type: "install", ItemType: "skill", Name: "s1", Source: "src"},
		{Type: "remove", ItemType: "skill", Name: "s2"},
	}

	result := orch.Execute(actions)

	if result.Succeeded != 4 {
		t.Errorf("Succeeded: got %d, want 4", result.Succeeded)
	}
	if result.Failed != 0 {
		t.Errorf("Failed: got %d, want 0", result.Failed)
	}
	if len(pi.installed) != 1 || pi.installed[0] != "p1" {
		t.Errorf("plugins installed: got %v, want [p1]", pi.installed)
	}
	if len(pi.removed) != 1 || pi.removed[0] != "p2" {
		t.Errorf("plugins removed: got %v, want [p2]", pi.removed)
	}
	if len(si.installed) != 1 || si.installed[0] != "s1" {
		t.Errorf("skills installed: got %v, want [s1]", si.installed)
	}
	if len(si.removed) != 1 || si.removed[0] != "s2" {
		t.Errorf("skills removed: got %v, want [s2]", si.removed)
	}
}

func TestExecute_PartialFailure(t *testing.T) {
	pi := &mockPluginInstaller{failOn: map[string]bool{"p2": true}}
	si := &mockSkillInstaller{}
	orch := New(pi, si, &mockSystemReader{}, &bytes.Buffer{})

	actions := []types.Action{
		{Type: "remove", ItemType: "plugin", Name: "p1"},
		{Type: "remove", ItemType: "plugin", Name: "p2"},
		{Type: "remove", ItemType: "plugin", Name: "p3"},
	}

	result := orch.Execute(actions)

	if result.Succeeded != 2 {
		t.Errorf("Succeeded: got %d, want 2", result.Succeeded)
	}
	if result.Failed != 1 {
		t.Errorf("Failed: got %d, want 1", result.Failed)
	}
	if len(result.Errors) != 1 {
		t.Errorf("Errors count: got %d, want 1", len(result.Errors))
	}
	// p1 and p3 should still be removed despite p2 failing
	if len(pi.removed) != 2 {
		t.Fatalf("removed count: got %d, want 2", len(pi.removed))
	}
	if pi.removed[0] != "p1" || pi.removed[1] != "p3" {
		t.Errorf("removed: got %v, want [p1 p3]", pi.removed)
	}
}

func TestExecute_Empty(t *testing.T) {
	pi := &mockPluginInstaller{}
	si := &mockSkillInstaller{}
	orch := New(pi, si, &mockSystemReader{}, &bytes.Buffer{})

	result := orch.Execute(nil)

	if result.Succeeded != 0 || result.Failed != 0 {
		t.Errorf("got %d succeeded, %d failed; want 0, 0", result.Succeeded, result.Failed)
	}
}

func TestPlanActions_OrphansDetected(t *testing.T) {
	pi := &mockPluginInstaller{}
	si := &mockSkillInstaller{}
	sys := &mockSystemReader{
		plugins: []types.InstalledPlugin{
			{Name: "wanted", Marketplace: "m"},
			{Name: "orphan-p1", Marketplace: "m"},
			{Name: "orphan-p2", Marketplace: "m"},
		},
		skills: []types.InstalledSkill{
			{Name: "wanted-s", Source: "src"},
			{Name: "orphan-s1", Source: "src"},
		},
	}
	orch := New(pi, si, sys, &bytes.Buffer{})

	manifest := &types.Manifest{
		Plugins: []types.ManifestPlugin{{Name: "wanted", Marketplace: "m"}},
		Skills:  []types.ManifestSkill{{Name: "wanted-s", Source: "src"}},
	}

	actions, err := orch.PlanActions(manifest)
	if err != nil {
		t.Fatalf("PlanActions: %v", err)
	}

	var removePlugins, removeSkills int
	for _, a := range actions {
		if a.Type == "remove" && a.ItemType == "plugin" {
			removePlugins++
		}
		if a.Type == "remove" && a.ItemType == "skill" {
			removeSkills++
		}
	}

	if removePlugins != 2 {
		t.Errorf("remove plugin actions: got %d, want 2", removePlugins)
	}
	if removeSkills != 1 {
		t.Errorf("remove skill actions: got %d, want 1", removeSkills)
	}
}
