/**
 * @file Test for system requirements definition
 * @author MoAI Team
 * @tags @TEST:SYSTEM-REQUIREMENTS-001 @REQ:AUTO-VERIFY-012
 */

import { describe, test, expect } from '@jest/globals';
import { SystemRequirement, requirementRegistry } from '@/core/system-checker/requirements';

describe('SystemRequirement', () => {
  describe('구조 검증', () => {
    test('should have all required properties', () => {
      // Given: SystemRequirement 타입 정의
      const requirement: SystemRequirement = {
        name: 'Git',
        category: 'runtime',
        minVersion: '2.30.0',
        installCommands: {
          darwin: 'brew install git',
          linux: 'sudo apt-get install git',
          win32: 'winget install Git.Git'
        },
        checkCommand: 'git --version',
        versionCommand: 'git --version'
      };

      // When: 속성 검증
      // Then: 모든 필수 속성이 존재해야 함
      expect(requirement).toHaveProperty('name');
      expect(requirement).toHaveProperty('category');
      expect(requirement).toHaveProperty('installCommands');
      expect(requirement).toHaveProperty('checkCommand');
      expect(typeof requirement.name).toBe('string');
      expect(['runtime', 'development', 'optional']).toContain(requirement.category);
    });

    test('should validate category values', () => {
      // Given: 유효한 카테고리들
      const validCategories = ['runtime', 'development', 'optional'] as const;

      // When & Then: 각 카테고리가 유효해야 함
      validCategories.forEach(category => {
        const requirement: SystemRequirement = {
          name: 'TestTool',
          category,
          installCommands: { darwin: 'test' },
          checkCommand: 'test --version'
        };
        expect(requirement.category).toBe(category);
      });
    });
  });

  describe('플랫폼별 설치 명령어', () => {
    test('should support all major platforms', () => {
      // Given: 멀티플랫폼 지원 요구사항
      const requirement: SystemRequirement = {
        name: 'Node.js',
        category: 'runtime',
        minVersion: '18.0.0',
        installCommands: {
          darwin: 'brew install node',
          linux: 'curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -',
          win32: 'winget install OpenJS.NodeJS'
        },
        checkCommand: 'node --version',
        versionCommand: 'node --version'
      };

      // When & Then: 각 플랫폼별 명령어가 정의되어야 함
      expect(requirement.installCommands).toHaveProperty('darwin');
      expect(requirement.installCommands).toHaveProperty('linux');
      expect(requirement.installCommands).toHaveProperty('win32');
      expect(typeof requirement.installCommands['darwin']).toBe('string');
      expect(typeof requirement.installCommands['linux']).toBe('string');
      expect(typeof requirement.installCommands['win32']).toBe('string');
    });
  });
});

describe('RequirementRegistry', () => {
  describe('사전 정의된 요구사항들', () => {
    test('should include Git requirement', () => {
      // Given: Git이 필수 요구사항
      // When: 요구사항 레지스트리에서 Git 조회
      const gitRequirement = requirementRegistry.getRequirement('Git');

      // Then: Git 요구사항이 정의되어야 함
      expect(gitRequirement).toBeDefined();
      expect(gitRequirement?.name).toBe('Git');
      expect(gitRequirement?.category).toBe('runtime');
      expect(gitRequirement?.checkCommand).toContain('git');
    });

    test('should include Node.js requirement', () => {
      // Given: Node.js가 필수 요구사항
      // When: 요구사항 레지스트리에서 Node.js 조회
      const nodeRequirement = requirementRegistry.getRequirement('Node.js');

      // Then: Node.js 요구사항이 정의되어야 함
      expect(nodeRequirement).toBeDefined();
      expect(nodeRequirement?.name).toBe('Node.js');
      expect(nodeRequirement?.category).toBe('runtime');
      expect(nodeRequirement?.minVersion).toBe('18.0.0');
    });

    test('should include SQLite3 requirement', () => {
      // Given: SQLite3가 선택적 요구사항
      // When: 요구사항 레지스트리에서 SQLite3 조회
      const sqliteRequirement = requirementRegistry.getRequirement('SQLite3');

      // Then: SQLite3 요구사항이 정의되어야 함
      expect(sqliteRequirement).toBeDefined();
      expect(sqliteRequirement?.name).toBe('SQLite3');
      expect(sqliteRequirement?.category).toBe('development');
    });

    test('should return all runtime requirements', () => {
      // Given: 런타임 요구사항들이 정의됨
      // When: 런타임 카테고리 요구사항들 조회
      const runtimeRequirements = requirementRegistry.getByCategory('runtime');

      // Then: 최소 Git, Node.js가 포함되어야 함
      expect(runtimeRequirements.length).toBeGreaterThan(0);
      const names = runtimeRequirements.map(req => req.name);
      expect(names).toContain('Git');
      expect(names).toContain('Node.js');
    });

    test('should return empty array for unknown category', () => {
      // Given: 존재하지 않는 카테고리
      // When: 알 수 없는 카테고리로 조회
      const unknownRequirements = requirementRegistry.getByCategory('unknown' as any);

      // Then: 빈 배열을 반환해야 함
      expect(unknownRequirements).toEqual([]);
    });
  });

  describe('요구사항 검색', () => {
    test('should return undefined for unknown requirement', () => {
      // Given: 존재하지 않는 요구사항
      // When: 알 수 없는 요구사항 조회
      const unknown = requirementRegistry.getRequirement('UnknownTool');

      // Then: undefined를 반환해야 함
      expect(unknown).toBeUndefined();
    });

    test('should list all requirements', () => {
      // Given: 요구사항 레지스트리가 초기화됨
      // When: 모든 요구사항 조회
      const allRequirements = requirementRegistry.getAllRequirements();

      // Then: 최소 3개 이상의 요구사항이 있어야 함
      expect(allRequirements.length).toBeGreaterThanOrEqual(3);

      // 각 요구사항이 유효한 구조를 가져야 함
      allRequirements.forEach(req => {
        expect(req).toHaveProperty('name');
        expect(req).toHaveProperty('category');
        expect(req).toHaveProperty('installCommands');
        expect(req).toHaveProperty('checkCommand');
      });
    });
  });
});