package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/smahovkic/agent-army/army/internal/core/manifest"
	"github.com/smahovkic/agent-army/army/internal/core/types"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var scope string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show manifest contents with install status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if scope != "" && scope != "user" && scope != "project" {
				return fmt.Errorf("invalid --scope %q: must be \"user\" or \"project\"", scope)
			}

			d, err := resolveDeps()
			if err != nil {
				return err
			}

			// Override manifest if --scope is set
			if scope != "" {
				overridePath, err := manifestPathForScope(scope)
				if err != nil {
					return err
				}
				m, err := manifest.Load(overridePath)
				if err != nil {
					return fmt.Errorf("loading manifest: %w", err)
				}
				d.manifest = m
				d.manifestPath = overridePath
			}

			installedPlugins, err := d.state.InstalledPlugins()
			if err != nil {
				return fmt.Errorf("reading installed plugins: %w", err)
			}

			installedSkills, err := d.state.InstalledSkills()
			if err != nil {
				return fmt.Errorf("reading installed skills: %w", err)
			}

			pluginMap := make(map[string]types.InstalledPlugin)
			for _, p := range installedPlugins {
				pluginMap[p.Name] = p
			}

			skillSet := make(map[string]bool)
			for _, s := range installedSkills {
				skillSet[s.Name] = true
			}

			if globalFlags.JSON {
				return listJSON(d, pluginMap, skillSet)
			}

			if len(d.manifest.Plugins) == 0 && len(d.manifest.Skills) == 0 {
				fmt.Println("Manifest is empty. Run 'army setup' to get started.")
				return nil
			}

			home, _ := os.UserHomeDir()

			if len(d.manifest.Plugins) > 0 {
				installedCount := 0
				for _, p := range d.manifest.Plugins {
					if _, found := pluginMap[p.Name]; found {
						installedCount++
					}
				}
				fmt.Printf("%sPlugins (%d/%d):%s\n", bold, installedCount, len(d.manifest.Plugins), reset)
				for _, p := range d.manifest.Plugins {
					ip, found := pluginMap[p.Name]
					status := pluginStatus(found, ip)
					fmt.Printf("  %s %s (%s, %s)\n", status, p.Name, p.Marketplace, p.Destination)
				}
				fmt.Println()
			}

			if len(d.manifest.Skills) > 0 {
				installedCount := 0
				for _, s := range d.manifest.Skills {
					if skillSet[s.Name] {
						installedCount++
					}
				}
				fmt.Printf("%sSkills (%d/%d):%s\n", bold, installedCount, len(d.manifest.Skills), reset)
				for _, s := range d.manifest.Skills {
					status := skillStatus(skillSet[s.Name], s.Name, home)
					fmt.Printf("  %s %s (%s, %s)\n", status, s.Name, s.Source, s.Destination)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "Force manifest scope (\"user\" or \"project\")")
	return cmd
}

func manifestPathForScope(scope string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	if scope == "user" {
		return filepath.Join(home, ".army", "manifest.json"), nil
	}
	// project: walk up from CWD
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}
	projectPath, err := manifest.ResolveFromDir(cwd)
	if err != nil {
		return "", err
	}
	if manifest.IsDefault(projectPath) {
		// No project manifest found, use cwd-based path
		return filepath.Join(cwd, ".army", "manifest.json"), nil
	}
	return projectPath, nil
}

func pluginStatus(inJSON bool, ip types.InstalledPlugin) string {
	if !inJSON {
		return red + "✗" + reset
	}
	if ip.InstallPath != "" {
		if _, err := os.Stat(ip.InstallPath); os.IsNotExist(err) {
			return yellow + "⚠" + reset
		}
	}
	return green + "✓" + reset
}

func skillStatus(inLock bool, name, home string) string {
	if !inLock {
		return red + "✗" + reset
	}
	skillDir := filepath.Join(home, ".agents", "skills", name)
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return yellow + "⚠" + reset
	}
	return green + "✓" + reset
}

func listJSON(d *deps, pluginMap map[string]types.InstalledPlugin, skillSet map[string]bool) error {
	home, _ := os.UserHomeDir()

	scope := "user"
	if !manifest.IsDefault(d.manifestPath) {
		scope = "project"
	}

	type pluginEntry struct {
		Name        string   `json:"name"`
		Marketplace string   `json:"marketplace"`
		Destination string   `json:"destination"`
		Tags        []string `json:"tags"`
		Installed   bool     `json:"installed"`
		Status      string   `json:"status"`
	}

	type skillEntry struct {
		Name        string   `json:"name"`
		Source      string   `json:"source"`
		Destination string   `json:"destination"`
		Tags        []string `json:"tags"`
		Installed   bool     `json:"installed"`
		Status      string   `json:"status"`
	}

	type listOutput struct {
		ManifestPath  string        `json:"manifest_path"`
		ManifestScope string        `json:"manifest_scope"`
		Plugins       []pluginEntry `json:"plugins"`
		Skills        []skillEntry  `json:"skills"`
	}

	out := listOutput{
		ManifestPath:  d.manifestPath,
		ManifestScope: scope,
		Plugins:       make([]pluginEntry, 0, len(d.manifest.Plugins)),
		Skills:        make([]skillEntry, 0, len(d.manifest.Skills)),
	}

	for _, p := range d.manifest.Plugins {
		ip, found := pluginMap[p.Name]
		status := "missing"
		if found {
			status = "ok"
			if ip.InstallPath != "" {
				if _, err := os.Stat(ip.InstallPath); os.IsNotExist(err) {
					status = "drift"
				}
			}
		}
		tags := p.Tags
		if tags == nil {
			tags = []string{}
		}
		out.Plugins = append(out.Plugins, pluginEntry{
			Name:        p.Name,
			Marketplace: p.Marketplace,
			Destination: p.Destination,
			Tags:        tags,
			Installed:   found,
			Status:      status,
		})
	}

	for _, s := range d.manifest.Skills {
		installed := skillSet[s.Name]
		status := "missing"
		if installed {
			status = "ok"
			skillDir := filepath.Join(home, ".agents", "skills", s.Name)
			if _, err := os.Stat(skillDir); os.IsNotExist(err) {
				status = "drift"
			}
		}
		tags := s.Tags
		if tags == nil {
			tags = []string{}
		}
		out.Skills = append(out.Skills, skillEntry{
			Name:        s.Name,
			Source:      s.Source,
			Destination: s.Destination,
			Tags:        tags,
			Installed:   installed,
			Status:      status,
		})
	}

	return json.NewEncoder(os.Stdout).Encode(out)
}
