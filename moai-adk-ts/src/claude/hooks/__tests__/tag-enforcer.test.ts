/**
 * @TEST:ENFORCER-001 | Chain: @SPEC:TAG-001 -> @SPEC:CODE-FIRST-001 -> @CODE:ENFORCER-001 -> @TEST:ENFORCER-001
 * Related: @CODE:HOOK-004, @CODE:HOOK-004:API
 *
 * TAG Enforcer Hook Test Suite
 * Code-First TAG 블록 불변성 및 8-Core TAG 체계 검증 테스트
 */

import * as fs from 'node:fs/promises';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { HookInput } from '../tag-enforcer';
import { CodeFirstTAGEnforcer } from '../tag-enforcer';

// Mock fs/promises
vi.mock('fs/promises', () => ({
  readFile: vi.fn(),
  writeFile: vi.fn(),
}));

describe('CodeFirstTAGEnforcer Hook', () => {
  let enforcer: CodeFirstTAGEnforcer;

  beforeEach(() => {
    vi.clearAllMocks();
    enforcer = new CodeFirstTAGEnforcer();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('@TEST:ENFORCER-001-HAPPY: 정상 TAG 블록 검증', () => {
    it('should allow valid TAG block', async () => {
      const validContent = `/**
 * @DOC:FEATURE:AUTH-001
 * CHAIN: REQ:AUTH-001 -> DESIGN:AUTH-001 -> TASK:AUTH-001 -> TEST:AUTH-001
 * DEPENDS: NONE
 * STATUS: active
 * CREATED: 2025-01-01
 * @IMMUTABLE
 */

export class AuthService {}`;

      vi.spyOn(fs, 'readFile').mockResolvedValue('');

      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/src/auth.ts',
          content: validContent,
        },
      };

      const result = await enforcer.execute(input);

      expect(result.success).toBe(true);
      expect(result.message).toContain('검증 완료');
    });

    it('should pass through non-write operations', async () => {
      const input: HookInput = {
        tool_name: 'Read',
        tool_input: {
          file_path: '/project/src/auth.ts',
        },
      };

      const result = await enforcer.execute(input);

      expect(result.success).toBe(true);
    });

    it('should allow files without TAG blocks (warning only)', async () => {
      const content = 'export function helper() {}';

      vi.spyOn(fs, 'readFile').mockResolvedValue('');

      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/src/utils.ts',
          content,
        },
      };

      const result = await enforcer.execute(input);

      expect(result.success).toBe(true);
    });

    it('should skip test files', async () => {
      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/src/auth.test.ts',
          content: 'test content',
        },
      };

      const result = await enforcer.execute(input);

      expect(result.success).toBe(true);
    });
  });

  describe('@TEST:ENFORCER-001-IMMUTABLE: @IMMUTABLE TAG 보호', () => {
    it('should block modification of @IMMUTABLE TAG block', async () => {
      const oldContent = `/**
 * @DOC:FEATURE:AUTH-001
 * CHAIN: REQ:AUTH-001 -> DESIGN:AUTH-001
 * STATUS: active
 * CREATED: 2025-01-01
 * @IMMUTABLE
 */

export class AuthService {}`;

      const newContent = `/**
 * @DOC:FEATURE:AUTH-002
 * CHAIN: REQ:AUTH-002 -> DESIGN:AUTH-002
 * STATUS: active
 * CREATED: 2025-01-02
 * @IMMUTABLE
 */

export class AuthService {}`;

      vi.spyOn(fs, 'readFile').mockResolvedValue(oldContent);

      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/src/auth.ts',
          content: newContent,
        },
      };

      const result = await enforcer.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
      expect(result.message).toContain('@IMMUTABLE TAG 수정 금지');
    });

    it('should block deletion of @IMMUTABLE TAG block', async () => {
      const oldContent = `/**
 * @DOC:FEATURE:AUTH-001
 * @IMMUTABLE
 */

export class AuthService {}`;

      const newContent = 'export class AuthService {}';

      vi.spyOn(fs, 'readFile').mockResolvedValue(oldContent);

      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/src/auth.ts',
          content: newContent,
        },
      };

      const result = await enforcer.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
    });

    it('should allow modifying non-immutable TAG blocks', async () => {
      const oldContent = `/**
 * @DOC:FEATURE:AUTH-001
 * STATUS: active
 */

export class AuthService {}`;

      const newContent = `/**
 * @DOC:FEATURE:AUTH-001
 * STATUS: completed
 */

export class AuthService {}`;

      vi.spyOn(fs, 'readFile').mockResolvedValue(oldContent);

      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/src/auth.ts',
          content: newContent,
        },
      };

      const result = await enforcer.execute(input);

      expect(result.success).toBe(true);
    });
  });

  describe('@TEST:ENFORCER-001-VALIDATION: TAG 유효성 검증', () => {
    it('should reject invalid TAG category', async () => {
      const invalidContent = `/**
 * @DOC:INVALID:AUTH-001
 * CHAIN: REQ:AUTH-001 -> DESIGN:AUTH-001
 */

export class AuthService {}`;

      vi.spyOn(fs, 'readFile').mockResolvedValue('');

      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/src/auth.ts',
          content: invalidContent,
        },
      };

      const result = await enforcer.execute(input);

      expect(result.success).toBe(false);
      expect(result.blocked).toBe(true);
    });

    it('should validate TAG chain format', async () => {
      const contentWithChain = `/**
 * @DOC:FEATURE:AUTH-001
 * CHAIN: REQ:AUTH-001 -> DESIGN:AUTH-001 -> TASK:AUTH-001
 * @IMMUTABLE
 */

export class AuthService {}`;

      vi.spyOn(fs, 'readFile').mockResolvedValue('');

      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/src/auth.ts',
          content: contentWithChain,
        },
      };

      const result = await enforcer.execute(input);

      expect(result.success).toBe(true);
    });

    it('should validate dependencies format', async () => {
      const contentWithDeps = `/**
 * @DOC:FEATURE:AUTH-001
 * DEPENDS: API:USER-001, DATA:USER-001
 * @IMMUTABLE
 */

export class AuthService {}`;

      vi.spyOn(fs, 'readFile').mockResolvedValue('');

      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/src/auth.ts',
          content: contentWithDeps,
        },
      };

      const result = await enforcer.execute(input);

      expect(result.success).toBe(true);
    });
  });

  describe('@TEST:ENFORCER-001-EDGE: 경계 조건 처리', () => {
    it('should handle new file creation', async () => {
      vi.spyOn(fs, 'readFile').mockRejectedValue(new Error('ENOENT'));

      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/src/new-file.ts',
          content: 'export const value = 1;',
        },
      };

      const result = await enforcer.execute(input);

      expect(result.success).toBe(true);
    });

    it('should handle Edit tool with new_string', async () => {
      vi.spyOn(fs, 'readFile').mockResolvedValue('old content');

      const input: HookInput = {
        tool_name: 'Edit',
        tool_input: {
          file_path: '/project/src/file.ts',
          new_string: 'new content',
        },
      };

      const result = await enforcer.execute(input);

      expect(result.success).toBe(true);
    });

    it('should handle MultiEdit with edits array', async () => {
      vi.spyOn(fs, 'readFile').mockResolvedValue('');

      const input: HookInput = {
        tool_name: 'MultiEdit',
        tool_input: {
          file_path: '/project/src/file.ts',
          edits: [{ new_string: 'edit 1' }, { new_string: 'edit 2' }],
        },
      };

      const result = await enforcer.execute(input);

      expect(result.success).toBe(true);
    });

    it('should handle errors gracefully', async () => {
      vi.spyOn(fs, 'readFile').mockRejectedValue(
        new Error('Permission denied')
      );

      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/src/file.ts',
          content: 'content',
        },
      };

      const result = await enforcer.execute(input);

      // Should not block on errors
      expect(result.success).toBe(true);
    });

    it('should skip node_modules and dist directories', async () => {
      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/project/node_modules/package/index.ts',
          content: 'module code',
        },
      };

      const result = await enforcer.execute(input);

      expect(result.success).toBe(true);
    });
  });

  describe('@TEST:HOOKS-REFACTOR-001 - Multilang Support', () => {
    it('should recognize Ruby files', async () => {
      // GIVEN: Ruby 파일 (아직 getAllFileExtensions()이 없으므로 실패 예상)
      const input: HookInput = {
        tool_name: 'Write',
        tool_input: {
          file_path: '/path/to/service.rb',
          content: '# Ruby code',
        },
      };

      // Mock fs to simulate new file
      vi.spyOn(fs, 'readFile').mockRejectedValue(new Error('ENOENT'));

      // WHEN: tag-enforcer 실행
      const result = await enforcer.execute(input);

      // THEN: 현재는 Ruby 파일을 인식하지 못할 수 있음
      // (getAllFileExtensions 구현 후 통과 예상)
      expect(result.success).toBe(true);
    });

    it('should recognize all new language extensions', async () => {
      const newLanguages = [
        { ext: '.rb', lang: 'Ruby' },
        { ext: '.php', lang: 'PHP' },
        { ext: '.cs', lang: 'C#' },
        { ext: '.dart', lang: 'Dart' },
        { ext: '.swift', lang: 'Swift' },
        { ext: '.kt', lang: 'Kotlin' },
        { ext: '.ex', lang: 'Elixir' },
      ];

      for (const { ext, lang } of newLanguages) {
        const input: HookInput = {
          tool_name: 'Write',
          tool_input: {
            file_path: `/path/to/file${ext}`,
            content: `// ${lang} code`,
          },
        };

        vi.spyOn(fs, 'readFile').mockRejectedValue(new Error('ENOENT'));

        const result = await enforcer.execute(input);

        // 현재는 일부 확장자를 인식하지 못할 수 있음
        // (getAllFileExtensions 구현 후 모두 통과 예상)
        expect(result.success).toBe(true);
      }
    });
  });
});
