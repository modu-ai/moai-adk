# acceptance.md — SPEC-WORKTREE-BRANCH-GUARD-DISCRIM-001

> Verification layer. Given-When-Then scenarios. Each AC is binary-testable
> and names a command whose verbatim output decides PASS/FAIL per
> `.claude/rules/moai/core/verification-claim-integrity.md` §3.

## §A. Scope bound to acceptance

This acceptance matrix binds ONLY the discriminant directory correction and its
non-regression on the -001 / -OPTIN-001 contracts. It does NOT restate the
predecessor ACs (AC-WBG-001..013 of -001; AC-REQ-1..6 of -OPTIN-001); those
remain in force and are referenced as regression anchors where applicable.

## §B. Test infrastructure note

AC-WBG-D-001, AC-WBG-D-002, AC-WBG-D-003, AC-WBG-D-004, AC-WBG-D-005 require a
real git worktree fixture. The fixture creates a primary repository under
`t.TempDir()`, adds a worktree via `git worktree add`, and uses the two paths
as `input.CWD` values. The existing `execCommand` indirection
(`internal/hook/branch_guard.go:29`) is used ONLY for the fallback path test
(AC-WBG-D-005 simulates a non-git cwd; the -001 AC-WBG-005 fallback test for
older-git rejection remains unchanged). Where the test environment lacks `git`,
`t.Skip` is acceptable; CI runners have `git`.

## §C. Given-When-Then AC Matrix

### AC-WBG-D-001 — Worktree discriminant returns false

**Given** a real git repository at `<primary>` (under `t.TempDir()`) with a
worktree added at `<worktree>` via `git worktree add <worktree>`.

**When** `isPrimaryCheckout(<worktree>)` is called.

**Then** the result is `(false, nil)`.

**Command** (verification):
```bash
go test ./internal/hook/... -run 'TestIsPrimaryCheckout_Worktree' -count=1 -v
```
**PASS**: `--- PASS: TestIsPrimaryCheckout_Worktree` and no `FAIL` line in the
output.

### AC-WBG-D-002 — Worktree allow end-to-end

**Given** a `preToolHandler` with `Workflow.BranchGuard.Enabled = true`, and a
`HookInput` with `ToolName = "Bash"`, `ToolInput = {"command": "git rebase
origin/main"}`, `CWD = <worktree>` (a real worktree path).

**When** `handler.Handle(ctx, input)` is called.

**Then** the returned `HookOutput.hookSpecificOutput.permissionDecision` equals
`"allow"` (NOT `"deny"`; no `BRANCH_GUARD_VIOLATION` reason).

**Command**:
```bash
go test ./internal/hook/... -run 'TestBranchGuard_Worktree_Allow' -count=1 -v
```
**PASS**: `--- PASS:` and the assertion on `permissionDecision == "allow"`.

### AC-WBG-D-003 — Primary deny preserved (regression)

**Given** a `preToolHandler` with `Workflow.BranchGuard.Enabled = true`, and a
`HookInput` with `ToolName = "Bash"`, `ToolInput = {"command": "git switch
feature-x"}`, `CWD = <primary>` (the primary checkout path), and
`AgentType != "manager-git"` and `MOAI_BRANCH_GUARD_EXEMPT` unset.

**When** `handler.Handle(ctx, input)` is called.

**Then** the returned decision is `DecisionDeny` and the reason starts with
`BRANCH_GUARD_VIOLATION: git switch`.

**Command**:
```bash
go test ./internal/hook/... -run 'TestBranchGuard_Primary_Still_Denies' -count=1 -v
```
**PASS**: `--- PASS:` and the deny-reason assertion. This is the regression
anchor — failure here means the fix over-corrected and broke the primary path.

### AC-WBG-D-004 — Audit-log placement on primary (not worktree)

**Given** a git repository at `<primary>` with a worktree at `<worktree>`, AND
the `execCommand` indirection swapped to simulate `git rev-parse` failure at
`<worktree>` (forcing the fail-open path), AND `input.CWD = <worktree>`.

**When** `checkBranchState(input, <primary>)` is called (or whichever signature
the chosen seam produces; the audit-log project dir resolves to `<primary>`).

**Then** the fail-open advisory is appended to
`<primary>/.moai/logs/branch-guard-audit.log` (the file exists at that path
post-call), AND no file is created at
`<worktree>/.moai/logs/branch-guard-audit.log`.

**Command**:
```bash
go test ./internal/hook/... -run 'TestBranchGuard_AuditLog_Placement' -count=1 -v
```
**PASS**: `--- PASS:` AND a `stat`/`os.Lstat` assertion that
`<worktree>/.moai/logs/branch-guard-audit.log` does NOT exist.

### AC-WBG-D-005 — Fail-open preserved at non-git command cwd

**Given** `input.CWD = <t.TempDir()>` (a non-git directory), the
`execCommand` indirection NOT swapped (real `git` invoked; `git rev-parse` at a
non-git dir exits non-zero), AND `Workflow.BranchGuard.Enabled = true`.

**When** `handler.Handle(ctx, input)` is called with a branch-state Bash
command.

**Then** the returned `permissionDecision` is `"allow"` (fail-open) AND the
advisory audit entry is written per AC-WBG-D-004 placement.

**Command**:
```bash
go test ./internal/hook/... -run 'TestBranchGuard_FailOpen_NonGitCwd' -count=1 -v
```
**PASS**: `--- PASS:` and the allow-decision assertion.

