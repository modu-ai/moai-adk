---
name: moai-workflow-github
description: >
  GitHub integration workflow for issue-to-SPEC conversion, bidirectional
  issue-SPEC linking, branch prefix detection, worktree orchestration with
  Plan-Run-Sync integration, TRUST 5 quality validation, automated PR review
  and merge, multilingual issue closure, and tmux session automation.
  Use when parsing GitHub issues, linking issues to SPECs, orchestrating
  worktree workflows, reviewing PRs, merging with prerequisites, closing
  issues with comments, or managing tmux sessions.
license: Apache-2.0
compatibility: Designed for Claude Code
user-invocable: false
metadata:
  version: "1.2.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-16"
  tags: "github, issues, spec, linking, branch, worktree, pr, workflow, i18n, multilingual, closure, tmux, session, quality, trust5, review, merge, orchestrator"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["github", "issue", "parse-issue", "link-spec", "branch prefix", "issue-to-spec", "close-issue", "issue-closure", "tmux", "multilingual", "i18n", "worktree-orchestrator", "pr-reviewer", "pr-merger", "quality-gate", "auto-merge"]
  agents: ["expert-backend", "manager-spec", "manager-git"]
  phases: ["plan", "run"]
---

# Workflow: GitHub Integration

Purpose: Provide CLI helpers and library functions for integrating GitHub issues with MoAI's SPEC-driven development workflow. Covers issue parsing, bidirectional linking, branch naming, and the full issue-to-SPEC-to-PR pipeline.

Flow: Parse Issue -> Link to SPEC -> Detect Branch Prefix -> Create Worktree -> Plan-Run-Sync -> PR -> Merge -> Close Issue

---

## Component Overview

### CLI Commands (internal/cli/github.go)

Parent command: `moai github`

Subcommands:
- `moai github parse-issue <number>` - Parse and display a GitHub issue
- `moai github link-spec <issue-number> <spec-id>` - Link an issue to a SPEC document

### Library Packages

| Package | File | Purpose |
|---------|------|---------|
| `internal/github` | `gh.go` | GHClient interface wrapping `gh` CLI |
| `internal/github` | `issue_parser.go` | Parse GitHub issues via `gh` CLI |
| `internal/github` | `spec_linker.go` | Bidirectional issue-SPEC registry |
| `internal/github` | `pr_reviewer.go` | Automated PR review with quality gates |
| `internal/github` | `pr_merger.go` | Conditional PR merge with prerequisites |
| `internal/github` | `errors.go` | Sentinel errors for all GitHub operations |
| `internal/core/quality` | `worktree_validator.go` | TRUST 5 validation scoped to worktrees |
| `internal/workflow` | `worktree_orchestrator.go` | Plan-Run-Sync orchestration in worktrees |
| `internal/workflow` | `errors.go` | Sentinel errors for workflow operations |
| `internal/git` | `branch_detector.go` | Label-to-branch-prefix mapping |

---

## Subcommand: parse-issue

### Usage

```
moai github parse-issue <number>
```

### Behavior

1. Validates the issue number argument (must be positive integer)
2. Calls `gh issue view <number> --json number,title,body,labels,author,comments`
3. Parses JSON response into `github.Issue` struct
4. Displays formatted card with issue details:
   - Number, title, author
   - Labels (comma-separated)
   - Body (truncated to 200 characters)
   - Comment count

### Interface

```go
type IssueParser interface {
    ParseIssue(ctx context.Context, number int) (*Issue, error)
}
```

Default implementation uses `execGH` from `internal/github/gh.go`. Tests inject a mock via the `GithubIssueParser` package variable.

### Error Handling

- Invalid issue number: Returns `fmt.Errorf("invalid issue number %q: %w", ...)`
- `gh` CLI not found: Wrapped `ErrGHNotFound`
- Issue does not exist: Wrapped error from `gh` stderr
- Empty title in response: Returns parse validation error

---

## Subcommand: link-spec

