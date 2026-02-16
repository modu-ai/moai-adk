# Implementation Plan: Milestone 1-2 (SPEC-GITHUB-WORKFLOW)

## Context

SPEC-GITHUB-WORKFLOW requires building GitHub issue-to-SPEC integration and worktree auto-creation with intelligent branch naming. This plan covers Milestones 1 (Issue Parser & Linker) and 2 (Branch Prefix Detection). The worktree CLI commands (`new`, `go`, `done`, etc.) already exist in `internal/cli/worktree/` and will be reused.

## Milestone 1: Issue to SPEC Conversion

### 1.1 New Package: `internal/github/`

**File: `internal/github/issue_parser.go` (~120 lines)**

- `Issue` struct with JSON tags matching `gh issue view --json` output:
  ```go
  type Issue struct {
      Number   int       `json:"number"`
      Title    string    `json:"title"`
      Body     string    `json:"body"`
      Labels   []Label   `json:"labels"`
      Author   Author    `json:"author"`
      Comments []Comment `json:"comments"`
  }
  type Label struct { Name string `json:"name"` }
  type Author struct { Login string `json:"login"` }
  type Comment struct { Body string `json:"body"`; Author Author `json:"author"` }
  ```
- `IssueParser` interface + `ghIssueParser` concrete impl for testability
- `ParseIssue(ctx context.Context, number int) (*Issue, error)` using `exec.CommandContext` to run `gh issue view {number} --json number,title,body,labels,author,comments`
- Follow existing `execGit` pattern from `internal/core/git/manager.go` (LookPath, context timeout, stderr capture)
- Validate required fields (number > 0, title non-empty)

**File: `internal/github/issue_parser_test.go` (~150 lines)**

- TDD: Write tests first
- Table-driven tests for JSON parsing (valid issue, missing fields, empty body)
- Mock-based tests using `IssueParser` interface (no real `gh` calls)
- Test validation logic

**File: `internal/github/spec_linker.go` (~130 lines)**

- `Registry` struct: `Version string`, `Mappings []SpecMapping`
- `SpecMapping` struct: `IssueNumber int`, `SpecID string`, `CreatedAt time.Time`, `Status string`
- `SpecLinker` interface + `fileSpecLinker` concrete impl
- `LinkIssueToSpec(issueNum int, specID string) error` - add mapping + atomic write
- `GetLinkedSpec(issueNum int) (string, error)` - lookup by issue number
- `GetLinkedIssue(specID string) (int, error)` - reverse lookup
- Registry stored at `{projectRoot}/.moai/github-spec-registry.json`
- Follow `internal/manifest/manifest.go` pattern: atomic write (temp file + rename), graceful handling of missing file, backup corrupt files

**File: `internal/github/spec_linker_test.go` (~150 lines)**

- TDD: Write tests first
- Table-driven tests using `t.TempDir()` for file isolation
- Test: link creation, duplicate detection, lookup hit/miss, corrupt file recovery, concurrent access

### 1.2 CLI Command: `internal/cli/github.go` (~100 lines)

- Register `moai github` parent command on `rootCmd`
- `parse-issue` subcommand: `moai github parse-issue <number>`
  - Calls `IssueParser.ParseIssue()`
  - Outputs formatted issue summary using `wtCard` style helper (reuse render pattern from worktree)
- `link-spec` subcommand: `moai github link-spec <issue-number> <spec-id>`
  - Calls `SpecLinker.LinkIssueToSpec()`
  - Outputs success card

**File: `internal/cli/github_test.go` (~100 lines)**

- Mock-based tests for CLI commands
- Test argument validation, output formatting, error handling

### 1.3 Defs Extension

**File: `internal/defs/files.go` (edit)**

- Add `GithubSpecRegistryJSON = "github-spec-registry.json"` constant

---

## Milestone 2: Worktree Auto-Creation

### 2.1 Branch Prefix Detection

**File: `internal/git/branch_detector.go` (~60 lines)**

