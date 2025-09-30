// @FEATURE:GIT-001 | Chain: @REQ:GIT-001 -> @DESIGN:GIT-001 -> @TASK:GIT-001 -> @TEST:GIT-001
// Related: @API:GIT-001, @DATA:GIT-CFG-001

/**
 * @file Git operations manager
 * @author MoAI Team
 */

import * as path from 'node:path';
import * as fs from 'fs-extra';
import simpleGit, { type SimpleGit, type StatusResult } from 'simple-git';
import type {
  CreatePullRequestOptions,
  CreateRepositoryOptions,
  GitCommitResult,
  GitConfig,
  GitError,
  GitErrorType,
  GitInitResult,
  GitignoreTemplate,
  GitStatus,
} from '../../types/git';
import { InputValidator } from '../../utils/input-validator';
import {
  GitCommitTemplates,
  GitDefaults,
  GitignoreTemplates,
  GitNamingRules,
  GitTimeouts,
} from './constants';
import { GitLockManager } from './git-lock-manager';
import { GitHubIntegration } from './github-integration';

/**
 * Git 작업을 관리하는 메인 클래스
 * Python GitManager의 모든 기능을 TypeScript로 포팅
 *
 * TRUST 원칙 준수:
 * - Test First: 모든 메서드가 테스트로 검증됨
 * - Readable: 명확한 메서드명과 타입 정의
 * - Unified: 단일 책임 원칙 준수
 * - Secured: 입력 검증 및 에러 처리
 * - Trackable: 모든 Git 작업 추적 가능
 */
export class GitManager {
  private git: SimpleGit;
  private config: GitConfig;
  private currentWorkingDir: string;
  private githubIntegration?: GitHubIntegration;
  private lockManager: GitLockManager;

  /**
   * @tags @FEATURE:GIT-MANAGER-001 @API:CONSTRUCTOR-001
   */
  constructor(config: GitConfig, workingDir?: string) {
    this.validateConfig(config);
    this.config = config;
    this.currentWorkingDir = workingDir || process.cwd();
    this.git = this.createGitInstance(this.currentWorkingDir);

    // Initialize Git lock manager for concurrent operation safety
    this.lockManager = new GitLockManager(this.currentWorkingDir);

    // Team 모드인 경우 GitHub 연동 초기화
    if (config.mode === 'team') {
      this.githubIntegration = new GitHubIntegration(config);
    }
  }

  /**
   * Git 인스턴스 생성 (설정 최적화)
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
    if (!config.mode || !['personal', 'team'].includes(config.mode)) {
      throw new Error('Invalid mode: must be "personal" or "team"');
    }

    if (typeof config.autoCommit !== 'boolean') {
      throw new Error('autoCommit must be a boolean');
    }

    if (!config.branchPrefix || typeof config.branchPrefix !== 'string') {
      throw new Error('branchPrefix must be a non-empty string');
    }
  }

  /**
   * Git 저장소 초기화
   */
  async initializeRepository(projectPath: string): Promise<GitInitResult> {
    try {
      // 디렉토리 존재 확인
      if (!(await fs.pathExists(projectPath))) {
        await fs.ensureDir(projectPath);
      }

      const gitDir = path.join(projectPath, '.git');

      // 이미 Git 저장소인지 확인
      if (await fs.pathExists(gitDir)) {
        return {
          success: true,
          repositoryPath: projectPath,
          gitDir,
          defaultBranch: GitDefaults.DEFAULT_BRANCH,
          message: 'Repository already initialized',
        };
      }

      // 새 Git 인스턴스 생성 (프로젝트 경로용)
      const projectGit = simpleGit({
        baseDir: projectPath,
        binary: 'git',
        maxConcurrentProcesses: 1,
      });

      // Git 저장소 초기화
      await projectGit.init();

      // 기본 설정 적용
      await projectGit.addConfig(
        'init.defaultBranch',
        GitDefaults.DEFAULT_BRANCH
      );
      await projectGit.addConfig(
        'core.autocrlf',
        GitDefaults.CONFIG['core.autocrlf']
      );
      await projectGit.addConfig(
        'core.ignorecase',
        GitDefaults.CONFIG['core.ignorecase']
      );

      // 기본 브랜치로 체크아웃 (필요한 경우)
      try {
        await projectGit.checkoutLocalBranch(GitDefaults.DEFAULT_BRANCH);
      } catch {
        // 이미 기본 브랜치에 있거나 커밋이 없는 경우 무시
      }

      return {
        success: true,
        repositoryPath: projectPath,
        gitDir,
        defaultBranch: GitDefaults.DEFAULT_BRANCH,
      };
    } catch (error) {
      return {
        success: false,
        repositoryPath: projectPath,
        gitDir: path.join(projectPath, '.git'),
        defaultBranch: GitDefaults.DEFAULT_BRANCH,
        message: `Failed to initialize repository: ${(error as Error).message}`,
      };
    }
  }

