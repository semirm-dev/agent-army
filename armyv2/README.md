# armyv2

Interactive CLI for setting up and managing Claude Code plugins and skills. Maintains a personal manifest of desired plugins/skills, auto-detects your tech stack, and keeps everything in sync.

## Quick Start

```bash
make build-v2          # Build the binary
make v2 setup          # Launch interactive setup wizard
make v2 list           # See what's in your manifest
make v2 sync           # Install everything from your manifest
```

## Commands

| Command | Description |
|---------|-------------|
| `setup` | Interactive TUI wizard — pick destination, detect tech stack, select plugins & skills. Supports inline path editing (`d` key on confirm step) and cursor persistence across steps |
| `sync` | Install missing + remove extras to match manifest. Shows plan and asks for confirmation. Supports interactive destination editing |
| `add` | Add a plugin or skill to manifest (`add plugin context7`, `add skill golang-pro`) |
| `remove` | Remove a plugin or skill from manifest (`remove plugin context7`) |
| `list` | Show manifest items with install status (`✓` ok, `⚠` broken on disk, `✗` missing) |
| `update` | Fetch latest catalog from GitHub into `~/.armyv2/catalog.json` |
| `doctor` | Run health checks — missing items, orphans, disk drift (skill dirs + plugin installPaths) |

### Global Flags

| Flag | Description |
|------|-------------|
| `--dry-run` | Print commands without executing |
| `--manifest <path>` | Override manifest path (default: `~/.armyv2/manifest.json`) |
| `--verbose` | Verbose output |

### Sync Flags

| Flag | Description |
|------|-------------|
| `--destination <user\|project>` | Override destination for all actions |
| `--yes` / `-y` | Skip confirmation prompt |

### Using via Make

Simple commands work through make:

```bash
make v2 setup
make v2 list
make v2 doctor
```

For commands with flags, use the binary directly:

```bash
./armyv2/armyv2 add plugin context7 --no-install
./armyv2/armyv2 sync --dry-run
./armyv2/armyv2 sync --destination user --yes
./armyv2/armyv2 remove skill golang-pro --manifest-only
```

## How It Works

1. **Catalog** — Bundled JSON with all known plugins, skills, and tech profiles. Updated via `armyv2 update`.
2. **Manifest** — Personal file at `~/.armyv2/manifest.json` tracking your selected plugins and skills.
3. **Tech detection** — Scans project directory for markers (go.mod, package.json deps, tsconfig.json, etc.) and recommends relevant plugins/skills.
4. **Sync** — Compares manifest against installed state, installs missing items, optionally removes extras.

## Architecture

Ports & Adapters (hexagonal):

```
internal/
├── core/              # Pure domain logic (no I/O)
│   ├── types/         # Shared data structures
│   ├── catalog/       # Catalog loading, merging, embedded JSON
│   ├── manifest/      # Manifest CRUD with atomic writes
│   ├── detector/      # Tech stack detection from project files
│   ├── orchestrator/  # Action planning and execution coordination
│   ├── diff/          # Manifest vs installed comparison
│   └── doctor/        # Health checks
├── port/              # User-facing interfaces
│   ├── cli/           # Cobra commands
│   └── tui/           # Bubble Tea setup wizard
└── adapter/           # External system integration
    ├── runner/         # Command execution (real + dry-run)
    ├── plugin/         # claude plugin install/remove
    ├── skill/          # npx skills add + direct filesystem removal
    └── system/         # Reads installed_plugins.json, .skill-lock.json
```

## Testing

```bash
make test-v2                                              # All tests
cd armyv2 && go test ./internal/core/catalog/... -race    # Single package
```
