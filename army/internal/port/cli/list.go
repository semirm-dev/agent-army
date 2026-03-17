package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/smahovkic/agent-army/army/internal/core/types"
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

			pluginMap := make(map[string]types.InstalledPlugin)
			for _, p := range installedPlugins {
				pluginMap[p.Name] = p
			}

			skillSet := make(map[string]bool)
			for _, s := range installedSkills {
				skillSet[s.Name] = true
			}

			if len(d.manifest.Plugins) == 0 && len(d.manifest.Skills) == 0 {
				fmt.Println("Manifest is empty. Run 'army setup' to get started.")
				return nil
			}

			home, _ := os.UserHomeDir()

			if len(d.manifest.Plugins) > 0 {
				installedCount := 0
				for _, p := range d.manifest.Plugins {
					if _, found := pluginMap[p.Name]; found {
						installedCount++
					}
				}
				fmt.Printf("%sPlugins (%d/%d):%s\n", bold, installedCount, len(d.manifest.Plugins), reset)
				for _, p := range d.manifest.Plugins {
					ip, found := pluginMap[p.Name]
					status := pluginStatus(found, ip)
					fmt.Printf("  %s %s (%s, %s)\n", status, p.Name, p.Marketplace, p.Destination)
				}
				fmt.Println()
			}

			if len(d.manifest.Skills) > 0 {
				installedCount := 0
				for _, s := range d.manifest.Skills {
					if skillSet[s.Name] {
						installedCount++
					}
				}
				fmt.Printf("%sSkills (%d/%d):%s\n", bold, installedCount, len(d.manifest.Skills), reset)
				for _, s := range d.manifest.Skills {
					status := skillStatus(skillSet[s.Name], s.Name, home)
					fmt.Printf("  %s %s (%s, %s)\n", status, s.Name, s.Source, s.Destination)
				}
			}

			return nil
		},
	}
}

func pluginStatus(inJSON bool, ip types.InstalledPlugin) string {
	if !inJSON {
		return red + "✗" + reset
	}
	if ip.InstallPath != "" {
		if _, err := os.Stat(ip.InstallPath); os.IsNotExist(err) {
			return yellow + "⚠" + reset
		}
	}
	return green + "✓" + reset
}

func skillStatus(inLock bool, name, home string) string {
	if !inLock {
		return red + "✗" + reset
	}
	skillDir := filepath.Join(home, ".agents", "skills", name)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return yellow + "⚠" + reset
	}
	return green + "✓" + reset
}
