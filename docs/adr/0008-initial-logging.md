# ADR-0008: Emit Structured JSON Logs to Standard Output

## Status

Accepted — staged progression through Milestones 1, 4, and 5.

## Context

The application needs useful operational records without coupling availability to a log backend or leaking sensitive information.

## Forces and constraints

Use `log/slog`; write stdout; never log secrets, auth headers, cookies, payment/card data, PII, or raw sensitive bodies. Logs, metrics, traces, domain events, and audit events are separate concepts.

## Options considered

1. stdout directly collected to CloudWatch.
2. stdout → Fluent Bit/OpenTelemetry Collector → CloudWatch.
3. stdout → collector → Kafka → specialized consumers.
4. application synchronously publishes every log to Kafka.

## Real-world usage examples

Verified fact: Kubernetes' logging architecture captures container stdout/stderr and recommends a node-level logging agent as a common approach. Verified fact: OpenTelemetry Collector receives, processes, and exports telemetry. Verified fact: LinkedIn described Kafka-backed exception processing at a scale above one million events per second. Inference: that scale can justify buffering and specialized consumers; AuroraShop initially needs reliable application behavior and simple logs, not that platform.

## Decision

Local: JSON stdout with service/environment/version/request fields. Kubernetes: add a collector. AWS baseline: collector to CloudWatch. Kafka lab: collector additionally writes sanitized records to `observability.application-logs.v1`. Never use direct synchronous application-to-Kafka logging.

## Detailed reasoning

Stdout lets the runtime own transport and decouples the request path from log infrastructure. A collector centralizes enrichment, redaction, buffering, routing, and retries. Kafka remains optional and distinct from domain-event topics.

## Positive consequences

Machine-readable output, minimal application dependency, portable collection, and API survival during collector/Kafka outage.

## Negative consequences

Local logs lack query UI; collector configuration is another failure surface; stdout backpressure/drop behavior must be understood.

## Failure modes introduced

High-cardinality cost, accidental sensitive fields, multiline stack size, disk pressure, collector queue overflow, duplicated/dropped export, and Kafka lag in the optional path.

## Operational requirements

Field schema and redaction tests, retention/access controls, sampling/rate limits where justified, bounded collector queues, health/queue metrics, and deployment-version correlation.

## Metrics to observe

Log bytes/rate, dropped records, collector queue occupancy/export failures, ingestion cost, sensitive-data incidents, error groups, and optional Kafka consumer lag.

## Revisit triggers

Kubernetes deployment, incident investigations cannot meet recovery targets, volume/cost thresholds are exceeded, or a validated Kafka consumer use case needs buffered log streams.

## References

- [Go `log/slog`](https://pkg.go.dev/log/slog)
- [Kubernetes logging architecture](https://kubernetes.io/docs/concepts/cluster-administration/logging/)
- [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/)
- [LinkedIn Engineering: Inception exception logs](https://www.linkedin.com/blog/engineering/archive/inception-how-linkedin-deals-with-exception-logs)
