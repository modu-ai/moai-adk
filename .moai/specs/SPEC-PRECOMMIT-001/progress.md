# SPEC-PRECOMMIT-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-07-05
plan_auditor_iter2: PASS-WITH-DEBT 0.79 → D1-D7 closed (all additive; core design unchanged)
tier: S
artifact_set: spec.md + plan.md + acceptance.md + progress.md
req_count: 19 REQ groups (REQ-PC-001 .. REQ-PC-019; REQ-PC-010/011 each split into two GEARS clauses 010a/010b, 011a/011b for GEARS purity — 21 clauses total)
ac_count: 15 (AC-PC-001 .. AC-PC-015)
config_reconciliation: config-agnostic install (REQ-PC-017 — pre-push *installer* is verifiably config-agnostic); deferred follow-up is a hook-run-time severity dial mirroring resolvePrePushAction, NOT an install gate. Verified this session: only pre_commit is dead (validation.go checkStringField only); pre_push is LIVE (resolvePrePushAction reads ActiveModeProfile().Hooks.PrePush at hook-run time)

## §E.2 Run-phase Evidence

Run-phase implementation landed in an isolated agent worktree
(`worktree-agent-ae28b58a655773009`, base HEAD `7cffb9717`). Files:

- NEW `internal/template/templates/.git_hooks/pre-commit` (fast-subset POSIX sh, 1953 bytes).
- NEW `internal/cli/hook_install_precommit.go` (`moaiPreCommitMarker`, `preCommitHookContent`,
  `PreCommitInstaller`, `NewPreCommitInstaller`, `InstallPreCommitHook`, `installPreCommitHookOptional`).
- EDIT `internal/cli/hook_install.go` (`fileHasMoaiMarker` → thin wrapper over new `fileHasMarker(path, marker)`;
  pre-push behaviour preserved byte-for-byte — `fileHasMoaiMarker` 100% covered by unchanged pre-push tests).
- EDIT `internal/cli/init.go` + `internal/cli/update.go` (one-line `installPreCommitHookOptional` wiring beside pre-push).
- NEW `internal/cli/hook_install_precommit_test.go` (17 tests; byte-identity + installer + hook-behaviour).

### AC PASS/FAIL Matrix (AC-PC-001 .. AC-PC-015)

| AC | Status | Verification command | Actual output |
|----|--------|----------------------|---------------|
| AC-PC-001 | PASS | `go test -run TestPreCommitInstall_FreshRepo ./internal/cli/` | `--- PASS: TestPreCommitInstall_FreshRepo (0.03s)` (mode 0755, content == constant, nil err) |
| AC-PC-002 | PASS | `go test -run 'TestPreCommitInstall_PreservesForeignHook|TestPreCommitInstall_OptionalPreservedNote' ./internal/cli/` | both `--- PASS`; foreign hook byte-unchanged, `ErrUserHookExists`, "preserved" note |
| AC-PC-003 | PASS | `go test -run TestPreCommitInstall_OverwritesMoaiHook ./internal/cli/` | `--- PASS`; marker hook overwritten with constant, nil err |
| AC-PC-004 | PASS | `go test -run TestPreCommitTemplateMatchesConstant ./internal/cli/` | `--- PASS`; template ≡ `preCommitHookContent` byte-identical |
| AC-PC-005 | PASS | `go test -run TestPreCommitHook_GofmtBlocks ./internal/cli/` | `--- PASS`; staged un-gofmt'd .go → exit 1 + gofmt hint |
| AC-PC-006 | PASS | `go test -run TestPreCommitHook_SkipBypass ./internal/cli/` | `--- PASS`; `SKIP_MOAI_PRECOMMIT=1` → exit 0 + bypass notice |
| AC-PC-007 | PASS | `go build ./... && GOOS=windows GOARCH=amd64 go build ./...` | both exit 0 (no OS-specific primitives) |
| AC-PC-008 | PASS | `go test -run TestPreCommitHook_NoStagedGo ./internal/cli/` | `--- PASS`; only .txt staged → exit 0, no FAILED |
| AC-PC-009 | PASS | `go test -run TestPreCommitHook_GoVetBlocks ./internal/cli/` | `--- PASS`; gofmt-clean Printf-mismatch → exit 1 + go vet hint |
| AC-PC-010 | PASS | `go test -run 'TestPreCommitInstall_SkipFlag|TestPreCommitInstall_OptionalSkipSilent' ./internal/cli/` | both `--- PASS`; skip=true → no file written, no output |
| AC-PC-011 | PASS | `go test -run TestPreCommitContent_TwoTierBoundary ./internal/cli/` | `--- PASS`; constant has no `make ci-local` / `golangci-lint` / `go test` token |
| AC-PC-012 | PASS | `go test -run TestPreCommitInstall_Assertion ./internal/cli/` | `--- PASS`; `#!/bin/sh` shebang + 4 tokens (marker, `gofmt -l`, `go vet`, `SKIP_MOAI_PRECOMMIT`) |
| AC-PC-013 | PASS | `go test -run TestPreCommitHook_ToolchainAbsent ./internal/cli/` | `--- PASS`; shim PATH without go/gofmt + staged .go → exit 0 (checks skipped) |
| AC-PC-014 | PASS | `go test -run TestPreCommitInstall_NonFatalFailure ./internal/cli/` | `--- PASS`; unwritable `.git/hooks` → non-fatal "Warning:" + no abort; err ≠ ErrUserHookExists |
| AC-PC-015 | PASS | `go test -run TestPreCommitInstall_ConfigAgnostic ./internal/cli/` | `--- PASS`; skip/warn/enforce → identical 0755 hook == constant |

