# SPEC-001: 구현 계획

## @TAG BLOCK

```text
# @CODE:REFACTOR-001 | Chain: @SPEC:REFACTOR-001 -> @SPEC:REFACTOR-001 -> @CODE:REFACTOR-001
# Related: @CODE:REFACTOR-001
```

## 우선순위별 마일스톤

### 1차 목표: 핵심 매니저 분리 (High Priority)

**목표**: GitManager를 4개 모듈로 분리하고 기본 동작 확인

**작업 항목**:

#### 1.1 GitBranchManager 분리
- [ ] `src/core/git/git-branch-manager.ts` 파일 생성
- [ ] 브랜치 관련 메서드 이동
  - `createBranch()`
  - `createBranchWithLock()`
  - `getCurrentBranch()`
  - `linkRemoteRepository()`
- [ ] 브랜치명 검증 로직 통합
- [ ] 초기 커밋 처리 로직 추가
- [ ] 단위 테스트 작성: `git-branch-manager.test.ts`

**의존성**:
- SimpleGit 인스턴스 주입
- GitConfig 주입
- InputValidator 사용
- GitNamingRules 사용

---

#### 1.2 GitCommitManager 분리
- [ ] `src/core/git/git-commit-manager.ts` 파일 생성
- [ ] 커밋/푸시 관련 메서드 이동
  - `commitChanges()`
  - `commitWithLock()`
  - `pushChanges()`
  - `pushWithLock()`
  - `createCheckpoint()`
- [ ] 커밋 템플릿 적용 로직 이동
- [ ] 파일 검증 로직 추가
- [ ] 단위 테스트 작성: `git-commit-manager.test.ts`

**의존성**:
- SimpleGit 인스턴스 주입
- GitConfig 주입
- GitCommitTemplates 사용
- fs-extra 사용

---

#### 1.3 GitPRManager 분리
- [ ] `src/core/git/git-pr-manager.ts` 파일 생성
- [ ] PR/저장소 관련 메서드 이동
  - `createPullRequest()`
  - `createRepository()`
  - `isGitHubCliAvailable()`
  - `isGitHubAuthenticated()`
- [ ] Team 모드 검증 로직 추가
- [ ] 단위 테스트 작성: `git-pr-manager.test.ts`

**의존성**:
- SimpleGit 인스턴스 주입
- GitConfig 주입
- GitHubIntegration 주입 (선택적)

---

#### 1.4 GitManager 리팩토링
- [ ] 기존 메서드를 위임 메서드로 변경
- [ ] 각 매니저 인스턴스 생성 로직 추가
- [ ] 공통 설정 및 에러 처리 유지
- [ ] 저장소 초기화 로직 유지
- [ ] 기존 테스트 통과 확인

**목표 LOC**: 150 라인 이하

---

### 2차 목표: 테스트 및 품질 검증 (High Priority)

**목표**: 모든 테스트 통과 및 코드 품질 기준 충족

**작업 항목**:

#### 2.1 기존 테스트 검증
- [ ] `git-manager.test.ts` 전체 테스트 실행
- [ ] 실패한 테스트 디버깅 및 수정
- [ ] 테스트 커버리지 측정 (85% 이상 목표)

#### 2.2 새 매니저별 단위 테스트
- [ ] `git-branch-manager.test.ts` 작성
  - 브랜치 생성 시나리오
  - 브랜치명 검증 테스트
  - 원격 저장소 연결 테스트
  - Lock 통합 테스트
- [ ] `git-commit-manager.test.ts` 작성
  - 커밋 생성 시나리오
  - 파일 검증 테스트
  - 푸시 동작 테스트
  - 체크포인트 생성 테스트
- [ ] `git-pr-manager.test.ts` 작성
  - PR 생성 테스트 (Team 모드)
  - 저장소 생성 테스트 (Team 모드)
  - GitHub CLI 통합 테스트
  - Personal 모드 제한 테스트

#### 2.3 통합 테스트
- [ ] Personal 모드 전체 워크플로우 테스트
- [ ] Team 모드 전체 워크플로우 테스트
- [ ] 에러 시나리오 테스트
- [ ] 동시성 제어(Lock) 테스트

---

### 3차 목표: 코드 품질 최적화 (Medium Priority)

