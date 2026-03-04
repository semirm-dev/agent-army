package resolver

import (
	"testing"

	"github.com/semir/agent-army/internal/model"
)

func TestComputeAllFixes_Redundant(t *testing.T) {
	rules := []model.Rule{
		{Name: "A", UsesRules: []string{"B"}, Path: "spec/rules/a.md"},
		{Name: "B", UsesRules: nil, Path: "spec/rules/b.md"},
	}
	// Rule A depends on B. If an agent has both A and B in uses_rules, B is redundant.
	agents := []model.Agent{
		{Name: "coder", Path: "spec/agents/coder.md", UsesRules: []string{"A", "B"}},
	}

	fixes := ComputeAllFixes(rules, nil, agents, "/tmp")
	if len(fixes) != 1 {
		t.Fatalf("expected 1 fix, got %d", len(fixes))
	}
	fix := fixes[0]
	if fix.FilePath != "spec/agents/coder.md" {
		t.Errorf("filePath = %q", fix.FilePath)
	}
	if len(fix.After) != 1 || fix.After[0] != "A" {
		t.Errorf("after = %v, want [A]", fix.After)
	}
}

func TestComputeAllFixes_NoRedundancy(t *testing.T) {
	rules := []model.Rule{
		{Name: "A", Path: "spec/rules/a.md"},
		{Name: "B", Path: "spec/rules/b.md"},
	}
	agents := []model.Agent{
		{Name: "coder", Path: "spec/agents/coder.md", UsesRules: []string{"A", "B"}},
	}

	fixes := ComputeAllFixes(rules, nil, agents, "/tmp")
	if len(fixes) != 0 {
		t.Errorf("expected no fixes, got %d", len(fixes))
	}
}
