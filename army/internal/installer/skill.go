package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// userHomeDir is a package-level var so tests can override it.
var userHomeDir = os.UserHomeDir

// Skill handles skill install/remove operations.
type Skill struct {
	runner CommandRunner
}

// NewSkill creates a Skill with the given command runner.
func NewSkill(r CommandRunner) *Skill {
	return &Skill{runner: r}
}

// Install runs: npx skills add <source> -s <name> -g -y
func (a *Skill) Install(name, source string) error {
	_, err := a.runner.Run("npx", "skills", "add", source, "-s", name, "-g", "-y")
	if err != nil {
		return fmt.Errorf("installing skill %s from %s: %w", name, source, err)
	}
	return nil
}

// Remove does direct filesystem removal of a skill. This bypasses
// "npx skills remove" which refuses to remove plugin-provided skills.
//
// Steps:
//  1. Delete ~/.agents/skills/<name>/ directory
//  2. Delete ~/.claude/skills/<name>/ symlink (if it exists)
//  3. Remove entry from ~/.agents/.skill-lock.json
func (a *Skill) Remove(name string) error {
	home, err := userHomeDir()
	if err != nil {
		return fmt.Errorf("getting home dir: %w", err)
	}

	// 1. Remove skill directory
	skillDir := filepath.Join(home, ".agents", "skills", name)
	if err := os.RemoveAll(skillDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing skill dir %s: %w", skillDir, err)
	}

	// 2. Remove symlink at ~/.claude/skills/<name> if it exists
	symlinkPath := filepath.Join(home, ".claude", "skills", name)
	if err := os.RemoveAll(symlinkPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing skill symlink %s: %w", symlinkPath, err)
	}

	// 3. Remove entry from .skill-lock.json
	if err := removeFromLockFile(home, name); err != nil {
		return fmt.Errorf("updating skill-lock for %s: %w", name, err)
	}

	return nil
}

// removeFromLockFile reads ~/.agents/.skill-lock.json, removes the named
// skill entry, and writes it back. If the lock file doesn't exist or the
// skill isn't in it, this is a no-op.
func removeFromLockFile(home, skillName string) error {
	lockPath := filepath.Join(home, ".agents", ".skill-lock.json")

	data, err := os.ReadFile(lockPath)
	if err != nil {
		// No lock file means nothing to clean up.
		return nil
	}

	var lock map[string]interface{}
	if err := json.Unmarshal(data, &lock); err != nil {
		return fmt.Errorf("parsing skill-lock: %w", err)
	}

	skills, ok := lock["skills"].(map[string]interface{})
	if !ok {
		return nil
	}

	if _, exists := skills[skillName]; !exists {
		return nil
	}
	delete(skills, skillName)

	out, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling skill-lock: %w", err)
	}
	return os.WriteFile(lockPath, append(out, '\n'), 0644)
}
