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
│   ├── Architecture/structure → arch-reviewer (skill: code-architecture)
│   └── Database review → db-reviewer (plugin: code-review)
│
├── Write tests
│   ├── Go tests → go-tester (skill: testing-strategy)
│   ├── TypeScript tests → ts-tester (skill: testing-strategy)
│   ├── Python tests → py-tester (skill: testing-strategy)
│   ├── React component tests → react-tester (skill: testing-strategy)
│   ├── Database tests → db-tester (skill: testing-strategy)
│   └── Docker validation → docker-tester (skill: testing-strategy)
│
├── Write documentation → docs-writer
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

| Agent | Custom Skills Used | Plugins Used |
|-------|-------------------|--------------|
| go-coder | error-handling | golang-pro, context7, code-simplifier |
| go-reviewer | — | code-review, security-guidance |
| go-tester | testing-strategy | superpowers (TDD) |
| ts-coder | error-handling | context7, code-simplifier |
| ts-reviewer | — | code-review, security-guidance |
| ts-tester | testing-strategy | superpowers (TDD) |
| py-coder | error-handling | context7, code-simplifier |
| py-reviewer | — | code-review, security-guidance |
| py-tester | testing-strategy | superpowers (TDD) |
| react-coder | error-handling | frontend-design, context7, code-simplifier |
| react-reviewer | — | code-review, security-guidance |
| react-tester | testing-strategy | superpowers (TDD) |
| db-coder | migration-safety | database-schema-designer, context7, code-simplifier |
| db-reviewer | — | code-review, security-guidance |
| db-tester | testing-strategy | superpowers (TDD) |
| docker-builder | — | code-simplifier |
| docker-reviewer | — | code-review, security-guidance |
| docker-tester | testing-strategy | — |
| arch-reviewer | code-architecture | code-review, security-guidance |
| docs-writer | — | — |

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