### Usage

```
moai github link-spec <issue-number> <spec-id>
```

### Behavior

1. Validates issue number (positive integer) and spec ID (non-empty string)
2. Creates a `SpecLinker` via `GithubSpecLinkerFactory` using current working directory
3. Calls `linker.LinkIssueToSpec(issueNum, specID)` to persist the mapping
4. Displays success card with issue number, SPEC ID, registry path, and timestamp

### Interface

```go
type SpecLinker interface {
    LinkIssueToSpec(issueNum int, specID string) error
    GetLinkedSpec(issueNum int) (string, error)
    GetLinkedIssue(specID string) (int, error)
    ListMappings() []SpecMapping
}
```

### Registry Storage

Mappings are stored in `.moai/github-spec-registry.json`:

```json
{
  "version": "1.0.0",
  "mappings": [
    {
      "issue_number": 123,
      "spec_id": "SPEC-ISSUE-123",
      "created_at": "2026-02-16T12:00:00Z",
      "status": "active"
    }
  ]
}
```

Storage features:
- Atomic writes via temp file + `os.Rename`
- Graceful handling of missing registry file (creates empty)
- Corrupt file recovery (backs up `.corrupt` suffix, starts fresh)
- Duplicate detection: Returns `ErrMappingExists` if issue already linked

### Error Handling

- Duplicate mapping: `ErrMappingExists` ("github: issue already linked to a SPEC")
- Mapping not found: `ErrMappingNotFound` ("github: no SPEC linked to issue")
- Factory failure: Wrapped error from `NewSpecLinker`

---

## Branch Prefix Detection

### Location

`internal/git/branch_detector.go`

### Functions

```go
func DetectBranchPrefix(labels []string) string
func FormatIssueBranch(labels []string, issueNumber int) string
```

### Label-to-Prefix Mapping

| Label | Branch Prefix | Example |
|-------|---------------|---------|
| `bug` | `fix/` | `fix/issue-123` |
| `feature` | `feat/` | `feat/issue-456` |
| `enhancement` | `feat/` | `feat/issue-789` |
| `documentation` | `docs/` | `docs/issue-101` |
| `docs` | `docs/` | `docs/issue-102` |
| (no match) | `feat/` | `feat/issue-200` |

Rules:
- First matching label wins (scan order)
- Default prefix is `feat/` when no labels match
- Case-sensitive matching

### Integration with Issue Classification

The `moai:github issues` slash command (`.claude/commands/moai/github.md`) uses the same label-to-prefix mapping at Step 1.3 (Issue Classification). The `branch_detector.go` functions provide the canonical implementation that the slash command references.

---

## Integration Points

### With /moai:github Slash Command

The `/moai:github issues` workflow uses these CLI helpers internally:

1. **Issue Discovery** (Phase 1): `IssueParser.ParseIssue()` extracts issue data
2. **Issue Linking** (Phase 2): `SpecLinker.LinkIssueToSpec()` records issue-SPEC mapping after SPEC creation
3. **Branch Creation** (Phase 3): `FormatIssueBranch()` generates branch name from labels

### With manager-spec Agent

SPEC generation from GitHub issues follows this sequence:

1. `moai github parse-issue <number>` extracts structured issue data
2. `manager-spec` agent converts issue content to EARS format SPEC
3. `moai github link-spec <number> SPEC-ISSUE-<number>` creates bidirectional link
4. SPEC files created at `.moai/specs/SPEC-ISSUE-{number}/`

The `manager-spec` agent handles the EARS conversion. The CLI helpers provide data extraction and link persistence.

### With Worktree System

After SPEC creation and linking:

1. `FormatIssueBranch(issue.LabelNames(), issue.Number)` determines branch name
2. `moai worktree new <branch-name>` creates isolated worktree (existing command)
3. `moai worktree go <branch-name>` navigates to worktree (existing command)
4. `moai worktree done <branch-name>` completes and cleans up (existing command)

