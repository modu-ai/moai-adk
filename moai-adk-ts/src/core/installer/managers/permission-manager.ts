/**
 * @file Cross-platform permission management system
 * @author MoAI Team
 * @tags @DESIGN:PERMISSION-MANAGER-012 @REQ:CROSS-PLATFORM-PERMISSIONS-012
 */

import * as fs from 'fs-extra';
import * as path from 'path';
import {
  type FilePermissions,
  type PermissionStatus,
  type PermissionFixResult,
  type PlatformType,
  MoAIPermissions,
} from './permission-types';
import { PermissionUtils } from './permission-utils';
import { PermissionHelpers } from './permission-helpers';

/**
 * Cross-platform permission manager
 * @tags @FEATURE:PERMISSION-MANAGER-012
 */
export class PermissionManager {
  private readonly platform: PlatformType;

  constructor() {
    this.platform = PermissionUtils.getCurrentPlatform();
  }

  async setFilePermissions(
    filePath: string,
    permissions: FilePermissions
  ): Promise<void> {
    try {
      await fs.access(filePath);

      if (this.platform === 'unix') {
        await PermissionHelpers.setUnixPermissions(filePath, permissions);
      } else {
        await PermissionHelpers.setWindowsPermissions(filePath, permissions);
      }
    } catch (error) {
      throw new Error(
        `Failed to set permissions for ${filePath}: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  }

  async setDirectoryPermissions(
    dirPath: string,
    permissions: FilePermissions
  ): Promise<void> {
    try {
      await fs.access(dirPath);

      if (this.platform === 'unix') {
        await PermissionHelpers.setUnixPermissions(dirPath, permissions);
      } else {
        await PermissionHelpers.setWindowsPermissions(dirPath, permissions);
      }
    } catch (error) {
      throw new Error(
        `Failed to set directory permissions for ${dirPath}: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  }

  async makeExecutable(filePath: string): Promise<void> {
    try {
      if (this.platform === 'unix') {
        await PermissionHelpers.makeUnixExecutable(filePath);
      } else {
        await PermissionHelpers.makeWindowsExecutable(filePath);
      }
    } catch (error) {
      throw new Error(
        `Failed to make ${filePath} executable: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  }

  async checkPermissions(targetPath: string): Promise<PermissionStatus> {
    try {
      const stats = await fs.stat(targetPath);
      const permissions =
        await PermissionHelpers.getPermissionFlags(targetPath);
      const octalMode = PermissionHelpers.getOctalMode(stats, this.platform);

      return {
        path: targetPath,
        readable: permissions.readable,
        writable: permissions.writable,
        executable: permissions.executable,
        isOwner: PermissionHelpers.isOwner(stats, this.platform),
        ...(octalMode && { octalMode }),
      };
    } catch (error) {
      throw new Error(
        `Failed to check permissions for ${targetPath}: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  }

  async validatePermissions(
    targetPath: string,
    required: FilePermissions
  ): Promise<boolean> {
    try {
      const status = await this.checkPermissions(targetPath);

      if (required.owner.read && !status.readable) return false;
      if (required.owner.write && !status.writable) return false;
      if (required.owner.execute && !status.executable) return false;

      return true;
    } catch {
      return false;
    }
  }

  async fixPermissions(projectPath: string): Promise<PermissionFixResult> {
    const result = {
      fixedFiles: [] as string[],
      failures: [] as string[],
      warnings: [] as string[],
    };

    try {
      await this.fixDirectoryPermissions(projectPath, result);

      return {
        success: result.failures.length === 0,
        fixedFiles: result.fixedFiles,
        failures: result.failures,
        warnings: result.warnings,
      };
    } catch (error) {
      result.failures.push(
        `Failed to fix permissions: ${error instanceof Error ? error.message : String(error)}`
      );

      return {
        success: false,
        fixedFiles: result.fixedFiles,
        failures: result.failures,
        warnings: result.warnings,
      };
    }
  }

  private async fixDirectoryPermissions(
    dirPath: string,
    result: { fixedFiles: string[]; failures: string[]; warnings: string[] }
  ): Promise<void> {
    try {
      const entries = await fs.readdir(dirPath, { withFileTypes: true });

      // Set directory permissions
      await this.fixSingleDirectoryPermissions(dirPath, result);

      // Process entries
      for (const entry of entries) {
        const fullPath = path.join(dirPath, entry.name);

        if (entry.isDirectory()) {
          await this.fixDirectoryPermissions(fullPath, result);
        } else {
          await this.fixSingleFilePermissions(fullPath, result);
        }
      }
    } catch (error) {
      result.failures.push(
        `Failed to read directory ${dirPath}: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  }

  private async fixSingleDirectoryPermissions(
    dirPath: string,
    result: { fixedFiles: string[]; failures: string[]; warnings: string[] }
  ): Promise<void> {
    try {
      await this.setDirectoryPermissions(dirPath, MoAIPermissions.DIRECTORIES);
      result.fixedFiles.push(dirPath);
    } catch (error) {
      result.failures.push(
        `Directory ${dirPath}: ${error instanceof Error ? error.message : String(error)}`
      );
    }
  }

  private async fixSingleFilePermissions(
    filePath: string,
    result: { fixedFiles: string[]; failures: string[]; warnings: string[] }
  ): Promise<void> {
    try {
      const permissions = PermissionHelpers.determineFilePermissions(
        filePath,
        result.warnings
      );
      await this.setFilePermissions(filePath, permissions);
      result.fixedFiles.push(filePath);
    } catch (error) {
      result.failures.push(
        `File ${filePath}: ${error instanceof Error ? error.message : String(error)}`
      );
    }
  }
}

export default PermissionManager;
