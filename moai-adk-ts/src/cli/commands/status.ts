/**
 * @file CLI status command implementation
 * @author MoAI Team
 * @tags @FEATURE:CLI-STATUS-001 @REQ:CLI-FOUNDATION-012
 */

import * as path from 'node:path';
import chalk from 'chalk';
import * as fs from 'fs-extra';
import { logger } from '../../utils/winston-logger.js';

/**
 * Status command options
 * @tags @DESIGN:STATUS-OPTIONS-001
 */
export interface StatusOptions {
  readonly verbose: boolean;
  readonly projectPath?: string | undefined;
}

/**
 * Version information
 * @tags @DESIGN:VERSION-INFO-001
 */
export interface VersionInfo {
  readonly package: string;
  readonly resources: string;
  readonly available?: string;
  readonly outdated?: boolean;
}

/**
 * File count information
 * @tags @DESIGN:FILE-COUNT-001
 */
export interface FileCount {
  [key: string]: number;
}

/**
 * Project status information
 * @tags @DESIGN:PROJECT-STATUS-001
 */
export interface ProjectStatus {
  readonly path: string;
  readonly projectType: string;
  readonly moaiInitialized: boolean;
  readonly claudeInitialized: boolean;
  readonly memoryFile: boolean;
  readonly gitRepository: boolean;
  readonly versions?: VersionInfo;
  readonly fileCounts?: FileCount | undefined;
}

/**
 * Status result
 * @tags @DESIGN:STATUS-RESULT-001
 */
export interface StatusResult {
  readonly success: boolean;
  readonly status?: ProjectStatus;
  readonly recommendations?: string[];
  readonly error?: string;
}

/**
 * Status command for project status display
 * @tags @FEATURE:CLI-STATUS-001
 */
