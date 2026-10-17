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

### Run-phase closure - per-AC matrix, whole-change checks, gaps, residual risk

Judging build and tree for every number below: tree `.moai/worktrees/t1853` at HEAD `080caba72` (the last code-bearing commit; the one commit after it changes this file only). The spec lint binary was built from this tree with `go build -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Version=v0.0.0-t1853 -X github.com/modu-ai/moai-adk/pkg/version.Commit=080caba72" -o <scratchpad>/bin/moai ./cmd/moai` and invoked by path; `moai version` prints `v0.0.0-t1853   080caba72`. The installed build `ee981a932` is an ancestor of HEAD (`git merge-base --is-ancestor ee981a932 HEAD` exit 0) and was not used for any cited result.

**Per-AC result (one invocation per package, factory environment unscrubbed per the eleven variables above, swept-count rule: one `--- PASS` line per named test).**

| AC | Command | Observed |
|---|---|---|
| AC-LJR-001 | `go test ./internal/homestate -run '^(TestActiveRunFactsCountsNonTerminalCards\|TestActiveRunFactsLastActivityUsesParsedTimes\|TestActiveRunFactsIsReadOnly)$' -count=1 -v` | 3 `--- PASS` lines, `ok  github.com/modu-ai/moai-adk/internal/homestate 2.922s`, exit 0. PASS |
| AC-LJR-002 | `go test ./internal/factorymsg -run '^(TestFormatKST\|TestAmbiguousRunsErrorRendersStartInKST)$' -count=1 -v` | 2 `--- PASS` lines, `ok  .../internal/factorymsg 0.738s`, exit 0. PASS |
| AC-LJR-003 | the nine-name `internal/cli` command below (`TestLaneJoinTTYAmbiguityInvokesChooserWithFacts`, `TestLaneJoinChooserRendersFacts`) | both `--- PASS`. PASS |
| AC-LJR-004 | same command (`TestLaneJoinChooserOptionEligibility`, `TestRunChooserParsesInMemoryStreams`) | both `--- PASS`. PASS |
| AC-LJR-005 | same command (`TestLaneJoinChooserJoinAndCancel`) | `--- PASS`. PASS |
| AC-LJR-006 | same command (`TestLaneJoinChooserRetireThenJoin`) | `--- PASS`. PASS |
| AC-LJR-007 | same command (`TestLaneJoinRefusedRetirePresentsSurvivorWithoutJoining`) | `--- PASS` (3 subtests). PASS |
| AC-LJR-008 | same command (`TestLaneJoinNonTTYKeepsGuidedRefusalInKST`, `TestLaneJoinInteractiveSkipsChooserForExplicitSelectors`) plus the E4 guards in the existing-suite run below | both `--- PASS`; E4: `--- PASS: TestAmbiguousLeaderJoinSelectsNamedLeaderRun (1.49s)`, `--- PASS: TestAmbiguousRefusalCarriesResolutionGuidance (0.86s)`, `--- PASS: TestExplicitFactoryRunJoinsNamedRunAmongTwoLive (0.82s)`. PASS |

```text
$ go test ./internal/cli -run '^(TestLaneJoinTTYAmbiguityInvokesChooserWithFacts|TestLaneJoinChooserRendersFacts|TestLaneJoinChooserOptionEligibility|TestRunChooserParsesInMemoryStreams|TestLaneJoinChooserJoinAndCancel|TestLaneJoinChooserRetireThenJoin|TestLaneJoinRefusedRetirePresentsSurvivorWithoutJoining|TestLaneJoinNonTTYKeepsGuidedRefusalInKST|TestLaneJoinInteractiveSkipsChooserForExplicitSelectors)$' -count=1 -v
--- PASS: TestLaneJoinTTYAmbiguityInvokesChooserWithFacts (3.22s)
--- PASS: TestLaneJoinChooserJoinAndCancel (10.38s)
--- PASS: TestLaneJoinChooserRetireThenJoin (18.07s)
--- PASS: TestLaneJoinRefusedRetirePresentsSurvivorWithoutJoining (13.34s)
--- PASS: TestLaneJoinNonTTYKeepsGuidedRefusalInKST (2.54s)
--- PASS: TestLaneJoinInteractiveSkipsChooserForExplicitSelectors (6.65s)
--- PASS: TestLaneJoinChooserRendersFacts (0.00s)
--- PASS: TestLaneJoinChooserOptionEligibility (0.00s)
--- PASS: TestRunChooserParsesInMemoryStreams (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	55.392s
```

