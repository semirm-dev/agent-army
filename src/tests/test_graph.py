"""Tests for agent_army.graph."""

from __future__ import annotations

import pytest

from agent_army.graph import (
    find_redundant,
    find_redundant_via_skills,
    resolve_transitive,
    rules_covered_by_skills,
)
from agent_army.models import Redundancy


class TestResolveTransitive:
    """resolve_transitive() — BFS with deduplication."""

    @pytest.mark.parametrize(
        "seeds, deps, expected",
        [
            pytest.param([], {}, [], id="empty-seeds"),
            pytest.param(
                ["a"],
                {},
                ["a"],
                id="single-no-deps",
            ),
            pytest.param(
                ["a"],
                {"a": ["b"], "b": ["c"]},
                ["a", "b", "c"],
                id="linear-chain",
            ),
            pytest.param(
                ["a"],
                {"a": ["b", "c"], "b": ["d"], "c": ["d"]},
                ["a", "b", "c", "d"],
                id="diamond",
            ),
            pytest.param(
                ["a"],
                {"a": ["b"], "b": ["a"]},
                ["a", "b"],
                id="cycle",
            ),
            pytest.param(
                ["a", "b"],
                {"a": ["c"], "b": ["c"]},
                ["a", "b", "c"],
                id="multiple-seeds-shared-dep",
            ),
        ],
    )
    def test_resolve(
        self,
        seeds: list[str],
        deps: dict[str, list[str]],
        expected: list[str],
    ) -> None:
        result = resolve_transitive(seeds, lambda n: deps.get(n, []))
        assert result == expected


class TestFindRedundant:
    """find_redundant() — detect entries covered by other entries."""

    def test_no_redundancy(self) -> None:
        deps = {"a": ["x"], "b": ["y"]}
        result = find_redundant(["a", "b"], lambda n: deps.get(n, []))
        assert result == []

    def test_simple_redundancy(self) -> None:
        # b depends on a, so a is redundant when b is also in the list
        deps = {"b": ["a"]}
        result = find_redundant(["a", "b"], lambda n: deps.get(n, []))
        assert len(result) == 1
        assert result[0].target == "a"
        assert result[0].covered_by == "b"

    def test_chain_redundancy(self) -> None:
        # c -> b -> a. If all three are listed, a is covered by b or c
        deps = {"c": ["b"], "b": ["a"]}
        result = find_redundant(["a", "b", "c"], lambda n: deps.get(n, []))
        targets = {r.target for r in result}
        # a is covered by b (or c), b is covered by c
        assert "a" in targets
        assert "b" in targets

    def test_empty_list(self) -> None:
        assert find_redundant([], lambda n: []) == []

    def test_single_entry(self) -> None:
        assert find_redundant(["a"], lambda n: []) == []


class TestRulesCoveredBySkills:
    """rules_covered_by_skills() — transitive closure from skills."""

    def test_basic(self) -> None:
        skill_lookup = {"go/coder": ["go/patterns"]}
        rule_lookup = {"go/patterns": ["code-quality", "security"]}
        result = rules_covered_by_skills(
            ["go/coder"], skill_lookup, rule_lookup
        )
        assert "go/patterns" in result
        assert "code-quality" in result
        assert "security" in result

    def test_empty_skills(self) -> None:
        assert rules_covered_by_skills([], {}, {}) == set()

    def test_skill_with_no_rules(self) -> None:
        skill_lookup = {"empty-skill": []}
        assert rules_covered_by_skills(["empty-skill"], skill_lookup, {}) == set()


class TestFindRedundantViaSkills:
    """find_redundant_via_skills() — rules covered by skills."""

    def test_rule_covered_by_skill(self) -> None:
        skill_lookup = {"go/coder": ["go/patterns"]}
        rule_lookup = {"go/patterns": ["code-quality"]}
        result = find_redundant_via_skills(
            rule_entries=["code-quality"],
            skill_entries=["go/coder"],
            skill_lookup=skill_lookup,
            rule_lookup=rule_lookup,
        )
        assert len(result) == 1
        assert result[0].target == "code-quality"
        assert "skill" in result[0].covered_by

    def test_no_overlap(self) -> None:
        skill_lookup = {"go/coder": ["go/patterns"]}
        rule_lookup = {}
        result = find_redundant_via_skills(
            rule_entries=["unrelated"],
            skill_entries=["go/coder"],
            skill_lookup=skill_lookup,
            rule_lookup=rule_lookup,
        )
        assert result == []

    def test_empty_inputs(self) -> None:
        assert find_redundant_via_skills([], [], {}, {}) == []
