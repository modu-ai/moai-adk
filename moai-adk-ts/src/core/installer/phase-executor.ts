// @CODE:INST-002 |
// Related: @CODE:INST-002:API

/**
 * @file Installation phase execution engine
 * @author MoAI Team
 */

import { execSync } from 'node:child_process';
import * as fs from 'node:fs';
import * as path from 'node:path';
import { InstallationError } from '@/utils/errors';
import { isInsideMoAIPackage } from '@/utils/path-validator';
import { logger } from '@/utils/winston-logger';
import type { ContextManager } from './context-manager';
import { PhaseValidator } from './phase-validator';
import { ResourceInstaller } from './resource-installer';
import { TemplateProcessor } from './template-processor';
import type {
  InstallationConfig,
  InstallationContext,
  ProgressCallback,
} from './types';
import { executePhase } from './utils/phase-runner.js';

/**
 * Executes installation phases with dependency injection
 *
 * Responsibilities:
 * - Execute installation phases using template pattern
 * - Coordinate with ResourceInstaller and TemplateProcessor
 * - Track phase progress via ContextManager
 * - Delegate validation to PhaseValidator
 *
 * @tags @CODE:PHASE-EXECUTOR-001
 */
export class PhaseExecutor {
  private readonly contextManager: ContextManager;
  private readonly resourceInstaller: ResourceInstaller;
  private readonly templateProcessor: TemplateProcessor;
  private readonly validator: PhaseValidator;

  constructor(contextManager: ContextManager) {
    this.contextManager = contextManager;
    this.templateProcessor = new TemplateProcessor();
    this.resourceInstaller = new ResourceInstaller(this.templateProcessor);
    this.validator = new PhaseValidator();
  }

  /**
   * Execute preparation phase including backup creation
   */
  async executePreparationPhase(
    context: InstallationContext,
    progressCallback?: ProgressCallback
  ): Promise<void> {
    await executePhase(
      'preparation',
      'Phase 1: Preparation and backup...',
      context,
      this.contextManager,
      async () => {
        if (context.config.backupEnabled) {
          await this.createBackup(context.config);
        }
        await this.validator.validateSystemRequirements(context.config);
        return [];
      },
      progressCallback
    );
  }

  /**
   * Execute directory structure creation phase
   */
  async executeDirectoryPhase(
    context: InstallationContext,
    progressCallback?: ProgressCallback
  ): Promise<void> {
    await executePhase(
      'directory',
      'Phase 2: Creating directory structure...',
      context,
      this.contextManager,
      async () => {
        await this.createProjectDirectories(context.config);
        return [context.config.projectPath];
      },
      progressCallback
    );
  }

  /**
   * Execute resource installation phase
   */
  async executeResourcePhase(
    context: InstallationContext,
    progressCallback?: ProgressCallback
  ): Promise<void> {
    await executePhase(
      'resource',
      'Phase 3: Installing resources...',
      context,
      this.contextManager,
      async () => {
        const files: string[] = [];

        const claudeFiles = await this.resourceInstaller.installClaudeResources(
          context.config
        );
        files.push(...claudeFiles);

        const moaiFiles = await this.resourceInstaller.installMoaiResources(
          context.config
        );
        files.push(...moaiFiles);

        const memoryFile = await this.resourceInstaller.installProjectMemory(
          context.config
        );
        if (memoryFile) {
          files.push(memoryFile);
        }

        return files;
      },
      progressCallback
    );
  }

  /**
   * Execute configuration generation phase
   */
  async executeConfigurationPhase(
    context: InstallationContext,
    progressCallback?: ProgressCallback
  ): Promise<void> {
    await executePhase(
      'configuration',
      'Phase 4: Generating configurations...',
      context,
      this.contextManager,
      async () => {
        const files: string[] = [];

        // Generate Claude Code settings if needed
        const claudeSettingsPath = path.join(
          context.config.projectPath,
          '.claude',
          'settings.json'
        );

        if (!fs.existsSync(claudeSettingsPath)) {
          const settingsPath =
            await this.resourceInstaller.createClaudeSettings(context.config);
          if (settingsPath) {
            files.push(settingsPath);
          }
        } else {
          logger.debug('Claude settings already exists, skipping', {
            settingsPath: claudeSettingsPath,
            tag: 'DEBUG:CLAUDE-SETTINGS-EXISTS-001',
          });
          files.push(claudeSettingsPath);
        }

        // Generate MoAI configuration
        const moaiConfigPath = await this.resourceInstaller.createMoaiConfig(
          context.config
        );
        if (moaiConfigPath) {
          files.push(moaiConfigPath);
        }

        // Generate gitignore
        const gitignorePath = await this.resourceInstaller.createGitignore(
          context.config
        );
        if (gitignorePath) {
          files.push(gitignorePath);
        }

        return files;
      },
      progressCallback
    );
  }

