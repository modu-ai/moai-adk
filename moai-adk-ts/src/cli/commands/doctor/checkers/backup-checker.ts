// @CODE:BACKUP-CHECKER-001 | Chain: @REQ:DOCTOR-001 -> @DESIGN:DOCTOR-001 -> @TASK:DOCTOR-001
// Related: @TEST:BACKUP-CHECKER-001

/**
 * @file Backup directory checker and analyzer
 * @author MoAI Team
 */

import * as fs from 'node:fs/promises';
import * as os from 'node:os';
import * as path from 'node:path';

/**
 * BackupChecker handles finding and analyzing MoAI-ADK backup directories
 * @tags @CODE:BACKUP-CHECKER-001
 */
export class BackupChecker {
  /**
   * Find backup directories in specified or default search paths
   * @param searchPaths - Optional custom search paths
   * @returns Array of backup directory paths
   * @tags @CODE:FIND-BACKUPS-001:API
   */
  public async findBackupDirectories(
    searchPaths?: string[]
  ): Promise<string[]> {
    const backupPaths: string[] = [];
    const paths = searchPaths || this.getDefaultSearchPaths();

    for (const searchPath of paths) {
      try {
        const exists = await this.directoryExists(searchPath);
        if (exists) {
          const subdirs = await this.getBackupSubdirectories(searchPath);
          backupPaths.push(
            ...subdirs.map(subdir => path.join(searchPath, subdir))
          );
        }
      } catch {
        // Directory doesn't exist or can't be accessed
      }
    }

    return backupPaths.sort();
  }

  /**
   * Get backup directory contents summary
   * @param backupPath - Path to backup directory
   * @returns Array of content descriptions
   * @tags @CODE:GET-BACKUP-CONTENTS-001:API
   */
  public async getBackupContents(backupPath: string): Promise<string[]> {
    const contents: string[] = [];

    try {
      const entries = await fs.readdir(backupPath);

      if (entries.includes('.claude')) contents.push('Claude Code config');
      if (entries.includes('.moai')) contents.push('MoAI config');
      if (entries.includes('package.json')) contents.push('Package config');
      if (entries.includes('tsconfig.json')) contents.push('TypeScript config');
      if (entries.some(e => e.endsWith('.py'))) contents.push('Python files');
      if (entries.some(e => e.endsWith('.ts')))
        contents.push('TypeScript files');

      const totalFiles = entries.filter(e => !e.startsWith('.')).length;
      if (totalFiles > 0) {
        contents.push(`${totalFiles} files`);
      }
    } catch {
      // Can't read contents
    }

    return contents;
  }

  /**
   * Get default backup search paths
   * @returns Array of default search paths
   * @tags @UTIL:DEFAULT-PATHS-001
   */
  private getDefaultSearchPaths(): string[] {
    return [
      path.join(process.cwd(), '.moai-backup'),
      path.join(os.homedir(), '.moai', 'backups'),
    ];
  }

  /**
   * Check if directory exists
   * @param dirPath - Directory path to check
   * @returns True if directory exists
   * @tags @UTIL:DIR-EXISTS-001
   */
  private async directoryExists(dirPath: string): Promise<boolean> {
    try {
      const stat = await fs.stat(dirPath);
      return stat.isDirectory();
    } catch {
      return false;
    }
  }

  /**
   * Get backup subdirectories matching naming patterns
   * @param dirPath - Directory path to search
   * @returns Array of backup subdirectory names
   * @tags @UTIL:GET-BACKUP-SUBDIRS-001
   */
  private async getBackupSubdirectories(dirPath: string): Promise<string[]> {
    try {
      const entries = await fs.readdir(dirPath, { withFileTypes: true });
      return entries
        .filter(entry => entry.isDirectory())
        .map(entry => entry.name)
        .filter(
          name => name.startsWith('backup-') || /^\d{4}-\d{2}-\d{2}/.test(name)
        );
    } catch {
      return [];
    }
  }
}
