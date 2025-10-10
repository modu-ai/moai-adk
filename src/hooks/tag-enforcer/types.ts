// @CODE:REFACTOR-003 | SPEC: SPEC-REFACTOR-003.md | TEST: __tests__/claude/hooks/tag-enforcer/types.test.ts

/**
 * TAG Enforcer Type Definitions
 * @CODE:REFACTOR-003:DATA - TAG 시스템 타입 정의
 *
 * SPEC-REFACTOR-003 요구사항:
 * - 타입 안전성 보장
 * - 명확한 인터페이스 정의
 * - 재사용 가능한 타입 구조
 */

/**
 * TAG 블록 추출 결과
 */
export interface TagBlock {
  content: string;
  lineNumber: number;
}

/**
 * 불변성 검사 결과
 */
export interface ImmutabilityCheck {
  violated: boolean;
  modifiedTag?: string;
  violationDetails?: string;
}

/**
 * TAG 유효성 검증 결과
 */
export interface ValidationResult {
  isValid: boolean;
  violations: string[];
  warnings: string[];
  hasTag: boolean;
}
