auditor-model: claude-sonnet-5-5[1m]

# Delta sync-audit — card t1452 (SPEC-VERIFY-RUN-REUSE-001)

verdict: PASS-WITH-DEBT
audited_sha: 28843a6d1868616e67605086540c012b37b10146

Score: 90/100 (delta scope: repair a05ec62de + evidence 28843a6d1; baseline sync-audit.md was FAIL 62).

## Claim

1. F0 (P1, `moai verify run` unregistered in the real CLI) is fixed: the verb is registered directly in `verify.go` and works through the production binary.
2. F6 (`--tool-version-cmd` timeout left a grandchild) is fixed: `verifyRunPrepare(cmd)` now covers the tool-identity command.
3. Nothing outside the repair moved; no t1479-owned surface touched.
4. F1 (Ctrl-C debt) from the baseline is not in this delta and remains debt.

## Evidence

| Check | Command | Observed |
|---|---|---|
| Scope | `git diff --stat 710e0eed5 HEAD` | 4 files only: progress.md, internal/cli/verify.go (+3), verify_run.go (+1/-4), verify_run_test.go (+67). No t1479 paths. |
| Build | `go build -o /tmp/t1452-audit-bin ./cmd/moai` | exit 0 |
| help | `/tmp/t1452-audit-bin verify --help` | Verbs list contains `verify run`; `verify run --help` renders full usage |
| Miss then reuse (real binary, ignored-state git repo) | `verify run --project-root /tmp/t1452-play -- sh -c 'echo ran >> /tmp/t1452-count; echo out'` x3 | run 1: `verify run: miss (no snapshot recorded for the current tree)` + `out`; runs 2-3: `verify run: reuse key=04e4…:e25edfcac454f303 recorded_at=2026-10-03T21:15:18+09:00 duration_ms=5`; counter file 1 line |
| Exit passthrough | `verify run ... -- sh -c 'exit 3'` | exit=3 |
| Targeted tests | `go test -v -count=1 -run '^(TestVerifyRunRegisteredOnRoot\|TestVerifyRunThroughRootCommand\|TestVerifyRunToolVersionGrandchildKilled)$' ./internal/cli/` | `--- PASS` x3, `ok ... 7.556s`, exit 0 |
| VerifyRun suite | `go test -v -count=1 -run 'VerifyRun' ./internal/cli/` | exit 0, `ok ... 51.628s`; 17 top-level PASS lines (19 incl. subtests), 1 SKIP (`TestVerifyRunHelperProcess`, the re-exec helper), 0 FAIL, 0 `no tests to run` |
| verify pkg | `go test -count=1 ./internal/verify/` | `ok ... 3.114s` |
| vet / fmt / windows | `go vet ./internal/cli/`; `GOOS=windows go build ./internal/cli/`; `gofmt -l` on the 4 files | vet exit 0; windows build exit 0; gofmt empty |
| Mutation 1 (F6) | commented out `verifyRunPrepare(cmd)` in `verifyRunToolIdentity`, ran the F6 test | `--- FAIL: TestVerifyRunToolVersionGrandchildKilled` "the tool-version grandchild survived the timeout: its marker file exists"; reverted |
| Mutation 2 (F0) | restored init-append, removed direct AddCommand | `--- FAIL: TestVerifyRunRegisteredOnRoot`, `--- FAIL: TestVerifyRunThroughRootCommand`; reverted, `cmp` against saved originals identical, `git status --short` shows only the pre-existing `M .moai/reports/t1452/verdict.md` |
| sync-gate | `verify sync-gate --help` | prints the parent `verify` help (not a subcommand): still unregistered |

## Repair review (a05ec62de)

- No double AddCommand: the `init()` in verify_run.go is removed and `verifyExtraCommands` has no `newVerifyRunCmd` entry (grep of non-test sources). Mutation 2 confirms the root-tree tests detect the old shape.
- Process-group kill: `Setpgid: true` plus `cmd.Cancel = syscall.Kill(-pid, SIGKILL)` is correct on Unix; `WaitDelay` already set before the call. Windows split via build tags is a documented no-op (spec §D-5), builds clean.
- Tests go through `rootCmd` (the production tree), so the test gap that hid F0 is closed. They mutate global `rootCmd` args/out and restore in `t.Cleanup`; not parallel-safe, but none are `t.Parallel()`.
- The F6 test uses a fixed `time.Sleep(3s)` (flake-tolerant direction: only a false PASS is possible if the grandchild is slow to write, but mutation 1 shows it does go red).

## Findings

- F1 (carried, baseline) [debt] [optional] Ctrl-C handling debt from sync-audit.md is outside this delta; unchanged, not re-measured here.
- F7 [P3] [debt] [optional] internal/cli/verify.go:65 / verify_receipts.go:16 / audit_plan_cmd.go:46 / codex_review_receipt.go:177 — the same init-order hazard that caused F0 (`verifyExtraCommands` filled by other files' init() after verify.go's init runs) leaves `verify sync-gate` unregistered (confirmed above); `codex-review` and `audit-plan` use the same mechanism and were not individually probed. Pre-existing, out of scope per dispatch. Required fix (separate card): register those verbs directly or build the command lazily at Execute time.
- F8 [P3] [debt] [optional] verify_run_test.go F6 test relies on a fixed 3 s sleep; acceptable, noted only.

Blocking findings: none.

## Gaps

- No cross-model audit was run (audit_multi/codex_audit/glm_audit/claude_audit all deliberately skipped per the bounded-run instruction).
- The full HEAD SHA was resolved at the end of the audit with `git rev-parse HEAD` (28843a6d1868616e67605086540c012b37b10146); HEAD at audit start was `28843a6d1`.
- Windows process-group behaviour is not exercised (build only).
- F1 and baseline dimensions other than the repaired surfaces (security, coverage percentage) were not re-measured; carried from sync-audit.md.
- `verify codex-review` / `verify audit-plan` registration not probed individually.
- Build provenance: the binary was built from this tree's HEAD in this run (`/tmp/t1452-audit-bin`), so the judging build equals the tree.

## Residual-risk

`sync-gate` and sibling verbs stay unreachable from the real CLI until the init-order pattern is retired (F7); the verify-run reuse itself depends on callers ignoring `.moai/state` in git (true in this project; my first probe in an un-ignored scratch repo showed missed reuse for that reason, an artifact of the probe not the code).

## Iteration history

1. sync-audit.md: FAIL 62 (F0 P1, F1, F6).
2. This delta: F0 and F6 resolved with mutation proof; PASS-WITH-DEBT 90.
