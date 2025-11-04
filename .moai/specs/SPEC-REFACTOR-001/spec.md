---
id: REFACTOR-001
version: 0.2.0
status: completed
created: 2025-09-15
updated: 2025-10-06
completed: 2025-10-06
author: @goos
priority: high
---

# @SPEC:REFACTOR-001: Git Manager 리팩토링

## HISTORY

### v0.2.0 (2025-10-06)
- **COMPLETED**: Git Manager 리팩토링 구현 완료
- **AUTHOR**: @goos, @alfred
- **IMPLEMENTATION**: git-manager.ts → 3개 모듈로 분리 완료
  - git-branch-manager.ts
  - git-commit-manager.ts
  - git-pr-manager.ts
- **EVIDENCE**:
  - `b01403e docs(sync): Complete SPEC-INIT-001, REFACTOR-001, BRAND-001`
  - `16263b3 refactor(tags): Unify TAG chain for SPEC-REFACTOR-001`
  - @CODE:REFACTOR-001:* TAG 체인 구축 완료

### v0.1.0 (2025-09-15)
- **INITIAL**: Git Manager 리팩토링 명세 작성
- **AUTHOR**: @goos
- **SCOPE**: 689 LOC → <300 LOC 분리
- **BACKGROUND**: TRUST 원칙 중 Readable (가독성) 준수 필요

## TAG BLOCK

```text
# @SPEC:REFACTOR-001: Git Manager 리팩토링
# Related: @CODE:REFACTOR-001:*, @TEST:REFACTOR-001:*
```

## Environment (환경 및 가정사항)

### 현재 상황

- **파일**: `src/core/git/git-manager.ts`
- **현재 LOC**: 689 라인
- **개발 가이드 권장**: 300 LOC 이하
- **초과율**: 230% (389 라인 초과)
- **현재 테스트**: `__tests__/core/git/git-manager.test.ts` 존재
- **의존성**:
  - `git-lock-manager.ts` (326 라인) - 이미 분리됨
  - `github-integration.ts` (346 라인) - 이미 분리됨
  - `constants.ts` - 상수 정의

### 프로젝트 환경

- **언어**: TypeScript
- **런타임**: Node.js 18.x+
- **패키지 매니저**: npm
- **테스트 도구**: Vitest
- **품질 도구**: Biome (린터+포매터)

## Assumptions (전제 조건)

### 기술적 전제

1. `git-lock-manager.ts`와 `github-integration.ts`는 이미 적절히 분리되어 있음
2. 기존 API 호환성을 100% 유지해야 함
3. 모든 기존 테스트는 수정 없이 통과해야 함
4. 순환 의존성이 없어야 함

### 비즈니스 전제

1. Git 작업의 안정성이 최우선
2. 리팩토링은 점진적으로 진행 (한 번에 하나의 책임씩 분리)
3. Team/Personal 모드 모두 지원 유지

## Requirements (기능 요구사항)

### Ubiquitous Requirements (기본 요구사항)

- 시스템은 git-manager.ts를 300 LOC 이하로 유지해야 한다
- 시스템은 단일 책임 원칙(SRP)을 준수해야 한다
- 시스템은 기존 API 호환성을 100% 보장해야 한다
- 시스템은 모든 기존 테스트가 통과하도록 해야 한다

### Event-driven Requirements (이벤트 기반)

- WHEN 파일이 300 LOC를 초과하면, 시스템은 추가 분리를 제안해야 한다
- WHEN 리팩토링이 완료되면, 시스템은 자동으로 전체 테스트를 실행해야 한다
- WHEN 브랜치 작업이 요청되면, 해당 모듈만 담당해야 한다
- WHEN 커밋 작업이 요청되면, 해당 모듈만 담당해야 한다
- WHEN PR 작업이 요청되면, 해당 모듈만 담당해야 한다

### State-driven Requirements (상태 기반)

- WHILE 리팩토링 진행 중일 때, 기존 기능은 계속 작동해야 한다
- WHILE 테스트 실행 중일 때, 모든 테스트는 독립적으로 실행되어야 한다

### Optional Features (선택적 기능)

- WHERE 성능 최적화가 필요하면, 배치 작업 지원을 강화할 수 있다
- WHERE 추가 Git 기능이 필요하면, 새로운 매니저 모듈을 추가할 수 있다

### Constraints (제약사항)

- IF 모듈이 300 LOC를 초과하면, 추가 분리가 필수다
- IF 순환 의존성이 발생하면, 설계를 재검토해야 한다
- IF 테스트 커버리지가 85% 미만이면, 테스트를 추가해야 한다
- 각 모듈의 복잡도는 10을 초과하지 않아야 한다
- 함수당 LOC는 50을 초과하지 않아야 한다
- 매개변수는 5개 이하여야 한다

## Specifications (상세 명세)

### 1. 모듈 분리 전략

