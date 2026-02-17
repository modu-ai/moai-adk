# Implementation Plan: SPEC-GITHUB-WORKFLOW Milestone 3-4

## Context

SPEC-GITHUB-WORKFLOW enhances the GitHub Issues workflow with SPEC integration, worktree automation, and PR management. Milestone 3-4 covers Plan-Run-Sync integration within worktrees and automated PR review/merge.

Two new packages (`internal/workflow/`, `internal/github/`) plus one new file in existing `internal/core/quality/` package are needed. All files are new, so TDD (test-first) is used per hybrid mode configuration.

---

## File Ownership

| File | Owner |
|------|-------|
| `internal/workflow/*` | backend-dev-2 |
| `internal/core/quality/worktree_validator.go` | backend-dev-2 |
| `internal/github/pr_reviewer.go` | backend-dev-2 |
| `internal/github/pr_merger.go` | backend-dev-2 |

---

## Implementation Order

### Phase A: Foundation (error types, package docs)

**Files to create:**

1. `internal/workflow/doc.go` - Package documentation
2. `internal/workflow/errors.go` - Sentinel errors for workflow package
3. `internal/github/doc.go` - Package documentation
4. `internal/github/errors.go` - Sentinel errors for github package

**Sentinel errors:**

`internal/workflow/errors.go`:
- `ErrNotInWorktree` - current dir not inside a worktree
- `ErrSPECNotFound` - SPEC document not found in worktree
- `ErrInvalidSPECID` - SPEC ID format mismatch
- `ErrWorkflowInProgress` - workflow already running
- `ErrQualityGateFailed` - TRUST 5 gates did not pass
- `ErrPlanPhaseFailed`, `ErrRunPhaseFailed`, `ErrSyncPhaseFailed`

`internal/github/errors.go`:
- `ErrGHNotFound` - gh CLI not in PATH
- `ErrGHNotAuthenticated` - gh not authenticated
- `ErrPRNotFound` - PR does not exist
- `ErrPRAlreadyExists` - PR already exists for branch
- `ErrMergeBlocked` - merge prerequisites unmet
- `ErrMergeConflict` - merge conflicts detected
- `ErrCIFailed` - CI/CD checks failed
- `ErrReviewRequired` - review approval required
- `ErrAutoMergeNotRequested` - --auto-merge flag not specified

### Phase B: GitHub CLI Abstraction

**File:** `internal/github/gh.go` + `internal/github/gh_test.go`

GHClient interface wrapping `gh` CLI via `os/exec`, mirroring the `execGit` pattern from `internal/core/git/manager.go:245-271`.

```go
type GHClient interface {
    PRCreate(ctx context.Context, opts PRCreateOptions) (int, error)
    PRView(ctx context.Context, number int) (*PRDetails, error)
    PRMerge(ctx context.Context, number int, method MergeMethod) error
    PRChecks(ctx context.Context, number int) (*CheckStatus, error)
    Push(ctx context.Context, dir string) error
    IsAuthenticated(ctx context.Context) error
}
```

Key types: `PRCreateOptions`, `PRDetails`, `CheckStatus`, `Check`, `CheckConclusion`, `MergeMethod`

`execGH` helper mirrors `execGit`:
- Uses `exec.LookPath("gh")` for binary discovery
- Sets `cmd.Dir` for working directory
- Captures stdout/stderr via `bytes.Buffer`
- Wraps errors with context

**Tests:** Mock-based for unit tests. Integration tests guarded with `exec.LookPath("gh")` skip.

### Phase C: Worktree Quality Validator

**File:** `internal/core/quality/worktree_validator.go` + `internal/core/quality/worktree_validator_test.go`

Lives in existing `quality` package. Reuses `Gate`, `Report`, `QualityConfig`, `TrustGate` types from `trust.go`.

```go
type WorktreeValidator interface {
    Validate(ctx context.Context, wtPath string) (*Report, error)
    ValidateWithConfig(ctx context.Context, wtPath string, config QualityConfig) (*Report, error)
}
```

Implementation creates a scoped `TrustGate` configured for the worktree directory, running tools (go test, golangci-lint) in the worktree working directory, not the main repo root.

**Tests:** Uses existing `mockLSPClient` pattern from `validators_test.go:14-18`.

### Phase D: Worktree Orchestrator

**File:** `internal/workflow/worktree_orchestrator.go` + `internal/workflow/worktree_orchestrator_test.go`

```go
type WorktreeOrchestrator interface {
    DetectWorktreeContext(ctx context.Context, dir string) (*WorktreeContext, error)
    ExecuteWorkflow(ctx context.Context, specID string) (*WorkflowResult, error)
    ValidateQuality(ctx context.Context, specID string) (*quality.Report, error)
    PrepareForReview(ctx context.Context, specID string) (*ReviewReadiness, error)
}
```

Key types:
- `WorktreeContext` - specID, worktreeDir, branch, baseBranch, issueNumber
- `WorkflowResult` - plan/run/sync status, quality report, timestamps
- `ReviewReadiness` - ready flag, quality passed, failure reasons
- `WorkflowPhaseStatus` - pending, running, completed, failed, skipped

Dependencies: `git.WorktreeManager` (existing), `quality.WorktreeValidator` (new)

