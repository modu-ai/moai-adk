/**
 * @file Backup and rollback management for update operations
 * @author MoAI Team
 * @tags @FEATURE:UPDATE-STRATEGY-001 | Chain: @REQ:UPDATE-REAL-001 -> @DESIGN:BACKUP-MANAGER-001 -> @TASK:BACKUP-MANAGER-001 -> @TEST:BACKUP-MANAGER-001
 * Related: @SEC:UPDATE-STRATEGY-001, @DOCS:UPDATE-STRATEGY-001
 */

import { promises as fs } from 'node:fs';
import * as path from 'node:path';

/**
 * Backup metadata for tracking snapshots
 * @tags @DESIGN:BACKUP-METADATA-001
 */
export interface BackupMetadata {
  readonly id: string;
  readonly timestamp: string;
  readonly projectPath: string;
  readonly backupPath: string;
  readonly files: readonly BackupFileInfo[];
  readonly totalSize: number;
  readonly reason: string;
  readonly version: {
    readonly before: string;
    readonly after: string;
  };
}

/**
 * Individual file backup information
 * @tags @DESIGN:BACKUP-FILE-INFO-001
 */
export interface BackupFileInfo {
  readonly originalPath: string;
  readonly backupPath: string;
  readonly size: number;
  readonly hash: string;
  readonly lastModified: string;
}

/**
 * Backup operation result
 * @tags @DESIGN:BACKUP-RESULT-001
 */
export interface BackupResult {
  readonly success: boolean;
  readonly backupId: string;
  readonly backupPath: string;
  readonly filesBackedUp: number;
  readonly totalSize: number;
  readonly duration: number;
  readonly error?: string;
}

/**
 * Rollback operation result
 * @tags @DESIGN:ROLLBACK-RESULT-001
 */
export interface RollbackResult {
  readonly success: boolean;
  readonly backupId: string;
  readonly filesRestored: number;
  readonly totalSize: number;
  readonly duration: number;
  readonly error?: string;
}

/**
 * Backup manager for creating snapshots and handling rollbacks
 * @tags @FEATURE:BACKUP-MANAGER-001
 */
export class BackupManager {
  private readonly backupRoot: string;

  constructor(projectPath: string) {
    this.backupRoot = path.join(projectPath, '.moai', 'backups');
  }

  /**
   * Create backup snapshot of specified files
   * @param filePaths - List of files to backup (relative paths)
   * @param projectPath - Project root directory
   * @param reason - Reason for backup (e.g., "pre-update")
   * @param versionInfo - Version information
   * @returns Backup operation result
   * @tags @API:CREATE-BACKUP-001
   */
  public async createBackup(
    filePaths: readonly string[],
    projectPath: string,
    reason: string,
    versionInfo: { before: string; after: string }
  ): Promise<BackupResult> {
    const startTime = Date.now();
    const backupId = this.generateBackupId();
    const backupPath = path.join(this.backupRoot, backupId);

    try {
      // Ensure backup directory exists
      await fs.mkdir(backupPath, { recursive: true });

      const backupFiles: BackupFileInfo[] = [];
      let totalSize = 0;

      // Backup each file
      for (const filePath of filePaths) {
        const fullPath = path.join(projectPath, filePath);

        // Check if file exists before backup
        try {
          await fs.access(fullPath);
        } catch {
          continue; // Skip non-existent files
        }

        const fileInfo = await this.backupSingleFile(
          fullPath,
          filePath,
          backupPath
        );

        if (fileInfo) {
          backupFiles.push(fileInfo);
          totalSize += fileInfo.size;
        }
      }

      // Create backup metadata
      const metadata: BackupMetadata = {
        id: backupId,
        timestamp: new Date().toISOString(),
        projectPath,
        backupPath,
        files: backupFiles,
        totalSize,
        reason,
        version: versionInfo,
      };

      // Save metadata
      await this.saveBackupMetadata(metadata);

      const duration = Date.now() - startTime;

      return {
        success: true,
        backupId,
        backupPath,
        filesBackedUp: backupFiles.length,
        totalSize,
        duration,
      };
    } catch (error) {
      const duration = Date.now() - startTime;
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';

      // Clean up partial backup on failure
      try {
        await fs.rm(backupPath, { recursive: true, force: true });
      } catch {
        // Ignore cleanup errors
      }

      return {
        success: false,
        backupId,
        backupPath,
        filesBackedUp: 0,
        totalSize: 0,
        duration,
        error: errorMessage,
      };
    }
  }

  /**
   * Restore files from backup
   * @param backupId - Backup ID to restore from
   * @param projectPath - Project root directory
   * @param selectiveFiles - Optional list of specific files to restore
   * @returns Rollback operation result
   * @tags @API:ROLLBACK-BACKUP-001
   */
  public async rollback(
    backupId: string,
    projectPath: string,
    selectiveFiles?: readonly string[]
  ): Promise<RollbackResult> {
    const startTime = Date.now();

    try {
      // Load backup metadata
      const metadata = await this.loadBackupMetadata(backupId);

      if (!metadata) {
        throw new Error(`Backup ${backupId} not found`);
      }

      let filesToRestore = metadata.files;

      // Filter to selective files if specified
      if (selectiveFiles && selectiveFiles.length > 0) {
        filesToRestore = metadata.files.filter(file =>
          selectiveFiles.includes(file.originalPath)
        );
      }

      let filesRestored = 0;
      let totalSize = 0;

      // Restore each file
      for (const fileInfo of filesToRestore) {
        const success = await this.restoreSingleFile(fileInfo, projectPath);

        if (success) {
          filesRestored++;
          totalSize += fileInfo.size;
        }
      }

      const duration = Date.now() - startTime;

      return {
        success: true,
        backupId,
        filesRestored,
        totalSize,
        duration,
      };
    } catch (error) {
      const duration = Date.now() - startTime;
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';

      return {
        success: false,
        backupId,
        filesRestored: 0,
        totalSize: 0,
        duration,
        error: errorMessage,
      };
    }
  }

