"use client";

import { useMemo, useState } from "react";
import type { FormFieldInfo } from "@/lib/config-studio/graph";

interface FieldRow {
  id: string;
  parent: string;
  name: string;
  type: string;
  reference: string;
  required: boolean;
}

const ROOT_PARENT = "__root__";
let seq = 0;
const nextId = () => `f${++seq}`;

function fieldToRow(field: FormFieldInfo, parent: string): FieldRow {
  const reference = field.source
    ? field.source
    : field.options && field.options.length > 0
      ? field.options.join(", ")
      : "";
  return {
    id: nextId(),
    parent,
    name: field.name,
    type: field.type ?? "text",
    reference,
    required: field.required ?? false,
  };
}

function collectRows(fields: FormFieldInfo[]): FieldRow[] {
  const rows: FieldRow[] = [];
  for (const field of fields) {
    rows.push(fieldToRow(field, ROOT_PARENT));
    if (field.columns && field.columns.length > 0) {
      for (const column of field.columns) {
        rows.push(fieldToRow(column, field.name));
      }
    }
  }
  return rows;
}

function rowsToFields(rows: FieldRow[], parentName: string): FormFieldInfo[] {
  return rows
    .filter((r) => r.parent === parentName)
    .map((r) => {
      const field: FormFieldInfo = {
        name: r.name.trim(),
        type: r.type.trim() || "text",
        required: r.required,
      };
      const ref = r.reference.trim();
      if (ref.includes(",")) {
        field.options = ref
          .split(",")
          .map((o) => o.trim())
          .filter(Boolean);
      } else if (ref) {
        field.source = ref;
      }
      const children = rows.filter((c) => c.parent === r.name && c.id !== r.id);
      if (children.length > 0) {
        field.columns = rowsToFields(rows, r.name);
      }
      return field;
    });
}

interface JsonFormEditorProps {
  value: unknown;
  onChange: (json: unknown) => void;
}

