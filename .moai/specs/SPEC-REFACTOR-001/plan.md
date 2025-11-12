# @SPEC:001 구현 계획

## TAG BLOCK

```text
# @SPEC:REFACTOR-001 | Chain: @SPEC:REFACTOR-001 -> @SPEC:REFACTOR-001 -> @CODE:REFACTOR-001 -> @TEST:REFACTOR-001
# Related: @CODE:REFACTOR-001
```

## 우선순위별 마일스톤

### 1차 목표: Git Branch Manager 분리

**우선순위**: High

**목표**:
- git-branch-manager.ts 생성 (200 LOC 이하)
- 브랜치 관련 모든 로직 이동
- 단위 테스트 작성 및 통과
- git-manager.ts에서 위임 패턴 적용

**주요 작업**:
1. GitBranchManager 클래스 구조 설계
2. 브랜치 생성/전환/삭제 메서드 이동
3. 브랜치명 검증 로직 통합
4. Lock 통합 메서드 구현
5. 단위 테스트 작성 (커버리지 ≥85%)

**의존성**:
- SimpleGit 인스턴스 전달 필요
- InputValidator 통합
- GitLockManager 연동

**완료 조건**:
- git-branch-manager.ts ≤ 200 LOC
- 모든 단위 테스트 통과
- 기존 통합 테스트 통과
- Biome 린트 통과

---

### 2차 목표: Git Commit Manager 분리

**우선순위**: High

**목표**:
- git-commit-manager.ts 생성 (200 LOC 이하)
- 커밋/푸시 관련 모든 로직 이동
- 단위 테스트 작성 및 통과
- git-manager.ts에서 위임 패턴 적용

**주요 작업**:
1. GitCommitManager 클래스 구조 설계
2. 커밋/스테이징/푸시 메서드 이동
3. 커밋 템플릿 적용 로직 이동
4. 체크포인트 생성 기능 이동
5. Lock 통합 메서드 구현
6. 단위 테스트 작성 (커버리지 ≥85%)

**의존성**:
- SimpleGit 인스턴스 전달 필요
- GitCommitTemplates 통합
- GitLockManager 연동

**완료 조건**:
- git-commit-manager.ts ≤ 200 LOC
- 모든 단위 테스트 통과
- 기존 통합 테스트 통과
- Biome 린트 통과

---

### 3차 목표: Git PR Manager 분리

**우선순위**: Medium

**목표**:
- git-pr-manager.ts 생성 (150 LOC 이하)
- PR/GitHub 관련 모든 로직 이동
- 단위 테스트 작성 및 통과
- git-manager.ts에서 위임 패턴 적용

**주요 작업**:
1. GitPRManager 클래스 구조 설계
2. PR 생성/저장소 생성 메서드 이동
3. GitHub CLI 연동 로직 이동
4. .gitignore 생성 기능 이동
5. 원격 저장소 연결 로직 이동
6. 단위 테스트 작성 (커버리지 ≥85%)

**의존성**:
- SimpleGit 인스턴스 전달 필요
- GitHubIntegration 연동
- GitignoreTemplates 통합

**완료 조건**:
- git-pr-manager.ts ≤ 150 LOC
- 모든 단위 테스트 통과
- 기존 통합 테스트 통과
- Biome 린트 통과

---

### 4차 목표: Git Manager 최종 정리

**우선순위**: High

**목표**:
- git-manager.ts 슬림화 (150 LOC 이하)
- 위임 패턴 완성
- 전체 통합 테스트 통과
- 문서화 및 TAG 검증

**주요 작업**:
1. git-manager.ts에서 모든 구현 제거
2. 순수 오케스트레이터 역할로 변경
3. 생성자에서 모든 매니저 초기화
4. 위임 메서드 구현 완료
5. 설정 검증 및 SimpleGit 인스턴스 관리
6. 배치 작업 및 상태 조회 유지

**완료 조건**:
- git-manager.ts ≤ 150 LOC
- 모든 통합 테스트 통과 (0 실패)
- 모든 단위 테스트 통과
- 테스트 커버리지 ≥ 85%
- Biome 린트 통과
- TAG 체인 검증 완료

