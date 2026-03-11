package cli

import (
	"fmt"
	"path/filepath"

	"github.com/semir/agent-army/internal/plugindoc"
	"github.com/spf13/cobra"
)

func newUpdatePluginsSkillsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update-plugins-skills",
		Short: "Regenerate PLUGINS_AND_SKILLS.md from system state",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := findRoot()
			out := filepath.Join(root, "PLUGINS_AND_SKILLS.md")
			if err := plugindoc.WritePluginsAndSkills(out); err != nil {
				return err
			}
			fmt.Printf("Generated %s\n", out)
			return nil
		},
	}
}
