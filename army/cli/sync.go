package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/smahovkic/agent-army/army/internal/core/manifest"
	"github.com/smahovkic/agent-army/army/internal/core/types"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	var destination string
	var yes bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Apply manifest to machine — install missing, remove extras",
		RunE: func(cmd *cobra.Command, args []string) error {
			if destination != "" && destination != "user" && destination != "project" {
				return fmt.Errorf("invalid --destination %q: must be \"user\" or \"project\"", destination)
			}

			if globalFlags.JSON && !yes {
				return fmt.Errorf("--json requires --yes for non-interactive mode")
			}

			d, err := resolveDeps()
			if err != nil {
				return err
			}

			actions, err := d.orchestrator.PlanActions(d.manifest)
			if err != nil {
				return fmt.Errorf("planning actions: %w", err)
			}

			// Project-level manifests should not remove orphans — they only
			// describe what this project needs, not the full system state.
			if !manifest.IsDefault(d.manifestPath) {
				filtered := actions[:0]
				for _, a := range actions {
					if a.Type != "remove" {
						filtered = append(filtered, a)
					}
				}
				actions = filtered
			}

			if len(actions) == 0 {
				if globalFlags.JSON {
					emitNDJSON(map[string]interface{}{"event": "complete", "succeeded": 0, "failed": 0})
					return nil
				}
				fmt.Println("Everything is in sync.")
				return nil
			}

			// Apply destination override from flag
			if destination != "" {
				applyDestination(actions, destination)
			}

			if globalFlags.JSON {
				return syncJSON(d, actions)
			}

			printPlan(actions, destination)

			// Confirmation prompt (skip for --dry-run or --yes)
			if !globalFlags.DryRun && !yes {
				tty, err := os.Open("/dev/tty")
				if err != nil {
					return fmt.Errorf("cannot open terminal for confirmation (use --yes to skip): %w", err)
				}
				defer tty.Close()
				scanner := bufio.NewScanner(tty)
				for {
					fmt.Print("Proceed? [Y/n/d(estination)] ")
					if !scanner.Scan() {
						fmt.Println("\nSync cancelled.")
						return nil
					}
					answer := strings.TrimSpace(scanner.Text())
					switch strings.ToLower(answer) {
					case "", "y":
						fmt.Println()
						goto execute
					case "n":
						fmt.Println("Sync cancelled.")
						return nil
					case "d":
						newDest := promptDestination(scanner, destination)
						if newDest != "" {
							destination = newDest
							applyDestination(actions, destination)
						}
						fmt.Println()
						printPlan(actions, destination)
					default:
						fmt.Println("Invalid option. Use Y, n, or d.")
					}
				}
			}

		execute:
			result := d.orchestrator.Execute(actions)

			fmt.Printf("\nDone: %d succeeded, %d failed\n", result.Succeeded, result.Failed)
			if result.Failed > 0 {
				return fmt.Errorf("%d action(s) failed", result.Failed)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&destination, "destination", "", "Override destination for all actions (\"user\" or \"project\")")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func applyDestination(actions []types.Action, dest string) {
	for i := range actions {
		actions[i].Destination = dest
	}
}

func printPlan(actions []types.Action, destOverride string) {
	if destOverride != "" {
		fmt.Printf("Destination override: %s\n\n", destOverride)
	}
	fmt.Printf("Planned %d action(s):\n", len(actions))
	for _, a := range actions {
		fmt.Printf("  %s %s %s (%s)\n", a.Type, a.ItemType, a.Name, a.Destination)
	}
	fmt.Println()
}

func promptDestination(scanner *bufio.Scanner, current string) string {
	hint := "user/project"
	if current != "" {
		hint = fmt.Sprintf("current: %s", current)
	}
	fmt.Printf("  Destination [%s]: ", hint)
	if !scanner.Scan() {
		return ""
	}
	val := strings.TrimSpace(scanner.Text())
	if val != "user" && val != "project" {
		fmt.Println("  Invalid destination. Must be \"user\" or \"project\".")
		return ""
	}
	return val
}

func emitNDJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "json marshal error: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func syncJSON(d *deps, actions []types.Action) error {
	// Emit plan event
	emitNDJSON(map[string]interface{}{
		"event":   "plan",
		"actions": actions,
	})

	succeeded := 0
	failed := 0

	for _, a := range actions {
		// Emit action_start
		emitNDJSON(map[string]interface{}{
			"event":     "action_start",
			"type":      a.Type,
			"item_type": a.ItemType,
			"name":      a.Name,
		})

		err := d.orchestrator.ExecuteAction(a)

		event := map[string]interface{}{
			"event":     "action_done",
			"type":      a.Type,
			"item_type": a.ItemType,
			"name":      a.Name,
			"success":   err == nil,
		}
		if err != nil {
			event["error"] = err.Error()
			failed++
		} else {
			succeeded++
		}
		emitNDJSON(event)
	}

	// Emit complete event
	emitNDJSON(map[string]interface{}{
		"event":     "complete",
		"succeeded": succeeded,
		"failed":    failed,
	})

	if failed > 0 {
		return fmt.Errorf("%d action(s) failed", failed)
	}
	return nil
}
