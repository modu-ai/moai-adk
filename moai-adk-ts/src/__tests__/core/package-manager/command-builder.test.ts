/**
 * @file Command builder test suite
 * @author MoAI Team
 * @tags @TEST:REFACTOR-007 @SPEC:REFACTOR-007
 */

import { beforeEach, describe, expect, test } from 'vitest';
import '@/__tests__/setup';
import { CommandBuilder } from '@/core/package-manager/command-builder';
import {
  type PackageInstallOptions,
  PackageManagerType,
} from '@/types/package-manager';

describe('CommandBuilder', () => {
  let builder: CommandBuilder;

  beforeEach(() => {
    builder = new CommandBuilder();
  });

  describe('Install Command Building', () => {
    // @TEST:REFACTOR-007-001: npm install 명령어 생성
    test('should build npm install command', () => {
      // Arrange
      const packages = ['express', 'lodash'];
      const options: PackageInstallOptions = {
        packageManager: PackageManagerType.NPM,
        isDevelopment: false,
      };

      // Act
      const command = builder.buildInstallCommand(packages, options);

      // Assert
      expect(command).toBe('npm install express lodash');
    });

    // @TEST:REFACTOR-007-002: yarn add 명령어 생성
    test('should build yarn add command', () => {
      // Arrange
      const packages = ['express', 'lodash'];
      const options: PackageInstallOptions = {
        packageManager: PackageManagerType.YARN,
        isDevelopment: false,
      };

      // Act
      const command = builder.buildInstallCommand(packages, options);

      // Assert
      expect(command).toBe('yarn add express lodash');
    });

    // @TEST:REFACTOR-007-003: pnpm add 명령어 생성
    test('should build pnpm add command', () => {
      // Arrange
      const packages = ['express', 'lodash'];
      const options: PackageInstallOptions = {
        packageManager: PackageManagerType.PNPM,
        isDevelopment: false,
      };

      // Act
      const command = builder.buildInstallCommand(packages, options);

      // Assert
      expect(command).toBe('pnpm add express lodash');
    });

    // @TEST:REFACTOR-007-004: npm 개발 의존성 플래그
    test('should include --save-dev flag for npm dev dependencies', () => {
      // Arrange
      const packages = ['typescript'];
      const options: PackageInstallOptions = {
        packageManager: PackageManagerType.NPM,
        isDevelopment: true,
      };

      // Act
      const command = builder.buildInstallCommand(packages, options);

      // Assert
      expect(command).toBe('npm install --save-dev typescript');
    });

    // @TEST:REFACTOR-007-005: yarn 개발 의존성 플래그
    test('should include --dev flag for yarn dev dependencies', () => {
      // Arrange
      const packages = ['@types/node', 'typescript'];
      const options: PackageInstallOptions = {
        packageManager: PackageManagerType.YARN,
        isDevelopment: true,
      };

      // Act
      const command = builder.buildInstallCommand(packages, options);

      // Assert
      expect(command).toBe('yarn add --dev @types/node typescript');
    });

    // @TEST:REFACTOR-007-006: pnpm 개발 의존성 플래그
    test('should include --save-dev flag for pnpm dev dependencies', () => {
      // Arrange
      const packages = ['jest'];
      const options: PackageInstallOptions = {
        packageManager: PackageManagerType.PNPM,
        isDevelopment: true,
      };

      // Act
      const command = builder.buildInstallCommand(packages, options);

      // Assert
      expect(command).toBe('pnpm add --save-dev jest');
    });

    // @TEST:REFACTOR-007-007: npm 글로벌 설치
    test('should include --global flag for npm global installation', () => {
      // Arrange
      const packages = ['typescript'];
      const options: PackageInstallOptions = {
        packageManager: PackageManagerType.NPM,
        isGlobal: true,
      };

      // Act
      const command = builder.buildInstallCommand(packages, options);

      // Assert
      expect(command).toBe('npm install --global typescript');
    });

    // @TEST:REFACTOR-007-008: yarn 글로벌 설치
    test('should use global add for yarn global installation', () => {
      // Arrange
      const packages = ['@vue/cli'];
      const options: PackageInstallOptions = {
        packageManager: PackageManagerType.YARN,
        isGlobal: true,
      };

      // Act
      const command = builder.buildInstallCommand(packages, options);

      // Assert
      expect(command).toBe('yarn global add @vue/cli');
    });

    // @TEST:REFACTOR-007-009: pnpm 글로벌 설치
    test('should include --global flag for pnpm global installation', () => {
      // Arrange
      const packages = ['typescript', '@vue/cli'];
      const options: PackageInstallOptions = {
        packageManager: PackageManagerType.PNPM,
        isGlobal: true,
      };

      // Act
      const command = builder.buildInstallCommand(packages, options);

      // Assert
      expect(command).toBe('pnpm add --global typescript @vue/cli');
    });
  });

  describe('Run Command Building', () => {
    // @TEST:REFACTOR-007-010: npm run 명령어
    test('should build npm run command', () => {
      // Act
      const command = builder.buildRunCommand(PackageManagerType.NPM);

      // Assert
      expect(command).toBe('npm run');
    });

    // @TEST:REFACTOR-007-011: yarn run 명령어
    test('should build yarn run command', () => {
      // Act
      const command = builder.buildRunCommand(PackageManagerType.YARN);

      // Assert
      expect(command).toBe('yarn run');
    });

    // @TEST:REFACTOR-007-012: pnpm run 명령어
    test('should build pnpm run command', () => {
      // Act
      const command = builder.buildRunCommand(PackageManagerType.PNPM);

      // Assert
      expect(command).toBe('pnpm run');
    });
  });

  describe('Test Command Building', () => {
    // @TEST:REFACTOR-007-013: npm 기본 테스트 명령어
    test('should build default npm test command', () => {
      // Act
      const command = builder.buildTestCommand(PackageManagerType.NPM);

      // Assert
      expect(command).toBe('npm test');
    });

    // @TEST:REFACTOR-007-014: jest 테스트 명령어
    test('should build jest test command', () => {
      // Act
      const command = builder.buildTestCommand(PackageManagerType.NPM, 'jest');

      // Assert
      expect(command).toBe('jest');
    });

    // @TEST:REFACTOR-007-015: yarn 기본 테스트 명령어
    test('should build default yarn test command', () => {
      // Act
      const command = builder.buildTestCommand(PackageManagerType.YARN);

      // Assert
      expect(command).toBe('yarn test');
    });

    // @TEST:REFACTOR-007-016: pnpm 기본 테스트 명령어
    test('should build default pnpm test command', () => {
      // Act
      const command = builder.buildTestCommand(PackageManagerType.PNPM);

      // Assert
      expect(command).toBe('pnpm test');
    });
  });

  describe('Init Command Building', () => {
    // @TEST:REFACTOR-007-017: npm init 명령어
    test('should build npm init command', () => {
      // Act
      const command = builder.buildInitCommand(PackageManagerType.NPM);

      // Assert
      expect(command).toBe('npm init -y');
    });

    // @TEST:REFACTOR-007-018: yarn init 명령어
    test('should build yarn init command', () => {
      // Act
      const command = builder.buildInitCommand(PackageManagerType.YARN);

      // Assert
      expect(command).toBe('yarn init -y');
    });

    // @TEST:REFACTOR-007-019: pnpm init 명령어
    test('should build pnpm init command', () => {
      // Act
      const command = builder.buildInitCommand(PackageManagerType.PNPM);

      // Assert
      expect(command).toBe('pnpm init');
    });
  });

  describe('Engine Requirements', () => {
    // @TEST:REFACTOR-007-020: npm 엔진 요구사항
    test('should return npm engine requirement', () => {
      // Act
      const engines = builder.getPackageManagerEngine(PackageManagerType.NPM);

      // Assert
      expect(engines).toEqual({ npm: '>=9.0.0' });
    });

    // @TEST:REFACTOR-007-021: yarn 엔진 요구사항
    test('should return yarn engine requirement', () => {
      // Act
      const engines = builder.getPackageManagerEngine(PackageManagerType.YARN);

      // Assert
      expect(engines).toEqual({ yarn: '>=1.22.0' });
    });

    // @TEST:REFACTOR-007-022: pnpm 엔진 요구사항
    test('should return pnpm engine requirement', () => {
      // Act
      const engines = builder.getPackageManagerEngine(PackageManagerType.PNPM);

      // Assert
      expect(engines).toEqual({ pnpm: '>=8.0.0' });
    });
  });

  describe('Error Handling', () => {
    // @TEST:REFACTOR-007-023: 지원하지 않는 패키지 매니저 오류
    test('should throw error for unsupported package manager', () => {
      // Arrange
      const packages = ['express'];
      const options: PackageInstallOptions = {
        packageManager: 'unsupported' as any,
      };

      // Act & Assert
      expect(() => builder.buildInstallCommand(packages, options)).toThrow(
        'Unsupported package manager'
      );
    });
  });
});