The existing `resolveSpecBranch()` in `internal/cli/worktree/new.go` already handles `SPEC-ISSUE-{number}` patterns by mapping them to `feature/SPEC-ISSUE-{number}`.

---

## Data Types

### Issue (internal/github/issue_parser.go)

```go
type Issue struct {
    Number   int       `json:"number"`
    Title    string    `json:"title"`
    Body     string    `json:"body"`
    Labels   []Label   `json:"labels"`
    Author   Author    `json:"author"`
    Comments []Comment `json:"comments"`
}

type Label struct {
    Name string `json:"name"`
}

type Author struct {
    Login string `json:"login"`
}

type Comment struct {
    Body      string    `json:"body"`
    Author    Author    `json:"author"`
    CreatedAt time.Time `json:"createdAt"`
}
```

Helper: `issue.LabelNames()` returns `[]string` of label names for use with `DetectBranchPrefix`.

### SpecMapping (internal/github/spec_linker.go)

```go
type SpecMapping struct {
    IssueNumber int       `json:"issue_number"`
    SpecID      string    `json:"spec_id"`
    CreatedAt   time.Time `json:"created_at"`
    Status      string    `json:"status"`
}
```

---

## Sentinel Errors (internal/github/errors.go)

| Error | Description | Milestone |
|-------|-------------|-----------|
| `ErrGHNotFound` | `gh` CLI binary not in PATH | 1-2 |
| `ErrGHNotAuthenticated` | `gh` not authenticated | 1-2 |
| `ErrIssueNotFound` | Specified issue does not exist | 1-2 |
| `ErrMappingExists` | Issue already linked to a SPEC | 1-2 |
| `ErrMappingNotFound` | No SPEC linked to issue | 1-2 |
| `ErrMaxRetriesExceeded` | All retry attempts exhausted | 5-6 |
| `ErrPRNotFound` | Pull request does not exist | 3-4 |
| `ErrPRAlreadyExists` | PR already exists for branch | 3-4 |
| `ErrMergeBlocked` | Merge prerequisites unmet | 3-4 |
| `ErrMergeConflict` | Merge conflicts detected | 3-4 |
| `ErrCIFailed` | CI/CD checks failed | 3-4 |
| `ErrReviewRequired` | Review approval required | 3-4 |
| `ErrAutoMergeNotRequested` | `--auto-merge` flag not specified | 3-4 |
| `ErrCommentFailed` | Failed to post issue comment | 5-6 |
| `ErrCloseFailed` | Failed to close issue | 5-6 |
| `ErrLabelFailed` | Failed to add label | 5-6 |

---

## Milestone 3: Plan-Run-Sync Integration within Worktrees

### Worktree Quality Validator (internal/core/quality/worktree_validator.go)

#### Interface

```go
type WorktreeValidator interface {
    Validate(ctx context.Context, wtPath string) (*Report, error)
    ValidateWithConfig(ctx context.Context, wtPath string, config QualityConfig) (*Report, error)
}

type GateFactory func(config QualityConfig) Gate
```

`WorktreeValidator` runs TRUST 5 quality gates scoped to a worktree directory rather than the main repository root. This enables isolated quality validation per-SPEC.

#### Dependency Injection

The `GateFactory` function type decouples gate construction from the validator, enabling both production use and test mocking:

- **Production**: `DefaultGateFactory(lsp)` creates `TrustGate` instances with `TestedValidator` and `ReadableValidator`
- **Testing**: Inject a mock factory that returns a mock `Gate`

```go
// Production setup
factory := quality.DefaultGateFactory(lspClient)
validator := quality.NewWorktreeValidator(factory, config, logger)

// Test setup
mockFactory := func(_ quality.QualityConfig) quality.Gate { return mockGate }
validator := quality.NewWorktreeValidator(mockFactory, config, nil)
```

#### Validation Flow

