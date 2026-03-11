package pluginsync

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/semir/agent-army/internal/termcolor"
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
	pluginMatches := pluginCmdRe.FindAllStringSubmatch(content, -1)
	fmt.Fprintln(w, termcolor.Header("Installing Plugins", len(pluginMatches)))
	for _, m := range pluginMatches {
		pluginRef := m[1]
		args := []string{"plugin", "install", pluginRef}
		cmdStr := "claude " + strings.Join(args, " ")
		fmt.Fprintln(w, termcolor.Arrow(cmdStr))
		if err := runner.Run("claude", args); err != nil {
			fmt.Fprintln(w, "  "+termcolor.Err("Failed: "+cmdStr))
			failures = append(failures, cmdStr)
		}
	}

	// Install skills (only from sections before "### Plugin-Provided Skills")
	skillContent := content
	if idx := strings.Index(content, "### Plugin-Provided Skills"); idx >= 0 {
		skillContent = content[:idx]
	}

	var skillCommands [][]string
	for _, m := range skillCmdRe.FindAllStringSubmatch(skillContent, -1) {
		if strings.Contains(m[1], "<") {
			continue
		}
		skillCommands = append(skillCommands, strings.Fields(m[1]))
	}

	fmt.Fprintln(w, termcolor.Header("Installing Skills", len(skillCommands)))
	for _, cmdParts := range skillCommands {
		cmdParts = append(cmdParts, "-y")
		cmdStr := strings.Join(cmdParts, " ")
		fmt.Fprintln(w, termcolor.Arrow(cmdStr))
		if err := runner.Run(cmdParts[0], cmdParts[1:]); err != nil {
			fmt.Fprintln(w, "  "+termcolor.Err("Failed: "+cmdStr))
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
		redundantMatches := redundantSkillRe.FindAllStringSubmatch(redundantSection, -1)
		fmt.Fprintln(w, termcolor.Header("Cleaning Up Redundant Skills", len(redundantMatches)))
		for _, m := range redundantMatches {
			skillName := m[1]
			cmdParts := []string{"skills", "remove", skillName, "-y"}
			cmdStr := "npx " + strings.Join(cmdParts, " ")
			fmt.Fprintln(w, termcolor.Arrow(cmdStr))
			if err := runner.Run("npx", cmdParts); err != nil {
				fmt.Fprintln(w, "  "+termcolor.Err("Failed: "+cmdStr))
				failures = append(failures, cmdStr)
			}
		}
	}

	if len(failures) > 0 {
		fmt.Fprint(w, termcolor.ErrMsg("Some commands failed. Check output above."))
		return fmt.Errorf("%d commands failed", len(failures))
	}

	fmt.Fprint(w, termcolor.DoneMsg("All plugins and skills installed."))
	return nil
}
