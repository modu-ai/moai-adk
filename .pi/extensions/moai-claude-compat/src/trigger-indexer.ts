import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative } from "node:path";
import { asString, parseMarkdownFrontmatter } from "./frontmatter.ts";
import { PI_SKILLS_SOURCE_PATH } from "./constants.ts";

export interface SkillTriggerEntry {
  name: string;
  path: string;
  description: string;
  keywords: string[];
}

function walkSkillFiles(root: string): string[] {
  if (!existsSync(root)) return [];
  const out: string[] = [];
  for (const entry of readdirSync(root, { withFileTypes: true })) {
    const path = join(root, entry.name);
    if (entry.isDirectory()) out.push(...walkSkillFiles(path));
    else if (entry.isFile() && entry.name === "SKILL.md") out.push(path);
  }
  return out;
}

function extractKeywords(name: string, description: string, body: string): string[] {
  const text = `${name} ${description} ${body.slice(0, 2000)}`.toLowerCase();
  const words = new Set<string>();
  for (const word of text.match(/[a-z0-9][a-z0-9_-]{2,}/g) ?? []) {
    if (["the", "and", "for", "with", "when", "use", "this", "that", "from", "your", "into"].includes(word)) continue;
    words.add(word);
  }
  return [...words].slice(0, 80);
}

export function buildSkillTriggerIndex(root = PI_SKILLS_SOURCE_PATH): SkillTriggerEntry[] {
  return walkSkillFiles(root).map((path) => {
    const text = readFileSync(path, "utf8");
    const parsed = parseMarkdownFrontmatter(text);
    const name = asString(parsed.frontmatter.name, relative(root, path).split(/[\\/]/)[0]);
    const description = asString(parsed.frontmatter.description, "");
    return {
      name,
      path,
      description,
      keywords: extractKeywords(name, description, parsed.body),
    };
  });
}

export function buildSkillTriggerHints(input: string, limit = 5): string[] {
  const queryWords = new Set((input.toLowerCase().match(/[a-z0-9][a-z0-9_-]{2,}/g) ?? []).filter(Boolean));
  if (queryWords.size === 0) return [];

  const scored = buildSkillTriggerIndex().map((entry) => {
    let score = 0;
    for (const word of queryWords) {
      if (entry.name.toLowerCase().includes(word)) score += 4;
      if (entry.description.toLowerCase().includes(word)) score += 2;
      if (entry.keywords.includes(word)) score += 1;
    }
    return { entry, score };
  });

  return scored
    .filter((s) => s.score > 0)
    .sort((a, b) => b.score - a.score || a.entry.name.localeCompare(b.entry.name))
    .slice(0, limit)
    .map((s) => `Relevant MoAI skill hint: read ${s.entry.path} (${s.entry.name}) before acting.`);
}

export function getSkillIndexStatus(expected = 38): string {
  const index = buildSkillTriggerIndex();
  let latest = 0;
  for (const entry of index) latest = Math.max(latest, statSync(entry.path).mtimeMs);
  const status = index.length === expected ? "ok" : "warn(non-blocking)";
  return `${status}: skill trigger index source count ${index.length}/${expected}${latest ? ` latest=${new Date(latest).toISOString()}` : ""}`;
}
