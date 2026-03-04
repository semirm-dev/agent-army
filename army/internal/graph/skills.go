package graph

import (
	"github.com/semir/agent-army/internal/model"
)

// RulesCoveredBySkills computes the transitive closure of all rules covered by skills.
func RulesCoveredBySkills(skillNames []string, skillLookup, ruleLookup map[string][]string) map[string]bool {
	var union []string
	for _, skill := range skillNames {
		union = append(union, skillLookup[skill]...)
	}

	if len(union) == 0 {
		return nil
	}

	resolved := ResolveTransitive(union, func(name string) []string {
		return ruleLookup[name]
	})

	result := make(map[string]bool, len(resolved))
	for _, r := range resolved {
		result[r] = true
	}
	return result
}

// FindRedundantViaSkills finds rules that are already covered transitively by skills.
func FindRedundantViaSkills(
	ruleEntries, skillEntries []string,
	skillLookup, ruleLookup map[string][]string,
) []model.Redundancy {
	if len(ruleEntries) == 0 || len(skillEntries) == 0 {
		return nil
	}

	covered := RulesCoveredBySkills(skillEntries, skillLookup, ruleLookup)
	if len(covered) == 0 {
		return nil
	}

	var redundancies []model.Redundancy

	for _, rule := range ruleEntries {
		if !covered[rule] {
			continue
		}
		for _, skill := range skillEntries {
			skillRules := skillLookup[skill]
			if len(skillRules) == 0 {
				continue
			}
			skillClosure := ResolveTransitive(skillRules, func(name string) []string {
				return ruleLookup[name]
			})
			if contains(skillClosure, rule) {
				redundancies = append(redundancies, model.Redundancy{
					Target:    rule,
					CoveredBy: "skill " + skill,
				})
				break
			}
		}
	}

	return redundancies
}