Nine of nine swept. Eight of eight criteria satisfy the swept-count rule at the final tree. AC-LJR-007 is one test name carrying three subtests (join after the refusal, cancel after the refusal, the re-read does not reconcile); a failure inside it reports as one name plus the subtest line.

**Whole-change checks.**

- Existing PRESERVE suites, one invocation (all test functions of `factory_ambiguous_join_test.go`, `factory_join_discovery_test.go`, `codex_factory_retire_test.go`, `factory_run_adopt_test.go`; 44 names): 44 `--- PASS` lines, `ok  github.com/modu-ai/moai-adk/internal/cli 45.663s`, `exit=0`. Plus all 12 tests of `internal/homestate/factory_run_retire_test.go` (`ok ... homestate 7.295s`) and all 6 of `internal/factorymsg/factory_run_retire_test.go` (`ok ... factorymsg 9.226s`), run at M1.
- `gofmt -l` over the ten touched Go files: no output, exit 0. `go vet ./internal/homestate ./internal/factorymsg ./internal/cli`: exit 0.
- `golangci-lint run --timeout=10m ./internal/homestate/... ./internal/factorymsg/... ./internal/cli/...`: 3 issues, all `errcheck` in `internal/cli/factory_run_adopt.go` (unchecked `fmt.Fprintf` return values), a file this change does not touch (absent from `git diff --name-only 8d7b16f01..HEAD`); baseline, not introduced here. `golangci-lint run --timeout=10m --new-from-rev=8d7b16f01 <same three packages>`: `0 issues.`, exit 0. (An earlier full run had also reported one `ST1008` in a test helper of this change; it was repaired and the later run no longer lists it.)
- `GOOS=windows GOARCH=amd64 go build ./...`: exit 0.
- Coverage of the new files by the named and added tests (`go test -coverprofile`, statements summed from the profile): `internal/cli/factory_run_select.go` 100/111 = 90.1%, `internal/homestate/factory_run_select.go` 28/31 = 90.3%, `internal/factorymsg/factory_run_select.go` 4/4 = 100.0%.
- Call-site inventory (`grep -rn "enterFactoryLaneRun(" internal --include='*.go'` without test files): `glm.go:299`, `cc.go:240`, `codex_launcher.go:959`, `codex_factory.go:103` plus the definition in `factory.go`; the same four callers plan.md A names.
- Disjointness with SPEC-FACTORY-RETIRE-GUARD-001 (card t1852): that tree's changed non-`.moai` files (HEAD `05253e87a`; against its merge-base with `main`, plus its uncommitted tracked changes) are `internal/cli/factory_handoff_recover.go`, `factory_relaunch.go`, `factory_retire_guard_text.go`, `factory_runs_retire_guard_test.go`, `internal/homestate/factory_run_cards.go`, `factory_run_retire_guard_test.go`, `factory_run_retire.go`; `comm -12` against this change's non-`.moai` files printed nothing. `RetireRunIfDead` is called unchanged and no t1852 identifier is defined here.
- No file under `internal/template/templates/` is touched, so no `make build` was required (plan.md F).
- `moai spec lint .moai/specs/SPEC-LANE-JOIN-RUN-SELECT-001` (tree-built binary above): `✓ No findings - all SPEC documents are valid`, exit 0. `moai spec lint --strict` on the same directory: the same line, exit 0. The `OwnershipTransitionUnmeasured` INFO the plan audit saw is absent now that the `draft -> in-progress` commit carries `Authored-By-Agent: manager-develop`.

**Gaps (not observed in this run).**

- No package-wide or repo-wide `go test`, `make build`, or `make test` ran (the single-slot rule); the full `internal/cli`, `internal/homestate` and `internal/factorymsg` suites are CI's and the lane's landing remeasure's. Only the named tests and the PRESERVE suites above ran.
- No real terminal was driven: the default `runSelectionInteractive` (stdin terminal and stderr terminal) and the `cc.go` / `glm.go` call sites are not exercised by any test; the gate function, the codex relaunch door and the text chooser over in-memory streams are. A terminal-supervising codex launcher with piped stdio would fail the interactivity gate and take the refusal path; untested.
- The Korean strings were authored by this run and have not been read by a native reviewer.
- The write-free claim of the read path is scoped to the `runs` and `events` tables, as spec.md H says; the legacy-peer fallback of `ReconcileOptionsFor` (which opens an existing broker database) is exercised only with no broker present in the fixtures.
- Whether a dead leader's name survives in the leader-name registry was not observed; the name column renders `-` for an unmapped pid, which the tests pin.
- The docs-site `advanced/factory-mode.md` pages (en, ko, ja, zh) were not edited: plan.md I assigns that edit to the sync phase, which also reads all four pages first.
- The production factory database and the `tmvgvn` evidence case were not read.
- The mutant probes above are single-edit mutants chosen from the mutants acceptance.md names plus the N2 variants; they are not an exhaustive mutation run.
- A background timeout once moved a mutant-loop command to the background mid-run (no SIGTERM landed; the loop finished and `cmp` confirmed both mutated files byte-identical to their saved copies).

