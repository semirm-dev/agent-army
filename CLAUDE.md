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

Ports & Adapters. Three layers, strict dependency direction: adapters → core ← ports.

- **`internal/core/`** — Pure domain logic, no I/O, no external deps beyond stdlib. Packages: `types`, `catalog`, `manifest`, `detector`, `orchestrator`, `diff`, `doctor`.
- **`internal/port/`** — User-facing. CLI (Cobra) and TUI (Bubble Tea).
- **`internal/adapter/`** — External integration. Plugin installer (`claude plugin install`), skill installer (`npx skills add`), system reader (parses `installed_plugins.json` and `.skill-lock.json`), command runner (real + dry-run).

Dependencies: `cobra`, `bubbletea`, `bubbles`, `lipgloss`

## Key Data Paths

- **Bundled catalog**: `army/internal/core/catalog/catalog.json` (embedded via `go:embed`)
- **Updated catalog**: `~/.army/catalog.json` (fetched by `update` command, merged over bundled)
- **Manifest**: `~/.army/manifest.json` (user's plugin/skill selections)
- **Installed plugins**: `~/.claude/plugins/installed_plugins.json`
- **Installed skills**: `~/.agents/.skill-lock.json`
- **Skill directories**: `~/.agents/skills/<name>/`

## Commands

`setup`, `sync`, `add`, `remove`, `list`, `update`, `doctor`

Global flags: `--dry-run`, `--manifest <path>`, `--verbose`

### sync flags

- `--destination <user|project>` — Override destination for all actions
- `--yes` / `-y` — Skip confirmation prompt
- Reads from `/dev/tty` for interactive confirmation (works through `make`)

### setup TUI

- Saves/restores cursor positions when navigating between steps
- Confirm step supports inline manifest path editing via `d` key

## Conventions

- Go CLI uses standard `internal/` package layout — no exported API
- Core packages must not import adapter or port packages
- Adapters implement interfaces defined in `orchestrator.go` (`PluginInstaller`, `SkillInstaller`, `SystemReader`)
- `deps.go` in the CLI package wires everything together
- Manifest writes use atomic temp-file + rename pattern
- Plugins install in parallel (goroutines), skills install sequentially
- Skill removal is direct filesystem deletion (dir + symlink + lock entry) — `npx skills remove` doesn't work for plugin-provided skills
- All name comparisons are case-insensitive (`strings.EqualFold`)
- Commit messages follow Conventional Commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`

## Testing

57 tests across 5 core packages: `catalog`, `manifest`, `detector`, `diff`, `doctor`. Tests use `t.TempDir()` and don't touch real system state (except `doctor_test.go` which creates temp dirs under `~/.agents/skills/`).

## Development Workflow

1. Make changes in `army/`
2. Run `make build` to build the binary
3. Run `make test` to verify nothing broke
