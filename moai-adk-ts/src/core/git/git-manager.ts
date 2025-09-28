/**
 * GitManager Implementation
 * SPEC-012 Week 2 Track D: Git System Integration
 *
 * @fileoverview Git operations manager using simple-git
 */

import simpleGit, { SimpleGit, StatusResult } from 'simple-git';
import * as fs from 'fs-extra';
import * as path from 'path';
import {
  GitConfig,
  GitInitResult,
  GitStatus,
  GitCommitResult,
  CreatePullRequestOptions,
  CreateRepositoryOptions,
  GitignoreTemplate,
  GitError,
  GitErrorType
} from '../../types/git';
import {
  GitNamingRules,
  GitCommitTemplates,
  GitignoreTemplates,
  GitDefaults,
  GitTimeouts
} from './constants';

/**
 * Git 작업을 관리하는 메인 클래스
 * Python GitManager의 모든 기능을 TypeScript로 포팅
 */
export class GitManager {
  private git: SimpleGit;
  private config: GitConfig;
  private currentWorkingDir: string;

  constructor(config: GitConfig, workingDir?: string) {
    this.config = config;
    this.currentWorkingDir = workingDir || process.cwd();
    this.git = simpleGit({
      baseDir: this.currentWorkingDir,
      binary: 'git',
      maxConcurrentProcesses: 1,
      timeout: {
        block: GitTimeouts.DEFAULT
      }
    });
  }

  /**
   * Git 저장소 초기화
   */
  async initializeRepository(projectPath: string): Promise<GitInitResult> {
    try {
      // 디렉토리 존재 확인
      if (!await fs.pathExists(projectPath)) {
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
          message: 'Repository already initialized'
        };
      }

      // 새 Git 인스턴스 생성 (프로젝트 경로용)
      const projectGit = simpleGit({
        baseDir: projectPath,
        binary: 'git',
        maxConcurrentProcesses: 1
      });

      // Git 저장소 초기화
      await projectGit.init();

      // 기본 설정 적용
      await projectGit.addConfig('init.defaultBranch', GitDefaults.DEFAULT_BRANCH);
      await projectGit.addConfig('core.autocrlf', GitDefaults.CONFIG['core.autocrlf']);
      await projectGit.addConfig('core.ignorecase', GitDefaults.CONFIG['core.ignorecase']);

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
        defaultBranch: GitDefaults.DEFAULT_BRANCH
      };

    } catch (error) {
      return {
        success: false,
        repositoryPath: projectPath,
        gitDir: path.join(projectPath, '.git'),
        defaultBranch: GitDefaults.DEFAULT_BRANCH,
        message: `Failed to initialize repository: ${(error as Error).message}`
      };
    }
  }

  /**
   * 새 브랜치 생성
   */
  async createBranch(branchName: string, baseBranch?: string): Promise<void> {
    try {
      // 브랜치명 검증
      if (!GitNamingRules.isValidBranchName(branchName)) {
        throw new Error(`Invalid branch name: ${branchName}`);
      }

      if (baseBranch) {
        await this.git.checkoutBranch(branchName, baseBranch);
      } else {
        await this.git.checkoutLocalBranch(branchName);
      }
    } catch (error) {
      throw this.createGitError('BRANCH_NOT_FOUND', `Failed to create branch: ${(error as Error).message}`);
    }
  }

  /**
   * 변경사항 커밋
   */
  async commitChanges(message: string, files?: string[]): Promise<GitCommitResult> {
    try {
      // 파일 스테이징
      if (files && files.length > 0) {
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
        author: latestCommit.author_name
      };

    } catch (error) {
      throw this.createGitError('UNKNOWN_ERROR', `Failed to commit changes: ${(error as Error).message}`);
    }
  }

  /**
   * 원격 저장소로 푸시
   */
  async pushChanges(branch?: string, remote?: string): Promise<void> {
    try {
      const targetBranch = branch || await this.getCurrentBranch();
      const targetRemote = remote || GitDefaults.DEFAULT_REMOTE;

      await this.git.push(targetRemote, targetBranch, ['--set-upstream']);
    } catch (error) {
      throw this.createGitError('NETWORK_ERROR', `Failed to push changes: ${(error as Error).message}`);
    }
  }

  /**
   * .gitignore 파일 생성
   */
  async createGitignore(projectPath: string, template: GitignoreTemplate = 'moai'): Promise<string> {
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
      throw new Error(`Failed to create .gitignore: ${(error as Error).message}`);
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
        behind: status.behind
      };
    } catch (error) {
      throw this.createGitError('REPOSITORY_NOT_FOUND', `Failed to get status: ${(error as Error).message}`);
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
      throw this.createGitError('UNKNOWN_ERROR', `Failed to create checkpoint: ${(error as Error).message}`);
    }
  }

  /**
   * 원격 저장소 연결
   */
  async linkRemoteRepository(repoUrl: string, remoteName: string = GitDefaults.DEFAULT_REMOTE): Promise<void> {
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
      throw this.createGitError('INVALID_REMOTE', `Failed to link remote repository: ${(error as Error).message}`);
    }
  }

  /**
   * GitHub 저장소 생성 (Team 모드)
   */
  async createRepository(_options: CreateRepositoryOptions): Promise<void> {
    if (this.config.mode !== 'team') {
      throw new Error('Repository creation is only available in team mode');
    }

    // GitHub CLI를 사용한 저장소 생성 (실제 구현 필요)
    throw new Error('GitHub repository creation not yet implemented');
  }

  /**
   * Pull Request 생성 (Team 모드)
   */
  async createPullRequest(_options: CreatePullRequestOptions): Promise<string> {
    if (this.config.mode !== 'team') {
      throw new Error('Pull request creation is only available in team mode');
    }

    // GitHub CLI를 사용한 PR 생성 (실제 구현 필요)
    throw new Error('Pull request creation not yet implemented');
  }

  // === Private 헬퍼 메서드 ===

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
      case 'moai':
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
      /^git@gitlab\.com:[\w\-_.]+\/[\w\-_.]+(?:\.git)?$/
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