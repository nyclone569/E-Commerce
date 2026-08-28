"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";
import { AuthAPIError, login } from "@/lib/auth";

export function LoginForm() {
  const router = useRouter();
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setError("");
    const form = new FormData(event.currentTarget);
    try {
      await login({
        email: String(form.get("email") ?? ""),
        password: String(form.get("password") ?? ""),
      });
      router.push("/");
      router.refresh();
    } catch (caught) {
      setError(caught instanceof AuthAPIError ? caught.message : "Login could not be completed");
    } finally {
      setPending(false);
    }
  }

  return (
    <form onSubmit={submit} className="grid gap-5">
      <div>
        <label htmlFor="email" className="text-sm font-bold text-emerald-950">Email</label>
        <input id="email" name="email" type="email" autoComplete="email" required maxLength={254} className="mt-2 w-full rounded-xl border border-emerald-950/20 bg-white px-4 py-3 outline-none focus:border-emerald-700 focus:ring-2 focus:ring-emerald-700/15" />
      </div>
      <div>
        <label htmlFor="password" className="text-sm font-bold text-emerald-950">Password</label>
        <input id="password" name="password" type="password" autoComplete="current-password" required maxLength={128} className="mt-2 w-full rounded-xl border border-emerald-950/20 bg-white px-4 py-3 outline-none focus:border-emerald-700 focus:ring-2 focus:ring-emerald-700/15" />
      </div>
      {error ? <p role="alert" className="rounded-xl bg-red-50 px-4 py-3 text-sm font-semibold text-red-800">{error}</p> : null}
      <button type="submit" disabled={pending} className="rounded-xl bg-emerald-950 px-5 py-3 font-bold text-white disabled:cursor-wait disabled:opacity-60">
        {pending ? "Signing in…" : "Sign in"}
      </button>
    </form>
  );
}
