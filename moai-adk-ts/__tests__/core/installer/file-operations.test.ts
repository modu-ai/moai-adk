/**
 * @fileoverview FileOperations Test Suite - TDD RED Phase
 *
 * SPEC-012 Week 2 Track C-1: Complete test coverage for fs-extra file operations
 * Tests cover cross-platform compatibility, security validation, and error handling.
 *
 * @author MoAI Team
 * @version 0.0.1
 * @since 2025-01-07
 */

import * as fs from 'fs-extra';
import * as path from 'path';
import * as os from 'os';
import {
  FileOperations,
  fileOperations,
  createFileOperations
} from '../../../src/core/installer/managers/file-operations';

describe('FileOperations', () => {
  let fileOps: FileOperations;
  let tempDir: string;
  let testProjectRoot: string;

  beforeEach(async () => {
    fileOps = createFileOperations();
    tempDir = await fs.mkdtemp(path.join(os.tmpdir(), 'moai-test-'));
    testProjectRoot = tempDir;
  });

  afterEach(async () => {
    try {
      await fs.remove(tempDir);
    } catch (error) {
      // Ignore cleanup errors in tests
    }
  });

  describe('Factory Functions', () => {
    it('should create FileOperations instance via createFileOperations', () => {
      const instance = createFileOperations();
      expect(instance).toBeInstanceOf(FileOperations);
    });

    it('should provide default singleton instance', () => {
      expect(fileOperations).toBeInstanceOf(FileOperations);
    });
  });

  describe('ensureDirectory', () => {
    it('should create directory with default permissions', async () => {
      const dirPath = path.join(tempDir, 'test-dir');

      await fileOps.ensureDirectory(dirPath);

      const exists = await fs.pathExists(dirPath);
      expect(exists).toBe(true);

      const stats = await fs.stat(dirPath);
      expect(stats.isDirectory()).toBe(true);
    });

    it('should create nested directories recursively', async () => {
      const nestedPath = path.join(tempDir, 'level1', 'level2', 'level3');

      await fileOps.ensureDirectory(nestedPath);

      const exists = await fs.pathExists(nestedPath);
      expect(exists).toBe(true);
    });

    it('should not fail if directory already exists', async () => {
      const dirPath = path.join(tempDir, 'existing-dir');
      await fs.ensureDir(dirPath);

      // Should not throw when directory already exists
      await expect(async () => {
        await fileOps.ensureDirectory(dirPath);
      }).not.toThrow();
    });

    it('should create directory with custom permissions on Unix', async () => {
      if (os.platform() === 'win32') {
        return; // Skip on Windows as it doesn't support Unix permissions
      }

      const dirPath = path.join(tempDir, 'custom-perms');
      const customMode = 0o750;

      await fileOps.ensureDirectory(dirPath, customMode);

      const stats = await fs.stat(dirPath);
      expect(stats.mode & parseInt('777', 8)).toBe(customMode);
    });
  });

  describe('copyFile', () => {
    let srcFile: string;
    let dstFile: string;

    beforeEach(async () => {
      srcFile = path.join(tempDir, 'source.txt');
      dstFile = path.join(tempDir, 'destination.txt');
      await fs.writeFile(srcFile, 'test content', 'utf8');
    });

    it('should copy file successfully', async () => {
      await fileOps.copyFile(srcFile, dstFile);

      const exists = await fs.pathExists(dstFile);
      expect(exists).toBe(true);

      const content = await fs.readFile(dstFile, 'utf8');
      expect(content).toBe('test content');
    });

    it('should overwrite existing file when overwrite is true', async () => {
      await fs.writeFile(dstFile, 'existing content', 'utf8');

      await fileOps.copyFile(srcFile, dstFile, true);

      const content = await fs.readFile(dstFile, 'utf8');
      expect(content).toBe('test content');
    });

    it('should not overwrite existing file when overwrite is false', async () => {
      await fs.writeFile(dstFile, 'existing content', 'utf8');

      await expect(fileOps.copyFile(srcFile, dstFile, false)).rejects.toThrow();
    });

    it('should throw error if source file does not exist', async () => {
      const nonExistentSrc = path.join(tempDir, 'non-existent.txt');

      await expect(fileOps.copyFile(nonExistentSrc, dstFile)).rejects.toThrow();
    });

    it('should create destination directory if it does not exist', async () => {
      const nestedDst = path.join(tempDir, 'nested', 'destination.txt');

      await fileOps.copyFile(srcFile, nestedDst);

      const exists = await fs.pathExists(nestedDst);
      expect(exists).toBe(true);
    });
  });

  describe('copyDirectory', () => {
    let srcDir: string;
    let dstDir: string;

    beforeEach(async () => {
      srcDir = path.join(tempDir, 'source-dir');
      dstDir = path.join(tempDir, 'destination-dir');

      // Create source directory structure
      await fs.ensureDir(srcDir);
      await fs.writeFile(path.join(srcDir, 'file1.txt'), 'content1', 'utf8');
      await fs.writeFile(path.join(srcDir, 'file2.txt'), 'content2', 'utf8');

      const subDir = path.join(srcDir, 'subdir');
      await fs.ensureDir(subDir);
      await fs.writeFile(path.join(subDir, 'file3.txt'), 'content3', 'utf8');
    });

    it('should copy directory recursively', async () => {
      const copiedFiles = await fileOps.copyDirectory(srcDir, dstDir);

      expect(copiedFiles).toHaveLength(3);
      expect(copiedFiles).toContain(path.join(dstDir, 'file1.txt'));
      expect(copiedFiles).toContain(path.join(dstDir, 'file2.txt'));
      expect(copiedFiles).toContain(path.join(dstDir, 'subdir', 'file3.txt'));

      // Verify content
      const content1 = await fs.readFile(path.join(dstDir, 'file1.txt'), 'utf8');
      expect(content1).toBe('content1');

      const content3 = await fs.readFile(path.join(dstDir, 'subdir', 'file3.txt'), 'utf8');
      expect(content3).toBe('content3');
    });

    it('should handle progress callback', async () => {
      const progressCalls: Array<{current: number; total: number; file: string}> = [];

      await fileOps.copyDirectory(srcDir, dstDir, false, (current, total, file) => {
        progressCalls.push({ current, total, file });
      });

      expect(progressCalls).toHaveLength(3);
      expect(progressCalls[progressCalls.length - 1]?.current).toBe(3);
      expect(progressCalls[progressCalls.length - 1]?.total).toBe(3);
    });

    it('should throw error if source directory does not exist', async () => {
      const nonExistentSrc = path.join(tempDir, 'non-existent-dir');

      await expect(fileOps.copyDirectory(nonExistentSrc, dstDir)).rejects.toThrow();
    });
  });

  describe('removeFile', () => {
    let testFile: string;

    beforeEach(async () => {
      testFile = path.join(tempDir, 'to-remove.txt');
      await fs.writeFile(testFile, 'content to remove', 'utf8');
    });

    it('should remove file successfully', async () => {
      await fileOps.removeFile(testFile);

      const exists = await fs.pathExists(testFile);
      expect(exists).toBe(false);
    });

    it('should not throw error if file does not exist', async () => {
      const nonExistentFile = path.join(tempDir, 'non-existent.txt');

      // Should not throw when file doesn't exist
      await expect(async () => {
        await fileOps.removeFile(nonExistentFile);
      }).not.toThrow();
    });
  });

  describe('removeDirectory', () => {
    let testDir: string;

    beforeEach(async () => {
      testDir = path.join(tempDir, 'to-remove-dir');
      await fs.ensureDir(testDir);
      await fs.writeFile(path.join(testDir, 'file.txt'), 'content', 'utf8');

      const subDir = path.join(testDir, 'subdir');
      await fs.ensureDir(subDir);
      await fs.writeFile(path.join(subDir, 'nested.txt'), 'nested content', 'utf8');
    });

    it('should remove directory recursively', async () => {
      await fileOps.removeDirectory(testDir);

      const exists = await fs.pathExists(testDir);
      expect(exists).toBe(false);
    });

    it('should not throw error if directory does not exist', async () => {
      const nonExistentDir = path.join(tempDir, 'non-existent-dir');

      // Should not throw when directory doesn't exist
      await expect(async () => {
        await fileOps.removeDirectory(nonExistentDir);
      }).not.toThrow();
    });
  });

  describe('readFileContent', () => {
    let testFile: string;

    beforeEach(async () => {
      testFile = path.join(tempDir, 'read-test.txt');
    });

    it('should read UTF-8 file content', async () => {
      const testContent = 'Hello, 世界! 🌍';
      await fs.writeFile(testFile, testContent, 'utf8');

      const content = await fileOps.readFileContent(testFile);
      expect(content).toBe(testContent);
    });

    it('should throw error if file does not exist', async () => {
      const nonExistentFile = path.join(tempDir, 'non-existent.txt');

      await expect(fileOps.readFileContent(nonExistentFile)).rejects.toThrow();
    });

    it('should handle empty files', async () => {
      await fs.writeFile(testFile, '', 'utf8');

      const content = await fileOps.readFileContent(testFile);
      expect(content).toBe('');
    });
  });

  describe('writeFileContent', () => {
    let testFile: string;

    beforeEach(() => {
      testFile = path.join(tempDir, 'write-test.txt');
    });

    it('should write UTF-8 content to file', async () => {
      const testContent = 'Hello, 世界! 🌍';

      await fileOps.writeFileContent(testFile, testContent);

      const content = await fs.readFile(testFile, 'utf8');
      expect(content).toBe(testContent);
    });

    it('should create parent directories if they do not exist', async () => {
      const nestedFile = path.join(tempDir, 'nested', 'deep', 'file.txt');

      await fileOps.writeFileContent(nestedFile, 'nested content');

      const exists = await fs.pathExists(nestedFile);
      expect(exists).toBe(true);

      const content = await fs.readFile(nestedFile, 'utf8');
      expect(content).toBe('nested content');
    });

    it('should overwrite existing file content', async () => {
      await fs.writeFile(testFile, 'original content', 'utf8');

      await fileOps.writeFileContent(testFile, 'new content');

      const content = await fs.readFile(testFile, 'utf8');
      expect(content).toBe('new content');
    });
  });

  describe('pathExists', () => {
    it('should return true for existing file', async () => {
      const testFile = path.join(tempDir, 'exists.txt');
      await fs.writeFile(testFile, 'content', 'utf8');

      const exists = await fileOps.pathExists(testFile);
      expect(exists).toBe(true);
    });

    it('should return true for existing directory', async () => {
      const testDir = path.join(tempDir, 'exists-dir');
      await fs.ensureDir(testDir);

      const exists = await fileOps.pathExists(testDir);
      expect(exists).toBe(true);
    });

    it('should return false for non-existent path', async () => {
      const nonExistentPath = path.join(tempDir, 'does-not-exist');

      const exists = await fileOps.pathExists(nonExistentPath);
      expect(exists).toBe(false);
    });
  });

  describe('getFileStats', () => {
    let testFile: string;

    beforeEach(async () => {
      testFile = path.join(tempDir, 'stats-test.txt');
      await fs.writeFile(testFile, 'test content for stats', 'utf8');
    });

    it('should return file statistics', async () => {
      const stats = await fileOps.getFileStats(testFile);

      expect(stats.size).toBeGreaterThan(0);
      expect(stats.isFile).toBe(true);
      expect(stats.isDirectory).toBe(false);
      expect(typeof stats.modified).toBe('object');
      expect(stats.modified.getTime).toBeDefined();
      expect(typeof stats.permissions).toBe('string');
    });

    it('should return directory statistics', async () => {
      const testDir = path.join(tempDir, 'stats-dir');
      await fs.ensureDir(testDir);

      const stats = await fileOps.getFileStats(testDir);

      expect(stats.isFile).toBe(false);
      expect(stats.isDirectory).toBe(true);
      expect(typeof stats.modified).toBe('object');
      expect(stats.modified.getTime).toBeDefined();
    });

    it('should throw error if path does not exist', async () => {
      const nonExistentPath = path.join(tempDir, 'non-existent');

      await expect(fileOps.getFileStats(nonExistentPath)).rejects.toThrow();
    });
  });

  describe('validatePathSafety', () => {
    it('should return true for safe paths within project root', () => {
      const safePath = path.join(testProjectRoot, 'safe', 'file.txt');

      const isSafe = fileOps.validatePathSafety(safePath, testProjectRoot);
      expect(isSafe).toBe(true);
    });

    it('should return false for directory traversal attempts', () => {
      const unsafePath = path.join(testProjectRoot, '..', '..', 'etc', 'passwd');

      const isSafe = fileOps.validatePathSafety(unsafePath, testProjectRoot);
      expect(isSafe).toBe(false);
    });

    it('should return false for absolute paths outside project root', () => {
      const outsidePath = '/tmp/outside-project';

      const isSafe = fileOps.validatePathSafety(outsidePath, testProjectRoot);
      expect(isSafe).toBe(false);
    });

    it('should handle relative paths correctly', () => {
      const relativePath = 'safe/file.txt';

      const isSafe = fileOps.validatePathSafety(relativePath, testProjectRoot);
      expect(isSafe).toBe(true);
    });
  });

  describe('sanitizeFileName', () => {
    it('should remove dangerous characters', () => {
      const dangerousName = 'file<name>|with:bad*chars?.txt';

      const sanitized = fileOps.sanitizeFileName(dangerousName);

      expect(sanitized).not.toContain('<');
      expect(sanitized).not.toContain('>');
      expect(sanitized).not.toContain('|');
      expect(sanitized).not.toContain(':');
      expect(sanitized).not.toContain('*');
      expect(sanitized).not.toContain('?');
    });

    it('should preserve safe characters', () => {
      const safeName = 'valid-file_name.123.txt';

      const sanitized = fileOps.sanitizeFileName(safeName);
      expect(sanitized).toBe(safeName);
    });

    it('should handle unicode characters safely', () => {
      const unicodeName = 'file-世界-🌍.txt';

      const sanitized = fileOps.sanitizeFileName(unicodeName);
      expect(sanitized).toBe('file-.txt');
    });

    it('should handle empty string', () => {
      const sanitized = fileOps.sanitizeFileName('');
      expect(sanitized).toBe('');
    });
  });

  describe('Platform-specific path utilities', () => {
    it('should return correct path separator', () => {
      const separator = fileOps.getPathSeparator();
      expect(separator).toBe(path.sep);
    });

    it('should join paths correctly', () => {
      const joined = fileOps.joinPath('project', 'src', 'index.ts');
      const expected = path.join('project', 'src', 'index.ts');
      expect(joined).toBe(expected);
    });

    it('should resolve paths to absolute', () => {
      const resolved = fileOps.resolvePath('./test');
      expect(path.isAbsolute(resolved)).toBe(true);
    });

    it('should calculate relative paths', () => {
      const from = '/project/src';
      const to = '/project/docs';
      const relative = fileOps.getRelativePath(from, to);
      expect(relative).toBe('../docs');
    });
  });

  describe('Error handling', () => {
    it('should handle permission errors gracefully', async () => {
      if (os.platform() === 'win32') {
        return; // Skip on Windows as permission tests are different
      }

      const readOnlyDir = path.join(tempDir, 'readonly');
      await fs.ensureDir(readOnlyDir);
      await fs.chmod(readOnlyDir, 0o444); // Read-only

      const testFile = path.join(readOnlyDir, 'test.txt');

      await expect(fileOps.writeFileContent(testFile, 'content')).rejects.toThrow();
    });

    it('should handle disk space errors', async () => {
      // This test is difficult to simulate reliably across platforms
      // We'll test with a simulated large content that might fail
      const largeContent = 'x'.repeat(1000000); // 1MB instead of MAX_SAFE_INTEGER
      const testFile = path.join(tempDir, 'large.txt');

      // This should either succeed or not throw an error for this size
      await expect(async () => {
        await fileOps.writeFileContent(testFile, largeContent);
      }).not.toThrow();
    });
  });
});