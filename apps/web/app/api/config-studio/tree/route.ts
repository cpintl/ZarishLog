import { NextResponse } from "next/server";
import { listConfigFiles, listDocsFiles } from "@/lib/config-studio/server/fs";

export const runtime = "nodejs";

export async function GET() {
  try {
    const [config, docs] = await Promise.all([listConfigFiles(), listDocsFiles()]);
    return NextResponse.json({ config, docs });
  } catch (err) {
    return NextResponse.json(
      { error: err instanceof Error ? err.message : "Failed to list files" },
      { status: 500 },
    );
  }
}
