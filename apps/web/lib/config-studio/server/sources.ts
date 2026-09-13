import type {
  ConfigFileEntry,
  FilePayload,
  GraphPayload,
  StudioDomain,
  StudioGraph,
} from "../types";
import type { FormSource, RoleInfo, ServiceInfo } from "../graph";
import {
  buildBusinessLogicGraph,
  buildConfigGraph,
  buildDocsGraph,
  buildMetadataGraph,
  buildOverviewGraph,
  buildSetupGraph,
} from "../graph";
import { layoutGraph, type PositionedNode } from "../layout";
import {
  extractMarkdownHeadings,
  extractMarkdownLinks,
  extractMarkdownTables,
  findFileNameMentions,
  parseCsv,
  parseJson,
} from "../parse";
import { getFilePayload, listConfigFiles, listDocsFiles, readFileEntry, readRepoText } from "./fs";

const DOMAIN_TITLES: Record<StudioDomain, { title: string; subtitle: string }> = {
  overview: {
    title: "System Overview",
    subtitle: "Every configuration and document in the platform, and how they relate",
  },
  config: {
    title: "Configuration & Forms",
    subtitle: "Data-entry forms, import templates, and the master data they reference",
  },
  metadata: {
    title: "Master Data & Roles",
    subtitle: "Reference catalogues, role hierarchy, and the standardized glossary",
  },
  "business-logic": {
    title: "Business Logic Flows",
    subtitle: "FEFO picking, AMC & reorder, goods receipt, offline sync, and cold-chain QA",
  },
  setup: {
    title: "Dependent Setup",
    subtitle: "Infrastructure services, volumes, environment groups, and app dependencies",
  },
  docs: {
    title: "Documentation Map",
    subtitle: "All project documents and the links between them",
  },
};

export function domainMeta(domain: StudioDomain) {
  return DOMAIN_TITLES[domain];
}

export function toGraphPayload(graph: StudioGraph): GraphPayload {
  const positioned = layoutGraph(graph.nodes, graph.edges);
  return {
    domain: graph.domain,
    nodes: positioned,
    edges: graph.edges,
    ...DOMAIN_TITLES[graph.domain],
  };
}

async function readConfigEntry(entry: ConfigFileEntry): Promise<string> {
  return readFileEntry(entry);
}

export async function resolveFilePayload(relPath: string): Promise<FilePayload> {
  const { entry, text } = await getFilePayload(relPath, readConfigEntry);
  const payload: FilePayload = { entry, text };
  if (entry.type === "json") {
    payload.parsed = { json: parseJson(text).value };
  } else if (entry.type === "csv") {
    payload.parsed = { csv: parseCsv(text) };
  } else if (entry.type === "md") {
    payload.parsed = {
      headings: extractMarkdownHeadings(text),
      links: extractMarkdownLinks(text),
    };
  }
  return payload;
}

async function listEntries(): Promise<{ config: ConfigFileEntry[]; docs: ConfigFileEntry[] }> {
  const [config, docs] = await Promise.all([listConfigFiles(), listDocsFiles()]);
  return { config, docs };
}

function parseRoles(text: string): RoleInfo[] {
  const roles: RoleInfo[] = [];
  for (const table of extractMarkdownTables(text)) {
    const codeIdx = table.headers.findIndex((h) => h.toLowerCase().includes("code"));
    const nameIdx = table.headers.findIndex((h) => h.toLowerCase().includes("name"));
    const levelIdx = table.headers.findIndex((h) => h.toLowerCase().includes("level"));
    const purposeIdx = table.headers.findIndex((h) => h.toLowerCase().includes("purpose"));
    if (codeIdx < 0 || levelIdx < 0) continue;
    for (const row of table.rows) {
      const level = Number.parseInt(row[levelIdx] ?? "", 10);
      if (!Number.isNaN(level)) {
        roles.push({
          code: row[codeIdx] ?? "",
          name: row[nameIdx] ?? "",
          level,
          purpose: row[purposeIdx] ?? "",
        });
      }
    }
  }
  return roles;
}

