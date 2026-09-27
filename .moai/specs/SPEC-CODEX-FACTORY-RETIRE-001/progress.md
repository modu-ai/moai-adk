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

### M2 — dead entry code and test moves (base `9f53ac65f`, M1 commit)

Evidence files (card worktree, gitignored): `.moai/reports/t1242/m2-{targeted,ac013-014,m1regress-1,m1regress-2,cli-full}.txt`.

RED-now at M2 start (tree `9f53ac65f`):
- AC-CFR-011: `grep -rnwE '<AC-011 pattern>' internal cmd pkg --include='*.go' | wc -l` → `31` (want 0);
  `ls internal/cli/codex_kanban.go internal/cli/codex_factory.go` → both listed, rc=0.
- AC-CFR-014: `ls internal/cli/codex_factory_test.go` → listed, rc=0.
- Lint carry-over from M1: 5 `unused` (codex_kanban.go ×4, `stripFactoryRunFlag`).

Order kept (plan §G): `TestNextFactoryWorkerNumber` + `NextFactoryWorkerNumberForTest` moved into
`factory_worker_naming_test.go` and run green (`--- PASS: TestNextFactoryWorkerNumber`) before
`codex_factory_test.go` / `codex_factory_helper_test.go` were deleted.

Changes: deleted `codex_factory.go`, `codex_kanban.go`, `codex_factory_test.go`,
`codex_factory_helper_test.go`; removed `stripFactoryRunFlag` (plan M2.5 re-grep: 0 callers);
removed `TestStripCodexFactoryFlagWorkerVocabulary`; codex LIVE cases removed from
`factory_live_test.go` (6 funcs) and the two `moai codex -f` operational live proofs from
`factory_operational_live_test.go` together with the 10 helpers + one struct field only they used
(each flagged by `unused` after the removal); POSIX failure fixture carries no lane identity
(`TestCodexDirectPOSIX{ChdirFailure,ExecFailure}IsReported`); fixtures switched to claude/glm in
`factory_mixed_test.go` (4 values) and `factory_operational_fixture_test.go` (3 values).
M3 files (`factory_lane_handoff*.go`, `factory_handoff_abandon_test.go`) untouched.

| AC | Status | Command | Actual output |
|----|--------|---------|---------------|
| AC-CFR-011 | PASS | `grep -rnwE 'stripCodexKanbanFlag\|applyCodexKanbanEntry\|codexKanbanEntry\|codexKanbanUsageDiag\|stripCodexFactoryFlag\|applyCodexFactoryEntry\|codexFactoryBackend' internal cmd pkg --include='*.go'`; `ls internal/cli/codex_kanban.go internal/cli/codex_factory.go`; control `git grep -nwE '<same>' 553e224f3 -- internal \| wc -l` | grep 0 lines rc=1; ls "No such file or directory" ×2 rc=1; control `39` |
| AC-CFR-012 | PASS | `grep -nE 'registerFactoryLaunchPending\|…\|stripFactoryRunFlag' codex_launcher.go codex_direct_posix.go codex_direct_windows.go`; `grep -c registerFactoryLaunchPending internal/cli/launch_exec_posix.go` | 0 lines rc=1; `1` |
| AC-CFR-013 | PASS | nine-name `go test ./internal/cli -run '^(…)$' -count=1 -v`; `grep -c 'syscall.Exec'`; `head -1` | 9 × `--- PASS`, `ok`; `2`; `//go:build !windows` / `//go:build windows` |
| AC-CFR-014 | PASS | `go test ./internal/cli -run '^TestNextFactoryWorkerNumber$' -count=1 -v`; `ls internal/cli/codex_factory_test.go` | `--- PASS: TestNextFactoryWorkerNumber`; "No such file or directory" rc=1 |
| AC-CFR-015 | PASS | `go build ./...`; `GOOS=windows GOARCH=amd64 go build ./...`; `GOOS=linux go build ./...` | rc=0 ×3 |

