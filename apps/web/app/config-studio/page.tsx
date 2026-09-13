"use client";

import Link from "next/link";
import { useConfigStudio } from "@/hooks/useConfigStudio";
import type { StudioDomain } from "@/lib/config-studio/types";
import { GraphCanvas } from "./components/GraphCanvas";
import { DetailsDrawer } from "./components/DetailsDrawer";

const DOMAINS: Array<{ id: StudioDomain; label: string }> = [
  { id: "overview", label: "Overview" },
  { id: "config", label: "Forms & Config" },
  { id: "metadata", label: "Master Data & Roles" },
  { id: "business-logic", label: "Business Logic" },
  { id: "setup", label: "Setup & Dependencies" },
  { id: "docs", label: "Documents" },
];

const NAV = [
  { href: "/", label: "Home" },
  { href: "/products", label: "Products" },
  { href: "/status", label: "Status" },
  { href: "/config-studio", label: "Config Studio", active: true },
];

export default function ConfigStudioPage() {
  const studio = useConfigStudio();

  return (
    <div className="flex h-dvh flex-col bg-slate-50">
      <header className="flex h-14 shrink-0 items-center gap-6 border-b border-slate-200 bg-white px-5">
        <div className="flex items-center gap-2">
          <span className="rounded bg-brand-700 px-1.5 py-0.5 text-[11px] font-bold tracking-wide text-white">
            ZL
          </span>
          <nav className="flex items-center gap-1">
            {NAV.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                className={`rounded-md px-2 py-1 text-xs font-medium ${
                  item.active
                    ? "bg-brand-700 text-white"
                    : "text-slate-500 hover:bg-slate-100 hover:text-slate-700"
                }`}
              >
                {item.label}
              </Link>
            ))}
          </nav>
        </div>
        <h1 className="hidden text-sm font-semibold text-slate-800 md:block">Config Studio</h1>
      </header>

      <section className="shrink-0 border-b border-slate-200 bg-white px-5 pb-3 pt-2">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div>
            <h2 className="text-base font-semibold text-slate-800">
              {studio.payload?.title ?? "Loading…"}
            </h2>
            <p className="text-xs text-slate-500">
              {studio.payload?.subtitle ?? "Building the live system graph…"}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <span className="flex items-center gap-1.5 text-[11px] text-slate-400">
              <span
                className={`h-2 w-2 rounded-full ${studio.payloadLoading ? "bg-amber-400" : "bg-emerald-500"}`}
              />
              {studio.payloadLoading ? "refreshing" : "live"}
            </span>
            <button
              onClick={() => void studio.refresh()}
              disabled={studio.payloadLoading}
              className="rounded-md border border-slate-300 px-2.5 py-1 text-xs font-medium text-slate-600 hover:bg-slate-50 disabled:opacity-50"
            >
              ↻ Refresh graph
            </button>
          </div>
        </div>
        <div className="mt-2 flex flex-wrap gap-1">
          {DOMAINS.map((d) => (
            <button
              key={d.id}
              onClick={() => studio.setDomain(d.id)}
              className={`rounded-md px-2.5 py-1 text-xs font-medium transition-colors ${
                studio.domain === d.id
                  ? "bg-sky-600 text-white"
                  : "bg-slate-100 text-slate-600 hover:bg-slate-200"
              }`}
            >
              {d.label}
            </button>
          ))}
        </div>
      </section>

      {studio.notice && (
        <div
          className={`fixed left-1/2 top-20 z-50 -translate-x-1/2 rounded-md px-3 py-2 text-xs font-medium shadow-lg ${
            studio.notice.type === "success"
              ? "bg-emerald-600 text-white"
              : "bg-rose-600 text-white"
          }`}
        >
          {studio.notice.text}
        </div>
      )}

      <main className="relative flex min-h-0 flex-1">
        {studio.payloadError ? (
          <div className="m-auto max-w-sm rounded-lg border border-rose-200 bg-rose-50 p-5 text-center">
            <p className="text-sm font-medium text-rose-700">Could not load the system graph</p>
            <p className="mt-1 text-xs text-rose-600">{studio.payloadError}</p>
            <button
              onClick={() => void studio.refresh()}
              className="mt-3 rounded-md bg-rose-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-rose-500"
            >
              Retry
            </button>
          </div>
        ) : studio.payload ? (
          <div className="flex-1">
            <GraphCanvas
              nodes={studio.payload.nodes}
              edges={studio.payload.edges}
              onNodeClick={(id) => {
                const n = studio.payload?.nodes.find((x) => x.id === id);
                studio.select(id, n?.refs ?? []);
              }}
            />
            {!studio.selectedNode && (
              <p className="pointer-events-none absolute bottom-4 left-4 rounded-md bg-slate-900/85 px-3 py-1.5 text-[11px] text-white shadow">
                Click any node to inspect it · drag to pan · scroll to zoom
              </p>
            )}
          </div>
        ) : (
          <div className="m-auto flex items-center gap-3 text-slate-400">
            <span className="h-4 w-4 animate-spin rounded-full border-2 border-slate-300 border-t-sky-500" />
            <span className="text-sm">Building graph…</span>
          </div>
        )}

        {studio.selectedNode && (
          <div className="hidden w-[440px] shrink-0 sm:block">
            <DetailsDrawer
              node={studio.selectedNode}
              file={studio.file}
              fileLoading={studio.fileLoading}
              fileError={studio.fileError}
              edit={studio.edit}
              setEdit={studio.setEdit}
              save={() => void studio.save()}
              saving={studio.saving}
              savedAt={studio.savedAt}
              onClose={studio.closePanel}
            />
          </div>
        )}
      </main>
    </div>
  );
}
