// @CODE:INST-004 |
// Related: @CODE:INST-004:API, @CODE:INST-RES-001

/**
 * @file Resource installation manager
 * @author MoAI Team
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import mustache from 'mustache';
import { logger } from '@/utils/winston-logger';
import { FallbackBuilder } from './fallback-builder';
import type { TemplateProcessor } from './template-processor';
import type { InstallationConfig } from './types';

/**
 * Default gitignore template for team mode
 * @tags CONST:GITIGNORE-TEMPLATE-001
 */
const DEFAULT_GITIGNORE = `# MoAI-ADK Generated .gitignore

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

/**
 * Handles resource installation and configuration file generation
 * @tags @CODE:RESOURCE-INSTALLER-001
 */
export class ResourceInstaller {
  private readonly templateProcessor: TemplateProcessor;
  private readonly fallbackBuilder: FallbackBuilder;

  constructor(templateProcessor: TemplateProcessor) {
    this.templateProcessor = templateProcessor;
    this.fallbackBuilder = new FallbackBuilder();
  }

  /**
   * Install Claude Code resources from templates
   * @param config Installation configuration
   * @returns List of installed files
   * @tags @CODE:INSTALL-CLAUDE-001:API
   */
  async installClaudeResources(config: InstallationConfig): Promise<string[]> {
    const installedFiles: string[] = [];
    const claudeDir = path.join(config.projectPath, '.claude');
    const templatesPath = this.templateProcessor.getTemplatesPath();
    const claudeTemplatesPath = path.join(templatesPath, '.claude');

    const templateVars = this.templateProcessor.createTemplateVariables(config);

    if (fs.existsSync(claudeTemplatesPath)) {
      // CRITICAL: Exclude documentation directories that contain Mustache syntax examples
      // These files document the template system and should not be processed as templates
      const copiedFiles = await this.templateProcessor.copyTemplateDirectory(
        claudeTemplatesPath,
        claudeDir,
        templateVars,
        {
          excludePaths: ['commands', 'agents', 'output-styles'], // Documentation with {{syntax}} examples
        }
      );
      installedFiles.push(...copiedFiles);

      // Copy documentation directories without template processing
      // These files contain Mustache syntax examples (e.g., `{{PROJECT_NAME}}`) that should not be interpolated
      const docDirs = ['commands', 'agents', 'output-styles'];
      for (const docDir of docDirs) {
        const srcDocDir = path.join(claudeTemplatesPath, docDir);
        const dstDocDir = path.join(claudeDir, docDir);
        if (fs.existsSync(srcDocDir)) {
          await this.templateProcessor.copyDirectory(srcDocDir, dstDocDir);
          logger.debug(`Copied documentation directory without template processing: ${docDir}`, {
            tag: 'DEBUG:DOC-COPY-001',
          });
        }
      }

      await this.fallbackBuilder.validateClaudeSettings(claudeDir);
    } else {
      logger.warn('Claude templates not found, creating minimal structure', {
        templatesPath: claudeTemplatesPath,
        tag: 'WARN:CLAUDE-TEMPLATES-001',
      });
      await this.fallbackBuilder.createMinimalClaudeStructure(
        claudeDir,
        installedFiles
      );
    }

    return installedFiles;
  }

  /**
   * Install MoAI resources from templates
   * @param config Installation configuration
   * @returns List of installed files
   * @tags @CODE:INSTALL-MOAI-001:API
   *
   * **User Data Protection**:
   * - `.moai/specs/` directory is NEVER touched (user's SPEC files)
   * - `.moai/reports/` directory is NEVER touched (user's sync reports)
   */
  async installMoaiResources(config: InstallationConfig): Promise<string[]> {
    const installedFiles: string[] = [];
    const moaiDir = path.join(config.projectPath, '.moai');
    const templatesPath = this.templateProcessor.getTemplatesPath();
    const moaiTemplatesPath = path.join(templatesPath, '.moai');

    const templateVars = this.templateProcessor.createTemplateVariables(config);

    if (fs.existsSync(moaiTemplatesPath)) {
      // CRITICAL: Exclude user data directories to prevent accidental overwrite
      const copiedFiles = await this.templateProcessor.copyTemplateDirectory(
        moaiTemplatesPath,
        moaiDir,
        templateVars,
        {
          excludePaths: ['specs', 'reports'], // User SPEC files and sync reports
        }
      );
      installedFiles.push(...copiedFiles);

      logger.info('User data directories protected during installation', {
        excluded: ['specs', 'reports'],
        tag: 'INFO:USER-DATA-PROTECTED-001',
      });
    } else {
      logger.warn('MoAI templates not found, creating minimal structure', {
        templatesPath: moaiTemplatesPath,
        tag: 'WARN:MOAI-TEMPLATES-001',
      });
      await this.fallbackBuilder.createMinimalMoaiStructure(
        moaiDir,
        installedFiles
      );
    }

    try {
      await this.fallbackBuilder.initializeTagSystem(moaiDir, config);
      logger.debug('JSON-based TAG system initialized successfully', {
        tag: 'SUCCESS:TAG-INIT-001',
      });
    } catch (error) {
      logger.warn(`Failed to initialize TAG system: ${error}`, {
        tag: 'WARN:TAG-INIT-001',
      });
    }

    return installedFiles;
  }

  /**
   * Install project memory file from template
   * @param config Installation configuration
   * @returns Path to memory file
   * @tags @CODE:INSTALL-MEMORY-001:API
   */
  async installProjectMemory(
    config: InstallationConfig
  ): Promise<string | null> {
    try {
      const memoryPath = path.join(config.projectPath, 'CLAUDE.md');
      const templatesPath = this.templateProcessor.getTemplatesPath();
      const templatePath = path.join(templatesPath, 'CLAUDE.md');

      const templateVars =
        this.templateProcessor.createTemplateVariables(config);

      let memoryContent: string;

      if (fs.existsSync(templatePath)) {
        const templateContent = await fs.promises.readFile(
          templatePath,
          'utf-8'
        );
        memoryContent = mustache.render(templateContent, templateVars);
      } else {
        logger.warn('CLAUDE.md template not found, using fallback', {
          templatePath,
          tag: 'WARN:CLAUDE-TEMPLATE-001',
        });
        memoryContent =
          this.fallbackBuilder.createFallbackMemoryContent(config);
      }

      await fs.promises.writeFile(memoryPath, memoryContent);
      return memoryPath;
    } catch (error) {
      logger.error('Failed to create project memory', {
        error,
        tag: 'ERROR:MEMORY-001',
      });
      return null;
    }
  }

  /**
   * Create Claude Code settings (delegates to FallbackBuilder)
   * @param config Installation configuration
   * @returns Path to settings file
   * @tags @CODE:CREATE-CLAUDE-SETTINGS-001:API
   */
  async createClaudeSettings(
    config: InstallationConfig
  ): Promise<string | null> {
    return this.fallbackBuilder.createClaudeSettings(config);
  }

  /**
   * Create MoAI configuration (delegates to FallbackBuilder)
   * @param config Installation configuration
   * @returns Path to config file
   * @tags @CODE:CREATE-MOAI-CONFIG-001:API
   */
  async createMoaiConfig(config: InstallationConfig): Promise<string | null> {
    return this.fallbackBuilder.createMoaiConfig(config);
  }

  /**
   * Create gitignore file
   * @param config Installation configuration
   * @returns Path to gitignore file
   * @tags @CODE:CREATE-GITIGNORE-001:API
   */
  async createGitignore(config: InstallationConfig): Promise<string | null> {
    if (config.mode === 'personal') return null;

    try {
      const gitignorePath = path.join(config.projectPath, '.gitignore');
      await fs.promises.writeFile(gitignorePath, DEFAULT_GITIGNORE);
      return gitignorePath;
    } catch (error) {
      logger.error('Failed to create gitignore', {
        error,
        tag: 'ERROR:GITIGNORE-001',
      });
      return null;
    }
  }
}
