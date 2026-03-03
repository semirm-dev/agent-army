"""Tests for agent_army.loader."""

from __future__ import annotations

from pathlib import Path

from agent_army.loader import (
    find_md_files,
    load_agents,
    load_plugins,
    load_rules,
    load_skills,
)


class TestFindMdFiles:
    """find_md_files() — recursive sorted discovery."""

    def test_finds_nested_files(self, sample_tree: Path) -> None:
        files = find_md_files(sample_tree / "rules")
        names = [f.name for f in files]
        assert "cross-cutting.md" in names
        assert "patterns.md" in names

    def test_sorted_by_path(self, sample_tree: Path) -> None:
        files = find_md_files(sample_tree / "rules")
        paths = [str(f) for f in files]
        assert paths == sorted(paths)

    def test_empty_dir(self, tmp_path: Path) -> None:
        empty = tmp_path / "empty"
        empty.mkdir()
        assert find_md_files(empty) == []


class TestLoadRules:
    """load_rules() — name from path, description from H1."""

    def test_loads_rules(self, sample_tree: Path) -> None:
        rules = load_rules(sample_tree)
        names = [r.name for r in rules]
        assert "cross-cutting" in names
        assert "go/patterns" in names

    def test_rule_name_from_path(self, sample_tree: Path) -> None:
        rules = load_rules(sample_tree)
        go_rule = next(r for r in rules if r.name == "go/patterns")
        assert go_rule.scope == "language-specific"
        assert go_rule.languages == ["go"]
        assert "code-quality" in go_rule.uses_rules

    def test_rule_description_from_h1(self, sample_tree: Path) -> None:
        rules = load_rules(sample_tree)
        cc = next(r for r in rules if r.name == "cross-cutting")
        assert cc.description == "Cross-Cutting Standards"

    def test_rule_path_relative(self, sample_tree: Path) -> None:
        rules = load_rules(sample_tree)
        go_rule = next(r for r in rules if r.name == "go/patterns")
        assert str(go_rule.path) == "rules/go/patterns.md"

    def test_missing_dir(self, tmp_path: Path) -> None:
        assert load_rules(tmp_path / "nonexistent") == []


class TestLoadSkills:
    """load_skills() — name from frontmatter, description from H1."""

    def test_loads_skills(self, sample_tree: Path) -> None:
        skills = load_skills(sample_tree)
        names = [s.name for s in skills]
        assert "error-handling" in names
        assert "go/coder" in names

    def test_skill_name_from_frontmatter(self, sample_tree: Path) -> None:
        skills = load_skills(sample_tree)
        eh = next(s for s in skills if s.name == "error-handling")
        assert eh.scope == "universal"
        assert "cross-cutting" in eh.uses_rules

    def test_skill_path_relative(self, sample_tree: Path) -> None:
        skills = load_skills(sample_tree)
        eh = next(s for s in skills if s.name == "error-handling")
        assert str(eh.path) == "skills/error-handling.md"


class TestLoadAgents:
    """load_agents() — name from frontmatter, description from frontmatter."""

    def test_loads_agents(self, sample_tree: Path) -> None:
        agents = load_agents(sample_tree)
        names = [a.name for a in agents]
        assert "go/coder" in names
        assert "type-design-analyzer" in names

    def test_agent_fields(self, sample_tree: Path) -> None:
        agents = load_agents(sample_tree)
        gc = next(a for a in agents if a.name == "go/coder")
        assert gc.role == "coder"
        assert gc.access == "read-write"
        assert gc.scope == "language-specific"
        assert gc.languages == ["go"]
        assert "go/coder" in gc.uses_skills
        assert "code-simplifier" in gc.uses_plugins

    def test_agent_description_from_frontmatter(self, sample_tree: Path) -> None:
        agents = load_agents(sample_tree)
        gc = next(a for a in agents if a.name == "go/coder")
        assert "Go engineer" in gc.description


class TestLoadPlugins:
    """load_plugins() — from config.json."""

    def test_loads_plugins(self, sample_tree: Path) -> None:
        plugins = load_plugins(sample_tree)
        assert "code-simplifier" in plugins
        assert "context7" in plugins

    def test_missing_config(self, tmp_path: Path) -> None:
        assert load_plugins(tmp_path) == []
