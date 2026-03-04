package scaffold

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/semir/agent-army/internal/loader"
	"github.com/semir/agent-army/internal/manifest"
	"github.com/semir/agent-army/internal/tui"
)

var (
	scopes          = []string{"universal", "language-specific"}
	commonLanguages = []string{"go", "python", "typescript", "react", "rust", "java"}
	agentRoles      = []string{"coder", "reviewer", "tester", "analyzer", "writer", "builder"}
	readOnlyRoles   = map[string]bool{"reviewer": true, "analyzer": true}
	accessChoices   = []string{"read-write", "read-only"}
)

// ScaffoldFlow runs the interactive scaffold flow for creating a new entity.
func ScaffoldFlow(root, entityType string, p tui.Prompter, w io.Writer) error {
	switch entityType {
	case "rule":
		return scaffoldRule(root, p, w)
	case "skill":
		return scaffoldSkill(root, p, w)
	case "agent":
		return scaffoldAgent(root, p, w)
	default:
		return fmt.Errorf("unknown entity type: %s", entityType)
	}
}

func scaffoldRule(root string, p tui.Prompter, w io.Writer) error {
	fmt.Fprintln(w, "=== New Rule ===")
	fmt.Fprintln(w)

	name, err := p.Prompt("Rule name (e.g. 'security' or 'go/testing'): ")
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)

	filePath := filepath.Join(root, "spec", "rules", name+".md")
	if checkDuplicate(w, filePath) {
		return nil
	}

	description, err := tui.PromptWithDefault(p, "Description", defaultDescription("rule", name))
	if err != nil {
		return err
	}

	scope, err := tui.SelectOneWithDefault(p, w, "Scope:", scopes, "universal")
	if err != nil {
		return err
	}

	languages, err := promptLanguages(p, w, scope)
	if err != nil {
		return err
	}

	availableRules := ruleNames(root)
	usesRules, err := promptDependencies(p, w, "uses_rules", availableRules)
	if err != nil {
		return err
	}

	fields := buildFields(map[string]interface{}{
		"name": name, "description": description, "scope": scope,
		"languages": languages, "uses_rules": usesRules,
	})
	content := generateFrontmatter(fields) + "\n" + generateRuleBody(name)

	return previewConfirmWrite(p, w, filePath, content, root)
}

func scaffoldSkill(root string, p tui.Prompter, w io.Writer) error {
	fmt.Fprintln(w, "=== New Skill ===")
	fmt.Fprintln(w)

	name, err := p.Prompt("Skill name (e.g. 'security' or 'go/testing'): ")
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)

	filePath := filepath.Join(root, "spec", "skills", name+".md")
	if checkDuplicate(w, filePath) {
		return nil
	}

	description, err := tui.PromptWithDefault(p, "Description", defaultDescription("skill", name))
	if err != nil {
		return err
	}

	scope, err := tui.SelectOneWithDefault(p, w, "Scope:", scopes, "universal")
	if err != nil {
		return err
	}

	languages, err := promptLanguages(p, w, scope)
	if err != nil {
		return err
	}

	availableRules := ruleNames(root)
	usesRules, err := promptDependencies(p, w, "uses_rules", availableRules)
	if err != nil {
		return err
	}

	fields := buildFields(map[string]interface{}{
		"name": name, "description": description, "scope": scope,
		"languages": languages, "uses_rules": usesRules,
	})
	content := generateFrontmatter(fields) + "\n" + generateSkillBody(name)

	return previewConfirmWrite(p, w, filePath, content, root)
}

