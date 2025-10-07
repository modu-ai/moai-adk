// @CODE:INIT-003:DATA | SPEC: SPEC-INIT-003.md | TEST: __tests__/cli/commands/project/merge-strategies/markdown-merger.test.ts
// Related: @CODE:INIT-003:MERGE, @SPEC:INIT-003

/**
 * @file Markdown Section Merge Strategy for Backup Merge
 * @author MoAI Team
 * @tags @CODE:INIT-003:DATA
 */

import YAML from 'yaml';

/**
 * Merge two Markdown files with HISTORY section accumulation
 *
 * **Merge Strategy**:
 * - Front matter: Deep merge (backup priority)
 * - HISTORY section: Accumulate both (newest first)
 * - Other sections: Use current (template priority)
 *
 * **Use Cases**:
 * - Merge CLAUDE.md (preserve user customizations + update template)
 * - Merge SPEC files (preserve version history)
 * - Merge any Markdown documentation
 *
 * @param backup - Backup markdown content (user's old file)
 * @param current - Current markdown content (new template)
 * @returns Merged markdown with accumulated HISTORY
 *
 * @example
 * ```typescript
 * const backup = `---
 * version: 0.1.0
 * ---
 * ## HISTORY
 * ### v0.1.0
 * - Initial
 * `;
 * const current = `---
 * version: 0.2.0
 * ---
 * ## HISTORY
 * ### v0.2.0
 * - Updated
 * `;
 * const result = mergeMarkdown(backup, current);
 * // Result includes both v0.2.0 and v0.1.0 in HISTORY
 * ```
 */
export function mergeMarkdown(backup: string, current: string): string {
  // Handle empty cases
  if (!backup) return current;
  if (!current) return backup;

  // Parse both files
  const backupParts = parseMarkdown(backup);
  const currentParts = parseMarkdown(current);

  // Merge front matter (backup priority)
  const mergedFrontMatter = mergeFrontMatter(
    backupParts.frontMatter,
    currentParts.frontMatter
  );

  // Merge HISTORY sections
  const mergedHistory = mergeHistorySections(
    backupParts.history,
    currentParts.history
  );

  // Determine body strategy based on presence of HISTORY
  let mergedBody: string;

  if (currentParts.history || backupParts.history) {
    // If either has HISTORY, use current body as base and inject merged HISTORY
    mergedBody = currentParts.body;
    if (mergedHistory) {
      mergedBody = replaceHistorySection(currentParts.body, mergedHistory);
    }
  } else if (backupParts.body && !currentParts.history) {
    // If current has no HISTORY and backup has content, check if we need backup body
    // For now, prefer current body (template priority)
    mergedBody = currentParts.body || backupParts.body;
  } else {
    // Default: use current body (template priority)
    mergedBody = currentParts.body;
  }

  // Reconstruct markdown
  return reconstructMarkdown(mergedFrontMatter, mergedBody);
}

/**
 * Parse markdown into parts
 * @internal
 */
interface MarkdownParts {
  frontMatter: Record<string, unknown> | null;
  history: string | null;
  body: string;
}

function parseMarkdown(content: string): MarkdownParts {
  const frontMatterMatch = content.match(/^---\n([\s\S]*?)\n---\n/);
  let frontMatter: Record<string, unknown> | null = null;
  let restContent = content;

  if (frontMatterMatch) {
    try {
      frontMatter = YAML.parse(frontMatterMatch[1]) as Record<string, unknown>;
      restContent = content.slice(frontMatterMatch[0].length);
    } catch {
      // Ignore YAML parse errors
    }
  }

  // Extract HISTORY section more precisely
  // Look for "## HISTORY" followed by content until next "## " or end
  const historyRegex = /## HISTORY\n([\s\S]*?)(?=\n## [^H]|\n## $|$)/;
  const historyMatch = restContent.match(historyRegex);
  const history = historyMatch
    ? `## HISTORY\n${historyMatch[1].trimEnd()}`
    : null;

  return {
    frontMatter,
    history,
    body: restContent,
  };
}

/**
 * Merge front matter with backup priority
 * @internal
 */
