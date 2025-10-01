// @CODE:CFG-PKG-BUILDER-001 |
// Related: @CODE:CFG-001

/**
 * @file Package.json builder
 * @author MoAI Team
 */

import { logger } from '../../../utils/winston-logger.js';
import type {
  PackageConfig,
  PackageJsonResult,
  ProjectConfigInput,
} from '../types';
import {
  getPackageDependencies,
  getPackageDevDependencies,
  getPackageScripts,
} from '../utils/config-helpers.js';
import { writeJsonFile } from '../utils/config-file-utils.js';

/**
 * Build package.json for Node.js projects
 */
export async function buildPackageJson(
  packagePath: string,
  config: ProjectConfigInput
): Promise<PackageJsonResult> {
  try {
    // Skip if not needed
    if (!config.shouldCreatePackageJson) {
      return {
        success: true,
        skipped: true,
        reason: 'package.json not needed for this project type',
      };
    }

    const packageConfig: PackageConfig = {
      name: config.projectName,
      version: '1.0.0',
      description: `MoAI-ADK project: ${config.projectName}`,
      main: 'index.js',
      scripts: getPackageScripts(config),
      dependencies: getPackageDependencies(config),
      devDependencies: getPackageDevDependencies(config),
      keywords: ['moai-adk', ...config.techStack],
      author: 'MoAI Developer',
      license: 'MIT',
      engines: {
        node: '>=18.0.0',
      },
    };

    writeJsonFile(packagePath, packageConfig);

    logger.info(`Package.json created: ${packagePath}`);
    return {
      success: true,
      filePath: packagePath,
      packageConfig,
    };
  } catch (error) {
    const errorMessage =
      error instanceof Error ? error.message : 'Unknown error';
    logger.error('Error creating package.json:', errorMessage);
    return {
      success: false,
      error: errorMessage,
    };
  }
}
