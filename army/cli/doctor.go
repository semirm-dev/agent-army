package cli

import (
	"fmt"
	"os"

	"github.com/smahovkic/agent-army/army/internal/core/doctor"
	"github.com/smahovkic/agent-army/army/internal/core/manifest"
	"github.com/smahovkic/agent-army/army/internal/core/types"
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

			cwd, _ := os.Getwd()
			manifestPath, provenance := resolveManifestWithProvenance(cwd)
			fmt.Printf("Manifest: %s %s(%s)%s\n", manifestPath, dim, provenance, reset)

			installedPlugins, err := d.state.InstalledPlugins()
			if err != nil {
				return fmt.Errorf("reading installed plugins: %w", err)
			}

			installedSkills, err := d.state.InstalledSkills()
			if err != nil {
				return fmt.Errorf("reading installed skills: %w", err)
			}

			issues := doctor.Check(d.manifest, installedPlugins, installedSkills)

			// Project-level manifests should not report orphans — they only
			// describe what this project needs, not the full system state.
			if !manifest.IsDefault(d.manifestPath) {
				filtered := make([]types.DoctorIssue, 0, len(issues))
				for _, i := range issues {
					if i.Category != "orphan" {
						filtered = append(filtered, i)
					}
				}
				issues = filtered
			}

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
