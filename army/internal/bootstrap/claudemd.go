package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/semir/agent-army/internal/model"
)

const (
	agentBeginMarker   = "<!-- BEGIN:agent-definitions -->"
	agentEndMarker     = "<!-- END:agent-definitions -->"
	tipsBeginMarker    = "<!-- BEGIN:subagent-tips -->"
	tipsEndMarker      = "<!-- END:subagent-tips -->"
	pluginsBeginMarker = "<!-- BEGIN:plugins-overview -->"
	pluginsEndMarker   = "<!-- END:plugins-overview -->"
	skillBeginMarker   = "<!-- BEGIN:custom-skills -->"
	skillEndMarker         = "<!-- END:custom-skills -->"
	ruleBeginMarker        = "<!-- BEGIN:rules-table -->"
	ruleEndMarker          = "<!-- END:rules-table -->"

	basePlaceholder = "{{BASE}}"
)

// generateClaudeMD reads the CLAUDE.md template from templatePath and replaces
// marker sections with content generated from the selected agents, skills, rules, and plugins.
// The dest path is used to resolve the {{BASE}} placeholder to the correct path prefix.
func generateClaudeMD(dest, templatePath string, agents []model.Agent, skills []model.Skill, rules []model.Rule, plugins []model.Plugin) error {
	tmplBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read CLAUDE.md template: %w", err)
	}
	content := string(tmplBytes)

	content = replaceMarkerSection(content, agentBeginMarker, agentEndMarker, buildAgentDefinitions(agents))
	content = replaceMarkerSection(content, tipsBeginMarker, tipsEndMarker, buildSubagentTips(agents))
	content = replaceMarkerSection(content, pluginsBeginMarker, pluginsEndMarker, buildPluginsSection(plugins))
	content = replaceMarkerSection(content, skillBeginMarker, skillEndMarker, buildSkillDefinitions(skills))
	content = replaceMarkerSection(content, ruleBeginMarker, ruleEndMarker, buildRuleTable(rules))

	// Collapse 3+ consecutive newlines to 2 (one blank line).
	// Empty marker replacements leave behind surrounding newlines that accumulate.
	for strings.Contains(content, "\n\n\n") {
		content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
	}

	prefix := destPrefix(dest)
	content = strings.ReplaceAll(content, basePlaceholder, prefix)

	return writeOutput(dest, "CLAUDE.md", content)
}

// destPrefix determines the display path prefix for the destination directory.
// Global (~/.claude) uses "~/.claude", project-local uses ".claude", custom uses the path as-is.
func destPrefix(dest string) string {
	home, err := os.UserHomeDir()
	if err == nil {
		globalClaude := filepath.Join(home, ".claude")
		if filepath.Clean(dest) == filepath.Clean(globalClaude) {
			return "~/.claude"
		}
	}

	base := filepath.Base(dest)
	if base == ".claude" {
		return ".claude"
	}

	return dest
}

// replaceMarkerSection replaces the begin marker, end marker, and everything between them
// with the replacement content. Markers are stripped from the output — they only serve
// as injection points in the template.
func replaceMarkerSection(content, beginMarker, endMarker, replacement string) string {
	beginIdx := strings.Index(content, beginMarker)
	endIdx := strings.Index(content, endMarker)
	if beginIdx < 0 || endIdx < 0 || endIdx <= beginIdx {
		return content
	}

	before := content[:beginIdx]
	after := content[endIdx+len(endMarker):]

	return before + replacement + after
}

// buildAgentDefinitions generates the agent definitions list grouped by language/domain.
// Includes the lead-in prose and plugin/skill annotations from the agent's frontmatter.
// Returns empty string when no agents are selected.
func buildAgentDefinitions(agents []model.Agent) string {
	if len(agents) == 0 {
		return ""
	}

	groups := groupAgentsByDomain(agents)
	order := orderedDomains(groups)

	var sb strings.Builder
	sb.WriteString("- **Agent Definitions:** Reusable agent prompts live in `{{BASE}}/agents/`. Use these when delegating via the Task tool:\n")
	for _, domain := range order {
		entries := groups[domain]
		if len(entries) == 0 {
			continue
		}

		var parts []string
		for _, a := range entries {
			part := fmt.Sprintf("`%s.md`", flattenName(a.Name))
			annotations := agentAnnotations(a)
			if annotations != "" {
				part += " " + annotations
			}
			parts = append(parts, part)
		}
		sb.WriteString(fmt.Sprintf("  - **%s:** %s\n", domain, strings.Join(parts, ", ")))
	}
	return sb.String()
}

