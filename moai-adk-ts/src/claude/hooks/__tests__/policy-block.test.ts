/**
 * @TEST:POLICY-001 |
 * Related: @CODE:HOOK-001, @CODE:HOOK-001:API
 *
 * Policy Block Hook Test Suite
 * 위험한 명령어 차단 및 정책 검증 테스트
 */

import { describe, it, expect, beforeEach } from 'vitest';
import { PolicyBlock } from '../policy-block';
import type { HookInput, HookResult } from '../policy-block';

describe('PolicyBlock Hook', () => {
  let policyBlock: PolicyBlock;

  beforeEach(() => {
    policyBlock = new PolicyBlock();
  });

  describe('@TEST:POLICY-001-HAPPY: 정상 동작 시나리오', () => {
    it('should allow safe git commands', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: {
          command: 'git status',
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(true);
      expect(result.blocked).toBeUndefined();
    });

    it('should allow safe npm commands', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: {
          command: 'npm install',
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(true);
      expect(result.blocked).toBeUndefined();
    });

    it('should allow Python commands', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: {
          command: 'python --version',
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(true);
    });

    it('should pass through non-Bash tool invocations', async () => {
      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/test/file.txt',
          content: 'test',
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(true);
    });
  });

  describe('@TEST:POLICY-001-ERROR: 위험 명령 차단', () => {
    it('should block "rm -rf /" command', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: {
          command: 'rm -rf /',
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
      expect(result.message).toContain('위험 명령이 감지되었습니다');
      expect(result.exitCode).toBe(2);
    });

    it('should block "sudo rm" command', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: {
          command: 'sudo rm -rf /important/files',
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
    });

    it('should block fork bomb command', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: {
          command: ':(){:|:&};:',
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
    });

    it('should block mkfs commands', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: {
          command: 'mkfs.ext4 /dev/sda1',
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
    });
  });

  describe('@TEST:POLICY-001-EDGE: 경계 조건 처리', () => {
    it('should handle empty command', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: {
          command: '',
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(true);
    });

    it('should handle command as array', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: {
          command: ['git', 'status'],
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(true);
    });

    it('should handle missing tool_input', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(true);
    });

    it('should be case-insensitive for dangerous commands', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: {
          command: 'RM -RF /',
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
    });
  });
});
