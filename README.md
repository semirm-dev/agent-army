# Agent Army

A modular library of coding standards, workflows, and agent prompts for AI-assisted development.

## What's Inside

### Rules (`rules/`)

Coding standards and best practices. Markdown files with YAML frontmatter. Scoped as `universal` (all languages) or `language-specific` (e.g., `go/patterns`). Rules define **what** good code looks like — naming, error handling, security, testing, etc.

Examples: `api-design`, `security`, `go/patterns`, `typescript/testing`

### Skills (`skills/`)

Structured workflows and decision trees. Also markdown with frontmatter. Skills define **how** to accomplish tasks — designing APIs, setting up caching, hardening security. Skills declare `uses_rules` to reference the rules they depend on.

Examples: `api-designer`, `caching-strategy`, `go/coder`, `react/tester`

### Agents (`agents/`)

Prompt templates for specialized AI roles. Define a role (coder, reviewer, tester), available tools, and behavioral instructions. Agents invoke skills and follow rules when activated.

Examples: `go-coder`, `ts-reviewer`, `py-tester`, `docker-builder`

## How They Relate

```
Rules       → foundational standards         (the "what")
Skills      → task workflows that use rules  (the "how")
Agents      → specialized roles that invoke skills and follow rules  (the "who")
```

```
┌─────────┐     uses_rules     ┌─────────┐     invokes     ┌─────────┐
│  Rules  │◄───────────────────│ Skills  │◄────────────────│ Agents  │
└─────────┘                    └─────────┘                 └─────────┘
  api-design                     api-designer                go-coder
  security                       go/coder                    ts-reviewer
  go/patterns                    react/tester                py-tester
```

## Manifest (`manifest.json`)

Auto-generated index of all rules and skills. Each entry lists:

- **name** — identifier (e.g., `go/patterns`, `api-designer`)
- **scope** — `universal` or `language-specific`
- **languages** — applicable languages (for language-specific entries)
- **uses_rules** — resolved dependencies (transitive — includes indirect dependencies)
- **path** — file path relative to the repo root

Regenerate with `make manifest`.

## Make Commands

| Command | Description |
|---------|-------------|
| `make manifest` | Regenerate `manifest.json` from `rules/` and `skills/` frontmatter |
| `make edit-rules` | Interactively add/remove `uses_rules` entries on any rule or skill |
| `make resolve-rules` | Detect and remove redundant `uses_rules` entries (already covered transitively) |

## File Format

### Rule

```yaml
---
name: go/patterns
description: Go coding conventions, error handling, project structure, and concurrency
scope: language-specific
languages: [go]
uses_rules: [code-quality, security, cross-cutting, observability]
---

# Go Coding Patterns
...
```

### Skill

```yaml
---
name: api-designer
description: API style selection, REST resource design, versioning strategy, ...
scope: universal
languages: []
uses_rules: [api-design, cross-cutting, security]
---

# API Designer
...
```

### Agent

```yaml
---
name: go-coder
description: "Senior Go engineer. Writes production-grade Go code following project patterns."
tools: Read, Write, Edit, Bash, Glob, Grep
model: inherit
---

# Role
...
```

## TODO

- [ ] Add a way to add new rules/skills/agents/plugins from the CLI
