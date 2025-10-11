/**
 * @CODE:HOOK-002 | @CODE:HOOKS-REFACTOR-001 |
 * Related: @CODE:HOOK-002:API, @CODE:PREWRITE-001
 *
 * Pre-Write Guard Hook
 * 파일 쓰기 전 위험 패턴 검증 및 차단
 */

import type { HookInput, HookResult, MoAIHook } from '../types';
import { PROTECTED_PATHS, SENSITIVE_KEYWORDS } from './constants';
import { runHook } from './base';
import { extractFilePath } from './utils';

// Re-export types for test compatibility
export type { HookInput, HookResult } from '../types';

/**
 * Pre-Write Guard Hook - TypeScript port of pre_write_guard.py
 */
export class PreWriteGuard implements MoAIHook {
  name = 'pre-write-guard';

  async execute(input: HookInput): Promise<HookResult> {
    const toolName = input.tool_name;

    // Only check write operations
    if (!toolName || !['Write', 'Edit', 'MultiEdit'].includes(toolName)) {
      return { success: true };
    }

    const toolInput = input.tool_input || {};
    const filePath = extractFilePath(toolInput);

    if (!this.checkFileSafety(filePath || '')) {
      return {
        success: false,
        blocked: true,
        message: '민감한 파일은 편집할 수 없습니다.',
        exitCode: 2,
      };
    }

    return { success: true };
  }

  /**
   * Check if file is safe to edit
   */
  private checkFileSafety(filePath: string): boolean {
    if (!filePath) {
      return true;
    }

    const pathLower = filePath.toLowerCase();

    // Check sensitive keywords
    for (const keyword of SENSITIVE_KEYWORDS) {
      if (pathLower.includes(keyword)) {
        return false;
      }
    }

    // Check protected paths (except templates)
    const isTemplate = filePath.includes('/templates/.moai/');
    if (!isTemplate) {
      for (const protectedPath of PROTECTED_PATHS) {
        if (filePath.includes(protectedPath)) {
          return false;
        }
      }
    }

    return true;
  }
}

// Execute if run directly
if (require.main === module) {
  runHook(PreWriteGuard).catch(() => {
    // Silent failure to avoid breaking Claude Code session
    process.exit(0);
  });
}
