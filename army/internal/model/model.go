package model

// Rule loaded from rules/*.md frontmatter.
type Rule struct {
	Name        string
	Description string // H1 heading (title)
	Summary     string // frontmatter description (brief summary)
	Scope       string
	Languages   []string
	UsesRules   []string
	Path        string
}

// Skill loaded from skills/*.md frontmatter.
type Skill struct {
	Name        string
	Description string // H1 heading (title)
	Summary     string // frontmatter description (brief summary)
	Scope       string
	Languages   []string
	UsesRules   []string
	Path        string
}

// Workflow represents a named workflow provided by a plugin.
type Workflow struct {
	Name        string
	Description string
}

// Plugin represents an external plugin from spec/claude/settings.json.
type Plugin struct {
	Name        string
	Description string
	Marketplace string
	Workflows   []Workflow
}

// PluginNames extracts plugin names from a slice of Plugin.
func PluginNames(plugins []Plugin) []string {
	names := make([]string, len(plugins))
	for i, p := range plugins {
		names[i] = p.Name
	}
	return names
}

// ResolvedDeps holds the fully resolved dependency information for an agent.
type ResolvedDeps struct {
	Skills      []Skill
	Rules       []Rule
	Plugins     []string
	DelegatesTo []Agent
}

// Agent loaded from agents/*.md frontmatter.
type Agent struct {
	Name        string
	Description string
	Role        string
	Domain      string // display grouping category (e.g., "Go", "Infrastructure"); optional
	Scope       string
	Access      string
	Languages   []string
	UsesSkills  []string
	UsesRules   []string
	UsesPlugins []string
	DelegatesTo []string
	Path        string
}