---

### 최종 목표: 품질 검증 및 문서화

**우선순위**: High

**목표**:
- 전체 시스템 품질 검증
- 문서 동기화
- 성능 테스트
- TAG 추적성 확인

**주요 작업**:
1. 전체 테스트 스위트 실행
2. 코드 커버리지 보고서 생성
3. 성능 벤치마크 실행
4. 코드 복잡도 검증
5. TAG 체인 스캔 및 검증
6. API 문서 업데이트
7. 개발 가이드 업데이트

**완료 조건**:
- 모든 품질 게이트 통과
- 문서 동기화 완료 (`/alfred:3-sync`)
- TAG 무결성 확인
- 성능 저하 없음

---

## 기술적 접근 방법

### 1. 위임 패턴 (Delegation Pattern)

**개념**:
- GitManager는 인터페이스 역할만 수행
- 실제 구현은 각 전문 매니저가 담당
- 느슨한 결합으로 확장성 확보

**구현 방식**:

```typescript
export class GitManager {
  private git: SimpleGit;
  private config: GitConfig;
  private branchManager: GitBranchManager;
  private commitManager: GitCommitManager;
  private prManager: GitPRManager;
  private lockManager: GitLockManager;

  constructor(config: GitConfig, workingDir?: string) {
    this.validateConfig(config);
    this.config = config;
    this.git = this.createGitInstance(workingDir || process.cwd());

    // 각 매니저 초기화
    this.lockManager = new GitLockManager(workingDir);
    this.branchManager = new GitBranchManager(this.git, this.lockManager);
    this.commitManager = new GitCommitManager(this.git, this.lockManager, config);
    this.prManager = new GitPRManager(this.git, config);
  }

  // 위임 메서드
  async createBranch(name: string, baseBranch?: string): Promise<void> {
    return this.branchManager.createBranch(name, baseBranch);
  }

  async commitChanges(message: string, files?: string[]): Promise<GitCommitResult> {
    return this.commitManager.commitChanges(message, files);
  }

  // ... 나머지 메서드들도 동일하게 위임
}
```

**장점**:
- API 호환성 100% 유지
- 각 매니저의 독립적인 테스트 가능
- 책임 분리로 유지보수성 향상

---

### 2. 의존성 주입 (Dependency Injection)

**개념**:
- 각 매니저는 필요한 의존성을 생성자로 주입받음
- 테스트 시 Mock 객체로 대체 가능
- 순환 의존성 방지

**구현 방식**:

```typescript
export class GitBranchManager {
  constructor(
    private git: SimpleGit,
    private lockManager: GitLockManager
  ) {}

  async createBranch(name: string, baseBranch?: string): Promise<void> {
    // SimpleGit 사용
    // LockManager 사용
  }
}

export class GitCommitManager {
  constructor(
    private git: SimpleGit,
    private lockManager: GitLockManager,
    private config: GitConfig
  ) {}

  async commitChanges(message: string, files?: string[]): Promise<GitCommitResult> {
    // SimpleGit 사용
    // Config 사용 (템플릿)
    // LockManager 사용
  }
}
```

**장점**:
- 테스트 용이성
- 느슨한 결합
- 재사용성 향상

---

### 3. TDD 사이클 (Red-Green-Refactor)

**Phase 1 단계별 TDD**:

1. **RED**: 실패하는 테스트 작성
   ```typescript
   // git-branch-manager.test.ts
   describe('GitBranchManager', () => {
     it('should create a new branch', async () => {
       // Given: Git 저장소와 브랜치명
       // When: createBranch 호출
       // Then: 브랜치 생성 확인
     });
   });
   ```

2. **GREEN**: 테스트를 통과하는 최소 코드 작성
   ```typescript
   // git-branch-manager.ts
   export class GitBranchManager {
     async createBranch(name: string): Promise<void> {
       await this.git.checkoutLocalBranch(name);
     }
   }
   ```

3. **REFACTOR**: 코드 품질 개선
   - 입력 검증 추가
   - 에러 처리 추가
   - 로깅 추가
   - 주석 및 문서화

