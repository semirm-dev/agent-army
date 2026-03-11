package cli

import (
	"os"
	"path/filepath"

	"github.com/semir/agent-army/internal/pluginsync"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Install all plugins and skills listed in PLUGINS_AND_SKILLS.md",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := findRoot()
			docPath := filepath.Join(root, "PLUGINS_AND_SKILLS.md")
			return pluginsync.Run(docPath, os.Stdout, pluginsync.DefaultRunner{})
		},
	}
}
