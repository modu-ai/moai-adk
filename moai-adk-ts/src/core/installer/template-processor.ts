/**
 * @file Template Processing and File Operations
 * @author MoAI Team
 * @tags @FEATURE:TEMPLATE-PROCESSOR-001 @REQ:INSTALL-SYSTEM-012
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
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
   * Get templates directory path
   * @returns Path to templates directory
   * @tags @API:GET-TEMPLATES-PATH-001
   */
  getTemplatesPath(): string {
    const nodeExecPath = process.execPath;
    const nodeBinDir = path.dirname(nodeExecPath);
    const nodeInstallDir = path.dirname(nodeBinDir);

    const possiblePaths = [
      path.join(process.cwd(), 'templates'),
      path.join(process.cwd(), 'src', '..', 'templates'),
      path.join(process.cwd(), 'node_modules', 'moai-adk', 'templates'),
      path.join(nodeInstallDir, 'lib', 'node_modules', 'moai-adk', 'templates'),
      path.join(process.cwd(), '..', 'templates'),
      '/usr/local/lib/node_modules/moai-adk/templates',
      path.join(
        process.env['HOME'] || '~',
        '.npm-global',
        'lib',
        'node_modules',
        'moai-adk',
        'templates'
      ),
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