package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/smahovkic/agent-army/army/internal/core/manifest"
	"github.com/smahovkic/agent-army/army/internal/core/types"
	"github.com/spf13/cobra"
)

func newWriteManifestCmd() *cobra.Command {
	var destination string
	var inputFlag string

	cmd := &cobra.Command{
		Use:   "write-manifest",
		Short: "Write a complete manifest from JSON (via --input flag or stdin)",
		RunE: func(cmd *cobra.Command, args []string) error {
			var data []byte
			var err error

			if inputFlag != "" {
				data = []byte(inputFlag)
			} else {
				data, err = io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("reading stdin: %w", err)
				}
			}

			var m types.Manifest
			if err := json.Unmarshal(data, &m); err != nil {
				return fmt.Errorf("parsing manifest JSON: %w", err)
			}
			if m.Version == 0 {
				m.Version = 1
			}

			// Stamp destination on all items
			for i := range m.Plugins {
				m.Plugins[i].Destination = destination
			}
			for i := range m.Skills {
				m.Skills[i].Destination = destination
			}

			// Determine path
			var savePath string
			if destination == "project" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("getting working directory: %w", err)
				}
				savePath = filepath.Join(cwd, ".army", "manifest.json")
			} else {
				p, err := manifest.DefaultPath()
				if err != nil {
					return err
				}
				savePath = p
			}

			if err := manifest.Save(savePath, &m); err != nil {
				return fmt.Errorf("saving manifest: %w", err)
			}

			if globalFlags.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]string{
					"path": savePath,
				})
			}

			fmt.Printf("Manifest saved: %s\n", savePath)
			return nil
		},
	}

	cmd.Flags().StringVar(&destination, "destination", "user", "Manifest destination: user or project")
	cmd.Flags().StringVar(&inputFlag, "input", "", "Manifest JSON string (alternative to stdin)")
	return cmd
}
