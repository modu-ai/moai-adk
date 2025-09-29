/**
 * @API:FILE-IO-001 Core File I/O Operations
 * @FEATURE:FS-EXTRA-CORE-001 fs-extra 기반 핵심 파일 작업
 *
 * SPEC-012 Week 2 Track C-1: Basic file I/O operations
 * Separated from main FileOperations class to maintain TRUST principle (Under 300 LOC)
 *
 * @TASK:FILE-IO-SEPARATION-001 파일 I/O 클래스 분리
 * @DESIGN:TRUST-COMPLIANCE-001 TRUST 원칙 준수 (300 LOC 이하)
 * @PERF:ASYNC-IO-001 비동기 파일 I/O 최적화
 *
 * @author MoAI Team
 * @version 0.0.1
 * @since 2025-01-07
 */

import * as os from 'node:os';
import * as path from 'node:path';
import { promises as fs, constants } from 'node:fs';
import type { FileOperationOptions, FileStats } from './file-operations';

/**
 * Core file I/O operations
 */
export class FileIO {
  /**
   * Ensure directory exists with proper permissions
   *
   * @param dirPath - Target directory path
   * @param mode - Directory permissions (default: 0o755)
   * @returns Promise that resolves when directory is created
   */
  static async ensureDirectory(
    dirPath: string,
    mode: number = 0o755
  ): Promise<void> {
    try {
      await fs.mkdir(dirPath, { recursive: true, mode });
    } catch (error) {
      throw new Error(`Failed to ensure directory '${dirPath}': ${error}`);
    }
  }

  /**
   * Copy a single file with overwrite option
   *
   * @param src - Source file path
   * @param dst - Destination file path
   * @param overwrite - Whether to overwrite existing files
   * @returns Promise that resolves when file is copied
   */
  static async copyFile(
    src: string,
    dst: string,
    overwrite: boolean = false
  ): Promise<void> {
    try {
      // Check if source exists
      try {
        await fs.access(src, constants.F_OK);
      } catch {
        throw new Error(`Source file does not exist: ${src}`);
      }

      // Check if destination exists and handle overwrite
      if (!overwrite) {
        try {
          await fs.access(dst, constants.F_OK);
          throw new Error(
            `Destination file already exists and overwrite is false: ${dst}`
          );
        } catch (error: any) {
          if (error.code !== 'ENOENT') {
            throw error;
          }
        }
      }

      // Ensure destination directory exists
      const dstDir = path.dirname(dst);
      await fs.mkdir(dstDir, { recursive: true });

      // Copy the file
      await fs.copyFile(src, dst);
    } catch (error) {
      throw new Error(
        `Failed to copy file from '${src}' to '${dst}': ${error}`
      );
    }
  }

  /**
   * Safely remove a file
   *
   * @param filePath - Path to file to remove
   * @returns Promise that resolves when file is removed
   */
  static async removeFile(filePath: string): Promise<void> {
    try {
      await fs.unlink(filePath);
    } catch (error: any) {
      // Ignore if file doesn't exist
      if (error.code !== 'ENOENT') {
        throw new Error(`Failed to remove file '${filePath}': ${error}`);
      }
    }
  }

  /**
   * Safely remove directory recursively
   *
   * @param dirPath - Path to directory to remove
   * @returns Promise that resolves when directory is removed
   */
  static async removeDirectory(dirPath: string): Promise<void> {
    try {
      await fs.rmdir(dirPath, { recursive: true });
    } catch (error: any) {
      // Ignore if directory doesn't exist
      if (error.code !== 'ENOENT') {
        throw new Error(`Failed to remove directory '${dirPath}': ${error}`);
      }
    }
  }

  /**
   * Read file content as UTF-8 string
   *
   * @param filePath - Path to file to read
   * @returns Promise that resolves to file content
   */
  static async readFileContent(filePath: string): Promise<string> {
    try {
      return await fs.readFile(filePath, 'utf8');
    } catch (error) {
      throw new Error(`Failed to read file '${filePath}': ${error}`);
    }
  }

  /**
   * Write content to file as UTF-8
   *
   * @param filePath - Path to file to write
   * @param content - Content to write
   * @param options - Write options
   * @returns Promise that resolves when file is written
   */
  static async writeFileContent(
    filePath: string,
    content: string,
    options?: FileOperationOptions
  ): Promise<void> {
    try {
      // Ensure parent directory exists
      const dir = path.dirname(filePath);
      await fs.mkdir(dir, { recursive: true });

      // Write file with UTF-8 encoding
      await fs.writeFile(filePath, content, 'utf8');

      // Set permissions if specified and on Unix
      if (options?.mode && os.platform() !== 'win32') {
        await fs.chmod(filePath, options.mode);
      }
    } catch (error) {
      throw new Error(`Failed to write file '${filePath}': ${error}`);
    }
  }

  /**
   * Check if path exists
   *
   * @param targetPath - Path to check
   * @returns Promise that resolves to true if path exists
   */
  static async pathExists(targetPath: string): Promise<boolean> {
    return fs.pathExists(targetPath);
  }

  /**
   * Get file statistics
   *
   * @param filePath - Path to file
   * @returns Promise that resolves to file stats
   */
  static async getFileStats(filePath: string): Promise<FileStats> {
    try {
      const stats = await fs.stat(filePath);

      return {
        size: stats.size,
        isFile: stats.isFile(),
        isDirectory: stats.isDirectory(),
        modified: stats.mtime,
        permissions: stats.mode.toString(8),
      };
    } catch (error) {
      throw new Error(`Failed to get stats for '${filePath}': ${error}`);
    }
  }

  /**
   * Recursively get all files in a directory
   *
   * @param dirPath - Directory path to scan
   * @returns Promise that resolves to array of file paths
   */
  static async getAllFilesRecursively(dirPath: string): Promise<string[]> {
    const files: string[] = [];
    const items = await fs.readdir(dirPath);

    for (const item of items) {
      const fullPath = path.join(dirPath, item);
      const stats = await fs.stat(fullPath);

      if (stats.isFile()) {
        files.push(fullPath);
      } else if (stats.isDirectory()) {
        const subFiles = await FileIO.getAllFilesRecursively(fullPath);
        files.push(...subFiles);
      }
    }

    return files;
  }
}
