/**
 * @API:FILE-OPERATIONS-001 Cross-platform File System Operations
 * @FEATURE:FS-EXTRA-INTEGRATION-001 fs-extra based file system operations
 *
 * SPEC-012 Week 2 Track C-1: Complete fs-extra porting of Python file operations
 * Provides cross-platform file system operations with security validation.
 *
 * @TASK:FILE-OPS-TYPESCRIPT-001 TypeScript 파일 시스템 작업 포팅
 * @DESIGN:CROSS-PLATFORM-001 Windows/macOS/Linux 크로스 플랫폼 지원
 * @SEC:PATH-TRAVERSAL-001 디렉토리 순회 공격 방어
 *
 * @author MoAI Team
 * @version 0.0.1
 * @since 2025-01-07
 */

import * as path from 'path';
import { FileUtils } from './file-utils';
import { FileIO } from './file-io';

/**
 * File statistics information interface
 */
export interface FileStats {
  readonly size: number;
  readonly isFile: boolean;
  readonly isDirectory: boolean;
  readonly modified: Date;
  readonly permissions: string;
}

/**
 * Copy operation result interface
 */
export interface CopyResult {
  readonly success: boolean;
  readonly copiedFiles: readonly string[];
  readonly errors: readonly string[];
}

/**
 * File operation options interface
 */
export interface FileOperationOptions {
  readonly overwrite?: boolean;
  readonly preserveTimestamps?: boolean;
  readonly recursive?: boolean;
  readonly mode?: number;
}

/**
 * Progress callback function type
 */
export type ProgressCallback = (
  current: number,
  total: number,
  file: string
) => void;

/**
 * FileOperations class - Cross-platform file system operations manager
 *
 * Features:
 * - Full fs-extra integration with async/await patterns
 * - Cross-platform path handling (Windows/macOS/Linux)
 * - Security validation and path sanitization
 * - Progress tracking for large operations
 * - Comprehensive error handling with detailed messages
 *
 * Security Features:
 * - Directory traversal attack prevention
 * - Path safety validation against project root
 * - Filename sanitization for cross-platform compatibility
 * - Permission validation and safe defaults
 */
export class FileOperations {
  /**
   * Ensure directory exists with proper permissions
   */
  async ensureDirectory(dirPath: string, mode: number = 0o755): Promise<void> {
    return FileIO.ensureDirectory(dirPath, mode);
  }

  /**
   * Copy a single file with overwrite option
   */
  async copyFile(
    src: string,
    dst: string,
    overwrite: boolean = false
  ): Promise<void> {
    return FileIO.copyFile(src, dst, overwrite);
  }

  /**
   * Copy directory recursively with file tracking
   */
  async copyDirectory(
    src: string,
    dst: string,
    overwrite: boolean = false,
    onProgress?: ProgressCallback
  ): Promise<string[]> {
    try {
      // Check if source directory exists
      if (!(await FileIO.pathExists(src))) {
        throw new Error(`Source directory does not exist: ${src}`);
      }

      const srcStats = await FileIO.getFileStats(src);
      if (!srcStats.isDirectory) {
        throw new Error(`Source is not a directory: ${src}`);
      }

      const copiedFiles: string[] = [];
      const allFiles = await FileIO.getAllFilesRecursively(src);

      for (let i = 0; i < allFiles.length; i++) {
        const srcFile = allFiles[i]!;
        const relativePath = path.relative(src, srcFile);
        const dstFile = path.join(dst, relativePath);

        await FileIO.copyFile(srcFile, dstFile, overwrite);
        copiedFiles.push(dstFile);

        if (onProgress) {
          onProgress(i + 1, allFiles.length, dstFile);
        }
      }

      return copiedFiles;
    } catch (error) {
      throw new Error(
        `Failed to copy directory from '${src}' to '${dst}': ${error}`
      );
    }
  }

  /**
   * Safely remove a file
   */
  async removeFile(filePath: string): Promise<void> {
    return FileIO.removeFile(filePath);
  }

  /**
   * Safely remove directory recursively
   */
  async removeDirectory(dirPath: string): Promise<void> {
    return FileIO.removeDirectory(dirPath);
  }

  /**
   * Read file content as UTF-8 string
   */
  async readFileContent(filePath: string): Promise<string> {
    return FileIO.readFileContent(filePath);
  }

  /**
   * Write content to file as UTF-8
   */
  async writeFileContent(
    filePath: string,
    content: string,
    options?: FileOperationOptions
  ): Promise<void> {
    return FileIO.writeFileContent(filePath, content, options);
  }

  /**
   * Check if path exists
   */
  async pathExists(targetPath: string): Promise<boolean> {
    return FileIO.pathExists(targetPath);
  }

  /**
   * Get file statistics
   */
  async getFileStats(filePath: string): Promise<FileStats> {
    return FileIO.getFileStats(filePath);
  }

  /**
   * Validate path safety against directory traversal attacks
   *
   * @param targetPath - Path to validate
   * @param projectRoot - Project root directory for validation
   * @returns True if path is safe, false otherwise
   */
  validatePathSafety(targetPath: string, projectRoot: string): boolean {
    return FileUtils.validatePathSafety(targetPath, projectRoot);
  }

  /**
   * Sanitize filename for cross-platform compatibility
   *
   * @param fileName - Original filename
   * @returns Sanitized filename safe for all platforms
   */
  sanitizeFileName(fileName: string): string {
    return FileUtils.sanitizeFileName(fileName);
  }

  /**
   * Get platform-specific path separator
   */
  getPathSeparator(): string {
    return FileUtils.getPathSeparator();
  }

  /**
   * Join path components in a platform-safe way
   */
  joinPath(...components: string[]): string {
    return FileUtils.joinPath(...components);
  }

  /**
   * Resolve path to absolute path
   */
  resolvePath(targetPath: string): string {
    return FileUtils.resolvePath(targetPath);
  }

  /**
   * Get relative path between two paths
   */
  getRelativePath(from: string, to: string): string {
    return FileUtils.getRelativePath(from, to);
  }
}

/**
 * Default FileOperations instance for convenience
 */
export const fileOperations = new FileOperations();

/**
 * Create a new FileOperations instance
 *
 * @returns New FileOperations instance
 */
export function createFileOperations(): FileOperations {
  return new FileOperations();
}
