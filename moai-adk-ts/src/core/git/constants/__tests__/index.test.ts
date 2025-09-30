// @TEST:REFACTOR-004 연결: @SPEC:REFACTOR-004 ->  -> @CODE:REFACTOR-004
/**
 * @file Barrel Export Test Suite
 * @tags @TEST:REFACTOR-004 @TEST-BARREL-EXPORT-001
 * @description index.ts barrel export 검증 및 import 경로 호환성 테스트
 */

import { describe, it, expect } from 'vitest';

describe('@TEST:REFACTOR-004 - Barrel Export', () => {
  describe('Import Path Compatibility', () => {
    it('@TEST-BARREL-EXPORT-001: should import all exports from index', async () => {
      const module = await import('../index');

      // Branch constants
      expect(module.GitNamingRules).toBeDefined();

      // Commit constants
      expect(module.GitCommitTemplates).toBeDefined();

      // Config constants
      expect(module.GitDefaults).toBeDefined();
      expect(module.GitignoreTemplates).toBeDefined();
      expect(module.GitHubDefaults).toBeDefined();
      expect(module.GitTimeouts).toBeDefined();
    });

    it('@TEST-BARREL-EXPORT-002: should maintain backward compatibility', async () => {
      // 기존 import 경로: from './constants'
      // 새 import 경로: from './constants/index' 또는 './constants'
      const { GitNamingRules, GitCommitTemplates } = await import('../index');

      expect(GitNamingRules.FEATURE_PREFIX).toBe('feature/');
      expect(GitCommitTemplates.FEATURE).toBe('✨ feat: {message}');
    });
  });

  describe('Re-export Integrity', () => {
    it('@TEST-BARREL-EXPORT-003: should export GitNamingRules', async () => {
      const { GitNamingRules } = await import('../index');
      expect(GitNamingRules.createFeatureBranch('test')).toBe('feature/test');
    });

    it('@TEST-BARREL-EXPORT-004: should export GitCommitTemplates', async () => {
      const { GitCommitTemplates } = await import('../index');
      expect(GitCommitTemplates.getEmoji('feat')).toBe('✨');
    });

    it('@TEST-BARREL-EXPORT-005: should export GitDefaults', async () => {
      const { GitDefaults } = await import('../index');
      expect(GitDefaults.DEFAULT_BRANCH).toBe('main');
    });

    it('@TEST-BARREL-EXPORT-006: should export GitignoreTemplates', async () => {
      const { GitignoreTemplates } = await import('../index');
      expect(GitignoreTemplates.MOAI).toContain('# MoAI-ADK');
    });

    it('@TEST-BARREL-EXPORT-007: should export GitHubDefaults', async () => {
      const { GitHubDefaults } = await import('../index');
      expect(GitHubDefaults.API_BASE_URL).toBe('https://api.github.com');
    });

    it('@TEST-BARREL-EXPORT-008: should export GitTimeouts', async () => {
      const { GitTimeouts } = await import('../index');
      expect(GitTimeouts.DEFAULT).toBe(60000);
    });
  });

  describe('Type Safety', () => {
    it('@TEST-BARREL-EXPORT-009: should maintain as const types through re-export', async () => {
      const { GitDefaults } = await import('../index');
      const branch: 'main' = GitDefaults.DEFAULT_BRANCH;
      expect(branch).toBe('main');
    });
  });
});
