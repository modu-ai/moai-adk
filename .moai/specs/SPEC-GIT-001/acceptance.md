# SPEC-GIT-001: Acceptance Criteria

---
spec_id: SPEC-GIT-001
type: acceptance
created: 2026-02-03
tags: git, go-git, acceptance, test, given-when-then
---

## 1. Repository Operations

### AC-R001: Repository Open - Valid Repository

```gherkin
Scenario: Open a valid Git repository
  Given a directory that contains a valid Git repository
  When NewRepository(path) is called
  Then a non-nil gitManager instance is returned
  And err is nil
  And Root() returns the absolute path to the repository root
```

### AC-R002: Repository Open - Invalid Path

```gherkin
Scenario: Open a non-repository directory
  Given a directory that does not contain a .git directory
  When NewRepository(path) is called
  Then err wraps ErrNotRepository
  And the returned instance is nil
```

### AC-R003: Current Branch - Normal State

```gherkin
Scenario: Get current branch on a normal checkout
  Given a Git repository with branch "main" checked out
  When CurrentBranch() is called
  Then "main" is returned
  And err is nil
```

### AC-R004: Current Branch - Detached HEAD

```gherkin
Scenario: Get current branch with detached HEAD
  Given a Git repository with HEAD detached at a specific commit
  When CurrentBranch() is called
  Then an empty string is returned
  And err wraps ErrDetachedHEAD
```

### AC-R005: Status - Clean Working Tree

```gherkin
Scenario: Get status of a clean working tree
  Given a Git repository with no uncommitted changes
  When Status() is called
  Then GitStatus.Staged is empty
  And GitStatus.Modified is empty
  And GitStatus.Untracked is empty
```

### AC-R006: Status - Dirty Working Tree

```gherkin
Scenario: Get status of a dirty working tree
  Given a Git repository with:
    | staged file     | staged.txt    |
    | modified file   | modified.txt  |
    | untracked file  | untracked.txt |
  When Status() is called
  Then GitStatus.Staged contains "staged.txt"
  And GitStatus.Modified contains "modified.txt"
  And GitStatus.Untracked contains "untracked.txt"
```

### AC-R007: Status - Ahead/Behind Count

```gherkin
Scenario: Get ahead/behind count relative to upstream
  Given a Git repository with an upstream tracking branch
  And local branch has 2 commits ahead and 1 commit behind
  When Status() is called
  Then GitStatus.Ahead equals 2
  And GitStatus.Behind equals 1
```

### AC-R008: Status - No Upstream

```gherkin
Scenario: Get status without upstream tracking branch
  Given a Git repository with a local-only branch (no upstream)
  When Status() is called
  Then GitStatus.Ahead equals 0
  And GitStatus.Behind equals 0
  And err is nil
```

### AC-R009: Commit Log

```gherkin
Scenario: Retrieve last N commits from HEAD
  Given a Git repository with 5 commits
  When Log(3) is called
  Then 3 Commit records are returned
  And each Commit has non-empty Hash, Author, Date, and Message
  And commits are ordered newest-first
```

### AC-R010: Commit Log - Exceeding Total

```gherkin
Scenario: Request more commits than available
  Given a Git repository with 3 commits
  When Log(10) is called
  Then 3 Commit records are returned
  And err is nil
```

### AC-R011: Diff Between Refs

```gherkin
Scenario: Generate diff between two branches
  Given a Git repository with branches "main" and "feature"
  And "feature" has a modified file compared to "main"
  When Diff("main", "feature") is called
  Then a non-empty unified diff string is returned
  And the diff contains the modified file path
```

### AC-R012: IsClean - True

```gherkin
Scenario: Check cleanness on a clean working tree
  Given a Git repository with no uncommitted changes
  When IsClean() is called
  Then true is returned
  And err is nil
```

### AC-R013: IsClean - False

```gherkin
Scenario: Check cleanness on a dirty working tree
  Given a Git repository with a modified file
  When IsClean() is called
  Then false is returned
  And err is nil
```

### AC-R014: Root Path

```gherkin
Scenario: Get repository root path
  Given a Git repository at /tmp/test-repo
  When Root() is called
  Then the returned string equals the absolute path to /tmp/test-repo
  And the path is normalized via filepath.Clean()
```

---

## 2. Branch Operations

### AC-B001: Create Branch - Success

```gherkin
Scenario: Create a new branch from HEAD
  Given a Git repository with branch "main" checked out
  And no branch named "feature/login" exists
  When Create("feature/login") is called
  Then err is nil
  And branch "feature/login" appears in List()
```

