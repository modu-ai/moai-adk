# SPEC-GIT-001: Git Operations Module

---
spec_id: SPEC-GIT-001
title: Git Operations Module
status: Completed
priority: High
phase: "Phase 2 - Core Domains"
module: internal/core/git/
files:
  - manager.go
  - branch.go
  - conflict.go
  - event.go
estimated_loc: ~1,200
dependencies:
  - SPEC-CONFIG-001
created: 2026-02-03
lifecycle: spec-anchored
tags: git, go-git, branch, worktree, conflict, repository
---

## HISTORY

| Date | Version | Author | Description |
|------|---------|--------|-------------|
| 2026-02-03 | 1.0.0 | GOOS | Initial SPEC creation |

---

## 1. Environment

### 1.1 Project Context

MoAI-ADK (Go Edition)는 Python 기반 MoAI-ADK (~73,000 LOC)를 순수 Go로 재작성하는 프로젝트이다. Git Operations Module은 `internal/core/git/` 패키지로, go-git 라이브러리를 통한 순수 Go Git 연산과 시스템 Git 바이너리 폴백을 결합하여 브랜치 관리, 충돌 감지, worktree 관리, 이벤트 감지 기능을 제공한다.

### 1.2 Technical Environment

| Component | Specification |
|-----------|---------------|
| Language | Go 1.22+ |
| Module Path | `github.com/modu-ai/moai-adk-go` |
| Primary Library | `github.com/go-git/go-git/v5` v5.12+ |
| System Fallback | `git` binary (2.30+) via `os/exec` |
| Logging | `log/slog` (stdlib) |
| Concurrency | `errgroup.Group`, `context.Context` |
| Testing | `testing` (stdlib) + `github.com/stretchr/testify` v1.9+ |
| Mock Generation | `github.com/vektra/mockery` v2.46+ |

### 1.3 Architecture Context

```
internal/core/git/
    ├── manager.go      # Repository interface implementation, go-git/system git routing
    ├── branch.go       # BranchManager interface implementation
    ├── conflict.go     # Pre-merge conflict detection
    └── event.go        # Git event detection and hook integration
```

**Dependency Graph**:
```
git/ ------------> pkg/models/
                -> pkg/utils/
```

**Dependents** (git/ 모듈을 사용하는 패키지):
```
internal/core/project/   -> internal/core/git/
internal/core/quality/   -> internal/core/git/
internal/core/integration/ -> internal/core/git/
internal/loop/           -> internal/core/git/
internal/statusline/     -> internal/core/git/
internal/rank/           -> internal/core/git/
```

### 1.4 Design Decision (ADR-007)

go-git은 순수 Go 구현으로 CGo 의존성 없이 기본 Git 연산을 제공한다. 그러나 worktree 지원이 제한적이므로 시스템 Git으로 폴백한다. 오류 래핑을 통해 백엔드에 관계없이 일관된 오류 타입을 보장한다.

---

## 2. Assumptions

### 2.1 Technical Assumptions

- [A-001] go-git v5.12+가 Repository open, branch CRUD, status, log, diff, IsClean 연산을 안정적으로 지원한다 (Confidence: High)
- [A-002] go-git의 worktree API는 `git worktree add/list/remove/prune`에 대해 불완전하므로 시스템 Git 폴백이 필요하다 (Confidence: High, Evidence: ADR-007, go-git issue tracker)
- [A-003] 대상 시스템에 Git 2.30+ 바이너리가 PATH에 존재한다 (worktree 연산 시) (Confidence: High, Evidence: product.md External Dependencies)
- [A-004] SPEC-CONFIG-001이 완료되어 `internal/config/` 패키지를 통해 `GitStrategyConfig`에 접근 가능하다 (Confidence: Medium)
- [A-005] `CGO_ENABLED=0` 환경에서 go-git이 정상 동작한다 (Confidence: High, Evidence: go-git는 pure Go)

### 2.2 Business Assumptions

