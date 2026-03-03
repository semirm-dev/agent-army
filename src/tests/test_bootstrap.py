"""Tests for agent_army.bootstrap."""

from __future__ import annotations

from pathlib import Path
from textwrap import dedent

import pytest

from agent_army.bootstrap import (
    _assign_cursor_numbers,
    _categorize_rule,
    _cursor_short_name,
    _detect_language,
    _extract_body,
    _flatten_name,
    _resolve_collision,
    _rule_to_claude,
    _rule_to_cursor,
    _agent_to_claude,
    _agent_to_cursor,
    _generate_all,
    main_bootstrap,
)
from agent_army.models import Agent, Rule


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------


@pytest.fixture
def bootstrap_tree(tmp_path: Path) -> Path:
    """Create a minimal rules/skills/agents tree for bootstrap tests."""
    root = tmp_path / "repo"
    root.mkdir()

    # Rules
    rules_dir = root / "rules"
    rules_dir.mkdir()
    (rules_dir / "security.md").write_text(dedent("""\
        ---
        name: security
        description: Security patterns and conventions
        scope: universal
        languages: []
        uses_rules: []
        ---

        # Security Patterns

        ## Password Hashing
        - Use bcrypt or argon2.
    """))
    (rules_dir / "api-design.md").write_text(dedent("""\
        ---
        name: api-design
        description: API design patterns
        scope: universal
        languages: []
        uses_rules: []
        ---

        # API Design Patterns

        ## Error Response Format
        - Use consistent structure.
    """))
    go_rules = rules_dir / "go"
    go_rules.mkdir()
    (go_rules / "patterns.md").write_text(dedent("""\
        ---
        name: go/patterns
        description: Go coding conventions
        scope: language-specific
        languages: [go]
        uses_rules: [security]
        ---

        # Go Coding Patterns

        ## Naming
        - Use MixedCaps.
    """))

    # Skills
    skills_dir = root / "skills"
    skills_dir.mkdir()
    (skills_dir / "error-handling.md").write_text(dedent("""\
        ---
        name: error-handling
        description: Error handling workflow
        scope: universal
        languages: []
        uses_rules: [security]
        ---

        # Error Handling Skill

        ## When to Use
        Invoke when designing error handling.
    """))
    go_skills = skills_dir / "go"
    go_skills.mkdir()
    (go_skills / "coder.md").write_text(dedent("""\
        ---
        name: go/coder
        description: Go coder skill
        scope: language-specific
        languages: [go]
        uses_rules: [go/patterns]
        ---

        # Go Coder Skill

        ## Workflow
        Write Go code.
    """))

    # Agents
    agents_dir = root / "agents"
    agents_dir.mkdir()
    (agents_dir / "arch-reviewer.md").write_text(dedent("""\
        ---
        name: arch-reviewer
        description: "Senior architecture reviewer. Read-only analysis."
        role: reviewer
        scope: universal
        languages: []
        access: read-only
        uses_skills: []
        uses_rules: [security]
        uses_plugins: []
        delegates_to: []
        ---

        # Architecture Reviewer Agent

        ## Role
        You review architecture using `Edit` and `Bash` tools.
        Config at ~/.claude/rules.
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
        uses_plugins: []
        delegates_to: []
        ---

        # Go Coder Agent

        ## Role
        You write Go code.
    """))

    return root


# ---------------------------------------------------------------------------
# Unit tests — pure functions
# ---------------------------------------------------------------------------


class TestFlattenName:
    """_flatten_name() — converts path names to flat names."""

    @pytest.mark.parametrize(
        "name, expected",
        [
            ("security", "security"),
            ("go/patterns", "go-patterns"),
            ("go/testing", "go-testing"),
            ("typescript/patterns", "typescript-patterns"),
        ],
    )
    def test_flatten(self, name: str, expected: str) -> None:
        assert _flatten_name(name) == expected


class TestDetectLanguage:
    """_detect_language() — identifies the primary language for a rule."""

    def test_from_name_prefix(self) -> None:
        rule = Rule(name="go/patterns", description="", scope="language-specific", languages=["go"])
        assert _detect_language(rule) == "go"

    def test_from_languages_list(self) -> None:
        rule = Rule(name="custom", description="", scope="language-specific", languages=["python"])
        assert _detect_language(rule) == "python"

    def test_universal_no_language(self) -> None:
        rule = Rule(name="security", description="", scope="universal")
        assert _detect_language(rule) is None


