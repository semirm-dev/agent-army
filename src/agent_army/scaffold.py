"""Scaffold new rules, skills, and agents from the CLI.

Provides interactive prompts to gather metadata, generates a markdown
file with YAML frontmatter and template sections, and regenerates the
manifest.
"""

from __future__ import annotations

import os
from pathlib import Path

from agent_army.editor import (
    prompt_with_default,
    select_multi_optional,
    select_one_with_default,
)
from agent_army.loader import load_agents, load_plugins, load_rules, load_skills

# Scope choices
_SCOPES = ["universal", "language-specific"]

# Common languages for language-specific entities
_COMMON_LANGUAGES = ["go", "python", "typescript", "react", "rust", "java"]

# Agent role choices
_AGENT_ROLES = ["coder", "reviewer", "tester", "analyzer", "writer", "builder"]

# Roles that default to read-only access
_READ_ONLY_ROLES = {"reviewer", "analyzer"}

# Access choices
_ACCESS_CHOICES = ["read-write", "read-only"]


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------


def scaffold_flow(root: Path, entity_type: str) -> None:
    """Interactive scaffold flow for creating a new entity.

    Prompts for metadata, shows a preview, confirms, writes the file,
    and regenerates the manifest.
    """
    dispatch = {
        "rule": _scaffold_rule,
        "skill": _scaffold_skill,
        "agent": _scaffold_agent,
    }
    dispatch[entity_type](root)


# ---------------------------------------------------------------------------
# Per-entity scaffold flows
# ---------------------------------------------------------------------------


def _scaffold_rule(root: Path) -> None:
    """Scaffold a new rule."""
    print("=== New Rule ===\n")

    # 1. Name
    name = _prompt_name("rule")

    # 2. Check duplicate
    file_path = _build_file_path(root, "rules", name)
    if _check_duplicate(file_path):
        return

    # 3. Description
    description = prompt_with_default(
        "Description",
        _default_description("rule", name),
    )

    # 4. Scope
    scope = select_one_with_default("Scope:", _SCOPES, "universal")

    # 5. Languages
    languages = _prompt_languages(scope)

    # 6. Dependencies
    available_rules = [r.name for r in load_rules(root)]
    uses_rules = _prompt_dependencies("uses_rules", available_rules)

    # 7. Build content
    fields = _build_rule_fields(name, description, scope, languages, uses_rules)
    content = _generate_rule_content(fields, name)

    # 8. Preview + confirm + write
    _preview_confirm_write(file_path, content, root)


def _scaffold_skill(root: Path) -> None:
    """Scaffold a new skill."""
    print("=== New Skill ===\n")

    # 1. Name
    name = _prompt_name("skill")

    # 2. Check duplicate
    file_path = _build_file_path(root, "skills", name)
    if _check_duplicate(file_path):
        return

    # 3. Description
    description = prompt_with_default(
        "Description",
        _default_description("skill", name),
    )

    # 4. Scope
    scope = select_one_with_default("Scope:", _SCOPES, "universal")

    # 5. Languages
    languages = _prompt_languages(scope)

    # 6. Dependencies
    available_rules = [r.name for r in load_rules(root)]
    uses_rules = _prompt_dependencies("uses_rules", available_rules)

    # 7. Build content
    fields = _build_skill_fields(name, description, scope, languages, uses_rules)
    content = _generate_skill_content(fields, name)

    # 8. Preview + confirm + write
    _preview_confirm_write(file_path, content, root)


def _scaffold_agent(root: Path) -> None:
    """Scaffold a new agent."""
    print("=== New Agent ===\n")

    # 1. Name
    name = _prompt_name("agent")

    # 2. Check duplicate
    file_path = _build_file_path(root, "agents", name)
    if _check_duplicate(file_path):
        return

    # 3. Description
    description = prompt_with_default(
        "Description",
        _default_description("agent", name),
    )

    # 4. Role
    role = select_one_with_default("Role:", _AGENT_ROLES, "coder")

    # 5. Scope
    scope = select_one_with_default("Scope:", _SCOPES, "universal")

    # 6. Languages
    languages = _prompt_languages(scope)

    # 7. Access (smart default based on role)
    default_access = "read-only" if role in _READ_ONLY_ROLES else "read-write"
    access = select_one_with_default("Access:", _ACCESS_CHOICES, default_access)

    # 8. Dependencies
    available_skills = [s.name for s in load_skills(root)]
    uses_skills = _prompt_dependencies("uses_skills", available_skills)

    available_rules = [r.name for r in load_rules(root)]
    uses_rules = _prompt_dependencies("uses_rules", available_rules)

    available_plugins = load_plugins(root)
    uses_plugins = _prompt_dependencies("uses_plugins", available_plugins)

    available_agents = [a.name for a in load_agents(root)]
    delegates_to = _prompt_dependencies("delegates_to", available_agents)

    # 9. Build content
    fields = _build_agent_fields(
        name, description, role, scope, languages, access,
        uses_skills, uses_rules, uses_plugins, delegates_to,
    )
    content = _generate_agent_content(fields, name, access)

    # 10. Preview + confirm + write
    _preview_confirm_write(file_path, content, root)


