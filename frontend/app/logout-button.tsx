"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { logout } from "@/lib/auth";

export function LogoutButton() {
  const router = useRouter();
  const [pending, setPending] = useState(false);

  async function handleLogout() {
    setPending(true);
    try {
      await logout();
      router.push("/");
      router.refresh();
    } finally {
      setPending(false);
    }
  }

  return (
    <button type="button" onClick={handleLogout} disabled={pending} className="text-sm font-bold text-emerald-800 hover:text-emerald-950 disabled:opacity-50">
      {pending ? "Signing out…" : "Sign out"}
    </button>
  );
}