class TestCategorizeRule:
    """_categorize_rule() — assigns Cursor category to rules."""

    @pytest.mark.parametrize(
        "name, expected_cat",
        [
            ("go/patterns", "language"),
            ("go/testing", "language"),
            ("typescript/patterns", "language"),
            ("python/testing", "language"),
            ("react/patterns", "language"),
            ("git-workflow", "git"),
            ("api-design", "api-db"),
            ("database", "api-db"),
            ("security", "infrastructure"),
            ("observability", "infrastructure"),
            ("cross-cutting", "infrastructure"),
        ],
    )
    def test_categories(self, name: str, expected_cat: str) -> None:
        rule = Rule(name=name, description="", scope="universal")
        assert _categorize_rule(rule) == expected_cat


class TestCursorShortName:
    """_cursor_short_name() — generates short names for Cursor rules."""

    def test_explicit_mapping(self) -> None:
        rule = Rule(name="go/patterns", description="", scope="language-specific")
        assert _cursor_short_name(rule) == "golang"

    def test_fallback_to_flatten(self) -> None:
        rule = Rule(name="security", description="", scope="universal")
        assert _cursor_short_name(rule) == "security"

    def test_typescript_testing(self) -> None:
        rule = Rule(name="typescript/testing", description="", scope="language-specific")
        assert _cursor_short_name(rule) == "typescript-testing"


class TestAssignCursorNumbers:
    """_assign_cursor_numbers() — assigns numbers by category."""

    def test_numbering_preserves_order(self) -> None:
        rules = [
            Rule(name="go/patterns", description="", scope="language-specific", languages=["go"]),
            Rule(name="security", description="", scope="universal"),
            Rule(name="api-design", description="", scope="universal"),
            Rule(name="typescript/patterns", description="", scope="language-specific", languages=["typescript"]),
        ]
        result = _assign_cursor_numbers(rules)
        assert len(result) == 4
        # go/patterns -> language 100
        assert result[0] == (100, "golang")
        # security -> infrastructure 500
        assert result[1] == (500, "security")
        # api-design -> api-db 400
        assert result[2] == (400, "api-design")
        # typescript/patterns -> language 101
        assert result[3] == (101, "typescript")

    def test_sequential_within_category(self) -> None:
        rules = [
            Rule(name="go/patterns", description="", scope="language-specific"),
            Rule(name="go/testing", description="", scope="language-specific"),
            Rule(name="typescript/patterns", description="", scope="language-specific"),
        ]
        result = _assign_cursor_numbers(rules)
        assert result[0][0] == 100  # go/patterns
        assert result[1][0] == 101  # go/testing
        assert result[2][0] == 102  # typescript/patterns


class TestExtractBody:
    """_extract_body() — strips frontmatter from file content."""

    def test_strips_frontmatter(self, tmp_path: Path) -> None:
        f = tmp_path / "test.md"
        f.write_text(dedent("""\
            ---
            name: test
            scope: universal
            ---

            # Test Heading

            Body content here.
        """))
        body = _extract_body(f)
        assert body.startswith("# Test Heading")
        assert "name: test" not in body
        assert "---" not in body

    def test_no_frontmatter(self, tmp_path: Path) -> None:
        f = tmp_path / "test.md"
        content = "# Just a heading\n\nSome content."
        f.write_text(content)
        body = _extract_body(f)
        assert body == content


class TestResolveCollision:
    """_resolve_collision() — appends suffix when file exists."""

    def test_no_collision(self, tmp_path: Path) -> None:
        rel = Path("rules") / "100-golang.mdc"
        result = _resolve_collision(tmp_path, rel)
        assert result == rel

    def test_collision_appends_suffix(self, tmp_path: Path) -> None:
        (tmp_path / "rules").mkdir()
        (tmp_path / "rules" / "100-golang.mdc").write_text("existing")
        rel = Path("rules") / "100-golang.mdc"
        result = _resolve_collision(tmp_path, rel)
        assert result == Path("rules") / "100-golang_2.mdc"

    def test_multiple_collisions(self, tmp_path: Path) -> None:
        (tmp_path / "rules").mkdir()
        (tmp_path / "rules" / "100-golang.mdc").write_text("existing")
        (tmp_path / "rules" / "100-golang_2.mdc").write_text("existing")
        rel = Path("rules") / "100-golang.mdc"
        result = _resolve_collision(tmp_path, rel)
        assert result == Path("rules") / "100-golang_3.mdc"


