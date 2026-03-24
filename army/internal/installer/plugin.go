package installer

import "fmt"

// CommandRunner executes shell commands.
type CommandRunner interface {
	Run(cmd string, args ...string) (stdout string, err error)
}

// Plugin handles plugin install/remove operations via the claude CLI.
type Plugin struct {
	runner CommandRunner
}

// NewPlugin creates a Plugin with the given command runner.
func NewPlugin(r CommandRunner) *Plugin {
	return &Plugin{runner: r}
}

// Install runs: claude plugin install <name>
func (a *Plugin) Install(name string) error {
	_, err := a.runner.Run("claude", "plugin", "install", name)
	if err != nil {
		return fmt.Errorf("installing plugin %s: %w", name, err)
	}
	return nil
}

// Remove runs: claude plugin remove <name>
func (a *Plugin) Remove(name string) error {
	_, err := a.runner.Run("claude", "plugin", "remove", name)
	if err != nil {
		return fmt.Errorf("removing plugin %s: %w", name, err)
	}
	return nil
}
