/**
 * @TEST:PREWRITE-001 | 
 * Related: @CODE:HOOK-002, @CODE:HOOK-002:API
 *
 * Pre-Write Guard Hook Test Suite
 * 파일 쓰기 전 위험 패턴 검증 및 차단 테스트
 */

import { describe, it, expect, beforeEach } from 'vitest';
import { PreWriteGuard } from '../pre-write-guard';
import type { HookInput, HookResult } from '../pre-write-guard';

describe('PreWriteGuard Hook', () => {
  let preWriteGuard: PreWriteGuard;

  beforeEach(() => {
    preWriteGuard = new PreWriteGuard();
  });

  describe('@TEST:PREWRITE-001-HAPPY: 정상 파일 쓰기 허용', () => {
    it('should allow writing to normal files', async () => {
      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/src/index.ts',
          content: 'console.log("hello");'
        }
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(true);
      expect(result.blocked).toBeUndefined();
    });

    it('should allow editing normal files', async () => {
      const input: HookInput = {
        tool_name: 'Edit',
        tool_input: {
          file_path: '/project/src/utils.ts',
          old_string: 'old code',
          new_string: 'new code'
        }
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(true);
    });

    it('should allow MultiEdit operations', async () => {
      const input: HookInput = {
        tool_name: 'MultiEdit',
        tool_input: {
          file_path: '/project/src/config.ts',
          edits: []
        }
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(true);
    });

    it('should pass through non-write operations', async () => {
      const input: HookInput = {
        tool_name: 'Read',
        tool_input: {
          file_path: '/project/src/index.ts'
        }
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(true);
    });
  });

  describe('@TEST:PREWRITE-001-ERROR: 민감한 파일 차단', () => {
    it('should block writing to .env files', async () => {
      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/.env',
          content: 'API_KEY=secret'
        }
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
      expect(result.message).toContain('민감한 파일은 편집할 수 없습니다');
      expect(result.exitCode).toBe(2);
    });

    it('should block writing to secrets directory', async () => {
      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/secrets/credentials.json',
          content: '{}'
        }
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
    });

    it('should block writing to .git directory', async () => {
      const input: HookInput = {
        tool_name: 'Edit',
        tool_input: {
          file_path: '/project/.git/config',
          old_string: 'old',
          new_string: 'new'
        }
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
    });

    it('should block writing to .ssh directory', async () => {
      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/home/user/.ssh/id_rsa',
          content: 'private key'
        }
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
    });

    it('should block writing to .moai/memory/ directory', async () => {
      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/.moai/memory/development-guide.md',
          content: 'modified guide'
        }
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
    });
  });

  describe('@TEST:PREWRITE-001-EDGE: 경계 조건 처리', () => {
    it('should handle missing file_path gracefully', async () => {
      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          content: 'test'
        }
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(true);
    });

    it('should be case-insensitive for sensitive patterns', async () => {
      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/.ENV',
          content: 'test'
        }
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
    });

    it('should handle alternative field names for file_path', async () => {
      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          filePath: '/project/src/test.ts',
          content: 'test'
        }
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(true);
    });

    it('should handle path field name variant', async () => {
      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          path: '/project/.env.local',
          content: 'test'
        }
      };

      const result = await preWriteGuard.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
    });
  });
});