### AC-B002: Create Branch - Already Exists

```gherkin
Scenario: Create a branch that already exists
  Given a Git repository with branch "feature/login" already existing
  When Create("feature/login") is called
  Then err wraps ErrBranchExists
```

### AC-B003: Create Branch - Invalid Name

```gherkin
Scenario Outline: Create a branch with invalid name
  Given a Git repository
  When Create(<invalid_name>) is called
  Then err is returned with validation failure message

  Examples:
    | invalid_name     | reason                         |
    | "feat..ure"      | contains double dots (..)      |
    | "feat~1"         | contains tilde (~)             |
    | "feat^2"         | contains caret (^)             |
    | "feat:name"      | contains colon (:)             |
    | "feat name"      | contains space                 |
    | ""               | empty string                   |
    | ".feat"          | starts with dot                |
    | "feat.lock"      | ends with .lock                |
```

### AC-B004: Switch Branch - Success

```gherkin
Scenario: Switch to an existing branch
  Given a Git repository with branches "main" and "feature/login"
  And "main" is currently checked out
  And the working tree is clean
  When Switch("feature/login") is called
  Then err is nil
  And CurrentBranch() returns "feature/login"
```

### AC-B005: Switch Branch - Not Found

```gherkin
Scenario: Switch to a non-existent branch
  Given a Git repository with no branch named "nonexistent"
  When Switch("nonexistent") is called
  Then err wraps ErrBranchNotFound
```

### AC-B006: Switch Branch - Dirty Working Tree

```gherkin
Scenario: Switch branch with uncommitted changes
  Given a Git repository with modified files in the working tree
  When Switch("feature/login") is called
  Then err wraps ErrDirtyWorkingTree
  And CurrentBranch() returns the original branch (unchanged)
```

### AC-B007: Delete Branch - Success

```gherkin
Scenario: Delete a non-current branch
  Given a Git repository with "main" checked out
  And branch "feature/old" exists
  When Delete("feature/old") is called
  Then err is nil
  And "feature/old" does not appear in List()
```

### AC-B008: Delete Branch - Current Branch

```gherkin
Scenario: Delete the currently checked-out branch
  Given a Git repository with "feature/current" checked out
  When Delete("feature/current") is called
  Then err wraps ErrCannotDeleteCurrentBranch
  And the branch still exists
```

### AC-B009: List Branches

```gherkin
Scenario: List all local branches
  Given a Git repository with branches "main", "develop", "feature/a"
  And "develop" is currently checked out
  When List() is called
  Then 3 Branch records are returned
  And each has non-empty Name field
  And only the "develop" branch has IsCurrent == true
  And all branches have IsRemote == false
```

---

## 3. Conflict Detection

### AC-C001: No Conflicts

```gherkin
Scenario: Detect no conflicts between branches
  Given a Git repository with branches "main" and "feature"
  And "main" modified file "a.go"
  And "feature" modified file "b.go" (different file)
  When HasConflicts("feature") is called from "main"
  Then false is returned
  And err is nil
```

### AC-C002: Conflicts Detected

```gherkin
Scenario: Detect conflicts between branches
  Given a Git repository with branches "main" and "feature"
  And both "main" and "feature" modified the same file "shared.go"
  And the modifications are on different content within the same file
  When HasConflicts("feature") is called from "main"
  Then true is returned
  And err is nil
```

### AC-C003: Target Branch Not Found

```gherkin
Scenario: Conflict detection with non-existent target
  Given a Git repository with branch "main" checked out
  When HasConflicts("nonexistent") is called
  Then err wraps ErrBranchNotFound
```

### AC-C004: Working Tree Unmodified After Detection

```gherkin
Scenario: Conflict detection does not modify working tree
  Given a Git repository with modified files in the working tree
  And the current Status() returns specific staged/modified/untracked files
  When HasConflicts("feature") is called
  Then the working tree Status() returns the same staged/modified/untracked files as before
  And HEAD has not changed
  And the staging area has not changed
```

### AC-C005: Merge Base Calculation

```gherkin
Scenario: Calculate merge base between two branches
  Given a Git repository with:
    | commit A (common ancestor) |
    | main: commit B, commit C   |
    | feature: commit D, commit E|
  When MergeBase("main", "feature") is called
  Then the hash of commit A is returned
  And err is nil
```

