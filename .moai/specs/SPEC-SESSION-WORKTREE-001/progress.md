# Progress — SPEC-SESSION-WORKTREE-001

> **Tier L** lifecycle progress. §E section skeleton emitted at plan-phase; §E.2-§E.4 populated only by the respective phase-owning agents.

## §E.1 Plan-phase Audit-Ready Signal

_Populated at plan-phase close._

- plan_status: _pending iter-2 audit (v0.2.1 iter-2 audit-fix; iter-1 verdict FAIL 0.69); v0.2.2 run-phase inline-fix applied (D-NEW-1 non-transition correction of M8 trigger points — `moai worktree list` / `moai session start` were non-existent; corrected to `moai session register` / `moai session list`)_
- plan_complete_at: _pending_
- artifacts: spec.md v0.2.2, plan.md v0.2.2 (§A / §F M8 / §I Q3 corrected), acceptance.md v0.2.2 (AC-SW-022 / EC-13 corrected), design.md v0.2.2 (§C.3 / §D corrected), research.md v0.2.1 (unchanged — carries no trigger-point claims), progress.md (this file)
- tier: L (retained at v0.2.2 — REQ/AC budget 24/24, file surface ≥10; v0.2.0 escalation M→L stands)
- version_trace: v0.1.0 (initial Tier M) → v0.2.0 (tier escalated M→L, 3 additions) → v0.2.1 (iter-2 audit-fix: D1 gitconfig source, D2 Q1-Q4 resolved, D3 defaultBranchDetectorFunc documented, D4 on-touch trigger, D5 path-distinguishing notices, D6 REQ-SW-014 Event-driven relabel, D7 REQ-SW-021 SHALL) → v0.2.2 (run-phase inline-fix D-NEW-1: M8 trigger-point command-name correction — v0.2.1 `moai worktree list` retired + `moai session start` non-existent → verified-existing `moai session register` + `moai session list`; REQ/AC count unchanged 24/24, status stays in-progress, NOT an amendment)
- lint: _pending `moai spec lint`_
- self_check: SPEC ID regex PASS observed; frontmatter 12-field check pending; GEARS notation verified; Out-of-Scope rule satisfied (11 `### Out of Scope —` H3 sub-headings — 9 carried from v0.2.0 + 2 added at v0.2.1: Profile-vs-profile git identity + Hook isolation); Tier L artifact set complete (spec + plan + acceptance + design + research + progress); all v0.2.0 `[NEEDS CLARIFICATION]` markers resolved (no more blocker rounds); v0.2.2 inline-fix consistency: trigger-point references corrected across all 4 carrying artifacts (spec.md HISTORY + plan.md §A/§F M8/§I Q3 + acceptance.md AC-SW-022/EC-13 + design.md §C.3/§D), research.md carries no trigger-point claims so no edit needed, REQ/AC count preserved at 24/24.

## §E.2 Run-phase Evidence

### M2 — `moai init` auto-entry (COMPLETE)

**Deliverables:**
- `internal/cli/session_worktree.go` (NEW) — `enterSessionWorktree` wrapper + `materializeSessionWorktree` + function-variable seams (`sessionWorktreeGitWorktreeAdd`, `sessionWorktreeInGitWorktree`, `sessionWorktreeResolveSessionShort`, `sessionWorktreeGitCommonDir`, `sessionWorktreeGitConfigSet`) + `loadSessionWorktreeConfig`.
- `internal/cli/session_worktree_test.go` (NEW) — 9 tests covering default-off, env-forced-on, already-in-worktree skip, fail-back (2 cases), defaultBranch application, branch-name shape, random fallback, falsification.
- `internal/cli/init.go` (EDIT) — `runInit` calls `enterSessionWorktree(..., "init", ...)` at the very top (after printer setup, before any mutation); on success `os.Chdir(wtPath)` re-resolves cwd into the worktree (BI-2 / R1).

**BI-1 resolution (worktree.Add extraction vs shell-out):** SHELL-OUT. Evidence: `grep -rn "func Add\|func.*Add(" internal/cli/worktree/ internal/cli/launcher.go internal/workflow/` returns only mock-interface `Add(path, branch)` methods in `_test.go` files (`subcommands_test.go:22`, `done_test.go:128`, `worktree_orchestrator_test.go:25`) — NO exported reusable helper exists. M2 shells out to `git worktree add -b <branch> <dest>` directly (`gitWorktreeAddReal`). Extraction would be premature; M3/M6 will reuse the same wrapper.

**Q2 resolution ([WT] brackets):** BRACKETS REJECTED → `WT-` prefix fallback. Evidence: `git check-ref-format --branch '[WT]-x'` exits non-zero with `fatal: '[WT]-x' is not a valid branch name`. `SessionWorktreeBranchPrefix = "WT-"` (bracket-free). REQ-SW-006's literal `[WT]` intent is preserved via the `WT-` token; the bracket rejection is the documented EC-3 fallback.

**Q1 resolution:** Same as BI-1 (shell-out decision).

**M7 call-site hook:** `materializeSessionWorktree` applies the M2 init-specific config (`init.defaultBranch=main`, REQ-SW-020) directly and carries the `// M7 (REQ-SW-018/019/021): ApplyGitConfig call site` comment marking where safe.directory / global-gitconfig identity / opt-in options will land. No non-existent function is stubbed (build passes).

**AC binary PASS/FAIL matrix (M2-relevant):**

| AC | Status | Verification | Actual Output |
|---|---|---|---|
| AC-SW-002 (default-off init byte-identical) | PASS | `TestEnterSessionWorktree_DefaultOffReturnsEmpty` + `TestEnterSessionWorktree_DefaultOffWithConfigFalse` | `enterSessionWorktree(nil,...) → ""`; no git invocation; no notice (out.Len()==0) |
| AC-SW-005 (non-git fail-back + notice) | PASS | `TestEnterSessionWorktree_MaterializeFailFallsBack` + `TestEnterSessionWorktree_FailBackFalsification` | `→ ""` + notice contains "materialization failed" + the failure reason |
| AC-SW-007 (branch name shape) | PASS | `TestSessionWorktreeBranchName_Shape` | `WT-abcdef12-init` (Q2 bracket fallback applied) |
| AC-SW-012 (already-in-worktree skip) | PASS | `TestEnterSessionWorktree_AlreadyInWorktreeSkips` | `→ ""`; notice contains "already inside a git worktree"; git add NOT invoked |
| AC-SW-020 (init.defaultBranch=main) | PASS | `TestEnterSessionWorktree_SuccessAppliesDefaultBranch` | `git config init.defaultBranch main` called on the worktree path |

**E8 RED → GREEN proof (falsification round-trip):** the fail-back notice string was mutated from "materialization failed" to "REMOVED_FAILBACK" (simulating removal of the fail-back). Both fail-back tests went RED:
```
--- FAIL: TestEnterSessionWorktree_MaterializeFailFallsBack
    expected fail-back notice, got "...REMOVED_FAILBACK..."
--- FAIL: TestEnterSessionWorktree_FailBackFalsification
    expected fail-back notice for non-git dir, got "...REMOVED_FAILBACK..."
```
After restoring the file, both tests returned GREEN (`ok ... 1.224s`). This proves the fail-back code path is load-bearing.

**Coverage (new init-wrapper decision code):**
- `enterSessionWorktree`: 100.0%
- `sessionWorktreeBranchName`: 100.0%
- `materializeSessionWorktree`: 90.9%
- `resolveSessionShortReal`: 45.5% (random-fallback branch hit; session-id-available branch is exercised in production via the side-channel file)
- `*Real` exec shells (gitWorktreeAddReal, inGitWorktreeReal, gitCommonDirReal, gitConfigSetReal, loadSessionWorktreeConfig): 0% in unit tests — these are thin `exec.Command` wrappers exercised in production; the load-bearing decision logic is covered via the seams. Exceeds the 85% target on the wrapper decision code.

**Cross-platform build:**
- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0

**Lint:** `golangci-lint run --timeout=3m ./internal/cli/` → 0 issues.

