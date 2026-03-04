package bootstrap

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/semir/agent-army/internal/graph"
	"github.com/semir/agent-army/internal/loader"
	"github.com/semir/agent-army/internal/model"
	"github.com/semir/agent-army/internal/tui"
)

var targets = []string{"Claude Code", "Cursor"}

// MainBootstrap runs the interactive bootstrap flow.
func MainBootstrap(root string, p tui.Prompter, w io.Writer) error {
	fmt.Fprintln(w, "=== Bootstrap ===")
	fmt.Fprintln(w)

	target, err := selectTarget(p, w)
	if err != nil {
		return err
	}

	dest, err := selectDestination(p, w, target)
	if err != nil {
		return err
	}

	rules, err := loader.LoadRules(root)
	if err != nil {
		return err
	}
	skills, err := loader.LoadSkills(root)
	if err != nil {
		return err
	}
	agents, err := loader.LoadAgents(root)
	if err != nil {
		return err
	}

	allRuleNames := names(rules, func(r model.Rule) string { return r.Name })
	allSkillNames := names(skills, func(s model.Skill) string { return s.Name })
	allSkillSet := makeSet(allSkillNames)
	ruleLookup := make(map[string][]string)
	for _, r := range rules {
		ruleLookup[r.Name] = r.UsesRules
	}

	// Select agents
	selectedAgentNames, err := selectEntities(p, w, "agents", names(agents, func(a model.Agent) string { return a.Name }))
	if err != nil {
		return err
	}
	agentObjs := filterAgents(agents, selectedAgentNames)

	// Auto-compute skills
	var autoSkillNames []string
	seenSkills := make(map[string]bool)
	for _, a := range agentObjs {
		for _, s := range a.UsesSkills {
			if !seenSkills[s] && allSkillSet[s] {
				autoSkillNames = append(autoSkillNames, s)
				seenSkills[s] = true
			}
		}
	}

	finalSkillNames, err := selectAdditionalEntities(p, w, "skills", autoSkillNames, allSkillNames)
	if err != nil {
		return err
	}

	// Auto-compute rules transitively
	var ruleSeeds []string
	seenRules := make(map[string]bool)
	skillObjs := filterSkills(skills, finalSkillNames)
	for _, s := range skillObjs {
		for _, r := range s.UsesRules {
			if !seenRules[r] {
				ruleSeeds = append(ruleSeeds, r)
				seenRules[r] = true
			}
		}
	}
	for _, a := range agentObjs {
		for _, r := range a.UsesRules {
			if !seenRules[r] {
				ruleSeeds = append(ruleSeeds, r)
				seenRules[r] = true
			}
		}
	}

	existingRuleSet := makeSet(allRuleNames)
	resolved := graph.ResolveTransitive(ruleSeeds, func(name string) []string {
		return ruleLookup[name]
	})
	var autoRuleNames []string
	for _, r := range resolved {
		if existingRuleSet[r] {
			autoRuleNames = append(autoRuleNames, r)
		}
	}

	finalRuleNames, err := selectAdditionalEntities(p, w, "rules", autoRuleNames, allRuleNames)
	if err != nil {
		return err
	}

	ruleObjs := filterRules(rules, finalRuleNames)

	total := len(ruleObjs) + len(skillObjs) + len(agentObjs)
	if total == 0 {
		fmt.Fprintln(w, "\nNo entities selected. Nothing to generate.")
		return nil
	}

	fmt.Fprintln(w, "\n--- Preview ---")
	fmt.Fprintf(w, "  Target:      %s\n", target)
	fmt.Fprintf(w, "  Destination: %s\n", dest)
	fmt.Fprintf(w, "  Rules:       %d files\n", len(ruleObjs))
	fmt.Fprintf(w, "  Skills:      %d files\n", len(skillObjs))
	fmt.Fprintf(w, "  Agents:      %d files\n", len(agentObjs))
	fmt.Fprintf(w, "  Total:       %d files\n", total)
	fmt.Fprintln(w)

	confirm, err := p.Prompt("Proceed? [y/N] ")
	if err != nil {
		return err
	}
	if strings.TrimSpace(strings.ToLower(confirm)) != "y" {
		fmt.Fprintln(w, "Aborted. No files written.")
		return nil
	}

	isClaude := target == "Claude Code"
	written, err := generateAll(root, dest, ruleObjs, skillObjs, agentObjs, isClaude)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "\nDone. %d files written to %s\n", written, dest)
	return nil
}

