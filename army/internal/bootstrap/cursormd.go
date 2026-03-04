package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/semir/agent-army/internal/model"
)

// generateAgentsMD reads the AGENTS.md template from templatePath and replaces
// marker sections with content generated from the selected agents, skills, and rules.
// The dest path is used to resolve the {{BASE}} placeholder.
func generateAgentsMD(dest, templatePath string, agents []model.Agent, skills []model.Skill, rules []model.Rule, cursorRuleNames map[string]string) error {
	tmplBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read AGENTS.md template: %w", err)
	}
	content := string(tmplBytes)

	content = replaceMarkerSection(content, agentBeginMarker, agentEndMarker, buildCursorAgentDefinitions(agents))
	content = replaceMarkerSection(content, tipsBeginMarker, tipsEndMarker, buildCursorSubagentTips(agents))
	content = replaceMarkerSection(content, skillBeginMarker, skillEndMarker, buildCursorSkillDefinitions(skills))
	content = replaceMarkerSection(content, ruleBeginMarker, ruleEndMarker, buildCursorRuleTable(rules, cursorRuleNames))

	// Collapse 3+ consecutive newlines to 2 (one blank line).
	for strings.Contains(content, "\n\n\n") {
		content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
	}

	prefix := cursorDestPrefix(dest)
	content = strings.ReplaceAll(content, basePlaceholder, prefix)

	return writeOutput(dest, "AGENTS.md", content)
}

// cursorDestPrefix determines the display path prefix for the Cursor destination directory.
func cursorDestPrefix(dest string) string {
	home, err := os.UserHomeDir()
	if err == nil {
		globalCursor := filepath.Join(home, ".cursor")
		if filepath.Clean(dest) == filepath.Clean(globalCursor) {
			return "~/.cursor"
		}
	}

	base := filepath.Base(dest)
	if base == ".cursor" {
		return ".cursor"
	}

	return dest
}

// buildCursorAgentDefinitions generates the agent definitions list for Cursor.
// References agents as files in the agents/ directory.
func buildCursorAgentDefinitions(agents []model.Agent) string {
	if len(agents) == 0 {
		return ""
	}

	groups := groupAgentsByDomain(agents)
	order := orderedDomains(groups)

	var sb strings.Builder
	sb.WriteString("- **Agent Definitions:** Reusable agent prompts live in `{{BASE}}/agents/`. Cursor discovers and delegates to these subagents automatically:\n")
	for _, domain := range order {
		entries := groups[domain]
		if len(entries) == 0 {
			continue
		}

		var parts []string
		for _, a := range entries {
			part := fmt.Sprintf("`%s.md`", flattenName(a.Name))
			annotations := cursorAgentAnnotations(a)
			if annotations != "" {
				part += " " + annotations
			}
			parts = append(parts, part)
		}
		sb.WriteString(fmt.Sprintf("  - **%s:** %s\n", domain, strings.Join(parts, ", ")))
	}
	return sb.String()
}

// cursorAgentAnnotations builds a parenthetical annotation string for a Cursor agent entry.
func cursorAgentAnnotations(a model.Agent) string {
	var parts []string

	if a.Access == "read-only" {
		parts = append(parts, "read-only")
	}

	if len(parts) == 0 {
		return ""
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

// buildCursorSubagentTips generates the subagent launch tips section for Cursor.
func buildCursorSubagentTips(agents []model.Agent) string {
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
	sb.WriteString("- **Subagent Tips:**\n")

	if len(readOnlyNames) > 0 {
		sb.WriteString(fmt.Sprintf("  - Read-only agents (%s) are configured with `readonly: true` in their frontmatter.\n",
			strings.Join(readOnlyNames, ", ")))
	}

	sb.WriteString("  - Agents with `model: inherit` use the same model as the parent. Override per-agent if needed.\n")

	return sb.String()
}

// buildCursorSkillDefinitions generates the custom skills list for Cursor.
// References skills as SKILL.md file paths.
func buildCursorSkillDefinitions(skills []model.Skill) string {
	if len(skills) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("- **Workflow Skills:** Located in `{{BASE}}/skills/`. Read and follow these workflow files when the task matches:\n")
	for _, s := range skills {
		desc := skillDescription(s)
		sb.WriteString(fmt.Sprintf("  - `%s` -- %s\n", flattenName(s.Name), desc))
	}
	return sb.String()
}

// buildCursorRuleTable generates the rules table for Cursor using .mdc file names.
func buildCursorRuleTable(rules []model.Rule, cursorRuleNames map[string]string) string {
	if len(rules) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Rules are auto-applied from `{{BASE}}/rules/` based on file type and context:\n\n")
	sb.WriteString("| Rule File | Content |\n")
	sb.WriteString("|-----------|---------|")
	for _, r := range rules {
		displayName := flattenName(r.Name) + ".mdc"
		if cn, ok := cursorRuleNames[r.Name]; ok {
			displayName = cn
		}
		desc := ruleDescription(r)
		ruleFile := "rules/" + displayName
		sb.WriteString(fmt.Sprintf("\n| `%s` | %s |", ruleFile, desc))
	}
	sb.WriteString("\n\nRules with `globs` are auto-attached when matching files are referenced. Rules with `alwaysApply: true` are always active.\n")
	return sb.String()
}