**Vet:** `go vet ./internal/cli/... ./internal/config/...` → clean.

**Subagent boundary (C-HRA-008):** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/session_worktree.go internal/cli/session_worktree_test.go` → 0 matches (no new violations).

**Scope discipline (B10):** touched only init entry (`internal/cli/init.go` runInit top), the new wrapper (`internal/cli/session_worktree.go`), and tests. Did NOT modify web.go (M3), profile.go (M6), or the M7 git-config helper body.

### M3 — `moai web` auto-entry + advisory suppression (COMPLETE)

**Deliverables:**
- `internal/cli/web.go` (EDIT) — `runWeb` now calls `enterSessionWorktree(..., "web", ...)` at the very top (before `findProjectRootFn` / `ensurePortFree` / `emitWorktreeAdvisory`); on success `os.Chdir(wtPath)` re-resolves cwd into the worktree (BI-2 / R1 mitigation); `emitWorktreeAdvisory` is gated on `!wtMaterialized` (REQ-SW-013). `web.Run` is invoked via a new `webRunFn` test seam (mirrors `findProjectRootFn` / `findPortHolder`).
- `internal/cli/web_session_worktree_test.go` (NEW) — 3 tests: suppression-on-materialization (REQ-SW-013 positive), advisory-fires-when-off (REQ-SW-001 negative control), advisory-fires-on-failback (REQ-SW-004 negative control).

**Residual from M2 (resolved):** `loadSessionWorktreeConfig(cmd)` now receives runWeb's `*cobra.Command` (the future-proof `cmd` param is used). Evidence: `internal/cli/web.go:87` `enterSessionWorktree(loadSessionWorktreeConfig(cmd), "web", ...)`.

**AC binary PASS/FAIL matrix (M3-relevant):**

| AC | Status | Verification | Actual Output |
|---|---|---|---|
| AC-SW-001 (default-off web byte-identical) | PASS | `TestRunWeb_AutoEntryOff_AdvisoryFires` | feature unset → advisory fires (stdout contains `main-checkout-branch-guard.md`) exactly as today |
| AC-SW-003 (config ON materializes worktree for web) | PASS | `TestRunWeb_AutoEntryOn_SuppressesAdvisory` (materialize seams succeed, chdir lands, `wtMaterialized=true`) | "MoAI Web Console starting" line printed; advisory suppressed |
| AC-SW-013 (advisory suppressed when materialized; negative controls fire) | PASS | suppression test + 2 negative controls | ON+materialized → no advisory; OFF → advisory; ON+failback → advisory |
| AC-SW-015 (port collision preserved) | PASS | `ensurePortFree` logic untouched; REQ-SW-015 noted in code comment at web.go:108-110 | `ensurePortFree(cmd.ErrOrStderr(), webPort, !webNoReuse)` runs identically regardless of worktree entry |

**E8 RED → GREEN proof (falsification round-trip):** RED captured first against the unchanged `runWeb` — all three tests failed because runWeb still called the real `web.Run` (port bind error) and no suppression/gating existed. After wiring GREEN (auto-entry + `webRunFn` seam + `!wtMaterialized` gate), a falsification round-trip mutated the gate to `if true || !wtMaterialized` (force advisory always). The suppression test went RED:
```
--- FAIL: TestRunWeb_AutoEntryOn_SuppressesAdvisory
    REQ-SW-013: advisory should be SUPPRESSED on materialization, but stdout contains it:
        ...Tip: this checkout is shared...main-checkout-branch-guard.md.
```
After restoring `if !wtMaterialized`, all three tests returned GREEN (`ok ... 3.893s` under `-race`). This proves the `wtMaterialized` gate is load-bearing.

**E9 negative-control proof:** the two negative-control tests (`TestRunWeb_AutoEntryOff_AdvisoryFires`, `TestRunWeb_AutoEntryOn_FailBack_AdvisoryFires`) both assert the advisory marker IS present — confirming suppression is conditional, not blanket. Both pass.

**Coverage:** `go test -cover ./internal/cli/` → 76.2% of statements (package-level baseline; M3's new runWeb branches are fully covered by the 3 new tests + the existing web flag/help tests).

**Cross-platform build:**
- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0

**Lint:** `golangci-lint run --timeout=3m ./internal/cli/...` → 0 issues.

**Subagent boundary (C-HRA-008 / E4):** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/web.go internal/cli/web_session_worktree_test.go | grep -v "_test.go" | grep -v "// "` → 0 NEW matches.

**Scope discipline (B8/B10):** touched only `internal/cli/web.go` (runWeb + new `webRunFn` seam + `os` import) and `internal/cli/web_session_worktree_test.go` (NEW). Did NOT modify `session_worktree.go` core logic (reused as-is), `ensurePortFree`/port handling (REQ-SW-015), or `emitWorktreeAdvisory` (gated at the call site, not modified). spec.md/plan.md/acceptance.md/design.md/research.md body unchanged.

### M4 — Branch naming verify + session-exit disposal (COMPLETE)

**Branch naming (AC-SW-007 / REQ-SW-006/007):** M2's `sessionWorktreeBranchName` (produces `WT-<session-short>-<subcommand>`) and `resolveSessionShortReal` verified correct — NOT changed. The session-id-available branch of `resolveSessionShortReal` (first-8 chars of the registry session UUID via the `.moai/state/current-session-id.txt` side-channel) was UNTESTED at M2 (coverage 45.5% — only the random-fallback branch was hit). M4 ADDS two tests (`TestResolveSessionShortReal_SideChannelAvailable`, `TestResolveSessionShortReal_ShortSessionID`) plus `TestResolveSessionShortReal_NoSideChannelFallsBack` covering the EC-4 fallback; `resolveSessionShortReal` coverage rose 45.5% → 90.9%.

**Session-exit disposal (AC-SW-008/009/010 / REQ-SW-008/009/010):** NEW `cleanupSessionWorktree(cfg, wtPath, cleanExit, out)` honors the three cases + the dirty guard:

| Case | Trigger | Behavior |
|---|---|---|
| Default-manual (REQ-SW-008) | `auto_cleanup == false` (distributed default) | worktree PERSISTS — no-op, no notice |
| Opt-in clean-exit (REQ-SW-009) | `auto_cleanup == true` AND `cleanExit == true` AND clean worktree | `git worktree remove` + stderr notice `removed by session-exit cleanup: <path>` |
| Non-clean-exit preserve (REQ-SW-009) | `auto_cleanup == true` AND `cleanExit == false` | PRESERVE for post-mortem + notice |
| Dirty guard (REQ-SW-010) | `git status --porcelain` non-empty | SKIP removal + notice naming the worktree path (uncommitted changes NEVER deleted); guard is fail-open on status-check error |

The `worktreeIsDirty(wtPath)` helper is the SHARED guard factored for M8 PR-merge reuse (REQ-SW-010 + REQ-SW-024 — one helper, two call sites).

**Notice distinguishability (REQ-SW-009 vs REQ-SW-022 / EC-13):** the session-exit notice prefix constant `SessionExitCleanupNoticePrefix = "removed by session-exit cleanup:"` does NOT contain the substring `PR-merge`, so it is unambiguously distinguishable from the M8 PR-merge notice (`removed by PR-merge cleanup:`). The path in the notice carries the `WT-` branch prefix from M2/Q2 (the worktree dir basename IS the branch name).

**Wiring (init.go + web.go):** `runInit` and `runWeb` converted to named returns `(err error)`; each captures the worktree path at entry and defers `cleanupSessionWorktree(swCfg, wtPath, err == nil, stderr)` — `cleanExit` is derived from the named error return (`err == nil` means exit 0). No new config flag; reuses `workflow.worktree.auto_cleanup` (§22.8, default `false`).

