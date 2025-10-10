/**
 * @CODE:HOOKS-TYPES-001 | SPEC: SPEC-HOOKS-001.md
 *
 * Common types for Claude Code Hooks
 */

/**
 * Hook execution input
 */
export interface HookInput {
  [key: string]: unknown;
}

/**
 * Hook execution result
 */
export interface HookResult {
  success: boolean;
  message?: string;
  data?: unknown;
}

/**
 * MoAI Hook interface
 */
export interface MoAIHook {
  name: string;
  execute(input: HookInput): Promise<HookResult>;
}
