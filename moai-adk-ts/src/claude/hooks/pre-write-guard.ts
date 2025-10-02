/**
 * @CODE:HOOK-002 |
 * Related: @CODE:HOOK-002:API, @CODE:PREWRITE-001
 *
 * Pre-Write Guard Hook
 * 파일 쓰기 전 위험 패턴 검증 및 차단
 */

import type { HookInput, HookResult, MoAIHook } from '../types';

// Re-export types for test compatibility
export type { HookInput, HookResult } from '../types';

/**
 * Sensitive file patterns that should be protected
 */
const SENSITIVE_KEYWORDS: string[] = ['.env', '/secrets', '/.git/', '/.ssh'];

/**
 * Protected paths that should not be modified
 * Note: Templates in moai-adk-ts/templates/.moai/memory/ are allowed
 */
const PROTECTED_PATHS: string[] = [];

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
    const filePath = this.extractFilePath(toolInput);

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
   * Extract file path from tool input
   */
  private extractFilePath(toolInput: Record<string, any>): string | null {
    return toolInput.file_path || toolInput.filePath || toolInput.path || null;
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

    // Check protected paths
    for (const protectedPath of PROTECTED_PATHS) {
      if (filePath.includes(protectedPath)) {
        return false;
      }
    }

    return true;
  }
}

/**
 * CLI entry point for Claude Code compatibility
 */
export async function main(): Promise<void> {
  try {
    const { parseClaudeInput, outputResult } = await import('../index');
    const input = await parseClaudeInput();
    const preWriteGuard = new PreWriteGuard();
    const result = await preWriteGuard.execute(input);
    outputResult(result);
  } catch (_error) {
    // Silent failure to avoid breaking Claude Code session
    process.exit(0);
  }
}

// Execute if run directly
if (require.main === module) {
  main().catch(() => {
    // Silent failure to avoid breaking Claude Code session
    process.exit(0);
  });
}
