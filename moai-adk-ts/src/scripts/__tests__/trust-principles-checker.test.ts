/**
 * @TEST-TRUST-CHECKER-001: TrustPrinciplesChecker TypeScript 포팅 테스트
 * @연결: @TASK:CONSTITUTION-CHECK-011 → @FEATURE:TRUST-VALIDATION-001 → @TASK:TRUST-TS-PORT-001
 *
 * 이 테스트는 Python check_constitution.py의 기능을 TypeScript로 완전 포팅한
 * TrustPrinciplesChecker 클래스의 동작을 검증합니다.
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import { beforeEach, describe, expect, vi } from 'vitest';
import { TrustPrinciplesChecker } from '../trust-principles-checker';

// Mock fs module
vi.mock('fs');
const mockFs = fs as vi.Mocked<typeof fs>;

describe('TrustPrinciplesChecker', () => {
  let checker: TrustPrinciplesChecker;
  const mockProjectRoot = '/mock/project';

  beforeEach(() => {
    vi.clearAllMocks();
    checker = new TrustPrinciplesChecker(mockProjectRoot);
  });

  describe('constructor', () => {
    it('@TEST-INIT-001: should initialize with default project root and relaxed mode', () => {
      const defaultChecker = new TrustPrinciplesChecker();

      expect(defaultChecker).toBeDefined();
      expect((defaultChecker as any).strict).toBe(false);
      expect((defaultChecker as any).projectRoot).toBeDefined();
    });

    it('@TEST-INIT-002: should initialize with custom project root and strict mode', () => {
      const strictChecker = new TrustPrinciplesChecker(mockProjectRoot, true);

      expect((strictChecker as any).projectRoot).toBe(
        path.resolve(mockProjectRoot)
      );
      expect((strictChecker as any).strict).toBe(true);
      expect((strictChecker as any).violations).toEqual([]);
    });
  });

  describe('loadConfig', () => {
    it('@TEST-CONFIG-001: should load existing config file', () => {
      const mockConfig = {
        mode: 'personal',
        constitution: {
          principles: {
            simplicity: { max_projects: 3 },
          },
        },
      };

      mockFs.existsSync.mockReturnValue(true);
      mockFs.readFileSync.mockReturnValue(JSON.stringify(mockConfig));

      const result = checker.loadConfig();

      expect(result).toEqual(mockConfig);
      expect(mockFs.existsSync).toHaveBeenCalledWith(
        path.join(mockProjectRoot, '.moai', 'config.json')
      );
    });

    it('@TEST-CONFIG-002: should return empty object when config file does not exist', () => {
      mockFs.existsSync.mockReturnValue(false);

      const result = checker.loadConfig();

      expect(result).toEqual({});
    });

    it('@TEST-CONFIG-003: should handle JSON parsing errors gracefully', () => {
      mockFs.existsSync.mockReturnValue(true);
      mockFs.readFileSync.mockReturnValue('invalid json');

      const result = checker.loadConfig();

      expect(result).toEqual({});
    });
  });

  describe('checkSimplicity', () => {
    const mockConfig = {
      constitution: {
        principles: {
          simplicity: { max_projects: 3 },
        },
      },
    };

    beforeEach(() => {
      // Mock src directory exists
      mockFs.existsSync.mockImplementation(path => {
        return path.toString().includes('src');
      });
    });

    it('@TEST-SIMPLICITY-001: should pass when module count is within limit (relaxed mode)', () => {
      // Mock directory structure
      const mockDirs = [
        { name: 'module1', isDirectory: () => true },
        { name: 'module2', isDirectory: () => true },
        { name: '__pycache__', isDirectory: () => true },
      ];

      mockFs.readdirSync.mockReturnValue(mockDirs as any);

      // Mock Python files in modules
      mockFs.readdirSync.mockImplementation(dirPath => {
        if (
          dirPath.toString().includes('module1') ||
          dirPath.toString().includes('module2')
        ) {
          return [{ name: 'test.py', isFile: () => true }] as any;
        }
        return mockDirs as any;
      });

      const result = checker.checkSimplicity(mockConfig);

      expect(result).toBe(true);
      expect((checker as any).violations).toHaveLength(0);
    });

    it('@TEST-SIMPLICITY-002: should fail when module count exceeds limit (relaxed mode)', () => {
      const mockDirs = [
        { name: 'module1', isDirectory: () => true },
        { name: 'module2', isDirectory: () => true },
        { name: 'module3', isDirectory: () => true },
        { name: 'module4', isDirectory: () => true },
      ];

      mockFs.readdirSync.mockReturnValue(mockDirs as any);

      // Mock Python files in all modules
      mockFs.readdirSync.mockImplementation(dirPath => {
        if (dirPath.toString().includes('module')) {
          return [{ name: 'test.py', isFile: () => true }] as any;
        }
        return mockDirs as any;
      });

      const result = checker.checkSimplicity(mockConfig);

      expect(result).toBe(false);
      expect((checker as any).violations).toHaveLength(1);
      expect((checker as any).violations[0][0]).toBe('Simplicity');
    });

    it('@TEST-SIMPLICITY-003: should pass when src directory does not exist', () => {
      mockFs.existsSync.mockReturnValue(false);

      const result = checker.checkSimplicity(mockConfig);

      expect(result).toBe(true);
      expect((checker as any).violations).toHaveLength(0);
    });

    it('@TEST-SIMPLICITY-004: should check file count in strict mode', () => {
      const strictChecker = new TrustPrinciplesChecker(mockProjectRoot, true);

      // Mock getAllSourceFiles to return many files
      vi.spyOn(strictChecker as any, 'getAllSourceFiles').mockReturnValue([
        'file1.py',
        'file2.py',
        'file3.py',
        'file4.py',
        'file5.py',
      ]);

      const result = strictChecker.checkSimplicity(mockConfig);

      expect(result).toBe(false);
      expect((strictChecker as any).violations).toHaveLength(1);
    });
  });

  describe('checkArchitecture', () => {
    beforeEach(() => {
      mockFs.existsSync.mockImplementation(path => {
        return path.toString().includes('src');
      });
    });

    it('@TEST-ARCHITECTURE-001: should pass with expected directory structure (relaxed mode)', () => {
      const mockDirs = [
        { name: 'models', isDirectory: () => true },
        { name: 'utils', isDirectory: () => true },
        { name: 'other', isDirectory: () => true },
      ];

      mockFs.readdirSync.mockReturnValue(mockDirs as any);

      const result = checker.checkArchitecture();

      expect(result).toBe(true);
      expect((checker as any).violations).toHaveLength(0);
    });

    it('@TEST-ARCHITECTURE-002: should fail with no expected directories (relaxed mode)', () => {
      const mockDirs = [
        { name: 'random1', isDirectory: () => true },
        { name: 'random2', isDirectory: () => true },
      ];

      mockFs.readdirSync.mockReturnValue(mockDirs as any);

      const result = checker.checkArchitecture();

      expect(result).toBe(false);
      expect((checker as any).violations).toHaveLength(1);
      expect((checker as any).violations[0][0]).toBe('Architecture');
    });

    it('@TEST-ARCHITECTURE-003: should require 2+ directories in strict mode', () => {
      const strictChecker = new TrustPrinciplesChecker(mockProjectRoot, true);

      const mockDirs = [
        { name: 'models', isDirectory: () => true },
        { name: 'other', isDirectory: () => true },
      ];

      mockFs.readdirSync.mockReturnValue(mockDirs as any);

      const result = strictChecker.checkArchitecture();

      expect(result).toBe(false); // models만 있으므로 strict 모드에서 실패해야 함
      expect((strictChecker as any).violations).toHaveLength(1);
    });
  });

  describe('checkTesting', () => {
    it('@TEST-TESTING-001: should pass with sufficient test coverage', () => {
      // Mock test coverage calculation
      vi.spyOn(checker as any, 'countTestFiles').mockReturnValue(2);
      vi.spyOn(checker as any, 'countSourceFiles').mockReturnValue(2);
      vi.spyOn(checker as any, 'calculateCoverage').mockReturnValue(85);

      const result = checker.checkTesting();

      expect(result).toBe(true);
      expect((checker as any).violations).toHaveLength(0);
    });

    it('@TEST-TESTING-002: should fail with low test coverage', () => {
      vi.spyOn(checker as any, 'countTestFiles').mockReturnValue(1);
      vi.spyOn(checker as any, 'countSourceFiles').mockReturnValue(10);
      vi.spyOn(checker as any, 'calculateCoverage').mockReturnValue(30);

      const result = checker.checkTesting();

      expect(result).toBe(false);
      expect((checker as any).violations).toHaveLength(1);
      expect((checker as any).violations[0][0]).toBe('Testing');
    });

    it('@TEST-TESTING-003: should fail with no test files', () => {
      vi.spyOn(checker as any, 'countTestFiles').mockReturnValue(0);
      vi.spyOn(checker as any, 'countSourceFiles').mockReturnValue(5);

      const result = checker.checkTesting();

      expect(result).toBe(false);
      expect((checker as any).violations).toHaveLength(1);
    });
  });

  describe('checkObservability', () => {
    it('@TEST-OBSERVABILITY-001: should pass with logging infrastructure', () => {
      vi.spyOn(checker as any, 'hasLoggingInfrastructure').mockReturnValue(
        true
      );
      vi.spyOn(checker as any, 'hasErrorTracking').mockReturnValue(true);

      const result = checker.checkObservability();

      expect(result).toBe(true);
      expect((checker as any).violations).toHaveLength(0);
    });

    it('@TEST-OBSERVABILITY-002: should fail without logging infrastructure', () => {
      vi.spyOn(checker as any, 'hasLoggingInfrastructure').mockReturnValue(
        false
      );
      vi.spyOn(checker as any, 'hasErrorTracking').mockReturnValue(false);

      const result = checker.checkObservability();

      expect(result).toBe(false);
      expect((checker as any).violations).toHaveLength(1);
      expect((checker as any).violations[0][0]).toBe('Observability');
    });
  });

  describe('checkTraceability', () => {
    it('@TEST-TRACEABILITY-001: should pass with version control and documentation', () => {
      vi.spyOn(checker as any, 'hasVersionControl').mockReturnValue(true);
      vi.spyOn(checker as any, 'hasDocumentation').mockReturnValue(true);
      vi.spyOn(checker as any, 'hasTagSystem').mockReturnValue(true);

      const result = checker.checkTraceability();

      expect(result).toBe(true);
      expect((checker as any).violations).toHaveLength(0);
    });

    it('@TEST-TRACEABILITY-002: should fail without version control', () => {
      vi.spyOn(checker as any, 'hasVersionControl').mockReturnValue(false);
      vi.spyOn(checker as any, 'hasDocumentation').mockReturnValue(true);
      vi.spyOn(checker as any, 'hasTagSystem').mockReturnValue(false);

      const result = checker.checkTraceability();

      expect(result).toBe(false);
      expect((checker as any).violations).toHaveLength(1);
      expect((checker as any).violations[0][0]).toBe('Traceability');
    });
  });

  describe('checkAllPrinciples', () => {
    it('@TEST-ALL-PRINCIPLES-001: should check all TRUST principles', () => {
      const mockConfig = {
        constitution: {
          principles: {
            simplicity: { max_projects: 3 },
          },
        },
      };

      vi.spyOn(checker, 'loadConfig').mockReturnValue(mockConfig);
      vi.spyOn(checker, 'checkSimplicity').mockReturnValue(true);
      vi.spyOn(checker, 'checkArchitecture').mockReturnValue(true);
      vi.spyOn(checker, 'checkTesting').mockReturnValue(true);
      vi.spyOn(checker, 'checkObservability').mockReturnValue(true);
      vi.spyOn(checker, 'checkTraceability').mockReturnValue(true);

      const result = checker.checkAllPrinciples();

      expect(result.passed).toBe(true);
      expect(result.violations).toHaveLength(0);
      expect(result.summary).toEqual({
        simplicity: true,
        architecture: true,
        testing: true,
        observability: true,
        traceability: true,
      });
    });

    it('@TEST-ALL-PRINCIPLES-002: should report violations when principles fail', () => {
      const mockConfig = {};

      vi.spyOn(checker, 'loadConfig').mockReturnValue(mockConfig);

      // Mock methods to actually add violations
      vi.spyOn(checker, 'checkSimplicity').mockImplementation(() => {
        (checker as any).violations.push([
          'Simplicity',
          'Too many modules',
          'Reduce modules',
        ]);
        return false;
      });
      vi.spyOn(checker, 'checkArchitecture').mockImplementation(() => {
        (checker as any).violations.push([
          'Architecture',
          'No layered structure',
          'Add layers',
        ]);
        return false;
      });
      vi.spyOn(checker, 'checkTesting').mockReturnValue(true);
      vi.spyOn(checker, 'checkObservability').mockReturnValue(true);
      vi.spyOn(checker, 'checkTraceability').mockReturnValue(true);

      const result = checker.checkAllPrinciples();

      expect(result.passed).toBe(false);
      expect(result.violations).toHaveLength(2);
      expect(result.summary.simplicity).toBe(false);
      expect(result.summary.architecture).toBe(false);
    });
  });
});
