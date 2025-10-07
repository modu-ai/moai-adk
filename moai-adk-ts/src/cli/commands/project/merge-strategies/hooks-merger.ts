// @CODE:INIT-003:DATA | SPEC: SPEC-INIT-003.md | TEST: __tests__/cli/commands/project/merge-strategies/hooks-merger.test.ts
// Related: @CODE:INIT-003:MERGE, @SPEC:INIT-003

/**
 * @file Hooks Version Comparison Strategy for Backup Merge
 * @author MoAI Team
 * @tags @CODE:INIT-003:DATA
 */

import semver from 'semver';

/**
 * Merge two hook files by comparing versions
 *
 * **Merge Strategy**:
 * - Extract @version tag from each hook file
 * - Compare versions using semver
 * - Use the newer version
 * - On tie or missing versions: use current (template priority)
 *
 * **Version Extraction**:
 * - Looks for `@version X.Y.Z` in comments
 * - Supports semantic versioning (major.minor.patch)
 * - Supports prerelease tags (e.g., 1.0.0-beta.1)
 *
 * **Use Cases**:
 * - Merge .claude/hooks/*.cjs files
 * - Preserve user customizations if newer
 * - Update to latest template hooks
 *
 * @param backup - Backup hook content (user's old hook)
 * @param current - Current hook content (new template hook)
 * @returns The hook with the newer version
 *
 * @example
 * ```typescript
 * const backup = `/**
 *  * @version 2.0.0
 *  *\/
 * console.log('backup');
 * `;
 * const current = `/**
 *  * @version 1.0.0
 *  *\/
 * console.log('current');
 * `;
 * const result = mergeHooks(backup, current);
 * // Returns backup (2.0.0 > 1.0.0)
 * ```
 */
export function mergeHooks(backup: string, current: string): string {
  // Handle empty cases
  if (!backup) return current;
  if (!current) return backup;

  // Extract versions
  const backupVersion = extractVersion(backup);
  const currentVersion = extractVersion(current);

  // If both lack versions, prefer current (template priority)
  if (!backupVersion && !currentVersion) {
    return current;
  }

  // If only one has version, use that one
  if (backupVersion && !currentVersion) {
    return backup;
  }
  if (!backupVersion && currentVersion) {
    return current;
  }

  // Both have versions - compare using semver
  // semver.gt returns true if first > second
  if (semver.gt(backupVersion!, currentVersion!)) {
    return backup;
  }

  // Current is newer or equal - use current (template priority on tie)
  return current;
}

/**
 * Extract version string from hook file content
 *
 * Looks for @version tag in comments:
 * - `@version 1.2.3`
 * - `@version 1.0.0-beta.1`
 *
 * @param content - Hook file content
 * @returns Version string or null if not found
 * @internal
 */
function extractVersion(content: string): string | null {
  // Match @version X.Y.Z (with optional prerelease)
  const versionRegex = /@version\s+([\d.]+(?:-[\w.]+)?)/;
  const match = content.match(versionRegex);

  if (match?.[1]) {
    const version = match[1];
    // Validate it's a valid semver
    return semver.valid(version);
  }

  return null;
}
