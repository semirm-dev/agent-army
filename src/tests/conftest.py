"""Shared test fixtures for agent-army tests."""

from __future__ import annotations

from pathlib import Path
from textwrap import dedent

import pytest


@pytest.fixture
def sample_rule_content() -> str:
    return dedent("""\
        ---
        name: go/patterns
        description: Go coding conventions, error handling, project structure, and concurrency
        scope: language-specific
        languages: [go]
        uses_rules: [code-quality, security, cross-cutting, observability]
        ---

        # Go Coding Patterns

        Some content here.
    """)


@pytest.fixture
def sample_skill_content() -> str:
    return dedent("""\
        ---
        name: api-designer
        description: API style selection, REST resource design, versioning strategy, pagination decision tree, error format guidance, GraphQL patterns, and RPC streaming.
        scope: universal
        languages: []
        uses_rules: [api-design, cross-cutting, security]
        ---

        # API Designer

        Some content here.
    """)


@pytest.fixture
def sample_agent_content() -> str:
    return dedent("""\
        ---
        name: go/coder
        description: "Senior Go engineer. Writes production-grade Go code following project patterns."
        role: coder
        scope: language-specific
        languages: [go]
        access: read-write
        uses_skills: [go/coder, go/architect, error-handling, code-architecture, api-designer, refactoring-patterns, caching-strategy, messaging-patterns, observability-setup]
        uses_rules: []
        uses_plugins: [code-simplifier, context7]
        delegates_to: []
        ---

        # Go Coder Agent

        Some content here.
    """)


@pytest.fixture
def sample_tree(tmp_path: Path) -> Path:
    """Create a minimal rules/skills/agents tree for testing loaders."""
    root = tmp_path

    # Rules
    rules_dir = root / "rules"
    rules_dir.mkdir()
    (rules_dir / "cross-cutting.md").write_text(dedent("""\
        ---
        scope: universal
        ---

        # Cross-Cutting Standards
    """))
    (rules_dir / "security.md").write_text(dedent("""\
        ---
        scope: universal
        ---

        # Security Patterns
    """))
    (rules_dir / "code-quality.md").write_text(dedent("""\
        ---
        scope: universal
        ---

        # Code Quality
    """))
    go_rules = rules_dir / "go"
    go_rules.mkdir()
    (go_rules / "patterns.md").write_text(dedent("""\
        ---
        name: go/patterns
        scope: language-specific
        languages: [go]
        uses_rules: [code-quality, security, cross-cutting]
        ---

        # Go Coding Patterns
    """))

    # Skills
    skills_dir = root / "skills"
    skills_dir.mkdir()
    (skills_dir / "error-handling.md").write_text(dedent("""\
        ---
        name: error-handling
        scope: universal
        languages: []
        uses_rules: [cross-cutting]
        ---

        # Error Handling
    """))
    go_skills = skills_dir / "go"
    go_skills.mkdir()
    (go_skills / "coder.md").write_text(dedent("""\
        ---
        name: go/coder
        scope: language-specific
        languages: [go]
        uses_rules: [go/patterns]
        ---

        # Go Coder Skill
    """))

    # Agents
    agents_dir = root / "agents"
    agents_dir.mkdir()
    (agents_dir / "type-design-analyzer.md").write_text(dedent("""\
        ---
        name: type-design-analyzer
        description: "Expert type design analyst."
        role: analyzer
        scope: universal
        access: read-only
        uses_skills: []
        uses_rules: []
        uses_plugins: []
        delegates_to: []
        ---

        # Type Design Analyzer
    """))
    go_agents = agents_dir / "go"
    go_agents.mkdir()
    (go_agents / "coder.md").write_text(dedent("""\
        ---
        name: go/coder
        description: "Senior Go engineer. Writes production-grade Go code."
        role: coder
        scope: language-specific
        languages: [go]
        access: read-write
        uses_skills: [go/coder, error-handling]
        uses_rules: []
        uses_plugins: [code-simplifier]
        delegates_to: [type-design-analyzer]
        ---

        # Go Coder Agent
    """))

    # config.json
    import json
    config = {
        "public_plugins": [
            {"name": "code-simplifier"},
            {"name": "context7"},
        ]
    }
    (root / "config.json").write_text(json.dumps(config, indent=2))

    return root
