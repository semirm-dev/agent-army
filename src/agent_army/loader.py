"""Load rules, skills, and agents from the filesystem.

Scans markdown files with YAML frontmatter and builds typed domain
objects.  Mirrors the bash loading functions in ``scripts/lib-deps.sh``
and ``scripts/generate-manifest.sh``.
"""

from __future__ import annotations

import json
from pathlib import Path

from agent_army.frontmatter import extract_h1, parse_frontmatter
from agent_army.models import Agent, Rule, Skill


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------


def find_md_files(directory: Path) -> list[Path]:
    """Find all ``.md`` files recursively, sorted by path string."""
    return sorted(directory.rglob("*.md"), key=lambda p: str(p))


def load_rules(root: Path) -> list[Rule]:
    """Load all rules from ``root/rules/`` directory.

    Name is derived from the relative path minus the ``.md`` extension
    (e.g. ``rules/go/patterns.md`` becomes ``"go/patterns"``).
    Description comes from the first H1 heading after frontmatter.
    Scope defaults to ``"universal"`` when not set.
    """
    rules_dir = root / "rules"
    if not rules_dir.is_dir():
        return []

    rules: list[Rule] = []
    for md_path in find_md_files(rules_dir):
        content = md_path.read_text(encoding="utf-8")
        fm = parse_frontmatter(content)
        rel = md_path.relative_to(rules_dir)
        name = str(rel.with_suffix(""))

        rules.append(
            Rule(
                name=name,
                description=extract_h1(content),
                scope=_str_or_default(fm.get("scope"), "universal"),
                languages=_ensure_list(fm.get("languages")),
                uses_rules=_ensure_list(fm.get("uses_rules")),
                path=Path("rules") / rel,
            )
        )
    return rules


def load_skills(root: Path) -> list[Skill]:
    """Load all skills from ``root/skills/`` directory.

    Name comes from the frontmatter ``name`` field.  Falls back to the
    relative path minus ``.md`` when no name field is present.
    Description comes from the first H1 heading.
    Scope defaults to ``"universal"`` when not set.
    """
    skills_dir = root / "skills"
    if not skills_dir.is_dir():
        return []

    skills: list[Skill] = []
    for md_path in find_md_files(skills_dir):
        content = md_path.read_text(encoding="utf-8")
        fm = parse_frontmatter(content)
        rel = md_path.relative_to(skills_dir)
        name = _str_or_default(fm.get("name"), str(rel.with_suffix("")))

        skills.append(
            Skill(
                name=name,
                description=extract_h1(content),
                scope=_str_or_default(fm.get("scope"), "universal"),
                languages=_ensure_list(fm.get("languages")),
                uses_rules=_ensure_list(fm.get("uses_rules")),
                path=Path("skills") / rel,
            )
        )
    return skills


def load_agents(root: Path) -> list[Agent]:
    """Load all agents from ``root/agents/`` directory.

    Name comes from the frontmatter ``name`` field (fallback: relative path).
    Description comes from the frontmatter ``description`` field (NOT H1).
    Role comes from frontmatter.
    Access defaults to ``"read-write"``, scope defaults to ``"universal"``.
    """
    agents_dir = root / "agents"
    if not agents_dir.is_dir():
        return []

    agents: list[Agent] = []
    for md_path in find_md_files(agents_dir):
        content = md_path.read_text(encoding="utf-8")
        fm = parse_frontmatter(content)
        rel = md_path.relative_to(agents_dir)
        name = _str_or_default(fm.get("name"), str(rel.with_suffix("")))

        agents.append(
            Agent(
                name=name,
                description=_str_or_default(fm.get("description"), ""),
                role=_str_or_default(fm.get("role"), ""),
                scope=_str_or_default(fm.get("scope"), "universal"),
                access=_str_or_default(fm.get("access"), "read-write"),
                languages=_ensure_list(fm.get("languages")),
                uses_skills=_ensure_list(fm.get("uses_skills")),
                uses_rules=_ensure_list(fm.get("uses_rules")),
                uses_plugins=_ensure_list(fm.get("uses_plugins")),
                delegates_to=_ensure_list(fm.get("delegates_to")),
                path=Path("agents") / rel,
            )
        )
    return agents


def load_plugins(root: Path) -> list[str]:
    """Load plugin names from ``root/config.json`` -> ``public_plugins[].name``."""
    config_path = root / "config.json"
    if not config_path.is_file():
        return []

    try:
        data = json.loads(config_path.read_text(encoding="utf-8"))
    except (json.JSONDecodeError, OSError):
        return []

    plugins = data.get("public_plugins", [])
    return [p["name"] for p in plugins if isinstance(p, dict) and "name" in p]


# ---------------------------------------------------------------------------
# Private helpers
# ---------------------------------------------------------------------------


def _str_or_default(value: str | list[str] | None, default: str) -> str:
    """Return *value* as a string, or *default* if missing/empty."""
    if value is None:
        return default
    if isinstance(value, list):
        return default
    return value if value else default


def _ensure_list(value: str | list[str] | None) -> list[str]:
    """Coerce a frontmatter value to a list of strings."""
    if value is None:
        return []
    if isinstance(value, list):
        return value
    # Scalar -- should not normally happen for list fields, but be safe.
    stripped = value.strip()
    if not stripped:
        return []
    return [stripped]
