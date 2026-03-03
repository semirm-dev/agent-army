"""Interactive dependency editor for rules, skills, and agents.

Ports the interactive workflow from ``scripts/edit-deps.sh`` into Python.
Uses ``input()`` for all user interaction (testable via monkeypatch).
"""

from __future__ import annotations

from pathlib import Path

from agent_army.frontmatter import parse_frontmatter, write_field
from agent_army.graph import find_redundant, find_redundant_via_skills
from agent_army.loader import (
    find_md_files,
    load_agents,
    load_plugins,
    load_rules,
    load_skills,
)

# Entity type -> subdirectory name
_ENTITY_DIRS: dict[str, str] = {
    "rule": "rules",
    "skill": "skills",
    "agent": "agents",
}

# Fields available per entity type
_ENTITY_FIELDS: dict[str, list[str]] = {
    "rule": ["uses_rules"],
    "skill": ["uses_rules"],
    "agent": ["uses_rules", "uses_skills", "uses_plugins", "delegates_to"],
}


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------


def select_one(prompt: str, items: list[str]) -> str:
    """Numbered menu, single selection. Returns selected value.

    Displays a numbered list and prompts until the user enters a valid
    number.  Uses ``input()`` for testability.
    """
    _print_numbered_menu(items)
    while True:
        raw = input(f"{prompt} ")
        choice = _parse_int(raw)
        if choice is not None and 1 <= choice <= len(items):
            return items[choice - 1]
        print(f"Invalid choice. Enter a number between 1 and {len(items)}.")


def select_multi(prompt: str, items: list[str]) -> list[str]:
    """Numbered menu, comma-separated multi-selection. Returns list.

    Must select at least one valid item. Re-prompts on invalid input.
    """
    _print_numbered_menu(items)
    while True:
        raw = input(f"{prompt} (comma-separated, e.g. 1,3,5): ")
        selected = _parse_multi_choice(raw, len(items))
        if selected is not None and len(selected) > 0:
            return [items[i - 1] for i in selected]
        # Re-prompt (error already printed by _parse_multi_choice or implicit)


def edit_flow(root: Path) -> None:
    """Full interactive edit flow matching edit-deps.sh."""
    print("=== Edit Dependencies ===")

    # 1. Choose entity type
    entity_type = select_one("Choose entity type:", list(_ENTITY_DIRS.keys()))

    # 2. Choose file
    entity_dir = root / _ENTITY_DIRS[entity_type]
    file_path, display_name = _choose_file(entity_dir)

    # 3. Choose field
    field = _choose_field(entity_type)

    # 4. Show current values
    prefix = _ENTITY_DIRS[entity_type]
    print(f"\nFile: {prefix}/{display_name}")
    current = _read_current_values(file_path, field)
    _print_current_values(field, current)

    # 5. Choose action
    action = select_one("Action:", ["add", "remove"])

    # 6. Compute new values
    new_values = _apply_action(action, field, current, root)
    if new_values is None:
        return  # early exit (nothing to add/remove)

    # 7. Redundancy check
    _check_redundancy(field, new_values, entity_type, file_path, root)

    # 8. Diff preview
    _print_diff_preview(current, new_values)

    # 9. Confirm
    confirm = input("Apply this change? [y/N] ")
    if confirm.strip().lower() != "y":
        print("Aborted.")
        return

    # 10. Write and regenerate manifest
    write_field(file_path, field, new_values)
    print(f"Updated {prefix}/{display_name}")
    _regenerate_manifest(root)


# ---------------------------------------------------------------------------
# Private helpers — menu display
# ---------------------------------------------------------------------------


def _print_numbered_menu(items: list[str]) -> None:
    """Print a numbered list of items."""
    print()
    for i, item in enumerate(items, start=1):
        print(f"  {i}) {item}")
    print()


def _parse_int(raw: str) -> int | None:
    """Parse a string to int, returning None on failure."""
    try:
        return int(raw.strip())
    except ValueError:
        return None


def _parse_multi_choice(raw: str, max_val: int) -> list[int] | None:
    """Parse comma-separated numbers. Returns list of ints or None on error."""
    parts = [p.strip() for p in raw.split(",")]
    result: list[int] = []
    for part in parts:
        num = _parse_int(part)
        if num is None or num < 1 or num > max_val:
            print(f"Invalid number: {part.strip()}")
            return None
        result.append(num)
    return result if result else None


# ---------------------------------------------------------------------------
# Private helpers — flow steps
# ---------------------------------------------------------------------------


def _choose_file(entity_dir: Path) -> tuple[Path, str]:
    """Choose a file from the entity directory. Returns (abs_path, display_name)."""
    md_files = find_md_files(entity_dir)
    if not md_files:
        print(f"No files found in {entity_dir}")
        raise SystemExit(1)

    display_names = [str(f.relative_to(entity_dir)) for f in md_files]
    chosen_display = select_one("Choose file:", display_names)
    chosen_path = entity_dir / chosen_display
    return chosen_path, chosen_display


def _choose_field(entity_type: str) -> str:
    """Choose a field for the entity type. Auto-selects for rule/skill."""
    fields = _ENTITY_FIELDS[entity_type]
    if len(fields) == 1:
        print(f"\nField: {fields[0]} (auto-selected)")
        return fields[0]
    return select_one("Choose field:", fields)


