import type { NextRequest } from "next/server";

const maximumBodyBytes = 1 << 20;

type RouteContext = {
  params: Promise<{ path: string[] }>;
};

async function proxyRequest(request: NextRequest, context: RouteContext): Promise<Response> {
  const { path } = await context.params;
  const backendURL = process.env.BACKEND_URL ?? "http://localhost:8080";
  const upstreamURL = new URL(`/api/${path.map(encodeURIComponent).join("/")}`, backendURL);
  upstreamURL.search = request.nextUrl.search;

  const headers = new Headers();
  for (const name of ["content-type", "cookie", "x-csrf-token", "x-request-id", "sec-fetch-site"]) {
    const value = request.headers.get(name);
    if (value) headers.set(name, value);
  }
  headers.set("origin", request.headers.get("origin") ?? request.nextUrl.origin);

  let body: ArrayBuffer | undefined;
  if (request.method !== "GET" && request.method !== "HEAD") {
    const contentLength = Number.parseInt(request.headers.get("content-length") ?? "0", 10);
    if (Number.isFinite(contentLength) && contentLength > maximumBodyBytes) {
      return Response.json({ error: { code: "request_too_large", message: "Request body is too large" } }, { status: 413 });
    }
    body = await request.arrayBuffer();
    if (body.byteLength > maximumBodyBytes) {
      return Response.json({ error: { code: "request_too_large", message: "Request body is too large" } }, { status: 413 });
    }
  }

  try {
    const upstream = await fetch(upstreamURL, {
      method: request.method,
      headers,
      body,
      cache: "no-store",
      redirect: "manual",
      signal: request.signal,
    });
    const responseHeaders = new Headers();
    for (const name of ["content-type", "set-cookie", "x-request-id"]) {
      const value = upstream.headers.get(name);
      if (value) responseHeaders.set(name, value);
    }
    return new Response(upstream.body, {
      status: upstream.status,
      headers: responseHeaders,
    });
  } catch {
    return Response.json(
      { error: { code: "backend_unavailable", message: "The backend is temporarily unavailable" } },
      { status: 502 },
    );
  }
}

export const dynamic = "force-dynamic";

export function GET(request: NextRequest, context: RouteContext) {
  return proxyRequest(request, context);
}

export function POST(request: NextRequest, context: RouteContext) {
  return proxyRequest(request, context);
}
