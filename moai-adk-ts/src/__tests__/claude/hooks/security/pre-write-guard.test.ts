/**
 * @file pre-write-guard.test.ts
 * @description Tests for PreWriteGuard hook
 */

import { PreWriteGuard } from '../../../../claude/hooks/pre-write-guard';
import type { HookInput } from '../../../../claude/hooks/types';

describe('PreWriteGuard', () => {
  let preWriteGuard: PreWriteGuard;

  beforeEach(() => {
    preWriteGuard = new PreWriteGuard();
  });

  describe('execute', () => {
    it('should allow non-write operations', async () => {
      const input: HookInput = {
        tool_name: 'Read',
        tool_input: {
          file_path: '/some/file.txt',
        },
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(true);
      expect(result.blocked).toBeUndefined();
    });

    it('should allow safe file writes', async () => {
      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/src/main.py',
          content: 'print("hello world")',
        },
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(true);
      expect(result.blocked).toBeUndefined();
    });

    it('should block .env file writes', async () => {
      const envFiles = [
        '.env',
        'project/.env',
        '/path/to/.env.local',
        'config/.env.production',
      ];

      for (const filePath of envFiles) {
        const input: HookInput = {
          tool_name: 'Write',
          tool_input: { file_path: filePath },
        };

        const result = await preWriteGuard.execute(input);

        expect(result.success).toBe(false);
        expect(result.blocked).toBe(true);
        expect(result.message).toContain('민감한 파일은 편집할 수 없습니다');
        expect(result.exitCode).toBe(2);
      }
    });

    it('should block secrets directory writes', async () => {
      const secretPaths = [
        '/secrets/api_key.txt',
        'project/secrets/config.json',
        '/app/secrets/database.env',
      ];

      for (const filePath of secretPaths) {
        const input: HookInput = {
          tool_name: 'Edit',
          tool_input: { file_path: filePath },
        };

        const result = await preWriteGuard.execute(input);

        expect(result.success).toBe(false);
        expect(result.blocked).toBe(true);
        expect(result.message).toContain('민감한 파일은 편집할 수 없습니다');
      }
    });

    it('should block .git directory writes', async () => {
      const gitPaths = [
        '/.git/config',
        'project/.git/HEAD',
        '/repo/.git/hooks/pre-commit',
      ];

      for (const filePath of gitPaths) {
        const input: HookInput = {
          tool_name: 'MultiEdit',
          tool_input: { file_path: filePath },
        };

        const result = await preWriteGuard.execute(input);

        expect(result.success).toBe(false);
        expect(result.blocked).toBe(true);
      }
    });

    it('should block .ssh directory writes', async () => {
      const sshPaths = [
        '/.ssh/id_rsa',
        '/home/user/.ssh/authorized_keys',
        '/.ssh/config',
      ];

      for (const filePath of sshPaths) {
        const input: HookInput = {
          tool_name: 'Write',
          tool_input: { file_path: filePath },
        };

        const result = await preWriteGuard.execute(input);

        expect(result.success).toBe(false);
        expect(result.blocked).toBe(true);
      }
    });

    it('should block protected .moai/memory/ directory writes', async () => {
      const memoryPaths = [
        '.moai/memory/development-guide.md',
        'project/.moai/memory/config.json',
        '/app/.moai/memory/secrets.txt',
      ];

      for (const filePath of memoryPaths) {
        const input: HookInput = {
          tool_name: 'Edit',
          tool_input: { file_path: filePath },
        };

        const result = await preWriteGuard.execute(input);

        expect(result.success).toBe(false);
        expect(result.blocked).toBe(true);
      }
    });

    it('should handle different file_path parameter names', async () => {
      const parameterNames = [
        { file_path: '/safe/file.txt' },
        { filePath: '/safe/file.txt' },
        { path: '/safe/file.txt' },
      ];

      for (const toolInput of parameterNames) {
        const input: HookInput = {
          tool_name: 'Write',
          tool_input: toolInput,
        };

        const result = await preWriteGuard.execute(input);

        expect(result.success).toBe(true);
        expect(result.blocked).toBeUndefined();
      }
    });

    it('should handle missing file path gracefully', async () => {
      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          content: 'some content',
        },
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(true);
      expect(result.blocked).toBeUndefined();
    });

    it('should handle empty tool input', async () => {
      const input: HookInput = {
        tool_name: 'Edit',
        tool_input: {},
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(true);
      expect(result.blocked).toBeUndefined();
    });

    it('should be case insensitive for path checking', async () => {
      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/PROJECT/.ENV',
        },
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
    });

    it('should allow .moai/project/ directory writes', async () => {
      // Note: .moai/project/ is not in protected paths, only .moai/memory/
      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '.moai/project/structure.md',
        },
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(true);
      expect(result.blocked).toBeUndefined();
    });

    it('should handle supported write operations', async () => {
      const writeOperations = ['Write', 'Edit', 'MultiEdit'];

      for (const toolName of writeOperations) {
        const input: HookInput = {
          tool_name: toolName,
          tool_input: {
            file_path: '/safe/path/file.txt',
          },
        };

        const result = await preWriteGuard.execute(input);

        expect(result.success).toBe(true);
        expect(result.blocked).toBeUndefined();
      }
    });
  });
});
