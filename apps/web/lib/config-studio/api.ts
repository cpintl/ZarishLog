import type {
  ConfigFileEntry,
  FilePayload,
  FileUpdateRequest,
  GraphPayload,
  StudioDomain,
} from "./types";

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, init);
  if (!res.ok) {
    let message = `Request failed (${res.status})`;
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      // keep default message
    }
    throw new Error(message);
  }
  return (await res.json()) as T;
}

export async function fetchTree(): Promise<{ config: ConfigFileEntry[]; docs: ConfigFileEntry[] }> {
  return request("/api/config-studio/tree");
}

export async function fetchGraph(domain: StudioDomain): Promise<GraphPayload> {
  return request(`/api/config-studio/graph?domain=${encodeURIComponent(domain)}`);
}

export async function fetchFile(path: string): Promise<FilePayload> {
  return request(`/api/config-studio/file?path=${encodeURIComponent(path)}`);
}

export async function saveFile(
  path: string,
  update: FileUpdateRequest,
): Promise<{ entry: ConfigFileEntry; payload: FilePayload }> {
  return request("/api/config-studio/file", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ ...update, path }),
  });
}