#### 제안 구조

```
src/core/git/
├── git-manager.ts           (Main Orchestrator, ~150 LOC)
│   └── 역할: 전체 Git 작업 조율, 모듈 간 통합
│
├── git-branch-manager.ts    (Branch Operations, ~200 LOC)
│   └── 역할: 브랜치 생성, 전환, 삭제, 목록 조회
│
├── git-commit-manager.ts    (Commit & Push Operations, ~200 LOC)
│   └── 역할: 커밋, 스테이징, 푸시, 체크포인트
│
├── git-pr-manager.ts        (PR Operations, ~150 LOC)
│   └── 역할: PR 생성, 저장소 생성, GitHub 통합 래퍼
│
├── git-lock-manager.ts      (Lock Management, 326 LOC) [기존]
├── github-integration.ts    (GitHub API, 346 LOC) [기존]
└── constants.ts             (상수 정의) [기존]
```

### 2. 모듈별 책임 정의

#### git-manager.ts (Main Orchestrator)

**책임**: 전체 Git 작업의 진입점, 설정 관리, 모듈 간 통합

**주요 메서드**:
- `constructor(config, workingDir?)` - 설정 및 모듈 초기화
- `initializeRepository(projectPath)` - 저장소 초기화
- `getStatus()` - 저장소 상태 조회
- `isValidRepository()` - 저장소 유효성 확인
- `performBatchOperations(operations)` - 배치 작업 수행
- Private: `createGitInstance()`, `validateConfig()`

**의존성**:
- GitBranchManager
- GitCommitManager
- GitPRManager
- GitLockManager

**목표 LOC**: ~150 라인

---

#### git-branch-manager.ts (Branch Operations)

**책임**: 브랜치 관련 모든 작업

**주요 메서드**:
- `createBranch(name, baseBranch?)` - 브랜치 생성
- `createBranchWithLock(name, baseBranch?, wait?, timeout?)` - Lock과 함께 브랜치 생성
- `getCurrentBranch()` - 현재 브랜치 조회
- `listBranches()` - 브랜치 목록 조회
- `switchBranch(name)` - 브랜치 전환
- `deleteBranch(name, force?)` - 브랜치 삭제
- Private: `validateBranchName()`, `ensureInitialCommit()`

**의존성**:
- SimpleGit
- InputValidator
- GitLockManager

**목표 LOC**: ~200 라인

---

#### git-commit-manager.ts (Commit & Push Operations)

**책임**: 커밋, 스테이징, 푸시 관련 모든 작업

**주요 메서드**:
- `commitChanges(message, files?)` - 변경사항 커밋
- `commitWithLock(message, files?, wait?, timeout?)` - Lock과 함께 커밋
- `pushChanges(branch?, remote?)` - 원격 저장소로 푸시
- `pushWithLock(branch?, remote?, wait?, timeout?)` - Lock과 함께 푸시
- `createCheckpoint(message)` - 체크포인트 생성
- `stageFiles(files)` - 파일 스테이징
- Private: `applyCommitTemplate()`, `ensureInitialCommit()`

**의존성**:
- SimpleGit
- GitLockManager
- GitCommitTemplates

**목표 LOC**: ~200 라인

---

#### git-pr-manager.ts (PR Operations)

**책임**: Pull Request 및 GitHub 통합 작업

**주요 메서드**:
- `createPullRequest(options)` - PR 생성
- `createRepository(options)` - GitHub 저장소 생성
- `linkRemoteRepository(url, remoteName?)` - 원격 저장소 연결
- `isGitHubCliAvailable()` - GitHub CLI 확인
- `isGitHubAuthenticated()` - GitHub 인증 확인
- `createGitignore(projectPath, template?)` - .gitignore 생성
- Private: `isValidGitUrl()`, `getGitignoreTemplate()`

**의존성**:
- SimpleGit
- GitHubIntegration
- GitignoreTemplates

**목표 LOC**: ~150 라인

---

### 3. API 호환성 전략

#### 기존 API 유지

`git-manager.ts`는 기존 모든 public 메서드를 유지하되, 내부적으로 각 매니저에 위임:

```typescript
export class GitManager {
  private branchManager: GitBranchManager;
  private commitManager: GitCommitManager;
  private prManager: GitPRManager;

  // 기존 API - 위임 패턴
  async createBranch(name: string, baseBranch?: string): Promise<void> {
    return this.branchManager.createBranch(name, baseBranch);
  }

  async commitChanges(message: string, files?: string[]): Promise<GitCommitResult> {
    return this.commitManager.commitChanges(message, files);
  }

  async createPullRequest(options: CreatePullRequestOptions): Promise<string> {
    return this.prManager.createPullRequest(options);
  }

  // ... 나머지 메서드들도 동일하게 위임
}
```

### 4. 테스트 전략

#### 기존 테스트 유지

