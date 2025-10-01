// @CODE:REFACTOR-003 | SPEC: SPEC-REFACTOR-003.md | TEST: __tests__/claude/hooks/tag-enforcer/tag-patterns.test.ts

/**
 * TAG Pattern Definitions
 * @CODE:REFACTOR-003:DATA - CODE-FIRST TAG 정규식 패턴
 *
 * SPEC-REFACTOR-003 요구사항:
 * - 명확한 패턴 정의
 * - 재사용 가능한 정규식
 * - 유효한 TAG 카테고리 관리
 */

/**
 * Code-First TAG 패턴
 * @IMMUTABLE
 */
export const CODE_FIRST_PATTERNS = {
  // 전체 TAG 블록 매칭 (파일 최상단)
  TAG_BLOCK: /^\/\*\*\s*([\s\S]*?)\*\//m,

  // 핵심 TAG 라인들
  MAIN_TAG: /^\s*\*\s*@TAG:([A-Z]+):([A-Z0-9_-]+)\s*$/m,
  CHAIN_LINE: /^\s*\*\s*@CHAIN:\s*(.+)\s*$/m,
  DEPENDS_LINE: /^\s*\*\s*@DEPENDS:\s*(.+)\s*$/m,
  STATUS_LINE: /^\s*\*\s*@STATUS:\s*(\w+)\s*$/m,
  CREATED_LINE: /^\s*\*\s*@CREATED:\s*(\d{4}-\d{2}-\d{2})\s*$/m,
  IMMUTABLE_MARKER: /^\s*\*\s*@IMMUTABLE\s*$/m,

  // TAG 참조
  TAG_REFERENCE: /@([A-Z]+):([A-Z0-9-]+)/g,
} as const;

/**
 * 8-Core TAG 카테고리
 * @IMMUTABLE
 */
export const VALID_CATEGORIES = {
  // Lifecycle (필수 체인)
  lifecycle: ['SPEC', 'REQ', 'DESIGN', 'TASK', 'TEST'],

  // Implementation (선택적)
  implementation: ['FEATURE', 'API', 'FIX'],
} as const;
