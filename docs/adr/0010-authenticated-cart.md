# ADR-0010: Keep the Cart Account-Owned and Synchronous

## Status

Accepted for the Milestone 2 Cart slice.

## Context

The shopper needs a durable cart after signing in. Identity, Catalog, one Go API, and one PostgreSQL database already exist. Inventory and checkout do not yet exist, so adding an item cannot imply reservation or a final price.

## Forces and constraints

Cart data is private to the authenticated owner; VND is the only new sale currency; concurrent updates must not lose quantity; empty reads should not create rows; the module boundary must remain explicit.

## Options considered

Guest/browser cart versus account-owned cart; create at registration versus first add; price snapshot versus current Catalog price; PostgreSQL versus cache-backed cart; stock reservation on add versus checkout validation.

## Real-world usage examples

Verified fact: Shopify's Storefront API describes cart cost as an estimate that may change at checkout. Verified fact: commercetools documents a Recalculate cart action to refresh stale prices and discounts. These show that cart pricing and checkout guarantees require explicit policy. Architectural inference: larger stores can justify guest-cart merging, promotional repricing, and cache layers because conversion and request volume make those capabilities valuable. AuroraShop has not measured such a need.

## Decision

Use one authenticated, PostgreSQL-backed active cart per user. Create it lazily on first add. Store SKU IDs and quantities only. Read current prices through Catalog's public service. Keep mutation calls synchronous. Require session authentication, exact Origin, and a session-bound CSRF token. Use VND only. Do not reserve stock on add.

## Detailed reasoning

The user ID comes from Identity's authenticated request context and is applied to every Cart query, so callers cannot nominate another owner. PostgreSQL's unique partial index protects the one-active-cart invariant. A transactional upsert plus bounded item count prevents duplicate carts and lost increments, while the item quantity CHECK is a second line of defense. Cart asks Catalog for SKU details rather than importing Catalog's repository or changing its tables. A current-price estimate avoids a premature quote-expiry policy; checkout must make the final commitment later.

## Positive consequences

Small, teachable flow; durable cart across devices; explicit ownership; no distributed cache consistency; simple VND arithmetic; no misleading stock promise.

## Negative consequences

Guest shoppers cannot save a cart; prices can change between views; every cart read queries Catalog and PostgreSQL; CSRF protection adds a browser round trip; PostgreSQL is on every authenticated path.

## Failure modes introduced

Concurrent increments can hit quantity limits; a deleted SKU makes an old cart line unavailable; changed prices surprise shoppers; unavailable database blocks cart; a proxy that drops cookies or CSRF headers causes failures; upsert contention can raise latency for one hot cart.

## Operational requirements

Do not log cart contents, cookies, or CSRF values. Keep Cart and Catalog SQL ownership separate. Preserve `no-store` on private responses. Measure pool saturation and errors. Before checkout, define repricing, inventory, and stale-cart policies.

## Metrics to observe

Cart GET/add/update/remove latency, `401`/`403`/`409` rates, upsert lock wait, database pool wait, number of active carts, and cart-to-checkout conversion. Use bounded labels without raw user IDs.

## Revisit triggers

Evidence of guest-cart conversion loss, high cart database load, price-policy requirements, inventory reservation need, regional currency support, or independent deployment/fault-isolation requirements.

## References

- [Shopify Storefront API Cart](https://shopify.dev/docs/api/storefront/latest/objects/Cart)
- [commercetools Cart recalculation](https://docs.commercetools.com/api/projects/me-carts)
- [PostgreSQL INSERT / ON CONFLICT](https://www.postgresql.org/docs/current/sql-insert.html)
- [OWASP CSRF Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html)
