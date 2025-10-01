// @CODE:CLI-005 |
// Related: @CODE:CFG-001:API, @CODE:CFG-BACKUP-001

/**
 * @file CLI restore command for configuration recovery
 * @author MoAI Team
 */

import * as path from 'node:path';
import chalk from 'chalk';
import * as fs from 'fs-extra';
import { logger } from '../../utils/winston-logger.js';

/**
 * Backup validation result
 * @tags @SPEC:BACKUP-VALIDATION-001
 */
export interface BackupValidationResult {
  readonly isValid: boolean;
  readonly error?: string;
  readonly warning?: string;
  readonly missingItems: string[];
}

/**
 * Restore operation options
 * @tags @SPEC:RESTORE-OPTIONS-001
 */
export interface RestoreOptions {
  readonly dryRun: boolean;
  readonly force?: boolean | undefined;
}

/**
 * Restore operation result
 * @tags @SPEC:RESTORE-RESULT-001
 */
export interface RestoreResult {
  readonly success: boolean;
  readonly isDryRun: boolean;
  readonly restoredItems: string[];
  readonly skippedItems?: string[];
  readonly error?: string;
}

/**
 * Restore command for backup restoration
 * @tags @CODE:CLI-RESTORE-001
 */
export class RestoreCommand {
  private readonly requiredItems = ['.moai', '.claude', 'CLAUDE.md'];

  /**
   * Validate backup path and contents
   * @param backupPath - Path to backup directory
   * @returns Validation result
   * @tags @CODE:VALIDATE-BACKUP-001:API
   */
  public async validateBackupPath(
    backupPath: string
  ): Promise<BackupValidationResult> {
    try {
      // Check if backup path exists
      const exists = await fs.pathExists(backupPath);
      if (!exists) {
        return {
          isValid: false,
          error: 'Backup path does not exist',
          missingItems: [],
        };
      }

      // Check if backup path is a directory
      const stats = await fs.stat(backupPath);
      if (!stats.isDirectory()) {
        return {
          isValid: false,
          error: 'Backup path must be a directory',
          missingItems: [],
        };
      }

      // Check for required items in backup
      const missingItems: string[] = [];
      for (const item of this.requiredItems) {
        const itemPath = path.join(backupPath, item);
        const itemExists = await fs.pathExists(itemPath);
        if (!itemExists) {
          missingItems.push(item);
        }
      }

      // Return validation result
      if (missingItems.length > 0) {
        return {
          isValid: true,
          warning: `Backup may be incomplete. Missing: ${missingItems.join(', ')}`,
          missingItems,
        };
      }

      return {
        isValid: true,
        missingItems: [],
      };
    } catch (error) {
      return {
        isValid: false,
        error:
          error instanceof Error
            ? error.message
            : 'Unknown error during validation',
        missingItems: [],
      };
    }
  }

  /**
   * Perform restore operation
   * @param backupPath - Path to backup directory
   * @param options - Restore options
   * @returns Restore result
   * @tags @CODE:PERFORM-RESTORE-001:API
   */
  public async performRestore(
    backupPath: string,
    options: RestoreOptions
  ): Promise<RestoreResult> {
    try {
      const currentDir = process.cwd();
      const restoredItems: string[] = [];
      const skippedItems: string[] = [];

      // Dry run - just report what would be done
      if (options.dryRun) {
        for (const item of this.requiredItems) {
          const sourcePath = path.join(backupPath, item);
          const exists = await fs.pathExists(sourcePath);
          if (exists) {
            restoredItems.push(item);
          }
        }

        return {
          success: true,
          isDryRun: true,
          restoredItems,
          skippedItems,
        };
      }

      // Actual restore operation
      for (const item of this.requiredItems) {
        const sourcePath = path.join(backupPath, item);
        const targetPath = path.join(currentDir, item);

        const sourceExists = await fs.pathExists(sourcePath);
        if (!sourceExists) {
          continue; // Skip missing items
        }

        const targetExists = await fs.pathExists(targetPath);

        // Skip existing files unless force is enabled
        if (targetExists && !options.force) {
          skippedItems.push(item);
          continue;
        }

        // Remove existing target if it exists
        if (targetExists) {
          await fs.remove(targetPath);
        }

        // Copy from backup to current directory
        await fs.copy(sourcePath, targetPath);
        restoredItems.push(item);
      }

      return {
        success: true,
        isDryRun: false,
        restoredItems,
        skippedItems,
      };
    } catch (error) {
      return {
        success: false,
        isDryRun: options.dryRun,
        restoredItems: [],
        error:
          error instanceof Error
            ? error.message
            : 'Unknown error during restore',
      };
    }
  }

  /**
   * Run restore command
   * @param backupPath - Path to backup directory
   * @param options - Restore options
   * @returns Restore result
   * @tags @CODE:RESTORE-RUN-001:API
   */
  public async run(
    backupPath: string,
    options: RestoreOptions
  ): Promise<RestoreResult> {
    // Step 1: Validate backup path
    const validation = await this.validateBackupPath(backupPath);

    if (!validation.isValid) {
      logger.info(chalk.red(`❌ ${validation.error}`));
      return {
        success: false,
        isDryRun: options.dryRun,
        restoredItems: [],
        error: validation.error || 'Unknown validation error',
      };
    }

    // Step 2: Show warning if backup is incomplete
    if (validation.warning) {
      logger.info(chalk.yellow(`⚠️  Warning: ${validation.warning}`));
    }

    // Step 3: Perform restore operation
    const currentDir = process.cwd();

    if (options.dryRun) {
      logger.info(chalk.cyan(`🔍 Dry run - would restore to: ${currentDir}`));

      // Show what would be restored
      for (const item of this.requiredItems) {
        const sourcePath = path.join(backupPath, item);
        const targetPath = path.join(currentDir, item);
        const exists = await fs.pathExists(sourcePath);

        if (exists) {
          logger.info(`  Would restore: ${sourcePath} → ${targetPath}`);
        }
      }
    } else {
      logger.info(chalk.cyan(`🔄 Restoring backup to: ${currentDir}`));
    }

    // Step 4: Execute restore
    const result = await this.performRestore(backupPath, options);

    // Step 5: Display results
    if (result.success) {
      if (options.dryRun) {
        logger.info(chalk.green(`✅ Dry run completed successfully`));
        logger.info(`  Would restore ${result.restoredItems.length} items`);
      } else {
        logger.info(chalk.green(`✅ Backup restored successfully`));

        // Show restored items
        for (const item of result.restoredItems) {
          logger.info(`  Restored: ${item}`);
        }

        // Show skipped items
        if (result.skippedItems && result.skippedItems.length > 0) {
          logger.info(
            chalk.yellow(
              `  Skipped ${result.skippedItems.length} existing items (use --force to overwrite)`
            )
          );
        }
      }
    } else {
      logger.info(chalk.red(`❌ Failed to restore backup: ${result.error}`));
    }

    return result;
  }
}
