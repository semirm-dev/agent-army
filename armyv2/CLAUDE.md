# CLAUDE.md — armyv2

## What This Is

Interactive CLI for managing Claude Code plugins and skills. Lives in `armyv2/` within the agent-army repo. Separate Go module from `army/`.

## Build & Test

```bash
make build-v2                                             # Build binary
make test-v2                                              # All tests with race detection
cd armyv2 && go test ./internal/core/detector/... -race   # Single package
```

## Architecture

Ports & Adapters. Three layers, strict dependency direction: adapters → core ← ports.

- **`internal/core/`** — Pure domain logic, no I/O, no external deps beyond stdlib. Packages: `types`, `catalog`, `manifest`, `detector`, `orchestrator`, `diff`, `doctor`.
- **`internal/port/`** — User-facing. CLI (Cobra) and TUI (Bubble Tea).
- **`internal/adapter/`** — External integration. Plugin installer (`claude plugin install`), skill installer (`npx skills add`), system reader (parses `installed_plugins.json` and `.skill-lock.json`), command runner (real + dry-run).

Entry point: `cmd/armyv2/main.go` → `cli.NewRootCmd()`

## Key Data Paths

- **Bundled catalog**: `internal/core/catalog/catalog.json` (embedded via `go:embed`)
- **Updated catalog**: `~/.armyv2/catalog.json` (fetched by `update` command, merged over bundled)
- **Manifest**: `~/.armyv2/manifest.json` (user's plugin/skill selections)
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

- Core packages must not import adapter or port packages
- Adapters implement interfaces defined in `orchestrator.go` (`PluginInstaller`, `SkillInstaller`, `SystemReader`)
- `deps.go` in the CLI package wires everything together
- Manifest writes use atomic temp-file + rename pattern
- Plugins install in parallel (goroutines), skills install sequentially
- Skill removal is direct filesystem deletion (dir + symlink + lock entry) — `npx skills remove` doesn't work for plugin-provided skills
- All name comparisons are case-insensitive (`strings.EqualFold`)

## Testing

55 tests across 5 core packages: `catalog`, `manifest`, `detector`, `diff`, `doctor`. Tests use `t.TempDir()` and don't touch real system state (except `doctor_test.go` which creates temp dirs under `~/.agents/skills/`).
