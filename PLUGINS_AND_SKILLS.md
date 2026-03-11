# Claude Code — Installed Plugins & Skills

> Generated: 2026-03-11

## Plugins (11)

| # | Name | Description | Source | Install |
|---|------|-------------|--------|--------|
| 1 | **claude-code-setup** (v1.0.0) | Analyze codebases and recommend tailored Claude Code automations such as hooks, skills, MCP servers, and subagents. | `claude-plugins-official` | `/plugin install claude-code-setup@claude-plugins-official` |
| 2 | **claude-md-management** (v1.0.0) | Tools to maintain and improve CLAUDE.md files - audit quality, capture session learnings, and keep project memory current. | `claude-plugins-official` | `/plugin install claude-md-management@claude-plugins-official` |
| 3 | **code-review** | Automated code review for pull requests using multiple specialized agents with confidence-based scoring | `claude-plugins-official` | `/plugin install code-review@claude-plugins-official` |
| 4 | **code-simplifier** (v1.0.0) | Agent that simplifies and refines code for clarity, consistency, and maintainability while preserving functionality | `claude-plugins-official` | `/plugin install code-simplifier@claude-plugins-official` |
| 5 | **coderabbit** (v1.0.0) | AI-powered code review in Claude Code, powered by CodeRabbit | [coderabbitai/claude-plugin](https://github.com/coderabbitai/claude-plugin) via `claude-plugins-official` | `/plugin install coderabbit@claude-plugins-official` |
| 6 | **context7** | Upstash Context7 MCP server for up-to-date documentation lookup. Pull version-specific documentation and code examples directly from source repositories into your LLM context. | `claude-plugins-official` | `/plugin install context7@claude-plugins-official` |
| 7 | **frontend-design** | Frontend design skill for UI/UX implementation | `claude-plugins-official` | `/plugin install frontend-design@claude-plugins-official` |
| 8 | **gopls-lsp** (v1.0.0) |  | `claude-plugins-official` | `/plugin install gopls-lsp@claude-plugins-official` |
| 9 | **security-guidance** | Security reminder hook that warns about potential security issues when editing files, including command injection, XSS, and unsafe code patterns | `claude-plugins-official` | `/plugin install security-guidance@claude-plugins-official` |
| 10 | **skill-creator** | Create new skills, improve existing skills, and measure skill performance. Use when users want to create a skill from scratch, update or optimize an existing skill, run evals to test a skill, or benchmark skill performance with variance analysis. | `claude-plugins-official` | `/plugin install skill-creator@claude-plugins-official` |
| 11 | **superpowers** (v5.0.1) | Core skills library for Claude Code: TDD, debugging, collaboration patterns, and proven techniques | [obra/superpowers](https://github.com/obra/superpowers) via `claude-plugins-official` | `/plugin install superpowers@claude-plugins-official` |

---

## Skills (42)

Install skills globally with `npx skills add <repo> -g -s <skill-name>`. Add `-l` to list available skills before installing.

> **Note:** The 14 [obra/superpowers](https://github.com/obra/superpowers) skills (brainstorming, dispatching-parallel-agents, executing-plans, finishing-a-development-branch, receiving-code-review, requesting-code-review, subagent-driven-development, systematic-debugging, test-driven-development, using-git-worktrees, using-superpowers, verification-before-completion, writing-plans, writing-skills) are provided by the **superpowers plugin** and invoked via the `superpowers:` prefix (e.g., `superpowers:brainstorming`). They are not installed as standalone skills.
>
> **Deprecated aliases:** The following superpowers skill names are deprecated but still functional:
> - `superpowers:brainstorm` → use `superpowers:brainstorming`
> - `superpowers:execute-plan` → use `superpowers:executing-plans`
> - `superpowers:write-plan` → use `superpowers:writing-plans`

### From [browser-use/browser-use](https://github.com/browser-use/browser-use) (1 skill)

| Skill | Description | Install |
|-------|-------------|--------|
| `browser-use` |  | `npx skills add browser-use/browser-use -g -s browser-use` |

### From [jeffallan/claude-skills](https://github.com/jeffallan/claude-skills) (17 skills)

| Skill | Description | Install |
|-------|-------------|--------|
| `api-designer` | Use when designing REST or GraphQL APIs, creating OpenAPI specifications, or planning API architecture. Invoke for resource modeling, versioning strategies, pagination patterns, error handling standards. | `npx skills add jeffallan/claude-skills -g -s api-designer` |
| `architecture-designer` | Use when designing new high-level system architecture, reviewing existing designs, or making architectural decisions. Invoke to create architecture diagrams, write Architecture Decision Records (ADRs), evaluate technology trade-offs, design component interactions, and plan for scalability. Use for system design, architecture review, microservices structuring, ADR authoring, scalability planning, and infrastructure pattern selection — distinct from code-level design patterns or database-only design tasks. | `npx skills add jeffallan/claude-skills -g -s architecture-designer` |
| `cli-developer` | Use when building CLI tools, implementing argument parsing, or adding interactive prompts. Invoke for parsing flags and subcommands, displaying progress bars and spinners, generating bash/zsh/fish completion scripts, CLI design, shell completions, and cross-platform terminal applications using commander, click, typer, or cobra. | `npx skills add jeffallan/claude-skills -g -s cli-developer` |
| `database-optimizer` | Optimizes database queries and improves performance across PostgreSQL and MySQL systems. Use when investigating slow queries, analyzing execution plans, or optimizing database performance. Invoke for index design, query rewrites, configuration tuning, partitioning strategies, lock contention resolution. | `npx skills add jeffallan/claude-skills -g -s database-optimizer` |
| `fullstack-guardian` | Builds security-focused full-stack web applications by implementing integrated frontend and backend components with layered security at every level. Covers the complete stack from database to UI, enforcing auth, input validation, output encoding, and parameterized queries across all layers. Use when implementing features across frontend and backend, building REST APIs with corresponding UI, connecting frontend components to backend endpoints, creating end-to-end data flows from database to UI, or implementing CRUD operations with UI forms. Distinct from frontend-only, backend-only, or API-only skills in that it simultaneously addresses all three perspectives—Frontend, Backend, and Security—within a single implementation workflow. Invoke for full-stack feature work, web app development, authenticated API routes with views, microservices, real-time features, monorepo architecture, or technology selection decisions. | `npx skills add jeffallan/claude-skills -g -s fullstack-guardian` |
| `golang-pro` | Implements concurrent Go patterns using goroutines and channels, designs and builds microservices with gRPC or REST, optimizes Go application performance with pprof, and enforces idiomatic Go with generics, interfaces, and robust error handling. Use when building Go applications requiring concurrent programming, microservices architecture, or high-performance systems. Invoke for goroutines, channels, Go generics, gRPC integration, CLI tools, benchmarks, or table-driven testing. | `npx skills add jeffallan/claude-skills -g -s golang-pro` |
| `javascript-pro` | Writes, debugs, and refactors JavaScript code using modern ES2023+ features, async/await patterns, ESM module systems, and Node.js APIs. Use when building vanilla JavaScript applications, implementing Promise-based async flows, optimising browser or Node.js performance, working with Web Workers or Fetch API, or reviewing .js/.mjs/.cjs files for correctness and best practices. | `npx skills add jeffallan/claude-skills -g -s javascript-pro` |
| `laravel-specialist` | Build and configure Laravel 10+ applications, including creating Eloquent models and relationships, implementing Sanctum authentication, configuring Horizon queues, designing RESTful APIs with API resources, and building reactive interfaces with Livewire. Use when creating Laravel models, setting up queue workers, implementing Sanctum auth flows, building Livewire components, optimising Eloquent queries, or writing Pest/PHPUnit tests for Laravel features. | `npx skills add jeffallan/claude-skills -g -s laravel-specialist` |
| `nestjs-expert` | Creates and configures NestJS modules, controllers, services, DTOs, guards, and interceptors for enterprise-grade TypeScript backend applications. Use when building NestJS REST APIs or GraphQL services, implementing dependency injection, scaffolding modular architecture, adding JWT/Passport authentication, integrating TypeORM or Prisma, or working with .module.ts, .controller.ts, and .service.ts files. Invoke for guards, interceptors, pipes, validation, Swagger documentation, and unit/E2E testing in NestJS projects. | `npx skills add jeffallan/claude-skills -g -s nestjs-expert` |
| `nextjs-developer` | Use when building Next.js 14+ applications with App Router, server components, or server actions. Invoke to configure route handlers, implement middleware, set up API routes, add streaming SSR, write generateMetadata for SEO, scaffold loading.tsx/error.tsx boundaries, or deploy to Vercel. Triggers on: Next.js, Next.js 14, App Router, RSC, use server, Server Components, Server Actions, React Server Components, generateMetadata, loading.tsx, Next.js deployment, Vercel, Next.js performance. | `npx skills add jeffallan/claude-skills -g -s nextjs-developer` |
| `php-pro` | Use when building PHP applications with modern PHP 8.3+ features, Laravel, or Symfony frameworks. Invokes strict typing, PHPStan level 9, async patterns with Swoole, and PSR standards. Creates controllers, configures middleware, generates migrations, writes PHPUnit/Pest tests, defines typed DTOs and value objects, sets up dependency injection, and scaffolds REST/GraphQL APIs. Use when working with Eloquent, Doctrine, Composer, Psalm, ReactPHP, or any PHP API development. | `npx skills add jeffallan/claude-skills -g -s php-pro` |
| `postgres-pro` | Use when optimizing PostgreSQL queries, configuring replication, or implementing advanced database features. Invoke for EXPLAIN analysis, JSONB operations, extension usage, VACUUM tuning, performance monitoring. | `npx skills add jeffallan/claude-skills -g -s postgres-pro` |
| `prompt-engineer` | Writes, refactors, and evaluates prompts for LLMs — generating optimized prompt templates, structured output schemas, evaluation rubrics, and test suites. Use when designing prompts for new LLM applications, refactoring existing prompts for better accuracy or token efficiency, implementing chain-of-thought or few-shot learning, creating system prompts with personas and guardrails, building JSON/function-calling schemas, or developing prompt evaluation frameworks to measure and improve model performance. | `npx skills add jeffallan/claude-skills -g -s prompt-engineer` |
| `react-expert` | Use when building React 18+ applications in .jsx or .tsx files, Next.js App Router projects, or create-react-app setups. Creates components, implements custom hooks, debugs rendering issues, migrates class components to functional, and implements state management. Invoke for Server Components, Suspense boundaries, useActionState forms, performance optimization, or React 19 features. | `npx skills add jeffallan/claude-skills -g -s react-expert` |
| `sql-pro` | Optimizes SQL queries, designs database schemas, and troubleshoots performance issues. Use when a user asks why their query is slow, needs help writing complex joins or aggregations, mentions database performance issues, or wants to design or migrate a schema. Invoke for complex queries, window functions, CTEs, indexing strategies, query plan analysis, covering index creation, recursive queries, EXPLAIN/ANALYZE interpretation, before/after query benchmarking, or migrating queries between database dialects (PostgreSQL, MySQL, SQL Server, Oracle). | `npx skills add jeffallan/claude-skills -g -s sql-pro` |
| `vue-expert` | Builds Vue 3 components with Composition API patterns, configures Nuxt 3 SSR/SSG projects, sets up Pinia stores, scaffolds Quasar/Capacitor mobile apps, implements PWA features, and optimises Vite builds. Use when creating Vue 3 applications with Composition API, writing reusable composables, managing state with Pinia, building hybrid mobile apps with Quasar or Capacitor, configuring service workers, or tuning Vite configuration and TypeScript integration. | `npx skills add jeffallan/claude-skills -g -s vue-expert` |
| `websocket-engineer` | Use when building real-time communication systems with WebSockets or Socket.IO. Invoke for bidirectional messaging, horizontal scaling with Redis, presence tracking, room management. | `npx skills add jeffallan/claude-skills -g -s websocket-engineer` |

### From [softaworks/agent-toolkit](https://github.com/softaworks/agent-toolkit) (1 skill)

| Skill | Description | Install |
|-------|-------------|--------|
| `database-schema-designer` |  | `npx skills add softaworks/agent-toolkit -g -s database-schema-designer` |

### From [vercel-labs/skills](https://github.com/vercel-labs/skills) (1 skill)

| Skill | Description | Install |
|-------|-------------|--------|
| `find-skills` |  | `npx skills add vercel-labs/skills -g -s find-skills` |

### Plugin-Provided Skills

Skills exposed by installed plugins, invoked via the `Skill` tool or `/skill-name` shorthand. These do not require separate installation.

| Skill | Description | Plugin Source |
|-------|-------------|---------------|
| `claude-code-setup:claude-automation-recommender` | Analyze a codebase and recommend Claude Code automations (hooks, subagents, skills, plugins, MCP servers). Use when user asks for automation recommendations, wants to optimize their Claude Code setup, mentions improving Claude Code workflows, asks how to first set up Claude Code for a project, or wants to know what Claude Code features they should use. | claude-code-setup |
| `claude-md-management:claude-md-improver` | Audit and improve CLAUDE.md files in repositories. Use when user asks to check, audit, update, improve, or fix CLAUDE.md files. Scans for all CLAUDE.md files, evaluates quality against templates, outputs quality report, then makes targeted updates. Also use when the user mentions "CLAUDE.md maintenance" or "project memory optimization". | claude-md-management |
| `claude-md-management:revise-claude-md` | Update CLAUDE.md with learnings from this session | claude-md-management |
| `code-review:code-review` | Code review a pull request | code-review |
| `coderabbit:code-review` | Reviews code changes using CodeRabbit AI. Use when user asks for code review, PR feedback, code quality checks, security issues, or wants autonomous fix-review cycles. | coderabbit |
| `coderabbit:review` | Run CodeRabbit AI code review on your changes | coderabbit |
| `frontend-design:frontend-design` | Create distinctive, production-grade frontend interfaces with high design quality. Use this skill when the user asks to build web components, pages, or applications. Generates creative, polished code that avoids generic AI aesthetics. | frontend-design |
| `skill-creator:skill-creator` | Create new skills, modify and improve existing skills, and measure skill performance. Use when users want to create a skill from scratch, update or optimize an existing skill, run evals to test a skill, benchmark skill performance with variance analysis, or optimize a skill's description for better triggering accuracy. | skill-creator |
| `superpowers:brainstorming` | You MUST use this before any creative work - creating features, building components, adding functionality, or modifying behavior. Explores user intent, requirements and design before implementation. | superpowers |
| `superpowers:dispatching-parallel-agents` | Use when facing 2+ independent tasks that can be worked on without shared state or sequential dependencies | superpowers |
| `superpowers:executing-plans` | Use when you have a written implementation plan to execute in a separate session with review checkpoints | superpowers |
| `superpowers:finishing-a-development-branch` | Use when implementation is complete, all tests pass, and you need to decide how to integrate the work - guides completion of development work by presenting structured options for merge, PR, or cleanup | superpowers |
| `superpowers:receiving-code-review` | Use when receiving code review feedback, before implementing suggestions, especially if feedback seems unclear or technically questionable - requires technical rigor and verification, not performative agreement or blind implementation | superpowers |
| `superpowers:requesting-code-review` | Use when completing tasks, implementing major features, or before merging to verify work meets requirements | superpowers |
| `superpowers:subagent-driven-development` | Use when executing implementation plans with independent tasks in the current session | superpowers |
| `superpowers:systematic-debugging` | Use when encountering any bug, test failure, or unexpected behavior, before proposing fixes | superpowers |
| `superpowers:test-driven-development` | Use when implementing any feature or bugfix, before writing implementation code | superpowers |
| `superpowers:using-git-worktrees` | Use when starting feature work that needs isolation from current workspace or before executing implementation plans - creates isolated git worktrees with smart directory selection and safety verification | superpowers |
| `superpowers:using-superpowers` | Use when starting any conversation - establishes how to find and use skills, requiring Skill tool invocation before ANY response including clarifying questions | superpowers |
| `superpowers:verification-before-completion` | Use when about to claim work is complete, fixed, or passing, before committing or creating PRs - requires running verification commands and confirming output before making any success claims; evidence before assertions always | superpowers |
| `superpowers:writing-plans` | Use when you have a spec or requirements for a multi-step task, before touching code | superpowers |
| `superpowers:writing-skills` | Use when creating new skills, editing existing skills, or verifying skills work before deployment | superpowers |

---

## Custom Agents (3)

| Agent | Description | Provided By |
|-------|-------------|-------------|
| `code-simplifier:code-simplifier` | Simplifies and refines code for clarity, consistency, and maintainability while preserving all functionality. | code-simplifier plugin |
| `coderabbit:code-reviewer` | Specialized CodeRabbit code review agent that performs thorough analysis of code changes | coderabbit plugin |
| `superpowers:code-reviewer` | Use this agent when a major project step has been completed and needs to be reviewed against the original plan and coding standards. | superpowers plugin |

---

## Plugin Marketplaces (1)

| Marketplace | Source | Browse |
|-------------|--------|--------|
| `claude-plugins-official` | [anthropics/claude-plugins-official](https://github.com/anthropics/claude-plugins-official) | `/plugins` |

---

## MCP Servers (1)

| Server | Transport | Endpoint |
|--------|-----------|----------|
| `context7` | stdio | `npx -y @upstash/context7-mcp` |
