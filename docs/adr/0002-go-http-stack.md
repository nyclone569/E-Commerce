# ADR-0002: Use `net/http` with Chi

## Status

Accepted — Milestone 1.

## Context

The API needs routing, middleware, cancellation, and standard HTTP lifecycle behavior without hiding Go's HTTP model.

## Forces and constraints

Prefer the standard library, small dependencies, explicit middleware, compatibility with `http.Handler`, and easy unit tests.

## Options considered

1. Standard `net/http` only: sufficient but verbose for nested routes and parameters.
2. Chi: small composable router built on `net/http` interfaces.
3. Gin or Echo: more integrated conveniences and a framework-specific context/API.

## Real-world usage examples

Verified fact: Go's standard server exposes timeout and graceful-shutdown primitives; Chi implements `http.Handler` and advertises stdlib compatibility. Public framework adoption counts vary by source and time, so popularity is not used as the decision criterion. Inference: preserving standard interfaces reduces learning and migration cost for this project.

## Decision

Use `net/http` for server/lifecycle and Chi only for routing and middleware composition.

## Detailed reasoning

AuroraShop needs no framework-specific binding, validation, or DI container. Chi removes routing boilerplate without replacing request, response, context, middleware, or test types.

## Positive consequences

Standard `httptest`, context cancellation, interoperable middleware, and a small conceptual surface.

## Negative consequences

JSON decoding, validation, error envelopes, and middleware must be designed explicitly.

## Failure modes introduced

Middleware ordering mistakes, inconsistent hand-written transport behavior, missing status recording for advanced interfaces, and accidental unbounded bodies.

## Operational requirements

Request/body limits, panic recovery, sanitized access logs, request IDs, server timeouts, graceful shutdown, and later trusted-proxy configuration.

## Metrics to observe

Request duration/status by route, panic count, rejected body count, active requests at shutdown, and handler allocation/CPU profiles.

## Revisit triggers

Repeated transport boilerplate causes defects, protocol needs extend beyond HTTP, or a well-evidenced framework capability materially reduces maintenance without obscuring core behavior.

## References

- [Go `net/http` package](https://pkg.go.dev/net/http)
- [Chi documentation](https://github.com/go-chi/chi)
- [Gin documentation](https://gin-gonic.com/docs/)
- [Echo documentation](https://echo.labstack.com/docs)
