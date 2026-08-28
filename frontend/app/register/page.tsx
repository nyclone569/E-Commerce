import Link from "next/link";
import { RegisterForm } from "./register-form";

export default function RegisterPage() {
  return (
    <main className="mx-auto grid min-h-[calc(100vh-5rem)] max-w-6xl items-center gap-12 px-6 py-14 lg:grid-cols-2">
      <section>
        <p className="text-sm font-bold uppercase tracking-[0.25em] text-amber-700">Identity / Register</p>
        <h1 className="mt-4 max-w-xl text-5xl font-black leading-none tracking-tight text-emerald-950 md:text-7xl">
          One account, explicit security choices.
        </h1>
        <p className="mt-6 max-w-lg text-lg leading-8 text-emerald-950/65">
          Use a passphrase of at least 15 characters. AuroraShop uses Argon2id and does not require arbitrary uppercase, number, or symbol rules.
        </p>
      </section>
      <section className="rounded-[2rem] border border-emerald-950/10 bg-white p-8 shadow-xl shadow-emerald-950/5">
        <RegisterForm />
        <p className="mt-6 text-sm text-emerald-950/60">
          Already registered? <Link href="/login" className="font-bold text-emerald-800 hover:text-emerald-950">Sign in</Link>
        </p>
      </section>
    </main>
  );
}
