# ADR-0001: Start with a Modular Monolith

## Status

Accepted — Milestone 1.

## Context

AuroraShop needs clear commerce boundaries but has one learner/team, no production load evidence, and workflows that will require atomic consistency.

## Forces and constraints

- Correctness and learnability precede independent deployment.
- Next.js and Go are separate deployables; the Go backend must remain one deployable and use one PostgreSQL database in version 1.
- Future extraction must be possible without pretending local calls are remote calls today.

## Options considered

1. Unstructured monolith: lowest initial ceremony, highest boundary-erosion risk.
2. Modular monolith: domain packages, public service boundaries, local calls, shared transaction capability.
3. Microservices: independent deployment/scaling and fault domains, plus network, compatibility, observability, and distributed-data costs.

## Real-world usage examples

Verified fact: Shopify described componentizing its large Rails monolith and enforcing component privacy/dependencies. Verified fact: Uber described adopting microservices when its engineering organization and operational problems grew, and later organizing roughly 2,200 services into about 70 domains to reduce complexity. Inference: those experiences support learning boundaries now while waiting for AuroraShop-specific extraction evidence; they do not prove one topology is always superior.

## Decision

Use one Go modular monolith with `identity`, `catalog`, `cart`, `inventory`, `order`, and `payment` modules and one PostgreSQL database. Modules interact synchronously through public service methods.

## Detailed reasoning

In-process calls and ACID transactions make early correctness observable. Package and table ownership teach the same boundary reasoning needed later. Deployment remains coarse, which is acceptable until measured friction outweighs distributed-system cost.

## Positive consequences

- Simple local development, debugging, transactions, and refactoring.
- Module APIs and ownership are explicit before extraction.
- Lower infrastructure and operational burden.

## Negative consequences

- One backend deployment and primary fault domain.
- A shared process and pool can create noisy neighbors.
- Discipline, review, and tests must prevent boundary bypass.

## Failure modes introduced

A panic or resource leak can affect every module; one slow query can exhaust the shared pool; schema access can couple modules covertly.

## Operational requirements

Graceful shutdown, bounded database pool, per-route telemetry, module dependency checks later, and database authorization boundaries where useful.

## Metrics to observe

p95/p99 latency by route/module, error rate, pool wait duration, CPU/memory, deploy frequency/failure rate, and change-coupling between modules.

## Revisit triggers

Measured need for independent scaling/deployment, different data ownership, stronger security/fault isolation, separate team ownership, or a load-test bottleneck that cannot be solved economically in-process.

## References

- [Shopify Engineering: Under Deconstruction](https://shopify.engineering/shopify-monolith)
- [Uber Engineering: Domain-Oriented Microservice Architecture](https://www.uber.com/blog/microservice-architecture/)
- [Martin Fowler: Microservice Prerequisites](https://martinfowler.com/bliki/MicroservicePrerequisites.html)
