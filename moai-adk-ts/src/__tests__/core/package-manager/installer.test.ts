/**
 * @file Package manager installer test suite
 * @author MoAI Team
 * @tags @TEST:PACKAGE-MANAGER-INSTALLER-001 @REQ:PACKAGE-MANAGER-003
 */

import { beforeEach, describe, expect, test, vi, type MockedFunction } from 'vitest';
import '@/__tests__/setup';
import { execa } from 'execa';
import { PackageManagerInstaller } from '@/core/package-manager/installer';
import {
  type PackageInstallOptions,
  PackageManagerType,
} from '@/types/package-manager';

// Mock execa
vi.mock('execa');
const mockExeca = execa as MockedFunction<typeof execa>;

describe('PackageManagerInstaller', () => {
  let installer: PackageManagerInstaller;

  beforeEach(() => {
    vi.clearAllMocks();
    installer = new PackageManagerInstaller();
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

  describe('Package.json Generation', () => {
    test('should generate package.json for Node.js project', async () => {
      // Arrange
      const projectConfig = {
        name: 'my-node-app',
        version: '1.0.0',
        description: 'A Node.js application',
        author: 'John Doe',
        license: 'MIT',
        type: 'module' as const,
      };

      const packageManagerType = PackageManagerType.NPM;

      // Act
      const packageJson = installer.generatePackageJson(
        projectConfig,
        packageManagerType
      );

      // Assert
      expect(packageJson.name).toBe('my-node-app');
      expect(packageJson.version).toBe('1.0.0');
      expect(packageJson.type).toBe('module');
      expect(packageJson.scripts?.['build']).toBeDefined();
      expect(packageJson.scripts?.['test']).toBeDefined();
      expect(packageJson.scripts?.['start']).toBeDefined();
      expect(packageJson.engines?.['node']).toBeDefined();
    });

    test('should generate package.json with TypeScript configuration', async () => {
      // Arrange
      const projectConfig = {
        name: 'my-ts-app',
        version: '0.1.0',
        description: 'A TypeScript application',
        main: 'dist/index.js',
      };

      const packageManagerType = PackageManagerType.PNPM;
      const includeTypeScript = true;

      // Act
      const packageJson = installer.generatePackageJson(
        projectConfig,
        packageManagerType,
        includeTypeScript
      );

      // Assert
      expect(packageJson.scripts?.['build']).toContain('tsc');
      expect(packageJson.scripts?.['type-check']).toBeDefined();
      expect(packageJson.devDependencies?.['typescript']).toBeDefined();
      expect(packageJson.devDependencies?.['@types/node']).toBeDefined();
    });

    test('should generate package.json with testing framework', async () => {
      // Arrange
      const projectConfig = {
        name: 'my-test-app',
        version: '1.0.0',
      };

      const packageManagerType = PackageManagerType.YARN;
      const includeTypeScript = true;
      const testingFramework = 'jest';

      // Act
      const packageJson = installer.generatePackageJson(
        projectConfig,
        packageManagerType,
        includeTypeScript,
        testingFramework
      );

      // Assert
      expect(packageJson.scripts?.['test']).toContain('jest');
      expect(packageJson.scripts?.['test:watch']).toBeDefined();
      expect(packageJson.scripts?.['test:coverage']).toBeDefined();
      expect(packageJson.devDependencies?.['jest']).toBeDefined();
      expect(packageJson.devDependencies?.['@types/jest']).toBeDefined();
    });
  });

  describe('Dependency Management', () => {
    test('should add dependencies to existing package.json', async () => {
      // Arrange
      const existingPackageJson = {
        name: 'existing-project',
        version: '1.0.0',
        dependencies: {
          express: '^4.18.0',
        },
      };

      const newDependencies = {
        lodash: '^4.17.21',
        axios: '^1.5.0',
      };

      // Act
      const updatedPackageJson = installer.addDependencies(
        existingPackageJson,
        newDependencies
      );

      // Assert
      expect(updatedPackageJson.dependencies).toEqual({
        express: '^4.18.0',
        lodash: '^4.17.21',
        axios: '^1.5.0',
      });
    });

    test('should add dev dependencies without affecting regular dependencies', async () => {
      // Arrange
      const existingPackageJson = {
        name: 'existing-project',
        version: '1.0.0',
        dependencies: {
          express: '^4.18.0',
        },
        devDependencies: {
          typescript: '^5.0.0',
        },
      };

      const newDevDependencies = {
        '@types/express': '^4.17.17',
        jest: '^29.0.0',
      };

      // Act
      const updatedPackageJson = installer.addDevDependencies(
        existingPackageJson,
        newDevDependencies
      );

      // Assert
      expect(updatedPackageJson.dependencies?.['express']).toBe('^4.18.0');
      expect(updatedPackageJson.devDependencies).toEqual({
        typescript: '^5.0.0',
        '@types/express': '^4.17.17',
        jest: '^29.0.0',
      });
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
