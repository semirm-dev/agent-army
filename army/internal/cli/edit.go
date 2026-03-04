package cli

import (
	"os"

	"github.com/semir/agent-army/internal/editor"
	"github.com/semir/agent-army/internal/tui"
	"github.com/spf13/cobra"
)

func newEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Interactive dependency editor",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := findRoot()
			p := tui.NewStdinPrompter(os.Stdin, os.Stdout)
			return editor.EditFlow(root, p, os.Stdout)
		},
	}
}
