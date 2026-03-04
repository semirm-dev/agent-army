package model

// Rule loaded from rules/*.md frontmatter.
type Rule struct {
	Name        string
	Description string
	Scope       string
	Languages   []string
	UsesRules   []string
	Path        string
}

// Skill loaded from skills/*.md frontmatter.
type Skill struct {
	Name        string
	Description string
	Scope       string
	Languages   []string
	UsesRules   []string
	Path        string
}

// Agent loaded from agents/*.md frontmatter.
type Agent struct {
	Name        string
	Description string
	Role        string
	Scope       string
	Access      string
	Languages   []string
	UsesSkills  []string
	UsesRules   []string
	UsesPlugins []string
	DelegatesTo []string
	Path        string
}
