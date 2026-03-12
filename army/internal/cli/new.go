package cli

import (
	"os"

	"github.com/semir/agent-army/internal/scaffold"
	"github.com/semir/agent-army/internal/tui"
	"github.com/spf13/cobra"
)

func newNewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Scaffold a new skill or agent",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "skill",
		Short: "Create a new skill",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := findRoot()
			p := tui.NewStdinPrompter(os.Stdin, os.Stdout)
			return scaffold.ScaffoldFlow(root, "skill", p, os.Stdout)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "agent",
		Short: "Create a new agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := findRoot()
			p := tui.NewStdinPrompter(os.Stdin, os.Stdout)
			return scaffold.ScaffoldFlow(root, "agent", p, os.Stdout)
		},
	})

	return cmd
}
