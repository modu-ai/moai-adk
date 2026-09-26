# SPEC-CODEX-FACTORY-RETIRE-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
spec_id: SPEC-CODEX-FACTORY-RETIRE-001
spec_version: 0.3.0
card: t1242
tier: L
base_tree: 553e224f3
branch: WT-codex-factory-retire
artifacts: [spec.md, plan.md, acceptance.md, design.md, research.md, progress.md]
requirements: 23
acceptance_criteria: 25
open_clarification_markers: 0
plan_audit_history:
  - "iter-1 FAIL 0.71 -> revised in v0.2.0"
  - "iter-2 FAIL 0.83 -> revised in v0.3.0"
  - "iter-3 PASS 0.91 (.moai/reports/plan-audit/SPEC-CODEX-FACTORY-RETIRE-001-review-3.md)"
plan_audit_verdict: PASS
implementation_kickoff:
  approved: true
  approved_by: operator (relayed by the lead, verbatim "승인 · 반자율")
  recorded_at: 2026-09-26T06:22:15Z   # time the approval reached this agent, not the operator's click time
  progression_mode: semi-autonomous   # pause after each of M1-M4 for operator review; M5 only in the merge window after lead confirmation
status: draft
```

- Coordinates re-measured on `553e224f3`; divergences from the design recorded in
  `research.md` §R2.
- The `[NEEDS CLARIFICATION: exact 6 paths from lead]` marker was resolved in this
  plan phase with the lead-provided list (plan.md §M5, research.md §R6). The list is the
  lead's observation; this agent did not read the develop worktree.
- Token-budget reading on the plan tree: 77539 / 77600 (headroom 61). Context only;
  AC-CFR-022 compares the merge commit with its develop parent.
- Fields the run lane writes into §E.2 (plan.md §C, §M5): `run_base:`, `budget_base:`,
  the AC-CFR-022 figure pair, the AC-CFR-020 mutant result, `m5_lead_confirmation:`
  (citing the lead-authored `m5-lead-confirm.md`, before M5 step 1), and the
  foreign-6.patch sha256 (after the confirmation line).

## §E.2 Run-phase Evidence

### Pre-flight (plan.md §C) — card worktree `.claude/worktrees/t1242`, branch `WT-codex-factory-retire`

- Arrival HEAD `d622d2373`; local develop `e62c3e183` was 15 commits ahead of the plan base
  (`git rev-list --count --left-right develop...HEAD` → `15	5`), and 7 ahead of origin/develop
  (`0	7`, lead batch-push pending). Absorbed with `git merge --no-ff develop` → `d2fc5cd69`.
- `run_base: e62c3e183` (`git merge-base develop HEAD` after the absorb; every `<base>` in
  acceptance.md means this value).
- `budget_base: 77539` — `go test ./internal/config -run 'TestAlwaysLoadedTokenBudget$' -count=1 -v`
  → `always-loaded surface = 77539 tokens (budget 77600, headroom 61, 16 entries)`, `ok`
  (measured on `d2fc5cd69`; context only, AC-CFR-022 compares merge vs merge^1).
- Builds on `d2fc5cd69`: `go build ./...` rc=0, `GOOS=windows GOARCH=amd64 go build ./...` rc=0,
  `GOOS=linux go build ./...` rc=0.
- Lint baseline on `d2fc5cd69`: `golangci-lint run --timeout=5m ./internal/cli/...` → `0 issues.` rc=0.

### M1 — entry refusal, lane-env defense, codex-run join refusal, codex hook peer refusal

Evidence files (card worktree, gitignored): `.moai/reports/t1242/m1/{red-1,red-2,red-3,green-1,green-2,green-3,cli-full}.txt`, coverage profile `.moai/reports/t1242/m1/c.out`.

RED (E8) — captured before any behavior change (only the `codexLaneLaunchEnvKeys` list declaration
existed): 35 `--- FAIL` lines across red-1..3. Verbatim excerpts:
`codex [-k]: err = <nil>, want exit code 1` · `codex [-f]: err = <nil>, want exit code 1` ·
`factory state DB changed: absent -> 017db7c6…` · 11/11 `TestCodexChildEnvScrubsLaneKeys/<KEY>` FAIL ·
11/11 `TestCodexSpawnCommandBlanksLaneKeys/<KEY>` FAIL · `join accepted, want a refusal` (×3) ·
`lead adopted a codex run, want a refusal` · `after --harness codex: peers for (r1, worker-1) = 1, want 0`
and `(r1, lead) = 1, want 0`. Preservation guards green at arrival, as designed: AC-CFR-004
(`TestCodexEntryTokensAfterDashDashPassThrough`), AC-CFR-009 (`TestFactoryJoinAndLeadAcceptNonCodexRun`).
The pre-change RED of AC-CFR-025 is the "guard removed" mutant (both shapes read 1).

| AC | Status | Command | Actual output |
|----|--------|---------|---------------|
| AC-CFR-001 | PASS | `go test ./internal/cli -run TestCodexKanbanEntryIsRefused -v` + built binary `moai codex -k SPEC-X-001` / `-k --spawn` | `--- PASS: TestCodexKanbanEntryIsRefused`; binary stderr `KANBAN_MODE_UNSUPPORTED_BACKEND: moai codex no longer enters Kanban Mode; use 'moai cc -k' or 'moai glm -k' instead`, rc=1 |
| AC-CFR-002 | PASS | `… -run TestCodexFactoryEntryIsRefused -v` + binary `codex --factory-run r1`, `-f status`, `-f worker-2` | `--- PASS: TestCodexFactoryEntryIsRefused`; binary stderr `FACTORY_MODE_UNSUPPORTED_BACKEND: moai codex no longer enters Factory Mode; use 'moai cc -f' or 'moai glm -f' instead`, rc=1; `-f worker-2 2>/dev/null` stdout empty, rc=1 |
| AC-CFR-003 | PASS | `… -run TestCodexEntryRefusalHasNoStateEffect -v` | `--- PASS` (20 refusal shapes: 0 launches, factory.db + workers.json absent→absent, lane env unchanged, no `.claude/worktrees`; bare form = 1 launch) |
| AC-CFR-004 | PASS | `… -run TestCodexEntryTokensAfterDashDashPassThrough -v` | `--- PASS` (1 launch, argv tail `-f worker -k`) |
| AC-CFR-005 | PASS | `… -run TestCodexChildEnvScrubsLaneKeys -v` | `--- PASS` + 11 per-key subtests PASS; posture keys, `T1242_KEEP`, `MOAI_HOME`, resolved `CODEX_HOME` kept; session pair absent |
| AC-CFR-006 | PASS | `… -run TestCodexSpawnCommandBlanksLaneKeys -v` | `--- PASS` + 11 per-key subtests; `MOAI_HOME=<set value>` present |
| AC-CFR-007 | PASS | `… -run 'TestCodexDirectPOSIXExecRegistersNoFactoryPeer\|TestCodexSpawnUnderLaneEnvRegistersNoFactoryPeer' -v`; control `… -run '^TestFactoryLauncherRegistersLaunchPendingPeers$' -v` | both `--- PASS` (0 lanes, owner `(424242, t1242-lead)` unchanged; POSIX exec pid/cwd/argv preserved); control `--- PASS: TestFactoryLauncherRegistersLaunchPendingPeers` |
| AC-CFR-008 | PASS | `… -run TestFactoryJoinRefusesCodexLedRun -v` | `--- PASS` cc `-f worker`, cc `-f worker-2`, glm `-f worker`: message has `rc`/`codex`/`moai factory runs --retire`; 0 launches; factory.db digest + worker rows unchanged |
| AC-CFR-009 | PASS | `… -run TestFactoryJoinAndLeadAcceptNonCodexRun -v` | `--- PASS` worker (1 launch, +1 slot) and lead (1 launch, run `(active, claude)`) |
| AC-CFR-010 | PASS | `… -run TestFactoryLeadRefusesCodexLedRun -v` | `--- PASS` (refused; `lead_backend`, `lead_pid` unchanged) |
| AC-CFR-025 | PASS | `… -run TestCodexHarnessHooksRegisterNoFactoryPeer -v` | `--- PASS` worker + lead shapes: codex-harness count 0, Claude count 1 |
| AC-CFR-013 (pre-check) | PASS | the nine-name `go test ./internal/cli -run '^(TestWorktreeLaunchRejectsConcurrentWriter\|…\|TestCodexAuditVerbRunsInCallerWorktree)$' -count=1 -v`; `grep -c 'syscall.Exec' codex_direct_posix.go`; `head -1` both direct files | 9 × `--- PASS`, `ok`; grep 2; `//go:build !windows` / `//go:build windows` |
| AC-CFR-017 (guard) | PASS | `git diff --exit-code develop...HEAD -- internal/factorymsg internal/homestate ':!*_test.go'` + same on the working tree | rc=0 / rc=0, empty |
| AC-CFR-021 (guard) | PASS | `git diff --exit-code develop...HEAD -- internal/codexwiring/configtoml.go .codex/config.toml`; `grep -c MOAI_KANBAN_ID internal/codexwiring/configtoml.go` | rc=0; `1` |

