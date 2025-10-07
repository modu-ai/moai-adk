// @TEST:INIT-003:DATA | SPEC: SPEC-INIT-003.md | CODE: src/core/installer/backup-metadata.ts
// Related: @CODE:INIT-003:BACKUP, @SPEC:INIT-003

/**
 * @file Test for Backup Metadata System
 * @author MoAI Team
 * @tags @TEST:INIT-003:DATA
 */

import { describe, test, expect, beforeEach, afterEach } from 'vitest';
import * as fs from 'node:fs';
import * as path from 'node:path';
import {
  type BackupMetadata,
  saveBackupMetadata,
  loadBackupMetadata,
  validateBackupMetadata,
} from '@/core/installer/backup-metadata';

const TEST_PROJECT_PATH = path.join(process.cwd(), '__test-project__');
const BACKUP_METADATA_PATH = path.join(
  TEST_PROJECT_PATH,
  '.moai',
  'backups',
  'latest.json'
);

describe('Backup Metadata System', () => {
  beforeEach(async () => {
    // Clean up test directory
    if (fs.existsSync(TEST_PROJECT_PATH)) {
      await fs.promises.rm(TEST_PROJECT_PATH, { recursive: true });
    }
    await fs.promises.mkdir(TEST_PROJECT_PATH, { recursive: true });
  });

  afterEach(async () => {
    // Clean up test directory
    if (fs.existsSync(TEST_PROJECT_PATH)) {
      await fs.promises.rm(TEST_PROJECT_PATH, { recursive: true });
    }
  });

  describe('BackupMetadata Interface', () => {
    test('should have correct structure', () => {
      // Given: Valid backup metadata
      const metadata: BackupMetadata = {
        timestamp: new Date().toISOString(),
        backup_path: '.moai-backup-20251006-143000',
        backed_up_files: ['.claude/', '.moai/', 'CLAUDE.md'],
        status: 'pending',
        created_by: 'moai init',
      };

      // Then: All required fields exist
      expect(metadata.timestamp).toBeDefined();
      expect(metadata.backup_path).toBeDefined();
      expect(metadata.backed_up_files).toBeInstanceOf(Array);
      expect(metadata.status).toBeDefined();
      expect(metadata.created_by).toBeDefined();
    });

    test('should support all status values', () => {
      // Given: Different status values
      const statuses: Array<'pending' | 'merged' | 'ignored'> = [
        'pending',
        'merged',
        'ignored',
      ];

      // When/Then: All statuses are valid
      for (const status of statuses) {
        const metadata: BackupMetadata = {
          timestamp: new Date().toISOString(),
          backup_path: '.moai-backup-test',
          backed_up_files: [],
          status,
          created_by: 'test',
        };
        expect(metadata.status).toBe(status);
      }
    });
  });

  describe('saveBackupMetadata', () => {
    test('should create .moai/backups/ directory if not exists', async () => {
      // Given: No backups directory
      const backupsDir = path.join(TEST_PROJECT_PATH, '.moai', 'backups');
      expect(fs.existsSync(backupsDir)).toBe(false);

      // When: Save metadata
      const metadata: BackupMetadata = {
        timestamp: new Date().toISOString(),
        backup_path: '.moai-backup-test',
        backed_up_files: [],
        status: 'pending',
        created_by: 'moai init',
      };
      await saveBackupMetadata(TEST_PROJECT_PATH, metadata);

      // Then: Directory created
      expect(fs.existsSync(backupsDir)).toBe(true);
    });

    test('should save metadata to latest.json', async () => {
      // Given: Valid metadata
      const metadata: BackupMetadata = {
        timestamp: '2025-10-06T14:30:00.000Z',
        backup_path: '.moai-backup-20251006-143000',
        backed_up_files: ['.claude/', '.moai/', 'CLAUDE.md'],
        status: 'pending',
        created_by: 'moai init',
      };

      // When: Save metadata
      await saveBackupMetadata(TEST_PROJECT_PATH, metadata);

      // Then: File exists and contains correct data
      expect(fs.existsSync(BACKUP_METADATA_PATH)).toBe(true);
      const saved = JSON.parse(
        await fs.promises.readFile(BACKUP_METADATA_PATH, 'utf-8')
      );
      expect(saved.timestamp).toBe(metadata.timestamp);
      expect(saved.backup_path).toBe(metadata.backup_path);
      expect(saved.backed_up_files).toEqual(metadata.backed_up_files);
      expect(saved.status).toBe('pending');
    });

    test('should overwrite existing metadata', async () => {
      // Given: Existing metadata
      const oldMetadata: BackupMetadata = {
        timestamp: '2025-10-06T10:00:00.000Z',
        backup_path: '.moai-backup-old',
        backed_up_files: [],
        status: 'merged',
        created_by: 'moai init',
      };
      await saveBackupMetadata(TEST_PROJECT_PATH, oldMetadata);

      // When: Save new metadata
      const newMetadata: BackupMetadata = {
        timestamp: '2025-10-06T14:30:00.000Z',
        backup_path: '.moai-backup-new',
        backed_up_files: ['.claude/'],
        status: 'pending',
        created_by: 'moai init',
      };
      await saveBackupMetadata(TEST_PROJECT_PATH, newMetadata);

      // Then: New metadata overwrites old
      const saved = JSON.parse(
        await fs.promises.readFile(BACKUP_METADATA_PATH, 'utf-8')
      );
      expect(saved.timestamp).toBe(newMetadata.timestamp);
      expect(saved.backup_path).toBe(newMetadata.backup_path);
    });

    test('should throw error if project path does not exist', async () => {
      // Given: Non-existent project path
      const invalidPath = path.join(process.cwd(), '__non-existent__');

      const metadata: BackupMetadata = {
        timestamp: new Date().toISOString(),
        backup_path: '.moai-backup-test',
        backed_up_files: [],
        status: 'pending',
        created_by: 'moai init',
      };

      // When/Then: Throws error
      await expect(saveBackupMetadata(invalidPath, metadata)).rejects.toThrow();
    });
  });

  describe('loadBackupMetadata', () => {
    test('should return null if metadata file does not exist', async () => {
      // Given: No metadata file
      expect(fs.existsSync(BACKUP_METADATA_PATH)).toBe(false);

      // When: Load metadata
      const result = await loadBackupMetadata(TEST_PROJECT_PATH);

      // Then: Returns null
      expect(result).toBeNull();
    });

    test('should load valid metadata', async () => {
      // Given: Saved metadata
      const metadata: BackupMetadata = {
        timestamp: '2025-10-06T14:30:00.000Z',
        backup_path: '.moai-backup-20251006-143000',
        backed_up_files: ['.claude/', '.moai/', 'CLAUDE.md'],
        status: 'pending',
        created_by: 'moai init',
      };
      await saveBackupMetadata(TEST_PROJECT_PATH, metadata);

      // When: Load metadata
      const loaded = await loadBackupMetadata(TEST_PROJECT_PATH);

      // Then: Loaded data matches saved data
      expect(loaded).not.toBeNull();
      expect(loaded?.timestamp).toBe(metadata.timestamp);
      expect(loaded?.backup_path).toBe(metadata.backup_path);
      expect(loaded?.backed_up_files).toEqual(metadata.backed_up_files);
      expect(loaded?.status).toBe('pending');
    });

    test('should return null if metadata is invalid', async () => {
      // Given: Invalid metadata (missing required fields)
      const invalidMetadata = {
        timestamp: '2025-10-06T14:30:00.000Z',
        // missing backup_path
        backed_up_files: [],
      };

      await fs.promises.mkdir(path.dirname(BACKUP_METADATA_PATH), {
        recursive: true,
      });
      await fs.promises.writeFile(
        BACKUP_METADATA_PATH,
        JSON.stringify(invalidMetadata)
      );

      // When: Load metadata
      const result = await loadBackupMetadata(TEST_PROJECT_PATH);

      // Then: Returns null due to validation failure
      expect(result).toBeNull();
    });
  });

  describe('validateBackupMetadata', () => {
    test('should return true for valid metadata', () => {
      // Given: Valid metadata
      const metadata = {
        timestamp: '2025-10-06T14:30:00.000Z',
        backup_path: '.moai-backup-20251006-143000',
        backed_up_files: ['.claude/', '.moai/'],
        status: 'pending',
        created_by: 'moai init',
      };

      // When: Validate
      const result = validateBackupMetadata(metadata);

      // Then: Validation passes
      expect(result).toBe(true);
    });

    test('should return false if timestamp is missing', () => {
      // Given: Missing timestamp
      const metadata = {
        backup_path: '.moai-backup-test',
        backed_up_files: [],
        status: 'pending',
        created_by: 'moai init',
      };

      // When: Validate
      const result = validateBackupMetadata(metadata);

      // Then: Validation fails
      expect(result).toBe(false);
    });

    test('should return false if backup_path is missing', () => {
      // Given: Missing backup_path
      const metadata = {
        timestamp: '2025-10-06T14:30:00.000Z',
        backed_up_files: [],
        status: 'pending',
        created_by: 'moai init',
      };

      // When: Validate
      const result = validateBackupMetadata(metadata);

      // Then: Validation fails
      expect(result).toBe(false);
    });

    test('should return false if backed_up_files is not an array', () => {
      // Given: Invalid backed_up_files
      const metadata = {
        timestamp: '2025-10-06T14:30:00.000Z',
        backup_path: '.moai-backup-test',
        backed_up_files: 'not-an-array',
        status: 'pending',
        created_by: 'moai init',
      };

      // When: Validate
      const result = validateBackupMetadata(metadata);

      // Then: Validation fails
      expect(result).toBe(false);
    });

    test('should return false if status is invalid', () => {
      // Given: Invalid status
      const metadata = {
        timestamp: '2025-10-06T14:30:00.000Z',
        backup_path: '.moai-backup-test',
        backed_up_files: [],
        status: 'invalid-status',
        created_by: 'moai init',
      };

      // When: Validate
      const result = validateBackupMetadata(metadata);

      // Then: Validation fails
      expect(result).toBe(false);
    });

    test('should return false if created_by is missing', () => {
      // Given: Missing created_by
      const metadata = {
        timestamp: '2025-10-06T14:30:00.000Z',
        backup_path: '.moai-backup-test',
        backed_up_files: [],
        status: 'pending',
      };

      // When: Validate
      const result = validateBackupMetadata(metadata);

      // Then: Validation fails
      expect(result).toBe(false);
    });

    test('should return false for null or undefined', () => {
      // When/Then: Invalid inputs
      expect(validateBackupMetadata(null)).toBe(false);
      expect(validateBackupMetadata(undefined)).toBe(false);
    });
  });
});
