// @CODE:GIT-COMMIT-001 | 
// Related: @CODE:GIT-COMMIT-001

/**
 * @file Git commit operations manager
 * @author MoAI Team
 * @description 커밋 관리 전담 클래스 (GitManager에서 분리)
 */

import * as path from 'node:path';
import * as fs from 'fs-extra';
import simpleGit, { type SimpleGit, type StatusResult } from 'simple-git';
import type { GitConfig, GitCommitResult, GitStatus } from '../../types/git';
import { GitCommitTemplates, GitDefaults, GitTimeouts } from './constants';
import { GitLockManager } from './git-lock-manager';

/**
 * Git 커밋 작업 전담 클래스
 *
 * 책임:
 * - 변경사항 커밋
 * - 커밋 메시지 템플릿 적용
 * - 체크포인트 생성
 * - Lock을 사용한 안전한 커밋 작업
 */
export class GitCommitManager {
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
        const readmePath = path.join(this.workingDir, 'README.md');
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
            : path.join(this.workingDir, file);
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
      throw new Error(`Failed to commit changes: ${(error as Error).message}`);
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
      throw new Error(
        `Failed to create checkpoint: ${(error as Error).message}`
      );
    }
  }

  /**
   * 저장소 상태 조회
   */
  async getStatus(): Promise<GitStatus> {
    try {
      // Git 상태를 명시적으로 갱신
      const status: StatusResult = await this.git.status();
      const branch = await this.git.branch();  // 브랜치 정보 명시적 조회

      return {
        clean: status.isClean(),
        currentBranch: branch.current,  // branchResult에서 직접 가져오기
        modified: status.modified,
        added: status.staged,
        deleted: status.deleted,
        untracked: status.not_added,
        ahead: status.ahead,
        behind: status.behind,
      };
    } catch (error) {
      throw new Error(`Failed to get status: ${(error as Error).message}`);
    }
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
   * Lock을 사용한 안전한 커밋 실행
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
   * 커밋 메시지 템플릿 적용
   */
  private applyCommitTemplate(message: string): string {
    if (!this.config.commitMessageTemplate) {
      return message;
    }

    return GitCommitTemplates.apply(this.config.commitMessageTemplate, message);
  }
}
