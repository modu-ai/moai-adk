# Implementation Plan: Milestone 5-6 (Issue Closure & tmux Automation)

## Context

SPEC-GITHUB-WORKFLOW requires automated issue closure with multilingual comments (Milestone 5) and optional tmux session management for parallel development (Milestone 6). No `internal/github/`, `internal/i18n/`, or `internal/tmux/` packages exist yet -- all files are new, so TDD (RED-GREEN-REFACTOR) applies per `quality.yaml` hybrid mode.

---

## File Structure

```
internal/
  github/
    errors.go              # Sentinel errors + ExecError, RetryError types
    executor.go            # CommandExecutor interface + ghExecutor (wraps gh CLI)
    executor_test.go
    issue_closer.go        # IssueCloser interface + DefaultIssueCloser
    issue_closer_test.go
  i18n/
    errors.go              # Sentinel errors
    templates.go           # CommentGenerator interface + TemplateCommentGenerator
    templates_test.go
  tmux/
    errors.go              # Sentinel errors
    detector.go            # Detector interface + SystemDetector
    detector_test.go
    session.go             # SessionManager interface + DefaultSessionManager
    session_test.go
```

No modifications to existing files (deps.go wiring is outside my file ownership).

---

## Implementation Order (TDD)

### Step 1: `internal/i18n` (zero dependencies on other new packages)

**1a. `internal/i18n/errors.go`** - Sentinel errors:
- `ErrTemplateNotFound`, `ErrTemplateExecution`, `ErrInvalidData`

**1b. `internal/i18n/templates_test.go`** (RED) - Table-driven tests:
- `TestCommentGenerator_Generate_English` - langCode="en" produces English output
- `TestCommentGenerator_Generate_Korean` - langCode="ko" produces Korean output
- `TestCommentGenerator_Generate_Japanese` - langCode="ja" produces Japanese output
- `TestCommentGenerator_Generate_Chinese` - langCode="zh" produces Chinese output
- `TestCommentGenerator_Generate_Fallback` - langCode="de" falls back to English
- `TestCommentGenerator_Generate_EmptyLang` - langCode="" falls back to English
- `TestCommentGenerator_Generate_ContainsPRLink` - output contains `#456`
- `TestCommentGenerator_Generate_ContainsTimestamp` - output contains formatted time

**1c. `internal/i18n/templates.go`** (GREEN) - Implementation:

```go
// CommentData holds template variables for comment generation.
type CommentData struct {
    Summary         string
    PRNumber        int
    IssueNumber     int
    MergedAt        time.Time
    TimeZone        string
    CoveragePercent int
}

// CommentGenerator generates multilingual comments for GitHub issues.
type CommentGenerator interface {
    Generate(langCode string, data *CommentData) (string, error)
}
```

- Templates stored as Go `const` strings (no file I/O)
- Uses `text/template` for variable substitution
- 4 languages: en, ko, ja, zh
- Fallback: unknown lang code -> English
- Uses `pkg/models.IsValidLanguageCode()` for validation
- Reuses `pkg/models.LangNameMap` for supported languages

**Key dependencies**: `pkg/models/lang.go` (`IsValidLanguageCode`)

### Step 2: `internal/github/errors.go`

Sentinel errors following `internal/core/git/errors.go` pattern:
- `ErrGHNotFound` - gh CLI not in PATH
- `ErrNotAuthenticated` - gh not authenticated
- `ErrIssueNotFound` - issue not found
- `ErrCommentFailed` - comment post failed
- `ErrCloseFailed` - issue close failed
- `ErrLabelFailed` - label add failed
- `ErrMaxRetriesExceeded` - all retries exhausted

Custom error types:
- `ExecError{Command, ExitCode, Stderr}` - command execution failure
- `RetryError{Operation, Attempts, LastError}` - wraps final error after retries (implements `Unwrap()`)

### Step 3: `internal/github/executor.go` + tests

**3a. `internal/github/executor_test.go`** (RED):
- `TestGHExecutor_Run_Success` - valid command returns stdout
- `TestGHExecutor_Run_ExitCode` - non-zero exit returns ExecError
- `TestGHExecutor_Run_Timeout` - context cancellation returns context error

**3b. `internal/github/executor.go`** (GREEN):

```go
// CommandExecutor abstracts shell command execution for testability.
type CommandExecutor interface {
    Run(ctx context.Context, name string, args ...string) (string, error)
}
```

- `ghExecutor` struct wraps `exec.CommandContext`
- Captures stdout/stderr separately
- Non-zero exit -> `*ExecError` with stderr
- Compile-time check: `var _ CommandExecutor = (*ghExecutor)(nil)`
- Functional option: `WithWorkDir(dir string)`

### Step 4: `internal/github/issue_closer.go` + tests

**4a. `internal/github/issue_closer_test.go`** (RED) - Uses `mockExecutor`:
- `TestIssueCloser_Close_Success` - all 3 gh commands succeed -> full CloseResult
- `TestIssueCloser_Close_CommentFails` - comment fails all retries -> RetryError
- `TestIssueCloser_Close_RetryThenSuccess` - fails twice, succeeds third -> Retries=2
- `TestIssueCloser_Close_MaxRetries` - all retries fail -> ErrMaxRetriesExceeded
- `TestIssueCloser_Close_LabelFailsCloseSucceeds` - label fails, close succeeds -> partial result
- `TestIssueCloser_Close_ContextCancelled` - cancelled context -> context error

