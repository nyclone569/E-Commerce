import { afterEach, describe, expect, it, vi } from "vitest";
import { AuthAPIError, login, logout, register } from "./auth";

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("authentication API client", () => {
  it("registers through the same-origin API path without storing a token", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ data: { id: "user-1", email: "learner@example.com", display_name: "Learner", status: "active", created_at: "now", updated_at: "now" } }), {
        status: 201,
        headers: { "Content-Type": "application/json" },
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    const user = await register({ email: "learner@example.com", password: "a sufficiently long passphrase", display_name: "Learner" });

    expect(user.id).toBe("user-1");
    expect(fetchMock).toHaveBeenCalledWith("/api/auth/register", expect.objectContaining({ method: "POST", credentials: "same-origin" }));
    const request = fetchMock.mock.calls[0][1] as RequestInit;
    expect(JSON.parse(String(request.body))).toEqual({ email: "learner@example.com", password: "a sufficiently long passphrase", display_name: "Learner" });
  });

  it("preserves bounded backend validation details", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ error: { code: "validation_failed", message: "Request validation failed", details: { password: "must be longer" } } }), {
        status: 400,
        headers: { "Content-Type": "application/json" },
      }),
    ));

    await expect(login({ email: "learner@example.com", password: "short" })).rejects.toMatchObject({
      code: "validation_failed",
      details: { password: "must be longer" },
    } satisfies Partial<AuthAPIError>);
  });

  it("uses idempotent logout without expecting a JSON body", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }));
    vi.stubGlobal("fetch", fetchMock);

    await expect(logout()).resolves.toBeUndefined();
    expect(fetchMock).toHaveBeenCalledWith("/api/auth/logout", { method: "POST", credentials: "same-origin" });
  });
});
