package loader

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/semir/agent-army/internal/frontmatter"
	"github.com/semir/agent-army/internal/model"
)

// FindMDFiles returns all .md files under dir, sorted by path.
func FindMDFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

// LoadRules loads all rules from root/rules/ directory.
func LoadRules(root string) ([]model.Rule, error) {
	rulesDir := filepath.Join(root, "spec", "rules")
	if !isDir(rulesDir) {
		return nil, nil
	}

	files, err := FindMDFiles(rulesDir)
	if err != nil {
		return nil, err
	}

	var rules []model.Rule
	for _, fp := range files {
		content, err := os.ReadFile(fp)
		if err != nil {
			return nil, err
		}
		fm := frontmatter.ParseFrontmatter(string(content))
		rel, _ := filepath.Rel(rulesDir, fp)
		name := strings.TrimSuffix(rel, ".md")

		rules = append(rules, model.Rule{
			Name:        name,
			Description: frontmatter.ExtractH1(string(content)),
			Summary:     fm.StringVal("description", ""),
			Scope:       fm.StringVal("scope", "universal"),
			Languages:   ensureList(fm, "languages"),
			UsesRules:   ensureList(fm, "uses_rules"),
			Path:        filepath.Join("spec", "rules", rel),
		})
	}
	return rules, nil
}

// LoadSkills loads all skills from root/skills/ directory.
func LoadSkills(root string) ([]model.Skill, error) {
	skillsDir := filepath.Join(root, "spec", "skills")
	if !isDir(skillsDir) {
		return nil, nil
	}

	files, err := FindMDFiles(skillsDir)
	if err != nil {
		return nil, err
	}

	var skills []model.Skill
	for _, fp := range files {
		content, err := os.ReadFile(fp)
		if err != nil {
			return nil, err
		}
		fm := frontmatter.ParseFrontmatter(string(content))
		rel, _ := filepath.Rel(skillsDir, fp)
		name := fm.StringVal("name", strings.TrimSuffix(rel, ".md"))

		skills = append(skills, model.Skill{
			Name:        name,
			Description: frontmatter.ExtractH1(string(content)),
			Summary:     fm.StringVal("description", ""),
			Scope:       fm.StringVal("scope", "universal"),
			Languages:   ensureList(fm, "languages"),
			UsesRules:   ensureList(fm, "uses_rules"),
			Path:        filepath.Join("spec", "skills", rel),
		})
	}
	return skills, nil
}

// LoadAgents loads all agents from root/agents/ directory.
func LoadAgents(root string) ([]model.Agent, error) {
	agentsDir := filepath.Join(root, "spec", "agents")
	if !isDir(agentsDir) {
		return nil, nil
	}

	files, err := FindMDFiles(agentsDir)
	if err != nil {
		return nil, err
	}

	var agents []model.Agent
	for _, fp := range files {
		content, err := os.ReadFile(fp)
		if err != nil {
			return nil, err
		}
		fm := frontmatter.ParseFrontmatter(string(content))
		rel, _ := filepath.Rel(agentsDir, fp)
		name := fm.StringVal("name", strings.TrimSuffix(rel, ".md"))

		agents = append(agents, model.Agent{
			Name:        name,
			Description: fm.StringVal("description", ""),
			Role:        fm.StringVal("role", ""),
			Domain:      fm.StringVal("domain", ""),
			Scope:       fm.StringVal("scope", "universal"),
			Access:      fm.StringVal("access", "read-write"),
			Languages:   ensureList(fm, "languages"),
			UsesSkills:  ensureList(fm, "uses_skills"),
			UsesRules:   ensureList(fm, "uses_rules"),
			UsesPlugins: ensureList(fm, "uses_plugins"),
			DelegatesTo: ensureList(fm, "delegates_to"),
			Path:        filepath.Join("spec", "agents", rel),
		})
	}
	return agents, nil
}

// LoadPlugins loads plugins from root/spec/claude/settings.json -> external_plugins[].
func LoadPlugins(root string) ([]model.Plugin, error) {
	pluginsPath := filepath.Join(root, "spec", "claude", "settings.json")
	data, err := os.ReadFile(pluginsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var config struct {
		ExternalPlugins []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Marketplace string `json:"marketplace"`
			Workflows   []struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"workflows"`
		} `json:"external_plugins"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, nil
	}

	var plugins []model.Plugin
	for _, p := range config.ExternalPlugins {
		if p.Name != "" {
			plugin := model.Plugin{
				Name:        p.Name,
				Description: p.Description,
				Marketplace: p.Marketplace,
			}
			for _, w := range p.Workflows {
				if w.Name != "" {
					plugin.Workflows = append(plugin.Workflows, model.Workflow{
						Name:        w.Name,
						Description: w.Description,
					})
				}
			}
			plugins = append(plugins, plugin)
		}
	}
	return plugins, nil
}

// LoadPluginsConfig extracts external_plugins and external_skills from
// spec/claude/settings.json for inclusion in the manifest.
func LoadPluginsConfig(root string) (json.RawMessage, error) {
	settingsPath := filepath.Join(root, "spec", "claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var full map[string]json.RawMessage
	if err := json.Unmarshal(data, &full); err != nil {
		return nil, nil
	}

	subset := make(map[string]json.RawMessage)
	if v, ok := full["external_plugins"]; ok {
		subset["external_plugins"] = v
	}
	if v, ok := full["external_skills"]; ok {
		subset["external_skills"] = v
	}
	if len(subset) == 0 {
		return nil, nil
	}

	out, err := json.Marshal(subset)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(out), nil
}

func ensureList(fm frontmatter.Frontmatter, key string) []string {
	list := fm.ListVal(key)
	if list == nil {
		return nil
	}
	return list
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
