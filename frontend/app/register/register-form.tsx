"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";
import { AuthAPIError, register } from "@/lib/auth";

export function RegisterForm() {
  const router = useRouter();
  const [error, setError] = useState("");
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [pending, setPending] = useState(false);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setPending(true);
    setError("");
    setFieldErrors({});
    const form = new FormData(event.currentTarget);
    try {
      await register({
        email: String(form.get("email") ?? ""),
        password: String(form.get("password") ?? ""),
        display_name: String(form.get("display_name") ?? ""),
      });
      router.push("/");
      router.refresh();
    } catch (caught) {
      if (caught instanceof AuthAPIError) {
        setError(caught.message);
        setFieldErrors(caught.details);
      } else {
        setError("Registration could not be completed");
      }
    } finally {
      setPending(false);
    }
  }

  return (
    <form onSubmit={submit} className="grid gap-5">
      <div>
        <label htmlFor="display_name" className="text-sm font-bold text-emerald-950">Display name</label>
        <input id="display_name" name="display_name" autoComplete="name" required maxLength={100} aria-describedby={fieldErrors.display_name ? "display-name-error" : undefined} className="mt-2 w-full rounded-xl border border-emerald-950/20 bg-white px-4 py-3 outline-none focus:border-emerald-700 focus:ring-2 focus:ring-emerald-700/15" />
        {fieldErrors.display_name ? <p id="display-name-error" className="mt-2 text-sm text-red-700">{fieldErrors.display_name}</p> : null}
      </div>
      <div>
        <label htmlFor="email" className="text-sm font-bold text-emerald-950">Email</label>
        <input id="email" name="email" type="email" autoComplete="email" required maxLength={254} aria-describedby={fieldErrors.email ? "email-error" : undefined} className="mt-2 w-full rounded-xl border border-emerald-950/20 bg-white px-4 py-3 outline-none focus:border-emerald-700 focus:ring-2 focus:ring-emerald-700/15" />
        {fieldErrors.email ? <p id="email-error" className="mt-2 text-sm text-red-700">{fieldErrors.email}</p> : null}
      </div>
      <div>
        <label htmlFor="password" className="text-sm font-bold text-emerald-950">Passphrase</label>
        <input id="password" name="password" type="password" autoComplete="new-password" required minLength={15} maxLength={128} aria-describedby={fieldErrors.password ? "password-error" : "password-help"} className="mt-2 w-full rounded-xl border border-emerald-950/20 bg-white px-4 py-3 outline-none focus:border-emerald-700 focus:ring-2 focus:ring-emerald-700/15" />
        <p id="password-help" className="mt-2 text-xs leading-5 text-emerald-950/55">15–128 characters; spaces are allowed and no character mixture is required.</p>
        {fieldErrors.password ? <p id="password-error" className="mt-2 text-sm text-red-700">{fieldErrors.password}</p> : null}
      </div>
      {error ? <p role="alert" className="rounded-xl bg-red-50 px-4 py-3 text-sm font-semibold text-red-800">{error}</p> : null}
      <button type="submit" disabled={pending} className="rounded-xl bg-emerald-950 px-5 py-3 font-bold text-white disabled:cursor-wait disabled:opacity-60">
        {pending ? "Creating account…" : "Create account"}
      </button>
    </form>
  );
}
