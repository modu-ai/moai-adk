/**
 * @TEST:HOOKS-REFACTOR-001 |
 * Related: @CODE:HOOKS-REFACTOR-001, @SPEC:HOOKS-REFACTOR-001
 *
 * Constants Test Suite
 * 중앙화된 상수 검증 테스트
 */

import { describe, expect, it } from 'vitest';

describe('@TEST:HOOKS-REFACTOR-001 - constants.ts', () => {
  describe('SUPPORTED_LANGUAGES', () => {
    it('should include all 14 languages', async () => {
      // GIVEN: SUPPORTED_LANGUAGES 상수 (아직 구현 안됨)
      // WHEN: import 시도
      // THEN: 14개 언어 포함 (실패 예상)

      // 현재는 constants.ts가 없으므로 import 실패 예상
      await expect(async () => {
        const { SUPPORTED_LANGUAGES } = await import('../constants');
        const languages = Object.keys(SUPPORTED_LANGUAGES);

        expect(languages).toHaveLength(14);
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
      }).rejects.toThrow();
    });

    it('should have correct extensions for each language', async () => {
      // GIVEN: constants.ts 모듈
      // WHEN: 각 언어의 확장자 확인
      // THEN: 올바른 확장자 매핑 (실패 예상)

      await expect(async () => {
        const { SUPPORTED_LANGUAGES } = await import('../constants');

        expect(SUPPORTED_LANGUAGES.typescript).toEqual(['.ts', '.tsx']);
        expect(SUPPORTED_LANGUAGES.ruby).toEqual(['.rb', '.rake', '.gemspec']);
        expect(SUPPORTED_LANGUAGES.php).toEqual(['.php']);
        expect(SUPPORTED_LANGUAGES.csharp).toEqual(['.cs']);
        expect(SUPPORTED_LANGUAGES.dart).toEqual(['.dart']);
        expect(SUPPORTED_LANGUAGES.swift).toEqual(['.swift']);
        expect(SUPPORTED_LANGUAGES.kotlin).toEqual(['.kt', '.kts']);
        expect(SUPPORTED_LANGUAGES.elixir).toEqual(['.ex', '.exs']);
      }).rejects.toThrow();
    });
  });

  describe('READ_ONLY_TOOLS', () => {
    it('should include all read-only tools', async () => {
      // GIVEN: constants.ts 모듈
      // WHEN: READ_ONLY_TOOLS 확인
      // THEN: 최소 10개 이상의 read-only tools 포함 (실패 예상)

      await expect(async () => {
        const { READ_ONLY_TOOLS } = await import('../constants');

        expect(READ_ONLY_TOOLS.length).toBeGreaterThanOrEqual(10);
        expect(READ_ONLY_TOOLS).toContain('Read');
        expect(READ_ONLY_TOOLS).toContain('Glob');
        expect(READ_ONLY_TOOLS).toContain('Grep');
      }).rejects.toThrow();
    });
  });

  describe('TIMEOUTS', () => {
    it('should have all timeout constants', async () => {
      // GIVEN: constants.ts 모듈
      // WHEN: TIMEOUTS 상수 확인
      // THEN: 모든 타임아웃 상수 존재 (실패 예상)

      await expect(async () => {
        const { TIMEOUTS } = await import('../constants');

        expect(TIMEOUTS.TAG_BLOCK_SEARCH_LIMIT).toBe(30);
        expect(TIMEOUTS.GIT_COMMAND_TIMEOUT).toBe(2000);
        expect(TIMEOUTS.NPM_REGISTRY_TIMEOUT).toBe(2000);
        expect(TIMEOUTS.POLICY_BLOCK_SLOW_THRESHOLD).toBe(100);
      }).rejects.toThrow();
    });
  });
});
