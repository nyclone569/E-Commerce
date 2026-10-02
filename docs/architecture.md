# AuroraShop Architecture

## Purpose

AuroraShop is a learning system for building correct commerce flows before distributing them. Version 1 has two deployables: a Next.js frontend and one Go API modular monolith. The Go application owns one PostgreSQL database and exposes explicit domain-module boundaries.

```text
Browser
  -> Next.js server (frontend deployable)
      -> HTTP /api/* (Go deployable)
          -> catalog handler
              -> catalog service
                  -> catalog repository
                      -> sqlc-generated pgx queries
                          -> PostgreSQL
```

The `catalog`, `identity`, and `cart` packages have behavior. The empty `inventory`, `order`, and `payment` packages mark planned ownership; empty packages are not services and are not deployed independently.

## Functional map

AuroraShop uses a functional map to show business capabilities without deciding deployment topology prematurely. A box in this map is a capability or possible domain module, not automatically a microservice, database, Kafka topic, or team boundary.

```text
AuroraShop
├── Core commerce — current roadmap
│   ├── Identity
│   ├── Catalog
│   ├── Cart
│   ├── Inventory
│   ├── Order
│   └── Payment
│
├── Customer experience — post-roadmap expansion
│   ├── Favourites
│   ├── Ratings and reviews
│   ├── Product search
│   └── Product media
│
├── Commerce operations — post-roadmap expansion
│   ├── Promotions
│   ├── Tax
│   ├── Shipping and fulfilment
│   ├── Returns and refunds
│   └── Notifications
│
└── Merchant operations — post-roadmap expansion
    ├── Product administration
    ├── Order administration
    ├── Inventory adjustments
    └── Audit history
```

Only **Core commerce** is in the current Milestone 1–8 roadmap. The other branches are an expansion backlog and must not delay checkout correctness, observability, deployment learning, or the evidence-based extraction experiment. Their order is intentionally undecided.

When the current roadmap is complete, select one expansion capability by starting from a user journey and a measurable learning objective. Before implementation, write a new Decision Review covering its requirements, ownership, transaction boundary, security and privacy risks, expected load, failure behavior, and whether an existing module should own it. The default remains to add the capability as a module in the Go monolith and the existing PostgreSQL database. A separate deployable, datastore, cache, search engine, object store, or event flow requires its own evidence and ADR.

Examples of justified future infrastructure include object storage when Product media needs binary assets, a search engine when measured PostgreSQL search quality or latency is inadequate, and asynchronous delivery when Notifications must tolerate slow external providers. These examples are revisit conditions, not predetermined technology choices.

## Milestone 2 Identity flow

The browser uses one public origin. During local development, a Next.js route handler proxies `/api/*` to the internal Go URL. In Kubernetes, ingress will route `/api/*` directly to Go and other paths to Next.js.

```text
Browser cookie (raw opaque token)
  -> GET /api/me
  -> Go authentication middleware
  -> SHA-256 token digest
  -> identity repository
  -> PostgreSQL active session + user
  -> safe user response
```

Go owns credential validation, Argon2id password hashing, session generation, expiry, revocation, and authentication context. Next.js only renders forms/current-user state and forwards same-origin API traffic. PostgreSQL stores Argon2id records and session-token digests, never raw passwords or raw session tokens. The session uses a configurable fixed absolute TTL, initially 24 hours.

## Milestone 2 Cart flow

`GET /cart` is server-rendered by Next.js. It forwards the session cookie to `GET /api/cart`; Go authenticates the cookie before invoking Cart. Cart's handler takes the user ID only from Identity's request context. The Cart service loads SKU IDs and quantities through its repository and calls the public Catalog service for current VND product details and prices. No cart row is created by GET. On first add, a PostgreSQL transaction creates one active cart and atomically upserts the SKU quantity. Mutations require authentication, an exact Origin, and a session-bound CSRF header. See [Feature 003](features/003-authenticated-cart.md) and [ADR-0010](adr/0010-authenticated-cart.md).

## Module rules

1. Handlers translate HTTP concerns and never contain business rules.
2. Services validate input and coordinate domain behavior.
3. Repositories own persistence and expose domain-oriented operations.
4. A module may use its own generated queries but must not mutate another module's tables.
5. Cross-module work uses a public service method and an explicit transaction coordinator when atomicity spans modules.
6. Generated database records are mapped to domain models at the repository boundary.
7. Interfaces are added only at a substitution boundary; the catalog repository interface exists so service tests do not need PostgreSQL.

## Milestone 1 request flow

`GET /` is rendered by a Next.js Server Component. It calls `GET /api/products?page=1&page_size=12` with a five-second cancellation timeout. Chi routes the request, middleware assigns a request ID and records a sanitized JSON access log, the handler parses pagination, the service validates it, and the repository executes typed sqlc queries over a bounded pgx pool. PostgreSQL returns products ordered by `created_at DESC, id DESC`; SKUs are ordered by `product_id, code ASC`.

`GET /products/{slug}` follows the same server-rendered path through `GET /api/products/{slug}`. Catalog maps a missing row to a domain not-found error, the API returns 404, and Next.js renders the route-specific not-found state. Slug is the public lookup key; rename and redirect policy is intentionally deferred.

`POST /api/products` accepts a product and one to 100 SKUs. The service normalizes and validates it. The repository creates the product and every SKU in one PostgreSQL transaction. A duplicate slug or SKU code maps to HTTP 409. Any failed SKU insert rolls the whole transaction back.

## Health and lifecycle

