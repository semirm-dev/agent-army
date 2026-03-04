package cli

import (
	"fmt"

	"github.com/semir/agent-army/internal/manifest"
	"github.com/spf13/cobra"
)

func newManifestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "manifest",
		Short: "Regenerate manifest.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := findRoot()
			if err := manifest.WriteManifest(root); err != nil {
				return err
			}
			fmt.Println("manifest.json regenerated.")
			return nil
		},
	}
}
