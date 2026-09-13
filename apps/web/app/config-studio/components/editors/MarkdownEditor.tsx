"use client";

import { useState } from "react";

interface MarkdownEditorProps {
  text: string;
  onChange: (patch: string) => void;
}

export function MarkdownEditor({ text, onChange }: MarkdownEditorProps) {
  const [value, setValue] = useState(text);

  const handleChange = (next: string) => {
    setValue(next);
    onChange(next);
  };

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <span className="h-2 w-2 rounded-full bg-indigo-400" />
        <span className="text-[11px] text-slate-500">
          Markdown — {value.split(/\r?\n/).length} lines · {value.length.toLocaleString()}{" "}
          characters
        </span>
      </div>
      <textarea
        value={value}
        onChange={(e) => handleChange(e.target.value)}
        spellCheck
        className="h-80 w-full resize-y rounded-md border border-slate-300 bg-slate-50 p-3 font-mono text-[12px] leading-relaxed text-slate-800 outline-none focus:border-sky-400 focus:bg-white"
      />
    </div>
  );
}