Full targeted suite: `go test ./internal/cli/ -run 'PreCommit' -count=1` → `ok  github.com/modu-ai/moai-adk/internal/cli  0.918s` (17 PASS, 0 FAIL).

### Pre-existing baseline breakage (NOT introduced by this SPEC)

`go test ./...` reports one FAIL package: `internal/cli`, from two pre-existing panicking tests
(`TestRunHookEvent_ReadInputError` @ `coverage_test.go:77`, `TestRunAgentHook_ReadInputError`
@ `misc_coverage_test.go:375`) — nil-pointer derefs in parallel-session-modified files this SPEC never touched.
Proven pre-existing: with my 6 changes stashed/set-aside, both panic identically on the pristine `7cffb9717` tree.
With both excluded, `go test ./internal/cli/ -skip 'TestRunHookEvent_ReadInputError|TestRunAgentHook_ReadInputError'`
→ `ok ... 9.507s`.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-07-05
run_commit_sha: "c72da7395"
run_status: PASS
ac_pass_count: 15
ac_fail_count: 0
preserve_list_post_run_count: pre-push installer/constant/template/test + ErrUserHookExists sentinel + all config layers — UNCHANGED (TestPrePushTemplateMatchesConstant + pre-push installer tests still green)
l44_pre_commit_fetch: N/A — isolated agent worktree; orchestrator owns origin/main fetch + integration (no push this agent)
l44_post_push_fetch: N/A — this agent does NOT push (per task instruction; commit left local on worktree branch)
new_warnings_or_lints_introduced: 0 (golangci-lint run --timeout=5m ./internal/cli/... → "0 issues."; gofmt -l on 5 changed files → clean; go vet ./internal/cli/ → exit 0)
cross_platform_build:
  native: exit 0 (go build ./...)
  windows: exit 0 (GOOS=windows GOARCH=amd64 go build ./...)
coverage_new_functions: NewPreCommitInstaller 100.0% · InstallPreCommitHook 86.7% · installPreCommitHookOptional 100.0% · fileHasMarker 91.7% · fileHasMoaiMarker 100.0% (all ≥ 85%)
total_run_phase_files: 6 (1 template NEW + 1 installer NEW + 1 test NEW + hook_install.go/init.go/update.go EDIT)
m1_to_mN_commit_strategy: single combined commit (Tier S, mechanical mirror); M1 draft→in-progress transition rides this first run-phase commit
template_neutrality: PASS — pre-commit template has no SPEC IDs / REQ tokens / dates / SHAs / macOS paths / CLAUDE.local refs (generic POSIX sh)
make_build: PASS — embedded FS regenerated (REQ-PC-016); catalog.yaml unchanged (template not part of catalog hash set)

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_
