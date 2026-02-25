# Capabilities

## Plugins (`claude plugins add`)

Plugins are maintained upstream and auto-update. No local files to manage.

| Plugin | What it provides | Status |
|--------|-----------------|--------|
| `superpowers` | brainstorming, systematic-debugging, writing-plans, TDD, code-review, parallel agents, git worktrees | Active — core workflow |
| `context7` | Documentation lookup for any library | Active — used via MCP |
| `frontend-design` | UI/design guidance and component generation | Active — use for UI work |
| `code-review` | PR review command (`/review-pr`) | Active — use for PRs |
| `security-guidance` | Security analysis hooks | Underutilized — consider wiring into reviewer agents or removing if unused |
| `code-simplifier` | Refactoring and code simplification agent | Active — use for refactoring |

## npm Skills (`npx skills add`)

Skills are installed locally and symlinked into `~/.claude/skills/`.

| Skill | Install command | Status |
|-------|----------------|--------|
| `golang-pro` | `npx skills add https://github.com/jeffallan/claude-skills --skill golang-pro` | Review overlap with CLAUDE.md Go patterns; keep only if it provides unique value (concurrency templates, generics patterns) |
| `browser-use` | `npx skills add https://github.com/browser-use/browser-use --skill browser-use` | Consider removing unless actively used for browser automation |
| `database-schema-designer` | `npx skills add https://github.com/softaworks/agent-toolkit --skill database-schema-designer` | Keep — add DB migration patterns to CLAUDE.md to complement |
| `skill-creator` | `npx skills add https://github.com/anthropics/skills --skill skill-creator` | Keep — useful for creating custom skills |
| `find-skills` | `npx skills add https://github.com/anthropics/skills --skill find-skills` | Consider removing — only needed during initial setup |

## Custom Skills to Create

These are recommended custom skills to build using `skill-creator`:

| Skill | Purpose |
|-------|---------|
| `git-conventions` | Enforce branch naming, commit format, PR templates. Uses rules from CLAUDE.md Git Workflow section |
| `api-designer` | REST/gRPC API design patterns, error formats, pagination. Uses rules from CLAUDE.md API Design section |

