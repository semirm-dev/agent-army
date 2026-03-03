"""Generate model-specific rules, skills, and agents from repo templates.

Reads the canonical YAML+markdown templates from rules/, skills/, agents/
and transforms them into Claude Code or Cursor format, writing output
to a user-chosen destination directory.
"""

from __future__ import annotations

import os
import re
from pathlib import Path

from agent_army.frontmatter import parse_frontmatter
from agent_army.graph import resolve_transitive
from agent_army.loader import load_agents, load_rules, load_skills
from agent_army.models import Agent, Rule, Skill

# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------

_TARGETS = ["Claude Code", "Cursor"]

_DESTINATIONS = ["Local project", "Global home", "Custom directory"]

# Category -> starting number for Cursor rule numbering
_CURSOR_CATEGORIES: dict[str, int] = {
    "language": 100,
    "git": 300,
    "api-db": 400,
    "infrastructure": 500,
}

# Rule name patterns -> category
_CATEGORY_PATTERNS: list[tuple[str, str]] = [
    ("go/", "language"),
    ("python/", "language"),
    ("typescript/", "language"),
    ("react/", "language"),
    ("git", "git"),
    ("api-design", "api-db"),
    ("database", "api-db"),
]

# Everything else falls into "infrastructure"

# Language-specific globs for Cursor rules
_LANGUAGE_GLOBS: dict[str, str] = {
    "go": '"**/*.go"',
    "typescript": '"**/*.ts,**/*.tsx,**/*.js,**/*.jsx"',
    "python": '"**/*.py"',
    "react": '"**/*.tsx,**/*.jsx"',
}

# Cursor short name mapping for language rules
_CURSOR_LANG_NAMES: dict[str, str] = {
    "go/patterns": "golang",
    "go/testing": "golang-testing",
    "typescript/patterns": "typescript",
    "typescript/testing": "typescript-testing",
    "python/patterns": "python",
    "python/testing": "python-testing",
    "react/patterns": "react",
    "react/testing": "react-testing",
}

# Claude tools by access level
_CLAUDE_TOOLS_RW = "Read, Write, Edit, Bash, Glob, Grep"
_CLAUDE_TOOLS_RO = "Read, Glob, Grep, Bash"


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------


def main_bootstrap(root: Path) -> None:
    """Interactive bootstrap flow: target -> destination -> entities -> generate."""
    print("=== Bootstrap ===\n")

    try:
        # Step 1: Target
        target = _select_target()

        # Step 2: Destination
        dest = _select_destination(target)

        # Step 3: Load all entities
        rules = load_rules(root)
        skills = load_skills(root)
        agents = load_agents(root)

        all_rule_names = [r.name for r in rules]
        all_skill_names = [s.name for s in skills]
        all_skill_set = set(all_skill_names)

        # Build lookup maps for dependency resolution
        rule_lookup: dict[str, list[str]] = {r.name: r.uses_rules for r in rules}

        # Step 4: Select agents (primary entry point)
        selected_agent_names = _select_entities("agents", [a.name for a in agents])
        agent_objs = [a for a in agents if a.name in set(selected_agent_names)]

        # Step 5: Auto-compute required skills from agents' uses_skills
        auto_skill_names: list[str] = []
        seen_skills: set[str] = set()
        for agent in agent_objs:
            for skill_name in agent.uses_skills:
                if skill_name not in seen_skills and skill_name in all_skill_set:
                    auto_skill_names.append(skill_name)
                    seen_skills.add(skill_name)

        # Step 6: Offer additional skill selection
        final_skill_names = _select_additional_entities(
            "skills", auto_skill_names, all_skill_names,
        )

        # Step 7: Auto-compute required rules transitively
        rule_seeds: list[str] = []
        seen_rules: set[str] = set()

        # From selected skills' uses_rules
        skill_objs = [s for s in skills if s.name in set(final_skill_names)]
        for skill in skill_objs:
            for rule_name in skill.uses_rules:
                if rule_name not in seen_rules:
                    rule_seeds.append(rule_name)
                    seen_rules.add(rule_name)

        # From agents' direct uses_rules
        for agent in agent_objs:
            for rule_name in agent.uses_rules:
                if rule_name not in seen_rules:
                    rule_seeds.append(rule_name)
                    seen_rules.add(rule_name)

        # Resolve transitively and filter to existing rules
        existing_rule_set = set(all_rule_names)
        auto_rule_names = [
            r for r in resolve_transitive(
                rule_seeds, lambda name: rule_lookup.get(name, []),
            )
            if r in existing_rule_set
        ]

        # Step 8: Offer additional rule selection
        final_rule_names = _select_additional_entities(
            "rules", auto_rule_names, all_rule_names,
        )

        rule_objs = [r for r in rules if r.name in set(final_rule_names)]

        total = len(rule_objs) + len(skill_objs) + len(agent_objs)
        if total == 0:
            print("\nNo entities selected. Nothing to generate.")
            return

        # Step 4: Preview
        print(f"\n--- Preview ---")
        print(f"  Target:      {target}")
        print(f"  Destination: {dest}")
        print(f"  Rules:       {len(rule_objs)} files")
        print(f"  Skills:      {len(skill_objs)} files")
        print(f"  Agents:      {len(agent_objs)} files")
        print(f"  Total:       {total} files")
        print()

        confirm = input("Proceed? [y/N] ")
        if confirm.strip().lower() != "y":
            print("Aborted. No files written.")
            return

        # Step 5: Generate
        is_claude = target == "Claude Code"
        written = _generate_all(
            root, dest, rule_objs, skill_objs, agent_objs, is_claude,
        )

        print(f"\nDone. {written} files written to {dest}")

    except (EOFError, KeyboardInterrupt):
        print("\nAborted.")


