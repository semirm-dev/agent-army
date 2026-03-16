# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Agent Army (armyv2) is an interactive CLI for managing Claude Code plugins and skills. It provides a TUI wizard for selecting plugins/skills, syncs installations to match a manifest, and runs health checks to detect drift.

## Build & Test Commands

```bash
make build-v2           # Build armyv2 CLI binary (armyv2/armyv2)
make test-v2            # Run armyv2 tests with race detection
make v2 setup           # Interactive TUI wizard — select plugins/skills, save manifest
make v2 sync            # Install missing + remove extras to match manifest (with confirmation)
make v2 list            # Show manifest items with install status (✓ ok, ⚠ broken, ✗ missing)
make v2 doctor          # Run health checks — missing, orphan, and disk drift detection
make v2 update          # Fetch latest catalog from GitHub into ~/.armyv2/catalog.json
make v2 add             # Add a plugin or skill to manifest (e.g. make v2 add plugin context7)
make v2 remove          # Remove a plugin or skill from manifest (e.g. make v2 remove skill golang-pro)
```

Run a single test package:
```bash
cd armyv2 && go test ./internal/core/detector/... -race -count=1
```

## Architecture

### Go CLI (`armyv2/`)

Entry point: `armyv2/cmd/armyv2/main.go` → `cli.NewRootCmd()`

Ports & Adapters architecture:
- **`internal/core/`** — Pure domain logic: catalog, manifest, detector, orchestrator, diff, doctor, types
- **`internal/port/`** — Presentation: TUI (Bubble Tea) + CLI (Cobra)
- **`internal/adapter/`** — System integration: plugin installer, skill installer, system reader, command runner

Commands: `setup`, `sync`, `add`, `remove`, `list`, `update`, `doctor`

Dependencies: `cobra`, `bubbletea`, `bubbles`, `lipgloss`

### Key Files

- **`Makefile`** — All build orchestration
- **`armyv2/internal/core/catalog/catalog.json`** — Bundled catalog (embedded via `go:embed`)
- **`~/.armyv2/catalog.json`** — Updated catalog (fetched by `update` command)
- **`~/.armyv2/manifest.json`** — User's plugin/skill selections

## Development Workflow

1. Make changes in `armyv2/`
2. Run `make build-v2` to build the binary
3. Run `make test-v2` to verify nothing broke

## Conventions

- Go CLI uses standard `internal/` package layout — no exported API
- Core packages must not import adapter or port packages
- Commit messages follow Conventional Commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`
