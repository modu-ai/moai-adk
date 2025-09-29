/**
 * @file Installation Process Orchestrator
 * @author MoAI Team
 * @tags @FEATURE:INSTALL-ORCHESTRATOR-001 @REQ:INSTALL-SYSTEM-012
 */

import { logger } from '@/utils/logger';
import * as fs from 'fs';
import * as path from 'path';
import mustache from 'mustache';
import type {
  InstallationConfig,
  InstallationResult,
  InstallationContext,
  ProgressCallback,
  PhaseStatus,
} from './types';
// import { getDefaultTagDatabase } from '../tag-system/tag-database';

/**
 * Central coordinator for MoAI-ADK installation
 * @tags @FEATURE:INSTALL-ORCHESTRATOR-001
 */
export class InstallationOrchestrator {
  private readonly config: InstallationConfig;
  private context: InstallationContext;

  constructor(config: InstallationConfig) {
    this.config = config;
    this.context = this.createInitialContext(config);

    logger.debug('InstallationOrchestrator initialized', {
      projectPath: config.projectPath,
      mode: config.mode,
      tag: '@INIT:ORCHESTRATOR-001',
    });
  }

  /**
   * Execute complete MoAI-ADK project installation
   * @param progressCallback Progress callback function
   * @returns Complete installation result
   * @tags @API:EXECUTE-INSTALLATION-001
   */
  public async executeInstallation(
    progressCallback?: ProgressCallback
  ): Promise<InstallationResult> {
    const startTime = Date.now();

    try {
      await this.executePreparationPhase(progressCallback);
      await this.executeDirectoryPhase(progressCallback);
      await this.executeResourcePhase(progressCallback);
      await this.executeConfigurationPhase(progressCallback);
      await this.executeValidationPhase(progressCallback);

      this.updateProgress('Installation complete!', progressCallback);

      return this.createSuccessResult(startTime);
    } catch (error) {
      logger.error('Installation failed', {
        error,
        tag: '@ERROR:INSTALL-FAILED-001',
      });
      return this.createFailureResult(startTime, error);
    }
  }