- `DetectBranchPrefix(labels []string) string` - pure function
- Label-to-prefix mapping:
  - `"bug"` -> `"fix/"`
  - `"feature"` or `"enhancement"` -> `"feat/"`
  - `"documentation"` or `"docs"` -> `"docs/"`
  - Default (no match) -> `"feat/"`
- `FormatIssueBranch(labels []string, issueNumber int) string` - combines prefix + `issue-{number}`
  - Example: `fix/issue-123`, `feat/issue-456`

**File: `internal/git/branch_detector_test.go` (~80 lines)**

- TDD: Write tests first
- Table-driven tests covering all label mappings
- Test: single label, multiple labels (first match wins by priority), no labels, case sensitivity

### 2.2 Existing Worktree CLI (No New Files Needed)

The worktree commands already exist in `internal/cli/worktree/`:
- `new.go` - `moai worktree new [branch-name]` with `--path` and `--base` flags
- `go.go` - `moai worktree go [branch-name]` outputs path for shell eval
- `done.go` - `moai worktree done [branch-name]` with `--force` and `--delete-branch`

The existing `resolveSpecBranch()` in `new.go` already handles `SPEC-XXX-NNN` -> `feature/SPEC-XXX-NNN`. For SPEC-ISSUE patterns, the same logic applies since `SPEC-ISSUE-123` has 3+ parts.

**No modifications needed** to existing worktree commands for Milestone 2. The `branch_detector.go` integrates at a higher level (issue-based worktree creation uses `FormatIssueBranch` to generate the branch name, then calls `moai worktree new` with that name).

---

## Files to Create (New)

| File | Lines | Purpose |
|------|-------|---------|
| `internal/github/issue_parser.go` | ~120 | Parse GitHub issues via gh CLI |
| `internal/github/issue_parser_test.go` | ~150 | Tests for issue parser |
| `internal/github/spec_linker.go` | ~130 | Bidirectional issue-SPEC linking |
| `internal/github/spec_linker_test.go` | ~150 | Tests for spec linker |
| `internal/cli/github.go` | ~100 | `moai github` CLI commands |
| `internal/cli/github_test.go` | ~100 | Tests for CLI commands |
| `internal/git/branch_detector.go` | ~60 | Label-to-branch-prefix mapping |
| `internal/git/branch_detector_test.go` | ~80 | Tests for branch detector |

## Files to Edit (Existing)

| File | Change |
|------|--------|
| `internal/defs/files.go` | Add `GithubSpecRegistryJSON` constant |

**Total: ~890 lines (implementation + tests)**

---

## Dependencies

- `gh` CLI for issue parsing (runtime dependency, mocked in tests)
- `internal/core/git/worktree.go` for worktree operations (existing)
- `internal/manifest/manifest.go` patterns for JSON registry (reference only)
- `internal/defs/` for shared constants

## Development Order (TDD for all new code)

1. `internal/git/branch_detector_test.go` -> `internal/git/branch_detector.go`
2. `internal/github/issue_parser_test.go` -> `internal/github/issue_parser.go`
3. `internal/github/spec_linker_test.go` -> `internal/github/spec_linker.go`
4. `internal/defs/files.go` (add constant)
5. `internal/cli/github_test.go` -> `internal/cli/github.go`

## Reusable Patterns

- **Command execution**: `exec.LookPath` + `exec.CommandContext` from `internal/core/git/manager.go:execGit()`
- **JSON registry**: Atomic write from `internal/manifest/manifest.go:Save()`
- **CLI commands**: Factory pattern from `internal/cli/worktree/new.go:newNewCmd()`
- **Testing**: Mock provider swap from `internal/cli/worktree/subcommands_test.go`
- **Output cards**: `wtSuccessCard` / `wtCard` style from `internal/cli/worktree/render.go`

## Verification

1. Run `go test ./internal/git/...` - branch detector tests pass
2. Run `go test ./internal/github/...` - issue parser and spec linker tests pass
3. Run `go test ./internal/cli/...` - CLI command tests pass
4. Run `go vet ./...` - no vet warnings
5. Run `golangci-lint run ./internal/github/... ./internal/git/... ./internal/cli/github*.go` - clean lint
6. Verify `go build ./cmd/moai/` compiles successfully
