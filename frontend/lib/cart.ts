export type CartItem = {
  sku_id: string;
  product_id: string;
  product_name: string;
  product_slug: string;
  sku_code: string;
  quantity: number;
  unit_price_minor: number;
  line_total_minor: number;
};

export type Cart = {
  items: CartItem[];
  subtotal_minor: number;
  currency: "VND";
};

type APIError = { error?: { message?: string } };

async function cartMutation(method: "POST" | "PUT" | "DELETE", path: string, body?: object): Promise<void> {
  const tokenResponse = await fetch("/api/auth/csrf", { credentials: "same-origin", cache: "no-store" });
  if (tokenResponse.status === 401) throw new Error("Please sign in to use your cart.");
  if (!tokenResponse.ok) throw new Error("Could not prepare the cart request.");
  const { csrf_token } = (await tokenResponse.json()) as { csrf_token: string };
  const response = await fetch(path, {
    method,
    credentials: "same-origin",
    cache: "no-store",
    headers: { "X-CSRF-Token": csrf_token, ...(body ? { "Content-Type": "application/json" } : {}) },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!response.ok) {
    const payload = (await response.json().catch(() => ({}))) as APIError;
    throw new Error(payload.error?.message ?? "Cart request failed.");
  }
}

export function addCartItem(skuID: string): Promise<void> {
  return cartMutation("POST", "/api/cart/items", { sku_id: skuID, quantity: 1 });
}

export function setCartQuantity(skuID: string, quantity: number): Promise<void> {
  return cartMutation("PUT", `/api/cart/items/${encodeURIComponent(skuID)}`, { quantity });
}

export function removeCartItem(skuID: string): Promise<void> {
  return cartMutation("DELETE", `/api/cart/items/${encodeURIComponent(skuID)}`);
}
