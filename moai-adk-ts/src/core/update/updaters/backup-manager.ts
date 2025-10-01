// @CODE:UPD-BAK-001 | SPEC: update-orchestrator.ts 리팩토링
// Related: @CODE:UPD-001

/**
 * @file Backup management functionality
 * @author MoAI Team
 */

import { promises as fs } from 'node:fs';
import * as path from 'node:path';
import chalk from 'chalk';
import { logger } from '../../../utils/winston-logger.js';

/**
 * Backup manager for update operations
 * @tags @CODE:BACKUP-MANAGER-001
 */
export class BackupManager {
  private readonly projectPath: string;

  constructor(projectPath: string) {
    this.projectPath = projectPath;
  }

  /**
   * Create backup of existing files
   * @returns Backup directory path
   * @tags @CODE:CREATE-BACKUP-001:API
   */
  public async createBackup(): Promise<string> {
    logger.log(chalk.cyan('\n💾 백업 생성 중...'));

    const timestamp = new Date()
      .toISOString()
      .replace(/T/, '-')
      .replace(/\..+/, '')
      .replace(/:/g, '-');

    const backupDir = path.join(this.projectPath, '.moai-backup', timestamp);

    // Backup directories and files
    const itemsToBackup = ['.claude', '.moai', 'CLAUDE.md'];

    for (const item of itemsToBackup) {
      const sourcePath = path.join(this.projectPath, item);
      try {
        await fs.access(sourcePath);
        const targetPath = path.join(backupDir, item);

        if ((await fs.stat(sourcePath)).isDirectory()) {
          await this.copyDirectory(sourcePath, targetPath);
        } else {
          await fs.mkdir(path.dirname(targetPath), { recursive: true });
          await fs.copyFile(sourcePath, targetPath);
        }
      } catch {
        // File/directory doesn't exist, skip
      }
    }

    logger.log(chalk.green(`   → ${backupDir}`));
    return backupDir;
  }

  /**
   * Copy directory recursively
   * @param source - Source directory
   * @param target - Target directory
   * @tags UTIL:COPY-DIRECTORY-001
   */
  private async copyDirectory(source: string, target: string): Promise<void> {
    await fs.mkdir(target, { recursive: true });

    const entries = await fs.readdir(source, { withFileTypes: true });

    for (const entry of entries) {
      const sourcePath = path.join(source, entry.name);
      const targetPath = path.join(target, entry.name);

      if (entry.isDirectory()) {
        await this.copyDirectory(sourcePath, targetPath);
      } else {
        await fs.copyFile(sourcePath, targetPath);
      }
    }
  }
}
