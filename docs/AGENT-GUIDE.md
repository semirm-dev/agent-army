# Agent Decision Guide

Quick reference for selecting the right agent, skill, and plugin combination.

## Decision Tree

```
What do you need to do?
│
├── Write code
│   ├── Go → go-coder (plugin: golang-pro, context7)
│   ├── TypeScript/JS → ts-coder (plugin: context7, code-simplifier)
│   ├── Python → py-coder (plugin: context7, code-simplifier)
│   ├── React/Frontend → react-coder (plugin: frontend-design, context7)
│   ├── Database (migrations, queries) → db-coder (plugin: database-schema-designer, context7)
│   └── Docker/CI/CD → docker-builder
│
├── Review code
│   ├── Language-specific review → {lang}-reviewer (plugin: code-review, security-guidance)
│   ├── Architecture/structure → arch-reviewer (skill: code-architecture, api-designer)
│   └── Database review → db-reviewer (skill: migration-safety, plugin: code-review)
│
├── Write tests
│   ├── Go tests → go-tester (skill: testing-strategy)
│   ├── TypeScript tests → ts-tester (skill: testing-strategy)
│   ├── Python tests → py-tester (skill: testing-strategy)
│   ├── React component tests → react-tester (skill: testing-strategy)
│   ├── Database tests → db-tester (skill: testing-strategy)
│   └── Docker validation → docker-tester (skill: testing-strategy)
│
├── Write documentation → docs-writer (skill: api-designer, git-conventions)
│
├── Design API → invoke api-designer skill
├── Audit dependencies → invoke dependency-audit skill
├── Review migration safety → invoke migration-safety skill
├── Refactor code → invoke refactoring-patterns skill
├── Handle errors → invoke error-handling skill
├── Plan architecture → invoke code-architecture skill
├── Build CLI tool → invoke cli-design skill
└── Git operations → invoke git-conventions skill
```

## Agent → Skill → Plugin Matrix

> **Terminology:** *Custom skills* are markdown files in `~/.claude/skills/` (Claude Code) or `~/.cursor/skills/` (Cursor). *npm skills* (golang-pro, database-schema-designer) are installed via `npx skills add`. *Subagents/MCP* lists Cursor built-in subagents (code-reviewer, code-simplifier, etc.) and MCP servers (context7) referenced in the agent body.

| Agent | Custom Skills | npm Skills | Subagents / MCP |
|-------|--------------|------------|-----------------|
| go-coder | error-handling, code-architecture, api-designer, refactoring-patterns | golang-pro | context7, code-simplifier, type-design-analyzer |
| go-reviewer | error-handling, api-designer, code-architecture | — | code-reviewer, silent-failure-hunter |
| go-tester | testing-strategy | — | context7, superpowers (TDD) |
| ts-coder | error-handling, code-architecture, api-designer, refactoring-patterns | — | context7, code-simplifier, type-design-analyzer |
| ts-reviewer | error-handling, api-designer, code-architecture | — | code-reviewer, silent-failure-hunter |
| ts-tester | testing-strategy | — | context7, superpowers (TDD) |
| py-coder | error-handling, code-architecture, api-designer, refactoring-patterns | — | context7, code-simplifier, type-design-analyzer |
| py-reviewer | error-handling, api-designer, code-architecture | — | code-reviewer, silent-failure-hunter |
| py-tester | testing-strategy | — | context7, superpowers (TDD) |
| react-coder | error-handling, code-architecture, api-designer, refactoring-patterns | — | frontend-design, context7, code-simplifier, type-design-analyzer |
| react-reviewer | error-handling, api-designer, code-architecture | — | code-reviewer, silent-failure-hunter |
| react-tester | testing-strategy | — | context7, superpowers (TDD) |
| db-coder | migration-safety, error-handling, code-architecture, refactoring-patterns | database-schema-designer | context7, code-simplifier, type-design-analyzer |
| db-reviewer | migration-safety | database-schema-designer | code-reviewer, silent-failure-hunter |
| db-tester | testing-strategy, migration-safety | — | context7, superpowers (TDD) |
| docker-builder | cli-design, error-handling | — | context7, code-simplifier |
| docker-reviewer | error-handling, cli-design | — | code-reviewer, silent-failure-hunter |
| docker-tester | testing-strategy | — | context7 |
| arch-reviewer | code-architecture, api-designer, dependency-audit | — | code-reviewer |
| docs-writer | api-designer, git-conventions | — | comment-analyzer |

