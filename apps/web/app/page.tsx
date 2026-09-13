import Image from "next/image";
import Link from "next/link";

const FEATURES = [
  {
    title: "Offline-first PWA",
    body: "Receive, issue, transfer, and adjust with zero connectivity. Workbox + IndexedDB queue everything and sync when you're back online.",
  },
  {
    title: "Multi-tenant by construction",
    body: "Every table carries an org_id and PostgreSQL Row-Level Security enforces isolation — built in, not bolted on.",
  },
  {
    title: "Config-as-files + Config Studio",
    body: "Master catalogues, forms, templates, and roles live as readable JSON/CSV/MD. The /config-studio page turns them into live, editable diagrams.",
  },
];

export default function HomePage() {
  return (
    <main>
      <section className="relative overflow-hidden bg-[linear-gradient(135deg,#0f766e_0%,#134e4a_55%,#0f172a_100%)] text-slate-50">
        <div
          className="pointer-events-none absolute inset-0 opacity-[0.07]"
          style={{
            backgroundImage: "radial-gradient(circle at 1px 1px, #ffffff 1px, transparent 0)",
            backgroundSize: "28px 28px",
          }}
        />

        <div className="relative mx-auto max-w-6xl px-6 py-20 sm:py-24">
          <div className="flex items-center gap-5">
            <Image
              src="/logo.svg"
              alt="ZarishLog logo"
              width={84}
              height={84}
              priority
              className="rounded-2xl shadow-lg shadow-black/20"
            />
            <div>
              <h1 className="text-4xl font-bold tracking-tight text-white sm:text-5xl">ZarishLog</h1>
              <p className="mt-1 text-lg font-medium text-brand-200">Humanitarian logistics, made calm.</p>
            </div>
          </div>

          <p className="mt-8 max-w-2xl text-lg leading-relaxed text-slate-100/90">
            Offline-first, multi-tenant supply chain and asset management for relief operations on the ground.
            Go + PostgreSQL backend, Next.js PWA frontend, and a no-code workspace that lets administrators
            steer the whole configuration visually.
          </p>

          <div className="mt-9 flex flex-wrap items-center gap-3">
            <Link
              href="/products"
              className="rounded-lg bg-accent-500 px-5 py-2.5 text-sm font-semibold text-slate-900 shadow-sm transition-colors hover:bg-accent-400"
            >
              View Product Catalogue
            </Link>
            <Link
              href="/config-studio"
              className="rounded-lg border border-white/30 px-5 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-white/10"
            >
              Open Config Studio
            </Link>
            <a
              href="https://github.com/cpintl/ZarishLog"
              className="rounded-lg border border-white/20 px-5 py-2.5 text-sm font-medium text-slate-200 transition-colors hover:bg-white/5"
            >
              Source on GitHub ↗
            </a>
          </div>
        </div>

        <div className="relative h-2 bg-gradient-to-r from-brand-400 via-accent-500 to-brand-400" />
      </section>

      <section className="mx-auto grid max-w-6xl gap-4 px-6 py-12 sm:grid-cols-3">
        {FEATURES.map((feature) => (
          <article key={feature.title} className="rounded-xl border border-brand-100 bg-white p-6 shadow-sm">
            <div className="mb-3 h-1.5 w-10 rounded-full bg-brand-500" />
            <h2 className="text-base font-semibold text-slate-800">{feature.title}</h2>
            <p className="mt-2 text-sm leading-relaxed text-slate-600">{feature.body}</p>
          </article>
        ))}
      </section>

      <section className="border-t border-slate-200 bg-white">
        <div className="mx-auto flex max-w-6xl flex-wrap items-center gap-x-8 gap-y-3 px-6 py-6 text-sm text-slate-600">
          <span className="font-semibold text-slate-800">Where to next:</span>
          <Link href="/products" className="text-brand-700 underline-offset-4 hover:underline">
            Product catalogue
          </Link>
          <Link href="/status" className="text-brand-700 underline-offset-4 hover:underline">
            Project status
          </Link>
          <Link href="/config-studio" className="text-brand-700 underline-offset-4 hover:underline">
            Config Studio
          </Link>
          <a
            href={`${process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080"}/api/v1/health`}
            className="text-brand-700 underline-offset-4 hover:underline"
          >
            API health
          </a>
        </div>
      </section>
    </main>
  );
}