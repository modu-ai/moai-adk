/**
 * @file Package manager installer test suite (Refactored)
 * @author MoAI Team
 * @tags @TEST:REFACTOR-007 @SPEC:REFACTOR-007
 */

import { beforeEach, describe, expect, test, vi, type MockedFunction } from 'vitest';
import '@/__tests__/setup';
import { execa } from 'execa';
import { PackageManagerInstaller } from '@/core/package-manager/installer';
import { CommandBuilder } from '@/core/package-manager/command-builder';
import {
  type PackageInstallOptions,
  PackageManagerType,
} from '@/types/package-manager';

// Mock execa
vi.mock('execa');
const mockExeca = execa as MockedFunction<typeof execa>;

describe('PackageManagerInstaller', () => {
  let installer: PackageManagerInstaller;
  let commandBuilder: CommandBuilder;

  beforeEach(() => {
    vi.clearAllMocks();
    commandBuilder = new CommandBuilder();
    installer = new PackageManagerInstaller(commandBuilder);
  });

  describe('Package Installation', () => {
    test('should install packages using npm', async () => {
      // Arrange
      const packages = ['express', 'lodash'];
      const options: PackageInstallOptions = {
        packageManager: PackageManagerType.NPM,
        isDevelopment: false,
      };

      mockExeca.mockResolvedValue({
        stdout: 'added 2 packages',
        stderr: '',
        exitCode: 0,
      } as any);

      // Act
      const result = await installer.installPackages(packages, options);

      // Assert
      expect(result.success).toBe(true);
      expect(result.installedPackages).toEqual(packages);
      expect(mockExeca).toHaveBeenCalledWith(
        'npm',
        ['install', 'express', 'lodash'],
        expect.objectContaining({
          cwd: process.cwd(),
        })
      );
    });

    test('should install dev dependencies using yarn', async () => {
      // Arrange
      const packages = ['@types/node', 'typescript'];
      const options: PackageInstallOptions = {
        packageManager: PackageManagerType.YARN,
        isDevelopment: true,
        workingDirectory: '/test/project',
      };

      mockExeca.mockResolvedValue({
        stdout: 'success Saved lockfile.',
        stderr: '',
        exitCode: 0,
      } as any);

      // Act
      const result = await installer.installPackages(packages, options);

      // Assert
      expect(result.success).toBe(true);
      expect(mockExeca).toHaveBeenCalledWith(
        'yarn',
        ['add', '--dev', '@types/node', 'typescript'],
        expect.objectContaining({
          cwd: '/test/project',
        })
      );
    });

    test('should install global packages using pnpm', async () => {
      // Arrange
      const packages = ['typescript', '@vue/cli'];
      const options: PackageInstallOptions = {
        packageManager: PackageManagerType.PNPM,
        isGlobal: true,
      };

      mockExeca.mockResolvedValue({
        stdout: 'Packages: +2',
        stderr: '',
        exitCode: 0,
      } as any);

      // Act
      const result = await installer.installPackages(packages, options);

      // Assert
      expect(result.success).toBe(true);
      expect(mockExeca).toHaveBeenCalledWith(
        'pnpm',
        ['add', '--global', 'typescript', '@vue/cli'],
        expect.any(Object)
      );
    });

    test('should handle installation failures', async () => {
      // Arrange
      const packages = ['nonexistent-package'];
      const options: PackageInstallOptions = {
        packageManager: PackageManagerType.NPM,
      };

      mockExeca.mockRejectedValue(new Error('Package not found'));

      // Act
      const result = await installer.installPackages(packages, options);

      // Assert
      expect(result.success).toBe(false);
      expect(result.error).toContain('Package not found');
      expect(result.installedPackages).toEqual([]);
    });
  });

  describe('Project Initialization', () => {
    test('should initialize project with package manager', async () => {
      // Arrange
      const projectPath = '/test/new-project';
      const packageManagerType = PackageManagerType.YARN;

      mockExeca.mockResolvedValue({
        stdout: 'success Saved package.json',
        stderr: '',
        exitCode: 0,
      } as any);

      // Act
      const result = await installer.initializeProject(
        projectPath,
        packageManagerType
      );

      // Assert
      expect(result.success).toBe(true);
      expect(mockExeca).toHaveBeenCalledWith(
        'yarn',
        ['init', '-y'],
        expect.objectContaining({
          cwd: projectPath,
        })
      );
    });

    test('should handle initialization failures', async () => {
      // Arrange
      const projectPath = '/invalid/path';
      const packageManagerType = PackageManagerType.NPM;

      mockExeca.mockRejectedValue(new Error('Directory not found'));

      // Act
      const result = await installer.initializeProject(
        projectPath,
        packageManagerType
      );

      // Assert
      expect(result.success).toBe(false);
      expect(result.error).toContain('Directory not found');
    });
  });
});
