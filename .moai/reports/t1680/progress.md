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