### AC-C006: No Common Ancestor

```gherkin
Scenario: Merge base with unrelated branches
  Given a Git repository with two orphan branches (no common history)
  When MergeBase("orphan1", "orphan2") is called
  Then err wraps ErrNoMergeBase
```

---

## 4. Worktree Operations

### AC-W001: Worktree Add - New Branch

```gherkin
Scenario: Create a new worktree with a new branch
  Given a Git repository with branch "main" checked out
  And no branch named "feature/parallel" exists
  And the path "/tmp/worktree-parallel" does not exist
  When Add("/tmp/worktree-parallel", "feature/parallel") is called
  Then err is nil
  And a directory exists at "/tmp/worktree-parallel"
  And the worktree appears in List()
  And the worktree's Branch field is "feature/parallel"
```

### AC-W002: Worktree Add - Existing Branch

```gherkin
Scenario: Create a worktree for an existing branch
  Given a Git repository with branch "feature/existing" already created
  And the path "/tmp/worktree-existing" does not exist
  When Add("/tmp/worktree-existing", "feature/existing") is called
  Then err is nil
  And the worktree is linked to branch "feature/existing"
```

### AC-W003: Worktree Add - Path Already Exists

```gherkin
Scenario: Create a worktree at an existing path
  Given the path "/tmp/worktree-conflict" already exists
  When Add("/tmp/worktree-conflict", "feature/new") is called
  Then err wraps ErrWorktreePathExists
```

### AC-W004: Worktree Add - System Git Not Found

```gherkin
Scenario: Create a worktree without system Git installed
  Given system Git is not available in PATH
  When Add("/tmp/worktree", "feature") is called
  Then err wraps ErrSystemGitNotFound
```

### AC-W005: Worktree Add - Timeout

```gherkin
Scenario: Worktree creation exceeds timeout
  Given a Git repository
  And the context timeout is set to 1 millisecond
  When Add("/tmp/worktree-timeout", "feature") is called with the short-timeout context
  Then err contains context deadline exceeded
```

### AC-W006: Worktree List

```gherkin
Scenario: List all active worktrees
  Given a Git repository with:
    | main worktree at /repo            |
    | worktree at /tmp/wt-1 on "feat-1" |
    | worktree at /tmp/wt-2 on "feat-2" |
  When List() is called
  Then 3 Worktree records are returned
  And each has non-empty Path, Branch, and HEAD fields
  And the main worktree is included
```

### AC-W007: Worktree List - Porcelain Parsing

```gherkin
Scenario: Parse git worktree list --porcelain output
  Given the following porcelain output:
    """
    worktree /Users/goos/project
    HEAD abc123def456
    branch refs/heads/main

    worktree /tmp/wt-feature
    HEAD def789abc012
    branch refs/heads/feature
    """
  When the porcelain parser processes this output
  Then 2 Worktree records are returned
  And the first has Path "/Users/goos/project" and Branch "main"
  And the second has Path "/tmp/wt-feature" and Branch "feature"
```

### AC-W008: Worktree Remove - Success

```gherkin
Scenario: Remove a clean worktree
  Given a worktree at "/tmp/wt-remove" with no uncommitted changes
  When Remove("/tmp/wt-remove") is called
  Then err is nil
  And the worktree no longer appears in List()
  And the directory "/tmp/wt-remove" no longer exists
```

### AC-W009: Worktree Remove - Dirty

```gherkin
Scenario: Remove a dirty worktree without force
  Given a worktree at "/tmp/wt-dirty" with uncommitted changes
  When Remove("/tmp/wt-dirty") is called
  Then err wraps ErrWorktreeDirty
  And the worktree still appears in List()
```

### AC-W010: Worktree Remove - Not Found

```gherkin
Scenario: Remove a non-existent worktree
  Given no worktree at path "/tmp/wt-nonexistent"
  When Remove("/tmp/wt-nonexistent") is called
  Then err wraps ErrWorktreeNotFound
```

### AC-W011: Worktree Prune

```gherkin
Scenario: Prune stale worktree references
  Given a worktree was previously created at "/tmp/wt-stale"
  And the directory "/tmp/wt-stale" was manually deleted
  When Prune() is called
  Then err is nil
  And the stale worktree reference is removed from the internal tracking
```

---

## 5. Event Detection

### AC-E001: Branch Switch Event

