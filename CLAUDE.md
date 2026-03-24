# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Agent Army (army) is an interactive CLI for managing Claude Code plugins and skills. It provides a TUI wizard for selecting plugins/skills, syncs installations to match a manifest, and runs health checks to detect drift.

## Build & Test

```bash
make build                                             # Build binary (army/army)
make export                                            # Add army to PATH (default: .build/army)
make export DEST=~/bin                                 # Copy binary to custom dir + add to PATH
make test                                              # All tests with race detection
cd army && go test ./internal/core/detector/... -race   # Single package
```

## Architecture

Entry point: `army/cmd/army/main.go` → `cli.NewRootCmd()`

Layered structure with strict dependency direction: outer packages → core.

- **`internal/core/`** — Pure domain logic, no I/O, no external deps beyond stdlib. Packages: `types`, `catalog`, `manifest`, `detector`, `orchestrator`, `diff`, `doctor`.
- **`internal/cli/`** — CLI commands (Cobra). **`internal/tui/`** — Setup wizard (Bubble Tea).
- **`internal/installer/`** — Plugin and skill install/remove operations (`claude plugin install`, `npx skills add`). **`internal/runner/`** — Command execution (`RealRunner` + `DryRunner`). **`internal/state/`** — Reads installed state from filesystem (`installed_plugins.json`, `.skill-lock.json`).

Dependencies: `cobra`, `bubbletea`, `bubbles`, `lipgloss`

## Key Data Paths

- **Bundled catalog**: `army/internal/core/catalog/catalog.json` (embedded via `go:embed`)
- **Updated catalog**: `~/.army/catalog.json` (fetched by `fetch-catalog` command, merged over bundled)
- **Manifest**: `~/.army/manifest.json` (user-level) or `<cwd>/.army/manifest.json` (project-level)
  - **Resolution**: walks up from cwd looking for `.army/manifest.json` (like `.git` discovery), falls back to `~/.army/manifest.json`
- **Installed plugins**: `~/.claude/plugins/installed_plugins.json`
- **Installed skills**: `~/.agents/.skill-lock.json`
- **Skill directories**: `~/.agents/skills/<name>/`

## Commands

`setup`, `sync`, `add`, `remove`, `list`, `fetch-catalog`, `doctor`

Global flags: `--dry-run`, `--verbose`

### Orphan handling

Orphans (installed items not in manifest) are only relevant for user-level manifests (`~/.army/manifest.json`). When using a project-level manifest, `sync` skips removal actions and `doctor` skips orphan warnings — a project manifest describes what one project needs, not the full system state.

### sync flags

- `--destination <user|project>` — Override destination for all actions
- `--yes` / `-y` — Skip confirmation prompt
- Reads from `/dev/tty` for interactive confirmation

### setup TUI

- Saves/restores cursor positions when navigating between steps
- Destination choice sets manifest path: user → `~/.army/manifest.json`, project → `<cwd>/.army/manifest.json` (fixed, not editable)

## Conventions

- Go CLI uses standard `internal/` package layout — no exported API
- Core packages must not import outer packages (installer, runner, state, cli, tui)
- `installer/`, `runner/`, `state/` implement interfaces defined in `orchestrator.go` (`PluginInstaller`, `SkillInstaller`, `SystemReader`)
- Interfaces are defined at the consumer side (Go idiom) — `installer/` defines its own `CommandRunner` interface rather than importing one from `runner/`
- `deps.go` in the CLI package wires everything together
- Manifest writes use atomic temp-file + rename pattern
- Plugins install in parallel (goroutines), skills install sequentially
- Skill removal is direct filesystem deletion (dir + symlink + lock entry) — `npx skills remove` doesn't work for plugin-provided skills
- All name comparisons are case-insensitive (`strings.EqualFold`)
- Commit messages follow Conventional Commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`
- **Versioning**: Always bump `VERSION` file (semver) with every code change. Patch (`0.1.2` → `0.1.3`) for `fix:`, `refactor:`, `test:`, `chore:`, `docs:`. Minor (`0.1.2` → `0.2.0`) for `feat:`. Major (`0.1.2` → `1.0.0`) for breaking changes.

## Testing

Tests across 5 core packages: `catalog`, `manifest`, `detector`, `diff`, `doctor`. Tests use `t.TempDir()` and don't touch real system state (except `doctor_test.go` which creates temp dirs under `~/.agents/skills/`).

## Development Workflow

1. Make changes in `army/`
2. Run `make build` to build the binary
3. Run `make test` to verify nothing broke