# ---------------------------------------------------------------------------
# Transformer tests
# ---------------------------------------------------------------------------


class TestRuleToClaude:
    """_rule_to_claude() — transforms rules for Claude Code output."""

    def test_strips_frontmatter_adds_sync(self, bootstrap_tree: Path) -> None:
        from agent_army.loader import load_rules
        rules = load_rules(bootstrap_tree)
        security = next(r for r in rules if r.name == "security")
        result = _rule_to_claude(bootstrap_tree, security)
        assert result.startswith("<!-- Sync: security from agent-army repo -->")
        assert "# Security Patterns" in result
        assert "name: security" not in result

    def test_go_rule(self, bootstrap_tree: Path) -> None:
        from agent_army.loader import load_rules
        rules = load_rules(bootstrap_tree)
        go = next(r for r in rules if r.name == "go/patterns")
        result = _rule_to_claude(bootstrap_tree, go)
        assert "<!-- Sync: go/patterns from agent-army repo -->" in result
        assert "# Go Coding Patterns" in result


class TestRuleToCursor:
    """_rule_to_cursor() — transforms rules for Cursor output."""

    def test_language_rule_gets_globs(self, bootstrap_tree: Path) -> None:
        from agent_army.loader import load_rules
        rules = load_rules(bootstrap_tree)
        go = next(r for r in rules if r.name == "go/patterns")
        result = _rule_to_cursor(bootstrap_tree, go)
        assert 'globs: "**/*.go"' in result
        assert "alwaysApply" not in result
        assert "# Go Coding Patterns" in result

    def test_universal_rule_gets_always_apply(self, bootstrap_tree: Path) -> None:
        from agent_army.loader import load_rules
        rules = load_rules(bootstrap_tree)
        security = next(r for r in rules if r.name == "security")
        result = _rule_to_cursor(bootstrap_tree, security)
        assert "alwaysApply: true" in result
        assert "globs" not in result

    def test_cursor_frontmatter_has_description(self, bootstrap_tree: Path) -> None:
        from agent_army.loader import load_rules
        rules = load_rules(bootstrap_tree)
        security = next(r for r in rules if r.name == "security")
        result = _rule_to_cursor(bootstrap_tree, security)
        assert "description: Security Patterns" in result


class TestAgentToClaude:
    """_agent_to_claude() — transforms agents for Claude Code output."""

    def test_read_write_tools(self, bootstrap_tree: Path) -> None:
        from agent_army.loader import load_agents
        agents = load_agents(bootstrap_tree)
        coder = next(a for a in agents if a.name == "go/coder")
        result = _agent_to_claude(bootstrap_tree, coder)
        assert "tools: Read, Write, Edit, Bash, Glob, Grep" in result
        assert "name: go-coder" in result
        assert "model: inherit" in result

    def test_read_only_tools(self, bootstrap_tree: Path) -> None:
        from agent_army.loader import load_agents
        agents = load_agents(bootstrap_tree)
        reviewer = next(a for a in agents if a.name == "arch-reviewer")
        result = _agent_to_claude(bootstrap_tree, reviewer)
        assert "tools: Read, Glob, Grep, Bash" in result

    def test_body_preserved(self, bootstrap_tree: Path) -> None:
        from agent_army.loader import load_agents
        agents = load_agents(bootstrap_tree)
        coder = next(a for a in agents if a.name == "go/coder")
        result = _agent_to_claude(bootstrap_tree, coder)
        assert "# Go Coder Agent" in result
        assert "You write Go code." in result


