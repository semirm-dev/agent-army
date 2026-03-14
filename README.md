# Agent Army

A modular library of workflows and agent prompts for AI-assisted development. Includes a Go CLI (`army`) to manage specs, resolve dependencies, and generate platform-specific output for Claude Code or Cursor.

> NOTE: **Caution!!** Always backup your current destination setup because `bootstrap` will always do remove-all-existing -> generate new. Test in a safe place (destination) first.

## What's Inside

### Skills (`spec/skills/`)

Structured workflows, decision trees, and coding standards. Markdown files with YAML frontmatter. Skills define **how** to accomplish tasks and **what** good code looks like — designing APIs, setting up caching, hardening security, naming conventions, error handling, testing patterns, etc. Skills can depend on other skills via `uses_skills`.

Examples: `api-design`, `caching-strategy`, `go/coder`, `react/tester`

### Agents (`spec/agents/`)

Prompt templates for specialized AI roles. Grouped by language/domain in subdirectories, with cross-cutting agents at the root. Platform-agnostic — no tool names, model references, or IDE-specific paths. Agents declare `uses_skills` (which transitively bring dependent skills) and `uses_plugins` for extensions.

Examples: `go/coder`, `typescript/reviewer`, `python/tester`, `infrastructure/builder`, `arch-reviewer`

## How They Relate

```
Skills      → task workflows and standards    (the "what" and "how")
Agents      → specialized roles that invoke skills  (the "who")
```

```
┌─────────┐   uses_skills   ┌─────────┐
│ Skills  │◄────────────────│ Agents  │
└─────────┘                 └─────────┘
  api-design                  go/coder
  go/coder                    typescript/reviewer
  react/tester                python/tester
```

## CLIs

Two Go CLIs live side by side. Both are invoked via namespaced make targets:

- **`army/`** (v1) — Spec bootstrapper. Build with `make build`, run commands via `make v1 <cmd>`.
- **`armyv2/`** — Plugin & skill manager. Build with `make build-v2`, run commands via `make v2 <cmd>`.

## Make Commands

| Target | Description |
|--------|-------------|
| `make help` | Show all available targets |
| **army (v1)** | |
| `make build` | Build the army CLI binary |
| `make test` | Run army Go tests with race detection |
| `make v1 manifest` | Scan `spec/` frontmatter and regenerate `manifest.json`. Resolves `uses_skills` and `delegates_to` transitively |
| `make v1 resolve-deps` | Validate all dependency references across `spec/`. Detect and remove redundant `uses_skills` and `delegates_to` entries covered by transitive dependencies |
| `make v1 bootstrap` | Generate model-specific skills and agents (output in `.build/`) |
| `make v1 sync` | Install all plugins and skills listed in `PLUGINS_AND_SKILLS.md` |
| `make v1 update-plugins-skills` | Regenerate `PLUGINS_AND_SKILLS.md` from installed system state |
| `make v1 analyze` | Show installed plugins, skills, and duplicate report (terminal only) |
| `make v1 analyze --fix` | Analyze and fix skill lock drift (remove stale entries) |
| **armyv2** | |
| `make build-v2` | Build armyv2 CLI binary (`armyv2/armyv2`) |
| `make test-v2` | Run armyv2 tests with race detection |
| `make v2 setup` | Interactive setup wizard for plugins and skills |
| `make v2 sync` | Apply manifest — install missing, remove extras (with confirmation) |
| `make v2 list` | Show manifest contents with install status |
| `make v2 diff` | Compare manifest vs installed state |
| `make v2 doctor` | Run health checks on plugins and skills |
| `make v2 update` | Fetch latest catalog from GitHub |
| `make v2 add` | Add a plugin or skill (e.g., `make v2 add plugin context7`) |
| `make v2 remove` | Remove a plugin or skill (e.g., `make v2 remove skill golang-pro`) |

## Plugin & Skill Management

Three v1 commands manage Claude Code plugins and standalone skills:

1. **`make v1 update-plugins-skills`** — Scans installed Claude Code plugins and standalone skills, writes `PLUGINS_AND_SKILLS.md`, and flags redundant standalone skills already covered by plugins.
2. **`make v1 sync`** — Reads `PLUGINS_AND_SKILLS.md` and installs all listed plugins and skills. Also removes standalone skills flagged as redundant.
3. **`make v1 analyze`** — Read-only terminal report showing installed plugins, skills, and any duplicates. Use for verification without modifying anything.

### Workflow

1. **Install locally** — Add a skill or plugin on your machine (e.g., `/skill add`, `/plugin install`, or `npx skills add`)
2. **Capture** — Run `make v1 update-plugins-skills` to record it in `PLUGINS_AND_SKILLS.md`
3. **Verify** — Run `make v1 analyze` to confirm no duplicates or conflicts
4. **Sync other devices** — On any other machine, run `make v1 sync` to install the same plugins and skills from `PLUGINS_AND_SKILLS.md`

## Manifest (`manifest.json`) - for now unused index file

Auto-generated index of all skills and agents. Each entry lists:

- **name** — identifier (e.g., `api-design`, `go/coder`)
- **scope** — `universal` or `language-specific`
- **languages** — applicable languages (for language-specific entries)
- **uses_skills** — resolved dependencies (transitive — includes indirect dependencies)
- **path** — file path relative to the repo root

Agent entries additionally include: **role**, **access**, **uses_plugins**, **delegates_to**.

Regenerate with `make v1 manifest`.

## Bootstrap Output (`.build/`)

### Claude Code

`make bootstrap` and pick Claude Code, generates platform-specific output in `.build/claude/` (adjustable):

- `CLAUDE.md` — orchestrator with agent definitions, safety constraints, and plugin references
- `agents/`, `skills/` — resolved spec files ready for Claude Code consumption
- `settings.json` — Claude Code settings from `spec/claude/settings.json`

### Cursor

`make bootstrap` and pick Cursor, generates platform-specific output in `.build/cursor/` (adjustable):

- `AGENTS.md` — orchestrator with agent definitions, safety constraints, and plugin references
- `agents/`, `skills/` — resolved spec files ready for Cursor consumption

## File Format

### Skill

```yaml
---
name: api-designer
description: API style selection, REST resource design, versioning strategy, ...
scope: universal
languages: []
uses_skills: [api-design, security]
---

# API Designer
...
```

### Agent

```yaml
---
name: go/coder
description: "Senior Go engineer. Writes production-grade Go code following project patterns."
role: coder
scope: language-specific
languages: [go]
access: read-write
uses_skills: [go/coder]
uses_plugins: [code-simplifier, context7]
delegates_to: []
---
```

