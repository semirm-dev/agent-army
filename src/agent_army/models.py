"""Domain models for rules, skills, agents, and validation results."""

from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path


@dataclass
class Rule:
    """A rule definition loaded from a markdown file with YAML frontmatter.

    Attributes:
        name: Derived from relative path (e.g. "go/patterns").
        description: First H1 heading after frontmatter.
        scope: "universal" or "language-specific".
        languages: Language tags (e.g. ["go"]).
        uses_rules: Direct rule dependencies.
        path: Relative path from repo root (e.g. Path("rules/go/patterns.md")).
    """

    name: str
    description: str
    scope: str
    languages: list[str] = field(default_factory=list)
    uses_rules: list[str] = field(default_factory=list)
    path: Path = field(default_factory=lambda: Path())


@dataclass
class Skill:
    """A skill definition loaded from a markdown file with YAML frontmatter.

    Attributes:
        name: From frontmatter 'name' field, fallback to relative path.
        description: First H1 heading after frontmatter.
        scope: "universal" or "language-specific".
        languages: Language tags.
        uses_rules: Direct rule dependencies.
        path: Relative path from repo root (e.g. Path("skills/api-designer.md")).
    """

    name: str
    description: str
    scope: str
    languages: list[str] = field(default_factory=list)
    uses_rules: list[str] = field(default_factory=list)
    path: Path = field(default_factory=lambda: Path())


@dataclass
class Agent:
    """An agent definition loaded from a markdown file with YAML frontmatter.

    Attributes:
        name: From frontmatter 'name' field, fallback to relative path.
        description: From frontmatter 'description' field.
        role: Agent role (e.g. "coder", "reviewer", "tester").
        scope: "universal" or "language-specific".
        access: "read-write" or "read-only".
        languages: Language tags.
        uses_skills: Direct skill dependencies.
        uses_rules: Direct rule dependencies.
        uses_plugins: Plugin references.
        delegates_to: Agent delegation targets.
        path: Relative path from repo root (e.g. Path("agents/go/coder.md")).
    """

    name: str
    description: str
    role: str
    scope: str
    access: str
    languages: list[str] = field(default_factory=list)
    uses_skills: list[str] = field(default_factory=list)
    uses_rules: list[str] = field(default_factory=list)
    uses_plugins: list[str] = field(default_factory=list)
    delegates_to: list[str] = field(default_factory=list)
    path: Path = field(default_factory=lambda: Path())


@dataclass(frozen=True)
class Redundancy:
    """A dependency entry that is transitively covered by another entry.

    Attributes:
        target: The redundant entry name.
        covered_by: The entry that already covers the target transitively.
    """

    target: str
    covered_by: str


@dataclass(frozen=True)
class ValidationError:
    """A reference validation error found during dependency resolution.

    Attributes:
        file_label: Display label for the file (e.g. "rules/go/patterns.md").
        field: The frontmatter field containing the bad reference.
        ref: The invalid reference value.
        severity: "error" or "warning".
    """

    file_label: str
    field: str
    ref: str
    severity: str


@dataclass
class Fix:
    """A proposed fix for a frontmatter field in a file.

    Attributes:
        label: Display label (e.g. "agents/go/coder.md").
        field: The frontmatter field to modify.
        file_path: Absolute or relative path to the file.
        before: Original list of values.
        after: Corrected list of values.
        reasons: Human-readable reasons for each change.
    """

    label: str
    field: str
    file_path: Path
    before: list[str] = field(default_factory=list)
    after: list[str] = field(default_factory=list)
    reasons: list[str] = field(default_factory=list)
