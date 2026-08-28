export default function ProductDetailsLoading() {
  return (
    <main className="mx-auto grid max-w-6xl animate-pulse gap-10 px-6 py-20 lg:grid-cols-2" aria-busy="true">
      <div className="min-h-[28rem] rounded-[2.5rem] bg-emerald-950/10" />
      <div>
        <div className="h-14 rounded-2xl bg-emerald-950/10" />
        <div className="mt-6 h-28 rounded-2xl bg-emerald-950/10" />
        <div className="mt-10 h-20 rounded-2xl bg-emerald-950/10" />
      </div>
      <span className="sr-only">Loading product details</span>
    </main>
  );
}
