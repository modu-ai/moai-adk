// @DATA:GIT-TYPES-001 | Chain: @REQ:GIT-001 -> @DESIGN:GIT-001 -> @TASK:GIT-001 -> @TEST:GIT-001
// Related: @FEATURE:GIT-001

/**
 * @file Git operations type definitions
 * @author MoAI Team
 */

/**
 * Git 저장소 초기화 결과
 */
export interface GitInitResult {
  readonly success: boolean;
  readonly repositoryPath: string;
  readonly gitDir: string;
  readonly defaultBranch: string;
  readonly message?: string;
}

/**
 * Git 저장소 상태 정보
 */
export interface GitStatus {
  readonly clean: boolean;
  readonly modified: readonly string[];
  readonly added: readonly string[];
  readonly deleted: readonly string[];
  readonly untracked: readonly string[];
  readonly currentBranch: string;
  readonly ahead: number;
  readonly behind: number;
}

/**
 * Git 커밋 결과
 */
export interface GitCommitResult {
  readonly hash: string;
  readonly message: string;
  readonly timestamp: Date;
  readonly filesChanged: number;
  readonly author: string;
}

/**
 * Git 설정 정보
 */
export interface GitConfig {
  readonly mode: 'personal' | 'team';
  readonly autoCommit: boolean;
  readonly branchPrefix: string;
  readonly commitMessageTemplate: string;
  readonly remote?: {
    readonly name: string;
    readonly url: string;
  };
  readonly github?: {
    readonly token?: string;
    readonly owner?: string;
    readonly repo?: string;
  };
}

/**
 * 브랜치 정보
 */
export interface GitBranch {
  readonly name: string;
  readonly current: boolean;
  readonly remote?: string;
  readonly upstream?: string;
  readonly ahead: number;
  readonly behind: number;
}

/**
 * Git 리모트 정보
 */
export interface GitRemote {
  readonly name: string;
  readonly url: string;
  readonly type: 'fetch' | 'push';
}

/**
 * GitHub Pull Request 생성 옵션
 */
export interface CreatePullRequestOptions {
  readonly title: string;
  readonly body: string;
  readonly baseBranch: string;
  readonly headBranch?: string;
  readonly draft?: boolean;
  readonly assignees?: readonly string[];
  readonly reviewers?: readonly string[];
  readonly labels?: readonly string[];
}

/**
 * GitHub 저장소 생성 옵션
 */
export interface CreateRepositoryOptions {
  readonly name: string;
  readonly description?: string;
  readonly private: boolean;
  readonly autoInit?: boolean;
  readonly gitignoreTemplate?: string;
  readonly licenseTemplate?: string;
}

/**
 * Git 체크포인트 정보
 */
export interface GitCheckpoint {
  readonly hash: string;
  readonly message: string;
  readonly timestamp: Date;
  readonly tag?: string;
  readonly branch: string;
}

/**
 * Git 작업 진행률 콜백
 */
export type GitProgressCallback = (progress: {
  phase: string;
  loaded: number;
  total: number;
  percent: number;
}) => void;

/**
 * Git 에러 타입
 */
export type GitErrorType =
  | 'REPOSITORY_NOT_FOUND'
  | 'BRANCH_NOT_FOUND'
  | 'MERGE_CONFLICT'
  | 'PERMISSION_DENIED'
  | 'NETWORK_ERROR'
  | 'AUTHENTICATION_FAILED'
  | 'INVALID_REMOTE'
  | 'WORKING_TREE_DIRTY'
  | 'UNKNOWN_ERROR';

/**
 * Git 에러 정보
 */
export interface GitError extends Error {
  readonly type: GitErrorType;
  readonly code?: string;
  readonly details?: Record<string, unknown>;
}

/**
 * Git 명령어 실행 옵션
 */
export interface GitCommandOptions {
  readonly cwd?: string;
  readonly timeout?: number;
  readonly onProgress?: GitProgressCallback;
  readonly force?: boolean;
  readonly dryRun?: boolean;
}

/**
 * .gitignore 템플릿 타입
 */
export type GitignoreTemplate = 'node' | 'python' | 'moai' | 'custom';

/**
 * Git 로그 엔트리
 */
export interface GitLogEntry {
  readonly hash: string;
  readonly shortHash: string;
  readonly message: string;
  readonly author: string;
  readonly authorEmail: string;
  readonly date: Date;
  readonly filesChanged: readonly string[];
  readonly insertions: number;
  readonly deletions: number;
}

/**
 * Git lock information interface
 * @tags @DESIGN:GIT-LOCK-001
 */
export interface GitLockInfo {
  pid: number;
  timestamp: number;
  operation: string;
  user: string;
  hostname?: string;
  workingDir?: string;
}

/**
 * Git lock status interface
 * @tags @DESIGN:GIT-LOCK-STATUS-001
 */
export interface GitLockStatus {
  isLocked: boolean;
  lockFileExists: boolean;
  lockInfo: GitLockInfo | null;
  processRunning?: boolean;
  lockAgeSeconds?: number;
}

/**
 * Git lock context interface for context manager pattern
 * @tags @DESIGN:GIT-LOCK-CONTEXT-001
 */
export interface GitLockContext {
  readonly lockInfo: GitLockInfo;
  readonly acquired: Date;
  release(): Promise<void>;
}

/**
 * Git lock exception class
 * @tags @DESIGN:GIT-LOCK-EXCEPTION-001
 */
export class GitLockedException extends Error {
  constructor(
    message: string,
    public readonly lockInfo?: GitLockInfo,
    public readonly timeout?: number
  ) {
    super(message);
    this.name = 'GitLockedException';
  }
}
