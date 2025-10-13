/**
 * @TEST:HOOKS-REFACTOR-001 |
 * Related: @CODE:HOOKS-REFACTOR-001, @SPEC:HOOKS-REFACTOR-001
 *
 * Constants Test Suite
 * 중앙화된 상수 검증 테스트
 */

import { describe, expect, it } from 'vitest';
import { READ_ONLY_TOOLS, SUPPORTED_LANGUAGES, TIMEOUTS } from '../constants';

describe('@TEST:HOOKS-REFACTOR-001 - constants.ts', () => {
  describe('SUPPORTED_LANGUAGES', () => {
    it('should include all 15 languages', () => {
      // GIVEN: SUPPORTED_LANGUAGES 상수
      // WHEN: 키 개수 확인
      const languages = Object.keys(SUPPORTED_LANGUAGES);

      // THEN: 15개 언어 포함 (markdown 포함)
      expect(languages).toHaveLength(15);
      expect(languages).toContain('typescript');
      expect(languages).toContain('javascript');
      expect(languages).toContain('python');
      expect(languages).toContain('java');
      expect(languages).toContain('go');
      expect(languages).toContain('rust');
      expect(languages).toContain('cpp');
      expect(languages).toContain('ruby');
      expect(languages).toContain('php');
      expect(languages).toContain('csharp');
      expect(languages).toContain('dart');
      expect(languages).toContain('swift');
      expect(languages).toContain('kotlin');
      expect(languages).toContain('elixir');
      expect(languages).toContain('markdown');
    });

    it('should have correct extensions for each language', () => {
      // THEN: 각 언어의 확장자가 올바름
      expect(SUPPORTED_LANGUAGES.typescript).toEqual(['.ts', '.tsx']);
      expect(SUPPORTED_LANGUAGES.ruby).toEqual(['.rb', '.rake', '.gemspec']);
      expect(SUPPORTED_LANGUAGES.php).toEqual(['.php']);
      expect(SUPPORTED_LANGUAGES.csharp).toEqual(['.cs']);
      expect(SUPPORTED_LANGUAGES.dart).toEqual(['.dart']);
      expect(SUPPORTED_LANGUAGES.swift).toEqual(['.swift']);
      expect(SUPPORTED_LANGUAGES.kotlin).toEqual(['.kt', '.kts']);
      expect(SUPPORTED_LANGUAGES.elixir).toEqual(['.ex', '.exs']);
    });
  });

  describe('READ_ONLY_TOOLS', () => {
    it('should include all read-only tools', () => {
      // THEN: 최소 10개 이상의 read-only tools 포함
      expect(READ_ONLY_TOOLS.length).toBeGreaterThanOrEqual(10);
      expect(READ_ONLY_TOOLS).toContain('Read');
      expect(READ_ONLY_TOOLS).toContain('Glob');
      expect(READ_ONLY_TOOLS).toContain('Grep');
    });
  });

  describe('TIMEOUTS', () => {
    it('should have all timeout constants', () => {
      // THEN: 모든 타임아웃 상수 존재
      expect(TIMEOUTS.TAG_BLOCK_SEARCH_LIMIT).toBe(30);
      expect(TIMEOUTS.GIT_COMMAND_TIMEOUT).toBe(2000);
      expect(TIMEOUTS.NPM_REGISTRY_TIMEOUT).toBe(2000);
      expect(TIMEOUTS.POLICY_BLOCK_SLOW_THRESHOLD).toBe(100);
    });
  });
});