**4b. `internal/github/issue_closer.go`** (GREEN):

```go
// CloseResult holds the outcome of an issue closure operation.
type CloseResult struct {
    IssueNumber   int
    CommentPosted bool
    IssueClosed   bool
    LabelAdded    bool
    Retries       int
}

// IssueCloser posts a comment, adds a label, and closes a GitHub issue.
type IssueCloser interface {
    Close(ctx context.Context, issueNumber int, comment string) (*CloseResult, error)
}
```

- `DefaultIssueCloser` with fields: `executor`, `maxRetries` (default 3), `retryDelay` (default 2s)
- Functional options: `WithMaxRetries(n)`, `WithRetryDelay(d)`, `WithLogger(l)`
- `Close()` runs 3 sequential `gh` commands:
  1. `gh issue comment {n} --body "{comment}"`
  2. `gh issue edit {n} --add-label resolved`
  3. `gh issue close {n}`
- Each step has independent retry loop with exponential backoff
- Returns partial `CloseResult` on partial failure
- Compile-time check: `var _ IssueCloser = (*DefaultIssueCloser)(nil)`

### Step 5: `internal/tmux/errors.go` + `detector.go` + tests

**5a. `internal/tmux/errors.go`** - Sentinel errors:
- `ErrTmuxNotFound`, `ErrSessionExists`, `ErrSessionFailed`, `ErrNoPanes`

**5b. `internal/tmux/detector_test.go`** (RED):
- `TestSystemDetector_IsAvailable_Found` - tmux in PATH -> true
- `TestSystemDetector_IsAvailable_NotFound` - tmux not in PATH -> false
- `TestSystemDetector_Version_Success` - parses "tmux 3.4" -> "3.4"

**5c. `internal/tmux/detector.go`** (GREEN):

```go
type Detector interface {
    IsAvailable() bool
    Version() (string, error)
}
```

- `SystemDetector` uses `exec.LookPath("tmux")` + `tmux -V`
- Testable via `CommandRunner` injection

### Step 6: `internal/tmux/session.go` + tests

**6a. `internal/tmux/session_test.go`** (RED) - Uses `mockRunner`:
- `TestSessionManager_Create_SinglePane` - 1 pane, no splits
- `TestSessionManager_Create_ThreePanes` - 3 panes, vertical splits
- `TestSessionManager_Create_FourPanes` - 3 vertical + 1 horizontal overflow
- `TestSessionManager_Create_NoPanes` - ErrNoPanes
- `TestSessionManager_Create_CommandFailure` - tmux error propagated

**6b. `internal/tmux/session.go`** (GREEN):

```go
type PaneConfig struct {
    SpecID  string // e.g., "SPEC-ISSUE-123"
    Command string // e.g., "moai worktree go SPEC-ISSUE-123"
}

type SessionConfig struct {
    Name       string
    Panes      []PaneConfig
    MaxVisible int // default 3
}

type SessionResult struct {
    SessionName string
    PaneCount   int
    Attached    bool
}

type SessionManager interface {
    Create(ctx context.Context, cfg *SessionConfig) (*SessionResult, error)
}
```

- Layout algorithm: vertical splits for panes 1-3, horizontal for 4+
- Commands: `tmux new-session -d -s`, `split-window -v/-h`, `send-keys`, `select-pane`
- Rebalance: `tmux select-layout -t {name} tiled`

---

## Key Design Decisions

1. **gh CLI over GitHub REST API**: Reuses `gh auth`, no new HTTP dependencies, consistent with SPEC
2. **`text/template` over `fmt.Sprintf`**: Better for multilingual templates with conditional fields
3. **Decoupled packages**: `i18n` and `github` are independent. Caller generates comment via `i18n.CommentGenerator`, then passes string to `github.IssueCloser.Close()`
4. **`CommandExecutor` interface**: All `exec.Command` calls behind interface for full testability
5. **No deps.go modification**: Outside my file ownership. Team lead or integration task will wire into Dependencies

---

## Reference Files

| File | Purpose |
|------|---------|
| `internal/rank/client.go` | Interface + functional options + error types pattern |
| `internal/core/git/errors.go` | Sentinel error pattern |
| `internal/cli/mock_test.go` | Mock with function fields pattern |
| `internal/cli/deps.go` | Composition Root (for future integration) |
| `pkg/models/lang.go` | Language code validation + name mapping |
| `pkg/models/config.go` | `LanguageConfig` struct definition |
| `internal/shell/detect.go` | External tool detection pattern (for tmux Detector) |

---

## Verification

```bash
# Run all new package tests
go test -race -cover ./internal/github/... ./internal/i18n/... ./internal/tmux/...

# Verify coverage >= 85%
go test -coverprofile=coverage.out ./internal/github/... ./internal/i18n/... ./internal/tmux/...
go tool cover -func=coverage.out

# Lint check
golangci-lint run ./internal/github/... ./internal/i18n/... ./internal/tmux/...

# Vet check
go vet ./internal/github/... ./internal/i18n/... ./internal/tmux/...
```
