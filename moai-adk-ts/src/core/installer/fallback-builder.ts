/**
 * @file Fallback Structure Builder
 * @author MoAI Team
 * @tags @FEATURE:FALLBACK-BUILDER-001 @REQ:INSTALL-SYSTEM-012
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { logger } from '@/utils/logger';
import { InstallationError, getErrorMessage } from '@/utils/errors';
import type { InstallationConfig } from './types';

/**
 * Builds fallback directory structures and configuration files
 * @tags @FEATURE:FALLBACK-BUILDER-001
 */
export class FallbackBuilder {
  /**
   * Initialize JSON-based TAG system
   * @param moaiDir MoAI directory path
   * @param config Installation configuration
   * @tags @API:INIT-TAG-SYSTEM-001
   */
  async initializeTagSystem(
    moaiDir: string,
    config: InstallationConfig
  ): Promise<void> {
    try {
      const indexesDir = path.join(moaiDir, 'indexes');
      await fs.promises.mkdir(indexesDir, { recursive: true });

      const metaPath = path.join(indexesDir, 'meta.json');
      const metaData = {
        project: config.projectName,
        created: new Date().toISOString(),
        lastSync: new Date().toISOString(),
        totalTags: 0,
        totalFiles: 0,
      };

      await fs.promises.writeFile(metaPath, JSON.stringify(metaData, null, 2));

      const cacheDir = path.join(indexesDir, 'cache');
      await fs.promises.mkdir(cacheDir, { recursive: true });

      logger.debug('TAG system initialized successfully (code-scan based)', {
        metaPath,
        cacheDir,
        tag: '@SUCCESS:TAG-SYSTEM-INIT-001',
      });
    } catch (error) {
      const errorMessage = getErrorMessage(error);
      logger.error('Failed to initialize TAG system', {
        error: errorMessage,
        moaiDir,
        tag: '@ERROR:TAG-SYSTEM-INIT-001',
      });
      throw new InstallationError('Failed to initialize TAG system', {
        error: error instanceof Error ? error : undefined,
        errorMessage,
        context: { moaiDir },
      });
    }
  }

  /**
   * Create minimal Claude structure (fallback)
   * @param claudeDir Claude directory path
   * @param installedFiles Array to track installed files
   * @tags @API:CREATE-MINIMAL-CLAUDE-001
   */
  async createMinimalClaudeStructure(
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

    logger.info('Minimal Claude structure created', {
      fileCount: Object.keys(minimalStructure).length,
      tag: '@SUCCESS:MINIMAL-CLAUDE-001',
    });
  }

  /**
   * Create minimal MoAI structure (fallback)
   * @param moaiDir MoAI directory path
   * @param installedFiles Array to track installed files
   * @tags @API:CREATE-MINIMAL-MOAI-001
   */
  async createMinimalMoaiStructure(
    moaiDir: string,
    installedFiles: string[]
  ): Promise<void> {
    const minimalStructure = {
      'config.json': JSON.stringify(
        {
          version: '0.0.1',
          mode: 'personal',
          projectName: 'default',
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

    logger.info('Minimal MoAI structure created', {
      fileCount: Object.keys(minimalStructure).length,
      tag: '@SUCCESS:MINIMAL-MOAI-001',
    });
  }

  /**
   * Create fallback memory content
   * @param config Installation configuration
   * @returns Fallback memory content
   * @tags @API:CREATE-FALLBACK-MEMORY-001
   */
  createFallbackMemoryContent(config: InstallationConfig): string {
    return `# ${config.projectName} - MoAI Project

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
- **Mode**: ${config.mode}
- **Project**: ${config.projectName}
- **Backup**: ${config.backupEnabled ? 'Enabled' : 'Disabled'}
`;
  }

  /**
   * Validate Claude settings file from template
   * @param claudeDir Claude directory path
   * @tags @API:VALIDATE-CLAUDE-SETTINGS-001
   */
  async validateClaudeSettings(claudeDir: string): Promise<void> {
    try {
      const settingsPath = path.join(claudeDir, 'settings.json');

      if (!fs.existsSync(settingsPath)) {
        logger.warn('Claude settings.json not found after template copy', {
          settingsPath,
          tag: '@WARN:SETTINGS-NOT-FOUND-001',
        });
        return;
      }

      const settingsContent = await fs.promises.readFile(settingsPath, 'utf-8');
      const settings = JSON.parse(settingsContent);

      const hasEnv = !!settings.env;
      const hasHooks = !!settings.hooks;
      const hasPermissions = !!settings.permissions;
      const hasAgents = !!settings.agents;
      const hasCommands = !!settings.commands;

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
      const errorMessage = getErrorMessage(error);
      logger.error('Failed to validate Claude settings', {
        error: errorMessage,
        claudeDir,
        tag: '@ERROR:SETTINGS-VALIDATION-001',
      });
    }
  }

  /**
   * Create Claude Code settings (fallback)
   */
  async createClaudeSettings(
    config: InstallationConfig
  ): Promise<string | null> {
    const settingsPath = path.join(config.projectPath, '.claude', 'settings.json');
    if (fs.existsSync(settingsPath)) return settingsPath;

    return this.writeJsonFile(settingsPath, {
      outputStyle: 'study',
      statusLine: { enabled: true, format: 'MoAI-ADK TypeScript v{version}' },
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
    }, '@ERROR:CLAUDE-SETTINGS-001');
  }

  /**
   * Create MoAI configuration
   */
  async createMoaiConfig(
    config: InstallationConfig
  ): Promise<string | null> {
    const configPath = path.join(config.projectPath, '.moai', 'config.json');

    return this.writeJsonFile(configPath, {
      version: '0.0.1',
      mode: config.mode,
      projectName: config.projectName,
      features: config.additionalFeatures,
      backup: { enabled: config.backupEnabled, retentionDays: 30 },
      git: { enabled: config.mode === 'team', autoCommit: true, branchPrefix: 'feature/' },
    }, '@ERROR:MOAI-CONFIG-001');
  }

  /**
   * Write JSON file with error handling
   */
  private async writeJsonFile(
    filePath: string,
    content: unknown,
    errorTag: string
  ): Promise<string | null> {
    try {
      await fs.promises.writeFile(filePath, JSON.stringify(content, null, 2));
      return filePath;
    } catch (error) {
      const errorMessage = getErrorMessage(error);
      logger.error(`Failed to write ${filePath}`, {
        error: errorMessage,
        tag: errorTag,
      });
      return null;
    }
  }
}