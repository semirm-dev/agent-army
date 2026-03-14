package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Show manifest contents with install status",
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

			pluginSet := make(map[string]bool)
			for _, p := range installedPlugins {
				pluginSet[p.Name] = true
			}

			skillSet := make(map[string]bool)
			for _, s := range installedSkills {
				skillSet[s.Name] = true
			}

			if len(d.manifest.Plugins) == 0 && len(d.manifest.Skills) == 0 {
				fmt.Println("Manifest is empty. Run 'armyv2 setup' to get started.")
				return nil
			}

			if len(d.manifest.Plugins) > 0 {
				fmt.Println("Plugins:")
				for _, p := range d.manifest.Plugins {
					status := "\033[32m✓\033[0m"
					if !pluginSet[p.Name] {
						status = "\033[31m✗\033[0m"
					}
					fmt.Printf("  %s %s (%s, %s)\n", status, p.Name, p.Marketplace, p.Destination)
				}
				fmt.Println()
			}

			if len(d.manifest.Skills) > 0 {
				fmt.Println("Skills:")
				for _, s := range d.manifest.Skills {
					status := "\033[32m✓\033[0m"
					if !skillSet[s.Name] {
						status = "\033[31m✗\033[0m"
					}
					fmt.Printf("  %s %s (%s, %s)\n", status, s.Name, s.Source, s.Destination)
				}
			}

			return nil
		},
	}
}
