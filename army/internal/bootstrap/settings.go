package bootstrap

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/semir/agent-army/internal/model"
)

// generateSettings reads the settings.json template from templatePath, replaces
// the enabledPlugins section with entries derived from the loaded plugins, and
// writes the result to the destination directory.
func generateSettings(dest, templatePath string, plugins []model.Plugin) error {
	tmplBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read settings template: %w", err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(tmplBytes, &settings); err != nil {
		return fmt.Errorf("parse settings template: %w", err)
	}

	// Build enabledPlugins from loaded plugins.
	// Format: "name@marketplace": true
	enabledPlugins := buildEnabledPlugins(plugins)
	if len(enabledPlugins) > 0 {
		settings["enabledPlugins"] = enabledPlugins
	}

	// Strip plugin metadata — these are source data for bootstrap,
	// not Claude Code settings for the generated output.
	delete(settings, "external_plugins")
	delete(settings, "external_skills")

	output, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}

	return writeOutput(dest, "settings.json", string(output)+"\n")
}

// buildEnabledPlugins constructs the enabledPlugins map from plugin metadata.
// Plugins without a marketplace field are skipped.
func buildEnabledPlugins(plugins []model.Plugin) map[string]bool {
	m := make(map[string]bool)
	for _, p := range plugins {
		if p.Name != "" && p.Marketplace != "" {
			key := p.Name + "@" + p.Marketplace
			m[key] = true
		}
	}
	return m
}
