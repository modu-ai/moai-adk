# SPEC-GIT-001: Implementation Plan

---
spec_id: SPEC-GIT-001
type: plan
created: 2026-02-03
tags: git, go-git, implementation, plan
---

## 1. Implementation Strategy

### 1.1 Overall Approach

Git Operations Module은 **인터페이스 우선(Interface-First)** 접근 방식으로 구현한다. Go의 implicit interface satisfaction을 활용하여 소비자 코드가 인터페이스에만 의존하도록 설계하고, 구현체를 자유롭게 교체할 수 있도록 한다.

**핵심 원칙**:
- go-git을 1차 백엔드로 사용하고, 시스템 Git을 2차 폴백으로 사용한다
- 모든 public API는 인터페이스로 정의한다
- mockery를 통한 테스트 mock 자동 생성을 지원한다
- `context.Context`를 통한 취소와 timeout을 모든 장기 연산에 적용한다

### 1.2 go-git vs System Git Decision Tree

```
Operation Requested
    |
    v
[go-git 지원 여부?]
    |
    ├── YES: go-git으로 실행
    |   ├── Repository Open (PlainOpen)
    |   ├── CurrentBranch (Head reference)
    |   ├── Status (Worktree status)
    |   ├── Log (Commit iterator)
    |   ├── Diff (Tree diff)
    |   ├── IsClean (Status check)
    |   ├── Root (Repository path)
    |   ├── Branch Create (CreateBranch)
    |   ├── Branch Switch (Checkout)
    |   ├── Branch Delete (DeleteBranch)
    |   ├── Branch List (Branches iterator)
    |   └── MergeBase (Commit graph traversal)
    |
    └── NO: 시스템 Git 폴백
        |
        v
    [시스템 Git 존재 여부?]
        |
        ├── YES: os/exec로 실행
        |   ├── Worktree Add    → git worktree add <path> <branch>
        |   ├── Worktree List   → git worktree list --porcelain
        |   ├── Worktree Remove → git worktree remove <path>
        |   ├── Worktree Prune  → git worktree prune
        |   └── HasConflicts    → git merge-tree (complex cases)
        |
        └── NO: ErrSystemGitNotFound 반환
```

### 1.3 System Git Execution Pattern

```go
// 시스템 Git 호출 표준 패턴
func (m *manager) execGit(ctx context.Context, args ...string) (string, error) {
    gitPath, err := exec.LookPath("git")
    if err != nil {
        return "", fmt.Errorf("system git lookup: %w", ErrSystemGitNotFound)
    }

    cmd := exec.CommandContext(ctx, gitPath, args...)
    cmd.Dir = m.root
    cmd.Env = append(os.Environ(),
        "GIT_TERMINAL_PROMPT=0",  // 대화형 프롬프트 비활성화
        "LC_ALL=C",               // 일관된 출력 로케일
    )

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("git %s: %s: %w", args[0], stderr.String(), err)
    }

    return strings.TrimSpace(stdout.String()), nil
}
```

---

## 2. Milestones

### Milestone 1: Foundation (Primary Goal)

**목표**: Repository 인터페이스 구현 및 go-git 통합

**구현 범위**:
- `manager.go`: `gitManager` struct 정의, `NewRepository()` 생성자
- Repository 인터페이스의 6개 메서드 구현 (CurrentBranch, Status, Log, Diff, IsClean, Root)
- sentinel error 정의
- 시스템 Git 실행 헬퍼 (`execGit`)
- 구조화된 로깅 통합 (`log/slog`)

**구현 파일**:
| File | Contents | Estimated LOC |
|------|----------|---------------|
| manager.go | gitManager struct, Repository impl, execGit helper, errors | ~350 |

**테스트**:
- Unit tests: git.PlainInit으로 임시 저장소 생성 후 각 메서드 검증
- Table-driven tests: 다양한 저장소 상태 (clean, dirty, detached HEAD)
- 85% 이상 커버리지 목표

**완료 기준**:
- [ ] Repository 인터페이스 6개 메서드 구현 완료
- [ ] 모든 sentinel error 정의 완료
- [ ] execGit 헬퍼 구현 완료
- [ ] Unit test 85%+ 커버리지

---

### Milestone 2: Branch Management (Primary Goal)

**목표**: BranchManager 인터페이스 구현

**구현 범위**:
- `branch.go`: `branchManager` struct 정의
- Branch CRUD 연산 4개 (Create, Switch, Delete, List)
- 브랜치명 검증 로직 (Git ref naming rules)
- MoAI naming convention 적용 (`moai/` prefix 옵션)

**구현 파일**:
| File | Contents | Estimated LOC |
|------|----------|---------------|
| branch.go | branchManager struct, BranchManager impl, validation | ~300 |

**테스트**:
- Branch lifecycle test: create -> list -> switch -> delete
- 오류 케이스: 중복 브랜치, 존재하지 않는 브랜치, dirty working tree
- naming validation: 유효/무효 브랜치명 테이블 테스트

