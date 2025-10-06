// @TEST:INIT-003:MERGE | SPEC: SPEC-INIT-003.md
// Related: @CODE:INIT-003:MERGE, @SPEC:INIT-003

/**
 * @file Backup Merger Integration Tests
 * @author MoAI Team
 * @tags @TEST:INIT-003:MERGE
 */

import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import * as fs from 'node:fs/promises';
import * as path from 'node:path';
import { BackupMerger } from '../../../../src/cli/commands/project/backup-merger.js';

describe('BackupMerger', () => {
  let TEST_DIR: string;
  let BACKUP_DIR: string;
  let CURRENT_DIR: string;

  beforeEach(async () => {
    // Create unique test directory for each test
    TEST_DIR = `/tmp/moai-backup-merger-test-${Date.now()}-${Math.random().toString(36).slice(2)}`;
    BACKUP_DIR = path.join(TEST_DIR, '.moai-backup-test');
    CURRENT_DIR = TEST_DIR;

    // Clean up and create test directory
    await fs.rm(TEST_DIR, { recursive: true, force: true });
    await fs.mkdir(TEST_DIR, { recursive: true });
    await fs.mkdir(BACKUP_DIR, { recursive: true });
  });

  afterEach(async () => {
    // Clean up after tests
    await fs.rm(TEST_DIR, { recursive: true, force: true });
  });

  describe('File type detection', () => {
    it('should merge JSON files using JSON merger', async () => {
      // Setup backup
      const backupConfig = { mode: 'personal', custom: true };
      await fs.writeFile(
        path.join(BACKUP_DIR, 'config.json'),
        JSON.stringify(backupConfig, null, 2)
      );

      // Setup current
      const currentConfig = { mode: 'team', version: '1.0.0' };
      await fs.writeFile(
        path.join(CURRENT_DIR, 'config.json'),
        JSON.stringify(currentConfig, null, 2)
      );

      // Merge
      const merger = new BackupMerger(BACKUP_DIR, CURRENT_DIR);
      await merger.mergeFile('config.json');

      // Verify result
      const result = JSON.parse(
        await fs.readFile(path.join(CURRENT_DIR, 'config.json'), 'utf-8')
      );

      // Backup values should take priority
      expect(result.mode).toBe('personal');
      expect(result.custom).toBe(true);

      // New fields from current should be added
      expect(result.version).toBe('1.0.0');
    });

    it('should merge Markdown files using Markdown merger', async () => {
      // Setup backup
      const backup = `---
version: 0.1.0
---

# Test

## HISTORY

### v0.1.0
- Initial
`;
      await fs.writeFile(path.join(BACKUP_DIR, 'README.md'), backup);

      // Setup current
      const current = `---
version: 0.2.0
---

# Test

## HISTORY

### v0.2.0
- Updated
`;
      await fs.writeFile(path.join(CURRENT_DIR, 'README.md'), current);

      // Merge
      const merger = new BackupMerger(BACKUP_DIR, CURRENT_DIR);
      await merger.mergeFile('README.md');

      // Verify result
      const result = await fs.readFile(path.join(CURRENT_DIR, 'README.md'), 'utf-8');

      // Should have both HISTORY entries
      expect(result).toContain('### v0.2.0');
      expect(result).toContain('### v0.1.0');

      // Backup version should take priority
      expect(result).toContain('version: 0.1.0');
    });

    it('should merge hook files using Hooks merger', async () => {
      // Setup backup (older version)
      const backup = `/**
 * @version 1.0.0
 */
console.log('backup');
`;
      await fs.writeFile(path.join(BACKUP_DIR, 'hook.cjs'), backup);

      // Setup current (newer version)
      const current = `/**
 * @version 2.0.0
 */
console.log('current');
`;
      await fs.writeFile(path.join(CURRENT_DIR, 'hook.cjs'), current);

      // Merge
      const merger = new BackupMerger(BACKUP_DIR, CURRENT_DIR);
      await merger.mergeFile('hook.cjs');

      // Verify result
      const result = await fs.readFile(path.join(CURRENT_DIR, 'hook.cjs'), 'utf-8');

      // Current should win (newer version)
      expect(result).toContain('current');
      expect(result).toContain('@version 2.0.0');
    });

    it('should copy other files without merging', async () => {
      // Setup backup
      await fs.writeFile(path.join(BACKUP_DIR, 'file.txt'), 'backup content');

      // Setup current (different content)
      await fs.writeFile(path.join(CURRENT_DIR, 'file.txt'), 'current content');

      // Merge
      const merger = new BackupMerger(BACKUP_DIR, CURRENT_DIR);
      await merger.mergeFile('file.txt');

      // Verify result
      const result = await fs.readFile(path.join(CURRENT_DIR, 'file.txt'), 'utf-8');

      // Should preserve current (template priority for non-mergeable files)
      expect(result).toBe('current content');
    });
  });

  describe('Directory merging', () => {
    it('should merge all files in a directory', async () => {
      // Setup backup directory
      await fs.mkdir(path.join(BACKUP_DIR, '.claude'), { recursive: true });
      await fs.writeFile(
        path.join(BACKUP_DIR, '.claude', 'settings.json'),
        JSON.stringify({ mode: 'personal' }, null, 2)
      );

      // Setup current directory
      await fs.mkdir(path.join(CURRENT_DIR, '.claude'), { recursive: true });
      await fs.writeFile(
        path.join(CURRENT_DIR, '.claude', 'settings.json'),
        JSON.stringify({ mode: 'team', version: '1.0.0' }, null, 2)
      );

      // Merge
      const merger = new BackupMerger(BACKUP_DIR, CURRENT_DIR);
      await merger.mergeDirectory('.claude');

      // Verify result
      const result = JSON.parse(
        await fs.readFile(path.join(CURRENT_DIR, '.claude', 'settings.json'), 'utf-8')
      );

      expect(result.mode).toBe('personal');
      expect(result.version).toBe('1.0.0');
    });

    it('should handle nested directories', async () => {
      // Setup backup
      await fs.mkdir(path.join(BACKUP_DIR, '.claude', 'hooks'), { recursive: true });
      await fs.writeFile(
        path.join(BACKUP_DIR, '.claude', 'hooks', 'test.cjs'),
        'backup hook'
      );

      // Setup current
      await fs.mkdir(path.join(CURRENT_DIR, '.claude', 'hooks'), { recursive: true });
      await fs.writeFile(
        path.join(CURRENT_DIR, '.claude', 'hooks', 'test.cjs'),
        'current hook'
      );

      // Merge
      const merger = new BackupMerger(BACKUP_DIR, CURRENT_DIR);
      await merger.mergeDirectory('.claude');

      // Verify result exists
      const result = await fs.readFile(
        path.join(CURRENT_DIR, '.claude', 'hooks', 'test.cjs'),
        'utf-8'
      );

      expect(result).toBeDefined();
    });
  });

  describe('Error handling', () => {
    it('should handle missing backup file', async () => {
      // Setup current only
      await fs.writeFile(path.join(CURRENT_DIR, 'file.txt'), 'current');

      // Merge (backup doesn't exist)
      const merger = new BackupMerger(BACKUP_DIR, CURRENT_DIR);
      await merger.mergeFile('file.txt');

      // Should keep current
      const result = await fs.readFile(path.join(CURRENT_DIR, 'file.txt'), 'utf-8');
      expect(result).toBe('current');
    });

    it('should handle missing current file', async () => {
      // Setup backup only
      await fs.writeFile(path.join(BACKUP_DIR, 'file.txt'), 'backup');

      // Merge (current doesn't exist)
      const merger = new BackupMerger(BACKUP_DIR, CURRENT_DIR);
      await merger.mergeFile('file.txt');

      // Should create current from backup
      const result = await fs.readFile(path.join(CURRENT_DIR, 'file.txt'), 'utf-8');
      expect(result).toBe('backup');
    });

    it('should handle missing directories gracefully', async () => {
      const merger = new BackupMerger(BACKUP_DIR, CURRENT_DIR);

      // Should not throw when merging non-existent directory
      await merger.mergeDirectory('nonexistent');

      // Should succeed without errors
      const report = merger.getReport();
      expect(report.errors.length).toBe(0);
    });
  });

  describe('Merge report generation', () => {
    it('should track merged files', async () => {
      // Setup files
      await fs.writeFile(path.join(BACKUP_DIR, 'file1.json'), JSON.stringify({ a: 1 }));
      await fs.writeFile(path.join(CURRENT_DIR, 'file1.json'), JSON.stringify({ b: 2 }));

      // Merge
      const merger = new BackupMerger(BACKUP_DIR, CURRENT_DIR);
      await merger.mergeFile('file1.json');

      // Get report
      const report = merger.getReport();

      expect(report.merged).toContain('file1.json');
    });

    it('should track skipped files', async () => {
      // Setup current only (no backup)
      await fs.writeFile(path.join(CURRENT_DIR, 'file1.txt'), 'current');

      // Merge (backup missing)
      const merger = new BackupMerger(BACKUP_DIR, CURRENT_DIR);
      await merger.mergeFile('file1.txt');

      // Get report
      const report = merger.getReport();

      expect(report.skipped).toContain('file1.txt');
    });

    it('should track errors', async () => {
      // Setup backup file with invalid permissions (simulate read error)
      // Note: We'll create a file and then try to read from a non-readable location
      const merger = new BackupMerger('/nonexistent/invalid/path', CURRENT_DIR);

      // Try to merge (should add to errors but not throw)
      await merger.mergeFile('file.txt');

      // Get report
      const report = merger.getReport();

      // Error should be tracked (backup path doesn't exist is handled gracefully)
      // Actually, non-existent backup results in skip, not error
      // Let's test with a different scenario - create backup but simulate error during merge

      // Better test: Create valid backup and current, but cause error in merge
      const merger2 = new BackupMerger(BACKUP_DIR, CURRENT_DIR);

      // Create backup and current with same path
      await fs.mkdir(path.join(BACKUP_DIR, 'subdir'), { recursive: true });
      await fs.writeFile(path.join(BACKUP_DIR, 'subdir', 'file.txt'), 'backup');

      // Create current as directory (will cause error when trying to write)
      await fs.mkdir(path.join(CURRENT_DIR, 'subdir', 'file.txt'), { recursive: true });

      // Try to merge (file vs directory conflict)
      await merger2.mergeFile('subdir/file.txt');

      const report2 = merger2.getReport();
      expect(report2.errors.length).toBeGreaterThan(0);
    });
  });
});
