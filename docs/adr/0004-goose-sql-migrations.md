# ADR-0004: Use Goose SQL Migrations

## Status

Accepted — Milestone 1.

## Context

Schema evolution must be explicit, ordered, testable, and independent from API replica startup.

## Forces and constraints

The project is Go-focused, migrations should teach SQL, production will use expand–migrate–contract, and application rollback must not assume schema rollback.

## Options considered

1. Goose: Go CLI/library, SQL or Go migrations, simple annotations.
2. golang-migrate: focused multi-source migration CLI/library and SQL files.
3. Atlas: declarative/desired-state workflows, linting, diff and richer governance.
4. Flyway: mature JVM-based, cross-language database migration platform.

## Real-world usage examples

Verified fact: Goose runs SQL migrations in a transaction by default and supports explicit non-transactional files; Atlas provides declarative schema and migration lint workflows; Flyway targets heterogeneous stacks. Inference: centralized platform teams may justify richer policy/drift tooling, while this project's scale benefits from a smaller Go-native workflow.

## Decision

Use sequential Goose SQL migrations. Use Go migration code only when the change genuinely cannot be expressed safely in SQL. Never migrate during API startup.

## Detailed reasoning

Goose has enough ordering and transaction behavior while keeping the artifact as readable SQL. Local operators run it explicitly. Later Argo CD runs a single PreSync Job before application sync.

## Positive consequences

Small tool surface, visible SQL, transactional defaults, and easy local/CI execution.

## Negative consequences

Less declarative drift detection and policy analysis than Atlas; developers design safe rollout phases manually.

## Failure modes introduced

Locking DDL can block traffic, a non-transactional change can partially apply, multiple runners can contend, and destructive down migrations can destroy data.

## Operational requirements

Validate migrations on a fresh and upgraded database, inspect locks/duration, back up production, run one controlled job, use expand–migrate–contract, and prefer forward fixes over production down migrations.

## Metrics to observe

Migration duration, lock wait/blocking, failure count, schema version, replica lag, and application error changes after migration.

## Revisit triggers

Multiple languages/teams need centralized governance, drift detection becomes material, or complex schema policy needs Atlas/Flyway-class tooling.

## References

- [Goose SQL annotations](https://pressly.github.io/goose/documentation/annotations/)
- [golang-migrate project](https://github.com/golang-migrate/migrate)
- [Atlas documentation](https://atlasgo.io/docs)
- [Flyway documentation](https://documentation.red-gate.com/flyway)