func scaffoldAgent(root string, p tui.Prompter, w io.Writer) error {
	fmt.Fprintln(w, "=== New Agent ===")
	fmt.Fprintln(w)

	name, err := p.Prompt("Agent name (e.g. 'security' or 'go/testing'): ")
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)

	filePath := filepath.Join(root, "spec", "agents", name+".md")
	if checkDuplicate(w, filePath) {
		return nil
	}

	description, err := tui.PromptWithDefault(p, "Description", defaultDescription("agent", name))
	if err != nil {
		return err
	}

	role, err := tui.SelectOneWithDefault(p, w, "Role:", agentRoles, "coder")
	if err != nil {
		return err
	}

	scope, err := tui.SelectOneWithDefault(p, w, "Scope:", scopes, "universal")
	if err != nil {
		return err
	}

	languages, err := promptLanguages(p, w, scope)
	if err != nil {
		return err
	}

	defaultAccess := "read-write"
	if readOnlyRoles[role] {
		defaultAccess = "read-only"
	}
	access, err := tui.SelectOneWithDefault(p, w, "Access:", accessChoices, defaultAccess)
	if err != nil {
		return err
	}

	usesSkills, err := promptDependencies(p, w, "uses_skills", skillNames(root))
	if err != nil {
		return err
	}
	usesRules, err := promptDependencies(p, w, "uses_rules", ruleNames(root))
	if err != nil {
		return err
	}
	usesPlugins, err := promptDependencies(p, w, "uses_plugins", pluginNames(root))
	if err != nil {
		return err
	}
	delegatesTo, err := promptDependencies(p, w, "delegates_to", agentNames(root))
	if err != nil {
		return err
	}

	fields := buildFields(map[string]interface{}{
		"name": name, "description": description, "role": role, "scope": scope,
		"languages": languages, "access": access,
		"uses_skills": usesSkills, "uses_rules": usesRules,
		"uses_plugins": usesPlugins, "delegates_to": delegatesTo,
	})
	content := generateFrontmatter(fields) + "\n" + generateAgentBody(name, access)

	return previewConfirmWrite(p, w, filePath, content, root)
}

func promptLanguages(p tui.Prompter, w io.Writer, scope string) ([]string, error) {
	if scope != "language-specific" {
		return nil, nil
	}
	fmt.Fprintln(w, "\nCommon languages:")
	for i, lang := range commonLanguages {
		fmt.Fprintf(w, "  %d) %s\n", i+1, lang)
	}
	fmt.Fprintln(w)

	raw, err := p.Prompt("Languages (comma-separated numbers, or type custom names): ")
	if err != nil {
		return nil, err
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var result []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if idx, ok := parseInt(part); ok && idx >= 1 && idx <= len(commonLanguages) {
			result = append(result, commonLanguages[idx-1])
		} else {
			result = append(result, part)
		}
	}
	return result, nil
}

func parseInt(s string) (int, bool) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	if len(s) == 0 {
		return 0, false
	}
	return n, true
}

func promptDependencies(p tui.Prompter, w io.Writer, field string, available []string) ([]string, error) {
	if len(available) == 0 {
		return nil, nil
	}
	fmt.Fprintf(w, "\nAvailable %s:\n", field)
	result, err := tui.SelectMultiOptional(p, w, fmt.Sprintf("Select %s", field), available)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func checkDuplicate(w io.Writer, filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		fmt.Fprintf(w, "File already exists: %s\n", filePath)
		fmt.Fprintln(w, "Aborted. Use 'agent-army edit' to modify existing entities.")
		return true
	}
	return false
}

func defaultDescription(entityType, name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool { return r == '/' || r == '-' })
	var titled []string
	for _, p := range parts {
		titled = append(titled, strings.Title(p))
	}
	title := strings.Join(titled, " ")

	suffixes := map[string]string{
		"rule":  "patterns and conventions",
		"skill": "workflow and decision tree",
		"agent": "specialist agent",
	}
	return strings.TrimSpace(title + " " + suffixes[entityType])
}

func nameToTitle(name string) string {
	parts := strings.FieldsFunc(name, func(r rune) bool { return r == '/' || r == '-' })
	var titled []string
	for _, p := range parts {
		titled = append(titled, strings.Title(p))
	}
	return strings.Join(titled, " ")
}

type fieldEntry struct {
	key   string
	value interface{}
}

func buildFields(m map[string]interface{}) []fieldEntry {
	order := []string{"name", "description", "role", "scope", "languages", "access",
		"uses_skills", "uses_rules", "uses_plugins", "delegates_to"}
	var fields []fieldEntry
	for _, k := range order {
		if v, ok := m[k]; ok {
			fields = append(fields, fieldEntry{k, v})
		}
	}
	return fields
}

