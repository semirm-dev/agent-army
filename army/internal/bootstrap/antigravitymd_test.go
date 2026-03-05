package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semir/agent-army/internal/model"
)

func TestGenerateAntigravityMD(t *testing.T) {
	dir := t.TempDir()
	tmplDir := filepath.Join(dir, "spec", "antigravity")
	os.MkdirAll(tmplDir, 0755)

	template := `# Orchestrator

Agent references in ` + "`{{BASE}}/agents/`" + `.

## Agents
<!-- BEGIN:agent-definitions -->
old agents
<!-- END:agent-definitions -->
<!-- BEGIN:subagent-tips -->
old tips
<!-- END:subagent-tips -->
<!-- BEGIN:custom-skills -->
old skills
<!-- END:custom-skills -->

## Rules
<!-- BEGIN:rules-table -->
old rules
<!-- END:rules-table -->

End of doc.
`
	tmplPath := filepath.Join(tmplDir, "GEMINI.md")
	os.WriteFile(tmplPath, []byte(template), 0644)

	destDir := filepath.Join(dir, "project", ".agent")

	agents := []model.Agent{
		{Name: "go/coder", Access: "read-write"},
		{Name: "go/reviewer", Access: "read-only"},
	}
	skills := []model.Skill{{Name: "error-handling", Summary: "Error taxonomy"}}
	rules := []model.Rule{{Name: "security", Summary: "Auth patterns"}}

	err := generateAntigravityMD(destDir, tmplPath, agents, skills, rules)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "GEMINI.md"))
	if err != nil {
		t.Fatal(err)
	}

	result := string(content)

	// Verify old content was replaced
	for _, old := range []string{"old agents", "old skills", "old rules", "old tips"} {
		if strings.Contains(result, old) {
			t.Errorf("%q should be replaced", old)
		}
	}

	// Verify markers are stripped from output
	if strings.Contains(result, "<!-- BEGIN:") {
		t.Error("markers should be stripped from generated output")
	}
	if strings.Contains(result, "<!-- END:") {
		t.Error("markers should be stripped from generated output")
	}

	// Verify generated content
	if !strings.Contains(result, "Agent References:") {
		t.Error("missing agent references lead-in")
	}
	if !strings.Contains(result, "go-coder.md") {
		t.Error("missing generated agent")
	}
	if !strings.Contains(result, "Workflow Skills:") {
		t.Error("missing skills lead-in")
	}
	if !strings.Contains(result, "skills/error-handling/SKILL.md") {
		t.Error("missing generated skill with plain file path")
	}
	if !strings.Contains(result, "Rules are available") {
		t.Error("missing rules lead-in")
	}
	if !strings.Contains(result, "security") {
		t.Error("missing generated rule")
	}

	// Verify agent reference tips
	if !strings.Contains(result, "Agent Reference Tips:") {
		t.Error("missing agent reference tips")
	}
	if !strings.Contains(result, "`go-reviewer`") {
		t.Error("missing go-reviewer in reference tips")
	}

	// Verify {{BASE}} was replaced with .agent (project-local)
	if strings.Contains(result, "{{BASE}}") {
		t.Error("{{BASE}} placeholder should be replaced")
	}
	if !strings.Contains(result, ".agent/agents/") {
		t.Errorf("expected .agent/agents/ path for project-local dest, got:\n%s", result)
	}
	if !strings.Contains(result, ".agent/skills/") {
		t.Error("expected .agent/skills/ path")
	}
	if !strings.Contains(result, ".agent/rules/") {
		t.Error("expected .agent/rules/ path")
	}

	// Verify no excessive blank lines
	if strings.Contains(result, "\n\n\n") {
		t.Error("should not have 3+ consecutive newlines after blank line collapse")
	}
}

func TestGenerateAntigravityMD_EmptyBuilders(t *testing.T) {
	dir := t.TempDir()
	tmplDir := filepath.Join(dir, "spec", "antigravity")
	os.MkdirAll(tmplDir, 0755)

	template := `# Header
- **Role:** Delegate tasks.
<!-- BEGIN:agent-definitions -->
<!-- END:agent-definitions -->
<!-- BEGIN:subagent-tips -->
<!-- END:subagent-tips -->
<!-- BEGIN:custom-skills -->
<!-- END:custom-skills -->
- **Verification:** Run build.

## Rules
<!-- BEGIN:rules-table -->
<!-- END:rules-table -->
`
	tmplPath := filepath.Join(tmplDir, "GEMINI.md")
	os.WriteFile(tmplPath, []byte(template), 0644)

	destDir := filepath.Join(dir, "project", ".agent")

	err := generateAntigravityMD(destDir, tmplPath, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "GEMINI.md"))
	if err != nil {
		t.Fatal(err)
	}

	result := string(content)

	// No 3+ consecutive newlines
	if strings.Contains(result, "\n\n\n") {
		t.Errorf("blank lines should be collapsed, got:\n%s", result)
	}

	// No marker remnants
	if strings.Contains(result, "<!-- BEGIN:") || strings.Contains(result, "<!-- END:") {
		t.Error("markers should be stripped")
	}

	// Static content preserved
	if !strings.Contains(result, "Role:") {
		t.Error("missing static content before markers")
	}
	if !strings.Contains(result, "Verification:") {
		t.Error("missing static content after markers")
	}
}

func TestAntigravityDestPrefix(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name string
		dest string
		want string
	}{
		{
			name: "global antigravity home",
			dest: filepath.Join(home, ".gemini", "antigravity"),
			want: "~/.gemini/antigravity",
		},
		{
			name: "project local .agent",
			dest: "/some/project/.agent",
			want: ".agent",
		},
		{
			name: "custom path",
			dest: "/opt/custom/output",
			want: "/opt/custom/output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := antigravityDestPrefix(tt.dest)
			if got != tt.want {
				t.Errorf("antigravityDestPrefix(%q) = %q, want %q", tt.dest, got, tt.want)
			}
		})
	}
}

