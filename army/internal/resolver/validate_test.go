package resolver

import (
	"testing"

	"github.com/semir/agent-army/internal/model"
)

func TestValidateAllRefs_Valid(t *testing.T) {
	skills := []model.Skill{
		{Name: "api", Path: "spec/skills/api.md", UsesSkills: []string{"error-handling"}},
		{Name: "error-handling", Path: "spec/skills/error-handling.md"},
	}
	agents := []model.Agent{{Name: "coder", Path: "spec/agents/coder.md", UsesSkills: []string{"api"}}}

	errs := ValidateAllRefs(skills, agents, nil)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateAllRefs_MissingSkill(t *testing.T) {
	skills := []model.Skill{{Name: "api", Path: "spec/skills/api.md", UsesSkills: []string{"nonexistent"}}}

	errs := ValidateAllRefs(skills, nil, nil)
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

	errs := ValidateAllRefs(nil, agents, nil)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if errs[0].Severity != "warning" {
		t.Errorf("severity = %q, want warning", errs[0].Severity)
	}
}