## Common Workflows

### New Feature (e.g., "Add user authentication")
1. Orchestrator creates plan
2. `db-coder` → writes migration + schema (skill: migration-safety)
3. `{lang}-coder` → writes business logic (skill: error-handling)
4. `{lang}-tester` → writes tests (skill: testing-strategy)
5. `{lang}-reviewer` + `arch-reviewer` → review code and structure
6. Orchestrator verifies: build + tests pass

### Bug Fix
1. Orchestrator invokes `systematic-debugging` (superpowers plugin)
2. `{lang}-coder` → implements fix
3. `{lang}-tester` → writes regression test
4. `{lang}-reviewer` → reviews fix

### Refactoring
1. Invoke `refactoring-patterns` skill
2. `{lang}-tester` → ensure test coverage exists
3. `{lang}-coder` → refactor (one move at a time)
4. `arch-reviewer` → verify structural improvement

### API Design
1. Invoke `api-designer` skill
2. `{lang}-coder` → implement endpoints
3. `{lang}-tester` → write integration tests
4. `docs-writer` → write API documentation

### Dependency Audit
1. Invoke `dependency-audit` skill
2. Run audit commands per language
3. Triage and update dependencies
4. `{lang}-tester` → verify nothing broke

## Cursor Built-in Agents

These additional `subagent_type` values are provided by Cursor and complement the custom agents above. Use them via the Task tool:

| Agent | When to Use |
|-------|-------------|
| `code-reviewer` | Cross-language code review against project guidelines. Use after writing or modifying code, before commits or PRs. |
| `code-simplifier` | Simplify recently modified code for clarity and maintainability. Use after completing a coding task. |
| `comment-analyzer` | Analyze code comments for accuracy and long-term maintainability. Use after generating docstrings or before finalizing PRs. |
| `docs-researcher` | Fetch library documentation without cluttering main context. Use when looking up unfamiliar APIs. |
| `pr-test-analyzer` | Review PR test coverage quality and completeness. Use after creating or updating a PR. |
| `silent-failure-hunter` | Identify silent failures and inadequate error handling. Use after work involving error handling or catch blocks. |
| `type-design-analyzer` | Analyze type design for encapsulation and invariant expression. Use when introducing or refactoring types. |

## Rules Reference

| Rule File | Loaded By | Content |
|-----------|-----------|---------|
| go-patterns.md | go-{coder,reviewer,tester} | Go coding + testing standards |
| ts-patterns.md | ts-{coder,reviewer,tester} | TypeScript coding + testing standards |
| py-patterns.md | py-{coder,reviewer,tester} | Python coding + testing standards |
| react-patterns.md | react-{coder,reviewer,tester} | React component patterns |
| git-workflow.md | orchestrator | Git conventions |
| api-design.md | arch-reviewer, orchestrator | API design patterns |
| database.md | db-{coder,reviewer,tester} | Database patterns |
| observability.md | docker-{builder,reviewer,tester} | Logging, metrics, health checks |
| security.md | all reviewers, docker-{builder,tester} | Security patterns |
| cross-cutting.md | arch-reviewer, orchestrator | Error taxonomy, coverage |
| concurrency.md | arch-reviewer | Concurrency patterns |
| testing-patterns.md | all testers | Testing patterns |
| caching-patterns.md | orchestrator | Caching patterns |
| messaging-patterns.md | orchestrator | Messaging patterns |
| ai-assisted-development.md | orchestrator | AI-friendly code patterns |