# ---------------------------------------------------------------------------
# Interactive prompts
# ---------------------------------------------------------------------------


def _select_target() -> str:
    """Step 1: Choose target AI model/tool."""
    print("Step 1 — Target AI model/tool:")
    for i, t in enumerate(_TARGETS, 1):
        print(f"  {i}) {t}")
    print()
    while True:
        raw = input("Select target: ")
        try:
            idx = int(raw.strip())
            if 1 <= idx <= len(_TARGETS):
                return _TARGETS[idx - 1]
        except ValueError:
            pass
        print(f"Invalid choice. Enter 1-{len(_TARGETS)}.")


def _select_destination(target: str) -> Path:
    """Step 2: Choose output destination."""
    suffix = ".claude" if target == "Claude Code" else ".cursor"

    local = Path.cwd() / suffix
    global_home = Path.home() / suffix

    print(f"\nStep 2 — Output destination:")
    print(f"  1) Local project ({local})  (*)")
    print(f"  2) Global ({global_home})")
    print(f"  3) Custom directory")
    print()
    while True:
        raw = input("Select destination [1]: ").strip()
        if raw == "" or raw == "1":
            return local
        if raw == "2":
            return global_home
        if raw == "3":
            custom = input("Enter path (absolute or relative): ").strip()
            if not custom:
                print("Path cannot be empty.")
                continue
            p = Path(custom)
            if not p.is_absolute():
                p = Path.cwd() / p
            return p
        print("Invalid choice. Enter 1, 2, or 3.")


def _select_entities(entity_type: str, names: list[str]) -> list[str]:
    """Step 3: Select entities of a given type.

    Enter = all, 'none' = skip, comma-separated numbers = specific.
    """
    if not names:
        return []

    print(f"\nAvailable {entity_type} ({len(names)}):")
    for i, name in enumerate(names, 1):
        print(f"  {i}) {name}")
    print()

    while True:
        raw = input(
            f"Select {entity_type} (comma-separated, Enter for all, 'none' to skip): "
        ).strip()

        if raw == "":
            return list(names)

        if raw.lower() == "none":
            return []

        parts = [p.strip() for p in raw.split(",")]
        selected: list[str] = []
        valid = True
        for part in parts:
            try:
                idx = int(part)
                if 1 <= idx <= len(names):
                    selected.append(names[idx - 1])
                else:
                    print(f"Invalid number: {part}")
                    valid = False
                    break
            except ValueError:
                print(f"Invalid number: {part}")
                valid = False
                break

        if valid and selected:
            return selected


