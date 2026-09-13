import { describe, expect, it } from "vitest";
import { LAYER_SPACING_X, layoutGraph } from "./layout";
import type { StudioGraphEdge, StudioGraphNode } from "./types";

const nodeOf = (id: string): StudioGraphNode => ({
  id,
  label: id,
  kind: "logic",
  domain: "business-logic",
});
const edgeOf = (source: string, target: string): StudioGraphEdge => ({
  id: `${source}-${target}`,
  source,
  target,
});

describe("layoutGraph", () => {
  it("lays a simple chain left to right", () => {
    const nodes = [nodeOf("a"), nodeOf("b"), nodeOf("c")];
    const edges = [edgeOf("a", "b"), edgeOf("b", "c")];
    const positioned = layoutGraph(nodes, edges);
    const a = positioned.find((n) => n.id === "a")!;
    const b = positioned.find((n) => n.id === "b")!;
    const c = positioned.find((n) => n.id === "c")!;
    expect(c.position.x).toBeGreaterThan(b.position.x);
    expect(b.position.x).toBeGreaterThan(a.position.x);
    expect(c.position.x).toBeCloseTo(2 * LAYER_SPACING_X);
    expect(a.position.y).toBe(0);
  });

  it("handles a node referenced by two sources by using the longest path", () => {
    const nodes = [nodeOf("start"), nodeOf("mid"), nodeOf("final")];
    const edges = [edgeOf("start", "final"), edgeOf("mid", "final")];
    const positioned = layoutGraph(nodes, edges);
    const final = positioned.find((n) => n.id === "final")!;
    expect(final.position.x).toBeCloseTo(LAYER_SPACING_X);
  });

  it("keeps sibling nodes on different rows within a layer", () => {
    const nodes = [nodeOf("a"), nodeOf("b"), nodeOf("c")];
    const edges = [edgeOf("a", "c"), edgeOf("b", "c")];
    const positioned = layoutGraph(nodes, edges);
    const byColumn = new Map<number, number[]>();
    for (const n of positioned) {
      const xs = byColumn.get(n.position.x) ?? [];
      xs.push(n.position.y);
      byColumn.set(n.position.x, xs);
    }
    for (const ys of byColumn.values()) {
      expect(new Set(ys).size).toBe(ys.length);
    }
  });
});