  /**
   * 새 브랜치 생성 (보안 강화)
   */
  async createBranch(branchName: string, baseBranch?: string): Promise<void> {
    try {
      // 강화된 브랜치명 검증
      const validation = InputValidator.validateBranchName(branchName);
      if (!validation.isValid) {
        throw new Error(
          `Branch name validation failed: ${validation.errors.join(', ')}`
        );
      }

      // 추가적인 Git 명명 규칙 검증
      if (!GitNamingRules.isValidBranchName(branchName)) {
        throw new Error(`Invalid branch name: ${branchName}`);
      }

      // 안전한 브랜치명 사용
      const safeBranchName = validation.sanitizedValue || branchName;

      // 초기 커밋이 없는 경우 기본 파일 생성
      const status = await this.git.status();
      if (status.files.length === 0) {
        // README.md 파일 생성하여 초기 커밋 만들기
        const readmePath = path.join(this.currentWorkingDir, 'README.md');
        if (!(await fs.pathExists(readmePath))) {
          await fs.writeFile(readmePath, '# Project\n\nInitial commit\n');
          await this.git.add('README.md');
          await this.git.commit('Initial commit');
        }
      }

      if (baseBranch) {
        // 베이스 브랜치가 존재하는지 확인
        const branches = await this.git.branch();
        // 베이스 브랜치명도 검증
        const baseBranchValidation =
          InputValidator.validateBranchName(baseBranch);
        if (!baseBranchValidation.isValid) {
          throw new Error(
            `Base branch name validation failed: ${baseBranchValidation.errors.join(', ')}`
          );
        }

        if (
          !branches.all.includes(baseBranch) &&
          baseBranch !== 'main' &&
          baseBranch !== 'master'
        ) {
          throw new Error(`Base branch '${baseBranch}' does not exist`);
        }
        await this.git.checkoutBranch(safeBranchName, baseBranch);
      } else {
        await this.git.checkoutLocalBranch(safeBranchName);
      }
    } catch (error) {
      throw this.createGitError(
        'BRANCH_NOT_FOUND',
        `Failed to create branch: ${(error as Error).message}`
      );
    }
  }

  /**
   * 변경사항 커밋
   */
  async commitChanges(
    message: string,
    files?: string[]
  ): Promise<GitCommitResult> {
    try {
      // 초기 커밋이 없는 경우 처리
      const status = await this.git.status();
      if (status.files.length === 0 && (!files || files.length === 0)) {
        // README.md 파일 생성하여 초기 커밋 만들기
        const readmePath = path.join(this.currentWorkingDir, 'README.md');
        if (!(await fs.pathExists(readmePath))) {
          await fs.writeFile(readmePath, '# Project\n\nInitial commit\n');
        }
      }

      // 파일 스테이징
      if (files && files.length > 0) {
        // 파일 존재 여부 확인
        for (const file of files) {
          const filePath = path.isAbsolute(file)
            ? file
            : path.join(this.currentWorkingDir, file);
          if (!(await fs.pathExists(filePath))) {
            throw new Error(`File not found: ${file}`);
          }
        }
        await this.git.add(files);
      } else {
        await this.git.add('.');
      }

      // 커밋 템플릿 적용
      const formattedMessage = this.applyCommitTemplate(message);

      // 커밋 실행
      const commitResult = await this.git.commit(formattedMessage);

      // 커밋 정보 조회
      const logResult = await this.git.log(['-1']);
      const latestCommit = logResult.latest;

      if (!latestCommit) {
        throw new Error('Failed to retrieve commit information');
      }

      return {
        hash: latestCommit.hash,
        message: latestCommit.message,
        timestamp: new Date(latestCommit.date),
        filesChanged: commitResult.summary.changes,
        author: latestCommit.author_name,
      };
    } catch (error) {
      throw this.createGitError(
        'UNKNOWN_ERROR',
        `Failed to commit changes: ${(error as Error).message}`
      );
    }
  }

