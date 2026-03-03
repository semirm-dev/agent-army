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

Prompt templates for specialized AI roles. Grouped by language/domain in subdirectories, with cross-cutting agents at the root. Platform-agnostic — no tool names, model references, or IDE-specific paths. Agents declare `uses_skills` (which transitively bring rules) and `uses_plugins` for extensions.

Examples: `go/coder`, `typescript/reviewer`, `python/tester`, `infrastructure/builder`, `arch-reviewer`

## How They Relate

```
Rules       → foundational standards         (the "what")
Skills      → task workflows that use rules  (the "how")
Agents      → specialized roles that invoke skills and follow rules  (the "who")
```

```
┌─────────┐     uses_rules     ┌─────────┐   uses_skills   ┌─────────┐
│  Rules  │◄───────────────────│ Skills  │◄────────────────│ Agents  │
└─────────┘                    └─────────┘                 └─────────┘
  api-design                     api-designer                go/coder
  security                       go/coder                    typescript/reviewer
  go/patterns                    react/tester                python/tester
```

## Manifest (`manifest.json`)

Auto-generated index of all rules, skills, and agents. Each entry lists:

- **name** — identifier (e.g., `go/patterns`, `api-designer`, `go/coder`)
- **scope** — `universal` or `language-specific`
- **languages** — applicable languages (for language-specific entries)
- **uses_rules** — resolved dependencies (transitive — includes indirect dependencies)
- **path** — file path relative to the repo root

Agent entries additionally include: **role**, **access**, **uses_skills**, **uses_plugins**, **delegates_to**.

Regenerate with `make manifest`.

## Make Commands

| Command | Description |
|---------|-------------|
| `make manifest` | Scan `rules/`, `skills/`, and `agents/` frontmatter and regenerate `manifest.json`. Resolves `uses_rules` and `delegates_to` transitively, including rules inherited through skills. |
| `make edit-deps` | Interactively add or remove dependency entries (`uses_rules`, `uses_skills`, `uses_plugins`, `delegates_to`) on any rule, skill, or agent file. Rewrites YAML frontmatter in-place, then auto-regenerates the manifest. |
| `make resolve-deps` | Validate all dependency references (`uses_rules`, `uses_skills`, `uses_plugins`, `delegates_to`) across `rules/`, `skills/`, and `agents/`. Detect and remove redundant entries covered by transitive dependencies. |

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
name: go/coder
description: "Senior Go engineer. Writes production-grade Go code following project patterns."
role: coder
scope: language-specific
languages: [go]
access: read-write
uses_skills: [go/coder]
uses_rules: []
uses_plugins: [code-simplifier, context7]
delegates_to: []
---

# Go Coder Agent
...
```

## TODO

- [ ] Add a way to add new rules/skills/agents/plugins from the CLI