**완료 기준**:
- [ ] BranchManager 인터페이스 6개 메서드 구현 완료
- [ ] 브랜치명 검증 로직 구현
- [ ] MoAI naming convention 통합
- [ ] Unit test 85%+ 커버리지

---

### Milestone 3: Conflict Detection (Secondary Goal)

**목표**: Pre-merge 충돌 감지 구현

**구현 범위**:
- `conflict.go`: 충돌 감지 알고리즘
- merge-base 계산 (go-git commit graph traversal)
- tree-level diff를 통한 동시 수정 파일 감지
- read-only 보장 (working tree 미수정)

**구현 파일**:
| File | Contents | Estimated LOC |
|------|----------|---------------|
| conflict.go | conflictDetector, HasConflicts impl, MergeBase impl | ~250 |

**테스트**:
- 충돌 시나리오: 동일 파일 양쪽 수정, 한쪽만 수정, 신규 파일 충돌
- 안전성 검증: 감지 전후 working tree 상태 불변 확인
- merge-base 계산 정확성 검증

**완료 기준**:
- [ ] HasConflicts dry-run 감지 구현
- [ ] MergeBase 계산 구현
- [ ] Working tree 불변성 보장
- [ ] Unit test 85%+ 커버리지

---

### Milestone 4: Worktree Management (Secondary Goal)

**목표**: WorktreeManager 인터페이스 구현 (시스템 Git 폴백)

**구현 범위**:
- Worktree Add/List/Remove/Prune 구현
- 시스템 Git CLI 호출 및 출력 파싱
- `git worktree list --porcelain` 출력 파서
- context timeout 적용 (기본 10s)

**구현 위치**: `manager.go` 내 `worktreeManager` 구현 또는 별도 internal helper

**구현 파일**:
| File | Contents | Estimated LOC |
|------|----------|---------------|
| manager.go (확장) | worktreeManager impl, porcelain parser | ~150 |

**테스트**:
- Integration tests: 실제 시스템 Git을 사용한 worktree 생성/목록/제거
- 오류 케이스: Git 미설치, 경로 존재, dirty worktree
- `--porcelain` 출력 파싱 정확성 검증

**완료 기준**:
- [ ] WorktreeManager 인터페이스 4개 메서드 구현
- [ ] porcelain 출력 파서 구현
- [ ] context timeout 적용
- [ ] Integration test 통과

---

### Milestone 5: Event Detection (Optional Goal)

**목표**: Git event detection 및 hook 통합

**구현 범위**:
- `event.go`: 이벤트 감지 로직
- Git 상태 변화 감지 (브랜치 전환, 커밋 생성)
- 폴링 메커니즘 (statusline 통합용)
- `internal/hook/` 패키지와의 이벤트 인터페이스

**구현 파일**:
| File | Contents | Estimated LOC |
|------|----------|---------------|
| event.go | eventDetector, polling, event types | ~150 |

**테스트**:
- 이벤트 감지: 브랜치 전환 전후 상태 비교
- 폴링 테스트: context 취소 시 정상 종료 확인
- 이벤트 타입 정확성 검증

**완료 기준**:
- [ ] Event detection 기본 구현
- [ ] Polling mechanism 구현
- [ ] Hook integration 인터페이스 정의
- [ ] Unit test 80%+ 커버리지

---

## 3. Technical Approach

### 3.1 Package Structure

```go
package git

import (
    "context"
    "errors"
    "fmt"
    "log/slog"
    "os/exec"
    "strings"
    "time"

    gogit "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
    "github.com/go-git/go-git/v5/plumbing/object"
)
```

### 3.2 Manager Implementation Pattern

```go
// gitManager implements Repository, BranchManager, and WorktreeManager.
type gitManager struct {
    repo   *gogit.Repository
    root   string
    logger *slog.Logger
}

// NewRepository opens a Git repository at the given path.
// Implements lazy initialization pattern.
func NewRepository(path string, opts ...Option) (*gitManager, error) {
    repo, err := gogit.PlainOpen(path)
    if err != nil {
        return nil, fmt.Errorf("open repository at %s: %w", path, ErrNotRepository)
    }

    wt, err := repo.Worktree()
    if err != nil {
        return nil, fmt.Errorf("get worktree: %w", err)
    }

    return &gitManager{
        repo:   repo,
        root:   wt.Filesystem.Root(),
        logger: slog.Default().With("module", "git"),
    }, nil
}
```

### 3.3 Testing Strategy

| Test Type | Framework | Target | Coverage |
|-----------|-----------|--------|----------|
| Unit Tests | `testing` + `testify` | 개별 메서드 검증 | 85%+ |
| Integration Tests | `testing` + `*_integration_test.go` | go-git + 시스템 Git 통합 | Key paths |
| Table-Driven Tests | `testing` | 매개변수화된 입력/출력 | All methods |
| Mock Tests | `mockery` generated | 소비자 패키지 테스트 | Interfaces |
| Benchmark Tests | `testing.B` | 성능 요구사항 검증 | Critical paths |