**RED → GREEN proof (E8):** tests written first against a no-op stub `cleanupSessionWorktree`; verbatim RED output (4 FAILs) captured before the real logic landed:
```
--- FAIL: TestCleanupSessionWorktree_CleanExitRemoves — clean-exit: expected remove("/repo/.claude/worktrees/WT-abcdef12-web"), got ""
--- FAIL: TestCleanupSessionWorktree_NonCleanExitPreserves — non-clean-exit: expected preserve notice, got ""
--- FAIL: TestCleanupSessionWorktree_DirtyPreserves — dirty: notice must name the worktree path "...", got ""
--- FAIL: TestCleanupSessionWorktree_DirtyCheckErrorPreserves — dirty-check-error: notice must name worktree path, got ""
```
GREEN: all 19 session-worktree tests (M2 + M4) PASS.

**Falsification round-trip (a) — dirty guard (REQ-SW-010) load-bearing:** bypassed the dirty-guard block (`dirty, derr := false, error(nil)`) and re-ran `TestCleanupSessionWorktree_DirtyPreserves` → FAIL:
```
--- FAIL: TestCleanupSessionWorktree_DirtyPreserves — dirty: remove MUST NOT run (uncommitted changes preserved)
```
Restoring the guard → PASS. The guard prevents a dirty worktree from being wrongly removed.

**Falsification round-trip (b) — clean-exit-only (REQ-SW-009) load-bearing:** bypassed the `if !cleanExit` preserve branch (`if !cleanExit && false`) and re-ran `TestCleanupSessionWorktree_NonCleanExitPreserves` → FAIL:
```
--- FAIL: TestCleanupSessionWorktree_NonCleanExitPreserves — non-clean-exit: remove MUST NOT run (preserve for post-mortem)
```
Restoring the check → PASS. The exit-code gate prevents a worktree from being removed after a non-zero exit.

**E1 — AC matrix (M4-relevant):**

| AC | Status | Command | Output |
|---|---|---|---|
| AC-SW-007 | PASS (M2, verified) | `go test -run TestSessionWorktreeBranchName ./internal/cli/` | `--- PASS: TestSessionWorktreeBranchName_Shape` + M4-added `TestResolveSessionShortReal_SideChannelAvailable` |
| AC-SW-008 | PASS | `go test -run TestCleanupSessionWorktree_DefaultManualPersists ./internal/cli/` | `--- PASS` — auto_cleanup=false → remove NOT called, no removal notice |
| AC-SW-009 | PASS | `go test -run 'TestCleanupSessionWorktree_CleanExitRemoves\|TestCleanupSessionWorktree_NonCleanExitPreserves\|TestCleanupSessionWorktree_NoticeDistinguishableFromPRMerge' ./internal/cli/` | `--- PASS` — clean-exit removes + notice; non-clean preserves; notice prefix lacks "PR-merge" |
| AC-SW-010 | PASS | `go test -run 'TestCleanupSessionWorktree_DirtyPreserves\|TestCleanupSessionWorktree_DirtyCheckErrorPreserves' ./internal/cli/` | `--- PASS` — dirty porcelain → skip removal + path-naming notice; status-check error → fail-open preserve |

