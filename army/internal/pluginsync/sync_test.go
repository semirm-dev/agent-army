package pluginsync

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockRunner struct {
	commands []string
	failOn   map[string]bool
}

func (m *mockRunner) Run(name string, args []string) error {
	cmd := name + " " + strings.Join(args, " ")
	m.commands = append(m.commands, cmd)
	if m.failOn != nil && m.failOn[cmd] {
		return fmt.Errorf("mock failure: %s", cmd)
	}
	return nil
}

const sampleDoc = `# Plugins

| # | Name | Install |
|---|------|---------|
| 1 | **foo** | ` + "`" + `/plugin install foo@marketplace` + "`" + ` |
| 2 | **bar** | ` + "`" + `/plugin install bar@marketplace` + "`" + ` |

## Skills (2)

Install skills globally with ` + "`" + `npx skills add <repo> -g -s <skill-name>` + "`" + `.

### From [owner/repo](https://github.com/owner/repo) (1 skill)

| Skill | Description | Install |
|-------|-------------|---------|
| ` + "`" + `my-skill` + "`" + ` | desc | ` + "`" + `npx skills add owner/repo -g -s my-skill` + "`" + ` |

### Plugin-Provided Skills

| Skill | Description | Plugin Source |
|-------|-------------|---------------|
| ` + "`" + `plugin:skill` + "`" + ` | desc | ` + "`" + `npx skills add plugin/repo -g -s plugin-skill` + "`" + ` |
`

func TestParsePluginCommands(t *testing.T) {
	matches := pluginCmdRe.FindAllStringSubmatch(sampleDoc, -1)
	if len(matches) != 2 {
		t.Fatalf("expected 2 plugin commands, got %d", len(matches))
	}
	if matches[0][1] != "foo@marketplace" {
		t.Errorf("first plugin: got %q", matches[0][1])
	}
	if matches[1][1] != "bar@marketplace" {
		t.Errorf("second plugin: got %q", matches[1][1])
	}
}

func TestParseSkillCommands(t *testing.T) {
	// Only skills before "### Plugin-Provided Skills" should be extracted
	content := sampleDoc
	idx := strings.Index(content, "### Plugin-Provided Skills")
	skillContent := content[:idx]

	matches := skillCmdRe.FindAllStringSubmatch(skillContent, -1)
	// Should find my-skill but not the <repo> template or plugin-skill
	found := 0
	for _, m := range matches {
		if !strings.Contains(m[1], "<") {
			found++
		}
	}
	if found != 1 {
		t.Errorf("expected 1 non-template skill command, got %d", found)
	}
}

func TestRunWithMockRunner(t *testing.T) {
	dir := t.TempDir()
	docPath := filepath.Join(dir, "PLUGINS_AND_SKILLS.md")
	os.WriteFile(docPath, []byte(sampleDoc), 0644)

	runner := &mockRunner{}
	var buf bytes.Buffer
	err := Run(docPath, &buf, runner)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Should have 2 plugin installs + 1 skill install
	if len(runner.commands) != 3 {
		t.Fatalf("expected 3 commands, got %d: %v", len(runner.commands), runner.commands)
	}

	if runner.commands[0] != "claude plugin install foo@marketplace" {
		t.Errorf("cmd[0] = %q", runner.commands[0])
	}
	if runner.commands[1] != "claude plugin install bar@marketplace" {
		t.Errorf("cmd[1] = %q", runner.commands[1])
	}
	if !strings.HasPrefix(runner.commands[2], "npx skills add owner/repo -g -s my-skill") {
		t.Errorf("cmd[2] = %q", runner.commands[2])
	}
}

const sampleDocWithRedundant = sampleDoc + `
> **Redundant standalone skills:** These are already provided by plugins and can be removed:
> - ` + "`frontend-design`" + ` (provided by **frontend-design** plugin) — ` + "`npx skills remove frontend-design`" + `
> - ` + "`skill-creator`" + ` (provided by **skill-creator** plugin) — ` + "`npx skills remove skill-creator`" + `
`

func TestCleanupParsing(t *testing.T) {
	matches := redundantSkillRe.FindAllStringSubmatch(sampleDocWithRedundant, -1)
	if len(matches) != 2 {
		t.Fatalf("expected 2 redundant skills, got %d", len(matches))
	}
	if matches[0][1] != "frontend-design" {
		t.Errorf("first redundant skill: got %q, want %q", matches[0][1], "frontend-design")
	}
	if matches[1][1] != "skill-creator" {
		t.Errorf("second redundant skill: got %q, want %q", matches[1][1], "skill-creator")
	}
}

func TestRunWithCleanup(t *testing.T) {
	dir := t.TempDir()
	docPath := filepath.Join(dir, "PLUGINS_AND_SKILLS.md")
	os.WriteFile(docPath, []byte(sampleDocWithRedundant), 0644)

	runner := &mockRunner{}
	var buf bytes.Buffer
	err := Run(docPath, &buf, runner)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Should have 2 plugin installs + 1 skill install + 2 cleanup removals = 5
	if len(runner.commands) != 5 {
		t.Fatalf("expected 5 commands, got %d: %v", len(runner.commands), runner.commands)
	}

	// Verify cleanup commands
	if runner.commands[3] != "npx skills remove frontend-design -y" {
		t.Errorf("cmd[3] = %q, want cleanup of frontend-design", runner.commands[3])
	}
	if runner.commands[4] != "npx skills remove skill-creator -y" {
		t.Errorf("cmd[4] = %q, want cleanup of skill-creator", runner.commands[4])
	}

	output := buf.String()
	if !strings.Contains(output, "Cleaning Up Redundant Skills") {
		t.Error("expected cleanup section header in output")
	}
}

func TestRunWithFailures(t *testing.T) {
	dir := t.TempDir()
	docPath := filepath.Join(dir, "PLUGINS_AND_SKILLS.md")
	os.WriteFile(docPath, []byte(sampleDoc), 0644)

	runner := &mockRunner{
		failOn: map[string]bool{
			"claude plugin install foo@marketplace": true,
		},
	}
	var buf bytes.Buffer
	err := Run(docPath, &buf, runner)
	if err == nil {
		t.Fatal("expected error for failed command")
	}
	if !strings.Contains(buf.String(), "Failed") {
		t.Error("expected failure message in output")
	}
}
