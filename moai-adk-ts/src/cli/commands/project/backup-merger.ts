// @CODE:INIT-003:MERGE | SPEC: SPEC-INIT-003.md | TEST: __tests__/cli/commands/project/backup-merger.test.ts
// Related: @CODE:INIT-003:DATA, @SPEC:INIT-003

/**
 * @file Backup Merger - Intelligent merge of backup and current files
 * @author MoAI Team
 * @tags @CODE:INIT-003:MERGE
 */

import * as fs from 'node:fs/promises';
import * as path from 'node:path';
import { mergeJSON } from './merge-strategies/json-merger.js';
import { mergeMarkdown } from './merge-strategies/markdown-merger.js';
import { mergeHooks } from './merge-strategies/hooks-merger.js';
import {
  hasAnyMoAIFiles as hasAnyMoAIFilesSync,
  generateBackupDirName,
  getBackupTargets,
  copyDirectoryRecursive,
} from '../../../core/installer/backup-utils.js';

/**
 * Merge report structure
 */
export interface MergeReport {
  /** Successfully merged files */
  merged: string[];
  /** Files skipped (no backup or no merge needed) */
  skipped: string[];
  /** Files with errors */
  errors: Array<{ file: string; error: string }>;
  /** Timestamp of merge operation */
  timestamp: string;
}

/**
 * BackupMerger - Intelligently merges backup files with current files
 *
 * **Merge Strategies**:
 * - JSON files: Deep merge with backup priority
 * - Markdown files: HISTORY section accumulation
 * - Hook files (.cjs): Version comparison
 * - Other files: Current priority (template wins)
 *
 * **Usage**:
 * ```typescript
 * const merger = new BackupMerger('.moai-backup-123/', './');
 * await merger.mergeFile('config.json');
 * await merger.mergeDirectory('.claude');
 * const report = merger.getReport();
 * ```
 */
export class BackupMerger {
  private backupDir: string;
  private currentDir: string;
  private report: MergeReport;

  /**
   * Create a new BackupMerger
   *
   * @param backupDir - Path to backup directory
   * @param currentDir - Path to current directory (usually project root)
   */
  constructor(backupDir: string, currentDir: string) {
    this.backupDir = backupDir;
    this.currentDir = currentDir;
    this.report = {
      merged: [],
      skipped: [],
      errors: [],
      timestamp: new Date().toISOString(),
    };
  }

  /**
   * Merge a single file
   *
   * Automatically detects file type and applies appropriate merge strategy.
   *
   * @param relativePath - Relative path to file (e.g., 'config.json' or '.claude/settings.json')
   */
  async mergeFile(relativePath: string): Promise<void> {
    const backupPath = path.join(this.backupDir, relativePath);
    const currentPath = path.join(this.currentDir, relativePath);

    try {
      // Check if backup file exists
      const backupExists = await fileExists(backupPath);
      if (!backupExists) {
        this.report.skipped.push(relativePath);
        return;
      }

      // Check if current file exists
      const currentExists = await fileExists(currentPath);
      if (!currentExists) {
        // Copy backup to current
        await fs.mkdir(path.dirname(currentPath), { recursive: true });
        await fs.copyFile(backupPath, currentPath);
        this.report.merged.push(relativePath);
        return;
      }

      // Read both files
      const backupContent = await fs.readFile(backupPath, 'utf-8');
      const currentContent = await fs.readFile(currentPath, 'utf-8');

      // Determine merge strategy based on file extension
      const merged = this.mergeContent(relativePath, backupContent, currentContent);

      // Write merged content
      await fs.writeFile(currentPath, merged, 'utf-8');
      this.report.merged.push(relativePath);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      this.report.errors.push({
        file: relativePath,
        error: errorMessage,
      });
    }
  }