def _select_additional_entities(
    entity_type: str,
    auto_names: list[str],
    all_names: list[str],
) -> list[str]:
    """Show auto-resolved items, offer optional selection from remaining pool.

    Args:
        entity_type: Display label (e.g. "skills", "rules").
        auto_names: Names auto-resolved from dependency graph.
        all_names: All available names of this entity type.

    Returns:
        Final list of selected names (auto + any user-added extras).
    """
    auto_set = set(auto_names)
    remaining = [n for n in all_names if n not in auto_set]

    if not auto_names and not remaining:
        return []

    if auto_names and not remaining:
        print(f"\n  Auto-included {entity_type}: {', '.join(auto_names)}")
        print(f"  All available {entity_type} are already included.")
        return list(auto_names)

    if auto_names:
        print(f"\n  Auto-included {entity_type}: {', '.join(auto_names)}")
    else:
        print(f"\n  No auto-included {entity_type}.")

    print(f"\n  Additional {entity_type} available:")
    for i, name in enumerate(remaining, 1):
        print(f"    {i}) {name}")
    print()

    if auto_names:
        prompt = f"Add extra {entity_type}? (comma-separated, Enter for none, 'all' for all): "
    else:
        prompt = f"Select {entity_type}? (comma-separated, Enter for none, 'all' for all): "

    while True:
        raw = input(prompt).strip()

        if raw == "":
            return list(auto_names)

        if raw.lower() == "all":
            return list(auto_names) + remaining

        parts = [p.strip() for p in raw.split(",")]
        selected = list(auto_names)
        valid = True
        for part in parts:
            try:
                idx = int(part)
                if 1 <= idx <= len(remaining):
                    selected.append(remaining[idx - 1])
                else:
                    print(f"Invalid number: {part}")
                    valid = False
                    break
            except ValueError:
                print(f"Invalid number: {part}")
                valid = False
                break

        if valid:
            return selected


# ---------------------------------------------------------------------------
# Generation orchestrator
# ---------------------------------------------------------------------------


def _generate_all(
    root: Path,
    dest: Path,
    rules: list[Rule],
    skills: list,
    agents: list[Agent],
    is_claude: bool,
) -> int:
    """Generate all selected entities. Returns count of files written."""
    written = 0

    # Rules
    if rules:
        if is_claude:
            for rule in rules:
                content = _rule_to_claude(root, rule)
                rel = Path("rules") / f"{_flatten_name(rule.name)}.md"
                _write_output(dest, rel, content)
                written += 1
        else:
            assignments = _assign_cursor_numbers(rules)
            for rule, (number, short_name) in zip(rules, assignments):
                content = _rule_to_cursor(root, rule)
                rel = Path("rules") / f"{number}-{short_name}.mdc"
                rel = _resolve_collision(dest, rel)
                _write_output(dest, rel, content)
                written += 1

    # Skills (identical for both targets)
    for skill in skills:
        flat = _flatten_name(skill.name)
        content = _read_file_content(root, skill.path)
        rel = Path("skills") / flat / "SKILL.md"
        _write_output(dest, rel, content)
        written += 1

    # Agents
    for agent in agents:
        flat = _flatten_name(agent.name)
        if is_claude:
            content = _agent_to_claude(root, agent)
        else:
            content = _agent_to_cursor(root, agent)
        rel = Path("agents") / f"{flat}.md"
        _write_output(dest, rel, content)
        written += 1

    return written


# ---------------------------------------------------------------------------
# Rule transformers
# ---------------------------------------------------------------------------


def _rule_to_claude(root: Path, rule: Rule) -> str:
    """Transform a rule for Claude Code output.

    Strips repo frontmatter, keeps body.
    """
    body = _extract_body(root / rule.path)
    return body


def _rule_to_cursor(root: Path, rule: Rule) -> str:
    """Transform a rule for Cursor output.

    Builds new YAML frontmatter with description, globs/alwaysApply.
    """
    body = _extract_body(root / rule.path)
    description = rule.description

    # Build cursor frontmatter
    lines = ["---"]
    lines.append(f"description: {description}")

    # Language-specific rules get globs; universal get alwaysApply
    lang = _detect_language(rule)
    if lang and lang in _LANGUAGE_GLOBS:
        lines.append(f"globs: {_LANGUAGE_GLOBS[lang]}")
    else:
        lines.append("alwaysApply: true")

    lines.append("---")
    fm = "\n".join(lines)

    return fm + "\n\n" + body


# ---------------------------------------------------------------------------
# Agent transformers
# ---------------------------------------------------------------------------


def _agent_to_claude(root: Path, agent: Agent) -> str:
    """Transform an agent for Claude Code output.

    Builds Claude-style frontmatter with tools derived from access level.
    """
    body = _extract_body(root / agent.path)
    flat = _flatten_name(agent.name)

    tools = _CLAUDE_TOOLS_RO if agent.access == "read-only" else _CLAUDE_TOOLS_RW

    lines = ["---"]
    lines.append(f"name: {flat}")
    desc = agent.description
    if ":" in desc:
        lines.append(f'description: "{desc}"')
    else:
        lines.append(f"description: {desc}")
    lines.append(f"tools: {tools}")
    lines.append("model: inherit")
    lines.append("---")
    fm = "\n".join(lines)

    return fm + "\n\n" + body