**E2 — Cross-platform build:**
```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

**E3 — Coverage (M4 + M2 session-worktree functions):**
```
$ go test -run 'TestCleanupSessionWorktree|TestResolveSessionShortReal|TestSessionWorktreeBranchName|TestEnterSessionWorktree' -coverprofile=... ./internal/cli/
go tool cover -func=... | grep -E 'cleanup|dirty|resolve|branchName|enter':
  enterSessionWorktree      100.0%
  sessionWorktreeBranchName 100.0%
  resolveSessionShortReal    90.9%  (was 45.5% at M2 — session-id-available branch now covered)
  cleanupSessionWorktree     88.9%  (NEW; uncovered = dirty-check-error notice + remove-failure notice — exceptional paths)
  worktreeIsDirty           100.0%  (NEW; shared with M8 PR-merge path REQ-SW-024)
  gitWorktreeRemoveReal       0.0%  (real git subprocess — exercised via the function-variable seam, mirroring M2's gitWorktreeAddReal=0%)
  gitStatusPorcelainReal      0.0%  (real git subprocess — exercised via the seam)
```

**E4 — Subagent boundary (C-HRA-008):** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/session_worktree.go internal/cli/init.go internal/cli/web.go | grep -v '_test.go' | grep -v '// '` → 0 matches.

**E5 — Lint:** `golangci-lint run --timeout=3m ./internal/cli/...` → 0 issues (NEW vs baseline: 0).

**Full suite:** `go test -count=1 ./internal/cli/` → `ok ... 161.745s` (all existing tests still pass).

**Scope discipline (B8/B10):** touched `internal/cli/session_worktree.go` (NEW `cleanupSessionWorktree` + `worktreeIsDirty` + 2 seams + real impls + notice constant), `internal/cli/session_worktree_test.go` (extended `swSeams` + 10 NEW tests), `internal/cli/init.go` (runInit → named return + defer cleanup), `internal/cli/web.go` (runWeb → named return + defer cleanup). Did NOT implement M8's PR-merge path (M4 owns only session-exit; M8 adds on-touch PR-merge). Did NOT add a new config flag (reuses `workflow.worktree.auto_cleanup`). spec.md/plan.md/acceptance.md/design.md/research.md body unchanged.

---

### M5 — coexistence + anti-regression + coverage (test-focused milestone)

**M5 deliverables — all test-only (production logic unchanged):** 5 NEW tests in `internal/cli/session_worktree_test.go` covering AC-SW-011/012/014 + anti-regression + E8 falsification round-trip. No production code changes (test-focused milestone; no defect surfaced).

**E1 — AC matrix (M5-relevant):**

| AC | Status | Command | Output |
|----|--------|---------|--------|
| AC-SW-011 (coexistence no-nest, web path) | PASS | `go test -run TestEnterSessionWorktree_WebCoexistenceNoNested ./internal/cli/` | `--- PASS: TestEnterSessionWorktree_WebCoexistenceNoNested (0.00s)` — inWt=true ⇒ add NOT called, notice names `"web"`, returns `""`. |
| AC-SW-012 (already-in-worktree skip, web path) | PASS | `go test -run TestEnterSessionWorktree_AlreadyInWorktreeWebNoticeContent ./internal/cli/` | `--- PASS: ...` — materialize seams (commonDir/add/configSet) all dormant; notice contains `"already inside a git worktree"`. |
| AC-SW-014 (parallel isolation deterministic invariant) | PASS | `go test -run TestSessionWorktreeBranchName_ParallelDistinctSessions ./internal/cli/` | `--- PASS: ...` — distinct session-shorts (aabbcc11 vs aabbcc22) ⇒ distinct branch names `WT-aabbcc11-web` vs `WT-aabbcc22-web` ⇒ distinct landing paths. |
| Anti-regression (OFF byte-identical, init + web) | PASS | `go test -run TestEnterSessionWorktree_OffByteIdentical_InitAndWeb ./internal/cli/` | `--- PASS: ...` — feature OFF ⇒ `""` + zero output + add NOT called, for BOTH `init` and `web`. |

**E8 — Falsification round-trip (REQ-SW-012, already-in-worktree skip):** `go test -run TestEnterSessionWorktree_AlreadyInWorktreeSkip_Falsification ./internal/cli/` → `--- PASS: ...`. Same materialize-capable setup run twice; only difference is `inWt`: true ⇒ 0 add calls (skip), false ⇒ 1 add call (proceed). The discrimination is the `sessionWorktreeInGitWorktree()` guard — removing it would make the `inWt=true` case also call add and FAIL this test. (M2's `TestEnterSessionWorktree_FailBackFalsification` covers the fail-back guard; M4's `DirtyPreserves` / `NonCleanExitPreserves` / `DirtyCheckErrorPreserves` cover the dirty + clean-exit-only guards — all still PASS, not duplicated.)

**E9 — Coverage-gap resolution (resolveSessionShortReal):** CLOSED. The orchestrator-flagged 45.5% measurement was STALE (taken before M4's `TestResolveSessionShortReal_SideChannelAvailable`). On current HEAD (`1b806e6ac`, M4 complete) the function measures **90.9%** — verified via TWO independent cover profiles:
- Focused run (session-worktree tests only): `go test -run 'SessionWorktree|CleanupSessionWorktree|ResolveSessionShortReal' -coverprofile=/tmp/m5_base.out ./internal/cli/` → `resolveSessionShortReal 90.9%`.
- Full-package run: `go test -count=1 -coverprofile=/tmp/m5_fullpkg.out ./internal/cli/` → `resolveSessionShortReal 90.9%`.

The side-channel mechanism staged by the M4 test IS the one the production code reads: `resolveCurrentSessionID` (session.go:214) reads `<$CLAUDE_PROJECT_DIR>/.moai/state/current-session-id.txt` (constant `session.CurrentSideChannelFile = ".moai/state/current-session-id.txt"`), and `TestResolveSessionShortReal_SideChannelAvailable` stages exactly that file under a `t.TempDir()` pointed at by `t.Setenv("CLAUDE_PROJECT_DIR", tmp)`. The test PASSES (`--- PASS: TestResolveSessionShortReal_SideChannelAvailable (0.00s)`) and returns the first-8 chars `"abcdef12"` of the staged UUID, proving the session-id-available branch executes. The remaining 9.1% gap is the `rand.Read` error branch (line 217-221: `crypto/rand` returning an error — exceptional, cannot be staged without injecting a failing reader; the seam `sessionWorktreeResolveSessionShort` wraps the whole function, not the rand call, so it cannot isolate that branch).

**E2 — Cross-platform build:** `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.

**E3 — Coverage (BEFORE → AFTER on the four key functions, M5 did not change production code so AFTER == M4 baseline):**
```
                              BEFORE(M4)  AFTER(M5)
enterSessionWorktree           100.0%     100.0%
resolveSessionShortReal         90.9%      90.9%   (gap already closed at M4; orchestrator 45.5% was stale)
cleanupSessionWorktree          88.9%      88.9%
worktreeIsDirty                100.0%     100.0%
```
Command: `go test -run 'SessionWorktree|CleanupSessionWorktree|ResolveSessionShortReal|EnterSessionWorktree' -count=1 -coverprofile=/tmp/m5_after.out ./internal/cli/` then `go tool cover -func=/tmp/m5_after.out | grep -iE "resolveSessionShortReal|cleanupSessionWorktree|enterSessionWorktree|worktreeIsDirty"`.

**E4 — Subagent boundary grep:** `grep -n 'AskUserQuestion\|mcp__askuser' internal/cli/session_worktree_test.go | grep -v '^[0-9]*:[ \t]*//'` → 0 matches (test-only milestone; no production code touched, so the existing M1-M4 production boundary is unchanged and already 0).

**E5 — Lint:** `golangci-lint run --timeout=3m ./internal/cli/` → `0 issues.` (NEW vs baseline: 0). `go vet ./internal/cli/` → exit 0.

**E6 — Branch HEAD + Push:** M5 commit + push to `feat/SPEC-SESSION-WORKTREE-001` (see commit SHAs below).

**Full suite:** `go test -count=1 -timeout=8m ./internal/cli/` → `ok ... 188.964s` (all existing tests still pass; no regressions from the 5 NEW M5 tests).

**Scope discipline (B8/B10 — test-only):** touched ONLY `internal/cli/session_worktree_test.go` (5 NEW tests, no edits to existing tests or production code) + this progress.md §E.2 evidence append. spec.md/plan.md/acceptance.md/design.md/research.md body unchanged. `internal/cli/session_worktree.go` production logic unchanged (test-focused milestone — no defect surfaced by the M5 tests, so no blocker report). spec.md status stays `in-progress`.

**Residual risk — AC-SW-014 timing non-determinism:** the "within the same second" concurrency aspect of AC-SW-014 cannot be reliably reproduced in a unit test (inherent to the OS scheduler, not a moai bug). The deterministic invariant — distinct session-shorts ⇒ distinct branch names ⇒ distinct worktree paths ⇒ distinct state trees + settings.local.json surfaces — is proven by `TestSessionWorktreeBranchName_ParallelDistinctSessions`. A collision under real concurrency would require two sessions to resolve the SAME 8-char session-short (the session-id-available branch) or collide on the 6-byte random fallback (≈ 2^48 space) — neither is tested for true concurrency here.

### M6 — moai profile auto-entry (project-scoped, REQ-SW-016/017, EC-7)

**M6 deliverables:**
1. **Profile entry function (BI-5):** `runProfileSetup` in `internal/cli/profile_setup.go` — the single handler invoked by BOTH `moai profile setup [name]` (subcommand) and `moai profile --setup/-s` (flag on root, delegated via `runProfileCmd`). Auto-entry wired HERE, before any shared-state mutation (`profile.ReadPreferences`/`WritePreferences`/`persistProjectConfig`), mirroring M2 init / M3 web: `loadSessionWorktreeConfig(cmd)` → `enterSessionWorktree(swCfg, "profile", cmd.ErrOrStderr())` → chdir + `emitProfileScopeNotice` → deferred `cleanupSessionWorktree`. Named return `(err error)` so cleanup derives `cleanExit`.
2. **REQ-SW-016/017 + branch naming:** `enterSessionWorktree(..., "profile", ...)` reuses the shared wrapper (no forked impl). Branch name `WT-<session-short>-profile` (REQ-SW-007). Worktree scopes to the PROJECT (working tree) — materialize path derives project root from `git rev-parse --git-common-dir` parent, exactly like init/web; NOT `~/.moai/claude-profiles/`.
3. **EXPLICIT NON-ACTION (load-bearing):** M6 does NOT isolate the profile dir. `~/.moai/claude-profiles/` remains GLOBAL; the launch-ledger race on `launch.yaml` is documented out of scope (§3) and deferred to a follow-up SPEC. `runProfileDelete` (global-dir mutator) is NOT wired to auto-entry.
4. **Honest stderr notice (REQ-SW-017, AC-SW-017c):** `emitProfileScopeNotice(out, wtPath)` emits a notice naming BOTH the project worktree path AND `profile.GetBaseDir()`, stating EXPLICITLY the profile dir is NOT isolated. The "NOT isolated" clause is load-bearing (falsification round-trip in E8).
5. **EC-7 read-only gate (subverb enumeration):**

   | Subverb | Handler | Mutates PROJECT tree? | Auto-entry? |
   |---|---|---|---|
   | `moai profile` (bare/help) | `runProfileCmd`→`cmd.Help()` | No | NO |
   | `moai profile --setup`/`-s` + `moai profile setup [name]` | `runProfileSetup` | Yes (`persistProjectConfig` writes `.moai/config/sections/*.yaml`) | **YES** |
   | `moai profile list`/`ls` | `runProfileList` | No (read-only) | NO |
   | `moai profile current` | `runProfileCurrent` | No (read-only) | NO |
   | `moai profile delete`/`rm` | `runProfileDelete` | No (mutates global `~/.moai/claude-profiles/`, NOT project) | NO |

   Wiring auto-entry into `runProfileSetup` (the only project-mutating subverb) naturally excludes all read-only and global-only subverbs.

6. **Unused-cmd debt (E10): CARRIED.** `loadSessionWorktreeConfig(cmd)` accepts `cmd` but profile setup resolves the project root via cwd (`runProfileSetup` uses `os.Getwd()` + `filepath.Join(cwd, ".moai")`); the profile command defines NO `--root`/`--project` flag that resolves the root more precisely than cwd. The deliverable explicitly permits carrying the debt forward when the profile shape offers no such flag. Lint baseline is clean (`golangci-lint run ./internal/cli/...` → 0 issues), so the `cmd` param remains a documented future-proof, not an active defect.

**E1 — AC matrix (M6-relevant):**

| AC | Status | Command | Output |
|----|--------|---------|--------|
| AC-SW-016 (profile auto-enters worktree) | PASS | `go test -run TestEnterSessionWorktree_ProfileSubcommand ./internal/cli/` | `ok ... 0.659s` — asserts `enterSessionWorktree(cfg,"profile",out)` materializes branch `WT-abcdef12-profile` + names subcommand `"profile"` |
| AC-SW-017(a) worktree under PROJECT | PASS (structural) | `enterSessionWorktree` reuses shared `materializeSessionWorktree` (project root = parent of `git rev-parse --git-common-dir`); NOT under `~/.moai/claude-profiles/` | (proven by M2 `TestEnterSessionWorktree_EnvForcesOn` path-prefix assert + reuse; no forked path logic) |
| AC-SW-017(b) launch.yaml at GLOBAL path | PASS (non-action) | `grep -rn "BaseDirOverride\|GetBaseDir\|launchLedgerFile" internal/cli/profile_setup.go` | launch.yaml write site untouched — `RecordLastUsedProfile` keeps using `GetBaseDir()` (global); M6 does NOT redirect it |
| AC-SW-017(c) stderr notice states NOT isolated | PASS | `go test -run TestEmitProfileScopeNotice_LoadBearing ./internal/cli/` | `ok ... 0.659s` — asserts notice contains worktree path + `claude-profiles` dir + literal `NOT isolated` |
| AC-SW-017(d) two parallel profiles → distinct project worktrees, same global launch.yaml | PASS (structural) | distinct-worktree invariant proven by M5 `TestSessionWorktreeBranchName_ParallelDistinctSessions`; same-global-ledger is the EXPLICIT NON-ACTION (#3) + notice (#4) | (compositional: distinct-branches ⇒ distinct-paths; global ledger never redirected) |
| EC-7 read-only subverbs do NOT auto-enter | PASS | `go test -run TestProfileReadOnlySubverbs_DoNotInvokeAutoEntry ./internal/cli/` | `ok ... 0.659s` — sentinel `sessionWorktreeGitWorktreeAdd` fatals if invoked; list+current leave it dormant |

**E2 — Cross-Platform Build:**
```
$ go build ./...                          → exit 0 (no output)
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0 (no output)
```

**E3 — Coverage (function-level, M6-added code via `-run` M6 filter):**
```
$ go test -run '<M6 tests>' -coverprofile=/tmp/m6cov.out ./internal/cli/
ok  github.com/modu-ai/moai-adk/internal/cli  0.659s  coverage: 6.6% of statements
$ go tool cover -func=/tmp/m6cov.out | grep -E "emitProfileScopeNotice|enterSessionWorktree"
profile.go:75:        emitProfileScopeNotice   100.0%
session_worktree.go:101: enterSessionWorktree   58.3% (profile-suffix branch covered; other branches by M2 suite)
```
`emitProfileScopeNotice` (the M6-specific production logic) is **100%** covered. `runProfileSetup`'s wiring block is 0.0% in the headless run — the `huh` interactive wizard requires a TTY and cannot be executed in a unit test (same constraint M2 `runInit` / M3 `runWeb` face; the wrapper is tested directly + EC-7 proves the gating). Full-package coverage: `go test -cover ./internal/cli/` → **76.3%**.

**E4 — Subagent Boundary Grep (C-HRA-008):**
```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/profile.go internal/cli/profile_setup.go internal/cli/profile_setup_translations.go | grep -v '_test.go' | grep -v '^[^:]*:[0-9]*:[[:space:]]*//'
(no output — exit 1 = 0 matches = clean)
```

**E5 — Lint:**
```
$ golangci-lint run --timeout=3m ./internal/cli/...
0 issues.
```
(The `loadSessionWorktreeConfig` unused-`cmd` param is NOT an active finding — debt CARRIED per E10.)

**E6 — Branch HEAD + Push:** see commit SHAs below (M6 commit + push to `feat/SPEC-SESSION-WORKTREE-001`).

**E7 — Blocker Report:** none. The profile subverb mutation/read-only split was clean (no ambiguity): `runProfileSetup` is the only project-mutating subverb. The `delete` subverb mutates only the global profile dir (not the project tree), so per deliverable #5 it is correctly excluded from auto-entry.

**E8 — RED → GREEN proof + falsification round-trip:**
RED (verbatim, before `emitProfileScopeNotice` existed):
```
$ go test -run 'TestEmitProfileScopeNotice|...' ./internal/cli/
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/profile_worktree_test.go:70:2: undefined: emitProfileScopeNotice
internal/cli/profile_worktree_test.go:109:2: undefined: emitProfileScopeNotice
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
FAIL
```
GREEN: `ok github.com/modu-ai/moai-adk/internal/cli 0.659s`.
Falsification round-trip (profile-dir-NOT-isolated notice is load-bearing): `TestEmitProfileScopeNotice_FalsificationRoundTrip` constructs the FALSE notice (`"... IS isolated by this entry"`), asserts it does NOT contain `NOT isolated`, then asserts the honest `emitProfileScopeNotice` output DOES — proving the clause is load-bearing. If the notice wording were mutated to claim the dir IS isolated, `TestEmitProfileScopeNotice_LoadBearing` would fail on the `strings.Contains(got, "NOT isolated")` assertion.

**E9 — EC-7 read-only gate proof:** `TestProfileReadOnlySubverbs_DoNotInvokeAutoEntry` swaps `sessionWorktreeGitWorktreeAdd` to a sentinel that `t.Fatal`s if invoked, sets the feature ON (`MOAI_SESSION_WORKTREE=1`), and invokes `profileListCmd.RunE` + `profileCurrentCmd.RunE`. The sentinel stays dormant → the materialize seam is not reached by read-only subverbs.

**E10 — unused-cmd debt status: CARRIED.** `loadSessionWorktreeConfig(cmd)`'s `cmd` param remains unused. Reason: the profile command (M6's consumer) defines no `--root`/`--project` flag; profile setup resolves the project root via cwd (`os.Getwd()` + `filepath.Join(cwd, ".moai")`), so no flag improves on cwd. Closing the debt would require touching M2's `init` (which has `--root`) — out of M6 scope (Behavior #5 Scope Discipline). The lint baseline is clean (0 issues), so this is a documented future-proof, not an active defect.

**Scope discipline (B8/B10):** touched ONLY `internal/cli/profile.go` (notice helper + `io` import) + `internal/cli/profile_setup.go` (runProfileSetup signature → named return + wiring block) + `internal/cli/profile_worktree_test.go` (NEW, 3 tests) + this progress.md §E.2 evidence append. spec.md/plan.md/acceptance.md/design.md/research.md body unchanged. `internal/cli/session_worktree.go` core logic UNCHANGED (reused `enterSessionWorktree`/`cleanupSessionWorktree`/`loadSessionWorktreeConfig` as-is; the unused-cmd debt in `loadSessionWorktreeConfig` was NOT closed — see E10). spec.md status stays `in-progress`.

---

### M7 — Worktree-scoped git config helper (REQ-SW-018/019/020/021)

**Pre-flight findings:**
- HEAD before M7: `13caf0826` (M6 COMPLETE, independently verified) ✓.
- M7 call-site hook: `internal/cli/session_worktree.go:163` comment `// M7 (REQ-SW-018/019/021): ApplyGitConfig call site` ✓.
- M2 direct `init.defaultBranch=main` application: `sessionWorktreeGitConfigSet(wtPath, "init.defaultBranch", "main")` at the materialize step ✓ (kept in place — see consolidation decision below).
- Profile opt-in fields: `grep -rn "signingkey\|SigningKey\|gpgsign\|GPGSign" internal/profile/ internal/cli/ pkg/models/` → NO schema fields exist. REQ-SW-021 is a **verified no-op for now** (tracked schema debt — E7 blocker).
- M4 safe.directory-unset: NOT present before M7. Wired in M7 (E9 below).
- Git version detection helper: none pre-existed; added `gitVersionInfo` + `gitVersionReal` + `parseGitVersion`.

**Empirical grounding (git worktree config scoping):** verified at M7 pre-flight that `git -C <worktree> config user.name X` writes to the SHARED main repo `.git/config` (the linked worktree reads it), NOT a per-worktree config file. `git config --worktree` requires `extensions.worktreeConfig` (git ≥ 2.20) to be enabled first. Conclusion: our helper uses the universal `git -C <wt> config` (local repo config) path which works in all git versions; the `--worktree` flag is never required. This satisfies AC-SW-019's observable invariant — "global `~/.gitconfig` UNCHANGED + worktree commits under the identity" — because writing local repo config leaves the global gitconfig untouched. Cross-worktree identity isolation is OUT OF SCOPE per EC-12 (both worktrees read the same global gitconfig → same identity).

**Design decision (M2 consolidation):** the helper is ADDITIVE — M2's direct `init.defaultBranch=main` call in `materializeSessionWorktree` stays in place; the helper handles tiers 1 (safe.directory), 2 (identity), and 4 (options) only. Rationale: consolidating tier 3 into the helper would change M2's `TestEnterSessionWorktree_SuccessAppliesDefaultBranch` seam expectations and risk test-machine global-gitconfig pollution (the new safe.directory/identity seams default to real git). The additive design preserves every existing M2/M3/M4/M5/M6 test byte-identical and keeps the helper focused on the NEW tiers.

**Implementation (internal/cli/session_worktree.go):**
- 5 new test-injectable seams: `sessionWorktreeGitSafeDirAdd` / `SafeDirUnset` / `SafeDirGetAll` / `GlobalGet` / `GitVersion`, each with a Real counterpart.
- `applyWorktreeGitConfig(wtPath string, out io.Writer) worktreeGitConfigResult` — applies tiers 1/2/4 (tier 3 owned by M2), emits the identity notice (REQ-SW-019 / R7) + the git-version fallback notice (EC-9). Fail-open: every tier is best-effort.
- `worktreeGitConfigResult` struct — per-tier observable fields so tests assert without coupling to notice text.
- `gitVersionInfo` + `SupportsWorktreeConfig()` — git ≥ 2.20 boundary (BI-6). The helper's local-config path is version-independent; the predicate drives the EC-9 notice only.
- Wired into `materializeSessionWorktree` (replaced the M7 comment hook with `_ = applyWorktreeGitConfig(wtPath, out)`). `materializeSessionWorktree` now takes `out io.Writer` (threaded from `enterSessionWorktree`).
- **E9 — safe.directory unset on cleanup:** wired into `cleanupSessionWorktree` immediately after a successful `git worktree remove` (R5 mitigation). Best-effort; an unset of an already-absent entry is a no-op. Gated by the same auto_cleanup + clean-exit + clean-worktree conditions as the removal itself.

**AC matrix (M7-relevant):**

| AC | Status | Verification |
|----|--------|--------------|
| AC-SW-018 (safe.directory registered, idempotent, `git -C W status` exits 0) | PASS | `TestApplyGitConfig_SafeDirectoryRegistered` + `TestApplyGitConfig_SafeDirectoryIdempotent` (3 applies → 1 entry) + `TestApplyGitConfig_SafeDirectoryFalsification` (applied toggles with the seam) |
| AC-SW-019 (worktree commits under global user.name/email; global gitconfig UNCHANGED; no-global-identity no-op; profile NOT consulted) | PASS | `TestApplyGitConfig_IdentityAppliedFromGlobal` (both fields worktree-scoped via `git -C config`) + `TestApplyGitConfig_NoGlobalIdentityIsNoOp` (EC-14 verified no-op + notice) + `TestApplyGitConfig_PartialIdentity` (EC-8) + `TestApplyGitConfig_ProfileNotConsulted` (structural: helper takes no profile arg) + `TestApplyGitConfig_IdentityFalsification` (configSet fires ≥2× for name+email) |
| AC-SW-020 (init defaultBranch main — helper exposure + M2 wiring) | PASS | M2's direct call preserved in `materializeSessionWorktree` (additive design); `TestEnterSessionWorktree_SuccessAppliesDefaultBranch` (M2) still passes unchanged |
| AC-SW-021 (opt-in gpgsign/signingkey when present; absent = silent skip; hooksPath NOT set; new-repo sane defaults) | PASS (no-op) | `TestApplyGitConfig_OptionsNoOpNoProfileFields` — verified no-op: no REQ-SW-021 key is set; `core.hooksPath` is NEVER set (v0.2.1 removal). Schema gap recorded as E7 blocker. |

**E8 — RED → GREEN + falsification round-trips:**
RED (verbatim, before symbols existed):
```
$ go test -run 'TestApplyGitConfig|TestGitVersion|...' ./internal/cli/
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/session_worktree_m7_test.go:33:24: undefined: gitVersionInfo
internal/cli/session_worktree_m7_test.go:42:18: undefined: sessionWorktreeGitSafeDirAdd
...
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```
GREEN: `ok github.com/modu-ai/moai-adk/internal/cli 0.730s`.
Falsification round-trips:
- **(a) safe.directory:** `TestApplyGitConfig_SafeDirectoryFalsification` — with `safeDirAdd` present → `safeDirectoryApplied=true`; with `safeDirAdd` erroring → `safeDirectoryApplied=false`. The discrimination proves the tier is load-bearing.
- **(b) identity:** `TestApplyGitConfig_IdentityFalsification` — globalGet returns the identity AND configSet fires ≥2× (name+email); if the worktree-scoped apply were removed, configSet would fire 0× for user.name/email and the worktree would commit under git's default identity (wrong).
- **(c) no-global-identity no-op:** `TestApplyGitConfig_NoGlobalIdentityIsNoOp` — empty globalGet → NO user.name/email configSet call + `identityNoop=true` + notice names the no-op.
- **(d) idempotency:** `TestApplyGitConfig_SafeDirectoryIdempotent` — 3 applies of the same path → `get-all` shows it exactly once.

**E9 — safe.directory unset on cleanup (R5 mitigation):**
`TestCleanupSessionWorktree_UnsetsSafeDirectoryOnRemoval` — auto_cleanup ON + clean exit + clean worktree → `safeDirUnset` called once for the removed path. `TestCleanupSessionWorktree_DoesNotUnsetWhenPreserved` — negative controls: default-manual / non-clean-exit / dirty → `safeDirUnset` NOT called (the worktree still needs its entry). M8 PR-merge cleanup will unset via the same seam when it lands.

**E10 — git < 2.20 fallback (BI-6 / EC-9):**
`TestApplyGitConfig_GitUnder220FallbackNotice` — `gitVersion` returns `{2,19}` → `gitVersionFallback=true` + notice names "git" + "2.20". `TestApplyGitConfig_Git220PlusNoFallbackNotice` — `{2,50}` → `gitVersionFallback=false`, no notice. The application path is version-independent (`git -C config` writes local repo config in all git versions); the `--worktree` extension flag is never used, so no separate fallback code path is needed — the notice is the user-facing EC-9 signal. Residual: the real `gitSafeDirAddReal` / `gitGlobalGetReal` / `gitVersionReal` error-path branches (25% uncovered) require a git failure to hit; happy paths covered by `TestGitVersionReal_ParsesModernGit`.

**E2 — Cross-platform build:**
```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

**E3 — Coverage (M7 helper + tiers, file-level):**
```
$ go test -cover -coverpkg=./internal/cli ./internal/cli/ (session_worktree tests)
materializeSessionWorktree      100.0%
applyWorktreeGitConfig           89.5%
emitWorktreeGitConfigNotice      85.7%
cleanupSessionWorktree           89.5%   (includes new safe.directory unset)
parseGitVersion                  86.7%
gitVersionReal                   75.0%   (error branch needs git failure)
gitGlobalGetReal                 75.0%
gitSafeDirAddReal                75.0%
gitSafeDirUnsetReal              75.0%
gitSafeDirGetAllReal              0.0%   (test-verification helper; real impl thin wrapper, seam exercised)
```
All NEW-logic functions ≥85%; git-exec Real wrappers at 75% (error-path branches need git failures to hit; happy paths covered).

**E4 — Subagent boundary grep (C-HRA-008):** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/session_worktree.go internal/cli/session_worktree_m7_test.go` → 0 NEW matches (exit 1). PASS.

**E5 — Lint:** `golangci-lint run --timeout=3m ./internal/cli/...` → 0 issues (one SA9003 empty-branch on the M2 init.defaultBranch call site fixed by switching to `_ =` best-effort).

**E6 — Branch HEAD + push:** M7 commit + push to `feat/SPEC-SESSION-WORKTREE-001` (this commit).

**E7 — Blocker report (REQ-SW-021 schema gap):** `ProfilePreferences` (`internal/profile/preferences.go`) carries NO `signingkey` / `gpgsign` / `gpg` opt-in fields (verified `grep -rn "signingkey\|SigningKey\|gpgsign\|GPGSign" internal/profile/ pkg/models/` → 0 matches outside tests). REQ-SW-021 is therefore a **verified no-op for now**: tier 4 applies nothing. This is NOT a defect — the SPEC (plan.md §F M7) explicitly says "if `signingkey`/gpg fields do not yet exist, this tier is a verified no-op for now (document; do NOT invent a schema)". Tracked as schema debt: when a profile schema change adds the opt-in fields, tier 4 is implemented (apply `commit.gpgsign=true` + `user.signingkey=<key>` worktree-scoped; new-repo defaults `core.autocrlf=input` / `fetch.prune=true` / `push.default=current`; `core.hooksPath` NEVER set per v0.2.1). The helper's `optionsApplied` field is reserved for that future state.

**Global-state minimality (acceptance.md §E):** `grep -n '\-\-global' internal/cli/session_worktree.go` → all `--global` uses are on `safe.directory` (REQ-SW-018 add + R5 unset + get-all verification) and `--get` reads (REQ-SW-019 read-only identity). NO other global mutation. PASS.

**Scope discipline (B8/B10):** touched ONLY `internal/cli/session_worktree.go` (5 new seams + Real impls + `gitVersionInfo` + `applyWorktreeGitConfig` + `emitWorktreeGitConfigNotice` + wiring in `materializeSessionWorktree` + safe.directory unset in `cleanupSessionWorktree` + `materializeSessionWorktree` signature `+out io.Writer`) + `internal/cli/session_worktree_m7_test.go` (NEW, 14 tests) + this progress.md §E.2 evidence append. spec.md/plan.md/acceptance.md/design.md/research.md body unchanged. spec.md status stays `in-progress`.

---

### M8 — PR-merge auto-cleanup (on-touch session register / list) — REQ-SW-022/023/024

**Trigger-point correction (v0.2.2 D-NEW-1 inline-fix):** the v0.2.1 plan named `moai worktree list` / `moai session start` as the on-touch trigger sites. Both were non-existent: `worktree list` is intentionally retired (`internal/cli/worktree/root_test.go:61`) and `session start` never existed (`internal/cli/session.go:53-59` — the session command tree has register/heartbeat/deregister/list/purge/current/doctor, no `start`). Manager-spec corrected the SPEC to v0.2.2 naming the verified-existing pair `moai session register` (`internal/cli/session.go:66` newSessionRegisterCmd) + `moai session list` (`internal/cli/session.go:132` newSessionListCmd). Q3 rationale shape (on-touch, fires when user present, stale-until-next-invocation acceptable) preserved; only the command names changed. This M8 run implements the v0.2.2 trigger points.

**Deliverables:**
- NEW `internal/cli/session_worktree_prmerge.go`: `prMergeCleanup(cfg, out)` helper + `PRMergeCleanupNoticePrefix` constant + 4 seams (`sessionWorktreeGitWorktreeList`, `sessionWorktreeGhLookPath`, `sessionWorktreeGhPRViewState`, `sessionWorktreeGitBranchMerged`) + Real impls + `parseWorktreeList` porcelain parser + `parseGhPRStateJSON` + `branchMergedForCleanup` decision helper. `@MX:ANCHOR` on `prMergeCleanup` (REQ-SW-022/023/024 — the on-touch sweep; one helper, gh primary + squash-blind fallback, dirty guard re-checked before removal EC-11, notice distinct from session-exit).
- Wired on-touch invocation at TWO RunE sites in `internal/cli/session.go`: `newSessionRegisterCmd` RunE (line ~70) + `newSessionListCmd` RunE (line ~136). Each calls `prMergeCleanup(loadSessionWorktreeConfig(cmd), cmd.ErrOrStderr())` BEFORE the subcommand's main work, gated by the AutoCleanup toggle inside the helper.
- NEW `internal/cli/session_worktree_prmerge_test.go`: 23 tests (toggle-off, nil-cfg, gh-primary-merged, gh-primary-open, gh-primary-squash-sees-merge, gh-absent-branch-merged, squash-blind-fallback-preserves, gh-absent-blindness-notice-once, dirty-preserve, dirty-check-error-preserve, worktree-list-error-fail-open, only-WT-branches, detached-ignored, removal-failure-fail-open, fallback-branch-merged-error-preserve, notice-distinguishable, on-touch-fires-register, on-touch-fires-list, on-touch-toggle-off-register, trigger-invariant-static, parseWorktreeList ×2, parseGhPRStateJSON ×1) + `chdirTemp` helper + `writeAutoCleanupConfig` helper.

**Reused surface (no duplication):**
- M4 `worktreeIsDirty` (`session_worktree.go:634`) — the SHARED dirty guard, one helper two call sites (REQ-SW-010 + REQ-SW-024). Re-checked immediately before removal for EC-11 race.
- M7 `sessionWorktreeGitSafeDirUnset` / `gitSafeDirUnsetReal` (`session_worktree.go:94`/`:301`) — the R5 safe.directory-unset seam, reused on the PR-merge removal path.
- M4 `sessionWorktreeGitWorktreeRemove` (`session_worktree.go:73`) — the `git worktree remove` seam, reused.
- `SessionWorktreeBranchPrefix = "WT-"` (`session_worktree.go:39`) — the WT- filter predicate.
- `cfg.Workflow.Worktree.AutoCleanup` (`types.go:492`, default false) — the SAME toggle as session-exit cleanup (REQ-SW-022); no new flag.
- `gh` runtime detection via `exec.LookPath("gh")` — pattern already assumed at `doctor.go:326` + `wizard/config_helpers.go:77` (BI-8).
- `loadSessionWorktreeConfig(cmd)` (`session_worktree.go:670`) — reads config from cwd, returns nil on failure (fail-safe to OFF).

**E1 — AC matrix (M8):**

| AC | Status | Verification Command | Actual Output |
|---|---|---|---|
| AC-SW-022 (on-touch at register+list, distinguishable notice, toggle-off negative control, trigger invariant ONLY two subcommands) | PASS | `go test ./internal/cli/ -run 'TestPRMergeCleanup_OnTouchFiresAtSessionRegister\|TestPRMergeCleanup_OnTouchFiresAtSessionList\|TestPRMergeCleanup_OnTouchToggleOffSkipsAtSessionRegister\|TestPRMergeCleanup_TriggerInvariant_OtherSubcommandsDoNotFire\|TestPRMergeCleanup_NoticeDistinguishableFromSessionExit' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli` — register RunE invokes prMergeCleanup (git worktree list called); list RunE invokes it; toggle-off register does NOT invoke it; TriggerInvariant asserts session.go `prMergeCleanup(` count == 2 (register+list only); notice prefix `"removed by PR-merge cleanup:"` ≠ `"removed by session-exit cleanup:"` |
| AC-SW-023 (gh primary MERGED; git branch --merged fallback; squash-blind documented+tested; primary sees squash merges) | PASS | `go test ./internal/cli/ -run 'TestPRMergeCleanup_GhPresentMergedRemoves\|TestPRMergeCleanup_GhPresentSeesSquashMerge\|TestPRMergeCleanup_GhAbsentBranchMergedRemoves\|TestPRMergeCleanup_SquashBlindFallbackPreserves\|TestPRMergeCleanup_GhAbsentEmitsBlindnessNoticeOnce' -count=1` | `ok` — gh+MERGED→removed+notice; gh+squash-merge(state==MERGED)→removed (primary sees it); gh-absent+branch-in---merged→removed; gh-absent+squash(branch NOT in --merged)→NOT removed + "squash-merge blind" notice emitted exactly once; blindness notice documents the fallback's limitation |
| AC-SW-024 (dirty preserved on PR-merge path) | PASS | `go test ./internal/cli/ -run 'TestPRMergeCleanup_DirtyPreserves\|TestPRMergeCleanup_DirtyCheckErrorPreserves' -count=1` | `ok` — merged+dirty→NOT removed + "preserved" notice names path; merged+status-error→NOT removed (fail-open preserve, EC-11) |

**E2 — Cross-platform build:**
```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

**E3 — Coverage (M8 logic, `go test -coverprofile`):**
```
session_worktree_prmerge.go:
  parseWorktreeList        100.0%
  prMergeCleanup           100.0%
  branchMergedForCleanup   100.0%
  parseGhPRStateJSON       100.0%
  gitWorktreeListReal        0.0%   (seam-swapped Real; shells out to git, exercised via seam)
  ghLookPathReal             0.0%   (seam-swapped Real; shells out to LookPath, exercised via seam)
  ghPRViewStateReal          0.0%   (seam-swapped Real; shells out to gh, exercised via seam)
  gitBranchMergedReal        0.0%   (seam-swapped Real; shells out to git, exercised via seam)
```
All M8 LOGIC functions at 100%. The four `*Real` shell-out wrappers are 0% by design — they are swapped via function-variable seams in every test (matching the M4 `gitWorktreeAddReal` / M7 `gitSafeDirAddReal` pattern: the real impls are thin `exec.Command` wrappers not unit-tested; the seams carry the behavioral coverage). Package-wide `go test -cover ./internal/cli/` on the M8 run: 7.2% of statements (the M8 subset).

**E4 — Subagent boundary grep (C-HRA-008):** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/session_worktree_prmerge.go internal/cli/session.go | grep -v "_test.go" | grep -v "// "` → **0 NEW matches**. PASS. (CLI returns exit codes + notices to stderr only; orchestrator owns user interaction per the session.go header invariant.)

**E5 — Lint:** `golangci-lint run --timeout=3m ./internal/cli/...` → **0 issues**. PASS. `go vet ./internal/cli/...` → exit 0.

**E6 — Branch HEAD + push:** M8 commit + push to `feat/SPEC-SESSION-WORKTREE-001` (this commit).

**E7 — Blocker report:** NONE. The v0.2.2 trigger-point blocker (prior M8 attempt) is RESOLVED — `session register` + `session list` are verified-existing CLI commands. No new blocker raised this run.

**E8 — RED → GREEN + falsification round-trips:**
- **(a) dirty guard (REQ-SW-024):** RED evidence — the M8 tests were written first referencing `prMergeCleanup` / `PRMergeCleanupNoticePrefix` / the 4 seams before any existed; `go test` failed with `undefined: sessionWorktreeGitWorktreeList ...` (build failed, verbatim RED). Falsification round-trip: `sed`-bypass `worktreeIsDirty` → `dirty=false` → `TestPRMergeCleanup_DirtyPreserves` + `TestPRMergeCleanup_DirtyCheckErrorPreserves` **FAIL** (merged-but-dirty wrongly removed / status-error wrongly removed); restore → **PASS**. Observed exit: FAIL then ok.
- **(b) trigger invariant (AC-SW-022):** falsification round-trip — `perl` fully removed ONE `prMergeCleanup(...)` call line from session.go (count 2→1) → `TestPRMergeCleanup_TriggerInvariant_OtherSubcommandsDoNotFire` **FAIL** (`expected exactly 2 ... got 1`); restore (count 1→2) → **PASS**. Observed: FAIL then ok. A NON-trigger subcommand cannot fire prMergeCleanup because the call sites exist ONLY in `newSessionRegisterCmd` + `newSessionListCmd` RunE (the static count guard enforces this).
- **(c) squash-blind fallback (REQ-SW-023):** `TestPRMergeCleanup_SquashBlindFallbackPreserves` — gh absent + squash-merged (branch NOT in `git branch --merged`) → NOT removed + "squash-merge blind" notice. The complementary `TestPRMergeCleanup_GhPresentSeesSquashMerge` proves the primary path catches what the fallback misses.
- **(d) toggle-off negative control (REQ-SW-022):** `TestPRMergeCleanup_ToggleOffNoOp` (helper-level: AutoCleanup=false → no git worktree list invocation, no notice) + `TestPRMergeCleanup_OnTouchToggleOffSkipsAtSessionRegister` (integration-level: register RunE with no config → loadSessionWorktreeConfig returns nil → prMergeCleanup no-op → list seam NOT called).

**E9 — Notice distinguishability (EC-13):** `PRMergeCleanupNoticePrefix = "removed by PR-merge cleanup:"` (verbatim, `session_worktree_prmerge.go:47`). Asserted ≠ `SessionExitCleanupNoticePrefix = "removed by session-exit cleanup:"` via `TestPRMergeCleanup_NoticeDistinguishableFromSessionExit`. EC-13 combined-output attribution: the two notices carry distinct prefixes (`PR-merge` vs `session-exit`) so a user reading combined stderr can attribute each removal. The M8 on-touch path (AC-SW-022) fires at `session register`/`session list` BEFORE session-exit cleanup (AC-SW-009) fires at subcommand exit.

**E10 — on-touch wiring sites:**
1. `internal/cli/session.go` `newSessionRegisterCmd().RunE` (line ~70): `prMergeCleanup(loadSessionWorktreeConfig(cmd), cmd.ErrOrStderr())` — gated by `cfg.Workflow.Worktree.AutoCleanup` inside the helper.
2. `internal/cli/session.go` `newSessionListCmd().RunE` (line ~136): same call, same gate.
Both fire BEFORE the subcommand's main work (RegisterSession / QueryActiveWork). Fail-open, non-blocking. Static guard `TestPRMergeCleanup_TriggerInvariant_OtherSubcommandsDoNotFire` asserts `prMergeCleanup(` appears exactly twice in session.go (register + list only).

**Pre-existing baseline note (NOT M8-caused):** `TestDoctorGolden_{Light,Dark,NoColor}` fail at the pre-M8 HEAD (`0350b9c907`) — verified by `git stash`-isolating the M8 changes and re-running: the 3 doctor golden tests FAIL identically without any M8 code present. M8 touches `session.go` (2 lines added) + 2 NEW files (`session_worktree_prmerge.go` + its test); doctor golden is a separate subsystem (golden-file comparison for `moai doctor` rendered output) with no symbol overlap. Running the full cli suite with `-skip 'TestDoctorGolden'` → `ok` (all other tests pass). The doctor golden failures are pre-existing and out of M8 scope.

**Scope discipline (B8/B10):** touched ONLY `internal/cli/session_worktree_prmerge.go` (NEW) + `internal/cli/session_worktree_prmerge_test.go` (NEW) + `internal/cli/session.go` (2 on-touch call lines, 2 comment blocks) + this progress.md §E.2 evidence append. spec.md/plan.md/acceptance.md/design.md/research.md body unchanged (manager-spec owns them; v0.2.2 already corrected). spec.md status stays `in-progress`.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- tier: L
- scope (file count): 11
- domain count: 3 (config loader, CLI entrypoints init/profile/web, worktree subsystem)
- file language mix: Go (100%)
- concurrency benefit: LOW (coding-heavy, per Anthropic coding-task parallelism caveat)
- Decision: sub-agent (Mode 5) — sequential per-milestone delegation
- Justification: Tier L coding-heavy work; Anthropic caveat "most coding tasks involve fewer truly parallelizable tasks than research". Mode 5 sequential sub-agent is the safe default. M1→M8 ordered by decision-reversibility (config flag first, PR-merge cleanup last).
- Implementation Kickoff Approval: PASSED (user explicit, this session).
- plan-audit: iter-2 PASS (0.92).
