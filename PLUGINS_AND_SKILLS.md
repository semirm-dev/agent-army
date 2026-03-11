# Claude Code — Installed Plugins & Skills

> Generated: 2026-03-11

## Plugins (11)

| # | Name | Description | Source | Install |
|---|------|-------------|--------|---------|
| 1 | **superpowers** (v5.0.1) | Core skills library — TDD, debugging, planning, code review, collaboration patterns. Includes SessionStart hook. | [obra/superpowers](https://github.com/obra/superpowers) via `claude-plugins-official` | `/plugin install superpowers@claude-plugins-official` |
| 2 | **context7** | Up-to-date documentation lookup for any library via MCP server (`@upstash/context7-mcp`). | [upstash/context7](https://github.com/upstash/context7) via `claude-plugins-official` | `/plugin install context7@claude-plugins-official` |
| 3 | **code-review** | Automated code review with specialized agents, confidence scoring, and structured feedback. | `claude-plugins-official` | `/plugin install code-review@claude-plugins-official` |
| 4 | **code-simplifier** (v1.0.0) | Simplifies and refines code for clarity, consistency, and maintainability. Custom agent subtype. | `claude-plugins-official` | `/plugin install code-simplifier@claude-plugins-official` |
| 5 | **coderabbit** (v1.0.0) | AI-powered code review powered by [CodeRabbit](https://github.com/coderabbitai/claude-plugin). | `claude-plugins-official` | `/plugin install coderabbit@claude-plugins-official` |
| 6 | **frontend-design** | Skill for creating distinctive, production-grade frontend interfaces with high design quality. | `claude-plugins-official` | `/plugin install frontend-design@claude-plugins-official` |
| 7 | **security-guidance** | PreToolUse hook that warns about security issues when editing files (runs Python hook on Edit/Write). | `claude-plugins-official` | `/plugin install security-guidance@claude-plugins-official` |
| 8 | **claude-code-setup** (v1.0.0) | Analyzes codebases and recommends Claude Code automations (hooks, skills, MCP servers, subagents). | `claude-plugins-official` | `/plugin install claude-code-setup@claude-plugins-official` |
| 9 | **claude-md-management** (v1.0.0) | Tools to maintain CLAUDE.md files — audit quality, capture session learnings, improve docs. | `claude-plugins-official` | `/plugin install claude-md-management@claude-plugins-official` |
| 10 | **skill-creator** | Create, improve, and benchmark skills with evals and performance variance analysis. | `claude-plugins-official` | `/plugin install skill-creator@claude-plugins-official` |
| 11 | **gopls-lsp** (v1.0.0) | Go Language Server Protocol integration for diagnostics and code intelligence. | `claude-plugins-official` | `/plugin install gopls-lsp@claude-plugins-official` |

---

## Skills (28)

Install skills globally with `npx skills add <repo> -g -s <skill-name>`. Add `-l` to list available skills before installing.

> **Note:** The 14 [obra/superpowers](https://github.com/obra/superpowers) skills (brainstorming, systematic-debugging, TDD, writing-plans, executing-plans, dispatching-parallel-agents, subagent-driven-development, finishing-a-development-branch, receiving-code-review, requesting-code-review, using-git-worktrees, using-superpowers, verification-before-completion, writing-skills) are provided by the **superpowers plugin** and invoked via the `superpowers:` prefix (e.g., `superpowers:brainstorming`). They are not installed as standalone skills.
>
> **Deprecated aliases:** The following superpowers skill names are deprecated but still functional:
> - `superpowers:brainstorm` → use `superpowers:brainstorming`
> - `superpowers:write-plan` → use `superpowers:writing-plans`
> - `superpowers:execute-plan` → use `superpowers:executing-plans`

### From [jeffallan/claude-skills](https://github.com/jeffallan/claude-skills) (17 skills)

| Skill | Description | Install |
|-------|-------------|---------|
| `golang-pro` | Concurrent Go patterns, microservices (gRPC/REST), pprof optimization, idiomatic Go. | `npx skills add jeffallan/claude-skills -g -s golang-pro` |
| `api-designer` | REST/GraphQL API design, OpenAPI specs, versioning, pagination, error handling. | `npx skills add jeffallan/claude-skills -g -s api-designer` |
| `architecture-designer` | System architecture diagrams, ADRs, technology trade-off evaluation. | `npx skills add jeffallan/claude-skills -g -s architecture-designer` |
| `cli-developer` | CLI tools, argument parsing, interactive prompts, shell completion scripts. | `npx skills add jeffallan/claude-skills -g -s cli-developer` |
| `database-optimizer` | Slow query investigation, execution plans, index design, query rewrites. | `npx skills add jeffallan/claude-skills -g -s database-optimizer` |
| `fullstack-guardian` | Security-focused full-stack web apps with layered security at every level. | `npx skills add jeffallan/claude-skills -g -s fullstack-guardian` |
| `javascript-pro` | Modern ES2023+ JS, async/await, ESM modules, Node.js APIs. | `npx skills add jeffallan/claude-skills -g -s javascript-pro` |
| `laravel-specialist` | Laravel 10+: Eloquent, Sanctum auth, Horizon queues, API resources. | `npx skills add jeffallan/claude-skills -g -s laravel-specialist` |
| `nestjs-expert` | NestJS modules, controllers, services, DTOs, guards, interceptors. | `npx skills add jeffallan/claude-skills -g -s nestjs-expert` |
| `nextjs-developer` | Next.js 14+ App Router, server components, server actions, streaming SSR. | `npx skills add jeffallan/claude-skills -g -s nextjs-developer` |
| `php-pro` | Modern PHP 8.3+, Laravel, Symfony, strict typing, PHPStan, async with Swoole. | `npx skills add jeffallan/claude-skills -g -s php-pro` |
| `postgres-pro` | PostgreSQL: EXPLAIN analysis, JSONB, extensions, VACUUM tuning, monitoring. | `npx skills add jeffallan/claude-skills -g -s postgres-pro` |
| `prompt-engineer` | Write, refactor, and evaluate LLM prompts — templates, schemas, test suites. | `npx skills add jeffallan/claude-skills -g -s prompt-engineer` |
| `react-expert` | React 18+: components, custom hooks, rendering debugging, class→functional migration. | `npx skills add jeffallan/claude-skills -g -s react-expert` |
| `sql-pro` | SQL query optimization, schema design, performance troubleshooting. | `npx skills add jeffallan/claude-skills -g -s sql-pro` |
| `vue-expert` | Vue 3 Composition API, Nuxt 3 SSR/SSG, Pinia stores, Quasar/Capacitor. | `npx skills add jeffallan/claude-skills -g -s vue-expert` |
| `websocket-engineer` | Real-time WebSockets/Socket.IO, Redis scaling, presence tracking, rooms. | `npx skills add jeffallan/claude-skills -g -s websocket-engineer` |

### From [anthropics/skills](https://github.com/anthropics/skills) (2 skills)

| Skill | Description | Install |
|-------|-------------|---------|
| `frontend-design` | Distinctive, production-grade frontend interfaces with high design quality. | `npx skills add anthropics/skills -g -s frontend-design` |
| `skill-creator` | Create new skills, modify existing, measure performance with evals. | `npx skills add anthropics/skills -g -s skill-creator` |

### Plugin-Provided Skills (9)

Skills exposed by installed plugins, invoked via the `Skill` tool or `/skill-name` shorthand. These do not require separate installation.

| Skill | Description | Plugin Source |
|-------|-------------|---------------|
| `simplify` | Review changed code for reuse, quality, and efficiency, then fix issues found. | code-simplifier |
| `keybindings-help` | Customize keyboard shortcuts, rebind keys, add chord bindings, modify keybindings.json. | built-in |
| `loop` | Run a prompt or slash command on a recurring interval (e.g., `/loop 5m /foo`). | built-in |
| `claude-api` | Build apps with the Claude API or Anthropic SDK. Triggers on `anthropic`/`@anthropic-ai/sdk` imports. | built-in |
| `code-review:code-review` | Code review a pull request. | code-review |
| `claude-md-management:revise-claude-md` | Update CLAUDE.md with learnings from the current session. | claude-md-management |
| `claude-md-management:claude-md-improver` | Audit and improve CLAUDE.md files in repositories. | claude-md-management |
| `claude-code-setup:claude-automation-recommender` | Analyze a codebase and recommend Claude Code automations (hooks, subagents, skills, plugins, MCP servers). | claude-code-setup |
| `coderabbit:review` | Run CodeRabbit AI code review on your changes. | coderabbit |

---

## Custom Agents (3)

| Agent | Description | Provided By |
|-------|-------------|-------------|
| `superpowers:code-reviewer` | Reviews completed code against plan and coding standards. Use after major implementation steps. | superpowers plugin |
| `code-simplifier:code-simplifier` | Simplifies code for clarity, consistency, and maintainability. Focuses on recently modified code. | code-simplifier plugin |
| `coderabbit:code-reviewer` | AI-powered code review with CodeRabbit analysis of code changes. | coderabbit plugin |

---

## Plugin Marketplaces (1)

| Marketplace | Source | Browse |
|-------------|--------|--------|
| `claude-plugins-official` | [anthropics/claude-plugins-official](https://github.com/anthropics/claude-plugins-official) | `/plugins` |

---

## MCP Servers (1)

| Server | Transport | Endpoint |
|--------|-----------|----------|
| `context7` | HTTP | `https://mcp.context7.com/mcp` |
