package editor

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/semir/agent-army/internal/frontmatter"
	"github.com/semir/agent-army/internal/graph"
	"github.com/semir/agent-army/internal/loader"
	"github.com/semir/agent-army/internal/manifest"
	"github.com/semir/agent-army/internal/model"
	"github.com/semir/agent-army/internal/tui"
)

var entityDirs = map[string]string{
	"rule":  "spec/rules",
	"skill": "spec/skills",
	"agent": "spec/agents",
}

var entityFields = map[string][]string{
	"rule":  {"uses_rules"},
	"skill": {"uses_rules"},
	"agent": {"uses_rules", "uses_skills", "uses_plugins", "delegates_to"},
}

// EditFlow runs the full interactive dependency editor.
func EditFlow(root string, p tui.Prompter, w io.Writer) error {
	fmt.Fprintln(w, "=== Edit Dependencies ===")

	entityType, err := tui.SelectOne(p, w, "Choose entity type:", []string{"rule", "skill", "agent"})
	if err != nil {
		return err
	}

	entityDir := filepath.Join(root, entityDirs[entityType])
	filePath, displayName, err := chooseFile(p, w, entityDir)
	if err != nil {
		return err
	}

	field, err := chooseField(p, w, entityType)
	if err != nil {
		return err
	}

	prefix := entityDirs[entityType]
	fmt.Fprintf(w, "\nFile: %s/%s\n", prefix, displayName)
	current := readCurrentValues(filePath, field)
	printCurrentValues(w, field, current)

	action, err := tui.SelectOne(p, w, "Action:", []string{"add", "remove"})
	if err != nil {
		return err
	}

	newValues, err := applyAction(p, w, action, field, current, root)
	if err != nil {
		return err
	}
	if newValues == nil {
		return nil
	}

	checkRedundancy(w, field, newValues, entityType, filePath, root)
	printDiffPreview(w, current, newValues)

	confirm, err := p.Prompt("Apply this change? [y/N] ")
	if err != nil {
		return err
	}
	if strings.TrimSpace(strings.ToLower(confirm)) != "y" {
		fmt.Fprintln(w, "Aborted.")
		return nil
	}

	if err := frontmatter.WriteField(filePath, field, newValues); err != nil {
		return err
	}
	fmt.Fprintf(w, "Updated %s/%s\n", prefix, displayName)

	fmt.Fprintln(w, "\nRegenerating manifest.json...")
	if err := manifest.WriteManifest(root); err != nil {
		fmt.Fprintf(w, "(manifest generation failed: %v)\n", err)
	}

	return nil
}

func chooseFile(p tui.Prompter, w io.Writer, entityDir string) (string, string, error) {
	mdFiles, err := loader.FindMDFiles(entityDir)
	if err != nil {
		return "", "", err
	}
	if len(mdFiles) == 0 {
		return "", "", fmt.Errorf("no files found in %s", entityDir)
	}

	displayNames := make([]string, len(mdFiles))
	for i, f := range mdFiles {
		rel, _ := filepath.Rel(entityDir, f)
		displayNames[i] = rel
	}

	chosen, err := tui.SelectOne(p, w, "Choose file:", displayNames)
	if err != nil {
		return "", "", err
	}
	return filepath.Join(entityDir, chosen), chosen, nil
}

func chooseField(p tui.Prompter, w io.Writer, entityType string) (string, error) {
	fields := entityFields[entityType]
	if len(fields) == 1 {
		fmt.Fprintf(w, "\nField: %s (auto-selected)\n", fields[0])
		return fields[0], nil
	}
	return tui.SelectOne(p, w, "Choose field:", fields)
}

func readCurrentValues(filePath, field string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	fm := frontmatter.ParseFrontmatter(string(content))
	return fm.ListVal(field)
}

func printCurrentValues(w io.Writer, field string, current []string) {
	if len(current) == 0 {
		fmt.Fprintf(w, "Current %s: (none)\n", field)
	} else {
		fmt.Fprintf(w, "Current %s: [%s]\n", field, strings.Join(current, ", "))
	}
}

func applyAction(p tui.Prompter, w io.Writer, action, field string, current []string, root string) ([]string, error) {
	if action == "add" {
		return actionAdd(p, w, field, current, root)
	}
	return actionRemove(p, w, field, current)
}