1. Verify worktree path exists and is a directory
2. Create a `Gate` via the `GateFactory` with the current config
3. Run `gate.Validate(ctx)` to produce a `*Report`
4. Log results (passed/failed, score)

### Worktree Orchestrator (internal/workflow/worktree_orchestrator.go)

#### Interface

```go
type WorktreeOrchestrator interface {
    DetectWorktreeContext(ctx context.Context, dir string) (*WorktreeContext, error)
    ExecuteWorkflow(ctx context.Context, specID string) (*WorkflowResult, error)
    ValidateQuality(ctx context.Context, specID string) (*Report, error)
    PrepareForReview(ctx context.Context, specID string) (*ReviewReadiness, error)
}
```

#### Data Types

```go
type WorktreeContext struct {
    SpecID      string    // "SPEC-ISSUE-123"
    WorktreeDir string    // Absolute filesystem path
    Branch      string    // Git branch name
    BaseBranch  string    // Target merge branch (default: "main")
    IssueNumber int       // Extracted from SPEC ID
}

type WorkflowResult struct {
    SpecID        string
    PlanStatus    WorkflowPhaseStatus  // pending | running | completed | failed | skipped
    RunStatus     WorkflowPhaseStatus
    SyncStatus    WorkflowPhaseStatus
    QualityReport *quality.Report
    StartedAt     time.Time
    CompletedAt   time.Time
}

type ReviewReadiness struct {
    Ready          bool
    QualityPassed  bool
    QualityReport  *quality.Report
    FailureReasons []string
}
```

#### SPEC ID Validation

All operations validate the SPEC ID against `^SPEC-ISSUE-\d+$`. Invalid IDs return `ErrInvalidSPECID`.

#### DetectWorktreeContext

Identifies which worktree contains a given directory and extracts SPEC metadata:

1. List all worktrees via `WorktreeManager.List()`
2. Resolve `dir` to absolute path
3. Find the worktree whose path is a prefix of `dir`
4. Extract SPEC ID from the worktree directory name (last path component)
5. Verify SPEC document exists at `.moai/specs/{SPEC-ID}/` within the worktree
6. Extract issue number from SPEC ID (e.g., `SPEC-ISSUE-123` → `123`)

#### ExecuteWorkflow

Runs the full Plan-Run-Sync sequence for a SPEC:

| Phase | On Success | On Failure |
|-------|-----------|------------|
| Plan | Continue to Run | Mark Run+Sync as `skipped`, return `ErrPlanPhaseFailed` |
| Run | Quality validation, continue to Sync | Mark Sync as `skipped`, return `ErrRunPhaseFailed` |
| Quality | Log warning if error (non-fatal) | Continue to Sync |
| Sync | Mark all `completed` | Return `ErrSyncPhaseFailed` |

Quality validation runs after the Run phase but before Sync. A quality failure is logged as a warning but does not block the Sync phase.

#### ValidateQuality

Standalone quality validation for a SPEC:

1. Validate SPEC ID format
2. Find worktree for the SPEC via `findWorktreeForSpec()`
3. Run `validator.Validate(ctx, worktreeDir)`

#### PrepareForReview

Checks whether a worktree is ready for PR creation:

1. Validate SPEC ID format
2. Find worktree for the SPEC
3. Run quality validation
4. If validation errors: add to `FailureReasons`, return `Ready: false`
5. If quality report failed: collect severity-error issues into `FailureReasons`
6. If quality passed: return `Ready: true`

#### PhaseExecutor

The orchestrator delegates actual phase execution to a `PhaseExecutor` interface:

```go
type PhaseExecutor interface {
    ExecutePlan(ctx context.Context, specID, workDir string) error
    ExecuteRun(ctx context.Context, specID, workDir string) error
    ExecuteSync(ctx context.Context, specID, workDir string) error
}
```

This abstraction allows testing the orchestrator without invoking real Claude Code commands.

### Sentinel Errors (internal/workflow/errors.go)

