/**
 * @file Package.json builder test suite
 * @author MoAI Team
 * @tags @TEST:REFACTOR-007 @SPEC:REFACTOR-007
 */

import { beforeEach, describe, expect, test } from 'vitest';
import '@/__tests__/setup';
import { CommandBuilder } from '@/core/package-manager/command-builder';
import { PackageJsonBuilder } from '@/core/package-manager/package-json-builder';
import { PackageManagerType } from '@/types/package-manager';

describe('PackageJsonBuilder', () => {
  let builder: PackageJsonBuilder;
  let commandBuilder: CommandBuilder;

  beforeEach(() => {
    commandBuilder = new CommandBuilder();
    builder = new PackageJsonBuilder(commandBuilder);
  });

  describe('Package.json Generation', () => {
    // @TEST:REFACTOR-007-101: 기본 package.json 생성
    test('should generate basic package.json', () => {
      // Arrange
      const projectConfig = {
        name: 'test-project',
        version: '1.0.0',
      };
      const packageManager = PackageManagerType.NPM;

      // Act
      const packageJson = builder.generatePackageJson(
        projectConfig,
        packageManager
      );

      // Assert
      expect(packageJson.name).toBe('test-project');
      expect(packageJson.version).toBe('1.0.0');
      expect(packageJson.scripts).toBeDefined();
      expect(packageJson.engines).toBeDefined();
      expect(packageJson.engines?.node).toBe('>=18.0.0');
      expect(packageJson.engines?.npm).toBe('>=9.0.0');
      expect(packageJson.dependencies).toBeDefined();
      expect(packageJson.devDependencies).toBeDefined();
    });

    // @TEST:REFACTOR-007-102: 기본값 설정 확인
    test('should use default values for missing fields', () => {
      // Arrange
      const projectConfig = {};
      const packageManager = PackageManagerType.NPM;

      // Act
      const packageJson = builder.generatePackageJson(
        projectConfig,
        packageManager
      );

      // Assert
      expect(packageJson.name).toBe('unnamed-project');
      expect(packageJson.version).toBe('1.0.0');
      expect(packageJson.main).toBe('index.js');
      expect(packageJson.type).toBe('commonjs');
      expect(packageJson.license).toBe('MIT');
    });

    // @TEST:REFACTOR-007-103: 사용자 제공 구성 사용
    test('should use provided configuration', () => {
      // Arrange
      const projectConfig = {
        name: 'my-app',
        version: '2.0.0',
        description: 'My application',
        author: 'John Doe',
        license: 'Apache-2.0',
        type: 'module' as const,
      };
      const packageManager = PackageManagerType.YARN;

      // Act
      const packageJson = builder.generatePackageJson(
        projectConfig,
        packageManager
      );

      // Assert
      expect(packageJson.name).toBe('my-app');
      expect(packageJson.version).toBe('2.0.0');
      expect(packageJson.description).toBe('My application');
      expect(packageJson.author).toBe('John Doe');
      expect(packageJson.license).toBe('Apache-2.0');
      expect(packageJson.type).toBe('module');
    });
  });

  describe('TypeScript Integration', () => {
    // @TEST:REFACTOR-007-104: TypeScript 의존성 포함
    test('should include TypeScript dependencies', () => {
      // Arrange
      const projectConfig = {
        name: 'ts-project',
        version: '1.0.0',
      };
      const packageManager = PackageManagerType.NPM;
      const includeTypeScript = true;

      // Act
      const packageJson = builder.generatePackageJson(
        projectConfig,
        packageManager,
        includeTypeScript
      );

      // Assert
      expect(packageJson.devDependencies?.['typescript']).toBe('^5.0.0');
      expect(packageJson.devDependencies?.['@types/node']).toBe('^20.0.0');
    });

    // @TEST:REFACTOR-007-105: TypeScript 스크립트 포함
    test('should include TypeScript scripts', () => {
      // Arrange
      const projectConfig = {
        name: 'ts-project',
        version: '1.0.0',
      };
      const packageManager = PackageManagerType.NPM;
      const includeTypeScript = true;

      // Act
      const packageJson = builder.generatePackageJson(
        projectConfig,
        packageManager,
        includeTypeScript
      );

      // Assert
      expect(packageJson.scripts?.['build']).toBe('tsc');
      expect(packageJson.scripts?.['type-check']).toBe('tsc --noEmit');
      expect(packageJson.scripts?.['dev']).toBe('ts-node src/index.ts');
    });
  });

  describe('Testing Framework Integration', () => {
    // @TEST:REFACTOR-007-106: Jest 의존성 포함
    test('should include Jest dependencies', () => {
      // Arrange
      const projectConfig = {
        name: 'test-app',
        version: '1.0.0',
      };
      const packageManager = PackageManagerType.NPM;
      const includeTypeScript = false;
      const testingFramework = 'jest';

      // Act
      const packageJson = builder.generatePackageJson(
        projectConfig,
        packageManager,
        includeTypeScript,
        testingFramework
      );

      // Assert
      expect(packageJson.devDependencies?.['jest']).toBe('^29.0.0');
    });

    // @TEST:REFACTOR-007-107: Jest 스크립트 포함
    test('should include Jest scripts', () => {
      // Arrange
      const projectConfig = {
        name: 'test-app',
        version: '1.0.0',
      };
      const packageManager = PackageManagerType.YARN;
      const includeTypeScript = false;
      const testingFramework = 'jest';

      // Act
      const packageJson = builder.generatePackageJson(
        projectConfig,
        packageManager,
        includeTypeScript,
        testingFramework
      );

      // Assert
      expect(packageJson.scripts?.['test']).toBe('jest');
      expect(packageJson.scripts?.['test:watch']).toBe('jest --watch');
      expect(packageJson.scripts?.['test:coverage']).toBe('jest --coverage');
    });

    // @TEST:REFACTOR-007-108: TypeScript + Jest 통합
    test('should include TypeScript Jest dependencies when both enabled', () => {
      // Arrange
      const projectConfig = {
        name: 'ts-test-app',
        version: '1.0.0',
      };
      const packageManager = PackageManagerType.NPM;
      const includeTypeScript = true;
      const testingFramework = 'jest';

      // Act
      const packageJson = builder.generatePackageJson(
        projectConfig,
        packageManager,
        includeTypeScript,
        testingFramework
      );

      // Assert
      expect(packageJson.devDependencies?.['jest']).toBe('^29.0.0');
      expect(packageJson.devDependencies?.['@types/jest']).toBe('^29.0.0');
      expect(packageJson.devDependencies?.['ts-jest']).toBe('^29.0.0');
    });
  });

  describe('Scripts Generation', () => {
    // @TEST:REFACTOR-007-109: 기본 스크립트 생성
    test('should generate basic scripts', () => {
      // Arrange
      const packageManager = PackageManagerType.NPM;
      const includeTypeScript = false;

      // Act
      const scripts = builder.generateScripts(
        packageManager,
        includeTypeScript
      );

      // Assert
      expect(scripts['start']).toBe('node index.js');
      expect(scripts['test']).toBe('npm test');
      expect(scripts['build']).toBe('echo "No build step configured"');
    });

    // @TEST:REFACTOR-007-110: TypeScript 스크립트 생성
    test('should generate TypeScript scripts', () => {
      // Arrange
      const packageManager = PackageManagerType.PNPM;
      const includeTypeScript = true;

      // Act
      const scripts = builder.generateScripts(
        packageManager,
        includeTypeScript
      );

      // Assert
      expect(scripts['build']).toBe('tsc');
      expect(scripts['type-check']).toBe('tsc --noEmit');
      expect(scripts['dev']).toBe('ts-node src/index.ts');
    });

    // @TEST:REFACTOR-007-111: Jest 스크립트 생성
    test('should generate Jest scripts', () => {
      // Arrange
      const packageManager = PackageManagerType.YARN;
      const includeTypeScript = false;
      const testingFramework = 'jest';

      // Act
      const scripts = builder.generateScripts(
        packageManager,
        includeTypeScript,
        testingFramework
      );

      // Assert
      expect(scripts['test']).toBe('jest');
      expect(scripts['test:watch']).toBe('jest --watch');
      expect(scripts['test:coverage']).toBe('jest --coverage');
    });
  });

  describe('Dependency Management', () => {
    // @TEST:REFACTOR-007-112: 의존성 추가
    test('should add dependencies to existing package.json', () => {
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
      const updatedPackageJson = builder.addDependencies(
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

    // @TEST:REFACTOR-007-113: 개발 의존성 추가
    test('should add dev dependencies without affecting regular dependencies', () => {
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
      const updatedPackageJson = builder.addDevDependencies(
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

    // @TEST:REFACTOR-007-114: 기존 구성 유지
    test('should preserve existing configuration when adding dependencies', () => {
      // Arrange
      const existingPackageJson = {
        name: 'my-app',
        version: '2.0.0',
        description: 'My app',
        author: 'John Doe',
        dependencies: {
          express: '^4.18.0',
        },
      };
      const newDependencies = {
        lodash: '^4.17.21',
      };

      // Act
      const updatedPackageJson = builder.addDependencies(
        existingPackageJson,
        newDependencies
      );

      // Assert
      expect(updatedPackageJson.name).toBe('my-app');
      expect(updatedPackageJson.version).toBe('2.0.0');
      expect(updatedPackageJson.description).toBe('My app');
      expect(updatedPackageJson.author).toBe('John Doe');
    });
  });

  describe('Package Manager Specific', () => {
    // @TEST:REFACTOR-007-115: Yarn 엔진 요구사항
    test('should include Yarn engine requirements', () => {
      // Arrange
      const projectConfig = {
        name: 'yarn-project',
        version: '1.0.0',
      };
      const packageManager = PackageManagerType.YARN;

      // Act
      const packageJson = builder.generatePackageJson(
        projectConfig,
        packageManager
      );

      // Assert
      expect(packageJson.engines?.['yarn']).toBe('>=1.22.0');
    });

    // @TEST:REFACTOR-007-116: PNPM 엔진 요구사항
    test('should include PNPM engine requirements', () => {
      // Arrange
      const projectConfig = {
        name: 'pnpm-project',
        version: '1.0.0',
      };
      const packageManager = PackageManagerType.PNPM;

      // Act
      const packageJson = builder.generatePackageJson(
        projectConfig,
        packageManager
      );

      // Assert
      expect(packageJson.engines?.['pnpm']).toBe('>=8.0.0');
    });
  });
});
