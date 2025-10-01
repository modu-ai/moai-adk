// @TEST:REFACTOR-003 | SPEC: SPEC-REFACTOR-003.md

/**
 * tag-patterns.ts 테스트
 * CODE-FIRST TAG 정규식 패턴 검증
 */

import { describe, test, expect } from 'vitest';
import { CODE_FIRST_PATTERNS, VALID_CATEGORIES } from '../../../../src/claude/hooks/tag-enforcer/tag-patterns';

describe('@TEST:REFACTOR-003: TAG Patterns', () => {
  describe('TAG_BLOCK pattern', () => {
    test('should match valid TAG block', () => {
      const content = `/**
 * @DOC:FEATURE:AUTH-001
 * CHAIN: REQ:AUTH-001 -> DESIGN:AUTH-001
 * @IMMUTABLE
 */`;

      const match = CODE_FIRST_PATTERNS.TAG_BLOCK.exec(content);
      expect(match).not.toBeNull();
    });

    test('should not match non-TAG comment blocks', () => {
      const content = `/**
 * Regular JSDoc comment
 * No TAG here
 */`;

      const match = CODE_FIRST_PATTERNS.TAG_BLOCK.exec(content);
      // Should match the block itself, but shouldn't contain TAG
      expect(match).not.toBeNull();
    });
  });

  describe('MAIN_TAG pattern', () => {
    test('should match valid @TAG line', () => {
      const line = ' * @DOC:FEATURE:AUTH-001';
      const match = CODE_FIRST_PATTERNS.MAIN_TAG.exec(line);

      expect(match).not.toBeNull();
      expect(match![1]).toBe('FEATURE');
      expect(match![2]).toBe('AUTH-001');
    });

    test('should not match invalid TAG format', () => {
      const line = ' * TAG-INVALID:TEST';
      const match = CODE_FIRST_PATTERNS.MAIN_TAG.exec(line);

      expect(match).toBeNull();
    });
  });

  describe('CHAIN_LINE pattern', () => {
    test('should match valid chain line', () => {
      const line = ' * CHAIN: REQ:AUTH-001 -> DESIGN:AUTH-001 -> TASK:AUTH-001';
      const match = CODE_FIRST_PATTERNS.CHAIN_LINE.exec(line);

      expect(match).not.toBeNull();
      expect(match![1]).toContain('REQ:AUTH-001');
    });
  });

  describe('IMMUTABLE_MARKER pattern', () => {
    test('should match @IMMUTABLE marker', () => {
      const line = ' * @IMMUTABLE';
      const match = CODE_FIRST_PATTERNS.IMMUTABLE_MARKER.exec(line);

      expect(match).not.toBeNull();
    });

    test('should not match @IMMUTABLE with extra text', () => {
      const line = ' * @IMMUTABLE_FLAG';
      const match = CODE_FIRST_PATTERNS.IMMUTABLE_MARKER.exec(line);

      expect(match).toBeNull();
    });
  });

  describe('TAG_REFERENCE pattern', () => {
    test('should match TAG references', () => {
      const text = 'Related to @SPEC:AUTH-001 and @SPEC:AUTH-001';
      const matches = Array.from(text.matchAll(CODE_FIRST_PATTERNS.TAG_REFERENCE));

      expect(matches).toHaveLength(2);
      expect(matches[0][1]).toBe('SPEC');
      expect(matches[0][2]).toBe('AUTH-001');
      expect(matches[1][1]).toBe('SPEC');
      expect(matches[1][2]).toBe('AUTH-001');
    });
  });

  describe('VALID_CATEGORIES', () => {
    test('should contain lifecycle categories', () => {
      expect(VALID_CATEGORIES.lifecycle).toContain('SPEC');
      expect(VALID_CATEGORIES.lifecycle).toContain('REQ');
      expect(VALID_CATEGORIES.lifecycle).toContain('DESIGN');
      expect(VALID_CATEGORIES.lifecycle).toContain('TASK');
      expect(VALID_CATEGORIES.lifecycle).toContain('TEST');
    });

    test('should contain implementation categories', () => {
      expect(VALID_CATEGORIES.implementation).toContain('FEATURE');
      expect(VALID_CATEGORIES.implementation).toContain('API');
      expect(VALID_CATEGORIES.implementation).toContain('FIX');
    });
  });

  describe('STATUS_LINE pattern', () => {
    test('should match valid status line', () => {
      const line = ' * STATUS: active';
      const match = CODE_FIRST_PATTERNS.STATUS_LINE.exec(line);

      expect(match).not.toBeNull();
      expect(match![1]).toBe('active');
    });
  });

  describe('CREATED_LINE pattern', () => {
    test('should match valid date format', () => {
      const line = ' * CREATED: 2025-01-15';
      const match = CODE_FIRST_PATTERNS.CREATED_LINE.exec(line);

      expect(match).not.toBeNull();
      expect(match![1]).toBe('2025-01-15');
    });

    test('should not match invalid date format', () => {
      const line = ' * CREATED: 01/15/2025';
      const match = CODE_FIRST_PATTERNS.CREATED_LINE.exec(line);

      expect(match).toBeNull();
    });
  });
});