| Error | Description |
|-------|-------------|
| `ErrNotInWorktree` | Current directory is not inside a worktree |
| `ErrSPECNotFound` | SPEC document not found in the worktree |
| `ErrInvalidSPECID` | SPEC ID does not match `SPEC-ISSUE-\d+` format |
| `ErrWorkflowInProgress` | Workflow already running |
| `ErrQualityGateFailed` | TRUST 5 quality gates did not pass |
| `ErrPlanPhaseFailed` | Plan phase failed |
| `ErrRunPhaseFailed` | Run phase failed |
| `ErrSyncPhaseFailed` | Sync phase failed |

---

## Milestone 4: PR Review and Automated Merge

### GitHub CLI Abstraction (internal/github/gh.go)

#### Interface

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

#### Data Types

```go
type PRCreateOptions struct {
    Title       string
    Body        string
    BaseBranch  string
    HeadBranch  string
    Labels      []string
    IssueNumber int
}

type PRDetails struct {
    Number     int
    Title      string
    State      string     // "OPEN", "MERGED", "CLOSED"
    Mergeable  string     // "MERGEABLE", "CONFLICTING", "UNKNOWN"
    HeadBranch string
    BaseBranch string
    URL        string
    CreatedAt  time.Time
}

type CheckStatus struct {
    Overall CheckConclusion  // "pass", "fail", "pending"
    Checks  []Check
}

type MergeMethod string  // "merge", "squash", "rebase"
```

#### Implementation

The `ghClient` struct wraps the `gh` CLI binary via `os/exec`, mirroring the `execGit` pattern from `internal/core/git/manager.go`:

- Uses `exec.LookPath("gh")` for binary discovery
- Sets `cmd.Dir` for working directory
- Captures stdout/stderr via `bytes.Buffer`
- Wraps all errors with context

### PR Reviewer (internal/github/pr_reviewer.go)

#### Interface

```go
type PRReviewer interface {
    Review(ctx context.Context, prNumber int, specID string) (*ReviewReport, error)
}

type ReviewReport struct {
    PRNumber      int
    Decision      ReviewDecision    // APPROVE | REQUEST_CHANGES | COMMENT
    QualityReport *quality.Report
    CheckStatus   *CheckStatus
    Summary       string            // Human-readable Markdown summary
    Issues        []string          // Specific problems found
}
```

#### Review Decision Logic

The reviewer runs two independent checks (quality gates + CI status) and combines results:

| Quality Gate | CI Status | Decision |
|-------------|-----------|----------|
| Passed | Pass | `APPROVE` |
| Passed | Fail | `REQUEST_CHANGES` |
| Passed | Pending | `COMMENT` |
| Passed | Error | `COMMENT` |
| Failed | Any | `REQUEST_CHANGES` |
| Error | Any | `REQUEST_CHANGES` |

#### Review Flow

1. Fetch PR details via `GHClient.PRView()` — verify PR exists and is `OPEN`
2. Run quality validation via `quality.Gate.Validate()`
3. Check CI status via `GHClient.PRChecks()`
4. Determine decision based on combined results
5. Collect issues (quality errors, CI failures) into `Issues` list
6. Generate Markdown summary via `buildSummary()`

#### Summary Output Format

```markdown
## PR #42 Review

**Decision: APPROVE**

All quality gates passed and CI/CD checks are green.

### Quality Score: 95.0%
```

### PR Merger (internal/github/pr_merger.go)

#### Interface

```go
type PRMerger interface {
    Merge(ctx context.Context, prNumber int, opts MergeOptions) (*MergeResult, error)
    CheckPrerequisites(ctx context.Context, prNumber int, opts MergeOptions) (*PrerequisiteCheck, error)
}
```

#### Data Types

