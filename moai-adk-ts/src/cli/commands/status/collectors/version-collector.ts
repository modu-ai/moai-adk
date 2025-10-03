// @CODE:CLI-VER-001 | SPEC: status.ts 리팩토링
// Related: @CODE:CLI-003

/**
 * @file Version information collection functionality
 * @author MoAI Team
 */

import * as path from 'node:path';
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
      // Try to get package version from package.json in moai-adk-ts
      let packageVersion = '0.2.0';
      try {
        const packageJsonPath = path.join(
          __dirname,
          '../../../..',
          'package.json'
        );
        const packageJsonExists = await fs.pathExists(packageJsonPath);
        if (packageJsonExists) {
          const packageJson = await fs.readJson(packageJsonPath);
          packageVersion = packageJson.version || '0.2.0';
        }
      } catch {
        // Fallback to default version
        packageVersion = '0.2.0';
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
}
