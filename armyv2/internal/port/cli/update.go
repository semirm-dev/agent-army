package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/smahovkic/agent-army/armyv2/internal/core/catalog"
	"github.com/smahovkic/agent-army/armyv2/internal/core/types"
	"github.com/spf13/cobra"
)

const catalogURL = "https://raw.githubusercontent.com/smahovkic/agent-army/main/armyv2/internal/core/catalog/catalog.json"

func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Fetch latest catalog from GitHub",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Fetching latest catalog...")

			resp, err := http.Get(catalogURL)
			if err != nil {
				return fmt.Errorf("fetching catalog: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status %d fetching catalog", resp.StatusCode)
			}

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("reading response: %w", err)
			}

			// Validate the fetched catalog
			if err := validateCatalog(data); err != nil {
				return fmt.Errorf("invalid catalog: %w", err)
			}

			// Also validate via the catalog service
			if err := catalog.Validate(data); err != nil {
				return fmt.Errorf("catalog validation failed: %w", err)
			}

			// Write to ~/.armyv2/catalog.json
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("getting home dir: %w", err)
			}

			dir := filepath.Join(home, ".armyv2")
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("creating directory: %w", err)
			}

			path := filepath.Join(dir, "catalog.json")
			if err := os.WriteFile(path, data, 0644); err != nil {
				return fmt.Errorf("writing catalog: %w", err)
			}

			fmt.Printf("Catalog updated: %s\n", path)
			return nil
		},
	}
}

func validateCatalog(data []byte) error {
	var cat types.Catalog
	if err := json.Unmarshal(data, &cat); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	if cat.Version < 1 {
		return fmt.Errorf("version field missing or invalid")
	}

	for i, p := range cat.Plugins {
		if p.Name == "" {
			return fmt.Errorf("plugin at index %d missing name", i)
		}
		if p.Marketplace == "" {
			return fmt.Errorf("plugin %q missing marketplace", p.Name)
		}
	}

	for i, s := range cat.Skills {
		if s.Name == "" {
			return fmt.Errorf("skill at index %d missing name", i)
		}
		if s.Source == "" {
			return fmt.Errorf("skill %q missing source", s.Name)
		}
	}

	return nil
}
