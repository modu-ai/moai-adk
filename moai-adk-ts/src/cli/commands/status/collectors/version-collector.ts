// @CODE:CLI-VER-001 | SPEC: status.ts 리팩토링
// Related: @CODE:CLI-003

/**
 * @file Version information collection functionality
 * @author MoAI Team
 */

import { readFile } from 'node:fs/promises';
import * as path from 'node:path';
import { fileURLToPath } from 'node:url';
import * as fs from 'fs-extra';

/**
 * Version information
 * @tags @SPEC:VERSION-INFO-002
 */
export interface VersionInfo {
  readonly package: string;
  readonly resources: string;
  readonly available?: string;
  readonly outdated?: boolean;
}

/**
 * Version collector for MoAI-ADK
 * @tags @CODE:VERSION-COLLECTOR-001
 */
export class VersionCollector {
  /**
   * Get version information
   * @param projectPath - Path to project directory
   * @returns Version information
   * @tags @CODE:COLLECT-VERSION-001:API
   */
  public async collectVersionInfo(projectPath: string): Promise<VersionInfo> {
    try {
      // Try to get package version from VERSION file or package.json
      let packageVersion = 'unknown';
      try {
        // Use import.meta.url to get the current module path
        const currentFilePath = fileURLToPath(import.meta.url);
        const currentDir = path.dirname(currentFilePath);

        // Try VERSION file first (more reliable after bundling)
        // From: dist/cli/index.js -> ../../VERSION
        const versionFilePath = path.join(currentDir, '../../VERSION');
        if (await fs.pathExists(versionFilePath)) {
          const versionContent = await readFile(versionFilePath, 'utf8');
          packageVersion = versionContent.trim();
        } else {
          // Fallback to package.json
          const packageJsonPath = path.join(currentDir, '../../package.json');
          if (await fs.pathExists(packageJsonPath)) {
            const packageJsonContent = await readFile(packageJsonPath, 'utf8');
            const packageJson = JSON.parse(packageJsonContent);
            packageVersion = packageJson.version || 'unknown';
          }
        }
      } catch (error) {
        // Fallback: try alternative paths
        console.error('Failed to read version:', error);
        packageVersion = 'unknown';
      }

      // Try to get resource version from .moai/version.json
      let resourceVersion = packageVersion;
      let outdated = false;

      try {
        const versionFilePath = path.join(projectPath, '.moai', 'version.json');
        const versionFileExists = await fs.pathExists(versionFilePath);
        if (versionFileExists) {
          const versionFileContent = await readFile(versionFilePath, 'utf8');
          const versionInfo = JSON.parse(versionFileContent);
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
}
