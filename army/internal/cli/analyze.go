package cli

import (
	"fmt"

	"github.com/semir/agent-army/internal/plugindoc"
	"github.com/spf13/cobra"
)

func newAnalyzeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "analyze",
		Short: "Analyze installed plugins and skills, report duplicates",
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := plugindoc.Analyze()
			if err != nil {
				return err
			}
			fmt.Print(report)
			return nil
		},
	}
}
