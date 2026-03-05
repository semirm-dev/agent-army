package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/semir/agent-army/internal/model"
)

// generateAntigravityMD reads the GEMINI.md template for Antigravity and fills marker sections
// with content generated from the selected agents, skills, and rules.
// The dest path is used to resolve the {{BASE}} placeholder.
func generateAntigravityMD(dest, templatePath string, agents []model.Agent, skills []model.Skill, rules []model.Rule) error {
	tmplBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read Antigravity GEMINI.md template: %w", err)
	}
	content := string(tmplBytes)

	content = replaceMarkerSection(content, agentBeginMarker, agentEndMarker, buildAntigravityAgentReferences(agents))
	content = replaceMarkerSection(content, tipsBeginMarker, tipsEndMarker, buildAntigravitySubagentTips(agents))
	content = replaceMarkerSection(content, skillBeginMarker, skillEndMarker, buildAntigravitySkillDefinitions(skills))
	content = replaceMarkerSection(content, ruleBeginMarker, ruleEndMarker, buildAntigravityRuleTable(rules))

	// Collapse 3+ consecutive newlines to 2 (one blank line).
	for strings.Contains(content, "\n\n\n") {
		content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
	}

	prefix := antigravityDestPrefix(dest)
	content = strings.ReplaceAll(content, basePlaceholder, prefix)

	return writeOutput(dest, "GEMINI.md", content)
}

// antigravityDestPrefix determines the display path prefix for the Antigravity destination directory.
// Global (~/.gemini/antigravity) uses "~/.gemini/antigravity", project-local uses ".agent", custom uses the path as-is.
func antigravityDestPrefix(dest string) string {
	home, err := os.UserHomeDir()
	if err == nil {
		globalAntigravity := filepath.Join(home, ".gemini", "antigravity")
		if filepath.Clean(dest) == filepath.Clean(globalAntigravity) {
			return "~/.gemini/antigravity"
		}
	}

	base := filepath.Base(dest)
	if base == ".agent" {
		return ".agent"
	}

	return dest
}

// buildAntigravityAgentReferences generates the agent reference list for Antigravity.
// Agents are listed as reference documents since Antigravity does not support native subagents.
func buildAntigravityAgentReferences(agents []model.Agent) string {
	if len(agents) == 0 {
		return ""
	}

	groups := groupAgentsByDomain(agents)
	order := orderedDomains(groups)

	var sb strings.Builder
	sb.WriteString("- **Agent References:** Reference documents for specialized roles in `{{BASE}}/agents/`:\n")
	for _, domain := range order {
		entries := groups[domain]
		if len(entries) == 0 {
			continue
		}

		var parts []string
		for _, a := range entries {
			part := fmt.Sprintf("`%s.md`", flattenName(a.Name))
			if a.Access == "read-only" {
				part += " (read-only)"
			}
			parts = append(parts, part)
		}
		sb.WriteString(fmt.Sprintf("  - **%s:** %s\n", domain, strings.Join(parts, ", ")))
	}
	return sb.String()
}

// buildAntigravitySubagentTips generates tips for working with agent references in Antigravity.
func buildAntigravitySubagentTips(agents []model.Agent) string {
	if len(agents) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("- **Agent Reference Tips:**\n")
	sb.WriteString("  - Agents are reference documents. Read them for domain expertise before working on related tasks.\n")

	var readOnlyNames []string
	for _, a := range agents {
		if a.Access == "read-only" {
			readOnlyNames = append(readOnlyNames, "`"+flattenName(a.Name)+"`")
		}
	}

	if len(readOnlyNames) > 0 {
		sb.WriteString(fmt.Sprintf("  - Read-only references (%s) describe review and analysis workflows.\n",
			strings.Join(readOnlyNames, ", ")))
	}

	return sb.String()
}

// buildAntigravitySkillDefinitions generates the skills list for Antigravity using plain file paths.
func buildAntigravitySkillDefinitions(skills []model.Skill) string {
	if len(skills) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("- **Workflow Skills:** Located in `{{BASE}}/skills/`. Read and follow these workflow files when the task matches:\n")
	for _, s := range skills {
		desc := skillDescription(s)
		sb.WriteString(fmt.Sprintf("  - `skills/%s/SKILL.md` -- %s\n", flattenName(s.Name), desc))
	}
	return sb.String()
}

// buildAntigravityRuleTable generates the rules table for Antigravity using plain file paths.
func buildAntigravityRuleTable(rules []model.Rule) string {
	if len(rules) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Rules are available in `{{BASE}}/rules/`:\n\n")
	sb.WriteString("| Rule File | Content |\n")
	sb.WriteString("|-----------|---------|")
	for _, r := range rules {
		flat := flattenName(r.Name)
		desc := ruleDescription(r)
		ruleFile := fmt.Sprintf("rules/%s.md", flat)
		sb.WriteString(fmt.Sprintf("\n| `%s` | %s |", ruleFile, desc))
	}
	sb.WriteString("\n\nRead the relevant rule file before working on tasks in that domain.\n")
	return sb.String()
}