  /**
   * Execute preparation phase including backup creation
   * @param progressCallback Progress callback
   * @tags @PHASE:PREPARATION-001
   */
  private async executePreparationPhase(
    progressCallback?: ProgressCallback
  ): Promise<void> {
    const phaseStartTime = Date.now();
    const filesCreated: string[] = [];
    const errors: string[] = [];

    this.updateProgress('Phase 1: Preparation and backup...', progressCallback);

    try {
      if (this.config.backupEnabled) {
        await this.createBackup();
      }

      await this.validateSystemRequirements();

      this.recordPhaseCompletion(
        'preparation',
        phaseStartTime,
        filesCreated,
        errors
      );
    } catch (error) {
      errors.push(`Preparation phase failed: ${error}`);
      this.recordPhaseCompletion(
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
   * @param progressCallback Progress callback
   * @tags @PHASE:DIRECTORY-001
   */
  private async executeDirectoryPhase(
    progressCallback?: ProgressCallback
  ): Promise<void> {
    const phaseStartTime = Date.now();
    const filesCreated: string[] = [];
    const errors: string[] = [];

    this.updateProgress(
      'Phase 2: Creating directory structure...',
      progressCallback
    );

    try {
      await this.createProjectDirectories();
      filesCreated.push(this.config.projectPath);

      this.recordPhaseCompletion(
        'directory',
        phaseStartTime,
        filesCreated,
        errors
      );
    } catch (error) {
      errors.push(`Directory phase failed: ${error}`);
      this.recordPhaseCompletion(
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
   * @param progressCallback Progress callback
   * @tags @PHASE:RESOURCE-001
   */
  private async executeResourcePhase(
    progressCallback?: ProgressCallback
  ): Promise<void> {
    const phaseStartTime = Date.now();
    const filesCreated: string[] = [];
    const errors: string[] = [];

    this.updateProgress('Phase 3: Installing resources...', progressCallback);

    try {
      const installedFiles = await this.installResources();
      filesCreated.push(...installedFiles);

      this.recordPhaseCompletion(
        'resource',
        phaseStartTime,
        filesCreated,
        errors
      );
    } catch (error) {
      errors.push(`Resource phase failed: ${error}`);
      this.recordPhaseCompletion(
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
   * @param progressCallback Progress callback
   * @tags @PHASE:CONFIGURATION-001
   */
  private async executeConfigurationPhase(
    progressCallback?: ProgressCallback
  ): Promise<void> {
    const phaseStartTime = Date.now();
    const filesCreated: string[] = [];
    const errors: string[] = [];

    this.updateProgress(
      'Phase 4: Generating configurations...',
      progressCallback
    );

    try {
      const configFiles = await this.generateConfigurations();
      filesCreated.push(...configFiles);

      this.recordPhaseCompletion(
        'configuration',
        phaseStartTime,
        filesCreated,
        errors
      );
    } catch (error) {
      errors.push(`Configuration phase failed: ${error}`);
      this.recordPhaseCompletion(
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
   * @param progressCallback Progress callback
   * @tags @PHASE:VALIDATION-001
   */
  private async executeValidationPhase(
    progressCallback?: ProgressCallback
  ): Promise<void> {
    const phaseStartTime = Date.now();
    const filesCreated: string[] = [];
    const errors: string[] = [];

    this.updateProgress(
      'Phase 5: Validation and finalization...',
      progressCallback
    );

    try {
      await this.validateInstallation();
      await this.finalizeInstallation();

      this.recordPhaseCompletion(
        'validation',
        phaseStartTime,
        filesCreated,
        errors
      );
    } catch (error) {
      errors.push(`Validation phase failed: ${error}`);
      this.recordPhaseCompletion(
        'validation',
        phaseStartTime,
        filesCreated,
        errors
      );
      throw error;
    }
  }

  /**
   * Create initial installation context
   * @param config Installation configuration
   * @returns Initial context
   * @tags @UTIL:CREATE-CONTEXT-001
   */
  private createInitialContext(
    config: InstallationConfig
  ): InstallationContext {
    return {
      config,
      startTime: new Date(),
      phases: [],
      allFilesCreated: [],
      allErrors: [],
    };
  }

  /**
   * Update progress and notify callback
   * @param message Progress message
   * @param callback Progress callback
   * @tags @UTIL:UPDATE-PROGRESS-001
   */
  private updateProgress(message: string, callback?: ProgressCallback): void {
    const current = this.context.phases.length;
    const total = 5; // Total number of phases

    logger.debug(message, { current, total, tag: '@PROGRESS:UPDATE-001' });

    if (callback) {
      callback(message, current, total);
    }
  }

  /**
   * Record phase completion
   * @param phaseName Phase name
   * @param startTime Phase start time
   * @param filesCreated Files created in phase
   * @param errors Errors in phase
   * @tags @UTIL:RECORD-PHASE-001
   */
  private recordPhaseCompletion(
    phaseName: string,
    startTime: number,
    filesCreated: string[],
    errors: string[]
  ): void {
    const phase: PhaseStatus = {
      name: phaseName,
      completed: errors.length === 0,
      duration: Date.now() - startTime,
      errors: [...errors],
      filesCreated: [...filesCreated],
    };

    this.context.phases.push(phase);
    this.context.allFilesCreated.push(...filesCreated);
    this.context.allErrors.push(...errors);
  }

  /**
   * Create success result
   * @param startTime Installation start time
   * @returns Success result
   * @tags @UTIL:CREATE-SUCCESS-RESULT-001
   */
  private createSuccessResult(startTime: number): InstallationResult {
    return {
      success: true,
      projectPath: this.config.projectPath,
      filesCreated: [...this.context.allFilesCreated],
      errors: [...this.context.allErrors],
      nextSteps: this.generateNextSteps(),
      config: this.config,
      timestamp: new Date(),
      duration: Date.now() - startTime,
    };
  }

  /**
   * Create failure result
   * @param startTime Installation start time
   * @param error Failure error
   * @returns Failure result
   * @tags @UTIL:CREATE-FAILURE-RESULT-001
   */
  private createFailureResult(
    startTime: number,
    error: unknown
  ): InstallationResult {
    const errorMessage = error instanceof Error ? error.message : String(error);

    return {
      success: false,
      projectPath: this.config.projectPath,
      filesCreated: [...this.context.allFilesCreated],
      errors: [
        ...this.context.allErrors,
        `Installation failed: ${errorMessage}`,
      ],
      nextSteps: ['Fix the errors above and retry installation'],
      config: this.config,
      timestamp: new Date(),
      duration: Date.now() - startTime,
    };
  }

  /**
   * Generate next steps for user
   * @returns Array of next steps
   * @tags @UTIL:GENERATE-NEXT-STEPS-001
   */
  private generateNextSteps(): string[] {
    const steps = [
      `cd ${this.config.projectName}`,
      'claude',
      '',  // Empty line for better readability
      `💡 Tip: Next time use "moai init ${this.config.projectName}" to recreate this setup`,
    ];

    if (this.config.mode === 'team') {
      steps.push('Initialize Git repository and configure team settings');
    }

    return steps;
  }

  /**
   * Create backup of existing project
   * @tags @IMPL:BACKUP-001
   */
  private async createBackup(): Promise<void> {
    try {
      const backupDir = path.join(this.config.projectPath, '.moai-backup');
      const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
      const backupPath = path.join(backupDir, `backup-${timestamp}`);

      if (fs.existsSync(this.config.projectPath)) {
        await fs.promises.mkdir(backupPath, { recursive: true });
        // Copy critical files (.claude, .moai directories if they exist)
        const criticalDirs = ['.claude', '.moai'];
        for (const dir of criticalDirs) {
          const srcPath = path.join(this.config.projectPath, dir);
          const dstPath = path.join(backupPath, dir);
          if (fs.existsSync(srcPath)) {
            await this.copyDirectory(srcPath, dstPath);
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
   * Validate system requirements
   * @tags @IMPL:SYSTEM-VALIDATION-001
   */
  private async validateSystemRequirements(): Promise<void> {
    try {
      // Check Node.js version
      const nodeVersion = process.version;
      const requiredVersion = '18.0.0';
      if (!this.isVersionSatisfied(nodeVersion.slice(1), requiredVersion)) {
        throw new Error(
          `Node.js version ${requiredVersion}+ required, found ${nodeVersion}`
        );
      }

      // Check write permissions on project path (create directory if needed)
      await fs.promises.mkdir(this.config.projectPath, { recursive: true });
      const testPath = path.join(this.config.projectPath, '.test-write');
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
   * Create project directory structure
   * @tags @IMPL:CREATE-DIRECTORIES-001
   */
  private async createProjectDirectories(): Promise<void> {
    try {
      const directories = [
        this.config.projectPath,
        path.join(this.config.projectPath, '.claude'),
        path.join(this.config.projectPath, '.claude', 'logs'),
        path.join(this.config.projectPath, '.moai'),
        path.join(this.config.projectPath, '.moai', 'project'),
        path.join(this.config.projectPath, '.moai', 'specs'),
        path.join(this.config.projectPath, '.moai', 'reports'),
        path.join(this.config.projectPath, '.moai', 'memory'),
        path.join(this.config.projectPath, '.moai', 'indexes'),
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
   * Install MoAI-ADK resources
   * @returns List of installed files
   * @tags @IMPL:INSTALL-RESOURCES-001
   */
  private async installResources(): Promise<string[]> {
    const installedFiles: string[] = [];

    try {
      // Install Claude Code resources
      const claudeFiles = await this.installClaudeResources();
      installedFiles.push(...claudeFiles);

      // Install MoAI resources
      const moaiFiles = await this.installMoaiResources();
      installedFiles.push(...moaiFiles);

      // Install project memory
      const memoryFile = await this.installProjectMemory();
      if (memoryFile) {
        installedFiles.push(memoryFile);
      }

      logger.debug('Resources installed', {
        count: installedFiles.length,
        tag: '@SUCCESS:INSTALL-RESOURCES-001',
      });
      return installedFiles;
    } catch (error) {
      logger.error('Resource installation failed', {
        error,
        tag: '@ERROR:INSTALL-RESOURCES-001',
      });
      throw error;
    }
  }

  /**
   * Generate configuration files
   * @returns List of generated config files
   * @tags @IMPL:GENERATE-CONFIG-001
   */
  private async generateConfigurations(): Promise<string[]> {
    const configFiles: string[] = [];

    try {
      // Generate Claude Code settings only if not already installed from template
      const existingClaudeSettingsPath = path.join(this.config.projectPath, '.claude', 'settings.json');
      if (!fs.existsSync(existingClaudeSettingsPath)) {
        const claudeSettingsPath = await this.createClaudeSettings();
        if (claudeSettingsPath) {
          configFiles.push(claudeSettingsPath);
        }
      } else {
        logger.debug('Claude settings already exists from template, skipping creation', {
          settingsPath: existingClaudeSettingsPath,
          tag: '@DEBUG:CLAUDE-SETTINGS-EXISTS-001',
        });
        configFiles.push(existingClaudeSettingsPath);
      }

      // Generate MoAI configuration
      const moaiConfigPath = await this.createMoaiConfig();
      if (moaiConfigPath) {
        configFiles.push(moaiConfigPath);
      }

      // Generate gitignore if needed
      const gitignorePath = await this.createGitignore();
      if (gitignorePath) {
        configFiles.push(gitignorePath);
      }

      logger.debug('Configuration files generated', {
        count: configFiles.length,
        tag: '@SUCCESS:GENERATE-CONFIG-001',
      });
      return configFiles;
    } catch (error) {
      logger.error('Configuration generation failed', {
        error,
        tag: '@ERROR:GENERATE-CONFIG-001',
      });
      throw error;
    }
  }

  /**
   * Validate installation completeness
   * @tags @IMPL:VALIDATE-INSTALLATION-001
   */
  private async validateInstallation(): Promise<void> {
    try {
      const requiredPaths = [
        path.join(this.config.projectPath, '.claude'),
        path.join(this.config.projectPath, '.moai'),
        path.join(this.config.projectPath, '.claude', 'settings.json'),
        path.join(this.config.projectPath, '.moai', 'config.json'),
      ];

      for (const requiredPath of requiredPaths) {
        if (!fs.existsSync(requiredPath)) {
          throw new Error(`Required path missing: ${requiredPath}`);
        }
      }

      // Validate configuration files are valid JSON
      const settingsPath = path.join(
        this.config.projectPath,
        '.claude',
        'settings.json'
      );
      const configPath = path.join(
        this.config.projectPath,
        '.moai',
        'config.json'
      );

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
   * Finalize installation with post-install tasks
   * @tags @IMPL:FINALIZE-INSTALLATION-001
   */
  private async finalizeInstallation(): Promise<void> {
    try {
      // Set appropriate permissions
      if (process.platform !== 'win32') {
        const scriptsDir = path.join(
          this.config.projectPath,
          '.claude',
          'hooks'
        );
        if (fs.existsSync(scriptsDir)) {
          await this.setExecutablePermissions(scriptsDir);
        }
      }

      // Initialize Git repository if requested
      if (this.config.mode === 'team') {
        await this.initializeGitRepository();
      }

      // Log successful completion
      logger.debug('Installation finalized successfully', {
        projectPath: this.config.projectPath,
        mode: this.config.mode,
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

  // Helper methods for implementation

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

  /**
   * Copy directory recursively
   * @param src Source directory
   * @param dst Destination directory
   * @tags @UTIL:COPY-DIRECTORY-001
   */
  private async copyDirectory(src: string, dst: string): Promise<void> {
    await fs.promises.mkdir(dst, { recursive: true });
    const entries = await fs.promises.readdir(src, { withFileTypes: true });

    for (const entry of entries) {
      const srcPath = path.join(src, entry.name);
      const dstPath = path.join(dst, entry.name);

      if (entry.isDirectory()) {
        await this.copyDirectory(srcPath, dstPath);
      } else {
        await fs.promises.copyFile(srcPath, dstPath);
      }
    }
  }

  /**
   * Install Claude Code resources from templates
   * @returns List of installed files
   * @tags @UTIL:INSTALL-CLAUDE-001
   */
  private async installClaudeResources(): Promise<string[]> {
    const installedFiles: string[] = [];
    const claudeDir = path.join(this.config.projectPath, '.claude');
    const templatesPath = this.getTemplatesPath();
    const claudeTemplatesPath = path.join(templatesPath, '.claude');

    // Create template variables
    const templateVars = this.createTemplateVariables();

    // Copy .claude directory structure from templates
    if (fs.existsSync(claudeTemplatesPath)) {
      const copiedFiles = await this.copyTemplateDirectory(
        claudeTemplatesPath,
        claudeDir,
        templateVars
      );
      installedFiles.push(...copiedFiles);

      // Validate critical Claude settings file
      await this.validateClaudeSettings(claudeDir);
    } else {
      // Fallback: create minimal structure if templates not found
      logger.warn('Claude templates not found, creating minimal structure', {
        templatesPath: claudeTemplatesPath,
        tag: '@WARN:CLAUDE-TEMPLATES-001',
      });
      await this.createMinimalClaudeStructure(claudeDir, installedFiles);
    }

    return installedFiles;
  }

  /**
   * Install MoAI resources from templates
   * @returns List of installed files
   * @tags @UTIL:INSTALL-MOAI-001
   */
  private async installMoaiResources(): Promise<string[]> {
    const installedFiles: string[] = [];
    const moaiDir = path.join(this.config.projectPath, '.moai');
    const templatesPath = this.getTemplatesPath();
    const moaiTemplatesPath = path.join(templatesPath, '.moai');

    // Create template variables
    const templateVars = this.createTemplateVariables();

    // Copy .moai directory structure from templates
    if (fs.existsSync(moaiTemplatesPath)) {
      const copiedFiles = await this.copyTemplateDirectory(
        moaiTemplatesPath,
        moaiDir,
        templateVars
      );
      installedFiles.push(...copiedFiles);
    } else {
      // Fallback: create minimal structure if templates not found
      logger.warn('MoAI templates not found, creating minimal structure', {
        templatesPath: moaiTemplatesPath,
        tag: '@WARN:MOAI-TEMPLATES-001',
      });
      await this.createMinimalMoaiStructure(moaiDir, installedFiles);
    }

    // Initialize JSON-based TAG database
    try {
      await this.initializeTagSystem(moaiDir);
      logger.debug('JSON-based TAG system initialized successfully', {
        tag: '@SUCCESS:TAG-INIT-001',
      });
    } catch (error) {
      logger.warn(`Failed to initialize TAG system: ${error}`, {
        tag: '@WARN:TAG-INIT-001',
      });
    }

    return installedFiles;
  }

  /**
   * Install project memory file from template
   * @returns Path to memory file
   * @tags @UTIL:INSTALL-MEMORY-001
   */
  private async installProjectMemory(): Promise<string | null> {
    try {
      const memoryPath = path.join(this.config.projectPath, 'CLAUDE.md');
      const templatesPath = this.getTemplatesPath();
      const templatePath = path.join(templatesPath, 'CLAUDE.md');

      // Create template variables
      const templateVars = this.createTemplateVariables();

      let memoryContent: string;

      if (fs.existsSync(templatePath)) {
        // Use template file with variable substitution
        const templateContent = await fs.promises.readFile(templatePath, 'utf-8');
        memoryContent = mustache.render(templateContent, templateVars);
      } else {
        // Fallback: use minimal content
        logger.warn('CLAUDE.md template not found, using fallback', {
          templatePath,
          tag: '@WARN:CLAUDE-TEMPLATE-001',
        });
        memoryContent = `# ${this.config.projectName} - MoAI Project

**Spec-First TDD Development Guide**

## Core Philosophy
- **Spec-First**: No code without specification
- **TDD-First**: No implementation without tests
- **GitFlow Support**: Automated Git workflows and Living Document sync

## Development Workflow
\`\`\`bash
/moai:0-project  # Initialize project documents
/moai:1-spec     # Create specifications
/moai:2-build    # TDD implementation
/moai:3-sync     # Document synchronization
\`\`\`

## Project Configuration
- **Mode**: ${this.config.mode}
- **Project**: ${this.config.projectName}
- **Backup**: ${this.config.backupEnabled ? 'Enabled' : 'Disabled'}
`;
      }

      await fs.promises.writeFile(memoryPath, memoryContent);
      return memoryPath;
    } catch (error) {
      logger.error('Failed to create project memory', {
        error,
        tag: '@ERROR:MEMORY-001',
      });
      return null;
    }
  }

  /**
   * Create Claude Code settings
   * @returns Path to settings file
   * @tags @UTIL:CREATE-CLAUDE-SETTINGS-001
   */
  private async createClaudeSettings(): Promise<string | null> {
    try {
      const settingsPath = path.join(
        this.config.projectPath,
        '.claude',
        'settings.json'
      );

      // Check if settings file already exists from template
      if (fs.existsSync(settingsPath)) {
        logger.debug('Settings file already exists from template, skipping creation', {
          settingsPath,
          tag: '@DEBUG:CLAUDE-SETTINGS-EXISTS-002',
        });
        return settingsPath;
      }

      // Fallback: create minimal settings if template was not found
      logger.warn('Creating fallback Claude settings (template not found)', {
        settingsPath,
        tag: '@WARN:CLAUDE-SETTINGS-FALLBACK-001',
      });

      const settings = {
        outputStyle: 'study',
        statusLine: {
          enabled: true,
          format: 'MoAI-ADK TypeScript v{version}',
        },
        agents: {
          'moai/spec-builder': { enabled: true },
          'moai/code-builder': { enabled: true },
          'moai/doc-syncer': { enabled: true },
          'moai/cc-manager': { enabled: true },
          'moai/debug-helper': { enabled: true },
        },
        commands: {
          'moai:0-project': { enabled: true },
          'moai:1-spec': { enabled: true },
          'moai:2-build': { enabled: true },
          'moai:3-sync': { enabled: true },
          'moai:4-debug': { enabled: true },
        },
      };

      await fs.promises.writeFile(
        settingsPath,
        JSON.stringify(settings, null, 2)
      );
      return settingsPath;
    } catch (error) {
      logger.error('Failed to create Claude settings', {
        error,
        tag: '@ERROR:CLAUDE-SETTINGS-001',
      });
      return null;
    }
  }

  /**
   * Create MoAI configuration
   * @returns Path to config file
   * @tags @UTIL:CREATE-MOAI-CONFIG-001
   */
  private async createMoaiConfig(): Promise<string | null> {
    try {
      const configPath = path.join(
        this.config.projectPath,
        '.moai',
        'config.json'
      );
      const config = {
        version: '0.0.1',
        mode: this.config.mode,
        projectName: this.config.projectName,
        features: this.config.additionalFeatures,
        backup: {
          enabled: this.config.backupEnabled,
          retentionDays: 30,
        },
        git: {
          enabled: this.config.mode === 'team',
          autoCommit: true,
          branchPrefix: 'feature/',
        },
      };

      await fs.promises.writeFile(configPath, JSON.stringify(config, null, 2));
      return configPath;
    } catch (error) {
      logger.error('Failed to create MoAI config', {
        error,
        tag: '@ERROR:MOAI-CONFIG-001',
      });
      return null;
    }
  }

  /**
   * Create gitignore file
   * @returns Path to gitignore file
   * @tags @UTIL:CREATE-GITIGNORE-001
   */
  private async createGitignore(): Promise<string | null> {
    if (this.config.mode === 'personal') return null;

    try {
      const gitignorePath = path.join(this.config.projectPath, '.gitignore');
      const gitignoreContent = `# MoAI-ADK Generated .gitignore

# Logs and temporary files
.claude/logs/
.moai/logs/
*.log
*.tmp

# Backup directories
.moai-backup/

# Node.js
node_modules/
npm-debug.log*
yarn-debug.log*
yarn-error.log*

# Environment variables
.env
.env.local
.env.*.local

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db
`;

      await fs.promises.writeFile(gitignorePath, gitignoreContent);
      return gitignorePath;
    } catch (error) {
      logger.error('Failed to create gitignore', {
        error,
        tag: '@ERROR:GITIGNORE-001',
      });
      return null;
    }
  }

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
   * @tags @UTIL:INIT-GIT-001
   */
  private async initializeGitRepository(): Promise<void> {
    const { execSync } = require('child_process');

    try {
      process.chdir(this.config.projectPath);
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

  // Template-related helper methods

  /**
   * Get templates directory path
   * @returns Path to templates directory
   * @tags @UTIL:GET-TEMPLATES-PATH-001
   */
  private getTemplatesPath(): string {
    // Get Node.js executable path to find global node_modules
    const nodeExecPath = process.execPath;
    const nodeBinDir = path.dirname(nodeExecPath);
    const nodeInstallDir = path.dirname(nodeBinDir);

    // Try to find templates directory using multiple strategies
    const possiblePaths = [
      // From project root (development)
      path.join(process.cwd(), 'templates'),
      // From source directory (development)
      path.join(process.cwd(), 'src', '..', 'templates'),
      // From npm package (local node_modules)
      path.join(process.cwd(), 'node_modules', 'moai-adk', 'templates'),
      // From current Node.js installation (nvm, nodenv, etc.)
      path.join(nodeInstallDir, 'lib', 'node_modules', 'moai-adk', 'templates'),
      // From parent directory (various installations)
      path.join(process.cwd(), '..', 'templates'),
      // Common npm global paths
      '/usr/local/lib/node_modules/moai-adk/templates',
      path.join(process.env.HOME || '~', '.npm-global', 'lib', 'node_modules', 'moai-adk', 'templates'),
    ];

    for (const templatePath of possiblePaths) {
      const resolvedPath = path.resolve(templatePath);
      if (fs.existsSync(resolvedPath)) {
        logger.debug('Found templates directory', {
          templatePath: resolvedPath,
          tag: '@DEBUG:TEMPLATES-PATH-001',
        });
        return resolvedPath;
      }
    }

    // Fallback to project root templates
    const fallbackPath = path.join(process.cwd(), 'templates');
    logger.warn('Templates directory not found, using fallback', {
      fallbackPath,
      searchedPaths: possiblePaths.map(p => path.resolve(p)),
      cwd: process.cwd(),
      tag: '@WARN:TEMPLATES-PATH-001',
    });
    return fallbackPath;
  }

  /**
   * Create template variables for Mustache rendering
   * @returns Template variables object
   * @tags @UTIL:CREATE-TEMPLATE-VARS-001
   */
  private createTemplateVariables(): Record<string, any> {
    return {
      PROJECT_NAME: this.config.projectName,
      PROJECT_DESCRIPTION: `A ${this.config.projectName} project built with MoAI-ADK`,
      PROJECT_VERSION: '0.1.0',
      PROJECT_MODE: this.config.mode,
      TIMESTAMP: new Date().toISOString(),
      AUTHOR: 'MoAI Developer',
      LICENSE: 'MIT',
    };
  }

  /**
   * Copy template directory with variable substitution
   * @param srcDir Source template directory
   * @param dstDir Destination directory
   * @param variables Template variables
   * @returns List of copied files
   * @tags @UTIL:COPY-TEMPLATE-DIR-001
   */
  private async copyTemplateDirectory(
    srcDir: string,
    dstDir: string,
    variables: Record<string, any>
  ): Promise<string[]> {
    const copiedFiles: string[] = [];

    try {
      await fs.promises.mkdir(dstDir, { recursive: true });
      const entries = await fs.promises.readdir(srcDir, { withFileTypes: true });

      for (const entry of entries) {
        const srcPath = path.join(srcDir, entry.name);
        const dstPath = path.join(dstDir, entry.name);

        if (entry.isDirectory()) {
          // Recursively copy subdirectories
          const subFiles = await this.copyTemplateDirectory(srcPath, dstPath, variables);
          copiedFiles.push(...subFiles);
        } else {
          // Copy and process file
          await this.copyTemplateFile(srcPath, dstPath, variables);
          copiedFiles.push(dstPath);
        }
      }

      return copiedFiles;
    } catch (error) {
      logger.error('Failed to copy template directory', {
        error,
        srcDir,
        dstDir,
        tag: '@ERROR:COPY-TEMPLATE-DIR-001',
      });
      throw error;
    }
  }

  /**
   * Copy and process template file with variable substitution
   * @param srcPath Source file path
   * @param dstPath Destination file path
   * @param variables Template variables
   * @tags @UTIL:COPY-TEMPLATE-FILE-001
   */
  private async copyTemplateFile(
    srcPath: string,
    dstPath: string,
    variables: Record<string, any>
  ): Promise<void> {
    try {
      await fs.promises.mkdir(path.dirname(dstPath), { recursive: true });

      const content = await fs.promises.readFile(srcPath, 'utf-8');

      // Apply Mustache template processing for text files
      const fileExt = path.extname(srcPath).toLowerCase();
      const isTextFile = ['.md', '.json', '.js', '.ts', '.py', '.txt', '.yml', '.yaml'].includes(fileExt);

      let processedContent: string;
      if (isTextFile) {
        processedContent = mustache.render(content, variables);
      } else {
        processedContent = content;
      }

      await fs.promises.writeFile(dstPath, processedContent);

      // Preserve executable permissions for script files
      if (['.py', '.sh', '.js'].includes(fileExt)) {
        try {
          await fs.promises.chmod(dstPath, 0o755);
        } catch (chmodError) {
          // Ignore chmod errors on Windows
          if (process.platform !== 'win32') {
            logger.warn('Failed to set executable permissions', {
              dstPath,
              error: chmodError,
              tag: '@WARN:CHMOD-001',
            });
          }
        }
      }

      logger.debug('Template file copied and processed', {
        srcPath,
        dstPath,
        isTextFile,
        tag: '@DEBUG:COPY-TEMPLATE-FILE-001',
      });
    } catch (error) {
      logger.error('Failed to copy template file', {
        error,
        srcPath,
        dstPath,
        tag: '@ERROR:COPY-TEMPLATE-FILE-001',
      });
      throw error;
    }
  }

  /**
   * Initialize JSON-based TAG system
   * @param moaiDir MoAI directory path
   * @tags @UTIL:INIT-TAG-SYSTEM-001
   */
  private async initializeTagSystem(moaiDir: string): Promise<void> {
    try {
      const indexesDir = path.join(moaiDir, 'indexes');
      await fs.promises.mkdir(indexesDir, { recursive: true });

      // Create initial tags.json
      const tagsPath = path.join(indexesDir, 'tags.json');
      const initialTags = {
        version: '1.0.0',
        project: this.config.projectName,
        created: new Date().toISOString(),
        tags: {},
        chains: [],
        categories: {
          primary: ['REQ', 'DESIGN', 'TASK', 'TEST'],
          implementation: ['FEATURE', 'API', 'UI', 'DATA'],
          quality: ['PERF', 'SEC', 'DOCS', 'TAG'],
          project: ['VISION', 'STRUCT', 'TECH', 'ADR'],
        },
      };

      await fs.promises.writeFile(tagsPath, JSON.stringify(initialTags, null, 2));

      // Create meta.json
      const metaPath = path.join(indexesDir, 'meta.json');
      const metaData = {
        project: this.config.projectName,
        created: new Date().toISOString(),
        lastSync: new Date().toISOString(),
        totalTags: 0,
        totalFiles: 0,
      };

      await fs.promises.writeFile(metaPath, JSON.stringify(metaData, null, 2));

      // Create cache directory and summary
      const cacheDir = path.join(indexesDir, 'cache');
      await fs.promises.mkdir(cacheDir, { recursive: true });

      const summaryPath = path.join(cacheDir, 'summary.json');
      const summaryData = {
        lastUpdate: new Date().toISOString(),
        totalTags: 0,
        recentTags: [],
        orphanTags: [],
        brokenLinks: [],
      };

      await fs.promises.writeFile(summaryPath, JSON.stringify(summaryData, null, 2));

      logger.debug('TAG system initialized successfully', {
        tagsPath,
        metaPath,
        summaryPath,
        tag: '@SUCCESS:TAG-SYSTEM-INIT-001',
      });
    } catch (error) {
      logger.error('Failed to initialize TAG system', {
        error,
        moaiDir,
        tag: '@ERROR:TAG-SYSTEM-INIT-001',
      });
      throw error;
    }
  }

  /**
   * Create minimal Claude structure (fallback)
   * @param claudeDir Claude directory path
   * @param installedFiles Array to track installed files
   * @tags @UTIL:CREATE-MINIMAL-CLAUDE-001
   */
  private async createMinimalClaudeStructure(
    claudeDir: string,
    installedFiles: string[]
  ): Promise<void> {
    const minimalStructure = {
      'settings.json': JSON.stringify(
        {
          outputStyle: 'study',
          agents: {
            'moai/spec-builder': { enabled: true },
            'moai/code-builder': { enabled: true },
            'moai/doc-syncer': { enabled: true },
          },
        },
        null,
        2
      ),
      'agents/moai/spec-builder.md':
        '# SPEC Builder Agent\n\nBuilds SPEC documents using EARS methodology.',
      'commands/moai/1-spec.md':
        '# SPEC Command\n\nCreates new SPEC documents.',
      'hooks/moai/steering_guard.py':
        '# Steering Guard Hook\n\n# Validates development guidelines',
    };

    for (const [relativePath, content] of Object.entries(minimalStructure)) {
      const fullPath = path.join(claudeDir, relativePath);
      await fs.promises.mkdir(path.dirname(fullPath), { recursive: true });
      await fs.promises.writeFile(fullPath, content);
      installedFiles.push(fullPath);
    }
  }

  /**
   * Create minimal MoAI structure (fallback)
   * @param moaiDir MoAI directory path
   * @param installedFiles Array to track installed files
   * @tags @UTIL:CREATE-MINIMAL-MOAI-001
   */
  private async createMinimalMoaiStructure(
    moaiDir: string,
    installedFiles: string[]
  ): Promise<void> {
    const minimalStructure = {
      'config.json': JSON.stringify(
        {
          version: '0.0.1',
          mode: this.config.mode,
          projectName: this.config.projectName,
        },
        null,
        2
      ),
      'memory/development-guide.md':
        '# Development Guide\n\nTRUST 5 principles and guidelines.',
    };

    for (const [relativePath, content] of Object.entries(minimalStructure)) {
      const fullPath = path.join(moaiDir, relativePath);
      await fs.promises.mkdir(path.dirname(fullPath), { recursive: true });
      await fs.promises.writeFile(fullPath, content);
      installedFiles.push(fullPath);
    }
  }

  /**
   * Validate Claude settings file from template
   * @param claudeDir Claude directory path
   * @tags @UTIL:VALIDATE-CLAUDE-SETTINGS-001
   */
  private async validateClaudeSettings(claudeDir: string): Promise<void> {
    try {
      const settingsPath = path.join(claudeDir, 'settings.json');

      if (!fs.existsSync(settingsPath)) {
        logger.warn('Claude settings.json not found after template copy', {
          settingsPath,
          tag: '@WARN:SETTINGS-NOT-FOUND-001',
        });
        return;
      }

      // Parse and validate settings content
      const settingsContent = await fs.promises.readFile(settingsPath, 'utf-8');
      const settings = JSON.parse(settingsContent);

      // Check for critical sections
      const hasEnv = !!settings.env;
      const hasHooks = !!settings.hooks;
      const hasPermissions = !!settings.permissions;
      const hasAgents = !!settings.agents;
      const hasCommands = !!settings.commands;

      // Log validation results
      logger.debug('Claude settings validation results', {
        settingsPath,
        hasEnv,
        hasHooks,
        hasPermissions,
        hasAgents,
        hasCommands,
        fileSize: settingsContent.length,
        tag: '@DEBUG:SETTINGS-VALIDATION-001',
      });

      // Warn if critical sections are missing
      if (!hasEnv || !hasHooks || !hasPermissions) {
        logger.warn('Incomplete Claude settings detected', {
          settingsPath,
          missingEnv: !hasEnv,
          missingHooks: !hasHooks,
          missingPermissions: !hasPermissions,
          tag: '@WARN:SETTINGS-INCOMPLETE-001',
        });
      } else {
        logger.debug('Claude settings validation successful', {
          settingsPath,
          tag: '@SUCCESS:SETTINGS-VALIDATION-001',
        });
      }
    } catch (error) {
      logger.error('Failed to validate Claude settings', {
        error,
        claudeDir,
        tag: '@ERROR:SETTINGS-VALIDATION-001',
      });
      // Don't throw - validation failure shouldn't stop installation
    }
  }
}
