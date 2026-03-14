# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Agent Army is a bootstrapping system that generates platform-specific orchestration files (CLAUDE.md, Cursor AGENTS.md) from a unified spec library. It organizes AI development guidance into two layers: **Skills** (standards + workflows) → **Agents** (specialized roles), with transitive dependency resolution.

## Build & Test Commands

```bash
make build              # Build the Go CLI binary (army/army)
make test               # Run all Go tests with race detection
make manifest           # Scan spec/ frontmatter, resolve transitive deps, generate manifest.json
make resolve-deps       # Validate all dependency references, remove redundancies
make bootstrap          # Generate platform-specific output into .build/
make sync               # Install all plugins and skills from PLUGINS_AND_SKILLS.md
make update-plugins-skills  # Regenerate PLUGINS_AND_SKILLS.md from system state
make analyze            # Analyze installed plugins and skills, report duplicates
make analyze-fix        # Analyze and fix skill lock drift (remove stale entries)
make build-v2           # Build armyv2 CLI binary (armyv2/armyv2)
make test-v2            # Run armyv2 tests with race detection
```

Run a single test package:
```bash
cd army && go test ./internal/graph/... -race -count=1
```

## Architecture

### Go CLI (`army/`)

Entry point: `army/cmd/army/main.go` → `cli.NewRootCmd()`

Key internal packages:
- **`bootstrap/`** — Generates CLAUDE.md and Cursor AGENTS.md from spec templates + manifest
- **`manifest/`** — Builds manifest.json with transitive dependency resolution
- **`graph/`** — Dependency graph traversal for skills/agents
- **`frontmatter/`** — YAML frontmatter parsing/writing for spec files
- **`loader/`** — Loads skills, agents from `spec/` directory
- **`resolver/`** — Conflict resolution for transitive dependencies
- **`model/`** — Core data types: Skill, Agent
- **`plugindoc/`** — Generates PLUGINS_AND_SKILLS.md and terminal analysis reports for installed plugins/skills
- **`pluginsync/`** — Reads PLUGINS_AND_SKILLS.md and executes plugin/skill install + redundant skill cleanup
- **`termcolor/`** — ANSI color helpers for formatted CLI output

Dependencies: `cobra` (CLI framework), `gopkg.in/yaml.v3` (YAML parsing)

### Go CLI (`armyv2/`)

Entry point: `armyv2/cmd/armyv2/main.go` → `cli.NewRootCmd()`

Ports & Adapters architecture:
- **`internal/core/`** — Pure domain logic: catalog, manifest, detector, orchestrator, diff, doctor, types
- **`internal/port/`** — Presentation: TUI (Bubble Tea) + CLI (Cobra)
- **`internal/adapter/`** — System integration: plugin installer, skill installer, system reader, command runner

Commands: `setup`, `sync`, `add`, `remove`, `list`, `diff`, `update`, `doctor`

Dependencies: `cobra`, `bubbletea`, `bubbles`, `lipgloss`

### Spec Library (`spec/`)

All specs use YAML frontmatter + Markdown content:
- **`skills/`** (30 files) — Standards + workflow definitions with `uses_skills` dependencies
- **`agents/`** (21 files) — Role definitions with `uses_skills`, `delegates_to`
- **`claude/`** — Claude Code platform template (`CLAUDE.md`, `settings.json`)
- **`cursor/`** — Cursor platform template

### Key Files

- **`manifest.json`** — Auto-generated index of all skills and agents with resolved transitive dependencies. Regenerate with `make manifest` after any spec change.
- **`Makefile`** — All build orchestration
- **`.build/`** — Generated output directory (gitignored)
- **`PLUGINS_AND_SKILLS.md`** — Auto-generated report of installed Claude Code plugins and skills. Regenerate with `make update-plugins-skills`.

## Development Workflow

1. Edit specs in `spec/` (skills or agents)
2. Run `make resolve-deps` to validate dependency references
3. Run `make manifest` to regenerate `manifest.json`
4. Run `make bootstrap` to produce platform output in `.build/`
5. Run `make test` to verify nothing broke

## Conventions

- Go CLI uses standard `internal/` package layout — no exported API
- Spec frontmatter keys vary by type: skills have `uses_skills` and workflow context; agents add `role`/`access`/`delegates_to`/`uses_skills`/`uses_plugins`
- Commit messages follow Conventional Commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`
