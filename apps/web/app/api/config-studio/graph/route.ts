import { NextResponse, type NextRequest } from "next/server";
import type { StudioDomain } from "@/lib/config-studio/types";
import { buildGraph, domainMeta } from "@/lib/config-studio/server/sources";

export const runtime = "nodejs";

const DOMAINS: StudioDomain[] = [
  "overview",
  "config",
  "metadata",
  "business-logic",
  "setup",
  "docs",
];

export async function GET(req: NextRequest) {
  const raw = req.nextUrl.searchParams.get("domain") ?? "overview";
  const domain: StudioDomain = DOMAINS.includes(raw as StudioDomain)
    ? (raw as StudioDomain)
    : "overview";
  try {
    const payload = await buildGraph(domain);
    return NextResponse.json({ ...payload, meta: domainMeta(domain) });
  } catch (err) {
    return NextResponse.json(
      { error: err instanceof Error ? err.message : "Failed to build graph" },
      { status: 500 },
    );
  }
}
