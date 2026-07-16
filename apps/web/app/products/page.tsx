// Server component: fetches the seeded product catalogue directly from the API.
// This is the first end-to-end demoable slice per BLUEPRINT.md Phase 1/2:
// seeded DB -> NestJS API -> Next.js UI.

interface Product {
  id: string;
  sku: string;
  name: string;
  itemType: string;
  status: string;
  category: { name: string };
  uom: { code: string };
}

async function getProducts(): Promise<Product[]> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:4000";
  // DEV-ONLY placeholder org id — matches the fixed id set for ORG_001 in
  // seed.ts. Replace with the authenticated user's organization once the
  // auth module (Keycloak/OIDC) is wired in.
  const res = await fetch(`${apiUrl}/products`, {
    headers: { "x-organization-id": "org_001_seed" },
    cache: "no-store",
  });
  if (!res.ok) return [];
  return res.json();
}

export default async function ProductsPage() {
  const products = await getProducts();

  return (
    <main className="mx-auto max-w-5xl px-6 py-12">
      <h1 className="text-2xl font-semibold">Product Catalogue</h1>
      <p className="mt-1 text-sm text-slate-500">{products.length} items loaded from the master catalogue</p>

      <div className="mt-6 overflow-hidden rounded-lg border border-slate-200">
        <table className="w-full text-left text-sm">
          <thead className="bg-slate-100 text-slate-600">
            <tr>
              <th className="px-4 py-2">SKU</th>
              <th className="px-4 py-2">Name</th>
              <th className="px-4 py-2">Category</th>
              <th className="px-4 py-2">Type</th>
              <th className="px-4 py-2">UoM</th>
              <th className="px-4 py-2">Status</th>
            </tr>
          </thead>
          <tbody>
            {products.map((p) => (
              <tr key={p.id} className="border-t border-slate-100">
                <td className="px-4 py-2 font-mono text-xs">{p.sku}</td>
                <td className="px-4 py-2">{p.name}</td>
                <td className="px-4 py-2">{p.category?.name}</td>
                <td className="px-4 py-2">{p.itemType}</td>
                <td className="px-4 py-2">{p.uom?.code}</td>
                <td className="px-4 py-2">{p.status}</td>
              </tr>
            ))}
            {products.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-slate-400">
                  No products found. Run <code>pnpm db:seed</code> and confirm the API is running.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </main>
  );
}
