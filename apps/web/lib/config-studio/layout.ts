import type { StudioGraphEdge, StudioGraphNode } from "./types";

export interface PositionedNode extends StudioGraphNode {
  position: { x: number; y: number };
}

export const NODE_WIDTH = 210;
export const LAYER_SPACING_X = 290;
export const LAYER_SPACING_Y = 120;

export function layoutGraph(nodes: StudioGraphNode[], edges: StudioGraphEdge[]): PositionedNode[] {
  const layer = new Map<string, number>();
  const visiting = new Set<string>();

  const longestPathLayer = (id: string): number => {
    if (layer.has(id)) return layer.get(id)!;
    if (visiting.has(id)) return 0;
    visiting.add(id);
    let depth = 0;
    for (const edge of edges) {
      if (edge.target === id) {
        depth = Math.max(depth, longestPathLayer(edge.source) + 1);
      }
    }
    visiting.delete(id);
    layer.set(id, depth);
    return depth;
  };

  for (const node of nodes) longestPathLayer(node.id);

  const index = new Map<string, number>();
  nodes.forEach((n, i) => index.set(n.id, i));

  const byLayer = new Map<number, StudioGraphNode[]>();
  for (const node of nodes) {
    const l = layer.get(node.id) ?? 0;
    if (!byLayer.has(l)) byLayer.set(l, []);
    byLayer.get(l)!.push(node);
  }

  const positioned: PositionedNode[] = [];
  for (const [l, layerNodes] of byLayer) {
    layerNodes.forEach((node, i) => {
      positioned.push({ ...node, position: { x: l * LAYER_SPACING_X, y: i * LAYER_SPACING_Y } });
    });
  }

  return positioned.sort((a, b) => {
    const aLayer = layer.get(a.id) ?? 0;
    const bLayer = layer.get(b.id) ?? 0;
    return aLayer - bLayer || (index.get(a.id) ?? 0) - (index.get(b.id) ?? 0);
  });
}
