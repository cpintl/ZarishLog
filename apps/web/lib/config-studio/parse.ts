export interface ParsedCsv {
  headers: string[];
  rows: string[][];
  comments: string[];
}

export interface ParsedJson {
  ok: boolean;
  value?: unknown;
  error?: string;
}

export interface MarkdownHeading {
  level: number;
  title: string;
}

export interface MarkdownTable {
  headers: string[];
  rows: string[][];
}

function isCommentLine(line: string): boolean {
  return line.trim().startsWith("#");
}

function splitCsvLine(line: string): string[] {
  const out: string[] = [];
  let current = "";
  let inQuotes = false;
  for (let i = 0; i < line.length; i++) {
    const ch = line[i];
    if (inQuotes) {
      if (ch === '"') {
        if (line[i + 1] === '"') {
          current += '"';
          i++;
        } else {
          inQuotes = false;
        }
      } else {
        current += ch;
      }
    } else if (ch === '"') {
      inQuotes = true;
    } else if (ch === ",") {
      out.push(current);
      current = "";
    } else {
      current += ch;
    }
  }
  out.push(current);
  return out.map((c) => c.trim());
}

export function parseCsv(text: string): ParsedCsv {
  const comments: string[] = [];
  const records: string[][] = [];
  for (const line of text.split(/\r?\n/)) {
    if (isCommentLine(line)) {
      comments.push(line.replace(/^\s*#\s*/, ""));
      continue;
    }
    if (line.trim() === "") continue;
    records.push(splitCsvLine(line));
  }
  return {
    headers: records[0] ?? [],
    rows: records.slice(1),
    comments,
  };
}

export function escapeCsvCell(cell: string): string {
  if (/[",\n]/.test(cell)) {
    return `"${cell.replace(/"/g, '""')}"`;
  }
  return cell;
}

export function serializeCsv(parsed: ParsedCsv): string {
  const lines: string[] = [parsed.headers.map(escapeCsvCell).join(",")];
  for (const row of parsed.rows) {
    lines.push(row.map(escapeCsvCell).join(","));
  }
  const body = lines.join("\n");
  const commentBlock =
    parsed.comments.length > 0 ? parsed.comments.map((c) => `# ${c}`).join("\n") + "\n" : "";
  return commentBlock + body + "\n";
}

export function parseJson(text: string): ParsedJson {
  try {
    return { ok: true, value: JSON.parse(text) };
  } catch (err) {
    return { ok: false, error: err instanceof Error ? err.message : String(err) };
  }
}

export function extractMarkdownHeadings(text: string): MarkdownHeading[] {
  const headings: MarkdownHeading[] = [];
  for (const line of text.split(/\r?\n/)) {
    const match = /^(#{1,6})\s+(.*)$/.exec(line.trim());
    if (match) headings.push({ level: match[1].length, title: match[2].trim() });
  }
  return headings;
}

export function findFileNameMentions(text: string, names: string[]): string[] {
  const haystack = text.toLowerCase();
  return names.filter((name) => {
    const token = name.replace(/\.md$/, "");
    return token.length > 0 && haystack.includes(token.toLowerCase());
  });
}

export function extractMarkdownLinks(text: string): string[] {
  const links: string[] = [];
  for (const line of text.split(/\r?\n/)) {
    for (const match of line.matchAll(/\[[^\]]*\]\(([^)]+)\)/g)) {
      const target = match[1].replace(/^\.\//, "").split("#")[0];
      if (target && /\.(md|json|csv)$/.test(target)) links.push(target);
    }
  }
  return [...new Set(links)];
}

export function isMarkdownSeparator(line: string): boolean {
  return /^\s*\|?\s*:?-{3,}:?\s*(\|\s*:?-{3,}:?\s*)*\|?\s*$/.test(line);
}

export function extractMarkdownTables(text: string): MarkdownTable[] {
  const tables: MarkdownTable[] = [];
  const lines = text.split(/\r?\n/);
  let i = 0;
  while (i < lines.length) {
    if (!lines[i].includes("|")) {
      i++;
      continue;
    }
    const headerCells = lines[i]
      .split("|")
      .map((c) => c.trim())
      .filter(Boolean);
    if (headerCells.length === 0 || !isMarkdownSeparator(lines[i + 1] ?? "")) {
      i++;
      continue;
    }
    const rows: string[][] = [];
    let j = i + 2;
    while (j < lines.length && lines[j].includes("|")) {
      const cells = lines[j]
        .split("|")
        .map((c) => c.trim())
        .filter(Boolean);
      if (cells.length > 0) rows.push(cells);
      j++;
    }
    tables.push({ headers: headerCells, rows });
    i = j;
  }
  return tables;
}
