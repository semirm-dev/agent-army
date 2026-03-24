package cli

import (
	"github.com/spf13/cobra"
)

// GlobalFlags holds flags shared across all commands.
type GlobalFlags struct {
	DryRun  bool
	Verbose bool
	JSON    bool
}

var globalFlags GlobalFlags

// NewRootCmd creates the root army command with all subcommands.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "army",
		Short: "Claude Code plugin & skill manager",
		Long:  "Interactive setup and lifecycle management for Claude Code plugins and skills.",
	}

	cmd.PersistentFlags().BoolVar(&globalFlags.DryRun, "dry-run", false, "Print commands without executing")
	cmd.PersistentFlags().BoolVar(&globalFlags.Verbose, "verbose", false, "Verbose output")
	cmd.PersistentFlags().BoolVar(&globalFlags.JSON, "json", false, "Output in JSON format")

	cmd.AddCommand(
		newSetupCmd(),
		newSyncCmd(),
		newAddCmd(),
		newRemoveCmd(),
		newClearCmd(),
		newListCmd(),
		newCatalogCmd(),
		newFetchCatalogCmd(),
		newDoctorCmd(),
		newDetectCmd(),
		newServeCmd(),
		newVersionCmd(),
	)

	return cmd
}
