package resolver

import (
	"testing"

	"github.com/semir/agent-army/internal/model"
)

func TestComputeAllFixes_Redundant(t *testing.T) {
	// Skill A depends on skill B. If a skill has both A and B in uses_skills, B is redundant.
	skills := []model.Skill{
		{Name: "A", UsesSkills: []string{"B"}, Path: "spec/skills/a.md"},
		{Name: "B", UsesSkills: nil, Path: "spec/skills/b.md"},
		{Name: "C", UsesSkills: []string{"A", "B"}, Path: "spec/skills/c.md"},
	}

	fixes := ComputeAllFixes(skills, nil, "/tmp")
	if len(fixes) != 1 {
		t.Fatalf("expected 1 fix, got %d", len(fixes))
	}
	fix := fixes[0]
	if fix.FilePath != "spec/skills/c.md" {
		t.Errorf("filePath = %q", fix.FilePath)
	}
	if len(fix.After) != 1 || fix.After[0] != "A" {
		t.Errorf("after = %v, want [A]", fix.After)
	}
}

func TestComputeAllFixes_NoRedundancy(t *testing.T) {
	skills := []model.Skill{
		{Name: "A", Path: "spec/skills/a.md"},
		{Name: "B", Path: "spec/skills/b.md"},
		{Name: "C", UsesSkills: []string{"A", "B"}, Path: "spec/skills/c.md"},
	}

	fixes := ComputeAllFixes(skills, nil, "/tmp")
	if len(fixes) != 0 {
		t.Errorf("expected no fixes, got %d", len(fixes))
	}
}