- [A-006] Python 전작의 Git 관련 기능(branch.py, branch_manager.py, conflict_detector.py, event_detector.py)과 기능 동등성(feature parity)을 유지해야 한다
- [A-007] MoAI naming convention (`moai/` prefix)을 브랜치 생성 시 적용해야 한다
- [A-008] Worktree 기능은 병렬 SPEC 개발 워크플로우의 핵심이다

### 2.3 Risk If Wrong

| Assumption | Risk | Mitigation |
|------------|------|------------|
| A-002 | go-git가 worktree를 완전 지원하면 시스템 Git 폴백 불필요 | 인터페이스 추상화로 구현 교체 용이 |
| A-003 | Git 미설치 시 worktree 연산 실패 | 적절한 오류 메시지와 `moai doctor` 검증 |
| A-004 | config 모듈 미완성 시 기본값 하드코딩 필요 | defaults fallback 구현 |

---

## 3. Requirements

### 3.1 Repository Interface Requirements

#### REQ-GIT-R001: Repository Open (Ubiquitous)
시스템은 **항상** 현재 작업 디렉토리 또는 지정된 경로에서 Git 저장소를 열 수 있어야 한다.

- go-git의 `git.PlainOpen()` 또는 `git.PlainOpenWithOptions()`를 사용한다.
- 저장소가 아닌 경우 `ErrNotRepository` 오류를 반환한다.
- 열린 저장소 인스턴스는 재사용 가능해야 한다 (lazy initialization).

#### REQ-GIT-R002: Current Branch (Event-Driven)
**WHEN** `CurrentBranch()`가 호출되면 **THEN** 현재 체크아웃된 브랜치 이름을 문자열로 반환해야 한다.

- HEAD가 detached 상태인 경우 빈 문자열과 `ErrDetachedHEAD` 오류를 반환한다.
- go-git의 `repo.Head()` reference를 파싱한다.

#### REQ-GIT-R003: Working Tree Status (Event-Driven)
**WHEN** `Status()`가 호출되면 **THEN** staged, modified, untracked 파일 목록과 ahead/behind 카운트를 포함하는 `*GitStatus` 구조체를 반환해야 한다.

- go-git의 `worktree.Status()` 결과를 `GitStatus` 구조체로 변환한다.
- ahead/behind 카운트는 upstream 추적 브랜치 대비 계산한다.
- upstream이 없는 경우 ahead/behind는 0으로 설정한다.

#### REQ-GIT-R004: Commit Log (Event-Driven)
**WHEN** `Log(n int)`이 호출되면 **THEN** HEAD로부터 최근 n개의 커밋을 `[]Commit` 슬라이스로 반환해야 한다.

- 각 `Commit`은 Hash, Author, Date, Message 필드를 포함한다.
- n이 총 커밋 수보다 큰 경우 가능한 모든 커밋을 반환한다.
- go-git의 `repo.Log()` iterator를 사용한다.

#### REQ-GIT-R005: Diff (Event-Driven)
**WHEN** `Diff(ref1, ref2 string)`이 호출되면 **THEN** 두 참조 간의 diff 출력을 문자열로 반환해야 한다.

- ref1, ref2는 브랜치명, 커밋 해시, 또는 HEAD~N 형태를 지원한다.
- go-git의 tree diff를 사용하여 unified diff 형식을 생성한다.

#### REQ-GIT-R006: Clean Check (Event-Driven)
**WHEN** `IsClean()`이 호출되면 **THEN** working tree에 uncommitted 변경사항이 없으면 `true`를, 있으면 `false`를 반환해야 한다.

- `Status()`를 내부적으로 호출하여 staged, modified, untracked 목록이 모두 비어있는지 확인한다.

#### REQ-GIT-R007: Repository Root (Ubiquitous)
시스템은 **항상** `Root()`를 통해 저장소 루트의 절대 경로를 반환해야 한다.

- `filepath.Clean()`을 적용한 정규화된 경로를 반환한다.
- 오류 없이 항상 문자열을 반환한다 (Repository가 유효하게 열린 상태에서).

