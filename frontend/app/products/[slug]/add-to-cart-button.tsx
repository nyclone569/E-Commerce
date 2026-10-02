"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { addCartItem } from "@/lib/cart";

export function AddToCartButton({ skuID }: { skuID: string }) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");

  async function add() {
    setBusy(true);
    setMessage("");
    try {
      await addCartItem(skuID);
      setMessage("Added to cart.");
      router.refresh();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Could not add item.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div>
      <button type="button" onClick={add} disabled={busy} className="mt-2 rounded-full bg-emerald-950 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">
        {busy ? "Adding…" : "Add to cart"}
      </button>
      {message && <p role="status" className="mt-2 text-xs text-emerald-900">{message}</p>}
    </div>
  );
}
