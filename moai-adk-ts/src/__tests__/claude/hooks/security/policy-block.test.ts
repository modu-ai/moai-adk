// @TEST:SEC-001 |
// Related: @CODE:SEC-POLICY-001

/**
 * @file Policy block security tests
 * @author MoAI Team
 */

import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';

import { PolicyBlock } from '../../../../claude/hooks/security/policy-block';
import type { HookInput } from '../../../../claude/hooks/types';

describe('PolicyBlock', () => {
  let policyBlock: PolicyBlock;

  beforeEach(() => {
    policyBlock = new PolicyBlock();

    // Mock console.error to avoid output during tests
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('execute', () => {
    it('should allow non-Bash tools', async () => {
      const input: HookInput = {
        tool_name: 'Read',
        tool_input: {
          file_path: '/some/file.txt',
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(true);
      expect(result.blocked).toBeUndefined();
    });

    it('should allow safe bash commands', async () => {
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

    it('should block dangerous rm -rf commands', async () => {
      const dangerousCommands = [
        'rm -rf /',
        'rm -rf --no-preserve-root',
        'sudo rm -rf /home',
      ];

      for (const command of dangerousCommands) {
        const input: HookInput = {
          tool_name: 'Bash',
          tool_input: { command },
        };

        const result = await policyBlock.execute(input);

        expect(result.success).toBe(false);
        expect(result.blocked).toBe(true);
        expect(result.message).toContain('위험 명령이 감지되었습니다');
        expect(result.exitCode).toBe(2);
      }
    });

    it('should block dangerous dd commands', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: {
          command: 'dd if=/dev/zero of=/dev/sda',
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
      expect(result.message).toContain('위험 명령이 감지되었습니다');
    });

    it('should block fork bombs', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: {
          command: ':(){:|:&};:',
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
      expect(result.message).toContain('위험 명령이 감지되었습니다');
    });

    it('should block mkfs commands', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: {
          command: 'mkfs.ext4 /dev/sdb1',
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
      expect(result.message).toContain('위험 명령이 감지되었습니다');
    });

    it('should handle array command format', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: {
          command: ['git', 'status'],
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(true);
      expect(result.blocked).toBeUndefined();
    });

    it('should handle cmd parameter name', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: {
          cmd: 'ls -la',
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(true);
      expect(result.blocked).toBeUndefined();
    });

    it('should handle missing command gracefully', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: {},
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(true);
      expect(result.blocked).toBeUndefined();
    });

    it('should show notice for unregistered commands', async () => {
      const input: HookInput = {
        tool_name: 'Bash',
        tool_input: {
          command: 'some-unknown-command --flag',
        },
      };

      const result = await policyBlock.execute(input);

      expect(result.success).toBe(true);
      expect(console.error).toHaveBeenCalledWith(
        expect.stringContaining('등록되지 않은 명령입니다')
      );
    });

    it('should allow whitelisted command prefixes', async () => {
      const allowedCommands = [
        'git push origin main',
        'python -m pytest',
        'npm install package',
        'node script.js',
        'go test ./...',
        'cargo build --release',
        'poetry install',
        'pnpm start',
        'rg pattern file.txt',
        'ls -la',
        'cat file.txt',
        'echo "hello"',
        'which node',
        'make build',
        'moai init project',
      ];

      for (const command of allowedCommands) {
        const input: HookInput = {
          tool_name: 'Bash',
          tool_input: { command },
        };

        const result = await policyBlock.execute(input);

        expect(result.success).toBe(true);
        expect(result.blocked).toBeUndefined();
      }
    });

    it('should be case sensitive for dangerous commands', async () => {
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
