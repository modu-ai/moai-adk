# SPEC-CLIFIX-CONTRACT-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- 2026-07-10: plan-phase artifact set authored (spec.md / plan.md / acceptance.md) by manager-spec from CLI audit 2026-07-10 §4 + §5 P1. Status: draft. Pending plan-audit. Sequenced after SPEC-CLIFIX-CRITICAL-001.
- 2026-07-10: plan-audit iteration-1 verdict FAIL (score 0.83) — 8 blocking defects (D1-D8) remediated by manager-spec (re-delegation, plan-phase body edit only, no Go code touched, status remains draft). Fixed: D1 stale anchors (8→11 sites, correct line numbers); D2 spec↔plan anchor consistency (both cite the same 11-site list); D3 multi-site enumeration (hook.go 231+362, spec_lint.go 70+96, migrate_agency.go 650+652 all explicit; harness/execute.go 327→333, agentlint/workflow_lint.go 159→158 corrected); D4 AC-CONT-001-001 grep now excludes comment lines (`grep -vE ':[0-9]+:[[:space:]]*//'`) to avoid false-positive on hook_pre_push.go comment literals; D5 launch_exec_windows.go + update.go:487 documented as approved Windows process-replacement exceptions in plan.md §D Constraints; D6 spec_audit exit-1→exit-2 CI risk clause added to plan.md §G mirroring the spec_lint clause; D7 --confirm (syncConfirm) classified as dead flag (registered :63, never read), --yes (syncYes) canonical (wired :40, gated :205), REQ-CONT-001-003 updated to remove --confirm; D8 acceptance.md reordered §A→§B→§C→§D→§D.5 (§B traceability summary inserted). Finding beyond orchestrator scope: update.go:487 is a third Windows process-replacement exception the orchestrator's D5 did not name — added to plan.md §D and acceptance.md §D AC-CONT-001-001 expected-outcome so the AC grep is self-consistent (otherwise update.go:487 would match as a non-exception site). Pending re-audit iteration 2.
- 2026-07-10: plan-audit iteration-2 verdict **PASS** (score 0.88; Tier M threshold 0.80; monotonic improvement over iter-1 0.83). D1-D8 all RESOLVED — auditor independently re-verified against the live tree (11 os.Exit removal sites, comment-exclusion filter, update.go:487 Windows runtime gate, --confirm dead/--yes live). 2 new minor debt introduced by remediation (N1: spec.md §D AC-CONT-001-003 still names --confirm, but acceptance.md §D SSOT correctly uses --yes; N2: plan.md §F M1 says "8 sites" actual 11) — neither blocking; deferred to sync-phase. Remediation committed at 5b66a683a (push). Implementation Kickoff Approval obtained (user-selected run-phase entry). → Phase 0.95 Mode 5 + run-phase.

## §E.2 Run-phase Evidence

All 8 ACs PASS. Evidence per acceptance.md §D AC Matrix (verbatim commands run against the worktree at origin/main + the 4 run-phase commits):

| AC | Status | Verification command | Actual outcome |
|---|---|---|---|
| AC-CONT-001-001 | PASS | `grep -rn 'os\.Exit' internal/cli internal/cli/harness internal/cli/agentlint --include='*.go' \| grep -v '_test.go' \| grep -vE ':[0-9]+:[[:space:]]*//'` | Only `launch_exec_windows.go:37,41` + `update.go:487` (Windows process-replacement exceptions). Zero os.Exit in RunE/PostRunE bodies of the 8 files. |
| AC-CONT-001-002 | PASS | `go test ./internal/cli/ -run 'GithubDryRun' -count=1 -v` | `--- PASS: TestGithubDryRun` — --dry-run prints planned mutation, no registry write |
| AC-CONT-001-003 | PASS | `go test ./internal/cli/ -run 'SpecStatusConfirm' -count=1 -v` | `--- PASS` — --confirm removed; non-TTY without --yes aborts (no hang, no mutation) |
| AC-CONT-001-004 | PASS | `go test ./internal/cli/ -run 'AstgrepExitCode' -count=1 -v` | `--- PASS` for text/json/sarif — error findings → exit 1 under all formats |
| AC-CONT-001-005 | PASS | `go test ./internal/cli/ -run 'ExitCodeContract' -count=1 -v` | `--- PASS` — spec lint invalid args=3, spec audit MUST-FIX=2, constitution=2 |
| AC-CONT-001-006 | PASS | `go test ./internal/cli/ -run 'PrePushStderr' -count=1 -v` | `--- PASS: TestPrePushStderr` — exit 2 + violation details on stderr |
| AC-CONT-001-007 | PASS | `ls internal/cli/team_spawn_lock_unix_test.go && go test ./internal/cli/ -run 'Flock\|ClaimLock' -count=1 -v` | File exists with `_test.go` suffix; `--- PASS: TestFlock`; `go list -deps ./cmd/moai \| grep -c '^testing$'` = 0 |
| AC-CONT-001-008 | PASS | `go test ./internal/cli/... -run 'HelpExitContract' -count=1 -v` | `--- PASS` for SpecLint/SpecAudit/Constitution — declared exit codes match produced |

Exit-code contract table (command → declared → produced):

