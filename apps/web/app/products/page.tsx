interface Product {
  id: string;
  sku: string;
  name: string;
  category?: { name: string };
  item_type: string;
  status: string;
}

interface ProductsResponse {
  data: Product[];
}

async function getProducts(): Promise<Product[]> {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
  const res = await fetch(`${apiUrl}/api/v1/products`, {
    cache: "no-store",
  });
  if (!res.ok) return [];
  const body: ProductsResponse = await res.json();
  return body.data ?? [];
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
                <td className="px-4 py-2">{p.item_type}</td>
                <td className="px-4 py-2">-</td>
                <td className="px-4 py-2">{p.status}</td>
              </tr>
            ))}
            {products.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-slate-400">
                  No products found. Run <code>make db-seed</code> and confirm the API is running.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </main>
  );
}
