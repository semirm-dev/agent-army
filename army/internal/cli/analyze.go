package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/semir/agent-army/internal/plugindoc"
	"github.com/semir/agent-army/internal/termcolor"
	"github.com/spf13/cobra"
)

func newAnalyzeCmd() *cobra.Command {
	var fix bool

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze installed plugins and skills, report duplicates",
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := plugindoc.Analyze()
			if err != nil {
				return err
			}
			fmt.Print(report)

			if !fix {
				return nil
			}

			driftEntries, err := plugindoc.DetectDrift()
			if err != nil {
				return err
			}
			if len(driftEntries) == 0 {
				return nil
			}

			fmt.Println()
			fmt.Println(termcolor.Section("The following stale entries will be removed from .skill-lock.json:"))
			for _, e := range driftEntries {
				fmt.Println("  " + termcolor.Item(e.Name) + " (source: " + e.Source + ")")
			}
			fmt.Print("\nProceed? [y/N] ")

			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(strings.ToLower(answer))

			if answer != "y" && answer != "yes" {
				fmt.Println("Aborted.")
				return nil
			}

			if err := plugindoc.RemoveDriftEntries(driftEntries); err != nil {
				return fmt.Errorf("fixing drift: %w", err)
			}
			fmt.Println(termcolor.Success(fmt.Sprintf("Removed %d stale entries from .skill-lock.json.", len(driftEntries))))
			return nil
		},
	}

	cmd.Flags().BoolVar(&fix, "fix", false, "Remove stale skill entries from lock file")
	return cmd
}
