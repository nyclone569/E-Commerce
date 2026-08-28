# ADR-0007: Separate Jenkins CI from Argo CD

## Status

Accepted as the delivery target; only CI foundation exists in Milestone 1.

## Context

Build/test/image publication and cluster reconciliation have different trust, audit, and failure concerns.

## Forces and constraints

Images use immutable Git SHA tags, never `latest`. Jenkins must not run `kubectl apply`. Production desired state eventually belongs in a separate GitOps repository, which must not be created without authorization.

## Options considered

1. Jenkins builds and imperatively deploys.
2. Argo CD performs builds and deploys.
3. Jenkins owns CI/artifacts; Argo CD reconciles Git desired state and owns CD.

## Real-world usage examples

Verified fact: Argo CD's official CI automation guidance describes CI building/publishing an image and updating manifests in Git, after which Argo CD follows the GitOps desired state. Verified fact: Jenkins Pipeline models build/test/delivery stages as code. Inference: large organizations often separate duties and credentials because many teams/clusters increase blast radius; AuroraShop adopts the responsibility boundary but not a large platform.

## Decision

Jenkins checks dependencies, formatting, vet/static analysis, tests, generation/migrations, frontend quality/builds, and image builds/pushes with SHA tags. Later it proposes a GitOps change. Argo CD alone reconciles Kubernetes, runs a Goose PreSync migration Job, deploys workloads, checks health, and later runs PostSync smoke tests.

## Detailed reasoning

Git becomes the auditable deployment request and Argo CD reports drift. CI credentials do not need broad cluster mutation. The migration hook orders schema expansion before the compatible application.

## Positive consequences

Traceable desired state, drift visibility, reduced CI cluster privilege, reproducible immutable artifacts, and clear ownership.

## Negative consequences

Another repository/control loop is required later; promotion is asynchronous; bad Git can be reconciled automatically.

## Failure modes introduced

Stale image references, reconciliation loops, PreSync migration failure, Git/registry outage, bad health checks, and CI/CD status divergence.

## Operational requirements

SHA tags/digests, signed/scanned artifacts later, protected GitOps reviews, Argo RBAC, hook timeouts, migration idempotency, health checks, audit logs, and rollback/run-forward procedures.

## Metrics to observe

CI duration/failure, artifact age, lead time, Argo sync/health status, drift duration, failed migrations, deployment frequency, change failure rate, and recovery time.

## Revisit triggers

GitOps promotion latency blocks required delivery, control ownership becomes ambiguous, or risk analysis supports a simpler workflow for a non-production environment.

## References

- [Argo CD: Automation from CI Pipelines](https://argo-cd.readthedocs.io/en/stable/user-guide/ci_automation/)
- [Argo CD resource hooks](https://argo-cd.readthedocs.io/en/stable/user-guide/resource_hooks/)
- [Jenkins Pipeline documentation](https://www.jenkins.io/doc/book/pipeline/)
