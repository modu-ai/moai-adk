/**
 * @file Installation Phase Executor (Refactored)
 * @author MoAI Team
 * @tags @FEATURE:PHASE-EXECUTOR-001 @REQ:INSTALL-SYSTEM-012
 * @description Executes installation phases with dependency injection
 *
 * Chain: @REQ:INSTALL-SYSTEM-012 -> @DESIGN:PHASE-SPLIT-001 -> @TASK:EXECUTOR-001 -> @TEST:EXECUTOR-001
 * Related: @FEATURE:PHASE-VALIDATOR-001, @DOCS:INSTALL-001
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { execSync } from 'node:child_process';
import { logger } from '@/utils/logger';
import { InstallationError, getErrorMessage } from '@/utils/errors';
import { isInsideMoAIPackage } from '@/utils/path-validator';
import type {
  InstallationConfig,
  InstallationContext,
  ProgressCallback,
} from './types';
import { ContextManager } from './context-manager';
import { ResourceInstaller } from './resource-installer';
import { TemplateProcessor } from './template-processor';
import { PhaseValidator } from './phase-validator';

/**
 * Executes installation phases with dependency injection
 *
 * Responsibilities:
 * - Execute installation phases in sequence
 * - Coordinate with ResourceInstaller and TemplateProcessor
 * - Track phase progress via ContextManager
 * - Delegate validation to PhaseValidator
 *
 * @tags @FEATURE:PHASE-EXECUTOR-001
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
   * @param context Installation context
   * @param progressCallback Progress callback
   * @tags @API:EXECUTE-PREPARATION-001
   */
  async executePreparationPhase(
    context: InstallationContext,
    progressCallback?: ProgressCallback
  ): Promise<void> {
    const phaseStartTime = Date.now();
    const filesCreated: string[] = [];
    const errors: string[] = [];

    this.contextManager.updateProgress(
      context,
      'Phase 1: Preparation and backup...',
      5,
      progressCallback
    );

    try {
      if (context.config.backupEnabled) {
        await this.createBackup(context.config);
      }

      await this.validator.validateSystemRequirements(context.config);

      this.contextManager.recordPhaseCompletion(
        context,
        'preparation',
        phaseStartTime,
        filesCreated,
        errors
      );
    } catch (error) {
      errors.push(`Preparation phase failed: ${error}`);
      this.contextManager.recordPhaseCompletion(
        context,
        'preparation',
        phaseStartTime,
        filesCreated,
        errors
      );
      throw error;
    }
  }

  /**
   * Execute directory structure creation phase
   * @param context Installation context
   * @param progressCallback Progress callback
   * @tags @API:EXECUTE-DIRECTORY-001
   */
  async executeDirectoryPhase(
    context: InstallationContext,
    progressCallback?: ProgressCallback
  ): Promise<void> {
    const phaseStartTime = Date.now();
    const filesCreated: string[] = [];
    const errors: string[] = [];

    this.contextManager.updateProgress(
      context,
      'Phase 2: Creating directory structure...',
      5,
      progressCallback
    );

    try {
      await this.createProjectDirectories(context.config);
      filesCreated.push(context.config.projectPath);

      this.contextManager.recordPhaseCompletion(
        context,
        'directory',
        phaseStartTime,
        filesCreated,
        errors
      );
    } catch (error) {
      errors.push(`Directory phase failed: ${error}`);
      this.contextManager.recordPhaseCompletion(
        context,
        'directory',
        phaseStartTime,
        filesCreated,
        errors
      );
      throw error;
    }
  }

  /**
   * Execute resource installation phase
   * @param context Installation context
   * @param progressCallback Progress callback
   * @tags @API:EXECUTE-RESOURCE-001
   */
  async executeResourcePhase(
    context: InstallationContext,
    progressCallback?: ProgressCallback
  ): Promise<void> {
    const phaseStartTime = Date.now();
    const filesCreated: string[] = [];
    const errors: string[] = [];

    this.contextManager.updateProgress(
      context,
      'Phase 3: Installing resources...',
      5,
      progressCallback
    );

    try {
      const claudeFiles = await this.resourceInstaller.installClaudeResources(
        context.config
      );
      filesCreated.push(...claudeFiles);

      const moaiFiles = await this.resourceInstaller.installMoaiResources(
        context.config
      );
      filesCreated.push(...moaiFiles);

      const memoryFile = await this.resourceInstaller.installProjectMemory(
        context.config
      );
      if (memoryFile) {
        filesCreated.push(memoryFile);
      }

      this.contextManager.recordPhaseCompletion(
        context,
        'resource',
        phaseStartTime,
        filesCreated,
        errors
      );
    } catch (error) {
      errors.push(`Resource phase failed: ${error}`);
      this.contextManager.recordPhaseCompletion(
        context,
        'resource',
        phaseStartTime,
        filesCreated,
        errors
      );
      throw error;
    }
  }

  /**
   * Execute configuration generation phase
   * @param context Installation context
   * @param progressCallback Progress callback
   * @tags @API:EXECUTE-CONFIGURATION-001
   */
  async executeConfigurationPhase(
    context: InstallationContext,
    progressCallback?: ProgressCallback
  ): Promise<void> {
    const phaseStartTime = Date.now();
    const filesCreated: string[] = [];
    const errors: string[] = [];

    this.contextManager.updateProgress(
      context,
      'Phase 4: Generating configurations...',
      5,
      progressCallback
    );

    try {
      // Generate Claude Code settings if needed
      const existingClaudeSettingsPath = path.join(
        context.config.projectPath,
        '.claude',
        'settings.json'
      );
      if (!fs.existsSync(existingClaudeSettingsPath)) {
        const claudeSettingsPath = await this.resourceInstaller.createClaudeSettings(
          context.config
        );
        if (claudeSettingsPath) {
          filesCreated.push(claudeSettingsPath);
        }
      } else {
        logger.debug(
          'Claude settings already exists from template, skipping creation',
          {
            settingsPath: existingClaudeSettingsPath,
            tag: '@DEBUG:CLAUDE-SETTINGS-EXISTS-001',
          }
        );
        filesCreated.push(existingClaudeSettingsPath);
      }

      // Generate MoAI configuration
      const moaiConfigPath = await this.resourceInstaller.createMoaiConfig(
        context.config
      );
      if (moaiConfigPath) {
        filesCreated.push(moaiConfigPath);
      }

      // Generate gitignore if needed
      const gitignorePath = await this.resourceInstaller.createGitignore(
        context.config
      );
      if (gitignorePath) {
        filesCreated.push(gitignorePath);
      }

      this.contextManager.recordPhaseCompletion(
        context,
        'configuration',
        phaseStartTime,
        filesCreated,
        errors
      );
    } catch (error) {
      errors.push(`Configuration phase failed: ${error}`);
      this.contextManager.recordPhaseCompletion(
        context,
        'configuration',
        phaseStartTime,
        filesCreated,
        errors
      );
      throw error;
    }
  }

  /**
   * Execute validation and finalization phase
   * @param context Installation context
   * @param progressCallback Progress callback
   * @tags @API:EXECUTE-VALIDATION-001
   */
  async executeValidationPhase(
    context: InstallationContext,
    progressCallback?: ProgressCallback
  ): Promise<void> {
    const phaseStartTime = Date.now();
    const filesCreated: string[] = [];
    const errors: string[] = [];

    this.contextManager.updateProgress(
      context,
      'Phase 5: Validation and finalization...',
      5,
      progressCallback
    );

    try {
      await this.validator.validateInstallation(context.config);
      await this.finalizeInstallation(context.config);

      this.contextManager.recordPhaseCompletion(
        context,
        'validation',
        phaseStartTime,
        filesCreated,
        errors
      );
    } catch (error) {
      errors.push(`Validation phase failed: ${error}`);
      this.contextManager.recordPhaseCompletion(
        context,
        'validation',
        phaseStartTime,
        filesCreated,
        errors
      );
      throw error;
    }
  }

  // ============================================================================
  // Phase Implementation Methods
  // ============================================================================

  /**
   * Create backup of existing project
   * @param config Installation configuration
   * @tags @IMPL:BACKUP-001
   */
  private async createBackup(config: InstallationConfig): Promise<void> {
    try {
      // Prevent backup creation inside package directory
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

        logger.debug('Backup created at', {
          backupPath,
          tag: '@SUCCESS:BACKUP-001',
        });
      }
    } catch (error) {
      logger.error('Backup creation failed', {
        error,
        tag: '@ERROR:BACKUP-001',
      });
      throw error;
    }
  }

  /**
   * Create project directory structure
   * @param config Installation configuration
   * @tags @IMPL:CREATE-DIRECTORIES-001
   */
  private async createProjectDirectories(
    config: InstallationConfig
  ): Promise<void> {
    try {
      const directories = [
        config.projectPath,
        path.join(config.projectPath, '.claude'),
        path.join(config.projectPath, '.claude', 'logs'),
        path.join(config.projectPath, '.moai'),
        path.join(config.projectPath, '.moai', 'project'),
        path.join(config.projectPath, '.moai', 'specs'),
        path.join(config.projectPath, '.moai', 'reports'),
        path.join(config.projectPath, '.moai', 'memory'),
        // NOTE: .moai/indexes removed - CODE-FIRST TAG system doesn't use cached indexes
      ];

      for (const dir of directories) {
        await fs.promises.mkdir(dir, { recursive: true });
        logger.debug('Created directory', {
          dir,
          tag: '@DEBUG:DIR-CREATE-001',
        });
      }

      logger.debug('Project directories created', {
        tag: '@SUCCESS:CREATE-DIRECTORIES-001',
      });
    } catch (error) {
      logger.error('Directory creation failed', {
        error,
        tag: '@ERROR:CREATE-DIRECTORIES-001',
      });
      throw error;
    }
  }

  /**
   * Finalize installation with post-install tasks
   * @param config Installation configuration
   * @tags @IMPL:FINALIZE-INSTALLATION-001
   */
  private async finalizeInstallation(
    config: InstallationConfig
  ): Promise<void> {
    try {
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

      logger.debug('Installation finalized successfully', {
        projectPath: config.projectPath,
        mode: config.mode,
        tag: '@SUCCESS:FINALIZE-INSTALLATION-001',
      });
    } catch (error) {
      logger.error('Installation finalization failed', {
        error,
        tag: '@ERROR:FINALIZE-INSTALLATION-001',
      });
      throw error;
    }
  }

  // ============================================================================
  // Helper Methods
  // ============================================================================

  /**
   * Set executable permissions on scripts
   * @param scriptsDir Directory containing scripts
   * @tags @UTIL:SET-PERMISSIONS-001
   */
  private async setExecutablePermissions(scriptsDir: string): Promise<void> {
    try {
      const files = await fs.promises.readdir(scriptsDir);
      for (const file of files) {
        if (file.endsWith('.py') || file.endsWith('.sh')) {
          const filePath = path.join(scriptsDir, file);
          await fs.promises.chmod(filePath, 0o755);
        }
      }
    } catch (error) {
      logger.error('Failed to set permissions', {
        error,
        tag: '@ERROR:PERMISSIONS-001',
      });
    }
  }

  /**
   * Initialize Git repository
   * @param config Installation configuration
   * @tags @UTIL:INIT-GIT-001
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
        tag: '@SUCCESS:INIT-GIT-001',
      });
    } catch (error) {
      logger.error('Git initialization failed', {
        error,
        tag: '@ERROR:INIT-GIT-001',
      });
    }
  }
}