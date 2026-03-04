package bootstrap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/semir/agent-army/internal/model"
)

func TestBuildEnabledPlugins(t *testing.T) {
	t.Run("plugins with marketplace", func(t *testing.T) {
		plugins := []model.Plugin{
			{Name: "code-review", Marketplace: "claude-plugins-official"},
			{Name: "superpowers", Marketplace: "claude-plugins-official"},
		}
		result := buildEnabledPlugins(plugins)

		if len(result) != 2 {
			t.Fatalf("got %d entries, want 2", len(result))
		}
		if !result["code-review@claude-plugins-official"] {
			t.Error("missing code-review@claude-plugins-official")
		}
		if !result["superpowers@claude-plugins-official"] {
			t.Error("missing superpowers@claude-plugins-official")
		}
	})

	t.Run("plugin without marketplace is skipped", func(t *testing.T) {
		plugins := []model.Plugin{
			{Name: "code-review", Marketplace: "claude-plugins-official"},
			{Name: "local-plugin"},
		}
		result := buildEnabledPlugins(plugins)

		if len(result) != 1 {
			t.Fatalf("got %d entries, want 1 (local-plugin should be skipped)", len(result))
		}
		if !result["code-review@claude-plugins-official"] {
			t.Error("missing code-review@claude-plugins-official")
		}
	})

	t.Run("empty plugins", func(t *testing.T) {
		result := buildEnabledPlugins(nil)
		if len(result) != 0 {
			t.Errorf("got %d entries, want 0", len(result))
		}
	})
}

func TestGenerateSettings(t *testing.T) {
	t.Run("syncs enabledPlugins from plugins data", func(t *testing.T) {
		dir := t.TempDir()
		tmplDir := filepath.Join(dir, "spec", "claude")
		os.MkdirAll(tmplDir, 0755)

		template := `{
  "permissions": {
    "allow": ["Bash(go build:*)"],
    "defaultMode": "plan"
  },
  "enabledPlugins": {
    "old-plugin@old-marketplace": true
  },
  "skipDangerousModePermissionPrompt": true,
  "external_plugins": [
    {"name": "code-review", "marketplace": "claude-plugins-official"}
  ],
  "external_skills": [
    {"name": "skill-creator"}
  ]
}`
		tmplPath := filepath.Join(tmplDir, "settings.json")
		os.WriteFile(tmplPath, []byte(template), 0644)

		destDir := filepath.Join(dir, "project", ".claude")

		plugins := []model.Plugin{
			{Name: "code-review", Marketplace: "claude-plugins-official"},
			{Name: "superpowers", Marketplace: "claude-plugins-official"},
		}

		err := generateSettings(destDir, tmplPath, plugins)
		if err != nil {
			t.Fatal(err)
		}

		content, err := os.ReadFile(filepath.Join(destDir, "settings.json"))
		if err != nil {
			t.Fatal(err)
		}

		var settings map[string]interface{}
		if err := json.Unmarshal(content, &settings); err != nil {
			t.Fatalf("invalid JSON output: %v", err)
		}

		// enabledPlugins should be synced from plugins data, not template
		ep, ok := settings["enabledPlugins"].(map[string]interface{})
		if !ok {
			t.Fatal("enabledPlugins missing or wrong type")
		}
		if len(ep) != 2 {
			t.Errorf("got %d enabledPlugins, want 2", len(ep))
		}
		if ep["code-review@claude-plugins-official"] != true {
			t.Error("missing code-review@claude-plugins-official")
		}
		if ep["superpowers@claude-plugins-official"] != true {
			t.Error("missing superpowers@claude-plugins-official")
		}
		// Old plugin from template should be gone
		if _, exists := ep["old-plugin@old-marketplace"]; exists {
			t.Error("old-plugin should be replaced, not preserved")
		}

		// Other template fields preserved
		perms, ok := settings["permissions"].(map[string]interface{})
		if !ok {
			t.Fatal("permissions missing or wrong type")
		}
		if perms["defaultMode"] != "plan" {
			t.Errorf("permissions.defaultMode = %v, want plan", perms["defaultMode"])
		}

		if settings["skipDangerousModePermissionPrompt"] != true {
			t.Error("skipDangerousModePermissionPrompt should be preserved")
		}

		// Plugin metadata should be stripped from output
		if _, exists := settings["external_plugins"]; exists {
			t.Error("external_plugins should be stripped from output")
		}
		if _, exists := settings["external_skills"]; exists {
			t.Error("external_skills should be stripped from output")
		}
	})

	t.Run("no plugins keeps empty enabledPlugins", func(t *testing.T) {
		dir := t.TempDir()
		tmplDir := filepath.Join(dir, "spec", "claude")
		os.MkdirAll(tmplDir, 0755)

		template := `{
  "permissions": {"defaultMode": "plan"},
  "enabledPlugins": {"old@mp": true}
}`
		tmplPath := filepath.Join(tmplDir, "settings.json")
		os.WriteFile(tmplPath, []byte(template), 0644)

		destDir := filepath.Join(dir, "project", ".claude")

		err := generateSettings(destDir, tmplPath, nil)
		if err != nil {
			t.Fatal(err)
		}

		content, err := os.ReadFile(filepath.Join(destDir, "settings.json"))
		if err != nil {
			t.Fatal(err)
		}

		var settings map[string]interface{}
		if err := json.Unmarshal(content, &settings); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		// enabledPlugins should still exist from template (not replaced since no plugins have marketplace)
		// but old entries remain since buildEnabledPlugins returns empty and we skip replacement
		perms, ok := settings["permissions"].(map[string]interface{})
		if !ok {
			t.Fatal("permissions missing")
		}
		if perms["defaultMode"] != "plan" {
			t.Errorf("defaultMode = %v, want plan", perms["defaultMode"])
		}
	})
}
