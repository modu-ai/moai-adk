// @CODE:INST-002 |
// Related: @CODE:INST-002:API, @CODE:INIT-003:BACKUP

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
import {
  type BackupMetadata,
  saveBackupMetadata,
} from './backup-metadata';
import {
  hasAnyMoAIFiles,
  generateBackupDirName,
  getBackupTargets,
} from './backup-utils';

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
   * @CODE:INIT-003:BACKUP - Phase A implementation (v0.2.1: Refactored with backup-utils)
   */
  private async createBackup(config: InstallationConfig): Promise<void> {
    if (isInsideMoAIPackage(config.projectPath)) {
      throw new InstallationError(
        'Cannot create backup inside MoAI-ADK package directory',
        {
          phase: 'BACKUP_IN_PACKAGE_DIR',
          projectPath: config.projectPath,
        }
      );
    }

    // v0.2.1: Check if ANY MoAI-ADK files exist (OR condition) - using backup-utils
    if (!hasAnyMoAIFiles(config.projectPath)) {
      logger.debug('No MoAI-ADK files found, skipping backup', {
        projectPath: config.projectPath,
        tag: 'INFO:SKIP-BACKUP-001',
      });
      return;
    }

    // Generate backup directory name - using backup-utils
    const backupDirName = generateBackupDirName();
    const backupPath = path.join(config.projectPath, backupDirName);

    // Create backup directory
    await fs.promises.mkdir(backupPath, { recursive: true });

    // v0.2.1: Get selective backup targets - using backup-utils
    const targets = getBackupTargets(config.projectPath);
    const backedUpFiles: string[] = [];

    // Backup files and directories
    for (const target of targets) {
      const srcPath = path.join(config.projectPath, target);
      const dstPath = path.join(backupPath, target);

      if (fs.statSync(srcPath).isDirectory()) {
        await this.templateProcessor.copyDirectory(srcPath, dstPath);
        backedUpFiles.push(`${target}/`);
      } else {
        await fs.promises.copyFile(srcPath, dstPath);
        backedUpFiles.push(target);
      }
    }

    logger.debug('Backup created', {
      backupPath,
      backedUpFiles,
      tag: 'SUCCESS:BACKUP-001',
    });

    // Save backup metadata (Phase A: SPEC-INIT-003 v0.2.1)
    const metadata: BackupMetadata = {
      timestamp: new Date().toISOString(),
      backup_path: backupDirName,
      backed_up_files: backedUpFiles,
      status: 'pending',
      created_by: 'moai init',
    };

    await saveBackupMetadata(config.projectPath, metadata);

    logger.info('Backup metadata saved', {
      metadataPath: path.join(
        config.projectPath,
        '.moai',
        'backups',
        'latest.json'
      ),
      backedUpFiles,
      tag: 'SUCCESS:BACKUP-METADATA-001',
    });
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
      // Only .sh scripts need executable permissions now (Python hooks replaced with TS)
      if (file.endsWith('.sh')) {
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