```gherkin
Scenario: Detect branch switch event
  Given the repository is on branch "main"
  And the event detector has captured the initial state
  When Switch("feature") is called
  And the event detector polls for changes
  Then a BranchSwitch event is emitted
  And the event's PreviousBranch is "main"
  And the event's CurrentBranch is "feature"
```

### AC-E002: Commit Event

```gherkin
Scenario: Detect new commit event
  Given the repository is on branch "main" with HEAD at commit A
  And the event detector has captured the initial state
  When a new commit B is created
  And the event detector polls for changes
  Then a NewCommit event is emitted
  And the event's PreviousHEAD is commit A's hash
  And the event's CurrentHEAD is commit B's hash
```

### AC-E003: Polling Cancellation

```gherkin
Scenario: Event polling respects context cancellation
  Given an event detector polling every 5 seconds
  When the context is cancelled
  Then the polling goroutine terminates gracefully
  And no further events are emitted
```

---

## 6. Cross-Cutting Concerns

### AC-X001: Error Wrapping Consistency

```gherkin
Scenario: All errors are wrapped with context
  Given any Git operation that encounters an internal error
  When the error is returned to the caller
  Then the error can be unwrapped with errors.Is() to match the sentinel error
  And the error message contains the operation context (e.g., "create branch")
```

### AC-X002: Structured Logging

```gherkin
Scenario: Operations produce structured log entries
  Given a configured slog.Logger
  When any Git operation is executed
  Then a debug-level log entry is produced at operation start
  And a debug-level log entry is produced at operation completion
  And error-level entries include the operation context and error details
```

### AC-X003: System Git Fallback Logging

```gherkin
Scenario: System Git fallback is logged
  Given a worktree operation is requested
  When the system Git fallback is invoked
  Then an info-level log entry is produced
  And the log entry contains the fallback reason
  And the log entry contains the executed git command
```

### AC-X004: Context Timeout Enforcement

```gherkin
Scenario: Long operations respect context timeout
  Given a context with a 5-second timeout
  When a worktree operation takes longer than 5 seconds
  Then the operation is cancelled
  And err contains context.DeadlineExceeded
```

### AC-X005: Goroutine Safety

```gherkin
Scenario: Concurrent read operations are safe
  Given a Git repository
  When 10 goroutines simultaneously call CurrentBranch()
  And 10 goroutines simultaneously call Status()
  And 10 goroutines simultaneously call IsClean()
  Then all operations complete without race conditions
  And the -race flag detects no data races
```

---

## 7. Performance Benchmarks

### AC-P001: Operation Latency

```gherkin
Scenario: Operations meet latency requirements
  Given a Git repository with 100 commits and 50 files
  Then CurrentBranch() completes within 10ms
  And Status() completes within 50ms
  And Log(10) completes within 30ms
  And IsClean() completes within 50ms
  And Branch Create/Switch/Delete completes within 100ms each
  And Worktree List completes within 100ms
```

### AC-P002: Large Repository Performance

```gherkin
Scenario: Operations perform acceptably on large repositories
  Given a Git repository with 10,000+ commits and 1,000+ files
  Then Status() completes within 500ms
  And Log(50) completes within 200ms
  And Branch List completes within 200ms
```

---

## 8. Quality Gates

### Definition of Done

- [ ] 모든 인터페이스 메서드 구현 완료 (Repository: 6, BranchManager: 6, WorktreeManager: 4)
- [ ] 모든 sentinel error 정의 및 사용 완료 (11 errors)
- [ ] 단위 테스트 커버리지 85% 이상
- [ ] 통합 테스트 key paths 커버리지 확보
- [ ] 벤치마크 테스트 성능 요구사항 통과
- [ ] `go vet` 경고 0건
- [ ] `golangci-lint` 오류 0건
- [ ] `-race` 플래그 테스트 통과
- [ ] godoc 주석 모든 exported 타입/함수/메서드에 작성
- [ ] mockery로 생성된 mock 파일 정상 동작 확인

### Verification Methods

| Category | Method | Tool |
|----------|--------|------|
| Correctness | Unit + Integration tests | `go test ./internal/core/git/...` |
| Coverage | Coverage report | `go test -coverprofile=coverage.out` |
| Race Safety | Race detector | `go test -race ./internal/core/git/...` |
| Performance | Benchmark tests | `go test -bench=. ./internal/core/git/...` |
| Lint | Static analysis | `golangci-lint run ./internal/core/git/...` |
| Security | Security scan | `gosec ./internal/core/git/...` |
| Documentation | godoc review | `go doc ./internal/core/git/` |