export class StatusCommand {
  /**
   * Check project status and components
   * @param projectPath - Path to project directory
   * @returns Project status information
   * @tags @API:CHECK-STATUS-001
   */
  public async checkProjectStatus(projectPath: string): Promise<ProjectStatus> {
    try {
      const resolvedPath = path.resolve(projectPath);

      // Check if MoAI is initialized
      const moaiDir = path.join(resolvedPath, '.moai');
      const moaiInitialized = await fs.pathExists(moaiDir);

      // Check if Claude is initialized
      const claudeDir = path.join(resolvedPath, '.claude');
      const claudeInitialized = await fs.pathExists(claudeDir);

      // Check if memory file exists
      const memoryFile = path.join(resolvedPath, 'CLAUDE.md');
      const memoryFileExists = await fs.pathExists(memoryFile);

      // Check if Git repository exists
      const gitDir = path.join(resolvedPath, '.git');
      const gitRepository = await fs.pathExists(gitDir);

      // Determine project type
      let projectType = 'Regular Directory';
      if (moaiInitialized && claudeInitialized) {
        projectType = 'MoAI Project (Full)';
      } else if (moaiInitialized) {
        projectType = 'MoAI Project (Partial)';
      } else if (claudeInitialized) {
        projectType = 'Claude Project';
      }

      return {
        path: resolvedPath,
        projectType,
        moaiInitialized,
        claudeInitialized,
        memoryFile: memoryFileExists,
        gitRepository,
      };
    } catch (error) {
      throw new Error(
        `Failed to check project status: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
    }
  }

  /**
   * Get version information
   * @param projectPath - Path to project directory
   * @returns Version information
   * @tags @API:GET-VERSION-INFO-001
   */
  public async getVersionInfo(projectPath: string): Promise<VersionInfo> {
    try {
      // Try to get package version from package.json in moai-adk-ts
      let packageVersion = '0.0.1';
      try {
        const packageJsonPath = path.join(
          __dirname,
          '../../..',
          'package.json'
        );
        const packageJsonExists = await fs.pathExists(packageJsonPath);
        if (packageJsonExists) {
          const packageJson = await fs.readJson(packageJsonPath);
          packageVersion = packageJson.version || '0.0.1';
        }
      } catch {
        // Fallback to default version
        packageVersion = '0.0.1';
      }

      // Try to get resource version from .moai/version.json
      let resourceVersion = packageVersion;
      let outdated = false;

      try {
        const versionFilePath = path.join(projectPath, '.moai', 'version.json');
        const versionFileExists = await fs.pathExists(versionFilePath);
        if (versionFileExists) {
          const versionInfo = await fs.readJson(versionFilePath);
          resourceVersion =
            versionInfo.template_version ||
            versionInfo.version ||
            packageVersion;

          // Check if outdated
          if (resourceVersion !== packageVersion) {
            outdated = true;
          }
        }
      } catch {
        // Use package version as fallback
        resourceVersion = packageVersion;
      }

      return {
        package: packageVersion,
        resources: resourceVersion,
        available: packageVersion,
        outdated,
      };
    } catch (error) {
      throw new Error(
        `Failed to get version information: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
    }
  }

  /**
   * Count project files
   * @param projectPath - Path to project directory
   * @returns File count information
   * @tags @API:COUNT-FILES-001
   */
  public async countProjectFiles(projectPath: string): Promise<FileCount> {
    try {
      const counts: FileCount = {};
      let total = 0;

      // Count files in .moai directory
      const moaiDir = path.join(projectPath, '.moai');
      const moaiExists = await fs.pathExists(moaiDir);
      if (moaiExists) {
        const moaiCount = await this.countFilesInDirectory(moaiDir);
        counts['.moai'] = moaiCount;
        total += moaiCount;
      } else {
        counts['.moai'] = 0;
      }

      // Count files in .claude directory
      const claudeDir = path.join(projectPath, '.claude');
      const claudeExists = await fs.pathExists(claudeDir);
      if (claudeExists) {
        const claudeCount = await this.countFilesInDirectory(claudeDir);
        counts['.claude'] = claudeCount;
        total += claudeCount;
      } else {
        counts['.claude'] = 0;
      }

      // Count CLAUDE.md if it exists
      const memoryFile = path.join(projectPath, 'CLAUDE.md');
      const memoryExists = await fs.pathExists(memoryFile);
      if (memoryExists) {
        counts['CLAUDE.md'] = 1;
        total += 1;
      } else {
        counts['CLAUDE.md'] = 0;
      }

      counts['total'] = total;
      return counts;
    } catch (error) {
      throw new Error(
        `Failed to count project files: ${error instanceof Error ? error.message : 'Unknown error'}`
      );
    }
  }

  /**
   * Count files in a directory recursively
   * @param dirPath - Directory path
   * @returns Number of files
   * @tags @UTIL:COUNT-FILES-DIRECTORY-001
   */
  private async countFilesInDirectory(dirPath: string): Promise<number> {
    try {
      const items = await fs.readdir(dirPath, { withFileTypes: true });
      let count = 0;

      for (const item of items) {
        if (item.isFile()) {
          count++;
        } else if (item.isDirectory()) {
          const subDirPath = path.join(dirPath, item.name);
          count += await this.countFilesInDirectory(subDirPath);
        }
      }

      return count;
    } catch {
      // Return 0 if directory cannot be read
      return 0;
    }
  }

  /**
   * Run status command
   * @param options - Status options
   * @returns Status result
   * @tags @API:STATUS-RUN-001
   */
  public async run(options: StatusOptions): Promise<StatusResult> {
    try {
      const projectPath = options.projectPath || process.cwd();

      logger.info(chalk.cyan('📊 MoAI-ADK Project Status'));

      // Step 1: Get project status
      const status = await this.checkProjectStatus(projectPath);

      logger.info(`\n📂 Project: ${status.path}`);
      logger.info(`   Type: ${status.projectType}`);

      // Step 2: Display core status
      logger.info('\n🗿 MoAI-ADK Components:');
      logger.info(`   MoAI System: ${status.moaiInitialized ? '✅' : '❌'}`);
      logger.info(
        `   Claude Integration: ${status.claudeInitialized ? '✅' : '❌'}`
      );
      logger.info(`   Memory File: ${status.memoryFile ? '✅' : '❌'}`);
      logger.info(`   Git Repository: ${status.gitRepository ? '✅' : '❌'}`);

      // Step 3: Get and display version information
      const versions = await this.getVersionInfo(projectPath);
      logger.info('\n🧭 Versions:');
      logger.info(`   Package: v${versions.package}`);
      logger.info(`   Templates: v${versions.resources}`);

      if (versions.available && versions.available !== versions.resources) {
        logger.info(`   Available template update: v${versions.available}`);
      }

      if (versions.outdated) {
        logger.info(
          chalk.yellow(
            "   ⚠️  Templates are outdated. Run 'moai update' to refresh."
          )
        );
      }

      // Step 4: Display file counts if verbose
      if (options.verbose) {
        const fileCounts = await this.countProjectFiles(projectPath);
        logger.info('\n📁 File Counts:');

        for (const [component, count] of Object.entries(fileCounts)) {
          if (component !== 'total') {
            logger.info(`   ${component}: ${count} files`);
          }
        }
      }

      // Step 5: Generate recommendations
      const recommendations: string[] = [];
      if (!status.moaiInitialized) {
        recommendations.push("Run 'moai init' to initialize MoAI-ADK");
      }
      if (!status.gitRepository) {
        recommendations.push('Initialize Git repository: git init');
      }

      if (recommendations.length > 0) {
        logger.info('\n💡 Recommendations:');
        for (const rec of recommendations) {
          logger.info(`   - ${rec}`);
        }
      }

      // Attach additional data to status
      const fileCounts = options.verbose
        ? await this.countProjectFiles(projectPath)
        : undefined;

      const enrichedStatus: ProjectStatus = {
        ...status,
        versions,
        fileCounts,
      };

      return {
        success: true,
        status: enrichedStatus,
        recommendations,
      };
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      logger.info(
        chalk.red(`❌ Failed to get project status: ${errorMessage}`)
      );

      return {
        success: false,
        error: errorMessage,
      };
    }
  }
}
