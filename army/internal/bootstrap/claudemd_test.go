package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/semir/agent-army/internal/model"
)

func TestReplaceMarkerSection(t *testing.T) {
	content := "before\n<!-- BEGIN:test -->\nold content\n<!-- END:test -->\nafter"
	result := replaceMarkerSection(content, "<!-- BEGIN:test -->", "<!-- END:test -->", "new content\n")
	if !strings.Contains(result, "new content") {
		t.Error("replacement content not found")
	}
	if strings.Contains(result, "old content") {
		t.Error("old content should be replaced")
	}
	if !strings.Contains(result, "before") || !strings.Contains(result, "after") {
		t.Error("surrounding content should be preserved")
	}
	if strings.Contains(result, "<!-- BEGIN:test -->") {
		t.Error("begin marker should be stripped from output")
	}
	if strings.Contains(result, "<!-- END:test -->") {
		t.Error("end marker should be stripped from output")
	}
}

func TestReplaceMarkerSection_NoMarkers(t *testing.T) {
	content := "no markers here"
	result := replaceMarkerSection(content, "<!-- BEGIN:test -->", "<!-- END:test -->", "new")
	if result != content {
		t.Error("content should be unchanged when markers not found")
	}
}

func TestBuildAgentDefinitions(t *testing.T) {
	agents := []model.Agent{
		{Name: "go/coder", Access: "read-write", UsesPlugins: []string{"code-simplifier", "context7"}},
		{Name: "go/reviewer", Access: "read-only"},
		{Name: "typescript/coder", Access: "read-write"},
		{Name: "arch-reviewer", Access: "read-only"},
	}

	result := buildAgentDefinitions(agents)

	if !strings.Contains(result, "Agent Definitions:") {
		t.Error("missing lead-in prose")
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
}

func TestBuildAgentDefinitions_Empty(t *testing.T) {
	result := buildAgentDefinitions(nil)
	if result != "" {
		t.Errorf("expected empty string for no agents, got: %s", result)
	}
}

func TestBuildAgentDefinitions_PluginAnnotations(t *testing.T) {
	agents := []model.Agent{
		{Name: "go/coder", Access: "read-write", UsesPlugins: []string{"code-simplifier", "context7"}},
	}

	result := buildAgentDefinitions(agents)

	if !strings.Contains(result, "uses `code-simplifier` plugin") {
		t.Error("missing code-simplifier plugin annotation")
	}
	if !strings.Contains(result, "uses `context7` plugin") {
		t.Error("missing context7 plugin annotation")
	}
}

func TestBuildAgentDefinitions_ReadOnlyWithPlugins(t *testing.T) {
	agents := []model.Agent{
		{Name: "go/reviewer", Access: "read-only", UsesPlugins: []string{"code-review"}},
	}

	result := buildAgentDefinitions(agents)

	if !strings.Contains(result, "(read-only, uses `code-review` plugin)") {
		t.Errorf("expected combined annotation, got: %s", result)
	}
}

func TestAgentAnnotations(t *testing.T) {
	tests := []struct {
		name   string
		agent  model.Agent
		want   string
	}{
		{
			name:  "no annotations",
			agent: model.Agent{Access: "read-write"},
			want:  "",
		},
		{
			name:  "read-only only",
			agent: model.Agent{Access: "read-only"},
			want:  "(read-only)",
		},
		{
			name:  "plugins only",
			agent: model.Agent{Access: "read-write", UsesPlugins: []string{"context7"}},
			want:  "(uses `context7` plugin)",
		},
		{
			name:  "read-only with plugins",
			agent: model.Agent{Access: "read-only", UsesPlugins: []string{"code-review"}},
			want:  "(read-only, uses `code-review` plugin)",
		},
		{
			name:  "multiple plugins",
			agent: model.Agent{Access: "read-write", UsesPlugins: []string{"code-simplifier", "context7"}},
			want:  "(uses `code-simplifier` plugin, uses `context7` plugin)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agentAnnotations(tt.agent)
			if got != tt.want {
				t.Errorf("agentAnnotations() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildSubagentTips(t *testing.T) {
	t.Run("with read-only agents", func(t *testing.T) {
		agents := []model.Agent{
			{Name: "go/coder", Access: "read-write"},
			{Name: "go/reviewer", Access: "read-only"},
			{Name: "arch-reviewer", Access: "read-only"},
		}

		result := buildSubagentTips(agents)

		if !strings.Contains(result, "`go-reviewer`") {
			t.Error("missing go-reviewer in read-only list")
		}
		if !strings.Contains(result, "`arch-reviewer`") {
			t.Error("missing arch-reviewer in read-only list")
		}
		if strings.Contains(result, "`go-coder`") {
			t.Error("read-write agent should not appear in read-only list")
		}
		if !strings.Contains(result, "readonly: true") {
			t.Error("missing readonly tip")
		}
		if !strings.Contains(result, `model: "fast"`) {
			t.Error("missing fast model tip")
		}
	})

	t.Run("without read-only agents", func(t *testing.T) {
		agents := []model.Agent{
			{Name: "go/coder", Access: "read-write"},
		}

		result := buildSubagentTips(agents)

		if strings.Contains(result, "readonly: true") {
			t.Error("readonly tip should not appear when no read-only agents")
		}
		if !strings.Contains(result, `model: "fast"`) {
			t.Error("fast model tip should always appear")
		}
	})

	t.Run("no agents", func(t *testing.T) {
		result := buildSubagentTips(nil)
		if result != "" {
			t.Errorf("expected empty string for no agents, got: %s", result)
		}
	})
}

func TestBuildPluginsSection(t *testing.T) {
	superpowersWithWorkflows := model.Plugin{
		Name:        "superpowers",
		Description: "Dev workflows",
		Workflows: []model.Workflow{
			{Name: "brainstorming", Description: "Before creative work."},
			{Name: "debugging", Description: "When encountering bugs."},
		},
	}

	t.Run("all plugins with descriptions", func(t *testing.T) {
		plugins := []model.Plugin{
			{Name: "context7", Description: "Docs lookup"},
			superpowersWithWorkflows,
			{Name: "code-review", Description: "Code review"},
		}
		result := buildPluginsSection(plugins)

		if !strings.Contains(result, "Configured Plugins:") {
			t.Error("missing section header")
		}
		if !strings.Contains(result, "`context7` -- Docs lookup") {
			t.Error("missing context7 with description")
		}
		if !strings.Contains(result, "`superpowers` -- Dev workflows") {
			t.Error("missing superpowers with description")
		}
		if !strings.Contains(result, "`code-review` -- Code review") {
			t.Error("missing code-review with description")
		}
	})

	t.Run("workflows from data", func(t *testing.T) {
		plugins := []model.Plugin{
			{Name: "context7", Description: "Docs lookup"},
			superpowersWithWorkflows,
		}
		result := buildPluginsSection(plugins)

		if !strings.Contains(result, "Superpowers Workflows:") {
			t.Error("missing superpowers workflows header")
		}
		if !strings.Contains(result, "`brainstorming` -- Before creative work.") {
			t.Error("missing brainstorming workflow from data")
		}
		if !strings.Contains(result, "`debugging` -- When encountering bugs.") {
			t.Error("missing debugging workflow from data")
		}
	})

	t.Run("no workflows when plugin has none", func(t *testing.T) {
		plugins := []model.Plugin{
			{Name: "context7", Description: "Docs lookup"},
			{Name: "code-review", Description: "Code review"},
		}
		result := buildPluginsSection(plugins)

		if !strings.Contains(result, "Configured Plugins:") {
			t.Error("missing section header")
		}
		if strings.Contains(result, "Workflows:") {
			t.Error("no workflow section should appear when no plugin has workflows")
		}
	})

	t.Run("multiple plugins with workflows", func(t *testing.T) {
		plugins := []model.Plugin{
			{Name: "alpha", Description: "First", Workflows: []model.Workflow{
				{Name: "wf-a", Description: "Workflow A"},
			}},
			{Name: "beta", Description: "Second", Workflows: []model.Workflow{
				{Name: "wf-b", Description: "Workflow B"},
			}},
		}
		result := buildPluginsSection(plugins)

		if !strings.Contains(result, "Alpha Workflows:") {
			t.Error("missing alpha workflows header")
		}
		if !strings.Contains(result, "Beta Workflows:") {
			t.Error("missing beta workflows header")
		}
		if !strings.Contains(result, "`wf-a` -- Workflow A") {
			t.Error("missing workflow A")
		}
		if !strings.Contains(result, "`wf-b` -- Workflow B") {
			t.Error("missing workflow B")
		}
	})

	t.Run("plugin without description falls back to name", func(t *testing.T) {
		plugins := []model.Plugin{
			{Name: "custom-plugin"},
		}
		result := buildPluginsSection(plugins)

		if !strings.Contains(result, "`custom-plugin` -- custom-plugin") {
			t.Errorf("expected name fallback, got: %s", result)
		}
	})

	t.Run("no plugins", func(t *testing.T) {
		result := buildPluginsSection(nil)
		if result != "" {
			t.Errorf("expected empty string for no plugins, got: %s", result)
		}
	})
}

func TestBuildSkillDefinitions(t *testing.T) {
	skills := []model.Skill{
		{Name: "error-handling", Summary: "Error taxonomy and patterns"},
		{Name: "api-designer", Description: "API Designer", Summary: ""},
	}

	result := buildSkillDefinitions(skills)

	if !strings.Contains(result, "Custom Skills:") {
		t.Error("missing lead-in prose")
	}
	if !strings.Contains(result, "`error-handling` -- Error taxonomy and patterns") {
		t.Error("missing skill with summary")
	}
	if !strings.Contains(result, "`api-designer` -- API Designer") {
		t.Error("missing skill with H1 fallback")
	}
}

func TestBuildSkillDefinitions_Empty(t *testing.T) {
	result := buildSkillDefinitions(nil)
	if result != "" {
		t.Errorf("expected empty string for no skills, got: %s", result)
	}
}

func TestBuildRuleTable(t *testing.T) {
	rules := []model.Rule{
		{Name: "go/patterns", Summary: "Go coding conventions"},
		{Name: "security", Description: "Security Patterns", Summary: "Auth, CORS, rate limiting"},
	}

	result := buildRuleTable(rules)

	if !strings.Contains(result, "Detailed patterns are loaded on-demand") {
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
	if !strings.Contains(result, "Agents load their relevant pattern file at activation") {
		t.Error("missing trailing prose")
	}
}

func TestBuildRuleTable_Empty(t *testing.T) {
	result := buildRuleTable(nil)
	if result != "" {
		t.Errorf("expected empty string for no rules, got: %s", result)
	}
}

func TestAgentDomain(t *testing.T) {
	tests := []struct {
		desc  string
		agent model.Agent
		want  string
	}{
		{"go prefix", model.Agent{Name: "go/coder"}, "Go"},
		{"typescript prefix", model.Agent{Name: "typescript/reviewer"}, "TypeScript/JS"},
		{"react prefix", model.Agent{Name: "react/coder"}, "React"},
		{"python prefix", model.Agent{Name: "python/tester"}, "Python"},
		{"database prefix", model.Agent{Name: "database/coder"}, "Database"},
		{"infrastructure prefix", model.Agent{Name: "infrastructure/builder"}, "Infrastructure"},
		{"arch-reviewer exact", model.Agent{Name: "arch-reviewer"}, "Architecture"},
		{"docs-writer exact", model.Agent{Name: "docs-writer"}, "Documentation"},
		{"unknown falls to Quality", model.Agent{Name: "type-design-analyzer"}, "Quality"},
		{"frontmatter domain overrides prefix", model.Agent{Name: "go/coder", Domain: "Custom"}, "Custom"},
		{"frontmatter domain on new prefix", model.Agent{Name: "java/coder", Domain: "Java"}, "Java"},
		{"frontmatter domain on standalone", model.Agent{Name: "some-agent", Domain: "DevOps"}, "DevOps"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := agentDomain(tt.agent)
			if got != tt.want {
				t.Errorf("agentDomain(%+v) = %q, want %q", tt.agent, got, tt.want)
			}
		})
	}
}

func TestOrderedDomains(t *testing.T) {
	t.Run("preferred order only", func(t *testing.T) {
		groups := map[string][]model.Agent{
			"Go":             {{Name: "go/coder"}},
			"Quality":        {{Name: "linter"}},
			"Infrastructure": {{Name: "infrastructure/builder"}},
		}
		order := orderedDomains(groups)
		want := []string{"Go", "Infrastructure", "Quality"}
		if len(order) != len(want) {
			t.Fatalf("got %v, want %v", order, want)
		}
		for i, d := range want {
			if order[i] != d {
				t.Errorf("order[%d] = %q, want %q", i, order[i], d)
			}
		}
	})

	t.Run("extras appended alphabetically", func(t *testing.T) {
		groups := map[string][]model.Agent{
			"Go":     {{Name: "go/coder"}},
			"Java":   {{Name: "java/coder", Domain: "Java"}},
			"DevOps": {{Name: "deploy-agent", Domain: "DevOps"}},
		}
		order := orderedDomains(groups)
		// Go is in preferred list, then DevOps and Java are extras (alphabetical)
		want := []string{"Go", "DevOps", "Java"}
		if len(order) != len(want) {
			t.Fatalf("got %v, want %v", order, want)
		}
		for i, d := range want {
			if order[i] != d {
				t.Errorf("order[%d] = %q, want %q", i, order[i], d)
			}
		}
	})
}

func TestDestPrefix(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name string
		dest string
		want string
	}{
		{
			name: "global home",
			dest: filepath.Join(home, ".claude"),
			want: "~/.claude",
		},
		{
			name: "project local",
			dest: "/some/project/.claude",
			want: ".claude",
		},
		{
			name: "custom path",
			dest: "/opt/custom/output",
			want: "/opt/custom/output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := destPrefix(tt.dest)
			if got != tt.want {
				t.Errorf("destPrefix(%q) = %q, want %q", tt.dest, got, tt.want)
			}
		})
	}
}

func TestGenerateClaudeMD(t *testing.T) {
	dir := t.TempDir()
	tmplDir := filepath.Join(dir, "spec", "claude")
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
<!-- BEGIN:plugins-overview -->
old plugins
<!-- END:plugins-overview -->
<!-- BEGIN:custom-skills -->
old skills
<!-- END:custom-skills -->

## Rules
<!-- BEGIN:rules-table -->
old rules
<!-- END:rules-table -->

End of doc.
`
	tmplPath := filepath.Join(tmplDir, "CLAUDE.md")
	os.WriteFile(tmplPath, []byte(template), 0644)

	destDir := filepath.Join(dir, "project", ".claude")

	agents := []model.Agent{
		{Name: "go/coder", Access: "read-write", UsesPlugins: []string{"context7"}},
		{Name: "go/reviewer", Access: "read-only"},
	}
	skills := []model.Skill{{Name: "error-handling", Summary: "Error taxonomy"}}
	rules := []model.Rule{{Name: "security", Summary: "Auth patterns"}}

	plugins := []model.Plugin{
		{Name: "context7", Description: "Docs lookup"},
		{Name: "superpowers", Description: "Dev workflows", Workflows: []model.Workflow{
			{Name: "brainstorming", Description: "Before creative work."},
			{Name: "verification", Description: "Before claiming done."},
		}},
	}

	err := generateClaudeMD(destDir, tmplPath, agents, skills, rules, plugins)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}

	result := string(content)

	// Verify old content was replaced
	for _, old := range []string{"old agents", "old skills", "old rules", "old tips", "old plugins"} {
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
	if !strings.Contains(result, "go-coder.md") {
		t.Error("missing generated agent")
	}
	if !strings.Contains(result, "Custom Skills:") {
		t.Error("missing custom skills lead-in")
	}
	if !strings.Contains(result, "error-handling") {
		t.Error("missing generated skill")
	}
	if !strings.Contains(result, "Detailed patterns are loaded on-demand") {
		t.Error("missing rules lead-in")
	}
	if !strings.Contains(result, "security") {
		t.Error("missing generated rule")
	}

	// Verify plugin annotations on agents
	if !strings.Contains(result, "uses `context7` plugin") {
		t.Error("missing plugin annotation on go-coder")
	}

	// Verify subagent tips
	if !strings.Contains(result, "`go-reviewer`") {
		t.Error("missing go-reviewer in subagent tips")
	}

	// Verify unified plugins section — all plugins listed with descriptions
	if !strings.Contains(result, "Configured Plugins:") {
		t.Error("missing plugins section header")
	}
	if !strings.Contains(result, "`context7` -- Docs lookup") {
		t.Error("missing context7 plugin with description")
	}
	if !strings.Contains(result, "`superpowers` -- Dev workflows") {
		t.Error("missing superpowers plugin with description")
	}

	// Verify superpowers workflows from data
	if !strings.Contains(result, "Superpowers Workflows:") {
		t.Error("missing superpowers workflows section")
	}
	if !strings.Contains(result, "`brainstorming` -- Before creative work.") {
		t.Error("missing brainstorming workflow from data")
	}
	if !strings.Contains(result, "`verification` -- Before claiming done.") {
		t.Error("missing verification workflow from data")
	}

	// Verify {{BASE}} was replaced with .claude (project-local)
	if strings.Contains(result, "{{BASE}}") {
		t.Error("{{BASE}} placeholder should be replaced")
	}
	if !strings.Contains(result, ".claude/agents/") {
		t.Errorf("expected .claude/agents/ path for project-local dest, got:\n%s", result)
	}
	if !strings.Contains(result, ".claude/skills/") {
		t.Error("expected .claude/skills/ path")
	}
	if !strings.Contains(result, ".claude/rules/") {
		t.Error("expected .claude/rules/ path")
	}

	// Verify trailing prose moved into rules table
	if !strings.Contains(result, "Agents load their relevant pattern file at activation") {
		t.Error("missing trailing prose from buildRuleTable")
	}

	// Verify no excessive blank lines (collapse check)
	if strings.Contains(result, "\n\n\n") {
		t.Error("should not have 3+ consecutive newlines after blank line collapse")
	}
}

func TestGenerateClaudeMD_EmptyBuilders(t *testing.T) {
	dir := t.TempDir()
	tmplDir := filepath.Join(dir, "spec", "claude")
	os.MkdirAll(tmplDir, 0755)

	template := `# Header
- **Parallelism:** Split work.
<!-- BEGIN:agent-definitions -->
<!-- END:agent-definitions -->
<!-- BEGIN:subagent-tips -->
<!-- END:subagent-tips -->
<!-- BEGIN:plugins-overview -->
<!-- END:plugins-overview -->
<!-- BEGIN:custom-skills -->
<!-- END:custom-skills -->
- **Verification:** Run build.

## Rules
<!-- BEGIN:rules-table -->
<!-- END:rules-table -->
`
	tmplPath := filepath.Join(tmplDir, "CLAUDE.md")
	os.WriteFile(tmplPath, []byte(template), 0644)

	destDir := filepath.Join(dir, "project", ".claude")

	// All empty inputs — every builder returns ""
	err := generateClaudeMD(destDir, tmplPath, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "CLAUDE.md"))
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

func TestGenerateClaudeMD_GlobalDest(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	dir := t.TempDir()
	tmplDir := filepath.Join(dir, "spec", "claude")
	os.MkdirAll(tmplDir, 0755)

	template := "Paths: `{{BASE}}/agents/`, `{{BASE}}/skills/`, `{{BASE}}/rules/`\n"
	tmplPath := filepath.Join(tmplDir, "CLAUDE.md")
	os.WriteFile(tmplPath, []byte(template), 0644)

	destDir := filepath.Join(home, ".claude")

	err = generateClaudeMD(destDir, tmplPath, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(destDir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filepath.Join(destDir, "CLAUDE.md"))

	result := string(content)
	if !strings.Contains(result, "`~/.claude/agents/`") {
		t.Errorf("expected ~/.claude prefix for global dest, got:\n%s", result)
	}
}
