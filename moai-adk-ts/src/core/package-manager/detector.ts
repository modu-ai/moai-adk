// @FEATURE:PKG-001 | Chain: @REQ:PKG-001 -> @DESIGN:PKG-001 -> @TASK:PKG-001 -> @TEST:PKG-001
// Related: @API:PKG-001, @DATA:PKG-INFO-001

/**
 * @file Package manager detector and analyzer
 * @author MoAI Team
 */

import * as fs from 'node:fs/promises';
import * as path from 'node:path';
import { execa } from 'execa';
import {
  type PackageManagerCommands,
  type PackageManagerDetectionResult,
  type PackageManagerInfo,
  PackageManagerType,
} from '@/types/package-manager';

/**
 * Package manager detector for automatic detection and recommendation
 * @tags @FEATURE:PACKAGE-MANAGER-DETECTOR-001
 */
export class PackageManagerDetector {
  /**
   * Detect single package manager availability and version
   * @param type - Package manager type to detect
   * @returns Package manager information
   * @tags @API:DETECT-SINGLE-001
   */
  public async detectPackageManager(
    type: PackageManagerType
  ): Promise<PackageManagerInfo> {
    try {
      const result = await execa(type, ['--version'], {
        timeout: 10000,
        reject: false,
      });

      if (result.exitCode === 0) {
        return {
          type,
          version: result.stdout.trim(),
          isAvailable: true,
        };
      } else {
        return {
          type,
          version: 'unknown',
          isAvailable: false,
        };
      }
    } catch (_error) {
      return {
        type,
        version: 'unknown',
        isAvailable: false,
      };
    }
  }

  /**
   * Detect all available package managers
   * @param workingDirectory - Directory to check for lock files
   * @returns Detection result with all available package managers
   * @tags @API:DETECT-ALL-001
   */
  public async detectAllPackageManagers(
    workingDirectory?: string
  ): Promise<PackageManagerDetectionResult> {
    const packageManagers = Object.values(PackageManagerType);

    // Detect all package managers concurrently
    const detectionPromises = packageManagers.map(type =>
      this.detectPackageManager(type)
    );

    const allResults = await Promise.all(detectionPromises);
    const available = allResults.filter(result => result.isAvailable);

    // Detect preferred package manager from lock files
    const preferredType = await this.detectLockFile(workingDirectory);
    const preferred = preferredType
      ? available.find(pm => pm.type === preferredType)
      : undefined;

    if (preferred) {
      preferred.isPreferred = true;
    }

    // Recommend fastest/most modern if no preference
    const recommended = preferred || this.recommendPackageManager(available);

    return {
      available,
      preferred,
      recommended,
    };
  }

  /**
   * Get commands for specific package manager
   * @param type - Package manager type
   * @returns Commands object
   * @tags @API:GET-COMMANDS-001
   */
  public getCommands(type: PackageManagerType): PackageManagerCommands {
    switch (type) {
      case PackageManagerType.NPM:
        return {
          install: 'npm install',
          installDev: 'npm install --save-dev',
          installGlobal: 'npm install --global',
          run: 'npm run',
          build: 'npm run build',
          test: 'npm test',
          init: 'npm init -y',
        };

      case PackageManagerType.YARN:
        return {
          install: 'yarn add',
          installDev: 'yarn add --dev',
          installGlobal: 'yarn global add',
          run: 'yarn run',
          build: 'yarn build',
          test: 'yarn test',
          init: 'yarn init -y',
        };

      case PackageManagerType.PNPM:
        return {
          install: 'pnpm add',
          installDev: 'pnpm add --save-dev',
          installGlobal: 'pnpm add --global',
          run: 'pnpm run',
          build: 'pnpm build',
          test: 'pnpm test',
          init: 'pnpm init',
        };

      default:
        throw new Error(`Unsupported package manager: ${type}`);
    }
  }

  /**
   * Recommend best package manager from available options
   * @param available - Available package managers
   * @returns Recommended package manager
   * @tags @API:RECOMMEND-001
   */
  public recommendPackageManager(
    available: PackageManagerInfo[]
  ): PackageManagerInfo {
    if (available.length === 0) {
      throw new Error('No package managers available');
    }

    // Priority order: pnpm > yarn > npm (based on performance and features)
    const priority = [
      PackageManagerType.PNPM,
      PackageManagerType.YARN,
      PackageManagerType.NPM,
    ];

    for (const preferredType of priority) {
      const found = available.find(pm => pm.type === preferredType);
      if (found) {
        return found;
      }
    }

    // Fallback to first available
    return available[0]!;
  }

  /**
   * Detect preferred package manager from lock files
   * @param workingDirectory - Directory to check
   * @returns Detected package manager type or null
   * @tags @API:DETECT-LOCK-001
   */
  private async detectLockFile(
    workingDirectory?: string
  ): Promise<PackageManagerType | null> {
    const baseDir = workingDirectory || process.cwd();

    const lockFiles = [
      { file: 'package-lock.json', type: PackageManagerType.NPM },
      { file: 'yarn.lock', type: PackageManagerType.YARN },
      { file: 'pnpm-lock.yaml', type: PackageManagerType.PNPM },
    ];

    for (const { file, type } of lockFiles) {
      if (await this.fileExists(path.join(baseDir, file))) {
        return type;
      }
    }

    return null;
  }

  /**
   * Check if file exists
   * @param filePath - File path to check
   * @returns Whether file exists
   * @tags @UTIL:FILE-EXISTS-001
   */
  private async fileExists(filePath: string): Promise<boolean> {
    try {
      await fs.access(filePath);
      return true;
    } catch {
      return false;
    }
  }
}
