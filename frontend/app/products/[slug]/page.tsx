import Link from "next/link";
import { notFound } from "next/navigation";
import { CatalogNotFoundError, formatMoney, getProduct, type Product } from "@/lib/catalog";
import { AddToCartButton } from "./add-to-cart-button";

type ProductDetailsProps = {
  params: Promise<{ slug: string }>;
};

export default async function ProductDetails({ params }: ProductDetailsProps) {
  const { slug } = await params;
  let product: Product;
  try {
    product = await getProduct(slug);
  } catch (error) {
    if (error instanceof CatalogNotFoundError) {
      notFound();
    }
    throw error;
  }

  return (
    <main className="mx-auto max-w-6xl px-6 py-12">
      <Link href="/" className="text-sm font-bold text-emerald-800 hover:text-emerald-950">
        ← Back to catalog
      </Link>
      <div className="mt-8 grid gap-10 lg:grid-cols-[1.05fr_0.95fr]">
        <div className="flex min-h-[28rem] items-end rounded-[2.5rem] bg-gradient-to-br from-emerald-100 via-teal-50 to-amber-100 p-10">
          <span className="font-mono text-8xl font-black text-emerald-950/10">AURORA</span>
        </div>
        <section className="py-4">
          <p className="text-sm font-bold uppercase tracking-[0.25em] text-amber-700">Product details</p>
          <h1 className="mt-4 text-5xl font-black tracking-tight text-emerald-950">{product.name}</h1>
          <p className="mt-6 text-lg leading-8 text-emerald-950/65">{product.description || "No description yet."}</p>

          <div className="mt-10">
            <h2 className="text-sm font-black uppercase tracking-widest text-emerald-950">Available SKUs</h2>
            <div className="mt-4 grid gap-3">
              {product.skus.map((sku) => (
                <div key={sku.id} className="flex items-center justify-between rounded-2xl border border-emerald-950/10 bg-white px-5 py-4">
                  <div>
                    <p className="font-bold text-emerald-950">{sku.code}</p>
                    <p className="mt-1 text-xs uppercase tracking-widest text-emerald-950/45">SKU option</p>
                  </div>
                  <div className="text-right">
                    <p className="text-xl font-black text-emerald-800">{formatMoney(sku.price_minor, sku.currency)}</p>
                    {sku.currency === "VND" && <AddToCartButton skuID={sku.id} />}
                  </div>
                </div>
              ))}
            </div>
          </div>

          <aside className="mt-8 rounded-2xl bg-emerald-950 px-5 py-4 text-sm leading-6 text-emerald-50/80">
            Cart quantities are estimates at current catalog prices. Stock and final price are checked during checkout in a later milestone.
          </aside>
        </section>
      </div>
    </main>
  );
}