**각 Phase마다 반복**:
- Phase 1 (Branch) → RED-GREEN-REFACTOR
- Phase 2 (Commit) → RED-GREEN-REFACTOR
- Phase 3 (PR) → RED-GREEN-REFACTOR
- Phase 4 (Manager) → RED-GREEN-REFACTOR

---

### 4. 점진적 마이그레이션

**전략**:
1. 한 번에 하나의 책임만 분리
2. 분리 후 즉시 테스트 검증
3. 통합 테스트로 회귀 방지
4. 각 단계별 커밋으로 롤백 가능

**단계별 체크리스트**:

**Step 1**: 새 매니저 클래스 생성
- [ ] 클래스 구조 정의
- [ ] 생성자 구현
- [ ] 의존성 주입 설정

**Step 2**: 메서드 이동
- [ ] Private 헬퍼 메서드 이동
- [ ] Public 메서드 이동
- [ ] 입력 검증 로직 이동

**Step 3**: 테스트 작성
- [ ] 단위 테스트 작성
- [ ] Edge case 테스트
- [ ] 에러 케이스 테스트

**Step 4**: 통합
- [ ] GitManager에 매니저 추가
- [ ] 위임 메서드 구현
- [ ] 기존 테스트 실행

**Step 5**: 검증
- [ ] 테스트 커버리지 확인
- [ ] 린트 통과
- [ ] LOC 검증
- [ ] 복잡도 검증

---

## 아키텍처 설계 방향

### 계층 구조

```
┌─────────────────────────────────────────┐
│         GitManager (Orchestrator)       │
│  - 설정 관리                            │
│  - SimpleGit 인스턴스 생성              │
│  - API 인터페이스 제공                  │
└────────────┬─────────┬─────────┬────────┘
             │         │         │
    ┌────────▼────┐ ┌──▼──────┐ ┌▼────────┐
    │   Branch    │ │ Commit  │ │   PR    │
    │   Manager   │ │ Manager │ │ Manager │
    └─────┬───────┘ └────┬────┘ └────┬────┘
          │              │            │
          │    ┌─────────▼────────────▼─────┐
          └────►    GitLockManager           │
               │   (Concurrent Safety)       │
               └─────────────────────────────┘
                          │
               ┌──────────▼──────────┐
               │  GitHubIntegration  │
               │   (Team Mode Only)  │
               └─────────────────────┘
```

### 의존성 방향

- **단방향 의존성**: 상위 → 하위만 가능
- **순환 방지**: 매니저 간 직접 의존 금지
- **공통 타입**: types/git.ts에 중앙 집중

### 인터페이스 설계

```typescript
// 각 매니저의 공통 인터페이스 패턴
interface GitOperationManager {
  // SimpleGit 인스턴스 접근
  readonly git: SimpleGit;

  // Lock 기반 안전 작업
  readonly lockManager: GitLockManager;

  // 설정 접근 (필요한 경우)
  readonly config?: GitConfig;
}
```

---

## 리스크 및 대응 방안

### 리스크 매트릭스

| 리스크 | 확률 | 영향도 | 대응 전략 |
|--------|------|--------|-----------|
| API 호환성 깨짐 | 낮음 | 높음 | 위임 패턴, 기존 테스트 유지 |
| 순환 의존성 발생 | 중간 | 높음 | 단방향 의존성 강제, 아키텍처 리뷰 |
| 성능 저하 | 낮음 | 중간 | 성능 테스트, 벤치마크 비교 |
| 테스트 복잡도 증가 | 중간 | 낮음 | 단위 테스트 독립화, Mock 활용 |
| 마이그레이션 중 버그 | 중간 | 중간 | 점진적 마이그레이션, 각 단계별 검증 |

### 대응 계획

#### 리스크 1: API 호환성
- **예방**: 위임 패턴으로 모든 기존 메서드 유지
- **탐지**: 기존 테스트 자동 실행
- **복구**: Git 롤백 (각 Phase별 커밋)

#### 리스크 2: 순환 의존성
- **예방**: 의존성 그래프 작성 및 리뷰
- **탐지**: TypeScript 컴파일러 경고
- **복구**: 공통 타입 추출, 인터페이스 분리