func TestBuildAntigravityAgentReferences(t *testing.T) {
	agents := []model.Agent{
		{Name: "go/coder", Access: "read-write"},
		{Name: "go/reviewer", Access: "read-only"},
		{Name: "typescript/coder", Access: "read-write"},
		{Name: "arch-reviewer", Access: "read-only"},
	}

	result := buildAntigravityAgentReferences(agents)

	if !strings.Contains(result, "Agent References:") {
		t.Error("missing lead-in prose")
	}
	if !strings.Contains(result, "Reference documents for specialized roles") {
		t.Error("missing reference documents description")
	}
	if !strings.Contains(result, "**Go:**") {
		t.Error("missing Go group")
	}
	if !strings.Contains(result, "go-coder.md") {
		t.Error("missing go-coder")
	}
	if !strings.Contains(result, "go-reviewer.md` (read-only)") {
		t.Error("missing read-only marker on go-reviewer")
	}
	if !strings.Contains(result, "**TypeScript/JS:**") {
		t.Error("missing TypeScript group")
	}
	if !strings.Contains(result, "**Architecture:**") {
		t.Error("missing Architecture group")
	}
	// Verify no plugin annotations (Antigravity doesn't support plugins)
	if strings.Contains(result, "plugin") {
		t.Error("Antigravity agents should not have plugin annotations")
	}
}

func TestBuildAntigravityAgentReferences_Empty(t *testing.T) {
	result := buildAntigravityAgentReferences(nil)
	if result != "" {
		t.Errorf("expected empty string for no agents, got: %s", result)
	}
}

func TestBuildAntigravitySubagentTips(t *testing.T) {
	t.Run("with read-only agents", func(t *testing.T) {
		agents := []model.Agent{
			{Name: "go/coder", Access: "read-write"},
			{Name: "go/reviewer", Access: "read-only"},
			{Name: "arch-reviewer", Access: "read-only"},
		}

		result := buildAntigravitySubagentTips(agents)

		if !strings.Contains(result, "Agent Reference Tips:") {
			t.Error("missing tips header")
		}
		if !strings.Contains(result, "reference documents") {
			t.Error("missing reference documents tip")
		}
		if !strings.Contains(result, "`go-reviewer`") {
			t.Error("missing go-reviewer in read-only list")
		}
		if !strings.Contains(result, "`arch-reviewer`") {
			t.Error("missing arch-reviewer in read-only list")
		}
		if strings.Contains(result, "`go-coder`") {
			t.Error("read-write agent should not appear in read-only list")
		}
	})

	t.Run("without read-only agents", func(t *testing.T) {
		agents := []model.Agent{
			{Name: "go/coder", Access: "read-write"},
		}

		result := buildAntigravitySubagentTips(agents)

		if !strings.Contains(result, "Agent Reference Tips:") {
			t.Error("tips header should always appear when agents exist")
		}
		if strings.Contains(result, "Read-only references") {
			t.Error("read-only tip should not appear when no read-only agents")
		}
	})

	t.Run("no agents", func(t *testing.T) {
		result := buildAntigravitySubagentTips(nil)
		if result != "" {
			t.Errorf("expected empty string for no agents, got: %s", result)
		}
	})
}

func TestBuildAntigravitySkillDefinitions(t *testing.T) {
	skills := []model.Skill{
		{Name: "error-handling", Summary: "Error taxonomy and patterns"},
		{Name: "api-designer", Description: "API Designer", Summary: ""},
	}

	result := buildAntigravitySkillDefinitions(skills)

	if !strings.Contains(result, "Workflow Skills:") {
		t.Error("missing lead-in prose")
	}
	if !strings.Contains(result, "`skills/error-handling/SKILL.md` -- Error taxonomy and patterns") {
		t.Error("missing skill with plain file path and summary")
	}
	if !strings.Contains(result, "`skills/api-designer/SKILL.md` -- API Designer") {
		t.Error("missing skill with H1 fallback")
	}
}

func TestBuildAntigravitySkillDefinitions_Empty(t *testing.T) {
	result := buildAntigravitySkillDefinitions(nil)
	if result != "" {
		t.Errorf("expected empty string for no skills, got: %s", result)
	}
}

func TestBuildAntigravityRuleTable(t *testing.T) {
	rules := []model.Rule{
		{Name: "go/patterns", Summary: "Go coding conventions"},
		{Name: "security", Description: "Security Patterns", Summary: "Auth, CORS, rate limiting"},
	}

	result := buildAntigravityRuleTable(rules)

	if !strings.Contains(result, "Rules are available") {
		t.Error("missing lead-in prose")
	}
	if !strings.Contains(result, "| Rule File | Content |") {
		t.Error("missing table header")
	}
	if !strings.Contains(result, "`rules/go-patterns.md`") {
		t.Error("missing flattened rule path")
	}
	if !strings.Contains(result, "Go coding conventions") {
		t.Error("missing rule description")
	}
	if !strings.Contains(result, "Auth, CORS, rate limiting") {
		t.Error("missing rule summary")
	}
	if !strings.Contains(result, "Read the relevant rule file before working on tasks") {
		t.Error("missing trailing prose")
	}
}

func TestBuildAntigravityRuleTable_Empty(t *testing.T) {
	result := buildAntigravityRuleTable(nil)
	if result != "" {
		t.Errorf("expected empty string for no rules, got: %s", result)
	}
}
