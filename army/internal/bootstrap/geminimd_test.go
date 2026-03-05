package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semir/agent-army/internal/model"
)

func TestGenerateGeminiMD(t *testing.T) {
	dir := t.TempDir()
	tmplDir := filepath.Join(dir, "spec", "gemini")
	os.MkdirAll(tmplDir, 0755)

	template := `# Orchestrator

Agent prompts live in ` + "`{{BASE}}/agents/`" + `.

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

	destDir := filepath.Join(dir, "project", ".gemini")

	agents := []model.Agent{
		{Name: "go/coder", Access: "read-write"},
		{Name: "go/reviewer", Access: "read-only"},
	}
	skills := []model.Skill{{Name: "error-handling", Summary: "Error taxonomy"}}
	rules := []model.Rule{{Name: "security", Summary: "Auth patterns"}}

	err := generateGeminiMD(destDir, tmplPath, agents, skills, rules)
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

	// Verify generated content with prose lead-ins
	if !strings.Contains(result, "Agent Definitions:") {
		t.Error("missing agent definitions lead-in")
	}
	if !strings.Contains(result, "@agents/go-coder.md") {
		t.Error("missing generated agent with @file reference")
	}
	if !strings.Contains(result, "Workflow Skills:") {
		t.Error("missing workflow skills lead-in")
	}
	if !strings.Contains(result, "@skills/error-handling/SKILL.md") {
		t.Error("missing generated skill with @file reference")
	}
	if !strings.Contains(result, "Detailed patterns are loaded via @file") {
		t.Error("missing rules lead-in")
	}
	if !strings.Contains(result, "@rules/security.md") {
		t.Error("missing generated rule with @file reference")
	}

	// Verify subagent tips
	if !strings.Contains(result, "`go-reviewer`") {
		t.Error("missing go-reviewer in subagent tips")
	}
	if !strings.Contains(result, "gemini-2.5-pro") {
		t.Error("missing gemini model tip")
	}

	// Verify {{BASE}} was replaced with .gemini (project-local)
	if strings.Contains(result, "{{BASE}}") {
		t.Error("{{BASE}} placeholder should be replaced")
	}
	if !strings.Contains(result, ".gemini/agents/") {
		t.Errorf("expected .gemini/agents/ path for project-local dest, got:\n%s", result)
	}
	if !strings.Contains(result, ".gemini/skills/") {
		t.Error("expected .gemini/skills/ path")
	}
	if !strings.Contains(result, ".gemini/rules/") {
		t.Error("expected .gemini/rules/ path")
	}

	// Verify trailing prose from buildGeminiRuleTable
	if !strings.Contains(result, "Rules are loaded via @file imports") {
		t.Error("missing trailing prose from buildGeminiRuleTable")
	}

	// Verify no excessive blank lines (collapse check)
	if strings.Contains(result, "\n\n\n") {
		t.Error("should not have 3+ consecutive newlines after blank line collapse")
	}
}

func TestGenerateGeminiMD_EmptyBuilders(t *testing.T) {
	dir := t.TempDir()
	tmplDir := filepath.Join(dir, "spec", "gemini")
	os.MkdirAll(tmplDir, 0755)

	template := `# Header
- **Parallelism:** Split work.
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

	destDir := filepath.Join(dir, "project", ".gemini")

	// All empty inputs -- every builder returns ""
	err := generateGeminiMD(destDir, tmplPath, nil, nil, nil)
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
	if !strings.Contains(result, "Parallelism:") {
		t.Error("missing static content before markers")
	}
	if !strings.Contains(result, "Verification:") {
		t.Error("missing static content after markers")
	}
}

func TestGeminiDestPrefix(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name string
		dest string
		want string
	}{
		{
			name: "global home",
			dest: filepath.Join(home, ".gemini"),
			want: "~/.gemini",
		},
		{
			name: "project local",
			dest: "/some/project/.gemini",
			want: ".gemini",
		},
		{
			name: "custom path",
			dest: "/opt/custom/output",
			want: "/opt/custom/output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := geminiDestPrefix(tt.dest)
			if got != tt.want {
				t.Errorf("geminiDestPrefix(%q) = %q, want %q", tt.dest, got, tt.want)
			}
		})
	}
}

func TestBuildGeminiAgentDefinitions(t *testing.T) {
	agents := []model.Agent{
		{Name: "go/coder", Access: "read-write"},
		{Name: "go/reviewer", Access: "read-only"},
		{Name: "typescript/coder", Access: "read-write"},
		{Name: "arch-reviewer", Access: "read-only"},
	}

	result := buildGeminiAgentDefinitions(agents)

	if !strings.Contains(result, "Agent Definitions:") {
		t.Error("missing lead-in prose")
	}
	if !strings.Contains(result, "@file imports") {
		t.Error("missing @file imports reference in lead-in")
	}
	if !strings.Contains(result, "**Go:**") {
		t.Error("missing Go group")
	}
	if !strings.Contains(result, "@agents/go-coder.md") {
		t.Error("missing go-coder with @agents prefix")
	}
	if !strings.Contains(result, "@agents/go-reviewer.md` (read-only)") {
		t.Error("missing read-only marker on go-reviewer")
	}
	if !strings.Contains(result, "**TypeScript/JS:**") {
		t.Error("missing TypeScript group")
	}
	if !strings.Contains(result, "**Architecture:**") {
		t.Error("missing Architecture group")
	}
}

