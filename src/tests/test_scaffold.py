"""Tests for agent_army.scaffold."""

from __future__ import annotations

from pathlib import Path
from textwrap import dedent

import pytest

from agent_army.scaffold import (
    _build_file_path,
    _check_duplicate,
    _default_description,
    _generate_agent_content,
    _generate_frontmatter,
    _generate_rule_content,
    _generate_skill_content,
    _name_to_title,
    scaffold_flow,
)


# ---------------------------------------------------------------------------
# Unit tests — pure functions
# ---------------------------------------------------------------------------


class TestBuildFilePath:
    """_build_file_path() — constructs target path from root + dir + name."""

    def test_flat_name(self, tmp_path: Path) -> None:
        result = _build_file_path(tmp_path, "rules", "security")
        assert result == tmp_path / "rules" / "security.md"

    def test_nested_name(self, tmp_path: Path) -> None:
        result = _build_file_path(tmp_path, "agents", "go/coder")
        assert result == tmp_path / "agents" / "go" / "coder.md"


class TestCheckDuplicate:
    """_check_duplicate() — returns True and prints message when file exists."""

    def test_no_duplicate(self, tmp_path: Path) -> None:
        path = tmp_path / "rules" / "new-rule.md"
        assert _check_duplicate(path) is False

    def test_duplicate_exists(self, tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
        path = tmp_path / "rules" / "existing.md"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text("existing content")
        assert _check_duplicate(path) is True
        captured = capsys.readouterr()
        assert "already exists" in captured.out


class TestDefaultDescription:
    """_default_description() — generates sensible defaults."""

    @pytest.mark.parametrize(
        "entity_type, name, expected",
        [
            ("rule", "security", "Security patterns and conventions"),
            ("rule", "go/patterns", "Go Patterns patterns and conventions"),
            ("skill", "error-handling", "Error Handling workflow and decision tree"),
            ("skill", "go/testing", "Go Testing workflow and decision tree"),
            ("agent", "go/coder", "Go Coder specialist agent"),
            ("agent", "arch-reviewer", "Arch Reviewer specialist agent"),
        ],
    )
    def test_descriptions(self, entity_type: str, name: str, expected: str) -> None:
        assert _default_description(entity_type, name) == expected


class TestNameToTitle:
    """_name_to_title() — converts names to title case."""

    @pytest.mark.parametrize(
        "name, expected",
        [
            ("security", "Security"),
            ("go/patterns", "Go Patterns"),
            ("error-handling", "Error Handling"),
            ("go/testing-patterns", "Go Testing Patterns"),
        ],
    )
    def test_titles(self, name: str, expected: str) -> None:
        assert _name_to_title(name) == expected


class TestGenerateFrontmatter:
    """_generate_frontmatter() — builds YAML frontmatter."""

    def test_scalar_fields(self) -> None:
        result = _generate_frontmatter({"name": "test", "scope": "universal"})
        assert result == "---\nname: test\nscope: universal\n---"

    def test_list_fields(self) -> None:
        result = _generate_frontmatter({"uses_rules": ["a", "b"]})
        assert result == "---\nuses_rules: [a, b]\n---"

    def test_empty_list(self) -> None:
        result = _generate_frontmatter({"uses_rules": []})
        assert result == "---\nuses_rules: []\n---"

    def test_description_with_colon(self) -> None:
        result = _generate_frontmatter({"description": "Something: with colon"})
        assert result == '---\ndescription: "Something: with colon"\n---'

    def test_description_without_colon(self) -> None:
        result = _generate_frontmatter({"description": "Simple description"})
        assert result == "---\ndescription: Simple description\n---"


class TestGenerateRuleContent:
    """_generate_rule_content() — generates rule template."""

    def test_contains_sections(self) -> None:
        fields = {
            "name": "test",
            "description": "Test rule",
            "scope": "universal",
            "languages": [],
            "uses_rules": [],
        }
        content = _generate_rule_content(fields, "test")
        assert "# Test Patterns" in content
        assert "## Overview" in content
        assert "## Patterns" in content
        assert "## Anti-Patterns" in content
        assert "---" in content

    def test_frontmatter_present(self) -> None:
        fields = {
            "name": "security",
            "description": "Security patterns",
            "scope": "universal",
            "languages": [],
            "uses_rules": ["cross-cutting"],
        }
        content = _generate_rule_content(fields, "security")
        assert "name: security" in content
        assert "uses_rules: [cross-cutting]" in content


class TestGenerateSkillContent:
    """_generate_skill_content() — generates skill template."""

    def test_contains_sections(self) -> None:
        fields = {
            "name": "test-skill",
            "description": "Test skill",
            "scope": "universal",
            "languages": [],
            "uses_rules": [],
        }
        content = _generate_skill_content(fields, "test-skill")
        assert "# Test Skill" in content
        assert "## When to Use" in content
        assert "## Workflow" in content
        assert "## Decision Tree" in content
        assert "## Checklist" in content


class TestGenerateAgentContent:
    """_generate_agent_content() — generates agent template."""

    def test_read_write_capabilities(self) -> None:
        fields = {
            "name": "go/coder",
            "description": "Go coder agent",
            "role": "coder",
            "scope": "language-specific",
            "languages": ["go"],
            "access": "read-write",
            "uses_skills": [],
            "uses_rules": [],
            "uses_plugins": [],
            "delegates_to": [],
        }
        content = _generate_agent_content(fields, "go/coder", "read-write")
        assert "# Go Coder Agent" in content
        assert "Read and write source files" in content
        assert "Cannot modify" not in content

    def test_read_only_capabilities(self) -> None:
        fields = {
            "name": "arch-reviewer",
            "description": "Architecture reviewer",
            "role": "reviewer",
            "scope": "universal",
            "languages": [],
            "access": "read-only",
            "uses_skills": [],
            "uses_rules": [],
            "uses_plugins": [],
            "delegates_to": [],
        }
        content = _generate_agent_content(fields, "arch-reviewer", "read-only")
        assert "# Arch Reviewer Agent" in content
        assert "Cannot modify any files" in content
        assert "Read and write" not in content

    def test_contains_all_sections(self) -> None:
        fields = {
            "name": "test",
            "description": "Test agent",
            "role": "coder",
            "scope": "universal",
            "languages": [],
            "access": "read-write",
            "uses_skills": [],
            "uses_rules": [],
            "uses_plugins": [],
            "delegates_to": [],
        }
        content = _generate_agent_content(fields, "test", "read-write")
        for section in ["## Role", "## Activation", "## Capabilities",
                        "## Standards", "## Workflow", "## Output Format", "## Constraints"]:
            assert section in content


# ---------------------------------------------------------------------------
# Integration tests — full scaffold flow with monkeypatched input
# ---------------------------------------------------------------------------


class TestScaffoldRuleIntegration:
    """Integration tests for scaffold_flow('rule')."""

    def test_happy_path_flat_rule(
        self,
        sample_tree: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Create a flat rule with all defaults accepted."""
        responses = iter([
            "new-patterns",              # name
            "",                          # description (accept default)
            "",                          # scope (accept default: universal)
            "",                          # uses_rules (skip)
            "y",                         # confirm
        ])
        monkeypatch.setattr("builtins.input", lambda _: next(responses))

        scaffold_flow(sample_tree, "rule")

        created = sample_tree / "rules" / "new-patterns.md"
        assert created.exists()
        content = created.read_text(encoding="utf-8")
        assert "name: new-patterns" in content
        assert "scope: universal" in content
        assert "# New Patterns Patterns" in content

    def test_nested_rule_creates_subdirectory(
        self,
        sample_tree: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Create a nested rule (go/new-thing) that creates a subdirectory."""
        responses = iter([
            "go/new-thing",              # name
            "",                          # description (accept default)
            "2",                         # scope: language-specific
            "1",                         # languages: go
            "",                          # uses_rules (skip)
            "y",                         # confirm
        ])
        monkeypatch.setattr("builtins.input", lambda _: next(responses))

        scaffold_flow(sample_tree, "rule")

        created = sample_tree / "rules" / "go" / "new-thing.md"
        assert created.exists()
        content = created.read_text(encoding="utf-8")
        assert "name: go/new-thing" in content
        assert "scope: language-specific" in content
        assert "languages: [go]" in content

    def test_duplicate_aborts(
        self,
        sample_tree: Path,
        monkeypatch: pytest.MonkeyPatch,
        capsys: pytest.CaptureFixture[str],
    ) -> None:
        """Attempting to create an existing rule aborts."""
        responses = iter([
            "security",                  # name (already exists)
        ])
        monkeypatch.setattr("builtins.input", lambda _: next(responses))

        scaffold_flow(sample_tree, "rule")

        captured = capsys.readouterr()
        assert "already exists" in captured.out

    def test_user_declines_confirmation(
        self,
        sample_tree: Path,
        monkeypatch: pytest.MonkeyPatch,
        capsys: pytest.CaptureFixture[str],
    ) -> None:
        """User says 'n' at confirmation — no file created."""
        responses = iter([
            "declined-rule",             # name
            "",                          # description (default)
            "",                          # scope (default)
            "",                          # uses_rules (skip)
            "n",                         # decline
        ])
        monkeypatch.setattr("builtins.input", lambda _: next(responses))

        scaffold_flow(sample_tree, "rule")

        assert not (sample_tree / "rules" / "declined-rule.md").exists()
        captured = capsys.readouterr()
        assert "Aborted" in captured.out


class TestScaffoldSkillIntegration:
    """Integration tests for scaffold_flow('skill')."""

    def test_happy_path(
        self,
        sample_tree: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        responses = iter([
            "new-skill",                 # name
            "",                          # description (default)
            "",                          # scope (default: universal)
            "",                          # uses_rules (skip)
            "y",                         # confirm
        ])
        monkeypatch.setattr("builtins.input", lambda _: next(responses))

        scaffold_flow(sample_tree, "skill")

        created = sample_tree / "skills" / "new-skill.md"
        assert created.exists()
        content = created.read_text(encoding="utf-8")
        assert "name: new-skill" in content
        assert "## When to Use" in content


class TestScaffoldAgentIntegration:
    """Integration tests for scaffold_flow('agent')."""

    def test_agent_with_dependencies(
        self,
        sample_tree: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Create an agent selecting dependencies from available entities."""
        responses = iter([
            "new-agent",                 # name
            "",                          # description (default)
            "",                          # role (default: coder)
            "",                          # scope (default: universal)
            "",                          # access (default: read-write for coder)
            "1",                         # uses_skills: first available
            "",                          # uses_rules (skip)
            "",                          # uses_plugins (skip)
            "",                          # delegates_to (skip)
            "y",                         # confirm
        ])
        monkeypatch.setattr("builtins.input", lambda _: next(responses))

        scaffold_flow(sample_tree, "agent")

        created = sample_tree / "agents" / "new-agent.md"
        assert created.exists()
        content = created.read_text(encoding="utf-8")
        assert "name: new-agent" in content
        assert "role: coder" in content
        assert "access: read-write" in content

    def test_read_only_agent(
        self,
        sample_tree: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Reviewer role defaults to read-only and gets correct capabilities."""
        responses = iter([
            "new-reviewer",              # name
            "",                          # description (default)
            "2",                         # role: reviewer
            "",                          # scope (default: universal)
            "",                          # access (default: read-only for reviewer)
            "",                          # uses_skills (skip)
            "",                          # uses_rules (skip)
            "",                          # uses_plugins (skip)
            "",                          # delegates_to (skip)
            "y",                         # confirm
        ])
        monkeypatch.setattr("builtins.input", lambda _: next(responses))

        scaffold_flow(sample_tree, "agent")

        created = sample_tree / "agents" / "new-reviewer.md"
        assert created.exists()
        content = created.read_text(encoding="utf-8")
        assert "access: read-only" in content
        assert "Cannot modify any files" in content
