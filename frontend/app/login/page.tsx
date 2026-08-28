import Link from "next/link";
import { LoginForm } from "./login-form";

export default function LoginPage() {
  return (
    <main className="mx-auto grid min-h-[calc(100vh-5rem)] max-w-6xl items-center gap-12 px-6 py-14 lg:grid-cols-2">
      <section>
        <p className="text-sm font-bold uppercase tracking-[0.25em] text-amber-700">Identity / Login</p>
        <h1 className="mt-4 max-w-xl text-5xl font-black leading-none tracking-tight text-emerald-950 md:text-7xl">
          Continue your Aurora journey.
        </h1>
        <p className="mt-6 max-w-lg text-lg leading-8 text-emerald-950/65">
          The browser keeps an opaque cookie. Go verifies the session against PostgreSQL; the password and raw token never enter application logs.
        </p>
      </section>
      <section className="rounded-[2rem] border border-emerald-950/10 bg-white p-8 shadow-xl shadow-emerald-950/5">
        <LoginForm />
        <p className="mt-6 text-sm text-emerald-950/60">
          New here? <Link href="/register" className="font-bold text-emerald-800 hover:text-emerald-950">Create an account</Link>
        </p>
      </section>
    </main>
  );
}
