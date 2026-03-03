"""Generate manifest.json from rules/, skills/, and agents/ frontmatter.

Produces output identical to ``scripts/generate-manifest.sh``: compact
one-object-per-line entries with 2-space section indent and 4-space item
indent.
"""

from __future__ import annotations

import json
from pathlib import Path

from agent_army.graph import resolve_transitive
from agent_army.loader import load_agents, load_rules, load_skills
from agent_army.models import Agent, Rule, Skill


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------


def generate_manifest(root: Path) -> dict:
    """Load all entities, resolve transitive deps, build manifest dict.

    Args:
        root: Repository root containing rules/, skills/, agents/ dirs.

    Returns:
        Manifest dict with "rules", "skills", and "agents" keys.
    """
    rules = load_rules(root)
    skills = load_skills(root)
    agents = load_agents(root)

    rule_lookup = _build_rule_lookup(rules)
    skill_lookup = _build_skill_lookup(skills)
    agent_lookup = _build_agent_lookup(agents)

    return {
        "rules": [_rule_entry(r, rule_lookup) for r in rules],
        "skills": [_skill_entry(s, rule_lookup) for s in skills],
        "agents": [
            _agent_entry(a, rule_lookup, skill_lookup, agent_lookup)
            for a in agents
        ],
    }


def write_manifest(root: Path) -> None:
    """Generate manifest and write to ``root/manifest.json``.

    Uses the custom compact JSON format matching the bash script output.
    """
    manifest = generate_manifest(root)
    output = _format_manifest_json(manifest)
    target = root / "manifest.json"
    target.write_text(output, encoding="utf-8")


# ---------------------------------------------------------------------------
# Entry builders
# ---------------------------------------------------------------------------


def _rule_entry(rule: Rule, rule_lookup: dict[str, list[str]]) -> dict:
    """Build a single rule manifest entry with transitive uses_rules."""
    resolved = _resolve_rule_deps(rule.uses_rules, rule_lookup)

    entry: dict = {
        "name": rule.name,
        "scope": rule.scope,
    }

    if rule.scope == "language-specific" and rule.languages:
        entry["languages"] = rule.languages

    if resolved:
        entry["uses_rules"] = resolved

    entry["path"] = str(rule.path)
    return entry


def _skill_entry(skill: Skill, rule_lookup: dict[str, list[str]]) -> dict:
    """Build a single skill manifest entry with transitive uses_rules."""
    resolved = _resolve_rule_deps(skill.uses_rules, rule_lookup)

    entry: dict = {
        "name": skill.name,
        "scope": skill.scope,
    }

    if skill.scope == "language-specific" and skill.languages:
        entry["languages"] = skill.languages

    # Skills always include uses_rules, even when empty
    entry["uses_rules"] = resolved
    entry["path"] = str(skill.path)
    return entry


def _agent_entry(
    agent: Agent,
    rule_lookup: dict[str, list[str]],
    skill_lookup: dict[str, list[str]],
    agent_lookup: dict[str, list[str]],
) -> dict:
    """Build a single agent manifest entry with transitive resolution."""
    combined_rules = _merge_agent_rules(agent, skill_lookup)
    resolved_rules = _resolve_rule_deps(combined_rules, rule_lookup)
    resolved_delegates = _resolve_agent_delegates(
        agent.delegates_to, agent_lookup
    )

    entry: dict = {
        "name": agent.name,
        "role": agent.role,
        "scope": agent.scope,
        "access": agent.access,
    }

    if agent.languages:
        entry["languages"] = agent.languages

    entry["uses_skills"] = agent.uses_skills
    entry["uses_rules"] = resolved_rules
    entry["uses_plugins"] = agent.uses_plugins
    entry["delegates_to"] = resolved_delegates
    entry["path"] = str(agent.path)
    return entry


# ---------------------------------------------------------------------------
# Lookup builders
# ---------------------------------------------------------------------------


def _build_rule_lookup(rules: list[Rule]) -> dict[str, list[str]]:
    """Map rule name to its direct uses_rules."""
    return {r.name: r.uses_rules for r in rules}


def _build_skill_lookup(skills: list[Skill]) -> dict[str, list[str]]:
    """Map skill name to its direct uses_rules."""
    return {s.name: s.uses_rules for s in skills}


def _build_agent_lookup(agents: list[Agent]) -> dict[str, list[str]]:
    """Map agent name to its direct delegates_to."""
    return {a.name: a.delegates_to for a in agents}


# ---------------------------------------------------------------------------
# Resolution helpers
# ---------------------------------------------------------------------------


def _resolve_rule_deps(
    seeds: list[str],
    rule_lookup: dict[str, list[str]],
) -> list[str]:
    """Resolve rule dependencies transitively via BFS."""
    if not seeds:
        return []

    def _get_deps(name: str) -> list[str]:
        return rule_lookup.get(name, [])

    return resolve_transitive(seeds, _get_deps)


def _resolve_agent_delegates(
    seeds: list[str],
    agent_lookup: dict[str, list[str]],
) -> list[str]:
    """Resolve agent delegations transitively via BFS."""
    if not seeds:
        return []

    def _get_deps(name: str) -> list[str]:
        return agent_lookup.get(name, [])

    return resolve_transitive(seeds, _get_deps)


def _merge_agent_rules(
    agent: Agent,
    skill_lookup: dict[str, list[str]],
) -> list[str]:
    """Merge agent's own uses_rules with rules from its skills.

    Collects direct uses_rules from each skill in uses_skills,
    then appends the agent's own direct uses_rules. Order matches
    the bash script: agent's own rules first, then skill rules.
    """
    own_rules = list(agent.uses_rules)
    skill_rules: list[str] = []
    for skill_name in agent.uses_skills:
        skill_rules.extend(skill_lookup.get(skill_name, []))

    if own_rules and skill_rules:
        return own_rules + skill_rules
    if skill_rules:
        return skill_rules
    return own_rules


# ---------------------------------------------------------------------------
# JSON formatting
# ---------------------------------------------------------------------------


def _format_entry(entry: dict) -> str:
    """Format a single manifest entry as ``{ "key": "val", ... }``.

    Matches the bash script output: spaces inside braces, compact
    key-value pairs separated by ``", "``.
    """
    pairs: list[str] = []
    for key, value in entry.items():
        encoded_value = json.dumps(value, ensure_ascii=False, separators=(", ", ": "))
        pairs.append(f'"{key}": {encoded_value}')
    return "{ " + ", ".join(pairs) + " }"


def _format_manifest_json(manifest: dict) -> str:
    """Format manifest dict as JSON matching the bash script output.

    Produces compact one-object-per-line entries:
    - 2-space indent for section keys
    - 4-space indent for array items
    - Trailing newline at end of file
    """
    lines: list[str] = ["{"]

    sections = list(manifest.keys())
    for sec_idx, section in enumerate(sections):
        entries = manifest[section]
        is_last_section = sec_idx == len(sections) - 1
        section_suffix = "" if is_last_section else ","

        lines.append(f'  "{section}": [')
        for i, entry in enumerate(entries):
            comma = "" if i == len(entries) - 1 else ","
            compact = _format_entry(entry)
            lines.append(f"    {compact}{comma}")
        lines.append(f"  ]{section_suffix}")

    lines.append("}")
    return "\n".join(lines) + "\n"
