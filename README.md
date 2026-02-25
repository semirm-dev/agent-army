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
│   ├── CLAUDE.md              # Main instructions (safety, coding patterns, planning)
│   ├── settings.json          # Reference settings template
│   ├── statusline-command.sh  # Status line script (deployed to ~/.claude/)
│   └── agents/
│       ├── go-coder.md        # Go code writer
│       ├── go-reviewer.md     # Go code reviewer (read-only)
│       ├── go-tester.md       # Go test writer
│       ├── ts-coder.md        # TypeScript/JS code writer
│       ├── ts-reviewer.md     # TypeScript/JS code reviewer (read-only)
│       └── ts-tester.md       # TypeScript/JS test writer
├── cursor/
│   ├── 000-index.mdc          # Safety & communication (alwaysApply)
│   ├── 100-golang.mdc         # Go coding patterns (globs: **/*.go)
│   ├── 101-typescript.mdc     # TypeScript patterns (globs: **/*.ts,tsx,js,jsx)
│   └── 200-planning.mdc       # Planning template (alwaysApply)
├── scripts/
│   ├── bootstrap.sh           # Interactive new-device setup
│   ├── rsync-rules.sh         # Sync repo → ~/.claude/ or ~/.cursor/rules/
│   ├── check-sync.sh          # Verify CLAUDE.md ↔ Cursor .mdc parity
│   └── test-check-sync.sh     # Tests for check-sync drift detection
├── SKILLS.md                  # What to install (plugins vs npm skills)
└── README.md
```

## What Gets Deployed Where

```
Repo                            → Deployed to
─────────────────────────────────────────────────────────
claude/CLAUDE.md                → ~/.claude/CLAUDE.md
claude/agents/*.md              → ~/.claude/agents/*.md
claude/statusline-command.sh    → ~/.claude/statusline-command.sh
claude/settings.json            → ~/.claude/settings.json (bootstrap only)
cursor/*.mdc                    → ~/.cursor/rules/*.mdc
```

Excluded from sync (user-managed): `~/.claude/settings.json`, `skills/`, `plugins/`, `projects/`, `todos/`.

## Commands

| Command            | Purpose                          |
|--------------------|----------------------------------|
| `make`             | Show all available targets       |
| `make bootstrap`   | First-time interactive setup     |
| `make sync`        | Sync rules to Claude and Cursor  |
| `make sync-claude` | Sync rules to Claude only        |
| `make sync-cursor` | Sync rules to Cursor only        |
| `make check`       | Verify nothing drifted           |
| `make deploy`      | Sync + check (day-to-day loop)   |
| `make test`        | Run test suite                   |

`make check` exit codes:

| Exit code | Meaning                             |
| --------- | ----------------------------------- |
| `0`       | All sections in sync                |
| `1`       | Drift detected — shows unified diff |

Sections checked: Go Coding Patterns, Go Testing, Safety, Communication, TypeScript Coding Patterns, TypeScript Testing, Planning.

## Capabilities

| Source           | What                                      | Items                                                                                   |
| ---------------- | ----------------------------------------- | --------------------------------------------------------------------------------------- |
| **Plugins**      | Auto-updating, managed by Claude CLI      | superpowers, context7, frontend-design, code-review, security-guidance, code-simplifier |
| **npm Skills**   | Installed locally via `npx skills add`    | golang-pro, browser-use, database-schema-designer, skill-creator, find-skills           |
| **Agents**       | Reusable prompts for Task tool delegation | go-coder, go-reviewer, go-tester, ts-coder, ts-reviewer, ts-tester                      |
| **Cursor Rules** | Glob-matched coding standards             | 000-index, 100-golang, 101-typescript, 200-planning                                     |

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

# 3. Run tests (if you changed check-sync.sh)
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