# ---------------------------------------------------------------------------
# Shared prompts
# ---------------------------------------------------------------------------


def _prompt_name(entity_type: str) -> str:
    """Prompt for entity name (supports nested paths like go/testing)."""
    return input(f"{entity_type.capitalize()} name (e.g. 'security' or 'go/testing'): ").strip()


def _prompt_languages(scope: str) -> list[str]:
    """Prompt for languages when scope is language-specific."""
    if scope != "language-specific":
        return []

    print("\nCommon languages:")
    for i, lang in enumerate(_COMMON_LANGUAGES, start=1):
        print(f"  {i}) {lang}")
    print()

    raw = input("Languages (comma-separated numbers, or type custom names): ").strip()
    if not raw:
        return []

    result: list[str] = []
    for part in raw.split(","):
        part = part.strip()
        try:
            idx = int(part)
            if 1 <= idx <= len(_COMMON_LANGUAGES):
                result.append(_COMMON_LANGUAGES[idx - 1])
            else:
                result.append(part)
        except ValueError:
            result.append(part)
    return result


def _prompt_dependencies(field: str, available: list[str]) -> list[str]:
    """Prompt for optional dependency selection from available entities."""
    if not available:
        return []
    print(f"\nAvailable {field}:")
    return select_multi_optional(f"Select {field}", available)


# ---------------------------------------------------------------------------
# File path helpers
# ---------------------------------------------------------------------------


def _build_file_path(root: Path, entity_dir: str, name: str) -> Path:
    """Build the target file path from root, entity directory, and name."""
    return root / entity_dir / f"{name}.md"


def _check_duplicate(file_path: Path) -> bool:
    """Check if the file already exists. Prints a message and returns True if so."""
    if file_path.exists():
        print(f"File already exists: {file_path}")
        print("Aborted. Use 'agent-army edit' to modify existing entities.")
        return True
    return False


# ---------------------------------------------------------------------------
# Default description
# ---------------------------------------------------------------------------


def _default_description(entity_type: str, name: str) -> str:
    """Generate a sensible default description from the entity name.

    Converts path-style names like ``go/testing`` into
    ``Go Testing patterns and conventions``.
    """
    parts = name.replace("/", " ").replace("-", " ").split()
    title = " ".join(p.capitalize() for p in parts)

    suffixes = {
        "rule": "patterns and conventions",
        "skill": "workflow and decision tree",
        "agent": "specialist agent",
    }
    return f"{title} {suffixes.get(entity_type, '')}".strip()


# ---------------------------------------------------------------------------
# Frontmatter generation
# ---------------------------------------------------------------------------


def _generate_frontmatter(fields: dict[str, str | list[str]]) -> str:
    """Build YAML frontmatter string from a fields dict.

    Lists are formatted inline: ``field: [a, b]``.
    Descriptions containing colons are quoted.
    """
    lines = ["---"]
    for key, value in fields.items():
        if isinstance(value, list):
            if not value:
                lines.append(f"{key}: []")
            else:
                joined = ", ".join(value)
                lines.append(f"{key}: [{joined}]")
        else:
            if ":" in value:
                lines.append(f'{key}: "{value}"')
            else:
                lines.append(f"{key}: {value}")
    lines.append("---")
    return "\n".join(lines)


def _build_rule_fields(
    name: str,
    description: str,
    scope: str,
    languages: list[str],
    uses_rules: list[str],
) -> dict[str, str | list[str]]:
    """Build frontmatter fields dict for a rule."""
    fields: dict[str, str | list[str]] = {
        "name": name,
        "description": description,
        "scope": scope,
    }
    if languages:
        fields["languages"] = languages
    else:
        fields["languages"] = []
    fields["uses_rules"] = uses_rules
    return fields


def _build_skill_fields(
    name: str,
    description: str,
    scope: str,
    languages: list[str],
    uses_rules: list[str],
) -> dict[str, str | list[str]]:
    """Build frontmatter fields dict for a skill."""
    fields: dict[str, str | list[str]] = {
        "name": name,
        "description": description,
        "scope": scope,
    }
    if languages:
        fields["languages"] = languages
    else:
        fields["languages"] = []
    fields["uses_rules"] = uses_rules
    return fields


