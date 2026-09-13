import type {
  ConfigFileEntry,
  StudioDomain,
  StudioGraph,
  StudioGraphEdge,
  StudioGraphNode,
} from "./types";

export interface FormFieldInfo {
  name: string;
  type: string;
  source?: string;
  options?: string[];
  columns?: FormFieldInfo[];
  required?: boolean;
}

export interface FormSource {
  path: string;
  formId?: string;
  title?: string;
  fields: FormFieldInfo[];
}

export interface EntitySource {
  sourceKey: string;
  label: string;
  path: string;
  count: number;
  headers: string[];
}

export interface RoleInfo {
  code: string;
  name: string;
  level: number;
  purpose: string;
}

export interface ServiceInfo {
  name: string;
  image?: string;
  ports?: string[];
  volume?: string;
  env?: string[];
}

export interface DocsInfo {
  path: string;
  name: string;
  headingCount: number;
  links: string[];
}

let edgeSeq = 0;
const node = (
  id: string,
  label: string,
  kind: StudioGraphNode["kind"],
  domain: StudioDomain,
  meta: StudioGraphNode["meta"] = {},
  refs: string[] = [],
): StudioGraphNode => ({ id, label, kind, domain, meta, refs });

const edge = (source: string, target: string, label?: string): StudioGraphEdge => ({
  id: `e${++edgeSeq}`,
  source,
  target,
  label,
});

const fileId = (path: string, domain: StudioDomain) => `${domain}:${path}`;

const FORM_TWIN_SUFFIX = "_form.json";
const TEMPLATE_TWIN_SUFFIX = "_template.xlsx.csv";

function formTemplateTwin(formPath: string): string | undefined {
  const base = formPath.replace(FORM_TWIN_SUFFIX, "");
  return `${base}${TEMPLATE_TWIN_SUFFIX}`;
}

const SOURCE_KEY_TO_ENTITY: Record<string, string> = {
  products: "products",
  product: "products",
  warehouses: "warehouses",
  warehouse: "warehouses",
  uom: "uom",
  programs: "programs",
  departments: "departments",
  organization: "organization",
  roles: "roles",
};

function collectFields(fields: FormFieldInfo[], acc: FormFieldInfo[] = []): FormFieldInfo[] {
  for (const field of fields) {
    acc.push(field);
    if (field.columns) collectFields(field.columns, acc);
  }
  return acc;
}

export function buildConfigGraph(
  forms: FormSource[],
  templates: Array<{ path: string; name: string }>,
  entities: EntitySource[],
  domain: StudioDomain = "config",
): StudioGraph {
  const nodes: StudioGraphNode[] = [];
  const edges: StudioGraphEdge[] = [];
  const entityById: Record<string, EntitySource> = {};
  for (const entity of entities) {
    entityById[entity.sourceKey] = entity;
    nodes.push(
      node(
        fileId(entity.path, domain),
        entity.label,
        "entity",
        domain,
        { rows: entity.count, columns: entity.headers.length, path: entity.path },
        [entity.path],
      ),
    );
  }

  for (const form of forms) {
    const formNodeId = fileId(form.path, domain);
    nodes.push(
      node(
        formNodeId,
        form.title ?? form.formId ?? form.path,
        "config",
        domain,
        { type: "form", fields: collectFields(form.fields).length, path: form.path },
        [form.path],
      ),
    );
    for (const field of collectFields(form.fields)) {
      const source = field.source;
      if (!source) continue;
      const targetKey = SOURCE_KEY_TO_ENTITY[source.toLowerCase()] ?? source.toLowerCase();
      const targetEntity = entityById[targetKey];
      if (targetEntity) {
        edges.push(
          edge(formNodeId, fileId(targetEntity.path, domain), `field: ${field.name} → ${source}`),
        );
      }
    }
    const twin = formTemplateTwin(form.path);
    if (twin && templates.some((t) => t.path === twin)) {
      edges.push(edge(formNodeId, fileId(twin, domain), "form drives import template"));
    }
  }

  for (const template of templates) {
    nodes.push(
      node(fileId(template.path, domain), template.name, "config", domain, {
        type: "template",
        path: template.path,
      }),
    );
    const twinForm = template.path.replace(TEMPLATE_TWIN_SUFFIX, FORM_TWIN_SUFFIX);
    if (twinForm !== template.path && forms.some((f) => f.path === twinForm)) {
      edges.push(
        edge(fileId(template.path, domain), fileId(twinForm, domain), "import template for form"),
      );
    }
  }

  return { domain, nodes, edges };
}

