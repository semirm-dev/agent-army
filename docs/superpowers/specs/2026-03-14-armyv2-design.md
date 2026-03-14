# armyv2 — Claude Code Setup & Plugin/Skill Manager

## Problem

Setting up Claude Code on a new machine or project requires manually discovering, selecting, and installing plugins and skills. There's no single tool that manages the full lifecycle: initial setup, ongoing sync, drift detection, and health checks. The existing `army` CLI handles spec-based bootstrapping but not interactive device setup.

## Solution

A new Go CLI (`armyv2`) in the agent-army repo that provides:
- Interactive TUI wizard for initial setup (Bubble Tea)
- Personal manifest file tracking desired plugins/skills
- Bundled catalog of all available plugins/skills with tech-stack mappings
- Full lifecycle management: setup, sync, add, remove, list, diff, update, doctor
- Layered architecture (ports & adapters) enabling a future web UI

## Architecture

### Ports & Adapters (Layered)

```
armyv2/
  cmd/armyv2/main.go                 # Entry point, wires everything together

  internal/
    core/                             # Pure domain logic, no I/O
      catalog/                        # Registry of all known plugins & skills + tech mappings
        catalog.go                    # Load/merge catalog from embedded + updated sources
        embed.go                      # go:embed for bundled catalog.json
      manifest/                       # User profile CRUD
        manifest.go                   # Load, save, add, remove, merge operations
      detector/                       # Tech stack detection
        detector.go                   # Scan project files → detected languages/frameworks
      orchestrator/                   # Install/remove/sync coordination
        orchestrator.go               # Reconcile manifest vs installed state, produce action plan
      diff/                           # Compare manifest vs installed state
        diff.go                       # Produce structured diff result
      doctor/                         # Health checks
        doctor.go                     # Broken installs, orphans, lock drift, missing deps

    port/                             # Presentation layer (swappable)
      tui/                            # Bubble Tea models & views
        setup.go                      # Setup wizard (destination → tech → plugins → skills → confirm)
        list.go                       # List view with install status
        diff.go                       # Diff view
        components/                   # Reusable TUI components (multi-select, filter, progress)
      cli/                            # Cobra command definitions
        root.go                       # Root command + global flags
        setup.go                      # armyv2 setup
        sync.go                       # armyv2 sync
        add.go                        # armyv2 add plugin|skill <name>
        remove.go                     # armyv2 remove plugin|skill <name>
        list.go                       # armyv2 list
        diff.go                       # armyv2 diff
        update.go                     # armyv2 update
        doctor.go                     # armyv2 doctor

    adapter/                          # System integration (swappable)
      plugin/                         # Claude plugin install/remove
        plugin.go                     # Wraps "claude plugin install/remove" commands
      skill/                          # npx skill install/remove
        skill.go                      # Wraps "npx skills add/remove" commands
      system/                         # Read installed state
        system.go                     # Reads ~/.claude/plugins/installed_plugins.json
                                      # Reads ~/.agents/.skill-lock.json
      runner/                         # Command execution abstraction
        runner.go                     # Interface: Run(cmd string) (stdout, error)
        real.go                       # Executes actual shell commands
        dryrun.go                     # Prints commands without executing
```

### Data Flow

```
TUI / CLI (port)
    ↓ calls
Core Domain (catalog + manifest + orchestrator)
    ↓ uses interfaces
Adapters (plugin installer, skill installer, system reader, runner)
    ↓ executes
System (~/.claude/, ~/.agents/, shell commands)
```

The core never imports port or adapter packages. It defines interfaces that adapters implement:

```go
// core/orchestrator/interfaces.go
type PluginInstaller interface {
    // Install runs: claude plugin install <name>
    // The marketplace field is metadata only — the claude CLI resolves it.
    Install(name string) error
    // Remove runs: claude plugin remove <name>
    Remove(name string) error
}

type SkillInstaller interface {
    // Install runs: npx @anthropic-ai/claude-code-skills add <name> -s <source> -y
    Install(name, source string) error
    // Remove does direct filesystem removal (delete ~/.agents/skills/<name>/
    // and remove entry from ~/.agents/.skill-lock.json) because npx skills remove
    // refuses to remove plugin-provided skills. This matches the existing army approach.
    Remove(name string) error
}

type SystemReader interface {
    InstalledPlugins() ([]InstalledPlugin, error)
    InstalledSkills() ([]InstalledSkill, error)
}

// CommandRunner abstracts shell execution. The real implementation streams
// stdout/stderr to the terminal for user feedback during installs.
// The dry-run implementation prints the command without executing.
// Return value captures stdout for commands that need it (e.g., version checks).
type CommandRunner interface {
    Run(cmd string, args ...string) (stdout string, err error)
}
```

