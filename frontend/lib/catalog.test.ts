import { afterEach, describe, expect, it, vi } from "vitest";
import { CatalogNotFoundError, formatMoney, getProduct } from "./catalog";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("formatMoney", () => {
  it("formats integer minor units as currency", () => {
    expect(formatMoney(1299, "USD")).toBe("$12.99");
  });
});

describe("getProduct", () => {
  it("maps an API 404 to a catalog not-found error", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 404 })));

    await expect(getProduct("missing-product")).rejects.toBeInstanceOf(CatalogNotFoundError);
  });

  it("encodes the public slug in the request URL", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ id: "1", slug: "aurora mug", skus: [] }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await getProduct("aurora mug");

    expect(fetchMock).toHaveBeenCalledWith(
      "http://localhost:8080/api/products/aurora%20mug",
      expect.objectContaining({ cache: "no-store" }),
    );
  });
});
