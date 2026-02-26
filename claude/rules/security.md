<!-- Sync: Must stay in sync with cursor/501-security.mdc -->

# 🔒 Security Patterns

## Password Hashing
- **bcrypt or argon2 only.** Never MD5, SHA-1, or SHA-256 for password storage.
- **Cost factor:** bcrypt cost >= 12. Argon2: memory >= 64MB, iterations >= 3, parallelism >= 1.
- **Upgrade strategy:** Re-hash on login if cost factor has increased since last hash.

## JWT & Token Management
- **Access tokens:** 15-minute expiry maximum. Short-lived, stateless.
- **Refresh tokens:** Opaque (not JWT), server-stored, rotated on every use. Revoke on logout.
- **Signing:** Use RS256 or ES256 for production. HS256 only for internal services with shared secrets.
- **Claims:** Include `sub`, `iat`, `exp`, `iss`. Never store sensitive data (passwords, PII) in JWT payload.
- **Validation:** Always verify signature, expiry, issuer, and audience. Reject tokens with `alg: none`.

## OAuth 2.0 / OIDC

- **Authorization Code + PKCE:** Default flow for all clients (web, mobile, SPA). Never use Implicit flow.
- **Client Credentials:** Service-to-service authentication only. Scope tokens to minimum required permissions.
- **OIDC Discovery:** Use `/.well-known/openid-configuration` to fetch provider endpoints. Never hardcode provider URLs.
- **Token storage:**
  - **Server-rendered web:** Store tokens in HTTP-only, Secure cookies.
  - **SPA:** Use Backend-for-Frontend (BFF) pattern — tokens stay server-side, session cookie to browser.
  - **Mobile:** Store in platform keychain (iOS Keychain, Android Keystore).
  - **Never store tokens in localStorage or sessionStorage.**
- **State parameter:** Always include `state` parameter to prevent CSRF. Validate on callback.
- **PKCE:** Use `S256` code challenge method. Never use `plain`.
- **Scopes:** Request minimum scopes needed. Use incremental consent for additional permissions.
- **ID token validation:** Verify `iss`, `aud`, `exp`, `nonce` (if used). Check `at_hash` when ID token accompanies access token.
- **Logout:** Implement both local session cleanup and provider logout (`end_session_endpoint`).

## Authorization
- **RBAC/ABAC at service layer.** Check permissions after authentication, never skip.
- **Never rely on client-side role checks.** Server must validate every request.
- **Principle of least privilege.** Default deny. Grant only required permissions.
- **Audit logging:** Log all permission changes, role assignments, and access denials.

## CORS
- **Whitelist specific origins.** Never use `*` in production.
- **Credentials:** Include `Access-Control-Allow-Credentials` only when needed (cookie-based auth).
- **Methods/Headers:** Whitelist only required methods and headers. Don't allow everything.
- **Preflight caching:** Set `Access-Control-Max-Age` to reduce OPTIONS requests.

## Rate Limiting
- **Per-IP and per-user on public endpoints.** Use sliding window algorithm.
- **Return 429** with `Retry-After` header indicating when the client can retry.
- **Graduated limits:** Stricter for auth endpoints (login, signup, password reset).
- **Distributed:** Use Redis or equivalent for rate limiting across multiple instances.

## Input Sanitization
- **Validate at boundary** (handler/controller layer). Never trust client input past the handler.
- **Allowlists over denylists.** Define what's accepted, not what's rejected.
- **Strip HTML before storage** when rich text is not needed. Use a sanitizer library (DOMPurify, bleach) when HTML is required.
- **Size limits:** Enforce maximum lengths on all string inputs, file uploads, and request bodies.

## Session Management
- **Regenerate session ID after login.** Prevents session fixation attacks.
- **Cookie flags:** `Secure` (HTTPS only), `HttpOnly` (no JS access), `SameSite=Strict` (or `Lax` if cross-site navigation needed).
- **Expiry:** Set reasonable session timeouts. Absolute timeout (max lifetime) + idle timeout (inactivity).
- **Server-side storage:** Store session data server-side. Cookie should contain only session ID.

## Secrets Management
- **Environment variables or secret manager** (Vault, AWS Secrets Manager, GCP Secret Manager).
- **Never commit `.env` files.** Add `.env` to `.gitignore` in every project.
- **Rotation:** Secrets should be rotatable without application restart when possible.
- **Logging:** Never log secrets, tokens, passwords, or API keys. Mask in error messages.
- **CI/CD:** Use pipeline secret injection (GitHub Actions secrets, GitLab CI variables). Never hardcode in pipeline files.
