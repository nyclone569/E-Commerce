import type { Metadata } from "next";
import Link from "next/link";
import { getCurrentUser } from "@/lib/auth-server";
import { LogoutButton } from "./logout-button";
import "./globals.css";

export const dynamic = "force-dynamic";

export const metadata: Metadata = {
  title: "AuroraShop",
  description: "A commerce system built to learn architecture through trade-offs.",
};

export default async function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  const user = await getCurrentUser();
  return (
    <html lang="en">
      <body className="min-h-screen">
        <header className="border-b border-emerald-950/10 bg-white/70 backdrop-blur">
          <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-5">
            <Link href="/" className="text-xl font-black tracking-tight text-emerald-950">
              Aurora<span className="text-amber-600">Shop</span>
            </Link>
            <nav className="flex items-center gap-4" aria-label="Account navigation">
              {user ? (
                <>
                  <span className="hidden text-sm text-emerald-950/65 sm:inline">Hello, <strong className="text-emerald-950">{user.display_name}</strong></span>
                  <LogoutButton />
                </>
              ) : (
                <>
                  <Link href="/login" className="text-sm font-bold text-emerald-800 hover:text-emerald-950">Sign in</Link>
                  <Link href="/register" className="rounded-full bg-emerald-950 px-4 py-2 text-sm font-bold text-white">Register</Link>
                </>
              )}
              <span className="hidden rounded-full bg-amber-100 px-3 py-1 text-xs font-bold uppercase tracking-widest text-amber-800 md:inline">
                M2 / Identity
              </span>
            </nav>
          </div>
        </header>
        {children}
      </body>
    </html>
  );
}
