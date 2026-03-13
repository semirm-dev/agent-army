package cli

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command with all subcommands.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "army",
		Short: "Manage dependencies across rules, skills, and agents",
	}

	cmd.AddCommand(newManifestCmd())
	cmd.AddCommand(newResolveCmd())
	cmd.AddCommand(newBootstrapCmd())
	cmd.AddCommand(newUpdatePluginsSkillsCmd())
	cmd.AddCommand(newSyncCmd())
	cmd.AddCommand(newAnalyzeCmd())

	return cmd
}

func findRoot() string {
	cwd, _ := os.Getwd()
	for _, candidate := range []string{cwd, filepath.Dir(cwd), filepath.Dir(filepath.Dir(cwd))} {
		if isDir(filepath.Join(candidate, "spec", "rules")) {
			return candidate
		}
	}
	return cwd
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
