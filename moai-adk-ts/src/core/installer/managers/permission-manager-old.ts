/**
 * @file Cross-platform permission management system
 * @author MoAI Team
 * @tags @DESIGN:PERMISSION-MANAGER-012 @REQ:CROSS-PLATFORM-PERMISSIONS-012
 */

import * as fs from 'fs-extra';
import * as path from 'path';

/**
 * File permission structure interface
 * @tags @DESIGN:FILE-PERMISSIONS-001
 */
export interface FilePermissions {
  readonly owner: {
    readonly read: boolean;
    readonly write: boolean;
    readonly execute: boolean;
  };
  readonly group: {
    readonly read: boolean;
    readonly write: boolean;
    readonly execute: boolean;
  };
  readonly others: {
    readonly read: boolean;
    readonly write: boolean;
    readonly execute: boolean;
  };
}

/**
 * Permission status for a file or directory
 * @tags @DESIGN:PERMISSION-STATUS-001
 */
export interface PermissionStatus {
  readonly path: string;
  readonly readable: boolean;
  readonly writable: boolean;
  readonly executable: boolean;
  readonly octalMode?: string; // Unix systems only
  readonly isOwner: boolean;
}

/**
 * Result of permission fix operation
 * @tags @DESIGN:PERMISSION-FIX-RESULT-001
 */
export interface PermissionFixResult {
  readonly success: boolean;
  readonly fixedFiles: readonly string[];
  readonly failures: readonly string[];
  readonly warnings: readonly string[];
}

/**
 * Platform type detection
 * @tags @DESIGN:PLATFORM-TYPE-001
 */
export type PlatformType = 'windows' | 'unix';

/**
 * Standard MoAI permission policies
 * @tags @DESIGN:MOAI-PERMISSIONS-001
 */
export const MoAIPermissions = {
  SCRIPT_FILES: { // .py, .sh, .js files
    owner: { read: true, write: true, execute: true },
    group: { read: true, write: false, execute: true },
    others: { read: true, write: false, execute: false }
  },
  CONFIG_FILES: { // .json, .yaml files
    owner: { read: true, write: true, execute: false },
    group: { read: true, write: false, execute: false },
    others: { read: false, write: false, execute: false }
  },
  DIRECTORIES: {
    owner: { read: true, write: true, execute: true },
    group: { read: true, write: false, execute: true },
    others: { read: true, write: false, execute: true }
  },
  SENSITIVE_FILES: { // secret files
    owner: { read: true, write: true, execute: false },
    group: { read: false, write: false, execute: false },
    others: { read: false, write: false, execute: false }
  }
} as const;

/**
 * Permission utility helper class
 * @tags @DESIGN:PERMISSION-UTILS-001
 */
export class PermissionUtils {
  /**
   * Convert octal permission string to FilePermissions object
   * @param octal - Octal permission string (e.g., "755")
   * @returns FilePermissions object
   */
  static octalToPermissions(octal: string): FilePermissions {
    if (!/^[0-7]{3}$/.test(octal)) {
      throw new Error(`Invalid octal permission: ${octal}`);
    }

    const digits = octal.split('').map(digit => parseInt(digit, 8));

    if (digits.length !== 3) {
      throw new Error(`Invalid octal permission: ${octal}`);
    }

    const owner = digits[0]!;
    const group = digits[1]!;
    const others = digits[2]!;

    const parseOctal = (value: number) => ({
      read: (value & 4) !== 0,
      write: (value & 2) !== 0,
      execute: (value & 1) !== 0
    });

    return {
      owner: parseOctal(owner),
      group: parseOctal(group),
      others: parseOctal(others)
    };
  }

  /**
   * Convert FilePermissions object to octal string
   * @param permissions - FilePermissions object
   * @returns Octal permission string
   */
  static permissionsToOctal(permissions: FilePermissions): string {
    const toOctal = (perm: { read: boolean; write: boolean; execute: boolean }) => {
      let value = 0;
      if (perm.read) value += 4;
      if (perm.write) value += 2;
      if (perm.execute) value += 1;
      return value;
    };

    const owner = toOctal(permissions.owner);
    const group = toOctal(permissions.group);
    const others = toOctal(permissions.others);

    return `${owner}${group}${others}`;
  }

  /**
   * Check if file is a script file
   * @param filePath - Path to check
   * @returns True if it's a script file
   */
  static isScript(filePath: string): boolean {
    const ext = path.extname(filePath).toLowerCase();
    return ['.py', '.sh', '.js', '.ts', '.bat', '.cmd', '.ps1'].includes(ext);
  }

  /**
   * Check if file is a configuration file
   * @param filePath - Path to check
   * @returns True if it's a config file
   */
  static isConfig(filePath: string): boolean {
    const ext = path.extname(filePath).toLowerCase();
    const basename = path.basename(filePath).toLowerCase();

    return ['.json', '.yaml', '.yml', '.toml', '.ini', '.conf'].includes(ext) ||
           basename.startsWith('.env') ||
           basename.includes('config');
  }

  /**
   * Check if file contains sensitive information
   * @param filePath - Path to check
   * @returns True if it's a sensitive file
   */
  static isSensitive(filePath: string): boolean {
    const basename = path.basename(filePath).toLowerCase();
    const dirname = path.dirname(filePath).toLowerCase();

    const sensitivePatterns = [
      '.env', 'secret', 'credential', 'key', 'token', 'password',
      'auth', 'private', '.ssh', '.pem', '.p12', '.pfx'
    ];

    return sensitivePatterns.some(pattern =>
      basename.includes(pattern) || dirname.includes(pattern)
    );
  }

