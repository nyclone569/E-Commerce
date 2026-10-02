# AuroraShop Roadmap

## Milestone 1 — Foundation (implemented)

- Modular-monolith module map and catalog vertical slice
- Go `net/http` + Chi API, pgx pool, sqlc queries, Goose SQL migration
- Product creation and deterministic paginated listing
- Unit and PostgreSQL Testcontainers integration tests
- Next.js App Router product list with loading, empty, and error states
- Docker/Compose, Jenkins CI foundation, ADRs, and runbook
- Explicit exclusion: Kafka, Kubernetes runtime, authentication, and checkout

## Milestone 2 — Core synchronous commerce (in progress)

Implement identity, cart, inventory, order, and internal mock payment. Write module ownership contracts first. Design one-database transaction boundaries for cart-to-order and inventory deduction. Never call a slow payment adapter while holding a database transaction. Add authentication and authorization threat review.

Implementation order is feature-first: Identity register/login/session, Cart, Inventory, Order checkout, Mock Payment, then Order History. Identity and authenticated VND Cart are implemented; Inventory, Order checkout, Mock Payment, and Order History remain. Milestone 2 is not complete.

Exit criteria: happy-path commerce flows, explicit state transitions, module dependency tests, integration tests, and no cross-module table mutation.

## Milestone 3 — Correctness under failure

Add `Idempotency-Key`, atomic stock deduction, concurrent PostgreSQL tests, reservation expiry, retry semantics, order/payment audit transitions, and failure injection. Practice expand–migrate–contract and application rollback without database rollback.

Exit criteria: retry creates no duplicate order; inventory never goes negative under concurrent checkout; state-machine invariants are tested.

## Milestone 4 — Observability

Add OpenTelemetry traces and metrics, propagate request/trace/correlation IDs, deploy a collector, define SLOs and alerting, and write incident runbooks. Keep logs, metrics, traces, domain events, and audit events semantically separate.

Exit criteria: a failed checkout can be followed end-to-end and service health is visible without reading raw logs manually.

## Milestone 5 — Kafka learning track

Kafka is introduced only after synchronous transactions work:

1. Add Kafka to the local lab Compose profile, not the baseline stack.
2. Add an `outbox_events` table. A business transaction changes domain state and inserts one versioned JSON event in the same PostgreSQL transaction.
3. Run the existing `cmd/worker` as an outbox publisher. It claims rows in bounded batches (using locking such as `FOR UPDATE SKIP LOCKED`), publishes, and records success. It never promises exactly-once end-to-end delivery.
4. Use the envelope: `event_id`, `event_type`, `event_version`, `aggregate_id`, `occurred_at`, `correlation_id`, `causation_id`, `producer`, and `payload`.
5. Key order events by `order_id`; test ordering, rebalance behavior, duplicate delivery, lag, retry, poison events, and dead-letter policy.
6. Add an idempotent consumer with a durable processed-event key or naturally idempotent database operation.
7. Write an ADR comparing JSON, Avro, and Protobuf before choosing a schema registry. Begin with versioned JSON.
8. Keep application logs on stdout. A collector may additionally send sanitized logs to `observability.application-logs.v1`; the application never synchronously writes each log to Kafka.
9. Build a separate read-only DevOps-agent consumer. It groups errors, correlates request/trace/deployment IDs, drafts incident reports and runbooks, and suggests commands. It cannot execute commands or mutate AWS, Kubernetes, IAM, Kafka, or databases.
10. Enforce redaction, topic ACLs, least privilege, retention, auditability, consumer-lag alerts, and failure drills. Kafka/collector/agent outage must not stop checkout or API serving.

Exit criteria: no PostgreSQL/Kafka naive dual write, duplicate delivery is tested, consumer lag is observable, and removing Kafka does not break synchronous business correctness.

## Milestone 6 — Kubernetes and GitOps

Create Helm charts, immutable SHA-tagged images, health probes, resource budgets, HPA experiments, and a separate GitOps repository only after explicit authorization. Argo CD PreSync runs one Goose migration Job; PostSync later runs smoke tests.

## Milestone 7 — AWS

Model infrastructure in Terraform, then evaluate EKS, RDS, CloudWatch, ALB Controller, and IAM least privilege. Compare MSK cost/operations before creation. No paid resource is created without explicit authorization.

## Milestone 8 — Evidence-based extraction

Use measured evidence to select Notification or Payment. Establish database ownership, compatibility, SLOs, and operational readiness. Compare choreography with orchestration and use a saga only when a cross-service workflow requires it. Measure latency and operational cost before and after.
