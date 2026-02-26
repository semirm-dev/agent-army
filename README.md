# agent-rules

Portable AI development setup for **Claude Code** and **Cursor**. One repo, one `bootstrap.sh`, consistent rules across devices.

## Quick Start

```bash
git clone <this-repo> ~/workspace/agent-rules
cd ~/workspace/agent-rules
make bootstrap
```

Bootstrap configures your machine for the agent-rules workflow — it installs skills, deploys settings (which enables plugins), and sets up shell aliases. Every step lists what will be installed and asks for confirmation.

| Step | Action                                                                |
| ---- | --------------------------------------------------------------------- |
| 1    | Check prerequisites (node, npx, claude CLI, rsync)                    |
| 2    | Sync rules to `~/.claude/` and `~/.cursor/rules/`                     |
| 3    | Install 5 npm skills (golang-pro, browser-use, etc.)                  |
| 4    | Deploy `settings.json` + enable 6 plugins (shows diff if file exists) |
| 5    | Add `sync-rules` and `check-sync` aliases to `~/.zshrc`               |
| 6    | Verify installation (list skills, agents, run check-sync)             |

Idempotent — skips already-installed components on re-run.

## Directory Structure

```
agent-rules/
├── claude/
│   ├── CLAUDE.md              # Main instructions (safety, patterns, planning)
│   ├── settings.json          # Reference settings template
│   ├── SKILLS.md              # What to install (plugins vs npm skills)
│   ├── statusline-command.sh  # Status line script (deployed to ~/.claude/)
│   ├── agents/                # 18 reusable agent prompts
│   │   ├── go-coder.md        # Go code writer (uses golang-pro skill)
│   │   ├── go-reviewer.md     # Go code reviewer (read-only)
│   │   ├── go-tester.md       # Go test writer
│   │   ├── ts-coder.md        # TypeScript/JS code writer
│   │   ├── ts-reviewer.md     # TypeScript/JS code reviewer (read-only)
│   │   ├── ts-tester.md       # TypeScript/JS test writer
│   │   ├── py-coder.md        # Python code writer
│   │   ├── py-reviewer.md     # Python code reviewer (read-only)
│   │   ├── py-tester.md       # Python test writer
│   │   ├── react-coder.md     # React/frontend component writer
│   │   ├── react-reviewer.md  # React/frontend code reviewer (read-only)
│   │   ├── react-tester.md    # React component test writer
│   │   ├── db-coder.md        # Database engineer (migrations, queries)
│   │   ├── db-reviewer.md     # Database reviewer (read-only)
│   │   ├── db-tester.md       # Database test writer
│   │   ├── docker-builder.md  # Dockerfile, compose, CI/CD writer
│   │   ├── docker-reviewer.md # Docker/infra reviewer (read-only)
│   │   └── arch-reviewer.md   # Architecture reviewer (read-only)
│   └── rules/                 # 12 domain-specific rule files
│       ├── go-patterns.md     # Go coding + testing standards
│       ├── ts-patterns.md     # TypeScript coding + testing standards
│       ├── py-patterns.md     # Python coding + testing standards
│       ├── react-patterns.md  # React component, accessibility, error boundaries
│       ├── git-workflow.md    # Git conventions (branch, commit, PR)
│       ├── api-design.md      # REST/gRPC API patterns
│       ├── database.md        # Migrations, pooling, transactions, ORMs
│       ├── observability.md   # Logging, metrics, health checks, Docker, CI/CD
│       ├── security.md        # Auth, CORS, rate limiting, secrets
│       ├── cross-cutting.md   # Error taxonomy, coverage targets, deps
│       ├── concurrency.md     # Concurrency (goroutines, promises, asyncio)
│       └── testing-patterns.md # Testing patterns (naming, fixtures, CI)
├── cursor/                    # 14 Cursor IDE rules
│   ├── 000-index.mdc          # Safety & communication (alwaysApply)
│   ├── 100-golang.mdc         # Go coding patterns (globs: **/*.go)
│   ├── 101-typescript.mdc     # TypeScript patterns (globs: **/*.ts,tsx,js,jsx)
│   ├── 102-python.mdc         # Python patterns (globs: **/*.py)
│   ├── 103-react.mdc          # React patterns (globs: **/*.tsx,jsx)
│   ├── 200-planning.mdc       # Planning template (alwaysApply)
│   ├── 300-git.mdc            # Git workflow conventions (alwaysApply)
│   ├── 400-api-design.mdc     # API design patterns
│   ├── 401-database.mdc       # Database patterns
│   ├── 500-observability.mdc  # Observability & infrastructure
│   ├── 501-security.mdc       # Security patterns
│   ├── 502-cross-cutting.mdc  # Error taxonomy, coverage, deps (alwaysApply)
│   ├── 503-concurrency.mdc    # Concurrency patterns
│   └── 504-testing.mdc        # Testing patterns
├── skills/                    # 7 custom skills
│   ├── api-designer.md        # API design checklist and scaffolding
│   ├── dependency-audit.md    # Dependency audit and update workflow
│   ├── git-conventions.md     # Branch naming, commit format, PR templates
│   ├── migration-safety.md    # Database migration safety checklist
│   ├── error-handling.md      # Error taxonomy and propagation patterns
│   ├── code-architecture.md   # Architecture decisions and DI patterns
│   └── testing-strategy.md    # Test pyramid and strategy guidance
├── scripts/
│   ├── bootstrap.sh           # Interactive new-device setup
│   ├── rsync-rules.sh         # Sync repo → ~/.claude/ or ~/.cursor/rules/
│   ├── check-sync.sh          # Verify CLAUDE.md ↔ Cursor .mdc parity
│   ├── validate-structure.sh  # Structural validation (agents, rules, triads)
│   ├── verify-deployed.sh     # Verify deployed state matches repo
│   └── test-check-sync.sh     # Tests for check-sync drift detection
├── .githooks/
│   └── pre-commit             # Sync check before commit
├── templates/
│   └── PROJECT-CLAUDE.md      # Project-level CLAUDE.md starter template
└── README.md
```