export function JsonFormEditor({ value, onChange }: JsonFormEditorProps) {
  const base = value as { formId?: string; title?: string; fields?: FormFieldInfo[] } | undefined;
  const [fields, setFields] = useState<FormFieldInfo[]>(base?.fields ?? []);
  const rows = useMemo(() => collectRows(fields), [fields]);
  const tableNames = useMemo(
    () => rows.filter((r) => r.type === "table").map((r) => r.name),
    [rows],
  );

  const commit = (nextRows: FieldRow[]) => {
    const nextFields = rowsToFields(nextRows, ROOT_PARENT);
    setFields(nextFields);
    onChange({ ...base, title: base?.title, formId: base?.formId, fields: nextFields });
  };

  const update = (id: string, patch: Partial<FieldRow>) => {
    commit(rows.map((r) => (r.id === id ? { ...r, ...patch } : r)));
  };

  const addTo = (parent: string) => {
    commit([
      ...rows,
      {
        id: nextId(),
        parent,
        name: `field${rows.filter((r) => r.parent === parent).length + 1}`,
        type: "text",
        reference: "",
        required: false,
      },
    ]);
  };

  const remove = (id: string) => {
    commit(rows.filter((r) => r.id !== id && r.parent !== rows.find((x) => x.id === id)?.name));
  };

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-2 gap-2 text-xs">
        <label className="block">
          <span className="mb-1 block font-medium text-slate-500">Form ID</span>
          <input
            value={base?.formId ?? ""}
            onChange={(e) =>
              onChange({ ...base, formId: e.target.value, title: base?.title, fields })
            }
            className="w-full rounded-md border border-slate-300 px-2 py-1.5 outline-none focus:border-sky-400"
          />
        </label>
        <label className="block">
          <span className="mb-1 block font-medium text-slate-500">Title</span>
          <input
            value={base?.title ?? ""}
            onChange={(e) =>
              onChange({ ...base, formId: base?.formId, title: e.target.value, fields })
            }
            className="w-full rounded-md border border-slate-300 px-2 py-1.5 outline-none focus:border-sky-400"
          />
        </label>
      </div>

      <div className="overflow-hidden rounded-md border border-slate-200">
        <table className="w-full text-left text-[12px]">
          <thead className="bg-slate-100 text-slate-500">
            <tr>
              <th className="px-2 py-1.5 font-medium">###</th>
              <th className="px-2 py-1.5 font-medium">Section</th>
              <th className="px-2 py-1.5 font-medium">Field name</th>
              <th className="px-2 py-1.5 font-medium">Type</th>
              <th className="px-2 py-1.5 font-medium">Source / options</th>
              <th className="px-2 py-1.5 font-medium">Req</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {rows.map((r, i) => {
              const isTableRoot = r.type === "table";
              return (
                <tr
                  key={r.id}
                  className={`border-t border-slate-100 ${r.parent === ROOT_PARENT ? "" : "bg-sky-50/40"} ${isTableRoot ? "bg-slate-50" : ""}`}
                >
                  <td className="px-2 py-1 text-[10px] text-slate-400">{i + 1}</td>
                  <td className="px-2 py-1 text-[11px] text-slate-500">
                    {r.parent === ROOT_PARENT
                      ? isTableRoot
                        ? "table ·"
                        : "root"
                      : `lines of ${r.parent}`}
                  </td>
                  <td className="px-1 py-1">
                    <input
                      value={r.name}
                      onChange={(e) => update(r.id, { name: e.target.value })}
                      className="w-32 rounded border border-transparent px-1.5 py-1 outline-none focus:border-sky-300 focus:bg-white"
                    />
                  </td>
                  <td className="px-1 py-1">
                    <select
                      value={r.type}
                      onChange={(e) => update(r.id, { type: e.target.value })}
                      className="rounded border border-slate-200 bg-white px-1 py-1 text-[11px] outline-none focus:border-sky-400"
                    >
                      {[
                        "text",
                        "number",
                        "date",
                        "datetime",
                        "select",
                        "table",
                        "checkbox",
                        "textarea",
                      ].map((t) => (
                        <option key={t} value={t}>
                          {t}
                        </option>
                      ))}
                    </select>
                  </td>
                  <td className="px-1 py-1">
                    <input
                      value={r.reference}
                      onChange={(e) => update(r.id, { reference: e.target.value })}
                      placeholder={r.type === "select" ? "products, 1, 2, 3 or source name" : "─"}
                      disabled={r.type === "table"}
                      className="w-40 rounded border border-slate-200 px-1.5 py-1 text-[11px] outline-none focus:border-sky-400 disabled:bg-slate-50 disabled:text-slate-400"
                    />
                  </td>
                  <td className="px-1 py-1 text-center">
                    <input
                      type="checkbox"
                      checked={r.required}
                      onChange={(e) => update(r.id, { required: e.target.checked })}
                      className="mt-1 accent-sky-600"
                      aria-label={`Required ${r.name}`}
                    />
                  </td>
                  <td className="px-1 py-1 text-center">
                    <button
                      onClick={() => remove(r.id)}
                      className="rounded px-1 text-[11px] text-rose-600 hover:bg-rose-50"
                      aria-label="Remove field"
                    >
                      ✕
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <div className="flex flex-wrap gap-2">
        <button
          onClick={() => addTo(ROOT_PARENT)}
          className="rounded-md bg-slate-800 px-3 py-1.5 text-xs font-medium text-white hover:bg-slate-700"
        >
          + Add field
        </button>
        {tableNames.map((name) => (
          <button
            key={name}
            onClick={() => addTo(name)}
            className="rounded-md border border-sky-300 bg-sky-50 px-3 py-1.5 text-xs font-medium text-sky-700 hover:bg-sky-100"
          >
            + Column line for {name}
          </button>
        ))}
      </div>
      <p className="text-[11px] leading-relaxed text-slate-400">
        Tip: in the Source / options column, a comma-separated value becomes an options list; a
        single value becomes the data source (e.g. products, warehouses, uom).
      </p>
    </div>
  );
}
