package resolver

import (
	"path/filepath"

	"github.com/semir/agent-army/internal/frontmatter"
	"github.com/semir/agent-army/internal/graph"
	"github.com/semir/agent-army/internal/model"
)

// ComputeAllFixes detects redundancies and produces auto-fixable Fix entries.
func ComputeAllFixes(
	skills []model.Skill,
	agents []model.Agent,
	root string,
) []model.Fix {
	skillLookup := buildSkillLookup(skills)
	agentLookup := buildAgentLookup(agents)

	var fixes []model.Fix

	// (a) uses_skills redundancies
	getSkillDeps := func(name string) []string { return skillLookup[name] }

	for _, s := range skills {
		if len(s.UsesSkills) == 0 {
			continue
		}
		redundancies := graph.FindRedundant(s.UsesSkills, getSkillDeps)
		if fix := redundanciesToFix(s.Path, s.Path, "uses_skills", s.UsesSkills, redundancies); fix != nil {
			fixes = append(fixes, *fix)
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

func buildSkillLookup(skills []model.Skill) map[string][]string {
	m := make(map[string][]string, len(skills))
	for _, s := range skills {
		m[s.Name] = s.UsesSkills
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