- `__tests__/core/git/git-manager.test.ts` - 수정 없이 통과해야 함
- 통합 테스트 관점 유지

#### 신규 테스트 추가

```
__tests__/core/git/
├── git-manager.test.ts              [기존 - 통합 테스트]
├── git-branch-manager.test.ts       [신규 - 단위 테스트]
├── git-commit-manager.test.ts       [신규 - 단위 테스트]
├── git-pr-manager.test.ts           [신규 - 단위 테스트]
└── git-lock-manager.test.ts         [기존]
```

### 5. 마이그레이션 절차

#### Phase 1: git-branch-manager.ts 분리
1. 브랜치 관련 메서드 추출
2. GitBranchManager 클래스 생성
3. 단위 테스트 작성 및 통과
4. git-manager.ts에서 위임 패턴 적용
5. 기존 통합 테스트 통과 확인

#### Phase 2: git-commit-manager.ts 분리
1. 커밋/푸시 관련 메서드 추출
2. GitCommitManager 클래스 생성
3. 단위 테스트 작성 및 통과
4. git-manager.ts에서 위임 패턴 적용
5. 기존 통합 테스트 통과 확인

#### Phase 3: git-pr-manager.ts 분리
1. PR/GitHub 관련 메서드 추출
2. GitPRManager 클래스 생성
3. 단위 테스트 작성 및 통과
4. git-manager.ts에서 위임 패턴 적용
5. 기존 통합 테스트 통과 확인

#### Phase 4: 최종 검증
1. 모든 테스트 실행 (통합 + 단위)
2. 테스트 커버리지 확인 (≥85%)
3. Biome 린트/포맷 통과
4. LOC 검증 (각 파일 ≤300)
5. 복잡도 검증 (각 함수 ≤10)

### 6. 품질 게이트

#### 코드 품질

- 각 파일 ≤ 300 LOC
- 각 함수 ≤ 50 LOC
- 매개변수 ≤ 5개
- 순환 복잡도 ≤ 10

#### 테스트 품질

- 전체 커버리지 ≥ 85%
- 모든 기존 테스트 통과 (0 실패)
- 신규 단위 테스트 추가 (각 매니저별)

#### 성능 기준

- 기존 성능 유지 (성능 저하 없음)
- Lock 기반 작업의 동시성 보장

## Traceability (추적성)

### TAG 체인

```
@SPEC:REFACTOR-001 (요구사항)
  └─> @SPEC:REFACTOR-001 (설계)
        └─> @CODE:REFACTOR-001 (작업)
              └─> @TEST:REFACTOR-001 (검증)
```

### 관련 TAG

- `@CODE:REFACTOR-001` - 리팩토링 기능
- `@CODE:GIT-MGR-001` - Git Manager API
- `@CODE:GIT-001` - Git 데이터 타입
- `@TEST:REFACTOR-001` - 리팩토링 테스트

### 기존 TAG 유지

모든 기존 TAG는 새로운 모듈로 이동하되, 추적성 유지:

- `@CODE:GIT-001` → 유지 및 확장
- `@SPEC:GIT-001` → 참조 유지
- `@SPEC:GIT-001` → 업데이트

## 성공 지표

### 정량적 지표

- git-manager.ts: 689 LOC → ~150 LOC (78% 감소)
- 모듈화: 1개 파일 → 4개 모듈 (책임 분리)
- 테스트 커버리지: 현재 수준 이상 유지 (≥85%)
- 테스트 실패: 0건

### 정성적 지표

- 단일 책임 원칙 준수
- 코드 가독성 향상
- 유지보수성 개선
- 확장성 확보

## 리스크 및 완화 방안

### 리스크 1: API 호환성 깨짐

**완화 방안**:
- 위임 패턴으로 기존 API 100% 유지
- 기존 테스트 수정 없이 통과 확인

### 리스크 2: 순환 의존성 발생

**완화 방안**:
- 각 매니저는 GitManager에 의존하지 않음
- 공통 타입은 types/git.ts에 정의
- 의존성 방향: GitManager → *Manager (단방향)

### 리스크 3: 성능 저하

**완화 방안**:
- 위임은 단순 메서드 호출로 오버헤드 최소
- Lock 매커니즘 유지로 동시성 보장
- 성능 테스트 추가

### 리스크 4: 테스트 복잡도 증가

**완화 방안**:
- 통합 테스트는 기존 유지
- 단위 테스트는 각 매니저별 독립 작성
- Mock 객체로 의존성 격리

## 다음 단계

1. `/alfred:2-run SPEC-001` - TDD 구현 시작
2. Phase별 순차 진행 (Branch → Commit → PR)
3. 각 Phase 완료 후 테스트 검증
4. `/alfred:3-sync` - 문서 동기화 및 TAG 검증

---

**작성일**: 2025-10-01
**버전**: 1.0
**상태**: Draft
