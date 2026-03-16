package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/smahovkic/agent-army/army/internal/core/config"
	"github.com/smahovkic/agent-army/army/internal/core/manifest"
	"github.com/spf13/cobra"
)

// ANSI colors — same convention as list.go
const (
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	cyan   = "\033[36m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	reset  = "\033[0m"
)

func newDetectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "detect",
		Short: "Show loaded config files for the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDetect()
		},
	}
}

type tableRow struct {
	label string
	value string
	sub   string // optional second line (e.g. provenance)
}

func runDetect() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	// Config
	configPath, err := config.Path()
	if err != nil {
		return fmt.Errorf("resolving config path: %w", err)
	}
	configValue := dim + "none" + reset + "  " + yellow + "(using defaults)" + reset
	if fileExists(configPath) {
		configValue = configPath
	}

	// Catalog
	overlayPath := filepath.Join(home, ".army", "catalog.json")
	catalogValue := "embedded"
	if fileExists(overlayPath) {
		catalogValue = "embedded + " + overlayPath
	}

	// Manifest
	manifestPath, provenance := resolveManifestWithProvenance(configPath, cwd)
	if manifestPath == "" {
		return fmt.Errorf("could not determine manifest path")
	}
	manifestValue := dim + "none" + reset + "  " + yellow + "(" + manifestPath + " does not exist)" + reset
	if fileExists(manifestPath) {
		manifestValue = manifestPath
	}

	// Load manifest for summary
	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	// Installed state files
	pluginsDB := filepath.Join(home, ".claude", "plugins", "installed_plugins.json")
	skillsDB := filepath.Join(home, ".agents", ".skill-lock.json")
	pluginsValue := dim + "none" + reset
	if fileExists(pluginsDB) {
		pluginsValue = pluginsDB
	}
	skillsValue := dim + "none" + reset
	if fileExists(skillsDB) {
		skillsValue = skillsDB
	}

	rows := []tableRow{
		{"Config", configValue, ""},
		{"Catalog", catalogValue, ""},
		{"Manifest", manifestValue, provenance},
		{"Plugins DB", pluginsValue, ""},
		{"Skills DB", skillsValue, ""},
	}

	labelWidth := 12
	maxValLen := 0
	for _, r := range rows {
		vl := visibleLen(r.value)
		if vl > maxValLen {
			maxValLen = vl
		}
		if r.sub != "" {
			sl := visibleLen(r.sub) + 2 // "↳ " prefix
			if sl > maxValLen {
				maxValLen = sl
			}
		}
	}
	colW := labelWidth + 2 // label + ": "
	valW := maxValLen
	innerWidth := colW + valW + 3 // "│ " + " " + " │"

	// Header
	fmt.Println()
	fmt.Printf("  📂 %sConfig Resolution%s  %s(cwd: %s)%s\n", bold, reset, dim, cwd, reset)
	fmt.Println()

	// Top border
	fmt.Printf("  ┌─%s─┬─%s─┐\n", strings.Repeat("─", colW), strings.Repeat("─", valW))

	for i, r := range rows {
		// Data row
		valPad := valW - visibleLen(r.value)
		fmt.Printf("  │ %s%-*s%s │ %s%s │\n",
			cyan, colW, r.label+":", reset,
			r.value, strings.Repeat(" ", valPad))

		// Sub-line (provenance)
		if r.sub != "" {
			subText := dim + "↳ " + r.sub + reset
			subPad := valW - visibleLen("↳ "+r.sub)
			fmt.Printf("  │ %-*s │ %s%s │\n",
				colW, "",
				subText, strings.Repeat(" ", subPad))
		}

		// Separator or bottom
		if i < len(rows)-1 {
			fmt.Printf("  ├─%s─┼─%s─┤\n", strings.Repeat("─", colW), strings.Repeat("─", valW))
		}
	}

	// Bottom border
	fmt.Printf("  └─%s─┴─%s─┘\n", strings.Repeat("─", colW), strings.Repeat("─", valW))

	// Summary below
	_ = innerWidth
	fmt.Printf("\n  %s Manifest:%s %d plugins, %d skills\n\n", bold, reset, len(m.Plugins), len(m.Skills))

	return nil
}

func resolveManifestWithProvenance(configPath, cwd string) (string, string) {
	if globalFlags.ManifestPath != "" {
		return globalFlags.ManifestPath, "via --manifest flag"
	}

	cfg, err := config.Load()
	if err == nil {
		if resolved := config.Resolve(cfg, cwd); resolved != "" {
			return resolved, "via config.json dir_map"
		}
	}

	defaultPath, err := manifest.DefaultPath()
	if err != nil {
		return "", ""
	}
	return defaultPath, "via default (~/.army/manifest.json)"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// visibleLen returns the display width of a string, stripping ANSI escape codes.
func visibleLen(s string) int {
	n := 0
	inEsc := false
	for _, r := range s {
		if r == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		n++
	}
	return n
}
