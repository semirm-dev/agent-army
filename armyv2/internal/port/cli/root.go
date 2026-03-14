package cli

import (
	"github.com/spf13/cobra"
)

// GlobalFlags holds flags shared across all commands.
type GlobalFlags struct {
	DryRun       bool
	ManifestPath string
	Verbose      bool
}

var globalFlags GlobalFlags

// NewRootCmd creates the root armyv2 command with all subcommands.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "armyv2",
		Short: "Claude Code plugin & skill manager",
		Long:  "Interactive setup and lifecycle management for Claude Code plugins and skills.",
	}

	cmd.PersistentFlags().BoolVar(&globalFlags.DryRun, "dry-run", false, "Print commands without executing")
	cmd.PersistentFlags().StringVar(&globalFlags.ManifestPath, "manifest", "", "Override manifest path (default: ~/.armyv2/manifest.json)")
	cmd.PersistentFlags().BoolVar(&globalFlags.Verbose, "verbose", false, "Verbose output")

	cmd.AddCommand(
		newSetupCmd(),
		newSyncCmd(),
		newAddCmd(),
		newRemoveCmd(),
		newListCmd(),
		newUpdateCmd(),
		newDoctorCmd(),
	)

	return cmd
}
