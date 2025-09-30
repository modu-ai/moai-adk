// @TEST:REFACTOR-004 연결: @SPEC:REFACTOR-004 ->  -> @CODE:REFACTOR-004
/**
 * @file Config Constants Test Suite
 * @tags @TEST:REFACTOR-004 @TEST-CONFIG-CONSTANTS-001
 * @description GitDefaults, GitignoreTemplates, GitHubDefaults, GitTimeouts 분리 및 검증
 */

import { describe, it, expect } from 'vitest';
import {
  GitDefaults,
  GitignoreTemplates,
  GitHubDefaults,
  GitTimeouts,
} from '../config-constants';

describe('@TEST:REFACTOR-004 - Git Configuration Constants', () => {
  describe('GitDefaults', () => {
    it('@TEST-CONFIG-CONSTANTS-001: should have correct default values', () => {
      expect(GitDefaults.DEFAULT_BRANCH).toBe('main');
      expect(GitDefaults.DEFAULT_REMOTE).toBe('origin');
      expect(GitDefaults.COMMIT_MESSAGE_MAX_LENGTH).toBe(72);
      expect(GitDefaults.DESCRIPTION_MAX_LENGTH).toBe(100);
    });

    it('@TEST-CONFIG-CONSTANTS-002: should have config object', () => {
      expect(GitDefaults.CONFIG).toHaveProperty('init.defaultBranch', 'main');
      expect(GitDefaults.CONFIG).toHaveProperty('core.ignorecase', 'false');
      expect(GitDefaults.CONFIG).toHaveProperty('pull.rebase', 'false');
      expect(GitDefaults.CONFIG).toHaveProperty('push.default', 'current');
    });

    it('@TEST-CONFIG-CONSTANTS-003: should have platform-specific autocrlf', () => {
      const autocrlf = GitDefaults.CONFIG['core.autocrlf'];
      expect(['true', 'input']).toContain(autocrlf);
    });

    it('@TEST-CONFIG-CONSTANTS-004: should have safe commands list', () => {
      expect(GitDefaults.SAFE_COMMANDS).toContain('status');
      expect(GitDefaults.SAFE_COMMANDS).toContain('log');
      expect(GitDefaults.SAFE_COMMANDS).toContain('diff');
      expect(GitDefaults.SAFE_COMMANDS).toContain('branch');
    });

    it('@TEST-CONFIG-CONSTANTS-005: should have dangerous commands list', () => {
      expect(GitDefaults.DANGEROUS_COMMANDS).toContain('reset --hard');
      expect(GitDefaults.DANGEROUS_COMMANDS).toContain('clean -fd');
      expect(GitDefaults.DANGEROUS_COMMANDS).toContain('push --force');
    });
  });

  describe('GitignoreTemplates', () => {
    it('@TEST-CONFIG-CONSTANTS-006: should have MOAI template', () => {
      expect(GitignoreTemplates.MOAI).toContain('# MoAI-ADK Generated .gitignore');
      expect(GitignoreTemplates.MOAI).toContain('.claude/logs/');
      expect(GitignoreTemplates.MOAI).toContain('.moai/logs/');
      expect(GitignoreTemplates.MOAI).toContain('node_modules/');
    });

    it('@TEST-CONFIG-CONSTANTS-007: should have NODE template', () => {
      expect(GitignoreTemplates.NODE).toContain('# Node.js');
      expect(GitignoreTemplates.NODE).toContain('node_modules/');
      expect(GitignoreTemplates.NODE).toContain('.env');
    });

    it('@TEST-CONFIG-CONSTANTS-008: should have PYTHON template', () => {
      expect(GitignoreTemplates.PYTHON).toContain('# Python');
      expect(GitignoreTemplates.PYTHON).toContain('__pycache__/');
      expect(GitignoreTemplates.PYTHON).toContain('.venv');
    });
  });

  describe('GitHubDefaults', () => {
    it('@TEST-CONFIG-CONSTANTS-009: should have correct GitHub values', () => {
      expect(GitHubDefaults.API_BASE_URL).toBe('https://api.github.com');
      expect(GitHubDefaults.DEFAULT_BRANCH).toBe('main');
    });

    it('@TEST-CONFIG-CONSTANTS-010: should have PR template', () => {
      expect(GitHubDefaults.PR_TEMPLATE).toContain('## Summary');
      expect(GitHubDefaults.PR_TEMPLATE).toContain('## Test Plan');
      expect(GitHubDefaults.PR_TEMPLATE).toContain('## Checklist');
      expect(GitHubDefaults.PR_TEMPLATE).toContain('MoAI-ADK');
    });

    it('@TEST-CONFIG-CONSTANTS-011: should have Issue template', () => {
      expect(GitHubDefaults.ISSUE_TEMPLATE).toContain('## Description');
      expect(GitHubDefaults.ISSUE_TEMPLATE).toContain('## Steps to Reproduce');
      expect(GitHubDefaults.ISSUE_TEMPLATE).toContain('## Environment');
    });

    it('@TEST-CONFIG-CONSTANTS-012: should have default labels', () => {
      expect(GitHubDefaults.DEFAULT_LABELS).toHaveLength(8);
      const bugLabel = GitHubDefaults.DEFAULT_LABELS.find((l) => l.name === 'bug');
      expect(bugLabel).toBeDefined();
      expect(bugLabel?.color).toBe('d73a4a');
    });
  });

  describe('GitTimeouts', () => {
    it('@TEST-CONFIG-CONSTANTS-013: should have correct timeout values', () => {
      expect(GitTimeouts.CLONE).toBe(300000); // 5분
      expect(GitTimeouts.FETCH).toBe(120000); // 2분
      expect(GitTimeouts.PUSH).toBe(180000); // 3분
      expect(GitTimeouts.COMMIT).toBe(30000); // 30초
      expect(GitTimeouts.STATUS).toBe(10000); // 10초
      expect(GitTimeouts.DEFAULT).toBe(60000); // 1분
    });

    it('@TEST-CONFIG-CONSTANTS-014: should have reasonable timeout ranges', () => {
      expect(GitTimeouts.STATUS).toBeLessThan(GitTimeouts.COMMIT);
      expect(GitTimeouts.COMMIT).toBeLessThan(GitTimeouts.DEFAULT);
      expect(GitTimeouts.DEFAULT).toBeLessThan(GitTimeouts.FETCH);
      expect(GitTimeouts.FETCH).toBeLessThan(GitTimeouts.CLONE);
    });
  });

  describe('Type Safety', () => {
    it('@TEST-CONFIG-CONSTANTS-015: should maintain as const type', () => {
      const branch: 'main' = GitDefaults.DEFAULT_BRANCH;
      expect(branch).toBe('main');
    });
  });
});
