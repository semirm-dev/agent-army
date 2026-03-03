"""CLI entry points for agent-army.

Usage::

    python -m agent_army manifest   # regenerate manifest.json
    python -m agent_army resolve    # validate refs + fix redundancies
    python -m agent_army edit       # interactive dependency editor
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path


def _find_root() -> Path:
    """Locate the repository root (directory containing rules/)."""
    cwd = Path.cwd()
    for candidate in [cwd, cwd.parent, cwd.parent.parent]:
        if (candidate / "rules").is_dir():
            return candidate
    return cwd


def main_manifest(root: Path) -> None:
    """Regenerate manifest.json."""
    from agent_army.manifest import write_manifest

    write_manifest(root)


def main_resolve(root: Path) -> None:
    """Validate all dependency references and remove redundancies."""
    from agent_army.loader import load_agents, load_plugins, load_rules, load_skills
    from agent_army.manifest import write_manifest
    from agent_army.resolver import (
        apply_fixes,
        compute_all_fixes,
        format_report,
        validate_all_refs,
    )

    rules = load_rules(root)
    skills = load_skills(root)
    agents = load_agents(root)
    plugins = load_plugins(root)

    errors = validate_all_refs(rules, skills, agents, plugins)
    fixes = compute_all_fixes(rules, skills, agents, root)

    report = format_report(errors, fixes)
    print(report)

    real_errors = [e for e in errors if e.severity == "error"]
    if real_errors:
        sys.exit(1)

    if not fixes:
        return

    try:
        confirm = input("Remove redundant entries? [y/N] ")
    except (EOFError, KeyboardInterrupt):
        print("\nAborted. No files changed.")
        return

    if confirm.strip().lower() != "y":
        print("Aborted. No files changed.")
        return

    print()
    apply_fixes(fixes, root)
    for fix in fixes:
        print(f"Updated {fix.label}")

    print()
    print("Regenerating manifest.json...")
    write_manifest(root)


def main_edit(root: Path) -> None:
    """Interactive dependency editor."""
    from agent_army.editor import edit_flow

    edit_flow(root)


def main() -> None:
    """Parse arguments and dispatch to subcommand."""
    parser = argparse.ArgumentParser(
        prog="agent-army",
        description="Manage dependencies across rules, skills, and agents.",
    )
    sub = parser.add_subparsers(dest="command")

    sub.add_parser("manifest", help="Regenerate manifest.json")
    sub.add_parser("resolve", help="Validate refs and fix redundancies")
    sub.add_parser("edit", help="Interactive dependency editor")

    args = parser.parse_args()
    if not args.command:
        parser.print_help()
        sys.exit(1)

    root = _find_root()

    dispatch = {
        "manifest": main_manifest,
        "resolve": main_resolve,
        "edit": main_edit,
    }
    dispatch[args.command](root)
