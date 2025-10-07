// @TEST:GIT-LOCALE-001 | SPEC: SPEC-GIT-LOCALE-001.md
/**
 * @file Commit Message Locales Test Suite
 * @tags @TEST:GIT-LOCALE-001
 * @description 다국어 커밋 메시지 템플릿 검증
 */

import { describe, expect, it } from 'vitest';
import {
  type CommitLocale,
  CommitMessageTemplates,
  getTDDCommitMessage,
  getTDDCommitWithTag,
  getValidatedLocale,
  isValidCommitLocale,
  type TDDStage,
} from '../commit-message-locales';

describe('@TEST:GIT-LOCALE-001 - Commit Message Locales', () => {
  describe('Locale Validation', () => {
    it('should validate correct locales', () => {
      expect(isValidCommitLocale('ko')).toBe(true);
      expect(isValidCommitLocale('en')).toBe(true);
      expect(isValidCommitLocale('ja')).toBe(true);
      expect(isValidCommitLocale('zh')).toBe(true);
    });

    it('should reject invalid locales', () => {
      expect(isValidCommitLocale('fr')).toBe(false);
      expect(isValidCommitLocale('es')).toBe(false);
      expect(isValidCommitLocale('')).toBe(false);
      expect(isValidCommitLocale('invalid')).toBe(false);
    });

    it('should return validated locale with fallback', () => {
      expect(getValidatedLocale('ko')).toBe('ko');
      expect(getValidatedLocale('en')).toBe('en');
      expect(getValidatedLocale('ja')).toBe('ja');
      expect(getValidatedLocale('zh')).toBe('zh');
      expect(getValidatedLocale('invalid')).toBe('en');
      expect(getValidatedLocale(undefined)).toBe('en');
    });
  });

  describe('Template Structure', () => {
    it('should have templates for all locales', () => {
      expect(CommitMessageTemplates.ko).toBeDefined();
      expect(CommitMessageTemplates.en).toBeDefined();
      expect(CommitMessageTemplates.ja).toBeDefined();
      expect(CommitMessageTemplates.zh).toBeDefined();
    });

    it('should have all TDD stages in each locale', () => {
      const locales: CommitLocale[] = ['ko', 'en', 'ja', 'zh'];
      const stages: TDDStage[] = ['RED', 'GREEN', 'REFACTOR', 'DOCS'];

      for (const locale of locales) {
        for (const stage of stages) {
          expect(CommitMessageTemplates[locale][stage]).toBeDefined();
          expect(CommitMessageTemplates[locale][stage]).toContain('{message}');
        }
      }
    });

    it('should use consistent emojis across locales', () => {
      expect(CommitMessageTemplates.ko.RED).toContain('🔴');
      expect(CommitMessageTemplates.en.RED).toContain('🔴');
      expect(CommitMessageTemplates.ja.RED).toContain('🔴');
      expect(CommitMessageTemplates.zh.RED).toContain('🔴');

      expect(CommitMessageTemplates.ko.GREEN).toContain('🟢');
      expect(CommitMessageTemplates.en.GREEN).toContain('🟢');
      expect(CommitMessageTemplates.ja.GREEN).toContain('🟢');
      expect(CommitMessageTemplates.zh.GREEN).toContain('🟢');

      expect(CommitMessageTemplates.ko.REFACTOR).toContain('♻️');
      expect(CommitMessageTemplates.en.REFACTOR).toContain('♻️');
      expect(CommitMessageTemplates.ja.REFACTOR).toContain('♻️');
      expect(CommitMessageTemplates.zh.REFACTOR).toContain('♻️');

      expect(CommitMessageTemplates.ko.DOCS).toContain('📝');
      expect(CommitMessageTemplates.en.DOCS).toContain('📝');
      expect(CommitMessageTemplates.ja.DOCS).toContain('📝');
      expect(CommitMessageTemplates.zh.DOCS).toContain('📝');
    });
  });

  describe('getTDDCommitMessage', () => {
    it('should generate Korean commit messages', () => {
      const result = getTDDCommitMessage('ko', 'RED', '로그인 테스트 추가');
      expect(result).toBe('🔴 RED: 로그인 테스트 추가');
    });

    it('should generate English commit messages', () => {
      const result = getTDDCommitMessage('en', 'GREEN', 'implement login');
      expect(result).toBe('🟢 GREEN: implement login');
    });

    it('should generate Japanese commit messages', () => {
      const result = getTDDCommitMessage('ja', 'REFACTOR', 'コード改善');
      expect(result).toBe('♻️ REFACTOR: コード改善');
    });

    it('should generate Chinese commit messages', () => {
      const result = getTDDCommitMessage('zh', 'DOCS', '更新文档');
      expect(result).toBe('📝 DOCS: 更新文档');
    });

    it('should handle all TDD stages', () => {
      expect(getTDDCommitMessage('en', 'RED', 'test')).toContain('🔴 RED:');
      expect(getTDDCommitMessage('en', 'GREEN', 'impl')).toContain('🟢 GREEN:');
      expect(getTDDCommitMessage('en', 'REFACTOR', 'clean')).toContain(
        '♻️ REFACTOR:'
      );
      expect(getTDDCommitMessage('en', 'DOCS', 'doc')).toContain('📝 DOCS:');
    });

    it('should fallback to English for invalid locale', () => {
      // getValidatedLocale converts invalid to 'en'
      const locale = getValidatedLocale('invalid');
      const result = getTDDCommitMessage(locale, 'RED', 'test');
      expect(result).toBe('🔴 RED: test');
    });
  });

  describe('getTDDCommitWithTag', () => {
    it('should add RED @TAG correctly', () => {
      const result = getTDDCommitWithTag(
        'ko',
        'RED',
        '테스트 추가',
        'AUTH-001'
      );
      expect(result).toContain('🔴 RED: 테스트 추가');
      expect(result).toContain('@TEST:AUTH-001-RED');
    });

    it('should add GREEN @TAG correctly', () => {
      const result = getTDDCommitWithTag(
        'en',
        'GREEN',
        'implement feature',
        'AUTH-001'
      );
      expect(result).toContain('🟢 GREEN: implement feature');
      expect(result).toContain('@CODE:AUTH-001-GREEN');
    });

    it('should add REFACTOR @TAG correctly', () => {
      const result = getTDDCommitWithTag(
        'ja',
        'REFACTOR',
        'コード改善',
        'AUTH-001'
      );
      expect(result).toContain('♻️ REFACTOR: コード改善');
      expect(result).toContain('REFACTOR:AUTH-001-CLEAN');
    });

    it('should add DOCS @TAG correctly', () => {
      const result = getTDDCommitWithTag('zh', 'DOCS', '更新文档', 'AUTH-001');
      expect(result).toContain('📝 DOCS: 更新文档');
      expect(result).toContain('@DOC:AUTH-001');
    });

    it('should format with newline between message and tag', () => {
      const result = getTDDCommitWithTag('en', 'RED', 'test', 'FEAT-001');
      const lines = result.split('\n');
      expect(lines).toHaveLength(3); // message, empty line, tag
      expect(lines[0]).toContain('🔴 RED: test');
      expect(lines[1]).toBe('');
      expect(lines[2]).toContain('@TEST:FEAT-001-RED');
    });
  });

  describe('Locale Templates Access', () => {
    it('should allow direct access to locale templates', () => {
      expect(CommitMessageTemplates.localeTemplates.ko).toBe(
        CommitMessageTemplates.ko
      );
      expect(CommitMessageTemplates.localeTemplates.en).toBe(
        CommitMessageTemplates.en
      );
      expect(CommitMessageTemplates.localeTemplates.ja).toBe(
        CommitMessageTemplates.ja
      );
      expect(CommitMessageTemplates.localeTemplates.zh).toBe(
        CommitMessageTemplates.zh
      );
    });
  });

  describe('Real-world Usage Scenarios', () => {
    it('should generate complete TDD workflow commits in Korean', () => {
      const specId = 'LOGIN-001';

      const red = getTDDCommitWithTag(
        'ko',
        'RED',
        '로그인 실패 테스트',
        specId
      );
      expect(red).toMatch(/🔴 RED: 로그인 실패 테스트\n\n@TEST:LOGIN-001-RED/);

      const green = getTDDCommitWithTag('ko', 'GREEN', '로그인 구현', specId);
      expect(green).toMatch(/🟢 GREEN: 로그인 구현\n\n@CODE:LOGIN-001-GREEN/);

      const refactor = getTDDCommitWithTag(
        'ko',
        'REFACTOR',
        '로그인 코드 정리',
        specId
      );
      expect(refactor).toMatch(
        /♻️ REFACTOR: 로그인 코드 정리\n\nREFACTOR:LOGIN-001-CLEAN/
      );
    });

    it('should generate complete TDD workflow commits in English', () => {
      const specId = 'LOGIN-001';

      const red = getTDDCommitWithTag('en', 'RED', 'add login test', specId);
      expect(red).toMatch(/🔴 RED: add login test\n\n@TEST:LOGIN-001-RED/);

      const green = getTDDCommitWithTag(
        'en',
        'GREEN',
        'implement login',
        specId
      );
      expect(green).toMatch(
        /🟢 GREEN: implement login\n\n@CODE:LOGIN-001-GREEN/
      );

      const refactor = getTDDCommitWithTag(
        'en',
        'REFACTOR',
        'clean up login code',
        specId
      );
      expect(refactor).toMatch(
        /♻️ REFACTOR: clean up login code\n\nREFACTOR:LOGIN-001-CLEAN/
      );
    });
  });
});
