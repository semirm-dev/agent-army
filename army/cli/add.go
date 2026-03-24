package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/smahovkic/agent-army/army/internal/core/manifest"
	"github.com/smahovkic/agent-army/army/internal/core/types"
	"github.com/spf13/cobra"
)

type addResult struct {
	Action           string `json:"action"`
	ItemType         string `json:"item_type"`
	Name             string `json:"name"`
	AddedToManifest  bool   `json:"added_to_manifest"`
	Installed        bool   `json:"installed"`
	Error            string `json:"error"`
}

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
				if globalFlags.JSON {
					r := addResult{Action: "add", ItemType: "plugin", Name: name, Error: fmt.Sprintf("plugin %q not found in catalog", name)}
					return json.NewEncoder(os.Stdout).Encode(r)
				}
				return fmt.Errorf("plugin %q not found in catalog. Run 'army fetch-catalog' to refresh", name)
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

			r := addResult{Action: "add", ItemType: "plugin", Name: cp.Name}

			added := manifest.AddPlugin(d.manifest, mp)
			if added {
				if err := manifest.Save(d.manifestPath, d.manifest); err != nil {
					if globalFlags.JSON {
						r.Error = fmt.Sprintf("saving manifest: %v", err)
						return json.NewEncoder(os.Stdout).Encode(r)
					}
					return fmt.Errorf("saving manifest: %w", err)
				}
				r.AddedToManifest = true
				if !globalFlags.JSON {
					fmt.Printf("Added plugin %q to manifest.\n", name)
				}
			} else {
				if !globalFlags.JSON {
					fmt.Printf("Plugin %q is already in the manifest.\n", name)
				}
			}

			if !noInstall && added {
				if !globalFlags.JSON {
					fmt.Printf("Installing %s...\n", name)
				}
				result := d.orchestrator.InstallItems([]types.ManifestPlugin{mp}, nil)
				if result.Failed > 0 {
					if globalFlags.JSON {
						r.Error = fmt.Sprintf("failed to install plugin %q", name)
						return json.NewEncoder(os.Stdout).Encode(r)
					}
					return fmt.Errorf("failed to install plugin %q", name)
				}
				r.Installed = true
			}

			if globalFlags.JSON {
				return json.NewEncoder(os.Stdout).Encode(r)
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
				if globalFlags.JSON {
					r := addResult{Action: "add", ItemType: "skill", Name: name, Error: fmt.Sprintf("skill %q not found in catalog", name)}
					return json.NewEncoder(os.Stdout).Encode(r)
				}
				return fmt.Errorf("skill %q not found in catalog. Run 'army fetch-catalog' to refresh", name)
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

			r := addResult{Action: "add", ItemType: "skill", Name: cs.Name}

			added := manifest.AddSkill(d.manifest, ms)
			if added {
				if err := manifest.Save(d.manifestPath, d.manifest); err != nil {
					if globalFlags.JSON {
						r.Error = fmt.Sprintf("saving manifest: %v", err)
						return json.NewEncoder(os.Stdout).Encode(r)
					}
					return fmt.Errorf("saving manifest: %w", err)
				}
				r.AddedToManifest = true
				if !globalFlags.JSON {
					fmt.Printf("Added skill %q to manifest.\n", name)
				}
			} else {
				if !globalFlags.JSON {
					fmt.Printf("Skill %q is already in the manifest.\n", name)
				}
			}

			if !noInstall && added {
				if !globalFlags.JSON {
					fmt.Printf("Installing %s...\n", name)
				}
				result := d.orchestrator.InstallItems(nil, []types.ManifestSkill{ms})
				if result.Failed > 0 {
					if globalFlags.JSON {
						r.Error = fmt.Sprintf("failed to install skill %q", name)
						return json.NewEncoder(os.Stdout).Encode(r)
					}
					return fmt.Errorf("failed to install skill %q", name)
				}
				r.Installed = true
			}

			if globalFlags.JSON {
				return json.NewEncoder(os.Stdout).Encode(r)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&noInstall, "no-install", false, "Add to manifest without installing")
	cmd.Flags().BoolVar(&project, "project", false, "Set destination to project scope")
	return cmd
}
