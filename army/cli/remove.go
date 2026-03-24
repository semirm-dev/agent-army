package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/smahovkic/agent-army/army/internal/core/manifest"
	"github.com/spf13/cobra"
)

type removeResult struct {
	Action              string `json:"action"`
	ItemType            string `json:"item_type"`
	Name                string `json:"name"`
	RemovedFromManifest bool   `json:"removed_from_manifest"`
	Uninstalled         bool   `json:"uninstalled"`
	Error               string `json:"error"`
}

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

			r := removeResult{Action: "remove", ItemType: "plugin", Name: name}

			if !manifest.RemovePlugin(d.manifest, name) {
				if globalFlags.JSON {
					r.Error = fmt.Sprintf("plugin %q not found in manifest", name)
					return json.NewEncoder(os.Stdout).Encode(r)
				}
				return fmt.Errorf("plugin %q not found in manifest", name)
			}

			if err := manifest.Save(d.manifestPath, d.manifest); err != nil {
				if globalFlags.JSON {
					r.Error = fmt.Sprintf("saving manifest: %v", err)
					return json.NewEncoder(os.Stdout).Encode(r)
				}
				return fmt.Errorf("saving manifest: %w", err)
			}
			r.RemovedFromManifest = true
			if !globalFlags.JSON {
				fmt.Printf("Removed plugin %q from manifest.\n", name)
			}

			if !manifestOnly {
				actions, err := d.orchestrator.PlanActions(d.manifest)
				if err != nil {
					if globalFlags.JSON {
						r.Error = fmt.Sprintf("planning actions: %v", err)
						return json.NewEncoder(os.Stdout).Encode(r)
					}
					return fmt.Errorf("planning actions: %w", err)
				}
				if len(actions) > 0 {
					result := d.orchestrator.Execute(actions)
					if result.Failed > 0 {
						if globalFlags.JSON {
							r.Error = fmt.Sprintf("failed to uninstall plugin %q", name)
							return json.NewEncoder(os.Stdout).Encode(r)
						}
						return fmt.Errorf("failed to uninstall plugin %q", name)
					}
					r.Uninstalled = true
				}
			}

			if globalFlags.JSON {
				return json.NewEncoder(os.Stdout).Encode(r)
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

			r := removeResult{Action: "remove", ItemType: "skill", Name: name}

			if !manifest.RemoveSkill(d.manifest, name) {
				if globalFlags.JSON {
					r.Error = fmt.Sprintf("skill %q not found in manifest", name)
					return json.NewEncoder(os.Stdout).Encode(r)
				}
				return fmt.Errorf("skill %q not found in manifest", name)
			}

			if err := manifest.Save(d.manifestPath, d.manifest); err != nil {
				if globalFlags.JSON {
					r.Error = fmt.Sprintf("saving manifest: %v", err)
					return json.NewEncoder(os.Stdout).Encode(r)
				}
				return fmt.Errorf("saving manifest: %w", err)
			}
			r.RemovedFromManifest = true
			if !globalFlags.JSON {
				fmt.Printf("Removed skill %q from manifest.\n", name)
			}

			if !manifestOnly {
				actions, err := d.orchestrator.PlanActions(d.manifest)
				if err != nil {
					if globalFlags.JSON {
						r.Error = fmt.Sprintf("planning actions: %v", err)
						return json.NewEncoder(os.Stdout).Encode(r)
					}
					return fmt.Errorf("planning actions: %w", err)
				}
				if len(actions) > 0 {
					result := d.orchestrator.Execute(actions)
					if result.Failed > 0 {
						if globalFlags.JSON {
							r.Error = fmt.Sprintf("failed to uninstall skill %q", name)
							return json.NewEncoder(os.Stdout).Encode(r)
						}
						return fmt.Errorf("failed to uninstall skill %q", name)
					}
					r.Uninstalled = true
				}
			}

			if globalFlags.JSON {
				return json.NewEncoder(os.Stdout).Encode(r)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&manifestOnly, "manifest-only", false, "Remove from manifest without uninstalling")
	return cmd
}