func TestBuildGeminiAgentDefinitions_Empty(t *testing.T) {
	result := buildGeminiAgentDefinitions(nil)
	if result != "" {
		t.Errorf("expected empty string for no agents, got: %s", result)
	}
}

func TestBuildGeminiSubagentTips(t *testing.T) {
	t.Run("with read-only agents", func(t *testing.T) {
		agents := []model.Agent{
			{Name: "go/coder", Access: "read-write"},
			{Name: "go/reviewer", Access: "read-only"},
			{Name: "arch-reviewer", Access: "read-only"},
		}

		result := buildGeminiSubagentTips(agents)

		if !strings.Contains(result, "`go-reviewer`") {
			t.Error("missing go-reviewer in read-only list")
		}
		if !strings.Contains(result, "`arch-reviewer`") {
			t.Error("missing arch-reviewer in read-only list")
		}
		if strings.Contains(result, "`go-coder`") {
			t.Error("read-write agent should not appear in read-only list")
		}
		if !strings.Contains(result, "read_file") {
			t.Error("missing read tool reference for read-only agents")
		}
		if !strings.Contains(result, "gemini-2.5-pro") {
			t.Error("missing gemini model tip")
		}
	})

	t.Run("without read-only agents", func(t *testing.T) {
		agents := []model.Agent{
			{Name: "go/coder", Access: "read-write"},
		}

		result := buildGeminiSubagentTips(agents)

		if strings.Contains(result, "Read-only agents") {
			t.Error("read-only tip should not appear when no read-only agents")
		}
		if !strings.Contains(result, "gemini-2.5-pro") {
			t.Error("gemini model tip should always appear")
		}
	})

	t.Run("no agents", func(t *testing.T) {
		result := buildGeminiSubagentTips(nil)
		if result != "" {
			t.Errorf("expected empty string for no agents, got: %s", result)
		}
	})
}

func TestBuildGeminiSkillDefinitions(t *testing.T) {
	skills := []model.Skill{
		{Name: "error-handling", Summary: "Error taxonomy and patterns"},
		{Name: "api-designer", Description: "API Designer", Summary: ""},
	}

	result := buildGeminiSkillDefinitions(skills)

	if !strings.Contains(result, "Workflow Skills:") {
		t.Error("missing lead-in prose")
	}
	if !strings.Contains(result, "via @file") {
		t.Error("missing @file reference in lead-in")
	}
	if !strings.Contains(result, "`@skills/error-handling/SKILL.md` -- Error taxonomy and patterns") {
		t.Error("missing skill with @file path and summary")
	}
	if !strings.Contains(result, "`@skills/api-designer/SKILL.md` -- API Designer") {
		t.Error("missing skill with @file path and H1 fallback")
	}
}

func TestBuildGeminiSkillDefinitions_Empty(t *testing.T) {
	result := buildGeminiSkillDefinitions(nil)
	if result != "" {
		t.Errorf("expected empty string for no skills, got: %s", result)
	}
}

func TestBuildGeminiRuleTable(t *testing.T) {
	rules := []model.Rule{
		{Name: "go/patterns", Summary: "Go coding conventions"},
		{Name: "security", Description: "Security Patterns", Summary: "Auth, CORS, rate limiting"},
	}

	result := buildGeminiRuleTable(rules)

	if !strings.Contains(result, "Detailed patterns are loaded via @file") {
		t.Error("missing lead-in prose")
	}
	if !strings.Contains(result, "| Rule File | Content |") {
		t.Error("missing table header")
	}
	if !strings.Contains(result, "`@rules/go-patterns.md`") {
		t.Error("missing flattened rule path with @rules prefix")
	}
	if !strings.Contains(result, "Go coding conventions") {
		t.Error("missing rule description")
	}
	if !strings.Contains(result, "Auth, CORS, rate limiting") {
		t.Error("missing rule summary")
	}
	if !strings.Contains(result, "Rules are loaded via @file imports") {
		t.Error("missing trailing prose")
	}
}

func TestBuildGeminiRuleTable_Empty(t *testing.T) {
	result := buildGeminiRuleTable(nil)
	if result != "" {
		t.Errorf("expected empty string for no rules, got: %s", result)
	}
}

func TestGeminiAgentAnnotations(t *testing.T) {
	tests := []struct {
		name  string
		agent model.Agent
		want  string
	}{
		{
			name:  "no annotations",
			agent: model.Agent{Access: "read-write"},
			want:  "",
		},
		{
			name:  "read-only",
			agent: model.Agent{Access: "read-only"},
			want:  "(read-only)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := geminiAgentAnnotations(tt.agent)
			if got != tt.want {
				t.Errorf("geminiAgentAnnotations() = %q, want %q", got, tt.want)
			}
		})
	}
}
