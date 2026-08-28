export type CurrentUser = {
  id: string;
  email: string;
  display_name: string;
  status: string;
  created_at: string;
  updated_at: string;
};

type UserEnvelope = {
  data: CurrentUser;
};

type ErrorEnvelope = {
  error?: {
    code?: string;
    message?: string;
    details?: Record<string, string>;
  };
};

export class AuthAPIError extends Error {
  readonly code: string;
  readonly details: Record<string, string>;

  constructor(code: string, message: string, details: Record<string, string> = {}) {
    super(message);
    this.name = "AuthAPIError";
    this.code = code;
    this.details = details;
  }
}

async function authRequest(path: string, body?: object): Promise<CurrentUser> {
  const response = await fetch(path, {
    method: "POST",
    headers: body ? { "Content-Type": "application/json" } : undefined,
    body: body ? JSON.stringify(body) : undefined,
    credentials: "same-origin",
  });
  if (!response.ok) {
    let envelope: ErrorEnvelope = {};
    try {
      envelope = (await response.json()) as ErrorEnvelope;
    } catch {
      // Use the bounded fallback below when an intermediary did not return JSON.
    }
    throw new AuthAPIError(
      envelope.error?.code ?? "request_failed",
      envelope.error?.message ?? "The request could not be completed",
      envelope.error?.details,
    );
  }
  return ((await response.json()) as UserEnvelope).data;
}

export function register(input: { email: string; password: string; display_name: string }) {
  return authRequest("/api/auth/register", input);
}

export function login(input: { email: string; password: string }) {
  return authRequest("/api/auth/login", input);
}

export async function logout(): Promise<void> {
  const response = await fetch("/api/auth/logout", {
    method: "POST",
    credentials: "same-origin",
  });
  if (!response.ok) {
    throw new AuthAPIError("logout_failed", "Logout could not be completed");
  }
}
