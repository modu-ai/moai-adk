# SPEC-001: git-manager.ts 리팩토링

## @TAG BLOCK

```text
# @CODE:REFACTOR-001 | Chain: @SPEC:REFACTOR-001 -> @SPEC:REFACTOR-001 -> @CODE:REFACTOR-001 -> @TEST:REFACTOR-001
# Related: @CODE:GIT-001:API, @CODE:GIT-CFG-001:DATA
```

## Environment (환경 및 가정사항)

### 현재 환경
- **파일**: `src/core/git/git-manager.ts`
- **현재 LOC**: 689 라인
- **목표 LOC**: 각 파일 300 라인 이하
- **초과율**: 230% (개발 가이드 권장 기준 대비)

### 기술 스택
- TypeScript 5.x
- simple-git 라이브러리
- Vitest (테스트 프레임워크)
- fs-extra (파일 시스템 작업)

### 전제 조건
- 기존 테스트 스위트가 존재함
- 외부 API(GitHubIntegration)와의 의존성이 있음
- GitLockManager와의 통합이 필요함
- Personal/Team 모드를 지원해야 함

## Assumptions (전제 조건)

### 리팩토링 원칙
- 기존 API 호환성 100% 유지
- 모든 기존 테스트가 통과해야 함
- 단일 책임 원칙(SRP) 준수
- 순환 의존성 제거
- @TAG 추적성 유지

### 설계 결정
- 오케스트레이터 패턴 적용 (GitManager가 중앙 조정자 역할)
- 도메인별 매니저 분리 (브랜치, 커밋, PR)
- 공통 유틸리티는 별도 모듈로 분리하지 않음 (중복 코드 최소화)

## Requirements (기능 요구사항)

### EARS 형식 요구사항

#### Ubiquitous Requirements (기본 요구사항)
- 시스템은 git-manager.ts를 300 LOC 이하의 2-3개 모듈로 분리해야 한다
- 시스템은 기존 API 호환성을 100% 유지해야 한다
- 시스템은 모든 기존 테스트가 통과하도록 보장해야 한다
- 시스템은 단일 책임 원칙(SRP)을 준수해야 한다

#### Event-driven Requirements (이벤트 기반)
- WHEN 파일이 300 LOC를 초과하면, 시스템은 추가 분리를 제안해야 한다
- WHEN 리팩토링 후, 시스템은 자동으로 전체 테스트 스위트를 실행해야 한다
- WHEN 순환 의존성이 감지되면, 시스템은 빌드를 실패시켜야 한다

#### State-driven Requirements (상태 기반)
- WHILE 리팩토링 진행 중일 때, 시스템은 기존 기능을 100% 유지해야 한다
- WHILE 각 Phase 완료 시, 시스템은 테스트 통과를 검증해야 한다
- WHILE Personal/Team 모드일 때, 시스템은 각 모드의 특화 기능을 제공해야 한다

#### Optional Features (선택적 기능)
- WHERE 성능 개선 기회가 발견되면, 시스템은 최적화를 적용할 수 있다
- WHERE 코드 중복이 3회 이상 발견되면, 공통 유틸리티로 추출할 수 있다

#### Constraints (제약사항)
- IF 리팩토링 결과 테스트가 실패하면, 시스템은 변경을 롤백해야 한다
- 각 분리된 모듈은 300 LOC 이하여야 한다
- 각 함수는 50 LOC 이하여야 한다
- 매개변수는 5개 이하여야 한다
- 순환 복잡도(Cyclomatic Complexity)는 10 이하여야 한다
- 테스트 커버리지는 85% 이상을 유지해야 한다

## Specifications (상세 명세)

### 1. 모듈 분리 전략

#### 1.1 GitManager (메인 오케스트레이터, ~150 LOC)
**@SPEC:REFACTOR-001-MANAGER**

**책임**:
- 전체 Git 작업의 중앙 조정자
- 설정 검증 및 관리
- 각 매니저 인스턴스 생성 및 위임
- 공통 에러 처리