| Command | Declared (help text) | Produced (AC test) |
|---|---|---|
| `spec lint` invalid args | `3 = invalid arguments` | 3 — ExitCodeContract_SpecLintInvalidArgs |
| `spec audit --strict` MUST-FIX | `2 = strict mode + MUST-FIX` | 2 — ExitCodeContract_SpecAuditMustFix |
| `spec lint` HasErrors | `1 = errors found` | 1 — exitCodeError{1} (M1) |
| `spec lint` linter crash | `2 = linter crash` | 2 — exitCodeError{2} (M1) |
| `constitution validate` missing source | `2=fatal (missing source file)` | 2 — ExitCodeContract_Constitution |
| `astgrep` error findings | `1` (AC4, all formats) | 1 — AstgrepExitCode text/json/sarif |
| `spec drift --exit-code-on-drift` | `Exit 1 if drift` | 1 — exitCodeError{1} (M1) |
| `hook`/`hook agent` ExitCode==2 | (generic plumbing) | 2 — exitCodeError{2} (M1) |
| `migrate agency` ErrMigrateNoSource | (system error) | 2 — exitCodeError{2} (M1) |
| `migrate agency` other MigrateError | (user error) | 1 — exitCodeError{1} (M1) |
| `harness execute` Apply error | (ExitCodeForError classification) | 1/2 per error type (M1) |
| `workflow lint` malformed YAML | `Exit 2` | 2 — exitCodeError{2} (M1) |
| `pre-push` enforce+violations | exit 2 (Claude Code protocol) | 2 — PrePushStderr |

Approved os.Exit exceptions (NOT removal targets, verified remaining): `launch_exec_windows.go:37,41` (Windows process-replacement) + `update.go:487` (Windows re-exec boundary, runtime-gated `if runtime.GOOS == "windows"`). `cmd/moai/main.go` ExitCoder boundary is out of the grep scope (verified separately by AC-CONT-001-008 + TestExitCodeErrorSatisfiesExitCoder).

Cross-platform build: `go build ./...` exit 0 (darwin); `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (windows). `go vet ./internal/cli/... ./cmd/moai/...` clean. `golangci-lint run --timeout=3m ./internal/cli/...` → 0 issues (after a self-introduced errcheck fix in a test file, fixed in-run).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-07-10"
run_commit_sha: "64ebeacca"  # M4 (last run-phase commit); M1=6027424d4 M2=6fa48a7e9 M3=017f176f3 M4=64ebeacca
run_status: "complete"
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "origin/main was 3 commits ahead at spawn (parallel WEB-CONSOLE-012 docs commits, no scope-file overlap); worktree rebased onto origin/main tip 99ad0a898 cleanly before M1"
l44_post_push_fetch: "pending (push in closure step)"
new_warnings_or_lints_introduced: 0  # 1 self-introduced errcheck in a test file, fixed in-run
cross_platform_build:
  darwin: "exit 0"
  windows: "exit 0 (GOOS=windows GOARCH=amd64)"
total_run_phase_files: 14  # 8 source conversions + cmd boundary (no change) + 3 new test files + flock rename + spec.md frontmatter
m1_to_mN_commit_strategy: "per-milestone Conventional Commits (M1 ExitCoder adoption / M2 exit-code contracts / M3 flags+streams / M4 test enablement+closure) + this evidence update"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: "2026-07-11"
sync_commit_sha: "pending-backfill"  # populated by follow-up commit (D3 self-referential-hazard exemption)
sync_status: "complete"
ac_pass_count: 8
ac_fail_count: 0
changelog_entry_added: true
n1_fixed: true  # spec.md §D AC-CONT-001-003 --confirm → --yes
n2_fixed: true  # plan.md §F M1 "8 os.Exit sites" → "11 os.Exit sites across 8 files"
frontmatter_status_completed: true  # spec.md frontmatter status: in-progress → implemented → completed (3-phase close)
breaking_change_documented: true  # REQ-CONT-001-005: moai spec audit MUST-FIX drift exit 1→2 (CI pipeline impact)
```

## §F Phase 0.95 Mode Selection

- tier: M
- scope: ~10-12 files (8 os.Exit source files + main.go ExitCoder boundary mapping + new contract/flock tests)
- domain count: 1-2 (internal/cli primary; harness/ + agentlint/ sub-packages)
- file language mix: 100% Go
- concurrency benefit: LOW (coding-heavy + inter-file dependency — ExitCoder boundary mapping at main.go coordinates with the 8 source-file os.Exit removals)
- Agent Teams prereqs: OFF (workflow.team.enabled default false)

Mode evaluation: trivial=No (8 REQ, multi-milestone) | background=No (write-heavy implementation) | agent-team=No (prereqs off; coding-heavy) | parallel=No (coding-heavy per Anthropic parallelism caveat) | workflow=No (inter-file dependency, not mechanical-uniform ≥30 files) | **sub-agent=SELECTED**

Decision: sub-agent (Mode 5)
Justification: coding-heavy + inter-file dependency; the ExitCoder boundary mapping at main.go must coordinate with the 8 source-file os.Exit removals, so sequential per-milestone is safer than parallel fan-out per Anthropic's coding-task parallelism caveat. manager-develop, cycle_type=tdd, M1→M2→M3→M4 sequential per plan.md §F.
Implementation Kickoff Approval: obtained (user selected run-phase entry; score-independent gate cleared).
