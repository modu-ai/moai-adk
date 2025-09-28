/**
 * @file Permission utility functions
 * @author MoAI Team
 * @tags @DESIGN:PERMISSION-UTILS-012 @REQ:CROSS-PLATFORM-PERMISSIONS-012
 */

import * as path from 'path';
import { FilePermissions, PlatformType } from './permission-types';

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