**주요 메서드**:
```typescript
// 생성자 및 초기화
constructor(config: GitConfig, workingDir?: string)
private validateConfig(config: GitConfig): void
private createGitInstance(baseDir: string): SimpleGit

// 저장소 초기화
async initializeRepository(projectPath: string): Promise<GitInitResult>
async isValidRepository(): Promise<boolean>
async createGitignore(projectPath: string, template?: GitignoreTemplate): Promise<string>

// 위임 메서드 (각 매니저에게 작업 위임)
async createBranch(branchName: string, baseBranch?: string): Promise<void>
async commitChanges(message: string, files?: string[]): Promise<GitCommitResult>
async pushChanges(branch?: string, remote?: string): Promise<void>
async createPullRequest(options: CreatePullRequestOptions): Promise<string>
async getStatus(): Promise<GitStatus>

// 공통 유틸리티
async getCurrentBranch(): Promise<string>
private createGitError(type: GitErrorType, message: string): GitError
```

**의존성**:
- GitBranchManager
- GitCommitManager
- GitPRManager
- GitLockManager
- GitHubIntegration

---

#### 1.2 GitBranchManager (브랜치 관리, ~200 LOC)
**@SPEC:REFACTOR-001-BRANCH**

**책임**:
- 브랜치 생성 및 검증
- 브랜치 전환 및 삭제
- 브랜치 네이밍 규칙 검증
- 원격 저장소 연결

**주요 메서드**:
```typescript
constructor(git: SimpleGit, config: GitConfig, workingDir: string)

// 브랜치 작업
async createBranch(branchName: string, baseBranch?: string): Promise<void>
async createBranchWithLock(branchName: string, baseBranch?: string, lockManager: GitLockManager): Promise<void>
async getCurrentBranch(): Promise<string>

// 브랜치 검증
private validateBranchName(branchName: string): void

// 원격 저장소 관리
async linkRemoteRepository(repoUrl: string, remoteName?: string): Promise<void>
private isValidGitUrl(url: string): boolean

// 초기 커밋 처리
private async ensureInitialCommit(): Promise<void>
```

**보안 강화**:
- InputValidator를 통한 브랜치명 검증
- GitNamingRules 준수 확인
- SQL Injection, Path Traversal 방지

---

#### 1.3 GitCommitManager (커밋/푸시 관리, ~200 LOC)
**@SPEC:REFACTOR-001-COMMIT**

**책임**:
- 파일 스테이징 및 커밋
- 커밋 메시지 템플릿 적용
- 체크포인트 생성
- 원격 저장소 푸시

**주요 메서드**:
```typescript
constructor(git: SimpleGit, config: GitConfig, workingDir: string)

// 커밋 작업
async commitChanges(message: string, files?: string[]): Promise<GitCommitResult>
async commitWithLock(message: string, files?: string[], lockManager: GitLockManager): Promise<GitCommitResult>
async createCheckpoint(message: string): Promise<string>

// 커밋 템플릿
private applyCommitTemplate(message: string): string

// 푸시 작업
async pushChanges(branch?: string, remote?: string): Promise<void>
async pushWithLock(branch?: string, remote?: string, lockManager: GitLockManager): Promise<void>

// 파일 검증
private async validateFiles(files: string[]): Promise<void>

// 초기 커밋 처리
private async ensureInitialCommit(): Promise<void>
```

**성능 최적화**:
- 배치 작업 지원
- 파일 존재 여부 캐싱

---

#### 1.4 GitPRManager (PR 및 저장소 관리, ~150 LOC)
**@SPEC:REFACTOR-001-PR**

**책임**:
- Pull Request 생성 (Team 모드)
- GitHub 저장소 생성 (Team 모드)
- GitHub CLI 통합
- GitHub 인증 상태 확인

**주요 메서드**:
```typescript
constructor(git: SimpleGit, config: GitConfig, githubIntegration?: GitHubIntegration)

// PR 작업 (Team 모드 전용)
async createPullRequest(options: CreatePullRequestOptions): Promise<string>

// 저장소 작업 (Team 모드 전용)
async createRepository(options: CreateRepositoryOptions): Promise<void>

// GitHub CLI 통합
async isGitHubCliAvailable(): Promise<boolean>
async isGitHubAuthenticated(): Promise<boolean>

// 모드 검증
private ensureTeamMode(): void
```

