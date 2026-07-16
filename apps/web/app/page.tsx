import Link from "next/link";

export default function HomePage() {
  return (
    <main className="mx-auto max-w-3xl px-6 py-16">
      <h1 className="text-3xl font-semibold">ZarishLog</h1>
      <p className="mt-2 text-slate-600">
        Humanitarian logistics, supply chain, and asset management — offline-first, open source. Go + Next.js stack.
      </p>
      <div className="mt-8 flex gap-4">
        <Link href="/products" className="rounded-md bg-slate-900 px-4 py-2 text-white hover:bg-slate-800">
          View Product Catalogue
        </Link>
        <a
          href={`${process.env.NEXT_PUBLIC_API_URL}/api/v1/health`}
          className="rounded-md border border-slate-300 px-4 py-2 hover:bg-slate-100"
        >
          API Health
        </a>
      </div>
    </main>
  );
}
