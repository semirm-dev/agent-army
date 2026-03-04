package graph

import (
	"testing"
)

func TestRulesCoveredBySkills_Empty(t *testing.T) {
	got := RulesCoveredBySkills(nil, nil, nil)
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestRulesCoveredBySkills_Basic(t *testing.T) {
	skillLookup := map[string][]string{
		"skill-a": {"rule1", "rule2"},
	}
	ruleLookup := map[string][]string{
		"rule1": {"rule3"},
		"rule2": {},
		"rule3": {},
	}
	got := RulesCoveredBySkills([]string{"skill-a"}, skillLookup, ruleLookup)

	for _, want := range []string{"rule1", "rule2", "rule3"} {
		if !got[want] {
			t.Errorf("missing %q in coverage", want)
		}
	}
}

func TestFindRedundantViaSkills_Empty(t *testing.T) {
	got := FindRedundantViaSkills(nil, nil, nil, nil)
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestFindRedundantViaSkills_Covered(t *testing.T) {
	skillLookup := map[string][]string{
		"skill-a": {"rule1"},
	}
	ruleLookup := map[string][]string{
		"rule1": {"rule2"},
		"rule2": {},
	}
	got := FindRedundantViaSkills(
		[]string{"rule2"},
		[]string{"skill-a"},
		skillLookup,
		ruleLookup,
	)
	if len(got) != 1 {
		t.Fatalf("got %d, want 1", len(got))
	}
	if got[0].Target != "rule2" {
		t.Errorf("target = %q, want rule2", got[0].Target)
	}
	if got[0].CoveredBy != "skill skill-a" {
		t.Errorf("coveredBy = %q, want 'skill skill-a'", got[0].CoveredBy)
	}
}

func TestFindRedundantViaSkills_NotCovered(t *testing.T) {
	skillLookup := map[string][]string{
		"skill-a": {"rule1"},
	}
	ruleLookup := map[string][]string{
		"rule1": {},
		"rule3": {},
	}
	got := FindRedundantViaSkills(
		[]string{"rule3"},
		[]string{"skill-a"},
		skillLookup,
		ruleLookup,
	)
	if len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}