export function buildMetadataGraph(
  csvEntities: EntitySource[],
  roles: RoleInfo[],
  glossaryHeadingCount: number,
  domain: StudioDomain = "metadata",
): StudioGraph {
  const nodes: StudioGraphNode[] = [];
  const edges: StudioGraphEdge[] = [];

  for (const entity of csvEntities) {
    nodes.push(
      node(
        fileId(entity.path, domain),
        entity.label,
        "entity",
        domain,
        {
          rows: entity.count,
          columns: entity.headers.length,
          columnsList: entity.headers.join(", "),
        },
        [entity.path],
      ),
    );
  }

  const levelNodes = new Set<number>();
  for (const role of roles) {
    nodes.push(
      node(
        `metadata:role:${role.code}`,
        `${role.code} · ${role.name}`,
        "entity",
        domain,
        { level: role.level, purpose: role.purpose },
        ["config/metadata/roles.md"],
      ),
    );
    if (!levelNodes.has(role.level)) {
      levelNodes.add(role.level);
      nodes.push(
        node(`metadata:level:${role.level}`, `Org Level ${role.level}`, "narrative", domain, {}),
      );
    }
    edges.push(
      edge(
        `metadata:role:${role.code}`,
        `metadata:level:${role.level}`,
        `operates at level ${role.level}`,
      ),
    );
  }

  nodes.push(
    node(
      "metadata:glossary",
      "Reference Glossary",
      "docs",
      domain,
      { definitions: glossaryHeadingCount },
      ["config/reference_data/GLOSSARY.md"],
    ),
  );

  return { domain, nodes, edges };
}

export function buildBusinessLogicGraph(domain: StudioDomain = "business-logic"): StudioGraph {
  const nodes: StudioGraphNode[] = [];
  const edges: StudioGraphEdge[] = [];
  const add = (
    id: string,
    label: string,
    meta: StudioGraphNode["meta"] = {},
    refs: string[] = [],
  ) => nodes.push(node(id, label, "logic", domain, meta, refs));

  add("bl:fefo:in", "Batches in stock", { input: "stock_movements ledger" });
  add("bl:fefo:sort", "Sort by expiry (FEFO)", { rule: "earliest expiry first" }, [
    "packages/business-logic/fefo.go",
  ]);
  add("bl:fefo:filter", "Drop expired & zero qty", { rule: "quantity > 0 and not expired" }, [
    "packages/business-logic/fefo.go",
  ]);
  add("bl:fefo:pick", "Allocate to required qty", {}, ["packages/business-logic/fefo.go"]);
  add("bl:fefo:out", "Dispatch / issue", {});
  edges.push(edge("bl:fefo:in", "bl:fefo:sort", "candidates"));
  edges.push(edge("bl:fefo:sort", "bl:fefo:filter"));
  edges.push(edge("bl:fefo:filter", "bl:fefo:pick"));
  edges.push(edge("bl:fefo:pick", "bl:fefo:out", "batch_id → qty"));

  add("bl:amc:in", "Monthly consumption history", {}, ["packages/business-logic/amc.go"]);
  add("bl:amc:aggregate", "Aggregate per month", {}, ["packages/business-logic/amc.go"]);
  add("bl:amc:avg", "Average over window (3/6/12 mo)", { rule: "last N months" }, [
    "packages/business-logic/amc.go",
  ]);
  add("bl:amc:reorder", "Reorder point = AMC × lead time + safety stock", {}, [
    "packages/business-logic",
  ]);
  add("bl:amc:out", "Purchase request + forecast", {});
  edges.push(edge("bl:amc:in", "bl:amc:aggregate"));
  edges.push(edge("bl:amc:aggregate", "bl:amc:avg"));
  edges.push(edge("bl:amc:avg", "bl:amc:reorder"));
  edges.push(edge("bl:amc:reorder", "bl:amc:out"));

  add("bl:grn:supplier", "Supplier delivers", {});
  add("bl:grn:receive", "Goods Receipt Note", {}, ["config/templates/goods_receipt_form.json"]);
  add("bl:grn:qa", "QA inspection & disposition", {}, [
    "config/templates/inspection_checklist_template.xlsx.csv",
  ]);
  add("bl:grn:put", "Put-away to location", {});
  add("bl:grn:ledger", "Append-only stock ledger", {});
  edges.push(edge("bl:grn:supplier", "bl:grn:receive"));
  edges.push(edge("bl:grn:receive", "bl:grn:qa", "batch / expiry check"));
  edges.push(edge("bl:grn:qa", "bl:grn:put"));
  edges.push(edge("bl:grn:put", "bl:grn:ledger"));

  add("bl:offline:write", "Offline write (Dexie/IndexedDB)", {}, ["apps/web/lib/db.ts"]);
  add("bl:offline:queue", "Sync queue (5 attempts)", {}, ["apps/web/lib/sync.ts"]);
  add("bl:offline:push", "Push to API", {}, ["apps/web/lib/sync.ts"]);
  add("bl:offline:commit", "Server ledger + audit", {});
  edges.push(edge("bl:offline:write", "bl:offline:queue"));
  edges.push(edge("bl:offline:queue", "bl:offline:push", "retry up to 5"));
  edges.push(edge("bl:offline:push", "bl:offline:commit"));

  add("bl:qa:coldchain", "Temperature monitoring", {}, [
    "config/templates/temperature_log_form.json",
  ]);
  add("bl:qa:excursion", "Excursion → review", {});
  add("bl:qa:capa", "CAPA + documentation", {}, [
    "config/templates/inspection_checklist_template.xlsx.csv",
  ]);
  edges.push(edge("bl:qa:coldchain", "bl:qa:excursion"));
  edges.push(edge("bl:qa:excursion", "bl:qa:capa"));

  return { domain, nodes, edges };
}