```go
type MergeOptions struct {
    AutoMerge     bool         // Required: must be true to merge
    Method        MergeMethod  // "merge" (default), "squash", "rebase"
    DeleteBranch  bool         // Delete branch after merge
    RequireReview bool         // Require APPROVE review decision
    RequireChecks bool         // Require CI checks to pass
    SpecID        string       // SPEC identifier for quality context
}

type MergeResult struct {
    Merged        bool
    PRNumber      int
    Method        MergeMethod
    BranchDeleted bool
    MergedAt      time.Time
}

type PrerequisiteCheck struct {
    AllMet         bool      // All conditions satisfied
    AutoMergeFlag  bool      // --auto-merge was specified
    ReviewApproved bool      // PR review approved
    ChecksPassed   bool      // CI/CD checks passed
    QualityPassed  bool      // TRUST 5 quality gates passed
    Mergeable      bool      // No merge conflicts
    FailureReasons []string  // Unmet prerequisites
}
```

#### Prerequisite Checks

`CheckPrerequisites` verifies all merge conditions without merging:

| Check | Source | On Failure |
|-------|--------|-----------|
| Auto-merge flag | `opts.AutoMerge` | Adds to `FailureReasons` |
| Mergeability | `GHClient.PRView()` → `Mergeable` field | `CONFLICTING` → failure; `UNKNOWN` → allowed |
| Review approval | `PRReviewer.Review()` | Non-`APPROVE` → failure |
| Quality gates | Via review's `QualityReport.Passed` | `false` → failure |
| CI checks | `GHClient.PRChecks()` → `Overall` | Non-`pass` → failure |

`AllMet` is computed as: `AutoMergeFlag && ReviewApproved && ChecksPassed && QualityPassed && Mergeable`

When `RequireReview` is `false`, both `ReviewApproved` and `QualityPassed` are set to `true` (skipped). When `RequireChecks` is `false`, `ChecksPassed` is set to `true` (skipped).

#### Merge Flow

1. Verify `AutoMerge` is `true` (returns `ErrAutoMergeNotRequested` if false)
2. Run `CheckPrerequisites()` to verify all conditions
3. If `AllMet` is `false`, return `ErrMergeBlocked` with failure reasons
4. Apply default merge method (`merge`) if not specified
5. Execute `GHClient.PRMerge()` with the chosen method
6. Return `MergeResult` with merge details

### Sentinel Errors (internal/github/errors.go) — Milestones 3-4

| Error | Description |
|-------|-------------|
| `ErrPRNotFound` | Pull request does not exist |
| `ErrPRAlreadyExists` | PR already exists for branch |
| `ErrMergeBlocked` | Merge prerequisites unmet |
| `ErrMergeConflict` | Merge conflicts detected |
| `ErrCIFailed` | CI/CD checks failed |
| `ErrReviewRequired` | Review approval required |
| `ErrAutoMergeNotRequested` | `--auto-merge` flag not specified |

### Integration with Existing Workflow

The Milestone 3-4 components integrate into the full GitHub Issues workflow:

```
Issue → SPEC → Worktree → Plan-Run-Sync → Quality Gate → PR Review → Merge → Close
         ↑                      ↑                ↑            ↑          ↑
     Milestone 1-2         Milestone 3      Milestone 3   Milestone 4  Milestone 4
```

#### Connection Points

| From | To | Interface |
|------|----|-----------|
| `SpecLinker` (M1-2) | `WorktreeOrchestrator` (M3) | SPEC ID links worktree to issue |
| `WorktreeOrchestrator` (M3) | `WorktreeValidator` (M3) | Quality validation in worktree dir |
| `WorktreeOrchestrator.PrepareForReview` (M3) | `PRReviewer` (M4) | Review readiness gates PR creation |
| `PRReviewer` (M4) | `PRMerger` (M4) | Reviewer is a prerequisite of merger |
| `PRMerger` (M4) | `IssueCloser` (M5-6) | Successful merge triggers issue closure |

---

## Milestone 5: Issue Closure with Multilingual Comments

### Multilingual Comment Generation (internal/i18n)

#### Interface