**목표**: TRUST 원칙 준수 및 코드 품질 향상

**작업 항목**:

#### 3.1 정적 분석
- [ ] ESLint/Biome 규칙 준수 확인
- [ ] TypeScript 타입 체크 (strict 모드)
- [ ] 순환 의존성 검사
- [ ] 복잡도 분석 (각 함수 ≤ 10)

#### 3.2 코드 리뷰 체크리스트
- [ ] 단일 책임 원칙(SRP) 준수
- [ ] 함수당 50 LOC 이하
- [ ] 매개변수 5개 이하
- [ ] 의도 드러내는 이름 사용
- [ ] 가드절 우선 사용

#### 3.3 문서화
- [ ] JSDoc 주석 추가
- [ ] @TAG 체인 업데이트
- [ ] README 업데이트 (필요 시)

---

### 최종 목표: 문서 동기화 및 배포 (Low Priority)

**목표**: 리팩토링 결과 문서화 및 팀 공유

**작업 항목**:

#### 4.1 문서 동기화
- [ ] `/alfred:3-sync` 실행하여 TAG 검증
- [ ] API 문서 갱신 (필요 시)
- [ ] 변경 로그 작성

#### 4.2 Git 작업
- [ ] 브랜치 생성: `feature/SPEC-001-refactor-git-manager`
- [ ] 커밋: "refactor: git-manager를 4개 모듈로 분리"
- [ ] Draft PR 생성 (Team 모드)

---

## 기술적 접근 방법

### 1. 오케스트레이터 패턴 적용

**GitManager가 중앙 조정자 역할**:

```typescript
export class GitManager {
  private branchManager: GitBranchManager;
  private commitManager: GitCommitManager;
  private prManager: GitPRManager;
  private lockManager: GitLockManager;

  constructor(config: GitConfig, workingDir?: string) {
    this.validateConfig(config);
    this.config = config;
    this.currentWorkingDir = workingDir || process.cwd();
    this.git = this.createGitInstance(this.currentWorkingDir);

    // 각 매니저 인스턴스 생성
    this.branchManager = new GitBranchManager(this.git, this.config, this.currentWorkingDir);
    this.commitManager = new GitCommitManager(this.git, this.config, this.currentWorkingDir);
    this.prManager = new GitPRManager(this.git, this.config, this.githubIntegration);
    this.lockManager = new GitLockManager(this.currentWorkingDir);
  }

  // 위임 메서드
  async createBranch(branchName: string, baseBranch?: string): Promise<void> {
    return this.branchManager.createBranch(branchName, baseBranch);
  }

  async commitChanges(message: string, files?: string[]): Promise<GitCommitResult> {
    return this.commitManager.commitChanges(message, files);
  }

  async createPullRequest(options: CreatePullRequestOptions): Promise<string> {
    return this.prManager.createPullRequest(options);
  }
}
```

---

### 2. 의존성 주입 전략

**SimpleGit 인스턴스 공유**:

```typescript
// GitManager가 SimpleGit 인스턴스를 생성하고 각 매니저에 주입
constructor(config: GitConfig, workingDir?: string) {
  this.git = this.createGitInstance(this.currentWorkingDir);

  // 동일한 git 인스턴스를 각 매니저에 주입
  this.branchManager = new GitBranchManager(this.git, this.config, this.currentWorkingDir);
  this.commitManager = new GitCommitManager(this.git, this.config, this.currentWorkingDir);
  this.prManager = new GitPRManager(this.git, this.config, this.githubIntegration);
}
```

**GitLockManager 통합**:

```typescript
// Lock이 필요한 작업은 각 매니저의 WithLock 메서드 호출
async createBranchWithLock(branchName: string, baseBranch?: string): Promise<void> {
  return this.lockManager.withLock(
    () => this.branchManager.createBranch(branchName, baseBranch),
    `create-branch: ${branchName}`
  );
}
```

---

### 3. 에러 처리 전략

**공통 에러 생성**:

```typescript
// GitManager에서 createGitError 메서드 유지
private createGitError(type: GitErrorType, message: string): GitError {
  const error = new Error(message) as GitError;
  (error as any).type = type;
  return error;
}

// 각 매니저는 에러를 throw하고 GitManager가 catch
async createBranch(branchName: string, baseBranch?: string): Promise<void> {
  try {
    return await this.branchManager.createBranch(branchName, baseBranch);
  } catch (error) {
    throw this.createGitError('BRANCH_NOT_FOUND', (error as Error).message);
  }
}
```

