export type SKU = {
  id: string;
  product_id: string;
  code: string;
  price_minor: number;
  currency: string;
};

export type Product = {
  id: string;
  name: string;
  slug: string;
  description: string;
  skus: SKU[];
  created_at: string;
  updated_at: string;
};

export type ProductPage = {
  products: Product[];
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
};

export class CatalogNotFoundError extends Error {
  constructor() {
    super("Product was not found");
    this.name = "CatalogNotFoundError";
  }
}

export async function getProducts(page: number): Promise<ProductPage> {
  const backendURL = process.env.BACKEND_URL ?? "http://localhost:8080";
  const response = await fetch(`${backendURL}/api/products?page=${page}&page_size=12`, {
    cache: "no-store",
    signal: AbortSignal.timeout(5000),
  });
  if (!response.ok) {
    throw new Error(`Catalog request failed with status ${response.status}`);
  }
  return response.json() as Promise<ProductPage>;
}

export async function getProduct(slug: string): Promise<Product> {
  const backendURL = process.env.BACKEND_URL ?? "http://localhost:8080";
  const response = await fetch(`${backendURL}/api/products/${encodeURIComponent(slug)}`, {
    cache: "no-store",
    signal: AbortSignal.timeout(5000),
  });
  if (response.status === 404) {
    throw new CatalogNotFoundError();
  }
  if (!response.ok) {
    throw new Error(`Catalog request failed with status ${response.status}`);
  }
  return response.json() as Promise<Product>;
}

export function formatMoney(priceMinor: number, currency: string): string {
  return new Intl.NumberFormat(currency === "VND" ? "vi-VN" : "en-US", {
    style: "currency",
    currency,
  }).format(currency === "VND" ? priceMinor : priceMinor / 100);
}
