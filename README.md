# AuroraShop

AuroraShop is a self-learning ecommerce system that starts as a Go modular monolith and evolves only when measurements justify more infrastructure. Milestone 1 implements the catalog foundation; the first Milestone 2 vertical slice adds register, login, current-user, and logout behavior. Kafka is intentionally absent from the running stack.

## Current architecture

```text
Browser -> Next.js App Router (deployable) -> Go net/http + Chi (modular monolith) -> PostgreSQL
              local /api proxy                 catalog  -> service -> repository -> sqlc/pgx
                                                identity -> service -> repository -> sqlc/pgx
```

The Go process contains identity, catalog, cart, inventory, order, and payment domain packages. Catalog and Identity are implemented; the remaining packages are planned module boundaries, not separately deployed services. All modules use one PostgreSQL database in version 1, with logical table ownership and explicit service boundaries.

Read [architecture.md](docs/architecture.md), the [roadmap](docs/roadmap.md), and the [ADRs](docs/adr/) before adding infrastructure.

## Prerequisites

- Go 1.25.7 or newer (required by the pinned integration-test toolchain)
- Node.js 24 and `npx` (pnpm is pinned and can run through `npx`)
- Docker with Compose
- Git for generation verification in CI

No global sqlc or Goose installation is required; the Makefile runs pinned Go tool versions.

## Quick start

The API does not run migrations. Start PostgreSQL, migrate once, then start the applications:

```bash
docker compose up -d postgres
make migrate-up
docker compose up --build backend frontend
```

Open <http://localhost:3000>. Check health independently:

```bash
curl -i http://localhost:8080/api/health/live
curl -i http://localhost:8080/api/health/ready
```

Liveness can remain healthy while readiness is 503 if PostgreSQL or the required schema is unavailable.

## Create a product

```bash
curl --fail-with-body \
  -X POST http://localhost:8080/api/products \
  -H 'Content-Type: application/json' \
  -H 'X-Request-ID: tutorial-create-001' \
  --data '{
    "name": "Aurora Camp Mug",
    "slug": "aurora-camp-mug",
    "description": "A durable mug for slow mornings.",
    "skus": [
      {"code": "MUG-MOSS-12OZ", "price_cents": 2499, "currency": "USD"}
    ]
  }'
```

List products with deterministic newest-first ordering:

```bash
curl --fail-with-body 'http://localhost:8080/api/products?page=1&page_size=20'
```

Read a product and its code-ordered SKUs by public slug:

```bash
curl --fail-with-body 'http://localhost:8080/api/products/aurora-camp-mug'
```

Product cards link to the server-rendered `/products/{slug}` detail page. The function decisions and acceptance criteria are documented in [Feature 001](docs/features/001-product-details.md).

## Register and login

Use the browser at <http://localhost:3000/register>, or exercise the same-origin API without printing the session token:

```bash
cookie_jar="$(mktemp)"
trap 'rm -f "$cookie_jar"' EXIT

curl --fail-with-body \
  -c "$cookie_jar" \
  -X POST http://localhost:3000/api/auth/register \
  -H 'Origin: http://localhost:3000' \
  -H 'Content-Type: application/json' \
  --data '{
    "email": "learner@example.com",
    "password": "correct horse battery staple",
    "display_name": "Aurora Learner"
  }'

curl --fail-with-body -b "$cookie_jar" http://localhost:3000/api/me

curl --fail-with-body \
  -b "$cookie_jar" \
  -X POST http://localhost:3000/api/auth/logout \
  -H 'Origin: http://localhost:3000'
```

The browser stores the raw opaque token in an `HttpOnly` cookie; PostgreSQL stores only its SHA-256 digest. Sessions have a fixed configurable 24-hour lifetime. Read [Feature 002](docs/features/002-register-and-login.md) and [ADR-0009](docs/adr/0009-browser-authentication-and-sessions.md) for the password, session, JWT, CSRF, and routing trade-offs.

Error responses share one envelope:

```json
{
  "error": {
    "code": "validation_failed",
    "message": "Request validation failed",
    "details": { "slug": "must be lowercase words separated by single hyphens" }
  }
}
```

## Local development without application containers

```bash
docker compose up -d postgres
make migrate-up

cd backend
DATABASE_URL='postgres://aurora:aurora@localhost:5432/aurora?sslmode=disable' go run ./cmd/api
```

In another shell:

```bash
npx --yes pnpm@11.23.0 --dir frontend install --frozen-lockfile
BACKEND_URL=http://localhost:8080 npx --yes pnpm@11.23.0 --dir frontend dev
```