func selectTarget(p tui.Prompter, w io.Writer) (string, error) {
	fmt.Fprintln(w, "Step 1 — Target AI model/tool:")
	for i, t := range targets {
		fmt.Fprintf(w, "  %d) %s\n", i+1, t)
	}
	fmt.Fprintln(w)

	for {
		raw, err := p.Prompt("Select target: ")
		if err != nil {
			return "", err
		}
		idx, err := strconv.Atoi(strings.TrimSpace(raw))
		if err == nil && idx >= 1 && idx <= len(targets) {
			return targets[idx-1], nil
		}
		fmt.Fprintf(w, "Invalid choice. Enter 1-%d.\n", len(targets))
	}
}

func selectDestination(p tui.Prompter, w io.Writer, target string) (string, error) {
	suffix := ".claude"
	if target != "Claude Code" {
		suffix = ".cursor"
	}

	cwd, _ := os.Getwd()
	local := filepath.Join(cwd, suffix)
	home, _ := os.UserHomeDir()
	globalHome := filepath.Join(home, suffix)

	fmt.Fprintln(w, "\nStep 2 — Output destination:")
	fmt.Fprintf(w, "  1) Local project (%s)  (*)\n", local)
	fmt.Fprintf(w, "  2) Global (%s)\n", globalHome)
	fmt.Fprintln(w, "  3) Custom directory")
	fmt.Fprintln(w)

	for {
		raw, err := p.Prompt("Select destination [1]: ")
		if err != nil {
			return "", err
		}
		raw = strings.TrimSpace(raw)
		if raw == "" || raw == "1" {
			return local, nil
		}
		if raw == "2" {
			return globalHome, nil
		}
		if raw == "3" {
			custom, err := p.Prompt("Enter path (absolute or relative): ")
			if err != nil {
				return "", err
			}
			custom = strings.TrimSpace(custom)
			if custom == "" {
				fmt.Fprintln(w, "Path cannot be empty.")
				continue
			}
			if !filepath.IsAbs(custom) {
				custom = filepath.Join(cwd, custom)
			}
			return custom, nil
		}
		fmt.Fprintln(w, "Invalid choice. Enter 1, 2, or 3.")
	}
}

func selectEntities(p tui.Prompter, w io.Writer, entityType string, entityNames []string) ([]string, error) {
	if len(entityNames) == 0 {
		return nil, nil
	}

	fmt.Fprintf(w, "\nAvailable %s (%d):\n", entityType, len(entityNames))
	for i, name := range entityNames {
		fmt.Fprintf(w, "  %d) %s\n", i+1, name)
	}
	fmt.Fprintln(w)

	for {
		raw, err := p.Prompt(fmt.Sprintf("Select %s (comma-separated, Enter for all, 'none' to skip): ", entityType))
		if err != nil {
			return nil, err
		}
		raw = strings.TrimSpace(raw)

		if raw == "" {
			return append([]string{}, entityNames...), nil
		}
		if strings.ToLower(raw) == "none" {
			return nil, nil
		}

		parts := strings.Split(raw, ",")
		var selected []string
		valid := true
		for _, part := range parts {
			idx, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil || idx < 1 || idx > len(entityNames) {
				fmt.Fprintf(w, "Invalid number: %s\n", strings.TrimSpace(part))
				valid = false
				break
			}
			selected = append(selected, entityNames[idx-1])
		}
		if valid && len(selected) > 0 {
			return selected, nil
		}
	}
}

