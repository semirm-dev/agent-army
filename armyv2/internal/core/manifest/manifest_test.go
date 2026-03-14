package manifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/smahovkic/agent-army/armyv2/internal/core/types"
)

func TestLoad_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	m, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Version != 1 {
		t.Errorf("got version %d, want 1", m.Version)
	}
	if m.Plugins == nil || m.Skills == nil {
		t.Error("slices should be non-nil")
	}
	if len(m.Plugins) != 0 || len(m.Skills) != 0 {
		t.Error("slices should be empty")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	os.WriteFile(path, []byte("not json"), 0o644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "manifest.json")

	original := &types.Manifest{
		Version: 1,
		Plugins: []types.ManifestPlugin{
			{Name: "my-plugin", Marketplace: "mkt", Tags: []string{"a"}, Destination: "user"},
		},
		Skills: []types.ManifestSkill{
			{Name: "my-skill", Source: "src/repo", Tags: []string{"b"}, Destination: "project"},
		},
	}

	if err := Save(path, original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Version != original.Version {
		t.Errorf("version: got %d, want %d", loaded.Version, original.Version)
	}
	if len(loaded.Plugins) != 1 || loaded.Plugins[0].Name != "my-plugin" {
		t.Error("plugin round-trip failed")
	}
	if len(loaded.Skills) != 1 || loaded.Skills[0].Name != "my-skill" {
		t.Error("skill round-trip failed")
	}
	if loaded.Skills[0].Destination != "project" {
		t.Errorf("skill destination: got %q, want project", loaded.Skills[0].Destination)
	}
}

func TestSave_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "manifest.json")

	m := &types.Manifest{Version: 1, Plugins: []types.ManifestPlugin{}, Skills: []types.ManifestSkill{}}
	if err := Save(path, m); err != nil {
		t.Fatalf("Save should create parent dirs: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file should exist: %v", err)
	}
}

func TestSave_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")

	m := &types.Manifest{Version: 1, Plugins: []types.ManifestPlugin{}, Skills: []types.ManifestSkill{}}
	if err := Save(path, m); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file ends with newline (per implementation)
	data, _ := os.ReadFile(path)
	if len(data) == 0 {
		t.Fatal("file is empty")
	}
	if data[len(data)-1] != '\n' {
		t.Error("file should end with newline")
	}

	// Verify it's valid JSON
	var check types.Manifest
	if err := json.Unmarshal(data, &check); err != nil {
		t.Errorf("saved file is not valid JSON: %v", err)
	}
}

func TestLoad_ZeroVersionDefaultsToOne(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	os.WriteFile(path, []byte(`{"version":0,"plugins":[],"skills":[]}`), 0o644)

	m, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Version != 1 {
		t.Errorf("got version %d, want 1 for zero-version manifest", m.Version)
	}
}

func TestAddPlugin(t *testing.T) {
	m := emptyManifest()
	p := types.ManifestPlugin{Name: "test-plugin", Marketplace: "mkt"}

	if !AddPlugin(m, p) {
		t.Error("first AddPlugin should return true")
	}
	if len(m.Plugins) != 1 {
		t.Fatalf("got %d plugins, want 1", len(m.Plugins))
	}

	// Duplicate (case-insensitive)
	dup := types.ManifestPlugin{Name: "Test-Plugin", Marketplace: "other"}
	if AddPlugin(m, dup) {
		t.Error("duplicate AddPlugin should return false")
	}
	if len(m.Plugins) != 1 {
		t.Error("duplicate should not be added")
	}
}

func TestRemovePlugin(t *testing.T) {
	m := emptyManifest()
	AddPlugin(m, types.ManifestPlugin{Name: "keep-me"})
	AddPlugin(m, types.ManifestPlugin{Name: "remove-me"})

	if !RemovePlugin(m, "Remove-Me") {
		t.Error("RemovePlugin should return true for existing plugin (case-insensitive)")
	}
	if len(m.Plugins) != 1 {
		t.Errorf("got %d plugins, want 1", len(m.Plugins))
	}
	if m.Plugins[0].Name != "keep-me" {
		t.Errorf("wrong plugin remaining: %q", m.Plugins[0].Name)
	}

	if RemovePlugin(m, "nonexistent") {
		t.Error("RemovePlugin should return false for missing plugin")
	}
}

func TestAddSkill(t *testing.T) {
	m := emptyManifest()
	s := types.ManifestSkill{Name: "test-skill", Source: "src"}

	if !AddSkill(m, s) {
		t.Error("first AddSkill should return true")
	}
	if len(m.Skills) != 1 {
		t.Fatalf("got %d skills, want 1", len(m.Skills))
	}

	dup := types.ManifestSkill{Name: "TEST-SKILL", Source: "other"}
	if AddSkill(m, dup) {
		t.Error("duplicate AddSkill should return false")
	}
	if len(m.Skills) != 1 {
		t.Error("duplicate should not be added")
	}
}

func TestRemoveSkill(t *testing.T) {
	m := emptyManifest()
	AddSkill(m, types.ManifestSkill{Name: "keep"})
	AddSkill(m, types.ManifestSkill{Name: "drop"})

	if !RemoveSkill(m, "DROP") {
		t.Error("RemoveSkill should return true (case-insensitive)")
	}
	if len(m.Skills) != 1 || m.Skills[0].Name != "keep" {
		t.Error("wrong skill remaining after removal")
	}

	if RemoveSkill(m, "nope") {
		t.Error("RemoveSkill should return false for missing skill")
	}
}

func TestHasPlugin(t *testing.T) {
	m := emptyManifest()
	AddPlugin(m, types.ManifestPlugin{Name: "exists"})

	tests := []struct {
		name string
		want bool
	}{
		{"exists", true},
		{"EXISTS", true},
		{"Exists", true},
		{"nope", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasPlugin(m, tt.name); got != tt.want {
				t.Errorf("HasPlugin(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestHasSkill(t *testing.T) {
	m := emptyManifest()
	AddSkill(m, types.ManifestSkill{Name: "present"})

	tests := []struct {
		name string
		want bool
	}{
		{"present", true},
		{"PRESENT", true},
		{"absent", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasSkill(m, tt.name); got != tt.want {
				t.Errorf("HasSkill(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
