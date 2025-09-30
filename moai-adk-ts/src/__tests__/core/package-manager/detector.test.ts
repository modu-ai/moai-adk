/**
 * @file Package manager detector test suite
 * @author MoAI Team
 * @tags @TEST:PACKAGE-MANAGER-DETECTOR-001 @SPEC:PACKAGE-MANAGER-002
 */

import { beforeEach, describe, expect, test, vi, type MockedFunction } from 'vitest';
import '@/__tests__/setup';
import { execa } from 'execa';
import { PackageManagerDetector } from '@/core/package-manager/detector';
import {
  type PackageManagerInfo,
  PackageManagerType,
} from '@/types/package-manager';

// Mock execa
vi.mock('execa');
const mockExeca = execa as MockedFunction<typeof execa>;

describe('PackageManagerDetector', () => {
  let detector: PackageManagerDetector;

  beforeEach(() => {
    vi.clearAllMocks();
    detector = new PackageManagerDetector();
  });

  describe('Single Package Manager Detection', () => {
    test('should detect npm when available', async () => {
      // Arrange
      mockExeca.mockResolvedValue({
        stdout: '9.8.1',
        stderr: '',
        exitCode: 0,
      } as any);

      // Act
      const result = await detector.detectPackageManager(
        PackageManagerType.NPM
      );

      // Assert
      expect(result.isAvailable).toBe(true);
      expect(result.type).toBe(PackageManagerType.NPM);
      expect(result.version).toBe('9.8.1');
      expect(mockExeca).toHaveBeenCalledWith(
        'npm',
        ['--version'],
        expect.any(Object)
      );
    });

    test('should detect yarn when available', async () => {
      // Arrange
      mockExeca.mockResolvedValue({
        stdout: '1.22.19',
        stderr: '',
        exitCode: 0,
      } as any);

      // Act
      const result = await detector.detectPackageManager(
        PackageManagerType.YARN
      );

      // Assert
      expect(result.isAvailable).toBe(true);
      expect(result.type).toBe(PackageManagerType.YARN);
      expect(result.version).toBe('1.22.19');
      expect(mockExeca).toHaveBeenCalledWith(
        'yarn',
        ['--version'],
        expect.any(Object)
      );
    });

    test('should detect pnpm when available', async () => {
      // Arrange
      mockExeca.mockResolvedValue({
        stdout: '8.7.0',
        stderr: '',
        exitCode: 0,
      } as any);

      // Act
      const result = await detector.detectPackageManager(
        PackageManagerType.PNPM
      );

      // Assert
      expect(result.isAvailable).toBe(true);
      expect(result.type).toBe(PackageManagerType.PNPM);
      expect(result.version).toBe('8.7.0');
      expect(mockExeca).toHaveBeenCalledWith(
        'pnpm',
        ['--version'],
        expect.any(Object)
      );
    });

    test('should handle unavailable package manager', async () => {
      // Arrange
      mockExeca.mockRejectedValue(new Error('Command not found'));

      // Act
      const result = await detector.detectPackageManager(
        PackageManagerType.YARN
      );

      // Assert
      expect(result.isAvailable).toBe(false);
      expect(result.type).toBe(PackageManagerType.YARN);
      expect(result.version).toBe('unknown');
    });
  });

  describe('All Package Managers Detection', () => {
    test('should detect all available package managers', async () => {
      // Arrange
      mockExeca
        .mockResolvedValueOnce({
          stdout: '9.8.1',
          stderr: '',
          exitCode: 0,
        } as any) // npm
        .mockResolvedValueOnce({
          stdout: '1.22.19',
          stderr: '',
          exitCode: 0,
        } as any) // yarn
        .mockRejectedValueOnce(new Error('Command not found')); // pnpm not available

      // Act
      const result = await detector.detectAllPackageManagers();

      // Assert
      expect(result.available).toHaveLength(2);
      expect(result.available[0]?.type).toBe(PackageManagerType.NPM);
      expect(result.available[1]?.type).toBe(PackageManagerType.YARN);
      expect(result.available[0]?.isAvailable).toBe(true);
      expect(result.available[1]?.isAvailable).toBe(true);
    });

    test('should identify preferred package manager based on lock files', async () => {
      // Arrange
      mockExeca
        .mockResolvedValueOnce({
          stdout: '9.8.1',
          stderr: '',
          exitCode: 0,
        } as any) // npm
        .mockResolvedValueOnce({
          stdout: '1.22.19',
          stderr: '',
          exitCode: 0,
        } as any) // yarn
        .mockResolvedValueOnce({
          stdout: '8.7.0',
          stderr: '',
          exitCode: 0,
        } as any); // pnpm

      // Mock file system to simulate yarn.lock exists
      vi.spyOn(detector as any, 'detectLockFile').mockResolvedValue(
        PackageManagerType.YARN
      );

      // Act
      const result = await detector.detectAllPackageManagers();

      // Assert
      expect(result.preferred?.type).toBe(PackageManagerType.YARN);
      expect(result.preferred?.isPreferred).toBe(true);
    });

    test('should recommend fastest package manager when no preference exists', async () => {
      // Arrange
      mockExeca
        .mockResolvedValueOnce({
          stdout: '9.8.1',
          stderr: '',
          exitCode: 0,
        } as any) // npm
        .mockResolvedValueOnce({
          stdout: '1.22.19',
          stderr: '',
          exitCode: 0,
        } as any) // yarn
        .mockResolvedValueOnce({
          stdout: '8.7.0',
          stderr: '',
          exitCode: 0,
        } as any); // pnpm

      // Mock no lock file found
      vi.spyOn(detector as any, 'detectLockFile').mockResolvedValue(null);

      // Act
      const result = await detector.detectAllPackageManagers();

      // Assert
      expect(result.recommended?.type).toBe(PackageManagerType.PNPM); // Fastest
    });
  });

  describe('Package Manager Commands', () => {
    test('should return correct commands for npm', () => {
      // Act
      const commands = detector.getCommands(PackageManagerType.NPM);

      // Assert
      expect(commands.install).toBe('npm install');
      expect(commands.installDev).toBe('npm install --save-dev');
      expect(commands.installGlobal).toBe('npm install --global');
      expect(commands.run).toBe('npm run');
      expect(commands.build).toBe('npm run build');
      expect(commands.test).toBe('npm test');
      expect(commands.init).toBe('npm init -y');
    });

    test('should return correct commands for yarn', () => {
      // Act
      const commands = detector.getCommands(PackageManagerType.YARN);

      // Assert
      expect(commands.install).toBe('yarn add');
      expect(commands.installDev).toBe('yarn add --dev');
      expect(commands.installGlobal).toBe('yarn global add');
      expect(commands.run).toBe('yarn run');
      expect(commands.build).toBe('yarn build');
      expect(commands.test).toBe('yarn test');
      expect(commands.init).toBe('yarn init -y');
    });

    test('should return correct commands for pnpm', () => {
      // Act
      const commands = detector.getCommands(PackageManagerType.PNPM);

      // Assert
      expect(commands.install).toBe('pnpm add');
      expect(commands.installDev).toBe('pnpm add --save-dev');
      expect(commands.installGlobal).toBe('pnpm add --global');
      expect(commands.run).toBe('pnpm run');
      expect(commands.build).toBe('pnpm build');
      expect(commands.test).toBe('pnpm test');
      expect(commands.init).toBe('pnpm init');
    });
  });

  describe('Lock File Detection', () => {
    test('should detect npm from package-lock.json', async () => {
      // Arrange
      vi.spyOn(detector as any, 'fileExists')
        .mockResolvedValueOnce(true) // package-lock.json exists
        .mockResolvedValueOnce(false) // yarn.lock doesn't exist
        .mockResolvedValueOnce(false); // pnpm-lock.yaml doesn't exist

      // Act
      const result = await (detector as any).detectLockFile();

      // Assert
      expect(result).toBe(PackageManagerType.NPM);
    });

    test('should detect yarn from yarn.lock', async () => {
      // Arrange
      vi.spyOn(detector as any, 'fileExists')
        .mockResolvedValueOnce(false) // package-lock.json doesn't exist
        .mockResolvedValueOnce(true) // yarn.lock exists
        .mockResolvedValueOnce(false); // pnpm-lock.yaml doesn't exist

      // Act
      const result = await (detector as any).detectLockFile();

      // Assert
      expect(result).toBe(PackageManagerType.YARN);
    });

    test('should detect pnpm from pnpm-lock.yaml', async () => {
      // Arrange
      vi.spyOn(detector as any, 'fileExists')
        .mockResolvedValueOnce(false) // package-lock.json doesn't exist
        .mockResolvedValueOnce(false) // yarn.lock doesn't exist
        .mockResolvedValueOnce(true); // pnpm-lock.yaml exists

      // Act
      const result = await (detector as any).detectLockFile();

      // Assert
      expect(result).toBe(PackageManagerType.PNPM);
    });

    test('should return null when no lock files exist', async () => {
      // Arrange
      vi.spyOn(detector as any, 'fileExists').mockResolvedValue(false); // No lock files exist

      // Act
      const result = await (detector as any).detectLockFile();

      // Assert
      expect(result).toBe(null);
    });
  });

  describe('Version Comparison and Recommendations', () => {
    test('should recommend latest version when multiple package managers available', async () => {
      // Arrange
      const npmInfo: PackageManagerInfo = {
        type: PackageManagerType.NPM,
        version: '9.8.1',
        isAvailable: true,
      };

      const yarnInfo: PackageManagerInfo = {
        type: PackageManagerType.YARN,
        version: '1.22.19',
        isAvailable: true,
      };

      const pnpmInfo: PackageManagerInfo = {
        type: PackageManagerType.PNPM,
        version: '8.7.0',
        isAvailable: true,
      };

      // Act
      const recommended = detector.recommendPackageManager([
        npmInfo,
        yarnInfo,
        pnpmInfo,
      ]);

      // Assert
      expect(recommended.type).toBe(PackageManagerType.PNPM); // Fastest and most modern
    });

    test('should prefer npm when only npm is available', async () => {
      // Arrange
      const npmInfo: PackageManagerInfo = {
        type: PackageManagerType.NPM,
        version: '9.8.1',
        isAvailable: true,
      };

      // Act
      const recommended = detector.recommendPackageManager([npmInfo]);

      // Assert
      expect(recommended.type).toBe(PackageManagerType.NPM);
    });
  });
});
