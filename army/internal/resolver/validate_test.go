package resolver

import (
	"testing"

	"github.com/semir/agent-army/internal/model"
)

func TestValidateAllRefs_Valid(t *testing.T) {
	rules := []model.Rule{{Name: "security", Path: "spec/rules/security.md"}}
	skills := []model.Skill{{Name: "api", Path: "spec/skills/api.md", UsesRules: []string{"security"}}}
	agents := []model.Agent{{Name: "coder", Path: "spec/agents/coder.md", UsesSkills: []string{"api"}, UsesRules: []string{"security"}}}

	errs := ValidateAllRefs(rules, skills, agents, nil)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateAllRefs_MissingRule(t *testing.T) {
	rules := []model.Rule{{Name: "security", Path: "spec/rules/security.md"}}
	skills := []model.Skill{{Name: "api", Path: "spec/skills/api.md", UsesRules: []string{"nonexistent"}}}

	errs := ValidateAllRefs(rules, skills, nil, nil)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Severity != "error" {
		t.Errorf("severity = %q, want error", errs[0].Severity)
	}
	if errs[0].Ref != "nonexistent" {
		t.Errorf("ref = %q, want nonexistent", errs[0].Ref)
	}
}

func TestValidateAllRefs_MissingPlugin(t *testing.T) {
	agents := []model.Agent{{Name: "coder", Path: "spec/agents/coder.md", UsesPlugins: []string{"missing-plugin"}}}

	errs := ValidateAllRefs(nil, nil, agents, nil)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Severity != "warning" {
		t.Errorf("severity = %q, want warning", errs[0].Severity)
	}
}
