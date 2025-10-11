// @CODE:INST-005 |
// Related: @CODE:INST-005:API, @CODE:INST-TPL-001

/**
 * @file Template processing and interpolation
 * @author MoAI Team
 */

import * as fs from 'node:fs';
import * as os from 'node:os';
import * as path from 'node:path';
import { fileURLToPath } from 'node:url';
import mustache from 'mustache';
import { logger } from '@/utils/winston-logger';
import type { InstallationConfig } from './types';

/**
 * Handles template processing and file operations
 * @tags @CODE:TEMPLATE-PROCESSOR-001
 */
export class TemplateProcessor {
  /**
   * Get user home directory in a cross-platform way
   *
   * @returns Home directory path
   * @private
   *
   * Priority:
   * 1. process.env.HOME (Unix/macOS)
   * 2. process.env.USERPROFILE (Windows)
   * 3. os.homedir() (fallback)
   */
  private getHomeDirectory(): string {
    return process.env.HOME || process.env.USERPROFILE || os.homedir();
  }

  /**
   * Try to find templates relative to package installation
   * Strategy 1: Package-relative (HIGHEST PRIORITY)
   *
   * @param currentDir Current file directory
   * @returns Templates path if found, null otherwise
   * @private
   */
  private tryPackageRelativeTemplates(currentDir: string): string | null {
    const packageRoot = path.resolve(currentDir, '../../..');
    const packageTemplates = path.join(packageRoot, 'templates');

    if (fs.existsSync(packageTemplates)) {
      logger.debug('Found templates in package root', {
        templatePath: packageTemplates,
        strategy: 'package-relative',
        tag: 'DEBUG:TEMPLATES-PATH-001',
      });
      return packageTemplates;
    }

    return null;
  }

  /**
   * Try to find templates in development directory
   * Strategy 2: Development environment
   *
   * @param currentDir Current file directory
   * @returns Templates path if found, null otherwise
   * @private
   */
  private tryDevelopmentTemplates(currentDir: string): string | null {
    const devTemplates = path.resolve(currentDir, '../../../templates');

    if (fs.existsSync(devTemplates)) {
      logger.debug('Found templates in development directory', {
        templatePath: devTemplates,
        strategy: 'development',
        tag: 'DEBUG:TEMPLATES-PATH-002',
      });
      return devTemplates;
    }

    return null;
  }

  /**
   * Try to find templates in user's project node_modules
   * Strategy 3: User's project node_modules
   *
   * @returns Templates path if found, null otherwise
   * @private
   */
  private tryUserNodeModulesTemplates(): string | null {
    const userNodeModules = path.join(
      process.cwd(),
      'node_modules',
      'moai-adk',
      'templates'
    );

    if (fs.existsSync(userNodeModules)) {
      logger.debug('Found templates in user node_modules', {
        templatePath: userNodeModules,
        strategy: 'user-node-modules',
        tag: 'DEBUG:TEMPLATES-PATH-003',
      });
      return userNodeModules;
    }

    return null;
  }

