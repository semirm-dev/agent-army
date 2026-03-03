"""BFS transitive resolution and redundancy detection for dependency graphs.

Ports the BFS algorithms from lib-deps.sh into a single generic module.
The bash originals had separate but identical functions for rules vs delegates;
this module unifies them via a callback-based design.
"""

from __future__ import annotations

from collections import deque
from collections.abc import Callable

from agent_army.models import Redundancy


def resolve_transitive(
    seeds: list[str],
    get_deps: Callable[[str], list[str]],
) -> list[str]:
    """BFS transitive resolution.

    Walks the dependency graph breadth-first starting from *seeds*,
    using *get_deps* to discover neighbors. Handles cycles via a visited set.

    Args:
        seeds: Initial items to resolve.
        get_deps: Callback returning direct dependencies for a given name.

    Returns:
        Deduplicated list in BFS discovery order (preserves insertion order).
    """
    if not seeds:
        return []

    visited: set[str] = set()
    result: list[str] = []
    queue: deque[str] = deque(seeds)

    while queue:
        current = queue.popleft()
        if current in visited:
            continue
        visited.add(current)
        result.append(current)
        queue.extend(get_deps(current))

    return result


def find_redundant(
    entries: list[str],
    get_deps: Callable[[str], list[str]],
) -> list[Redundancy]:
    """Find entries that are transitively covered by other entries in the same list.

    For each entry R, checks whether any *other* entry O in the list has R
    in its transitive closure. Returns at most one Redundancy per target
    (the first covering entry found).

    Args:
        entries: List of entry names to check for redundancy.
        get_deps: Callback returning direct dependencies for a given name.

    Returns:
        List of Redundancy(target=redundant_entry, covered_by=covering_entry).
    """
    if not entries:
        return []

    redundancies: list[Redundancy] = []

    for i, target in enumerate(entries):
        for j, other in enumerate(entries):
            if i == j:
                continue
            closure = resolve_transitive([other], get_deps)
            if target in closure:
                redundancies.append(Redundancy(target=target, covered_by=other))
                break  # one cover is enough per target

    return redundancies


def rules_covered_by_skills(
    skill_names: list[str],
    skill_lookup: dict[str, list[str]],
    rule_lookup: dict[str, list[str]],
) -> set[str]:
    """Compute transitive closure of all rules covered by a set of skills.

    Collects the direct uses_rules from each skill, unions them, then
    resolves the full transitive closure through the rule dependency graph.

    Args:
        skill_names: Skills whose rule coverage to compute.
        skill_lookup: Mapping of skill_name to its direct uses_rules.
        rule_lookup: Mapping of rule_name to its direct uses_rules.

    Returns:
        Set of all rule names transitively covered by the given skills.
    """
    union: list[str] = []
    for skill in skill_names:
        union.extend(skill_lookup.get(skill, []))

    if not union:
        return set()

    def _get_rule_deps(name: str) -> list[str]:
        return rule_lookup.get(name, [])

    resolved = resolve_transitive(union, _get_rule_deps)
    return set(resolved)


def find_redundant_via_skills(
    rule_entries: list[str],
    skill_entries: list[str],
    skill_lookup: dict[str, list[str]],
    rule_lookup: dict[str, list[str]],
) -> list[Redundancy]:
    """Find rules that are already covered transitively by skills.

    Computes the full rule coverage of the skill set, then checks each
    rule entry against that coverage. For each covered rule, identifies
    the first skill that provides it.

    Args:
        rule_entries: Rule names to check for redundancy.
        skill_entries: Skill names that may cover the rules.
        skill_lookup: Mapping of skill_name to its direct uses_rules.
        rule_lookup: Mapping of rule_name to its direct uses_rules.

    Returns:
        List of Redundancy where covered_by is ``"skill <name>"``.
    """
    if not rule_entries or not skill_entries:
        return []

    covered = rules_covered_by_skills(skill_entries, skill_lookup, rule_lookup)
    if not covered:
        return []

    def _get_rule_deps(name: str) -> list[str]:
        return rule_lookup.get(name, [])

    redundancies: list[Redundancy] = []

    for rule in rule_entries:
        if rule not in covered:
            continue
        # Find the first skill whose transitive closure includes this rule
        for skill in skill_entries:
            skill_rules = skill_lookup.get(skill, [])
            if not skill_rules:
                continue
            skill_closure = resolve_transitive(skill_rules, _get_rule_deps)
            if rule in skill_closure:
                redundancies.append(
                    Redundancy(target=rule, covered_by=f"skill {skill}")
                )
                break

    return redundancies
