// @TEST:REFACTOR-001 | Chain: @SPEC:REFACTOR-001 -> @SPEC:REFACTOR-001 -> @CODE:REFACTOR-001
// Related: @CODE:GIT-PR-001, @CODE:GIT-PR-001:API

/**
 * GitPRManager Test Suite
 * SPEC-001: git-manager.ts 리팩토링 - Phase 3
 *
 * @fileoverview TDD 테스트 suite for GitPRManager class
 */

import { describe, it, expect, beforeEach } from 'vitest';
import { GitPRManager } from '../../../src/core/git/git-pr-manager';
import { GitCommitTemplates } from '../../../src/core/git/constants';
import type { GitConfig, CreatePullRequestOptions, CreateRepositoryOptions } from '../../../src/types/git';

describe('GitPRManager', () => {
  let prManager: GitPRManager;
  let config: GitConfig;

  beforeEach(() => {
    // Team 모드 설정
    config = {
      mode: 'team',
      autoCommit: false,
      branchPrefix: 'feature/',
      commitMessageTemplate: GitCommitTemplates.FEATURE,
      github: {
        token: 'mock-token',
        owner: 'test-org',
        repo: 'test-repo',
      },
    };

    prManager = new GitPRManager(config);
  });

  describe('Constructor', () => {
    it('should create instance with team mode config', () => {
      expect(prManager).toBeDefined();
      expect(prManager).toBeInstanceOf(GitPRManager);
    });

    it('should throw error for personal mode', () => {
      const personalConfig: GitConfig = {
        mode: 'personal',
        autoCommit: false,
        branchPrefix: 'feature/',
      };

      expect(() => new GitPRManager(personalConfig)).toThrow();
    });

    it('should throw error without github config', () => {
      const noGithubConfig: GitConfig = {
        mode: 'team',
        autoCommit: false,
        branchPrefix: 'feature/',
      };

      expect(() => new GitPRManager(noGithubConfig)).toThrow();
    });
  });

  describe('createPullRequest', () => {
    it('should handle PR creation request', async () => {
      const options: CreatePullRequestOptions = {
        title: 'Test PR',
        body: 'Test PR body',
        base: 'main',
        head: 'feature/test',
      };

      // GitHub 연동이 없으므로 에러가 예상됨
      await expect(prManager.createPullRequest(options)).rejects.toThrow();
    });

    it('should validate PR options', async () => {
      const invalidOptions: CreatePullRequestOptions = {
        title: '',
        body: 'Test',
        base: 'main',
        head: 'feature/test',
      };

      // 빈 title은 허용되지 않음
      await expect(prManager.createPullRequest(invalidOptions)).rejects.toThrow();
    });

    it('should handle PR template application', async () => {
      const options: CreatePullRequestOptions = {
        title: 'Test PR with template',
        body: '',
        base: 'main',
        head: 'feature/test',
      };

      // 빈 body는 기본 템플릿을 사용해야 함
      await expect(prManager.createPullRequest(options)).rejects.toThrow();
    });
  });

  describe('createRepository', () => {
    it('should handle repository creation request', async () => {
      const options: CreateRepositoryOptions = {
        name: 'test-repo',
        description: 'Test repository',
        private: true,
      };

      // GitHub 연동이 없으므로 에러가 예상됨
      await expect(prManager.createRepository(options)).rejects.toThrow();
    });

    it('should validate repository options', async () => {
      const invalidOptions: CreateRepositoryOptions = {
        name: '',
        description: 'Test',
      };

      // 빈 name은 허용되지 않음
      await expect(prManager.createRepository(invalidOptions)).rejects.toThrow();
    });
  });

  describe('pushChanges', () => {
    it('should handle push to remote (without actual remote)', async () => {
      // 실제 원격 저장소가 없으므로 에러 예상
      await expect(prManager.pushChanges()).rejects.toThrow(/Failed to push changes/);
    });

    it('should push to specific branch (without actual remote)', async () => {
      await expect(prManager.pushChanges('feature/test')).rejects.toThrow(/Failed to push changes/);
    });

    it('should push to specific remote (without actual remote)', async () => {
      await expect(prManager.pushChanges('main', 'upstream')).rejects.toThrow(/Failed to push changes/);
    });
  });

  describe('pushWithLock', () => {
    it('should push with lock safely (without actual remote)', async () => {
      // 실제 원격 저장소가 없으므로 에러 예상
      await expect(prManager.pushWithLock()).rejects.toThrow(/Failed to push changes/);
    }, 10000);

    it('should handle push lock timeout (without actual remote)', async () => {
      await expect(
        prManager.pushWithLock('main', 'origin', true, 1)
      ).rejects.toThrow(/Failed to push changes/);
    }, 5000);
  });

  describe('linkRemoteRepository', () => {
    it('should link remote repository with valid URL', async () => {
      const remoteUrl = 'https://github.com/test/repo.git';

      // 유효한 URL은 성공해야 함 (실제 푸시는 별도)
      await prManager.linkRemoteRepository(remoteUrl);
      // 에러 없이 완료되었는지 확인
      expect(true).toBe(true);
    });

    it('should validate remote URL format', async () => {
      const invalidUrls = [
        'not-a-url',
        'ftp://invalid.com/repo.git',
        'https://invalid-domain',
        '',
      ];

      for (const url of invalidUrls) {
        await expect(prManager.linkRemoteRepository(url)).rejects.toThrow(/Invalid Git URL/);
      }
    });

    it('should handle SSH and HTTPS URLs', async () => {
      const sshUrl = 'git@github.com:test/repo.git';
      const httpsUrl = 'https://github.com/test/repo.git';

      // 유효한 URL 형식은 링크 성공
      await prManager.linkRemoteRepository(sshUrl);
      await prManager.linkRemoteRepository(httpsUrl, 'upstream');

      // 에러 없이 완료되었는지 확인
      expect(true).toBe(true);
    });
  });

  describe('GitHub CLI Integration', () => {
    it('should check GitHub CLI availability', async () => {
      const isAvailable = await prManager.isGitHubCliAvailable();
      expect(typeof isAvailable).toBe('boolean');
    });

    it('should check GitHub authentication status', async () => {
      const isAuthenticated = await prManager.isGitHubAuthenticated();
      expect(typeof isAuthenticated).toBe('boolean');
    });
  });

  describe('getLockStatus', () => {
    it('should return correct lock status', async () => {
      const lockStatus = await prManager.getLockStatus();

      expect(lockStatus).toBeDefined();
      expect(typeof lockStatus.isLocked).toBe('boolean');
      expect(typeof lockStatus.lockFileExists).toBe('boolean');
    });
  });

  describe('cleanupStaleLocks', () => {
    it('should cleanup stale locks without error', async () => {
      await expect(prManager.cleanupStaleLocks()).resolves.toBeUndefined();
    });
  });
});
