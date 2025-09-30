// @TEST:REFACTOR-003 | SPEC: SPEC-REFACTOR-003.md

/**
 * 단순화된 tag-enforcer 테스트
 */

import { describe, test, expect } from 'vitest';
import { CodeFirstTAGEnforcer } from '../../../../src/claude/hooks/tag-enforcer-refactored';
import type { HookInput } from '../../../../src/claude/types';

describe('@TEST:REFACTOR-003: TAG Enforcer Simple Tests', () => {
  test('should process non-write operations', async () => {
    const enforcer = new CodeFirstTAGEnforcer();
    const input: HookInput = {
      tool_name: 'Read',
      tool_input: {
        file_path: '/test/service.ts'
      }
    };

    const result = await enforcer.execute(input);
    expect(result.success).toBe(true);
  });

  test('should skip test files', async () => {
    const enforcer = new CodeFirstTAGEnforcer();
    const input: HookInput = {
      tool_name: 'Write',
      tool_input: {
        file_path: '/test/service.test.ts',
        content: 'test content'
      }
    };

    const result = await enforcer.execute(input);
    expect(result.success).toBe(true);
  });

  test('should skip non-code files', async () => {
    const enforcer = new CodeFirstTAGEnforcer();
    const input: HookInput = {
      tool_name: 'Write',
      tool_input: {
        file_path: '/test/data.json',
        content: '{}'
      }
    };

    const result = await enforcer.execute(input);
    expect(result.success).toBe(true);
  });
});
