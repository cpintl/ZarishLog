import fs from "node:fs/promises";
import { existsSync } from "node:fs";
import path from "node:path";
import type { ConfigFileEntry, FileType } from "../types";

const FILE_EXTENSIONS: Record<string, FileType> = {
  ".json": "json",
  ".csv": "csv",
  ".md": "md",
  ".txt": "txt",
};

const READ_ONLY_REPO_FILES = new Set([
  "docker-compose.yml",
  ".env.example",
  "Makefile",
  "README.md",
  "SETUP.md",
  "CONFIGURE.md",
]);

const READ_ONLY_PREFIXES = ["config/", "docs/", "packages/business-logic/"];

let cachedRoot: string | undefined;

export function repoRoot(): string {
  if (cachedRoot) return cachedRoot;
  let dir = process.cwd();
  for (let i = 0; i < 8; i++) {
    if (existsSync(path.join(dir, "config")) && existsSync(path.join(dir, "docs"))) {
      cachedRoot = dir;
      return dir;
    }
    const parent = path.dirname(dir);
    if (parent === dir) break;
    dir = parent;
  }
  cachedRoot = process.cwd();
  return cachedRoot;
}

export function resolveRepoFile(relPath: string): {
  dir: "config" | "docs" | "root";
  absolute: string;
} {
  const cleaned = relPath.replace(/^\/+/, "").replace(/\\/g, "/");
  const root = repoRoot();
  const absolute = path.resolve(/*turbopackIgnore: true*/ root, cleaned);
  const prefix = `${root}${path.sep}`;
  if (!absolute.startsWith(prefix)) {
    throw new Error("Path escapes the repository root");
  }
  if (cleaned.startsWith("config/")) return { dir: "config", absolute };
  if (cleaned.startsWith("docs/")) return { dir: "docs", absolute };
  if (READ_ONLY_REPO_FILES.has(cleaned)) return { dir: "root", absolute };
  throw new Error("Path is outside the readable scope");
}

export function isWritable(relPath: string): boolean {
  const cleaned = relPath.replace(/^\/+/, "").replace(/\\/g, "/");
  if (!cleaned.startsWith("config/") && !cleaned.startsWith("docs/")) return false;
  const ext = path.extname(cleaned).toLowerCase();
  return Object.hasOwn(FILE_EXTENSIONS, ext);
}

export function isReadable(relPath: string): boolean {
  const cleaned = relPath.replace(/^\/+/, "").replace(/\\/g, "/");
  return READ_ONLY_PREFIXES.some((p) => cleaned.startsWith(p)) || READ_ONLY_REPO_FILES.has(cleaned);
}

function toEntry(root: "config" | "docs", absolute: string, rel: string): ConfigFileEntry | null {
  const ext = path.extname(absolute).toLowerCase();
  const type = FILE_EXTENSIONS[ext];
  if (!type) return null;
  const parts = rel.split("/");
  const name = parts.pop() ?? rel;
  return {
    path: rel,
    root,
    type,
    name,
    category: parts.slice(1).join("/"),
    sizeBytes: 0,
    modifiedAt: "",
  };
}

async function walkDir(
  dirAbs: string,
  baseRel: string,
  root: "config" | "docs",
): Promise<ConfigFileEntry[]> {
  const entries: ConfigFileEntry[] = [];
  const dirents = await fs.readdir(dirAbs, { withFileTypes: true });
  for (const dirent of dirents) {
    if (dirent.name.startsWith(".")) continue;
    const abs = path.join(dirAbs, dirent.name);
    const rel = `${baseRel}/${dirent.name}`;
    if (dirent.isDirectory()) {
      entries.push(...(await walkDir(abs, rel, root)));
      continue;
    }
    const entry = toEntry(root, abs, rel);
    if (!entry) continue;
    const stat = await fs.stat(abs);
    entry.sizeBytes = stat.size;
    entry.modifiedAt = stat.mtime.toISOString();
    entries.push(entry);
  }
  return entries.sort((a, b) => a.path.localeCompare(b.path));
}

export async function listConfigFiles(): Promise<ConfigFileEntry[]> {
  const root = path.join(/*turbopackIgnore: true*/ repoRoot(), "config");
  return walkDir(root, "config", "config");
}

export async function listDocsFiles(): Promise<ConfigFileEntry[]> {
  const root = path.join(/*turbopackIgnore: true*/ repoRoot(), "docs");
  return walkDir(root, "docs", "docs");
}

export async function readFileEntry(entry: ConfigFileEntry): Promise<string> {
  return fs.readFile(path.join(/*turbopackIgnore: true*/ repoRoot(), entry.path), "utf8");
}

export async function readRepoText(relPath: string): Promise<string> {
  const { absolute } = resolveRepoFile(relPath);
  return fs.readFile(absolute, "utf8");
}

export async function writeRepoText(relPath: string, content: string): Promise<ConfigFileEntry> {
  const { absolute } = resolveRepoFile(relPath);
  await fs.writeFile(absolute, content, "utf8");
  const rel = relPath.replace(/^\/+/, "").replace(/\\/g, "/");
  const root: "config" | "docs" = rel.startsWith("config/") ? "config" : "docs";
  const stat = await fs.stat(absolute);
  const entry = toEntry(root, absolute, rel);
  if (!entry) throw new Error("Unsupported file type for update");
  entry.sizeBytes = stat.size;
  entry.modifiedAt = stat.mtime.toISOString();
  return entry;
}

export async function getFilePayload(
  relPath: string,
  read: (entry: ConfigFileEntry) => Promise<string>,
): Promise<{
  entry: ConfigFileEntry;
  text: string;
}> {
  const cleaned = relPath.replace(/^\/+/, "").replace(/\\/g, "/");
  if (cleaned.startsWith("config/") || cleaned.startsWith("docs/")) {
    const { absolute } = resolveRepoFile(cleaned);
    const root: "config" | "docs" = cleaned.startsWith("config/") ? "config" : "docs";
    const entry = toEntry(root, absolute, cleaned);
    if (!entry) throw new Error("Unsupported file type");
    const stat = await fs.stat(absolute);
    entry.sizeBytes = stat.size;
    entry.modifiedAt = stat.mtime.toISOString();
    const text = await read(entry);
    return { entry, text };
  }
  const { absolute } = resolveRepoFile(cleaned);
  const stat = await fs.stat(absolute);
  const ext = path.extname(cleaned).toLowerCase();
  const type = FILE_EXTENSIONS[ext];
  if (!type) throw new Error("Unsupported file type");
  const parts = cleaned.split("/");
  const name = parts.pop() ?? cleaned;
  const entry: ConfigFileEntry = {
    path: cleaned,
    root: "docs",
    type,
    name,
    category: parts.slice(1).join("/"),
    sizeBytes: stat.size,
    modifiedAt: stat.mtime.toISOString(),
  };
  const text = await read(entry);
  return { entry, text };
}
