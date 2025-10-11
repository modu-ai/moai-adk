/**
 * @TEST:HOOKS-REFACTOR-001 |
 * Related: @CODE:HOOKS-REFACTOR-001, @SPEC:HOOKS-REFACTOR-001
 *
 * Utility Functions Test Suite
 * 공통 유틸리티 함수 테스트
 */

import { describe, expect, it } from 'vitest';
import {
  extractCommand,
  extractFilePath,
  getAllFileExtensions,
} from '../utils';

describe('@TEST:HOOKS-REFACTOR-001 - utils.ts', () => {
  describe('extractFilePath()', () => {
    it('should extract file_path', () => {
      const input = { file_path: '/path/to/file.ts' };
      expect(extractFilePath(input)).toBe('/path/to/file.ts');
    });

    it('should extract filePath', () => {
      const input = { filePath: '/path/to/file.ts' };
      expect(extractFilePath(input)).toBe('/path/to/file.ts');
    });

    it('should extract path', () => {
      const input = { path: '/path/to/file.ts' };
      expect(extractFilePath(input)).toBe('/path/to/file.ts');
    });

    it('should extract notebook_path', () => {
      const input = { notebook_path: '/path/to/notebook.ipynb' };
      expect(extractFilePath(input)).toBe('/path/to/notebook.ipynb');
    });

    it('should return null if no path found', () => {
      const input = { other: 'value' };
      expect(extractFilePath(input)).toBeNull();
    });
  });

  describe('extractCommand()', () => {
    it('should extract command string', () => {
      const input = { command: 'git status' };
      expect(extractCommand(input)).toBe('git status');
    });

    it('should extract cmd string', () => {
      const input = { cmd: 'npm install' };
      expect(extractCommand(input)).toBe('npm install');
    });

    it('should join command array', () => {
      const input = { command: ['git', 'commit', '-m', 'test'] };
      expect(extractCommand(input)).toBe('git commit -m test');
    });

    it('should return null if no command found', () => {
      const input = { other: 'value' };
      expect(extractCommand(input)).toBeNull();
    });
  });

  describe('getAllFileExtensions()', () => {
    it('should return all extensions from SUPPORTED_LANGUAGES', () => {
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
    });

    it('should return unique extensions', () => {
      const extensions = getAllFileExtensions();
      const uniqueExtensions = [...new Set(extensions)];
      expect(extensions.length).toBe(uniqueExtensions.length);
    });
  });
});