```go
type CommentGenerator interface {
    Generate(langCode string, data *CommentData) (string, error)
}

type CommentData struct {
    Summary         string
    PRNumber        int
    IssueNumber     int
    MergedAt        time.Time
    TimeZone        string
    CoveragePercent int    // Zero omits the coverage line
}
```

#### Supported Languages

| Code | Language | Detection Source |
|------|----------|-----------------|
| `en` | English | Default fallback |
| `ko` | Korean (한국어) | `.moai/config/sections/language.yaml` |
| `ja` | Japanese (日本語) | `.moai/config/sections/language.yaml` |
| `zh` | Chinese (中文) | `.moai/config/sections/language.yaml` |

Language is detected from `language.conversation_language` in `.moai/config/sections/language.yaml`. Unsupported codes (e.g., `de`, `fr`) fall back to English.

#### Template Output Example (Korean)

```
✅ 이슈가 성공적으로 해결되었습니다!

**구현 내용:**
Added user authentication feature

**테스트 커버리지:** 92%

**관련 PR:** #456
**병합 시간:** 2026-02-16 16:30 KST

이슈를 자동으로 종료합니다. 추가 문제가 있으면 새 이슈를 생성해주세요.
```

#### Usage

```go
gen := i18n.NewCommentGenerator()
comment, err := gen.Generate(langCode, &i18n.CommentData{
    Summary:         "Added user authentication",
    PRNumber:        456,
    IssueNumber:     123,
    MergedAt:        time.Now(),
    TimeZone:        "KST",
    CoveragePercent: 92,
})
```

### Issue Closure (internal/github/issue_closer.go)

#### Interface

```go
type IssueCloser interface {
    Close(ctx context.Context, issueNumber int, comment string) (*CloseResult, error)
}

type CloseResult struct {
    IssueNumber   int
    CommentPosted bool
    LabelAdded    bool
    IssueClosed   bool
}
```

#### 3-Step Closure Process

| Step | Command | Critical | On Failure |
|------|---------|----------|------------|
| 1. Post comment | `gh issue comment {n} --body "{comment}"` | Yes | Abort, return error |
| 2. Add label | `gh issue edit {n} --add-label resolved` | No | Log warning, continue |
| 3. Close issue | `gh issue close {n}` | Yes | Abort, return error |

#### Retry Logic

Each step retries independently with exponential backoff:

- Default: 3 retries with 2s base delay (2s, 4s, 8s)
- Configurable via `WithMaxRetries(n)` and `WithRetryDelay(d)`
- Context cancellation interrupts retry waits immediately
- Returns `*RetryError` wrapping `ErrMaxRetriesExceeded` on exhaustion

#### Partial Failure Handling

`CloseResult` tracks which steps succeeded, enabling recovery:

```go
closer := github.NewIssueCloser(repoRoot,
    github.WithMaxRetries(3),
    github.WithRetryDelay(2 * time.Second),
)
result, err := closer.Close(ctx, issueNumber, comment)
// result.CommentPosted, result.LabelAdded, result.IssueClosed
```

### Workflow: Close Issue

The full issue closure workflow combines i18n and github packages:

1. Read `conversation_language` from `.moai/config/sections/language.yaml`
2. Generate multilingual comment via `i18n.CommentGenerator.Generate()`
3. Execute 3-step closure via `github.IssueCloser.Close()`
4. Update SPEC status to `completed` if SPEC exists

### Additional Errors (internal/github/errors.go)

| Error | Description |
|-------|-------------|
| `ErrCommentFailed` | Comment posting failed |
| `ErrCloseFailed` | Issue close failed |
| `ErrLabelFailed` | Label addition failed |
| `ErrMaxRetriesExceeded` | All retry attempts exhausted |

Custom type: `RetryError{Operation, Attempts, LastError}` implements `error` and `Unwrap()`.

---

## Milestone 6: tmux Session Automation

### tmux Detection (internal/tmux/detector.go)

#### Interface