**Residual notes for the orchestrator (not run-phase work).**

1. DEBT-3 (plan audit): `acceptance.md` AC-LJR-003 says the fixture dates contain no digit 7, but the seeded `created_at` `2026-10-18T01:31:00.123456789Z` contains one. Run-phase code could not and did not edit `acceptance.md`; this run's own fixtures sidestep it (live-row start `2026-10-18T02:11:00Z`, run ids carrying neither a `7` nor `x9`; the dead row's stored time keeps its `7` and is asserted by cell equality against `2026-10-18 10:31:00 KST`). Needs a manager-spec re-delegation to reword to "the rendered cells contain no 7 outside the count column".
2. The `draft -> in-progress` edit to `spec.md` (`status:` and `updated:` only) changed the plan-artifact bytes, so the plan audit's `plan_artifact_hash` no longer matches those files; skip-eligibility of the earlier plan verdict does not carry past this commit. This is the transition the run phase owns; it is stated so no one reads the stale hash as current.
3. If this change lands before SPEC-FACTORY-RETIRE-GUARD-001, the chooser's only card check is the read-time count (plan.md G): a card landing between render and cleanup is not stopped, and a dead-owner run with registered worker rows and no non-terminal card is offered cleanup, and the table shows no worker count. After that SPEC lands, `RetireRunIfDead` closes the window and the loop prints its typed refusal as ordinary error text.
4. A cleanup that leaves exactly one run joins it without a second look at that run's facts (plan audit residual risk; REQ-LJR-005 states it).

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-10-18
run_commit_sha: 080caba72
run_status: complete
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 44
l44_pre_commit_fetch: not-run (card tree; no push, no PR)
l44_post_push_fetch: not-applicable (nothing pushed)
new_warnings_or_lints_introduced: 0 (golangci-lint --new-from-rev=8d7b16f01 reported 0 issues; 3 errcheck issues in untouched factory_run_adopt.go are baseline)
cross_platform_build:
  windows_amd64: exit 0 (GOOS=windows GOARCH=amd64 go build ./...)
total_run_phase_files: 12
m1_to_mN_commit_strategy: three milestone commits on WT-10-17-class-9 (M1 6d4db8e31, M2 068535cf1, M3 080caba72) plus one closing commit carrying this file only; not pushed

audit_ready: true

## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-10-18
sync_commit_sha: e73e296b9b5f6dd4bfdbf5c643c19ce696666aed
sync_status: complete
b12_self_test_a: PASS (pre-emission `grep -c 'SPEC-LANE-JOIN-RUN-SELECT-001' CHANGELOG.md` printed `0` before the entry was added; `1` after)
b12_self_test_b: DEVIATION, reported not softened (see "B12 self-test 2" below): the CHANGELOG entry states 8 criteria (AC-LJR-001..008); the reference counter printed 16
b12_self_test_c: PASS (every path the CHANGELOG entry names exists; `ls` exit 0)
changelog_entry_position: top of `## [Unreleased]` / `### Added`, above SPEC-JEV-RESOLVED-MODEL-001
frontmatter_status_transitions: spec.md `in-progress -> implemented -> completed` applied atomically on this single sync commit (final value `completed`); `updated: 2026-10-18` unchanged because it already equals the sync commit date; plan.md and acceptance.md untouched (stateless, bodies unedited)
canary_compliance_check: not applicable (this SPEC defines no forward-looking policy that its own sync tests)

audit_ready: true

### Sync-phase evidence (judging build and tree)

