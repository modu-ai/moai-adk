/**
 * @CODE:HOOK-001 |
 * Related: @CODE:HOOK-001:API, @CODE:POLICY-001
 *
 * Policy Blocking Hook - Steering Guard
 * 위험한 프롬프트 차단 및 헌법/지침 무시 방지
 */

import type { HookInput, HookResult, MoAIHook } from '../types';

// Re-export types for test compatibility
export type { HookInput, HookResult } from '../types';

/**
 * Dangerous commands that should always be blocked
 */
const DANGEROUS_COMMANDS: string[] = [
  'rm -rf /',
  'rm -rf --no-preserve-root',
  'sudo rm',
  'dd if=/dev/zero',
  ':(){:|:&};:',
  'mkfs.',
];

/**
 * Command prefixes that are allowed
 */
const ALLOWED_PREFIXES: string[] = [
  'git ',
  'python',
  'pytest',
  'npm ',
  'node ',
  'go ',
  'cargo ',
  'poetry ',
  'pnpm ',
  'rg ',
  'ls ',
  'cat ',
  'echo ',
  'which ',
  'make ',
  'moai ',
];

/**
 * Read-only tools that should bypass policy checks
 * These tools don't modify the system and can be fast-tracked
 */
const READ_ONLY_TOOLS: string[] = [
  'Read',
  'Glob',
  'Grep',
  'WebFetch',
  'WebSearch',
  'TodoWrite',
  'BashOutput',
  'mcp__context7__resolve-library-id',
  'mcp__context7__get-library-docs',
  'mcp__ide__getDiagnostics',
  'mcp__ide__executeCode',
];

/**
 * Policy Block Hook - TypeScript port of policy_block.py
 */
export class PolicyBlock implements MoAIHook {
  name = 'policy-block';

  async execute(input: HookInput): Promise<HookResult> {
    const startTime = Date.now();

    // Fast-track: Read-only tools bypass all checks
    if (this.isReadOnlyTool(input.tool_name)) {
      return { success: true };
    }

    // Only process Bash tool invocations
    if (input.tool_name !== 'Bash') {
      return { success: true };
    }

    const command = this.extractCommand(input.tool_input || {});
    if (!command) {
      return { success: true };
    }

    const commandLower = command.toLowerCase();

    // Check for dangerous commands
    for (const dangerousCommand of DANGEROUS_COMMANDS) {
      if (commandLower.includes(dangerousCommand)) {
        const duration = Date.now() - startTime;
        if (duration > 100) {
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

    // Log slow executions (>100ms)
    const duration = Date.now() - startTime;
    if (duration > 100) {
      console.error(
        `[policy-block] Slow execution: ${duration}ms for ${input.tool_name}`
      );
    }

    return { success: true };
  }

  /**
   * Extract command from tool input
   */
  private extractCommand(toolInput: Record<string, any>): string | null {
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
    return READ_ONLY_TOOLS.includes(toolName);
  }
}

/**
 * CLI entry point for Claude Code compatibility
 */
export async function main(): Promise<void> {
  try {
    const { parseClaudeInput, outputResult } = await import('../index');
    const input = await parseClaudeInput();
    const policyBlock = new PolicyBlock();
    const result = await policyBlock.execute(input);
    outputResult(result);
  } catch (error) {
    console.error(
      `ERROR policy_block: ${error instanceof Error ? error.message : 'Unknown error'}`
    );
    process.exit(1);
  }
}

// Execute if run directly
if (require.main === module) {
  main().catch(error => {
    console.error(
      `ERROR policy_block: ${error instanceof Error ? error.message : 'Unknown error'}`
    );
    process.exit(1);
  });
}