  /**
   * Try to find templates in global installation paths
   * Strategy 4: Global installation paths
   *
   * @returns Templates path if found, null otherwise
   * @private
   */
  private tryGlobalInstallTemplates(): string | null {
    const nodeExecPath = process.execPath;
    const nodeBinDir = path.dirname(nodeExecPath);
    const nodeInstallDir = path.dirname(nodeBinDir);
    const homeDir = this.getHomeDirectory();

    const globalPaths = [
      path.join(nodeInstallDir, 'lib', 'node_modules', 'moai-adk', 'templates'),
      path.join(
        homeDir,
        '.bun',
        'install',
        'global',
        'node_modules',
        'moai-adk',
        'templates'
      ),
      path.join(
        homeDir,
        '.npm-global',
        'lib',
        'node_modules',
        'moai-adk',
        'templates'
      ),
    ];

    // Platform-specific paths
    if (process.platform !== 'win32') {
      globalPaths.unshift('/usr/local/lib/node_modules/moai-adk/templates');
    }

    for (const globalPath of globalPaths) {
      const resolvedPath = path.resolve(globalPath);
      if (fs.existsSync(resolvedPath)) {
        logger.debug('Found templates in global installation', {
          templatePath: resolvedPath,
          strategy: 'global-install',
          tag: 'DEBUG:TEMPLATES-PATH-004',
        });
        return resolvedPath;
      }
    }

    return null;
  }
  /**
   * Get templates directory path with robust cross-platform resolution
   *
   * @returns Path to templates directory
   * @tags @CODE:GET-TEMPLATES-PATH-001:API
   *
   * Resolution strategies (in priority order):
   * 1. User's node_modules: When user runs moai init (HIGHEST PRIORITY)
   * 2. Package-relative: Works for npm/bun install (global/local)
   * 3. Development: moai-adk-ts/templates (for development)
   * 4. Platform-specific global paths
   *
   * Cross-platform considerations:
   * - Uses process.env.HOME (Unix) or process.env.USERPROFILE (Windows)
   * - Unix-specific paths are conditionally added only on non-Windows platforms
   *
   * Priority rationale:
   * - User's node_modules takes precedence to ensure correct version is used
   * - Prevents accidentally using development/source templates in production
   */
  getTemplatesPath(): string {
    const currentFilePath = fileURLToPath(import.meta.url);
    const currentDir = path.dirname(currentFilePath);

    // Try each strategy in priority order (User node_modules first!)
    const strategies = [
      () => this.tryUserNodeModulesTemplates(),
      () => this.tryPackageRelativeTemplates(currentDir),
      () => this.tryDevelopmentTemplates(currentDir),
      () => this.tryGlobalInstallTemplates(),
    ];

    for (const strategy of strategies) {
      const result = strategy();
      if (result !== null) {
        return result;
      }
    }

    // Fallback: Return package-relative path for better error messages
    const packageRoot = path.resolve(currentDir, '../../..');
    const packageTemplates = path.join(packageRoot, 'templates');

    logger.warn('Templates directory not found, using fallback', {
      fallbackPath: packageTemplates,
      currentFile: currentFilePath,
      tag: 'WARN:TEMPLATES-PATH-001',
    });

    return packageTemplates;
  }

  /**
   * Create template variables for Mustache rendering
   * @param config Installation configuration
   * @returns Template variables object
   * @tags @CODE:CREATE-TEMPLATE-VARS-001:API
   */
  createTemplateVariables(config: InstallationConfig): Record<string, any> {
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
   * @param options Optional configuration
   * @param options.excludePaths Paths to exclude (relative to srcDir, e.g., ['specs', 'reports'])
   * @returns List of copied files
   * @tags @CODE:COPY-TEMPLATE-DIR-001:API
   */
  async copyTemplateDirectory(
    srcDir: string,
    dstDir: string,
    variables: Record<string, any>,
    options?: {
      excludePaths?: string[];
    }
  ): Promise<string[]> {
    const copiedFiles: string[] = [];
    const excludePaths = options?.excludePaths || [];

    try {
      await fs.promises.mkdir(dstDir, { recursive: true });
      const entries = await fs.promises.readdir(srcDir, {
        withFileTypes: true,
      });

      for (const entry of entries) {
        const srcPath = path.join(srcDir, entry.name);
        const dstPath = path.join(dstDir, entry.name);

        // Check if this path should be excluded
        const isExcluded = excludePaths.some(excludePath => {
          const normalizedExclude = path.normalize(excludePath).toLowerCase();
          const normalizedEntry = entry.name.toLowerCase();
          return normalizedEntry === normalizedExclude;
        });

        if (isExcluded) {
          logger.info(`Skipping excluded path: ${entry.name}`, {
            srcPath,
            tag: 'INFO:EXCLUDE-PATH-001',
          });
          continue; // Skip this entry
        }

        if (entry.isDirectory()) {
          const subFiles = await this.copyTemplateDirectory(
            srcPath,
            dstPath,
            variables,
            options
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
        tag: 'ERROR:COPY-TEMPLATE-DIR-001',
      });
      throw error;
    }
  }

  /**
   * Copy and process template file with variable substitution
   * @param srcPath Source file path
   * @param dstPath Destination file path
   * @param variables Template variables
   * @tags @CODE:COPY-TEMPLATE-FILE-001:API
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
              tag: 'WARN:CHMOD-001',
            });
          }
        }
      }

      logger.debug('Template file copied and processed', {
        srcPath,
        dstPath,
        isTextFile,
        tag: 'DEBUG:COPY-TEMPLATE-FILE-001',
      });
    } catch (error) {
      logger.error('Failed to copy template file', {
        error,
        srcPath,
        dstPath,
        tag: 'ERROR:COPY-TEMPLATE-FILE-001',
      });
      throw error;
    }
  }

  /**
   * Copy directory recursively (without template processing)
   * @param src Source directory
   * @param dst Destination directory
   * @tags @CODE:COPY-DIRECTORY-001:API
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
