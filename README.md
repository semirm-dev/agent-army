# agent-army
This is intended for personal use. Use with caution and adjust as needed.

Portable AI development setup for **Claude Code** and **Cursor**. One repo, one `bootstrap.sh`, consistent rules across devices.

## Quick Start

```bash
git clone <this-repo> ~/workspace/agent-rules
cd ~/workspace/agent-rules
make bootstrap
```

Bootstrap configures your machine for the agent-rules workflow — it installs skills and deploys settings (which enables plugins). Every step lists what will be installed and asks for confirmation.

| Step | Action                                                                   |
| ---- | ------------------------------------------------------------------------ |
| 1    | Check prerequisites (node, npx, claude CLI, rsync)                       |
| 2    | Sync rules to `~/.claude/` and `~/.cursor/rules/`                        |
| 3    | Install Agent Skills (golang-pro, database-schema-designer, skill-creator) |
| 4    | Install Claude Plugins (deploy `settings.json`, shows diff if exists)    |
| 5    | Verify installation (list skills, agents, run check-sync)                |

Idempotent — skips already-installed components on re-run.

## What Gets Deployed Where

```
Repo                            → Deployed to
─────────────────────────────────────────────────────────
claude/CLAUDE.md                → ~/.claude/CLAUDE.md
claude/agents/*.md              → ~/.claude/agents/*.md
claude/rules/*.md               → ~/.claude/rules/*.md
claude/skills/*/SKILL.md        → ~/.claude/skills/*/SKILL.md
claude/skills/*/SKILL.md        → ~/.cursor/skills/*/SKILL.md
claude/statusline-command.sh    → ~/.claude/statusline-command.sh
claude/settings.json            → ~/.claude/settings.json (bootstrap only)
cursor/*.mdc                    → ~/.cursor/rules/*.mdc
```

## Capabilities

| Source             | What                                      | Count | Items                                                                                   |
| ------------------ | ----------------------------------------- | ----- | --------------------------------------------------------------------------------------- |
| **Plugins**        | Auto-updating, managed by Claude CLI      | 6     | superpowers, context7, frontend-design, code-review, security-guidance, code-simplifier |
| **npm Skills**     | Installed locally via `npx skills add`    | 3     | golang-pro, database-schema-designer, skill-creator                                     |
| **Custom Skills**  | Deployed to both Claude and Cursor        | 9     | api-designer, git-conventions, migration-safety, dependency-audit, error-handling, code-architecture, testing-strategy, cli-design, refactoring-patterns |
| **Agents**         | Reusable prompts for Task tool delegation | 20    | go-{coder,reviewer,tester}, ts-{coder,reviewer,tester}, py-{coder,reviewer,tester}, react-{coder,reviewer,tester}, db-{coder,reviewer,tester}, docker-{builder,reviewer,tester}, arch-{reviewer}, docs-{writer} |
| **Claude Rules**   | Domain-specific standards                 | 15    | go-patterns, ts-patterns, py-patterns, react-patterns, git-workflow, api-design, database, observability, security, cross-cutting, concurrency, testing-patterns, caching-patterns, messaging-patterns, ai-assisted-development |
| **Cursor Rules**   | Glob-matched coding standards             | 17    | 000-index, 100-golang, 101-typescript, 102-python, 103-react, 200-planning, 300-git, 400-api-design, 401-database, 500-observability, 501-security, 502-cross-cutting, 503-concurrency, 504-testing, 505-caching, 506-messaging, 507-ai-dev |

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

| Agent        | Role                    | Tools                               | Writes Code? |
| ------------ | ----------------------- | ----------------------------------- | ------------ |
| `*-coder`    | Write production code   | Read, Write, Edit, Bash, Glob, Grep | Yes          |
| `*-reviewer` | Read-only critique      | Read, Glob, Grep, Bash              | No           |
| `*-tester`   | Write and run tests     | Read, Write, Edit, Bash, Glob, Grep | Tests only   |
| `*-builder`  | Write infrastructure    | Read, Write, Edit, Bash, Glob, Grep | Config only  |
| `*-writer`   | Write documentation     | Read, Write, Edit, Glob, Grep       | Docs only    |


## Adding a New Language

```bash
make new-language
```

The script prompts for four inputs:

| Prompt | Example | Default |
| --- | --- | --- |
| Language short name | `rust`, `java`, `elixir` | — |
| Display name | `Rust`, `Java`, `Elixir` | Capitalized short name |
| File extension glob | `**/*.rs`, `**/*.java` | — |
| Cursor rule number | `104` | Next available in 1XX range |

It then scaffolds everything automatically:

- `claude/rules/<lang>-patterns.md` — coding + testing standards (with sync header)
- `cursor/<num>-<lang>.mdc` — glob-matched Cursor rule (synced content)
- `claude/agents/<lang>-coder.md`, `<lang>-reviewer.md`, `<lang>-tester.md` — agent files
- `config.json` — agent group and sync pair entries
- `CLAUDE.md` — regenerated to include the new agents

After running, fill in the TODO placeholders with language-specific patterns, then deploy:

```bash
make check && make sync
```

Always edit the repo first (single source of truth), then deploy. Never edit `~/.claude/` or `~/.cursor/rules/` directly.