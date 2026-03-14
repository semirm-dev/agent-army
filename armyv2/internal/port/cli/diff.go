package cli

import (
	"fmt"
	"os"

	"github.com/smahovkic/agent-army/armyv2/internal/core/diff"
	"github.com/spf13/cobra"
)

func newDiffCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diff",
		Short: "Compare manifest vs installed state",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := resolveDeps()
			if err != nil {
				return err
			}

			installedPlugins, err := d.system.InstalledPlugins()
			if err != nil {
				return fmt.Errorf("reading installed plugins: %w", err)
			}

			installedSkills, err := d.system.InstalledSkills()
			if err != nil {
				return fmt.Errorf("reading installed skills: %w", err)
			}

			result := diff.Compare(d.manifest, installedPlugins, installedSkills)

			if !diff.HasDrift(result) {
				fmt.Println("No drift detected. Manifest matches installed state.")
				return nil
			}

			if len(result.MissingPlugins) > 0 {
				fmt.Println("Missing plugins (in manifest, not installed):")
				for _, p := range result.MissingPlugins {
					fmt.Printf("  \033[31m- %s\033[0m\n", p.Name)
				}
				fmt.Println()
			}

			if len(result.ExtraPlugins) > 0 {
				fmt.Println("Extra plugins (installed, not in manifest):")
				for _, p := range result.ExtraPlugins {
					fmt.Printf("  \033[33m+ %s\033[0m\n", p.Name)
				}
				fmt.Println()
			}

			if len(result.MissingSkills) > 0 {
				fmt.Println("Missing skills (in manifest, not installed):")
				for _, s := range result.MissingSkills {
					fmt.Printf("  \033[31m- %s\033[0m\n", s.Name)
				}
				fmt.Println()
			}

			if len(result.ExtraSkills) > 0 {
				fmt.Println("Extra skills (installed, not in manifest):")
				for _, s := range result.ExtraSkills {
					fmt.Printf("  \033[33m+ %s\033[0m\n", s.Name)
				}
			}

			// Exit code 1 for CI/scripts
			os.Exit(1)
			return nil
		},
	}
}
