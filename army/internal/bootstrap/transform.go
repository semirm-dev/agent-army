package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/semir/agent-army/internal/model"
)

const (
	claudeToolsRW = "Read, Write, Edit, Bash, Glob, Grep"
	claudeToolsRO = "Read, Glob, Grep, Bash"
)

func agentToClaude(root string, agent model.Agent, deps model.ResolvedDeps) (string, error) {
	body, err := extractBody(filepath.Join(root, agent.Path))
	if err != nil {
		return "", err
	}

	body = enrichAgentBody(body, deps, TargetClaude)

	flat := flattenName(agent.Name)
	tools := claudeToolsRW
	if agent.Access == "read-only" {
		tools = claudeToolsRO
	}

	lines := []string{"---"}
	lines = append(lines, fmt.Sprintf("name: %s", flat))
	if strings.Contains(agent.Description, ":") {
		lines = append(lines, fmt.Sprintf("description: %q", agent.Description))
	} else {
		lines = append(lines, fmt.Sprintf("description: %s", agent.Description))
	}
	lines = append(lines, fmt.Sprintf("tools: %s", tools))
	lines = append(lines, "model: inherit")
	lines = append(lines, "---")

	return strings.Join(lines, "\n") + "\n\n" + body, nil
}

var (
	editRe = regexp.MustCompile("`Edit`")
	bashRe = regexp.MustCompile("`Bash`")
)

func agentToCursor(root string, agent model.Agent, deps model.ResolvedDeps) (string, error) {
	body, err := extractBody(filepath.Join(root, agent.Path))
	if err != nil {
		return "", err
	}

	body = enrichAgentBody(body, deps, TargetCursor)

	flat := flattenName(agent.Name)

	lines := []string{"---"}
	lines = append(lines, fmt.Sprintf("name: %s", flat))
	if strings.Contains(agent.Description, ":") {
		lines = append(lines, fmt.Sprintf("description: %q", agent.Description))
	} else {
		lines = append(lines, fmt.Sprintf("description: %s", agent.Description))
	}
	lines = append(lines, "model: inherit")
	if agent.Access == "read-only" {
		lines = append(lines, "readonly: true")
	}
	lines = append(lines, "---")

	body = editRe.ReplaceAllString(body, "`StrReplace`")
	body = bashRe.ReplaceAllString(body, "`Shell`")
	body = strings.ReplaceAll(body, "~/.claude/", "~/.cursor/")

	return strings.Join(lines, "\n") + "\n\n" + body, nil
}

func skillToClaude(root string, skill model.Skill) (string, error) {
	return extractBody(filepath.Join(root, skill.Path))
}

func skillToCursor(root string, skill model.Skill) (string, error) {
	body, err := extractBody(filepath.Join(root, skill.Path))
	if err != nil {
		return "", err
	}

	body = editRe.ReplaceAllString(body, "`StrReplace`")
	body = bashRe.ReplaceAllString(body, "`Shell`")
	body = strings.ReplaceAll(body, "~/.claude/", "~/.cursor/")

	flat := flattenName(skill.Name)
	desc := skillDescription(skill)

	lines := []string{"---"}
	lines = append(lines, fmt.Sprintf("name: %s", flat))
	if strings.Contains(desc, ":") {
		lines = append(lines, fmt.Sprintf("description: %q", desc))
	} else {
		lines = append(lines, fmt.Sprintf("description: %s", desc))
	}
	lines = append(lines, "---")

	return strings.Join(lines, "\n") + "\n\n" + body, nil
}

func extractBody(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	lines := strings.SplitAfter(string(content), "\n")
	dashCount := 0
	bodyStart := 0
	for i, line := range lines {
		if strings.TrimRight(line, "\n\r ") == "---" {
			dashCount++
			if dashCount == 2 {
				bodyStart = i + 1
				break
			}
		}
	}

	if dashCount < 2 {
		return string(content), nil
	}

	body := strings.Join(lines[bodyStart:], "")
	return strings.TrimLeft(body, "\n"), nil
}

func readFileContent(root string, relPath string) (string, error) {
	content, err := os.ReadFile(filepath.Join(root, relPath))
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func flattenName(name string) string {
	return strings.ReplaceAll(name, "/", "-")
}

func writeOutput(dest, relPath, content string) error {
	target := filepath.Join(dest, relPath)
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, target)
}

func resolveCollision(dest, relPath string) string {
	target := filepath.Join(dest, relPath)
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return relPath
	}

	dir := filepath.Dir(relPath)
	ext := filepath.Ext(relPath)
	stem := strings.TrimSuffix(filepath.Base(relPath), ext)

	for i := 2; i < 100; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s_%d%s", stem, i, ext))
		if _, err := os.Stat(filepath.Join(dest, candidate)); os.IsNotExist(err) {
			return candidate
		}
	}
	return relPath
}

// skillDescription returns a brief description for a skill, preferring Summary over Description.
func skillDescription(s model.Skill) string {
	if s.Summary != "" {
		return s.Summary
	}
	return s.Description
}
