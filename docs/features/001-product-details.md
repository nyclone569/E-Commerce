# Feature 001: View Product Details

## User function

As a shopper, I can open a product from the catalog and see its description, available SKU choices, and the price of each SKU.

## Scope

This slice is read-only. Every product created in the current model is considered storefront-visible because publication lifecycle is a separate catalog-management function. Selecting a SKU for a cart, inventory availability, images, and product editing are out of scope.

## Acceptance criteria

1. A product card links to `/products/{slug}`.
2. A valid slug displays product name, description, and SKUs ordered by code.
3. Each SKU displays its code, currency, and formatted price.
4. An unknown or invalid slug returns API HTTP 404 and renders the Next.js not-found state.
5. Backend failures render the existing error state and do not expose internal error details.
6. Request cancellation and the existing five-second frontend timeout remain active.

## Business rules

- Product slug is the public lookup key and is unique.
- SKU code is unique and deterministic display ordering uses `code ASC`.
- Prices use integer minor units and a three-letter currency code.
- This slice does not claim that a displayed SKU has inventory available.

## Decision Review

| Decision | Options | Selected | Why now | Main trade-off | Revisit trigger |
| --- | --- | --- | --- | --- | --- |
| Public identifier | UUID; slug; both | Slug | Human-readable storefront URL and existing uniqueness constraint | Rename requires redirect/alias policy | Product rename or localization is implemented |
| Frontend fetching | Server Component; TanStack Query | Server Component | Read-only initial request with no client cache/mutation need | No client background refresh | SKU selection needs live inventory or client mutation |
| Repository read | One joined query; product then SKUs | Product then SKUs | Reuses clear generated models and keeps mapping explicit | Two round trips and no shared read snapshot | Concurrent product/SKU editing or measured latency matters |
| Missing product | Empty success; redirect; 404 | 404 | Resource does not exist | Cache behavior must be understood later | Slug aliases or soft deletion are added |
| Publication | Add status now; treat created products as visible | Defer status | Publication is a separate operator function | Current API-created products are immediately visible | Catalog administration is implemented |

## Request and data flow

```text
Browser GET /products/{slug}
  -> Next.js Server Component
  -> Go GET /api/products/{slug}
  -> catalog handler
  -> catalog service
  -> catalog repository
  -> PostgreSQL product query
  -> PostgreSQL ordered SKU query
  -> JSON
  -> rendered HTML
```

## Failure modes

- Slug is syntactically invalid: return 404 to avoid distinguishing invalid from unknown public resources.
- Product is absent: map `pgx.ErrNoRows` to the catalog not-found domain error.
- Next.js renders the route-specific not-found UI and emits `noindex`. Because a loading boundary can flush a streamed response before the lookup completes, the frontend transport status can remain 200 even though the upstream API returned 404. Rework the loading/data boundary if strict CDN HTTP 404 semantics become a requirement.
- PostgreSQL is unavailable: return the standard 500 envelope; Next.js renders its error boundary.
- Product is deleted between the two repository queries: a stale product with no SKUs could be returned. There is no delete/update flow yet; use a joined query or repeatable-read transaction when that flow is added.

## Metrics to add later

- Product-detail request rate, status, and p95/p99 duration.
- Not-found rate by normalized route, never by raw unbounded slug label.
- PostgreSQL query duration and rows returned.

## Learning outcome

This slice demonstrates resource lookup, domain error mapping, server rendering, 404 behavior, and deterministic child ordering without adding client-side server-state management prematurely.
