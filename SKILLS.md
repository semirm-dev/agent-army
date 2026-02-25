# Capabilities

## Plugins (`claude plugins add`)

Plugins are maintained upstream and auto-update. No local files to manage.

| Plugin | What it provides | Status |
|--------|-----------------|--------|
| `superpowers` | brainstorming, systematic-debugging, writing-plans, TDD, code-review, parallel agents, git worktrees | Active — core workflow |
| `context7` | Documentation lookup for any library | Active — used via MCP |
| `frontend-design` | UI/design guidance and component generation | Active — use for UI work |
| `code-review` | PR review command (`/review-pr`) | Active — use for PRs |
| `security-guidance` | Security analysis hooks | Active — wired into reviewer agents |
| `code-simplifier` | Refactoring and code simplification agent | Active — use for refactoring |

## npm Skills (`npx skills add`)

Skills are installed locally and symlinked into `~/.claude/skills/`.

| Skill | Install command | Status |
|-------|----------------|--------|
| `golang-pro` | `npx skills add https://github.com/jeffallan/claude-skills --skill golang-pro` | Active — invoked by go-coder agent |
| `database-schema-designer` | `npx skills add https://github.com/softaworks/agent-toolkit --skill database-schema-designer` | Active — complements rules/database.md |
| `skill-creator` | `npx skills add https://github.com/anthropics/skills --skill skill-creator` | Active — use to build custom skills |

## Custom Skills to Create

These are recommended custom skills to build using `skill-creator`:

| Skill | Purpose |
|-------|---------|
| `git-conventions` | Enforce branch naming, commit format, PR templates. Uses rules from rules/git-workflow.md |
| `api-designer` | REST/gRPC API design patterns, error formats, pagination. Uses rules from rules/api-design.md |
