export default function Loading() {
  return (
    <main className="mx-auto max-w-6xl animate-pulse px-6 py-14" aria-busy="true">
      <div className="h-16 max-w-3xl rounded-2xl bg-emerald-950/10" />
      <div className="mt-12 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 6 }, (_, index) => <div key={index} className="aspect-[4/5] rounded-3xl bg-emerald-950/10" />)}
      </div>
      <span className="sr-only">Loading products</span>
    </main>
  );
}
