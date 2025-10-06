// @TEST:INIT-003:DATA | SPEC: SPEC-INIT-003.md | CODE: src/cli/commands/project/merge-strategies/json-merger.ts
// Related: @CODE:INIT-003:MERGE, @SPEC:INIT-003

/**
 * @file Test for JSON Deep Merge Strategy
 * @author MoAI Team
 * @tags @TEST:INIT-003:DATA
 */

import { describe, test, expect } from 'vitest';
import { mergeJSON } from '@/cli/commands/project/merge-strategies/json-merger';

describe('JSON Merger', () => {
  describe('Simple object merge', () => {
    test('should merge two simple objects', () => {
      // Given: Two simple objects
      const backup = { a: 1, b: 2 };
      const current = { b: 3, c: 4 };

      // When: Merge
      const result = mergeJSON(backup, current);

      // Then: Backup values take priority, new fields added
      expect(result).toEqual({ a: 1, b: 2, c: 4 });
    });

    test('should add new fields from current', () => {
      // Given: Backup without some fields
      const backup = { mode: 'personal' };
      const current = { mode: 'team', newFeature: true };

      // When: Merge
      const result = mergeJSON(backup, current);

      // Then: Backup value kept, new field added
      expect(result).toEqual({ mode: 'personal', newFeature: true });
    });
  });

  describe('Nested object merge', () => {
    test('should merge nested objects (3 levels)', () => {
      // Given: Nested objects
      const backup = {
        level1: {
          level2: {
            level3: { user: 'old' },
          },
        },
      };
      const current = {
        level1: {
          level2: {
            level3: { user: 'new', added: true },
          },
        },
      };

      // When: Merge
      const result = mergeJSON(backup, current);

      // Then: Deep merge preserves backup values
      expect(result).toEqual({
        level1: {
          level2: {
            level3: { user: 'old', added: true },
          },
        },
      });
    });

    test('should handle partial nested structures', () => {
      // Given: Partial nested structures
      const backup = {
        hooks: {
          PreToolUse: ['custom-hook.cjs'],
        },
      };
      const current = {
        hooks: {
          PreToolUse: ['default-hook.cjs'],
          SessionStart: ['session-hook.cjs'],
        },
      };

      // When: Merge
      const result = mergeJSON(backup, current);

      // Then: Both hooks preserved
      expect(result).toEqual({
        hooks: {
          PreToolUse: ['custom-hook.cjs'],
          SessionStart: ['session-hook.cjs'],
        },
      });
    });
  });

  describe('Array handling', () => {
    test('should replace arrays (not merge)', () => {
      // Given: Arrays in backup and current
      const backup = { tags: ['a', 'b'] };
      const current = { tags: ['c', 'd'] };

      // When: Merge
      const result = mergeJSON(backup, current);

      // Then: Backup array takes priority
      expect(result).toEqual({ tags: ['a', 'b'] });
    });

    test('should preserve empty arrays', () => {
      // Given: Empty array in backup
      const backup = { list: [] };
      const current = { list: [1, 2, 3] };

      // When: Merge
      const result = mergeJSON(backup, current);

      // Then: Empty array preserved
      expect(result).toEqual({ list: [] });
    });
  });

  describe('Special cases', () => {
    test('should handle null values', () => {
      // Given: Null values
      const backup = { value: null };
      const current = { value: 'something' };

      // When: Merge
      const result = mergeJSON(backup, current);

      // Then: Backup null preserved
      expect(result).toEqual({ value: null });
    });

    test('should handle undefined values', () => {
      // Given: Undefined in backup
      const backup = { defined: 'yes' };
      const current = { defined: 'no', other: 'value' };

      // When: Merge
      const result = mergeJSON(backup, current);

      // Then: Backup value kept
      expect(result).toEqual({ defined: 'yes', other: 'value' });
    });

    test('should handle empty objects', () => {
      // Given: Empty objects
      const backup = {};
      const current = { a: 1, b: 2 };

      // When: Merge
      const result = mergeJSON(backup, current);

      // Then: Current values used
      expect(result).toEqual({ a: 1, b: 2 });
    });
  });

  describe('Real-world scenarios', () => {
    test('should merge Claude Code settings.json', () => {
      // Given: Real settings.json structures
      const backup = {
        mode: 'personal',
        hooks: {
          PreToolUse: ['custom-security.cjs', 'custom-tag.cjs'],
        },
      };
      const current = {
        mode: 'team',
        hooks: {
          PreToolUse: ['default-security.cjs'],
          SessionStart: ['session-notice.cjs'],
        },
        newFeature: {
          enabled: true,
        },
      };

      // When: Merge
      const result = mergeJSON(backup, current);

      // Then: User settings preserved, new features added
      expect(result).toEqual({
        mode: 'personal',
        hooks: {
          PreToolUse: ['custom-security.cjs', 'custom-tag.cjs'],
          SessionStart: ['session-notice.cjs'],
        },
        newFeature: {
          enabled: true,
        },
      });
    });

    test('should merge .moai/config.json', () => {
      // Given: Config structures
      const backup = {
        project_name: 'my-app',
        project_language: 'typescript',
        customField: 'user-value',
      };
      const current = {
        project_name: 'default',
        project_language: 'javascript',
        version: '0.3.0',
      };

      // When: Merge
      const result = mergeJSON(backup, current);

      // Then: User config preserved, version updated
      expect(result).toEqual({
        project_name: 'my-app',
        project_language: 'typescript',
        customField: 'user-value',
        version: '0.3.0',
      });
    });
  });
});
