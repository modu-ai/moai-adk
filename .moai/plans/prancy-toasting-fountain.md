# Security Review: PR #390 — SPEC-GITHUB-WORKFLOW (Milestones 1-6)

## Context

PR #390 adds 50 files (+11,352/-104) implementing GitHub issue-to-SPEC conversion, worktree automation, PR review/merge, multilingual issue closure, and tmux session management. This review covers all new code for OWASP compliance, injection risks, authentication handling, data exposure, and concurrency safety.

---

## Findings Summary

| # | Severity | Category | File | Description |
|---|----------|----------|------|-------------|
| 1 | Important | Race Condition | `internal/github/spec_linker.go` | TOCTOU in concurrent registry file access |
| 2 | Important | Input Validation | `internal/cli/github.go` | Missing SPEC ID format validation in `link-spec` |
| 3 | Suggestion | Input Validation | `internal/tmux/session.go` | Tmux session name not sanitized |
| 4 | Suggestion | Info Leakage | `internal/github/gh.go` | gh CLI stderr forwarded in error messages |
| 5 | Suggestion | Defense-in-Depth | `internal/i18n/templates.go` | text/template with externally-sourced data |

**No Critical findings. No hardcoded credentials, API keys, or tokens found.**

---

## Detailed Findings

### IMPORTANT-1: Race Condition in SpecLinker File Operations

**File:** `internal/github/spec_linker.go` (load/save/LinkIssueToSpec methods)

**Description:** `fileSpecLinker` has no mutex or file locking. The read-check-append-write sequence in `LinkIssueToSpec` is vulnerable to TOCTOU:

```
Goroutine A: load() → reads file (0 mappings)
Goroutine B: load() → reads file (0 mappings)
Goroutine A: LinkIssueToSpec(1, "SPEC-1") → appends, saves (1 mapping)
Goroutine B: LinkIssueToSpec(2, "SPEC-2") → appends to stale state, saves (1 mapping — A's mapping lost!)
```

The concurrent test (`spec_linker_concurrent_test.go`) explicitly acknowledges this: *"The fileSpecLinker has no internal locking"* and *"Due to lack of file locking, multiple goroutines may succeed"*.

Atomic write (temp+rename) prevents corrupt partial writes but does NOT prevent lost updates.

**Impact:** Data loss when multiple `moai github link-spec` invocations run concurrently (e.g., in tmux parallel sessions). Lost registry mappings could cause orphaned SPECs.

**Recommendation:** Add `sync.Mutex` for in-process safety and `syscall.Flock`/`os.OpenFile` with `O_EXCL` for cross-process safety. Alternatively, document this as a known limitation since CLI commands are typically sequential.

---

### IMPORTANT-2: Missing SPEC ID Format Validation in CLI

**File:** `internal/cli/github.go:runLinkSpec` (lines ~148-155)

**Description:** The `link-spec` command validates issue number via `strconv.Atoi()` but only checks SPEC ID for emptiness:

```go
specID := args[1]
if specID == "" {
    return fmt.Errorf("spec ID cannot be empty")
}
```

Meanwhile, `worktree_orchestrator.go` enforces strict regex `^SPEC-ISSUE-\d+$`. A malformed SPEC ID (e.g., `../../../etc/passwd` or `SPEC; rm -rf /`) would be written to the registry JSON file. While this doesn't cause path traversal (the registry path is fixed), it creates inconsistent data that `worktree_orchestrator.go` would later reject.

**Impact:** Medium — invalid data in registry file; no direct exploit but breaks workflow consistency.

**Recommendation:** Reuse `specIDPattern` regex validation in `runLinkSpec` or create a `ValidateSpecID()` function in the github package.

---

### SUGGESTION-1: Tmux Session Name Not Sanitized

**File:** `internal/tmux/session.go:Create` and `sendKeys`

**Description:** `SessionConfig.Name` is used directly in tmux commands:
```go
m.run(ctx, "tmux", "new-session", "-d", "-s", cfg.Name)
m.run(ctx, "tmux", "send-keys", "-t", target, command, "Enter")
```

While `exec.Command` array arguments prevent shell injection, tmux itself interprets certain characters in session names (e.g., `.`, `:`, spaces). A session name containing `:` would conflict with tmux's `session:window.pane` target syntax.

**Impact:** Low — session names are currently generated internally (not from direct user input). But if future code passes user strings as session names, tmux command parsing could break.

**Recommendation:** Add a `validateSessionName` function that rejects or sanitizes names containing `:`, `.`, spaces, or non-printable characters.

