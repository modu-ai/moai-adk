/**
 * @CODE:HOOKS-REFACTOR-001 |
 * SPEC: ../.moai/specs/SPEC-HOOKS-REFACTOR-001/spec.md |
 * TEST: __tests__/utils.test.ts
 *
 * Hook Utility Functions
 * 훅에서 공통으로 사용하는 유틸리티 함수
 */

import { SUPPORTED_LANGUAGES } from './constants';

/**
 * Extract file path from tool input
 *
 * @param toolInput - Tool input object
 * @returns File path or null if not found
 */
export function extractFilePath(
  toolInput: Record<string, any>
): string | null {
  return (
    toolInput.file_path ||
    toolInput.filePath ||
    toolInput.path ||
    toolInput.notebook_path ||
    null
  );
}

/**
 * Extract command from tool input
 *
 * @param toolInput - Tool input object
 * @returns Command string or null if not found
 */
export function extractCommand(
  toolInput: Record<string, any>
): string | null {
  const raw = toolInput.command || toolInput.cmd;

  if (Array.isArray(raw)) {
    return raw.map(String).join(' ');
  }

  if (typeof raw === 'string') {
    return raw.trim();
  }

  return null;
}

/**
 * Get all file extensions from supported languages
 *
 * @returns Array of all file extensions
 */
export function getAllFileExtensions(): string[] {
  return Object.values(SUPPORTED_LANGUAGES).flat();
}
