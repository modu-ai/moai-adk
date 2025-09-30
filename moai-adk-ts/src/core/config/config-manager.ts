// @FEATURE:CFG-001 | Chain: @REQ:CFG-001 -> @DESIGN:CFG-001 -> @TASK:CFG-001 -> @TEST:CFG-001
// Related: @API:CFG-001, @DATA:CFG-001

/**
 * @file Configuration management system
 * @author MoAI Team
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { logger } from '../../utils/winston-logger.js';
import type {
  BackupResult,
  ClaudeSettings,
  ClaudeSettingsResult,
  FullConfigResult,
  MoAIConfig,
  MoAIConfigResult,
  PackageConfig,
  PackageJsonResult,
  ProjectConfigInput,
  ValidationResult,
} from './types';

/**
 * ConfigManager class for creating and managing configuration files
 * @tags @TASK:CONFIG-MANAGER-001
 */
export class ConfigManager {
  private readonly version = '0.0.1';

  /**
   * Create Claude Code settings.json file
   * @param settingsPath Path to settings file
   * @param config Project configuration
   * @returns Claude settings creation result
   * @tags @API:CREATE-CLAUDE-SETTINGS-001
   */
  public async createClaudeSettings(
    settingsPath: string,
    config: ProjectConfigInput
  ): Promise<ClaudeSettingsResult> {
    try {
      const settingsDir = path.dirname(settingsPath);

      // Ensure directory exists
      if (!fs.existsSync(settingsDir)) {
        fs.mkdirSync(settingsDir, { recursive: true });
      }

      const settings: ClaudeSettings = {
        mode: config.mode,
        agents: {
          enabled: this.getEnabledAgents(config.mode),
          disabled: [],
        },
        commands: {
          enabled: this.getEnabledCommands(config.mode),
          shortcuts: this.getCommandShortcuts(),
        },
        hooks: {
          enabled: this.getEnabledHooks(config.mode),
          configuration: this.getHookConfiguration(config),
        },
        outputStyles: {
          default: 'study',
          available: ['study', 'professional', 'debug', 'minimal'],
        },
        features: {
          autoSync: config.mode === 'team',
          gitIntegration: true,
          tagTracking: true,
        },
        security: {
          allowedCommands: ['git', 'npm', 'node', 'python'],
          blockedPatterns: ['rm -rf', 'sudo', 'chmod 777'],
          requireApproval: ['--force', '--hard'],
        },
      };

      fs.writeFileSync(
        settingsPath,
        JSON.stringify(settings, null, 2),
        'utf-8'
      );

      logger.info(`Claude settings created: ${settingsPath}`);
      return {
        success: true,
        filePath: settingsPath,
        settings,
      };
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      logger.error('Error creating Claude settings:', errorMessage);
      return {
        success: false,
        error: errorMessage,
      };
    }
  }

  /**
   * Create .moai/config.json file
   * @param configPath Path to config file
   * @param config Project configuration
   * @returns MoAI config creation result
   * @tags @API:CREATE-MOAI-CONFIG-001
   */
  public async createMoAIConfig(
    configPath: string,
    config: ProjectConfigInput
  ): Promise<MoAIConfigResult> {
    try {
      const configDir = path.dirname(configPath);
      let backupCreated = false;

      // Ensure directory exists
      if (!fs.existsSync(configDir)) {
        fs.mkdirSync(configDir, { recursive: true });
      }

      // Backup existing config if it exists
      if (fs.existsSync(configPath) && config.backup !== false) {
        const backupResult = await this.backupConfigFile(configPath);
        backupCreated = backupResult.success;
      }

      const moaiConfig: MoAIConfig = {
        projectName: config.projectName,
        version: this.version,
        mode: config.mode,
        runtime: config.runtime,
        techStack: config.techStack,
        features: {
          tdd: true,
          tagSystem: true,
          gitAutomation: config.mode === 'team',
          documentSync: config.mode === 'team',
        },
        directories: {
          moai: '.moai',
          claude: '.claude',
          specs: '.moai/specs',
          templates: '.moai/templates',
        },
        createdAt: new Date(),
        updatedAt: new Date(),
      };

      fs.writeFileSync(
        configPath,
        JSON.stringify(moaiConfig, null, 2),
        'utf-8'
      );

      logger.info(`MoAI config created: ${configPath}`);
      return {
        success: true,
        filePath: configPath,
        config: moaiConfig,
        backupCreated,
      };
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      logger.error('Error creating MoAI config:', errorMessage);
      return {
        success: false,
        error: errorMessage,
      };
    }
  }

  /**
   * Create package.json file for Node.js projects
   * @param packagePath Path to package.json
   * @param config Project configuration
   * @returns Package.json creation result
   * @tags @API:CREATE-PACKAGE-JSON-001
   */
  public async createPackageJson(
    packagePath: string,
    config: ProjectConfigInput
  ): Promise<PackageJsonResult> {
    try {
      // Skip if not needed
      if (!config.shouldCreatePackageJson) {
        return {
          success: true,
          skipped: true,
          reason: 'package.json not needed for this project type',
        };
      }

      const packageConfig: PackageConfig = {
        name: config.projectName,
        version: '1.0.0',
        description: `MoAI-ADK project: ${config.projectName}`,
        main: 'index.js',
        scripts: this.getPackageScripts(config),
        dependencies: this.getPackageDependencies(config),
        devDependencies: this.getPackageDevDependencies(config),
        keywords: ['moai-adk', ...config.techStack],
        author: 'MoAI Developer',
        license: 'MIT',
        engines: {
          node: '>=18.0.0',
        },
      };

      fs.writeFileSync(
        packagePath,
        JSON.stringify(packageConfig, null, 2),
        'utf-8'
      );

      logger.info(`Package.json created: ${packagePath}`);
      return {
        success: true,
        filePath: packagePath,
        packageConfig,
      };
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      logger.error('Error creating package.json:', errorMessage);
      return {
        success: false,
        error: errorMessage,
      };
    }
  }

