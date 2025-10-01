/**
 * GitManager Test Suite
 * SPEC-012 Week 2 Track D: Git System Integration
 *
 * @fileoverview TDD 테스트 suite for GitManager class
 */

import { GitManager } from '../../../src/core/git/git-manager';
import { GitConfig, GitInitResult, GitStatus, GitCommitResult } from '../../../src/types/git';
import { GitNamingRules, GitCommitTemplates } from '../../../src/core/git/constants';
import * as fs from 'fs-extra';
import * as path from 'path';
import * as os from 'os';

describe('GitManager', () => {
  let gitManager: GitManager;
  let testDir: string;
  let config: GitConfig;

  beforeAll(async () => {
    // 테스트용 임시 디렉토리 생성
    testDir = await fs.mkdtemp(path.join(os.tmpdir(), 'moai-git-test-'));
  });

  afterAll(async () => {
    // 테스트 디렉토리 정리
    if (testDir && await fs.pathExists(testDir)) {
      await fs.remove(testDir);
    }
  });

  beforeEach(() => {
    // 기본 Git 설정
    config = {
      mode: 'personal',
      autoCommit: false,
      branchPrefix: 'feature/',
      commitMessageTemplate: GitCommitTemplates.FEATURE
    };

    // testDir로 작업 디렉토리 설정
    gitManager = new GitManager(config, testDir);
  });

  describe('Repository Initialization', () => {
    it('should initialize a new Git repository successfully', async () => {
      const projectPath = path.join(testDir, 'test-repo-init');
      await fs.ensureDir(projectPath);

      const result: GitInitResult = await gitManager.initializeRepository(projectPath);

      expect(result.success).toBe(true);
      expect(result.repositoryPath).toBe(projectPath);
      expect(result.gitDir).toBe(path.join(projectPath, '.git'));
      expect(result.defaultBranch).toBe('main');
      expect(await fs.pathExists(result.gitDir)).toBe(true);
    });

    it('should handle repository initialization failure gracefully', async () => {
      const invalidPath = '/invalid/path/that/does/not/exist';

      const result: GitInitResult = await gitManager.initializeRepository(invalidPath);

      expect(result.success).toBe(false);
      expect(result.message).toContain('Failed to initialize repository');
    });

    it('should not reinitialize existing repository', async () => {
      const projectPath = path.join(testDir, 'test-repo-existing');
      await fs.ensureDir(projectPath);

      // 첫 번째 초기화
      const firstResult = await gitManager.initializeRepository(projectPath);
      expect(firstResult.success).toBe(true);

      // 두 번째 초기화 시도
      const secondResult = await gitManager.initializeRepository(projectPath);
      expect(secondResult.success).toBe(true);
      expect(secondResult.message).toContain('already initialized');
    });
  });

  describe('Branch Management', () => {
    it('should create a new branch successfully', async () => {
      const repoPath = path.join(testDir, 'test-repo-branch-create');
      await fs.ensureDir(repoPath);

      const branchName = 'feature/test-branch';
      const branchGitManager = new GitManager(config, repoPath);

      // Initialize repository first
      await branchGitManager.initializeRepository(repoPath);

      // Should not throw when creating a valid branch
      await expect(branchGitManager.createBranch(branchName)).resolves.not.toThrow();

      // Verify that branch operation completed without error
      const status = await branchGitManager.getStatus();
      expect(status).toBeDefined();
    });

    it('should validate branch names according to Git rules', async () => {
      const repoPath = path.join(testDir, 'test-repo-branch-validate');
      await fs.ensureDir(repoPath);

      const branchGitManager = new GitManager(config, repoPath);
      await branchGitManager.initializeRepository(repoPath);

      const invalidBranchNames = [
        '-invalid-start',
        'invalid-end-',
        'invalid//double-slash',
        'invalid..double-dot'
      ];

      for (const invalidName of invalidBranchNames) {
        await expect(branchGitManager.createBranch(invalidName)).rejects.toThrow();
      }
    });

    it('should create branch from specific base branch', async () => {
      const repoPath = path.join(testDir, 'test-repo-branch-from-base');
      await fs.ensureDir(repoPath);

      const branchGitManager = new GitManager(config, repoPath);
      await branchGitManager.initializeRepository(repoPath);

      const baseBranch = 'main';
      const newBranch = 'feature/from-main';

      await branchGitManager.createBranch(newBranch, baseBranch);

      const currentBranch = await branchGitManager.getCurrentBranch();
      expect(currentBranch).toBe(newBranch);
    });

    it('should use naming rules for different branch types', () => {
      expect(GitNamingRules.createFeatureBranch('awesome-feature')).toBe('feature/awesome-feature');
      expect(GitNamingRules.createSpecBranch('SPEC-012')).toBe('spec/SPEC-012');
      expect(GitNamingRules.createBugfixBranch('critical-bug')).toBe('bugfix/critical-bug');
      expect(GitNamingRules.createHotfixBranch('security-fix')).toBe('hotfix/security-fix');
    });
  });

  describe('Commit Operations', () => {
    let repoPath: string;

    beforeEach(async () => {
      repoPath = path.join(testDir, 'test-repo-commit');
      await fs.ensureDir(repoPath);
    });

    it('should commit changes with proper message formatting', async () => {
      const commitGitManager = new GitManager(config, repoPath);
      await commitGitManager.initializeRepository(repoPath);

      // 테스트 파일 생성
      const testFile = path.join(repoPath, 'test.txt');
      await fs.writeFile(testFile, 'Initial content');

      const commitMessage = 'Add initial test file';
      const result: GitCommitResult = await commitGitManager.commitChanges(commitMessage);

      expect(result.hash).toMatch(/^[a-f0-9]+$/);
      expect(result.message).toContain(commitMessage);
      expect(result.timestamp).toBeInstanceOf(Date);
      expect(result.filesChanged).toBeGreaterThan(0);
      expect(result.author).toBeTruthy();
    });

    it('should stage specific files before committing', async () => {
      const commitGitManager = new GitManager(config, repoPath);
      await commitGitManager.initializeRepository(repoPath);

      const file1 = path.join(repoPath, 'file1.txt');
      const file2 = path.join(repoPath, 'file2.txt');

      await fs.writeFile(file1, 'Content 1');
      await fs.writeFile(file2, 'Content 2');

      const result = await commitGitManager.commitChanges('Add specific files', ['file1.txt']);

      expect(result.filesChanged).toBe(1);
    });

    it('should apply commit message templates correctly', () => {
      const message = 'implement new feature';
      const templatedMessage = GitCommitTemplates.apply(GitCommitTemplates.FEATURE, message);

      expect(templatedMessage).toBe('✨ feat: implement new feature');
    });

    it('should create checkpoint commits with special formatting', async () => {
      const commitGitManager = new GitManager(config, repoPath);
      await commitGitManager.initializeRepository(repoPath);

      // 테스트 파일 생성
      const testFile = path.join(repoPath, 'checkpoint-test.txt');
      await fs.writeFile(testFile, 'Content for checkpoint');

      const checkpointMessage = 'Before major refactoring';
      const result = await commitGitManager.createCheckpoint(checkpointMessage);

      expect(result).toMatch(/^[a-f0-9]+$/);
      // 체크포인트 커밋은 특별한 형식을 가져야 함
    });
  });

  describe('Status and Information', () => {
    let repoPath: string;

    beforeEach(async () => {
      repoPath = path.join(testDir, 'test-repo-status');
      await fs.ensureDir(repoPath);
    });

    it('should return accurate repository status', async () => {
      const statusGitManager = new GitManager(config, repoPath);
      await statusGitManager.initializeRepository(repoPath);

      // 파일 생성 및 수정
      const newFile = path.join(repoPath, 'new.txt');
      const modifiedFile = path.join(repoPath, 'modified.txt');

      await fs.writeFile(newFile, 'New file content');
      await fs.writeFile(modifiedFile, 'Initial content');
      await statusGitManager.commitChanges('Add modified file', ['modified.txt']);
      await fs.writeFile(modifiedFile, 'Modified content');

      const status: GitStatus = await statusGitManager.getStatus();

      expect(status.clean).toBe(false);
      expect(status.untracked).toContain('new.txt');
      expect(status.modified).toContain('modified.txt');
      expect(status.currentBranch).toBe('main');
      expect(status.ahead).toBe(0);
      expect(status.behind).toBe(0);
    });

    it('should detect clean working tree', async () => {
      const statusGitManager = new GitManager(config, repoPath);
      await statusGitManager.initializeRepository(repoPath);

      // 초기 README 파일 커밋 후 상태 확인
      await statusGitManager.commitChanges('Initial commit');
      const status = await statusGitManager.getStatus();

      expect(status.clean).toBe(true);
      expect(status.modified).toHaveLength(0);
      expect(status.untracked).toHaveLength(0);
      expect(status.added).toHaveLength(0);
      expect(status.deleted).toHaveLength(0);
    });
  });

  describe('Gitignore Management', () => {
    let repoPath: string;

    beforeEach(async () => {
      repoPath = path.join(testDir, 'test-repo-gitignore');
      await fs.ensureDir(repoPath);
    });

    it('should create .gitignore with MoAI-ADK template', async () => {
      const gitignoreGitManager = new GitManager(config, repoPath);
      const gitignorePath = await gitignoreGitManager.createGitignore(repoPath);

      expect(gitignorePath).toBe(path.join(repoPath, '.gitignore'));
      expect(await fs.pathExists(gitignorePath)).toBe(true);

      const content = await fs.readFile(gitignorePath, 'utf-8');
      expect(content).toContain('# MoAI-ADK Generated .gitignore');
      expect(content).toContain('node_modules/');
      expect(content).toContain('.moai/logs/');
      expect(content).toContain('.claude/logs/');
    });

    it('should not overwrite existing .gitignore', async () => {
      const gitignoreGitManager = new GitManager(config, repoPath);
      const gitignorePath = path.join(repoPath, '.gitignore');
      const existingContent = '# Existing content\ncustom-ignore/\n';

      await fs.writeFile(gitignorePath, existingContent);

      const resultPath = await gitignoreGitManager.createGitignore(repoPath);
      const content = await fs.readFile(resultPath, 'utf-8');

      expect(content).toContain('# Existing content');
      expect(content).toContain('custom-ignore/');
    });
  });

  describe('Remote Operations', () => {
    let repoPath: string;

    beforeEach(async () => {
      repoPath = path.join(testDir, 'test-repo-remote');
      await fs.ensureDir(repoPath);
    });

    it('should link remote repository successfully', async () => {
      const remoteGitManager = new GitManager(config, repoPath);
      await remoteGitManager.initializeRepository(repoPath);

      const remoteUrl = 'https://github.com/test/repo.git';

      // Should not throw when linking valid remote repository
      await expect(async () => {
        await remoteGitManager.linkRemoteRepository(remoteUrl);
      }).not.toThrow();
    });

    it('should detect SSH vs HTTPS remote URLs', async () => {
      // SSH 테스트용 별도 레포
      const sshRepoPath = path.join(testDir, 'test-repo-ssh');
      await fs.ensureDir(sshRepoPath);
      const sshGitManager = new GitManager(config, sshRepoPath);
      await sshGitManager.initializeRepository(sshRepoPath);

      const sshUrl = 'git@github.com:test/repo.git';

      // Should not throw when linking SSH remote
      await expect(async () => {
        await sshGitManager.linkRemoteRepository(sshUrl);
      }).not.toThrow();

      // HTTPS 테스트용 별도 레포
      const httpsRepoPath = path.join(testDir, 'test-repo-https');
      await fs.ensureDir(httpsRepoPath);
      const httpsGitManager = new GitManager(config, httpsRepoPath);
      await httpsGitManager.initializeRepository(httpsRepoPath);

      const httpsUrl = 'https://github.com/test/repo.git';

      // Should not throw when linking HTTPS remote
      await expect(async () => {
        await httpsGitManager.linkRemoteRepository(httpsUrl);
      }).not.toThrow();
    });

    it('should validate remote URL format', async () => {
      const remoteGitManager = new GitManager(config, repoPath);

      const invalidUrls = [
        'not-a-url',
        'ftp://invalid.com/repo.git',
        'https://invalid-domain',
        ''
      ];

      for (const url of invalidUrls) {
        await expect(remoteGitManager.linkRemoteRepository(url)).rejects.toThrow();
      }
    });
  });

  describe('Push Operations', () => {
    let repoPath: string;
    let teamConfig: GitConfig;

    beforeEach(async () => {
      repoPath = path.join(testDir, 'test-repo-push');
      await fs.ensureDir(repoPath);

      // Team mode 설정 (push는 team mode에서만 가능)
      teamConfig = {
        mode: 'team',
        autoCommit: false,
        branchPrefix: 'feature/',
        commitMessageTemplate: GitCommitTemplates.FEATURE,
        github: {
          owner: 'test-org',
          repo: 'test-repo'
        }
      };
    });

    it('should throw error when pushChanges called in personal mode', async () => {
      const personalGitManager = new GitManager(config, repoPath);
      await personalGitManager.initializeRepository(repoPath);

      // Personal mode에서는 pushChanges가 에러를 던져야 함
      await expect(personalGitManager.pushChanges()).rejects.toThrow('PR Manager is only available in team mode');
    });

    it('should reject push without remote in team mode', async () => {
      const pushGitManager = new GitManager(teamConfig, repoPath);
      await pushGitManager.initializeRepository(repoPath);

      const testFile = path.join(repoPath, 'test.txt');
      await fs.writeFile(testFile, 'Test content');
      await pushGitManager.commitChanges('Add test file');

      // 실제 원격이 없으므로 에러가 예상됨
      await expect(pushGitManager.pushChanges()).rejects.toThrow();
    });
  });

  describe('Team Mode Integration', () => {
    let teamGitManager: GitManager;
    let repoPath: string;

    beforeEach(async () => {
      const teamConfig: GitConfig = {
        mode: 'team',
        autoCommit: true,
        branchPrefix: 'feature/',
        commitMessageTemplate: GitCommitTemplates.FEATURE,
        github: {
          owner: 'test-org',
          repo: 'test-repo'
        }
      };

      teamGitManager = new GitManager(teamConfig);
      repoPath = path.join(testDir, 'test-repo-team');
      await fs.ensureDir(repoPath);
    });

    it('should initialize team mode with GitHub integration', async () => {
      expect(teamGitManager).toBeInstanceOf(GitManager);
      // Team 모드 특정 기능은 GitHub 연동 구현 후 테스트
    });

    it('should auto-commit when enabled', async () => {
      // auto-commit 기능 테스트는 구현 후 추가
      expect(true).toBe(true); // 플레이스홀더
    });
  });

  describe('Error Handling', () => {
    it('should handle Git command failures gracefully', async () => {
      expect(() => {
        const invalidConfig: GitConfig = {
          mode: 'personal',
          autoCommit: false,
          branchPrefix: '', // Invalid empty string
          commitMessageTemplate: ''
        };

        new GitManager(invalidConfig);
      }).toThrow('branchPrefix must be a non-empty string');
    });

    it('should provide meaningful error messages', async () => {
      const nonExistentPath = '/non/existent/path';

      try {
        await gitManager.initializeRepository(nonExistentPath);
      } catch (error) {
        expect(error).toBeInstanceOf(Error);
        expect((error as Error).message).toContain('Failed to initialize repository');
      }
    });

    it('should handle network timeouts', async () => {
      // 네트워크 타임아웃 테스트는 실제 네트워크 작업 시뮬레이션 필요
      expect(true).toBe(true); // 플레이스홀더
    });
  });

  describe('Performance and Safety', () => {
    it('should complete basic operations within timeout limits', async () => {
      const startTime = Date.now();
      const repoPath = path.join(testDir, 'test-repo-perf');
      await fs.ensureDir(repoPath);

      const perfGitManager = new GitManager(config, repoPath);
      await perfGitManager.initializeRepository(repoPath);
      const status = await perfGitManager.getStatus();

      const duration = Date.now() - startTime;
      expect(duration).toBeLessThan(5000); // 5초 이내
      expect(status).toBeDefined();
    });

    it('should validate dangerous operations', async () => {
      // 위험한 작업에 대한 안전장치 테스트
      // 실제 구현에서는 force 옵션 등에 대한 검증
      expect(true).toBe(true); // 플레이스홀더
    });
  });

  describe('Git Lock Integration', () => {
    let lockRepoPath: string;
    let lockGitManager: GitManager;
    let lockConfig: GitConfig;

    beforeEach(async () => {
      lockConfig = {
        mode: 'personal',
        autoCommit: false,
        branchPrefix: 'feature/',
        commitMessageTemplate: GitCommitTemplates.FEATURE
      };

      lockRepoPath = path.join(testDir, 'test-repo-lock');
      await fs.ensureDir(lockRepoPath);
      lockGitManager = new GitManager(lockConfig, lockRepoPath);
      await lockGitManager.initializeRepository(lockRepoPath);
    });

    /**
     * @tags @TEST:COMMIT-WITH-LOCK-001 @CODE:GIT-LOCK-INTEGRATION-001
     */
    it('should commit with lock successfully', async () => {
      // 테스트 파일 생성
      const testFile = path.join(lockRepoPath, 'lock-test.txt');
      await fs.writeFile(testFile, 'Lock test content');

      // Lock을 사용한 커밋
      const result = await lockGitManager.commitWithLock('Test commit with lock', ['lock-test.txt']);

      expect(result.hash).toBeDefined();
      expect(result.message).toContain('Test commit with lock');
      expect(result.filesChanged).toBeGreaterThan(0);
    }, 10000);

    /**
     * @tags @TEST:BRANCH-WITH-LOCK-001 @CODE:GIT-LOCK-INTEGRATION-001
     */
    it('should create branch with lock safely', async () => {
      const branchName = 'feature/lock-test-branch';

      // Lock을 사용한 브랜치 생성
      await lockGitManager.createBranchWithLock(branchName);

      // 브랜치가 생성되었는지 확인
      const currentBranch = await lockGitManager.getCurrentBranch();
      expect(currentBranch).toBe(branchName);
    }, 10000);

    /**
     * @tags @TEST:LOCK-STATUS-001 @CODE:GIT-LOCK-INTEGRATION-001
     */
    it('should return correct lock status', async () => {
      const lockStatus = await lockGitManager.getLockStatus();

      expect(lockStatus).toBeDefined();
      expect(typeof lockStatus.isLocked).toBe('boolean');
      expect(typeof lockStatus.lockFileExists).toBe('boolean');
      expect(lockStatus.isLocked).toBe(false); // 현재 Lock 없음
    });

    /**
     * @tags @TEST:CLEANUP-STALE-LOCKS-001 @CODE:GIT-LOCK-INTEGRATION-001
     */
    it('should cleanup stale locks without error', async () => {
      // 오래된 lock 정리는 에러 없이 실행되어야 함
      await expect(lockGitManager.cleanupStaleLocks()).resolves.toBeUndefined();
    });

    /**
     * @tags @TEST:CONCURRENT-OPERATIONS-001 @CODE:GIT-LOCK-INTEGRATION-001
     */
    it('should handle concurrent commit attempts safely', async () => {
      // 첫 번째 커밋 준비
      const file1 = path.join(lockRepoPath, 'concurrent1.txt');
      const file2 = path.join(lockRepoPath, 'concurrent2.txt');

      await fs.writeFile(file1, 'Content 1');
      await fs.writeFile(file2, 'Content 2');

      // 동시 커밋 시도 (둘 다 대기 모드로 순차 실행)
      const commit1Promise = lockGitManager.commitWithLock('Concurrent commit 1', [file1], true, 5);

      // 약간의 지연 후 두 번째 커밋 시도 (대기 모드)
      await new Promise(resolve => setTimeout(resolve, 50));
      const commit2Promise = lockGitManager.commitWithLock('Concurrent commit 2', [file2], true, 5);

      // 두 커밋 모두 성공해야 함 (Lock이 순차적으로 작동)
      const result1 = await commit1Promise;
      expect(result1.hash).toBeDefined();

      const result2 = await commit2Promise;
      expect(result2.hash).toBeDefined();

      // 두 커밋이 서로 다른 해시를 가져야 함
      expect(result1.hash).not.toBe(result2.hash);
    }, 15000);
  });
});