func selectAdditionalEntities(p tui.Prompter, w io.Writer, entityType string, autoNames, allNames []string) ([]string, error) {
	autoSet := makeSet(autoNames)
	var remaining []string
	for _, n := range allNames {
		if !autoSet[n] {
			remaining = append(remaining, n)
		}
	}

	if len(autoNames) == 0 && len(remaining) == 0 {
		return nil, nil
	}

	if len(autoNames) > 0 && len(remaining) == 0 {
		fmt.Fprintf(w, "\n  Auto-included %s: %s\n", entityType, strings.Join(autoNames, ", "))
		fmt.Fprintf(w, "  All available %s are already included.\n", entityType)
		return append([]string{}, autoNames...), nil
	}

	if len(autoNames) > 0 {
		fmt.Fprintf(w, "\n  Auto-included %s: %s\n", entityType, strings.Join(autoNames, ", "))
	} else {
		fmt.Fprintf(w, "\n  No auto-included %s.\n", entityType)
	}

	fmt.Fprintf(w, "\n  Additional %s available:\n", entityType)
	for i, name := range remaining {
		fmt.Fprintf(w, "    %d) %s\n", i+1, name)
	}
	fmt.Fprintln(w)

	prompt := fmt.Sprintf("Add extra %s? (comma-separated, Enter for none, 'all' for all): ", entityType)
	if len(autoNames) == 0 {
		prompt = fmt.Sprintf("Select %s? (comma-separated, Enter for none, 'all' for all): ", entityType)
	}

	for {
		raw, err := p.Prompt(prompt)
		if err != nil {
			return nil, err
		}
		raw = strings.TrimSpace(raw)

		if raw == "" {
			return append([]string{}, autoNames...), nil
		}
		if strings.ToLower(raw) == "all" {
			result := append([]string{}, autoNames...)
			return append(result, remaining...), nil
		}

		parts := strings.Split(raw, ",")
		selected := append([]string{}, autoNames...)
		valid := true
		for _, part := range parts {
			idx, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil || idx < 1 || idx > len(remaining) {
				fmt.Fprintf(w, "Invalid number: %s\n", strings.TrimSpace(part))
				valid = false
				break
			}
			selected = append(selected, remaining[idx-1])
		}
		if valid {
			return selected, nil
		}
	}
}

func generateAll(root, dest string, rules []model.Rule, skills []model.Skill, agents []model.Agent, isClaude bool) (int, error) {
	written := 0

	if len(rules) > 0 {
		if isClaude {
			for _, r := range rules {
				content, err := ruleToClaude(root, r)
				if err != nil {
					return written, err
				}
				rel := filepath.Join("rules", flattenName(r.Name)+".md")
				if err := writeOutput(dest, rel, content); err != nil {
					return written, err
				}
				written++
			}
		} else {
			assignments := assignCursorNumbers(rules)
			for i, r := range rules {
				content, err := ruleToCursor(root, r)
				if err != nil {
					return written, err
				}
				rel := filepath.Join("rules", fmt.Sprintf("%d-%s.mdc", assignments[i].Number, assignments[i].ShortName))
				rel = resolveCollision(dest, rel)
				if err := writeOutput(dest, rel, content); err != nil {
					return written, err
				}
				written++
			}
		}
	}

	for _, s := range skills {
		flat := flattenName(s.Name)
		content, err := readFileContent(root, s.Path)
		if err != nil {
			return written, err
		}
		rel := filepath.Join("skills", flat, "SKILL.md")
		if err := writeOutput(dest, rel, content); err != nil {
			return written, err
		}
		written++
	}

	for _, a := range agents {
		flat := flattenName(a.Name)
		var content string
		var err error
		if isClaude {
			content, err = agentToClaude(root, a)
		} else {
			content, err = agentToCursor(root, a)
		}
		if err != nil {
			return written, err
		}
		rel := filepath.Join("agents", flat+".md")
		if err := writeOutput(dest, rel, content); err != nil {
			return written, err
		}
		written++
	}

	return written, nil
}

func names[T any](items []T, key func(T) string) []string {
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = key(item)
	}
	return result
}

func makeSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, s := range items {
		m[s] = true
	}
	return m
}

func filterRules(rules []model.Rule, selected []string) []model.Rule {
	set := makeSet(selected)
	var result []model.Rule
	for _, r := range rules {
		if set[r.Name] {
			result = append(result, r)
		}
	}
	return result
}

func filterSkills(skills []model.Skill, selected []string) []model.Skill {
	set := makeSet(selected)
	var result []model.Skill
	for _, s := range skills {
		if set[s.Name] {
			result = append(result, s)
		}
	}
	return result
}

func filterAgents(agents []model.Agent, selected []string) []model.Agent {
	set := makeSet(selected)
	var result []model.Agent
	for _, a := range agents {
		if set[a.Name] {
			result = append(result, a)
		}
	}
	return result
}
