---
name: go/patterns
description: Go coding conventions, error handling, project structure, and concurrency
scope: language-specific
languages: [go]
uses_rules: [code-quality, security, cross-cutting, observability, testing-patterns]
---

# Go Coding Patterns

## Naming and Structure
- **Packages:** Avoid "stuttering." Use `auth.Service` instead of `auth.AuthService`.
- **Naming:** Use `MixedCaps` (Acronyms like `ID`, `HTTP`, `URL` should be consistent case).
- **Project structure:** Follow vertical-slices architecture (feature + hexagonal/clean), package by feature. Keep `cmd/` thin -- delegate to packages immediately.
- **Godoc:** All exported types, functions, and methods must have a godoc comment starting with the identifier name.

## Error Handling
- **ALWAYS wrap errors with context:** `fmt.Errorf("domain: operation: %w", err)`.
- **Sentinel errors:** Define package-level error values for expected conditions: `var ErrNotFound = errors.New("not found")`. Callers check with `errors.Is(err, ErrNotFound)`.
- Use `errors.Is` and `errors.As` for checking error types.
- **Panics:** Never use `panic()` for normal error paths. Reserve for truly unrecoverable situations.

## Interfaces and Types
- **Interfaces:** "Accept interfaces, return concrete types." Keep interfaces small (2-3 methods max).
- **Receivers:** Use pointer receivers (`*T`) when the method mutates state or the struct is large. Use value receivers (`T`) for small, immutable types. Never mix receiver types on the same type.
- **Type assertions:** Always use the two-value form: `v, ok := x.(Type)`. Never use single-value form that panics.
- **Generics:** Use generics for type-safe collections and utilities; prefer interfaces for domain logic.

## Concurrency
- **Context:** Always pass `context.Context` as the first parameter to blocking/IO operations.
- **Context lifecycle:** Create cancellable contexts with `context.WithTimeout` or `context.WithCancel` at the entry point. Propagate the same context down the call chain -- never create a fresh `context.Background()` mid-call to bypass a parent's cancellation.
- **defer:** Use `defer` for resource cleanup. Be aware of loop and closure pitfalls (e.g., `defer` in a loop defers until function exit, not iteration end).

## Dependencies and Configuration
- **Structured logging:** Use `log/slog` (Go 1.21+) for structured logging. Pass `slog.Logger` via dependency injection, not package-level globals.
- **Linting:** Use `golangci-lint` with project config. Fix all warnings before committing.
- **Dependencies:** Use `go get` to add/update dependencies. Run `go mod tidy` after changes. Never manually edit `go.mod` or `go.sum`.
- **Configuration:** Use environment variables, config files, or functional options.
- **init():** Avoid `init()` functions -- they make testing difficult and create hidden dependencies. Document if truly unavoidable.
- **Global state:** Avoid package-level `var` for mutable state. Prefer dependency injection.
- **Build tags:** Use `//go:build` constraints for platform-specific code and to separate integration tests (`//go:build integration`).