async function loadConfigEntities(): Promise<{
  entities: Array<{
    sourceKey: string;
    label: string;
    path: string;
    count: number;
    headers: string[];
  }>;
  forms: FormSource[];
  templates: Array<{ path: string; name: string }>;
}> {
  const { config } = await listEntries();
  const entities: Array<{
    sourceKey: string;
    label: string;
    path: string;
    count: number;
    headers: string[];
  }> = [];
  const forms: FormSource[] = [];
  const templates: Array<{ path: string; name: string }> = [];

  const csvMeta: Array<{ sourceKey: string; label: string; path: string; headers?: string[] }> = [
    { sourceKey: "uom", label: "UoM · Units of Measure", path: "config/metadata/uom.csv" },
    { sourceKey: "programs", label: "Programs", path: "config/metadata/programs.csv" },
    { sourceKey: "departments", label: "Departments", path: "config/metadata/departments.csv" },
    { sourceKey: "organization", label: "Organizations", path: "config/metadata/organization.csv" },
    {
      sourceKey: "products",
      label: "Products (Master Catalogue)",
      path: "config/metadata/master_product_list.csv",
    },
    {
      sourceKey: "catalogue",
      label: "Expanded Product Catalogue",
      path: "config/metadata/master_product_catalogue.csv",
    },
  ];

  for (const meta of csvMeta) {
    try {
      const text = await readRepoText(meta.path);
      const parsed = parseCsv(text);
      entities.push({
        sourceKey: meta.sourceKey,
        label: meta.label,
        path: meta.path,
        count: parsed.rows.length,
        headers: meta.headers ?? parsed.headers,
      });
    } catch {
      continue;
    }
  }

  entities.push({
    sourceKey: "warehouses",
    label: "Warehouses & Office Directory",
    path: "config/location/warehouse.json",
    count: 0,
    headers: ["id", "name", "office_category", "level"],
  });

  const configEntries = config.filter((e) => e.path.startsWith("config/templates/"));
  for (const entry of configEntries) {
    if (entry.type === "json") {
      const text = await readConfigEntry(entry);
      const parsed = parseJson(text);
      if (!parsed.ok || !parsed.value || typeof parsed.value !== "object") continue;
      const value = parsed.value as { formId?: string; title?: string; fields?: unknown[] };
      forms.push({
        path: entry.path,
        formId: typeof value.formId === "string" ? value.formId : undefined,
        title: typeof value.title === "string" ? value.title : undefined,
        fields: Array.isArray(value.fields) ? (value.fields as FormSource["fields"]) : [],
      });
    } else if (entry.type === "csv") {
      templates.push({ path: entry.path, name: entry.name });
    }
  }

  return { entities, forms, templates };
}

async function loadMetadata(): Promise<{
  csvEntities: Array<{
    sourceKey: string;
    label: string;
    path: string;
    count: number;
    headers: string[];
  }>;
  roles: RoleInfo[];
  glossaryHeadingCount: number;
}> {
  const deps = await loadConfigEntities();
  const csvEntities = deps.entities.filter((e) => e.path.endsWith(".csv"));
  csvEntities.push({
    sourceKey: "warehouses",
    label: "Warehouses & Office Directory",
    path: "config/location/warehouse.json",
    count: 0,
    headers: ["id", "name", "office_category", "level"],
  });

  let roles: RoleInfo[] = [];
  let glossaryHeadingCount = 0;
  try {
    const rolesText = await readRepoText("config/metadata/roles.md");
    roles = parseRoles(rolesText);
  } catch {
    roles = [];
  }
  try {
    const glossaryText = await readRepoText("config/reference_data/GLOSSARY.md");
    glossaryHeadingCount = extractMarkdownHeadings(glossaryText).filter((h) => h.level >= 2).length;
  } catch {
    glossaryHeadingCount = 0;
  }
  return { csvEntities, roles, glossaryHeadingCount };
}