func generateFrontmatter(fields []fieldEntry) string {
	lines := []string{"---"}
	for _, f := range fields {
		switch v := f.value.(type) {
		case string:
			if strings.Contains(v, ":") {
				lines = append(lines, fmt.Sprintf("%s: %q", f.key, v))
			} else {
				lines = append(lines, fmt.Sprintf("%s: %s", f.key, v))
			}
		case []string:
			if len(v) == 0 {
				lines = append(lines, fmt.Sprintf("%s: []", f.key))
			} else {
				lines = append(lines, fmt.Sprintf("%s: [%s]", f.key, strings.Join(v, ", ")))
			}
		case nil:
			lines = append(lines, fmt.Sprintf("%s: []", f.key))
		}
	}
	lines = append(lines, "---")
	return strings.Join(lines, "\n")
}

func generateRuleBody(name string) string {
	title := nameToTitle(name)
	return fmt.Sprintf(`
# %s Patterns

## Overview

<!-- Describe the purpose and scope of these patterns. -->

## Patterns

<!-- List the key patterns, conventions, and best practices. -->

## Anti-Patterns

<!-- List common mistakes and what to do instead. -->
`, title)
}

func generateSkillBody(name string) string {
	title := nameToTitle(name)
	return fmt.Sprintf(`
# %s

## When to Use

<!-- Describe when this skill should be invoked. -->

## Workflow

<!-- Step-by-step workflow for this skill. -->

## Decision Tree

<!-- Decision tree or flowchart for key choices. -->

## Checklist

<!-- Pre-completion checklist items. -->
`, title)
}

func generateAgentBody(name, access string) string {
	title := nameToTitle(name)
	var capabilities string
	if access == "read-only" {
		capabilities = `- Read source files, configuration, and documentation
- Search for patterns, imports, and dependencies
- Run read-only analysis commands
- Cannot modify any files`
	} else {
		capabilities = `- Read and write source files
- Run build, test, and lint commands
- Create new files and directories
- Modify existing code following project patterns`
	}

	return fmt.Sprintf(`
# %s Agent

## Role

<!-- Describe the agent's role and expertise. -->

## Activation

<!-- When does the orchestrator activate this agent? -->

## Capabilities

%s

## Standards

<!-- Key standards this agent enforces or follows. -->

## Workflow

<!-- Step-by-step workflow the agent follows. -->

## Output Format

<!-- Describe the expected output format. -->

## Constraints

<!-- Hard constraints the agent must never violate. -->
`, title, capabilities)
}

func previewConfirmWrite(p tui.Prompter, w io.Writer, filePath, content, root string) error {
	rel, _ := filepath.Rel(root, filePath)
	fmt.Fprintln(w, "\n--- Preview ---")
	fmt.Fprint(w, content)
	fmt.Fprintln(w, "--- End Preview ---")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "File: %s\n", rel)

	confirm, err := p.Prompt("Create this file? [y/N] ")
	if err != nil {
		fmt.Fprintln(w, "\nAborted.")
		return nil
	}
	if strings.TrimSpace(strings.ToLower(confirm)) != "y" {
		fmt.Fprintln(w, "Aborted. No files created.")
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	tmpPath := filePath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, filePath); err != nil {
		return err
	}

	fmt.Fprintf(w, "Created %s\n", rel)
	fmt.Fprintln(w, "Regenerating manifest.json...")
	if err := manifest.WriteManifest(root); err != nil {
		fmt.Fprintf(w, "(manifest module not available: %v)\n", err)
	}
	return nil
}

func ruleNames(root string) []string {
	rules, _ := loader.LoadRules(root)
	names := make([]string, len(rules))
	for i, r := range rules {
		names[i] = r.Name
	}
	return names
}

func skillNames(root string) []string {
	skills, _ := loader.LoadSkills(root)
	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	return names
}

func agentNames(root string) []string {
	agents, _ := loader.LoadAgents(root)
	names := make([]string, len(agents))
	for i, a := range agents {
		names[i] = a.Name
	}
	return names
}

func pluginNames(root string) []string {
	plugins, _ := loader.LoadPlugins(root)
	return plugins
}
