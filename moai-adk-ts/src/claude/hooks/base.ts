/**
 * @CODE:HOOKS-REFACTOR-001 |
 * SPEC: ../.moai/specs/SPEC-HOOKS-REFACTOR-001/spec.md |
 * TEST: __tests__/base.test.ts
 *
 * Base Hook Utilities
 * 모든 훅에서 사용하는 공통 유틸리티
 */

import type { MoAIHook } from '../types';

/**
 * Run a hook instance with standardized CLI handling
 *
 * @param HookClass - Hook class constructor
 */
export async function runHook(HookClass: new () => MoAIHook): Promise<void> {
  try {
    const { parseClaudeInput, outputResult } = await import('../index');
    const input = await parseClaudeInput();
    const hook = new HookClass();
    const result = await hook.execute(input);
    outputResult(result);
  } catch (error) {
    console.error(
      `ERROR ${HookClass.name}: ${error instanceof Error ? error.message : 'Unknown error'}`
    );
    process.exit(1);
  }
}
