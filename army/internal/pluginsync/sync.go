package pluginsync

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

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

// userHomeDir is overridden in tests to use a temp directory.
var userHomeDir = os.UserHomeDir

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

	// Install plugins in parallel
	type result struct {
		cmdStr string
		err    error
	}

	pluginMatches := pluginCmdRe.FindAllStringSubmatch(content, -1)
	fmt.Fprintln(w, termcolor.Header("Installing Plugins", len(pluginMatches)))

	var pluginWg sync.WaitGroup
	pluginResults := make(chan result, len(pluginMatches))

	for _, m := range pluginMatches {
		pluginRef := m[1]
		args := []string{"plugin", "install", pluginRef}
		cmdStr := "claude " + strings.Join(args, " ")
		fmt.Fprintln(w, termcolor.Arrow(cmdStr))

		pluginWg.Add(1)
		go func(cmdStr string, args []string) {
			defer pluginWg.Done()
			err := runner.Run("claude", args)
			pluginResults <- result{cmdStr, err}
		}(cmdStr, args)
	}

	pluginWg.Wait()
	close(pluginResults)

	for r := range pluginResults {
		if r.err != nil {
			fmt.Fprintln(w, "  "+termcolor.Err("Failed: "+r.cmdStr))
			failures = append(failures, r.cmdStr)
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
			if err := removeSkillDirect(skillName); err != nil {
				fmt.Fprintln(w, "  "+termcolor.Err(fmt.Sprintf("Failed to remove %s: %v", skillName, err)))
				failures = append(failures, "remove "+skillName)
			} else {
				fmt.Fprintln(w, "  "+termcolor.Success("Removed standalone skill: "+skillName))
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

// removeSkillDirect removes a standalone skill by deleting its directory
// and removing its entry from .skill-lock.json. This bypasses `npx skills remove`
// which refuses to remove skills that have a pluginName field in the lock file.
func removeSkillDirect(skillName string) error {
	home, err := userHomeDir()
	if err != nil {
		return err
	}

	// Remove skill directory
	skillDir := filepath.Join(home, ".agents", "skills", skillName)
	if err := os.RemoveAll(skillDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing skill dir: %w", err)
	}

	// Remove entry from .skill-lock.json
	lockPath := filepath.Join(home, ".agents", ".skill-lock.json")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil // no lock file, nothing to clean
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
