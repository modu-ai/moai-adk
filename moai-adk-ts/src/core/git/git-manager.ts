// @CODE:GIT-001 |
// Related: @CODE:GIT-001:API, @CODE:GIT-CFG-001

/**
 * @file Git operations manager (Refactored)
 * @author MoAI Team
 * @description 리팩토링된 GitManager - 3개의 전문 매니저를 조합
 */

import * as path from 'node:path';
import * as fs from 'fs-extra';
import type {
  CreatePullRequestOptions,
  CreateRepositoryOptions,
  GitCommitResult,
  GitConfig,
  GitInitResult,
  GitignoreTemplate,
  GitStatus,
} from '../../types/git';
import { GitignoreTemplates } from './constants';
import { GitBranchManager } from './git-branch-manager';
import { GitCommitManager } from './git-commit-manager';
import { GitPRManager } from './git-pr-manager';

/**
 * Git 작업을 관리하는 메인 클래스 (리팩토링됨)
 *
 * 이전: 689 LOC의 단일 클래스
 * 이후: 3개의 전문 매니저를 조합하는 Facade 패턴
 *
 * TRUST 원칙 준수:
 * - Test First: 모든 메서드가 테스트로 검증됨
 * - Readable: 명확한 책임 분리와 메서드명
 * - Unified: 단일 책임 원칙 준수
 * - Secured: 입력 검증 및 에러 처리
 * - Trackable: 모든 Git 작업 추적 가능
 */
export class GitManager {
  private branchManager: GitBranchManager;
  private commitManager: GitCommitManager;
  private prManager?: GitPRManager;
  private config: GitConfig;
  private currentWorkingDir: string;

  constructor(config: GitConfig, workingDir?: string) {
    this.config = config;
    this.currentWorkingDir = workingDir || process.cwd();

    // 3개의 전문 매니저 초기화
    this.branchManager = new GitBranchManager(config, this.currentWorkingDir);
    this.commitManager = new GitCommitManager(config, this.currentWorkingDir);

    // Team 모드인 경우 PR 매니저 초기화
    if (config.mode === 'team') {
      this.prManager = new GitPRManager(config, this.currentWorkingDir);
    }
  }

  // ===== 브랜치 관리 (GitBranchManager 위임) =====

  /**
   * Git 저장소 초기화
   */
  async initializeRepository(projectPath?: string): Promise<GitInitResult> {
    // 임시로 새 브랜치 매니저 생성 (다른 경로용)
    if (projectPath && projectPath !== this.currentWorkingDir) {
      try {
        const tempBranchManager = new GitBranchManager(
          this.config,
          projectPath
        );
        return await tempBranchManager.initializeRepository();
      } catch (error) {
        // 실패 시 GitInitResult 반환
        return {
          success: false,
          repositoryPath: projectPath,
          gitDir: path.join(projectPath, '.git'),
          defaultBranch: 'main',
          message: `Failed to initialize repository: ${(error as Error).message}`,
        };
      }
    }
    return await this.branchManager.initializeRepository();
  }

  /**
   * 새 브랜치 생성
   */
  async createBranch(branchName: string, baseBranch?: string): Promise<void> {
    return await this.branchManager.createBranch(branchName, baseBranch);
  }

  /**
   * Lock을 사용한 안전한 브랜치 생성
   */
  async createBranchWithLock(
    branchName: string,
    baseBranch?: string,
    wait: boolean = true,
    timeout: number = 30
  ): Promise<void> {
    return await this.branchManager.createBranchWithLock(
      branchName,
      baseBranch,
      wait,
      timeout
    );
  }

  /**
   * 현재 브랜치명 조회
   */
  async getCurrentBranch(): Promise<string> {
    return await this.branchManager.getCurrentBranch();
  }

  // ===== 커밋 관리 (GitCommitManager 위임) =====

  /**
   * 변경사항 커밋
   */
  async commitChanges(
    message: string,
    files?: string[]
  ): Promise<GitCommitResult> {
    return await this.commitManager.commitChanges(message, files);
  }

