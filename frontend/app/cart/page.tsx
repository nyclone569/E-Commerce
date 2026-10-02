import Link from "next/link";
import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { formatMoney } from "@/lib/catalog";
import type { Cart } from "@/lib/cart";
import { CartControls } from "./cart-controls";

export const dynamic = "force-dynamic";

export default async function CartPage() {
  const session = (await cookies()).get("aurora_session");
  if (!session) redirect("/login");

  const backendURL = process.env.BACKEND_URL ?? "http://localhost:8080";
  const response = await fetch(`${backendURL}/api/cart`, {
    headers: { Cookie: `${session.name}=${session.value}` },
    cache: "no-store",
    signal: AbortSignal.timeout(5000),
  });
  if (response.status === 401) redirect("/login");
  if (!response.ok) throw new Error("Cart is temporarily unavailable");
  const cart = (await response.json()) as Cart;

  return (
    <main className="mx-auto max-w-4xl px-6 py-12">
      <Link href="/" className="text-sm font-bold text-emerald-800">← Continue shopping</Link>
      <h1 className="mt-6 text-5xl font-black text-emerald-950">Your cart</h1>
      <p className="mt-3 text-emerald-950/60">Prices are current estimates. Adding to cart does not reserve stock.</p>
      {cart.items.length === 0 ? (
        <section className="mt-10 rounded-3xl border border-dashed border-emerald-900/30 bg-white px-8 py-16 text-center">
          <p className="text-2xl font-bold text-emerald-950">Your cart is empty.</p>
          <Link href="/" className="mt-4 inline-block font-bold text-emerald-800">Browse products</Link>
        </section>
      ) : (
        <>
          <ul className="mt-10 space-y-4">
            {cart.items.map((item) => (
              <li key={item.sku_id} className="rounded-2xl border border-emerald-950/10 bg-white p-6">
                <div className="flex flex-wrap items-start justify-between gap-4">
                  <div>
                    <Link href={`/products/${item.product_slug}`} className="text-lg font-bold text-emerald-950 hover:underline">{item.product_name}</Link>
                    <p className="text-sm text-emerald-950/60">{item.sku_code} · {formatMoney(item.unit_price_minor, "VND")} each</p>
                  </div>
                  <p className="text-lg font-bold text-emerald-800">{formatMoney(item.line_total_minor, "VND")}</p>
                </div>
                <CartControls skuID={item.sku_id} quantity={item.quantity} />
              </li>
            ))}
          </ul>
          <div className="mt-8 flex justify-between border-t border-emerald-950/10 pt-6 text-xl font-black text-emerald-950">
            <span>Estimated subtotal</span><span>{formatMoney(cart.subtotal_minor, "VND")}</span>
          </div>
          <p className="mt-4 text-sm text-emerald-950/60">Checkout and payment arrive later in Milestone 2.</p>
        </>
      )}
    </main>
  );
}