  /**
   * Execute validation and finalization phase
   */
  async executeValidationPhase(
    context: InstallationContext,
    progressCallback?: ProgressCallback
  ): Promise<void> {
    await executePhase(
      'validation',
      'Phase 5: Validation and finalization...',
      context,
      this.contextManager,
      async () => {
        await this.validator.validateInstallation(context.config);
        await this.finalizeInstallation(context.config);
        return [];
      },
      progressCallback
    );
  }

  // ============================================================================
  // Phase Implementation Methods
  // ============================================================================

  /**
   * Create backup of existing project
   */
  private async createBackup(config: InstallationConfig): Promise<void> {
    if (isInsideMoAIPackage(config.projectPath)) {
      throw new InstallationError(
        'Cannot create backup inside MoAI-ADK package directory',
        'BACKUP_IN_PACKAGE_DIR'
      );
    }

    const backupDir = path.join(config.projectPath, '.moai-backup');
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const backupPath = path.join(backupDir, `backup-${timestamp}`);

    if (fs.existsSync(config.projectPath)) {
      await fs.promises.mkdir(backupPath, { recursive: true });

      // Backup critical directories
      const criticalDirs = ['.claude', '.moai'];
      for (const dir of criticalDirs) {
        const srcPath = path.join(config.projectPath, dir);
        const dstPath = path.join(backupPath, dir);
        if (fs.existsSync(srcPath)) {
          await this.templateProcessor.copyDirectory(srcPath, dstPath);
        }
      }

      // Backup critical files
      const criticalFiles = ['CLAUDE.md'];
      for (const file of criticalFiles) {
        const srcPath = path.join(config.projectPath, file);
        const dstPath = path.join(backupPath, file);
        if (fs.existsSync(srcPath)) {
          await fs.promises.copyFile(srcPath, dstPath);
        }
      }

      logger.debug('Backup created', {
        backupPath,
        tag: 'SUCCESS:BACKUP-001',
      });
    }
  }

  /**
   * Create project directory structure
   */
  private async createProjectDirectories(
    config: InstallationConfig
  ): Promise<void> {
    const directories = [
      config.projectPath,
      path.join(config.projectPath, '.claude'),
      path.join(config.projectPath, '.claude', 'logs'),
      path.join(config.projectPath, '.moai'),
      path.join(config.projectPath, '.moai', 'project'),
      path.join(config.projectPath, '.moai', 'specs'),
      path.join(config.projectPath, '.moai', 'reports'),
      path.join(config.projectPath, '.moai', 'memory'),
    ];

    for (const dir of directories) {
      await fs.promises.mkdir(dir, { recursive: true });
      logger.debug('Created directory', { dir, tag: 'DEBUG:DIR-CREATE-001' });
    }

    logger.debug('Project directories created', {
      tag: 'SUCCESS:CREATE-DIRECTORIES-001',
    });
  }

  /**
   * Finalize installation with post-install tasks
   */
  private async finalizeInstallation(
    config: InstallationConfig
  ): Promise<void> {
    // Set permissions on non-Windows
    if (process.platform !== 'win32') {
      const scriptsDir = path.join(config.projectPath, '.claude', 'hooks');
      if (fs.existsSync(scriptsDir)) {
        await this.setExecutablePermissions(scriptsDir);
      }
    }

    // Initialize Git if team mode
    if (config.mode === 'team') {
      await this.initializeGitRepository(config);
    }

    logger.debug('Installation finalized', {
      projectPath: config.projectPath,
      mode: config.mode,
      tag: 'SUCCESS:FINALIZE-INSTALLATION-001',
    });
  }

  // ============================================================================
  // Helper Methods
  // ============================================================================

  /**
   * Set executable permissions on scripts
   */
  private async setExecutablePermissions(scriptsDir: string): Promise<void> {
    const files = await fs.promises.readdir(scriptsDir);
    for (const file of files) {
      if (file.endsWith('.py') || file.endsWith('.sh')) {
        const filePath = path.join(scriptsDir, file);
        await fs.promises.chmod(filePath, 0o755);
      }
    }
  }

  /**
   * Initialize Git repository
   */
  private async initializeGitRepository(
    config: InstallationConfig
  ): Promise<void> {
    try {
      process.chdir(config.projectPath);
      execSync('git init', { stdio: 'ignore' });
      execSync('git add .', { stdio: 'ignore' });
      execSync('git commit -m "Initial MoAI-ADK project setup"', {
        stdio: 'ignore',
      });

      logger.debug('Git repository initialized', {
        tag: 'SUCCESS:INIT-GIT-001',
      });
    } catch (error) {
      logger.error('Git initialization failed', {
        error,
        tag: 'ERROR:INIT-GIT-001',
      });
    }
  }
}