## What Gets Deployed Where

```
Repo                            → Deployed to
─────────────────────────────────────────────────────────
claude/CLAUDE.md                → ~/.claude/CLAUDE.md
claude/agents/*.md              → ~/.claude/agents/*.md
claude/rules/*.md               → ~/.claude/rules/*.md
claude/statusline-command.sh    → ~/.claude/statusline-command.sh
claude/settings.json            → ~/.claude/settings.json (bootstrap only)
cursor/*.mdc                    → ~/.cursor/rules/*.mdc
```

Excluded from sync (user-managed): `~/.claude/settings.json`, `skills/`, `plugins/`, `projects/`, `todos/`.

## Commands

| Command                | Purpose                                        |
|------------------------|------------------------------------------------|
| `make`                 | Show all available targets                     |
| `make bootstrap`       | First-time interactive setup                   |
| `make sync`            | Sync rules to Claude and Cursor                |
| `make sync-claude`     | Sync rules to Claude only                      |
| `make sync-cursor`     | Sync rules to Cursor only                      |
| `make check`           | Verify nothing drifted                         |
| `make deploy`          | Sync + check (day-to-day loop)                 |
| `make validate`        | Structural validation (agents, rules, triads)  |
| `make verify-deployed` | Verify deployed state matches repo             |
| `make test`            | Run test suite                                 |
| `make install-hooks`   | Install git pre-commit hook                    |
| `make init-project`    | Scaffold a project-level CLAUDE.md             |
| `make watch`           | Watch for changes and auto-sync                |

`make check` exit codes:

| Exit code | Meaning                             |
| --------- | ----------------------------------- |
| `0`       | All sections in sync                |
| `1`       | Drift detected — shows unified diff |

Sections checked: Go, TypeScript, Python, React, Git Workflow, Safety, Communication, Planning, API Design, Database, Observability, Security, Cross-Cutting, Concurrency, Testing Patterns. Structural validation (`make validate`) checks agent triads, rule references, and sync pairs.

## Capabilities