---

### 4. 초기 커밋 처리 통합

**GitBranchManager와 GitCommitManager 공통 로직**:

```typescript
// 초기 커밋이 없는 경우 README.md 생성
private async ensureInitialCommit(): Promise<void> {
  const status = await this.git.status();
  if (status.files.length === 0) {
    const readmePath = path.join(this.workingDir, 'README.md');
    if (!(await fs.pathExists(readmePath))) {
      await fs.writeFile(readmePath, '# Project\n\nInitial commit\n');
      await this.git.add('README.md');
      await this.git.commit('Initial commit');
    }
  }
}
```

---

## 아키텍처 설계 방향

### 레이어 구조

```
Orchestrator Layer (GitManager)
    ↓ 위임
Domain Layer (각 매니저)
    ├── GitBranchManager (브랜치 도메인)
    ├── GitCommitManager (커밋/푸시 도메인)
    └── GitPRManager (PR/저장소 도메인)
    ↓ 사용
Infrastructure Layer (SimpleGit, GitLockManager, GitHubIntegration)
```

---

### 책임 분리

| 레이어 | 책임 | 클래스 |
|--------|------|--------|
| **Orchestrator** | 전체 조정, 설정 검증, 공통 에러 처리 | GitManager |
| **Domain** | 도메인별 비즈니스 로직 | GitBranchManager, GitCommitManager, GitPRManager |
| **Infrastructure** | Git CLI 래퍼, 동시성 제어, GitHub API | SimpleGit, GitLockManager, GitHubIntegration |

---

## 리스크 및 대응 방안

### 리스크 1: 기존 API 호환성 깨짐

**영향도**: 높음 (기존 코드 전체 영향)

**대응 방안**:
- GitManager의 Public API 시그니처 100% 유지
- 기존 테스트 스위트 전체 통과 필수
- 통합 테스트로 전체 워크플로우 검증

---

### 리스크 2: 순환 의존성 발생

**영향도**: 중간 (빌드 실패 가능)

**대응 방안**:
- 매니저들은 서로를 직접 참조하지 않음
- GitManager를 통한 간접 호출
- ESLint 플러그인으로 순환 의존성 자동 검출

---

### 리스크 3: 성능 저하

**영향도**: 낮음 (추가 레이어로 인한 오버헤드)

**대응 방안**:
- SimpleGit 인스턴스 공유로 오버헤드 최소화
- 캐싱 로직 유지 (repositoryInfoCache)
- 성능 벤치마크 테스트 추가

---

### 리스크 4: Lock 통합 복잡도 증가

**영향도**: 중간 (동시성 제어 오류 가능)

**대응 방안**:
- WithLock 메서드 패턴 표준화
- Lock 타임아웃 설정 명확화
- 동시성 테스트 강화

---

## 완료 조건 (Definition of Done)

### 필수 조건
- [ ] git-manager.ts가 150 LOC 이하로 감소
- [ ] 각 새 파일이 300 LOC 이하
- [ ] 모든 기존 테스트 통과
- [ ] 테스트 커버리지 85% 이상
- [ ] 순환 의존성 없음
- [ ] ESLint/Biome 규칙 준수
- [ ] TypeScript strict 모드 통과

### 선택 조건
- [ ] 성능 벤치마크 통과 (기존 대비 5% 이내)
- [ ] 코드 리뷰 승인
- [ ] 문서 동기화 완료

---

## 다음 단계

리팩토링 완료 후:

1. `/alfred:3-sync` 실행하여 TAG 검증
2. Draft PR 생성 (브랜치: `feature/SPEC-001-refactor-git-manager`)
3. 팀 리뷰 요청
4. 승인 후 develop 브랜치로 머지

---

## 참고 자료

- 개발 가이드: `.moai/memory/development-guide.md`
- TRUST 5원칙: Test First, Readable, Unified, Secured, Trackable
- 리팩토링 패턴: Martin Fowler's Refactoring Catalog
- Git Manager 원본: `src/core/git/git-manager.ts`