  /**
   * Lock을 사용한 안전한 커밋 실행
   */
  async commitWithLock(
    message: string,
    files?: string[],
    wait: boolean = true,
    timeout: number = 30
  ): Promise<GitCommitResult> {
    return await this.commitManager.commitWithLock(
      message,
      files,
      wait,
      timeout
    );
  }

  /**
   * 체크포인트 생성
   */
  async createCheckpoint(message: string): Promise<string> {
    return await this.commitManager.createCheckpoint(message);
  }

  /**
   * 저장소 상태 조회
   */
  async getStatus(): Promise<GitStatus> {
    return await this.commitManager.getStatus();
  }

  // ===== PR 및 원격 관리 (GitPRManager 위임) =====

  /**
   * 원격 저장소로 푸시
   */
  async pushChanges(branch?: string, remote?: string): Promise<void> {
    if (!this.prManager) {
      throw new Error('PR Manager is only available in team mode');
    }
    return await this.prManager.pushChanges(branch, remote);
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
    if (!this.prManager) {
      throw new Error('PR Manager is only available in team mode');
    }
    return await this.prManager.pushWithLock(branch, remote, wait, timeout);
  }

  /**
   * GitHub 저장소 생성 (Team 모드)
   */
  async createRepository(options: CreateRepositoryOptions): Promise<void> {
    if (!this.prManager) {
      throw new Error('Repository creation is only available in team mode');
    }
    return await this.prManager.createRepository(options);
  }

  /**
   * Pull Request 생성 (Team 모드)
   */
  async createPullRequest(options: CreatePullRequestOptions): Promise<string> {
    if (!this.prManager) {
      throw new Error('Pull request creation is only available in team mode');
    }
    return await this.prManager.createPullRequest(options);
  }

  /**
   * 원격 저장소 연결
   */
  async linkRemoteRepository(
    repoUrl: string,
    remoteName?: string
  ): Promise<void> {
    // personal 모드에서도 remote 연결 가능하도록 수정
    if (this.prManager) {
      return await this.prManager.linkRemoteRepository(repoUrl, remoteName);
    }

    // personal 모드: 브랜치 매니저의 git 인스턴스 직접 사용
    const git = this.branchManager.git;
    const targetRemote = remoteName || 'origin';

    // URL 검증
    if (!this.isValidGitUrl(repoUrl)) {
      throw new Error(`Invalid Git URL: ${repoUrl}`);
    }

    try {
      await git.removeRemote(targetRemote);
    } catch {
      // 원격이 없는 경우 무시
    }

    await git.addRemote(targetRemote, repoUrl);
  }

  /**
   * GitHub CLI 가용성 확인
   */
  async isGitHubCliAvailable(): Promise<boolean> {
    if (!this.prManager) {
      return false;
    }
    return await this.prManager.isGitHubCliAvailable();
  }

  /**
   * GitHub 인증 상태 확인
   */
  async isGitHubAuthenticated(): Promise<boolean> {
    if (!this.prManager) {
      return false;
    }
    return await this.prManager.isGitHubAuthenticated();
  }

  // ===== Lock 관리 =====

  /**
   * Lock 상태 조회
   */
  async getLockStatus() {
    // commitManager를 통해 lock 상태 조회
    return await this.commitManager.getLockStatus();
  }

  /**
   * 오래된 Lock 정리
   */
  async cleanupStaleLocks(): Promise<void> {
    return await this.commitManager.cleanupStaleLocks();
  }

  // ===== 기타 유틸리티 =====

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
   * 저장소 유효성 확인
   */
  async isValidRepository(): Promise<boolean> {
    try {
      await this.commitManager.getStatus();
      return true;
    } catch {
      return false;
    }
  }

  /**
   * 배치 작업 수행
   */
  async performBatchOperations(
    operations: (() => Promise<void>)[]
  ): Promise<void> {
    try {
      for (const operation of operations) {
        await operation();
      }
    } catch (error) {
      throw new Error(`Batch operation failed: ${(error as Error).message}`);
    }
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
}