M1 regression (same tree): all M1 AC tests `--- PASS` (m2-m1regress-1/2.txt).
Package: `go test ./internal/cli/ -count=1 -timeout 38m` → `ok  github.com/modu-ai/moai-adk/internal/cli  1121.582s`
(under `moai slot` `go-test-heavy`, acquired after the prior holder released; no RESIDUE line).
Lint (darwin): `golangci-lint run --timeout=5m ./internal/cli/...` → `0 issues.` (M1's 5 unused gone).
`GOOS=windows golangci-lint …` → 4 issues (`mcp_claude_process_windows.go` errcheck ×2,
`web_port.go` `moaiProcessName` unused, `factory_launch_pending.go` `rollbackFactoryLaunchPending`
unused); none in files M2 changed; the last one's only callers on develop were already POSIX-only
(`git grep -n 'rollbackFactoryLaunchPending(' develop -- internal` → `codex_direct_posix.go`,
`launch_exec_posix.go`), so it was unused on windows before this card. The windows lint baseline
itself was not measured on develop (gap).

### M3 — lane handoff CLI and codex relocation (base `5333a9e4d`, M2 commit)

Evidence files (card worktree, gitignored): `.moai/reports/t1242/m3-{mutant-conditional-guards,mutant-unconditional,hook-factory,cli-full,regress}.txt`.

Pre-flight §C.3 re-grep on `5333a9e4d`: every symbol slated for removal
(`prepareLaneHandoff`, `switchLaneHandoff{Interactive,Headless}`, `bindLaneHandoffHeadless`,
`recoverLaneHandoff`, `codexThreadRelocation`, `runCodexThreadRelocation`, `laneHandoffAppServer`,
`DefaultCodexHandoffRelocationTimeout`, `codexMethodThreadFork`, `codexNotifyThreadStarted`) had
production hits only in the four handoff files, `mcp_codex.go` and `config/defaults.go` — no new
production caller, M3 proceeds.

Order kept (plan §G): the fixture moved to `factory_handoff_fixture_test.go` (states driven by
`ReserveHandoff` / `MarkHandoffWTReady`; the target is a real `git worktree add -b WT-lane-handoff`
on the develop pin) and `TestFactoryLaneHandoffOperatorAbandon` ran green on it, with the eight old
handoff test files removed but all production handoff files still present. Only then were the four
production files deleted.

AC-CFR-020 two-cell evidence (E8), mutant = `wtReady` without the `materializeTarget` step:
- conditional guards (the arrival shape, `if pathExists(target)`): mutant run → all 5 subtests
  `--- PASS`, `ok` — a missing target passed vacuously (m3-mutant-conditional-guards.txt);
- unconditional assertions (after M3): the same mutant → `go test` rc=1, verbatim
  `factory_handoff_abandon_test.go:139: target worktree …/repo/.claude/worktrees/t1082 does not exist`
  for WT_READY, SWITCH_PENDING_INTERACTIVE, SWITCH_PENDING_HEADLESS (`--- FAIL` ×3), RESERVED
  `--- PASS` (m3-mutant-unconditional.txt). Mutant reverted; fixture back to green.

Also: `mcp_build_identity_test.go` sweep baseline drops the `factory_lane_handoff_recover.go:212`
coordinate (file deleted); `TestAuditLagUsesBinlagSeam` `--- PASS`.

| AC | Status | Command | Actual output |
|----|--------|---------|---------------|
| AC-CFR-016 | PASS | `grep -rnwE '<AC-016 pattern>' internal cmd pkg --include='*.go'`; `ls internal/cli/factory_lane_handoff*.go`; control `git grep -nwE '<same>' 553e224f3 -- internal \| wc -l` | 0 lines rc=1; "no matches found" rc=1; control `98` |
| AC-CFR-017 | PASS | `git diff --exit-code develop...HEAD -- internal/factorymsg internal/homestate ':!*_test.go'` (+ working tree) | rc=0 / rc=0 |
| AC-CFR-018 | PASS | `go test ./internal/factorymsg ./internal/kanban -count=1 -v`; `MOAI_HOME=<scratch> go test ./internal/hook -run 'Factory' -count=1 -v`; cli three-name run; `<bin> factory handoff abandon-lane --help` | `ok` ×2, 579 `--- PASS`; hook `ok … 128.713s`, `--- PASS: TestFactoryLaneHandoffInteractiveStateMachine`, 34 top-level PASS, no `[no tests to run]`; cli 3 × `--- PASS`; help rc=0 |
| AC-CFR-020 | PASS | `go test ./internal/cli -run '^TestFactoryLaneHandoffOperatorAbandon$' -count=1 -v` | `ok`; PASS for RESERVED, WT_READY, SWITCH_PENDING_INTERACTIVE, SWITCH_PENDING_HEADLESS and the pre-existing `no_handoff_on_slot` (5 subtest lines — the AC's "four" counts the four handoff states; the fifth subtest was kept, not removed); mutant above |

Package: `go test ./internal/cli/ -count=1 -timeout 40m` → `ok  github.com/modu-ai/moai-adk/internal/cli  1200.345s`
(under `moai slot` `go-test-heavy`; no RESIDUE line). `go test ./internal/config -count=1` → `ok`.
M1+M2 regression (m3-regress-1..3.txt): every M1/M2 AC test plus the nine AC-CFR-013 names `--- PASS`.
Builds: darwin / linux / `GOOS=windows GOARCH=amd64` rc=0; `go vet ./internal/cli ./internal/config` rc=0 (darwin, windows).
Lint: `golangci-lint run --timeout=5m ./internal/cli/... ./internal/config/...` → `0 issues.`;
`GOOS=windows …` → the same 4 pre-existing issues as M2 (none in an M3 file).
Mutant script kept as evidence: `.moai/reports/t1242/m3-mutant.py.txt`.

### M4 — docs, rules, AGENTS.md, token budget (base `34ac95e90`, M3 commit)

Edits (plan M4.1-M4.3): `AGENTS.md:26` + `AGENTS.md.tmpl:32` "Codex lanes:" → "Codex:"; root
`AGENTS.md` §8 drops "and `-f` lead/agents" from the launch shapes; the rule pair (local + template)
`moai-mcp-tools.md:75` "the Codex lane orchestrator" → "a Codex session" and
`moai-mcp-tools-catalogue.md:103-105` "Codex lane orchestrator" → "a Codex session" / "a Codex
lane's shell" → "a Codex session's shell"; docs-site ko first, then en/ja/zh:
`advanced/codex-dual-harness.md:33` drops the `-f` lead/agents launch shape, `guides/mcp-server.md`
rows 180-182 + line 184 move to the Codex-session wording (ko 세션 / en session / ja セッション /
zh 会话). No Mermaid added. Recorded, not fixed (spec.md §E): `AGENTS.md:26` still says
`moai codex -w` "never creates" a tree while `resolveOrCreateCodexWorktreeDir` creates one.

| AC | Status | Command | Actual output |
|----|--------|---------|---------------|
| AC-CFR-019 (1) | PASS | `grep -nE "codex -f\|codex -k\|Codex lanes?\|lane's shell\|Codex 레인\|레인 셸\|Codex レーン\|レーンのシェル\|Codex 泳道\|泳道 shell" <file set>`; control `git grep -nE '<same>' 553e224f3 -- <file set> \| wc -l` | 0 lines rc=1; control `26` |
| AC-CFR-019 (2) | PASS | `cat AGENTS.md docs-site/content/{ko,en,ja,zh}/advanced/codex-dual-harness.md \| tr '\n' ' ' \| grep -oE '\`-f\`[^\|]{0,6}lead/agent' \| wc -l`; control via `git show 553e224f3:<each>` | `0`; control `5` |
| AC-CFR-019 (3) | PASS | `cmp` local vs template, `moai-mcp-tools.md` and `moai-mcp-tools-catalogue.md` | rc=0, rc=0 |
| AC-CFR-022 (local) | PASS (context) | `go test ./internal/config -run 'TestAlwaysLoadedTokenBudget$' -count=1 -v` on the M4 working tree | `always-loaded surface = 77530 tokens (budget 77600, headroom 70, 16 entries)` vs `budget_base: 77539`; the binding merge vs merge^1 run is later |

Docs verify (hns-oss-docs-verify on the eight touched pages): `hugo --minify --gc` (to a scratch
destination) rc=0, 0 WARN/ERROR lines, sitemap present; URL blacklist 0; Mermaid LR/RL 0; H2+ counts
equal across locales (codex-dual-harness 8/8/8/8, mcp-server 26/26/26/26); emoji in added lines 0.
`make build` rc=0 (catalog.yaml regenerated byte-identically, no diff). `go test ./internal/template/...`
→ `ok` ×3; doc-parity tests in `./internal/cli` (`TestMCPToolCatalogueDocsStayMirrorIdentical`,
`TestMCPToolCatalogueFiguresMatchRegistry`, version-stamp and project-root doc tests) `--- PASS`.
No Go source changed in M4.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: completed   # first close: implemented (2026-09-27T04:08:45Z); this record: implemented -> completed
sync_complete_at: 2026-09-27T05:52:19Z
sync_commit_sha: pending-backfill-sync
changelog_entry_position: "CHANGELOG.md [Unreleased] > ### Added, first entry (inserted before SPEC-POWERSHELL-DENY-PARITY-001)"
b12_self_test_a: "grep -c 'SPEC-CODEX-FACTORY-RETIRE-001' CHANGELOG.md (before write) -> 0 -> PASS"
b12_self_test_b: "grep -oE 'AC-CFR-[0-9]+' acceptance.md | sort -u | wc -l -> 25; CHANGELOG entry states '25 acceptance criteria AC-CFR-001..025' -> PASS (count match)"
b12_self_test_c: "ls .moai/reports/t1242/{m1,m2-*,m3-*,m4-*} + ls .moai/project/codemaps/modules.md -> all exist -> PASS"
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented (2026-09-27) -> completed (this record, 2026-09-27)"
  plan_md: "no status: field (stateless per spec-frontmatter-schema.md Artifact Statelessness)"
  acceptance_md: "no status: field (stateless)"
canary_compliance_check:
  applicable: false
  reason: "This SPEC does not define a forward-looking policy that its own sync tests"
```

**Why `implemented`, not `completed`.** `plan.md` §I dispatches M5 (foreign-file
preservation in the develop worktree, AC-CFR-024) as an in-scope sync-phase
milestone — it is part of this SPEC's own Definition of Done (`acceptance.md`
§D.3), not a follow-up card. M5 has NOT been executed by this sync commit: no
`.moai/reports/t1242/m5-lead-confirm.md` exists, so AC-CFR-024 is unresolved.
Per the Status Transition Ownership Matrix (`spec-frontmatter-schema.md`), the
`implemented → completed` transition is available to manager-docs, but taking
it here would assert a Definition of Done this commit did not satisfy — an
unobserved completion claim (`verification-claim-integrity.md` §1.1 surface 1).
`spec.md` frontmatter `status:` is transitioned `in-progress → implemented`;
`updated:` is refreshed to this sync commit's date. A follow-up sync commit
(after the lead's M5 confirmation lands) closes `implemented → completed`.

**AC-CFR-020 discrepancy (sync-phase note for the auditor).** `progress.md`
§E.2 M3 evidence records `TestFactoryLaneHandoffOperatorAbandon` printing
**5** `--- PASS` subtest lines against `acceptance.md`'s AC text, which states
the assertion applies to "four" handoff states (WT_READY, SWITCH_PENDING_INTERACTIVE,
SWITCH_PENDING_HEADLESS, plus RESERVED exempted from the existence check). The
fifth subtest, `no_handoff_on_slot`, is a pre-existing case the M3 rebuild kept
rather than removed (`progress.md` §E.2 M3 section, "the existing `no_handoff_on_slot`" —
verbatim). This is not scope creep introduced by this SPEC; it predates card
t1242 and is surfaced here only because the AC text's literal "four" does not
match the observed test-output line count.

**Partial-supersession annotation — deferred, not performed.** `plan.md` §I
directs the sync phase to mark `SPEC-FACTORY-MIXED-HOOK-001`,
`SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001`, `SPEC-DUAL-HARNESS-RECOVERY-001`,
`SPEC-CODEX-LOCALMD-001`, and `SPEC-FACTORY-RUN-RETIRE-001` with
`partially_superseded_by: [SPEC-CODEX-FACTORY-RETIRE-001]`. This was NOT
performed: `partially_superseded_by:` is a new frontmatter field on five
**other, already-completed** SPECs' `spec.md` files, and the Status Transition
Ownership Matrix scopes manager-docs's allowed frontmatter edits to
`status:`/`updated:` on the SPEC it is itself closing — it does not grant a
general license to add new frontmatter fields to unrelated completed SPECs.
Recorded here as a **follow-up for manager-spec** rather than performed
directly, consistent with the Forbidden ownership crossings clause's
blocker-report discipline. No card is opened for this follow-up by this
commit; the lead may issue one.

**M5 status — closed as already disposed (equivalent to option (a)); no
disposition step executed by the lane.** `.moai/reports/t1242/m5-lead-confirm.md`
now exists (copied into this tree and the primary checkout) and records the
lead's own measurement, taken from the develop worktree, not this branch:
`git status --short` on develop is empty; `grep -rln 'factoryQueueCodexMessage'
internal/` returns 0 files in develop and in `git grep` on this branch; the
positive control `git log --all --oneline -S 'factoryQueueCodexMessage'`
finds it (commit `e92cdbde5` and plan artifact `b82332ea2`), so the symbol
did exist and the zero-hit result above is a measured absence, not a missed
grep. `git branch -a --contains e92cdbde5` shows the six-file residue
(`mcp_factory_msg.go`, `mcp_factory_msg_test.go`,
`mcp_factory_push_live_test.go`, `factorymsg/store.go` +4,
`reports/factory-cross-host-push-20260924.{md,html}`) preserved on
2026-09-26 17:55 KST by operator decision as WIP commit `e92cdbde5` on
branch `WT-factory-push-wip` (do NOT delete or merge that branch), and
already removed from the develop worktree. `factorymsg/store.go`'s +4 lines
add a `Duplicate bool` field on the envelope, read only by the preserved wake
path — not a change this card's own scope touches.

AC-CFR-024 is judged **PASS**, via this recorded lead confirmation — no
disposition step (removal, merge, or file edit) was executed by the lane in
`WT-codex-factory-retire`, and none was needed: option (a) had already
happened before this sync commit, on the develop worktree the lane never
entered. The confirmation file's own Gaps note: whether any other checkout
holds an uncommitted copy of these six files was not measured beyond the
develop worktree and the primary checkout.

**Codemaps regeneration.** `.moai/project/codemaps/modules.md` lines 71 and 73
(the `codex*` and `factory*`/`handoff*`/`profile*` cluster rows) were edited
in place to drop references to the files this card deleted (`codex_factory.go`,
`codex_kanban.go`, `factory_lane_handoff{,_switch,_bind,_recover}.go`) and to
narrate the deletion with the current file counts (`codex*` 21→20;
`factory*`/`handoff*`/`profile*` 14→10, sub-counted 4+1+5). This is a targeted
edit of the two affected rows, not a full codemaps re-scan.

### Gaps (this sync commit)

- `§E.3 Run-phase Audit-Ready Signal` above is still `_<pending run-phase>_` —
  that section is manager-develop's, not manager-docs's. This is no longer
  gated on M5 (M5 is now closed, see above); it remains open because no
  manager-develop commit has populated it.
- `run_commit_sha` / any run-phase-owned field is not backfilled by this commit.
- AC-CFR-022 (always-loaded budget on the merge tree, comparing `<merge>` vs
  `<merge>^1`) was not re-measured by this sync commit — it is defined against
  the develop merge commit, which does not exist yet on this branch.

### Status decision — remains `implemented`, not `completed`

Per `acceptance.md` §D.3 Definition of Done, `completed` requires: (1) all 25
ACs pass with evidence in §E.2 — AC-CFR-024 now PASSes via the lead
confirmation above, but AC-CFR-022 is unmeasured (gap, above); (2) CI on the
develop push that carries the merge is green — no merge has happened yet;
(3) the sync phase has marked the five partially-superseded SPECs
(`plan.md` §I) — deferred to manager-spec, not yet done (see the
partially-superseded-by note above). M5's closure removes one blocker but
two others (green develop-push CI, the five-SPEC marking) remain outstanding
independent of AC-CFR-022, so `spec.md` `status:` stays `implemented`. The
DoD's own wording defers AC-CFR-022's measurement to the merge window
(it compares `<merge>` against `<merge>^1`, which cannot exist before the
merge) — but that deferral does not, on its own, satisfy the other two DoD
items, so `completed` is not yet reachable on any reading of the contract.

### Final close record (2026-09-27T05:52:19Z) — `implemented → completed`

Both DoD items that kept `status:` at `implemented` are now satisfied.

**1. Green develop CI on the push carrying the merge.** Run `36296844622`,
head `6b523a5d3` — includes merge `6d514f9b7` of this card and the guard-fix
merge `6b523a5d3` (see the CI-fix note below). CI, Graph Freshness, CodeQL,
and lsel-leak-guard all report `success` (**lead-reported evidence** — this
lane did not independently query the GitHub Actions API for this run).

**AC-CFR-022 on the merge tree** — `TestAlwaysLoadedTokenBudget`: `77530 ≤
77539`, measured by the lane on merge tree `a7143c58e` (tree identical to
merge `6d514f9b7`). This is the merge-tree measurement the original AC-CFR-022
gap deferred to the merge window; it now PASSes.

**Post-merge CI fix (recorded for the audit trail).** Develop CI run
`36295706756` was red on `internal/hook` `TestDeferredScanOptOut_CrossPackageTestCallersOptOut`.
Fixed by commit `4a0078651` (`WithSynchronousDeferredScans` added at
`codex_factory_retire_test.go:506`), merged as `6b523a5d3` — the same head
named in item 1 above. This fix is downstream of this card's M1 test file and
is recorded here for traceability; it is not a change to this SPEC's own
milestones.

**2. F2 — `partially_superseded_by:` marking.** Done by commit `1fbac5f5b`
(`Authored-By-Agent: manager-spec`), which added
`partially_superseded_by: [SPEC-CODEX-FACTORY-RETIRE-001]` to the frontmatter
of the five sibling SPECs named in `plan.md` §I: `SPEC-DUAL-HARNESS-RECOVERY-001`,
`SPEC-FACTORY-MIXED-HOOK-001`, `SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001`,
`SPEC-FACTORY-RUN-RETIRE-001`, `SPEC-CODEX-LOCALMD-001`. Frontmatter-only
(non-transition frontmatter correction, `spec-frontmatter-schema.md` § Non-
transition frontmatter corrections); the follow-up recorded in the earlier
sync-close note above is now resolved.

**All three `acceptance.md` §D.3 Definition of Done items are satisfied**:
all 25 ACs pass (AC-CFR-024 via the M5 lead confirmation, AC-CFR-022 via the
merge-tree measurement above); CI on the develop push carrying the merge is
green; the five partially-superseded SPECs are marked. `spec.md` `status:`
transitions `implemented → completed`, `updated:` unchanged (already
2026-09-27). This closes the 3-phase close for card t1242.