def _read_current_values(file_path: Path, field: str) -> list[str]:
    """Read current field values from file frontmatter."""
    content = file_path.read_text(encoding="utf-8")
    fm = parse_frontmatter(content)
    raw = fm.get(field)
    if raw is None:
        return []
    if isinstance(raw, list):
        return raw
    stripped = raw.strip()
    return [stripped] if stripped else []


def _print_current_values(field: str, current: list[str]) -> None:
    """Print the current field values."""
    if not current:
        print(f"Current {field}: (none)")
    else:
        print(f"Current {field}: [{', '.join(current)}]")


def _apply_action(
    action: str,
    field: str,
    current: list[str],
    root: Path,
) -> list[str] | None:
    """Apply add/remove action. Returns new values or None to abort."""
    if action == "add":
        return _action_add(field, current, root)
    return _action_remove(field, current)


def _action_add(
    field: str,
    current: list[str],
    root: Path,
) -> list[str] | None:
    """Handle the 'add' action. Returns new values or None if nothing to add."""
    valid = _load_valid_values(field, root)
    current_set = set(current)
    available = [v for v in valid if v not in current_set]

    if not available:
        print("All values are already present. Nothing to add.")
        return None

    selected = select_multi("Select entries to add", available)
    return current + selected


def _action_remove(field: str, current: list[str]) -> list[str] | None:
    """Handle the 'remove' action. Returns new values or None if nothing to remove."""
    if not current:
        print(f"No {field} entries to remove.")
        return None

    to_remove = select_multi("Select entries to remove", current)
    remove_set = set(to_remove)
    return [v for v in current if v not in remove_set]


def _load_valid_values(field: str, root: Path) -> list[str]:
    """Load valid values for a given field from the filesystem."""
    if field == "uses_rules":
        return [r.name for r in load_rules(root)]
    if field == "uses_skills":
        return [s.name for s in load_skills(root)]
    if field == "uses_plugins":
        return load_plugins(root)
    if field == "delegates_to":
        return [a.name for a in load_agents(root)]
    return []


# ---------------------------------------------------------------------------
# Private helpers — redundancy
# ---------------------------------------------------------------------------


def _check_redundancy(
    field: str,
    new_values: list[str],
    entity_type: str,
    file_path: Path,
    root: Path,
) -> None:
    """Run redundancy checks and print warnings."""
    if not new_values:
        return

    if field == "uses_rules":
        _check_rule_redundancy(new_values, entity_type, file_path, root)
    elif field == "delegates_to":
        _check_delegate_redundancy(new_values, root)


def _check_rule_redundancy(
    new_values: list[str],
    entity_type: str,
    file_path: Path,
    root: Path,
) -> None:
    """Check rule-to-rule and skill-transitive redundancy."""
    rules = load_rules(root)
    rule_lookup = {r.name: r.uses_rules for r in rules}

    # Rule-to-rule redundancy
    redundancies = find_redundant(new_values, lambda name: rule_lookup.get(name, []))
    _print_redundancy_warnings(redundancies, "rule-to-rule")

    # Skill-transitive check (agents only)
    if entity_type != "agent":
        return

    content = file_path.read_text(encoding="utf-8")
    fm = parse_frontmatter(content)
    agent_skills = fm.get("uses_skills", [])
    if isinstance(agent_skills, str):
        agent_skills = [agent_skills] if agent_skills.strip() else []

    if not agent_skills:
        return

    skills = load_skills(root)
    skill_lookup = {s.name: s.uses_rules for s in skills}
    skill_redundancies = find_redundant_via_skills(
        new_values, agent_skills, skill_lookup, rule_lookup
    )
    _print_redundancy_warnings(skill_redundancies, "covered via skills")


def _check_delegate_redundancy(new_values: list[str], root: Path) -> None:
    """Check delegates_to redundancy."""
    agents = load_agents(root)
    agent_lookup = {a.name: a.delegates_to for a in agents}
    redundancies = find_redundant(
        new_values, lambda name: agent_lookup.get(name, [])
    )
    _print_redundancy_warnings(redundancies, "delegate")


def _print_redundancy_warnings(
    redundancies: list,
    label: str,
) -> None:
    """Print redundancy warnings if any found."""
    if not redundancies:
        return
    print(f"\nWarning: Redundant entries detected ({label}):")
    for r in redundancies:
        print(f'  - "{r.target}" is already included transitively by "{r.covered_by}"')


# ---------------------------------------------------------------------------
# Private helpers — diff and write
# ---------------------------------------------------------------------------


def _print_diff_preview(before: list[str], after: list[str]) -> None:
    """Print the before/after diff preview."""
    print("\n--- Change Preview ---")
    if not before:
        print("  Before: (none)")
    else:
        print(f"  Before: [{', '.join(before)}]")
    if not after:
        print("  After:  (none -- field will be cleared)")
    else:
        print(f"  After:  [{', '.join(after)}]")
    print()


def _regenerate_manifest(root: Path) -> None:
    """Regenerate manifest.json after editing."""
    print("\nRegenerating manifest.json...")
    try:
        from agent_army.manifest import write_manifest  # type: ignore[import-not-found]

        write_manifest(root)
    except ImportError:
        # Manifest module not yet available -- skip silently
        print("(manifest module not available, skipping)")
