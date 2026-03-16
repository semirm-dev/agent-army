package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFrom_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != 1 {
		t.Errorf("got version %d, want 1", cfg.Version)
	}
	if cfg.DirMap == nil {
		t.Error("DirMap should be non-nil")
	}
	if len(cfg.DirMap) != 0 {
		t.Error("DirMap should be empty")
	}
}

func TestLoadFrom_InvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	os.WriteFile(path, []byte("not json"), 0o644)

	_, err := LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestSaveToAndLoadFrom_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "config.json")

	original := &Config{
		Version: 1,
		DirMap: map[string]string{
			"/projects/app":   "/projects/app/manifest.json",
			"/projects/other": "/home/user/.army/manifests/other.json",
		},
	}

	if err := SaveTo(path, original); err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if loaded.Version != original.Version {
		t.Errorf("version: got %d, want %d", loaded.Version, original.Version)
	}
	if len(loaded.DirMap) != 2 {
		t.Fatalf("DirMap length: got %d, want 2", len(loaded.DirMap))
	}
	if loaded.DirMap["/projects/app"] != "/projects/app/manifest.json" {
		t.Error("DirMap round-trip failed for /projects/app")
	}
	if loaded.DirMap["/projects/other"] != "/home/user/.army/manifests/other.json" {
		t.Error("DirMap round-trip failed for /projects/other")
	}
}

func TestSaveTo_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "config.json")

	cfg := emptyConfig()
	if err := SaveTo(path, cfg); err != nil {
		t.Fatalf("SaveTo should create parent dirs: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file should exist: %v", err)
	}
}

func TestSaveTo_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := emptyConfig()
	if err := SaveTo(path, cfg); err != nil {
		t.Fatalf("SaveTo failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	if len(data) == 0 {
		t.Fatal("file is empty")
	}
	if data[len(data)-1] != '\n' {
		t.Error("file should end with newline")
	}

	var check Config
	if err := json.Unmarshal(data, &check); err != nil {
		t.Errorf("saved file is not valid JSON: %v", err)
	}
}

func TestResolve_ExactMatch(t *testing.T) {
	cfg := &Config{
		Version: 1,
		DirMap: map[string]string{
			"/projects/app": "/projects/app/manifest.json",
		},
	}

	got := Resolve(cfg, "/projects/app")
	if got != "/projects/app/manifest.json" {
		t.Errorf("got %q, want /projects/app/manifest.json", got)
	}
}

func TestResolve_ParentWalk(t *testing.T) {
	cfg := &Config{
		Version: 1,
		DirMap: map[string]string{
			"/projects": "/projects/manifest.json",
		},
	}

	got := Resolve(cfg, "/projects/app/src/pkg")
	if got != "/projects/manifest.json" {
		t.Errorf("got %q, want /projects/manifest.json", got)
	}
}

func TestResolve_DeepestWins(t *testing.T) {
	cfg := &Config{
		Version: 1,
		DirMap: map[string]string{
			"/projects":     "/projects/manifest.json",
			"/projects/app": "/projects/app/manifest.json",
		},
	}

	got := Resolve(cfg, "/projects/app/src")
	if got != "/projects/app/manifest.json" {
		t.Errorf("got %q, want /projects/app/manifest.json", got)
	}
}

func TestResolve_NoMatch(t *testing.T) {
	cfg := emptyConfig()

	got := Resolve(cfg, "/some/random/path")
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestRegister_AbsolutePaths(t *testing.T) {
	dir := t.TempDir()
	cfg := emptyConfig()

	manifestPath := filepath.Join(dir, "manifest.json")
	defaultPath := "/not/the/default"

	if err := Register(cfg, dir, manifestPath, defaultPath); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	absDir := filepath.Clean(dir)
	absManifest := filepath.Clean(manifestPath)

	if got, ok := cfg.DirMap[absDir]; !ok {
		t.Error("expected entry for dir")
	} else if got != absManifest {
		t.Errorf("got %q, want %q", got, absManifest)
	}
}

func TestRegister_DefaultRemoves(t *testing.T) {
	dir := t.TempDir()
	defaultPath := filepath.Join(dir, "default-manifest.json")

	cfg := &Config{
		Version: 1,
		DirMap: map[string]string{
			dir: filepath.Join(dir, "old-manifest.json"),
		},
	}

	// Registering with the default path should remove the entry.
	if err := Register(cfg, dir, defaultPath, defaultPath); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if _, ok := cfg.DirMap[dir]; ok {
		t.Error("entry should have been removed when registering default path")
	}
}

func TestRegister_Overwrite(t *testing.T) {
	dir := t.TempDir()
	cfg := emptyConfig()
	defaultPath := "/not/the/default"

	first := filepath.Join(dir, "first.json")
	second := filepath.Join(dir, "second.json")

	Register(cfg, dir, first, defaultPath)
	Register(cfg, dir, second, defaultPath)

	absDir := filepath.Clean(dir)
	if got := cfg.DirMap[absDir]; got != filepath.Clean(second) {
		t.Errorf("got %q, want %q", got, filepath.Clean(second))
	}
}

func TestRemove_Existing(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		Version: 1,
		DirMap: map[string]string{
			dir: "/some/manifest.json",
		},
	}

	Remove(cfg, dir)

	if _, ok := cfg.DirMap[dir]; ok {
		t.Error("entry should have been removed")
	}
}

func TestRemove_Nonexistent(t *testing.T) {
	cfg := emptyConfig()

	// Should not panic or error.
	Remove(cfg, "/nonexistent/path")

	if len(cfg.DirMap) != 0 {
		t.Error("DirMap should still be empty")
	}
}

func TestLoadFrom_ZeroVersionDefaultsToOne(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	os.WriteFile(path, []byte(`{"version":0,"dir_map":{}}`), 0o644)

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Version != 1 {
		t.Errorf("got version %d, want 1 for zero-version config", cfg.Version)
	}
}
