"""Tests for agent_army.frontmatter."""

from __future__ import annotations

from pathlib import Path
from textwrap import dedent

import pytest

from agent_army.frontmatter import extract_h1, parse_frontmatter, write_field


class TestParseFrontmatter:
    """parse_frontmatter() — scalar values and list formats."""

    @pytest.mark.parametrize(
        "content, expected",
        [
            pytest.param(
                "---\nname: api-design\nscope: universal\n---\n# Title\n",
                {"name": "api-design", "scope": "universal"},
                id="scalars",
            ),
            pytest.param(
                '---\ndescription: "Senior Go engineer. Writes code."\n---\n',
                {"description": "Senior Go engineer. Writes code."},
                id="quoted-scalar-with-colon",
            ),
            pytest.param(
                "---\nuses_rules: [api-design, cross-cutting, security]\n---\n",
                {"uses_rules": ["api-design", "cross-cutting", "security"]},
                id="inline-list",
            ),
            pytest.param(
                "---\nuses_rules: []\n---\n",
                {"uses_rules": []},
                id="empty-inline-list",
            ),
            pytest.param(
                "---\nlanguages: [go]\n---\n",
                {"languages": ["go"]},
                id="single-item-list",
            ),
            pytest.param(
                "---\nuses_rules:\n  - api-design\n  - security\n---\n",
                {"uses_rules": ["api-design", "security"]},
                id="block-list",
            ),
            pytest.param(
                "no frontmatter here\n",
                {},
                id="no-frontmatter",
            ),
            pytest.param(
                "---\n---\n",
                {},
                id="empty-frontmatter",
            ),
        ],
    )
    def test_parse(self, content: str, expected: dict) -> None:
        assert parse_frontmatter(content) == expected

    def test_full_agent_frontmatter(self, sample_agent_content: str) -> None:
        fm = parse_frontmatter(sample_agent_content)
        assert fm["name"] == "go/coder"
        assert fm["role"] == "coder"
        assert fm["scope"] == "language-specific"
        assert fm["access"] == "read-write"
        assert fm["languages"] == ["go"]
        assert "go/coder" in fm["uses_skills"]
        assert fm["uses_rules"] == []
        assert fm["uses_plugins"] == ["code-simplifier", "context7"]
        assert fm["delegates_to"] == []

    def test_full_rule_frontmatter(self, sample_rule_content: str) -> None:
        fm = parse_frontmatter(sample_rule_content)
        assert fm["name"] == "go/patterns"
        assert fm["scope"] == "language-specific"
        assert fm["languages"] == ["go"]
        assert fm["uses_rules"] == [
            "code-quality",
            "security",
            "cross-cutting",
            "observability",
        ]


class TestExtractH1:
    """extract_h1() — first heading after frontmatter."""

    @pytest.mark.parametrize(
        "content, expected",
        [
            pytest.param(
                "---\nname: foo\n---\n\n# My Title\n\nContent.",
                "My Title",
                id="normal",
            ),
            pytest.param(
                "---\n---\n\n# First\n\n# Second\n",
                "First",
                id="picks-first-h1",
            ),
            pytest.param(
                "---\nname: foo\n---\n\nNo heading here.\n",
                "",
                id="no-h1",
            ),
            pytest.param(
                "# Heading Without Frontmatter\n",
                "Heading Without Frontmatter",
                id="no-frontmatter-finds-h1",
            ),
        ],
    )
    def test_extract(self, content: str, expected: str) -> None:
        assert extract_h1(content) == expected


class TestWriteField:
    """write_field() — replace or insert frontmatter field."""

    def test_replace_existing_field(self, tmp_path: Path) -> None:
        md = tmp_path / "test.md"
        md.write_text(
            "---\nname: foo\nuses_rules: [old-rule]\n---\n\n# Title\n"
        )
        write_field(md, "uses_rules", ["new-a", "new-b"])
        result = md.read_text()
        assert "uses_rules: [new-a, new-b]" in result
        assert "old-rule" not in result

    def test_insert_missing_field(self, tmp_path: Path) -> None:
        md = tmp_path / "test.md"
        md.write_text("---\nname: foo\n---\n\n# Title\n")
        write_field(md, "uses_rules", ["alpha", "beta"])
        result = md.read_text()
        assert "uses_rules: [alpha, beta]" in result
        # Field inserted before closing ---
        lines = result.splitlines()
        field_idx = next(i for i, l in enumerate(lines) if "uses_rules:" in l)
        close_idx = next(
            i for i in range(field_idx, len(lines)) if lines[i].strip() == "---"
        )
        assert close_idx == field_idx + 1

    def test_write_empty_list(self, tmp_path: Path) -> None:
        md = tmp_path / "test.md"
        md.write_text("---\nuses_rules: [a, b]\n---\n")
        write_field(md, "uses_rules", [])
        assert "uses_rules: []" in md.read_text()

    def test_atomic_write(self, tmp_path: Path) -> None:
        md = tmp_path / "test.md"
        md.write_text("---\nname: foo\n---\n")
        write_field(md, "uses_rules", ["x"])
        # No leftover .tmp file
        assert not (tmp_path / "test.md.tmp").exists()
