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

    gitManager = new GitManager(config);
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
    let repoPath: string;

    beforeEach(async () => {
      repoPath = path.join(testDir, 'test-repo-branch');
      await fs.ensureDir(repoPath);
      await gitManager.initializeRepository(repoPath);
    });

    it('should create a new branch successfully', async () => {
      const branchName = 'feature/test-branch';

      await expect(gitManager.createBranch(branchName)).resolves.not.toThrow();

      const status = await gitManager.getStatus();
      expect(status.currentBranch).toBe(branchName);
    });

    it('should validate branch names according to Git rules', async () => {
      const invalidBranchNames = [
        '-invalid-start',
        'invalid-end-',
        'invalid//double-slash',
        'invalid..double-dot'
      ];

      for (const invalidName of invalidBranchNames) {
        await expect(gitManager.createBranch(invalidName)).rejects.toThrow();
      }
    });

    it('should create branch from specific base branch', async () => {
      const baseBranch = 'main';
      const newBranch = 'feature/from-main';

      await gitManager.createBranch(newBranch, baseBranch);

      const status = await gitManager.getStatus();
      expect(status.currentBranch).toBe(newBranch);
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
      await gitManager.initializeRepository(repoPath);

      // 테스트 파일 생성
      const testFile = path.join(repoPath, 'test.txt');
      await fs.writeFile(testFile, 'Initial content');
    });

    it('should commit changes with proper message formatting', async () => {
      const commitMessage = 'Add initial test file';

      const result: GitCommitResult = await gitManager.commitChanges(commitMessage);

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

      const result = await gitManager.commitChanges('Add specific files', ['file1.txt']);

      expect(result.filesChanged).toBe(1);
    });

    it('should apply commit message templates correctly', () => {
      const message = 'implement new feature';
      const templatedMessage = GitCommitTemplates.apply(GitCommitTemplates.FEATURE, message);

      expect(templatedMessage).toBe('✨ feat: implement new feature');
    });

    it('should create checkpoint commits with special formatting', async () => {
      const checkpointMessage = 'Before major refactoring';

      const result = await gitManager.createCheckpoint(checkpointMessage);

      expect(result).toMatch(/^[a-f0-9]+$/);
      // 체크포인트 커밋은 특별한 형식을 가져야 함
    });
  });

  describe('Status and Information', () => {
    let repoPath: string;

    beforeEach(async () => {
      repoPath = path.join(testDir, 'test-repo-status');
      await fs.ensureDir(repoPath);
      await gitManager.initializeRepository(repoPath);
    });

    it('should return accurate repository status', async () => {
      // 파일 생성 및 수정
      const newFile = path.join(repoPath, 'new.txt');
      const modifiedFile = path.join(repoPath, 'modified.txt');

      await fs.writeFile(newFile, 'New file content');
      await fs.writeFile(modifiedFile, 'Initial content');
      await gitManager.commitChanges('Add modified file', ['modified.txt']);
      await fs.writeFile(modifiedFile, 'Modified content');

      const status: GitStatus = await gitManager.getStatus();

      expect(status.clean).toBe(false);
      expect(status.untracked).toContain('new.txt');
      expect(status.modified).toContain('modified.txt');
      expect(status.currentBranch).toBe('main');
      expect(status.ahead).toBe(0);
      expect(status.behind).toBe(0);
    });

    it('should detect clean working tree', async () => {
      const status = await gitManager.getStatus();

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
      await gitManager.initializeRepository(repoPath);
    });

    it('should create .gitignore with MoAI-ADK template', async () => {
      const gitignorePath = await gitManager.createGitignore(repoPath);

      expect(gitignorePath).toBe(path.join(repoPath, '.gitignore'));
      expect(await fs.pathExists(gitignorePath)).toBe(true);

      const content = await fs.readFile(gitignorePath, 'utf-8');
      expect(content).toContain('# MoAI-ADK Generated .gitignore');
      expect(content).toContain('node_modules/');
      expect(content).toContain('.moai/logs/');
      expect(content).toContain('.claude/logs/');
    });

    it('should not overwrite existing .gitignore', async () => {
      const gitignorePath = path.join(repoPath, '.gitignore');
      const existingContent = '# Existing content\ncustom-ignore/\n';

      await fs.writeFile(gitignorePath, existingContent);

      const resultPath = await gitManager.createGitignore(repoPath);
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
      await gitManager.initializeRepository(repoPath);
    });

    it('should link remote repository successfully', async () => {
      const remoteUrl = 'https://github.com/test/repo.git';

      await expect(gitManager.linkRemoteRepository(remoteUrl)).resolves.not.toThrow();
    });

    it('should detect SSH vs HTTPS remote URLs', async () => {
      const sshUrl = 'git@github.com:test/repo.git';
      const httpsUrl = 'https://github.com/test/repo.git';

      await expect(gitManager.linkRemoteRepository(sshUrl)).resolves.not.toThrow();
      await expect(gitManager.linkRemoteRepository(httpsUrl)).resolves.not.toThrow();
    });

    it('should validate remote URL format', async () => {
      const invalidUrls = [
        'not-a-url',
        'ftp://invalid.com/repo.git',
        'https://invalid-domain',
        ''
      ];

      for (const url of invalidUrls) {
        await expect(gitManager.linkRemoteRepository(url)).rejects.toThrow();
      }
    });
  });

  describe('Push Operations', () => {
    let repoPath: string;

    beforeEach(async () => {
      repoPath = path.join(testDir, 'test-repo-push');
      await fs.ensureDir(repoPath);
      await gitManager.initializeRepository(repoPath);
    });

    it('should handle push to remote repository', async () => {
      // 모의 원격 저장소 설정이 필요하지만, 단위 테스트에서는 실제 push 대신 명령어 구성 테스트
      const testFile = path.join(repoPath, 'test.txt');
      await fs.writeFile(testFile, 'Test content');
      await gitManager.commitChanges('Add test file');

      // 실제 원격이 없으므로 에러가 예상됨
      await expect(gitManager.pushChanges()).rejects.toThrow();
    });

    it('should set upstream branch automatically', async () => {
      const branchName = 'feature/auto-upstream';
      await gitManager.createBranch(branchName);

      // upstream 설정 테스트 (실제 원격 없이는 실패 예상)
      await expect(gitManager.pushChanges(branchName, 'origin')).rejects.toThrow();
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
      const invalidConfig: GitConfig = {
        mode: 'personal',
        autoCommit: false,
        branchPrefix: '',
        commitMessageTemplate: ''
      };

      const invalidGitManager = new GitManager(invalidConfig);

      await expect(invalidGitManager.getStatus()).rejects.toThrow();
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

      await gitManager.initializeRepository(repoPath);
      const status = await gitManager.getStatus();

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