func actionAdd(p tui.Prompter, w io.Writer, field string, current []string, root string) ([]string, error) {
	valid := loadValidValues(field, root)
	currentSet := make(map[string]bool)
	for _, c := range current {
		currentSet[c] = true
	}

	var available []string
	for _, v := range valid {
		if !currentSet[v] {
			available = append(available, v)
		}
	}

	if len(available) == 0 {
		fmt.Fprintln(w, "All values are already present. Nothing to add.")
		return nil, nil
	}

	selected, err := tui.SelectMulti(p, w, "Select entries to add", available)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(current))
	copy(result, current)
	return append(result, selected...), nil
}

func actionRemove(p tui.Prompter, w io.Writer, field string, current []string) ([]string, error) {
	if len(current) == 0 {
		fmt.Fprintf(w, "No %s entries to remove.\n", field)
		return nil, nil
	}

	toRemove, err := tui.SelectMulti(p, w, "Select entries to remove", current)
	if err != nil {
		return nil, err
	}
	removeSet := make(map[string]bool)
	for _, r := range toRemove {
		removeSet[r] = true
	}
	var result []string
	for _, v := range current {
		if !removeSet[v] {
			result = append(result, v)
		}
	}
	return result, nil
}

func loadValidValues(field, root string) []string {
	switch field {
	case "uses_rules":
		rules, _ := loader.LoadRules(root)
		names := make([]string, len(rules))
		for i, r := range rules {
			names[i] = r.Name
		}
		return names
	case "uses_skills":
		skills, _ := loader.LoadSkills(root)
		names := make([]string, len(skills))
		for i, s := range skills {
			names[i] = s.Name
		}
		return names
	case "uses_plugins":
		plugins, _ := loader.LoadPlugins(root)
		return plugins
	case "delegates_to":
		agents, _ := loader.LoadAgents(root)
		names := make([]string, len(agents))
		for i, a := range agents {
			names[i] = a.Name
		}
		return names
	}
	return nil
}

func checkRedundancy(w io.Writer, field string, newValues []string, entityType, filePath, root string) {
	if len(newValues) == 0 {
		return
	}
	if field == "uses_rules" {
		checkRuleRedundancy(w, newValues, entityType, filePath, root)
	} else if field == "delegates_to" {
		checkDelegateRedundancy(w, newValues, root)
	}
}

func checkRuleRedundancy(w io.Writer, newValues []string, entityType, filePath, root string) {
	rules, _ := loader.LoadRules(root)
	ruleLookup := make(map[string][]string)
	for _, r := range rules {
		ruleLookup[r.Name] = r.UsesRules
	}

	redundancies := graph.FindRedundant(newValues, func(name string) []string {
		return ruleLookup[name]
	})
	printRedundancyWarnings(w, redundancies, "rule-to-rule")

	if entityType != "agent" {
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return
	}
	fm := frontmatter.ParseFrontmatter(string(content))
	agentSkills := fm.ListVal("uses_skills")
	if len(agentSkills) == 0 {
		return
	}

	skills, _ := loader.LoadSkills(root)
	skillLookup := make(map[string][]string)
	for _, s := range skills {
		skillLookup[s.Name] = s.UsesRules
	}

	skillRedundancies := graph.FindRedundantViaSkills(newValues, agentSkills, skillLookup, ruleLookup)
	printRedundancyWarnings(w, skillRedundancies, "covered via skills")
}

func checkDelegateRedundancy(w io.Writer, newValues []string, root string) {
	agents, _ := loader.LoadAgents(root)
	agentLookup := make(map[string][]string)
	for _, a := range agents {
		agentLookup[a.Name] = a.DelegatesTo
	}

	redundancies := graph.FindRedundant(newValues, func(name string) []string {
		return agentLookup[name]
	})
	printRedundancyWarnings(w, redundancies, "delegate")
}

func printRedundancyWarnings(w io.Writer, redundancies []model.Redundancy, label string) {
	if len(redundancies) == 0 {
		return
	}
	fmt.Fprintf(w, "\nWarning: Redundant entries detected (%s):\n", label)
	for _, r := range redundancies {
		fmt.Fprintf(w, "  - %q is already included transitively by %q\n", r.Target, r.CoveredBy)
	}
}

func printDiffPreview(w io.Writer, before, after []string) {
	fmt.Fprintln(w, "\n--- Change Preview ---")
	if len(before) == 0 {
		fmt.Fprintln(w, "  Before: (none)")
	} else {
		fmt.Fprintf(w, "  Before: [%s]\n", strings.Join(before, ", "))
	}
	if len(after) == 0 {
		fmt.Fprintln(w, "  After:  (none -- field will be cleared)")
	} else {
		fmt.Fprintf(w, "  After:  [%s]\n", strings.Join(after, ", "))
	}
	fmt.Fprintln(w)
}
