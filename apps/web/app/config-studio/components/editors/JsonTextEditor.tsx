"use client";

import { useState } from "react";

interface JsonTextEditorProps {
  value: unknown;
  onChange: (json: unknown) => void;
}

export function JsonTextEditor({ value, onChange }: JsonTextEditorProps) {
  const [text, setText] = useState(() => JSON.stringify(value, null, 2));
  const [error, setError] = useState<string | null>(null);

  const handleChange = (next: string) => {
    setText(next);
    try {
      onChange(JSON.parse(next));
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Invalid JSON");
    }
  };

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <span className={`h-2 w-2 rounded-full ${error ? "bg-rose-500" : "bg-emerald-500"}`} />
        <span className="text-[11px] text-slate-500">
          {error ? "Invalid JSON — fix before saving" : "Valid JSON — changes are ready to save"}
        </span>
      </div>
      <textarea
        value={text}
        onChange={(e) => handleChange(e.target.value)}
        spellCheck={false}
        className="h-64 w-full resize-y rounded-md border border-slate-300 bg-slate-50 p-3 font-mono text-[12px] leading-relaxed text-slate-800 outline-none focus:border-sky-400 focus:bg-white"
      />
      {error && <p className="rounded-md bg-rose-50 px-3 py-2 text-xs text-rose-600">{error}</p>}
    </div>
  );
}