**Test Fixtures**:
- `testdata/` 디렉토리에 사전 구성된 Git 저장소 스냅샷을 배치한다.
- `git.PlainInit()`으로 각 테스트에서 독립적인 임시 저장소를 생성한다.
- `t.Parallel()`을 사용하여 독립적인 테스트를 병렬 실행한다.

### 3.4 Mock Generation

```go
//go:generate mockery --name=Repository --output=./mocks --outpkg=mocks
//go:generate mockery --name=BranchManager --output=./mocks --outpkg=mocks
//go:generate mockery --name=WorktreeManager --output=./mocks --outpkg=mocks
```

소비자 패키지 (`internal/core/quality/`, `internal/statusline/` 등)에서 mock을 사용하여 git 모듈 없이 독립적으로 테스트한다.

---

## 4. Architecture Design Direction

### 4.1 Dependency Injection

```go
// 소비자 패키지에서의 사용 패턴
type QualityGate struct {
    repo git.Repository   // 인터페이스에만 의존
    // ...
}

func NewQualityGate(repo git.Repository) *QualityGate {
    return &QualityGate{repo: repo}
}
```

### 4.2 Constructor Wiring (cmd/moai/main.go)

```go
// Application wiring at startup
repo := gitpkg.NewRepository()
branchMgr := gitpkg.NewBranchManager(repo)
worktreeMgr := gitpkg.NewWorktreeManager(repo)
```

### 4.3 Concurrency Pattern

```go
// Parallel worktree operations
func (m *worktreeManager) AddMultiple(ctx context.Context, specs []WorktreeSpec) error {
    g, ctx := errgroup.WithContext(ctx)

    for _, spec := range specs {
        spec := spec
        g.Go(func() error {
            ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
            defer cancel()
            return m.Add(ctx, spec.Path, spec.Branch)
        })
    }

    return g.Wait()
}
```

---

## 5. Risk Analysis

### 5.1 Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| go-git API 변경 (v5.12+ 호환성) | Low | Medium | go.sum pinning, 통합 테스트로 조기 감지 |
| 시스템 Git 미설치 환경 | Medium | High | `moai doctor`에서 사전 검증, 적절한 오류 메시지 |
| Worktree porcelain 출력 형식 변경 | Low | Medium | Git 버전 확인, 파서 테스트 강화 |
| 대규모 저장소에서 성능 저하 | Medium | Medium | Benchmark 테스트, lazy loading, 캐싱 |
| Windows 경로 구분자 문제 | Medium | Low | `filepath.Clean()` 일관 적용, CI 매트릭스 |
| go-git/시스템 Git 동작 불일치 | Low | High | 통합 테스트에서 양쪽 결과 비교 검증 |

### 5.2 Dependency Risks

| Dependency | Risk | Mitigation |
|-----------|------|------------|
| SPEC-CONFIG-001 미완성 | `GitStrategyConfig` 접근 불가 | 기본값 하드코딩 후 config 연결 시 교체 |
| go-git v5 maintenance | 프로젝트 유지보수 중단 가능성 | 인터페이스 추상화로 대체 구현 가능 |
| pkg/utils/ | 유틸리티 패키지 미구현 | 필요 함수만 git 패키지 내부에 임시 구현 |

### 5.3 Mitigation Actions

**MA-001**: Integration test suite 구축
- 실제 Git 저장소를 사용한 end-to-end 테스트
- go-git과 시스템 Git 결과의 일관성 검증
- CI에서 macOS, Linux, Windows 매트릭스 실행

**MA-002**: Benchmark test suite 구축
- 각 연산의 성능 baseline 측정
- 10,000+ 파일 저장소에서의 성능 검증
- 성능 회귀 감지를 위한 CI 통합

**MA-003**: Graceful degradation
- 시스템 Git 미설치 시 worktree 연산만 비활성화
- 나머지 기능은 go-git만으로 정상 동작
- `moai doctor`에서 시스템 Git 존재 여부 보고

---

## 6. LOC Estimation

| File | Estimated LOC | Description |
|------|---------------|-------------|
| manager.go | ~350 | Repository impl, execGit, worktree routing |
| branch.go | ~300 | BranchManager impl, naming validation |
| conflict.go | ~250 | Conflict detection, merge-base |
| event.go | ~150 | Event detection, polling |
| errors.go | ~50 | Sentinel error definitions |
| doc.go | ~20 | Package documentation |
| **Total** | **~1,120** | |
| manager_test.go | ~400 | Repository + Worktree tests |
| branch_test.go | ~350 | BranchManager tests |
| conflict_test.go | ~250 | Conflict detection tests |
| event_test.go | ~150 | Event detection tests |
| **Test Total** | **~1,150** | |
| **Grand Total** | **~2,270** | Production + Test code |