Tree `.moai/worktrees/t1853` at HEAD `2011df0cc` (run-phase complete, before this commit). The spec lint binary was built from this tree: `go build -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Version=v0.0.0-t1853sync -X github.com/modu-ai/moai-adk/pkg/version.Commit=2011df0cc" -o <scratchpad>/bin/moai ./cmd/moai` (exit 0) and invoked by path; `moai version` printed `v0.0.0-t1853sync   2011df0cc`.

- Docs-site, four locales, one edit each: `git diff --stat` lists `docs-site/content/{en,ja,ko,zh}/advanced/factory-mode.md | 2 +-`; only the ambiguity-refusal sentence of line 67 changed (the `--retire` sentence is byte-identical, the first sentence is untouched). The edit describes the interactive chooser, the cleanup-then-join eligibility, KST times, the fail-closed non-interactive path and the explicit-selector skip.
- 4-locale parity (headings): `en h2=9 h3=1 lines=142`, `ko h2=9 h3=1 lines=142`, `ja h2=9 h3=1 lines=142`, `zh h2=9 h3=1 lines=142`.
- URL blacklist grep (`docs.moai-ai.dev|adk.moai.com|adk.moai.kr` over the four pages): no output, exit 1 (no hits). Mermaid `(flowchart|graph) (LR|RL)` grep: no output, exit 1. Body-emoji scan (code fences stripped): `en 0, ko 0, ja 0, zh 0, total 0`. Emphasis-spacing grep: no hits.
- `hugo --destination <scratchpad>/hugo-public --cleanDestinationDir` (hugo v0.160.1+extended) in `docs-site`: exit 0; page table `KO 187 / EN 185 / JA 185 / ZH 185`; `grep -ci 'warn|error'` over the build output printed `0`; `sitemap.xml` exists in the destination.
- Moved-page redirects: no page was moved or renamed, so no `vercel.json` change.
- `moai spec lint .moai/specs/SPEC-LANE-JOIN-RUN-SELECT-001`: `✓ No findings — all SPEC documents are valid`, exit 0. `moai spec lint --strict` on the same directory: the same line, exit 0. (Both ran after the `spec.md` status edit and before this section was written.)

- `mcp__moai__spec_audit` (filter SPEC-LANE-JOIN-RUN-SELECT-001, project_root = this tree): `total_specs 1, modern_era_clean 1, grandfathered 0`, one finding `EraAutoDetected` INFO (`H-4 (§E.2 + §E.4 + sync_commit_sha)`), no drift finding. The MCP server process is the installed build `ee981a932` (an ancestor of HEAD), not a tree-built one.

### B12 self-test 2 (acceptance count) - deviation, stated plainly

`ac_source=.moai/specs/SPEC-LANE-JOIN-RUN-SELECT-001/acceptance.md`, `tier=M`, state `resolved` (`test -s` passed). The reference counter from `.claude/agents/moai/manager-docs.md` printed `16` with stderr `live=16 excluded=0 ambiguous=0`, exit 0 (no ambiguity, so the halt obligation does not fire). The acceptance file declares 8 criteria: the AC matrix rows AC-LJR-001..008, and its own `Counts:` line says 8. The other 8 counted identifiers are the shorthand `AC-001`..`AC-008`, each appearing once, on the `Requirement-to-criterion map:` line (acceptance.md line 68), where they are cross-references to the same eight criteria rather than new ones. The counter cannot tell a shorthand reference from a declaration without the `[REF]` token, and adding `[REF]` after those occurrences is an `acceptance.md` body edit, which this sync phase may not make. The CHANGELOG entry therefore states 8, matching the matrix and `ac_pass_count: 8` in §E.3; it does not state 16. Resolution owner: a manager-spec re-delegation (the same re-delegation that carries the plan audit's DEBT-3 wording fix) places `[REF]` after the eight shorthand occurrences, after which the counter prints 8.

### Residual notes carried forward (not sync-phase work)

- DEBT-3 (plan audit): `acceptance.md` AC-LJR-003's "no digit 7" fixture sentence is unchanged and stays a reported residual; it needs a manager-spec re-delegation.
- `sync_commit_sha` was written as the canonical `pending-backfill` placeholder in the sync commit because a commit cannot cite its own hash; it has been backfilled with the real sync commit SHA `e73e296b9b5f6dd4bfdbf5c643c19ce696666aed` in a following commit (spec-workflow.md Sync (close); spec-frontmatter-schema.md D3 backfill exemption).
- The MX Tag sub-step added no tag: this change touches no `@MX` annotation, and no tag validation tool was run in this sync phase (see the report's Gaps).
