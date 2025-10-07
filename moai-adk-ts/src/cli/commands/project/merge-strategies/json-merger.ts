// @CODE:INIT-003:DATA | SPEC: SPEC-INIT-003.md | TEST: __tests__/cli/commands/project/merge-strategies/json-merger.test.ts
// Related: @CODE:INIT-003:MERGE, @SPEC:INIT-003

/**
 * @file JSON Deep Merge Strategy for Backup Merge
 * @author MoAI Team
 * @tags @CODE:INIT-003:DATA
 */

/**
 * Deep merge two JSON objects with backup priority
 *
 * **Merge Strategy**:
 * - Backup values take priority over current values
 * - New fields from current are added to result
 * - Nested objects are merged recursively
 * - Arrays are replaced (not merged)
 * - Null and undefined values are preserved
 *
 * **Use Cases**:
 * - Merge .claude/settings.json (preserve user settings)
 * - Merge .moai/config.json (preserve user config)
 * - Merge any JSON configuration files
 *
 * @param backup - Backup object (user's old config)
 * @param current - Current object (new template defaults)
 * @returns Merged object with backup priority
 *
 * @example
 * ```typescript
 * const backup = { mode: 'personal', custom: true };
 * const current = { mode: 'team', version: '1.0.0' };
 * const result = mergeJSON(backup, current);
 * // Result: { mode: 'personal', custom: true, version: '1.0.0' }
 * ```
 */
export function mergeJSON(
  backup: Record<string, unknown>,
  current: Record<string, unknown>
): Record<string, unknown> {
  // Start with current as base (includes new fields)
  const result: Record<string, unknown> = { ...current };

  // Override with backup values (user priority)
  for (const key in backup) {
    if (!Object.hasOwn(backup, key)) {
      continue;
    }

    const backupValue = backup[key];
    const currentValue = current[key];

    // If backup value doesn't exist, skip (keep current)
    if (backupValue === undefined) {
      continue;
    }

    // If current doesn't have this key, use backup value
    if (!(key in current)) {
      result[key] = backupValue;
      continue;
    }

    // If both are plain objects, merge recursively
    if (isPlainObject(backupValue) && isPlainObject(currentValue)) {
      result[key] = mergeJSON(
        backupValue as Record<string, unknown>,
        currentValue as Record<string, unknown>
      );
    } else {
      // For primitives, arrays, null, replace with backup value
      result[key] = backupValue;
    }
  }

  return result;
}

/**
 * Check if value is a plain object (not array, not null)
 * @param value - Value to check
 * @returns True if plain object
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