---

### 3.2 BranchManager Interface Requirements

#### REQ-GIT-B001: Branch Create (Event-Driven)
**WHEN** `Create(name string)`이 호출되면 **THEN** 현재 HEAD에서 새 로컬 브랜치를 생성해야 한다.

- 이미 존재하는 브랜치명인 경우 `ErrBranchExists` 오류를 반환한다.
- 브랜치명 검증: Git ref 이름 규칙을 따라야 한다 (공백, `..`, `~`, `^`, `:` 금지).
- go-git의 `repo.CreateBranch()`를 사용한다.

#### REQ-GIT-B002: Branch Switch (Event-Driven)
**WHEN** `Switch(name string)`이 호출되면 **THEN** 지정된 브랜치로 체크아웃해야 한다.

- 존재하지 않는 브랜치인 경우 `ErrBranchNotFound` 오류를 반환한다.
- working tree가 clean하지 않은 경우 `ErrDirtyWorkingTree` 오류를 반환한다.
- go-git의 `worktree.Checkout()`을 사용한다.

#### REQ-GIT-B003: Branch Delete (Event-Driven)
**WHEN** `Delete(name string)`이 호출되면 **THEN** 지정된 로컬 브랜치를 삭제해야 한다.

- 현재 체크아웃된 브랜치는 삭제할 수 없다 (`ErrCannotDeleteCurrentBranch`).
- 원격 브랜치는 삭제하지 않는다 (로컬 전용).
- go-git의 `repo.DeleteBranch()`를 사용한다.

#### REQ-GIT-B004: Branch List (Event-Driven)
**WHEN** `List()`가 호출되면 **THEN** 모든 로컬 브랜치를 `[]Branch` 슬라이스로 반환해야 한다.

- 각 `Branch`는 Name, IsRemote, IsCurrent 필드를 포함한다.
- 현재 브랜치는 `IsCurrent: true`로 표시한다.
- go-git의 `repo.Branches()` iterator를 사용한다.

#### REQ-GIT-B005: Conflict Detection (Event-Driven)
**WHEN** `HasConflicts(target string)`이 호출되면 **THEN** 현재 브랜치와 대상 브랜치를 병합할 때 충돌이 발생하는지 여부를 `bool`로 반환해야 한다.

- 실제 병합을 수행하지 않고 dry-run 방식으로 감지한다.
- merge base를 찾아 양쪽의 변경사항을 비교한다.
- conflict.go에서 상세 구현한다.

#### REQ-GIT-B006: Merge Base (Event-Driven)
**WHEN** `MergeBase(branch1, branch2 string)`이 호출되면 **THEN** 두 브랜치의 공통 조상 커밋 해시를 문자열로 반환해야 한다.

- 공통 조상이 없는 경우 `ErrNoMergeBase` 오류를 반환한다.
- go-git의 commit object graph traversal을 사용한다.

---

### 3.3 WorktreeManager Interface Requirements

#### REQ-GIT-W001: Worktree Add (Event-Driven)
**WHEN** `Add(path, branch string)`이 호출되면 **THEN** 지정된 경로에 지정된 브랜치의 worktree를 생성해야 한다.

- **시스템 Git 폴백 사용**: `git worktree add <path> <branch>`를 `os/exec`로 실행한다.
- 경로가 이미 존재하면 `ErrWorktreePathExists` 오류를 반환한다.
- 브랜치가 존재하지 않으면 자동 생성한다 (`-b` 플래그).
- `context.Context`를 통한 timeout 지원 (기본 10s).

#### REQ-GIT-W002: Worktree List (Event-Driven)
**WHEN** `List()`가 호출되면 **THEN** 모든 활성 worktree를 `[]Worktree` 슬라이스로 반환해야 한다.

- **시스템 Git 폴백 사용**: `git worktree list --porcelain`을 파싱한다.
- 각 `Worktree`는 Path, Branch, HEAD 필드를 포함한다.
- 메인 worktree도 목록에 포함한다.

