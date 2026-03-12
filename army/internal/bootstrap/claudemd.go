package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const basePlaceholder = "{{BASE}}"

// generateClaudeMD reads the CLAUDE.md template from templatePath and replaces
// the {{BASE}} placeholder with the correct path prefix for the destination.
func generateClaudeMD(dest, templatePath string) error {
	tmplBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read CLAUDE.md template: %w", err)
	}
	content := string(tmplBytes)

	prefix := destPrefix(dest)
	content = strings.ReplaceAll(content, basePlaceholder, prefix)

	return writeOutput(dest, "CLAUDE.md", content)
}

// destPrefix determines the display path prefix for the destination directory.
// Global (~/.claude) uses "~/.claude", project-local uses ".claude", custom uses the path as-is.
func destPrefix(dest string) string {
	home, err := os.UserHomeDir()
	if err == nil {
		globalClaude := filepath.Join(home, ".claude")
		if filepath.Clean(dest) == filepath.Clean(globalClaude) {
			return "~/.claude"
		}
	}

	base := filepath.Base(dest)
	if base == ".claude" {
		return ".claude"
	}

	return dest
}
