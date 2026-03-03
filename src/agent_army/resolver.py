"""Validate dependency references and detect/fix redundancies.

Ports ``scripts/resolve-deps.sh`` into pure Python.  Phase 2 validates
that every ``uses_rules``, ``uses_skills``, ``uses_plugins``, and
``delegates_to`` reference points to an existing entity.  Phase 3
detects transitive redundancies and produces auto-fixable ``Fix`` objects.
"""

from __future__ import annotations

from collections.abc import Callable
from pathlib import Path

from agent_army.frontmatter import write_field
from agent_army.graph import find_redundant, find_redundant_via_skills
from agent_army.models import Agent, Fix, Redundancy, Rule, Skill, ValidationError


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------


def validate_all_refs(
    rules: list[Rule],
    skills: list[Skill],
    agents: list[Agent],
    plugins: list[str],
) -> list[ValidationError]:
    """Check that every dependency reference points to an existing entity.

    Missing plugins receive ``severity="warning"``; all other missing
    references receive ``severity="error"``.
    """
    rule_names = {r.name for r in rules}
    skill_names = {s.name for s in skills}
    agent_names = {a.name for a in agents}
    plugin_names = set(plugins)

    errors: list[ValidationError] = []

    for rule in rules:
        errors.extend(
            _check_refs(
                str(rule.path), rule.uses_rules, "uses_rules", rule_names, "rules/"
            )
        )

    for skill in skills:
        errors.extend(
            _check_refs(
                str(skill.path), skill.uses_rules, "uses_rules", rule_names, "rules/"
            )
        )

    for agent in agents:
        label = str(agent.path)
        errors.extend(
            _check_refs(label, agent.uses_rules, "uses_rules", rule_names, "rules/")
        )
        errors.extend(
            _check_refs(label, agent.uses_skills, "uses_skills", skill_names, "skills/")
        )
        errors.extend(
            _check_refs(
                label,
                agent.uses_plugins,
                "uses_plugins",
                plugin_names,
                "config.json public_plugins",
                severity="warning",
            )
        )
        errors.extend(
            _check_refs(
                label, agent.delegates_to, "delegates_to", agent_names, "agents/"
            )
        )

    return errors


def compute_all_fixes(
    rules: list[Rule],
    skills: list[Skill],
    agents: list[Agent],
    root: Path,
) -> list[Fix]:
    """Detect redundancies and produce auto-fixable ``Fix`` entries.

    Runs three scans in order:
    a) ``uses_rules`` rule-to-rule redundancies across rules/, skills/, agents/.
    b) ``delegates_to`` agent-to-agent redundancies in agents/.
    c) ``uses_rules`` covered by agent skills -- merges into existing (a) fixes.
    """
    rule_lookup = _build_rule_lookup(rules)
    agent_lookup = _build_agent_lookup(agents)
    skill_lookup = _build_skill_lookup(skills)

    fixes: list[Fix] = []

    # (a) uses_rules redundancies for rules, skills, agents
    _scan_uses_rules(rules, rule_lookup, root, fixes)
    _scan_uses_rules_for_skills(skills, rule_lookup, root, fixes)
    _scan_uses_rules_for_agents(agents, rule_lookup, root, fixes)

    # (b) delegates_to redundancies for agents
    _scan_delegates_to(agents, agent_lookup, root, fixes)

    # (c) skill-transitive rule redundancies for agents (merges into existing)
    _scan_agent_skill_rules(agents, skill_lookup, rule_lookup, root, fixes)

    return fixes


def apply_fixes(fixes: list[Fix], root: Path) -> None:
    """Write each fix to disk by updating the frontmatter field."""
    for fix in fixes:
        file_path = root / fix.file_path
        write_field(file_path, fix.field, fix.after)


def format_report(
    errors: list[ValidationError],
    fixes: list[Fix],
) -> str:
    """Format a human-readable validation report matching the bash output."""
    real_errors = [e for e in errors if e.severity == "error"]
    warnings = [e for e in errors if e.severity == "warning"]

    if not real_errors and not warnings and not fixes:
        return "All dependency references are valid. No redundancies found."

    lines: list[str] = ["=== Dependency Validation Report ===", ""]
    _format_errors_section(real_errors, lines)
    _format_warnings_section(warnings, lines)
    _format_fixes_section(fixes, lines)
    _format_summary(real_errors, warnings, fixes, lines)

    if real_errors:
        lines.append("Fix errors above before auto-fixing redundancies.")

    return "\n".join(lines)


