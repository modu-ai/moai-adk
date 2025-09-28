/**
 * @file Cross-platform permission management system (refactored)
 * @author MoAI Team
 * @tags @DESIGN:PERMISSION-MANAGER-012 @REQ:CROSS-PLATFORM-PERMISSIONS-012
 */

import * as fs from 'fs-extra';
import * as path from 'path';
import {
  FilePermissions,
  PermissionStatus,
  PermissionFixResult,
  PlatformType,
  MoAIPermissions
} from './permission-types';
import { PermissionUtils } from './permission-utils';

/**
 * Cross-platform permission manager
 * @tags @FEATURE:PERMISSION-MANAGER-012
 */
export class PermissionManager {
  private readonly platform: PlatformType;

  constructor() {
    this.platform = PermissionUtils.getCurrentPlatform();
  }

  /**
   * Set permissions for a file
   * @param filePath - Path to the file
   * @param permissions - Desired permissions
   * @throws Error if permission setting fails
   */
  async setFilePermissions(filePath: string, permissions: FilePermissions): Promise<void> {
    try {
      await fs.access(filePath);

      if (this.platform === 'unix') {
        await this.setUnixPermissions(filePath, permissions);
      } else {
        await this.setWindowsPermissions(filePath, permissions);
      }
    } catch (error) {
      throw new Error(
        `Failed to set permissions for ${filePath}: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  }

  /**
   * Set permissions for a directory
   * @param dirPath - Path to the directory
   * @param permissions - Desired permissions
   * @throws Error if permission setting fails
   */
  async setDirectoryPermissions(dirPath: string, permissions: FilePermissions): Promise<void> {
    try {
      await fs.access(dirPath);

      if (this.platform === 'unix') {
        await this.setUnixPermissions(dirPath, permissions);
      } else {
        await this.setWindowsPermissions(dirPath, permissions);
      }
    } catch (error) {
      throw new Error(
        `Failed to set directory permissions for ${dirPath}: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  }

  /**
   * Make a file executable
   * @param filePath - Path to the file
   * @throws Error if making executable fails
   */
  async makeExecutable(filePath: string): Promise<void> {
    try {
      if (this.platform === 'unix') {
        await this.makeUnixExecutable(filePath);
      } else {
        await this.makeWindowsExecutable(filePath);
      }
    } catch (error) {
      throw new Error(
        `Failed to make ${filePath} executable: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  }

  /**
   * Check current permissions of a path
   * @param targetPath - Path to check
   * @returns Permission status
   */
  async checkPermissions(targetPath: string): Promise<PermissionStatus> {
    try {
      const stats = await fs.stat(targetPath);
      const permissions = await this.getPermissionFlags(targetPath);
      const octalMode = this.getOctalMode(stats);

      return {
        path: targetPath,
        readable: permissions.readable,
        writable: permissions.writable,
        executable: permissions.executable,
        isOwner: this.isOwner(stats),
        ...(octalMode && { octalMode })
      };
    } catch (error) {
      throw new Error(
        `Failed to check permissions for ${targetPath}: ${
          error instanceof Error ? error.message : String(error)
        }`
      );
    }
  }

  /**
   * Validate if path has required permissions
   * @param targetPath - Path to validate
   * @param required - Required permissions
   * @returns True if permissions are sufficient
   */
  async validatePermissions(targetPath: string, required: FilePermissions): Promise<boolean> {
    try {
      const status = await this.checkPermissions(targetPath);

      // For current user, check if we have the required access
      if (required.owner.read && !status.readable) return false;
      if (required.owner.write && !status.writable) return false;
      if (required.owner.execute && !status.executable) return false;

      return true;
    } catch {
      return false;
    }
  }

  /**
   * Fix permissions for a MoAI project
   * @param projectPath - Path to the project
   * @returns Result of the fix operation
   */
  async fixPermissions(projectPath: string): Promise<PermissionFixResult> {
    const result = {
      fixedFiles: [] as string[],
      failures: [] as string[],
      warnings: [] as string[]
    };

    try {
      await this.fixDirectoryPermissions(projectPath, result);

      return {
        success: result.failures.length === 0,
        fixedFiles: result.fixedFiles,
        failures: result.failures,
        warnings: result.warnings
      };
    } catch (error) {
      result.failures.push(
        `Failed to fix permissions: ${error instanceof Error ? error.message : String(error)}`
      );

      return {
        success: false,
        fixedFiles: result.fixedFiles,
        failures: result.failures,
        warnings: result.warnings
      };
    }
  }

  // Private helper methods below

  private async setUnixPermissions(targetPath: string, permissions: FilePermissions): Promise<void> {
    const octal = PermissionUtils.permissionsToOctal(permissions);
    await fs.chmod(targetPath, parseInt(octal, 8));
  }

  private async setWindowsPermissions(targetPath: string, permissions: FilePermissions): Promise<void> {
    try {
      const stats = await fs.stat(targetPath);

      if (!permissions.owner.write && stats.mode & 0o200) {
        // Remove write permission (make read-only)
        await fs.chmod(targetPath, stats.mode & ~0o200);
      } else if (permissions.owner.write && !(stats.mode & 0o200)) {
        // Add write permission
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

  private async makeUnixExecutable(filePath: string): Promise<void> {
    const stats = await fs.stat(filePath);
    const currentMode = stats.mode;
    const executableMode = currentMode | 0o111; // Add execute permissions for all
    await fs.chmod(filePath, executableMode);
  }

  private async makeWindowsExecutable(filePath: string): Promise<void> {
    // Windows: Execution is determined by file extension
    if (!PermissionUtils.isScript(filePath)) {
      console.warn(
        `Warning: ${filePath} may not be executable on Windows without proper extension`
      );
    }
  }

  private async getPermissionFlags(targetPath: string): Promise<{
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

  private getOctalMode(stats: fs.Stats): string | undefined {
    if (this.platform === 'unix') {
      return (stats.mode & parseInt('777', 8)).toString(8);
    }
    return undefined;
  }

  private isOwner(stats: fs.Stats): boolean {
    return stats.uid === process.getuid?.() || this.platform === 'windows';
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
      const permissions = this.determineFilePermissions(filePath, result);
      await this.setFilePermissions(filePath, permissions);
      result.fixedFiles.push(filePath);
    } catch (error) {
      result.failures.push(
        `File ${filePath}: ${error instanceof Error ? error.message : String(error)}`
      );
    }
  }

  private determineFilePermissions(
    filePath: string,
    result: { warnings: string[] }
  ): FilePermissions {
    if (PermissionUtils.isSensitive(filePath)) {
      result.warnings.push(`Applied strict permissions to sensitive file: ${filePath}`);
      return MoAIPermissions.SENSITIVE_FILES;
    } else if (PermissionUtils.isScript(filePath)) {
      return MoAIPermissions.SCRIPT_FILES;
    } else if (PermissionUtils.isConfig(filePath)) {
      return MoAIPermissions.CONFIG_FILES;
    } else {
      return MoAIPermissions.CONFIG_FILES; // Default to config file permissions
    }
  }
}

/**
 * Default export for convenience
 */
export default PermissionManager;