// @TEST:INIT-003:BACKUP | SPEC: SPEC-INIT-003.md | CODE: src/core/installer/phase-executor.ts
// Related: @CODE:INIT-003:DATA, @SPEC:INIT-003

/**
 * @file Test for Phase Executor with Backup Metadata
 * @author MoAI Team
 * @tags @TEST:INIT-003:BACKUP
 */

import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';
import * as fs from 'node:fs';
import * as path from 'node:path';
import { PhaseExecutor } from '@/core/installer/phase-executor';
import { ContextManager } from '@/core/installer/context-manager';
import { loadBackupMetadata } from '@/core/installer/backup-metadata';
import type { InstallationContext } from '@/core/installer/types';

const TEST_PROJECT_PATH = path.join('/tmp', '__test-phase-executor__');

describe('Phase Executor - Backup Metadata Integration', () => {
  let executor: PhaseExecutor;
  let contextManager: ContextManager;

  beforeEach(async () => {
    // Clean up test directory
    if (fs.existsSync(TEST_PROJECT_PATH)) {
      await fs.promises.rm(TEST_PROJECT_PATH, { recursive: true });
    }
    await fs.promises.mkdir(TEST_PROJECT_PATH, { recursive: true });

    // Create test context
    contextManager = new ContextManager();
    executor = new PhaseExecutor(contextManager);
  });

  afterEach(async () => {
    // Clean up test directory
    if (fs.existsSync(TEST_PROJECT_PATH)) {
      await fs.promises.rm(TEST_PROJECT_PATH, { recursive: true });
    }
  });

  describe('executePreparationPhase with backup', () => {
    test('should create backup metadata after backup creation', async () => {
      // Given: Valid context with backup enabled
      const context: InstallationContext = {
        config: {
          projectPath: TEST_PROJECT_PATH,
          projectName: 'test-project',
          projectLanguage: 'typescript',
          mode: 'personal',
          backupEnabled: true,
          npmInstallEnabled: false,
        },
        state: {},
        phases: [],
        allFilesCreated: [],
        allErrors: [],
      };

      // Create some files to backup
      await fs.promises.mkdir(path.join(TEST_PROJECT_PATH, '.claude'), {
        recursive: true,
      });
      await fs.promises.writeFile(
        path.join(TEST_PROJECT_PATH, 'CLAUDE.md'),
        '# Test'
      );

      // When: Execute preparation phase
      await executor.executePreparationPhase(context);

      // Then: Backup metadata file created
      const metadataPath = path.join(
        TEST_PROJECT_PATH,
        '.moai',
        'backups',
        'latest.json'
      );
      expect(fs.existsSync(metadataPath)).toBe(true);

      // And: Metadata has correct structure
      const metadata = await loadBackupMetadata(TEST_PROJECT_PATH);
      expect(metadata).not.toBeNull();
      expect(metadata?.status).toBe('pending');
      expect(metadata?.created_by).toBe('moai init');
      expect(metadata?.backed_up_files).toBeInstanceOf(Array);
    });

    test('should not create metadata if backup is disabled', async () => {
      // Given: Context with backup disabled
      const context: InstallationContext = {
        config: {
          projectPath: TEST_PROJECT_PATH,
          projectName: 'test-project',
          projectLanguage: 'typescript',
          mode: 'personal',
          backupEnabled: false,
          npmInstallEnabled: false,
        },
        state: {},
        phases: [],
        allFilesCreated: [],
        allErrors: [],
      };

      // When: Execute preparation phase
      await executor.executePreparationPhase(context);

      // Then: No metadata file created
      const metadataPath = path.join(
        TEST_PROJECT_PATH,
        '.moai',
        'backups',
        'latest.json'
      );
      expect(fs.existsSync(metadataPath)).toBe(false);
    });

    test('should not create metadata if backup creation fails', async () => {
      // Given: Invalid project path (backup will fail)
      const invalidPath = path.join(process.cwd(), '__non-existent__');
      const context: InstallationContext = {
        config: {
          projectPath: invalidPath,
          projectName: 'test-project',
          projectLanguage: 'typescript',
          mode: 'personal',
          backupEnabled: true,
          npmInstallEnabled: false,
        },
        state: {},
        phases: [],
        allFilesCreated: [],
        allErrors: [],
      };

      // When/Then: Backup fails, no metadata created
      await expect(
        executor.executePreparationPhase(context)
      ).rejects.toThrow();

      // And: Metadata file not created
      const metadataPath = path.join(
        invalidPath,
        '.moai',
        'backups',
        'latest.json'
      );
      expect(fs.existsSync(metadataPath)).toBe(false);
    });
  });

  describe('backup metadata content', () => {
    test('should include backed up file list in metadata', async () => {
      // Given: Project with multiple backup targets
      const context: InstallationContext = {
        config: {
          projectPath: TEST_PROJECT_PATH,
          projectName: 'test-project',
          projectLanguage: 'typescript',
          mode: 'personal',
          backupEnabled: true,
          npmInstallEnabled: false,
        },
        state: {},
        phases: [],
        allFilesCreated: [],
        allErrors: [],
      };

      // Create backup targets
      await fs.promises.mkdir(path.join(TEST_PROJECT_PATH, '.claude'), {
        recursive: true,
      });
      await fs.promises.mkdir(path.join(TEST_PROJECT_PATH, '.moai'), {
        recursive: true,
      });
      await fs.promises.writeFile(
        path.join(TEST_PROJECT_PATH, 'CLAUDE.md'),
        '# Test'
      );

      // When: Execute preparation phase
      await executor.executePreparationPhase(context);

      // Then: Metadata contains file list
      const metadata = await loadBackupMetadata(TEST_PROJECT_PATH);
      expect(metadata?.backed_up_files.length).toBeGreaterThan(0);
    });

    test('should include backup path in metadata', async () => {
      // Given: Valid context
      const context: InstallationContext = {
        config: {
          projectPath: TEST_PROJECT_PATH,
          projectName: 'test-project',
          projectLanguage: 'typescript',
          mode: 'personal',
          backupEnabled: true,
          npmInstallEnabled: false,
        },
        state: {},
        phases: [],
        allFilesCreated: [],
        allErrors: [],
      };

      // When: Execute preparation phase
      await executor.executePreparationPhase(context);

      // Then: Metadata contains backup path
      const metadata = await loadBackupMetadata(TEST_PROJECT_PATH);
      expect(metadata?.backup_path).toMatch(/\.moai-backup-/);
    });

    test('should include ISO timestamp in metadata', async () => {
      // Given: Valid context
      const context: InstallationContext = {
        config: {
          projectPath: TEST_PROJECT_PATH,
          projectName: 'test-project',
          projectLanguage: 'typescript',
          mode: 'personal',
          backupEnabled: true,
          npmInstallEnabled: false,
        },
        state: {},
        phases: [],
        allFilesCreated: [],
        allErrors: [],
      };

      // When: Execute preparation phase
      await executor.executePreparationPhase(context);

      // Then: Metadata contains valid ISO timestamp
      const metadata = await loadBackupMetadata(TEST_PROJECT_PATH);
      expect(metadata?.timestamp).toMatch(
        /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/
      );

      // And: Timestamp is recent (within last minute)
      const timestampDate = new Date(metadata?.timestamp || '');
      const now = new Date();
      const diffMs = now.getTime() - timestampDate.getTime();
      expect(diffMs).toBeLessThan(60000); // Less than 1 minute
    });
  });
});