function mergeFrontMatter(
  backup: Record<string, unknown> | null,
  current: Record<string, unknown> | null
): Record<string, unknown> | null {
  if (!backup && !current) return null;
  if (!backup) return current;
  if (!current) return backup;

  // Deep merge: current as base, backup overrides
  const result: Record<string, unknown> = {};

  // First add all current keys (new fields)
  for (const key in current) {
    if (Object.hasOwn(current, key)) {
      result[key] = current[key];
    }
  }

  // Then override with backup values (user priority)
  for (const key in backup) {
    if (Object.hasOwn(backup, key)) {
      const backupValue = backup[key];
      const currentValue = current[key];

      // If both are plain objects, merge recursively
      if (isPlainObject(backupValue) && isPlainObject(currentValue)) {
        result[key] = mergeFrontMatter(
          backupValue as Record<string, unknown>,
          currentValue as Record<string, unknown>
        );
      } else {
        // Backup value takes priority
        result[key] = backupValue;
      }
    }
  }

  return result;
}

/**
 * Check if value is a plain object (not array, not null)
 * @internal
 */
function isPlainObject(value: unknown): boolean {
  return (
    value !== null &&
    typeof value === 'object' &&
    !Array.isArray(value) &&
    Object.prototype.toString.call(value) === '[object Object]'
  );
}

/**
 * Merge HISTORY sections (accumulate both, newest first)
 * @internal
 */
function mergeHistorySections(
  backupHistory: string | null,
  currentHistory: string | null
): string | null {
  if (!backupHistory && !currentHistory) return null;
  if (!backupHistory) return currentHistory;
  if (!currentHistory) return backupHistory;

  // Extract version entries from both
  const backupEntries = extractHistoryEntries(backupHistory);
  const currentEntries = extractHistoryEntries(currentHistory);

  // Combine and deduplicate by version
  const allEntries = [...currentEntries, ...backupEntries];
  const uniqueEntries = deduplicateByVersion(allEntries);

  // Reconstruct HISTORY section
  return `## HISTORY\n\n${uniqueEntries.join('\n\n')}`;
}

/**
 * Extract individual version entries from HISTORY section
 * @internal
 */
function extractHistoryEntries(history: string): string[] {
  // Remove "## HISTORY" header
  const content = history.replace(/^## HISTORY\n+/, '');

  // Split by ### (version markers)
  const entries = content.split(/(?=### v)/);

  return entries.filter(entry => entry.trim().length > 0);
}

/**
 * Deduplicate history entries by version
 * @internal
 */
function deduplicateByVersion(entries: string[]): string[] {
  const seen = new Set<string>();
  const unique: string[] = [];

  for (const entry of entries) {
    const versionMatch = entry.match(/### (v[\d.]+)/);
    if (versionMatch) {
      const version = versionMatch[1];
      if (!seen.has(version)) {
        seen.add(version);
        unique.push(entry.trim());
      }
    } else {
      // Entry without version, keep it
      unique.push(entry.trim());
    }
  }

  return unique;
}

/**
 * Replace HISTORY section in body with merged version
 * @internal
 */
function replaceHistorySection(body: string, mergedHistory: string): string {
  // Try to find and replace existing HISTORY section
  const historyRegex = /^## HISTORY\n[\s\S]*?(?=\n## |\n---|\n$|$)/m;

  if (historyRegex.test(body)) {
    return body.replace(historyRegex, mergedHistory);
  }

  // If no HISTORY section exists, add it after front matter and title
  // Find first ## heading (not HISTORY)
  const firstSectionMatch = body.match(/^## (?!HISTORY)/m);

  if (firstSectionMatch && firstSectionMatch.index !== undefined) {
    return (
      body.slice(0, firstSectionMatch.index) +
      `${mergedHistory}\n\n` +
      body.slice(firstSectionMatch.index)
    );
  }

  // If no sections, append at end
  return `${body}\n\n${mergedHistory}`;
}

/**
 * Reconstruct markdown from parts
 * @internal
 */
function reconstructMarkdown(
  frontMatter: Record<string, unknown> | null,
  body: string
): string {
  if (!frontMatter) {
    return body;
  }

  // Stringify with consistent key order
  const yamlString = YAML.stringify(frontMatter, {
    sortMapEntries: false, // Preserve insertion order
  });

  // Ensure yamlString ends without extra newline (YAML.stringify adds one)
  const trimmedYaml = yamlString.endsWith('\n')
    ? yamlString.slice(0, -1)
    : yamlString;

  return `---\n${trimmedYaml}\n---\n${body}`;
}
