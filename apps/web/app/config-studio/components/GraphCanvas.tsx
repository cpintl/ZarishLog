"use client";

import { useCallback } from "react";
import {
  Background,
  BackgroundVariant,
  Controls,
  MarkerType,
  MiniMap,
  ReactFlow,
  type Edge,
  type Node,
  type NodeTypes,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import type { StudioGraphEdge, StudioGraphNode, StudioNodeKind } from "@/lib/config-studio/types";
import { StudioNode } from "./StudioNode";

const NODE_COLORS: Record<StudioNodeKind, string> = {
  config: "#0284c7",
  docs: "#4f46e5",
  entity: "#059669",
  service: "#d97706",
  volume: "#7c3aed",
  env: "#c026d3",
  logic: "#e11d48",
  narrative: "#64748b",
  file: "#0891b2",
};

const nodeTypes: NodeTypes = { studio: StudioNode };

interface GraphCanvasProps {
  nodes: Array<StudioGraphNode & { position: { x: number; y: number } }>;
  edges: StudioGraphEdge[];
  onNodeClick: (id: string) => void;
}

export function GraphCanvas({ nodes, edges, onNodeClick }: GraphCanvasProps) {
  const rfNodes: Node[] = nodes.map((n) => ({
    id: n.id,
    type: "studio" as const,
    position: n.position,
    data: { node: n },
  }));

  const rfEdges: Edge[] = edges.map((e) => ({
    id: e.id,
    source: e.source,
    target: e.target,
    label: e.label,
    labelStyle: { fontSize: 10, fill: "#475569", fontWeight: 500 },
    labelBgStyle: { fill: "#f8fafc", fillOpacity: 0.9 },
    labelBgPadding: [4, 2] as [number, number],
    labelBgBorderRadius: 4,
    markerEnd: { type: MarkerType.ArrowClosed, color: "#94a3b8" },
    style: { stroke: "#94a3b8", strokeWidth: 1.5 },
  }));

  const handleNodeClick = useCallback(
    (_: React.MouseEvent, node: Node) => onNodeClick(node.id),
    [onNodeClick],
  );

  return (
    <div className="h-full w-full">
      <ReactFlow
        nodes={rfNodes}
        edges={rfEdges}
        nodeTypes={nodeTypes}
        onNodeClick={handleNodeClick}
        fitView
        fitViewOptions={{ padding: 0.18, maxZoom: 1 }}
        minZoom={0.15}
        maxZoom={2}
        proOptions={{ hideAttribution: true }}
      >
        <Background variant={BackgroundVariant.Dots} gap={18} size={1.2} color="#cbd5e1" />
        <Controls position="bottom-left" showInteractive={false} />
        <MiniMap
          position="bottom-right"
          pannable
          zoomable
          nodeColor={(n) => {
            const kind = (n.data?.node as StudioGraphNode | undefined)?.kind;
            return kind ? NODE_COLORS[kind] : "#94a3b8";
          }}
          maskColor="rgba(241, 245, 249, 0.68)"
        />
      </ReactFlow>
    </div>
  );
}
