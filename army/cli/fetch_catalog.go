package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/smahovkic/agent-army/army/internal/core/catalog"
	"github.com/smahovkic/agent-army/army/internal/core/types"
	"github.com/spf13/cobra"
)

const catalogURL = "https://raw.githubusercontent.com/semirm-dev/agent-army/main/army/internal/core/catalog/catalog.json"

func newFetchCatalogCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fetch-catalog",
		Short: "Fetch latest catalog from GitHub",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !globalFlags.JSON {
				fmt.Println("Fetching latest catalog...")
			}

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

			// Write to ~/.army/catalog.json
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("getting home dir: %w", err)
			}

			dir := filepath.Join(home, ".army")
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("creating directory: %w", err)
			}

			path := filepath.Join(dir, "catalog.json")
			if err := os.WriteFile(path, data, 0644); err != nil {
				return fmt.Errorf("writing catalog: %w", err)
			}

			if globalFlags.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"path":    path,
					"updated": true,
				})
			}

			fmt.Printf("Catalog saved: %s\n", path)
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
