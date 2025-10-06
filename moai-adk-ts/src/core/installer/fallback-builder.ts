/**
 * @file Fallback Structure Builder
 * @author MoAI Team
 * @tags @CODE:FALLBACK-BUILDER-001 @SPEC:INSTALL-SYSTEM-012
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { getErrorMessage } from '@/utils/errors';
import { logger } from '@/utils/winston-logger';
import type { InstallationConfig } from './types';

/**
 * Builds fallback directory structures and configuration files
 * @tags @CODE:FALLBACK-BUILDER-001
 */
export class FallbackBuilder {
  /**
   * Initialize JSON-based TAG system
   * @param moaiDir MoAI directory path
   * @param config Installation configuration
   * @tags @CODE:INIT-TAG-SYSTEM-001:API
   */
  async initializeTagSystem(
    moaiDir: string,
    config: InstallationConfig
  ): Promise<void> {
    // NOTE: [v0.0.1] TAG 시스템 철학 - CODE-FIRST
    // - 이전: meta.json, indexes, cache 디렉토리 생성
    // - 현재: 코드 직접 스캔 (rg/grep) 기반 실시간 검증
    // - 이유: 단일 진실 소스(코드)로 동기화 문제 해결
    // - 모든 TAG 정보는 소스코드에만 존재
    // - indexes, cache, meta.json 불필요

    logger.debug('TAG system initialized (CODE-FIRST mode - no cache files)', {
      projectName: config.projectName,
      moaiDir,
      tag: 'SUCCESS:TAG-SYSTEM-INIT-001',
    });
  }

  /**
   * Create minimal Claude structure (fallback)
   * @param claudeDir Claude directory path
   * @param installedFiles Array to track installed files
   * @tags @CODE:CREATE-MINIMAL-CLAUDE-001:API
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
      'agents/alfred/spec-builder.md':
        '# SPEC Builder Agent\n\nBuilds SPEC documents using EARS methodology.',
      'commands/alfred/1-spec.md':
        '# SPEC Command\n\nCreates new SPEC documents.',
    };

    for (const [relativePath, content] of Object.entries(minimalStructure)) {
      const fullPath = path.join(claudeDir, relativePath);
      await fs.promises.mkdir(path.dirname(fullPath), { recursive: true });
      await fs.promises.writeFile(fullPath, content);
      installedFiles.push(fullPath);
    }

    logger.info('Minimal Claude structure created', {
      fileCount: Object.keys(minimalStructure).length,
      tag: 'SUCCESS:MINIMAL-CLAUDE-001',
    });
  }

  /**
   * Create minimal MoAI structure (fallback)
   * @param moaiDir MoAI directory path
   * @param installedFiles Array to track installed files
   * @tags @CODE:CREATE-MINIMAL-MOAI-001:API
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
      tag: 'SUCCESS:MINIMAL-MOAI-001',
    });
  }

  /**
   * Create fallback memory content
   * @param config Installation configuration
   * @returns Fallback memory content
   * @tags @CODE:CREATE-FALLBACK-MEMORY-001:API
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
/alfred:8-project  # Initialize project documents
/alfred:1-spec     # Create specifications
/alfred:2-build    # TDD implementation
/alfred:3-sync     # Document synchronization
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
   * @tags @CODE:VALIDATE-CLAUDE-SETTINGS-001:API
   */
  async validateClaudeSettings(claudeDir: string): Promise<void> {
    try {
      const settingsPath = path.join(claudeDir, 'settings.json');

      if (!fs.existsSync(settingsPath)) {
        logger.warn('Claude settings.json not found after template copy', {
          settingsPath,
          tag: 'WARN:SETTINGS-NOT-FOUND-001',
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
        tag: 'DEBUG:SETTINGS-VALIDATION-001',
      });

      if (!hasEnv || !hasHooks || !hasPermissions) {
        logger.warn('Incomplete Claude settings detected', {
          settingsPath,
          missingEnv: !hasEnv,
          missingHooks: !hasHooks,
          missingPermissions: !hasPermissions,
          tag: 'WARN:SETTINGS-INCOMPLETE-001',
        });
      } else {
        logger.debug('Claude settings validation successful', {
          settingsPath,
          tag: 'SUCCESS:SETTINGS-VALIDATION-001',
        });
      }
    } catch (error) {
      const errorMessage = getErrorMessage(error);
      logger.error('Failed to validate Claude settings', {
        error: errorMessage,
        claudeDir,
        tag: 'ERROR:SETTINGS-VALIDATION-001',
      });
    }
  }

  /**
   * Create Claude Code settings (fallback)
   */
  async createClaudeSettings(
    config: InstallationConfig
  ): Promise<string | null> {
    const settingsPath = path.join(
      config.projectPath,
      '.claude',
      'settings.json'
    );
    if (fs.existsSync(settingsPath)) return settingsPath;

    return this.writeJsonFile(
      settingsPath,
      {
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
          'alfred:8-project': { enabled: true },
          'alfred:1-spec': { enabled: true },
          'alfred:2-build': { enabled: true },
          'alfred:3-sync': { enabled: true },
          'alfred:4-debug': { enabled: true },
        },
      },
      'ERROR:CLAUDE-SETTINGS-001'
    );
  }

  /**
   * Create MoAI configuration
   */
  async createMoaiConfig(config: InstallationConfig): Promise<string | null> {
    const configPath = path.join(config.projectPath, '.moai', 'config.json');

    return this.writeJsonFile(
      configPath,
      {
        version: '0.0.1',
        mode: config.mode,
        projectName: config.projectName,
        features: config.additionalFeatures,
        backup: { enabled: config.backupEnabled, retentionDays: 30 },
        git: {
          enabled: config.mode === 'team',
          autoCommit: true,
          branchPrefix: 'feature/',
        },
      },
      'ERROR:MOAI-CONFIG-001'
    );
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