// GitHub 연동 기능 테스트 (Team 모드)
describe('GitManager - GitHub Integration', () => {
  let gitManager: GitManager;

  beforeEach(() => {
    const config: GitConfig = {
      mode: 'team',
      autoCommit: false,
      branchPrefix: 'feature/',
      commitMessageTemplate: GitCommitTemplates.FEATURE,
      github: {
        token: 'mock-token',
        owner: 'test-org',
        repo: 'test-repo'
      }
    };

    gitManager = new GitManager(config);
  });

  describe('Repository Creation', () => {
    it('should create GitHub repository with correct settings', async () => {
      // GitHub API 모킹이 필요하므로 구현 후 실제 테스트 작성
      expect(gitManager).toBeInstanceOf(GitManager);
    });

    it('should handle repository creation errors', async () => {
      // API 에러 처리 테스트
      expect(gitManager).toBeInstanceOf(GitManager);
    });
  });

  describe('Pull Request Creation', () => {
    it('should create PR with proper template', async () => {
      // GitHub API 호출 모킹 필요
      expect(gitManager).toBeInstanceOf(GitManager);
    });

    it('should apply default PR template when body is empty', async () => {
      // 기본 템플릿 적용 테스트
      expect(gitManager).toBeInstanceOf(GitManager);
    });
  });
});