// agentAnnotations builds a parenthetical annotation string for an agent entry.
// Combines access, plugin, and skill info: "(read-only, uses `context7` plugin)"
func agentAnnotations(a model.Agent) string {
	var parts []string

	if a.Access == "read-only" {
		parts = append(parts, "read-only")
	}

	for _, p := range a.UsesPlugins {
		parts = append(parts, fmt.Sprintf("uses `%s` plugin", p))
	}

	if len(parts) == 0 {
		return ""
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

// groupAgentsByDomain groups agents into display categories.
func groupAgentsByDomain(agents []model.Agent) map[string][]model.Agent {
	groups := make(map[string][]model.Agent)

	for _, a := range agents {
		domain := agentDomain(a)
		groups[domain] = append(groups[domain], a)
	}
	return groups
}

// agentDomain determines the display domain for an agent.
// Uses the frontmatter "domain" field when set, otherwise falls back to a
// prefix-based heuristic derived from the agent name.
func agentDomain(a model.Agent) string {
	if a.Domain != "" {
		return a.Domain
	}

	name := a.Name
	switch {
	case strings.HasPrefix(name, "go/"):
		return "Go"
	case strings.HasPrefix(name, "typescript/"):
		return "TypeScript/JS"
	case strings.HasPrefix(name, "react/"):
		return "React"
	case strings.HasPrefix(name, "python/"):
		return "Python"
	case strings.HasPrefix(name, "database/"):
		return "Database"
	case strings.HasPrefix(name, "infrastructure/"):
		return "Infrastructure"
	case name == "arch-reviewer":
		return "Architecture"
	case name == "docs-writer":
		return "Documentation"
	default:
		return "Quality"
	}
}

// orderedDomains returns domain names in a stable display order.
// Known domains appear in a predefined order; any new domains from frontmatter
// are appended alphabetically.
func orderedDomains(groups map[string][]model.Agent) []string {
	preferred := []string{"Go", "TypeScript/JS", "React", "Python", "Database", "Infrastructure", "Architecture", "Documentation", "Quality"}

	seen := make(map[string]bool)
	var order []string
	for _, domain := range preferred {
		if _, ok := groups[domain]; ok {
			order = append(order, domain)
			seen[domain] = true
		}
	}

	var extras []string
	for domain := range groups {
		if !seen[domain] {
			extras = append(extras, domain)
		}
	}
	sort.Strings(extras)
	return append(order, extras...)
}

// buildSubagentTips generates the subagent launch tips section based on selected agents.
// Only includes the readonly tip if read-only agents are present, listing them by name.
func buildSubagentTips(agents []model.Agent) string {
	if len(agents) == 0 {
		return ""
	}

	var readOnlyNames []string
	for _, a := range agents {
		if a.Access == "read-only" {
			readOnlyNames = append(readOnlyNames, "`"+flattenName(a.Name)+"`")
		}
	}

	var sb strings.Builder
	sb.WriteString("- **Subagent Launch Tips:**\n")

	if len(readOnlyNames) > 0 {
		sb.WriteString(fmt.Sprintf("  - Use `readonly: true` when launching read-only agents (%s) to enforce read-only access at the tool level.\n",
			strings.Join(readOnlyNames, ", ")))
	}

	sb.WriteString("  - Use `model: \"fast\"` for quick, scoped tasks (reviewers analyzing a few files, simple test generation, codebase exploration). Use the default model for complex coding tasks requiring deep reasoning.\n")

	return sb.String()
}

// buildPluginsSection generates the unified plugins section listing all configured plugins
// with descriptions. For any plugin that has workflows defined, appends them as a sub-section.
// Returns empty string when no plugins are configured.
func buildPluginsSection(plugins []model.Plugin) string {
	if len(plugins) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("- **Configured Plugins:** External plugins installed via `claude plugin install` and enabled in `settings.json`:\n")

	for _, p := range plugins {
		desc := p.Name
		if p.Description != "" {
			desc = p.Description
		}
		sb.WriteString(fmt.Sprintf("  - `%s` -- %s\n", p.Name, desc))
	}

	// Append workflow sections for any plugin that defines them
	for _, p := range plugins {
		if len(p.Workflows) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("- **%s Workflows:** Use these structured workflows from the `%s` plugin when applicable:\n", capitalize(p.Name), p.Name))
		for _, w := range p.Workflows {
			sb.WriteString(fmt.Sprintf("  - `%s` -- %s\n", w.Name, w.Description))
		}
	}

	return sb.String()
}

// capitalize returns s with the first letter uppercased.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// buildSkillDefinitions generates the custom skills list with its lead-in prose.
// Returns empty string when no skills are selected.
func buildSkillDefinitions(skills []model.Skill) string {
	if len(skills) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("- **Custom Skills:** Located in `{{BASE}}/skills/`. Use these when the task matches:\n")
	for _, s := range skills {
		desc := skillDescription(s)
		sb.WriteString(fmt.Sprintf("  - `%s` -- %s\n", flattenName(s.Name), desc))
	}
	return sb.String()
}

// buildRuleTable generates the rules table with its lead-in prose.
// Returns empty string when no rules are selected.
func buildRuleTable(rules []model.Rule) string {
	if len(rules) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Detailed patterns are loaded on-demand from `{{BASE}}/rules/`:\n\n")
	sb.WriteString("| Rule File | Content |\n")
	sb.WriteString("|-----------|---------|")
	for _, r := range rules {
		flat := flattenName(r.Name)
		desc := ruleDescription(r)
		ruleFile := filepath.ToSlash(fmt.Sprintf("rules/%s.md", flat))
		sb.WriteString(fmt.Sprintf("\n| `%s` | %s |", ruleFile, desc))
	}
	sb.WriteString("\n\nAgents load their relevant pattern file at activation. The orchestrator loads only this core file.\n")
	return sb.String()
}
