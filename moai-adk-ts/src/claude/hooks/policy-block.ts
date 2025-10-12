/**
 * @CODE:HOOK-001 | @CODE:HOOKS-REFACTOR-001 |
 * Related: @CODE:HOOK-001:API, @CODE:POLICY-001
 *
 * Policy Blocking Hook - Steering Guard
 * 위험한 프롬프트 차단 및 헌법/지침 무시 방지
 */

import type { HookInput, HookResult, MoAIHook } from '../types';
import { runHook } from './base';
import {
  ALLOWED_PREFIXES,
  DANGEROUS_COMMANDS,
  READ_ONLY_TOOLS,
  TIMEOUTS,
} from './constants';
import { extractCommand } from './utils';

// Re-export types for test compatibility
export type { HookInput, HookResult } from '../types';

/**
 * Policy Block Hook - TypeScript port of policy_block.py
 */
export class PolicyBlock implements MoAIHook {
  name = 'policy-block';

  async execute(input?: HookInput): Promise<HookResult> {
    const startTime = Date.now();

    // Handle missing input or tool_name
    if (!input || !input.tool_name) {
      return { success: true };
    }

    // Fast-track: Read-only tools bypass all checks
    if (this.isReadOnlyTool(input.tool_name)) {
      return { success: true };
    }

    // Only process Bash tool invocations
    if (input.tool_name !== 'Bash') {
      return { success: true };
    }

    const command = extractCommand(input.tool_input || {});
    if (!command) {
      return { success: true };
    }

    const commandLower = command.toLowerCase();

    // Check for dangerous commands
    for (const dangerousCommand of DANGEROUS_COMMANDS) {
      if (commandLower.includes(dangerousCommand)) {
        const duration = Date.now() - startTime;
        if (duration > TIMEOUTS.POLICY_BLOCK_SLOW_THRESHOLD) {
          console.error(`[policy-block] Blocked in ${duration}ms`);
        }
        return {
          success: false,
          blocked: true,
          message: `위험 명령이 감지되었습니다 (${dangerousCommand}).`,
          exitCode: 2,
        };
      }
    }

    // Check if command has allowed prefix
    if (!this.isAllowedPrefix(command)) {
      console.error(
        'NOTICE: 등록되지 않은 명령입니다. 필요 시 settings.json 의 allow 목록을 갱신하세요.'
      );
    }

    // Log slow executions
    const duration = Date.now() - startTime;
    if (duration > TIMEOUTS.POLICY_BLOCK_SLOW_THRESHOLD) {
      console.error(
        `[policy-block] Slow execution: ${duration}ms for ${input.tool_name}`
      );
    }

    return { success: true };
  }

  /**
   * Check if command starts with allowed prefix
   */
  private isAllowedPrefix(command: string): boolean {
    return ALLOWED_PREFIXES.some(prefix => command.startsWith(prefix));
  }

  /**
   * Check if tool is read-only and can bypass policy checks
   */
  private isReadOnlyTool(toolName: string): boolean {
    // Check for MCP tool pattern (mcp__*)
    if (toolName.startsWith('mcp__')) {
      return true;
    }

    // Check against known read-only tools
    return (READ_ONLY_TOOLS as readonly string[]).includes(toolName);
  }
}

// Execute if run directly
if (require.main === module) {
  runHook(PolicyBlock).catch(error => {
    console.error(
      `ERROR policy_block: ${error instanceof Error ? error.message : 'Unknown error'}`
    );
    process.exit(1);
  });
}