function parseDockerCompose(text: string): { services: ServiceInfo[]; volumes: string[] } {
  const services: ServiceInfo[] = [];
  const volumes: string[] = [];
  let section: "" | "services" | "volumes" = "";
  let current: ServiceInfo | undefined;
  const namedVolumes = new Set<string>();
  const pushVolume = (binding: string | undefined) => {
    if (!binding || !current) return;
    if (binding.includes(".") || binding.includes("/") || binding.includes(":")) return;
    if (binding.endsWith("-data")) binding = binding.slice(0, -"-data".length);
    if (namedVolumes.has(binding)) current.volume = binding;
  };
  for (const raw of text.split(/\r?\n/)) {
    const trimmed = raw.trim();
    if (!trimmed) continue;
    if (raw === trimmed) {
      const top = /^([a-z0-9_-]+):\s*$/.exec(trimmed);
      if (top) section = top[1] === "services" || top[1] === "volumes" ? top[1] : "";
      current = undefined;
      continue;
    }
    const indent = raw.match(/^(\s+)/)?.[1].length ?? 0;
    if (indent === 2) {
      const svc = /^([a-z0-9_-]+):\s*$/.exec(trimmed);
      if (section === "services" && svc) {
        current = { name: svc[1] };
        services.push(current);
        continue;
      }
      if (section === "volumes" && svc) {
        volumes.push(svc[1]);
        namedVolumes.add(svc[1]);
        continue;
      }
    }
    if (section === "services" && current) {
      const image = /^image:\s*(.+)$/.exec(trimmed);
      if (image) current.image = image[1];
      const port = /^"(\d+):(\d+)"$/.exec(trimmed) ?? /^(\d+):(\d+)$/.exec(trimmed);
      if (port) {
        current.ports = current.ports ?? [];
        current.ports.push(`${port[1]}:${port[2]}`);
      }
      const envKey = /^([A-Z][A-Z0-9_]*):/.exec(trimmed);
      if (envKey) {
        current.env = current.env ?? [];
        current.env.push(envKey[1]);
      }
      const volumeBinding = /^-\s+(.+):/.exec(trimmed);
      if (volumeBinding) pushVolume(volumeBinding[1]);
    }
  }
  return { services, volumes };
}

function envTarget(key: string): string {
  if (key.startsWith("DATABASE_")) return "postgres";
  if (key.startsWith("REDIS_")) return "redis";
  if (key.startsWith("MEILISEARCH_")) return "meilisearch";
  if (key.startsWith("SMTP_") || key.startsWith("NOTIFICATION_")) return "notification";
  if (key.startsWith("OIDC_")) return "keycloak";
  if (key.startsWith("NEXT_PUBLIC_")) return "web";
  if (key.startsWith("DEFAULT_")) return "api";
  return "api";
}

function groupEnvKeys(envText: string): Array<{ group: string; target: string; keys: string[] }> {
  const byTarget = new Map<string, { group: string; target: string; keys: string[] }>();
  for (const line of envText.split(/\r?\n/)) {
    const m = /^([A-Z][A-Z0-9_]*)=/.exec(line);
    if (!m) continue;
    const target = envTarget(m[1]);
    const existing = byTarget.get(target) ?? { group: target, target, keys: [] };
    existing.keys.push(m[1]);
    byTarget.set(target, existing);
  }
  return [...byTarget.values()];
}

export async function buildGraph(domain: StudioDomain): Promise<GraphPayload> {
  switch (domain) {
    case "overview": {
      const { config, docs } = await listEntries();
      return toGraphPayload(buildOverviewGraph(config, docs));
    }
    case "config": {
      const { entities, forms, templates } = await loadConfigEntities();
      return toGraphPayload(buildConfigGraph(forms, templates, entities));
    }
    case "metadata": {
      const { csvEntities, roles, glossaryHeadingCount } = await loadMetadata();
      return toGraphPayload(buildMetadataGraph(csvEntities, roles, glossaryHeadingCount));
    }
    case "business-logic":
      return toGraphPayload(buildBusinessLogicGraph());
    case "setup": {
      let services: ServiceInfo[] = [];
      let volumes: string[] = [];
      try {
        const compose = await readRepoText("docker-compose.yml");
        const parsed = parseDockerCompose(compose);
        services = parsed.services;
        volumes = parsed.volumes;
      } catch {
        services = [];
        volumes = [];
      }
      let envGroups: ReturnType<typeof groupEnvKeys> = [];
      try {
        const envText = await readRepoText(".env.example");
        envGroups = groupEnvKeys(envText);
      } catch {
        envGroups = [];
      }
      return toGraphPayload(buildSetupGraph(services, volumes, envGroups));
    }
    case "docs": {
      const { docs } = await listEntries();
      const names = docs.map((d) => d.name);
      const infos: Array<{
        path: string;
        name: string;
        headingCount: number;
        links: string[];
      }> = [];
      for (const entry of docs) {
        const text = await readFileEntry(entry);
        const mentionLinks = findFileNameMentions(text, names);
        infos.push({
          path: entry.path,
          name: entry.name,
          headingCount: extractMarkdownHeadings(text).length,
          links: [...new Set([...extractMarkdownLinks(text), ...mentionLinks])],
        });
      }
      return toGraphPayload(buildDocsGraph(infos));
    }
  }
}

export { parseRoles }; // for reuse in tests