export function buildSetupGraph(
  services: ServiceInfo[],
  volumes: string[],
  envGroups: Array<{ group: string; target: string; keys: string[] }>,
  domain: StudioDomain = "setup",
): StudioGraph {
  const nodes: StudioGraphNode[] = [];
  const edges: StudioGraphEdge[] = [];

  nodes.push(
    node("setup:web", "Web PWA (Next.js 16)", "service", domain, { layer: "client" }, ["apps/web"]),
    node("setup:api", "API (Go + Gin)", "service", domain, { layer: "server" }, ["apps/api"]),
  );

  const serviceNodeIds = new Set<string>();
  for (const service of services) {
    const id = `setup:service:${service.name}`;
    serviceNodeIds.add(id);
    nodes.push(
      node(id, service.name, "service", domain, {
        image: service.image ?? "",
        ports: service.ports?.join(", ") ?? "",
      }),
    );
    if (service.volume) {
      edges.push(edge(id, `setup:volume:${service.volume}`, "data volume"));
    }
    for (const envVar of service.env ?? []) {
      const group = envVar.split("_")[0];
      const groupNode = `setup:env:${group}`;
      if (!nodes.some((n) => n.id === groupNode)) {
        nodes.push(node(groupNode, `${group}* env`, "env", domain, {}));
      }
      if (!edges.some((e) => e.source === groupNode && e.target === id)) {
        edges.push(edge(groupNode, id, "env"));
      }
    }
  }

  for (const volume of volumes) {
    const id = `setup:volume:${volume}`;
    nodes.push(
      node(id, volume, "volume", domain, { type: "Docker named volume" }, ["docker-compose.yml"]),
    );
  }

  for (const group of envGroups) {
    const groupNode = `setup:env:${group.group}`;
    const targetId =
      group.target === "api"
        ? "setup:api"
        : group.target === "web"
          ? "setup:web"
          : serviceNodeIds.has(`setup:service:${group.target}`)
            ? `setup:service:${group.target}`
            : undefined;
    if (!targetId) continue;
    if (!nodes.some((n) => n.id === groupNode)) {
      nodes.push(
        node(groupNode, `${group.group}* env`, "env", domain, { keys: group.keys.length }, [
          ".env.example",
        ]),
      );
    }
    if (!edges.some((e) => e.source === groupNode && e.target === targetId)) {
      edges.push(edge(groupNode, targetId, "env"));
    }
  }

  edges.push(edge("setup:web", "setup:api", "REST / sync"));
  if (serviceNodeIds.has("setup:service:postgres")) {
    edges.push(edge("setup:api", "setup:service:postgres", "PostgreSQL/sqlc"));
  }
  if (serviceNodeIds.has("setup:service:redis")) {
    edges.push(edge("setup:api", "setup:service:redis", "cache / jobs"));
  }
  if (serviceNodeIds.has("setup:service:meilisearch")) {
    edges.push(edge("setup:api", "setup:service:meilisearch", "search"));
  }
  if (serviceNodeIds.has("setup:service:minio")) {
    edges.push(edge("setup:api", "setup:service:minio", "object storage"));
  }
  if (serviceNodeIds.has("setup:service:keycloak")) {
    edges.push(edge("setup:web", "setup:service:keycloak", "OIDC login"));
    edges.push(edge("setup:api", "setup:service:keycloak", "OIDC verify"));
  }

  return { domain, nodes, edges };
}

