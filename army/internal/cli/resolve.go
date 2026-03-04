package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/semir/agent-army/internal/loader"
	"github.com/semir/agent-army/internal/manifest"
	"github.com/semir/agent-army/internal/model"
	"github.com/semir/agent-army/internal/resolver"
	"github.com/spf13/cobra"
)

func newResolveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resolve",
		Short: "Validate refs and fix redundancies",
		RunE: func(cmd *cobra.Command, args []string) error {
			root := findRoot()

			rules, err := loader.LoadRules(root)
			if err != nil {
				return err
			}
			skills, err := loader.LoadSkills(root)
			if err != nil {
				return err
			}
			agents, err := loader.LoadAgents(root)
			if err != nil {
				return err
			}
			plugins, err := loader.LoadPlugins(root)
			if err != nil {
				return err
			}

			errors := resolver.ValidateAllRefs(rules, skills, agents, model.PluginNames(plugins))
			fixes := resolver.ComputeAllFixes(rules, skills, agents, root)

			report := resolver.FormatReport(errors, fixes)
			fmt.Println(report)

			realErrors := 0
			for _, e := range errors {
				if e.Severity == "error" {
					realErrors++
				}
			}
			if realErrors > 0 {
				os.Exit(1)
			}

			if len(fixes) == 0 {
				return nil
			}

			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Remove redundant entries? [y/N] ")
			line, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(line)) != "y" {
				fmt.Println("Aborted. No files changed.")
				return nil
			}

			fmt.Println()
			if err := resolver.ApplyFixes(fixes, root); err != nil {
				return err
			}
			for _, fix := range fixes {
				fmt.Printf("Updated %s\n", fix.Label)
			}

			fmt.Println()
			fmt.Println("Regenerating manifest.json...")
			return manifest.WriteManifest(root)
		},
	}
}
