/**
 * @file policy-block.ts
 * @description Policy guard for Bash commands in MoAI-ADK
 * @version 1.0.0
 * @tag @SEC:POLICY-BLOCK-013
 */

import type { HookInput, HookResult, MoAIHook } from '../types';

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
 * Policy Block Hook - TypeScript port of policy_block.py
 */
export class PolicyBlock implements MoAIHook {
  name = 'policy-block';

  async execute(input: HookInput): Promise<HookResult> {
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

    return { success: true };
  }

  /**
   * Extract command from tool input
   */
  private extractCommand(toolInput: Record<string, any>): string | null {
    const raw = toolInput['command'] || toolInput['cmd'];

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
