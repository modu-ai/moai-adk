// @TEST:REFACTOR-004 연결: @SPEC:REFACTOR-004 ->  -> @CODE:REFACTOR-004
/**
 * @file Commit Constants Test Suite
 * @tags @TEST:REFACTOR-004 @TEST-COMMIT-CONSTANTS-001
 * @description GitCommitTemplates 분리 및 타입 안전성 검증
 */

import { describe, it, expect } from 'vitest';
import { GitCommitTemplates } from '../commit-constants';

describe('@TEST:REFACTOR-004 - GitCommitTemplates', () => {
  describe('Commit Message Templates', () => {
    it('@TEST-COMMIT-CONSTANTS-001: should have correct template values', () => {
      expect(GitCommitTemplates.FEATURE).toBe('✨ feat: {message}');
      expect(GitCommitTemplates.BUGFIX).toBe('🐛 fix: {message}');
      expect(GitCommitTemplates.DOCS).toBe('📝 docs: {message}');
      expect(GitCommitTemplates.REFACTOR).toBe('♻️ refactor: {message}');
      expect(GitCommitTemplates.TEST).toBe('✅ test: {message}');
      expect(GitCommitTemplates.CHORE).toBe('🔧 chore: {message}');
      expect(GitCommitTemplates.STYLE).toBe('💄 style: {message}');
      expect(GitCommitTemplates.PERF).toBe('⚡ perf: {message}');
      expect(GitCommitTemplates.BUILD).toBe('👷 build: {message}');
      expect(GitCommitTemplates.CI).toBe('💚 ci: {message}');
      expect(GitCommitTemplates.REVERT).toBe('⏪ revert: {message}');
    });
  });

  describe('Template Application', () => {
    it('@TEST-COMMIT-CONSTANTS-002: should apply message to template', () => {
      const template = '✨ feat: {message}';
      const message = 'add login feature';
      expect(GitCommitTemplates.apply(template, message)).toBe('✨ feat: add login feature');
    });

    it('@TEST-COMMIT-CONSTANTS-003: should handle multiple placeholders', () => {
      const template = '{message} - {message}';
      const result = GitCommitTemplates.apply(template, 'test');
      // 첫 번째 placeholder만 치환
      expect(result).toBe('test - {message}');
    });
  });

  describe('Auto Commit Message Generation', () => {
    it('@TEST-COMMIT-CONSTANTS-004: should create auto commit without scope', () => {
      const result = GitCommitTemplates.createAutoCommit('feat');
      expect(result).toBe('✨ feat: Auto-generated commit');
    });

    it('@TEST-COMMIT-CONSTANTS-005: should create auto commit with scope', () => {
      const result = GitCommitTemplates.createAutoCommit('fix', 'auth');
      expect(result).toBe('🐛 fix(auth): Auto-generated commit');
    });

    it('@TEST-COMMIT-CONSTANTS-006: should handle unknown types', () => {
      const result = GitCommitTemplates.createAutoCommit('unknown');
      expect(result).toBe('📝 unknown: Auto-generated commit');
    });
  });

  describe('Checkpoint Message Generation', () => {
    it('@TEST-COMMIT-CONSTANTS-007: should create checkpoint message', () => {
      const result = GitCommitTemplates.createCheckpoint('Phase 1 complete');
      expect(result).toBe('🔖 checkpoint: Phase 1 complete');
    });
  });

  describe('Emoji Mapping', () => {
    it('@TEST-COMMIT-CONSTANTS-008: should return correct emoji for type', () => {
      expect(GitCommitTemplates.getEmoji('feat')).toBe('✨');
      expect(GitCommitTemplates.getEmoji('fix')).toBe('🐛');
      expect(GitCommitTemplates.getEmoji('docs')).toBe('📝');
      expect(GitCommitTemplates.getEmoji('refactor')).toBe('♻️');
    });

    it('@TEST-COMMIT-CONSTANTS-009: should return default emoji for unknown type', () => {
      expect(GitCommitTemplates.getEmoji('unknown')).toBe('📝');
      expect(GitCommitTemplates.getEmoji('')).toBe('📝');
    });
  });

  describe('Type Safety', () => {
    it('@TEST-COMMIT-CONSTANTS-010: should maintain as const type', () => {
      const template: '✨ feat: {message}' = GitCommitTemplates.FEATURE;
      expect(template).toBe('✨ feat: {message}');
    });
  });
});
