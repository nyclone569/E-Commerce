# ADR-0006: Use Synchronous Processing Before Kafka

## Status

Accepted — Kafka deferred to Milestone 5.

## Context

Kafka is a learning goal, but no Milestone 1 flow needs high-throughput fan-out, replay, decoupled availability, or independent consumers.

## Forces and constraints

Business correctness must work without Kafka. PostgreSQL/Kafka dual writes are forbidden. Consumers must later tolerate duplicates. Application logs must not synchronously depend on Kafka.

## Options considered

1. Synchronous in-process calls and PostgreSQL transactions.
2. Kafka from day one with direct producer calls.
3. Database transactional outbox followed by Kafka publishing.
4. A simpler task queue when only background jobs—not event replay—are required.

## Real-world usage examples

Verified fact: LinkedIn built Kafka to handle high-volume activity streams and later reported an ecosystem of many clusters, brokers, topics, partitions, schema tooling, mirroring, and operations. Verified fact: Apache Kafka documents at-least-once as a normal producer/consumer outcome depending on configuration. Inference: LinkedIn's scale explains the platform investment; copying it for a catalog CRUD slice would add operations without solving a present problem.

## Decision

Use synchronous calls through Milestones 1–4. In Milestone 5, add a PostgreSQL transactional outbox, separate publisher, versioned JSON envelope, keyed partitions, and idempotent consumers.

## Detailed reasoning

First establish transaction boundaries and failure semantics with the fewest moving parts. The outbox later closes the database/broker dual-write gap by committing domain state and event intent atomically, while accepting repeat publication.

## Positive consequences

Clear early failure behavior, local atomicity, fewer dependencies, and a meaningful Kafka comparison baseline.

## Negative consequences

Temporal coupling remains; long non-critical work cannot yet be decoupled; later outbox work is additional code and operations.

## Failure modes introduced

Synchronous dependency latency can propagate. Later, outbox backlog, duplicate publication, poison events, rebalances, skewed partitions, schema incompatibility, and consumer lag will appear.

## Operational requirements

Before Kafka: timeouts and transaction metrics. With Kafka: outbox age/size, publisher retries, topic ACLs/retention, consumer lag, idempotency store, dead-letter policy, schema compatibility, and failure drills.

## Metrics to observe

Current request latency/error and transaction duration; later outbox oldest age, publish error rate, end-to-end event latency, duplicate rate, partition skew, and consumer lag.

## Revisit triggers

At least one measured durable fan-out/replay/decoupling need, synchronous work harms an SLO, transaction design is stable, and the team can operate broker/publisher/consumers.

## References

- [LinkedIn Engineering: Kafka at 7 trillion messages per day](https://www.linkedin.com/blog/engineering/open-source/apache-kafka-trillion-messages)
- [Apache Kafka delivery semantics](https://kafka.apache.org/documentation/#semantics)
- [AWS Prescriptive Guidance: transactional outbox pattern](https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/transactional-outbox.html)