def _agent_to_cursor(root: Path, agent: Agent) -> str:
    """Transform an agent for Cursor output.

    Builds Cursor-style frontmatter (name + description only).
    Applies text substitutions for tool names and paths.
    """
    body = _extract_body(root / agent.path)
    flat = _flatten_name(agent.name)

    lines = ["---"]
    lines.append(f"name: {flat}")
    desc = agent.description
    if ":" in desc:
        lines.append(f'description: "{desc}"')
    else:
        lines.append(f"description: {desc}")
    lines.append("---")
    fm = "\n".join(lines)

    # Cursor tool substitutions
    body = re.sub(r"`Edit`", "`StrReplace`", body)
    body = re.sub(r"`Bash`", "`Shell`", body)
    body = body.replace("~/.claude/", "~/.cursor/")

    return fm + "\n\n" + body


# ---------------------------------------------------------------------------
# Cursor numbering
# ---------------------------------------------------------------------------


def _assign_cursor_numbers(rules: list[Rule]) -> list[tuple[int, str]]:
    """Assign sequential numbers to rules by category.

    Returns a list of (number, short_name) tuples, one per rule, in the
    same order as the input list.
    """
    # Categorize each rule
    categorized: dict[str, list[tuple[int, Rule]]] = {
        cat: [] for cat in _CURSOR_CATEGORIES
    }
    for idx, rule in enumerate(rules):
        cat = _categorize_rule(rule)
        categorized[cat].append((idx, rule))

    # Assign numbers per category
    result: list[tuple[int, tuple[int, str]]] = []
    for cat, start in _CURSOR_CATEGORIES.items():
        entries = categorized[cat]
        for offset, (orig_idx, rule) in enumerate(entries):
            number = start + offset
            short = _cursor_short_name(rule)
            result.append((orig_idx, (number, short)))

    # Sort back to original order
    result.sort(key=lambda x: x[0])
    return [r[1] for r in result]


def _categorize_rule(rule: Rule) -> str:
    """Determine the Cursor category for a rule."""
    for pattern, cat in _CATEGORY_PATTERNS:
        if rule.name.startswith(pattern) or rule.name == pattern.rstrip("/"):
            return cat
    return "infrastructure"


def _cursor_short_name(rule: Rule) -> str:
    """Derive a short Cursor filename from a rule name."""
    # Check explicit mapping first
    if rule.name in _CURSOR_LANG_NAMES:
        return _CURSOR_LANG_NAMES[rule.name]
    # Fallback: flatten and use as-is
    return _flatten_name(rule.name)


# ---------------------------------------------------------------------------
# File I/O helpers
# ---------------------------------------------------------------------------


def _extract_body(file_path: Path) -> str:
    """Read a file and strip YAML frontmatter, returning just the body."""
    content = file_path.read_text(encoding="utf-8")
    lines = content.splitlines(keepends=True)

    dash_count = 0
    body_start = 0
    for i, line in enumerate(lines):
        if line.rstrip() == "---":
            dash_count += 1
            if dash_count == 2:
                body_start = i + 1
                break

    if dash_count < 2:
        return content

    body = "".join(lines[body_start:]).lstrip("\n")
    return body


def _read_file_content(root: Path, rel_path: Path) -> str:
    """Read raw file content from the repo."""
    return (root / rel_path).read_text(encoding="utf-8")


def _write_output(dest: Path, rel_path: Path, content: str) -> None:
    """Write content to dest/rel_path atomically."""
    target = dest / rel_path
    target.parent.mkdir(parents=True, exist_ok=True)
    tmp = target.with_suffix(target.suffix + ".tmp")
    tmp.write_text(content, encoding="utf-8")
    os.replace(tmp, target)


def _resolve_collision(dest: Path, rel_path: Path) -> Path:
    """If dest/rel_path exists, try _2, _3, etc. suffixes."""
    target = dest / rel_path
    if not target.exists():
        return rel_path

    stem = rel_path.stem
    suffix = rel_path.suffix
    parent = rel_path.parent

    for i in range(2, 100):
        candidate = parent / f"{stem}_{i}{suffix}"
        if not (dest / candidate).exists():
            return candidate

    return rel_path  # give up after 99 attempts


def _flatten_name(name: str) -> str:
    """Flatten a path-style name: ``go/patterns`` -> ``go-patterns``."""
    return name.replace("/", "-")


def _detect_language(rule: Rule) -> str | None:
    """Detect the primary language from a rule's name or languages list."""
    # Check name prefix
    for lang in ("go", "typescript", "python", "react"):
        if rule.name.startswith(f"{lang}/"):
            return lang

    # Check languages list
    if rule.languages:
        return rule.languages[0]

    return None