  /**
   * Get current platform type
   * @returns Platform type
   */
  static getCurrentPlatform(): PlatformType {
    return process.platform === 'win32' ? 'windows' : 'unix';
  }
}

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
        const octal = PermissionUtils.permissionsToOctal(permissions);
        await fs.chmod(filePath, parseInt(octal, 8));
      } else {
        // Windows: Limited chmod support, handle gracefully
        await this.setWindowsPermissions(filePath, permissions);
      }
    } catch (error) {
      throw new Error(`Failed to set permissions for ${filePath}: ${error instanceof Error ? error.message : String(error)}`);
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
        const octal = PermissionUtils.permissionsToOctal(permissions);
        await fs.chmod(dirPath, parseInt(octal, 8));
      } else {
        // Windows: Handle directory permissions
        await this.setWindowsPermissions(dirPath, permissions);
      }
    } catch (error) {
      throw new Error(`Failed to set directory permissions for ${dirPath}: ${error instanceof Error ? error.message : String(error)}`);
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
        const stats = await fs.stat(filePath);
        const currentMode = stats.mode;
        const executableMode = currentMode | 0o111; // Add execute permissions for all
        await fs.chmod(filePath, executableMode);
      } else {
        // Windows: Execution is determined by file extension
        if (!PermissionUtils.isScript(filePath)) {
          console.warn(`Warning: ${filePath} may not be executable on Windows without proper extension`);
        }
      }
    } catch (error) {
      throw new Error(`Failed to make ${filePath} executable: ${error instanceof Error ? error.message : String(error)}`);
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

      let readable = false;
      let writable = false;
      let executable = false;
      let octalMode: string | undefined;

      try {
        await fs.access(targetPath, fs.constants.R_OK);
        readable = true;
      } catch {
        // Not readable
      }

      try {
        await fs.access(targetPath, fs.constants.W_OK);
        writable = true;
      } catch {
        // Not writable
      }

      try {
        await fs.access(targetPath, fs.constants.X_OK);
        executable = true;
      } catch {
        // Not executable
      }

      if (this.platform === 'unix') {
        octalMode = (stats.mode & parseInt('777', 8)).toString(8);
      }

      const result: PermissionStatus = {
        path: targetPath,
        readable,
        writable,
        executable,
        isOwner: stats.uid === process.getuid?.() || this.platform === 'windows'
      };

      if (octalMode !== undefined) {
        (result as any).octalMode = octalMode;
      }

      return result;
    } catch (error) {
      throw new Error(`Failed to check permissions for ${targetPath}: ${error instanceof Error ? error.message : String(error)}`);
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
    const fixedFiles: string[] = [];
    const failures: string[] = [];
    const warnings: string[] = [];

    try {
      await this.fixDirectoryPermissions(projectPath, fixedFiles, failures, warnings);

      return {
        success: failures.length === 0,
        fixedFiles,
        failures,
        warnings
      };
    } catch (error) {
      failures.push(`Failed to fix permissions: ${error instanceof Error ? error.message : String(error)}`);

      return {
        success: false,
        fixedFiles,
        failures,
        warnings
      };
    }
  }

  /**
   * Handle Windows-specific permission setting
   * @param targetPath - Path to set permissions for
   * @param permissions - Desired permissions
   * @private
   */
  private async setWindowsPermissions(targetPath: string, permissions: FilePermissions): Promise<void> {
    // On Windows, we have limited chmod functionality
    // Focus on basic read/write permissions
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
      console.warn(`Windows permission setting limited for ${targetPath}: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * Recursively fix directory permissions
   * @param dirPath - Directory path
   * @param fixedFiles - Array to track fixed files
   * @param failures - Array to track failures
   * @param warnings - Array to track warnings
   * @private
   */
  private async fixDirectoryPermissions(
    dirPath: string,
    fixedFiles: string[],
    failures: string[],
    warnings: string[]
  ): Promise<void> {
    try {
      const entries = await fs.readdir(dirPath, { withFileTypes: true });

      // Set directory permissions
      try {
        await this.setDirectoryPermissions(dirPath, MoAIPermissions.DIRECTORIES);
        fixedFiles.push(dirPath);
      } catch (error) {
        failures.push(`Directory ${dirPath}: ${error instanceof Error ? error.message : String(error)}`);
      }

      for (const entry of entries) {
        const fullPath = path.join(dirPath, entry.name);

        if (entry.isDirectory()) {
          await this.fixDirectoryPermissions(fullPath, fixedFiles, failures, warnings);
        } else {
          await this.fixFilePermissions(fullPath, fixedFiles, failures, warnings);
        }
      }
    } catch (error) {
      failures.push(`Failed to read directory ${dirPath}: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * Fix permissions for a single file
   * @param filePath - File path
   * @param fixedFiles - Array to track fixed files
   * @param failures - Array to track failures
   * @param warnings - Array to track warnings
   * @private
   */
  private async fixFilePermissions(
    filePath: string,
    fixedFiles: string[],
    failures: string[],
    warnings: string[]
  ): Promise<void> {
    try {
      let permissions: FilePermissions;

      if (PermissionUtils.isSensitive(filePath)) {
        permissions = MoAIPermissions.SENSITIVE_FILES;
        warnings.push(`Applied strict permissions to sensitive file: ${filePath}`);
      } else if (PermissionUtils.isScript(filePath)) {
        permissions = MoAIPermissions.SCRIPT_FILES;
      } else if (PermissionUtils.isConfig(filePath)) {
        permissions = MoAIPermissions.CONFIG_FILES;
      } else {
        permissions = MoAIPermissions.CONFIG_FILES; // Default to config file permissions
      }

      await this.setFilePermissions(filePath, permissions);
      fixedFiles.push(filePath);
    } catch (error) {
      failures.push(`File ${filePath}: ${error instanceof Error ? error.message : String(error)}`);
    }
  }
}

/**
 * Default export for convenience
 */
export default PermissionManager;