#### REQ-GIT-W003: Worktree Remove (Event-Driven)
**WHEN** `Remove(path string)`이 호출되면 **THEN** 지정된 worktree를 제거해야 한다.

- **시스템 Git 폴백 사용**: `git worktree remove <path>`를 실행한다.
- 변경사항이 있는 worktree는 `--force` 없이 제거하지 않는다 (`ErrWorktreeDirty`).
- 존재하지 않는 경로인 경우 `ErrWorktreeNotFound` 오류를 반환한다.

#### REQ-GIT-W004: Worktree Prune (Event-Driven)
**WHEN** `Prune()`이 호출되면 **THEN** stale worktree 참조를 정리해야 한다.

- **시스템 Git 폴백 사용**: `git worktree prune`을 실행한다.
- 삭제된 디렉토리나 이동된 worktree에 대한 참조를 정리한다.

---

### 3.4 Conflict Detection Requirements (conflict.go)

#### REQ-GIT-C001: Dry-Run Conflict Analysis (Event-Driven)
**WHEN** 충돌 감지가 요청되면 **THEN** 실제 병합 없이 tree-level diff를 통해 충돌 여부를 판단해야 한다.

- merge base commit을 기준으로 양쪽 브랜치의 변경 파일을 비교한다.
- 동일 파일이 양쪽에서 수정된 경우 충돌 가능성으로 판단한다.
- 충돌 파일 목록을 반환할 수 있는 확장 메서드를 지원한다.

#### REQ-GIT-C002: Conflict Safety (Unwanted)
시스템은 충돌 감지 과정에서 working tree를 수정**하지 않아야 한다**.

- read-only 연산만 수행한다.
- staging area, HEAD, working tree를 변경하지 않는다.

---

### 3.5 Event Detection Requirements (event.go)

#### REQ-GIT-E001: Git Event Detection (Event-Driven)
**WHEN** Git 상태 변경이 감지되면 **THEN** 해당 이벤트를 구조화된 형태로 보고해야 한다.

- 감지 대상 이벤트: 브랜치 전환, 커밋 생성, 병합, rebase
- 이벤트는 이전 상태와 현재 상태를 포함한다.
- hook 시스템(`internal/hook/`)과 통합을 위한 이벤트 인터페이스를 제공한다.

#### REQ-GIT-E002: Event Polling (State-Driven)
**IF** 이벤트 폴링 모드가 활성화된 상태이면 **THEN** 주기적으로 Git 상태를 확인하고 변경 시 이벤트를 발행해야 한다.

- statusline 통합을 위한 경량 폴링 메커니즘이다.
- 폴링 간격은 설정 가능하다 (기본값: 5초).
- `context.Context`를 통한 취소를 지원한다.

---

### 3.6 Cross-Cutting Requirements

#### REQ-GIT-X001: Error Wrapping (Ubiquitous)
시스템은 **항상** 모든 오류를 `fmt.Errorf("context: %w", err)` 패턴으로 래핑해야 한다.

- go-git 오류와 시스템 Git 오류 모두 일관된 오류 타입으로 래핑한다.
- 센티넬 오류를 정의한다: `ErrNotRepository`, `ErrDetachedHEAD`, `ErrBranchExists`, `ErrBranchNotFound`, `ErrDirtyWorkingTree`, `ErrCannotDeleteCurrentBranch`, `ErrNoMergeBase`, `ErrWorktreePathExists`, `ErrWorktreeDirty`, `ErrWorktreeNotFound`, `ErrSystemGitNotFound`.

#### REQ-GIT-X002: Context Propagation (Ubiquitous)
시스템은 **항상** 장기 실행 연산에 `context.Context`를 첫 번째 매개변수로 받아야 한다.

- worktree 연산, 원격 관련 연산, diff 생성에 context를 적용한다.
- timeout은 `context.WithTimeout`으로 설정한다 (worktree: 10s, 일반: 5s).