  /**
   * Lock을 사용한 안전한 커밋 실행
   * Python git_manager.py의 commit_with_lock 포팅
   * @param message 커밋 메시지
   * @param files 커밋할 파일 목록 (선택사항)
   * @param wait Lock 획득 대기 여부
   * @param timeout Lock 획득 타임아웃 (초)
   * @returns 커밋 결과
   * @tags @API:COMMIT-WITH-LOCK-001 @FEATURE:GIT-LOCK-INTEGRATION-001
   */
  async commitWithLock(
    message: string,
    files?: string[],
    wait: boolean = true,
    timeout: number = 30
  ): Promise<GitCommitResult> {
    return await this.lockManager.withLock(
      () => this.commitChanges(message, files),
      `commit: ${message.substring(0, 50)}...`,
      wait,
      timeout
    );
  }

  /**
   * Lock을 사용한 안전한 브랜치 생성
   * @param branchName 생성할 브랜치명
   * @param baseBranch 기준 브랜치 (선택사항)
   * @param wait Lock 획득 대기 여부
   * @param timeout Lock 획득 타임아웃 (초)
   * @tags @API:CREATE-BRANCH-WITH-LOCK-001 @FEATURE:GIT-LOCK-INTEGRATION-001
   */
  async createBranchWithLock(
    branchName: string,
    baseBranch?: string,
    wait: boolean = true,
    timeout: number = 30
  ): Promise<void> {
    return await this.lockManager.withLock(
      () => this.createBranch(branchName, baseBranch),
      `create-branch: ${branchName}`,
      wait,
      timeout
    );
  }

  /**
   * Lock을 사용한 안전한 푸시
   * @param branch 푸시할 브랜치 (선택사항)
   * @param remote 원격 저장소명 (선택사항)
   * @param wait Lock 획득 대기 여부
   * @param timeout Lock 획득 타임아웃 (초)
   * @tags @API:PUSH-WITH-LOCK-001 @FEATURE:GIT-LOCK-INTEGRATION-001
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
   * Lock 상태 조회
   * @returns Git lock 상태 정보
   * @tags @API:GET-LOCK-STATUS-001 @FEATURE:GIT-LOCK-INTEGRATION-001
   */
  async getLockStatus() {
    return await this.lockManager.getLockStatus();
  }

