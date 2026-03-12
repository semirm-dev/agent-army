package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generateAgentsMD reads the AGENTS.md template from templatePath and replaces
// the {{BASE}} placeholder with the correct path prefix for the destination.
func generateAgentsMD(dest, templatePath string) error {
	tmplBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read AGENTS.md template: %w", err)
	}
	content := string(tmplBytes)

	prefix := cursorDestPrefix(dest)
	content = strings.ReplaceAll(content, basePlaceholder, prefix)

	return writeOutput(dest, "AGENTS.md", content)
}

// cursorDestPrefix determines the display path prefix for the Cursor destination directory.
func cursorDestPrefix(dest string) string {
	home, err := os.UserHomeDir()
	if err == nil {
		globalCursor := filepath.Join(home, ".cursor")
		if filepath.Clean(dest) == filepath.Clean(globalCursor) {
			return "~/.cursor"
		}
	}

	base := filepath.Base(dest)
	if base == ".cursor" {
		return ".cursor"
	}

	return dest
}