  /**
   * Validate configuration file format and content
   * @param filePath Path to configuration file
   * @returns Validation result
   * @tags @API:VALIDATE-CONFIG-001
   */
  public async validateConfigFile(filePath: string): Promise<ValidationResult> {
    const result: ValidationResult = {
      isValid: true,
      errors: [],
      warnings: [],
      suggestions: [],
    };

    try {
      if (!fs.existsSync(filePath)) {
        result.isValid = false;
        result.errors.push('File does not exist');
        return result;
      }

      const content = fs.readFileSync(filePath, 'utf-8');
      JSON.parse(content); // Validate JSON format

      logger.info(`Configuration file validated: ${filePath}`);
      return result;
    } catch (error) {
      result.isValid = false;
      if (error instanceof SyntaxError) {
        result.errors.push('Invalid JSON format');
      } else {
        result.errors.push(
          error instanceof Error ? error.message : 'Unknown error'
        );
      }
      return result;
    }
  }

  /**
   * Create backup of configuration file
   * @param filePath Path to file to backup
   * @returns Backup result
   * @tags @API:BACKUP-CONFIG-001
   */
  public async backupConfigFile(filePath: string): Promise<BackupResult> {
    try {
      if (!fs.existsSync(filePath)) {
        return {
          success: false,
          error: 'File does not exist',
          timestamp: new Date(),
        };
      }

      const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
      const dir = path.dirname(filePath);
      const name = path.basename(filePath, path.extname(filePath));
      const ext = path.extname(filePath);
      const backupPath = path.join(dir, `${name}.backup.${timestamp}${ext}`);

      const content = fs.readFileSync(filePath, 'utf-8');
      fs.writeFileSync(backupPath, content, 'utf-8');

      logger.info(`Backup created: ${backupPath}`);
      return {
        success: true,
        backupPath,
        timestamp: new Date(),
      };
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      return {
        success: false,
        error: errorMessage,
        timestamp: new Date(),
      };
    }
  }

  /**
   * Setup complete project configuration
   * @param projectPath Project directory path
   * @param config Project configuration
   * @returns Full configuration result
   * @tags @API:SETUP-FULL-CONFIG-001
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

  // Private helper methods

  private getEnabledAgents(mode: string): string[] {
    const baseAgents = ['spec-builder', 'code-builder', 'doc-syncer'];
    if (mode === 'team') {
      return [...baseAgents, 'git-manager', 'debug-helper'];
    }
    return baseAgents;
  }

  private getEnabledCommands(mode: string): string[] {
    const baseCommands = [
      '/moai:8-project',
      '/moai:1-spec',
      '/moai:2-build',
      '/moai:3-sync',
    ];
    if (mode === 'team') {
      return [...baseCommands, '/moai:4-debug'];
    }
    return baseCommands;
  }

  private getCommandShortcuts(): Record<string, string> {
    return {
      spec: '/moai:1-spec',
      build: '/moai:2-build',
      sync: '/moai:3-sync',
    };
  }

  private getEnabledHooks(mode: string): string[] {
    const baseHooks = ['steering-guard', 'pre-write-guard', 'file-monitor'];
    if (mode === 'team') {
      return [...baseHooks, 'test-runner', 'policy-block'];
    }
    return baseHooks;
  }

  private getHookConfiguration(
    config: ProjectConfigInput
  ): Record<string, any> {
    return {
      'steering-guard': {
        enabled: true,
        checkPatterns: ['*.ts', '*.js', '*.py'],
      },
      'file-monitor': {
        watchPaths: ['src/', 'tests/'],
        extensions: config.techStack.includes('typescript')
          ? ['.ts', '.tsx']
          : ['.js', '.jsx'],
      },
    };
  }

  private getPackageScripts(
    config: ProjectConfigInput
  ): Record<string, string> {
    const scripts: Record<string, string> = {
      build: 'tsc',
      dev: 'tsx src/index.ts',
      test: 'jest',
      lint: 'eslint src/',
      format: 'prettier --write src/',
    };

    if (config.techStack.includes('typescript')) {
      scripts['type-check'] = 'tsc --noEmit';
    }

    if (config.techStack.includes('react')) {
      scripts['dev'] = 'vite';
      scripts['build'] = 'vite build';
    }

    return scripts;
  }

  private getPackageDependencies(
    config: ProjectConfigInput
  ): Record<string, string> {
    const deps: Record<string, string> = {};

    if (config.techStack.includes('react')) {
      deps['react'] = '^18.0.0';
      deps['react-dom'] = '^18.0.0';
    }

    if (config.techStack.includes('nextjs')) {
      deps['next'] = '^13.0.0';
    }

    if (config.techStack.includes('express')) {
      deps['express'] = '^4.18.0';
    }

    return deps;
  }

  private getPackageDevDependencies(
    config: ProjectConfigInput
  ): Record<string, string> {
    const devDeps: Record<string, string> = {
      jest: '^29.0.0',
      eslint: '^8.0.0',
      prettier: '^3.0.0',
    };

    if (config.techStack.includes('typescript')) {
      devDeps['typescript'] = '^5.0.0';
      devDeps['@types/node'] = '^20.0.0';
      devDeps['tsx'] = '^4.0.0';
    }

    if (config.techStack.includes('react')) {
      devDeps['@types/react'] = '^18.0.0';
      devDeps['@types/react-dom'] = '^18.0.0';
      devDeps['vite'] = '^4.0.0';
    }

    return devDeps;
  }
}
