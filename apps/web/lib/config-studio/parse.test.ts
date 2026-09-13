import { describe, expect, it } from "vitest";
import {
  extractMarkdownHeadings,
  extractMarkdownLinks,
  extractMarkdownTables,
  findFileNameMentions,
  isMarkdownSeparator,
  parseCsv,
  parseJson,
  serializeCsv,
} from "./parse";

describe("parseCsv", () => {
  it("parses headers and data rows", () => {
    const parsed = parseCsv("name,abbreviation,category\nEach,EA,count\nBox,BX,count\n");
    expect(parsed.headers).toEqual(["name", "abbreviation", "category"]);
    expect(parsed.rows).toEqual([
      ["Each", "EA", "count"],
      ["Box", "BX", "count"],
    ]);
  });

  it("handles quoted cells containing commas and quotes", () => {
    const parsed = parseCsv('sku,name\nA001,"Glass, 10ml ""premium"""\n');
    expect(parsed.headers).toEqual(["sku", "name"]);
    expect(parsed.rows[0][1]).toBe('Glass, 10ml "premium"');
  });

  it("preserves full-line # comments separately", () => {
    const parsed = parseCsv("# product master list\nsku,name\nA001,Aspirin\n");
    expect(parsed.comments).toEqual(["product master list"]);
    expect(parsed.headers).toEqual(["sku", "name"]);
    expect(parsed.rows).toEqual([["A001", "Aspirin"]]);
  });

  it("skips blank lines", () => {
    const parsed = parseCsv("a,b\n\n1,2\n");
    expect(parsed.rows).toEqual([["1", "2"]]);
  });
});

describe("serializeCsv", () => {
  it("round-trips through parseCsv", () => {
    const source = "# header comment\nsku,name\nA001," + '"Glass, 10ml"' + "\n";
    const parsed = parseCsv(source);
    expect(parseCsv(serializeCsv(parsed))).toEqual(parsed);
  });

  it("escapes commas, quotes, and newlines in cells", () => {
    const out = serializeCsv({ headers: ["a", "b"], rows: [["x,y", 'say "hi"']], comments: [] });
    expect(out).toContain('"x,y"');
    expect(out).toContain('"say ""hi"""');
  });
});

describe("parseJson", () => {
  it("parses valid JSON", () => {
    const result = parseJson('{"formId":"GRN","fields":[]}');
    expect(result.ok).toBe(true);
    expect(result.value).toEqual({ formId: "GRN", fields: [] });
  });

  it("reports errors for invalid JSON", () => {
    const result = parseJson("{oops");
    expect(result.ok).toBe(false);
    expect(result.error).toBeTruthy();
  });
});

describe("markdown helpers", () => {
  it("extracts headings", () => {
    const headings = extractMarkdownHeadings("# Title\n## Section\n### Sub\nNot a heading\n");
    expect(headings).toEqual([
      { level: 1, title: "Title" },
      { level: 2, title: "Section" },
      { level: 3, title: "Sub" },
    ]);
  });

  it("extracts internal markdown links only", () => {
    const links = extractMarkdownLinks(
      "See [ARCHITECTURE.md](ARCHITECTURE.md) and [web](https://example.com).\n",
    );
    expect(links).toEqual(["ARCHITECTURE.md"]);
  });

  it("finds prose mentions of sibling document names", () => {
    const mentions = findFileNameMentions("See the SANDBOX guide and STATUS report for progress.", [
      "ARCHITECTURE.md",
      "SANDBOX.md",
      "STATUS.md",
    ]);
    expect(mentions).toEqual(["SANDBOX.md", "STATUS.md"]);
  });

  it("detects table separators and extracts tables", () => {
    expect(isMarkdownSeparator("| --- | --- |")).toBe(true);
    const tables = extractMarkdownTables(
      "| Code | Name |\n| --- | --- |\n| R01 | Admin |\n| R02 | Rep |\n\nText.\n",
    );
    expect(tables).toEqual([
      {
        headers: ["Code", "Name"],
        rows: [
          ["R01", "Admin"],
          ["R02", "Rep"],
        ],
      },
    ]);
  });
});
