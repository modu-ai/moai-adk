/**
 * @fileoverview File Utilities - Cross-platform path and security utilities
 *
 * SPEC-012 Week 2 Track C-1: Security and path utility functions
 * Separated from main FileOperations class to maintain TRUST principle (Under 300 LOC)
 *
 * @author MoAI Team
 * @version 0.0.1
 * @since 2025-01-07
 */

import * as path from 'path';

/**
 * File security and path utility functions
 */
export class FileUtils {
  /**
   * Validate path safety against directory traversal attacks
   *
   * @param targetPath - Path to validate
   * @param projectRoot - Project root directory for validation
   * @returns True if path is safe, false otherwise
   *
   * @example
   * ```typescript
   * const isSafe = FileUtils.validatePathSafety('/project/file.txt', '/project');
   * // Returns true
   *
   * const isUnsafe = FileUtils.validatePathSafety('/project/../../../etc/passwd', '/project');
   * // Returns false
   * ```
   */
  static validatePathSafety(targetPath: string, projectRoot: string): boolean {
    try {
      // Handle relative paths by resolving them against the project root
      const resolvedTarget = path.isAbsolute(targetPath)
        ? path.resolve(targetPath)
        : path.resolve(projectRoot, targetPath);
      const resolvedRoot = path.resolve(projectRoot);

      // Check if the target path is within the project root
      const relativePath = path.relative(resolvedRoot, resolvedTarget);

      // If relative path starts with '..' or is absolute, it's outside the root
      return !relativePath.startsWith('..') && !path.isAbsolute(relativePath);
    } catch (error) {
      // If path resolution fails, consider it unsafe
      return false;
    }
  }

  /**
   * Sanitize filename for cross-platform compatibility
   *
   * @param fileName - Original filename
   * @returns Sanitized filename safe for all platforms
   *
   * @example
   * ```typescript
   * const safe = FileUtils.sanitizeFileName('file<name>?.txt');
   * // Returns 'file-name-.txt'
   * ```
   */
  static sanitizeFileName(fileName: string): string {
    if (!fileName) {
      return '';
    }

    // Replace dangerous characters with safe alternatives
    let sanitized = fileName
      .replace(/[<>:"|?*]/g, '-')    // Windows reserved chars
      // eslint-disable-next-line no-control-regex
      .replace(/[\u0000-\u001F\u0080-\u009F]/g, '') // Control characters (Unicode escape)
      .replace(/[\u{1F600}-\u{1F64F}]/gu, '-') // Emoticons
      .replace(/[\u{1F300}-\u{1F5FF}]/gu, '-') // Symbols
      .replace(/[\u{1F680}-\u{1F6FF}]/gu, '-') // Transport
      .replace(/[\u{2600}-\u{26FF}]/gu, '-')  // Misc symbols
      .replace(/[^\w\-_.]/g, '-');   // Non-word chars except safe ones

    // Remove multiple consecutive dashes
    sanitized = sanitized.replace(/-+/g, '-');

    // Remove leading/trailing dashes
    sanitized = sanitized.replace(/^-+|-+$/g, '');

    return sanitized;
  }

  /**
   * Get platform-specific path separator
   *
   * @returns Platform path separator
   */
  static getPathSeparator(): string {
    return path.sep;
  }

  /**
   * Join path components in a platform-safe way
   *
   * @param components - Path components to join
   * @returns Joined path
   *
   * @example
   * ```typescript
   * const fullPath = FileUtils.joinPath('project', 'src', 'index.ts');
   * // Returns 'project/src/index.ts' on Unix or 'project\\src\\index.ts' on Windows
   * ```
   */
  static joinPath(...components: string[]): string {
    return path.join(...components);
  }

  /**
   * Resolve path to absolute path
   *
   * @param targetPath - Path to resolve
   * @returns Absolute path
   */
  static resolvePath(targetPath: string): string {
    return path.resolve(targetPath);
  }

  /**
   * Get relative path between two paths
   *
   * @param from - From path
   * @param to - To path
   * @returns Relative path
   */
  static getRelativePath(from: string, to: string): string {
    return path.relative(from, to);
  }

  /**
   * Check if path is absolute
   *
   * @param targetPath - Path to check
   * @returns True if path is absolute
   */
  static isAbsolute(targetPath: string): boolean {
    return path.isAbsolute(targetPath);
  }

  /**
   * Get directory name from path
   *
   * @param targetPath - Path to extract directory from
   * @returns Directory path
   */
  static dirname(targetPath: string): string {
    return path.dirname(targetPath);
  }

  /**
   * Get base name from path
   *
   * @param targetPath - Path to extract basename from
   * @param ext - Optional extension to remove
   * @returns Base name
   */
  static basename(targetPath: string, ext?: string): string {
    return path.basename(targetPath, ext);
  }

  /**
   * Get file extension from path
   *
   * @param targetPath - Path to extract extension from
   * @returns File extension including the dot
   */
  static extname(targetPath: string): string {
    return path.extname(targetPath);
  }
}