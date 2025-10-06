// @TEST:INIT-001 | SPEC: SPEC-INIT-001.md | CODE: src/utils/tty-detector.ts
// Related: @CODE:INIT-001:TTY, @SPEC:INIT-001

/**
 * @file Test for TTY detection utility
 * @author MoAI Team
 * @tags @TEST:INIT-001:TTY
 */

import { describe, test, expect } from 'vitest';
import { isTTYAvailable } from '@/utils/tty-detector';

describe('TTY Detector', () => {
  describe('isTTYAvailable', () => {
    test('should check TTY availability based on current environment', () => {
      // When: TTY 감지 수행
      const result = isTTYAvailable();

      // Then: boolean 값 반환 (현재 환경에 따라)
      expect(typeof result).toBe('boolean');
    });

    test('should safely handle TTY check without throwing', () => {
      // When: TTY 감지 수행
      const check = () => isTTYAvailable();

      // Then: 예외 발생하지 않음
      expect(check).not.toThrow();
    });

    test('should return false in non-TTY environment (expected)', () => {
      // Given: Test environment is typically non-TTY (CI/CD, Claude Code)
      // When: TTY 감지 수행
      const result = isTTYAvailable();

      // Then: false 반환 (비대화형 모드로 안전하게 폴백)
      // Note: In test environment, this is expected to be false
      expect(result).toBe(false);
    });
  });
});