## Database and generated code

- Migrations: `backend/db/migrations/00001_create_catalog.sql` and `00002_create_identity.sql`
- Human-authored queries: `backend/db/queries/catalog.sql` and `identity.sql`
- Generated Go: `backend/internal/db/`
- Generator config: `backend/sqlc.yaml`

After changing migrations or queries:

```bash
make sqlc
```

Commit generated files. CI regenerates them and rejects a diff. For production, use forward-compatible expand–migrate–contract changes. Do not rely on a down migration as an application rollback strategy.

## Validation

```bash
make backend-check
make backend-integration
make frontend-check
```

`backend-integration` uses Testcontainers and requires a healthy Docker provider. It starts isolated PostgreSQL 17 containers, runs the real Goose migrations, and exercises catalog behavior plus Identity transaction rollback, password representation, session expiry, and revocation.

## Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `APP_ENV` | `development` | Structured-log environment |
| `APP_VERSION` | `dev` | Immutable deployment version/Git SHA in real environments |
| `HTTP_ADDRESS` | `:8080` | API listen address |
| `DATABASE_URL` | required | PostgreSQL DSN; supply via secret outside local development |
| `DATABASE_MAX_OPEN_CONNS` | `10` | pgx maximum connections per API replica |
| `DATABASE_MIN_IDLE_CONNS` | `2` | pgx warm idle connections per replica |
| `DATABASE_CONNECT_TIMEOUT` | `5s` | Startup connection timeout |
| `BACKEND_URL` | `http://localhost:8080` | Server-side Next.js API URL |
| `PUBLIC_ORIGIN` | `http://localhost:3000` | Exact browser origin accepted on authentication mutations |
| `SESSION_TTL` | `24h` | Fixed absolute session lifetime; maximum `720h` |
| `SESSION_COOKIE_SECURE` | environment-dependent | Require HTTPS for the session cookie; explicitly `false` only for local HTTP |

Never commit `.env`. The defaults and `.env.example` are for local learning only.

## CI/CD boundary

The Jenkinsfile currently validates backend/frontend code, runs PostgreSQL integration tests, checks sqlc generation, and builds images tagged with the immutable `GIT_COMMIT`. Registry push is not enabled because no registry or credential identifier has been authorized/configured. When configured, Jenkins will push the SHA tags and propose a GitOps manifest change; it will not use `latest` or deploy with `kubectl apply`.

Argo CD and Helm are planned for Milestone 6. Argo CD will own cluster reconciliation and run Goose through one PreSync Job. Production GitOps configuration will eventually live in a separate repository, but this project does not initialize or modify one.

## Logging and data safety

The Go API writes JSON with `log/slog` to stdout. Current access records include timestamp, level, message, service, environment, version, request ID, route, method, status, and duration. Do not add authorization headers, cookies, passwords, access tokens, payment card data, PII, or raw bodies.

Logs are operational records, metrics are numeric trends, traces follow request paths, domain events state business facts, and audit events record sensitive transitions. They will not share one Kafka topic.

## Kafka is not in Milestone 1

Milestone 5 adds Kafka only after synchronous commerce, correctness tests, and observability. The exact design is in [roadmap.md](docs/roadmap.md) and [ADR-0006](docs/adr/0006-synchronous-before-kafka.md): PostgreSQL transactional outbox, separate publisher, versioned JSON envelope, `order_id` partition keys, idempotent consumers, lag/retry/dead-letter experiments, and an optional collector-to-Kafka log path. The application never directly and synchronously publishes every log to Kafka.

## Repository map

```text
frontend/                 Next.js deployable
backend/cmd/api/          Go API entry point
backend/cmd/worker/       Reserved for the Milestone 5 outbox publisher
backend/internal/catalog/ Catalog handler/service/repository/domain model
backend/internal/identity/Identity password/session/handler/service/repository behavior
backend/internal/platform/Configuration, database, HTTP, logging
backend/db/               Goose migrations and sqlc queries
backend/tests/integration/PostgreSQL Testcontainers tests
docs/adr/                 Architecture decisions
docs/runbooks/            Operational procedures
deployments/              Placeholders for later Helm/observability/Kafka work
compose.yaml              PostgreSQL, backend, frontend only
Jenkinsfile               CI foundation; no direct deployment
```

## Troubleshooting

See [local-development.md](docs/runbooks/local-development.md) for diagnosis steps and safe reset instructions.
