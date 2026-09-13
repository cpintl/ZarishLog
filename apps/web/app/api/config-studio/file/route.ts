import { NextResponse, type NextRequest } from "next/server";
import type { FileUpdateRequest } from "@/lib/config-studio/types";
import { serializeCsv } from "@/lib/config-studio/parse";
import { isWritable, writeRepoText } from "@/lib/config-studio/server/fs";
import { resolveFilePayload } from "@/lib/config-studio/server/sources";

export const runtime = "nodejs";

export async function GET(req: NextRequest) {
  const relPath = req.nextUrl.searchParams.get("path");
  if (!relPath) {
    return NextResponse.json({ error: "Missing ?path=" }, { status: 400 });
  }
  try {
    const payload = await resolveFilePayload(relPath);
    return NextResponse.json(payload);
  } catch (err) {
    return NextResponse.json(
      { error: err instanceof Error ? err.message : "Failed to read file" },
      { status: 400 },
    );
  }
}

export async function PUT(req: NextRequest) {
  let body: FileUpdateRequest;
  try {
    body = (await req.json()) as FileUpdateRequest;
  } catch {
    return NextResponse.json({ error: "Invalid JSON body" }, { status: 400 });
  }
  if (!body.path || typeof body.path !== "string") {
    return NextResponse.json({ error: "Missing path" }, { status: 400 });
  }
  if (!isWritable(body.path)) {
    return NextResponse.json(
      { error: "Only files under config/ and docs/ are writable from the studio" },
      { status: 403 },
    );
  }

  let content: string;
  switch (body.contentType) {
    case "json": {
      if (body.json === undefined)
        return NextResponse.json({ error: "Missing json payload" }, { status: 400 });
      try {
        // round-trip through JSON.parse/stringify to guarantee valid output
        JSON.stringify(body.json);
      } catch {
        return NextResponse.json({ error: "Invalid JSON payload" }, { status: 400 });
      }
      content = JSON.stringify(body.json, null, 2) + "\n";
      break;
    }
    case "csv": {
      if (!body.csv) return NextResponse.json({ error: "Missing csv payload" }, { status: 400 });
      content = serializeCsv(body.csv);
      break;
    }
    case "md": {
      if (typeof body.text !== "string")
        return NextResponse.json({ error: "Missing text" }, { status: 400 });
      content = body.text;
      break;
    }
    default:
      return NextResponse.json({ error: "Unsupported contentType" }, { status: 400 });
  }

  try {
    const entry = await writeRepoText(body.path, content);
    const payload = await resolveFilePayload(body.path);
    return NextResponse.json({ entry, payload });
  } catch (err) {
    return NextResponse.json(
      { error: err instanceof Error ? err.message : "Failed to write file" },
      { status: 500 },
    );
  }
}
