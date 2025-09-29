/**
 * SPEC 검증 시스템 테스트
 * @file spec-validator.test.ts
 * @description SpecValidator 클래스의 기능 테스트
 *
 * @TEST:SPEC-VALIDATOR-001 SPEC 메타데이터 검증 테스트
 * @TEST:DEPENDENCY-VALIDATION-001 의존성 검증 테스트
 * @TEST:STATUS-TRANSITION-001 상태 전환 검증 테스트
 */

import { beforeEach, describe, expect, test } from 'vitest';
import {
  type SpecMetadata,
  SpecPriority,
  SpecStatus,
} from '../../types/spec-metadata.js';
import { SpecValidator } from '../spec-validator.js';

describe('SpecValidator', () => {
  let validator: SpecValidator;

  beforeEach(() => {
    validator = new SpecValidator();
  });

  describe('메타데이터 기본 검증', () => {
    test('유효한 메타데이터 검증', () => {
      const validMetadata: SpecMetadata = {
        spec_id: 'SPEC-014',
        status: SpecStatus.DRAFT,
        priority: SpecPriority.HIGH,
        dependencies: [],
        tags: ['feature', 'backend'],
      };

      const result = validator.validateMetadata(validMetadata);

      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    test('잘못된 SPEC ID 형식 검증', () => {
      const invalidMetadata: SpecMetadata = {
        spec_id: 'INVALID-ID',
        status: SpecStatus.DRAFT,
        priority: SpecPriority.MEDIUM,
      };

      const result = validator.validateMetadata(invalidMetadata);

      expect(result.isValid).toBe(false);
      expect(result.errors).toContain(
        'Invalid spec_id format: INVALID-ID. Expected format: SPEC-XXX'
      );
    });

    test('잘못된 상태 값 검증', () => {
      const invalidMetadata: SpecMetadata = {
        spec_id: 'SPEC-014',
        status: 'invalid-status' as SpecStatus,
        priority: SpecPriority.MEDIUM,
      };

      const result = validator.validateMetadata(invalidMetadata);

      expect(result.isValid).toBe(false);
      expect(result.errors[0]).toContain('Invalid status: invalid-status');
    });

    test('잘못된 우선순위 값 검증', () => {
      const invalidMetadata: SpecMetadata = {
        spec_id: 'SPEC-014',
        status: SpecStatus.DRAFT,
        priority: 'invalid-priority' as SpecPriority,
      };

      const result = validator.validateMetadata(invalidMetadata);

      expect(result.isValid).toBe(false);
      expect(result.errors[0]).toContain('Invalid priority: invalid-priority');
    });
  });

  describe('상태 전환 검증', () => {
    test('draft에서 active로 전환 가능', () => {
      const canTransition = validator.validateStatusTransition(
        SpecStatus.DRAFT,
        SpecStatus.ACTIVE
      );

      expect(canTransition).toBe(true);
    });

    test('draft에서 completed로 직접 전환 불가능', () => {
      const canTransition = validator.validateStatusTransition(
        SpecStatus.DRAFT,
        SpecStatus.COMPLETED
      );

      expect(canTransition).toBe(false);
    });

    test('active에서 completed로 전환 가능', () => {
      const canTransition = validator.validateStatusTransition(
        SpecStatus.ACTIVE,
        SpecStatus.COMPLETED
      );

      expect(canTransition).toBe(true);
    });

    test('deprecated에서 다른 상태로 전환 불가능', () => {
      const canTransition = validator.validateStatusTransition(
        SpecStatus.DEPRECATED,
        SpecStatus.ACTIVE
      );

      expect(canTransition).toBe(false);
    });

    test('같은 상태로의 전환은 항상 가능', () => {
      const canTransition = validator.validateStatusTransition(
        SpecStatus.ACTIVE,
        SpecStatus.ACTIVE
      );

      expect(canTransition).toBe(true);
    });
  });

  describe('의존성 검증', () => {
    test('존재하는 의존성은 유효', () => {
      const allSpecs = new Map([
        [
          'SPEC-010',
          {
            spec_id: 'SPEC-010',
            status: SpecStatus.COMPLETED,
            priority: SpecPriority.MEDIUM,
          },
        ],
        [
          'SPEC-011',
          {
            spec_id: 'SPEC-011',
            status: SpecStatus.COMPLETED,
            priority: SpecPriority.MEDIUM,
          },
        ],
      ]);

      const metadata: SpecMetadata = {
        spec_id: 'SPEC-012',
        status: SpecStatus.DRAFT,
        priority: SpecPriority.HIGH,
        dependencies: ['SPEC-010', 'SPEC-011'],
      };

      const result = validator.validateMetadata(metadata, allSpecs);

      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    test('존재하지 않는 의존성 검출', () => {
      const allSpecs = new Map([
        [
          'SPEC-010',
          {
            spec_id: 'SPEC-010',
            status: SpecStatus.COMPLETED,
            priority: SpecPriority.MEDIUM,
          },
        ],
      ]);

      const metadata: SpecMetadata = {
        spec_id: 'SPEC-012',
        status: SpecStatus.DRAFT,
        priority: SpecPriority.HIGH,
        dependencies: ['SPEC-010', 'SPEC-999'], // SPEC-999는 존재하지 않음
      };

      const result = validator.validateMetadata(metadata, allSpecs);

      expect(result.isValid).toBe(false);
      expect(result.errors).toContain('Missing dependencies: SPEC-999');
    });
  });

  describe('경고 메시지', () => {
    test('태그가 없는 경우 경고', () => {
      const metadata: SpecMetadata = {
        spec_id: 'SPEC-014',
        status: SpecStatus.DRAFT,
        priority: SpecPriority.MEDIUM,
        tags: [],
      };

      const result = validator.validateMetadata(metadata);

      expect(result.warnings).toContain(
        'No tags specified. Consider adding tags for better organization.'
      );
    });

    test('활성 상태인데 의존성이 없는 경우 경고', () => {
      const metadata: SpecMetadata = {
        spec_id: 'SPEC-014',
        status: SpecStatus.ACTIVE,
        priority: SpecPriority.MEDIUM,
        dependencies: [],
      };

      const result = validator.validateMetadata(metadata);

      expect(result.warnings).toContain(
        'Active SPEC without dependencies. Verify if this is intentional.'
      );
    });
  });

  describe('의존성 그래프', () => {
    test('구현 가능한 SPEC 식별', () => {
      const readySpecs = validator.getReadyToImplementSpecs();

      // 의존성이 모두 완료된 draft 상태의 SPEC들이 반환되어야 함
      expect(readySpecs).toBeDefined();
      expect(Array.isArray(readySpecs)).toBe(true);
    });
  });

  describe('ripgrep 통합', () => {
    test('상태별 SPEC 검색 (비동기)', async () => {
      // ripgrep이 설치되지 않은 환경에서는 fallback 사용
      const activeSpecs = await validator.getSpecsByStatus(SpecStatus.ACTIVE);

      expect(Array.isArray(activeSpecs)).toBe(true);
    });

    test('태그별 SPEC 검색 (비동기)', async () => {
      // ripgrep이 설치되지 않은 환경에서는 빈 배열 반환
      const typescriptSpecs = await validator.getSpecsByTag('typescript');

      expect(Array.isArray(typescriptSpecs)).toBe(true);
    });

    test('일반 패턴 검색 (비동기)', async () => {
      const searchResults = await validator.searchSpecs('priority: high');

      expect(Array.isArray(searchResults)).toBe(true);
    });
  });
});
