package cli

import (
	"fmt"
	"os"
	"path/filepath"

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

func runDetect() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	fmt.Printf("\n%s📂 Config Resolution%s %s(cwd: %s)%s\n\n", bold, reset, dim, cwd, reset)

	// Config
	configPath, err := config.Path()
	if err != nil {
		return fmt.Errorf("resolving config path: %w", err)
	}
	if fileExists(configPath) {
		printRow("⚙", cyan, "Config", configPath)
	} else {
		printRow("⚙", cyan, "Config", dim+"none"+reset+" "+yellow+"(using defaults)"+reset)
	}

	// Catalog
	overlayPath := filepath.Join(home, ".army", "catalog.json")
	if fileExists(overlayPath) {
		printRow("⚙", cyan, "Catalog", "embedded + "+overlayPath)
	} else {
		printRow("⚙", cyan, "Catalog", "embedded")
	}

	// Manifest resolution with provenance
	manifestPath, provenance := resolveManifestWithProvenance(configPath, cwd)
	if manifestPath == "" {
		return fmt.Errorf("could not determine manifest path")
	}
	if fileExists(manifestPath) {
		printRow("📄", "", "Manifest", manifestPath)
	} else {
		printRow("📄", "", "Manifest", dim+"none"+reset+" "+yellow+"("+manifestPath+" does not exist)"+reset)
	}
	fmt.Printf("               %s↳ %s%s\n", dim, provenance, reset)

	// Load manifest for summary
	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	// Installed state files
	pluginsDB := filepath.Join(home, ".claude", "plugins", "installed_plugins.json")
	skillsDB := filepath.Join(home, ".agents", ".skill-lock.json")
	if fileExists(pluginsDB) {
		printRow("🔌", "", "Plugins DB", pluginsDB)
	} else {
		printRow("🔌", "", "Plugins DB", dim+"none"+reset)
	}
	if fileExists(skillsDB) {
		printRow("🧩", "", "Skills DB", skillsDB)
	} else {
		printRow("🧩", "", "Skills DB", dim+"none"+reset)
	}

	// Summary
	fmt.Printf("\n  %s📋 Manifest:%s %d plugins, %d skills\n\n", bold, reset, len(m.Plugins), len(m.Skills))

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

func printRow(icon, iconColor, label, value string) {
	if iconColor != "" {
		fmt.Printf("  %s%s%s  %-12s %s\n", iconColor, icon, reset, label+":", value)
	} else {
		fmt.Printf("  %s  %-12s %s\n", icon, label+":", value)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