Package verification: `go test ./internal/cli/ -count=1 -timeout 35m` → `ok  github.com/modu-ai/moai-adk/internal/cli  1370.340s`
(run under `moai slot` resource `go-test-heavy`; an earlier run at the default 10m timeout panicked
on the timeout while `TestGateCmd_DegradedRunNeverReacquires` was running, and surfaced the
`TestRestampSeamIsCalledAtEveryNonReplaceCallSite` codex-door rows, fixed before the rerun).
`go vet ./internal/cli` rc=0 (darwin and `GOOS=windows`); `gofmt -l internal/cli` empty.

Coverage (§D.2, `go tool cover -func=.moai/reports/t1242/m1/c.out`): `codexEntryRefusal` 100.0%,
`codexChildEnv` 100.0%, `buildCodexSpawnCommand` 100.0%, `refuseCodexLedRun` 90.0%,
`unsetLaneEnvForCodexHook` 100.0% (also `defaultCodexSpawnLaunch` 93.8%, `runCodex` 89.7%,
`defaultCodexDirectLaunch` POSIX 88.9% — the exec success path runs in a helper subprocess).

Lint (E5): `golangci-lint run --timeout=5m ./internal/cli/...` → 5 NEW `unused` findings, all
transitional dead code M2 deletes by plan: `codex_kanban.go` `codexKanbanUsageDiag`,
`codexKanbanEntry`, `stripCodexKanbanFlag`, `applyCodexKanbanEntry` (plan M2.2) and
`factory.go` `stripFactoryRunFlag` (plan M2.5). `clearFactoryRunOwner` (Windows-only caller)
was kept reachable by a direct test instead.

Tests changed in M1 because M1 removes their subject (listed for M2's accounting):
`codex_kanban_test.go` deleted (whole file drove `moai codex -k`); `TestCodexDirectPOSIXExecPreservesFactoryOwner`
→ `TestCodexDirectPOSIXExecRegistersNoFactoryPeer`; `TestFactoryCodexSpawnRegistersLaunchPendingPeer`
→ `TestCodexSpawnUnderLaneEnvRegistersNoFactoryPeer`; `TestPaneDoorRefusalLeavesNoLauncherStampedRun`
/ `TestPaneDoorAnchorRefusalClearsFactoryRunOwner` → `TestCodexPaneDoor{Identity,Anchor}RefusalLeavesRunOwner`;
`TestRestampSeamIsCalledAtEveryNonReplaceCallSite` codex doors moved to the absent set;
`TestFactoryOperationalFixtureUsesProductionInit` now asserts the codex-harness hook binds no peer;
`TestCodexSpawn_RealAssembly*` expect the eleven `KEY=` blanks; the `-f` forms dropped from
`TestCodexLocalInstructions_AllFunnelsPreserveInputs`.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
