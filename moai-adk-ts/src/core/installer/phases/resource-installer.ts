/**
 * @file Resource Installation Logic
 * @author MoAI Team
 * @tags @FEATURE:RESOURCE-INSTALLER-001 @REQ:INSTALL-SYSTEM-012
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { logger } from '@/utils/logger';
import type {
  InstallationConfig,
  InstallationContext,
} from '../types';

/**
 * Handles installation of various resources
 * @tags @FEATURE:RESOURCE-INSTALLER-001
 */
export class ResourceInstaller {
  constructor(
    private readonly config: InstallationConfig,
    private readonly context: InstallationContext
  ) {}

  /**
   * Install all required resources
   * @returns List of created files
   * @tags @API:INSTALL-RESOURCES-001
   */
  async installResources(): Promise<string[]> {
    const filesCreated: string[] = [];

    // Install Claude resources
    const claudeFiles = await this.installClaudeResources();
    filesCreated.push(...claudeFiles);

    // Install MoAI resources
    const moaiFiles = await this.installMoaiResources();
    filesCreated.push(...moaiFiles);

    // Install project memory if requested
    const memoryFile = await this.installProjectMemory();
    if (memoryFile) {
      filesCreated.push(memoryFile);
    }

    logger.info(`Installed ${filesCreated.length} resource files`, {
      tag: '@SUCCESS:RESOURCES-INSTALLED-001',
    });

    return filesCreated;
  }

  /**
   * Install Claude Code related resources
   * @returns List of installed Claude files
   * @tags @API:INSTALL-CLAUDE-001
   */
  async installClaudeResources(): Promise<string[]> {
    const filesCreated: string[] = [];
    const claudeDir = path.join(this.config.projectPath, '.claude');

    if (!fs.existsSync(claudeDir)) {
      fs.mkdirSync(claudeDir, { recursive: true });
      logger.debug('Created Claude directory', {
        path: claudeDir,
        tag: '@DIR:CLAUDE-001',
      });
    }

    const templatesPath = this.getTemplatesPath();
    const claudeTemplatesPath = path.join(templatesPath, '.claude');

    if (fs.existsSync(claudeTemplatesPath)) {
      await this.copyTemplateDirectory(claudeTemplatesPath, claudeDir);

      // Get list of copied files
      const copiedFiles = await this.getDirectoryFiles(claudeDir);
      filesCreated.push(...copiedFiles);
    } else {
      // Create minimal Claude structure
      const minimalFiles = await this.createMinimalClaudeStructure(claudeDir);
      filesCreated.push(...minimalFiles);
    }

    logger.info(`Installed ${filesCreated.length} Claude files`, {
      claudeDir,
      tag: '@SUCCESS:CLAUDE-INSTALLED-001',
    });

    return filesCreated;
  }

  /**
   * Install MoAI specific resources
   * @returns List of installed MoAI files
   * @tags @API:INSTALL-MOAI-001
   */
  async installMoaiResources(): Promise<string[]> {
    const filesCreated: string[] = [];
    const moaiDir = path.join(this.config.projectPath, '.moai');

    if (!fs.existsSync(moaiDir)) {
      fs.mkdirSync(moaiDir, { recursive: true });
      logger.debug('Created MoAI directory', {
        path: moaiDir,
        tag: '@DIR:MOAI-001',
      });
    }

    const templatesPath = this.getTemplatesPath();
    const moaiTemplatesPath = path.join(templatesPath, '.moai');

    if (fs.existsSync(moaiTemplatesPath)) {
      await this.copyTemplateDirectory(moaiTemplatesPath, moaiDir);

      // Get list of copied files
      const copiedFiles = await this.getDirectoryFiles(moaiDir);
      filesCreated.push(...copiedFiles);
    } else {
      // Create minimal MoAI structure
      const minimalFiles = await this.createMinimalMoaiStructure(moaiDir);
      filesCreated.push(...minimalFiles);
    }

    // Initialize TAG system
    await this.initializeTagSystem(moaiDir);

    // Set permissions for scripts
    const scriptsDir = path.join(moaiDir, 'scripts');
    if (fs.existsSync(scriptsDir)) {
      await this.setExecutablePermissions(scriptsDir);
    }

    logger.info(`Installed ${filesCreated.length} MoAI files`, {
      moaiDir,
      tag: '@SUCCESS:MOAI-INSTALLED-001',
    });

    return filesCreated;
  }

  /**
   * Install project memory documentation
   * @returns Path to created memory file or null
   * @tags @API:INSTALL-MEMORY-001
   */
  async installProjectMemory(): Promise<string | null> {
    if (!this.config.includeProjectMemory) {
      return null;
    }

    const memoryDir = path.join(this.config.projectPath, '.moai', 'memory');
    const memoryFile = path.join(memoryDir, 'development-guide.md');

    if (!fs.existsSync(memoryDir)) {
      fs.mkdirSync(memoryDir, { recursive: true });
    }

    const templatesPath = this.getTemplatesPath();
    const memoryTemplate = path.join(
      templatesPath,
      '.moai',
      'memory',
      'development-guide.md'
    );

    if (fs.existsSync(memoryTemplate)) {
      const content = fs.readFileSync(memoryTemplate, 'utf8');
      const variables = this.createTemplateVariables();
      const processedContent = this.renderTemplate(content, variables);

      fs.writeFileSync(memoryFile, processedContent, 'utf8');

      logger.info('Installed project memory', {
        memoryFile,
        tag: '@SUCCESS:MEMORY-INSTALLED-001',
      });
    } else {
      // Create default memory structure
      const defaultContent = this.createDefaultMemoryContent();
      fs.writeFileSync(memoryFile, defaultContent, 'utf8');
    }

    return memoryFile;
  }