```go
type Detector interface {
    IsAvailable() bool
    Version() (string, error)
}
```

Uses `exec.LookPath("tmux")` and `tmux -V` to check availability. If tmux is unavailable, the workflow falls back to sequential execution.

### Session Management (internal/tmux/session.go)

#### Interface

```go
type SessionManager interface {
    Create(ctx context.Context, cfg *SessionConfig) (*SessionResult, error)
}

type SessionConfig struct {
    Name       string       // e.g., "github-issues-2026-02-16-18-30"
    Panes      []PaneConfig // One pane per SPEC worktree
    MaxVisible int          // Default 3; panes beyond this use horizontal splits
}

type PaneConfig struct {
    SpecID  string // e.g., "SPEC-ISSUE-123"
    Command string // e.g., "moai worktree go SPEC-ISSUE-123"
}

type SessionResult struct {
    SessionName string
    PaneCount   int
    Attached    bool
}
```

#### Layout Algorithm

| Pane | Split Direction | tmux Command |
|------|----------------|--------------|
| 1 (first) | Created with session | `tmux new-session -d -s {name}` |
| 2 to MaxVisible | Vertical | `tmux split-window -v -t {name}` |
| MaxVisible+1 onward | Horizontal | `tmux split-window -h -t {name}` |

After all panes are created:
- Focus returns to pane 0: `tmux select-pane -t {name}:0.0`
- Layout rebalanced: `tmux select-layout -t {name} tiled`

#### Error Handling

- Session creation failure: Fatal, returns wrapped error
- Split-window failure: Non-fatal, logs warning, continues with remaining panes
- Send-keys failure: Non-fatal, logs warning, continues

#### Usage

```go
detector := tmux.NewDetector()
if !detector.IsAvailable() {
    // Fall back to sequential mode
}

mgr := tmux.NewSessionManager()
result, err := mgr.Create(ctx, &tmux.SessionConfig{
    Name: "github-issues-20260216-1630",
    Panes: []tmux.PaneConfig{
        {SpecID: "SPEC-ISSUE-1", Command: "moai worktree go SPEC-ISSUE-1"},
        {SpecID: "SPEC-ISSUE-2", Command: "moai worktree go SPEC-ISSUE-2"},
    },
    MaxVisible: 3,
})
```

### Sentinel Errors (internal/tmux/errors.go)

| Error | Description |
|-------|-------------|
| `ErrTmuxNotFound` | tmux binary not in PATH |
| `ErrSessionExists` | Session with same name already exists |
| `ErrSessionFailed` | Failed to create or manage session |
| `ErrNoPanes` | No panes configured in SessionConfig |

### Integration with /moai:github Slash Command

The `--tmux` flag on `/moai:github issues` triggers tmux automation:

1. Check tmux availability via `Detector.IsAvailable()`
2. Create session with one pane per issue worktree via `SessionManager.Create()`
3. Each pane auto-executes `moai worktree go SPEC-ISSUE-{number}`
4. Developer sees all worktrees side-by-side in tmux

---

## Integration Points (Milestones 5-6)

### With /moai:github Slash Command

The `/moai:github issues` workflow uses Milestone 5-6 helpers:

1. **Issue Closure** (Phase 5): After PR merge, `IssueCloser.Close()` posts multilingual comment and closes issue
2. **tmux Automation** (Phase 7, optional): With `--tmux` flag, `SessionManager.Create()` opens parallel worktrees

### With manager-spec Agent

After issue closure:
- If SPEC document exists (`SPEC-ISSUE-{number}`), update metadata `status` to `completed`

### Decoupled Architecture

The i18n and github packages are intentionally independent:
- Caller generates comment string via `i18n.CommentGenerator.Generate()`
- Caller passes comment string to `github.IssueCloser.Close()`
- No direct dependency between packages

---

Version: 1.1.0
Updated: 2026-02-16
Source: SPEC-GITHUB-WORKFLOW Milestones 1-2 + 5-6 implementation
