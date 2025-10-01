// @CODE:GITHUB-001 | SPEC: github-integration
// Related: @CODE:GIT-003:API, @CODE:GIT-GH-001

/**
 * @file GitHub API integration facade
 * @author MoAI Team
 *
 * @fileoverview GitHub API integration using GitHub CLI
 * Refactored from monolithic github-integration.ts
 */

import type {
  CreatePullRequestOptions,
  CreateRepositoryOptions,
  GitConfig,
} from '@/types/git';
import { AuthChecker } from './auth-checker';
import { IssueManager } from './issue-manager';
import { PullRequestManager } from './pr-manager';
import { RepositoryManager } from './repo-manager';
import { WorkflowManager } from './workflow-manager';
import type { RepositoryInfo } from './types';

/**
 * GitHub CLI를 사용한 GitHub 연동 (Facade)
 */
export class GitHubIntegration {
  private readonly authChecker: AuthChecker;
  private readonly repoManager: RepositoryManager;
  private readonly prManager: PullRequestManager;
  private readonly issueManager: IssueManager;
  private readonly workflowManager: WorkflowManager;

  constructor(config: GitConfig) {
    if (config.mode !== 'team') {
      throw new Error('GitHub integration is only available in team mode');
    }

    // Initialize managers
    this.authChecker = new AuthChecker();
    this.repoManager = new RepositoryManager(this.authChecker);
    this.prManager = new PullRequestManager(this.authChecker);
    this.issueManager = new IssueManager(this.authChecker);
    this.workflowManager = new WorkflowManager();
  }

  /**
   * GitHub CLI 설치 확인
   */
  async isGitHubCliAvailable(): Promise<boolean> {
    return this.authChecker.isGitHubCliAvailable();
  }

  /**
   * GitHub CLI 인증 상태 확인
   */
  async isAuthenticated(): Promise<boolean> {
    return this.authChecker.isAuthenticated();
  }

  /**
   * GitHub 저장소 생성
   */
  async createRepository(options: CreateRepositoryOptions): Promise<void> {
    return this.repoManager.createRepository(options);
  }

  /**
   * Pull Request 생성
   */
  async createPullRequest(options: CreatePullRequestOptions): Promise<string> {
    return this.prManager.createPullRequest(options);
  }

  /**
   * 저장소 정보 조회
   */
  async getRepositoryInfo(): Promise<RepositoryInfo> {
    return this.repoManager.getRepositoryInfo();
  }

  /**
   * Issue 생성
   */
  async createIssue(
    title: string,
    body?: string,
    labels?: string[]
  ): Promise<string> {
    return this.issueManager.createIssue(title, body, labels);
  }

  /**
   * 라벨 생성
   */
  async createLabels(): Promise<void> {
    return this.issueManager.createLabels();
  }

  /**
   * GitHub Actions 워크플로우 파일 생성
   */
  async createWorkflowFile(
    workflowName: string,
    content: string
  ): Promise<void> {
    return this.workflowManager.createWorkflowFile(workflowName, content);
  }

  /**
   * 기본 GitHub Actions 워크플로우 생성
   */
  async setupDefaultWorkflows(): Promise<void> {
    return this.workflowManager.setupDefaultWorkflows();
  }
}

// Re-export types for backward compatibility
export type { RepositoryInfo } from './types';
