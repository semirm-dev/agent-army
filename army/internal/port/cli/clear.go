package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/smahovkic/agent-army/army/internal/core/types"
	"github.com/spf13/cobra"
)

func newClearCmd() *cobra.Command {
	var yes bool
	var full bool

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Uninstall plugins and skills from the system",
		Long: `Remove installed plugins and skills from the system.

By default, removes only items listed in the current manifest.
Use --full to remove ALL installed plugins and skills regardless of manifest.

The manifest itself is not modified — use 'army sync' to re-install.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := resolveDeps()
			if err != nil {
				return err
			}

			var actions []types.Action
			if full {
				actions, err = d.orchestrator.PlanFullClear()
			} else {
				actions, err = d.orchestrator.PlanClear(d.manifest)
			}
			if err != nil {
				return fmt.Errorf("planning clear actions: %w", err)
			}

			if len(actions) == 0 {
				if full {
					fmt.Println("Nothing to clear — no plugins or skills are installed.")
				} else {
					fmt.Println("Nothing to clear — no manifest items are currently installed.")
				}
				return nil
			}

			printPlan(actions, "")

			if !globalFlags.DryRun && !yes {
				tty, err := os.Open("/dev/tty")
				if err != nil {
					return fmt.Errorf("cannot open terminal for confirmation (use --yes to skip): %w", err)
				}
				defer tty.Close()
				scanner := bufio.NewScanner(tty)

				prompt := "Proceed? [y/N] "
				if full {
					prompt = "This will remove ALL installed plugins and skills. Proceed? [y/N] "
				}
				fmt.Print(prompt)

				if !scanner.Scan() {
					fmt.Println("\nClear cancelled.")
					return nil
				}
				answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
				if answer != "y" && answer != "yes" {
					fmt.Println("Clear cancelled.")
					return nil
				}
				fmt.Println()
			}

			result := d.orchestrator.Execute(actions)

			fmt.Printf("\nDone: %d succeeded, %d failed\n", result.Succeeded, result.Failed)
			if result.Failed > 0 {
				return fmt.Errorf("%d action(s) failed", result.Failed)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&full, "full", false, "Remove ALL installed plugins and skills, not just manifest items")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}
