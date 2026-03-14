package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Apply manifest to machine — install missing, remove extras",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := resolveDeps()
			if err != nil {
				return err
			}

			actions, err := d.orchestrator.PlanActions(d.manifest)
			if err != nil {
				return fmt.Errorf("planning actions: %w", err)
			}

			if len(actions) == 0 {
				fmt.Println("Everything is in sync.")
				return nil
			}

			fmt.Printf("Planned %d action(s):\n", len(actions))
			for _, a := range actions {
				fmt.Printf("  %s %s %s\n", a.Type, a.ItemType, a.Name)
			}
			fmt.Println()

			result := d.orchestrator.Execute(actions)

			fmt.Printf("\nDone: %d succeeded, %d failed\n", result.Succeeded, result.Failed)
			if result.Failed > 0 {
				return fmt.Errorf("%d action(s) failed", result.Failed)
			}
			return nil
		},
	}
}
