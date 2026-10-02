# Feature 003: Authenticated VND Cart

## User function

A signed-in shopper adds a SKU, opens their cart, changes quantity, and removes an item. A cart is a planning surface, not a stock reservation or a final checkout quote.

## Decision Review

| Decision | Options | Selected | Why now | Main trade-off | Revisit trigger |
| --- | --- | --- | --- | --- | --- |
| Cart identity | Guest; account-owned; merged | Account-owned | Identity already exists | Sign-in required | Guest conversion requirement |
| Creation | Create at registration; create on GET; create on first add | First add | Empty GET is read-only | First add needs an upsert | Measured write pressure |
| Prices | Snapshot at add; read current catalog price | Current catalog price | Cart is an estimate | Price may change before checkout | Promotions or quote guarantees |
| Money | Multi-currency; VND only | VND only | One market in this learning slice | Existing USD demo rows cannot enter cart | International checkout requirement |
| Stock | Reserve on add; check during checkout | Later checkout | No inventory implementation yet | Item can become unavailable | Checkout design and inventory module |
| Cart persistence | Browser; PostgreSQL; cache | PostgreSQL | Durable user ownership and simple atomic updates | Database write and read load | Measured pool/latency pressure |
| Browser mutation protection | SameSite; Origin; session-bound CSRF token | All three | Cookie-authenticated mutations start here | Extra token-fetch round trip | Different client/auth topology |

## Request and data flow

```text
Browser GET /cart
  -> Next.js Server Component forwards session cookie to Go GET /api/cart
  -> Identity Authenticate resolves session to user ID
  -> Cart handler reads user ID from context (never from URL/body)
  -> Cart service asks Cart repository for user-owned SKU IDs and quantities
  -> Cart service asks public Catalog service for current SKU details/prices
  -> Next.js renders a VND estimate

Browser POST /api/cart/items
  -> Next.js same-origin proxy -> Go Authenticate -> Origin check -> CSRF check
  -> Cart handler parses JSON -> Cart service validates SKU and quantity
  -> Catalog service checks VND SKU -> Cart repository transaction
  -> PostgreSQL creates one active cart on first add and atomically increments the item
```

`PUT /api/cart/items/{skuID}` sets an absolute quantity; `DELETE` removes it. Both SQL statements filter through the active cart owned by the authenticated user. `GET` does not insert a cart row. The active cart and items have no cached price: every read uses current Catalog values. Item order is deterministic by insertion time, then SKU ID.

## Invariants and failure behavior

- No user ID is accepted from client input. A missing or invalid session returns `401` before Cart queries run.
- Mutations require an exact public Origin and a CSRF token derived from the current opaque session. The raw session is never returned to JavaScript; `GET /api/auth/csrf` is authenticated and `no-store`.
- PostgreSQL has at most one active cart per user. An upsert serializes concurrent adds to that cart. The item upsert increments atomically and refuses a total above 99.
- A cart has at most 100 distinct SKUs. A transaction rolls back an insertion that would exceed the limit.
- A SKU foreign key prevents dangling cart items. A deleted or non-VND SKU is reported as unavailable; stock availability is not yet promised.
- Item prices are not locked. Checkout must later reprice and validate stock atomically, display any change, and require shopper confirmation when necessary.
- Catalog owns `products` and `skus`; Cart owns `carts` and `cart_items`. Cart calls the public Catalog service and does not mutate Catalog tables.
- New SKU prices are capped at 900,000,000,000 minor VND units so a maximum cart subtotal fits JavaScript's safe integer range.

## Limitations and revisit triggers

No guest cart, save-for-later, promotions, shipping, tax, checkout, inventory availability, product publication state, or price-lock policy exists. These are separate business decisions. `POST` adds quantity and is not idempotent if a client retries after an unknown response; `PUT` sets an absolute quantity. `carts.updated_at` currently tracks cart creation and add operations, not every quantity change or removal, so it must not drive stale-cart cleanup yet. The VND migration retains old USD catalog rows for reading but blocks new non-VND writes; those rows cannot be added to cart. The `price_cents` to `price_minor` rename is an incompatible database change: do not deploy old application code against the migrated schema. A production rollout would use expand–migrate–contract rather than this learning-repo cutover.

Observe cart request latency/errors, database pool wait, contention on active-cart upserts, conflict rates, and cart-to-checkout conversion before adding a cache or distributed cart service.

## Verification

Unit tests check empty carts, validation, VND gating, price calculation, and CSRF token binding. PostgreSQL integration tests check real migrations, read-only empty GET, ownership isolation, concurrent additions, one active cart, quantity limit, and removal.

## References

- [PostgreSQL INSERT and ON CONFLICT](https://www.postgresql.org/docs/current/sql-insert.html)
- [PostgreSQL ALTER TABLE and NOT VALID constraints](https://www.postgresql.org/docs/current/sql-altertable.html)
- [OWASP CSRF Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html)
