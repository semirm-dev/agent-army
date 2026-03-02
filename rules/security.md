---
name: security
description: Authentication, authorization, CORS, rate limiting, input sanitization, and secrets management
scope: universal
languages: []
---

# Security Patterns

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

## CORS
- **Whitelist specific origins.** Never use `*` in production.
- **Credentials:** Include `Access-Control-Allow-Credentials` only when needed (cookie-based auth).
- **Methods/Headers:** Whitelist only required methods and headers.
- **Preflight caching:** Set `Access-Control-Max-Age` to reduce OPTIONS requests.

## Rate Limiting
- **Per-IP and per-user on public endpoints.** Use sliding window for simplicity or token bucket for burst tolerance.
- **Return 429** with `Retry-After` header indicating when the client can retry.
- **Graduated limits:** Stricter for auth endpoints (login, signup, password reset).
- **Distributed:** Use distributed state for rate limiting across multiple instances.

## Input Sanitization
- **Validate at boundary** (handler/controller layer). Never trust client input past the handler.
- **Allowlists over denylists.** Define what is accepted, not what is rejected.
- **Strip HTML before storage** when rich text is not needed. Use a vetted sanitizer library when HTML is required.
- **Size limits:** Enforce maximum lengths on all string inputs, file uploads, and request bodies.

## Session Management
- **Regenerate session ID after login.** Prevents session fixation attacks.
- **Cookie flags:** `Secure` (HTTPS only), `HttpOnly` (no JS access), `SameSite=Strict` (or `Lax` if cross-site navigation needed).
- **Expiry:** Set reasonable session timeouts. Absolute timeout (max lifetime) + idle timeout (inactivity).
- **Server-side storage:** Store session data server-side. Cookie should contain only session ID.

## Secrets Management
- **Use a secrets manager** or environment variables. Never hardcode secrets in source code.
- **Never commit `.env` files.** Add `.env` to `.gitignore` in every project.
- **Rotation:** Secrets should be rotatable without application restart when possible.
- **Logging:** Never log secrets, tokens, passwords, or API keys. Mask in error messages.
- **CI/CD:** Use pipeline secret injection. Never hardcode secrets in pipeline configuration files.

## Transport Security
- **Enforce HTTPS everywhere.** No plaintext HTTP in production, including internal service-to-service calls.
- **HSTS:** Set `Strict-Transport-Security` header with a long max-age and `includeSubDomains`.
- **Minimum TLS version:** TLS 1.2 minimum, prefer TLS 1.3. Disable older protocols and weak cipher suites.
- **Certificate validation:** Always validate certificates. Never disable certificate checks, even in staging.
- **Certificate pinning:** Pin certificates or public keys for mobile clients calling known backends to prevent MITM attacks. Implement a rotation strategy to avoid bricking clients on certificate renewal. **Trade-off:** Failed rotation bricks clients until an app update ships. Evaluate whether your threat model justifies the operational risk.

## Injection & Output Safety
- **SQL injection:** Always use parameterized queries. Never string-concatenate user input into queries.
- **XSS:** Context-appropriate output encoding + strict CSP. Never insert untrusted data into raw HTML.
- **CSP:** Start with `default-src 'self'`. Add specific directives as needed. Use `report-uri` or `report-to` to detect violations before enforcing. Never use `unsafe-inline` or `unsafe-eval` in production.
- **CSRF:** Protect all state-changing endpoints with synchronizer tokens + SameSite cookies.
- **SSRF:** When the server fetches user-supplied URLs (webhooks, previews, imports), validate and restrict targets. Block private/internal IP ranges (127.0.0.0/8, 10.0.0.0/8, 172.16.0.0/12, 169.254.0.0/16). Allowlist permitted domains when possible.

## Dependency Security
- **Scan dependencies for known vulnerabilities** as part of CI. Block merges on critical/high severity CVEs.
- **Monitor for new CVEs** in deployed dependencies. New vulnerabilities appear after build time.
- **Minimize dependency count.** Every dependency is attack surface. Evaluate before adding.