---

### SUGGESTION-2: gh CLI Stderr Forwarded in Error Messages

**File:** `internal/github/gh.go:execGH` (line ~238)

**Description:**
```go
errMsg := strings.TrimSpace(stderr.String())
return "", fmt.Errorf("gh %s: %s: %w", args[0], errMsg, err)
```

The full stderr from `gh` CLI is included in error messages. While gh CLI errors are typically safe, in edge cases (misconfigured environment, debug mode), stderr could contain auth token fragments or internal server URLs.

**Impact:** Very low — gh CLI is well-behaved. This is a defense-in-depth concern.

**Recommendation:** Consider truncating stderr to a reasonable length (e.g., 500 chars) and/or sanitizing known patterns (URLs, tokens).

---

### SUGGESTION-3: text/template with Externally-Sourced Data

**File:** `internal/i18n/templates.go`

**Description:** Uses `text/template` (not `html/template`) for generating GitHub Markdown comments. The `CommentData.Summary` field is a string that gets interpolated into templates. While `text/template.Execute` does NOT re-evaluate template syntax in data values (so `{{.Summary}}` is safe even if Summary contains `{{...}}`), the output is posted as a GitHub comment which renders Markdown.

**Impact:** Very low — Summary comes from internal workflow, not raw user input. No XSS risk since GitHub sanitizes Markdown rendering.

**Recommendation:** Document that `CommentData.Summary` should be treated as trusted input.

---

## Positive Security Findings

| Area | Status | Details |
|------|--------|---------|
| **Command Injection** | SAFE | All `exec.Command` uses array arguments, never `sh -c` |
| **gh CLI Invocation** | SAFE | `execGH()` uses `exec.LookPath` + `CommandContext` |
| **Integer Inputs** | SAFE | All issue/PR numbers validated via `strconv.Atoi`/`strconv.Itoa` |
| **SPEC ID Validation** | SAFE (orchestrator) | `worktree_orchestrator.go` uses strict regex `^SPEC-ISSUE-\d+$` |
| **Path Traversal** | SAFE | Registry uses constant filename; worktree uses `filepath.Abs` + prefix matching |
| **Atomic File Writes** | SAFE | `spec_linker.go:save()` uses temp-file + rename pattern |
| **Authentication Check** | SAFE | `IsAuthenticated()` verifies `gh auth status` before operations |
| **No Hardcoded Secrets** | SAFE | No API keys, tokens, or credentials in source or tests |
| **Error Handling** | SAFE | Sentinel errors wrap properly via `fmt.Errorf("...: %w", err)` |
| **Structured Logging** | SAFE | Uses `slog` with module tags; no sensitive data logged |
| **Dependency Injection** | SAFE | All external dependencies use interfaces for testability |
| **Context Cancellation** | SAFE | All long operations use `context.Context` for cancellation |
| **Retry Logic** | SAFE | Exponential backoff with max retries (3) and context checking |
| **JSON Parsing** | SAFE | Uses `encoding/json.Unmarshal` with typed structs |
| **Corrupt File Recovery** | SAFE | `spec_linker.go:load()` renames corrupt files and starts fresh |

---

## OWASP Top 10 Compliance

| OWASP Category | Status | Notes |
|----------------|--------|-------|
| A01 Broken Access Control | N/A | CLI tool; relies on gh auth |
| A02 Cryptographic Failures | N/A | No cryptographic operations |
| A03 Injection | PASS | Array-based exec.Command throughout |
| A04 Insecure Design | NOTE | SpecLinker race condition (IMPORTANT-1) |
| A05 Security Misconfiguration | PASS | Proper defaults, no dangerous configs |
| A06 Vulnerable Components | N/A | Standard library + gh CLI only |
| A07 Auth Failures | PASS | IsAuthenticated() gate |
| A08 Data Integrity | NOTE | TOCTOU in SpecLinker (IMPORTANT-1) |
| A09 Logging Failures | PASS | Structured slog, no sensitive data |
| A10 SSRF | N/A | No outbound HTTP calls (uses gh CLI) |

---

## Verdict

**PR #390 demonstrates strong security practices overall.** The codebase consistently uses safe command execution patterns, proper input validation, interface-based design, and structured error handling.

The two Important findings (SpecLinker race condition and missing SPEC ID validation in CLI) should be addressed before merge or explicitly documented as known limitations. Neither represents an exploitable vulnerability in typical single-user CLI usage.
