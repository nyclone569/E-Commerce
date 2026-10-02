# Feature 002: Register and Login

## User function

As a shopper, I can create an account, sign in, see that AuroraShop recognizes me, and sign out.

## Scope

This slice adds email/password registration, login, current-user lookup, logout, Argon2id password hashing, opaque database sessions, browser cookies, and exact-origin checks. Authorization rules for carts and orders arrive with those functions.

Email verification, password reset, MFA, OAuth/OIDC, social login, roles, account administration, Redis, JWT, idle timeout, remembered devices, and guest-session merging are out of scope.

## Acceptance criteria

1. A shopper can register with a unique normalized email, display name, and valid passphrase.
2. Successful registration creates the user and initial session atomically and sets an `HttpOnly` cookie only after commit.
3. Passwords are represented only as Argon2id encoded hashes; API responses and logs never contain passwords, cookies, or tokens.
4. Login accepts valid credentials and returns the safe user profile.
5. Unknown email and wrong password return the same `INVALID_CREDENTIALS` response.
6. A fresh session uses 32 random bytes, stores only its SHA-256 digest, and expires after the configured fixed TTL (24 hours by default).
7. `GET /api/me` returns the user for an active session and `401` for a missing, invalid, revoked, or expired session.
8. Logout revokes the current session, expires the cookie, and remains idempotent.
9. Register, login, and logout reject a present origin that does not exactly match the configured public origin.
10. Next.js renders register/login forms and the current-user state without using browser storage as authentication proof.
11. Multiple successful logins create independent sessions.
12. Unit and PostgreSQL integration tests cover validation, password verification, conflict, session expiry, revocation, and raw-token non-persistence.

## Business and security rules

- User IDs are UUID v4 values and are stable, non-email identifiers.
- Email is trimmed and case-folded for uniqueness in this learning project.
- Passphrases contain 15 to 128 Unicode code points and are normalized to NFC; no uppercase/digit/symbol composition rule is imposed.
- `users.status` starts as `active`; disabled-account behavior is reserved for a later administration function.
- Session expiry in PostgreSQL is authoritative; cookie expiry is a browser convenience.
- A valid identity never implies ownership of another user's resource.
- Authentication fails closed when PostgreSQL is unavailable.

## Decision Review

| Decision | Options | Selected | Why now | Main trade-off | Revisit trigger |
| --- | --- | --- | --- | --- | --- |
| Browser credential | Opaque session; JWT; managed IdP | Opaque database session | One browser and one Go backend need immediate revocation, not federation | Database lookup per authenticated request | Public clients or independently verifying services |
| Session storage | PostgreSQL; Redis | PostgreSQL | Existing durable dependency and low expected load | Adds reads to the primary database | Measured session load or independent availability target |
| Password hashing | Argon2id; scrypt; bcrypt; PBKDF2 | Argon2id | Memory-hard modern default with visible parameters | Deliberate CPU and memory cost | Guidance, compliance, or benchmark changes |
| Lifetime | Fixed; idle + absolute; access/refresh tokens | Fixed 24-hour absolute TTL | Simplest observable expiration semantics | Less convenient than sliding sessions | User feedback or security policy requires another model |
| Token rotation | Every request; periodic; security boundaries | Register/login boundaries only | Avoid multi-tab races before a requirement exists | Stolen token survives until revoke/expiry | Elevated privilege, recovery, or replay evidence |
| Frontend auth state | Browser storage; client global store; server-derived `/api/me` | Server-derived `/api/me` | Backend session remains the source of truth | Current layout becomes dynamic | Static caching becomes valuable |
| Browser routing | Cross-origin API; Next proxy; edge path routing | Same-origin path, Next proxy locally | Avoid CORS complexity and match future ingress paths | Extra local hop | Proxy becomes a bottleneck or topology changes |
| CSRF baseline | SameSite only; Origin; session-bound token | SameSite + exact Origin for Identity | Covers this narrow first slice explicitly | Cart adds a session-bound token | Authenticated commerce mutation |
| User ID | Sequence; UUID v4; UUID v7 | UUID v4 | Existing project convention and non-PII public identifier | Larger, less index-local than sequence/v7 | Index measurements justify change |

## Request and data flow

```text
Browser POST /api/auth/login
  -> Next.js same-origin proxy
  -> Go origin check
  -> identity handler
  -> identity service
  -> identity repository
  -> sqlc/pgx
  -> PostgreSQL users + sessions
  -> Set-Cookie response through the proxy
  -> Browser cookie jar

Browser GET /api/me with cookie
  -> Next.js same-origin proxy
  -> Go authentication middleware
  -> hash presented session token
  -> PostgreSQL active-session lookup
  -> safe user JSON
```

## Failure modes

- Duplicate email: the database uniqueness constraint wins and the API returns `409 IDENTITY_CONFLICT`.
- Invalid credentials: perform a real or dummy Argon2id verification and return the same generic `401` response.
- Invalid or expired session: return `401 AUTHENTICATION_REQUIRED` without revealing token state.
- PostgreSQL failure: return a standard internal error and never manufacture an authenticated identity.
- Registration transaction failure: roll back both user and session and do not set a cookie.
- Proxy failure: render a bounded frontend error and do not store a fake login flag.
- Foreign origin: return `403 ORIGIN_NOT_ALLOWED` before invoking the identity service.

## Metrics to add later

- Authentication request count, status, and bounded error code.
- Argon2 hash/verify duration and resource saturation.
- Session lookup latency and expired-session rate.
- Origin rejection and login-throttling rate.

## Learning outcome

This slice teaches the separation among credential verification, session lifecycle, browser state, and authorization. It also provides a concrete baseline against which a later JWT exercise can be measured instead of selecting JWT for its own sake.
