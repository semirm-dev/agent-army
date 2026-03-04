---
name: security-hardening
description: Security implementation workflows -- auth flow selection, JWT vs session decision, input validation checklist, secrets management, CORS decision tree, and rate limiting strategy.
scope: universal
languages: []
uses_rules: [security, cross-cutting, api-design]
---

# Security Hardening Skill

## When to Use

Invoke this skill when:
- Implementing authentication or authorization
- Handling user input at API boundaries
- Configuring CORS for a new service or endpoint
- Managing secrets (env vars, vaults, CI/CD injection)
- Adding rate limiting to public endpoints
- Conducting a security review of existing code

## Auth Implementation Decision Tree

```
What type of client needs authentication?
  |
  +-- Single-page application (SPA)?
  |     YES --> Use Authorization Code + PKCE with BFF pattern
  |
  +-- Mobile application?
  |     YES --> Use Authorization Code + PKCE (platform keychain storage)
  |
  +-- Server-to-server (no user context)?
  |     YES --> Use Client Credentials flow (minimum required scopes)
  |
  +-- Server-rendered web application?
        YES --> Use Authorization Code + PKCE (HTTP-only cookies)
```

### JWT vs Session Decision

```
Do you need stateless auth across multiple services?
  YES --> JWT (access token)
          Pair with server-stored refresh token.
  NO  --> Server-side sessions
          Session ID in cookie, data server-side.
```

## Input Validation Workflow

### Boundary Validation Checklist

Run through this checklist for every handler/controller that accepts external input:

1. [ ] Validate at the handler boundary -- never trust input past this layer
2. [ ] Apply input sanitization (allowlists, size limits, type/range/format validation)
3. [ ] Reject unexpected fields -- do not silently accept unknown keys
4. [ ] Return 400 with specific field errors -- tell the caller what failed

## Secrets Management Checklist

1. [ ] No hardcoded secrets; all secrets loaded from environment or secret manager
2. [ ] Different secrets per environment (dev, staging, production) -- never share across environments
3. [ ] Access to production secrets restricted to minimum required personnel
4. [ ] `.env` added to `.gitignore` -- never commit environment files with real credentials
5. [ ] Secrets are never logged, printed, or included in error messages -- mask in all output
6. [ ] CI/CD uses pipeline secret injection (GitHub Actions secrets, GitLab CI variables) -- never hardcoded in pipeline files
7. [ ] Rotation strategy defined -- secrets should be rotatable without application restart when possible

### Environment Variables vs Secret Manager

```
Is this a single-service deployment or local dev?
  YES --> Environment variables are sufficient
  |       Load via a validated config module, not raw environment variable access
  |
  NO --> Multiple services sharing secrets?
    YES --> Use a secrets manager (Vault, AWS SM, GCP SM)
    |       - Centralized rotation
    |       - Audit trail for access
    |       - Fine-grained access policies
    |
    NO --> Environment variables with per-service isolation
            Consider secrets manager when the team or service count grows
```

## CORS Configuration Guide

### Decision Tree

```
Does your frontend and API share the same origin (scheme + host + port)?
  YES --> No CORS configuration needed
          Browser allows same-origin requests without CORS headers
  |
  NO --> Cross-origin setup required
          |
          Is the API public (called by third-party clients)?
            YES --> Whitelist known partner origins
            |       Consider a CORS proxy or API gateway for unknown callers
            |
            NO --> Whitelist only your own frontend origins
                    Use environment-specific allowlists
```

## Rate Limiting Setup

### Strategy Decision

```
What type of endpoint?
  |
  +-- Auth endpoint (login, signup, password reset)?
  |     Apply strict limits: 5-10 requests/minute per IP
  |     Use progressive delays or lockouts after repeated failures
  |
  +-- Public API endpoint?
  |     Apply per-IP AND per-user limits
  |     Use sliding window algorithm
  |     Return 429 + Retry-After header
  |
  +-- Internal/service-to-service?
        Use per-service rate limits based on expected throughput
        Monitor but start permissive
```

## Session Security

### Cookie Configuration

Set all session cookies with these flags:
- **`Secure`:** HTTPS only — prevents transmission over plain HTTP
- **`HttpOnly`:** No JavaScript access — prevents XSS from stealing session IDs
- **`SameSite`:** `Strict` (default) or `Lax` (if cross-site navigation is needed) — prevents CSRF

### Session Lifecycle

- **Regenerate session ID on login.** Prevents session fixation attacks where an attacker sets a known session ID before authentication.
- **Set two timeouts:** Idle timeout (inactivity, e.g., 30 minutes) and absolute timeout (max lifetime, e.g., 24 hours). Both are required.
- **Store session data server-side.** The cookie should contain only the session ID, never session data.

Implement input validation, rate limiting, auth middleware, and session management using your project's framework conventions.
