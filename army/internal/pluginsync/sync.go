package pluginsync

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// CommandRunner abstracts command execution for testability.
type CommandRunner interface {
	Run(name string, args []string) error
}

// DefaultRunner executes real commands.
type DefaultRunner struct{}

func (DefaultRunner) Run(name string, args []string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = nil // prevent interactive prompts from consuming input
	return cmd.Run()
}

var pluginCmdRe = regexp.MustCompile(`/plugin install ([^\s|` + "`" + `]+)`)
var skillCmdRe = regexp.MustCompile("`(npx skills add [^`]+)`")
var redundantSkillRe = regexp.MustCompile("`npx skills remove ([^`]+)`")

// Run reads the doc file, extracts install commands, and executes them.
func Run(docPath string, w io.Writer, runner CommandRunner) error {
	data, err := os.ReadFile(docPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", docPath, err)
	}
	content := string(data)

	var failures []string

	// Install plugins
	fmt.Fprintln(w, "=== Installing Plugins ===")
	for _, m := range pluginCmdRe.FindAllStringSubmatch(content, -1) {
		pluginRef := m[1]
		args := []string{"plugin", "install", pluginRef}
		cmdStr := "claude " + strings.Join(args, " ")
		fmt.Fprintf(w, "\u2192 %s\n", cmdStr)
		if err := runner.Run("claude", args); err != nil {
			fmt.Fprintf(w, "  \u2717 Failed: %s\n", cmdStr)
			failures = append(failures, cmdStr)
		}
	}

	// Install skills (only from sections before "### Plugin-Provided Skills")
	fmt.Fprintln(w, "\n=== Installing Skills ===")
	skillContent := content
	if idx := strings.Index(content, "### Plugin-Provided Skills"); idx >= 0 {
		skillContent = content[:idx]
	}

	for _, m := range skillCmdRe.FindAllStringSubmatch(skillContent, -1) {
		cmdParts := strings.Fields(m[1])
		// Skip commands containing < (template placeholders)
		if strings.Contains(m[1], "<") {
			continue
		}
		// Append -y flag for non-interactive
		cmdParts = append(cmdParts, "-y")
		cmdStr := strings.Join(cmdParts, " ")
		fmt.Fprintf(w, "\u2192 %s\n", cmdStr)
		if err := runner.Run(cmdParts[0], cmdParts[1:]); err != nil {
			fmt.Fprintf(w, "  \u2717 Failed: %s\n", cmdStr)
			failures = append(failures, cmdStr)
		}
	}

	// Cleanup redundant standalone skills
	redundantSection := ""
	if idx := strings.Index(content, "> **Redundant standalone skills:**"); idx >= 0 {
		// Extract until the next blank line (end of blockquote)
		rest := content[idx:]
		if endIdx := strings.Index(rest, "\n\n"); endIdx >= 0 {
			redundantSection = rest[:endIdx]
		} else {
			redundantSection = rest
		}
	}

	if redundantSection != "" {
		fmt.Fprintln(w, "\n=== Cleaning Up Redundant Skills ===")
		for _, m := range redundantSkillRe.FindAllStringSubmatch(redundantSection, -1) {
			skillName := m[1]
			cmdParts := []string{"skills", "remove", skillName, "-y"}
			cmdStr := "npx " + strings.Join(cmdParts, " ")
			fmt.Fprintf(w, "\u2192 %s\n", cmdStr)
			if err := runner.Run("npx", cmdParts); err != nil {
				fmt.Fprintf(w, "  \u2717 Failed: %s\n", cmdStr)
				failures = append(failures, cmdStr)
			}
		}
	}

	if len(failures) > 0 {
		fmt.Fprintf(w, "\nSome commands failed. Check output above.\n")
		return fmt.Errorf("%d commands failed", len(failures))
	}

	fmt.Fprintln(w, "\nDone. All plugins and skills installed.")
	return nil
}