# ---------------------------------------------------------------------------
# Private: validation helpers
# ---------------------------------------------------------------------------


def _check_refs(
    file_label: str,
    refs: list[str],
    field: str,
    valid_names: set[str],
    location: str,
    *,
    severity: str = "error",
) -> list[ValidationError]:
    """Return a ValidationError for each ref not in *valid_names*."""
    return [
        ValidationError(
            file_label=file_label,
            field=field,
            ref=ref,
            severity=severity,
        )
        for ref in refs
        if ref not in valid_names
    ]


# ---------------------------------------------------------------------------
# Private: lookup builders
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
# Private: scan (a) — uses_rules redundancies
# ---------------------------------------------------------------------------


def _redundancies_to_fix(
    label: str,
    file_path: Path,
    field: str,
    original: list[str],
    redundancies: list[Redundancy],
) -> Fix | None:
    """Convert redundancies into a Fix, or None if nothing is redundant."""
    if not redundancies:
        return None

    redundant_names = {r.target for r in redundancies}
    cleaned = [entry for entry in original if entry not in redundant_names]
    reasons = [f'"{r.target}" covered by "{r.covered_by}"' for r in redundancies]

    return Fix(
        label=label,
        field=field,
        file_path=file_path,
        before=list(original),
        after=cleaned,
        reasons=reasons,
    )


def _scan_uses_rules(
    rules: list[Rule],
    rule_lookup: dict[str, list[str]],
    root: Path,
    fixes: list[Fix],
) -> None:
    """Scan rules/ for uses_rules redundancies."""
    _get_deps = _make_rule_deps_fn(rule_lookup)
    for rule in rules:
        if not rule.uses_rules:
            continue
        redundancies = find_redundant(rule.uses_rules, _get_deps)
        fix = _redundancies_to_fix(
            str(rule.path), rule.path, "uses_rules", rule.uses_rules, redundancies
        )
        if fix:
            fixes.append(fix)


def _scan_uses_rules_for_skills(
    skills: list[Skill],
    rule_lookup: dict[str, list[str]],
    root: Path,
    fixes: list[Fix],
) -> None:
    """Scan skills/ for uses_rules redundancies."""
    _get_deps = _make_rule_deps_fn(rule_lookup)
    for skill in skills:
        if not skill.uses_rules:
            continue
        redundancies = find_redundant(skill.uses_rules, _get_deps)
        fix = _redundancies_to_fix(
            str(skill.path), skill.path, "uses_rules", skill.uses_rules, redundancies
        )
        if fix:
            fixes.append(fix)


def _scan_uses_rules_for_agents(
    agents: list[Agent],
    rule_lookup: dict[str, list[str]],
    root: Path,
    fixes: list[Fix],
) -> None:
    """Scan agents/ for uses_rules redundancies (rule-to-rule only)."""
    _get_deps = _make_rule_deps_fn(rule_lookup)
    for agent in agents:
        if not agent.uses_rules:
            continue
        redundancies = find_redundant(agent.uses_rules, _get_deps)
        fix = _redundancies_to_fix(
            str(agent.path), agent.path, "uses_rules", agent.uses_rules, redundancies
        )
        if fix:
            fixes.append(fix)


# ---------------------------------------------------------------------------
# Private: scan (b) — delegates_to redundancies
# ---------------------------------------------------------------------------


def _scan_delegates_to(
    agents: list[Agent],
    agent_lookup: dict[str, list[str]],
    root: Path,
    fixes: list[Fix],
) -> None:
    """Scan agents/ for delegates_to redundancies."""
    _get_deps = _make_agent_deps_fn(agent_lookup)
    for agent in agents:
        if not agent.delegates_to:
            continue
        redundancies = find_redundant(agent.delegates_to, _get_deps)
        fix = _redundancies_to_fix(
            str(agent.path),
            agent.path,
            "delegates_to",
            agent.delegates_to,
            redundancies,
        )
        if fix:
            fixes.append(fix)


# ---------------------------------------------------------------------------
# Private: scan (c) — agent skill-rule redundancies with merge
# ---------------------------------------------------------------------------


def _scan_agent_skill_rules(
    agents: list[Agent],
    skill_lookup: dict[str, list[str]],
    rule_lookup: dict[str, list[str]],
    root: Path,
    fixes: list[Fix],
) -> None:
    """Find agent rules covered by skills; merge into existing fixes."""
    for agent in agents:
        if not agent.uses_rules or not agent.uses_skills:
            continue

        redundancies = find_redundant_via_skills(
            agent.uses_rules,
            agent.uses_skills,
            skill_lookup,
            rule_lookup,
        )
        if not redundancies:
            continue

        _merge_or_append_skill_fix(agent, redundancies, fixes)


