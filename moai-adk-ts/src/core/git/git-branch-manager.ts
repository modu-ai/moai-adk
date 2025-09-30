// @CODE:GIT-BRANCH-001 | 
// Related: @CODE:GIT-BRANCH-001

/**
 * @file Git branch operations manager
 * @author MoAI Team
 * @description 브랜치 관리 전담 클래스 (GitManager에서 분리)
 */

import * as path from 'node:path';
import * as fs from 'fs-extra';
import simpleGit, { type SimpleGit } from 'simple-git';
import type { GitConfig, GitInitResult } from '../../types/git';
import { InputValidator } from '../../utils/input-validator';
import { GitDefaults, GitNamingRules, GitTimeouts } from './constants';
import { GitLockManager } from './git-lock-manager';

/**
 * Git 브랜치 작업 전담 클래스
 *
 * 책임:
 * - 브랜치 생성/전환/조회
 * - 브랜치명 검증
 * - Lock을 사용한 안전한 브랜치 작업
 */
export class GitBranchManager {
  public readonly git: SimpleGit; // 테스트를 위해 public으로 변경
  private config: GitConfig;
  private workingDir: string;
  private lockManager: GitLockManager;

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
  async initializeRepository(): Promise<GitInitResult> {
    try {
      // 디렉토리 존재 확인
      if (!(await fs.pathExists(this.workingDir))) {
        await fs.ensureDir(this.workingDir);
      }

      const gitDir = path.join(this.workingDir, '.git');

      // 이미 Git 저장소인지 확인
      if (await fs.pathExists(gitDir)) {
        return {
          success: true,
          repositoryPath: this.workingDir,
          gitDir,
          defaultBranch: GitDefaults.DEFAULT_BRANCH,
          message: 'Repository already initialized',
        };
      }

      // Git 저장소 초기화
      await this.git.init();

      // 기본 설정 적용
      await this.git.addConfig('init.defaultBranch', GitDefaults.DEFAULT_BRANCH);
      await this.git.addConfig('core.autocrlf', GitDefaults.CONFIG['core.autocrlf']);
      await this.git.addConfig('core.ignorecase', GitDefaults.CONFIG['core.ignorecase']);

      // 기본 브랜치로 체크아웃
      try {
        await this.git.checkoutLocalBranch(GitDefaults.DEFAULT_BRANCH);
      } catch {
        // 이미 기본 브랜치에 있거나 커밋이 없는 경우 무시
      }

      return {
        success: true,
        repositoryPath: this.workingDir,
        gitDir,
        defaultBranch: GitDefaults.DEFAULT_BRANCH,
      };
    } catch (error) {
      return {
        success: false,
        repositoryPath: this.workingDir,
        gitDir: path.join(this.workingDir, '.git'),
        defaultBranch: GitDefaults.DEFAULT_BRANCH,
        message: `Failed to initialize repository: ${(error as Error).message}`,
      };
    }
  }

  /**
   * 새 브랜치 생성
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

      // Git 명명 규칙 검증
      if (!GitNamingRules.isValidBranchName(branchName)) {
        throw new Error(`Invalid branch name: ${branchName}`);
      }

      const safeBranchName = validation.sanitizedValue || branchName;

      // 초기 커밋이 없는 경우 기본 파일 생성
      const status = await this.git.status();
      if (status.files.length === 0) {
        const readmePath = path.join(this.workingDir, 'README.md');
        if (!(await fs.pathExists(readmePath))) {
          await fs.writeFile(readmePath, '# Project\n\nInitial commit\n');
          await this.git.add('README.md');
          await this.git.commit('Initial commit');
        }
      }

      if (baseBranch) {
        // 베이스 브랜치 검증
        const branches = await this.git.branch();
        const baseBranchValidation = InputValidator.validateBranchName(baseBranch);

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
      throw new Error(`Failed to create branch: ${(error as Error).message}`);
    }
  }

  /**
   * 현재 브랜치명 조회
   */
  async getCurrentBranch(): Promise<string> {
    try {
      const branchResult = await this.git.branch();
      return branchResult.current;
    } catch {
      return GitDefaults.DEFAULT_BRANCH;
    }
  }

  /**
   * 모든 브랜치 목록 조회
   */
  async listBranches(): Promise<string[]> {
    try {
      const branchResult = await this.git.branch();
      return branchResult.all;
    } catch {
      return [];
    }
  }

  /**
   * 브랜치 전환
   */
  async switchBranch(branchName: string): Promise<void> {
    try {
      await this.git.checkout(branchName);
    } catch (error) {
      throw new Error(`Failed to switch branch: ${(error as Error).message}`);
    }
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
    return await this.lockManager.withLock(
      () => this.createBranch(branchName, baseBranch),
      `create-branch: ${branchName}`,
      wait,
      timeout
    );
  }
}