#### REQ-GIT-X003: System Git Fallback (State-Driven)
**IF** go-git이 특정 연산을 지원하지 않는 상태이면 **THEN** 시스템 Git 바이너리로 폴백해야 한다.

- 시스템 Git 경로를 `exec.LookPath("git")`로 검색한다.
- 시스템 Git이 없는 경우 `ErrSystemGitNotFound` 오류를 반환한다.
- 폴백 연산: worktree add/list/remove/prune, 복잡한 merge-base 계산.

#### REQ-GIT-X004: Structured Logging (Ubiquitous)
시스템은 **항상** `log/slog`를 통해 구조화된 로깅을 수행해야 한다.

- 모든 Git 연산의 시작과 완료를 debug 레벨로 로깅한다.
- 오류 발생 시 error 레벨로 context 정보와 함께 로깅한다.
- 시스템 Git 폴백 시 info 레벨로 폴백 사유를 로깅한다.

#### REQ-GIT-X005: Concurrency Safety (Ubiquitous)
시스템은 **항상** goroutine-safe하게 동작해야 한다.

- Repository 인스턴스는 여러 goroutine에서 안전하게 읽기 연산을 수행할 수 있어야 한다.
- 쓰기 연산 (branch create/switch/delete)은 동시 호출 시 적절한 동기화를 보장해야 한다.
- worktree 연산은 `errgroup.Group`을 통한 병렬 실행을 지원해야 한다.

#### REQ-GIT-X006: Performance (Ubiquitous)
시스템은 **항상** 다음 성능 요구사항을 충족해야 한다.

| Operation | Target Latency |
|-----------|---------------|
| CurrentBranch | < 10ms |
| Status | < 50ms |
| Log(10) | < 30ms |
| IsClean | < 50ms |
| Branch Create/Switch/Delete | < 100ms |
| Worktree Add | < 10s |
| Worktree List | < 100ms |

---

## 4. Specifications

### 4.1 Data Structures

```go
// GitStatus holds the working tree state.
type GitStatus struct {
    Staged    []string
    Modified  []string
    Untracked []string
    Ahead     int
    Behind    int
}

// Commit represents a Git commit record.
type Commit struct {
    Hash    string
    Author  string
    Date    time.Time
    Message string
}

// Branch represents a Git branch.
type Branch struct {
    Name      string
    IsRemote  bool
    IsCurrent bool
}

// Worktree represents a Git worktree entry.
type Worktree struct {
    Path   string
    Branch string
    HEAD   string
}
```

### 4.2 Interface Definitions

```go
// Repository provides read operations on a Git repository.
type Repository interface {
    CurrentBranch() (string, error)
    Status() (*GitStatus, error)
    Log(n int) ([]Commit, error)
    Diff(ref1, ref2 string) (string, error)
    IsClean() (bool, error)
    Root() string
}

// BranchManager provides branch lifecycle operations.
type BranchManager interface {
    Create(name string) error
    Switch(name string) error
    Delete(name string) error
    List() ([]Branch, error)
    HasConflicts(target string) (bool, error)
    MergeBase(branch1, branch2 string) (string, error)
}

// WorktreeManager manages Git worktrees for parallel development.
type WorktreeManager interface {
    Add(path, branch string) error
    List() ([]Worktree, error)
    Remove(path string) error
    Prune() error
}
```

### 4.3 Sentinel Errors

```go
var (
    ErrNotRepository            = errors.New("git: not a git repository")
    ErrDetachedHEAD             = errors.New("git: HEAD is detached")
    ErrBranchExists             = errors.New("git: branch already exists")
    ErrBranchNotFound           = errors.New("git: branch not found")
    ErrDirtyWorkingTree         = errors.New("git: working tree has uncommitted changes")
    ErrCannotDeleteCurrentBranch = errors.New("git: cannot delete currently checked-out branch")
    ErrNoMergeBase              = errors.New("git: no common ancestor found")
    ErrWorktreePathExists       = errors.New("git: worktree path already exists")
    ErrWorktreeDirty            = errors.New("git: worktree has uncommitted changes")
    ErrWorktreeNotFound         = errors.New("git: worktree not found")
    ErrSystemGitNotFound        = errors.New("git: system git binary not found")
)
```

