import "server-only";

import { cookies } from "next/headers";
import type { CurrentUser } from "@/lib/auth";

type UserEnvelope = {
  data: CurrentUser;
};

export async function getCurrentUser(): Promise<CurrentUser | null> {
  const session = (await cookies()).get("aurora_session");
  if (!session) return null;

  const backendURL = process.env.BACKEND_URL ?? "http://localhost:8080";
  try {
    const response = await fetch(`${backendURL}/api/me`, {
      headers: { Cookie: `${session.name}=${session.value}` },
      cache: "no-store",
      signal: AbortSignal.timeout(3000),
    });
    if (!response.ok) return null;
    return ((await response.json()) as UserEnvelope).data;
  } catch {
    return null;
  }
}