Constructor: `NewWorktreeOrchestrator(worktreeMgr, qualityGate, logger)`

SPEC ID validation: regex `^SPEC-ISSUE-\d+$`

**Tests:** Table-driven with mock `WorktreeManager` and mock `Gate`. Tests cover:
- Valid/invalid worktree detection
- Plan/Run/Sync phase success and failure combinations
- Quality gate pass/fail scenarios
- Context cancellation

### Phase E: PR Reviewer

**File:** `internal/github/pr_reviewer.go` + `internal/github/pr_reviewer_test.go`

```go
type PRReviewer interface {
    Review(ctx context.Context, prNumber int, specID string) (*ReviewReport, error)
}
```

Key types:
- `ReviewReport` - prNumber, decision, qualityReport, checkStatus, summary, issues
- `ReviewDecision` - APPROVE, REQUEST_CHANGES, COMMENT

Logic:
1. Fetch PR details via `GHClient.PRView()`
2. Run quality validation via `quality.Gate.Validate()`
3. Check CI status via `GHClient.PRChecks()`
4. Decision: all pass = APPROVE, quality/CI fail = REQUEST_CHANGES, CI pending = COMMENT

Dependencies: `GHClient`, `quality.Gate`

Constructor: `NewPRReviewer(gh, qualityGate, logger)`

**Tests:** Table-driven testing all decision combinations with mock `GHClient` and mock `Gate`.

### Phase F: PR Merger

**File:** `internal/github/pr_merger.go` + `internal/github/pr_merger_test.go`

```go
type PRMerger interface {
    Merge(ctx context.Context, prNumber int, opts MergeOptions) (*MergeResult, error)
    CheckPrerequisites(ctx context.Context, prNumber int, opts MergeOptions) (*PrerequisiteCheck, error)
}
```

Key types:
- `MergeOptions` - autoMerge, method, deleteBranch, requireReview, requireChecks
- `MergeResult` - merged, prNumber, method, branchDeleted, mergedAt
- `PrerequisiteCheck` - allMet, individual flags, failureReasons

Logic:
1. Check `--auto-merge` flag (return `ErrAutoMergeNotRequested` if false)
2. Run `PRReviewer.Review()` to verify quality
3. Check CI via `GHClient.PRChecks()`
4. Check mergeability via `GHClient.PRView()`
5. Execute merge via `GHClient.PRMerge()`
6. Optionally delete branch

Dependencies: `GHClient`, `PRReviewer`

Constructor: `NewPRMerger(gh, reviewer, logger)`

**Tests:** Table-driven testing all prerequisite combinations, merge methods, branch deletion.

---

## Existing Code to Reuse

| Component | Location | Usage |
|-----------|----------|-------|
| `WorktreeManager` interface | `internal/core/git/types.go:156-195` | Worktree operations in orchestrator |
| `Worktree` struct | `internal/core/git/types.go:51-60` | Worktree data |
| `Gate` interface | `internal/core/quality/trust.go:169-175` | Quality validation |
| `Report` struct | `internal/core/quality/trust.go:105-115` | Quality reports |
| `QualityConfig` | `internal/core/quality/trust.go:205-216` | Quality configuration |
| `TrustGate` | `internal/core/quality/trust.go:358+` | Concrete quality gate |
| `execGit` pattern | `internal/core/git/manager.go:245-271` | Mirror for `execGH` |
| `mockLSPClient` | `internal/core/quality/validators_test.go:14-18` | Test mock pattern |
| Sentinel error pattern | `internal/foundation/errors.go` | Error definitions |

---

## Conventions to Follow

- **Compile-time checks:** `var _ Interface = (*impl)(nil)` for all interfaces
- **Error wrapping:** `fmt.Errorf("context: %w", err)`
- **Logging:** `slog.Default().With("module", "github")`
- **Testing:** Table-driven with `t.Run()`, `t.Parallel()`, `t.Helper()`, `t.TempDir()`
- **Unexported impl, exported interface:** `type prReviewer struct` / `type PRReviewer interface`
- **Constructor returns concrete:** `func NewPRReviewer(...) *prReviewer`

---

## Verification Plan

1. **Unit tests:** `go test -race ./internal/workflow/... ./internal/github/... ./internal/core/quality/...`
2. **Coverage:** `go test -cover ./internal/workflow/... ./internal/github/...` (target 85%+)
3. **Lint:** `golangci-lint run ./internal/workflow/... ./internal/github/...`
4. **Vet:** `go vet ./internal/workflow/... ./internal/github/...`
5. **Build:** `go build ./...` (ensure compilation)

---

## Summary

| Phase | Files | Depends On |
|-------|-------|-----------|
| A | workflow/doc.go, workflow/errors.go, github/doc.go, github/errors.go | None |
| B | github/gh.go, github/gh_test.go | Phase A |
| C | core/quality/worktree_validator.go + test | Existing quality package |
| D | workflow/worktree_orchestrator.go + test | Phase A, C |
| E | github/pr_reviewer.go + test | Phase A, B |
| F | github/pr_merger.go + test | Phase E |

Total new files: ~14 (7 implementation + 5 test + 2 doc.go)
