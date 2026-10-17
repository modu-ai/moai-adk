# SPEC-LANE-JOIN-RUN-SELECT-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-17

Basis for the explicit signal: `moai spec lint` exited 0 on the SPEC directory in both the plain and the `--strict` form (record below). This is the author's own lint, not an audit verdict; no plan audit has run.

- Authoring: spec.md, plan.md, acceptance.md, decision-index.md written by manager-spec (Tier M, card t1853, Class C; the A-part of the split of card t1848). Tree `fb749d372e61a6b10c53091362c76b4412827427`, branch `WT-10-17-class-9`. Not yet audited.
- Counts: 7 requirements (Tier M ceiling 16), 8 acceptance criteria (ceiling 16), 3 milestones (M1-M3).
- SPEC ID self-check, executed as Bash: `ID="SPEC-LANE-JOIN-RUN-SELECT-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` printed `PASS`; `ls .moai/specs | grep -i "LANE-JOIN-RUN\|RUN-SELECT"` printed nothing (grep exit 1), so the ID is unique at this tree.
- RED-now evidence: `acceptance.md` ledger E1-E3 (empty sweeps, exit 0, swept-count FAIL; 3 + 2 + 9 named tests), E4 (three existing guards GREEN). No behavioral RED was observed.
- Decision rows: Q1 `DEFAULT-APPLIED` (implementation-level); none open.
- Lint (command, judging build, tree): a binary built from this tree (`go build -o <scratchpad>/bin/moai ./cmd/moai` run in the card tree at HEAD `fb749d372e61a6b10c53091362c76b4412827427`, exit 0) and invoked by path: `moai spec lint <SPEC dir>` printed `✓ No findings — all SPEC documents are valid`, exit 0; `moai spec lint --strict <SPEC dir>` printed the same line, exit 0. That binary carries no commit stamp (`moai version` prints `v3.1.3 none built unknown` because a bare `go build` omits the Makefile's `LDFLAGS`), so its provenance is the build command at the HEAD above, not a self-reported SHA. The installed `moai` is `v3.2.0-rc.36-t1822c` (commit `ee981a932`); `git merge-base --is-ancestor ee981a932 HEAD` exited 0, so it lags this tree and was not used for the cited result (verification-claim-integrity §2.2).
- Hygiene sweep over the five files: no `origin/*` ref, no duration-style time estimate (`grep -nEi` for the patterns printed nothing). Line numbers cited in the SPEC are coordinates at the tree above and are stated as such in `spec.md` §C and `plan.md` §A.
- Gaps carried forward (full list in `spec.md` §H): no behavioral RED; the E4 guards ran in a shell carrying lane environment variables (TestMain clears ambient factory state); the read-path write-free claim is scoped to the `runs` and `events` tables; the Korean strings are authored but not read by a native reviewer beyond the author; docs-site pages not read in full; no plan-auditor or cross-model audit has run.
- Residual judgment calls: (1) the card says the launchers "pick a run implicitly" while this tree already refuses with two or more survivors (`spec.md` §B states the observed behavior); (2) Q1 default is read-only re-presentation (no reconciler); (3) the one-row table never offers cleanup (`CanCleanup` needs two or more rows); (4) the tree-built lint binary has no commit stamp; (5) AC-LJR-007 is a single test name carrying five scenarios (the N2 scenario set), kept as one criterion so its sweep count is one.
- Inherited history (carried, not reset): the predecessor SPEC-FACTORY-RUN-SELECT-CORE-001 (card t1848) plan audits FAIL 0.81 -> 0.89 -> 0.90; D1-D5 and N1 closed and carried in `spec.md` §F; N2 resolved by REQ-LJR-005 / REQ-LJR-006 and pinned by AC-LJR-007 (plan §C.4). The three reports were read in the sibling tree `.moai/worktrees/t1848/.moai/reports/t1848/`.

## §E.2 Run-phase Evidence

Run-phase environment: card tree `.moai/worktrees/t1853` at base `8d7b16f01`; every test command ran with the eleven factory variables of acceptance.md unset (`unset MOAI_KANBAN_ID MOAI_FACTORY_WORKER MOAI_FACTORY_ROLE MOAI_FACTORY_WORKERS MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_AUTO_DISPATCH MOAI_KANBAN_BACKEND MOAI_KANBAN_SETTINGS_INJECTED && go test ...`; `MOAI_KANBAN_CARD`, `MOAI_KANBAN_ID`, `MOAI_KANBAN_LABEL`, `MOAI_KANBAN_LEAD_NAME`, `MOAI_KANBAN_LEAD_ADDR` were not set in the shell). Go 1.26.8, darwin/arm64. `TMPDIR` is the OS default under `/var/folders`, outside `$HOME` and the repository.

### M1 - facts read path and KST (AC-LJR-001, AC-LJR-002)

RED, behavioral (a stub `ActiveRunFacts` returning no rows, a stub `FormatKST` returning its input; tests written first):

```text
$ go test ./internal/homestate -run '^(TestActiveRunFactsCountsNonTerminalCards|TestActiveRunFactsLastActivityUsesParsedTimes|TestActiveRunFactsIsReadOnly)$' -count=1 -v
--- FAIL: TestActiveRunFactsCountsNonTerminalCards (0.69s)
    factory_run_select_test.go:69: run-a InProgress = 0, want 3 (leased, needs-decision, merged-local)
    factory_run_select_test.go:78: listed 0 runs, want 2
--- FAIL: TestActiveRunFactsLastActivityUsesParsedTimes (0.41s)
    factory_run_select_test.go:115: run-w: not listed   (x5: w h u e n)
--- FAIL: TestActiveRunFactsIsReadOnly (0.36s)
    factory_run_select_test.go:170: run-z was not listed
FAIL	github.com/modu-ai/moai-adk/internal/homestate	1.810s

$ go test ./internal/factorymsg -run '^(TestFormatKST|TestAmbiguousRunsErrorRendersStartInKST)$' -count=1 -v
    factory_run_select_test.go:33: utc-nanos: FormatKST("2026-10-17T09:31:00.123456789Z") = "2026-10-17T09:31:00.123456789Z", want "2026-10-17 18:31:00 KST"
    factory_run_select_test.go:47: refusal "AMBIGUOUS_FACTORY: run-kst (owner live, leader pid 4242, started 2026-10-17T09:31:00.123456789Z)" does not render the start as KST
--- FAIL: TestFormatKST (0.00s)
--- FAIL: TestAmbiguousRunsErrorRendersStartInKST (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/factorymsg	0.344s
```

GREEN:

```text
$ go test ./internal/homestate -run '^(TestActiveRunFacts.*)$' -count=1 -v
--- PASS: TestActiveRunFactsCountsNonTerminalCards (0.47s)
--- PASS: TestActiveRunFactsLastActivityUsesParsedTimes (0.40s)
--- PASS: TestActiveRunFactsIsReadOnly (0.45s)
--- PASS: TestActiveRunFactsSurfacesReadFailures (1.12s)   [3 subtests PASS]
ok  	github.com/modu-ai/moai-adk/internal/homestate	2.790s
$ go test ./internal/factorymsg -run '^(TestFormatKST|TestAmbiguousRunsErrorRendersStartInKST)$' -count=1 -v
--- PASS: TestFormatKST (0.00s)
--- PASS: TestAmbiguousRunsErrorRendersStartInKST (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/factorymsg	0.344s
```

Mutant probes executed against the GREEN tree (each applied to a copy, test run, file restored):

- AC-LJR-001 string-order mutant (`raw > bestRaw` in place of the parsed comparison): `--- FAIL: TestActiveRunFactsLastActivityUsesParsedTimes`, `run-w LastActivity = "2026-10-17T10:00:00.1Z", want "...00.11Z"` and `run-h LastActivity = "2026-10-17T10:00:00Z", want "...00.5Z"`. Killed.
- AC-LJR-001 reconciler mutant (`ReconcileActiveRuns` called before the read): `--- FAIL: TestActiveRunFactsIsReadOnly`, `run-z was not listed` (the reconciler retired the dead zero-card run). Killed.
- AC-LJR-002 host-zone mutant (`at.Local()` instead of the fixed zone, with `time.Local` pinned to UTC): `offset-input: ... = "2026-10-17 09:31:00 KST", want "2026-10-17 18:31:00 KST"`, `day-rollover` and `utc-no-fraction` failing alike. Killed.

Existing suites that share the touched code (unchanged, GREEN): all 12 tests of `internal/homestate/factory_run_retire_test.go` (`ok ... homestate 7.295s`) and all 6 of `internal/factorymsg/factory_run_retire_test.go` (`ok ... factorymsg 9.226s`); `GOOS=windows GOARCH=amd64 go build ./internal/homestate/... ./internal/factorymsg/...` exit 0.


### M2 - chooser seams (AC-LJR-003 render half, AC-LJR-004)

RED, behavioral (stub `buildLaneRows` returning no rows, stub `renderLaneTable` returning "", stub chooser always returning a join of "stub"; tests written first):

```text
$ go test ./internal/cli -run '^(TestLaneJoinChooserRendersFacts|TestLaneJoinChooserRowCellEdgeCases|TestLaneJoinChooserOptionEligibility|TestRunChooserParsesInMemoryStreams|TestRunChooserPromptText|TestRunChooserCatalogueHasNoUnicodeEscapes|TestRunChooserSurfacesWriteFailure)$' -count=1 -v
--- FAIL: TestLaneJoinChooserRendersFacts (0.00s)         factory_run_select_test.go:45: buildLaneRows returned 0 rows, want 2
--- FAIL: TestLaneJoinChooserRowCellEdgeCases (0.00s)     factory_run_select_test.go:89: buildLaneRows returned 0 rows, want 2
--- FAIL: TestLaneJoinChooserOptionEligibility (0.00s)    factory_run_select_test.go:120: a single dead zero-card row = [], want CanCleanup false ...
--- FAIL: TestRunChooserParsesInMemoryStreams (0.00s)     choice = {Kind:1 RunID:stub}, want {Kind:1 RunID:r2}  (and 19 sibling subtests)
--- FAIL: TestRunChooserPromptText (0.00s)
--- PASS: TestRunChooserCatalogueHasNoUnicodeEscapes (0.00s)   [guard, green by design: the catalogue constants live in the stub]
--- FAIL: TestRunChooserSurfacesWriteFailure (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.285s
```

GREEN:

```text
--- PASS: TestLaneJoinChooserRendersFacts (0.00s)
--- PASS: TestLaneJoinChooserRowCellEdgeCases (0.00s)
--- PASS: TestLaneJoinChooserOptionEligibility (0.00s)
--- PASS: TestRunChooserParsesInMemoryStreams (0.00s)      [19 subtests incl. the DEBT-1 entries 99, r99, garbage, 0, -1, r0, r, +2, overflow, two bad entries, invalid-then-EOF]
--- PASS: TestRunChooserPromptText (0.00s)
--- PASS: TestRunChooserCatalogueHasNoUnicodeEscapes (0.00s)
--- PASS: TestRunChooserSurfacesWriteFailure (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.398s	coverage: 4.7% of statements (package); factory_run_select.go functions 80.0%-100.0%
```

DEBT-1 (plan audit) disposed here: `TestRunChooserParsesInMemoryStreams` pins out-of-range and unparseable entries (`99`, `r99`, `garbage`, `0`, `-1`, `r0`, bare `r`, `+2`, a 20-digit number, two consecutive bad entries, a bad entry followed by end of input): each prints the catalogue's invalid-entry line exactly the expected number of times, the prompt repeats, and no choice is returned for the bad entry.

Mutant probes executed against the GREEN tree (applied to a copy, test run, file restored):

- AC-LJR-004 eligibility keyed on classification alone (drop `InProgress == 0`): `--- FAIL: TestLaneJoinChooserOptionEligibility`, `row 3 (r3, dead (종료), 2 cards) CanCleanup = true, want false`. Killed.
- AC-LJR-004 one-row table offering cleanup (drop `len(facts) >= 2`): `--- FAIL: TestLaneJoinChooserOptionEligibility`, `a single dead zero-card row = [{RunID:r4 ... CanCleanup:true}], want CanCleanup false`. Killed.
- AC-LJR-004 default-on-empty (an empty line joins row 1): `--- FAIL: TestRunChooserParsesInMemoryStreams`. Killed.
- AC-LJR-004 over-reading buffered reader (`bufio.NewReader(r).ReadString('\n')`): `source reader left "" unread, want "TYPEAHEAD" (the chooser over-read)`. Killed.
- AC-LJR-003 renderer dropping the count column: `line 1 has 6 cells, want 7: "1 | runa-live | live (생존) | lane-x9 | 2026-10-18 11:11:00 KST | -"`. Killed.

### M3 - gate wiring and the selection loop (AC-LJR-003 gate half, AC-LJR-005, 006, 007, 008)

RED, behavioral (the gate hook-up, the `FACTORY_JOIN_CANCELLED` sentinel, the interactive seam and a stub `chooseLaneRun`/`retireChosenRun` returning a fixed error; tests written first):

```text
$ go test ./internal/cli -run '^(TestLaneJoinTTYAmbiguityInvokesChooserWithFacts|TestLaneJoinChooserJoinAndCancel|TestLaneJoinChooserRetireThenJoin|TestLaneJoinRefusedRetirePresentsSurvivorWithoutJoining|TestLaneJoinNonTTYKeepsGuidedRefusalInKST|TestLaneJoinInteractiveSkipsChooserForExplicitSelectors)$' -count=1 -v
--- FAIL: TestLaneJoinTTYAmbiguityInvokesChooserWithFacts (1.72s)
    factory_run_select_gate_test.go:293: enterFactoryLaneRun: spy entered 0 times, want exactly once
--- FAIL: TestLaneJoinChooserJoinAndCancel (4.04s)
    factory_run_select_gate_test.go:322: join: stub chooseLaneRun
    factory_run_select_gate_test.go:353: err = stub chooseLaneRun, want FACTORY_JOIN_CANCELLED
--- FAIL: TestLaneJoinChooserRetireThenJoin (4.27s)
    factory_run_select_gate_test.go:435: gate: stub chooseLaneRun
    factory_run_select_gate_test.go:499: retireChosenRun(live) = stub retireChosenRun, want it to wrap ErrRunOwnerNotDead
--- FAIL: TestLaneJoinRefusedRetirePresentsSurvivorWithoutJoining (2.16s)
--- PASS: TestLaneJoinNonTTYKeepsGuidedRefusalInKST (1.06s)
--- PASS: TestLaneJoinInteractiveSkipsChooserForExplicitSelectors (2.65s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	17.048s
```

The two AC-LJR-008 tests are GREEN at this RED point and the table above says so rather than hiding it: M1 already landed the KST rendering, and the non-interactive path is unchanged by design, so those two are regression guards at this step (acceptance.md labels the E4 half of AC-LJR-008 a regression guard; the KST and spy half was RED on the pre-implementation tree, ledger E3). Their grip is shown by the mutants below, not by a red arrival.

GREEN:

```text
$ go test ./internal/cli -run '^(TestLaneJoinTTYAmbiguityInvokesChooserWithFacts|TestLaneJoinChooserRendersFacts|TestLaneJoinChooserOptionEligibility|TestRunChooserParsesInMemoryStreams|TestLaneJoinChooserJoinAndCancel|TestLaneJoinChooserRetireThenJoin|TestLaneJoinRefusedRetirePresentsSurvivorWithoutJoining|TestLaneJoinNonTTYKeepsGuidedRefusalInKST|TestLaneJoinInteractiveSkipsChooserForExplicitSelectors)$' -count=1 -v
--- PASS: TestLaneJoinTTYAmbiguityInvokesChooserWithFacts (2.85s)
--- PASS: TestLaneJoinChooserJoinAndCancel (9.15s)
--- PASS: TestLaneJoinChooserRetireThenJoin (16.25s)
--- PASS: TestLaneJoinRefusedRetirePresentsSurvivorWithoutJoining (9.48s)
--- PASS: TestLaneJoinNonTTYKeepsGuidedRefusalInKST (1.47s)
--- PASS: TestLaneJoinInteractiveSkipsChooserForExplicitSelectors (4.89s)
--- PASS: TestLaneJoinChooserRendersFacts (0.00s)
--- PASS: TestLaneJoinChooserOptionEligibility (0.00s)
--- PASS: TestRunChooserParsesInMemoryStreams (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	45.593s
```

Nine of nine named tests swept (nine `--- PASS` lines).

DEBT-2 (plan audit) disposed here: the REQ-LJR-005 clause "when none remain, fail with NO_ACTIVE_FACTORY" is pinned by the subtest `cleanup-leaving-no-run-fails-no-active-factory` of `TestLaneJoinChooserRetireThenJoin` (a callback retires the only other run after the chooser opened, then the operator chooses cleanup-then-join: the gate returns exactly `NO_ACTIVE_FACTORY`, the cleaned run is `retired`, nothing is joined, the variable stays empty). The edge case that a chosen join whose run vanished returns the explicit-join error with no re-presentation is pinned by `TestLaneJoinChooserJoinAndCancel/join-of-a-vanished-run-returns-the-explicit-join-error`.

Mutant probes executed against the GREEN tree (applied to a copy, the targeted test run, the file restored byte-identical, checked with `cmp`):

- AC-LJR-007 N2 mutant (after a cleanup attempt, `enterSelectedFactoryRun(root, "", true, timing)` re-resolves with an empty selector): `--- FAIL: TestLaneJoinRefusedRetirePresentsSurvivorWithoutJoining`, `spy entered 1 times, want 2 (the gate must present again, not join)` and `err = <nil>, want FACTORY_JOIN_CANCELLED` on both the join and cancel subtests. Killed.
- AC-LJR-007 N1 mutant (auto-join the sole survivor even after a refused cleanup, `case len(facts) == 1 && (afterCleanup || true)`): `spy entered 1 times, want 2` and `err = <nil>, want FACTORY_JOIN_CANCELLED`. Killed.
- AC-LJR-007 sub-scenario (e) mutant (the read path runs `ReconcileActiveRuns` first): `--- FAIL: TestLaneJoinRefusedRetirePresentsSurvivorWithoutJoining/the-reread-does-not-reconcile`. Killed.
- AC-LJR-005 mutant (the variable set before the choice): `variable was "rx-live" at the presentation entry, want empty (set before the choice)` and `MOAI_KANBAN_ID = "rx-live", want ""` on the cancel variants. Killed.
- AC-LJR-006 mutant (a chooser-side `UPDATE runs SET status='retired'` instead of `RetireRunIfDead`): `ra-dead has 0 run.retired events, want exactly 1`, `rc-dead has 0 run.retired events, want exactly 1`. Killed.
- AC-LJR-008 mutant (chooser consulted before the interactive seam is checked): `the chooser was consulted on a non-interactive session` and `refusal "FACTORY_JOIN_CANCELLED" does not begin with AMBIGUOUS_FACTORY:`. Killed.
- AC-LJR-008 mutant (the guidance suffix dropped): `refusal "AMBIGUOUS_FACTORY: ambigruna (owner live, ... started 2026-10-17 18:31:00 KST), ..." misses "--factory-run"` (and `--leader`). Killed.
- AC-LJR-008 mutant (a named leader reaches the chooser): `named-leader join: FACTORY_JOIN_CANCELLED` and `err = FACTORY_JOIN_CANCELLED, want the ambiguity refusal to stand`. Killed.
- AC-LJR-003 mutant (raw UTC start passed to the row builder): `row 1 = {... Start:2026-10-18T02:11:00Z ...}, want {... Start:2026-10-18 11:11:00 KST ...}`. Killed.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
