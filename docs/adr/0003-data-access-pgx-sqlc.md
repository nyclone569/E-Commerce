# ADR-0003: Use pgx and sqlc for PostgreSQL Access

## Status

Accepted — Milestone 1.

## Context

The project must teach SQL and PostgreSQL behavior while avoiding repetitive, unchecked row scanning.

## Forces and constraints

SQL should remain reviewable; queries and Go types should be checked at generation time; PostgreSQL features must stay accessible; generated records should not automatically become domain or HTTP models.

## Options considered

1. pgx + sqlc: PostgreSQL-native driver/pool plus generated typed methods from SQL.
2. `database/sql` with handwritten scanning: minimal tooling but runtime scan/query drift risk and repetition.
3. GORM: productive ORM conventions and associations, but more abstraction and runtime behavior.
4. Ent: schema-as-Go and generated graph API, strong types, but SQL is less central to the learning path.

## Real-world usage examples

Verified fact: sqlc supports pgx/v5 generation and database type overrides; pgx exposes PostgreSQL-specific functionality and pooling. GORM and Ent official docs demonstrate richer model/schema abstractions. Inference: large companies often standardize data libraries/platforms to reduce variance across many teams; AuroraShop's single-team scale does not justify building such a platform.

## Decision

Write SQL, generate typed pgx/v5 query methods with sqlc, and map generated rows to catalog domain types in the repository.

## Detailed reasoning

This keeps execution plans, ordering, transactions, and constraints visible. Compile-time method/parameter generation removes much scanning boilerplate but not the responsibility to understand SQL, indexes, nullability, isolation, or mapping.

## Positive consequences

Visible SQL, typed calls, PostgreSQL feature access, and testable query files.

## Negative consequences

Generation is a required build step; dynamic query composition is less convenient; explicit mapping and deeper SQL knowledge are required.

## Failure modes introduced

Stale generated files, generator-version drift, N+1 queries, wrong SQL despite correct types, and leaking generated types across module boundaries.

## Operational requirements

Pin sqlc, verify clean regeneration in CI, inspect query plans for hot paths, bound the pgx pool, set timeouts, and run integration tests against PostgreSQL.

## Metrics to observe

Query latency, rows scanned/returned, pool acquire time, connection utilization, slow-query frequency, generation drift, and database errors by SQLSTATE.

## Revisit triggers

Most required queries become safely composable/dynamic and sqlc becomes the bottleneck, or organizational governance requires a different access layer.

## References

- [pgx project documentation](https://github.com/jackc/pgx)
- [sqlc configuration reference](https://docs.sqlc.dev/en/latest/reference/config.html)
- [GORM documentation](https://gorm.io/docs/)
- [Ent documentation](https://entgo.io/docs/getting-started/)
