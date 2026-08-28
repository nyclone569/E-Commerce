# ADR-0005: Use One PostgreSQL Database Initially

## Status

Accepted — version 1.

## Context

Commerce data is relational and requires uniqueness, referential integrity, concurrent updates, and multi-step atomic transactions.

## Forces and constraints

Inventory must never become negative; retries must not duplicate orders; module tables need logical ownership; version 1 must use one database.

## Options considered

1. PostgreSQL: relational constraints, transactions, MVCC, row locking, JSON where appropriate.
2. MySQL: capable relational alternative with different operational/SQL details.
3. Document database: flexible aggregates, but cross-aggregate constraints and transaction learning are less direct.
4. Database per module from day one: stronger physical ownership at substantial transaction and operational cost.

## Real-world usage examples

Verified fact: PostgreSQL documents MVCC snapshots, transaction isolation, locks, constraints, and Serializable Snapshot Isolation. Public company architectures commonly use multiple specialized stores at high scale, but workload-specific choices do not imply AuroraShop needs polyglot persistence. Inference: one relational database is the smallest system that directly satisfies current correctness requirements.

## Decision

Use one PostgreSQL database. Each module logically owns named tables and only its repository/query package mutates them.

## Detailed reasoning

Local ACID transactions are valuable for checkout learning. Foreign keys and checks are a final invariant layer. Logical ownership prepares extraction without forcing distributed transactions now.

## Positive consequences

Strong integrity, mature SQL, atomic workflows, simpler backups and local operations.

## Negative consequences

One database is a shared resource and availability domain; physical access does not enforce every logical module rule.

## Failure modes introduced

Connection exhaustion, lock contention/deadlocks, long transactions, vacuum pressure, replica lag later, and accidental cross-module SQL.

## Operational requirements

Bound pools per replica, short transactions, statement/lock timeouts where appropriate, backups and restore drills, migration discipline, slow-query analysis, and least-privilege roles later.

## Metrics to observe

Connections/pool wait, query p95/p99, lock wait and deadlocks, transaction duration, CPU/IO, table/index growth, vacuum health, and recovery objectives.

## Revisit triggers

A module needs demonstrably different scaling/storage/security/availability, database contention persists after query/index/pool tuning, or organizational ownership requires physical separation.

## References

- [PostgreSQL concurrency control](https://www.postgresql.org/docs/current/mvcc.html)
- [PostgreSQL constraints](https://www.postgresql.org/docs/current/ddl-constraints.html)
- [PostgreSQL transaction isolation](https://www.postgresql.org/docs/current/transaction-iso.html)
