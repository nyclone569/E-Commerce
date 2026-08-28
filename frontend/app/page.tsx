import Link from "next/link";
import { formatMoney, getProducts } from "@/lib/catalog";

type HomeProps = {
  searchParams: Promise<{ page?: string }>;
};

export default async function Home({ searchParams }: HomeProps) {
  const params = await searchParams;
  const parsedPage = Number.parseInt(params.page ?? "1", 10);
  const page = Number.isFinite(parsedPage) && parsedPage > 0 ? parsedPage : 1;
  const result = await getProducts(page);

  return (
    <main className="mx-auto max-w-6xl px-6 py-14">
      <section className="mb-12 grid gap-6 md:grid-cols-[1fr_auto] md:items-end">
        <div>
          <p className="mb-3 text-sm font-bold uppercase tracking-[0.25em] text-amber-700">Catalog / Foundation</p>
          <h1 className="max-w-3xl text-5xl font-black leading-[0.95] tracking-tight text-emerald-950 md:text-7xl">
            Useful things, built on visible decisions.
          </h1>
        </div>
        <p className="max-w-xs border-l-2 border-amber-500 pl-4 text-sm leading-6 text-emerald-950/70">
          This first slice reads catalog data through Next.js, Go, sqlc, and PostgreSQL.
        </p>
      </section>

      {result.products.length === 0 ? (
        <section className="rounded-3xl border border-dashed border-emerald-900/30 bg-white/60 px-8 py-20 text-center">
          <p className="text-3xl font-black text-emerald-950">The shelf is ready.</p>
          <p className="mt-3 text-emerald-950/60">Create the first product through the catalog API.</p>
          <code className="mt-6 inline-block rounded-lg bg-emerald-950 px-4 py-2 text-sm text-emerald-50">POST /api/products</code>
        </section>
      ) : (
        <>
          <section className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
            {result.products.map((product, index) => {
              const firstSKU = product.skus[0];
              return (
                <Link key={product.id} href={`/products/${product.slug}`} className="group rounded-3xl focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-emerald-800">
                  <article className="h-full overflow-hidden rounded-3xl border border-emerald-950/10 bg-white shadow-sm transition group-hover:-translate-y-1 group-hover:shadow-xl">
                    <div className="flex aspect-[4/3] items-end bg-gradient-to-br from-emerald-100 via-teal-50 to-amber-100 p-6">
                      <span className="font-mono text-6xl font-black text-emerald-950/10">{String(index + 1).padStart(2, "0")}</span>
                    </div>
                    <div className="p-6">
                      <p className="text-xs font-bold uppercase tracking-widest text-amber-700">{firstSKU?.code ?? "NO-SKU"}</p>
                      <h2 className="mt-2 text-2xl font-black text-emerald-950">{product.name}</h2>
                      <p className="mt-3 line-clamp-2 min-h-12 text-sm leading-6 text-emerald-950/60">{product.description || "No description yet."}</p>
                      <p className="mt-5 text-lg font-bold text-emerald-800">
                        {firstSKU ? formatMoney(firstSKU.price_cents, firstSKU.currency) : "Price unavailable"}
                      </p>
                    </div>
                  </article>
                </Link>
              );
            })}
          </section>

          <nav className="mt-10 flex items-center justify-between" aria-label="Product pagination">
            {result.page > 1 ? <Link className="rounded-full border border-emerald-900/20 px-5 py-2 font-bold" href={`/?page=${result.page - 1}`}>Previous</Link> : <span />}
            <span className="text-sm text-emerald-950/60">Page {result.page} of {result.total_pages}</span>
            {result.page < result.total_pages ? <Link className="rounded-full bg-emerald-950 px-5 py-2 font-bold text-white" href={`/?page=${result.page + 1}`}>Next</Link> : <span />}
          </nav>
        </>
      )}
    </main>
  );
}
