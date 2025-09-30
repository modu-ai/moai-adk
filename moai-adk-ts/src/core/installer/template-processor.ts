/**
 * @file Template Processing and File Operations
 * @author MoAI Team
 * @tags @FEATURE:TEMPLATE-PROCESSOR-001 @REQ:INSTALL-SYSTEM-012
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { fileURLToPath } from 'node:url';
import mustache from 'mustache';
import { logger } from '@/utils/logger';
import { InstallationError, getErrorMessage } from '@/utils/errors';
import type { InstallationConfig } from './types';

/**
 * Handles template processing and file operations
 * @tags @FEATURE:TEMPLATE-PROCESSOR-001
 */
export class TemplateProcessor {
  /**
   * Get templates directory path with robust resolution for all installation scenarios
   * @returns Path to templates directory
   * @tags @API:GET-TEMPLATES-PATH-001
   *
   * Handles:
   * - Local development: moai-adk-ts/templates
   * - Global install: /usr/local/lib/node_modules/moai-adk/templates
   * - Local install: node_modules/moai-adk/templates
   * - Bun global: ~/.bun/install/global/node_modules/moai-adk/templates
   */
  getTemplatesPath(): string {
    // Get current file location (works in ESM)
    const currentFilePath = fileURLToPath(import.meta.url);
    const currentDir = path.dirname(currentFilePath);

    // Calculate package root from dist/core/installer/template-processor.js
    // dist/core/installer -> ../../.. -> package root
    const packageRoot = path.resolve(currentDir, '../../..');
    const packageTemplates = path.join(packageRoot, 'templates');

    // Strategy 1: Package-relative (HIGHEST PRIORITY)
    // This works for: npm install -g, npm install, bun install -g, bun install
    if (fs.existsSync(packageTemplates)) {
      logger.debug('Found templates in package root', {
        templatePath: packageTemplates,
        strategy: 'package-relative',
        tag: '@DEBUG:TEMPLATES-PATH-001',
      });
      return packageTemplates;
    }

    // Strategy 2: Development environment (for moai-adk-ts development)
    // dist/core/installer -> ../../../../templates
    const devTemplates = path.resolve(currentDir, '../../../templates');
    if (fs.existsSync(devTemplates)) {
      logger.debug('Found templates in development directory', {
        templatePath: devTemplates,
        strategy: 'development',
        tag: '@DEBUG:TEMPLATES-PATH-002',
      });
      return devTemplates;
    }

    // Strategy 3: User's project node_modules (when user runs moai init)
    const userNodeModules = path.join(process.cwd(), 'node_modules', 'moai-adk', 'templates');
    if (fs.existsSync(userNodeModules)) {
      logger.debug('Found templates in user node_modules', {
        templatePath: userNodeModules,
        strategy: 'user-node-modules',
        tag: '@DEBUG:TEMPLATES-PATH-003',
      });
      return userNodeModules;
    }

    // Strategy 4: Global installation paths
    const nodeExecPath = process.execPath;
    const nodeBinDir = path.dirname(nodeExecPath);
    const nodeInstallDir = path.dirname(nodeBinDir);

    const globalPaths = [
      // Standard npm global
      '/usr/local/lib/node_modules/moai-adk/templates',
      // Node version manager paths
      path.join(nodeInstallDir, 'lib', 'node_modules', 'moai-adk', 'templates'),
      // Bun global
      path.join(process.env['HOME'] || '~', '.bun', 'install', 'global', 'node_modules', 'moai-adk', 'templates'),
      // npm-global (alternative)
      path.join(process.env['HOME'] || '~', '.npm-global', 'lib', 'node_modules', 'moai-adk', 'templates'),
    ];

    for (const globalPath of globalPaths) {
      const resolvedPath = path.resolve(globalPath);
      if (fs.existsSync(resolvedPath)) {
        logger.debug('Found templates in global installation', {
          templatePath: resolvedPath,
          strategy: 'global-install',
          tag: '@DEBUG:TEMPLATES-PATH-004',
        });
        return resolvedPath;
      }
    }

    // Fallback: Return package-relative path even if it doesn't exist
    // This allows for better error messages downstream
    logger.warn('Templates directory not found, using package-relative fallback', {
      fallbackPath: packageTemplates,
      searchedPaths: [
        packageTemplates,
        devTemplates,
        userNodeModules,
        ...globalPaths.map(p => path.resolve(p)),
      ],
      currentFile: currentFilePath,
      packageRoot,
      tag: '@WARN:TEMPLATES-PATH-001',
    });

    return packageTemplates;
  }

  /**
   * Create template variables for Mustache rendering
   * @param config Installation configuration
   * @returns Template variables object
   * @tags @API:CREATE-TEMPLATE-VARS-001
   */
  createTemplateVariables(
    config: InstallationConfig
  ): Record<string, any> {
    return {
      PROJECT_NAME: config.projectName,
      PROJECT_DESCRIPTION: `A ${config.projectName} project built with MoAI-ADK`,
      PROJECT_VERSION: '0.1.0',
      PROJECT_MODE: config.mode,
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
   * @tags @API:COPY-TEMPLATE-DIR-001
   */
  async copyTemplateDirectory(
    srcDir: string,
    dstDir: string,
    variables: Record<string, any>
  ): Promise<string[]> {
    const copiedFiles: string[] = [];

    try {
      await fs.promises.mkdir(dstDir, { recursive: true });
      const entries = await fs.promises.readdir(srcDir, {
        withFileTypes: true,
      });

      for (const entry of entries) {
        const srcPath = path.join(srcDir, entry.name);
        const dstPath = path.join(dstDir, entry.name);

        if (entry.isDirectory()) {
          const subFiles = await this.copyTemplateDirectory(
            srcPath,
            dstPath,
            variables
          );
          copiedFiles.push(...subFiles);
        } else {
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
   * @tags @API:COPY-TEMPLATE-FILE-001
   */
  async copyTemplateFile(
    srcPath: string,
    dstPath: string,
    variables: Record<string, any>
  ): Promise<void> {
    try {
      await fs.promises.mkdir(path.dirname(dstPath), { recursive: true });

      const content = await fs.promises.readFile(srcPath, 'utf-8');

      const fileExt = path.extname(srcPath).toLowerCase();
      const isTextFile = [
        '.md',
        '.json',
        '.js',
        '.ts',
        '.py',
        '.txt',
        '.yml',
        '.yaml',
      ].includes(fileExt);

      let processedContent: string;
      if (isTextFile) {
        processedContent = mustache.render(content, variables);
      } else {
        processedContent = content;
      }

      await fs.promises.writeFile(dstPath, processedContent);

      if (['.py', '.sh', '.js'].includes(fileExt)) {
        try {
          await fs.promises.chmod(dstPath, 0o755);
        } catch (chmodError) {
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
   * Copy directory recursively (without template processing)
   * @param src Source directory
   * @param dst Destination directory
   * @tags @API:COPY-DIRECTORY-001
   */
  async copyDirectory(src: string, dst: string): Promise<void> {
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
}