# ADR-0009: Use Database-Backed Browser Sessions Before JWT

## Status

Accepted for the first Milestone 2 Identity slice.

## Context

AuroraShop needs register, login, current-user, and logout functions before carts and orders can be owned by a shopper. The current system has one browser frontend, one Go modular monolith, and one PostgreSQL database. Authentication must not introduce a distributed identity platform before the application has a client or service boundary that needs one.

## Forces and constraints

- The browser must not expose reusable credentials to client-side JavaScript.
- Sessions must expire, be revocable, and support independent logins on multiple devices.
- Passwords must never be stored, logged, or recoverable as plaintext.
- Next.js and Go remain separate deployables; Go owns identity rules and data.
- Local development uses HTTP, while production requires TLS.
- The first implementation should make session lifecycle and failure behavior visible to a learner.
- Authentication does not replace authorization; resource ownership remains a backend check.

## Options considered

1. Opaque session token in an `HttpOnly` cookie with the session stored in PostgreSQL.
2. Self-contained JWT in an `HttpOnly` cookie.
3. JWT access token plus refresh token stored and managed by the browser.
4. A managed identity provider such as Amazon Cognito, Auth0, or Firebase Authentication.

## Real-world usage examples

Verified fact: RFC 7519 defines JWT as a compact claims representation, not as a complete login, logout, refresh, or authorization system. Verified fact: RFC 8725 requires JWT consumers to constrain algorithms and validate issuer and audience in applicable deployments. Verified fact: Firebase supports both JWT ID tokens and server-side session cookies for traditional websites. Verified fact: Next.js documents both stateless and database sessions and recommends server-set cookies with `HttpOnly`, `Secure`, `SameSite`, expiry, and path attributes.

Architectural inference: independently deployed resource servers and third-party clients can justify self-verifiable tokens at large scale. AuroraShop has neither requirement yet, so copying that token topology would add key rotation, claim validation, refresh, and revocation behavior before it produces a measurable benefit.

## Decision

Use a cryptographically random 256-bit opaque session token in the `aurora_session` cookie. Store only a SHA-256 digest of that token in PostgreSQL. Use a configurable fixed absolute lifetime with a 24-hour default. Create a fresh token after register and login, revoke it on logout, and do not rotate it periodically in this slice.

Hash passwords with Argon2id using a unique random salt and an encoded record containing the algorithm version and cost parameters. Require 15 to 128 Unicode code points, normalize to NFC before hashing, and do not impose character-composition rules. A breached-password blocklist, rate limiting, password reset, email verification, MFA, and recovery are required before production but are separate learning slices.

Expose the browser API under the same public origin. Locally, a Next.js route handler proxies `/api/*` to Go; later, an ingress routes `/api/*` directly to Go and all other paths to Next.js. Go remains the authentication authority.

Protect the initial state-changing authentication endpoints with `SameSite=Lax` and exact `Origin` validation. The Cart slice adds a session-bound HMAC CSRF token, fetched through an authenticated, non-cacheable endpoint, for authenticated Cart mutations. Checkout must use the same or stronger protection. Production cookies use `Secure`; local HTTP explicitly disables it.

## Detailed reasoning

A database session makes revocation and expiry explicit: the API hashes the presented token and looks up an active record. A database breach does not immediately reveal a bearer token because the raw token exists only in the browser. PostgreSQL is already an operational dependency and the expected load does not justify Redis.

Short JWT access-token lifetimes are often paired with refresh tokens because the resource server cannot otherwise revoke a self-contained token cheaply. A PostgreSQL session can be revoked immediately, so copying that mechanism would add two credential lifecycles without solving a current problem.

Fixed expiry is selected before sliding expiry. Updating `last_seen_at` on every request creates write amplification; throttled refresh adds concurrent-tab and response-order races. A 24-hour absolute lifetime is easy to test and can be changed through configuration after usability evidence exists.

## Positive consequences

- Immediate logout and server-side revocation.
- No signing-key distribution or refresh-token protocol.
- Raw session tokens are absent from PostgreSQL and logs.
- Independent sessions per device are represented directly.
- Authentication behavior remains inside the Identity module.
- Same-origin browser calls avoid routine CORS complexity.

## Negative consequences

- Authenticated requests perform a PostgreSQL lookup.
- A PostgreSQL outage prevents authentication and fails closed.
- Rendering current-user state in the root layout makes the current pages dynamic.
- The local Next.js proxy is an additional network hop that production ingress later removes.
- The implementation is not production-ready without throttling and compromised-password checks.

## Failure modes introduced

- Session-table or pool pressure can increase latency for every authenticated request.
- A stolen raw cookie remains usable until expiry or revocation.
- Incorrect cookie attributes can expose or suppress the session.
- Incorrect proxy forwarding can drop `Cookie`, `Origin`, or `Set-Cookie` headers.
- Aggressive Argon2 parameters can exhaust CPU or memory during a login flood.
- User enumeration and timing differences can leak account existence if errors or verification paths diverge.
- Clock errors can expire sessions too early or too late.

## Operational requirements

- TLS and `Secure` cookies outside local development.
- A secret-safe log policy that excludes passwords, cookies, token values, and hashes.
- Bounded request bodies and password lengths.
- Login throttling by account and network signal before internet exposure.
- Periodic cleanup of expired/revoked sessions.
- Time synchronization and UTC timestamps.
- Password-hash parameters benchmarked on production-class hardware.
- Generic invalid-credential responses and a dummy hash check for unknown accounts.

## Metrics to observe

- Register/login success and failure rates by bounded error code.
- Login and Argon2 verification duration.
- Active, expired, and revoked session counts.
- Session lookup p95/p99 latency and database pool wait time.
- `401` and origin-rejection rates.
- Authentication database errors.
- Memory and CPU saturation during login load tests.

Do not label metrics with email, raw user ID, session ID, or token.

## Revisit triggers

- A mobile or third-party client needs delegated API access.
- Independently deployed services need local token verification.
- A managed identity provider is required for social login, enterprise federation, or MFA.
- Session lookups measurably saturate PostgreSQL.
- Users require a longer login with an idle timeout or explicit “remember me” policy.
- Security requirements demand phishing-resistant authentication such as WebAuthn.
- Multiple public origins make the same-origin routing model impractical.

## References

- [RFC 7519: JSON Web Token](https://www.rfc-editor.org/rfc/rfc7519.html)
- [RFC 8725: JWT Best Current Practices](https://www.rfc-editor.org/info/rfc8725/)
- [NIST SP 800-63B: Password Authenticators](https://pages.nist.gov/800-63-4/sp800-63b.html#passwordver)
- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [OWASP Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)
- [OWASP CSRF Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html)
- [Next.js Authentication Guide](https://nextjs.org/docs/app/guides/authentication)
- [Firebase Session Cookies](https://firebase.google.com/docs/auth/admin/manage-cookies)
