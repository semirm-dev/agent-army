package cli

import (
	"fmt"

	"github.com/smahovkic/agent-army/armyv2/internal/core/manifest"
	"github.com/smahovkic/agent-army/armyv2/internal/core/types"
	"github.com/spf13/cobra"
)

func newAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a plugin or skill to the manifest",
	}

	cmd.AddCommand(newAddPluginCmd(), newAddSkillCmd())
	return cmd
}

func newAddPluginCmd() *cobra.Command {
	var noInstall bool
	var project bool

	cmd := &cobra.Command{
		Use:   "plugin <name>",
		Short: "Add a plugin to the manifest and install it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			d, err := resolveDeps()
			if err != nil {
				return err
			}

			cp, found := d.catalog.FindPlugin(name)
			if !found {
				return fmt.Errorf("plugin %q not found in catalog. Run 'armyv2 update' to refresh", name)
			}

			dest := "user"
			if project {
				dest = "project"
			}

			mp := types.ManifestPlugin{
				Name:        cp.Name,
				Marketplace: cp.Marketplace,
				Tags:        cp.Tags,
				Destination: dest,
			}

			if !manifest.AddPlugin(d.manifest, mp) {
				fmt.Printf("Plugin %q is already in the manifest.\n", name)
				return nil
			}

			if err := manifest.Save(d.manifestPath, d.manifest); err != nil {
				return fmt.Errorf("saving manifest: %w", err)
			}
			fmt.Printf("Added plugin %q to manifest.\n", name)

			if !noInstall {
				fmt.Printf("Installing %s...\n", name)
				result := d.orchestrator.InstallItems([]types.ManifestPlugin{mp}, nil)
				if result.Failed > 0 {
					return fmt.Errorf("failed to install plugin %q", name)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&noInstall, "no-install", false, "Add to manifest without installing")
	cmd.Flags().BoolVar(&project, "project", false, "Set destination to project scope")
	return cmd
}

func newAddSkillCmd() *cobra.Command {
	var noInstall bool
	var project bool

	cmd := &cobra.Command{
		Use:   "skill <name>",
		Short: "Add a skill to the manifest and install it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			d, err := resolveDeps()
			if err != nil {
				return err
			}

			cs, found := d.catalog.FindSkill(name)
			if !found {
				return fmt.Errorf("skill %q not found in catalog. Run 'armyv2 update' to refresh", name)
			}

			dest := "user"
			if project {
				dest = "project"
			}

			ms := types.ManifestSkill{
				Name:        cs.Name,
				Source:      cs.Source,
				Tags:        cs.Tags,
				Destination: dest,
			}

			if !manifest.AddSkill(d.manifest, ms) {
				fmt.Printf("Skill %q is already in the manifest.\n", name)
				return nil
			}

			if err := manifest.Save(d.manifestPath, d.manifest); err != nil {
				return fmt.Errorf("saving manifest: %w", err)
			}
			fmt.Printf("Added skill %q to manifest.\n", name)

			if !noInstall {
				fmt.Printf("Installing %s...\n", name)
				result := d.orchestrator.InstallItems(nil, []types.ManifestSkill{ms})
				if result.Failed > 0 {
					return fmt.Errorf("failed to install skill %q", name)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&noInstall, "no-install", false, "Add to manifest without installing")
	cmd.Flags().BoolVar(&project, "project", false, "Set destination to project scope")
	return cmd
}
