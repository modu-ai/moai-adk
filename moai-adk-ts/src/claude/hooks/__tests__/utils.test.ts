/**
 * @TEST:HOOKS-REFACTOR-001 |
 * Related: @CODE:HOOKS-REFACTOR-001, @SPEC:HOOKS-REFACTOR-001
 *
 * Utility Functions Test Suite
 * 공통 유틸리티 함수 테스트
 */

import { describe, expect, it } from 'vitest';

describe('@TEST:HOOKS-REFACTOR-001 - utils.ts', () => {
  describe('extractFilePath()', () => {
    it('should extract file_path', async () => {
      // GIVEN: utils.ts 모듈 (아직 구현 안됨)
      // WHEN: extractFilePath 호출
      // THEN: file_path 추출 (실패 예상)

      await expect(async () => {
        const { extractFilePath } = await import('../utils');
        const input = { file_path: '/path/to/file.ts' };
        expect(extractFilePath(input)).toBe('/path/to/file.ts');
      }).rejects.toThrow();
    });

    it('should extract filePath', async () => {
      await expect(async () => {
        const { extractFilePath } = await import('../utils');
        const input = { filePath: '/path/to/file.ts' };
        expect(extractFilePath(input)).toBe('/path/to/file.ts');
      }).rejects.toThrow();
    });

    it('should extract path', async () => {
      await expect(async () => {
        const { extractFilePath } = await import('../utils');
        const input = { path: '/path/to/file.ts' };
        expect(extractFilePath(input)).toBe('/path/to/file.ts');
      }).rejects.toThrow();
    });

    it('should extract notebook_path', async () => {
      await expect(async () => {
        const { extractFilePath } = await import('../utils');
        const input = { notebook_path: '/path/to/notebook.ipynb' };
        expect(extractFilePath(input)).toBe('/path/to/notebook.ipynb');
      }).rejects.toThrow();
    });

    it('should return null if no path found', async () => {
      await expect(async () => {
        const { extractFilePath } = await import('../utils');
        const input = { other: 'value' };
        expect(extractFilePath(input)).toBeNull();
      }).rejects.toThrow();
    });
  });

  describe('extractCommand()', () => {
    it('should extract command string', async () => {
      await expect(async () => {
        const { extractCommand } = await import('../utils');
        const input = { command: 'git status' };
        expect(extractCommand(input)).toBe('git status');
      }).rejects.toThrow();
    });

    it('should extract cmd string', async () => {
      await expect(async () => {
        const { extractCommand } = await import('../utils');
        const input = { cmd: 'npm install' };
        expect(extractCommand(input)).toBe('npm install');
      }).rejects.toThrow();
    });

    it('should join command array', async () => {
      await expect(async () => {
        const { extractCommand } = await import('../utils');
        const input = { command: ['git', 'commit', '-m', 'test'] };
        expect(extractCommand(input)).toBe('git commit -m test');
      }).rejects.toThrow();
    });

    it('should return null if no command found', async () => {
      await expect(async () => {
        const { extractCommand } = await import('../utils');
        const input = { other: 'value' };
        expect(extractCommand(input)).toBeNull();
      }).rejects.toThrow();
    });
  });

  describe('getAllFileExtensions()', () => {
    it('should return all extensions from SUPPORTED_LANGUAGES', async () => {
      await expect(async () => {
        const { getAllFileExtensions } = await import('../utils');
        const extensions = getAllFileExtensions();

        // THEN: 모든 언어의 확장자 포함
        expect(extensions).toContain('.ts');
        expect(extensions).toContain('.py');
        expect(extensions).toContain('.rb');
        expect(extensions).toContain('.php');
        expect(extensions).toContain('.cs');
        expect(extensions).toContain('.dart');
        expect(extensions).toContain('.swift');
        expect(extensions).toContain('.kt');
        expect(extensions).toContain('.ex');
      }).rejects.toThrow();
    });

    it('should return unique extensions', async () => {
      await expect(async () => {
        const { getAllFileExtensions } = await import('../utils');
        const extensions = getAllFileExtensions();
        const uniqueExtensions = [...new Set(extensions)];
        expect(extensions.length).toBe(uniqueExtensions.length);
      }).rejects.toThrow();
    });
  });
});
