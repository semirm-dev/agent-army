package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
)

func newCatalogCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "catalog",
		Short: "Show the full catalog of available plugins and skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := resolveDeps()
			if err != nil {
				return err
			}

			cat := d.catalog.GetCatalog()

			if globalFlags.JSON {
				return json.NewEncoder(os.Stdout).Encode(cat)
			}

			// Human-readable summary
			fmt.Printf("%sCatalog%s (version %d, updated %s)\n\n", bold, reset, cat.Version, cat.UpdatedAt)

			fmt.Printf("%sPlugins (%d):%s\n", bold, len(cat.Plugins), reset)
			for _, p := range cat.Plugins {
				fmt.Printf("  %s — %s\n", p.Name, p.Description)
			}
			fmt.Println()

			fmt.Printf("%sSkills (%d):%s\n", bold, len(cat.Skills), reset)
			for _, s := range cat.Skills {
				fmt.Printf("  %s — %s\n", s.Name, s.Description)
			}
			fmt.Println()

			profileNames := make([]string, 0, len(cat.TechProfiles))
			for name := range cat.TechProfiles {
				profileNames = append(profileNames, name)
			}
			sort.Strings(profileNames)

			fmt.Printf("%sTech Profiles (%d):%s\n", bold, len(cat.TechProfiles), reset)
			for _, name := range profileNames {
				fmt.Printf("  %s\n", name)
			}

			return nil
		},
	}
}
