export type FileType = "json" | "csv" | "md" | "txt";

export interface ConfigFileEntry {
  path: string;
  root: "config" | "docs";
  type: FileType;
  name: string;
  category: string;
  sizeBytes: number;
  modifiedAt: string;
}

export type StudioDomain = "overview" | "config" | "metadata" | "business-logic" | "setup" | "docs";

export type StudioNodeKind =
  "config" | "docs" | "entity" | "service" | "volume" | "env" | "logic" | "narrative" | "file";

export interface StudioGraphNode {
  id: string;
  label: string;
  kind: StudioNodeKind;
  domain: StudioDomain;
  meta?: Record<string, string | number | boolean>;
  refs?: string[];
}

export interface StudioGraphEdge {
  id: string;
  source: string;
  target: string;
  label?: string;
}

export interface StudioGraph {
  domain: StudioDomain;
  nodes: StudioGraphNode[];
  edges: StudioGraphEdge[];
}

export interface GraphPayload {
  domain: StudioDomain;
  nodes: Array<StudioGraphNode & { position: { x: number; y: number } }>;
  edges: StudioGraphEdge[];
  title: string;
  subtitle: string;
}

export interface FilePayload {
  entry: ConfigFileEntry;
  text: string;
  parsed?: {
    json?: unknown;
    csv?: { headers: string[]; rows: string[][]; comments: string[] };
    headings?: { level: number; title: string }[];
    links?: string[];
  };
}

export interface FileUpdateRequest {
  path?: string;
  contentType: "json" | "csv" | "md";
  json?: unknown;
  csv?: { headers: string[]; rows: string[][]; comments: string[] };
  text?: string;
}

export interface StudioApiError {
  error: string;
}
