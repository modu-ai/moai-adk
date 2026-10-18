# t1680 progress record

Card t1680 (Class B, run tn1p4e, lane-27, mode parallelizable). Tree `.moai/worktrees/t1680`, branch `WT-mcp-verb-parity` (renamed in place from the leased `WT-t1663-sync-t1663`), base local main `ae34f26831` (HEAD equals main at lease time, tree clean).

## Cause evidence (read from code in this run, not yet reproduced; the RED tests below reproduce each one)

Baseline: `internal/cli/mcp_factory_card.go` and `internal/cli/factory_card.go` at `ae34f26831`.

| Item | Verb (reference rule) | MCP handler (divergent) |
|---|---|---|
| N1 lane label | `factoryLaneLabelFromEnv(verb)` at `factory_card.go:1697` (next), `:1832` (stage), `:2038` (complete); it refuses legacy labels (`worker-1`) before any work (`:148-164`) | `handleFactoryNext` (`mcp_factory_card.go:132`) and `handleFactoryComplete` (`:235`) read `MOAI_FACTORY_WORKER` directly; `handleFactoryStage` passes `os.Getenv(...)` inline at `:219`; only `handleFactoryRenew` (`:299`) calls the shared check |
| (a) decide root | `factoryDecideCards(ctx, factoryCardRoot(), ...)` at `factory_card.go:3289`, `factoryCardRoot() = factoryRootOf(resolveProjectDir())` (`:84`) | `handleFactoryDecide` uses `resolveToolProjectRoot(req)` (raw `project_root`) at `mcp_factory_card.go:257` and passes it at `:278` |
| (b) complete window root | `factoryCompleteCard(ctx, out, factoryCardRoot(), integrationLockRoot(), ...)` at `factory_card.go:2050` | `handleFactoryComplete` passes `launch` (the caller's raw tree) as the window root at `mcp_factory_card.go:247` |
| N3 / F8 / (c) | shared run resolution `resolveFactoryCardRun` (`factory_card.go:2646`) and the renew reconcile path | not yet examined in this block |

Already aligned by t1663 (do not redo): `handleFactoryRenew` (label check and `factoryRootOf(launch)`), `handleFactoryNext` / `handleFactoryStage` / `handleFactoryComplete` use `factoryRootOf(launch)` for the record root.

Debt definitions (from `.moai/specs/SPEC-FACTORY-LEASE-RENEW-001/progress.md` lines 2811-2813 and 3742-3744): N1 as above; N3 = the other lane verbs that resolve a run through `resolveFactoryCardRun` still refuse when several runs are active and no `--run` is given, only the renewal path was narrowed; F8 = a refused renewal still retires a dead run through reconcile and writes an event.

Scope guard from the leader: touch the MCP handler side only; t1688 (lane-26) changes the `complete` Args in `factory_card.go` and t1839 (lane-17) changes the `next` selection, so any change to shared code in `factory_card.go` needs an explicit reason and an absorb-conflict check at landing. No test skips, no timeout enlargement, `internal/cli` only by `-run` selection (ruling 35e9).

## Run evidence

Run commits: RED `5aefcc989b` (test file only), GREEN `4e6cc2e6ab` (handlers plus one test fix). Test file: `internal/cli/factory_mcp_verb_parity_test.go`. Handler file: `internal/cli/mcp_factory_card.go`. Environment for every `go` run: `TMPDIR=GOTMPDIR=/private/tmp/t1680-tmp`. Output below omits the `WARN config sections directory not found` log lines and the `=== RUN` lines; everything else is verbatim.

### RED (command: `go test -count=1 ./internal/cli -run '^TestMCPVerbParity_' -v`, tree at `5aefcc989b`, exit 1)

```
    factory_mcp_verb_parity_test.go:61: tool next with label "worker-1": err = <nil>, want the verb's refusal "factory next: \"worker-1\" is the legacy lane label; use \"lane-1\" (rejoin with -l)"
    factory_mcp_verb_parity_test.go:61: tool stage with label "worker-1": err = factory_stage: factory stage: factory card not found, want the verb's refusal "factory stage: \"worker-1\" is the legacy lane label; use \"lane-1\" (rejoin with -l)"
    factory_mcp_verb_parity_test.go:61: tool stage with label "": err = factory_stage: factory stage: factory card not found, want the verb's refusal "factory stage: MOAI_FACTORY_WORKER is empty — a lane session carries its lane label there"
    factory_mcp_verb_parity_test.go:61: tool complete with label "worker-1": err = factory_complete: factory complete: factory card not found, want the verb's refusal "factory complete: \"worker-1\" is the legacy lane label; use \"lane-1\" (rejoin with -l)"
--- FAIL: TestMCPVerbParity_LaneLabel (2.79s)
    --- FAIL: TestMCPVerbParity_LaneLabel/next/legacy_spelling (0.82s)
    --- PASS: TestMCPVerbParity_LaneLabel/next/empty (0.15s)
    --- FAIL: TestMCPVerbParity_LaneLabel/stage/legacy_spelling (0.40s)
    --- FAIL: TestMCPVerbParity_LaneLabel/stage/empty (0.42s)
    --- FAIL: TestMCPVerbParity_LaneLabel/complete/legacy_spelling (0.47s)
    --- PASS: TestMCPVerbParity_LaneLabel/complete/empty (0.13s)
    --- PASS: TestMCPVerbParity_LaneLabel/renew/legacy_spelling (0.12s)
    --- PASS: TestMCPVerbParity_LaneLabel/renew/empty (0.11s)
    --- PASS: TestMCPVerbParity_LaneLabel/decide/lane_boundary (0.17s)
--- PASS: TestMCPVerbParity_RecordRoot (6.92s)
    --- PASS: TestMCPVerbParity_RecordRoot/decide (1.81s)
    --- PASS: TestMCPVerbParity_RecordRoot/renew (1.82s)
    --- PASS: TestMCPVerbParity_RecordRoot/stage (3.30s)
    factory_mcp_verb_parity_test.go:246: tool arm: err = "factory_complete: factory complete: refused — no valid re-measure record for the candidate tree 92ffd31a8376 (...)", want the verb's refusal "factory complete: refused — the integration window is held by lane-1 (pid 0) ..."
    factory_mcp_verb_parity_test.go:259: tool arm: err = "factory_complete: factory complete: refused — no valid re-measure record for the candidate tree ad419b299fab (...)" state = "merge-ready", want the verb's clean landing
--- FAIL: TestMCPVerbParity_CompleteWindowRoot (12.51s)
    --- FAIL: TestMCPVerbParity_CompleteWindowRoot/a_window_another_lane_holds_refuses_the_second_lane_naming_the_holder (5.87s)
    --- FAIL: TestMCPVerbParity_CompleteWindowRoot/a_re-measure_record_written_at_the_parent_is_found_and_the_card_lands (6.63s)
    factory_mcp_verb_parity_test.go:293: decide (viaMCP=true): factory_decide: factory decide: 1 of 1 cards refused
    factory_mcp_verb_parity_test.go:302: tool arm: card = merged-local v1, verb arm = pushed v2: the handler read another tree's config or queue
--- FAIL: TestMCPVerbParity_DecideConfigRoot (5.12s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	28.460s
```

(The `(...)` marks a span I shortened in this record; the full text is in the run's output file, which is machine-local scratch.) The record-root rows `decide`, `renew` and `stage` pass on the unfixed tree: `homestate.OpenFactory` already resolves the parent record from a linked tree, so the decide defect shows only in the config and queue reads, which `TestMCPVerbParity_DecideConfigRoot` covers.

### GREEN (same command, tree at `4e6cc2e6ab`, exit 0)

```
--- PASS: TestMCPVerbParity_LaneLabel (0.76s)
    --- PASS: TestMCPVerbParity_LaneLabel/next/legacy_spelling (0.09s)
    --- PASS: TestMCPVerbParity_LaneLabel/next/empty (0.08s)
    --- PASS: TestMCPVerbParity_LaneLabel/stage/legacy_spelling (0.08s)
    --- PASS: TestMCPVerbParity_LaneLabel/stage/empty (0.08s)
    --- PASS: TestMCPVerbParity_LaneLabel/complete/legacy_spelling (0.08s)
    --- PASS: TestMCPVerbParity_LaneLabel/complete/empty (0.08s)
    --- PASS: TestMCPVerbParity_LaneLabel/renew/legacy_spelling (0.09s)
    --- PASS: TestMCPVerbParity_LaneLabel/renew/empty (0.07s)
    --- PASS: TestMCPVerbParity_LaneLabel/decide/lane_boundary (0.11s)
--- PASS: TestMCPVerbParity_RecordRoot (5.01s)
    --- PASS: TestMCPVerbParity_RecordRoot/decide (1.06s)
    --- PASS: TestMCPVerbParity_RecordRoot/renew (1.95s)
    --- PASS: TestMCPVerbParity_RecordRoot/stage (2.00s)
--- PASS: TestMCPVerbParity_CompleteWindowRoot (9.78s)
    --- PASS: TestMCPVerbParity_CompleteWindowRoot/a_window_another_lane_holds_refuses_the_second_lane_naming_the_holder (4.85s)
    --- PASS: TestMCPVerbParity_CompleteWindowRoot/a_re-measure_record_written_at_the_parent_is_found_and_the_card_lands (4.93s)
--- PASS: TestMCPVerbParity_DecideConfigRoot (2.72s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	19.230s
```

An intermediate GREEN run (before the test fix) failed only on `CompleteWindowRoot/a_window_another_lane_holds...`: both arms refuse naming lane-1, but the refusal text carries the window's acquired-at stamp, taken from the wall clock per fixture, so the byte comparison differed (`since 2026-10-18T04:34:03Z` against `...04:34:01Z`). The fix in `4e6cc2e6ab` masks that stamp in the comparison (a test defect, not a handler defect).

Also run on the GREEN tree, all exit 0 with the pass counts from `go test -v` output: `-run '^TestSD_AC'` 30 passes (includes the existing `TestSD_AC010_MCPNextParentCheck`, `AC012`, `AC013`, `AC014`, `AC015`, `AC016`, `AC024`, `AC025`); `-run 'Renew'` 16 passes; `-run 'TestFactoryNextNominate'` 16 passes; `-run 'TestQAS_AC008b'` 1; `-run 'TestQWR_AC008'` 1; `-run 'TestFactoryMsgSend'` 2; `-run 'DecidePush'` 3. `-race` on the new table: 4 top-level passes, 0 `DATA RACE` lines.

### Choices for (b) — the integration-window and re-measure root

`integrationLockRoot()` returns `canonicalTreeSpelling($CLAUDE_PROJECT_DIR)` when the variable is set, else the parent of `git rev-parse --git-common-dir`, else `resolveProjectDir()`. Claude Code sets the variable to the project (parent checkout) root even inside a worktree, and the git fallback yields the parent checkout. Both name the parent checkout. `moai integration remeasure` writes its record under that same root, and the window and the lock readers in `factoryCompleteCard` read it under `lockRoot`. The handler cannot call `integrationLockRoot()` as is: it reads the process environment and the process working directory of a long-lived server, which do not follow a caller's worktree. The handler therefore uses `factoryRootOf(project_root)` (= `factory.ResolveTodoQueueRoot`, the parent checkout of the caller's tree) for both `root` and `lockRoot`. Verified by `CompleteWindowRoot`: with `CLAUDE_PROJECT_DIR` pinned to the parent by the fixture, the verb and the handler (given a linked card tree) see the same window and the same re-measure record. A limit of this choice: where `CLAUDE_PROJECT_DIR` names a directory other than the parent checkout of the verb's launch tree, the verb would use that directory and the handler would not; no such launch exists in this repository's flow, and I did not test it.

### N3 / F8 / (c): investigated, no code change (all three need shared-code edits)

- N3 reproduced with a throwaway probe test (deleted, never committed): two live runs `run-a`, `run-b`; the lane bound to `run-a` and holding `t1` there; no `--run`. Result: `verb renew` and `factory_renew` succeed; `verb stage`, `factory_stage` and `factory_complete` refuse with `no --run given and no single active factory run (AMBIGUOUS_FACTORY: run-a (owner live ...), run-b (owner live ...))`. The refusal comes from `resolveFactoryCardRun` (`factory_card.go:2646`), which `factoryStageCard`, `factoryCompleteCard`, `handleFactoryNext` and `factoryDecideCards` call. A handler-only fix would make the MCP resolve a run the verb refuses, which breaks the parity this card locks, so it needs the shared edit below.
- F8 read from code, not reproduced: `factorymsg.ResolveActiveRun` (`internal/factorymsg/store.go:383`) calls `db.ReconcileActiveRuns` whenever two or more active runs exist, which retires provably dead runs and writes events before `resolveFactoryRenewRunRule` can refuse. It sits in shared code under every no-`--run` caller, so no handler can avoid it. A reproduction needs a dead-owner run fixture (`seedLiveRun` seeds live ones only).
- (c) read from code, not reproduced: the sole-run branch of `resolveFactoryRenewRunRule` (`factory_card.go:2684`) already refuses a roster binding naming another run, so the renew path is covered. The gap is in the other verbs: they pass no roster rule to the lease write (`RosterUnchecked`), so a label migrating onto the only active run is not refused by `stage` or `complete`. Same shared-code edit as N3.

### Spec-change / scope request (for the lane to put to the leader)

Proposed shared edit, for a decision before anything is changed:

1. `internal/cli/factory_card.go`: add `resolveFactoryLaneRun(ctx, root, explicit, lane, cardID)` that returns `(runID, rule, err)` by calling the existing `resolveFactoryRenewRunRule` unchanged; in `factoryStageCard` (one call site, near line 1866) and `factoryCompleteCard` (near line 2082) replace `resolveFactoryCardRun(ctx, root, run)` with it, and pass the returned rule to the lease write (`RenewLeaseUnderRoster`) instead of `RosterUnchecked`. That closes N3 and (c) for stage and complete on both surfaces at once, because the handlers call the same functions. `next` has no card yet, so it would use only the sole-run roster pre-check. `decide` is the operator verb and has no lane, so it stays on `resolveFactoryCardRun` (N3 as written lists decide; I recommend dropping it from N3).
2. F8: in `resolveFactoryRenewRunRule`, read the active set through `db.ActiveRunFacts` (read-only, already used by `readLaneRunFacts`, `factory_run_select.go:321`) first and call `factorymsg.ResolveActiveRun` (the reconciling path) only when the read-only set has exactly one run or the refusal would not already be decided. This touches `factory_card.go` and may need a read-only variant in `internal/factorymsg/store.go`.
3. Files touched: `internal/cli/factory_card.go`, possibly `internal/factorymsg/store.go`, plus tests (the table gains two rows; the existing `Renew` tests must keep passing).
4. Absorb-conflict risk: t1688 (lane-26) edits the `complete` command's Args block (`factory_card.go` around lines 2033-2050), which is adjacent to but separate from the `factoryCompleteCard` body edit at line 2082; t1839 (lane-17) changes the `next` selection (`factory_card.go` around lines 1693-1760 and `factoryNextLeaseOnce`), which my `next` change would not need to touch if the roster pre-check lives in `resolveFactoryLaneRun`. Medium risk of textual conflict in `factory_card.go`; semantic interaction with t1839 is the larger unknown.

### Gaps

- `/verify` (the pre-commit wrapper) was not invoked; the verification listed in the task was run directly instead.
- A compound `go test -run '^(A|B|...)$'` selection was refused by the worktree-session guard ("construct too complex to verify"); I split the selection into seven simpler `-run` patterns. The patterns `Renew` and `TestFactoryNextNominate` are unanchored substrings and matched more tests than the table's list (16 passes each).
- Not run: the full `internal/cli` package (barred), `internal/config`, and any other package. Vet and lint ran on `internal/cli` only; `./...` was built (linux and windows), not tested.
- F8 and (c) are code readings, not reproductions. N3 was reproduced by a probe that is not part of the tree.
- The `CLAUDE_PROJECT_DIR`-differs-from-parent case for (b) was not tested.
- The `@MX:ANCHOR` on `mcpRequiredProjectRoot` says "3 call sites"; it has four since `handleFactoryRenew` (t1663), and my change leaves the count at four. I did not edit the tag.

### Residual-risk

- The `complete` happy-path row runs real `git merge` inside temp worktrees on the fixture; a git version difference in merge output could change the row, though it asserts only the final card state.
- The handler now uses the same dir for record and window; if a deployment sets `CLAUDE_PROJECT_DIR` to a non-parent directory (see Gaps), verb and handler windows would diverge again and the table would not detect it, since the fixture pins the variable to the parent.
- CI on the integration branch owns the repository-wide test verdict; it is PENDING at report time.

## Sync

Scope split (leader 10-18): N3, F8 and (c) -> t1867; this card delivers N1, (a) and (b) on the MCP handler side only. t1867 is Class B, P2, queued behind t1839 and t1688.

What changed in the sync commit: one `### Fixed` entry under `[Unreleased]` in `CHANGELOG.md` (no SPEC, so no B12 AC count; the entry's claims were checked against the handler diff `git diff 6a03acf941 HEAD -- internal/cli/mcp_factory_card.go`, the four test names in `internal/cli/factory_mcp_verb_parity_test.go`, and the Run evidence above; `grep -c 't1680' CHANGELOG.md` was 0 before writing), and this section. No docs changed: the MCP catalogue (`.claude/rules/moai/core/moai-mcp-tools-catalogue.md` and its template mirror) and `docs-site/content/{ko,en,ja,zh}/guides/mcp-server.md` list the tools and the `project_root` requirement only, and none states a lane-label or window-root rule that became inaccurate. No template or docs-site file touched, so no `make build`.

Verification (HEAD a5b312d066, this tree): `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `gofmt -l internal/cli` printed nothing, exit 0; `TMPDIR=GOTMPDIR=/private/tmp/t1680-tmp go test -count=1 -run 'TestMCPVerbParity_' ./internal/cli/` -> `ok  github.com/modu-ai/moai-adk/internal/cli 26.595s`.

Observations (not in scope): the `@MX:ANCHOR` on `mcpRequiredProjectRoot` (`internal/cli/mcp_factory_card.go`) says 3 call sites while it has 4 since `handleFactoryRenew` (t1663); observation only, not changed here.

Gaps: `/verify` wrapper not invoked; the full `internal/cli` suite and every other package test were not run (ruling 35e9, no slot); the CHANGELOG sentence on `factory_decide` ("acts on another tree's config") rests on the RED row `DecideConfigRoot` (card = merged-local v1 vs verb pushed v2), not on a separate reproduction; CI owns the repository-wide verdict.

## Receipt-guard record removed (leader, board 6a81) — 2026-10-18T05:25:48Z

- The repair spawn was refused with `AUDIT_RECEIPT_VIOLATION ... sync-auditor / unknown-spec / verdict line missing` after sync-audit round 1 (FAIL 0.78, verdict file committed at 11e07d2c5f). The leader removed that single record under the standing rule (conditions checked: only file in `rejections/`, `starts/` empty, verdict file on disk with `verdict: FAIL` and receipts). Backup: `.moai/reports/t1680/backup-rejection-20261018T052548Z/` (sha256 717ee6dbaa7577ea, cmp match per the leader). The FAIL stands; completion needs a new PASS.
- Expectation recorded: this card has no SPEC, so an auditor's final line `spec=t1680` may form the same record again until t1861 lands and the binary is reinstalled; if it recurs the lane sends the raw record to the leader and does not edit guard records.

## Sync-audit round 1 repair (F1)

Supersedes the Run-evidence statements that the verb's window root and the handler's "name the same directory" / "both name the parent checkout": that held only when `CLAUDE_PROJECT_DIR` was unset or named the parent. Tree base 5ebcf93a62; RED e14014c273, GREEN a3094207c6.

**Rule (leader ruling, binding):** both surfaces resolve the window and re-measure root to the repository common root, independent of server env/cwd. Helper `integrationCommonRootOf(dir)` (`internal/cli/integration.go`): `git -C dir rev-parse --git-common-dir` under `gitenv.Env()`, relative answers joined onto `dir`, `Clean`ed, parent taken when the base is `.git`, then `canonicalTreeSpelling` (t766, git's own spelling). Where git cannot answer (missing path, not a repo, base not `.git` such as bare/separate-git-dir) it returns `canonicalTreeSpelling(dir)`, i.e. the caller's value verbatim, exactly as before. `integrationLockRoot()` applies it to `CLAUDE_PROJECT_DIR` when set (the unset branch is unchanged); `handleFactoryComplete` applies it to `project_root` for the window/re-measure root and keeps `factoryRootOf(launch)` as the record root. `factory_card.go` diff: 0 lines.

**RED (stub helper with the old behaviour, tests fail on assertions), verbatim:**

```
--- FAIL: TestMCPVerbParity_CompleteWindowRootNonParentEnv (4.67s)
    --- FAIL: .../a_window_another_lane_holds_refuses_the_second_lane_naming_the_holder (2.17s)
    --- FAIL: .../a_re-measure_record_written_at_the_parent_is_found_and_the_card_lands (2.50s)
factory_mcp_verb_parity_test.go:266: verb arm: err = "factory complete: refused — no valid re-measure record for the candidate tree 92ffd31a8376 (...); run moai integration remeasure, then re-acquire --wait ..." want a refusal naming lane-1
--- FAIL: TestIntegrationCommonRootOf_LinkedWorktreeResolvesToThePrimaryCheckout
--- FAIL: TestIntegrationCommonRootOf_SubdirectoryOfALinkedWorktreeResolvesToThePrimaryCheckout
--- FAIL: TestIntegrationCommonRootOf_ReturnsGitsOwnSpellingNotTheCallers
--- FAIL: TestIntegrationLockRoot_LinkedWorktreeVariableResolvesToThePrimaryCheckout
    integrationLockRoot() with CLAUDE_PROJECT_DIR=linked worktree = ".../001/Common-Root-Linked", want the primary checkout ".../001/CommonRootRepo"
--- FAIL: TestIntegrationWindowRootAgreesWithTheFactoryRecordRoot
    linked worktree: window root ".../Common-Root-Linked" != record root ".../CommonRootRepo"
FAIL	github.com/modu-ai/moai-adk/internal/cli	23.075s
```

**GREEN, verbatim** (`-run 'TestIntegrationCommonRootOf_|TestIntegrationLockRoot_Linked|TestIntegrationWindowRootAgrees|TestMCPVerbParity_|TestIntegrationLockRoot(Keeps|Corrects|FallsBack)'`): every listed test `--- PASS` (5 `TestMCPVerbParity_` parents, 7 new helper and root tests, 4 t766 tests), including `TestMCPVerbParity_CompleteWindowRootNonParentEnv` (both subtests), `TestIntegrationCommonRootOf_*` (5), `TestIntegrationLockRoot_LinkedWorktreeVariableResolvesToThePrimaryCheckout`, `TestIntegrationWindowRootAgreesWithTheFactoryRecordRoot`, and the four t766 tests unchanged; `ok  github.com/modu-ai/moai-adk/internal/cli 45.392s`. The bypass repro (window held at the parent by another lane) now refuses the second lane naming `lane-1` on the verb and the MCP handler with CLAUDE_PROJECT_DIR set to the lane's linked worktree; the re-measure record written at the parent is found by both.

**Root equality:** `TestIntegrationWindowRootAgreesWithTheFactoryRecordRoot` asserts window root == `factoryRootOf` record root == the verb's `integrationLockRoot()` for a linked worktree, the primary checkout and a non-repo path; green. They can differ where git records no `.git`-named common dir (bare, separate-git-dir, submodule): there `factoryRootOf` (via `homestate.CanonicalProjectRoot`) resolves to the git directory while the helper returns the caller's path. The helper is the single source for the window root, the record root is untouched per the ruling; not measured by a test, residual.

**Existing tests:** no assertion changed. The 13 `integrationLockRoot` references (`integration_lock_root_case_test.go` t766 x4, `integration_lock_cli_test.go`, `candidate_landing_gate_test.go`, the new table) pass; none pinned the verbatim-variable behaviour for a linked-worktree variable. One test refactor: the closure in `TestMCPVerbParity_CompleteWindowRoot` became `completeWindowRootRows(t, envLinked)` so both tables share it; its assertions are unchanged.

**Caller survey (all read the root as the shared lock/state/config root, none wants the variable's own worktree):** factory_card.go:2050 (record root + window), integration_policy.go:76 (policy file), integration_merge.go:47, integration_candidate.go:111, integration_settings_drift.go:218 (preserved copy root; the scanned `dir` stays separate), integration.go:145 (session side-channel slot, one per project), :301, :427, :620 (acquire/status/release lock record), factory_merge.go:93/:113/:173/:184/:296 (target root, re-measure record, git-flow config, window), local_main_resync.go:94, integration_remeasure.go:31. One caller outside the lead's list: `session_worktree_automerge.go:65` (`autoMergeLockRoot = integrationLockRoot`, reads the git-flow config and the window at the lock root); same shape, no per-tree read. No STOP condition.

**Verification (HEAD a3094207c6, this tree):** `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./internal/cli/` exit 0; `gofmt -l internal/cli` no output; `golangci-lint run --timeout=5m ./internal/cli/` -> `0 issues.`; `go test -race -count=1 -run 'TestMCPVerbParity_CompleteWindowRootNonParentEnv|TestIntegrationCommonRootOf_|TestIntegrationWindowRootAgrees' ./internal/cli/` -> `ok 20.101s`. Regression sweep `-run '^TestIntegration|^TestCandidate|^TestObserveCandidate|^TestComplete|^TestWaitPromotion|Remeasure|SettingsDrift|LocalMain|AutoMerge|^TestSD_AC|TestQAS_AC008b|TestQWR_AC008|DecidePush'` -> 156 top-level `--- PASS`, 0 FAIL, 0 SKIP, `ok 369.019s` (named: `TestSD_AC0xx` x30+, `TestQAS_AC008b_MCPFactoryNextHeld`, `TestQWR_AC008_SurfacesSeeWorktreeReading`, `TestFactoryDecidePushGateClosesAMergedPRCard`, `TestFR_AC018_DecidePushGate`, `TestFactoryDecidePushChainsApprovalVersion`).

**Gaps:** a worktree-guard refusal of one compound python heredoc edit (split into Edit calls) and one compound `go test; grep` line (split); `/verify` wrapper not invoked; the full `internal/cli` suite and other packages not run (ruling 35e9, no slot); `internal/config` TestNoHardcodedContexts (scripts/release.sh) known unrelated baseline, not re-run; the CHANGELOG/codex repro `TestReviewCompleteNonParent` itself was not re-run, my table row is its equivalent; judging build for the cited tool outputs is the Go toolchain building this tree (no installed-moai measurement was used); CI owns the repository-wide verdict (PENDING).

**Residual-risk:** a layout where git has no `.git`-named common dir (bare, separate-git-dir, submodule) still keeps the caller's path as the window root and `factoryRootOf` as the record root, which can differ there; behaviour for such layouts is unchanged from before this card. `CLAUDE_PROJECT_DIR` naming a plain subdirectory of a repository now resolves up to the repository root instead of the subdirectory (a deliberate consequence of the ruling; pinned by `TestIntegrationCommonRootOf_SubdirectoryOfALinkedWorktree...`). RecordRoot rows remain non-discriminating (F2, debt); the `@MX:ANCHOR` call-site count (3 vs 4, F3) is untouched.

## Sync re-audit round 2 PASS-WITH-DEBT — 2026-10-18 (lane-27, tn1p4e)

- Sync audit round 2 (delta `11e07d2c5f..e872022eb3`): PASS-WITH-DEBT 0.90, 0 blocking, required codex pass `rcpt-cf99d57ac958b0cb3612a404` (+ the auditor's second codex pass `rcpt-b3a0d6e93006956a23ec44f9`); judging moai build `ee981a932`. Verdict file `.moai/reports/t1680/sync-audit-iter2.md` (committed `41f65048dd`). Round 1 FAIL (F1) is resolved by the shared `integrationCommonRootOf` helper.
- Lane-observed on this tree at `e872022eb3`: `go build ./...` exit 0, `GOOS=windows GOARCH=amd64 go build ./...` exit 0, and `go test -count=1 -run 'TestIntegrationCommonRootOf_|TestIntegrationLockRoot_Linked|TestIntegrationWindowRootAgrees|TestMCPVerbParity_' ./internal/cli/` -> `ok  	github.com/modu-ai/moai-adk/internal/cli	42.190s` (12 top-level PASS).
- Optional debts D1-D5 (from the verdict, none blocking): D1 handler call to the helper is unpinned on layouts where the two roots agree; D2 bare / separate-git-dir / submodule layouts fall back to the directory itself (pre-existing for the verb, the handler wording says "repository common root" without that qualifier); D3 `integrationSessionID` side-channel read now lands at the primary checkout when `CLAUDE_PROJECT_DIR` names a linked worktree and `CLAUDE_CODE_SESSION_ID` is empty (degraded path); D4 `CLAUDE_PROJECT_DIR` naming a plain subdirectory now resolves to the repository root (accepted); D5 the env-unset branch keeps its own copy of the `.git` rule.
- card-review: `.moai/reports/t1680/card-review.md` (codex_review scope=card, pass, 0 findings).
- Gaps: no full `internal/cli` run (ruling 35e9); other packages not run; CI owns the repository-wide verdict. Out of scope by leader split: N3, F8, (c) -> t1867. Observation only: the `@MX:ANCHOR` call-site count on `mcpRequiredProjectRoot` (3 stated, 4 actual).
- Next: record `sync-audit` and `merge-ready`, take a landing turn from the leader (line end at last report: ... -> t1688 -> t1703), then absorb main and remeasure with `-run` selections.

## Sync-audit iter3 repair (rule C) — 2026-10-18

Finding R3-1 (P1, `.moai/reports/t1680/sync-audit-iter3.md`): `integrationSessionID` read the SessionStart side-channel file at `integrationLockRoot()`, which since the common-root normalization answers the repository common root. With `CLAUDE_PROJECT_DIR` naming a linked worktree and `CLAUDE_CODE_SESSION_ID` unset, the caller adopted the id recorded at the primary checkout and could release another session's window without `--force`. The hook writes the file at `input.ProjectDir` (`internal/hook/session_start.go:483`), so read and write roots had diverged.

**Fix (leader ruling, read location only):** new `integrationSidecarRoot()` in `internal/cli/integration.go`, used only by `integrationSessionID`: `CLAUDE_PROJECT_DIR` as given when set, else the unchanged `integrationLockRoot()` fallback. `integrationLockRoot`, `integrationCommonRootOf`, the MCP handler and `factory_card.go` are untouched.

**Commits:** RED `726259c51` (`test(cli): RED for integrationSessionID ...`), GREEN `ed56cb400f` (`fix(cli): read the session sidecar where the SessionStart hook writes it`). Tests: `internal/cli/integration_session_sidecar_test.go` (4 new tests, real temp repo with a linked worktree, two sidecar slots HOLDER-SESSION at the primary / CALLER-SESSION at the linked tree).

**RED (verbatim, at `726259c51`):**

```
integration_session_sidecar_test.go:75: variable names the linked worktree: integrationSessionID("") = "HOLDER-SESSION", want "CALLER-SESSION" (the slot at .../Common-Root-Linked)
integration_session_sidecar_test.go:75: variable names a subdir of the linked: integrationSessionID("") = "HOLDER-SESSION", want "SUBDIR-SESSION" (the slot at .../Common-Root-Linked/pkg/deep)
--- FAIL: TestIntegrationSessionID_ReadsTheSlotWhereTheHookWritesIt (0.41s)
integration_session_sidecar_test.go:96: release by CALLER-SESSION succeeded on a window held by HOLDER-SESSION without --force: release-integration window released (was HOLDER-SESSION on main)
--- FAIL: TestIntegrationRelease_LinkedWorktreeCallerCannotReleaseAnotherSessionsWindow (0.48s)
--- PASS: TestIntegrationSessionID_ControlCasesKeepTheirPrecedence (0.44s)
--- PASS: TestIntegrationSessionID_PlainProjectReadsWhereTheHookWrites (0.45s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.167s
```

**GREEN (verbatim, at `ed56cb400f`):**

```
--- PASS: TestIntegrationSessionID_ReadsTheSlotWhereTheHookWritesIt (0.22s)
--- PASS: TestIntegrationRelease_LinkedWorktreeCallerCannotReleaseAnotherSessionsWindow (0.33s)
--- PASS: TestIntegrationSessionID_ControlCasesKeepTheirPrecedence (0.24s)   (3 subtests PASS: explicit --session, CLAUDE_CODE_SESSION_ID, variable-unset git fallback)
--- PASS: TestIntegrationSessionID_PlainProjectReadsWhereTheHookWrites (0.30s)
--- PASS: TestIntegrationSessionID_EnvBeatsSidecar (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	2.612s
```

**Caller survey:** `grep CurrentSideChannelFile internal --include=*.go` (non-test): readers are `internal/cli/integration.go:176` (fixed here) and `internal/cli/session.go:282` (`resolveCurrentSessionID`, reads at `resolveProjectDir()` = `CLAUDE_PROJECT_DIR` as given, else cwd, which already equals the hook's write location); the writer is `internal/hook/session_start.go:483`. No other reader uses `integrationLockRoot()`. `integrationSessionID`'s callers (`factory_card.go`, `slot.go` x2, `integration_merge.go`, `integration.go` x2, `factory_merge.go`, `local_main_resync.go`, `session_worktree_automerge.go`) all gain the corrected read; none needed their own change. `integrationLockRoot()` still serves the lock record, config and re-measure paths.

**Verification (this tree, HEAD `ed56cb400f`, `-count=1`, TMPDIR=GOTMPDIR=/private/tmp/t1680-tmp):** `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` no output; `go vet ./internal/cli/` no output; `gofmt -l internal/cli` empty. Regression selections: `TestMCPVerbParity_|TestIntegrationCommonRootOf_|TestIntegrationLockRoot_Linked|TestIntegrationWindowRootAgrees` 12 PASS / 0 FAIL (ok 37.992s); grep-derived lock CLI and session-resolution tests (`TestIntegrationStatus_|TestIntegration_AcquireStatusRelease|TestIntegrationAcquire_|TestIntegrationRelease_|TestResolveCurrentSessionID_|TestResolveArmSessionID_`) 36 PASS / 0 FAIL (ok 7.942s); `^TestSD_AC|DecidePush` 33 PASS / 0 FAIL (ok 149.654s); `-race` of the new tests 5 PASS / 0 FAIL (ok 4.574s).

**Gaps:** `golangci-lint run --timeout=5m ./internal/cli/` reports 1 issue, `internal/cli/paths.go:172:79: ST1008 error should be returned as the last argument (staticcheck)`; `paths.go` is not touched by this card (last change `2a8a66bf59`, from main's absorbed t1827), so I treat it as a main-side baseline and did not alter it (scope discipline); no lint finding in the card's files. Test (b) seeds the holder through `factory.AcquireIntegrationLock` (the primitive `acquire` calls) with pid 0, not through the `acquire` CLI, because an in-process CLI acquire records the owning-session pid and the same process then matches the second release key (t951), which would make the refusal assertion depend on the machine's process ancestry. The full `internal/cli` suite was not run (ruling 35e9); no go-test-heavy slot taken; CHANGELOG not edited (the t1680 entry makes no claim about the session id source, `grep t1680 CHANGELOG.md` read); `/verify` wrapper not invoked; CI owns the repository-wide verdict (PENDING).

**Residual-risk:** with `CLAUDE_PROJECT_DIR` unset the fallback still reads the slot at the git common root (unchanged behavior; from inside a linked worktree that is the primary's slot, pinned by the control subtest), so a runtime that sets neither the session env var nor the project dir can still adopt the primary's last-started session; that is the pre-card behavior. The hook's own `input.ProjectDir` is assumed equal to the `CLAUDE_PROJECT_DIR` the CLI sees (Claude Code sets both); a runtime that gives the hook a different directory than the shell variable would still diverge. The sidecar remains one slot per project directory (last SessionStart wins), as before.
