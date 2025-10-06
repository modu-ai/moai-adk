// @TEST:INIT-003:DATA | SPEC: SPEC-INIT-003.md
// Related: @CODE:INIT-003:DATA, @SPEC:INIT-003

/**
 * @file Hooks Version Comparison Strategy Tests
 * @author MoAI Team
 * @tags @TEST:INIT-003:DATA
 */

import { describe, it, expect } from 'vitest';
import { mergeHooks } from '../../../../../src/cli/commands/project/merge-strategies/hooks-merger.js';

describe('Hooks Merger', () => {
  describe('Version extraction', () => {
    it('should extract version from hook file content', () => {
      const hookContent = `#!/usr/bin/env node
/**
 * @version 1.2.3
 */

console.log('Hook running');
`;

      const result = mergeHooks(hookContent, hookContent);

      // Should preserve the hook content
      expect(result).toContain('Hook running');
    });

    it('should handle missing version in hook', () => {
      const hookContent = `#!/usr/bin/env node

console.log('No version here');
`;

      const result = mergeHooks(hookContent, hookContent);

      expect(result).toContain('No version here');
    });
  });

  describe('Version comparison', () => {
    it('should use newer version (backup > current)', () => {
      const backup = `#!/usr/bin/env node
/**
 * @version 2.0.0
 */

console.log('Backup hook');
`;

      const current = `#!/usr/bin/env node
/**
 * @version 1.0.0
 */

console.log('Current hook');
`;

      const result = mergeHooks(backup, current);

      // Should use backup (newer version)
      expect(result).toContain('Backup hook');
      expect(result).toContain('@version 2.0.0');
      expect(result).not.toContain('Current hook');
    });

    it('should use newer version (current > backup)', () => {
      const backup = `#!/usr/bin/env node
/**
 * @version 1.0.0
 */

console.log('Backup hook');
`;

      const current = `#!/usr/bin/env node
/**
 * @version 2.0.0
 */

console.log('Current hook');
`;

      const result = mergeHooks(backup, current);

      // Should use current (newer version)
      expect(result).toContain('Current hook');
      expect(result).toContain('@version 2.0.0');
      expect(result).not.toContain('Backup hook');
    });

    it('should use current when versions are equal', () => {
      const backup = `#!/usr/bin/env node
/**
 * @version 1.0.0
 */

console.log('Backup hook');
`;

      const current = `#!/usr/bin/env node
/**
 * @version 1.0.0
 */

console.log('Current hook');
`;

      const result = mergeHooks(backup, current);

      // Should use current (template priority on tie)
      expect(result).toContain('Current hook');
      expect(result).not.toContain('Backup hook');
    });
  });

  describe('Semantic versioning', () => {
    it('should handle patch version differences', () => {
      const backup = `/**
 * @version 1.0.1
 */
console.log('backup');
`;

      const current = `/**
 * @version 1.0.0
 */
console.log('current');
`;

      const result = mergeHooks(backup, current);

      expect(result).toContain('backup');
      expect(result).toContain('1.0.1');
    });

    it('should handle minor version differences', () => {
      const backup = `/**
 * @version 1.1.0
 */
console.log('backup');
`;

      const current = `/**
 * @version 1.2.0
 */
console.log('current');
`;

      const result = mergeHooks(backup, current);

      expect(result).toContain('current');
      expect(result).toContain('1.2.0');
    });

    it('should handle major version differences', () => {
      const backup = `/**
 * @version 2.0.0
 */
console.log('backup');
`;

      const current = `/**
 * @version 1.9.9
 */
console.log('current');
`;

      const result = mergeHooks(backup, current);

      expect(result).toContain('backup');
      expect(result).toContain('2.0.0');
    });

    it('should handle prerelease versions', () => {
      const backup = `/**
 * @version 1.0.0-beta.1
 */
console.log('backup');
`;

      const current = `/**
 * @version 1.0.0-alpha.1
 */
console.log('current');
`;

      const result = mergeHooks(backup, current);

      // beta > alpha in semver
      expect(result).toContain('backup');
    });
  });

  describe('Edge cases', () => {
    it('should handle empty backup', () => {
      const backup = '';
      const current = `/**
 * @version 1.0.0
 */
console.log('current');
`;

      const result = mergeHooks(backup, current);

      expect(result).toBe(current);
    });

    it('should handle empty current', () => {
      const backup = `/**
 * @version 1.0.0
 */
console.log('backup');
`;
      const current = '';

      const result = mergeHooks(backup, current);

      expect(result).toBe(backup);
    });

    it('should handle both empty', () => {
      const result = mergeHooks('', '');

      expect(result).toBe('');
    });

    it('should prefer current when both lack version', () => {
      const backup = `console.log('backup');`;
      const current = `console.log('current');`;

      const result = mergeHooks(backup, current);

      // No version = use current (template priority)
      expect(result).toBe(current);
    });

    it('should handle backup version only', () => {
      const backup = `/**
 * @version 1.0.0
 */
console.log('backup');
`;
      const current = `console.log('current');`;

      const result = mergeHooks(backup, current);

      // Backup has version, current doesn't → use backup
      expect(result).toContain('backup');
    });

    it('should handle current version only', () => {
      const backup = `console.log('backup');`;
      const current = `/**
 * @version 1.0.0
 */
console.log('current');
`;

      const result = mergeHooks(backup, current);

      // Current has version, backup doesn't → use current
      expect(result).toContain('current');
    });
  });
});
