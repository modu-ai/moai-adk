export interface ParsedMarkdown {
  frontmatter: Record<string, unknown>;
  body: string;
}

function parseInlineArray(value: string): string[] | null {
  const trimmed = value.trim();
  if (!trimmed.startsWith("[") || !trimmed.endsWith("]")) return null;
  return trimmed
    .slice(1, -1)
    .split(",")
    .map((part) => part.trim().replace(/^['\"]|['\"]$/g, ""))
    .filter(Boolean);
}

function parseScalar(value: string): unknown {
  const trimmed = value.trim();
  const inlineArray = parseInlineArray(trimmed);
  if (inlineArray) return inlineArray;
  if (trimmed === "true") return true;
  if (trimmed === "false") return false;
  if (/^-?\d+(\.\d+)?$/.test(trimmed)) return Number(trimmed);
  return trimmed.replace(/^['\"]|['\"]$/g, "");
}

export function parseMarkdownFrontmatter(text: string): ParsedMarkdown {
  if (!text.startsWith("---\n")) return { frontmatter: {}, body: text };
  const end = text.indexOf("\n---", 4);
  if (end === -1) return { frontmatter: {}, body: text };

  const yaml = text.slice(4, end).replace(/\r\n/g, "\n");
  const bodyStart = text.indexOf("\n", end + 4);
  const body = bodyStart === -1 ? "" : text.slice(bodyStart + 1);
  const result: Record<string, unknown> = {};
  const lines = yaml.split("\n");

  for (let i = 0; i < lines.length; i++) {
    const raw = lines[i];
    if (!raw.trim() || raw.trimStart().startsWith("#")) continue;
    const match = raw.match(/^([A-Za-z0-9_-]+):\s*(.*)$/);
    if (!match) continue;
    const key = match[1];
    const value = match[2] ?? "";

    if (value === "|" || value === ">") {
      const chunks: string[] = [];
      while (i + 1 < lines.length && /^\s+/.test(lines[i + 1])) {
        i++;
        chunks.push(lines[i].replace(/^\s{2,}/, ""));
      }
      result[key] = value === ">" ? chunks.join(" ").trim() : chunks.join("\n").trim();
      continue;
    }

    if (value.trim() === "") {
      const items: string[] = [];
      while (i + 1 < lines.length && /^\s*-\s+/.test(lines[i + 1])) {
        i++;
        items.push(lines[i].replace(/^\s*-\s+/, "").trim().replace(/^['\"]|['\"]$/g, ""));
      }
      result[key] = items.length ? items : "";
      continue;
    }

    result[key] = parseScalar(value);
  }

  return { frontmatter: result, body };
}

export function asString(value: unknown, fallback = ""): string {
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "boolean") return String(value);
  if (Array.isArray(value)) return value.join(", ");
  return fallback;
}

export function asStringArray(value: unknown): string[] {
  if (Array.isArray(value)) return value.map((v) => String(v)).filter(Boolean);
  if (typeof value === "string") {
    return value
      .split(/[\n,]/)
      .map((part) => part.trim())
      .filter(Boolean);
  }
  return [];
}
