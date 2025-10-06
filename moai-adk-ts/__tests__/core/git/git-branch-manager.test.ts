// @TEST:REFACTOR-001:BRANCH | SPEC: SPEC-REFACTOR-001.md | CODE: src/core/git/git-branch-manager.ts
// Related: @CODE:REFACTOR-001:BRANCH, @CODE:GIT-MGR-001

/**
 * GitBranchManager Test Suite
 * SPEC-001: git-manager.ts 리팩토링 - Phase 1
 *
 * @fileoverview TDD 테스트 suite for GitBranchManager class
 */

import { describe, it, expect, beforeAll, afterAll, beforeEach } from 'vitest';
import { GitBranchManager } from '../../../src/core/git/git-branch-manager';
import { GitNamingRules } from '../../../src/core/git/constants';
import type { GitConfig } from '../../../src/types/git';
import * as fs from 'fs-extra';
import * as path from 'path';
import * as os from 'os';

describe('GitBranchManager', () => {
  let branchManager: GitBranchManager;
  let testDir: string;
  let repoPath: string;
  let config: GitConfig;

  beforeAll(async () => {
    // 테스트용 임시 디렉토리 생성
    testDir = await fs.mkdtemp(path.join(os.tmpdir(), 'moai-branch-test-'));
  });

  afterAll(async () => {
    // 테스트 디렉토리 정리
    if (testDir && await fs.pathExists(testDir)) {
      await fs.remove(testDir);
    }
  });

  beforeEach(async () => {
    // 기본 Git 설정
    config = {
      mode: 'personal',
      autoCommit: false,
      branchPrefix: 'feature/',
    };

    // 각 테스트마다 새 저장소 경로
    repoPath = path.join(testDir, `repo-${Date.now()}`);
    await fs.ensureDir(repoPath);

    branchManager = new GitBranchManager(config, repoPath);
  });

  describe('Constructor', () => {
    it('should create instance with valid config', () => {
      expect(branchManager).toBeDefined();
      expect(branchManager).toBeInstanceOf(GitBranchManager);
    });

    it('should throw error with invalid config', () => {
      const invalidConfig = {
        mode: 'invalid',
        autoCommit: false,
        branchPrefix: '',
      } as any;

      expect(() => new GitBranchManager(invalidConfig, repoPath)).toThrow();
    });
  });

  describe('createBranch', () => {
    it('should create a new branch successfully', async () => {
      // 먼저 저장소 초기화 필요
      await branchManager.initializeRepository();

      const branchName = 'feature/test-branch';
      await branchManager.createBranch(branchName);

      const currentBranch = await branchManager.getCurrentBranch();
      expect(currentBranch).toBe(branchName);
    });

    it('should validate branch names', async () => {
      await branchManager.initializeRepository();

      const invalidBranchNames = [
        '-invalid-start',
        'invalid-end-',
        'invalid//double-slash',
        'invalid..double-dot',
      ];

      for (const invalidName of invalidBranchNames) {
        await expect(branchManager.createBranch(invalidName)).rejects.toThrow();
      }
    });

    it('should create branch from specific base branch', async () => {
      await branchManager.initializeRepository();

      const baseBranch = 'main';
      const newBranch = 'feature/from-main';

      await branchManager.createBranch(newBranch, baseBranch);
      const currentBranch = await branchManager.getCurrentBranch();
      expect(currentBranch).toBe(newBranch);
    });

    it('should handle missing base branch error', async () => {
      await branchManager.initializeRepository();

      const nonExistentBase = 'feature/non-existent';
      const newBranch = 'feature/test';

      await expect(
        branchManager.createBranch(newBranch, nonExistentBase)
      ).rejects.toThrow();
    });
  });

  describe('getCurrentBranch', () => {
    it('should return current branch name', async () => {
      await branchManager.initializeRepository();

      // 초기 커밋 생성 (브랜치 이름이 설정되도록)
      const testFile = path.join(repoPath, 'test.txt');
      await fs.writeFile(testFile, 'test');
      await branchManager.git.add('.');
      await branchManager.git.commit('Initial commit');

      const currentBranch = await branchManager.getCurrentBranch();
      expect(currentBranch).toBe('main');
    });

    it('should return default branch if repository not initialized', async () => {
      const uninitDir = path.join(testDir, 'uninit');
      const uninitializedManager = new GitBranchManager(config, uninitDir);
      const currentBranch = await uninitializedManager.getCurrentBranch();
      expect(currentBranch).toBe('main'); // fallback to default
    });
  });

  describe('listBranches', () => {
    it('should list all branches', async () => {
      await branchManager.initializeRepository();

      // 추가 브랜치 생성
      await branchManager.createBranch('feature/branch1');
      await branchManager.createBranch('feature/branch2');

      const branches = await branchManager.listBranches();
      expect(branches).toContain('main');
      expect(branches).toContain('feature/branch1');
      expect(branches).toContain('feature/branch2');
    });

    it('should return empty array for uninitialized repository', async () => {
      const uninitDir2 = path.join(testDir, 'uninit2');
      const uninitializedManager = new GitBranchManager(config, uninitDir2);
      const branches = await uninitializedManager.listBranches();
      expect(branches).toEqual([]);
    });
  });

  describe('switchBranch', () => {
    it('should switch to existing branch', async () => {
      await branchManager.initializeRepository();

      const newBranch = 'feature/switch-test';
      await branchManager.createBranch(newBranch);

      // main으로 돌아가기
      await branchManager.switchBranch('main');
      const currentBranch = await branchManager.getCurrentBranch();
      expect(currentBranch).toBe('main');
    });

    it('should throw error when switching to non-existent branch', async () => {
      await branchManager.initializeRepository();

      await expect(
        branchManager.switchBranch('feature/non-existent')
      ).rejects.toThrow();
    });
  });

  describe('initializeRepository', () => {
    it('should initialize Git repository', async () => {
      const result = await branchManager.initializeRepository();

      expect(result.success).toBe(true);
      expect(result.defaultBranch).toBe('main');
      expect(await fs.pathExists(path.join(repoPath, '.git'))).toBe(true);
    });

    it('should not reinitialize existing repository', async () => {
      const firstResult = await branchManager.initializeRepository();
      expect(firstResult.success).toBe(true);

      const secondResult = await branchManager.initializeRepository();
      expect(secondResult.success).toBe(true);
      expect(secondResult.message).toContain('already initialized');
    });
  });

  describe('Branch Naming Rules Integration', () => {
    it('should use GitNamingRules for branch creation', () => {
      expect(GitNamingRules.createFeatureBranch('awesome-feature')).toBe('feature/awesome-feature');
      expect(GitNamingRules.createSpecBranch('SPEC-012')).toBe('spec/SPEC-012');
      expect(GitNamingRules.createBugfixBranch('critical-bug')).toBe('bugfix/critical-bug');
      expect(GitNamingRules.createHotfixBranch('security-fix')).toBe('hotfix/security-fix');
    });

    it('should validate branch names according to Git rules', () => {
      expect(GitNamingRules.isValidBranchName('feature/valid-branch')).toBe(true);
      expect(GitNamingRules.isValidBranchName('-invalid-start')).toBe(false);
      expect(GitNamingRules.isValidBranchName('invalid//double')).toBe(false);
    });
  });

  describe('createBranchWithLock', () => {
    it('should create branch safely with lock', async () => {
      await branchManager.initializeRepository();

      const branchName = 'feature/lock-test';
      await branchManager.createBranchWithLock(branchName);

      // 초기 커밋 생성 후 브랜치 확인
      const testFile = path.join(repoPath, 'lock-test.txt');
      await fs.writeFile(testFile, 'lock test');
      await branchManager.git.add('.');
      await branchManager.git.commit('Lock test commit');

      const currentBranch = await branchManager.getCurrentBranch();
      expect(currentBranch).toBe(branchName);
    }, 10000);

    it('should handle lock timeout with no-wait mode', async () => {
      await branchManager.initializeRepository();

      // wait=false로 설정하여 lock이 이미 있으면 즉시 실패하도록
      const branchName = 'feature/no-wait-test';
      await branchManager.createBranchWithLock(branchName, undefined, false, 1);

      // 초기 커밋 후 브랜치 확인
      const testFile = path.join(repoPath, 'no-wait.txt');
      await fs.writeFile(testFile, 'no wait test');
      await branchManager.git.add('.');
      await branchManager.git.commit('No-wait test commit');

      const currentBranch = await branchManager.getCurrentBranch();
      expect(currentBranch).toBe(branchName);
    }, 5000);
  });
});
