import { readdir, readFile, stat } from "node:fs/promises";
import { join } from "node:path";

export interface PageMapItem {
  name: string;
  route: string;
  title?: string;
  frontMatter?: Record<string, unknown>;
  children?: PageMapItem[];
  data?: Record<string, unknown>;
}

// Convert kebab-case or snake_case to Title Case (fallback)
function toTitleCase(str: string): string {
  return str
    .replace(/[-_]/g, " ")
    .replace(/\b\w/g, (char) => char.toUpperCase());
}

// Parse _meta.ts file to extract metadata
async function parseMetaFile(
  metaPath: string,
): Promise<Record<string, string | { title?: string; display?: string }>> {
  try {
    const content = await readFile(metaPath, "utf-8");

    // Try different formats

    // Format 1: export default { ... }
    let match = content.match(/export\s+default\s+({[\s\S]*?});?\s*$/);
    if (match) {
      return parseMetaObject(match[1]);
    }

    // Format 2: const meta: MetaRecord = { ... }; export default meta;
    match = content.match(
      /const\s+meta(?::\s*MetaRecord)?\s*=\s*({[\s\S]*?});\s*export\s+default\s+meta;?/,
    );
    if (match) {
      return parseMetaObject(match[1]);
    }

    // Format 3: One-line format - const meta = { ... }; export default meta;
    match = content.match(
      /const\s+meta\s*=\s*({[^}]+})\s*;\s*export\s+default\s+meta/,
    );
    if (match) {
      return parseMetaObject(match[1]);
    }

    return {};
  } catch {
    return {};
  }
}

// Simple parser for meta object
function parseMetaObject(
  objStr: string,
): Record<string, string | { title?: string; display?: string }> {
  const result: Record<string, string | { title?: string; display?: string }> =
    {};

  // Match simple key-value pairs: "key": "value" or key: "value
  const simpleRegex = /"?([^":\s,{}]+)"?\s*:\s*"([^"]+)"/g;

  for (
    let execResult = simpleRegex.exec(objStr);
    execResult !== null;
    execResult = simpleRegex.exec(objStr)
  ) {
    const key = execResult[1];
    const value = execResult[2];
    if (
      key &&
      value &&
      !key.startsWith("---") &&
      key !== "type" &&
      key !== "title" &&
      key !== "display"
    ) {
      result[key] = value;
    }
  }

  // Match object values: key: { title: "value", display: "hidden" }
  const objectRegex = /"?([^":\s,{}]+)"?\s*:\s*\{([^}]+)\}/g;

  for (
    let execResult = objectRegex.exec(objStr);
    execResult !== null;
    execResult = objectRegex.exec(objStr)
  ) {
    const key = execResult[1];
    const objContent = execResult[2];
    if (key && !key.startsWith("---")) {
      const titleMatch = objContent.match(/title:\s*"([^"]+)"/);
      const displayMatch = objContent.match(/display:\s*"([^"]+)"/);

      result[key] = {
        title: titleMatch ? titleMatch[1] : undefined,
        display: displayMatch ? displayMatch[1] : undefined,
      };
    }
  }

  return result;
}

async function buildPageMap(
  dir: string,
  baseRoute = "",
): Promise<PageMapItem[]> {
  const items: PageMapItem[] = [];
  let metaData: Record<string, string | { title?: string; display?: string }> =
    {};

  try {
    const entries = await readdir(dir, { withFileTypes: true });

    // First, check for _meta.ts to get ordering and titles
    const metaPath = join(dir, "_meta.ts");
    try {
      await stat(metaPath);
      metaData = await parseMetaFile(metaPath);
      // Add data entry for _meta
      items.push({ name: "_meta", route: baseRoute || "/", data: metaData });
    } catch {
      // No _meta.ts
    }

    // Build items map first
    const itemsMap: Map<string, PageMapItem> = new Map();

    for (const entry of entries) {
      if (
        entry.name.startsWith("_") ||
        entry.name.startsWith(".") ||
        entry.name.startsWith("[")
      ) {
        continue;
      }

      const fullPath = join(dir, entry.name);

      if (entry.isDirectory()) {
        const dirRoute = baseRoute
          ? `${baseRoute}/${entry.name}`
          : `/${entry.name}`;
        const children = await buildPageMap(fullPath, dirRoute);

        if (children.length > 0) {
          // Get title from meta data
          const metaValue = metaData[entry.name];
          const title =
            typeof metaValue === "string"
              ? metaValue
              : metaValue?.title || toTitleCase(entry.name);

          itemsMap.set(entry.name, {
            name: entry.name,
            route: dirRoute,
            title,
            children,
          });
        }
      } else if (entry.name.endsWith(".mdx") || entry.name.endsWith(".md")) {
        const name = entry.name.replace(/\.(mdx|md)$/, "");
        const isIndex = name === "page" || name === "index";
        const pageName = isIndex ? "index" : name;
        const pageRoute = isIndex
          ? baseRoute || "/"
          : baseRoute
            ? `${baseRoute}/${name}`
            : `/${name}`;

        // Get title from meta data
        const metaKey = isIndex ? "index" : name;
        const metaValue = metaData[metaKey];
        const title =
          typeof metaValue === "string"
            ? metaValue
            : metaValue?.title || (isIndex ? "Index" : toTitleCase(name));

        // Check if display is hidden
        const display =
          typeof metaValue === "object" ? metaValue?.display : undefined;
        if (display === "hidden") {
          continue; // Skip hidden items
        }

        itemsMap.set(pageName, {
          name: pageName,
          route: pageRoute,
          title,
          frontMatter: {},
        });
      }
    }

    // Order items based on _meta.ts order
    if (Object.keys(metaData).length > 0) {
      // Add items in meta order first
      for (const key of Object.keys(metaData)) {
        if (key === "index" || key.startsWith("---")) continue;
        const item = itemsMap.get(key);
        if (item) {
          items.push(item);
          itemsMap.delete(key);
        }
      }
      // Add index if exists
      const indexItem = itemsMap.get("index");
      if (indexItem) {
        items.unshift(indexItem);
        itemsMap.delete("index");
      }
    }

    // Add remaining items (not in meta)
    for (const item of itemsMap.values()) {
      items.push(item);
    }
  } catch (error) {
    console.error(`Error reading directory ${dir}:`, error);
  }

  return items;
}

export async function getPageMap(): Promise<PageMapItem[]> {
  const appDir = join(process.cwd(), "app");
  return buildPageMap(appDir);
}
