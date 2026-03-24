package plugin

import "fmt"

// CommandRunner executes shell commands.
type CommandRunner interface {
	Run(cmd string, args ...string) (stdout string, err error)
}

// Adapter handles plugin install/remove operations via the claude CLI.
type Adapter struct {
	runner CommandRunner
}

// New creates a plugin Adapter with the given command runner.
func New(r CommandRunner) *Adapter {
	return &Adapter{runner: r}
}

// Install runs: claude plugin install <name>
func (a *Adapter) Install(name string) error {
	_, err := a.runner.Run("claude", "plugin", "install", name)
	if err != nil {
		return fmt.Errorf("installing plugin %s: %w", name, err)
	}
	return nil
}

// Remove runs: claude plugin remove <name>
func (a *Adapter) Remove(name string) error {
	_, err := a.runner.Run("claude", "plugin", "remove", name)
	if err != nil {
		return fmt.Errorf("removing plugin %s: %w", name, err)
	}
	return nil
}
