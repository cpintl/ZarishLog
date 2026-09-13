"use client";

import type { FilePayload, FileUpdateRequest, StudioGraphNode } from "@/lib/config-studio/types";
import { CsvEditor } from "./editors/CsvEditor";
import { JsonFormEditor } from "./editors/JsonFormEditor";
import { JsonTextEditor } from "./editors/JsonTextEditor";
import { MarkdownEditor } from "./editors/MarkdownEditor";

interface DetailsDrawerProps {
  node: (StudioGraphNode & { position: { x: number; y: number } }) | null;
  file: FilePayload | null;
  fileLoading: boolean;
  fileError: string | null;
  edit: FileUpdateRequest | null;
  setEdit: (edit: FileUpdateRequest | null) => void;
  save: () => void;
  saving: boolean;
  savedAt: string | null;
  onClose: () => void;
}

type JsonShape = { formId?: unknown; title?: unknown; fields?: unknown; [key: string]: unknown };

function isFormJson(value: unknown): value is JsonShape {
  const v = value as JsonShape | undefined;
  return !!v && typeof v === "object" && Array.isArray(v.fields) && typeof v.formId === "string";
}

function EditorPanel({
  file,
  edit,
  setEdit,
}: Pick<DetailsDrawerProps, "file" | "edit" | "setEdit">) {
  if (!file) return null;
  const type = file.entry.type;

  if (type === "csv" && file.parsed?.csv) {
    return (
      <CsvEditor
        key={`csv:${file.entry.path}:${file.entry.modifiedAt}`}
        headers={file.parsed.csv.headers}
        rows={file.parsed.csv.rows}
        comments={file.parsed.csv.comments}
        onChange={(csv) => setEdit({ contentType: "csv", csv })}
      />
    );
  }

  if (type === "json" && file.parsed) {
    if (isFormJson(file.parsed.json)) {
      return (
        <JsonFormEditor
          key={`form:${file.entry.path}:${file.entry.modifiedAt}`}
          value={file.parsed.json}
          onChange={(json) => setEdit({ contentType: "json", json })}
        />
      );
    }
    return (
      <JsonTextEditor
        key={`json:${file.entry.path}:${file.entry.modifiedAt}`}
        value={file.parsed.json}
        onChange={(json) => setEdit({ contentType: "json", json })}
      />
    );
  }

  if (type === "md" || type === "txt") {
    return (
      <MarkdownEditor
        key={`md:${file.entry.path}:${file.entry.modifiedAt}`}
        text={file.text}
        onChange={(text) => setEdit({ contentType: "md", text })}
      />
    );
  }

  return <p className="text-xs text-slate-500">This file type is read-only in the studio.</p>;
}

export function DetailsDrawer({
  node,
  file,
  fileLoading,
  fileError,
  edit,
  setEdit,
  save,
  saving,
  savedAt,
  onClose,
}: DetailsDrawerProps) {
  if (!node) return null;

  const editable =
    !!file &&
    (file.entry.type === "json" ||
      file.entry.type === "csv" ||
      file.entry.type === "md" ||
      file.entry.type === "txt");

  return (
    <aside className="flex h-full w-full flex-col border-l border-slate-200 bg-white">
      <div className="flex items-start justify-between gap-2 border-b border-slate-200 px-4 py-3">
        <div className="min-w-0">
          <p className="truncate text-sm font-semibold text-slate-800">{node.label}</p>
          <p className="font-mono text-[11px] text-slate-400">{node.id}</p>
        </div>
        <button
          onClick={onClose}
          className="rounded-md p-1 text-slate-400 hover:bg-slate-100 hover:text-slate-600"
          aria-label="Close panel"
        >
          ✕
        </button>
      </div>

      <div className="flex-1 overflow-y-auto px-4 py-3">
        {node.meta && Object.keys(node.meta).length > 0 && (
          <dl className="mb-3 space-y-1">
            {Object.entries(node.meta).map(([key, value]) => {
              const text = String(value);
              return (
                <div key={key} className="flex gap-2 text-[11px]">
                  <dt className="w-24 shrink-0 font-medium uppercase tracking-wide text-slate-400">
                    {key}
                  </dt>
                  <dd className="break-words font-mono text-slate-700">{text}</dd>
                </div>
              );
            })}
          </dl>
        )}

        {node.refs && node.refs.length > 0 && (
          <div className="mb-3">
            <p className="mb-1 text-[10px] font-semibold uppercase tracking-wide text-slate-400">
              Source files
            </p>
            <div className="flex flex-wrap gap-1">
              {node.refs.map((ref) => (
                <span
                  key={ref}
                  className="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-[10px] text-slate-600"
                >
                  {ref}
                </span>
              ))}
            </div>
          </div>
        )}

        {fileLoading ? (
          <p className="py-8 text-center text-xs text-slate-400">Loading file…</p>
        ) : fileError ? (
          <p className="rounded-md bg-rose-50 px-3 py-2 text-xs text-rose-600">{fileError}</p>
        ) : file ? (
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <p className="font-mono text-[11px] text-slate-500">{file.entry.path}</p>
              {editable && edit && (
                <span className="rounded bg-amber-100 px-1.5 py-0.5 text-[10px] font-medium text-amber-700">
                  unsaved
                </span>
              )}
            </div>
            <EditorPanel file={file} edit={edit} setEdit={setEdit} />
            {editable && !edit && (
              <p className="rounded-md bg-emerald-50 px-3 py-2 text-[11px] text-emerald-700">
                In sync with the file. Edit something above to enable Save.
              </p>
            )}
          </div>
        ) : (
          <p className="py-6 text-center text-[11px] leading-relaxed text-slate-400">
            This node has no local file to edit.
            <br />
            Reference logic and services are view-only.
          </p>
        )}
      </div>

      <div className="border-t border-slate-200 px-4 py-3">
        {editable && (
          <button
            onClick={save}
            disabled={!edit || saving}
            className={`w-full rounded-md px-3 py-2 text-sm font-medium text-white transition-colors ${
              edit && !saving ? "bg-sky-600 hover:bg-sky-500" : "cursor-not-allowed bg-slate-300"
            }`}
          >
            {saving ? "Saving…" : edit ? "Save changes to file" : "No changes to save"}
          </button>
        )}
        <div className="mt-2 flex items-center justify-between text-[10px] text-slate-400">
          <span>{savedAt ? `Last synced ${savedAt}` : "Auto-refreshes every 8s while idle"}</span>
          <span>
            {file ? file.entry.sizeBytes.toLocaleString() : ""}
            {file ? " B" : ""}
          </span>
        </div>
      </div>
    </aside>
  );
}