## Data Formats

### Catalog (`catalog.json` — embedded in binary, updatable from GitHub)

```json
{
  "version": 1,
  "updated_at": "2026-03-14",
  "plugins": [
    {
      "name": "context7",
      "marketplace": "claude-plugins-official",
      "description": "Up-to-date docs and code examples for any library",
      "tags": ["docs", "universal"]
    }
  ],
  "skills": [
    {
      "name": "golang-pro",
      "source": "jeffallan/claude-skills",
      "description": "Go patterns, concurrency, gRPC, microservices",
      "tags": ["go"]
    }
  ],
  "tech_profiles": {
    "go": {
      "detect": ["go.mod", "go.sum"],
      "plugins": ["gopls-lsp"],
      "skills": ["golang-pro"]
    },
    "react": {
      "detect": ["package.json:react", "*.tsx"],
      "plugins": ["frontend-design"],
      "skills": ["react-expert", "javascript-pro"]
    },
    "python": {
      "detect": ["requirements.txt", "pyproject.toml", "*.py"],
      "plugins": [],
      "skills": []
    },
    "typescript": {
      "detect": ["tsconfig.json", "*.ts"],
      "plugins": [],
      "skills": ["javascript-pro"]
    },
    "php": {
      "detect": ["composer.json", "*.php"],
      "plugins": [],
      "skills": ["php-pro"]
    },
    "laravel": {
      "detect": ["artisan", "composer.json:laravel"],
      "plugins": [],
      "skills": ["laravel-specialist", "php-pro"]
    },
    "vue": {
      "detect": ["package.json:vue", "*.vue"],
      "plugins": ["frontend-design"],
      "skills": ["vue-expert", "javascript-pro"]
    },
    "nextjs": {
      "detect": ["package.json:next", "next.config.*"],
      "plugins": ["frontend-design"],
      "skills": ["nextjs-developer", "react-expert", "javascript-pro"]
    },
    "nestjs": {
      "detect": ["package.json:@nestjs/core"],
      "plugins": [],
      "skills": ["nestjs-expert"]
    },
    "postgres": {
      "detect": ["*.sql", "docker-compose.*:postgres"],
      "plugins": [],
      "skills": ["postgres-pro", "database-optimizer", "database-schema-designer"]
    }
  }
}
```

The `detect` field supports two forms:
- `"go.mod"` — file existence check (supports globs: `"*.tsx"`)
- `"package.json:react"` — file existence + JSON dependency key check. For JSON files, the detector parses the file and checks if the string appears as a key in `dependencies` or `devDependencies` (for package.json) or as a key in the top-level require map (for composer.json). For non-JSON files (e.g., `docker-compose.*:postgres`), falls back to substring match. This avoids false positives like matching `react-native` when looking for `react`.

### Manifest (`~/.armyv2/manifest.json` — user's personal selections)

```json
{
  "version": 1,
  "plugins": [
    {
      "name": "context7",
      "marketplace": "claude-plugins-official",
      "tags": ["docs", "universal"],
      "destination": "user"
    }
  ],
  "skills": [
    {
      "name": "golang-pro",
      "source": "jeffallan/claude-skills",
      "tags": ["go"],
      "destination": "user"
    }
  ]
}
```

- `destination`: `"user"` or `"project"`. This controls install behavior only:
  - `"user"`: `claude plugin install <name>` (global), `npx skills add <name> -s <source> -y` (global)
  - `"project"`: same commands but run from the project directory (skills install to project-level `.claude/skills/`)
  - The manifest itself is always personal at `~/.armyv2/manifest.json` — destination is a per-item install scope hint
- `tags` are copied from catalog for filtering in the TUI

### Catalog Update Source

`armyv2 update` fetches the latest `catalog.json` from `raw.githubusercontent.com/<owner>/agent-army/main/armyv2/catalog.json` and saves it to `~/.armyv2/catalog.json`. The embedded catalog is the fallback when no updated version exists.

**Validation:** Before writing a fetched catalog to disk, `update` validates:
1. JSON parses successfully
2. `version` field is present and >= current version
3. `plugins` and `skills` arrays exist
4. Each entry has required fields (`name`, and `marketplace`/`source` respectively)

