package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/smahovkic/agent-army/armyv2/internal/core/manifest"
	"github.com/smahovkic/agent-army/armyv2/internal/port/tui"
	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Interactive setup wizard for Claude Code plugins and skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := resolveDeps()
			if err != nil {
				return err
			}

			model := tui.NewSetupModel(d.catalog, d.manifest, d.manifestPath)
			p := tea.NewProgram(model, tea.WithAltScreen())

			finalModel, err := p.Run()
			if err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}

			// Save manifest if setup completed successfully
			if m, ok := finalModel.(tui.SetupModel); ok && m.Completed() {
				savePath := m.ManifestPath()
				if err := manifest.Save(savePath, m.ResultManifest()); err != nil {
					return fmt.Errorf("saving manifest: %w", err)
				}
				fmt.Println("Manifest saved to", savePath)
				fmt.Println("Run 'armyv2 sync' to install your selections.")
			}

			return nil
		},
	}
}
