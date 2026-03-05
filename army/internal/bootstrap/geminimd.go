package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/semir/agent-army/internal/model"
)

// generateGeminiMD reads the GEMINI.md template from templatePath and replaces
// marker sections with content generated from the selected agents, skills, and rules.
// The dest path is used to resolve the {{BASE}} placeholder.
// Note: Gemini CLI does not support plugins, so there is no plugins parameter.
func generateGeminiMD(dest, templatePath string, agents []model.Agent, skills []model.Skill, rules []model.Rule) error {
	tmplBytes, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read GEMINI.md template: %w", err)
	}
	content := string(tmplBytes)

	content = replaceMarkerSection(content, agentBeginMarker, agentEndMarker, buildGeminiAgentDefinitions(agents))
	content = replaceMarkerSection(content, tipsBeginMarker, tipsEndMarker, buildGeminiSubagentTips(agents))
	content = replaceMarkerSection(content, skillBeginMarker, skillEndMarker, buildGeminiSkillDefinitions(skills))
	content = replaceMarkerSection(content, ruleBeginMarker, ruleEndMarker, buildGeminiRuleTable(rules))

	// Collapse 3+ consecutive newlines to 2 (one blank line).
	for strings.Contains(content, "\n\n\n") {
		content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
	}

	prefix := geminiDestPrefix(dest)
	content = strings.ReplaceAll(content, basePlaceholder, prefix)

	return writeOutput(dest, "GEMINI.md", content)
}

// geminiDestPrefix determines the display path prefix for the Gemini destination directory.
func geminiDestPrefix(dest string) string {
	home, err := os.UserHomeDir()
	if err == nil {
		globalGemini := filepath.Join(home, ".gemini")
		if filepath.Clean(dest) == filepath.Clean(globalGemini) {
			return "~/.gemini"
		}
	}

	base := filepath.Base(dest)
	if base == ".gemini" {
		return ".gemini"
	}

	return dest
}

// buildGeminiAgentDefinitions generates the agent definitions list for Gemini CLI.
// References agents as files in the agents/ directory using @file syntax.
func buildGeminiAgentDefinitions(agents []model.Agent) string {
	if len(agents) == 0 {
		return ""
	}

	groups := groupAgentsByDomain(agents)
	order := orderedDomains(groups)

	var sb strings.Builder
	sb.WriteString("- **Agent Definitions:** Reusable agent prompts live in `{{BASE}}/agents/`. Spawn subagents via @file imports:\n")
	for _, domain := range order {
		entries := groups[domain]
		if len(entries) == 0 {
			continue
		}

		var parts []string
		for _, a := range entries {
			part := fmt.Sprintf("`@agents/%s.md`", flattenName(a.Name))
			annotations := geminiAgentAnnotations(a)
			if annotations != "" {
				part += " " + annotations
			}
			parts = append(parts, part)
		}
		sb.WriteString(fmt.Sprintf("  - **%s:** %s\n", domain, strings.Join(parts, ", ")))
	}
	return sb.String()
}

// geminiAgentAnnotations builds a parenthetical annotation string for a Gemini agent entry.
func geminiAgentAnnotations(a model.Agent) string {
	var parts []string

	if a.Access == "read-only" {
		parts = append(parts, "read-only")
	}

	if len(parts) == 0 {
		return ""
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

// buildGeminiSubagentTips generates the subagent launch tips section for Gemini CLI.
func buildGeminiSubagentTips(agents []model.Agent) string {
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
		sb.WriteString(fmt.Sprintf("  - Read-only agents (%s) should only use read tools (read_file, glob, search_file_content, run_shell_command).\n",
			strings.Join(readOnlyNames, ", ")))
	}

	sb.WriteString("  - Use `model: gemini-3.1-pro` for all subagent tasks. Override per-agent if needed.\n")

	return sb.String()
}

// buildGeminiSkillDefinitions generates the custom skills list for Gemini CLI.
// References skills as @file paths.
func buildGeminiSkillDefinitions(skills []model.Skill) string {
	if len(skills) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("- **Workflow Skills:** Located in `{{BASE}}/skills/`. Read and follow via @file:\n")
	for _, s := range skills {
		desc := skillDescription(s)
		sb.WriteString(fmt.Sprintf("  - `@skills/%s/SKILL.md` -- %s\n", flattenName(s.Name), desc))
	}
	return sb.String()
}

// buildGeminiRuleTable generates the rules table for Gemini CLI using @file references.
func buildGeminiRuleTable(rules []model.Rule) string {
	if len(rules) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("Detailed patterns are loaded via @file from `{{BASE}}/rules/`:\n\n")
	sb.WriteString("| Rule File | Content |\n")
	sb.WriteString("|-----------|---------|")
	for _, r := range rules {
		flat := flattenName(r.Name)
		desc := ruleDescription(r)
		ruleFile := fmt.Sprintf("@rules/%s.md", flat)
		sb.WriteString(fmt.Sprintf("\n| `%s` | %s |", ruleFile, desc))
	}
	sb.WriteString("\n\nRules are loaded via @file imports. Reference them in your prompts when relevant.\n")
	return sb.String()
}
