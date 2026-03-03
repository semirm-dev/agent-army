"""Tests for agent_army.manifest."""

from __future__ import annotations

import json
from pathlib import Path

from agent_army.manifest import generate_manifest


class TestGenerateManifest:
    """generate_manifest() — build manifest dict from sample tree."""

    def test_has_all_sections(self, sample_tree: Path) -> None:
        manifest = generate_manifest(sample_tree)
        assert "rules" in manifest
        assert "skills" in manifest
        assert "agents" in manifest

    def test_rule_entry_format(self, sample_tree: Path) -> None:
        manifest = generate_manifest(sample_tree)
        go_rule = next(
            r for r in manifest["rules"] if r["name"] == "go/patterns"
        )
        assert go_rule["scope"] == "language-specific"
        assert go_rule["languages"] == ["go"]
        assert "code-quality" in go_rule["uses_rules"]
        assert go_rule["path"] == "rules/go/patterns.md"

    def test_rule_omits_empty_optional_fields(self, sample_tree: Path) -> None:
        manifest = generate_manifest(sample_tree)
        cc = next(r for r in manifest["rules"] if r["name"] == "cross-cutting")
        assert "languages" not in cc
        assert "uses_rules" not in cc

    def test_skill_always_has_uses_rules(self, sample_tree: Path) -> None:
        manifest = generate_manifest(sample_tree)
        for skill in manifest["skills"]:
            assert "uses_rules" in skill

    def test_agent_always_has_all_fields(self, sample_tree: Path) -> None:
        manifest = generate_manifest(sample_tree)
        for agent in manifest["agents"]:
            assert "uses_skills" in agent
            assert "uses_rules" in agent
            assert "uses_plugins" in agent
            assert "delegates_to" in agent

    def test_agent_transitive_rules(self, sample_tree: Path) -> None:
        manifest = generate_manifest(sample_tree)
        gc = next(a for a in manifest["agents"] if a["name"] == "go/coder")
        # go/coder uses skills [go/coder, error-handling]
        # go/coder skill -> go/patterns -> [code-quality, security, cross-cutting]
        # error-handling skill -> cross-cutting
        assert "go/patterns" in gc["uses_rules"]
        assert "code-quality" in gc["uses_rules"]
        assert "cross-cutting" in gc["uses_rules"]

    def test_agent_transitive_delegates(self, sample_tree: Path) -> None:
        manifest = generate_manifest(sample_tree)
        gc = next(a for a in manifest["agents"] if a["name"] == "go/coder")
        assert "type-design-analyzer" in gc["delegates_to"]


class TestGoldenFile:
    """Golden-file comparison against the real manifest.json."""

    def test_matches_current_manifest(self) -> None:
        """Compare Python-generated manifest against the checked-in one."""
        repo_root = Path(__file__).resolve().parent.parent.parent
        manifest_path = repo_root / "manifest.json"
        if not manifest_path.exists():
            return

        manifest = generate_manifest(repo_root)

        current = json.loads(manifest_path.read_text(encoding="utf-8"))

        assert len(manifest["rules"]) == len(current["rules"])
        assert len(manifest["skills"]) == len(current["skills"])
        assert len(manifest["agents"]) == len(current["agents"])

        for gen, cur in zip(manifest["rules"], current["rules"]):
            assert gen == cur, f"Rule mismatch: {gen['name']}"

        for gen, cur in zip(manifest["skills"], current["skills"]):
            assert gen == cur, f"Skill mismatch: {gen['name']}"

        for gen, cur in zip(manifest["agents"], current["agents"]):
            assert gen == cur, f"Agent mismatch: {gen['name']}"