### AC-WBG-D-006 — Exemption unchanged

**Given** the same setup as AC-WBG-D-003 (primary checkout, branch-state
command), BUT `MOAI_BRANCH_GUARD_EXEMPT=1` is set via `t.Setenv` (the test
framework's env-var helper).

**When** `handler.Handle(ctx, input)` is called.

**Then** the returned `permissionDecision` is `"allow"` (exemption fires before
the discriminant).

**Command**:
```bash
go test ./internal/hook/... -run 'TestBranchGuard_Exempt' -count=1 -v
```
**PASS**: `--- PASS:`. This test SHOULD already exist from -001 (AC-WBG-011);
verify it is unchanged and still passes.

### AC-WBG-D-007 — Latency ceiling preserved (≤ 500ms)

**Given** a primary checkout fixture, `Workflow.BranchGuard.Enabled = true`,
and a branch-state Bash command in `input.ToolInput`.

**When** `handler.Handle(ctx, input)` is invoked N=100 times and the wall-time
per invocation is measured.

**Then** the median per-invocation wall-time is ≤ 500ms.

**Command**:
```bash
go test ./internal/hook/... -run 'TestBranchGuard_Latency' -count=1 -v
```
**PASS**: `--- PASS:` and the recorded median in the test output satisfies
the bound. The existing TestBranchGuard_Latency from -001 SHOULD be reused;
this AC asserts the ceiling is preserved post-fix, not lowered.

### AC-WBG-D-008 — Doctrine v1.2.0 + sanitized-pair mirror parity

**Given** the doctrine rule at
`.claude/rules/moai/workflow/main-checkout-branch-guard.md` and its sanitized
mirror at
`internal/template/templates/.claude/rules/moai/workflow/main-checkout-branch-guard.md`.

**When** the source rule is grepped for the version line and the sanitized-pair
parity test is run.

**Then** (a) `grep '^Version: 1.2.0' <source-rule>` returns exactly one line,
AND (b) `TestSanitizedPairParity` passes (the mirror is in parity with the
source modulo the §25 sanitization), AND (c)
`TestTemplateNoInternalContentLeak` passes for the mirror (no SPEC-ID / REQ
tokens leaked into the distributed template).

**Commands**:
```bash
grep -n '^Version: 1.2.0' .claude/rules/moai/workflow/main-checkout-branch-guard.md
go test ./internal/template/... -run 'TestSanitizedPairParity|TestTemplateNoInternalContentLeak' -count=1 -v
```
**PASS**: (a) exactly one matching line `Version: 1.2.0`; (b) and (c) both
`--- PASS:`.

## §D. Severity classification

| AC | Severity | Rationale |
|----|----------|-----------|
| AC-WBG-D-001 | MUST-PASS | headline fix; without it the bug persists |
| AC-WBG-D-002 | MUST-PASS | end-to-end demonstration of the fix |
| AC-WBG-D-003 | MUST-PASS | regression anchor for the primary-checkout deny |
| AC-WBG-D-004 | MUST-PASS | audit-log placement invariant; silent failure mode if violated |
| AC-WBG-D-005 | MUST-PASS | fail-open contract preserved (-001 REQ-WBG-012) |
| AC-WBG-D-006 | MUST-PASS | exemption unchanged (-OPTIN-001 REQ-6) |
| AC-WBG-D-007 | SHOULD-PASS | latency ceiling; failure is a perf regression, not a correctness bug |
| AC-WBG-D-008 | MUST-PASS | doctrine + template-parity discipline |

## §E. Indirect verification (covered by existing tests, not re-run)

- `branchStatePatterns` regex set (the -OPTIN-001 read-only refinement) is
  unchanged; existing -OPTIN-001 tests cover `git stash list`, `git stash show`,
  `git merge-base` exclusion. No new test required here.
- The `--path-format=absolute` primary path and the older-git fallback
  (`--absolute-git-dir` + cwd-normalized `--git-common-dir`) are unchanged;
  existing -001 AC-WBG-005 tests cover them. The fix changes only WHICH
  directory is queried, not HOW.

## §F. Closure gate (Definition of Done)

- All MUST-PASS AC green; AC-WBG-D-007 SHOULD-PASS with the recorded median
  documented in the §E self-verification report.
- `go test ./internal/hook/... -count=1` exits 0 (full package green, not just
  the new tests).
- `go test ./internal/template/... -count=1` exits 0 (sanitized-pair parity +
  no-leak guards green).
- `golangci-lint run --timeout=2m` reports no NEW findings in the changed
  files.
- `GOOS=windows GOARCH=amd64 go build ./...` exits 0.
- Doctrine rule carries `Version: 1.2.0` on both source and sanitized mirror.
- Commit + PR on a feature branch per Route B (repo-local PR-mandatory
  policy); `--no-verify` NOT used.

## §G. Forward-looking check (advisory, not blocking)

- If the fix surfaces that Claude Code does NOT reliably populate `input.CWD`
  for PreToolUse Bash events when the agent is operating inside a worktree (the
  concern raised in plan.md §C seam decision), STOP and return a blocker report
  rather than silently falling back to `$CLAUDE_PROJECT_DIR` (AP-D-003). A
  follow-up SPEC may be required to add an explicit `worktree_path` field to
  HookInput, analogous to the v2.1.49+ `WorktreePath` field already present at
  `types.go:277`.