If validation fails, the fetched catalog is rejected and the existing one is preserved.

## Commands

| Command | Purpose | Interactive |
|---------|---------|:-----------:|
| `armyv2 setup` | Interactive wizard: destination → tech detection → plugin selection → skill selection → confirm & install | TUI |
| `armyv2 sync` | Apply manifest to machine — install missing, remove extras. Idempotent. | No |
| `armyv2 add plugin <name>` | Add plugin to manifest + install (unless `--no-install`) | No |
| `armyv2 add skill <name>` | Add skill to manifest + install (unless `--no-install`) | No |
| `armyv2 remove plugin <name>` | Remove from manifest + uninstall (unless `--manifest-only`) | No |
| `armyv2 remove skill <name>` | Remove from manifest + uninstall (unless `--manifest-only`) | No |
| `armyv2 list` | Show manifest contents with install status (✓ installed, ✗ missing) | No |
| `armyv2 diff` | Compare manifest vs installed state, exit code 1 if drift | No |
| `armyv2 update` | Fetch latest catalog from GitHub | No |
| `armyv2 doctor` | Health check: broken installs, orphans, lock drift, missing deps | No |

### Global Flags

- `--dry-run` — print commands instead of executing
- `--manifest <path>` — override default manifest location (`~/.armyv2/manifest.json`)
- `--verbose` — detailed output

## Setup Wizard Flow

1. **Destination**: user-level (global) or project-level (current directory)
2. **Tech stack** (project only): auto-detect from project files (go.mod, package.json, etc.), show as pre-checked multi-select list, user can add/remove
3. **Plugins**: multi-select from catalog, tech-recommended items pre-checked and starred (★)
4. **Skills**: same as plugins, filterable with `/` key
5. **Confirm & install**: summary of selections, proceed/cancel, parallel progress with spinners

## Key Behaviors

- **Parallel plugin installs**: plugins install concurrently (goroutines + WaitGroup), skills sequentially (npx limitation)
- **Idempotent sync**: `sync` compares manifest vs installed state, only installs missing and removes extras
- **Partial failure handling**: on install/sync, continue installing remaining items on failure. Report all failures at the end with a summary (N succeeded, M failed). Manifest is not modified on failure — it represents desired state, not actual state. `diff` shows the gap.
- **Doctor checks**: installed_plugins.json vs manifest, .skill-lock.json vs manifest, orphaned skills (installed but not in manifest), lock file drift (in lock but missing from disk)
- **Catalog merge**: when `update` fetches a newer catalog, new entries are merged; removed entries are flagged but not auto-deleted from manifest
- **Add unknown items**: `armyv2 add` requires the item to exist in the catalog. If not found, error with suggestion to run `armyv2 update` first. No custom entries outside the catalog.
- **Add destination**: `armyv2 add` defaults to `destination: "user"`. Use `--project` flag to set `destination: "project"`.
- **Tech profile deduplication**: when multiple profiles are detected (e.g., react + nextjs both recommend `react-expert`), the orchestrator deduplicates before presenting to the user or producing install actions.
- **Output convention**: errors to stderr, progress to stdout. `diff` and `doctor` exit with code 1 if issues found (useful for CI/scripts).

## Tech Stack Detection

The detector scans the current directory for marker files defined in `tech_profiles[].detect`:

1. Glob for file existence (e.g., `go.mod`, `*.tsx`)
2. For `file:content` patterns, check if file exists AND contains the string (e.g., `package.json:react` → does package.json contain "react"?)
3. Return list of matched tech profile keys
4. TUI pre-checks matched items, user refines

## Dependencies

- **cobra** — CLI framework (same as army)
- **bubbletea** — TUI framework
- **bubbles** — TUI components (list, textinput, spinner, progress)
- **lipgloss** — TUI styling
- **gopkg.in/yaml.v3** — not needed (JSON only for armyv2)

## File Locations

| What | Where |
|------|-------|
| Manifest | `~/.armyv2/manifest.json` |
| Updated catalog | `~/.armyv2/catalog.json` (overrides embedded) |
| Bundled catalog | `armyv2/catalog.json` (go:embed) |
| Source code | `armyv2/` (parallel to `army/`) |
| Binary | `armyv2/armyv2` |

## Non-Goals

- No backward compatibility with PLUGINS_AND_SKILLS.md (deprecated)
- No project-level manifests (personal manifest only)
- No web UI in v1 (architecture supports it, not building it yet)
- No auto-update of the armyv2 binary itself
