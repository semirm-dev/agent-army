package doctor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/smahovkic/agent-army/armyv2/internal/core/types"
)

// Check runs all health checks and returns a list of issues found.
func Check(manifest *types.Manifest, plugins []types.InstalledPlugin, skills []types.InstalledSkill) []types.DoctorIssue {
	var issues []types.DoctorIssue

	issues = append(issues, checkMissingPlugins(manifest, plugins)...)
	issues = append(issues, checkMissingSkills(manifest, skills)...)
	issues = append(issues, checkOrphanPlugins(manifest, plugins)...)
	issues = append(issues, checkOrphanSkills(manifest, skills)...)
	issues = append(issues, checkSkillDiskDrift(skills)...)
	issues = append(issues, checkSkillOrphanDirs(skills)...)
	issues = append(issues, checkPluginDiskDrift(plugins)...)

	return issues
}

// checkMissingPlugins finds plugins in manifest but not installed.
func checkMissingPlugins(manifest *types.Manifest, installed []types.InstalledPlugin) []types.DoctorIssue {
	set := make(map[string]bool)
	for _, p := range installed {
		set[p.Name] = true
	}

	var issues []types.DoctorIssue
	for _, mp := range manifest.Plugins {
		if !set[mp.Name] {
			issues = append(issues, types.DoctorIssue{
				Severity:    "error",
				Category:    "missing",
				Description: fmt.Sprintf("Plugin %q is in manifest but not installed", mp.Name),
				Item:        mp.Name,
			})
		}
	}
	return issues
}

// checkMissingSkills finds skills in manifest but not installed.
func checkMissingSkills(manifest *types.Manifest, installed []types.InstalledSkill) []types.DoctorIssue {
	set := make(map[string]bool)
	for _, s := range installed {
		set[s.Name] = true
	}

	var issues []types.DoctorIssue
	for _, ms := range manifest.Skills {
		if !set[ms.Name] {
			issues = append(issues, types.DoctorIssue{
				Severity:    "error",
				Category:    "missing",
				Description: fmt.Sprintf("Skill %q is in manifest but not installed", ms.Name),
				Item:        ms.Name,
			})
		}
	}
	return issues
}

// checkOrphanPlugins finds plugins installed but not in manifest.
func checkOrphanPlugins(manifest *types.Manifest, installed []types.InstalledPlugin) []types.DoctorIssue {
	set := make(map[string]bool)
	for _, mp := range manifest.Plugins {
		set[mp.Name] = true
	}

	var issues []types.DoctorIssue
	for _, ip := range installed {
		if !set[ip.Name] {
			issues = append(issues, types.DoctorIssue{
				Severity:    "warning",
				Category:    "orphan",
				Description: fmt.Sprintf("Plugin %q is installed but not in manifest", ip.Name),
				Item:        ip.Name,
			})
		}
	}
	return issues
}

// checkOrphanSkills finds skills installed but not in manifest.
func checkOrphanSkills(manifest *types.Manifest, installed []types.InstalledSkill) []types.DoctorIssue {
	set := make(map[string]bool)
	for _, ms := range manifest.Skills {
		set[ms.Name] = true
	}

	var issues []types.DoctorIssue
	for _, is := range installed {
		if !set[is.Name] {
			issues = append(issues, types.DoctorIssue{
				Severity:    "warning",
				Category:    "orphan",
				Description: fmt.Sprintf("Skill %q is installed but not in manifest", is.Name),
				Item:        is.Name,
			})
		}
	}
	return issues
}

// checkSkillOrphanDirs finds skill directories on disk that are not in the lock file.
func checkSkillOrphanDirs(installed []types.InstalledSkill) []types.DoctorIssue {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	skillsDir := filepath.Join(home, ".agents", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil
	}

	known := make(map[string]bool)
	for _, s := range installed {
		known[s.Name] = true
	}

	var issues []types.DoctorIssue
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if !known[e.Name()] {
			issues = append(issues, types.DoctorIssue{
				Severity:    "warning",
				Category:    "drift",
				Description: fmt.Sprintf("Skill directory %q exists on disk at %s but is not in lock file", e.Name(), filepath.Join(skillsDir, e.Name())),
				Item:        e.Name(),
			})
		}
	}
	return issues
}

// checkPluginDiskDrift finds plugins whose installPath no longer exists on disk.
func checkPluginDiskDrift(installed []types.InstalledPlugin) []types.DoctorIssue {
	var issues []types.DoctorIssue
	for _, p := range installed {
		if p.InstallPath == "" {
			continue
		}
		if _, err := os.Stat(p.InstallPath); os.IsNotExist(err) {
			issues = append(issues, types.DoctorIssue{
				Severity:    "warning",
				Category:    "drift",
				Description: fmt.Sprintf("Plugin %q has installPath %s but directory is missing from disk", p.Name, p.InstallPath),
				Item:        p.Name,
			})
		}
	}
	return issues
}

// checkSkillDiskDrift finds skills in the lock file that are missing from disk.
func checkSkillDiskDrift(installed []types.InstalledSkill) []types.DoctorIssue {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var issues []types.DoctorIssue
	for _, s := range installed {
		skillDir := filepath.Join(home, ".agents", "skills", s.Name)
		if _, err := os.Stat(skillDir); os.IsNotExist(err) {
			issues = append(issues, types.DoctorIssue{
				Severity:    "error",
				Category:    "drift",
				Description: fmt.Sprintf("Skill %q is in lock file but missing from disk at %s", s.Name, skillDir),
				Item:        s.Name,
			})
		}
	}
	return issues
}
