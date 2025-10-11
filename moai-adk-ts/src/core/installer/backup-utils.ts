// @CODE:INIT-003:DATA | SPEC: SPEC-INIT-003.md
// Related: @CODE:INIT-003:BACKUP, @CODE:INIT-003:MERGE

/**
 * @file Backup Utility Functions (v0.2.1)
 * @author MoAI Team
 * @tags @CODE:INIT-003:DATA
 *
 * Common utilities for backup creation across Phase A (moai init)
 * and Phase B (/alfred:0-project emergency backup).
 */

import * as fs from 'node:fs';
import * as path from 'node:path';

/**
 * Check if any MoAI-ADK files exist in a directory (OR condition)
 * @CODE:INIT-003:DATA - v0.2.1 selective backup check
 *
 * @param projectPath - Project directory path
 * @returns True if at least one MoAI-ADK file exists
 */
export function hasAnyMoAIFiles(projectPath: string): boolean {
  const hasClaudeDir = fs.existsSync(path.join(projectPath, '.claude'));
  const hasMoaiDir = fs.existsSync(path.join(projectPath, '.moai'));
  const hasClaudeMd = fs.existsSync(path.join(projectPath, 'CLAUDE.md'));

  return hasClaudeDir || hasMoaiDir || hasClaudeMd;
}

/**
 * Generate backup directory name with timestamp
 * @CODE:INIT-003:DATA
 *
 * @returns Backup directory name (e.g., `.moai-backup-2025-10-07T05-04-03`)
 */
export function generateBackupDirName(): string {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5);
  return `.moai-backup-${timestamp}`;
}

/**
 * Get list of files that need to be backed up
 * @CODE:INIT-003:DATA - v0.2.1 selective backup
 *
 * @param projectPath - Project directory path
 * @returns Array of file/directory paths that should be backed up
 */
export function getBackupTargets(projectPath: string): string[] {
  const targets: string[] = [];

  // Check directories
  const criticalDirs = ['.claude', '.moai'];
  for (const dir of criticalDirs) {
    if (fs.existsSync(path.join(projectPath, dir))) {
      targets.push(dir);
    }
  }

  // Check files
  const criticalFiles = ['CLAUDE.md'];
  for (const file of criticalFiles) {
    if (fs.existsSync(path.join(projectPath, file))) {
      targets.push(file);
    }
  }

  return targets;
}

/**
 * Copy directory recursively (sync version for phase-executor)
 * @CODE:INIT-003:DATA
 *
 * @param src - Source directory
 * @param dest - Destination directory
 */
export async function copyDirectoryRecursive(
  src: string,
  dest: string
): Promise<void> {
  await fs.promises.mkdir(dest, { recursive: true });
  const entries = await fs.promises.readdir(src, { withFileTypes: true });

  for (const entry of entries) {
    const srcPath = path.join(src, entry.name);
    const destPath = path.join(dest, entry.name);

    if (entry.isDirectory()) {
      await copyDirectoryRecursive(srcPath, destPath);
    } else {
      await fs.promises.copyFile(srcPath, destPath);
    }
  }
}

/**
 * Validate backup metadata structure
 * @CODE:INIT-003:DATA
 *
 * @param metadata - Backup metadata object
 * @returns True if metadata is valid
 */
export function isValidBackupMetadata(metadata: unknown): boolean {
  if (!metadata || typeof metadata !== 'object') {
    return false;
  }

  const meta = metadata as Record<string, unknown>;

  return (
    typeof meta.timestamp === 'string' &&
    typeof meta.backup_path === 'string' &&
    Array.isArray(meta.backed_up_files) &&
    typeof meta.status === 'string' &&
    ['pending', 'merged', 'ignored'].includes(meta.status as string) &&
    typeof meta.created_by === 'string'
  );
}
