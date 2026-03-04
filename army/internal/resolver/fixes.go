package resolver

import (
	"path/filepath"

	"github.com/semir/agent-army/internal/frontmatter"
	"github.com/semir/agent-army/internal/graph"
	"github.com/semir/agent-army/internal/model"
)

// ComputeAllFixes detects redundancies and produces auto-fixable Fix entries.
func ComputeAllFixes(
	rules []model.Rule,
	skills []model.Skill,
	agents []model.Agent,
	root string,
) []model.Fix {
	ruleLookup := buildRuleLookup(rules)
	agentLookup := buildAgentLookup(agents)
	skillLookup := buildSkillLookup(skills)

	var fixes []model.Fix

	// (a) uses_rules redundancies
	getRuleDeps := func(name string) []string { return ruleLookup[name] }

	for _, r := range rules {
		if len(r.UsesRules) == 0 {
			continue
		}
		redundancies := graph.FindRedundant(r.UsesRules, getRuleDeps)
		if fix := redundanciesToFix(r.Path, r.Path, "uses_rules", r.UsesRules, redundancies); fix != nil {
			fixes = append(fixes, *fix)
		}
	}

	for _, s := range skills {
		if len(s.UsesRules) == 0 {
			continue
		}
		redundancies := graph.FindRedundant(s.UsesRules, getRuleDeps)
		if fix := redundanciesToFix(s.Path, s.Path, "uses_rules", s.UsesRules, redundancies); fix != nil {
			fixes = append(fixes, *fix)
		}
	}

	for _, a := range agents {
		if len(a.UsesRules) > 0 {
			redundancies := graph.FindRedundant(a.UsesRules, getRuleDeps)
			if fix := redundanciesToFix(a.Path, a.Path, "uses_rules", a.UsesRules, redundancies); fix != nil {
				fixes = append(fixes, *fix)
			}
		}
	}

	// (b) delegates_to redundancies
	getAgentDeps := func(name string) []string { return agentLookup[name] }
	for _, a := range agents {
		if len(a.DelegatesTo) == 0 {
			continue
		}
		redundancies := graph.FindRedundant(a.DelegatesTo, getAgentDeps)
		if fix := redundanciesToFix(a.Path, a.Path, "delegates_to", a.DelegatesTo, redundancies); fix != nil {
			fixes = append(fixes, *fix)
		}
	}

	// (c) skill-transitive rule redundancies (merges into existing fixes)
	for _, a := range agents {
		if len(a.UsesRules) == 0 || len(a.UsesSkills) == 0 {
			continue
		}
		redundancies := graph.FindRedundantViaSkills(a.UsesRules, a.UsesSkills, skillLookup, ruleLookup)
		if len(redundancies) == 0 {
			continue
		}
		mergeOrAppendSkillFix(a, redundancies, &fixes)
	}

	return fixes
}

// ApplyFixes writes each fix to disk.
func ApplyFixes(fixes []model.Fix, root string) error {
	for _, fix := range fixes {
		fp := filepath.Join(root, fix.FilePath)
		if err := frontmatter.WriteField(fp, fix.Field, fix.After); err != nil {
			return err
		}
	}
	return nil
}

func redundanciesToFix(label, filePath, field string, original []string, redundancies []model.Redundancy) *model.Fix {
	if len(redundancies) == 0 {
		return nil
	}

	redundantNames := make(map[string]bool)
	var reasons []string
	for _, r := range redundancies {
		redundantNames[r.Target] = true
		reasons = append(reasons, `"`+r.Target+`" covered by "`+r.CoveredBy+`"`)
	}

	var cleaned []string
	for _, entry := range original {
		if !redundantNames[entry] {
			cleaned = append(cleaned, entry)
		}
	}

	before := make([]string, len(original))
	copy(before, original)

	return &model.Fix{
		Label:    label,
		Field:    field,
		FilePath: filePath,
		Before:   before,
		After:    cleaned,
		Reasons:  reasons,
	}
}

func mergeOrAppendSkillFix(a model.Agent, redundancies []model.Redundancy, fixes *[]model.Fix) {
	redundantNames := make(map[string]bool)
	var reasons []string
	for _, r := range redundancies {
		redundantNames[r.Target] = true
		reasons = append(reasons, `"`+r.Target+`" covered by `+r.CoveredBy)
	}

	// Find existing fix for this file+field
	for i := range *fixes {
		if (*fixes)[i].FilePath == a.Path && (*fixes)[i].Field == "uses_rules" {
			var cleaned []string
			for _, e := range (*fixes)[i].After {
				if !redundantNames[e] {
					cleaned = append(cleaned, e)
				}
			}
			(*fixes)[i].After = cleaned
			(*fixes)[i].Reasons = append((*fixes)[i].Reasons, reasons...)
			return
		}
	}

	// No prior fix — create fresh one
	var cleaned []string
	for _, e := range a.UsesRules {
		if !redundantNames[e] {
			cleaned = append(cleaned, e)
		}
	}

	before := make([]string, len(a.UsesRules))
	copy(before, a.UsesRules)

	*fixes = append(*fixes, model.Fix{
		Label:    a.Path,
		Field:    "uses_rules",
		FilePath: a.Path,
		Before:   before,
		After:    cleaned,
		Reasons:  reasons,
	})
}

func buildRuleLookup(rules []model.Rule) map[string][]string {
	m := make(map[string][]string, len(rules))
	for _, r := range rules {
		m[r.Name] = r.UsesRules
	}
	return m
}

func buildSkillLookup(skills []model.Skill) map[string][]string {
	m := make(map[string][]string, len(skills))
	for _, s := range skills {
		m[s.Name] = s.UsesRules
	}
	return m
}

func buildAgentLookup(agents []model.Agent) map[string][]string {
	m := make(map[string][]string, len(agents))
	for _, a := range agents {
		m[a.Name] = a.DelegatesTo
	}
	return m
}
