// @CODE:INIT-003:DATA | SPEC: SPEC-INIT-003.md | TEST: __tests__/core/installer/backup-metadata.test.ts
// Related: @CODE:INIT-003:BACKUP, @SPEC:INIT-003

/**
 * @file Backup Metadata System for MoAI-ADK Init
 * @author MoAI Team
 * @tags @CODE:INIT-003:DATA
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { logger } from '@/utils/winston-logger';

/**
 * Backup metadata structure
 * Stored in .moai/backups/latest.json
 */
export interface BackupMetadata {
  /** ISO 8601 timestamp when backup was created */
  timestamp: string;
  /** Backup directory path relative to project root */
  backup_path: string;
  /** List of backed up files/directories */
  backed_up_files: string[];
  /** Backup status for /alfred:0-project workflow */
  status: 'pending' | 'merged' | 'ignored';
  /** Source of backup creation */
  created_by: string;
}

/**
 * Save backup metadata to .moai/backups/latest.json
 *
 * @param projectPath - Project root directory
 * @param metadata - Backup metadata to save
 * @throws Error if project path does not exist or file system error occurs
 */
export async function saveBackupMetadata(
  projectPath: string,
  metadata: BackupMetadata
): Promise<void> {
  // Validate project path exists
  if (!fs.existsSync(projectPath)) {
    throw new Error(`Project path does not exist: ${projectPath}`);
  }

  // Create .moai/backups directory if needed
  const backupsDir = path.join(projectPath, '.moai', 'backups');
  await fs.promises.mkdir(backupsDir, { recursive: true });

  // Save metadata
  const metadataPath = path.join(backupsDir, 'latest.json');
  await fs.promises.writeFile(
    metadataPath,
    JSON.stringify(metadata, null, 2),
    'utf-8'
  );

  logger.debug('Backup metadata saved', {
    metadataPath,
    status: metadata.status,
    tag: 'SUCCESS:BACKUP-METADATA-SAVE-001',
  });
}

/**
 * Load backup metadata from .moai/backups/latest.json
 *
 * @param projectPath - Project root directory
 * @returns Backup metadata if exists and valid, null otherwise
 */
export async function loadBackupMetadata(
  projectPath: string
): Promise<BackupMetadata | null> {
  const metadataPath = path.join(
    projectPath,
    '.moai',
    'backups',
    'latest.json'
  );

  // Check if metadata file exists
  if (!fs.existsSync(metadataPath)) {
    logger.debug('No backup metadata found', {
      metadataPath,
      tag: 'INFO:BACKUP-METADATA-NOT-FOUND-001',
    });
    return null;
  }

  try {
    // Load and parse metadata
    const content = await fs.promises.readFile(metadataPath, 'utf-8');
    const data = JSON.parse(content);

    // Validate metadata structure
    if (!validateBackupMetadata(data)) {
      logger.warn('Invalid backup metadata structure', {
        metadataPath,
        tag: 'WARN:BACKUP-METADATA-INVALID-001',
      });
      return null;
    }

    logger.debug('Backup metadata loaded', {
      metadataPath,
      status: data.status,
      tag: 'SUCCESS:BACKUP-METADATA-LOAD-001',
    });

    return data;
  } catch (error) {
    logger.error('Failed to load backup metadata', {
      error,
      metadataPath,
      tag: 'ERROR:BACKUP-METADATA-LOAD-001',
    });
    return null;
  }
}

/**
 * Validate backup metadata structure
 *
 * @param data - Data to validate
 * @returns True if data is valid BackupMetadata
 */
export function validateBackupMetadata(data: unknown): data is BackupMetadata {
  // Check if data is an object
  if (!data || typeof data !== 'object') {
    return false;
  }

  const obj = data as Record<string, unknown>;

  // Required fields check
  if (
    typeof obj.timestamp !== 'string' ||
    typeof obj.backup_path !== 'string' ||
    !Array.isArray(obj.backed_up_files) ||
    typeof obj.status !== 'string' ||
    typeof obj.created_by !== 'string'
  ) {
    return false;
  }

  // Validate status enum
  const validStatuses = ['pending', 'merged', 'ignored'];
  if (!validStatuses.includes(obj.status)) {
    return false;
  }

  return true;
}