| Source             | What                                      | Count | Items                                                                                   |
| ------------------ | ----------------------------------------- | ----- | --------------------------------------------------------------------------------------- |
| **Plugins**        | Auto-updating, managed by Claude CLI      | 6     | superpowers, context7, frontend-design, code-review, security-guidance, code-simplifier |
| **npm Skills**     | Installed locally via `npx skills add`    | 5     | golang-pro, browser-use, database-schema-designer, skill-creator, find-skills           |
| **Custom Skills**  | Built-in, located in `skills/`            | 7     | api-designer, git-conventions, migration-safety, dependency-audit, error-handling, code-architecture, testing-strategy |
| **Agents**         | Reusable prompts for Task tool delegation | 18    | go-{coder,reviewer,tester}, ts-{coder,reviewer,tester}, py-{coder,reviewer,tester}, react-{coder,reviewer,tester}, db-{coder,reviewer,tester}, docker-{builder,reviewer}, arch-{reviewer} |
| **Claude Rules**   | Domain-specific standards                 | 12    | go-patterns, ts-patterns, py-patterns, react-patterns, git-workflow, api-design, database, observability, security, cross-cutting, concurrency, testing-patterns |
| **Cursor Rules**   | Glob-matched coding standards             | 14    | 000-index, 100-golang, 101-typescript, 102-python, 103-react, 200-planning, 300-git, 400-api-design, 401-database, 500-observability, 501-security, 502-cross-cutting, 503-concurrency, 504-testing |

## How Agents Work

CLAUDE.md makes Claude act as a **Lead Product Architect** that delegates via the Task tool:

```
User: "Add a /health endpoint"

Claude (orchestrator):
  1. Creates a plan (planning template from CLAUDE.md)
  2. Spawns go-coder agent → writes the endpoint code
  3. Spawns go-tester agent → writes table-driven tests (in parallel)
  4. Spawns go-reviewer agent → read-only review of both
  5. Runs `go build` + `go test` to verify
```

The orchestrator writes no code itself — it plans, delegates, and verifies build/test output.

Each agent type has a specific role:

| Agent        | Role                  | Tools                               | Writes Code? |
| ------------ | --------------------- | ----------------------------------- | ------------ |
| `*-coder`    | Write production code | Read, Write, Edit, Bash, Glob, Grep | Yes          |
| `*-reviewer` | Read-only critique    | Read, Glob, Grep, Bash              | No           |
| `*-tester`   | Write and run tests   | Read, Write, Edit, Bash, Glob, Grep | Tests only   |

## Day-to-Day Workflow

```bash
# 1. Edit rules in this repo (single source of truth)
nano claude/CLAUDE.md

# 2. Deploy + verify
make deploy

# 3. Structural validation (agents, rules, triads)
make validate

# 4. Run tests (if you changed check-sync.sh)
make test
```

## How Sync Works

- CLAUDE.md and Cursor `.mdc` files share sections (safety, coding patterns, planning)
- `<!-- Sync: ... -->` comments mark what must stay in sync
- `check-sync.sh` extracts matching sections, strips heading levels, normalizes platform terms, and diffs
- If drift is found, edit the repo (single source of truth) and re-deploy

## Adding a New Language

1. Create 3 agent files: `claude/agents/<lang>-coder.md`, `<lang>-reviewer.md`, `<lang>-tester.md`
2. Create a Cursor rule: `cursor/1XX-<lang>.mdc` with appropriate globs
3. Add the language's coding patterns to `claude/CLAUDE.md` (with sync comments)
4. Add a `check-sync` section in `scripts/check-sync.sh` for the new rule
5. Update the agent definitions list in `claude/CLAUDE.md`
6. Run `make deploy`

## Troubleshooting

### `npx: command not found`
Node.js is not installed or not in your PATH. Install via:
```bash
brew install node   # macOS
```

### `EACCES: permission denied` during bootstrap
Don't use `sudo` with npm. Fix npm permissions:
```bash
mkdir -p ~/.npm-global && npm config set prefix '~/.npm-global'
export PATH="$HOME/.npm-global/bin:$PATH"  # add to ~/.zshrc
```

### Cursor not picking up rule changes
Cursor caches `.mdc` files. After `make sync-cursor`:
1. Close all Cursor windows
2. Reopen the project
3. Verify in Cursor settings → Rules that the rules appear

### Sync drift after editing
If `make check` shows drift:
```bash
# 1. See what drifted
make check

# 2. Edit the source file (claude/rules/ or claude/CLAUDE.md)
# 3. Re-deploy
make deploy
```

Always edit the repo first (single source of truth), then deploy. Never edit `~/.claude/` or `~/.cursor/rules/` directly.

### Plugin not available in Claude Code
Plugins require Claude Code CLI and the `settings.json` to be deployed:
```bash
# Re-deploy settings
make bootstrap
# Or manually copy
cp claude/settings.json ~/.claude/settings.json
```

Then restart Claude Code for plugins to load.

### `make watch` says `fswatch not found`
Install fswatch:
```bash
brew install fswatch   # macOS
```
