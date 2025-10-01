// @TEST:REFACTOR-001 | Chain: @SPEC:REFACTOR-001 -> @SPEC:REFACTOR-001 -> @CODE:REFACTOR-001
// Related: @CODE:GIT-COMMIT-001, @CODE:GIT-COMMIT-001:API

/**
 * GitCommitManager Test Suite
 * SPEC-001: git-manager.ts 리팩토링 - Phase 2
 *
 * @fileoverview TDD 테스트 suite for GitCommitManager class
 */

import { describe, it, expect, beforeAll, afterAll, beforeEach } from 'vitest';
import { GitCommitManager } from '../../../src/core/git/git-commit-manager';
import { GitCommitTemplates } from '../../../src/core/git/constants';
import type { GitConfig, GitCommitResult } from '../../../src/types/git';
import * as fs from 'fs-extra';
import * as path from 'path';
import * as os from 'os';
import simpleGit from 'simple-git';

describe('GitCommitManager', () => {
  let commitManager: GitCommitManager;
  let testDir: string;
  let repoPath: string;
  let config: GitConfig;

  beforeAll(async () => {
    // 테스트용 임시 디렉토리 생성
    testDir = await fs.mkdtemp(path.join(os.tmpdir(), 'moai-commit-test-'));
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
      commitMessageTemplate: GitCommitTemplates.FEATURE,
    };

    // 각 테스트마다 새 저장소 경로
    repoPath = path.join(testDir, `repo-${Date.now()}`);
    await fs.ensureDir(repoPath);

    // Git 저장소 초기화
    const git = simpleGit(repoPath);
    await git.init();
    await git.addConfig('user.name', 'Test User');
    await git.addConfig('user.email', 'test@example.com');

    commitManager = new GitCommitManager(config, repoPath);
  });

  describe('Constructor', () => {
    it('should create instance with valid config', () => {
      expect(commitManager).toBeDefined();
      expect(commitManager).toBeInstanceOf(GitCommitManager);
    });

    it('should throw error with invalid config', () => {
      const invalidConfig = {
        mode: 'invalid',
        autoCommit: false,
        branchPrefix: '',
      } as any;

      expect(() => new GitCommitManager(invalidConfig, repoPath)).toThrow();
    });
  });

  describe('commitChanges', () => {
    it('should commit changes with proper message formatting', async () => {
      // 테스트 파일 생성
      const testFile = path.join(repoPath, 'test.txt');
      await fs.writeFile(testFile, 'Initial content');

      const commitMessage = 'Add initial test file';
      const result: GitCommitResult = await commitManager.commitChanges(commitMessage);

      expect(result.hash).toMatch(/^[a-f0-9]+$/);
      expect(result.message).toContain(commitMessage);
      expect(result.timestamp).toBeInstanceOf(Date);
      expect(result.filesChanged).toBeGreaterThan(0);
      expect(result.author).toBeTruthy();
    });

    it('should stage specific files before committing', async () => {
      const file1 = path.join(repoPath, 'file1.txt');
      const file2 = path.join(repoPath, 'file2.txt');

      await fs.writeFile(file1, 'Content 1');
      await fs.writeFile(file2, 'Content 2');

      const result = await commitManager.commitChanges('Add specific files', ['file1.txt']);

      expect(result.filesChanged).toBe(1);
    });

    it('should handle empty repository with initial commit', async () => {
      // 빈 저장소에 커밋
      const result = await commitManager.commitChanges('Initial commit');

      expect(result.hash).toMatch(/^[a-f0-9]+$/);
      expect(result.message).toContain('Initial commit');
    });

    it('should throw error when file does not exist', async () => {
      await expect(
        commitManager.commitChanges('Add non-existent file', ['non-existent.txt'])
      ).rejects.toThrow();
    });
  });

  describe('createCheckpoint', () => {
    it('should create checkpoint commits with special formatting', async () => {
      // 테스트 파일 생성
      const testFile = path.join(repoPath, 'checkpoint-test.txt');
      await fs.writeFile(testFile, 'Content for checkpoint');

      const checkpointMessage = 'Before major refactoring';
      const result = await commitManager.createCheckpoint(checkpointMessage);

      expect(result).toMatch(/^[a-f0-9]+$/);
    });

    it('should handle checkpoint on empty repository', async () => {
      const result = await commitManager.createCheckpoint('Initial checkpoint');
      expect(result).toMatch(/^[a-f0-9]+$/);
    });
  });

  describe('getStatus', () => {
    it('should return accurate repository status', async () => {
      // 파일 생성 및 수정
      const newFile = path.join(repoPath, 'new.txt');
      const modifiedFile = path.join(repoPath, 'modified.txt');

      await fs.writeFile(newFile, 'New file content');
      await fs.writeFile(modifiedFile, 'Initial content');
      await commitManager.commitChanges('Add modified file', ['modified.txt']);
      await fs.writeFile(modifiedFile, 'Modified content');

      const status = await commitManager.getStatus();

      expect(status.clean).toBe(false);
      expect(status.untracked).toContain('new.txt');
      expect(status.modified).toContain('modified.txt');
      expect(status.currentBranch).toBeTruthy();
    });

    it('should detect clean working tree', async () => {
      // 초기 커밋 후 상태 확인
      await commitManager.commitChanges('Initial commit');
      const status = await commitManager.getStatus();

      expect(status.clean).toBe(true);
      expect(status.modified).toHaveLength(0);
      expect(status.untracked).toHaveLength(0);
    });
  });

  describe('commitWithLock', () => {
    it('should commit with lock successfully', async () => {
      // 테스트 파일 생성
      const testFile = path.join(repoPath, 'lock-test.txt');
      await fs.writeFile(testFile, 'Lock test content');

      const result = await commitManager.commitWithLock(
        'Test commit with lock',
        ['lock-test.txt']
      );

      expect(result.hash).toBeDefined();
      expect(result.message).toContain('Test commit with lock');
      expect(result.filesChanged).toBeGreaterThan(0);
    }, 10000);

    it('should handle concurrent commit attempts safely', async () => {
      // 첫 번째 커밋 준비
      const file1 = path.join(repoPath, 'concurrent1.txt');
      const file2 = path.join(repoPath, 'concurrent2.txt');

      await fs.writeFile(file1, 'Content 1');
      await fs.writeFile(file2, 'Content 2');

      // 동시 커밋 시도
      const commit1Promise = commitManager.commitWithLock('Concurrent commit 1', ['concurrent1.txt']);
      await new Promise(resolve => setTimeout(resolve, 50));
      const commit2Promise = commitManager.commitWithLock('Concurrent commit 2', ['concurrent2.txt']);

      const [result1, result2] = await Promise.all([commit1Promise, commit2Promise]);

      expect(result1.hash).toBeDefined();
      expect(result2.hash).toBeDefined();
      expect(result1.hash).not.toBe(result2.hash);
    }, 15000);
  });

  describe('Commit Message Templates', () => {
    it('should apply commit message templates correctly', async () => {
      const testFile = path.join(repoPath, 'template-test.txt');
      await fs.writeFile(testFile, 'Template test');

      const message = 'implement new feature';
      const result = await commitManager.commitChanges(message);

      // GitCommitTemplates.FEATURE 템플릿이 적용되었는지 확인
      expect(result.message).toContain('feat:');
    });

    it('should use different templates based on config', async () => {
      const bugfixConfig: GitConfig = {
        ...config,
        commitMessageTemplate: GitCommitTemplates.BUGFIX,
      };

      const bugfixManager = new GitCommitManager(bugfixConfig, repoPath);

      const testFile = path.join(repoPath, 'bugfix-test.txt');
      await fs.writeFile(testFile, 'Bugfix test');

      const result = await bugfixManager.commitChanges('fix critical bug');
      expect(result.message).toContain('fix:');
    });
  });

  describe('getLockStatus', () => {
    it('should return correct lock status', async () => {
      const lockStatus = await commitManager.getLockStatus();

      expect(lockStatus).toBeDefined();
      expect(typeof lockStatus.isLocked).toBe('boolean');
      expect(typeof lockStatus.lockFileExists).toBe('boolean');
    });
  });

  describe('cleanupStaleLocks', () => {
    it('should cleanup stale locks without error', async () => {
      await expect(commitManager.cleanupStaleLocks()).resolves.toBeUndefined();
    });
  });
});