  /**
   * 오래된 Lock 정리
   * @tags @API:CLEANUP-STALE-LOCKS-001 @FEATURE:GIT-LOCK-INTEGRATION-001
   */
  async cleanupStaleLocks(): Promise<void> {
    return await this.lockManager.cleanupStaleLocks();
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
      throw this.createGitError(
        'NETWORK_ERROR',
        `Failed to push changes: ${(error as Error).message}`
      );
    }
  }

  /**
   * .gitignore 파일 생성
   */
  async createGitignore(
    projectPath: string,
    template: GitignoreTemplate = 'moai'
  ): Promise<string> {
    const gitignorePath = path.join(projectPath, '.gitignore');

    try {
      // 기존 .gitignore 확인
      if (await fs.pathExists(gitignorePath)) {
        const existingContent = await fs.readFile(gitignorePath, 'utf-8');

        // MoAI 템플릿이 이미 있는지 확인
        if (existingContent.includes('# MoAI-ADK Generated .gitignore')) {
          return gitignorePath;
        }

        // 기존 내용과 템플릿 병합
        const templateContent = this.getGitignoreTemplate(template);
        const mergedContent = `${existingContent}\n\n${templateContent}`;
        await fs.writeFile(gitignorePath, mergedContent);
      } else {
        // 새 .gitignore 생성
        const templateContent = this.getGitignoreTemplate(template);
        await fs.writeFile(gitignorePath, templateContent);
      }

      return gitignorePath;
    } catch (error) {
      throw new Error(
        `Failed to create .gitignore: ${(error as Error).message}`
      );
    }
  }

  /**
   * 저장소 상태 조회
   */
  async getStatus(): Promise<GitStatus> {
    try {
      const status: StatusResult = await this.git.status();
      const currentBranch = await this.getCurrentBranch();

      return {
        clean: status.isClean(),
        modified: status.modified,
        added: status.staged,
        deleted: status.deleted,
        untracked: status.not_added,
        currentBranch,
        ahead: status.ahead,
        behind: status.behind,
      };
    } catch (error) {
      throw this.createGitError(
        'REPOSITORY_NOT_FOUND',
        `Failed to get status: ${(error as Error).message}`
      );
    }
  }

  /**
   * 체크포인트 생성
   */
  async createCheckpoint(message: string): Promise<string> {
    try {
      const checkpointMessage = GitCommitTemplates.createCheckpoint(message);
      const result = await this.commitChanges(checkpointMessage);
      return result.hash;
    } catch (error) {
      throw this.createGitError(
        'UNKNOWN_ERROR',
        `Failed to create checkpoint: ${(error as Error).message}`
      );
    }
  }

  /**
   * 저장소 정보 캐시
   */
  private repositoryInfoCache: {
    isRepo?: boolean;
    hasCommits?: boolean;
    lastChecked?: number;
  } = {};

  /**
   * 저장소 유효성 확인 (캐시 사용)
   */
  async isValidRepository(): Promise<boolean> {
    const now = Date.now();
    const cacheExpiry = 5000; // 5초 캐시

    if (
      this.repositoryInfoCache.isRepo !== undefined &&
      this.repositoryInfoCache.lastChecked &&
      now - this.repositoryInfoCache.lastChecked < cacheExpiry
    ) {
      return this.repositoryInfoCache.isRepo;
    }

    try {
      await this.git.status();
      this.repositoryInfoCache.isRepo = true;
      this.repositoryInfoCache.lastChecked = now;
      return true;
    } catch {
      this.repositoryInfoCache.isRepo = false;
      this.repositoryInfoCache.lastChecked = now;
      return false;
    }
  }

  /**
   * 배치 작업 수행 (성능 최적화)
   */
  async performBatchOperations(
    operations: (() => Promise<void>)[]
  ): Promise<void> {
    try {
      // 작업들을 순차적으로 실행 (Git은 동시 실행 불가)
      for (const operation of operations) {
        await operation();
      }
    } catch (error) {
      throw this.createGitError(
        'UNKNOWN_ERROR',
        `Batch operation failed: ${(error as Error).message}`
      );
    }
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
      throw this.createGitError(
        'INVALID_REMOTE',
        `Failed to link remote repository: ${(error as Error).message}`
      );
    }
  }

  /**
   * GitHub 저장소 생성 (Team 모드)
   */
  async createRepository(options: CreateRepositoryOptions): Promise<void> {
    if (this.config.mode !== 'team' || !this.githubIntegration) {
      throw new Error('Repository creation is only available in team mode');
    }

    try {
      await this.githubIntegration.createRepository(options);
    } catch (error) {
      throw this.createGitError(
        'UNKNOWN_ERROR',
        `Failed to create repository: ${(error as Error).message}`
      );
    }
  }

  /**
   * Pull Request 생성 (Team 모드)
   */
  async createPullRequest(options: CreatePullRequestOptions): Promise<string> {
    if (this.config.mode !== 'team' || !this.githubIntegration) {
      throw new Error('Pull request creation is only available in team mode');
    }

    try {
      return await this.githubIntegration.createPullRequest(options);
    } catch (error) {
      throw this.createGitError(
        'UNKNOWN_ERROR',
        `Failed to create pull request: ${(error as Error).message}`
      );
    }
  }

  /**
   * GitHub CLI 가용성 확인
   */
  async isGitHubCliAvailable(): Promise<boolean> {
    if (!this.githubIntegration) {
      return false;
    }

    return await this.githubIntegration.isGitHubCliAvailable();
  }

  /**
   * GitHub 인증 상태 확인
   */
  async isGitHubAuthenticated(): Promise<boolean> {
    if (!this.githubIntegration) {
      return false;
    }

    return await this.githubIntegration.isAuthenticated();
  }

  // === Private 헬퍼 메서드 ===

  /**
   * 현재 브랜치명 조회
   */
  public async getCurrentBranch(): Promise<string> {
    try {
      const branchResult = await this.git.branch();
      return branchResult.current;
    } catch {
      return GitDefaults.DEFAULT_BRANCH;
    }
  }

  /**
   * 커밋 메시지 템플릿 적용
   */
  private applyCommitTemplate(message: string): string {
    if (!this.config.commitMessageTemplate) {
      return message;
    }

    return GitCommitTemplates.apply(this.config.commitMessageTemplate, message);
  }

  /**
   * .gitignore 템플릿 내용 반환
   */
  private getGitignoreTemplate(template: GitignoreTemplate): string {
    switch (template) {
      case 'node':
        return GitignoreTemplates.NODE;
      case 'python':
        return GitignoreTemplates.PYTHON;
      default:
        return GitignoreTemplates.MOAI;
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

  /**
   * Git 에러 생성
   */
  private createGitError(type: GitErrorType, message: string): GitError {
    const error = new Error(message) as GitError;
    (error as any).type = type;
    return error;
  }
}
