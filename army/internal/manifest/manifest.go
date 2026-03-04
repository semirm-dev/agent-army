package manifest

import (
	"os"
	"path/filepath"

	"github.com/semir/agent-army/internal/graph"
	"github.com/semir/agent-army/internal/loader"
	"github.com/semir/agent-army/internal/model"
)

// GenerateManifest loads all entities, resolves transitive deps, and builds the manifest.
func GenerateManifest(root string) (OrderedMap, error) {
	rules, err := loader.LoadRules(root)
	if err != nil {
		return OrderedMap{}, err
	}
	skills, err := loader.LoadSkills(root)
	if err != nil {
		return OrderedMap{}, err
	}
	agents, err := loader.LoadAgents(root)
	if err != nil {
		return OrderedMap{}, err
	}

	ruleLookup := buildRuleLookup(rules)
	skillLookup := buildSkillLookup(skills)
	agentLookup := buildAgentLookup(agents)

	var ruleEntries []Entry
	for _, r := range rules {
		ruleEntries = append(ruleEntries, ruleEntry(r, ruleLookup))
	}

	var skillEntries []Entry
	for _, s := range skills {
		skillEntries = append(skillEntries, skillEntry(s, ruleLookup))
	}

	var agentEntries []Entry
	for _, a := range agents {
		agentEntries = append(agentEntries, agentEntry(a, ruleLookup, skillLookup, agentLookup))
	}

	m := OrderedMap{
		Keys: []string{"rules", "skills", "agents"},
		Sections: map[string][]Entry{
			"rules":  ruleEntries,
			"skills": skillEntries,
			"agents": agentEntries,
		},
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

func ruleEntry(r model.Rule, ruleLookup map[string][]string) Entry {
	resolved := resolveRuleDeps(r.UsesRules, ruleLookup)
	e := Entry{}
	e.Add("name", r.Name)
	e.Add("scope", r.Scope)
	if r.Scope == "language-specific" && len(r.Languages) > 0 {
		e.AddList("languages", r.Languages)
	}
	if len(resolved) > 0 {
		e.AddList("uses_rules", resolved)
	}
	e.Add("path", filepath.ToSlash(r.Path))
	return e
}

func skillEntry(s model.Skill, ruleLookup map[string][]string) Entry {
	resolved := resolveRuleDeps(s.UsesRules, ruleLookup)
	e := Entry{}
	e.Add("name", s.Name)
	e.Add("scope", s.Scope)
	if s.Scope == "language-specific" && len(s.Languages) > 0 {
		e.AddList("languages", s.Languages)
	}
	e.AddList("uses_rules", resolved)
	e.Add("path", filepath.ToSlash(s.Path))
	return e
}

func agentEntry(a model.Agent, ruleLookup, skillLookup map[string][]string, agentLookup map[string][]string) Entry {
	combinedRules := mergeAgentRules(a, skillLookup)
	resolvedRules := resolveRuleDeps(combinedRules, ruleLookup)
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
	e.AddList("uses_rules", resolvedRules)
	e.AddList("uses_plugins", a.UsesPlugins)
	e.AddList("delegates_to", resolvedDelegates)
	e.Add("path", filepath.ToSlash(a.Path))
	return e
}

func buildRuleLookup(rules []model.Rule) map[string][]string {
	m := make(map[string][]string, len(rules))
	for _, r := range rules {
		m[r.Name] = r.UsesRules
	}
	return m
}

func buildSkillLookup(skills []model.Skill) map[string][]string {
	m := make(map[string][]string, len(skills))
	for _, s := range skills {
		m[s.Name] = s.UsesRules
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

func resolveRuleDeps(seeds []string, ruleLookup map[string][]string) []string {
	if len(seeds) == 0 {
		return nil
	}
	return graph.ResolveTransitive(seeds, func(name string) []string {
		return ruleLookup[name]
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

func mergeAgentRules(a model.Agent, skillLookup map[string][]string) []string {
	own := make([]string, len(a.UsesRules))
	copy(own, a.UsesRules)

	var skillRules []string
	for _, skillName := range a.UsesSkills {
		skillRules = append(skillRules, skillLookup[skillName]...)
	}

	if len(own) > 0 && len(skillRules) > 0 {
		return append(own, skillRules...)
	}
	if len(skillRules) > 0 {
		return skillRules
	}
	return own
}
