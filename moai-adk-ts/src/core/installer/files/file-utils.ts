/**
 * @file Cross-platform File Security Utilities
 * @author MoAI Team
 * @tags @API:FILE-UTILS-001 @FEATURE:PATH-SECURITY-001 @TASK:PATH-UTILS-SEPARATION-001
 * @description Path security and utility functions for cross-platform compatibility
 */

import * as path from 'path';

/**
 * File security and path utility functions
 * @tags @API:FILE-UTILS-001
 */
export class FileUtils {
  /**
   * Validate path safety against directory traversal attacks
   * @param targetPath - Path to validate
   * @param projectRoot - Project root directory for validation
   * @returns True if path is safe, false otherwise
   * @tags @API:VALIDATE-PATH-SAFETY-001
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
   * @param fileName - Original filename
   * @returns Sanitized filename safe for all platforms
   * @tags @API:SANITIZE-FILENAME-001
   */
  static sanitizeFileName(fileName: string): string {
    if (!fileName) {
      return '';
    }

    // Replace dangerous characters with safe alternatives
    let sanitized = fileName
      .replace(/[<>:"|?*]/g, '-') // Windows reserved chars
      // eslint-disable-next-line no-control-regex
      .replace(/[\u0000-\u001F\u0080-\u009F]/g, '') // Control characters (Unicode escape)
      .replace(/[\u{1F600}-\u{1F64F}]/gu, '-') // Emoticons
      .replace(/[\u{1F300}-\u{1F5FF}]/gu, '-') // Symbols
      .replace(/[\u{1F680}-\u{1F6FF}]/gu, '-') // Transport
      .replace(/[\u{2600}-\u{26FF}]/gu, '-') // Misc symbols
      .replace(/[^\w\-_.]/g, '-'); // Non-word chars except safe ones

    // Remove multiple consecutive dashes
    sanitized = sanitized.replace(/-+/g, '-');

    // Remove leading/trailing dashes
    sanitized = sanitized.replace(/^-+|-+$/g, '');

    return sanitized;
  }

  /**
   * Get platform-specific path separator
   * @returns Platform path separator
   * @tags @API:GET-PATH-SEPARATOR-001
   */
  static getPathSeparator(): string {
    return path.sep;
  }

  /**
   * Join path components in a platform-safe way
   * @param components - Path components to join
   * @returns Joined path
   * @tags @API:JOIN-PATH-001
   */
  static joinPath(...components: string[]): string {
    return path.join(...components);
  }

  /**
   * Resolve path to absolute path
   * @param targetPath - Path to resolve
   * @returns Absolute path
   * @tags @API:RESOLVE-PATH-001
   */
  static resolvePath(targetPath: string): string {
    return path.resolve(targetPath);
  }

  /**
   * Get relative path between two paths
   * @param from - From path
   * @param to - To path
   * @returns Relative path
   * @tags @API:GET-RELATIVE-PATH-001
   */
  static getRelativePath(from: string, to: string): string {
    return path.relative(from, to);
  }

  /**
   * Check if path is absolute
   * @param targetPath - Path to check
   * @returns True if path is absolute
   * @tags @API:IS-ABSOLUTE-001
   */
  static isAbsolute(targetPath: string): boolean {
    return path.isAbsolute(targetPath);
  }

  /**
   * Get directory name from path
   * @param targetPath - Path to extract directory from
   * @returns Directory path
   * @tags @API:DIRNAME-001
   */
  static dirname(targetPath: string): string {
    return path.dirname(targetPath);
  }

  /**
   * Get base name from path
   * @param targetPath - Path to extract basename from
   * @param ext - Optional extension to remove
   * @returns Base name
   * @tags @API:BASENAME-001
   */
  static basename(targetPath: string, ext?: string): string {
    return path.basename(targetPath, ext);
  }

  /**
   * Get file extension from path
   * @param targetPath - Path to extract extension from
   * @returns File extension including the dot
   * @tags @API:EXTNAME-001
   */
  static extname(targetPath: string): string {
    return path.extname(targetPath);
  }
}
