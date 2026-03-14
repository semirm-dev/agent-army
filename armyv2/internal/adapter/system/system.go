package system

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/smahovkic/agent-army/armyv2/internal/core/types"
)

// userHomeDir is a package-level var so tests can override it.
var userHomeDir = os.UserHomeDir

// installedPluginsFile mirrors the JSON structure of installed_plugins.json.
type installedPluginsFile struct {
	Version int                        `json:"version"`
	Plugins map[string][]pluginInstance `json:"plugins"`
}

// pluginInstance represents a single install entry in the plugins map.
type pluginInstance struct {
	Scope       string `json:"scope"`
	InstallPath string `json:"installPath"`
	Version     string `json:"version"`
	InstalledAt string `json:"installedAt"`
	LastUpdated string `json:"lastUpdated"`
	GitCommitSha string `json:"gitCommitSha"`
}

// skillLockFile mirrors the JSON structure of .skill-lock.json.
type skillLockFile struct {
	Version int                    `json:"version"`
	Skills  map[string]skillEntry  `json:"skills"`
}

// skillEntry represents a single skill in the lock file.
type skillEntry struct {
	Source          string `json:"source"`
	SourceType      string `json:"sourceType"`
	SourceURL       string `json:"sourceUrl"`
	SkillPath       string `json:"skillPath"`
	SkillFolderHash string `json:"skillFolderHash"`
	InstalledAt     string `json:"installedAt"`
	UpdatedAt       string `json:"updatedAt"`
}

// Reader reads installed plugin and skill state from the filesystem.
type Reader struct{}

// New creates a system Reader.
func New() *Reader {
	return &Reader{}
}

// InstalledPlugins reads ~/.claude/plugins/installed_plugins.json and returns
// a list of installed plugins. Returns an empty slice if the file doesn't exist.
func (r *Reader) InstalledPlugins() ([]types.InstalledPlugin, error) {
	home, err := userHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home dir: %w", err)
	}

	path := filepath.Join(home, ".claude", "plugins", "installed_plugins.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []types.InstalledPlugin{}, nil
		}
		return nil, fmt.Errorf("reading installed_plugins.json: %w", err)
	}

	var file installedPluginsFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parsing installed_plugins.json: %w", err)
	}

	if file.Plugins == nil {
		return []types.InstalledPlugin{}, nil
	}

	var plugins []types.InstalledPlugin
	for key, instances := range file.Plugins {
		if len(instances) == 0 {
			continue
		}

		name, marketplace := parsePluginKey(key)
		inst := instances[0]

		plugins = append(plugins, types.InstalledPlugin{
			Name:        name,
			Marketplace: marketplace,
			Version:     inst.Version,
			Scope:       inst.Scope,
			InstallPath: inst.InstallPath,
		})
	}

	return plugins, nil
}

// InstalledSkills reads ~/.agents/.skill-lock.json and returns a list of
// installed skills. Returns an empty slice if the file doesn't exist.
func (r *Reader) InstalledSkills() ([]types.InstalledSkill, error) {
	home, err := userHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home dir: %w", err)
	}

	path := filepath.Join(home, ".agents", ".skill-lock.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []types.InstalledSkill{}, nil
		}
		return nil, fmt.Errorf("reading .skill-lock.json: %w", err)
	}

	var file skillLockFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parsing .skill-lock.json: %w", err)
	}

	if file.Skills == nil {
		return []types.InstalledSkill{}, nil
	}

	var skills []types.InstalledSkill
	for name, entry := range file.Skills {
		skills = append(skills, types.InstalledSkill{
			Name:      name,
			Source:    entry.Source,
			SourceURL: entry.SourceURL,
		})
	}

	return skills, nil
}

// parsePluginKey splits a "name@marketplace" key into its components.
// If there's no "@", marketplace is empty.
func parsePluginKey(key string) (name, marketplace string) {
	parts := strings.SplitN(key, "@", 2)
	name = parts[0]
	if len(parts) == 2 {
		marketplace = parts[1]
	}
	return
}