def _build_agent_fields(
    name: str,
    description: str,
    role: str,
    scope: str,
    languages: list[str],
    access: str,
    uses_skills: list[str],
    uses_rules: list[str],
    uses_plugins: list[str],
    delegates_to: list[str],
) -> dict[str, str | list[str]]:
    """Build frontmatter fields dict for an agent."""
    fields: dict[str, str | list[str]] = {
        "name": name,
        "description": description,
        "role": role,
        "scope": scope,
    }
    if languages:
        fields["languages"] = languages
    else:
        fields["languages"] = []
    fields["access"] = access
    fields["uses_skills"] = uses_skills
    fields["uses_rules"] = uses_rules
    fields["uses_plugins"] = uses_plugins
    fields["delegates_to"] = delegates_to
    return fields


# ---------------------------------------------------------------------------
# Content generation (frontmatter + template body)
# ---------------------------------------------------------------------------


def _generate_rule_content(
    fields: dict[str, str | list[str]],
    name: str,
) -> str:
    """Generate full markdown content for a rule."""
    title = _name_to_title(name)
    frontmatter = _generate_frontmatter(fields)

    body = f"""
# {title} Patterns

## Overview

<!-- Describe the purpose and scope of these patterns. -->

## Patterns

<!-- List the key patterns, conventions, and best practices. -->

## Anti-Patterns

<!-- List common mistakes and what to do instead. -->
"""
    return frontmatter + "\n" + body


def _generate_skill_content(
    fields: dict[str, str | list[str]],
    name: str,
) -> str:
    """Generate full markdown content for a skill."""
    title = _name_to_title(name)
    frontmatter = _generate_frontmatter(fields)

    body = f"""
# {title}

## When to Use

<!-- Describe when this skill should be invoked. -->

## Workflow

<!-- Step-by-step workflow for this skill. -->

## Decision Tree

<!-- Decision tree or flowchart for key choices. -->

## Checklist

<!-- Pre-completion checklist items. -->
"""
    return frontmatter + "\n" + body


def _generate_agent_content(
    fields: dict[str, str | list[str]],
    name: str,
    access: str,
) -> str:
    """Generate full markdown content for an agent."""
    title = _name_to_title(name)
    frontmatter = _generate_frontmatter(fields)

    if access == "read-only":
        capabilities = (
            "- Read source files, configuration, and documentation\n"
            "- Search for patterns, imports, and dependencies\n"
            "- Run read-only analysis commands\n"
            "- Cannot modify any files"
        )
    else:
        capabilities = (
            "- Read and write source files\n"
            "- Run build, test, and lint commands\n"
            "- Create new files and directories\n"
            "- Modify existing code following project patterns"
        )

    body = f"""
# {title} Agent

## Role

<!-- Describe the agent's role and expertise. -->

## Activation

<!-- When does the orchestrator activate this agent? -->

## Capabilities

{capabilities}

## Standards

<!-- Key standards this agent enforces or follows. -->

## Workflow

<!-- Step-by-step workflow the agent follows. -->

## Output Format

<!-- Describe the expected output format. -->

## Constraints

<!-- Hard constraints the agent must never violate. -->
"""
    return frontmatter + "\n" + body


# ---------------------------------------------------------------------------
# Preview, confirm, write
# ---------------------------------------------------------------------------


def _preview_confirm_write(file_path: Path, content: str, root: Path) -> None:
    """Show preview, ask for confirmation, write file, regenerate manifest."""
    print("\n--- Preview ---")
    print(content)
    print("--- End Preview ---\n")
    print(f"File: {file_path.relative_to(root)}")

    try:
        confirm = input("Create this file? [y/N] ")
    except (EOFError, KeyboardInterrupt):
        print("\nAborted.")
        return

    if confirm.strip().lower() != "y":
        print("Aborted. No files created.")
        return

    # Write atomically
    file_path.parent.mkdir(parents=True, exist_ok=True)
    tmp_path = file_path.with_suffix(file_path.suffix + ".tmp")
    tmp_path.write_text(content, encoding="utf-8")
    os.replace(tmp_path, file_path)

    print(f"Created {file_path.relative_to(root)}")

    # Regenerate manifest
    _regenerate_manifest(root)


def _regenerate_manifest(root: Path) -> None:
    """Regenerate manifest.json after scaffolding."""
    print("Regenerating manifest.json...")
    try:
        from agent_army.manifest import write_manifest

        write_manifest(root)
    except ImportError:
        print("(manifest module not available, skipping)")


# ---------------------------------------------------------------------------
# Utilities
# ---------------------------------------------------------------------------


def _name_to_title(name: str) -> str:
    """Convert a name like ``go/testing-patterns`` to ``Go Testing Patterns``."""
    parts = name.replace("/", " ").replace("-", " ").split()
    return " ".join(p.capitalize() for p in parts)
