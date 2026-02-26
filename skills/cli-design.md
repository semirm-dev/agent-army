---
name: cli-design
description: "CLI tool design patterns — flags, help text, exit codes, output formatting. Invoke when building command-line tools, admin scripts, or migration runners."
---

# CLI Design Skill

## Flag Conventions

- **Long flags:** `--output`, `--verbose`, `--dry-run` (double dash, kebab-case)
- **Short flags:** `-o`, `-v`, `-n` (single dash, single character) — only for common flags
- **Boolean flags:** `--verbose` (presence = true), `--no-color` for negation
- **Value flags:** `--output=json` or `--output json` (support both `=` and space)
- **Required vs optional:** Required flags should error with a clear message. Optional flags have sensible defaults.
- **Environment variable fallback:** Support `--token` flag with `APP_TOKEN` env var fallback. Flag takes precedence.

## Help Text Structure

```
Usage: myapp <command> [options]

Commands:
  serve     Start the HTTP server
  migrate   Run database migrations
  seed      Seed the database with sample data

Options:
  -p, --port <port>     Port to listen on (default: 8080)
  -v, --verbose         Enable verbose logging
  -h, --help            Show this help message
      --version         Show version number

Examples:
  myapp serve --port 3000
  myapp migrate --dry-run
  myapp seed --count 100
```

Key rules:
- Always include `--help` and `--version`
- Show defaults in help text
- Include 2-3 practical examples
- Group related flags with blank lines
- Align descriptions for readability

## Exit Codes

| Code | Meaning | When to Use |
|------|---------|-------------|
| `0`  | Success | Command completed normally |
| `1`  | General error | Runtime failure, unhandled exception |
| `2`  | Usage error | Invalid flags, missing required args |

- **Consistent exit codes.** Document them in `--help` output.
- **Non-zero on failure.** Scripts depend on exit codes for control flow.

## Error Messages

- **Write to stderr** (`os.Stderr`, `console.error`, `sys.stderr`). Never mix errors with stdout output.
- **Format:** `error: <what went wrong>` — lowercase, no period, actionable.
- **Suggestions:** Include the correct flag or command when user input is wrong.
  ```
  error: unknown flag --port
  Did you mean --port? Run 'myapp --help' for usage.
  ```

## Output Formatting

- **Default:** Human-readable text for TTY, machine-readable for pipes.
- **`--json` flag:** Output structured JSON. Every CLI that produces data should support `--json`.
- **`--quiet` flag:** Suppress non-essential output. Only errors and requested data.
- **TTY detection:** Auto-detect terminal and adjust output:
  - TTY: colors, progress bars, tables
  - Pipe: plain text, no ANSI codes, newline-delimited

## Per-Language Libraries

| Language | Library | Notes |
|----------|---------|-------|
| Go | `cobra` + `viper` | Industry standard. Cobra for commands, Viper for config |
| TypeScript | `commander` or `yargs` | Commander for simple CLIs, yargs for complex |
| Python | `click` or `typer` | Click is battle-tested, Typer adds type hints |

## Checklist

Before shipping a CLI tool:
- [ ] `--help` works and shows examples
- [ ] `--version` outputs version string
- [ ] Exit codes are consistent (0/1/2)
- [ ] Errors go to stderr
- [ ] `--json` flag for machine-readable output
- [ ] Required flags show clear error when missing
- [ ] Environment variable fallbacks for sensitive flags (tokens, passwords)
- [ ] No hardcoded paths — use `$HOME`, config dirs, or flags