export function buildDocsGraph(docs: DocsInfo[], domain: StudioDomain = "docs"): StudioGraph {
  const nodes: StudioGraphNode[] = [];
  const edges: StudioGraphEdge[] = [];
  const byName = new Map<string, string>();
  for (const doc of docs) {
    const id = `docs:${doc.path}`;
    byName.set(doc.name, id);
    nodes.push(
      node(id, doc.name.replace(/\.md$/, ""), "docs", domain, { headings: doc.headingCount }, [
        doc.path,
      ]),
    );
  }
  for (const doc of docs) {
    const sourceId = `docs:${doc.path}`;
    for (const link of doc.links) {
      const targetName = link.split("/").pop() ?? link;
      const targetId = byName.get(targetName);
      if (targetId && targetId !== sourceId) {
        edges.push(edge(sourceId, targetId, "references"));
      }
    }
  }
  return { domain, nodes, edges };
}

export function buildOverviewGraph(
  configEntries: ConfigFileEntry[],
  docsEntries: ConfigFileEntry[],
  domain: StudioDomain = "overview",
): StudioGraph {
  const nodes: StudioGraphNode[] = [];
  const edges: StudioGraphEdge[] = [];
  const configId = (path: string) => `overview:config:${path}`;
  const docsId = (path: string) => `overview:docs:${path}`;

  for (const entry of configEntries) {
    nodes.push(
      node(configId(entry.path), entry.name, entry.root === "config" ? "config" : "file", domain, {
        category: entry.category,
        path: entry.path,
      }),
    );
  }
  for (const entry of docsEntries) {
    nodes.push(
      node(docsId(entry.path), entry.name.replace(/\.md$/, ""), "docs", domain, {
        path: entry.path,
      }),
    );
  }

  const configByTeam = new Map<string, { base: string; form?: string; template?: string }>();
  for (const entry of configEntries) {
    if (entry.type !== "json" && entry.type !== "csv") continue;
    const name = entry.name;
    let base = name;
    let kind: "form" | "template" | "other";
    if (name.endsWith(FORM_TWIN_SUFFIX)) {
      base = name.slice(0, -FORM_TWIN_SUFFIX.length);
      kind = "form";
    } else if (name.endsWith(TEMPLATE_TWIN_SUFFIX)) {
      base = name.slice(0, -TEMPLATE_TWIN_SUFFIX.length);
      kind = "template";
    } else {
      base = name.replace(/\.(json|csv)$/, "");
      kind = "other";
    }
    const key = entry.path.split("/").slice(0, -1).concat(base).join("/");
    const bucket = configByTeam.get(key) ?? { base, form: undefined, template: undefined };
    if (kind === "form") bucket.form = entry.path;
    if (kind === "template") bucket.template = entry.path;
    configByTeam.set(key, bucket);
  }

  for (const bucket of configByTeam.values()) {
    if (bucket.form && bucket.template) {
      edges.push(edge(configId(bucket.form), configId(bucket.template), "form ↔ template"));
    }
  }

  const keywordDocs = new Map<string, string[]>([
    ["ARCHITECTURE.md", ["warehouse.json", "GLOSSARY.md"]],
    ["STATUS.md", ["catalogue_summary.txt", "master_product_list.csv"]],
    ["SANDBOX.md", ["organization.csv", "programs.csv"]],
    ["PRODUCT_REQUIREMENTS_DOCUMENT.md", ["templates", "GLOSSARY.md"]],
  ]);
  for (const [docName, keywords] of keywordDocs) {
    const docEntry = docsEntries.find((d) => d.name === docName);
    if (!docEntry) continue;
    for (const keyword of keywords) {
      const match = configEntries.find(
        (c) =>
          c.name.toLowerCase().includes(keyword.replace("templates", "").toLowerCase()) ||
          c.category === keyword,
      );
      if (match) {
        edges.push(edge(docsId(docEntry.path), configId(match.path), "governs"));
      }
    }
  }

  return { domain, nodes, edges };
}
