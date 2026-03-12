package bootstrap

import (
	"strings"
	"testing"

	"github.com/semir/agent-army/internal/model"
)

func TestBuildResolvedDeps(t *testing.T) {
	skillMap := map[string]model.Skill{
		"error-handling": {Name: "error-handling", Summary: "Error taxonomy"},
		"go/coder":       {Name: "go/coder", Summary: "Go workflow"},
	}
	agentMap := map[string]model.Agent{
		"type-design-analyzer": {Name: "type-design-analyzer", Description: "Type analyzer"},
	}

	agent := model.Agent{
		Name:        "go/coder",
		UsesSkills:  []string{"go/coder", "error-handling"},
		UsesPlugins: []string{"code-simplifier"},
		DelegatesTo: []string{"type-design-analyzer"},
	}

	deps := buildResolvedDeps(agent, skillMap, agentMap)

	if len(deps.Skills) != 2 {
		t.Errorf("expected 2 skills, got %d", len(deps.Skills))
	}

	if len(deps.Plugins) != 1 || deps.Plugins[0] != "code-simplifier" {
		t.Errorf("plugins = %v, want [code-simplifier]", deps.Plugins)
	}

	if len(deps.DelegatesTo) != 1 || deps.DelegatesTo[0].Name != "type-design-analyzer" {
		t.Errorf("delegates = %v, want [type-design-analyzer]", deps.DelegatesTo)
	}
}

func TestBuildResolvedDeps_MissingRefs(t *testing.T) {
	agent := model.Agent{
		Name:        "test",
		UsesSkills:  []string{"nonexistent-skill"},
		DelegatesTo: []string{"nonexistent-agent"},
	}

	deps := buildResolvedDeps(agent, nil, nil)

	if len(deps.Skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(deps.Skills))
	}
	if len(deps.DelegatesTo) != 0 {
		t.Errorf("expected 0 delegates, got %d", len(deps.DelegatesTo))
	}
}

func TestEnrichAgentBody_Claude(t *testing.T) {
	body := "# Agent\n\n## Role\nDoes things.\n\n## Workflow\n1. Step one\n"
	deps := model.ResolvedDeps{
		Skills:  []model.Skill{{Name: "error-handling", Summary: "Error taxonomy"}},
		Plugins: []string{"code-simplifier"},
		DelegatesTo: []model.Agent{{Name: "type-design-analyzer", Description: "Type analysis"}},
	}

	result := enrichAgentBody(body, deps, TargetClaude)

	if !strings.Contains(result, "## Resources Available") {
		t.Error("missing Resources Available section")
	}
	if !strings.Contains(result, "### Skills (Invoke via Skill Tool)") {
		t.Error("missing Skills subsection")
	}
	if !strings.Contains(result, "`error-handling` -- Error taxonomy") {
		t.Error("missing skill entry")
	}
	if !strings.Contains(result, "### Plugins") {
		t.Error("missing Plugins subsection")
	}
	if !strings.Contains(result, "`code-simplifier`") {
		t.Error("missing plugin entry")
	}
	if !strings.Contains(result, "### Delegate Agents") {
		t.Error("missing Delegate Agents subsection")
	}
	if !strings.Contains(result, "`type-design-analyzer`") {
		t.Error("missing delegate entry")
	}

	// Resources should appear before Workflow
	resIdx := strings.Index(result, "## Resources Available")
	wfIdx := strings.Index(result, "## Workflow")
	if resIdx > wfIdx {
		t.Error("Resources section should appear before Workflow")
	}
}

func TestEnrichAgentBody_Cursor(t *testing.T) {
	body := "# Agent\n\n## Workflow\n1. Step one\n"
	deps := model.ResolvedDeps{
		Skills:      []model.Skill{{Name: "error-handling", Summary: "Error taxonomy"}},
		Plugins:     []string{"code-simplifier"},
		DelegatesTo: []model.Agent{{Name: "type-design-analyzer", Description: "Type analysis"}},
	}

	result := enrichAgentBody(body, deps, TargetCursor)

	if !strings.Contains(result, "### Workflow References") {
		t.Error("missing Workflow References subsection")
	}
	if !strings.Contains(result, "skills/error-handling/SKILL.md") {
		t.Error("missing skill file path reference")
	}
	// Cursor should NOT have plugins or delegate sections
	if strings.Contains(result, "### Plugins") {
		t.Error("Cursor output should not have Plugins section")
	}
	if strings.Contains(result, "### Delegate Agents") {
		t.Error("Cursor output should not have Delegate Agents section")
	}
}

func TestEnrichAgentBody_EmptyDeps(t *testing.T) {
	body := "# Agent\n\n## Workflow\n1. Step one\n"
	deps := model.ResolvedDeps{}

	result := enrichAgentBody(body, deps, TargetClaude)

	// Should still have the section header but no subsections
	if !strings.Contains(result, "## Resources Available") {
		t.Error("should contain Resources Available even with empty deps")
	}
	if strings.Contains(result, "### Skills") {
		t.Error("should not contain Skills subsection with no skills")
	}
}

func TestRewriteBodyRefs_Claude(t *testing.T) {
	body := "invoke the `error-handling` skill for error patterns"
	result := rewriteBodyRefs(body, TargetClaude)
	if result != body {
		t.Error("Claude body should not be rewritten")
	}
}

func TestRewriteBodyRefs_Cursor(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "invoke skill for",
			input: "invoke the `error-handling` skill for error patterns",
			want:  "read and follow the workflow in `skills/error-handling/SKILL.md` for error patterns",
		},
		{
			name:  "invoke skill standalone",
			input: "invoke the `api-designer` skill",
			want:  "read and follow the workflow in `skills/api-designer/SKILL.md`",
		},
		{
			name:  "delegate to",
			input: "Delegate to `type-design-analyzer` when reviewing types",
			want:  "read the review checklist in `agents/type-design-analyzer.md` when reviewing types",
		},
		{
			name:  "loaded via skill",
			input: "Go coding patterns are loaded via the `go/coder` skill. Key emphasis:",
			want:  "Go coding patterns are defined in `skills/go/coder/SKILL.md`. Key emphasis:",
		},
		{
			name:  "loaded via skills generic",
			input: "Patterns are loaded via skills. Concurrency patterns included.",
			want:  "Patterns are defined in the workflow files listed under Resources Available. Concurrency patterns included.",
		},
		{
			name:  "loaded via rule",
			input: "Patterns are loaded via the `ai-assisted-development` rule.",
			want:  "Patterns are defined in the `ai-assisted-development` rule.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rewriteBodyRefs(tt.input, TargetCursor)
			if got != tt.want {
				t.Errorf("\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestInjectSection(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		section string
		marker  string // expected string that should appear before the section
	}{
		{
			name:    "before workflow",
			body:    "# Title\n\n## Role\nStuff\n\n## Workflow\n1. Step\n",
			section: "## Resources\nContent\n",
			marker:  "## Workflow",
		},
		{
			name:    "before constraints",
			body:    "# Title\n\n## Constraints\n- Rule\n",
			section: "## Resources\nContent\n",
			marker:  "## Constraints",
		},
		{
			name:    "append if no marker",
			body:    "# Title\n\nJust content.\n",
			section: "## Resources\nContent\n",
			marker:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := injectSection(tt.body, tt.section)
			if !strings.Contains(result, tt.section) {
				t.Error("section not found in result")
			}
			if tt.marker != "" {
				secIdx := strings.Index(result, tt.section)
				markerIdx := strings.Index(result, tt.marker)
				if secIdx > markerIdx {
					t.Error("section should appear before marker")
				}
			}
		})
	}
}
