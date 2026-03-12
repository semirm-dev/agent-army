package bootstrap

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/semir/agent-army/internal/model"
)

// buildResolvedDeps resolves all dependencies for an agent into full model objects.
func buildResolvedDeps(
	agent model.Agent,
	skillMap map[string]model.Skill,
	agentMap map[string]model.Agent,
) model.ResolvedDeps {
	var deps model.ResolvedDeps

	// Resolve skills
	for _, name := range agent.UsesSkills {
		if s, ok := skillMap[name]; ok {
			deps.Skills = append(deps.Skills, s)
		}
	}

	// Pass through plugins
	deps.Plugins = agent.UsesPlugins

	// Resolve delegates
	for _, name := range agent.DelegatesTo {
		if a, ok := agentMap[name]; ok {
			deps.DelegatesTo = append(deps.DelegatesTo, a)
		}
	}

	return deps
}

// enrichAgentBody injects a "Resources Available" section into the agent body
// and rewrites body text references to be target-native.
func enrichAgentBody(body string, deps model.ResolvedDeps, target string) string {
	resources := buildResourcesSection(deps, target)
	body = injectSection(body, resources)
	body = rewriteBodyRefs(body, target)
	return body
}

// buildResourcesSection generates the "## Resources Available" markdown section.
func buildResourcesSection(deps model.ResolvedDeps, target string) string {
	var sb strings.Builder

	sb.WriteString("## Resources Available\n")

	// Skills section
	if len(deps.Skills) > 0 {
		sb.WriteString("\n")
		switch target {
		case TargetClaude:
			sb.WriteString("### Skills (Invoke via Skill Tool)\n")
			sb.WriteString("Use the Skill tool to invoke these when the task matches:\n")
			for _, s := range deps.Skills {
				sb.WriteString(fmt.Sprintf("- `%s` -- %s\n", flattenName(s.Name), skillDescription(s)))
			}
		case TargetCursor:
			sb.WriteString("### Workflow References\n")
			sb.WriteString("Read and follow these workflow files when the task matches:\n")
			for _, s := range deps.Skills {
				sb.WriteString(fmt.Sprintf("- `skills/%s/SKILL.md` -- %s\n", flattenName(s.Name), skillDescription(s)))
			}
		}
	}

	// Plugins section (Claude only)
	if target == TargetClaude && len(deps.Plugins) > 0 {
		sb.WriteString("\n### Plugins\n")
		for _, p := range deps.Plugins {
			sb.WriteString(fmt.Sprintf("- `%s`\n", p))
		}
	}

	// Delegate agents section (Claude only)
	if target == TargetClaude && len(deps.DelegatesTo) > 0 {
		sb.WriteString("\n### Delegate Agents (Invoke via Task Tool)\n")
		sb.WriteString("Use the Task tool with these agent files when delegation is needed:\n")
		for _, a := range deps.DelegatesTo {
			sb.WriteString(fmt.Sprintf("- `%s` -- %s\n", flattenName(a.Name), a.Description))
		}
	}

	return sb.String()
}

// injectSection inserts the resources section before ## Workflow or ## Constraints,
// or appends it at the end if neither is found.
func injectSection(body string, section string) string {
	markers := []string{"## Workflow", "## Constraints"}
	for _, marker := range markers {
		idx := strings.Index(body, marker)
		if idx > 0 {
			return body[:idx] + section + "\n" + body[idx:]
		}
	}
	return body + "\n" + section
}

// Regex patterns for rewriting body text references.
var (
	// Matches: invoke the `skill-name` skill
	invokeSkillRe = regexp.MustCompile("(?i)invoke the `([^`]+)` skill")
	// Matches: invoke the `skill-name` skill for ...
	invokeSkillForRe = regexp.MustCompile("(?i)invoke the `([^`]+)` skill (for [^.\\n]+)")
	// Matches: Delegate to `agent-name`
	delegateToRe = regexp.MustCompile("(?i)delegate to `([^`]+)`")
	// Matches: via the Skill tool
	viaSkillToolRe = regexp.MustCompile("(?i) via the Skill tool")
	// Matches: via the `Skill` tool
	viaSkillToolBacktickRe = regexp.MustCompile("(?i) via the `Skill` tool")
	// Matches: loaded via the `skill-name` skill
	loadedViaSkillRe = regexp.MustCompile("(?i)loaded via the `([^`]+)` skill")
	// Matches: loaded via skills (generic)
	loadedViaSkillsRe = regexp.MustCompile("(?i)loaded via skills")
	// Matches: loaded via the `rule-name` rule
	loadedViaRuleRe = regexp.MustCompile("(?i)loaded via the `([^`]+)` rule")
)

// rewriteBodyRefs rewrites instructional references in the body text to be target-native.
func rewriteBodyRefs(body string, target string) string {
	switch target {
	case TargetClaude:
		// Claude body text is already native — no changes needed
		return body
	case TargetCursor:
		body = invokeSkillForRe.ReplaceAllString(body, "read and follow the workflow in `skills/$1/SKILL.md` $2")
		body = invokeSkillRe.ReplaceAllString(body, "read and follow the workflow in `skills/$1/SKILL.md`")
		body = loadedViaSkillRe.ReplaceAllString(body, "defined in `skills/$1/SKILL.md`")
		body = loadedViaSkillsRe.ReplaceAllString(body, "defined in the workflow files listed under Resources Available")
		body = loadedViaRuleRe.ReplaceAllString(body, "defined in the `$1` rule")
		body = delegateToRe.ReplaceAllString(body, "read the review checklist in `agents/$1.md`")
		body = viaSkillToolRe.ReplaceAllString(body, "")
		body = viaSkillToolBacktickRe.ReplaceAllString(body, "")
	}
	return body
}
