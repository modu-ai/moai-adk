# Progress — SPEC-SESSION-WORKTREE-001

> **Tier L** lifecycle progress. §E section skeleton emitted at plan-phase; §E.2-§E.4 populated only by the respective phase-owning agents.

## §E.1 Plan-phase Audit-Ready Signal

_Populated at plan-phase close._

- plan_status: _pending iter-2 audit (v0.2.1 iter-2 audit-fix; iter-1 verdict FAIL 0.69)_
- plan_complete_at: _pending_
- artifacts: spec.md v0.2.1, plan.md v0.2.1, acceptance.md v0.2.1, design.md v0.2.1, research.md v0.2.1, progress.md (this file)
- tier: L (retained at v0.2.1 — REQ/AC budget 24/24, file surface ≥10; v0.2.0 escalation M→L stands)
- version_trace: v0.1.0 (initial Tier M) → v0.2.0 (tier escalated M→L, 3 additions) → v0.2.1 (iter-2 audit-fix: D1 gitconfig source, D2 Q1-Q4 resolved, D3 defaultBranchDetectorFunc documented, D4 on-touch trigger, D5 path-distinguishing notices, D6 REQ-SW-014 Event-driven relabel, D7 REQ-SW-021 SHALL)
- lint: _pending `moai spec lint`_
- self_check: SPEC ID regex PASS observed; frontmatter 12-field check pending; GEARS notation verified; Out-of-Scope rule satisfied (11 `### Out of Scope —` H3 sub-headings at v0.2.1 — 9 carried from v0.2.0 + 2 added: Profile-vs-profile git identity + Hook isolation); Tier L artifact set complete (spec + plan + acceptance + design + research + progress); all v0.2.0 `[NEEDS CLARIFICATION]` markers resolved (no more blocker rounds).

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
