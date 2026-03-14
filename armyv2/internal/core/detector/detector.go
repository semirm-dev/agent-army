package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/smahovkic/agent-army/armyv2/internal/core/types"
)

// Detect scans the given directory for tech stack markers and returns
// the names of matched tech profiles from the catalog.
func Detect(dir string, profiles map[string]types.TechProfile) []string {
	var matched []string
	for name, profile := range profiles {
		if matchesProfile(dir, profile) {
			matched = append(matched, name)
		}
	}
	return matched
}

// RecommendedItems returns deduplicated plugin and skill names
// recommended for the given set of tech profile names.
func RecommendedItems(profileNames []string, profiles map[string]types.TechProfile) (plugins []string, skills []string) {
	pluginSet := make(map[string]bool)
	skillSet := make(map[string]bool)

	for _, name := range profileNames {
		p, ok := profiles[name]
		if !ok {
			continue
		}
		for _, pl := range p.Plugins {
			if !pluginSet[pl] {
				pluginSet[pl] = true
				plugins = append(plugins, pl)
			}
		}
		for _, sk := range p.Skills {
			if !skillSet[sk] {
				skillSet[sk] = true
				skills = append(skills, sk)
			}
		}
	}
	return plugins, skills
}

func matchesProfile(dir string, profile types.TechProfile) bool {
	for _, pattern := range profile.Detect {
		if matchesPattern(dir, pattern) {
			return true
		}
	}
	return false
}

func matchesPattern(dir, pattern string) bool {
	// Check for content match pattern: "file:content"
	if idx := strings.Index(pattern, ":"); idx > 0 {
		file := pattern[:idx]
		content := pattern[idx+1:]
		return matchesContentPattern(dir, file, content)
	}
	// Simple file existence / glob pattern
	return matchesFilePattern(dir, pattern)
}

func matchesFilePattern(dir, pattern string) bool {
	full := filepath.Join(dir, pattern)
	matches, err := filepath.Glob(full)
	if err != nil {
		return false
	}
	return len(matches) > 0
}

func matchesContentPattern(dir, filePattern, content string) bool {
	full := filepath.Join(dir, filePattern)
	matches, err := filepath.Glob(full)
	if err != nil || len(matches) == 0 {
		return false
	}

	for _, path := range matches {
		if fileContains(path, content) {
			return true
		}
	}
	return false
}

func fileContains(path, content string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	// For JSON files (package.json, composer.json), check dependency keys
	ext := filepath.Ext(path)
	base := filepath.Base(path)

	if ext == ".json" {
		switch base {
		case "package.json":
			return jsonHasDependency(data, content)
		case "composer.json":
			return jsonHasComposerDependency(data, content)
		}
	}

	// Fallback: substring match
	return strings.Contains(string(data), content)
}

// jsonHasDependency checks if a package.json has a dependency (or devDependency)
// matching the given key exactly.
func jsonHasDependency(data []byte, key string) bool {
	var pkg struct {
		Dependencies    map[string]interface{} `json:"dependencies"`
		DevDependencies map[string]interface{} `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		// Fallback to substring match on parse failure
		return strings.Contains(string(data), key)
	}

	if _, ok := pkg.Dependencies[key]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies[key]; ok {
		return true
	}
	return false
}

// jsonHasComposerDependency checks if a composer.json has a require entry
// matching the given key exactly.
func jsonHasComposerDependency(data []byte, key string) bool {
	var composer struct {
		Require    map[string]interface{} `json:"require"`
		RequireDev map[string]interface{} `json:"require-dev"`
	}
	if err := json.Unmarshal(data, &composer); err != nil {
		return strings.Contains(string(data), key)
	}

	if _, ok := composer.Require[key]; ok {
		return true
	}
	if _, ok := composer.RequireDev[key]; ok {
		return true
	}
	return false
}
