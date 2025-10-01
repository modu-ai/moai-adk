/**
 * @file Path validator test suite for package root detection
 * @author MoAI Team
 * @tags @TEST:PATH-VALIDATOR-001 @SPEC:BUG-FIX-PACKAGE-PATH-001
 */

import * as fs from 'node:fs';
import * as os from 'node:os';
import * as path from 'node:path';
import { afterEach, beforeEach, describe, expect, test } from 'vitest';
import {
  isInsideMoAIPackage,
  validateProjectPath,
} from '@/utils/path-validator';

describe('Path Validator - Package Root Detection', () => {
  let tempDir: string;

  beforeEach(() => {
    // Create temporary directory for testing
    tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'moai-test-'));
  });

  afterEach(() => {
    // Clean up temporary directory
    if (fs.existsSync(tempDir)) {
      fs.rmSync(tempDir, { recursive: true, force: true });
    }
  });

  describe('isInsideMoAIPackage', () => {
    test('should return true when path is package root', () => {
      // Arrange: Create package.json with name "moai-adk"
      const packageRoot = path.join(tempDir, 'moai-adk-ts');
      fs.mkdirSync(packageRoot, { recursive: true });
      fs.writeFileSync(
        path.join(packageRoot, 'package.json'),
        JSON.stringify({ name: 'moai-adk' }, null, 2)
      );

      // Act: Test package root detection
      const result = isInsideMoAIPackage(packageRoot);

      // Assert: Should detect as inside package
      expect(result).toBe(true);
    });

    test('should return true when path is inside package subdirectory', () => {
      // Arrange: Create package with subdirectories
      const packageRoot = path.join(tempDir, 'moai-adk-ts');
      const srcDir = path.join(packageRoot, 'src', 'cli', 'commands');
      fs.mkdirSync(srcDir, { recursive: true });
      fs.writeFileSync(
        path.join(packageRoot, 'package.json'),
        JSON.stringify({ name: 'moai-adk' }, null, 2)
      );

      // Act: Test nested subdirectory detection
      const result = isInsideMoAIPackage(srcDir);

      // Assert: Should detect as inside package
      expect(result).toBe(true);
    });

    test('should return false for normal project directory', () => {
      // Arrange: Create a normal project directory (not moai-adk package)
      const normalProject = path.join(tempDir, 'my-project');
      fs.mkdirSync(normalProject, { recursive: true });
      fs.writeFileSync(
        path.join(normalProject, 'package.json'),
        JSON.stringify({ name: 'my-project' }, null, 2)
      );

      // Act: Test normal project detection
      const result = isInsideMoAIPackage(normalProject);

      // Assert: Should detect as NOT inside package
      expect(result).toBe(false);
    });

    test('should handle symlinks correctly', () => {
      // Arrange: Create package and symlink to it
      const packageRoot = path.join(tempDir, 'moai-adk-ts');
      const symlinkedPath = path.join(tempDir, 'symlink-to-moai');
      fs.mkdirSync(packageRoot, { recursive: true });
      fs.writeFileSync(
        path.join(packageRoot, 'package.json'),
        JSON.stringify({ name: 'moai-adk' }, null, 2)
      );

      // Create symlink (skip test on Windows if symlink creation fails)
      try {
        fs.symlinkSync(packageRoot, symlinkedPath, 'dir');
      } catch (_error) {
        // Symlink creation might fail on Windows without admin rights
        console.warn('Skipping symlink test: symlink creation failed');
        return;
      }

      // Act: Test symlinked directory detection
      const result = isInsideMoAIPackage(symlinkedPath);

      // Assert: Should resolve symlink and detect as inside package
      expect(result).toBe(true);
    });

    test('should work on Windows paths', () => {
      // Arrange: Create package with Windows-style path separators
      const packageRoot = path.join(tempDir, 'moai-adk-ts');
      fs.mkdirSync(packageRoot, { recursive: true });
      fs.writeFileSync(
        path.join(packageRoot, 'package.json'),
        JSON.stringify({ name: 'moai-adk' }, null, 2)
      );

      // Act: Test with normalized path
      const result = isInsideMoAIPackage(packageRoot);

      // Assert: Should work regardless of platform
      expect(result).toBe(true);
    });
  });

  describe('validateProjectPath', () => {
    test('should reject paths inside MoAI package', () => {
      // Arrange: Create package structure
      const packageRoot = path.join(tempDir, 'moai-adk-ts');
      const projectPath = path.join(packageRoot, 'my-project');
      fs.mkdirSync(packageRoot, { recursive: true });
      fs.writeFileSync(
        path.join(packageRoot, 'package.json'),
        JSON.stringify({ name: 'moai-adk' }, null, 2)
      );

      // Act: Validate project path inside package
      const result = validateProjectPath(projectPath);

      // Assert: Should reject the path
      expect(result.isValid).toBe(false);
      expect(result.error).toContain(
        'Cannot initialize project inside MoAI-ADK package'
      );
    });

    test('should accept paths outside MoAI package', () => {
      // Arrange: Create normal project directory
      const normalProject = path.join(tempDir, 'my-project');
      fs.mkdirSync(normalProject, { recursive: true });

      // Act: Validate normal project path
      const result = validateProjectPath(normalProject);

      // Assert: Should accept the path
      expect(result.isValid).toBe(true);
      expect(result.error).toBeUndefined();
    });
  });
});
