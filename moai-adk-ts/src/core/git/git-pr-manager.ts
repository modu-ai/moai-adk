// @CODE:GIT-PR-001 |
// Related: @CODE:GIT-PR-001

/**
 * @file Git PR and remote operations manager
 * @author MoAI Team
 * @description Pull Request 및 원격 저장소 관리 전담 클래스 (GitManager에서 분리)
 */

import * as fs from 'fs-extra';
import simpleGit, { type SimpleGit } from 'simple-git';
import type {
  CreatePullRequestOptions,
  CreateRepositoryOptions,
  GitConfig,
} from '../../types/git';
import { GitDefaults, GitTimeouts } from './constants';
import { GitLockManager } from './git-lock-manager';
import { GitHubIntegration } from './github-integration';

/**
 * Git PR 및 원격 작업 전담 클래스
 *
 * 책임:
 * - Pull Request 생성
 * - 원격 저장소 생성
 * - 원격 저장소 푸시
 * - GitHub CLI 연동
 * - Lock을 사용한 안전한 원격 작업
 */
export class GitPRManager {
  public readonly git: SimpleGit;
  private workingDir: string;
  private lockManager: GitLockManager;
  private githubIntegration: GitHubIntegration;
  private config: GitConfig; // Used for PR configuration

  constructor(config: GitConfig, workingDir?: string) {
    this.validateConfig(config);
    this.config = config;
    this.workingDir = workingDir || process.cwd();

    // 디렉토리가 존재하지 않는 경우 동기적으로 생성
    if (!fs.pathExistsSync(this.workingDir)) {
      fs.ensureDirSync(this.workingDir);
    }

    this.git = this.createGitInstance(this.workingDir);
    this.lockManager = new GitLockManager(this.workingDir);
    this.githubIntegration = new GitHubIntegration(config);
  }

  /**
   * Git 인스턴스 생성
   */
  private createGitInstance(baseDir: string): SimpleGit {
    return simpleGit({
      baseDir,
      binary: 'git',
      maxConcurrentProcesses: 1,
      timeout: {
        block: GitTimeouts.DEFAULT,
      },
      config: [
        'core.autocrlf=input',
        'core.ignorecase=false',
        'init.defaultBranch=main',
      ],
    });
  }

  /**
   * 설정 검증
   */
  private validateConfig(config: GitConfig): void {
    if (config.mode !== 'team') {
      throw new Error('GitPRManager requires team mode');
    }

    if (!config.github) {
      throw new Error('GitHub configuration is required for team mode');
    }

    if (typeof config.autoCommit !== 'boolean') {
      throw new Error('autoCommit must be a boolean');
    }

    if (!config.branchPrefix || typeof config.branchPrefix !== 'string') {
      throw new Error('branchPrefix must be a non-empty string');
    }
  }

  /**
   * Pull Request 생성
   */
  async createPullRequest(options: CreatePullRequestOptions): Promise<string> {
    try {
      // 옵션 검증
      if (!options.title || options.title.trim() === '') {
        throw new Error('PR title is required');
      }

      return await this.githubIntegration.createPullRequest(options);
    } catch (error) {
      throw new Error(
        `Failed to create pull request: ${(error as Error).message}`
      );
    }
  }

  /**
   * GitHub 저장소 생성
   */
  async createRepository(options: CreateRepositoryOptions): Promise<void> {
    try {
      // 옵션 검증
      if (!options.name || options.name.trim() === '') {
        throw new Error('Repository name is required');
      }

      await this.githubIntegration.createRepository(options);
    } catch (error) {
      throw new Error(
        `Failed to create repository: ${(error as Error).message}`
      );
    }
  }

  /**
   * 원격 저장소로 푸시
   */
  async pushChanges(branch?: string, remote?: string): Promise<void> {
    try {
      const targetBranch = branch || (await this.getCurrentBranch());
      const targetRemote = remote || GitDefaults.DEFAULT_REMOTE;

      await this.git.push(targetRemote, targetBranch, ['--set-upstream']);
    } catch (error) {
      throw new Error(`Failed to push changes: ${(error as Error).message}`);
    }
  }

  /**
   * Lock을 사용한 안전한 푸시
   */
  async pushWithLock(
    branch?: string,
    remote?: string,
    wait: boolean = true,
    timeout: number = 30
  ): Promise<void> {
    return await this.lockManager.withLock(
      () => this.pushChanges(branch, remote),
      `push: ${branch || 'current'} -> ${remote || 'origin'}`,
      wait,
      timeout
    );
  }

  /**
   * 원격 저장소 연결
   */
  async linkRemoteRepository(
    repoUrl: string,
    remoteName: string = GitDefaults.DEFAULT_REMOTE
  ): Promise<void> {
    try {
      // URL 검증
      if (!this.isValidGitUrl(repoUrl)) {
        throw new Error(`Invalid Git URL: ${repoUrl}`);
      }

      // 기존 원격 제거 (있는 경우)
      try {
        await this.git.removeRemote(remoteName);
      } catch {
        // 원격이 없는 경우 무시
      }

      // 새 원격 추가
      await this.git.addRemote(remoteName, repoUrl);
    } catch (error) {
      throw new Error(
        `Failed to link remote repository: ${(error as Error).message}`
      );
    }
  }

  /**
   * GitHub CLI 가용성 확인
   */
  async isGitHubCliAvailable(): Promise<boolean> {
    return await this.githubIntegration.isGitHubCliAvailable();
  }

  /**
   * GitHub 인증 상태 확인
   */
  async isGitHubAuthenticated(): Promise<boolean> {
    return await this.githubIntegration.isAuthenticated();
  }

  /**
   * Lock 상태 조회
   */
  async getLockStatus() {
    return await this.lockManager.getLockStatus();
  }

  /**
   * 오래된 Lock 정리
   */
  async cleanupStaleLocks(): Promise<void> {
    return await this.lockManager.cleanupStaleLocks();
  }

  /**
   * 현재 브랜치명 조회
   */
  private async getCurrentBranch(): Promise<string> {
    try {
      const branchResult = await this.git.branch();
      return branchResult.current;
    } catch {
      return GitDefaults.DEFAULT_BRANCH;
    }
  }

  /**
   * Git URL 검증
   */
  private isValidGitUrl(url: string): boolean {
    const gitUrlPatterns = [
      /^https:\/\/github\.com\/[\w\-_.]+\/[\w\-_.]+(?:\.git)?$/,
      /^git@github\.com:[\w\-_.]+\/[\w\-_.]+(?:\.git)?$/,
      /^https:\/\/gitlab\.com\/[\w\-_.]+\/[\w\-_.]+(?:\.git)?$/,
      /^git@gitlab\.com:[\w\-_.]+\/[\w\-_.]+(?:\.git)?$/,
    ];

    return gitUrlPatterns.some(pattern => pattern.test(url));
  }
}