  /**
   * Merge all files in a directory recursively
   *
   * @param relativePath - Relative path to directory (e.g., '.claude' or '.moai')
   */
  async mergeDirectory(relativePath: string): Promise<void> {
    const backupPath = path.join(this.backupDir, relativePath);

    try {
      // Check if backup directory exists
      const backupExists = await fileExists(backupPath);
      if (!backupExists) {
        return;
      }

      // Read directory contents
      const entries = await fs.readdir(backupPath, { withFileTypes: true });

      for (const entry of entries) {
        const entryRelativePath = path.join(relativePath, entry.name);

        if (entry.isDirectory()) {
          // Recursively merge subdirectories
          await this.mergeDirectory(entryRelativePath);
        } else {
          // Merge file
          await this.mergeFile(entryRelativePath);
        }
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      this.report.errors.push({
        file: relativePath,
        error: errorMessage,
      });
    }
  }

  /**
   * Get merge report
   *
   * @returns Report of merged, skipped, and error files
   */
  getReport(): MergeReport {
    return { ...this.report };
  }

  /**
   * Merge content based on file type
   *
   * @param filePath - File path (for extension detection)
   * @param backup - Backup content
   * @param current - Current content
   * @returns Merged content
   * @internal
   */
  private mergeContent(filePath: string, backup: string, current: string): string {
    const ext = path.extname(filePath).toLowerCase();

    // JSON files: Deep merge
    if (ext === '.json') {
      try {
        const backupObj = JSON.parse(backup);
        const currentObj = JSON.parse(current);
        const merged = mergeJSON(backupObj, currentObj);
        return JSON.stringify(merged, null, 2);
      } catch {
        // If JSON parsing fails, use current
        return current;
      }
    }

    // Markdown files: HISTORY accumulation
    if (ext === '.md') {
      return mergeMarkdown(backup, current);
    }

    // Hook files: Version comparison
    if (ext === '.cjs' || ext === '.mjs' || filePath.includes('hooks/')) {
      return mergeHooks(backup, current);
    }

    // Other files: Current priority (template wins)
    return current;
  }
}

/**
 * Check if file or directory exists
 * @internal
 */
async function fileExists(filePath: string): Promise<boolean> {
  try {
    await fs.access(filePath);
    return true;
  } catch {
    return false;
  }
}

/**
 * Create emergency backup when metadata is missing but MoAI-ADK files exist
 * @CODE:INIT-003:MERGE - v0.2.1 Emergency backup (Refactored with backup-utils)
 *
 * @param projectPath - Project root path
 * @returns Backup metadata if backup was created, null otherwise
 */
export async function createEmergencyBackup(
  projectPath: string
): Promise<BackupMetadata | null> {
  // Check if ANY MoAI-ADK files exist (OR condition) - using backup-utils
  if (!hasAnyMoAIFilesSync(projectPath)) {
    return null;
  }

  // Generate backup directory name - using backup-utils
  const backupDirName = generateBackupDirName();
  const backupPath = path.join(projectPath, backupDirName);

  // Create backup directory
  await fs.mkdir(backupPath, { recursive: true });

  // Get selective backup targets - using backup-utils
  const targets = getBackupTargets(projectPath);
  const backedUpFiles: string[] = [];

  // Backup files and directories
  for (const target of targets) {
    const srcPath = path.join(projectPath, target);
    const dstPath = path.join(backupPath, target);

    // Check if directory or file (async version)
    const stats = await fs.stat(srcPath);
    if (stats.isDirectory()) {
      await copyDirectoryRecursive(srcPath, dstPath);
      backedUpFiles.push(`${target}/`);
    } else {
      await fs.copyFile(srcPath, dstPath);
      backedUpFiles.push(target);
    }
  }

  // Create backup metadata
  const metadata: BackupMetadata = {
    timestamp: new Date().toISOString(),
    backup_path: backupDirName,
    backed_up_files: backedUpFiles,
    status: 'pending',
    created_by: '/alfred:8-project (emergency backup)',
  };

  // Save metadata
  const metadataDir = path.join(projectPath, '.moai', 'backups');
  await fs.mkdir(metadataDir, { recursive: true });
  await fs.writeFile(
    path.join(metadataDir, 'latest.json'),
    JSON.stringify(metadata, null, 2),
    'utf-8'
  );

  return metadata;
}

/**
 * Backup metadata structure (aligned with Phase A)
 */
export interface BackupMetadata {
  timestamp: string;
  backup_path: string;
  backed_up_files: string[];
  status: 'pending' | 'merged' | 'ignored';
  created_by: string;
}