- Liveness (`/api/health/live`) proves the HTTP process can respond; it deliberately does not query dependencies.
- Readiness (`/api/health/ready`) pings PostgreSQL with a two-second timeout. A failed dependency returns 503 so a scheduler can stop routing traffic.
- The API configures read-header, read, write, idle, and shutdown timeouts and drains on SIGINT/SIGTERM.
- Database pool size is explicit and configurable. Connection count must later be budgeted across replicas against PostgreSQL capacity.
- The API never runs migrations. Locally, an operator runs Goose. In GitOps, an Argo CD PreSync Job will run Goose once per sync.

## Data ownership

Current tables have one logical owner per module:

```text
catalog
  products (unique slug)
  skus (unique code, product FK, non-negative price)

identity
  users (unique normalized email)
  sessions (hashed token, expiry, revocation)

cart
  carts (one active cart per user)
  cart_items (SKU reference, bounded quantity)
```

PostgreSQL constraints are a final correctness boundary; application validation exists to produce useful errors. New products use VND integer minor units (one unit is one đồng). Existing non-VND demo SKUs remain readable after migration but cannot enter the Cart. Cart stores only SKU IDs and quantities, not prices; checkout must later reprice and validate inventory. The maximum new SKU price and Cart limits keep JSON numbers in JavaScript's safe integer range.

## Security and data handling baseline

Configuration comes from environment variables. The example local password is development-only. Logs never include headers, cookies, bodies, credentials, payment data, or tokens. Request IDs from clients are length-limited; trace IDs and authenticated user IDs arrive in later milestones. Production requires TLS at ingress, secret management, authentication, authorization, rate limiting, and hardened headers.

## Decision Review — Milestone 1

| Decision | Options | Selected | Why now | Main trade-off | Revisit trigger |
| --- | --- | --- | --- | --- | --- |
| Backend topology | Modular monolith; microservices | Modular monolith | One team and unproven boundaries | Shared deployment and fault domain | Independent scaling/deploy/security/data ownership is measured |
| HTTP stack | `net/http`; Chi; Gin; Echo | `net/http` + Chi | Small routing layer over standard APIs | Fewer batteries included | Repeated missing cross-cutting capability |
| Data access | pgx + sqlc; manual scan; GORM; Ent | pgx + sqlc | Keep SQL visible and typed | More SQL and mapping work | Query generation blocks required dynamic behavior |
| Migration tool | Goose; golang-migrate; Atlas; Flyway | Goose SQL | Go-native CLI and explicit SQL | No declarative schema planning | Multi-language governance or drift management needs it |
| Primary store | PostgreSQL; MySQL; document store; per-module DB | One PostgreSQL | Transactions and relational integrity fit commerce | Shared database can enable boundary leaks | Proven workload/data-isolation need |
| Inter-module processing | In-process sync; Kafka; queue | Synchronous | Few flows and easiest failure reasoning | Temporal coupling | Durable async use case and outbox exist |
| Delivery ownership | Jenkins deploys; Argo CD deploys; split CI/GitOps CD | Jenkins CI, Argo CD CD | Auditable desired state without direct cluster mutation | Two control systems/repositories later | GitOps latency or ownership becomes harmful |
| Logs | Text; JSON stdout; collector; direct Kafka | JSON stdout | Immediately machine-readable, no runtime dependency | No centralized search yet | Kubernetes deployment starts Milestone 4/6 collector work |
| Frontend data fetching | Server Component; TanStack Query | Server Component | Read-only initial page | No rich client cache | Client mutations/polling/optimistic updates appear |

## Ten-question review for the topology decision

1. **Problem:** deliver correct catalog and later checkout behavior without distributed-system failure modes.
2. **Constraints:** a learning project, one team, one database, explicit future extraction, and no measured independent-scaling need.
3. **Options:** a layered monolith, modular monolith, or independently deployed services.
4. **Large-company pattern (verified):** Shopify described componentizing a very large Rails monolith; Uber described moving to microservices after operational and organizational pressure, then grouping thousands of services into domains.
5. **Why scale justifies it:** hundreds or thousands of engineers need independent ownership/deployment and very large workloads may need different scaling profiles.
6. **Why copying is harmful:** network latency, partial failures, compatibility, distributed tracing, multiple deployment pipelines, and data consistency arrive before AuroraShop has evidence that they pay for themselves.
7. **Choice:** one Go modular monolith, explicit modules, one PostgreSQL database, synchronous calls.
8. **Benefits:** local transactions, fast feedback, simple debugging, and discoverable boundaries.
9. **Introduced failures:** one API deployment/fault domain, shared pool contention, and boundary erosion if imports/table access are not reviewed.
10. **Revisit evidence:** per-module saturation, deploy contention, separate security boundary, team ownership, fault-isolation target, or persistent p95/p99 bottleneck confirmed by load tests.

Facts above are attributed to their sources; applying them to AuroraShop is an architectural inference, not a claim that the companies endorse this design.

## References

- [Shopify Engineering: Under Deconstruction—The State of Shopify's Monolith](https://shopify.engineering/shopify-monolith)
- [Uber Engineering: Domain-Oriented Microservice Architecture](https://www.uber.com/blog/microservice-architecture/)
- [Go package `net/http`](https://pkg.go.dev/net/http)
- [PostgreSQL concurrency control](https://www.postgresql.org/docs/current/mvcc.html)
- [Argo CD CI automation and GitOps flow](https://argo-cd.readthedocs.io/en/stable/user-guide/ci_automation/)
