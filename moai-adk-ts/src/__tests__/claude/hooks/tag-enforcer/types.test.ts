// @TEST:REFACTOR-003 | SPEC: SPEC-REFACTOR-003.md

/**
 * types.ts 테스트
 * TAG 시스템 타입 정의 검증
 */

import { describe, test, expect } from 'vitest';
import type {
  TagBlock,
  ImmutabilityCheck,
  ValidationResult
} from '../../../claude/hooks/tag-enforcer/types';

describe('@TEST:REFACTOR-003: TAG Types', () => {
  describe('TagBlock', () => {
    test('should accept valid TagBlock structure', () => {
      const tagBlock: TagBlock = {
        content: '/** @TAG:FEATURE:AUTH-001 */',
        lineNumber: 1
      };

      expect(tagBlock.content).toBe('/** @TAG:FEATURE:AUTH-001 */');
      expect(tagBlock.lineNumber).toBe(1);
    });
  });

  describe('ImmutabilityCheck', () => {
    test('should accept non-violated check', () => {
      const check: ImmutabilityCheck = {
        violated: false
      };

      expect(check.violated).toBe(false);
      expect(check.modifiedTag).toBeUndefined();
      expect(check.violationDetails).toBeUndefined();
    });

    test('should accept violated check with details', () => {
      const check: ImmutabilityCheck = {
        violated: true,
        modifiedTag: '@TAG:FEATURE:AUTH-001',
        violationDetails: '@IMMUTABLE TAG 블록이 수정되었습니다'
      };

      expect(check.violated).toBe(true);
      expect(check.modifiedTag).toBe('@TAG:FEATURE:AUTH-001');
      expect(check.violationDetails).toBe('@IMMUTABLE TAG 블록이 수정되었습니다');
    });
  });

  describe('ValidationResult', () => {
    test('should accept valid result with no violations', () => {
      const result: ValidationResult = {
        isValid: true,
        violations: [],
        warnings: [],
        hasTag: true
      };

      expect(result.isValid).toBe(true);
      expect(result.violations).toHaveLength(0);
      expect(result.warnings).toHaveLength(0);
      expect(result.hasTag).toBe(true);
    });

    test('should accept invalid result with violations', () => {
      const result: ValidationResult = {
        isValid: false,
        violations: ['Missing @TAG line'],
        warnings: ['Consider adding @IMMUTABLE'],
        hasTag: false
      };

      expect(result.isValid).toBe(false);
      expect(result.violations).toContain('Missing @TAG line');
      expect(result.warnings).toContain('Consider adding @IMMUTABLE');
      expect(result.hasTag).toBe(false);
    });
  });
});
