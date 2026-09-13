import { describe, expect, it } from "vitest";
import {
  buildBusinessLogicGraph,
  buildConfigGraph,
  buildDocsGraph,
  buildMetadataGraph,
  buildOverviewGraph,
  buildSetupGraph,
  type EntitySource,
  type FormSource,
  type RoleInfo,
} from "./graph";
import type { ConfigFileEntry, StudioGraphEdge, StudioGraphNode } from "./types";

const entityOf = (sourceKey: string): EntitySource => ({
  sourceKey,
  label: sourceKey,
  path: `config/metadata/${sourceKey}.csv`,
  count: 1,
  headers: ["id"],
});

describe("buildConfigGraph", () => {
  it("links form field sources to their entities", () => {
    const forms: FormSource[] = [
      {
        path: "config/templates/goods_receipt_form.json",
        formId: "GRN",
        title: "Goods Receipt Note",
        fields: [
          { name: "warehouseId", type: "select", source: "warehouses", required: true },
          {
            name: "lines",
            type: "table",
            columns: [{ name: "productId", type: "select", source: "products" }],
          },
        ],
      },
    ];
    const graph = buildConfigGraph(forms, [], [entityOf("warehouses"), entityOf("products")]);
    const references = graph.edges
      .filter((e) => e.source.includes("goods_receipt_form"))
      .map((e) => e.target);
    expect(references).toContain("config:config/metadata/warehouses.csv");
    expect(references).toContain("config:config/metadata/products.csv");
    const formNode = graph.nodes.find((n) => n.id.includes("goods_receipt_form"));
    expect(formNode?.meta?.fields).toBe(3);
  });

  it("pairs a form with its import template twin", () => {
    const graph = buildConfigGraph(
      [
        {
          path: "config/templates/goods_receipt_form.json",
          formId: "GRN",
          title: "GRN",
          fields: [],
        },
      ],
      [
        {
          path: "config/templates/goods_receipt_template.xlsx.csv",
          name: "goods_receipt_template.xlsx.csv",
        },
      ],
      [],
    );
    const labels = graph.edges.map((e) => e.label).filter(Boolean);
    expect(labels.some((l) => l?.includes("template") || l?.includes("form drives"))).toBe(true);
  });

  it("skips sources with no matching entity instead of dangling edges", () => {
    const graph = buildConfigGraph(
      [
        {
          path: "config/templates/x.json",
          formId: "X",
          title: "X",
          fields: [{ name: "a", type: "select", source: "missing" }],
        },
      ],
      [],
      [],
    );
    expect(graph.edges).toHaveLength(0);
  });
});

describe("buildMetadataGraph", () => {
  it("builds roles, level groups, and glossary node", () => {
    const roles: RoleInfo[] = [
      { code: "R01", name: "GLOBAL_ADMIN", level: 1, purpose: "Full admin" },
      { code: "R04", name: "WAREHOUSE_OFFICER", level: 3, purpose: "Ops" },
    ];
    const graph = buildMetadataGraph([entityOf("uom"), entityOf("programs")], roles, 12);
    const roleNode = graph.nodes.find((n) => n.id.includes("R01"));
    expect(roleNode).toBeDefined();
    expect(graph.nodes).toContainEqual(expect.objectContaining({ id: "metadata:level:1" }));
    expect(graph.edges.some((e) => e.source.includes("R01") && e.target.includes("level:1"))).toBe(
      true,
    );
    expect(
      graph.nodes.some((n) => n.id === "metadata:glossary" && n.meta?.definitions === 12),
    ).toBe(true);
  });
});

describe("buildBusinessLogicGraph", () => {
  it("includes FEFO, AMC, GRN, offline and cold-chain flows", () => {
    const graph = buildBusinessLogicGraph();
    const labels = graph.nodes.map((n) => n.label);
    for (const expected of [
      "Sort by expiry (FEFO)",
      "Reorder point",
      "Goods Receipt Note",
      "Sync queue",
      "Temperature monitoring",
    ]) {
      expect(labels.some((l) => l.includes(expected.split(" ")[0]) || l === expected)).toBe(true);
    }
    expect(graph.edges.length).toBeGreaterThan(10);
    expect(graph.nodes.every((n) => n.refs && n.refs.length >= 0)).toBe(true);
  });
});

describe("buildSetupGraph", () => {
  it("wires services to volumes and env groups", () => {
    const graph = buildSetupGraph(
      [{ name: "postgres", image: "postgres:18-alpine", volume: "pgdata", env: ["POSTGRES_USER"] }],
      ["pgdata", "meilidata"],
      [{ group: "DATABASE", target: "postgres", keys: ["DATABASE_URL"] }],
    );
    expect(graph.nodes.some((n) => n.id === "setup:service:postgres")).toBe(true);
    expect(
      graph.edges.some(
        (e) => e.source === "setup:service:postgres" && e.target === "setup:volume:pgdata",
      ),
    ).toBe(true);
    expect(
      graph.edges.some(
        (e) => e.source === "setup:env:DATABASE" && e.target === "setup:service:postgres",
      ),
    ).toBe(true);
  });

  it("links web and api to infrastructure", () => {
    const graph = buildSetupGraph(
      [{ name: "postgres", image: "postgres:18", volume: "pgdata" }],
      ["pgdata"],
      [],
    );
    expect(
      graph.edges.some((e) => e.source === "setup:api" && e.target === "setup:service:postgres"),
    ).toBe(true);
    expect(graph.edges.some((e) => e.source === "setup:web" && e.target === "setup:api")).toBe(
      true,
    );
  });
});

describe("buildDocsGraph", () => {
  it("links documents that reference each other", () => {
    const graph = buildDocsGraph([
      {
        path: "docs/ARCHITECTURE.md",
        name: "ARCHITECTURE.md",
        headingCount: 20,
        links: ["STATUS.md"],
      },
      { path: "docs/STATUS.md", name: "STATUS.md", headingCount: 30, links: [] },
    ]);
    expect(graph.edges).toHaveLength(1);
    expect(graph.edges[0].source).toContain("ARCHITECTURE");
    expect(graph.edges[0].target).toContain("STATUS");
  });
});

describe("buildOverviewGraph", () => {
  const configEntries: ConfigFileEntry[] = [
    {
      path: "config/templates/goods_receipt_form.json",
      root: "config",
      type: "json",
      name: "goods_receipt_form.json",
      category: "templates",
      sizeBytes: 1,
      modifiedAt: "",
    },
    {
      path: "config/templates/goods_receipt_template.xlsx.csv",
      root: "config",
      type: "csv",
      name: "goods_receipt_template.xlsx.csv",
      category: "templates",
      sizeBytes: 1,
      modifiedAt: "",
    },
  ];
  const docEntries: ConfigFileEntry[] = [
    {
      path: "docs/ARCHITECTURE.md",
      root: "docs",
      type: "md",
      name: "ARCHITECTURE.md",
      category: "",
      sizeBytes: 1,
      modifiedAt: "",
    },
  ];

  it("draws a form↔template edge", () => {
    const graph = buildOverviewGraph(configEntries, docEntries);
    expect(
      graph.edges.some(
        (e: StudioGraphEdge) =>
          e.source.includes("goods_receipt_form") && e.target.includes("goods_receipt_template"),
      ),
    ).toBe(true);
  });

  it("creates one node per file", () => {
    const graph = buildOverviewGraph(configEntries, docEntries);
    const ids = graph.nodes.map((n: StudioGraphNode) => n.id);
    expect(ids).toHaveLength(3);
  });
});
