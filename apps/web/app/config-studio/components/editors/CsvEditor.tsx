"use client";

import { useState } from "react";

export interface CsvPatch {
  headers: string[];
  rows: string[][];
  comments: string[];
}

interface CsvEditorProps {
  headers: string[];
  rows: string[][];
  comments: string[];
  onChange: (patch: CsvPatch) => void;
}

export function CsvEditor({ headers, rows, comments, onChange }: CsvEditorProps) {
  const [state, setState] = useState<CsvPatch>({
    headers,
    rows: rows.map((r) => [...r]),
    comments,
  });
  const [error, setError] = useState<string | null>(null);

  const commit = (next: CsvPatch) => {
    setState(next);
    try {
      const widths = next.headers.length;
      const badRow = next.rows.findIndex((r) => r.length !== widths);
      if (badRow >= 0) {
        setError(
          `Row ${badRow + 2} has ${next.rows[badRow].length} cells, expected ${widths}. Fix before saving.`,
        );
        return;
      }
      if (next.headers.some((h) => h.trim() === "")) {
        setError("Column headers cannot be empty.");
        return;
      }
      setError(null);
      onChange(next);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Invalid CSV shape");
    }
  };

  const setHeader = (index: number, value: string) => {
    const nextHeaders = [...state.headers];
    nextHeaders[index] = value;
    commit({ ...state, headers: nextHeaders });
  };

  const setCell = (rowIdx: number, colIdx: number, value: string) => {
    const nextRows = state.rows.map((r) => [...r]);
    nextRows[rowIdx][colIdx] = value;
    commit({ ...state, rows: nextRows });
  };

  const addRow = () => {
    if (state.rows.length === 0 && state.headers.length === 0) {
      commit({ ...state, headers: ["value"], rows: [[""]] });
      return;
    }
    commit({ ...state, rows: [...state.rows, state.headers.map(() => "")] });
  };

  const removeRow = (rowIdx: number) => {
    commit({ ...state, rows: state.rows.filter((_, i) => i !== rowIdx) });
  };

  const addColumn = () => {
    commit({
      ...state,
      headers: [...state.headers, `column${state.headers.length + 1}`],
      rows: state.rows.map((r) => [...r, ""]),
    });
  };

  const removeColumn = () => {
    if (state.headers.length <= 1) return;
    commit({
      ...state,
      headers: state.headers.slice(0, -1),
      rows: state.rows.map((r) => r.slice(0, -1)),
    });
  };

  return (
    <div className="space-y-3">
      {state.comments.length > 0 && (
        <div className="rounded-md bg-slate-100 px-3 py-2 text-[11px] text-slate-500">
          {state.comments.length} comment line(s) preserved on save
        </div>
      )}
      <div className="overflow-x-auto rounded-md border border-slate-200">
        <table className="w-full border-collapse text-left text-[12px]">
          <thead>
            <tr>
              <th className="w-8 border-b border-slate-200 bg-slate-100" />
              {state.headers.map((header, ci) => (
                <th key={`h${ci}`} className="border-b border-slate-200 bg-slate-100 p-0">
                  <input
                    value={header}
                    onChange={(e) => setHeader(ci, e.target.value)}
                    className="w-full bg-transparent px-2 py-1.5 font-semibold text-slate-700 outline-none focus:bg-white"
                    aria-label={`Column ${ci + 1} header`}
                  />
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {state.rows.map((row, ri) => (
              <tr key={`r${ri}`} className="border-b border-slate-100 last:border-0">
                <td className="border-r border-slate-100 bg-slate-50 px-2 py-1 text-center text-[10px] text-slate-400">
                  {ri + 2}
                </td>
                {state.headers.map((_, ci) => (
                  <td key={`c${ci}`} className="border-r border-slate-100 p-0 last:border-0">
                    <input
                      value={row[ci] ?? ""}
                      onChange={(e) => setCell(ri, ci, e.target.value)}
                      className="w-full px-2 py-1.5 outline-none focus:bg-white focus:ring-1 focus:ring-inset focus:ring-sky-300"
                      aria-label={`Row ${ri + 2} column ${ci + 1}`}
                    />
                  </td>
                ))}
                <td className="p-1 text-center">
                  <button
                    onClick={() => removeRow(ri)}
                    className="rounded px-1.5 py-0.5 text-[11px] text-rose-600 hover:bg-rose-50"
                    aria-label={`Delete row ${ri + 2}`}
                  >
                    ✕
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <button
          onClick={addRow}
          className="rounded-md bg-slate-800 px-3 py-1.5 text-xs font-medium text-white hover:bg-slate-700"
        >
          + Add row
        </button>
        <button
          onClick={addColumn}
          className="rounded-md border border-slate-300 px-3 py-1.5 text-xs font-medium text-slate-600 hover:bg-slate-50"
        >
          + Add column
        </button>
        <button
          onClick={removeColumn}
          disabled={state.headers.length <= 1}
          className="rounded-md border border-slate-300 px-3 py-1.5 text-xs font-medium text-slate-600 hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40"
        >
          − Remove last column
        </button>
        <span className="text-[11px] text-slate-400">
          {state.rows.length} row{state.rows.length === 1 ? "" : "s"} · {state.headers.length}{" "}
          columns
        </span>
      </div>
      {error && <p className="rounded-md bg-rose-50 px-3 py-2 text-xs text-rose-600">{error}</p>}
    </div>
  );
}