def _merge_or_append_skill_fix(
    agent: Agent,
    redundancies: list[Redundancy],
    fixes: list[Fix],
) -> None:
    """Merge skill-based removals into an existing fix, or create a new one."""
    redundant_names = {r.target for r in redundancies}
    reasons = [f'"{r.target}" covered by {r.covered_by}' for r in redundancies]

    existing = _find_existing_fix(fixes, agent.path, "uses_rules")
    if existing is not None:
        # Merge: apply skill-transitive removals on the already-cleaned list
        existing.after = [e for e in existing.after if e not in redundant_names]
        existing.reasons.extend(reasons)
        return

    # No prior fix -- create a fresh one
    cleaned = [e for e in agent.uses_rules if e not in redundant_names]
    fixes.append(
        Fix(
            label=str(agent.path),
            field="uses_rules",
            file_path=agent.path,
            before=list(agent.uses_rules),
            after=cleaned,
            reasons=reasons,
        )
    )


def _find_existing_fix(fixes: list[Fix], file_path: Path, field: str) -> Fix | None:
    """Find a fix matching *file_path* and *field*, or None."""
    for fix in fixes:
        if fix.file_path == file_path and fix.field == field:
            return fix
    return None


# ---------------------------------------------------------------------------
# Private: dependency getter factories
# ---------------------------------------------------------------------------


def _make_rule_deps_fn(
    rule_lookup: dict[str, list[str]],
) -> Callable[[str], list[str]]:
    """Return a closure that resolves rule dependencies."""

    def _get_deps(name: str) -> list[str]:
        return rule_lookup.get(name, [])

    return _get_deps


def _make_agent_deps_fn(
    agent_lookup: dict[str, list[str]],
) -> Callable[[str], list[str]]:
    """Return a closure that resolves agent delegation dependencies."""

    def _get_deps(name: str) -> list[str]:
        return agent_lookup.get(name, [])

    return _get_deps


# ---------------------------------------------------------------------------
# Private: report formatting
# ---------------------------------------------------------------------------


_FIELD_LOCATION: dict[str, str] = {
    "uses_rules": "rules/",
    "uses_skills": "skills/",
    "uses_plugins": "config.json public_plugins",
    "delegates_to": "agents/",
}


def _format_errors_section(
    errors: list[ValidationError],
    lines: list[str],
) -> None:
    """Append the errors section to *lines*."""
    if not errors:
        return
    lines.append("--- Errors (must fix manually) ---")
    lines.append("")
    for err in errors:
        location = _FIELD_LOCATION.get(err.field, "unknown")
        lines.append(f"  [ERROR] {err.file_label}")
        lines.append(f'    {err.field}: "{err.ref}" not found in {location}')
    lines.append("")


def _format_warnings_section(
    warnings: list[ValidationError],
    lines: list[str],
) -> None:
    """Append the warnings section to *lines*."""
    if not warnings:
        return
    lines.append("--- Warnings ---")
    lines.append("")
    for warn in warnings:
        location = _FIELD_LOCATION.get(warn.field, "unknown")
        lines.append(f"  [WARN] {warn.file_label}")
        lines.append(f'    {warn.field}: "{warn.ref}" not found in {location}')
    lines.append("")


def _format_fixes_section(
    fixes: list[Fix],
    lines: list[str],
) -> None:
    """Append the redundancy fixes section to *lines*."""
    if not fixes:
        return
    lines.append("--- Redundancies (auto-fixable) ---")
    lines.append("")
    for fix in fixes:
        lines.append(f"  [FIX] {fix.label}")
        for reason in fix.reasons:
            lines.append(f"    {fix.field}: {reason}")
        lines.append(f"    Before: [{', '.join(fix.before)}]")
        lines.append(f"    After:  [{', '.join(fix.after)}]")
        lines.append("")


def _format_summary(
    errors: list[ValidationError],
    warnings: list[ValidationError],
    fixes: list[Fix],
    lines: list[str],
) -> None:
    """Append the summary line to *lines*."""
    lines.append(
        f"Summary: {len(errors)} error(s), {len(warnings)} warning(s), "
        f"{len(fixes)} fixable redundanc(ies) across files."
    )
    lines.append("")
