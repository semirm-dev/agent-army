# army

Interactive CLI for setting up and managing Claude Code plugins and skills. Maintains a personal manifest of desired plugins/skills, auto-detects your tech stack, and keeps everything in sync.

## Quick Start

```bash
make build               # Build the binary
make export              # Add army to PATH (uses .build/army by default)
make export DEST=~/bin   # Copy binary to custom dir + add to PATH
army setup               # Launch interactive setup wizard
army list                # See what's in your manifest
army sync                # Install everything from your manifest
```

## Commands

| Command | Description |
|---------|-------------|
| `setup` | Interactive TUI wizard — pick destination, detect tech stack, select plugins & skills. Fixed manifest paths: user → `~/.army/manifest.json`, project → `<cwd>/.army/manifest.json` |
| `sync` | Install missing + remove extras to match manifest. Shows plan and asks for confirmation. **Project-level manifests skip orphan removal** — only installs missing items |
| `add` | Add a plugin or skill to manifest (`add plugin context7`, `add skill golang-pro`) |
| `remove` | Remove a plugin or skill from manifest (`remove plugin context7`) |
| `clear` | Uninstall plugins and skills from the system |
| `list` | Show manifest items with install status (`✓` ok, `⚠` broken on disk, `✗` missing) |
| `detect` | Show loaded config files for the current directory |
| `fetch-catalog` | Fetch latest catalog from GitHub into `~/.army/catalog.json` |
| `doctor` | Run health checks — missing items, orphans, disk drift. **Project-level manifests skip orphan warnings** |
| `catalog` | Show catalog summary (plugin/skill/profile counts) |
| `serve` | Start the web management UI at `http://localhost:3141` |
| `version` | Print the army version |

### Global Flags

| Flag | Description |
|------|-------------|
| `--dry-run` | Print commands without executing |
| `--verbose` | Verbose output |
| `--json` | Output structured JSON (for scripting and web UI) |

### Add Flags

| Flag | Description |
|------|-------------|
| `--no-install` | Add to manifest without installing |

### Remove Flags

| Flag | Description |
|------|-------------|
| `--manifest-only` | Remove from manifest without uninstalling |

### Sync Flags

| Flag | Description |
|------|-------------|
| `--destination <user\|project>` | Override destination for all actions |
| `--yes` / `-y` | Skip confirmation prompt |

### Serve Flags

| Flag | Description |
|------|-------------|
| `--port <int>` | Port to serve on (default: 3141) |
| `--no-open` | Don't open browser automatically |

## How It Works

1. **Catalog** — Bundled JSON with all known plugins, skills, and tech profiles. Updated via `army fetch-catalog`.
2. **Manifest** — Tracks your selected plugins and skills. Paths: `~/.army/manifest.json` (user-level) or `<cwd>/.army/manifest.json` (project-level). Resolution walks up from cwd looking for `.army/manifest.json` (like `.git` discovery), falls back to `~/.army/manifest.json`.
3. **Tech detection** — Scans project directory for markers (go.mod, package.json deps, tsconfig.json, etc.) and recommends relevant plugins/skills.
4. **Sync** — Compares manifest against installed state, installs missing items, removes extras (user-level manifests only — project-level manifests skip orphan removal since they describe a subset of the system).

## Web UI

A browser-based management console that covers all CLI operations.

```bash
make web-install           # Install frontend + backend dependencies
make web-build             # Build for production
army serve                 # Start web UI at http://localhost:3141
```

**Pages:** Catalog browser, Manifest manager, Sync with live progress, Doctor dashboard.

**Architecture:** React SPA (Vite + shadcn/ui + TanStack Query) → NestJS API (pure HTTP-to-CLI shell) → army CLI (`--json` output). Zero logic duplication — all domain logic stays in Go.

**Dev mode:** Run three processes: `make build`, NestJS (`cd army/web/be && ARMY_BIN=../../.build/army PORT=3141 npm run start:dev`), Vite (`cd army/web/fe && VITE_API_URL=http://localhost:3141/api npm run dev`).

## Architecture

```
army/
├── cmd/army/main.go       # Entry point → cli.NewRootCmd()
├── cli/                   # CLI commands (Cobra) — app's entry surface
├── web/
│   ├── be/                # NestJS backend (API shell, SSE sync streaming)
│   └── fe/                # React frontend (Vite + shadcn/ui + Tailwind)
└── internal/
    ├── core/              # Pure domain logic (no I/O)
    │   ├── types/         # Shared data structures
    │   ├── catalog/       # Catalog loading, merging, embedded JSON
    │   ├── manifest/      # Manifest CRUD with atomic writes, walk-up resolution
    │   ├── detector/      # Tech stack detection from project files
    │   ├── orchestrator/  # Action planning and execution coordination
    │   ├── diff/          # Manifest vs installed comparison
    │   └── doctor/        # Health checks
    ├── installer/         # Plugin + skill install/remove operations
    ├── runner/            # Command execution (real + dry-run)
    ├── state/             # Reads installed_plugins.json, .skill-lock.json
    └── tui/               # Bubble Tea setup wizard
```

## Testing

```bash
make test                                              # All Go tests
cd army && go test ./internal/core/catalog/... -race    # Single package
cd army/web/be && npm run build                         # Verify NestJS compiles
cd army/web/fe && npm run build                         # Verify React compiles
```
