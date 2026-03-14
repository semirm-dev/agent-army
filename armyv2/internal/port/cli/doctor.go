package cli

import (
	"fmt"
	"os"

	"github.com/smahovkic/agent-army/armyv2/internal/core/doctor"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run health checks on plugins and skills",
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

			issues := doctor.Check(d.manifest, installedPlugins, installedSkills)

			if len(issues) == 0 {
				fmt.Println("\033[32m✓\033[0m No issues found.")
				return nil
			}

			errors := 0
			warnings := 0

			for _, issue := range issues {
				var icon string
				switch issue.Severity {
				case "error":
					icon = "\033[31m✗\033[0m"
					errors++
				case "warning":
					icon = "\033[33m!\033[0m"
					warnings++
				default:
					icon = "\033[34mℹ\033[0m"
				}
				fmt.Printf("  %s [%s] %s\n", icon, issue.Category, issue.Description)
			}

			fmt.Printf("\n%d error(s), %d warning(s)\n", errors, warnings)

			if errors > 0 {
				os.Exit(1)
			}
			return nil
		},
	}
}
