"use client";

export default function ErrorPage({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
  return (
    <main className="mx-auto flex min-h-[70vh] max-w-3xl items-center px-6 py-14">
      <section className="w-full rounded-3xl border border-red-900/10 bg-white p-10 shadow-sm">
        <p className="text-sm font-bold uppercase tracking-widest text-red-700">Catalog unavailable</p>
        <h1 className="mt-3 text-4xl font-black text-emerald-950">We could not load the products.</h1>
        <p className="mt-4 leading-7 text-emerald-950/60">The API may be starting or PostgreSQL may not be ready. Try again after checking the readiness endpoint.</p>
        {error.digest ? <p className="mt-2 font-mono text-xs text-emerald-950/40">Reference: {error.digest}</p> : null}
        <button onClick={reset} className="mt-7 rounded-full bg-emerald-950 px-6 py-3 font-bold text-white">Try again</button>
      </section>
    </main>
  );
}
