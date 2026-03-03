"""Tests for agent_army.resolver."""

from __future__ import annotations

from pathlib import Path
from textwrap import dedent

import pytest

from agent_army.models import Agent, Fix, Rule, Skill, ValidationError
from agent_army.resolver import (
    apply_fixes,
    compute_all_fixes,
    format_report,
    validate_all_refs,
)


class TestValidateAllRefs:
    """validate_all_refs() — reference existence checks."""

    def test_all_valid(self) -> None:
        rules = [Rule(name="a", description="", scope="universal")]
        skills = [
            Skill(name="s1", description="", scope="universal", uses_rules=["a"])
        ]
        agents: list[Agent] = []
        errors = validate_all_refs(rules, skills, agents, [])
        assert errors == []

    def test_missing_rule_ref(self) -> None:
        rules = [Rule(name="a", description="", scope="universal")]
        skills = [
            Skill(
                name="s1",
                description="",
                scope="universal",
                uses_rules=["nonexistent"],
                path=Path("skills/s1.md"),
            )
        ]
        errors = validate_all_refs(rules, skills, [], [])
        assert len(errors) == 1
        assert errors[0].severity == "error"
        assert errors[0].ref == "nonexistent"

    def test_missing_plugin_is_warning(self) -> None:
        agents = [
            Agent(
                name="a1",
                description="",
                role="coder",
                scope="universal",
                access="read-write",
                uses_plugins=["unknown-plugin"],
                path=Path("agents/a1.md"),
            )
        ]
        errors = validate_all_refs([], [], agents, ["known-plugin"])
        assert len(errors) == 1
        assert errors[0].severity == "warning"

    def test_missing_skill_ref(self) -> None:
        agents = [
            Agent(
                name="a1",
                description="",
                role="coder",
                scope="universal",
                access="read-write",
                uses_skills=["nonexistent-skill"],
                path=Path("agents/a1.md"),
            )
        ]
        errors = validate_all_refs([], [], agents, [])
        assert len(errors) == 1
        assert errors[0].field == "uses_skills"

    def test_missing_delegates_ref(self) -> None:
        agents = [
            Agent(
                name="a1",
                description="",
                role="coder",
                scope="universal",
                access="read-write",
                delegates_to=["ghost-agent"],
                path=Path("agents/a1.md"),
            )
        ]
        errors = validate_all_refs([], [], agents, [])
        assert len(errors) == 1
        assert errors[0].field == "delegates_to"


class TestComputeAllFixes:
    """compute_all_fixes() — redundancy detection and fix computation."""

    def test_no_redundancies(self) -> None:
        rules = [
            Rule(name="a", description="", scope="universal"),
            Rule(name="b", description="", scope="universal"),
        ]
        skills = [
            Skill(
                name="s1",
                description="",
                scope="universal",
                uses_rules=["a", "b"],
            )
        ]
        fixes = compute_all_fixes(rules, skills, [], Path("/tmp"))
        assert fixes == []

    def test_rule_to_rule_redundancy(self) -> None:
        rules = [
            Rule(
                name="a",
                description="",
                scope="universal",
                uses_rules=["b"],
                path=Path("rules/a.md"),
            ),
            Rule(name="b", description="", scope="universal", path=Path("rules/b.md")),
        ]
        # Skill with [a, b] — b is covered by a
        skills = [
            Skill(
                name="s1",
                description="",
                scope="universal",
                uses_rules=["a", "b"],
                path=Path("skills/s1.md"),
            )
        ]
        fixes = compute_all_fixes(rules, skills, [], Path("/tmp"))
        assert len(fixes) == 1
        assert "b" not in fixes[0].after


class TestApplyFixes:
    """apply_fixes() — write fixes to disk."""

    def test_writes_field(self, tmp_path: Path) -> None:
        md = tmp_path / "rules" / "test.md"
        md.parent.mkdir(parents=True)
        md.write_text("---\nuses_rules: [a, b, c]\n---\n\n# Test\n")

        fix = Fix(
            label="rules/test.md",
            field="uses_rules",
            file_path=Path("rules/test.md"),
            before=["a", "b", "c"],
            after=["a", "c"],
            reasons=['"b" covered by "a"'],
        )
        apply_fixes([fix], tmp_path)

        content = md.read_text()
        assert "uses_rules: [a, c]" in content
        assert "b" not in content.split("---")[1]


class TestFormatReport:
    """format_report() — human-readable output."""

    def test_all_clean(self) -> None:
        result = format_report([], [])
        assert "All dependency references are valid" in result

    def test_errors_section(self) -> None:
        errors = [
            ValidationError(
                file_label="rules/x.md",
                field="uses_rules",
                ref="missing",
                severity="error",
            )
        ]
        result = format_report(errors, [])
        assert "[ERROR]" in result
        assert "missing" in result
        assert "Fix errors above" in result

    def test_warnings_section(self) -> None:
        errors = [
            ValidationError(
                file_label="agents/x.md",
                field="uses_plugins",
                ref="unknown",
                severity="warning",
            )
        ]
        result = format_report(errors, [])
        assert "[WARN]" in result
        assert "unknown" in result

    def test_fixes_section(self) -> None:
        fixes = [
            Fix(
                label="skills/s1.md",
                field="uses_rules",
                file_path=Path("skills/s1.md"),
                before=["a", "b"],
                after=["a"],
                reasons=['"b" covered by "a"'],
            )
        ]
        result = format_report([], fixes)
        assert "[FIX]" in result
        assert "Before: [a, b]" in result
        assert "After:  [a]" in result

    def test_summary_counts(self) -> None:
        errors = [
            ValidationError("x", "uses_rules", "y", "error"),
            ValidationError("x", "uses_rules", "z", "error"),
        ]
        fixes = [
            Fix("s", "uses_rules", Path("s.md"), ["a"], [], ['"a" covered by "b"'])
        ]
        result = format_report(errors, fixes)
        assert "2 error(s)" in result
        assert "1 fixable" in result
