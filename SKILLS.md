# Capabilities

## Plugins (`claude plugins add`)

Plugins are maintained upstream and auto-update. No local files to manage.

| Plugin | What it provides |
|--------|-----------------|
| `superpowers` | brainstorming, systematic-debugging, writing-plans, TDD, code-review, parallel agents, git worktrees |
| `context7` | Documentation lookup for any library |
| `frontend-design` | UI/design guidance and component generation |
| `code-review` | PR review command (`/review-pr`) |
| `security-guidance` | Security analysis hooks |
| `code-simplifier` | Refactoring and code simplification agent |

## npm Skills (`npx skills add`)

Skills are installed locally and symlinked into `~/.claude/skills/`.

| Skill | Install command |
|-------|----------------|
| `golang-pro` | `npx skills add https://github.com/jeffallan/claude-skills --skill golang-pro` |
| `browser-use` | `npx skills add https://github.com/browser-use/browser-use --skill browser-use` |
| `database-schema-designer` | `npx skills add https://github.com/softaworks/agent-toolkit --skill database-schema-designer` |
| `skill-creator` | `npx skills add https://github.com/anthropics/skills --skill skill-creator` |
| `find-skills` | `npx skills add https://github.com/anthropics/skills --skill find-skills` |

