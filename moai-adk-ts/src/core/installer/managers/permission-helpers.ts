/**
 * @file Permission management helper functions
 * @author MoAI Team
 * @tags @DESIGN:PERMISSION-HELPERS-012
 */

import * as fs from 'fs-extra';
import { type FilePermissions, MoAIPermissions } from './permission-types';
import { PermissionUtils } from './permission-utils';

/**
 * Permission operation helper functions
 */
export class PermissionHelpers {
  static async setUnixPermissions(
    targetPath: string,
    permissions: FilePermissions
  ): Promise<void> {
    const octal = PermissionUtils.permissionsToOctal(permissions);
    await fs.chmod(targetPath, parseInt(octal, 8));
  }

  static async setWindowsPermissions(
    targetPath: string,
    permissions: FilePermissions
  ): Promise<void> {
    try {
      const stats = await fs.stat(targetPath);

      if (!permissions.owner.write && stats.mode & 0o200) {
        await fs.chmod(targetPath, stats.mode & ~0o200);
      } else if (permissions.owner.write && !(stats.mode & 0o200)) {
        await fs.chmod(targetPath, stats.mode | 0o200);
      }
    } catch (error) {
      console.warn(
        `Windows permission setting limited for ${targetPath}: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  }

  static async makeUnixExecutable(filePath: string): Promise<void> {
    const stats = await fs.stat(filePath);
    const currentMode = stats.mode;
    const executableMode = currentMode | 0o111;
    await fs.chmod(filePath, executableMode);
  }

  static async makeWindowsExecutable(filePath: string): Promise<void> {
    if (!PermissionUtils.isScript(filePath)) {
      console.warn(
        `Warning: ${filePath} may not be executable on Windows without proper extension`
      );
    }
  }

  static async getPermissionFlags(targetPath: string): Promise<{
    readable: boolean;
    writable: boolean;
    executable: boolean;
  }> {
    const flags = { readable: false, writable: false, executable: false };

    try {
      await fs.access(targetPath, fs.constants.R_OK);
      flags.readable = true;
    } catch {
      // Not readable
    }

    try {
      await fs.access(targetPath, fs.constants.W_OK);
      flags.writable = true;
    } catch {
      // Not writable
    }

    try {
      await fs.access(targetPath, fs.constants.X_OK);
      flags.executable = true;
    } catch {
      // Not executable
    }

    return flags;
  }

  static getOctalMode(
    stats: fs.Stats,
    platform: 'windows' | 'unix'
  ): string | undefined {
    if (platform === 'unix') {
      return (stats.mode & 0o777).toString(8);
    }
    return undefined;
  }

  static isOwner(stats: fs.Stats, platform: 'windows' | 'unix'): boolean {
    return stats.uid === process.getuid?.() || platform === 'windows';
  }

  static determineFilePermissions(
    filePath: string,
    warnings: string[]
  ): FilePermissions {
    if (PermissionUtils.isSensitive(filePath)) {
      warnings.push(
        `Applied strict permissions to sensitive file: ${filePath}`
      );
      return MoAIPermissions.SENSITIVE_FILES;
    } else if (PermissionUtils.isScript(filePath)) {
      return MoAIPermissions.SCRIPT_FILES;
    } else if (PermissionUtils.isConfig(filePath)) {
      return MoAIPermissions.CONFIG_FILES;
    } else {
      return MoAIPermissions.CONFIG_FILES;
    }
  }
}
