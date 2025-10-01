// @CODE:INST-003 |
// Related: @CODE:INST-003:API

/**
 * @file Installation phase validation
 * @author MoAI Team
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { logger } from '@/utils/winston-logger';
import type { InstallationConfig } from './types';

/**
 * Validates installation phases before execution
 *
 * Responsibilities:
 * - System requirements validation
 * - File system permissions check
 * - Installation completeness verification
 * - Node.js version validation
 *
 * @tags @CODE:PHASE-VALIDATOR-001
 */
export class PhaseValidator {
  /**
   * Validate system requirements
   * @param config Installation configuration
   * @throws Error if requirements not met
   * @tags @CODE:VALIDATE-SYSTEM-001:API
   */
  async validateSystemRequirements(config: InstallationConfig): Promise<void> {
    try {
      // Check Node.js version
      const nodeVersion = process.version;
      const requiredVersion = '18.0.0';
      if (!this.isVersionSatisfied(nodeVersion.slice(1), requiredVersion)) {
        throw new Error(
          `Node.js version ${requiredVersion}+ required, found ${nodeVersion}`
        );
      }

      // Check write permissions
      await fs.promises.mkdir(config.projectPath, { recursive: true });
      const testPath = path.join(config.projectPath, '.test-write');
      await fs.promises.writeFile(testPath, 'test');
      await fs.promises.unlink(testPath);

      logger.debug('System requirements validated', {
        tag: '@SUCCESS:SYSTEM-VALIDATION-001',
      });
    } catch (error) {
      logger.error('System validation failed', {
        error,
        tag: '@ERROR:SYSTEM-VALIDATION-001',
      });
      throw error;
    }
  }

  /**
   * Validate installation completeness
   * @param config Installation configuration
   * @throws Error if validation fails
   * @tags @CODE:VALIDATE-INSTALLATION-001:API
   */
  async validateInstallation(config: InstallationConfig): Promise<void> {
    try {
      const requiredPaths = [
        path.join(config.projectPath, '.claude'),
        path.join(config.projectPath, '.moai'),
        path.join(config.projectPath, '.claude', 'settings.json'),
        path.join(config.projectPath, '.moai', 'config.json'),
      ];

      for (const requiredPath of requiredPaths) {
        if (!fs.existsSync(requiredPath)) {
          throw new Error(`Required path missing: ${requiredPath}`);
        }
      }

      // Validate JSON files
      const settingsPath = path.join(
        config.projectPath,
        '.claude',
        'settings.json'
      );
      const configPath = path.join(config.projectPath, '.moai', 'config.json');

      JSON.parse(await fs.promises.readFile(settingsPath, 'utf-8'));
      JSON.parse(await fs.promises.readFile(configPath, 'utf-8'));

      logger.debug('Installation validation successful', {
        tag: '@SUCCESS:VALIDATE-INSTALLATION-001',
      });
    } catch (error) {
      logger.error('Installation validation failed', {
        error,
        tag: '@ERROR:VALIDATE-INSTALLATION-001',
      });
      throw error;
    }
  }

  /**
   * Validate directory structure can be created
   * @param config Installation configuration
   * @returns true if directory structure is valid
   * @tags @CODE:VALIDATE-DIRECTORIES-001:API
   */
  async validateDirectoryStructure(
    config: InstallationConfig
  ): Promise<boolean> {
    try {
      const directories = [
        config.projectPath,
        path.join(config.projectPath, '.claude'),
        path.join(config.projectPath, '.moai'),
      ];

      // Check if we can create these directories
      for (const dir of directories) {
        const parentDir = path.dirname(dir);
        if (!fs.existsSync(parentDir)) {
          logger.warn('Parent directory does not exist', {
            parentDir,
            tag: '@WARN:PARENT-DIR-MISSING-001',
          });
          return false;
        }
      }

      return true;
    } catch (error) {
      logger.error('Directory structure validation failed', {
        error,
        tag: '@ERROR:DIR-VALIDATION-001',
      });
      return false;
    }
  }

  /**
   * Validate backup can be created
   * @param config Installation configuration
   * @returns true if backup can be created
   * @tags @CODE:VALIDATE-BACKUP-001:API
   */
  async validateBackupPossible(config: InstallationConfig): Promise<boolean> {
    try {
      if (!config.backupEnabled) {
        return true;
      }

      const backupDir = path.join(config.projectPath, '.moai-backup');
      const parentDir = path.dirname(backupDir);

      // Check if parent directory exists and is writable
      if (!fs.existsSync(parentDir)) {
        return false;
      }

      // Try to create backup directory
      await fs.promises.mkdir(backupDir, { recursive: true });
      return true;
    } catch (error) {
      logger.error('Backup validation failed', {
        error,
        tag: '@ERROR:BACKUP-VALIDATION-001',
      });
      return false;
    }
  }

  /**
   * Check if version satisfies minimum requirement
   * @param current Current version
   * @param required Required minimum version
   * @returns Whether version is satisfied
   * @tags @UTIL:VERSION-CHECK-001
   */
  private isVersionSatisfied(current: string, required: string): boolean {
    const currentParts = current.split('.').map(Number);
    const requiredParts = required.split('.').map(Number);

    for (
      let i = 0;
      i < Math.max(currentParts.length, requiredParts.length);
      i++
    ) {
      const currentPart = currentParts[i] || 0;
      const requiredPart = requiredParts[i] || 0;

      if (currentPart > requiredPart) return true;
      if (currentPart < requiredPart) return false;
    }
    return true;
  }
}