  private getTemplatesPath(): string {
    // This should use the actual templates path resolution logic
    return path.join(__dirname, '../../../../templates');
  }

  private async copyTemplateDirectory(src: string, dst: string): Promise<void> {
    if (!fs.existsSync(src)) {
      logger.warn(`Template directory not found: ${src}`);
      return;
    }

    const entries = fs.readdirSync(src, { withFileTypes: true });

    for (const entry of entries) {
      const srcPath = path.join(src, entry.name);
      const dstPath = path.join(dst, entry.name);

      if (entry.isDirectory()) {
        if (!fs.existsSync(dstPath)) {
          fs.mkdirSync(dstPath, { recursive: true });
        }
        await this.copyTemplateDirectory(srcPath, dstPath);
      } else if (entry.isFile()) {
        await this.copyTemplateFile(srcPath, dstPath);
      }
    }
  }

  private async copyTemplateFile(src: string, dst: string): Promise<void> {
    const content = fs.readFileSync(src, 'utf8');
    const variables = this.createTemplateVariables();
    const processedContent = this.renderTemplate(content, variables);

    fs.writeFileSync(dst, processedContent, 'utf8');
  }

  private async getDirectoryFiles(dir: string): Promise<string[]> {
    const files: string[] = [];

    const entries = fs.readdirSync(dir, { withFileTypes: true });
    for (const entry of entries) {
      const fullPath = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        const subFiles = await this.getDirectoryFiles(fullPath);
        files.push(...subFiles);
      } else {
        files.push(fullPath);
      }
    }

    return files;
  }

  private async createMinimalClaudeStructure(claudeDir: string): Promise<string[]> {
    const files: string[] = [];

    // Create basic settings file
    const settingsFile = path.join(claudeDir, 'settings.json');
    const settingsContent = {
      project: {
        name: this.config.projectName,
        description: `A ${this.config.projectName} project built with MoAI-ADK`,
        version: '1.0.0',
        mode: this.config.mode,
      },
    };

    fs.writeFileSync(settingsFile, JSON.stringify(settingsContent, null, 2), 'utf8');
    files.push(settingsFile);

    return files;
  }

  private async createMinimalMoaiStructure(moaiDir: string): Promise<string[]> {
    const files: string[] = [];

    // Create basic config file
    const configFile = path.join(moaiDir, 'config.json');
    const configContent = {
      project: {
        name: this.config.projectName,
        version: '1.0.0',
        mode: this.config.mode,
        language: 'typescript',
      },
      features: {
        tagSystem: true,
        livingDocs: true,
        tddWorkflow: true,
      },
    };

    fs.writeFileSync(configFile, JSON.stringify(configContent, null, 2), 'utf8');
    files.push(configFile);

    return files;
  }

  private async initializeTagSystem(moaiDir: string): Promise<void> {
    const indexesDir = path.join(moaiDir, 'indexes');
    if (!fs.existsSync(indexesDir)) {
      fs.mkdirSync(indexesDir, { recursive: true });
    }

    // Create initial tags.json
    const tagsFile = path.join(indexesDir, 'tags.json');
    const initialTags = {
      version: '1.0.0',
      lastUpdated: new Date().toISOString(),
      tags: [],
      chains: [],
      statistics: {
        totalTags: 0,
        totalChains: 0,
        coverage: 0,
      },
    };

    fs.writeFileSync(tagsFile, JSON.stringify(initialTags, null, 2), 'utf8');
  }

  private async setExecutablePermissions(scriptsDir: string): Promise<void> {
    const entries = fs.readdirSync(scriptsDir, { withFileTypes: true });

    for (const entry of entries) {
      if (entry.isFile() && (entry.name.endsWith('.sh') || entry.name.endsWith('.ts'))) {
        const scriptPath = path.join(scriptsDir, entry.name);
        try {
          fs.chmodSync(scriptPath, 0o755);
          logger.debug(`Set executable permission for: ${scriptPath}`);
        } catch (error) {
          logger.warn(`Failed to set permission for: ${scriptPath}`, error);
        }
      }
    }
  }

  private createTemplateVariables(): Record<string, any> {
    return {
      project_name: this.config.projectName,
      project_description: `A ${this.config.projectName} project built with MoAI-ADK`,
      project_version: '1.0.0',
      mode: this.config.mode,
      timestamp: new Date().toISOString(),
    };
  }

  private renderTemplate(template: string, variables: Record<string, any>): string {
    try {
      return template.replace(/\{\{(\w+)\}\}/g, (match, key) => {
        return variables[key] || match;
      });
    } catch (error) {
      logger.warn('Template rendering failed, using original content', error);
      return template;
    }
  }

  private createDefaultMemoryContent(): string {
    return `# ${this.config.projectName} Development Guide

> "No SPEC, no code. No tests, no implementation."

This development guide follows MoAI-ADK SPEC-First TDD principles.

## Project Information

- **Name**: ${this.config.projectName}
- **Version**: 1.0.0
- **Mode**: ${this.config.mode}
- **Created**: ${new Date().toISOString()}

## Development Workflow

1. **SPEC Creation** (\`/moai:1-spec\`) → No code without specification
2. **TDD Implementation** (\`/moai:2-build\`) → No implementation without tests
3. **Documentation Sync** (\`/moai:3-sync\`) → No completion without traceability

## TRUST 5 Principles

- **T**est First: Write tests before implementation
- **R**eadable: Code should be self-documenting
- **U**nified: Consistent patterns and architecture
- **S**ecured: Security by design
- **T**rackable: Complete traceability with @TAG system

## Getting Started

Run \`moai doctor\` to verify your development environment setup.
`;
  }
}