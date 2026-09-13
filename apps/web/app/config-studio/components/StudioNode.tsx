"use client";

import { memo } from "react";
import { Handle, Position, type NodeProps } from "@xyflow/react";
import type { StudioGraphNode, StudioNodeKind } from "@/lib/config-studio/types";

const KIND_STYLES: Record<StudioNodeKind, { card: string; chip: string }> = {
  config: { card: "border-sky-300 bg-sky-50", chip: "bg-sky-600" },
  docs: { card: "border-indigo-300 bg-indigo-50", chip: "bg-indigo-600" },
  entity: { card: "border-emerald-300 bg-emerald-50", chip: "bg-emerald-600" },
  service: { card: "border-amber-300 bg-amber-50", chip: "bg-amber-600" },
  volume: { card: "border-violet-300 bg-violet-50", chip: "bg-violet-600" },
  env: { card: "border-fuchsia-300 bg-fuchsia-50", chip: "bg-fuchsia-600" },
  logic: { card: "border-rose-300 bg-rose-50", chip: "bg-rose-600" },
  narrative: { card: "border-slate-300 bg-slate-100", chip: "bg-slate-500" },
  file: { card: "border-cyan-300 bg-cyan-50", chip: "bg-cyan-600" },
};

const META_PREFERENCE: Array<[key: string, label: string]> = [
  ["rows", "rows"],
  ["count", "items"],
  ["columns", "columns"],
  ["fields", "fields"],
  ["headings", "headings"],
  ["definitions", "definitions"],
  ["level", "level"],
  ["image", "image"],
  ["ports", "ports"],
  ["type", "type"],
  ["path", "path"],
];

function StudioNodeInner({ data }: NodeProps) {
  const node = data.node as StudioGraphNode;
  const styles = KIND_STYLES[node.kind] ?? KIND_STYLES.narrative;

  const metaLines: Array<[string, string]> = [];
  if (node.meta) {
    for (const [key, label] of META_PREFERENCE) {
      const value = node.meta[key];
      if (value === undefined || value === "") continue;
      const text = String(value);
      metaLines.push([label, text.length > 46 ? `${text.slice(0, 46)}…` : text]);
      if (metaLines.length >= 4) break;
    }
  }

  return (
    <div className={`w-52 rounded-lg border-2 ${styles.card} px-3 py-2 shadow-sm`}>
      <Handle
        type="target"
        position={Position.Left}
        className="!h-2 !w-2 !border-0 !bg-slate-400"
      />
      <div className="flex items-center gap-2">
        <span className={`h-2.5 w-2.5 shrink-0 rounded-full ${styles.chip}`} />
        <span className="text-[13px] font-semibold leading-tight text-slate-800">{node.label}</span>
      </div>
      {metaLines.length > 0 && (
        <div className="mt-1.5 space-y-0.5 border-t border-slate-200/80 pt-1.5">
          {metaLines.map(([k, v]) => (
            <div key={k} className="flex items-baseline gap-1.5 text-[11px] text-slate-600">
              <span className="shrink-0 font-medium uppercase tracking-wide text-slate-400">
                {k}
              </span>
              <span className="truncate font-mono">{v}</span>
            </div>
          ))}
        </div>
      )}
      <Handle
        type="source"
        position={Position.Right}
        className="!h-2 !w-2 !border-0 !bg-slate-400"
      />
    </div>
  );
}

export const StudioNode = memo(StudioNodeInner);