class TestAgentToCursor:
    """_agent_to_cursor() — transforms agents for Cursor output."""

    def test_no_tools_field(self, bootstrap_tree: Path) -> None:
        from agent_army.loader import load_agents
        agents = load_agents(bootstrap_tree)
        coder = next(a for a in agents if a.name == "go/coder")
        result = _agent_to_cursor(bootstrap_tree, coder)
        assert "tools:" not in result
        assert "model:" not in result
        assert "name: go-coder" in result

    def test_tool_substitutions(self, bootstrap_tree: Path) -> None:
        from agent_army.loader import load_agents
        agents = load_agents(bootstrap_tree)
        reviewer = next(a for a in agents if a.name == "arch-reviewer")
        result = _agent_to_cursor(bootstrap_tree, reviewer)
        assert "`StrReplace`" in result
        assert "`Shell`" in result
        assert "`Edit`" not in result
        assert "`Bash`" not in result

    def test_path_substitution(self, bootstrap_tree: Path) -> None:
        from agent_army.loader import load_agents
        agents = load_agents(bootstrap_tree)
        reviewer = next(a for a in agents if a.name == "arch-reviewer")
        result = _agent_to_cursor(bootstrap_tree, reviewer)
        assert "~/.cursor/rules" in result
        assert "~/.claude/" not in result

    def test_description_without_colon_unquoted(self, bootstrap_tree: Path) -> None:
        from agent_army.loader import load_agents
        agents = load_agents(bootstrap_tree)
        reviewer = next(a for a in agents if a.name == "arch-reviewer")
        result = _agent_to_cursor(bootstrap_tree, reviewer)
        assert "description: Senior architecture reviewer. Read-only analysis." in result


# ---------------------------------------------------------------------------
# Integration tests — _generate_all
# ---------------------------------------------------------------------------


class TestGenerateAll:
    """_generate_all() — writes files to destination directory."""

    def test_claude_output_structure(self, bootstrap_tree: Path, tmp_path: Path) -> None:
        from agent_army.loader import load_agents, load_rules, load_skills
        dest = tmp_path / "output"
        rules = load_rules(bootstrap_tree)
        skills = load_skills(bootstrap_tree)
        agents = load_agents(bootstrap_tree)

        written = _generate_all(bootstrap_tree, dest, rules, skills, agents, is_claude=True)

        assert written == len(rules) + len(skills) + len(agents)

        # Rules
        assert (dest / "rules" / "security.md").exists()
        assert (dest / "rules" / "api-design.md").exists()
        assert (dest / "rules" / "go-patterns.md").exists()

        # Skills
        assert (dest / "skills" / "error-handling" / "SKILL.md").exists()
        assert (dest / "skills" / "go-coder" / "SKILL.md").exists()

        # Agents
        assert (dest / "agents" / "arch-reviewer.md").exists()
        assert (dest / "agents" / "go-coder.md").exists()

    def test_cursor_output_structure(self, bootstrap_tree: Path, tmp_path: Path) -> None:
        from agent_army.loader import load_agents, load_rules, load_skills
        dest = tmp_path / "output"
        rules = load_rules(bootstrap_tree)
        skills = load_skills(bootstrap_tree)
        agents = load_agents(bootstrap_tree)

        written = _generate_all(bootstrap_tree, dest, rules, skills, agents, is_claude=False)

        assert written == len(rules) + len(skills) + len(agents)

        # Rules should be .mdc files with numbers
        mdc_files = list((dest / "rules").glob("*.mdc"))
        assert len(mdc_files) == len(rules)

        # Skills still use SKILL.md
        assert (dest / "skills" / "error-handling" / "SKILL.md").exists()

        # Agents
        assert (dest / "agents" / "arch-reviewer.md").exists()

    def test_claude_rule_content(self, bootstrap_tree: Path, tmp_path: Path) -> None:
        from agent_army.loader import load_rules
        dest = tmp_path / "output"
        rules = load_rules(bootstrap_tree)
        security = [r for r in rules if r.name == "security"]

        _generate_all(bootstrap_tree, dest, security, [], [], is_claude=True)

        content = (dest / "rules" / "security.md").read_text()
        assert "<!-- Sync:" in content
        assert "# Security Patterns" in content

    def test_cursor_rule_content(self, bootstrap_tree: Path, tmp_path: Path) -> None:
        from agent_army.loader import load_rules
        dest = tmp_path / "output"
        rules = load_rules(bootstrap_tree)
        go_rules = [r for r in rules if r.name == "go/patterns"]

        _generate_all(bootstrap_tree, dest, go_rules, [], [], is_claude=False)

        mdc_files = list((dest / "rules").glob("*.mdc"))
        assert len(mdc_files) == 1
        content = mdc_files[0].read_text()
        assert 'globs: "**/*.go"' in content
        assert "# Go Coding Patterns" in content

    def test_empty_selection(self, bootstrap_tree: Path, tmp_path: Path) -> None:
        dest = tmp_path / "output"
        written = _generate_all(bootstrap_tree, dest, [], [], [], is_claude=True)
        assert written == 0


# ---------------------------------------------------------------------------
# Integration tests — full interactive flow
# ---------------------------------------------------------------------------


