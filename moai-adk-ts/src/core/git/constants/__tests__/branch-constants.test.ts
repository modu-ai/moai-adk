// @TEST:REFACTOR-004 연결: @SPEC:REFACTOR-004 ->  -> @CODE:REFACTOR-004
/**
 * @file Branch Constants Test Suite
 * @tags @TEST:REFACTOR-004 @TEST-BRANCH-CONSTANTS-001
 * @description GitNamingRules 분리 및 타입 안전성 검증
 */

import { describe, it, expect } from 'vitest';
import { GitNamingRules } from '../branch-constants';

describe('@TEST:REFACTOR-004 - GitNamingRules', () => {
  describe('Branch Prefix Constants', () => {
    it('@TEST-BRANCH-CONSTANTS-001: should have correct prefix values', () => {
      expect(GitNamingRules.FEATURE_PREFIX).toBe('feature/');
      expect(GitNamingRules.BUGFIX_PREFIX).toBe('bugfix/');
      expect(GitNamingRules.HOTFIX_PREFIX).toBe('hotfix/');
      expect(GitNamingRules.SPEC_PREFIX).toBe('spec/');
      expect(GitNamingRules.CHORE_PREFIX).toBe('chore/');
    });
  });

  describe('Branch Name Creation', () => {
    it('@TEST-BRANCH-CONSTANTS-002: should create feature branch name', () => {
      expect(GitNamingRules.createFeatureBranch('login')).toBe('feature/login');
      expect(GitNamingRules.createFeatureBranch('user-auth')).toBe('feature/user-auth');
    });

    it('@TEST-BRANCH-CONSTANTS-003: should create spec branch name', () => {
      expect(GitNamingRules.createSpecBranch('SPEC-001')).toBe('spec/SPEC-001');
      expect(GitNamingRules.createSpecBranch('SPEC-004')).toBe('spec/SPEC-004');
    });

    it('@TEST-BRANCH-CONSTANTS-004: should create bugfix branch name', () => {
      expect(GitNamingRules.createBugfixBranch('fix-login')).toBe('bugfix/fix-login');
    });

    it('@TEST-BRANCH-CONSTANTS-005: should create hotfix branch name', () => {
      expect(GitNamingRules.createHotfixBranch('critical-bug')).toBe('hotfix/critical-bug');
    });
  });

  describe('Branch Name Validation', () => {
    it('@TEST-BRANCH-CONSTANTS-006: should validate correct branch names', () => {
      expect(GitNamingRules.isValidBranchName('feature/login')).toBe(true);
      expect(GitNamingRules.isValidBranchName('bugfix/fix-123')).toBe(true);
      expect(GitNamingRules.isValidBranchName('main')).toBe(true);
      expect(GitNamingRules.isValidBranchName('develop')).toBe(true);
    });

    it('@TEST-BRANCH-CONSTANTS-007: should reject invalid branch names', () => {
      // 시작/끝 하이픈
      expect(GitNamingRules.isValidBranchName('-invalid')).toBe(false);
      expect(GitNamingRules.isValidBranchName('invalid-')).toBe(false);

      // 연속 슬래시
      expect(GitNamingRules.isValidBranchName('feature//login')).toBe(false);

      // 연속 점
      expect(GitNamingRules.isValidBranchName('feature..login')).toBe(false);

      // 특수문자
      expect(GitNamingRules.isValidBranchName('feature@login')).toBe(false);
      expect(GitNamingRules.isValidBranchName('feature#login')).toBe(false);
    });
  });

  describe('Type Safety', () => {
    it('@TEST-BRANCH-CONSTANTS-008: should maintain as const type', () => {
      // TypeScript 컴파일 시 타입 추론 검증
      const prefix: 'feature/' = GitNamingRules.FEATURE_PREFIX;
      expect(prefix).toBe('feature/');
    });
  });
});
