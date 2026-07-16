export default async function StatusPage() {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

  async function fetchJson(url: string) {
    try {
      const res = await fetch(url, { cache: "no-store" });
      if (!res.ok) return null;
      return await res.json();
    } catch {
      return null;
    }
  }

  const health = await fetchJson(`${apiUrl}/api/v1/health`);
  const version = await fetchJson(`${apiUrl}/api/v1/version`);
  const products = await fetchJson(`${apiUrl}/api/v1/products`);

  const isHealthy = health?.status === "healthy";
  const productCount = products?.data?.length ?? 0;

  return (
    <main className="mx-auto max-w-3xl px-6 py-12">
      <h1 className="text-2xl font-semibold">System Status</h1>
      <p className="mt-1 text-sm text-slate-500">ZarishLog development sandbox status dashboard</p>

      <div className="mt-8 grid gap-4 sm:grid-cols-2">
        {/* API Server */}
        <div className={`rounded-lg border p-4 ${isHealthy ? "border-green-200 bg-green-50" : "border-red-200 bg-red-50"}`}>
          <div className="flex items-center justify-between">
            <h2 className="font-medium">API Server</h2>
            <span className={`h-3 w-3 rounded-full ${isHealthy ? "bg-green-500" : "bg-red-500"}`} />
          </div>
          <p className="mt-2 text-sm text-slate-600">
            {isHealthy ? "Healthy and responding" : "Not responding"}
          </p>
          <p className="text-xs text-slate-400">{apiUrl}</p>
        </div>

        {/* Database */}
        <div className={`rounded-lg border p-4 ${health?.db === "connected" ? "border-green-200 bg-green-50" : "border-yellow-200 bg-yellow-50"}`}>
          <div className="flex items-center justify-between">
            <h2 className="font-medium">Database</h2>
            <span className={`h-3 w-3 rounded-full ${health?.db === "connected" ? "bg-green-500" : "bg-yellow-500"}`} />
          </div>
          <p className="mt-2 text-sm text-slate-600">
            {health?.db === "connected" ? "Connected" : "Disconnected"}
          </p>
          <p className="text-xs text-slate-400">PostgreSQL 18</p>
        </div>

        {/* Product Catalogue */}
        <div className={`rounded-lg border p-4 ${productCount > 0 ? "border-green-200 bg-green-50" : "border-yellow-200 bg-yellow-50"}`}>
          <div className="flex items-center justify-between">
            <h2 className="font-medium">Product Catalogue</h2>
            <span className={`h-3 w-3 rounded-full ${productCount > 0 ? "bg-green-500" : "bg-yellow-500"}`} />
          </div>
          <p className="mt-2 text-sm text-slate-600">
            {productCount > 0 ? `${productCount} products loaded` : "No products"}
          </p>
        </div>

        {/* Version */}
        <div className="rounded-lg border border-slate-200 bg-slate-50 p-4">
          <h2 className="font-medium">Version</h2>
          <p className="mt-2 text-sm text-slate-600">
            {version?.version ?? "unknown"}
          </p>
          <p className="text-xs text-slate-400">
            Go {version?.go_version ?? "?"} · {version?.go_arch ?? "?"}
          </p>
        </div>
      </div>

      {/* Test Runner Section */}
      <div className="mt-8 rounded-lg border border-slate-200 p-6">
        <h2 className="font-medium">Test Runner</h2>
        <p className="mt-1 text-sm text-slate-500">Run tests from the terminal:</p>
        <div className="mt-4 space-y-2">
          <div className="rounded bg-slate-900 px-4 py-3 text-sm text-green-400 font-mono">
            <span className="text-slate-500">$</span> ./scripts/test.sh
          </div>
          <div className="rounded bg-slate-900 px-4 py-3 text-sm text-green-400 font-mono">
            <span className="text-slate-500">$</span> make test
          </div>
          <div className="rounded bg-slate-900 px-4 py-3 text-sm text-green-400 font-mono">
            <span className="text-slate-500">$</span> make test-coverage
          </div>
        </div>
      </div>

      {/* Configuration Section */}
      <div className="mt-8 rounded-lg border border-slate-200 p-6">
        <h2 className="font-medium">Configuration</h2>
        <p className="mt-1 text-sm text-slate-500">Validate and apply configuration:</p>
        <div className="mt-4 space-y-2">
          <div className="rounded bg-slate-900 px-4 py-3 text-sm text-green-400 font-mono">
            <span className="text-slate-500">$</span> ./scripts/validate-config.sh
          </div>
          <div className="rounded bg-slate-900 px-4 py-3 text-sm text-green-400 font-mono">
            <span className="text-slate-500">$</span> make db-seed
          </div>
        </div>
      </div>
    </main>
  );
}
