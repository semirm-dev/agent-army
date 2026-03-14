---
name: security
description: Enforce authentication, authorization, input sanitization, secrets management, transport security, injection prevention, data lifecycle controls, and guide auth flow selection, CORS configuration, rate limiting setup, and session hardening.
scope: universal
languages: []
uses_skills: []
---

# Security Patterns

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

## Password Hashing
- **Use modern adaptive hashing algorithms** (bcrypt, argon2id, scrypt). Never use general-purpose hash functions (MD5, SHA-1, SHA-256) for password storage. PBKDF2 is acceptable only with SHA-256 and >= 600,000 iterations per NIST SP 800-63B.
- **Tune cost parameters** to the maximum your hardware can sustain within acceptable login latency (~250ms).
- **Upgrade strategy:** Re-hash on login if the cost factor has increased since the last hash.

## JWT & Token Management
- **Access tokens:** 15-minute expiry maximum. Short-lived, stateless.
- **Refresh tokens:** Opaque (not JWT), server-stored, rotated on every use. Revoke on logout.
- **Signing:** Use asymmetric algorithms (RS256, ES256) for production. Symmetric signing only for internal services with shared secrets.
- **Claims:** Include `sub`, `iat`, `exp`, `iss`. Never store sensitive data (passwords, PII) in JWT payload.
- **Validation:** Always verify signature, expiry, issuer, and audience. Reject tokens with `alg: none`.

## OAuth 2.0 / OIDC
- **Authorization Code + PKCE:** Default flow for all clients. Never use Implicit flow.
- **Never store tokens in localStorage/sessionStorage.** Use HTTP-only cookies (server-rendered) or BFF pattern (SPA).
- **Request minimum scopes needed.** Use incremental consent.
- **Implement both local session cleanup and provider logout.**

## Authorization
- **RBAC/ABAC at service layer.** Check permissions after authentication, never skip.
- **Never rely on client-side role checks.** Server must validate every request.
- **Principle of least privilege.** Default deny. Grant only required permissions.
- **Audit logging:** Log all permission changes, role assignments, and access denials.

## Input Validation

- **Allowlists over denylists.** Define what is accepted, not what is rejected.
- **Strip HTML before storage** when rich text is not needed. Use a vetted sanitizer library when HTML is required.
- **Size limits:** Enforce maximum lengths on all string inputs, file uploads, and request bodies.

### Boundary Validation Checklist

Run through this checklist for every handler/controller that accepts external input:

1. [ ] Validate at the handler boundary -- never trust input past this layer
2. [ ] Apply input sanitization (allowlists, size limits, type/range/format validation)
3. [ ] Reject unexpected fields -- do not silently accept unknown keys
4. [ ] Return 400 with specific field errors -- tell the caller what failed

## CORS Configuration

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

- **Credentials:** Include `Access-Control-Allow-Credentials` only when needed (cookie-based auth).
- **Methods/Headers:** Whitelist only required methods and headers.
- **Preflight caching:** Set `Access-Control-Max-Age` to reduce OPTIONS requests.

## Rate Limiting

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

- **Distributed:** Use distributed state for rate limiting across multiple instances.

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

## Secrets Management

- **Use a secrets manager** or environment variables. Never hardcode secrets in source code.
- **Never commit `.env` files.** Add `.env` to `.gitignore` in every project.
- **Rotation:** Secrets should be rotatable without application restart when possible.
- **Logging:** Never log secrets, tokens, passwords, or API keys. Mask in error messages.
- **CI/CD:** Use pipeline secret injection. Never hardcode secrets in pipeline configuration files.
- **Different secrets per environment** (dev, staging, production) — never share across environments.
- **Access to production secrets** restricted to minimum required personnel.

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

## Transport Security
- **Enforce HTTPS everywhere.** No plaintext HTTP in production, including internal service-to-service calls.
- **HSTS:** Set `Strict-Transport-Security` header with a long max-age and `includeSubDomains`.
- **Minimum TLS version:** TLS 1.2 minimum, prefer TLS 1.3. Disable older protocols and weak cipher suites.
- **Certificate validation:** Always validate certificates. Never disable certificate checks, even in staging.
- **Certificate pinning:** Pin certificates or public keys for mobile clients calling known backends to prevent MITM attacks. Implement a rotation strategy. Trade-off: failed rotation bricks clients -- evaluate whether your threat model justifies the operational risk.

## Injection & Output Safety
- **SQL injection:** Always use parameterized queries. Never string-concatenate user input into queries.
- **XSS:** Context-appropriate output encoding + strict CSP. Never insert untrusted data into raw HTML.
- **CSP:** Start with `default-src 'self'`. Add specific directives as needed. Use `report-uri` or `report-to` to detect violations before enforcing. Never use `unsafe-inline` or `unsafe-eval` in production.
- **CSRF:** Protect all state-changing endpoints with synchronizer tokens + SameSite cookies.
- **SSRF:** When the server fetches user-supplied URLs (webhooks, previews, imports), validate and restrict targets. Block private/internal IP ranges (127.0.0.0/8, 10.0.0.0/8, 172.16.0.0/12, 169.254.0.0/16). Allowlist permitted domains when possible.

## Data Lifecycle
- Classify data by sensitivity: public, internal, confidential, restricted. Apply controls proportional to classification.
- **Retention policies:** Define per data class. Automate enforcement -- scheduled jobs to purge or archive expired data. Never retain data indefinitely without justification.
- **PII handling:** Identify all fields containing personally identifiable information. Support deletion and anonymization requests (GDPR right-to-erasure, CCPA). Track PII across replicas, caches, backups, and logs.
- **Data deletion:** Soft-delete first (recoverable), hard-delete after retention window. Verify deletion propagates to derived stores (caches, search indexes, analytics pipelines).
- **Audit trail:** Log data access and mutations for confidential and restricted data. Include who, what, when, and from where.

## Dependency Security
- **Scan dependencies for known vulnerabilities** as part of CI. Block merges on critical/high severity CVEs.
- **Monitor for new CVEs** in deployed dependencies. New vulnerabilities appear after build time.
- **Minimize dependency count.** Every dependency is attack surface. Evaluate before adding.