#### 리스크 3: 성능 저하
- **예방**: 단순 위임으로 오버헤드 최소화
- **탐지**: 성능 벤치마크 실행
- **복구**: 핫스팟 최적화, 캐싱 추가

---

## 테스트 전략

### 테스트 피라미드

```
        ┌────────────┐
        │ E2E Tests  │  (소수)
        ├────────────┤
        │Integration │  (중간)
        │   Tests    │
        ├────────────┤
        │   Unit     │  (다수)
        │   Tests    │
        └────────────┘
```

### 테스트 커버리지 목표

- **전체**: ≥ 85%
- **각 매니저**: ≥ 90%
- **Critical Path**: 100%

### 테스트 케이스 분류

#### 1. 단위 테스트 (Unit Tests)

**각 매니저별**:
- 정상 케이스 (Happy Path)
- Edge 케이스 (경계값)
- 에러 케이스 (예외 처리)
- Lock 통합 케이스

**예시**:
```typescript
describe('GitBranchManager', () => {
  describe('createBranch', () => {
    it('should create branch with valid name', async () => {});
    it('should reject invalid branch name', async () => {});
    it('should handle existing branch', async () => {});
    it('should create branch with base branch', async () => {});
    it('should use lock when requested', async () => {});
  });
});
```

#### 2. 통합 테스트 (Integration Tests)

**기존 테스트 유지**:
- `git-manager.test.ts` - 수정 없이 통과해야 함
- 전체 워크플로우 검증

**신규 통합 테스트**:
- 여러 매니저 간 협력 시나리오
- Lock 경쟁 상황 테스트

#### 3. E2E 테스트 (End-to-End Tests)

**실제 사용 시나리오**:
- 브랜치 생성 → 커밋 → 푸시 → PR 생성
- Team/Personal 모드별 전체 흐름

---

## 성능 최적화 고려사항

### 1. 인스턴스 재사용

**현재**:
- SimpleGit 인스턴스 1개 (GitManager)

**리팩토링 후**:
- SimpleGit 인스턴스 1개 (GitManager에서 생성, 각 매니저에 전달)
- Lock 인스턴스 1개 (공유)

### 2. 캐싱 전략

**유지**:
- Repository 정보 캐싱 (5초)
- Lock 상태 캐싱

**개선 가능**:
- 브랜치 목록 캐싱 (필요 시)
- 현재 브랜치 캐싱 강화

### 3. 배치 작업

**유지**:
- `performBatchOperations()` 메서드
- 여러 Git 작업을 순차적으로 실행

**개선 가능**:
- 각 매니저별 배치 작업 최적화

---

## 문서화 계획

### 코드 문서화

**각 파일 헤더**:
```typescript
// @CODE:<ID> | Chain: @SPEC:<ID> -> @SPEC:<ID> -> @CODE:<ID> -> @TEST:<ID>
// Related: @CODE:<ID>, @CODE:<ID>

/**
 * @file Brief description
 * @author MoAI Team
 * @tags @CODE:<ID>
 */
```

**메서드 문서화**:
- JSDoc 주석
- 매개변수 설명
- 반환값 설명
- 예외 설명
- 사용 예시

### API 문서

**생성 도구**: TypeDoc

**내용**:
- 각 매니저의 public API
- 사용 예시 코드
- 마이그레이션 가이드

### 개발 가이드 업데이트

**업데이트 항목**:
- 새로운 아키텍처 설명
- 각 매니저의 책임 설명
- 확장 가이드 (새 매니저 추가 방법)

---

## 다음 단계

1. **Phase 1 시작**: GitBranchManager 구현
   - `@agent-code-builder` 호출하여 TDD 구현 시작
   - RED → GREEN → REFACTOR 사이클 준수

2. **단계별 검증**: 각 Phase 완료 후
   - 테스트 실행 및 통과 확인
   - 코드 품질 검증
   - 커밋 생성

3. **최종 검증**: 모든 Phase 완료 후
   - `/alfred:3-sync` 실행
   - TAG 체인 검증
   - 문서 동기화

---

**작성일**: 2025-10-01
**버전**: 1.0
**상태**: Draft