class TestMainBootstrap:
    """Full flow tests with monkeypatched input."""

    def test_claude_local_all(
        self,
        bootstrap_tree: Path,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Generate all entities for Claude Code to a local directory."""
        monkeypatch.chdir(bootstrap_tree)
        responses = iter([
            "1",   # target: Claude Code
            "1",   # destination: local
            "",    # rules: all
            "",    # skills: all
            "",    # agents: all
            "y",   # confirm
        ])
        monkeypatch.setattr("builtins.input", lambda _: next(responses))

        main_bootstrap(bootstrap_tree)

        dest = bootstrap_tree / ".claude"
        assert (dest / "rules").is_dir()
        assert (dest / "skills").is_dir()
        assert (dest / "agents").is_dir()

    def test_cursor_local_all(
        self,
        bootstrap_tree: Path,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Generate all entities for Cursor to a local directory."""
        monkeypatch.chdir(bootstrap_tree)
        responses = iter([
            "2",   # target: Cursor
            "1",   # destination: local
            "",    # rules: all
            "",    # skills: all
            "",    # agents: all
            "y",   # confirm
        ])
        monkeypatch.setattr("builtins.input", lambda _: next(responses))

        main_bootstrap(bootstrap_tree)

        dest = bootstrap_tree / ".cursor"
        assert (dest / "rules").is_dir()
        mdc_files = list((dest / "rules").glob("*.mdc"))
        assert len(mdc_files) == 3  # 3 rules in fixture

    def test_selective_entities(
        self,
        bootstrap_tree: Path,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Select specific rules, skip skills, select one agent."""
        monkeypatch.chdir(bootstrap_tree)
        responses = iter([
            "1",      # target: Claude Code
            "1",      # destination: local
            "1",      # rules: just first
            "none",   # skills: skip
            "1",      # agents: just first
            "y",      # confirm
        ])
        monkeypatch.setattr("builtins.input", lambda _: next(responses))

        main_bootstrap(bootstrap_tree)

        dest = bootstrap_tree / ".claude"
        rule_files = list((dest / "rules").glob("*.md"))
        assert len(rule_files) == 1
        assert not (dest / "skills").exists()
        agent_files = list((dest / "agents").glob("*.md"))
        assert len(agent_files) == 1

    def test_user_declines(
        self,
        bootstrap_tree: Path,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
        capsys: pytest.CaptureFixture[str],
    ) -> None:
        """User declines at confirmation — no files written."""
        monkeypatch.chdir(bootstrap_tree)
        responses = iter([
            "1",   # target: Claude Code
            "1",   # destination: local
            "",    # rules: all
            "",    # skills: all
            "",    # agents: all
            "n",   # decline
        ])
        monkeypatch.setattr("builtins.input", lambda _: next(responses))

        main_bootstrap(bootstrap_tree)

        dest = bootstrap_tree / ".claude"
        assert not dest.exists()
        captured = capsys.readouterr()
        assert "Aborted" in captured.out

    def test_none_selected(
        self,
        bootstrap_tree: Path,
        monkeypatch: pytest.MonkeyPatch,
        capsys: pytest.CaptureFixture[str],
    ) -> None:
        """Selecting 'none' for all entity types exits early."""
        monkeypatch.chdir(bootstrap_tree)
        responses = iter([
            "1",      # target: Claude Code
            "1",      # destination: local
            "none",   # rules: skip
            "none",   # skills: skip
            "none",   # agents: skip
        ])
        monkeypatch.setattr("builtins.input", lambda _: next(responses))

        main_bootstrap(bootstrap_tree)

        captured = capsys.readouterr()
        assert "No entities selected" in captured.out

    def test_custom_destination(
        self,
        bootstrap_tree: Path,
        tmp_path: Path,
        monkeypatch: pytest.MonkeyPatch,
    ) -> None:
        """Generate to a custom output directory."""
        custom_dest = tmp_path / "custom-output"
        monkeypatch.chdir(bootstrap_tree)
        responses = iter([
            "1",                    # target: Claude Code
            "3",                    # destination: custom
            str(custom_dest),       # custom path
            "1",                    # rules: first only
            "none",                 # skills: skip
            "none",                 # agents: skip
            "y",                    # confirm
        ])
        monkeypatch.setattr("builtins.input", lambda _: next(responses))

        main_bootstrap(bootstrap_tree)

        assert (custom_dest / "rules").is_dir()