### 4.4 Configuration Integration

```go
// GitStrategyConfig (from SPEC-CONFIG-001)
type GitStrategyConfig struct {
    AutoBranch   bool   `yaml:"auto_branch" default:"false"`
    BranchPrefix string `yaml:"branch_prefix" default:"moai/"`
    CommitStyle  string `yaml:"commit_style" default:"conventional"`
    WorktreeRoot string `yaml:"worktree_root"`
}
```

### 4.5 File Responsibilities

| File | Primary Responsibility | Interfaces Implemented |
|------|----------------------|----------------------|
| `manager.go` | Repository open, go-git/system-git routing, lazy init | `Repository` |
| `branch.go` | Branch CRUD, naming validation | `BranchManager` |
| `conflict.go` | Tree-level conflict detection, merge-base calculation | (conflict analysis helpers) |
| `event.go` | Git state change detection, event emission | (event detection helpers) |

### 4.6 Traceability

| Requirement | File | Test |
|-------------|------|------|
| REQ-GIT-R001 | manager.go | TestRepositoryOpen |
| REQ-GIT-R002 | manager.go | TestCurrentBranch |
| REQ-GIT-R003 | manager.go | TestStatus |
| REQ-GIT-R004 | manager.go | TestLog |
| REQ-GIT-R005 | manager.go | TestDiff |
| REQ-GIT-R006 | manager.go | TestIsClean |
| REQ-GIT-R007 | manager.go | TestRoot |
| REQ-GIT-B001 | branch.go | TestBranchCreate |
| REQ-GIT-B002 | branch.go | TestBranchSwitch |
| REQ-GIT-B003 | branch.go | TestBranchDelete |
| REQ-GIT-B004 | branch.go | TestBranchList |
| REQ-GIT-B005 | conflict.go | TestHasConflicts |
| REQ-GIT-B006 | conflict.go | TestMergeBase |
| REQ-GIT-W001 | manager.go (system git) | TestWorktreeAdd |
| REQ-GIT-W002 | manager.go (system git) | TestWorktreeList |
| REQ-GIT-W003 | manager.go (system git) | TestWorktreeRemove |
| REQ-GIT-W004 | manager.go (system git) | TestWorktreePrune |
| REQ-GIT-C001 | conflict.go | TestDryRunConflictAnalysis |
| REQ-GIT-C002 | conflict.go | TestConflictSafety |
| REQ-GIT-E001 | event.go | TestEventDetection |
| REQ-GIT-E002 | event.go | TestEventPolling |
| REQ-GIT-X001~X006 | (all files) | (cross-cutting tests) |

---

## Implementation Notes

**Status**: Completed
**Implementation Date**: 2026-02-03
**Development Methodology**: Hybrid (TDD for new code, DDD for existing code)
**Test Coverage**: 88.1%

### Summary

Git operations package implemented with dual-backend architecture: go-git for in-process repository operations and system git fallback for worktree management. Includes branch CRUD, conflict detection with merge-base calculation, event-driven state change detection with polling, and worktree add/list/remove/prune operations. All operations support context-based cancellation and timeout.

### Files Created

- `internal/core/git/branch.go`
- `internal/core/git/branch_test.go`
- `internal/core/git/conflict.go`
- `internal/core/git/conflict_test.go`
- `internal/core/git/doc.go`
- `internal/core/git/errors.go`
- `internal/core/git/event.go`
- `internal/core/git/event_test.go`
- `internal/core/git/helpers_test.go`
- `internal/core/git/manager.go`
- `internal/core/git/manager_test.go`
- `internal/core/git/types.go`
- `internal/core/git/worktree.go`
- `internal/core/git/worktree_test.go`
