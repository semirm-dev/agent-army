package cli

import (
	"os"

	"github.com/semir/agent-army/internal/bootstrap"
	"github.com/semir/agent-army/internal/tui"
	"github.com/spf13/cobra"
)

func newBootstrapCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bootstrap",
		Short: "Generate model-specific rules, skills, and agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := findRoot()
			p := tui.NewStdinPrompter(os.Stdin, os.Stdout)
			return bootstrap.MainBootstrap(root, p, os.Stdout)
		},
	}
}
