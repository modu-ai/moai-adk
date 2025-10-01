// @TEST:REFACTOR-002-VALIDATOR | -VALIDATOR
// Related: @CODE:TEMPLATE-VALIDATOR-001

/**
 * @file TemplateValidator Test Suite - Phase 1
 * @tags @TEST:REFACTOR-002-VALIDATOR
 *
 * Bottom-up TDD: TemplateValidator 분리 테스트
 * - 프로젝트 이름 검증
 * - 설정 검증
 * - 경로 검증
 */

import { beforeEach, describe, expect, it } from 'vitest';
import { TemplateValidator } from '@/core/project/template-validator';
import type { ProjectConfig } from '@/types/project';
import { ProjectType } from '@/types/project';

describe('TemplateValidator - Phase 1: Validation Logic', () => {
  let validator: TemplateValidator;

  beforeEach(() => {
    validator = new TemplateValidator();
  });

  describe('@TEST:VALIDATOR-NAME-001 - Project Name Validation', () => {
    it('should accept valid project names', () => {
      // RED: validateProjectName 메서드가 존재하지 않음
      expect(validator.validateProjectName('my-project')).toBe(true);
      expect(validator.validateProjectName('my_project')).toBe(true);
      expect(validator.validateProjectName('MyProject123')).toBe(true);
      expect(validator.validateProjectName('project-2024')).toBe(true);
    });

    it('should reject invalid project names', () => {
      expect(validator.validateProjectName('my project')).toBe(false); // 공백
      expect(validator.validateProjectName('my@project')).toBe(false); // 특수문자
      expect(validator.validateProjectName('my.project')).toBe(false); // 점
      expect(validator.validateProjectName('')).toBe(false); // 빈 문자열
      expect(validator.validateProjectName('my/project')).toBe(false); // 슬래시
    });

    it('should return validation errors for invalid names', () => {
      // RED: getValidationErrors 메서드가 존재하지 않음
      const errors = validator.getValidationErrors('my@project');
      expect(errors).toBeDefined();
      expect(errors.length).toBeGreaterThan(0);
      expect(errors[0]).toContain('Invalid project name');
    });
  });

  describe('@TEST:VALIDATOR-CONFIG-001 - Config Validation', () => {
    it('should validate complete project config', () => {
      const config: ProjectConfig = {
        name: 'test-project',
        type: ProjectType.TYPESCRIPT,
        description: 'Test project',
        author: 'Test Author',
        license: 'MIT',
      };

      // RED: validateConfig 메서드가 존재하지 않음
      const result = validator.validateConfig(config);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should detect missing required fields', () => {
      const incompleteConfig = {
        name: 'test-project',
        // type 필드 누락
      } as ProjectConfig;

      const result = validator.validateConfig(incompleteConfig);
      expect(result.isValid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
      expect(result.errors.some(e => e.includes('type'))).toBe(true);
    });

    it('should validate project type', () => {
      const config: ProjectConfig = {
        name: 'test-project',
        type: 'invalid-type' as any,
      };

      const result = validator.validateConfig(config);
      expect(result.isValid).toBe(false);
      expect(result.errors.some(e => e.includes('type'))).toBe(true);
    });
  });

  describe('@TEST:VALIDATOR-PATH-001 - Path Validation', () => {
    it('should validate safe paths', () => {
      // RED: validatePath 메서드가 존재하지 않음
      expect(validator.validatePath('/valid/project/path')).toBe(true);
      expect(validator.validatePath('./relative/path')).toBe(true);
      expect(validator.validatePath('../parent/path')).toBe(true);
    });

    it('should reject unsafe paths', () => {
      expect(validator.validatePath('/etc/passwd')).toBe(false); // 시스템 경로
      expect(validator.validatePath('/root')).toBe(false); // 루트 경로
      expect(validator.validatePath('/sys')).toBe(false); // 시스템 경로
      expect(validator.validatePath('')).toBe(false); // 빈 경로
    });

    it('should handle cross-platform paths', () => {
      // Windows 경로
      if (process.platform === 'win32') {
        expect(validator.validatePath('C:\\Users\\test\\project')).toBe(true);
        expect(validator.validatePath('C:\\Windows\\System32')).toBe(false);
      }

      // Unix 경로
      expect(validator.validatePath('/home/user/project')).toBe(true);
      expect(validator.validatePath('/tmp/test-project')).toBe(true);
    });
  });

  describe('@TEST:VALIDATOR-FEATURE-001 - Feature Validation', () => {
    it('should validate feature configuration', () => {
      const features = [
        { name: 'typescript', enabled: true },
        { name: 'jest', enabled: true },
      ];

      // RED: validateFeatures 메서드가 존재하지 않음
      const result = validator.validateFeatures(features);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should detect invalid feature names', () => {
      const features = [
        { name: 'unknown-feature', enabled: true },
        { name: '', enabled: true },
      ];

      const result = validator.validateFeatures(features);
      expect(result.isValid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
    });

    it('should validate feature compatibility', () => {
      // Python 프로젝트에 TypeScript 전용 기능
      const config: ProjectConfig = {
        name: 'test-project',
        type: ProjectType.PYTHON,
        features: [
          { name: 'typescript', enabled: true }, // 호환되지 않음
        ],
      };

      const result = validator.validateConfig(config);
      expect(result.isValid).toBe(false);
      expect(result.errors.some(e => e.includes('incompatible'))).toBe(true);
    });
  });

  describe('@TEST:VALIDATOR-ERROR-001 - Error Message Quality', () => {
    it('should provide clear error messages', () => {
      const config: ProjectConfig = {
        name: 'invalid name!',
        type: 'wrong-type' as any,
      };

      const result = validator.validateConfig(config);
      expect(result.isValid).toBe(false);

      // 에러 메시지가 명확해야 함
      for (const error of result.errors) {
        expect(error.length).toBeGreaterThan(10); // 최소한의 설명
        expect(error).toMatch(/[A-Z]/); // 대문자로 시작
      }
    });

    it('should collect all validation errors', () => {
      const config = {
        name: 'invalid name!',
        // 여러 필드 누락/오류
      } as ProjectConfig;

      const result = validator.validateConfig(config);
      expect(result.isValid).toBe(false);
      expect(result.errors.length).toBeGreaterThanOrEqual(2); // 여러 에러 수집
    });
  });
});
