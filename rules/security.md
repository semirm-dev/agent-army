---
scope: universal
languages: []
---

# Security Patterns

## Password Hashing
- **Use modern adaptive hashing algorithms.** Never use general-purpose hash functions for password storage.
- **Tune cost parameters** to the maximum your hardware can sustain within acceptable login latency (~250ms).
- **Upgrade strategy:** Re-hash on login if the cost factor has increased since the last hash.

## JWT & Token Management
- **Access tokens:** 15-minute expiry maximum. Short-lived, stateless.
- **Refresh tokens:** Opaque (not JWT), server-stored, rotated on every use. Revoke on logout.
- **Signing:** Use asymmetric algorithms (RS256, ES256) for production. Symmetric signing only for internal services with shared secrets.
- **Claims:** Include `sub`, `iat`, `exp`, `iss`. Never store sensitive data (passwords, PII) in JWT payload.
- **Validation:** Always verify signature, expiry, issuer, and audience. Reject tokens with `alg: none`.

## OAuth 2.0 / OIDC

### Authorization Flows

- **Authorization Code + PKCE:** Default flow for all client types (web, mobile, SPA). Never use the Implicit flow.
- **PKCE:** Use `S256` code challenge method. Never use `plain`.
- **Client Credentials:** Service-to-service authentication only. Scope tokens to the minimum required permissions.

### OIDC Discovery

- **Use the provider's discovery endpoint** (`/.well-known/openid-configuration`) to fetch all endpoints dynamically.
- **Never hardcode provider URLs.** Discovery ensures resilience to provider infrastructure changes.

### Token Storage by Client Type

- **Server-rendered web:** Store tokens in HTTP-only, Secure cookies.
- **Single-page applications:** Use the Backend-for-Frontend (BFF) pattern. Tokens stay server-side; the browser receives only a session cookie.
- **Mobile applications:** Store in the platform's secure credential storage.
- **Never store tokens in browser-accessible storage** (localStorage, sessionStorage).

### CSRF and State

- **Always include a `state` parameter** in authorization requests. Validate it on the callback to prevent CSRF.
- **Bind state to the user's session** so it cannot be replayed across sessions.

### Scopes and Consent

- **Request minimum scopes needed.** Avoid requesting broad permissions upfront.
- **Use incremental consent** to request additional permissions only when the user needs the feature that requires them.

### ID Token Validation

- **Verify all required claims:** `iss`, `aud`, `exp`, `nonce` (if used in the request).
- **Check `at_hash`** when the ID token accompanies an access token to ensure token binding.

### Logout

- **Implement both local session cleanup and provider logout** via the provider's `end_session_endpoint`.
- **Revoke refresh tokens** on logout to prevent token reuse.

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
- **Per-IP and per-user on public endpoints.** Choose an appropriate algorithm for your use case.
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
- **Certificate pinning:** Pin certificates or public keys for mobile clients calling known backends to prevent MITM attacks. Implement a rotation strategy to avoid bricking clients on certificate renewal.

## SQL Injection Prevention
- **Always use parameterized queries.** Never string-concatenate user input into SQL.
- **ORM/query builder:** Use parameterized execution. Verify generated SQL during development.
- **Stored procedures** do not guarantee safety. Parameterize inputs to stored procedures the same way.
- See `database.md` "Query Safety" for additional query hygiene rules.

## XSS Prevention
- **Context-appropriate output encoding.** Encode for the correct context: HTML body, HTML attributes, JavaScript, URL parameters, CSS.
- **Content-Security-Policy header:** Define strict CSP. Avoid `unsafe-inline` and `unsafe-eval` where possible.
- **Never insert untrusted data into raw HTML** — use the framework's built-in escaping mechanisms.
- **Sanitize rich text input** with a vetted library before rendering.

## CSRF Protection
- **Protect all state-changing endpoints** (POST, PUT, PATCH, DELETE).
- **Synchronizer token pattern:** Generate a unique token per session, validate on every state-changing request.
- **SameSite cookies** provide defense-in-depth but are not sufficient alone in all browsers. Pair with token-based protection.
- **Custom request headers** (e.g., `X-Requested-With`) add an extra layer for AJAX-only endpoints, since browsers enforce CORS on custom headers.