**모드별 분기**:
- Personal 모드: PR 기능 비활성화
- Team 모드: GitHub 연동 활성화

---

### 2. 의존성 관계

```
GitManager (오케스트레이터)
    ├── GitBranchManager (브랜치 관리)
    ├── GitCommitManager (커밋/푸시 관리)
    ├── GitPRManager (PR/저장소 관리)
    ├── GitLockManager (동시성 제어)
    └── GitHubIntegration (GitHub API)

공통 의존성:
    ├── SimpleGit (Git CLI 래퍼)
    ├── GitConfig (설정)
    ├── InputValidator (입력 검증)
    └── Constants (상수)
```

**순환 의존성 방지**:
- GitManager가 매니저들을 생성하고 조정
- 매니저들은 서로를 직접 참조하지 않음
- 필요한 경우 GitManager를 통해 간접 호출

---

### 3. 기존 API 호환성

모든 기존 Public API는 GitManager에서 동일한 시그니처로 유지됩니다:

```typescript
// 변경 전 (현재)
await gitManager.createBranch('feature/new-feature');
await gitManager.commitChanges('feat: add new feature', ['file.ts']);
await gitManager.pushChanges();
await gitManager.createPullRequest({ title: 'PR Title', body: 'PR Body' });

// 변경 후 (리팩토링)
// ✅ 동일한 API 유지 (내부적으로 각 매니저에게 위임)
await gitManager.createBranch('feature/new-feature');
await gitManager.commitChanges('feat: add new feature', ['file.ts']);
await gitManager.pushChanges();
await gitManager.createPullRequest({ title: 'PR Title', body: 'PR Body' });
```

---

### 4. 테스트 전략

#### 4.1 기존 테스트 유지
- `src/core/git/__tests__/git-manager.test.ts` 전체 통과 필수
- 테스트 커버리지 85% 이상 유지

#### 4.2 새 매니저별 테스트 추가
- `git-branch-manager.test.ts`: 브랜치 관리 단위 테스트
- `git-commit-manager.test.ts`: 커밋/푸시 단위 테스트
- `git-pr-manager.test.ts`: PR 관리 단위 테스트

#### 4.3 통합 테스트
- 전체 워크플로우 시나리오 테스트
- Personal/Team 모드별 테스트

---

### 5. 성공 기준

#### 정량적 지표
- ✅ git-manager.ts: 689 LOC → 150 LOC (78% 감소)
- ✅ 각 새 파일: 300 LOC 이하
- ✅ 함수당: 50 LOC 이하
- ✅ 복잡도: 10 이하
- ✅ 테스트 커버리지: 85% 이상

#### 정성적 지표
- ✅ 단일 책임 원칙(SRP) 준수
- ✅ 순환 의존성 없음
- ✅ 기존 API 100% 호환
- ✅ 모든 기존 테스트 통과
- ✅ @TAG 체인 무결성 유지

---

## Traceability (추적성)

### TAG 체인 구성
```text
@SPEC:REFACTOR-001 (요구사항 정의)
  └─> @SPEC:REFACTOR-001 (전체 설계)
       ├─> @SPEC:REFACTOR-001-MANAGER (GitManager 설계)
       ├─> @SPEC:REFACTOR-001-BRANCH (GitBranchManager 설계)
       ├─> @SPEC:REFACTOR-001-COMMIT (GitCommitManager 설계)
       └─> @SPEC:REFACTOR-001-PR (GitPRManager 설계)
            └─> @CODE:REFACTOR-001 (구현 작업)
                 └─> @TEST:REFACTOR-001 (테스트 검증)
```

### 관련 TAG
- @CODE:GIT-001 (기존 Git 기능)
- @CODE:GIT-001:API (Git API)
- @CODE:GIT-CFG-001:DATA (Git 설정)

---

## 참고 자료

- MoAI-ADK 개발 가이드: `.moai/memory/development-guide.md`
- TRUST 5원칙: Test First, Readable, Unified, Secured, Trackable
- EARS 방법론: Easy Approach to Requirements Syntax
- Git Manager 원본: `src/core/git/git-manager.ts`
