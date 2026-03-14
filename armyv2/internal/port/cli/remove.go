package cli

import (
	"fmt"

	"github.com/smahovkic/agent-army/armyv2/internal/core/manifest"
	"github.com/spf13/cobra"
)

func newRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a plugin or skill from the manifest",
	}

	cmd.AddCommand(newRemovePluginCmd(), newRemoveSkillCmd())
	return cmd
}

func newRemovePluginCmd() *cobra.Command {
	var manifestOnly bool

	cmd := &cobra.Command{
		Use:   "plugin <name>",
		Short: "Remove a plugin from the manifest and uninstall it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			d, err := resolveDeps()
			if err != nil {
				return err
			}

			if !manifest.RemovePlugin(d.manifest, name) {
				return fmt.Errorf("plugin %q not found in manifest", name)
			}

			if err := manifest.Save(d.manifestPath, d.manifest); err != nil {
				return fmt.Errorf("saving manifest: %w", err)
			}
			fmt.Printf("Removed plugin %q from manifest.\n", name)

			if !manifestOnly {
				// Sync will handle the removal
				actions, err := d.orchestrator.PlanActions(d.manifest)
				if err != nil {
					return fmt.Errorf("planning actions: %w", err)
				}
				if len(actions) > 0 {
					result := d.orchestrator.Execute(actions)
					if result.Failed > 0 {
						return fmt.Errorf("failed to uninstall plugin %q", name)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&manifestOnly, "manifest-only", false, "Remove from manifest without uninstalling")
	return cmd
}

func newRemoveSkillCmd() *cobra.Command {
	var manifestOnly bool

	cmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Remove a skill from the manifest and uninstall it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			d, err := resolveDeps()
			if err != nil {
				return err
			}

			if !manifest.RemoveSkill(d.manifest, name) {
				return fmt.Errorf("skill %q not found in manifest", name)
			}

			if err := manifest.Save(d.manifestPath, d.manifest); err != nil {
				return fmt.Errorf("saving manifest: %w", err)
			}
			fmt.Printf("Removed skill %q from manifest.\n", name)

			if !manifestOnly {
				actions, err := d.orchestrator.PlanActions(d.manifest)
				if err != nil {
					return fmt.Errorf("planning actions: %w", err)
				}
				if len(actions) > 0 {
					result := d.orchestrator.Execute(actions)
					if result.Failed > 0 {
						return fmt.Errorf("failed to uninstall skill %q", name)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&manifestOnly, "manifest-only", false, "Remove from manifest without uninstalling")
	return cmd
}
