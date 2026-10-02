import { afterEach, describe, expect, it, vi } from "vitest";
import { addCartItem, setCartQuantity } from "./cart";

afterEach(() => vi.unstubAllGlobals());

describe("cart mutations", () => {
  it("fetches a session-bound CSRF token before adding a SKU", async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ csrf_token: "test-csrf" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }));
    vi.stubGlobal("fetch", fetchMock);

    await addCartItem("sku-123");

    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/auth/csrf", expect.objectContaining({ credentials: "same-origin", cache: "no-store" }));
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/cart/items", expect.objectContaining({
      method: "POST",
      headers: expect.objectContaining({ "X-CSRF-Token": "test-csrf" }),
      body: JSON.stringify({ sku_id: "sku-123", quantity: 1 }),
    }));
  });

  it("does not mutate when the session has expired", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 401 }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(setCartQuantity("sku-123", 2)).rejects.toThrow("Please sign in");
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});
