package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generateAntigravityMD reads the GEMINI.md template for Antigravity and writes it to the output directory.
// GEMINI.md is written to the gemini-level directory (parent of the antigravity subdir for global installs),
// since Antigravity does not have its own separate GEMINI.md — it shares the one at the gemini level.
func generateAntigravityMD(root, dest, templatePath string) error {
	tmplBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read Antigravity GEMINI.md template: %w", err)
	}
	content := string(tmplBytes)

	geminiDir, _ := antigravityOutputPaths(dest)

	// Collapse 3+ consecutive newlines to 2 (one blank line).
	for strings.Contains(content, "\n\n\n") {
		content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
	}

	prefix := antigravityDestPrefix(geminiDir)
	content = strings.ReplaceAll(content, basePlaceholder, prefix)

	return writeOutput(geminiDir, "GEMINI.md", content)
}

// antigravityOutputPaths determines where GEMINI.md should be written and the artifact path prefix.
// For global installs (~/.gemini/antigravity), GEMINI.md goes to the parent (~/.gemini/) and artifacts
// are prefixed with "antigravity/" so paths resolve correctly relative to GEMINI.md.
// For local (.agents) and custom paths, GEMINI.md stays in dest with no artifact prefix.
func antigravityOutputPaths(dest string) (geminiDir string, artifactPrefix string) {
	home, err := os.UserHomeDir()
	if err == nil {
		globalAntigravity := filepath.Join(home, ".gemini", "antigravity")
		if filepath.Clean(dest) == filepath.Clean(globalAntigravity) {
			return filepath.Join(home, ".gemini"), "antigravity/"
		}
	}
	return dest, ""
}

// antigravityDestPrefix determines the display path prefix for where GEMINI.md is written.
// Since Antigravity shares GEMINI.md at the gemini level, the prefix reflects that location.
func antigravityDestPrefix(geminiDir string) string {
	home, err := os.UserHomeDir()
	if err == nil {
		globalGemini := filepath.Join(home, ".gemini")
		if filepath.Clean(geminiDir) == filepath.Clean(globalGemini) {
			return "~/.gemini"
		}
	}

	base := filepath.Base(geminiDir)
	if base == ".agents" {
		return ".agents"
	}

	return geminiDir
}

// AntigravityGeminiMDPath returns the path where GEMINI.md will be written for an Antigravity bootstrap.
// Exported for use by the bootstrap flow to check for existing files.
func AntigravityGeminiMDPath(dest string) string {
	geminiDir, _ := antigravityOutputPaths(dest)
	return filepath.Join(geminiDir, "GEMINI.md")
}
