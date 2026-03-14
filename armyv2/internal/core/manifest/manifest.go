package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/smahovkic/agent-army/armyv2/internal/core/types"
)

// DefaultPath returns the default manifest path: ~/.armyv2/manifest.json.
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, ".armyv2", "manifest.json"), nil
}

// Load reads a manifest from the given path. If the file does not exist,
// it returns an empty manifest with version 1.
func Load(path string) (*types.Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return emptyManifest(), nil
		}
		return nil, fmt.Errorf("reading manifest %s: %w", path, err)
	}

	var m types.Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest %s: %w", path, err)
	}

	// Ensure non-nil slices.
	if m.Plugins == nil {
		m.Plugins = []types.ManifestPlugin{}
	}
	if m.Skills == nil {
		m.Skills = []types.ManifestSkill{}
	}
	if m.Version == 0 {
		m.Version = 1
	}

	return &m, nil
}

// Save writes the manifest to the given path using an atomic write pattern
// (write to temp file, then rename). Creates the parent directory if needed.
func Save(path string, m *types.Manifest) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling manifest: %w", err)
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(dir, "manifest-*.json.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	// Clean up the temp file on any failure path.
	success := false
	defer func() {
		if !success {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("renaming temp file to %s: %w", path, err)
	}

	success = true
	return nil
}

// AddPlugin adds a plugin to the manifest. Returns false if a plugin with
// the same name already exists (case-insensitive comparison).
func AddPlugin(m *types.Manifest, p types.ManifestPlugin) bool {
	if HasPlugin(m, p.Name) {
		return false
	}
	m.Plugins = append(m.Plugins, p)
	return true
}

// RemovePlugin removes a plugin from the manifest by name (case-insensitive).
// Returns false if the plugin was not found.
func RemovePlugin(m *types.Manifest, name string) bool {
	lower := strings.ToLower(name)
	for i, p := range m.Plugins {
		if strings.ToLower(p.Name) == lower {
			m.Plugins = append(m.Plugins[:i], m.Plugins[i+1:]...)
			return true
		}
	}
	return false
}

// AddSkill adds a skill to the manifest. Returns false if a skill with
// the same name already exists (case-insensitive comparison).
func AddSkill(m *types.Manifest, s types.ManifestSkill) bool {
	if HasSkill(m, s.Name) {
		return false
	}
	m.Skills = append(m.Skills, s)
	return true
}

// RemoveSkill removes a skill from the manifest by name (case-insensitive).
// Returns false if the skill was not found.
func RemoveSkill(m *types.Manifest, name string) bool {
	lower := strings.ToLower(name)
	for i, sk := range m.Skills {
		if strings.ToLower(sk.Name) == lower {
			m.Skills = append(m.Skills[:i], m.Skills[i+1:]...)
			return true
		}
	}
	return false
}

// HasPlugin checks if a plugin with the given name exists in the manifest
// (case-insensitive comparison).
func HasPlugin(m *types.Manifest, name string) bool {
	lower := strings.ToLower(name)
	for _, p := range m.Plugins {
		if strings.ToLower(p.Name) == lower {
			return true
		}
	}
	return false
}

// HasSkill checks if a skill with the given name exists in the manifest
// (case-insensitive comparison).
func HasSkill(m *types.Manifest, name string) bool {
	lower := strings.ToLower(name)
	for _, sk := range m.Skills {
		if strings.ToLower(sk.Name) == lower {
			return true
		}
	}
	return false
}

// emptyManifest returns a new empty manifest with version 1.
func emptyManifest() *types.Manifest {
	return &types.Manifest{
		Version: 1,
		Plugins: []types.ManifestPlugin{},
		Skills:  []types.ManifestSkill{},
	}
}
