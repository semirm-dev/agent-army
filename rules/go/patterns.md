---
name: go-patterns
description: Go coding conventions, error handling, project structure, and concurrency
scope: language-specific
languages: [go]
extends: [code-quality]
uses_rules: [security, cross-cutting, observability, testing-patterns]
---

# Go Coding Patterns
- **Linting:** Use `golangci-lint` with project config. Fix all warnings before committing.
- **Packages:** Avoid "stuttering." Use `auth.Service` instead of `auth.AuthService`.
- **Error Handling:** ALWAYS wrap errors with context: `fmt.Errorf("domain: operation: %w", err)`.
  - Use `errors.Is` and `errors.As` for checking error types.
- **Interfaces:** "Accept interfaces, return concrete types." Keep interfaces small (2-3 methods max).
- **Project structure:** Follow vertical-slices architecture (feature + hexagonal/clean), package by feature. Follow Golang best practices.
- **Naming:** Use `MixedCaps` (Acronyms like `ID`, `HTTP`, `URL` should be consistent case).
- **Context:** Always pass `context.Context` as the first parameter to blocking/IO operations.
- **Panics:** Never use `panic()` for normal error paths. Reserve for truly unrecoverable situations.
- **Configuration:** Use environment variables, config files, or functional options.
- **Godoc:** All exported types, functions, and methods must have a godoc comment starting with the identifier name.
- **Dependencies:** Use `go get` to add/update dependencies. Run `go mod tidy` after changes. Never manually edit `go.mod` or `go.sum`.
- **init():** Avoid `init()` functions -- they make testing difficult and create hidden dependencies. Document if truly unavoidable.
- **Global state:** Avoid package-level `var` for mutable state. Prefer dependency injection.
- **Type assertions:** Always use the two-value form: `v, ok := x.(Type)`. Never use single-value form that panics.
- **Generics:** Use generics for type-safe collections and utilities; prefer interfaces for domain logic.
- **defer:** Use `defer` for resource cleanup. Be aware of loop and closure pitfalls (e.g., `defer` in a loop defers until function exit, not iteration end).
