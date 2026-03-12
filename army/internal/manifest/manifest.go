package manifest

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/semir/agent-army/internal/graph"
	"github.com/semir/agent-army/internal/loader"
	"github.com/semir/agent-army/internal/model"
)

// GenerateManifest loads all entities, resolves transitive deps, and builds the manifest.
func GenerateManifest(root string) (OrderedMap, error) {
	skills, err := loader.LoadSkills(root)
	if err != nil {
		return OrderedMap{}, err
	}
	agents, err := loader.LoadAgents(root)
	if err != nil {
		return OrderedMap{}, err
	}

	skillLookup := buildSkillLookup(skills)
	agentLookup := buildAgentLookup(agents)

	var skillEntries []Entry
	for _, s := range skills {
		skillEntries = append(skillEntries, skillEntry(s, skillLookup))
	}

	var agentEntries []Entry
	for _, a := range agents {
		agentEntries = append(agentEntries, agentEntry(a, skillLookup, agentLookup))
	}

	// Load external plugins/skills config for inclusion in manifest
	pluginsRaw, err := loader.LoadPluginsConfig(root)
	if err != nil {
		return OrderedMap{}, err
	}

	m := OrderedMap{
		Keys: []string{"skills", "agents"},
		Sections: map[string][]Entry{
			"skills": skillEntries,
			"agents": agentEntries,
		},
		RawSections: make(map[string]json.RawMessage),
	}

	if pluginsRaw != nil {
		// Parse the plugins config to extract individual sections
		var pluginsConfig struct {
			ExternalPlugins json.RawMessage `json:"external_plugins"`
			ExternalSkills  json.RawMessage `json:"external_skills"`
		}
		if err := json.Unmarshal(pluginsRaw, &pluginsConfig); err == nil {
			if pluginsConfig.ExternalPlugins != nil {
				m.Keys = append(m.Keys, "external_plugins")
				m.RawSections["external_plugins"] = pluginsConfig.ExternalPlugins
			}
			if pluginsConfig.ExternalSkills != nil {
				m.Keys = append(m.Keys, "external_skills")
				m.RawSections["external_skills"] = pluginsConfig.ExternalSkills
			}
		}
	}

	return m, nil
}

// WriteManifest generates and writes manifest.json.
func WriteManifest(root string) error {
	m, err := GenerateManifest(root)
	if err != nil {
		return err
	}
	output := formatManifestJSON(m)
	return os.WriteFile(filepath.Join(root, "manifest.json"), []byte(output), 0644)
}

func skillEntry(s model.Skill, skillLookup map[string][]string) Entry {
	resolved := resolveSkillDeps(s.UsesSkills, skillLookup)
	e := Entry{}
	e.Add("name", s.Name)
	e.Add("scope", s.Scope)
	if s.Scope == "language-specific" && len(s.Languages) > 0 {
		e.AddList("languages", s.Languages)
	}
	e.AddList("uses_skills", resolved)
	e.Add("path", filepath.ToSlash(s.Path))
	return e
}

func agentEntry(a model.Agent, skillLookup map[string][]string, agentLookup map[string][]string) Entry {
	resolvedDelegates := resolveAgentDelegates(a.DelegatesTo, agentLookup)

	e := Entry{}
	e.Add("name", a.Name)
	e.Add("role", a.Role)
	e.Add("scope", a.Scope)
	e.Add("access", a.Access)
	if len(a.Languages) > 0 {
		e.AddList("languages", a.Languages)
	}
	e.AddList("uses_skills", a.UsesSkills)
	e.AddList("uses_plugins", a.UsesPlugins)
	e.AddList("delegates_to", resolvedDelegates)
	e.Add("path", filepath.ToSlash(a.Path))
	return e
}

func buildSkillLookup(skills []model.Skill) map[string][]string {
	m := make(map[string][]string, len(skills))
	for _, s := range skills {
		m[s.Name] = s.UsesSkills
	}
	return m
}

func buildAgentLookup(agents []model.Agent) map[string][]string {
	m := make(map[string][]string, len(agents))
	for _, a := range agents {
		m[a.Name] = a.DelegatesTo
	}
	return m
}

func resolveSkillDeps(seeds []string, skillLookup map[string][]string) []string {
	if len(seeds) == 0 {
		return nil
	}
	return graph.ResolveTransitive(seeds, func(name string) []string {
		return skillLookup[name]
	})
}

func resolveAgentDelegates(seeds []string, agentLookup map[string][]string) []string {
	if len(seeds) == 0 {
		return nil
	}
	return graph.ResolveTransitive(seeds, func(name string) []string {
		return agentLookup[name]
	})
}
