import Link from "next/link";

export default function ProductNotFound() {
  return (
    <main className="mx-auto flex min-h-[70vh] max-w-3xl items-center px-6 py-14">
      <section className="w-full rounded-3xl border border-emerald-950/10 bg-white p-10 shadow-sm">
        <p className="text-sm font-bold uppercase tracking-widest text-amber-700">Product not found</p>
        <h1 className="mt-3 text-4xl font-black text-emerald-950">This product is not on the shelf.</h1>
        <p className="mt-4 leading-7 text-emerald-950/60">The link may be incorrect, or the product may no longer exist.</p>
        <Link href="/" className="mt-7 inline-block rounded-full bg-emerald-950 px-6 py-3 font-bold text-white">Return to catalog</Link>
      </section>
    </main>
  );
}
