// @CODE:CFG-001 |
// Related: @CODE:CFG-001:API, @CODE:CFG-001:DATA

/**
 * @file Configuration management system (Facade)
 * @author MoAI Team
 */

import * as path from 'node:path';
import { logger } from '../../utils/winston-logger.js';
import { buildClaudeSettings } from './builders/claude-settings-builder.js';
import { buildMoAIConfig } from './builders/moai-config-builder.js';
import { buildPackageJson } from './builders/package-json-builder.js';
import type {
  BackupResult,
  ClaudeSettingsResult,
  FullConfigResult,
  MoAIConfigResult,
  PackageJsonResult,
  ProjectConfigInput,
  ValidationResult,
} from './types';
import {
  backupConfigFile,
  validateConfigFile,
} from './utils/config-file-utils.js';

/**
 * ConfigManager class - Facade for configuration builders
 * @tags @CODE:CONFIG-MANAGER-001
 */
export class ConfigManager {
  private readonly version = '0.0.1';

  /**
   * Create Claude Code settings.json file
   * @param settingsPath Path to settings file
   * @param config Project configuration
   * @returns Claude settings creation result
   * @tags @CODE:CREATE-CLAUDE-SETTINGS-001:API
   */
  public async createClaudeSettings(
    settingsPath: string,
    config: ProjectConfigInput
  ): Promise<ClaudeSettingsResult> {
    return buildClaudeSettings(settingsPath, config);
  }

  /**
   * Create .moai/config.json file
   * @param configPath Path to config file
   * @param config Project configuration
   * @returns MoAI config creation result
   * @tags @CODE:CREATE-MOAI-CONFIG-001:API
   */
  public async createMoAIConfig(
    configPath: string,
    config: ProjectConfigInput
  ): Promise<MoAIConfigResult> {
    return buildMoAIConfig(configPath, config, this.version);
  }

  /**
   * Create package.json file for Node.js projects
   * @param packagePath Path to package.json
   * @param config Project configuration
   * @returns Package.json creation result
   * @tags @CODE:CREATE-PACKAGE-JSON-001:API
   */
  public async createPackageJson(
    packagePath: string,
    config: ProjectConfigInput
  ): Promise<PackageJsonResult> {
    return buildPackageJson(packagePath, config);
  }

  /**
   * Validate configuration file format and content
   * @param filePath Path to configuration file
   * @returns Validation result
   * @tags @CODE:VALIDATE-CONFIG-001:API
   */
  public async validateConfigFile(filePath: string): Promise<ValidationResult> {
    return validateConfigFile(filePath);
  }

  /**
   * Create backup of configuration file
   * @param filePath Path to file to backup
   * @returns Backup result
   * @tags @CODE:BACKUP-CONFIG-001:API
   */
  public async backupConfigFile(filePath: string): Promise<BackupResult> {
    return backupConfigFile(filePath);
  }

  /**
   * Setup complete project configuration
   * @param projectPath Project directory path
   * @param config Project configuration
   * @returns Full configuration result
   * @tags @CODE:SETUP-FULL-CONFIG-001:API
   */
  public async setupFullProjectConfig(
    projectPath: string,
    config: ProjectConfigInput
  ): Promise<FullConfigResult> {
    const startTime = Date.now();
    const result: FullConfigResult = {
      success: true,
      filesCreated: [],
      errors: [],
      duration: 0,
      timestamp: new Date(),
    };

    try {
      // Create Claude settings
      const claudeSettingsPath = path.join(
        projectPath,
        '.claude',
        'settings.json'
      );
      const claudeResult = await this.createClaudeSettings(
        claudeSettingsPath,
        config
      );
      result.claudeSettings = claudeResult;

      if (claudeResult.success && claudeResult.filePath) {
        result.filesCreated.push(claudeResult.filePath);
      } else if (claudeResult.error) {
        result.errors.push(claudeResult.error);
        result.success = false;
      }

      // Create MoAI config
      const moaiConfigPath = path.join(projectPath, '.moai', 'config.json');
      const moaiResult = await this.createMoAIConfig(moaiConfigPath, config);
      result.moaiConfig = moaiResult;

      if (moaiResult.success && moaiResult.filePath) {
        result.filesCreated.push(moaiResult.filePath);
      } else if (moaiResult.error) {
        result.errors.push(moaiResult.error);
        result.success = false;
      }

      // Create package.json if needed
      if (config.shouldCreatePackageJson) {
        const packageJsonPath = path.join(projectPath, 'package.json');
        const packageResult = await this.createPackageJson(
          packageJsonPath,
          config
        );
        result.packageJson = packageResult;

        if (packageResult.success && packageResult.filePath) {
          result.filesCreated.push(packageResult.filePath);
        } else if (packageResult.error) {
          result.errors.push(packageResult.error);
          result.success = false;
        }
      }

      result.duration = Date.now() - startTime;
      logger.info(
        `Full project configuration completed in ${result.duration}ms`
      );
      return result;
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      result.success = false;
      result.errors.push(errorMessage);
      result.duration = Date.now() - startTime;
      return result;
    }
  }
}