  /**
   * List available backups with metadata
   * @returns List of backup metadata
   * @tags @API:LIST-BACKUPS-001
   */
  public async listBackups(): Promise<readonly BackupMetadata[]> {
    try {
      await fs.access(this.backupRoot);
    } catch {
      return []; // No backups directory
    }

    try {
      const entries = await fs.readdir(this.backupRoot, {
        withFileTypes: true,
      });
      const backups: BackupMetadata[] = [];

      for (const entry of entries) {
        if (entry.isDirectory()) {
          const metadata = await this.loadBackupMetadata(entry.name);
          if (metadata) {
            backups.push(metadata);
          }
        }
      }

      // Sort by timestamp (newest first)
      return backups.sort(
        (a, b) =>
          new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
      );
    } catch {
      return [];
    }
  }

  /**
   * Delete old backups to manage disk space
   * @param keepCount - Number of recent backups to keep
   * @returns Number of backups deleted
   * @tags @API:CLEANUP-BACKUPS-001
   */
  public async cleanupOldBackups(keepCount: number = 5): Promise<number> {
    const backups = await this.listBackups();

    if (backups.length <= keepCount) {
      return 0;
    }

    const toDelete = backups.slice(keepCount);
    let deletedCount = 0;

    for (const backup of toDelete) {
      try {
        await fs.rm(backup.backupPath, { recursive: true, force: true });
        deletedCount++;
      } catch {
        // Continue even if deletion fails
      }
    }

    return deletedCount;
  }

  /**
   * Generate unique backup ID
   * @returns Backup ID string
   * @tags @UTIL:GENERATE-BACKUP-ID-001
   */
  private generateBackupId(): string {
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const random = Math.random().toString(36).substring(2, 8);
    return `backup-${timestamp}-${random}`;
  }

  /**
   * Backup single file with metadata
   * @param sourcePath - Source file path
   * @param relativePath - Relative path for storage
   * @param backupDir - Backup directory
   * @returns File backup information
   * @tags @UTIL:BACKUP-SINGLE-FILE-001
   */
  private async backupSingleFile(
    sourcePath: string,
    relativePath: string,
    backupDir: string
  ): Promise<BackupFileInfo | null> {
    try {
      const stat = await fs.stat(sourcePath);
      const targetPath = path.join(backupDir, relativePath);

      // Ensure target directory exists
      await fs.mkdir(path.dirname(targetPath), { recursive: true });

      // Copy file
      await fs.copyFile(sourcePath, targetPath);

      // Generate simple hash (for integrity checking)
      const content = await fs.readFile(sourcePath);
      const hash = this.generateHash(content);

      return {
        originalPath: relativePath,
        backupPath: targetPath,
        size: stat.size,
        hash,
        lastModified: stat.mtime.toISOString(),
      };
    } catch {
      return null;
    }
  }

  /**
   * Restore single file from backup
   * @param fileInfo - File backup information
   * @param projectPath - Project root directory
   * @returns True if restoration successful
   * @tags @UTIL:RESTORE-SINGLE-FILE-001
   */
  private async restoreSingleFile(
    fileInfo: BackupFileInfo,
    projectPath: string
  ): Promise<boolean> {
    try {
      const targetPath = path.join(projectPath, fileInfo.originalPath);

      // Ensure target directory exists
      await fs.mkdir(path.dirname(targetPath), { recursive: true });

      // Copy file back
      await fs.copyFile(fileInfo.backupPath, targetPath);

      return true;
    } catch {
      return false;
    }
  }

  /**
   * Save backup metadata to disk
   * @param metadata - Backup metadata
   * @tags @UTIL:SAVE-METADATA-001
   */
  private async saveBackupMetadata(metadata: BackupMetadata): Promise<void> {
    const metadataPath = path.join(metadata.backupPath, 'metadata.json');
    await fs.writeFile(metadataPath, JSON.stringify(metadata, null, 2));
  }

  /**
   * Load backup metadata from disk
   * @param backupId - Backup ID
   * @returns Backup metadata or null if not found
   * @tags @UTIL:LOAD-METADATA-001
   */
  private async loadBackupMetadata(
    backupId: string
  ): Promise<BackupMetadata | null> {
    try {
      const metadataPath = path.join(
        this.backupRoot,
        backupId,
        'metadata.json'
      );
      const content = await fs.readFile(metadataPath, 'utf-8');
      return JSON.parse(content) as BackupMetadata;
    } catch {
      return null;
    }
  }

  /**
   * Generate simple hash for content integrity
   * @param content - File content buffer
   * @returns Hash string
   * @tags @UTIL:GENERATE-HASH-001
   */
  private generateHash(content: Buffer): string {
    // Simple hash implementation (production would use crypto)
    let hash = 0;
    const str = content.toString('base64');

    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i);
      hash = (hash << 5) - hash + char;
      hash = hash & hash; // Convert to 32-bit integer
    }

    return Math.abs(hash).toString(16);
  }
}
