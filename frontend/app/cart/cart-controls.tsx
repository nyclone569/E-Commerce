"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";
import { removeCartItem, setCartQuantity } from "@/lib/cart";

export function CartControls({ skuID, quantity }: { skuID: string; quantity: number }) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  async function change(nextQuantity: number | null) {
    setBusy(true);
    setError("");
    try {
      if (nextQuantity === null) await removeCartItem(skuID);
      else await setCartQuantity(skuID, nextQuantity);
      router.refresh();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Could not update cart.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="mt-5 flex items-center gap-3">
      <button type="button" disabled={busy || quantity <= 1} onClick={() => change(quantity - 1)} aria-label="Decrease quantity" className="rounded-lg border px-3 py-1 disabled:opacity-40">−</button>
      <span aria-label="Quantity">{quantity}</span>
      <button type="button" disabled={busy || quantity >= 99} onClick={() => change(quantity + 1)} aria-label="Increase quantity" className="rounded-lg border px-3 py-1 disabled:opacity-40">+</button>
      <button type="button" disabled={busy} onClick={() => change(null)} className="ml-3 text-sm font-bold text-red-700 disabled:opacity-40">Remove</button>
      {error && <span role="alert" className="text-sm text-red-700">{error}</span>}
    </div>
  );
}
