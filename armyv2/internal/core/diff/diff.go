package diff

import (
	"github.com/smahovkic/agent-army/armyv2/internal/core/types"
)

// Compare compares the manifest against the installed state and returns
// a structured diff result showing missing and extra items.
func Compare(manifest *types.Manifest, plugins []types.InstalledPlugin, skills []types.InstalledSkill) types.DiffResult {
	result := types.DiffResult{}

	installedPlugins := make(map[string]types.InstalledPlugin)
	for _, p := range plugins {
		installedPlugins[p.Name] = p
	}

	installedSkills := make(map[string]types.InstalledSkill)
	for _, s := range skills {
		installedSkills[s.Name] = s
	}

	// Find plugins in manifest but not installed
	manifestPlugins := make(map[string]bool)
	for _, mp := range manifest.Plugins {
		manifestPlugins[mp.Name] = true
		if _, found := installedPlugins[mp.Name]; !found {
			result.MissingPlugins = append(result.MissingPlugins, mp)
		}
	}

	// Find plugins installed but not in manifest
	for _, ip := range plugins {
		if !manifestPlugins[ip.Name] {
			result.ExtraPlugins = append(result.ExtraPlugins, ip)
		}
	}

	// Find skills in manifest but not installed
	manifestSkills := make(map[string]bool)
	for _, ms := range manifest.Skills {
		manifestSkills[ms.Name] = true
		if _, found := installedSkills[ms.Name]; !found {
			result.MissingSkills = append(result.MissingSkills, ms)
		}
	}

	// Find skills installed but not in manifest
	for _, is := range skills {
		if !manifestSkills[is.Name] {
			result.ExtraSkills = append(result.ExtraSkills, is)
		}
	}

	return result
}

// HasDrift returns true if there are any differences between manifest and installed state.
func HasDrift(d types.DiffResult) bool {
	return len(d.MissingPlugins) > 0 ||
		len(d.ExtraPlugins) > 0 ||
		len(d.MissingSkills) > 0 ||
		len(d.ExtraSkills) > 0
}
