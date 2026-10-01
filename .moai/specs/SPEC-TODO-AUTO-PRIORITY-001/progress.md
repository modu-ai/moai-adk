# SPEC-TODO-AUTO-PRIORITY-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-10-01T18:13:17Z
card: t1400
tier: M
plan_audit:
  iteration_1: { verdict: FAIL, score: 0.75, report: .moai/reports/t1400/plan-audit-iter1.md }
  iteration_2: { verdict: PASS, score: 0.91, threshold: 0.80, report: .moai/reports/t1400/plan-audit-iter2.md }
  audited_sha: 7d8a9bdbce81a6016c884b844fd05e8d49052081
  note: SPEC artifacts were untracked at audit time; audited_sha is the branch HEAD the audit ran against, and the artifact hashes below pin the audited content.
  evidence_locality: verdict files are local evidence under gitignored .moai/reports/
artifact_sha256:
  spec.md: a07ea11a1188d0784222dfc8d286f57dd7697fed91807e6f8e3cb700a862676f
  plan.md: 0d22a51c23f9a3fabd5d7663b5e23d268de627aeebcab4bd355959ececfb17d8
  acceptance.md: 8ced1468807221609f2cdc6190c55e07e324803c3ae01d94594f5cb865c73e8a
```

## §F Phase 4 Mode Selection

### Kickoff gate (plan→run) — autonomous form, `.claude/rules/moai/workflow/auto-semantics.md` §9.1

| Condition (§9.1) | Observed | Evidence |
|---|---|---|
| Independent plan-audit verdict is PASS | PASS, 0.91 against the Tier M threshold 0.80 | `.moai/reports/t1400/plan-audit-iter2.md` (`verdict: PASS` read from the file) |
| Plan phase records audit-ready | `plan_status: audit-ready` | §E.1 above, committed in `38b54f29b` |
| Plan-artifact hashes unchanged since the verdict | equal | `shasum -a 256` re-measured after the plan commit; matches §E.1; artifact mtimes (02:59, 03:03) precede the verdict file (03:11) |
| No blocker open | none | both iteration-1 clarification items were settled by the operator on 2026-10-02 (S-1, S-2 in plan.md); optional audit findings D-N1..D-N7 are carried into the run delegation, not into the SPEC |
| Keep-set case (environment-impossible / operator-held / irreversible external-shared) | none applies | doctrine amendment was operator-approved; no external shared system is touched; nothing is pushed |

```text
decision record: decided_by=claude-code lane-2 orchestrator (factory lane, Kickoff autonomous transition) evidence_refs=.moai/reports/t1400/plan-audit-iter2.md(verdict=PASS score=0.91 audited_sha=7d8a9bdbc),.moai/specs/SPEC-TODO-AUTO-PRIORITY-001/progress.md#E.1,commit 38b54f29b,sha256 spec=a07ea11a plan=0d22a51c acceptance=8ced1468 ladder_path=gate-row plan→run Kickoff (AUTONOMOUS, auto-semantics §9.1)
```

Recorded 2026-10-01T18:14:28Z. The decision board under the moai home (auto-semantics §11) was NOT written: `factory_decide` is refused for a lane session and no lane-writable board verb was found, so this record lives here and in the card's local evidence file. A reader must treat it as self-attested (auto-semantics §10).

### Mode evaluation

Input parameters: tier M; files affected about 12, counting each live file and its template mirror separately (estimated from `plan.md` §files list, not mechanically counted); domains 4 (Go source, rule document, skill workflow document, agent document) plus template mirrors; language mix Go + markdown; concurrency benefit LOW — the milestones are ordered (the stage function in M1 precedes the `runAutoCycle` wiring in M2). Agent Teams: not requested.

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | non-trivial, spans code, tests and doctrine |
| serial | **yes** | coding-heavy, ordered milestones M1-M4, default fallback |
| fanout | no | research is finished; the remaining work is implementation, not multi-domain reading |
| sweep | no | not a single uniform mechanical transform and well below ~30 files |

Decision: serial

Justification: the work is a ranking stage in Go with its tests plus a small set of paired doctrine edits whose wording depends on the code's behaviour; the milestones have ordering dependencies, so a single `manager-develop` per milestone is the safe default. Boundary case: none (files and domains are not at a threshold).

## §E.2 Run-phase Evidence

### M1 — Selection contract: ranking stage, record renderer, fallback keys

Run by manager-develop (cycle_type=tdd), card t1400, branch `WT-auto-priority-pick`. M1 commits: `96e2c71fe` (rank accessor, SPEC `draft → in-progress`, §F record), `110918d14` (ranking stage, renderer, tests). Every measurement below ran with the eleven lane variables scrubbed in one compound `unset … && <command>` invocation, on a tree whose Go content equals `110918d14` (measured before the commits, content unchanged by them; this section is the only later edit). Build identity: the Go toolchain compiled the tree under test directly, no installed `moai` binary was invoked.

**Claim 1 — M1 tests are GREEN and swept (not vacuous).**
- Command: `go test -list 'TestAutoRank' ./internal/cli/`. Output: `TestAutoRankFallbackOrder` `TestAutoRankDemotion` `TestAutoRankUnmeasuredSignal` `TestAutoRankBlockedExcluded` then `ok  	github.com/modu-ai/moai-adk/internal/cli	0.775s` — swept count 4 (the other 11 names of the final 15-name sweep join at M2-M4).
- Command: `go test -count=1 -v -run '^(TestAutoRankFallbackOrder|TestAutoRankDemotion|TestAutoRankUnmeasuredSignal|TestAutoRankBlockedExcluded|TestAutoFallbackReasonMapping)$' ./internal/cli/`. Output: five `--- PASS` lines (top-level), 31 `    --- PASS` subtests, `ok  	github.com/modu-ai/moai-adk/internal/cli	1.382s`. `TestAutoFallbackReasonMapping` is an M1 addition outside the 15-name sweep (named so `-list 'TestAutoRank'` stays at the planned count).
- Command: `go test -count=1 -v -run '^TestPriorityRank$' ./internal/kanban/`. Output: `--- PASS: TestPriorityRank (0.00s)` with 5 subtests, `ok  	github.com/modu-ai/moai-adk/internal/kanban	0.469s`.

**Claim 2 — characterization guard (plan §Findings (f), 10 tests) unchanged GREEN.** Baseline before any edit and after M1, same command: `go test -count=1 -v -run '^(TestTodoAutoPickupSelection|TestAutoPickTargetsRelationBlocked|TestAutoPickTargetsReturnsAfterDone|TestRunAutoCycleSkipsBlockedCards|TestAutoPickTargetsRescueArmUnfiltered|TestAutoPickTargetsNonSequencingRelation|TestTodoAutoSerialCycle|TestTodoAutoJevPoisonedValueCausesNoMutation|TestTodoAutoJevDegradedNonFinding|TestTodoAutoJevScriptPresentSignal)$' ./internal/cli/` — baseline `10 PASS / 0 FAIL`, `ok … 3.325s`; after: 10 `--- PASS`, `ok … 6.298s`. `todo_auto.go` and the two existing test files are byte-unchanged (`git diff --stat` on `internal/cli/todo_auto.go` empty).

**Claim 3 — the regression guard bites (AC-TAP-012 mutant).** With the relation filter temporarily disabled in `autoPickTargets` (`; false && len(blockers) > 0`), `go test -count=1 -v -run '^TestAutoPickTargetsRelationBlocked$' ./internal/cli/` printed `pickup = [t1,t2], want [t2] — t1 must be excluded while t2 stays a candidate` / `pickup = [t1,t2], want [t1] — t2 must be excluded while t1 stays a candidate`, `--- FAIL: TestAutoPickTargetsRelationBlocked`, `FAIL`. The line was restored; `git diff --stat -- internal/cli/todo_auto.go` was empty afterwards.

**Claim 4 — new tests bite (mutant probes on `todo_auto_rank.go`, each restored).** Demotion ignored in the comparator → `--- FAIL: TestAutoRankDemotion`. Hold marker matched anywhere (`strings.Contains`) → `--- FAIL: TestAutoRankUnmeasuredSignal/marker_in_the_middle_of_the_text_does_not_demote` and `…/a_character_before_the_marker_does_not_demote`. Blocked exclusion disabled → three `TestAutoRankBlockedExcluded` subtests FAIL. Per-card unknown note keyed on `kinds != nil` instead of a successful lookup → `todo_auto_rank_test.go:343: record = ["selection: source=fallback" "selection: ranked U C"], want a `selection: note` line naming the unmeasured "landed" signal`, `--- FAIL: TestAutoRankUnmeasuredSignal/lookup_answered_nothing`. No mutant residue: `grep -n "false &&" internal/cli/todo_auto_rank.go internal/cli/todo_auto.go` exit 1.

**E8 — verbatim RED output captured before GREEN.**

`go test -count=1 -v -run '^TestPriorityRank$' ./internal/kanban/` (no accessor yet):
```
# github.com/modu-ai/moai-adk/internal/kanban [github.com/modu-ai/moai-adk/internal/kanban.test]
internal/kanban/classification_test.go:255:14: undefined: PriorityRank
internal/kanban/classification_test.go:260:5: undefined: PriorityRank
internal/kanban/classification_test.go:261:3: undefined: PriorityRank
FAIL	github.com/modu-ai/moai-adk/internal/kanban [build failed]
```
`go test -count=1 -run '^(TestAutoRankFallbackOrder|TestAutoRankDemotion|TestAutoRankUnmeasuredSignal|TestAutoRankBlockedExcluded|TestAutoFallbackReasonMapping)$' ./internal/cli/` (no stage file yet, head):
```
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/todo_auto_rank_test.go:49:53: undefined: autoLandedLookup
internal/cli/todo_auto_rank_test.go:87:10: undefined: autoRankFallback
internal/cli/todo_auto_rank_test.go:91:20: undefined: autoRankSourceFallback
internal/cli/todo_auto_rank_test.go:119:12: undefined: renderAutoSelectionRecord
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```
Same selector against a signature-only skeleton (types and constants, zero-value bodies — replaced by the implementation, never committed), behavioral RED, selected lines:
```
todo_auto_rank_test.go:89: ranked = [], want [B E A D C]
todo_auto_rank_test.go:108: ranked = [], want [Y X Z W] — an absent classification must tie with the default priority, keeping queue order
todo_auto_rank_test.go:121: record = [], want at least a source line and a ranked line
todo_auto_rank_test.go:156: ranked = [], want [C H]
todo_auto_rank_test.go:182: ranked = [], want [C L]
todo_auto_rank_test.go:201: ranked = [], want [C N1 N2]
todo_auto_rank_test.go:308: ranked = [], want [U C] — unknown is not landed
todo_auto_rank_test.go:392: ranked = [], want [A]
todo_auto_rank_test.go:444: excluded = [], want [B1 B2] in queue order
todo_auto_rank_test.go:472: autoFallbackReason("disabled") = "", want "jev-disabled"
--- FAIL: TestAutoRankFallbackOrder (0.00s)
--- FAIL: TestAutoRankDemotion (0.00s)
--- FAIL: TestAutoRankUnmeasuredSignal (0.00s)
--- FAIL: TestAutoRankBlockedExcluded (0.00s)
--- FAIL: TestAutoFallbackReasonMapping (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.260s
```
The skeleton is itself an implementation-before-test shape only in the trivial sense that it holds signatures; its bodies return zero values and the failing run above preceded every real body.

**Claim 5 — builds, vet, lint, format.**
- `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (pre-flight and after).
- `go vet ./internal/cli/ ./internal/kanban/` → exit 0.
- `golangci-lint run --timeout=2m ./internal/cli/ ./internal/kanban/` (v2.1.6) → exit 0, `0 issues.` after fixing one new finding of mine (QF1002 tagged-switch on the first draft); no pre-existing finding in these two packages.
- `gofmt -l` on the four touched Go files → no output.

**Claim 6 — coverage.** `todo_auto_rank.go` under only its own five tests (`go test -count=1 -coverprofile=… -run '^(…five names…)$' ./internal/cli/`, then `go tool cover -func`): every function 100.0% (`poor`, `autoRankHoldMarked`, `autoRankNearDuplicate`, `autoRankPartition`, `autoRankCandidates`, `autoRankOrderFallback`, `autoRankFlags`, `autoRankFallback`, `renderAutoSelectionRecord`, `autoFallbackReason`). `kanban.PriorityRank` 100.0% (`-run '^TestPriorityRank$' ./internal/kanban/`). `go test -count=1 -cover ./internal/kanban/` → `ok … 226.941s coverage: 84.8% of statements` (whole-package figure, includes the pre-existing code; no `internal/cli` package-level figure was taken because that needs the full suite, which CI owns).

**Claim 7 — static boundaries on `todo_auto_rank.go`.** `grep -n 'AskUserQuestion\|mcp__askuser'` exit 1 (no match); `grep -n 'Mutate\|ArchiveCard'` exit 1; `grep -n '"high"\|"normal"\|"low"'` exit 1; `grep -c '^func '` → 10 (positive control: the file exists and is scanned). B2 pre-scan `grep -rn "Retired\|superseded" internal/cli/todo_auto.go internal/kanban/classification.go` exit 1 (no conflict).

**Deviations and interpretations (no scope growth).**
- `TestAutoFallbackReasonMapping` was added (the reason mapping is M1 scope; the planned cycle-level `TestAutoRankFallbackReasons` is M2). Its name stays outside the `TestAutoRank` prefix so the final sweep count remains 15.
- REQ-TAP-006's "one labelled non-finding per excluded card" is rendered as the `selection: excluded <id> (blocked)` record line (AC-TAP-007 asks for that line exactly once); no separate `non-finding:` line is printed by the stage. If the leader wants both, that is an M2 wiring choice.
- The landed-lookup seam is defined by type only (`autoLandedLookup`); its live implementation over `computeTodoPRRows` and the Jev ranker are M2 (plan D-5).
- `kanban.PriorityRank` returns -1 for a value outside the closed set (plan names the accessor only); no validated record can reach that branch.

**Gaps (not observed).** The cycle-level criteria AC-TAP-001, -002, -003, -007, -008, -009, -010 are not flipped at M1 and were not measured. No `internal/cli` package-wide test or coverage run (CI owns it). The doc/mirror guards (`TestTodoSkillDocumentsClassification`, `TestSanitizedPairParity`, `TestJevAmendmentLinkage`) were not run at M1 (no document changed). Raw logs live in the session scratchpad and are machine-local; the deciding lines are quoted above.

**Residual risk.** The stage is only exercised with injected seams until M2; the real `computeTodoPRRows` mapping to `autoLandedLookup` and the Jev score validation are untested here. `TestAutoRankDemotion` pins signals in the fixed order hold-marker, landed, near-duplicate; a later change to that order changes the flagged line text.

### M2 — Wire the stage and the Jev ordering consumer

Run by manager-develop (cycle_type=tdd), card t1400, branch `WT-auto-priority-pick`, starting HEAD `dcd0e44d5`. M2 commits: `719e6c1b4` (stage wiring, Jev consumer, live seams, tests, hermetic entry-point test, consumer-set declaration), `4050769f1` (source-line literal spelled whole). Every Go command ran with the eleven lane variables scrubbed in one compound `unset … && <command>` invocation; the Go toolchain (go1.26.8) compiled the tree under test directly and no installed `moai` binary was invoked. Pre-flight at `dcd0e44d5`: `go build ./...` exit 0, `GOOS=windows GOARCH=amd64 go build ./...` exit 0, baseline of the ten §Findings (f) tests plus `TestTodoAutoEntryPointFlag` plus the five M1 tests = 16 `--- PASS`, 0 FAIL (`ok … 10.055s`). B2 pre-scan `grep -rn "Retired\|superseded" internal/cli/todo_auto.go internal/jev` exit 1 (no conflict).

**Claim 1 — the M2 tests are GREEN and swept.** All figures are from one run on the tree of HEAD `4050769f1` (clean `git status`): `go test -count=1 -v -run '^(…15 rank/live names…|…10 characterization names…|TestTodoAutoEntryPointFlag|TestJevCallPath_HasExactlyTheDeclaredConsumers)$' ./internal/cli/` → 27 `--- PASS`, 0 `--- FAIL`, `ok  	github.com/modu-ai/moai-adk/internal/cli	17.356s`. `go test -list 'TestAutoRank' ./internal/cli/` → 11 names (`TestAutoRankFallbackOrder`, `…Demotion`, `…UnmeasuredSignal`, `…BlockedExcluded`, `…SelectionRecord`, `…JevOrdering`, `…FallbackReasons`, `…JevMalformedAnswer`, `…RescueFirst`, `…QueueUnchanged`, `…NoQueueWriteGuard`) then `ok … 0.788s` — the 4 doc/agent names join at M3-M4 for the planned 15. Three further M2 tests (`TestAutoLiveJevRanker`, `TestAutoLiveLandedLookup`, `TestAutoDefaultSeamsAreInert`) are named outside the `TestAutoRank` prefix so the final sweep stays at 15.

**Claim 2 — AC matrix (this run, this tree, HEAD `4050769f1`; each row's test is among the 27 PASS lines above).**

| AC | Test (command: `go test -count=1 -v -run '^<name>$' ./internal/cli/`) | Observed | Status |
|---|---|---|---|
| AC-TAP-001 | `TestAutoRankSelectionRecord` | `--- PASS: TestAutoRankSelectionRecord (0.57s)` (record before first accept, ranked order = accept order, D-6 order jev line < notes < record < accept, no record and no Jev call when no queued candidate) ; E1 flipped: `grep -rl --exclude="*_test.go" "selection: source=" internal/cli` → stdout `internal/cli/todo_auto_rank.go`, exit 0 (E1c control `…"jev: unavailable"…` → `internal/cli/todo_auto.go`, exit 0) | PASS |
| AC-TAP-002 | `TestAutoRankJevOrdering` | `--- PASS: TestAutoRankJevOrdering (0.86s)` (jev order ≠ fallback order and wins; score tie → confidence → fallback order; one request, one score question per candidate, blocked left out; surplus beyond the bound of 40 follows in fallback order with a `selection: note`) | PASS |
| AC-TAP-003 | `TestAutoRankFallbackReasons` | `--- PASS: TestAutoRankFallbackReasons (1.90s)` (11 subtests: no seam + the nine `jev.Availability` values + available-but-incomplete; exactly one source line and one `reason=`; cycle returns nil) | PASS |
| AC-TAP-007 | `TestAutoRankBlockedExcluded` | `--- PASS: TestAutoRankBlockedExcluded (0.43s)` (M1 stage subtests + cycle subtests on both sources + all-blocked ends on `no eligible card`; Jev not called) | PASS |
| AC-TAP-008 | `TestAutoRankRescueFirst` | `--- PASS: TestAutoRankRescueFirst (0.38s)` (rescue targets first on both sources, never in the Jev request) | PASS |
| AC-TAP-009 | `TestAutoRankQueueUnchanged`, `TestAutoRankNoQueueWriteGuard` | both `--- PASS` (0.39s, 0.00s); `grep -n 'Mutate\|ArchiveCard' internal/cli/todo_auto_rank.go` exit 1 (no output); `grep -c '^func ' internal/cli/todo_auto_rank.go` → `18`, exit 0 | PASS |
| AC-TAP-010 | `TestAutoRankJevMalformedAnswer` | `--- PASS: TestAutoRankJevMalformedAnswer (4.27s)` (11 defect subtests all → `source=fallback reason=jev-incomplete-answer`, fallback order, queue snapshot equal; positive controls 0, 4 and 2.5 → `source=jev`) | PASS |
| AC-TAP-012 | the ten §Findings (f) tests, unchanged files `todo_auto_test.go`/`todo_relation_filter_test.go` modified only as stated in D-N1 below | ten `--- PASS`, no `--- FAIL`, in the 27-PASS run above | PASS |

**Claim 3 — E2/E5 builds, vet, lint.** At HEAD `4050769f1`: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./internal/cli/ ./internal/kanban/` exit 0; `golangci-lint run --timeout=2m ./internal/cli/ ./internal/kanban/` (v2.1.6) → `0 issues.` after fixing one NEW finding of mine in a test (QF1001 De Morgan at `todo_auto_rank_test.go:830`); no pre-existing finding in these two packages. `gofmt -l internal/cli/ internal/kanban/` lists only `internal/cli/mcp_claude.go`, a file this milestone does not touch (pre-existing baseline). Other guards run and green at HEAD: `TestJevCallPath_UnreachableFromDecisionSurfaces`, `TestJevBoundaryScanPositiveControl`, `TestTodoSelectionPredicatesPositivelyEnumerateStates`, `TestTodoSweepSelectorMatchesFamily`, `TestTodoCmd_NoAskUserQuestion`, `TestPriorityRank` (`internal/kanban`), `TestPackageImports_AreStandardLibraryOnly` (`internal/jev`).

**Claim 4 — E3 coverage.** `go test -count=1 -coverprofile=… -run '^(…the rank, live and entry tests…)$' ./internal/cli/` then `go tool cover -func`: every function of `todo_auto_rank.go` 100.0% (`autoRank`, `autoRankTargets`, `autoRankJevOrder`, `autoRankJevRequest`, `autoRankJevAnswers`, `autoRankResultOf`, `liveAutoJevRanker`, `liveAutoLandedLookup`, plus the M1 functions); `runAutoCycle` 81.7% under that selection (the uncovered statements are pre-existing evidence/done paths other tests own). The whole-package figure was not taken (needs the full suite; CI owns it).

**Claim 5 — E4 static boundaries on `todo_auto_rank.go`.** `grep -n 'AskUserQuestion\|mcp__askuser'` exit 1; `grep -n 'Mutate\|ArchiveCard'` exit 1; `grep -n '"high"\|"normal"\|"low"'` exit 1; `grep -c '^func '` → 18 (positive control: the file exists and is scanned). `internal/jev` imports untouched (stdlib-only guard green). No `scripts/jev` reference anywhere in the new code.

**E9 / D-N1 — the existing entry-point test no longer reaches `gh`.** Hazard first, observed: with the live seams wired and `TestTodoAutoEntryPointFlag` still unmodified, `PATH=<fake gh dir>:$PATH go test -count=1 -v -run '^TestTodoAutoEntryPointFlag$' ./internal/cli/` printed `--- PASS` and the fake `gh` (a script that appends its argv to a marker file) had been invoked: marker content `gh pr list --state open --limit 100 --json number,title,body,state`. Change: `todo.go` gains package variables `todoAutoLandedLookup` / `todoAutoJevRanker` (defaults = the live seams) which the `--auto` flag path hands to `autoOptions`; `TestTodoAutoEntryPointFlag` replaces both with counting stubs, asserts `landed=1 jev=1` (so the production wiring really reaches both) and asserts the `selection: source=fallback reason=jev-disabled` record; its original two assertions and the `store.Add` line are unchanged. After the change, same command with the marker file removed first: `--- PASS: TestTodoAutoEntryPointFlag (5.51s)` and `ls <marker>` → `No such file or directory`, exit 1 (marker absent). No other existing test reaches the live wiring: the only `--auto` through the `todo` cobra path is `todo_auto_test.go:550` (the other `"--auto"` hits are the unrelated goal-mission command).

**E8 — verbatim RED output captured BEFORE GREEN.** Build-RED (tests written, nothing implemented):
```
# github.com/modu-ai/moai-adk/internal/cli [github.com/modu-ai/moai-adk/internal/cli.test]
internal/cli/todo_auto_rank_test.go:504:8: opts.jevRank undefined (type autoOptions has no field or method jevRank)
internal/cli/todo_auto_rank_test.go:942:24: undefined: autoRankJevLevelCount
internal/cli/todo_auto_rank_test.go:960:12: undefined: autoRankJevCandidateLimit
internal/cli/todo_auto_rank_test.go:969:10: undefined: autoRank
FAIL	github.com/modu-ai/moai-adk/internal/cli [build failed]
```
Behavioral RED against a signature-only skeleton (the two `autoOptions` fields, `autoJevRanker`, the constants, and `autoRank`/`liveAutoJevRanker`/`liveAutoLandedLookup` with zero-value bodies; `runAutoCycle` unwired; the skeleton was deleted before the real bodies were written and never committed). The first behavioral run panicked on my test helper (`index out of range [-1]` from indexing a missing source line), so the helper `rankSourceLine` was added and the run repeated; selected lines of that second run:
```
todo_auto_rank_test.go:491: accept order = [t1 t2], want [t2]
todo_auto_rank_test.go:493: `selection: excluded t1 (blocked)` printed 0 times, want exactly once
todo_auto_rank_test.go:802: source line at -1, first accept at 1 — the record must come first
todo_auto_rank_test.go:908: accept order = [t1 t2 t3 t4], want [t2 t4 t3 t1]
todo_auto_rank_test.go:943: calls = 0, want 1
todo_auto_rank_test.go:1053: 0 source lines, want exactly 1
todo_auto_rank_test.go:1204: accept order equals the stored order [t1 t2 t3 t4]; the fixture cannot tell a reorder from none
todo_auto_rank_test.go:1313: availability="" calls=0, want disabled with no transport call
todo_auto_rank_test.go:1390: kinds = map[], want t1 landed and t2 no-link
--- FAIL: TestAutoRankBlockedExcluded / SelectionRecord / JevOrdering / FallbackReasons / JevMalformedAnswer / RescueFirst / QueueUnchanged / LiveJevRanker / LiveLandedLookup / DefaultSeamsAreInert
```
(`TestAutoRankNoQueueWriteGuard` was already green at that point — the ranking file existed since M1 and carries no queue verb; its bite is shown by mutant 8 below.) E1 itself was RED-now at the start of M2's GREEN step and found a real gap: `grep -rl --exclude="*_test.go" "selection: source=" internal/cli` printed nothing, exit 1, because M1 built the source line from `autoRankLinePrefix + "source="`; commit `4050769f1` spells the constant whole, after which E1 reads `internal/cli/todo_auto_rank.go`, exit 0.

**E10 — mutant probes (each applied to the committed tree `719e6c1b4`, run against its target test, then restored; `git diff --stat` empty after every restore).** The `4050769f1` delta is the spelling of one constant, no behaviour.
- score ordering ignored (`return a.score > b.score` → `return false`) → `TestAutoRankJevOrdering` FAIL (`accept order = [t1 t2 t3 t4], want [t3 t1 t4 t2]`); ascending (`<`) → FAIL (`[t2 t4 t1 t3]`); confidence tie-break removed → FAIL (`want [t2 t4 t3 t1]`).
- Jev ranker never consulted (`autoRankJevOrder(fallback, nil)`) → `TestAutoRankJevOrdering` FAIL (`source line = "selection: source=fallback reason=jev-disabled", want "selection: source=jev"`, `calls = 0`).
- blocked exclusion skipped when a ranker is present → `TestAutoRankBlockedExcluded` FAIL (`accept order = [t1 t2], want [t2]`; `request ids = [t1 t2] … a blocked card is never sent to Jev`).
- rescue arm ranked with the queued suffix → `TestAutoRankRescueFirst` FAIL (`accept order = [t4 t3 t5 t1 t2], want [t1 t2 t4 t3 t5]`; `request ids = [t1 t2 t3 t4 t5]`).
- queue write added in `runAutoCycle` (items reversed through `store.Mutate`) → `TestAutoRankQueueUnchanged` FAIL (`queue changed:` ×2); a queue verb named in the ranking source (`ArchiveCard`) → `TestAutoRankNoQueueWriteGuard` FAIL (`carries forbidden tokens [ArchiveCard]`).
- partial application (unanswered-card check disabled) → `TestAutoRankJevMalformedAnswer` FAIL (`source line = "selection: source=jev"`); score range lower bound 1 → FAIL (`want "selection: source=jev"` on the boundary 0); upper bound 5 → FAIL (the score-5 subtest).
- fallback reason omitted (`res.Reason` not set) → `TestAutoRankFallbackReasons` FAIL (`reason printed 0 times, want once`).
- production wiring omits the Jev seam → `TestTodoAutoEntryPointFlag` FAIL (`landed=1 jev=0, want 1 each`); omits the landed seam → FAIL (`landed=0 jev=1`).
- demotion applied on the Jev source → `TestAutoRankJevOrdering` FAIL (`accept order = [t2 t1], want [t1 t2] — demotion applies on the fallback source only`).
- nil landed seam defaulted to the live lookup → `TestAutoDefaultSeamsAreInert` FAIL (`nil seams ran 2 subprocess calls, want none`).

**Deviations and interpretations (no scope growth beyond the M2 files plus one guard declaration).**
- REQ-TAP-006's "one labelled non-finding per excluded card" is satisfied by exactly one `selection: excluded <id> (blocked)` record line per excluded card, as AC-TAP-007 pins (orchestrator decision); no separate `non-finding:` line is printed.
- When no eligible candidate remains (every queued candidate blocked), no request is sent (the stage never sends an empty question list) and the record reads `source=fallback reason=jev-disabled` plus a `selection: note` saying Jev was not consulted. The closed reason vocabulary has no "not consulted" value; `jev-disabled` is the conservative pick and the note carries the truth. A flagged interpretation for the leader.
- REQ-TAP-011 names three defects; the stage also rejects, with the same `jev-incomplete-answer`, an answer that is not a score kind, a card answered twice, and a non-finite confidence. All are "answers defective" under REQ-TAP-003, and the kind check closes the in-range-zero decode R-7 describes.
- The request carries the first `autoRankJevCandidateLimit` (40) candidates in fallback order; the surplus follows the Jev-ordered cards in fallback order with a note naming the bound.
- `TestJevCallPath_HasExactlyTheDeclaredConsumers` (in `doctor_jev_test.go`, not in the plan's file list) was RED at the start of M2 on `dcd0e44d5`: `internal/cli files outside the declared consumer set import internal/jev: [todo_auto_rank.go]` — the M1 import of `internal/jev`, which the M1 verification did not catch because the guard is not among the ten §Findings (f) tests. M2 declares `todo_auto_rank.go` in that guard's allow-list (the guard's own comment names that as the way a consumer arrives).
- Definition of Done item 3 of `acceptance.md` asks that the two existing test files be unmodified; the orchestrator's D-N1 instruction asks for the hermetic change to `TestTodoAutoEntryPointFlag`, which lives in `todo_auto_test.go`. The two cannot both hold: `todo_auto_test.go` therefore differs from its pre-M2 state in `TestTodoAutoEntryPointFlag` and one added import (`internal/jev`) only; `todo_relation_filter_test.go` is untouched. A leader decision is needed on which statement governs at sync.
- The live consumer lives in `todo_auto_rank.go` (no new file); a package variable `autoRankJevDoer` is the test seam for the transport. `todo.go` gains the two seam variables and the two `autoOptions` fields, `todo_auto.go` one call and the two `autoOptions` fields; the "in queue order" strings in `todo.go` are left for M3.

**Gaps (not observed).** No live Jev request was made (credential-gated; the wire shape was exercised only against a fake transport) and the score level-index base (plan A-6) remains unobserved; the live `gh` path ran only through stubs and the fake-`gh` proof. No `internal/cli` package-wide test or coverage run and no CI run (CI owns the repository-wide verdict, PENDING at report time). The doc/mirror criteria AC-TAP-011, -013, -014, -015 and `TestSanitizedPairParity` were not run (no document changed in M2). The live landed seam discards the `computeTodoPRRows` degradation notes (`io.Discard`); the record's `selection: note` names the unmeasured signal but not the `gh` failure text. Mutant probes ran on `719e6c1b4`, not on `4050769f1`. Raw logs live in the session scratchpad and are machine-local; the deciding lines are quoted above.

**Residual risk.** The Jev question wording and the five level texts are unmeasured against the live capability: a score question whose levels read differently to the model than intended orders cards by an uncalibrated signal, and this SPEC claims no accuracy (S-2). `computeTodoPRRows` runs one local git query per card in the record, so `--auto` on a very large queue pays that cost once per invocation. The record prints on every invocation with a non-empty queued set, which adds `selection:` lines before the first `accept` for every consumer that parses `--auto` output.

### M3 — Doctrine amendment and marker disclosure

Run by manager-develop (cycle_type=tdd), card t1400, branch `WT-auto-priority-pick`, starting HEAD `946896945`. M3 commits: `5ddbd6c44` (the two doctrine documents, live and template, plus the `catalog.yaml` hash `make build` regenerated), `75ade0ccc` (flag help and refusal wording, `todo_auto_doc_test.go`). Every Go command ran with the eleven lane variables scrubbed in one compound `unset … && <command>` invocation; the Go toolchain compiled the tree under test directly, no installed `moai` binary was invoked. Measurements below were taken on the tree whose Go and document content equals `75ade0ccc` (this section is the only later edit).

**Pre-flight (HEAD `946896945`, clean tree).** `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; baseline of the 27 M1/M2 rank tests plus the ten characterization tests plus the live-seam tests = 27 `--- PASS`, 0 FAIL (`ok … 19.359s`).

**Guards that read the edited files (found by `grep -rn --include='*_test.go' 'kanban-dispatch\.md\|workflows/gtd\.md\|manager-todo\.md' internal` plus a sweep for budget and neutrality guards; source lines).** `internal/template/gtd_canonical_surface_test.go:13` (`TestGTDCanonicalSurfaceGolden`), `internal/template/contract_mode_guided_test.go:679` (`TestContractModeEmitterSites`, classifies kanban-dispatch.md), `internal/template/workflow_rule_paths_pinned_test.go:31` (`TestWorkflowRulePathsPinned`), `internal/template/backlog_json_disclosure_mirror_test.go:25` (`TestBacklogJSONDisclosure_*`), `internal/cli/doc_json_shape_test.go:98` (`TestTodoListJSONShapeMatchesDoc`), `internal/cli/init_headroom_export_test.go:43` (`TestHeadroomInitSurfaceExport`), `internal/cli/todo_classify_doc_parity_test.go` (`TestTodoSkillDocumentsClassification`), `todo_hold_doc_test.go`, `todo_skill_doc_test.go`, `todo_skill_doc_parity_test.go`, `todo_landed_doc_test.go` (`TestTodoDoctrine_MirrorParityAndStatedColumnCount`); template-tree walkers `TestTemplateNoInternalContentLeak`, `TestRuleTemplateMirrorDrift`, `TestDeclaredRuleMirrorForks`, `TestSanitizedPairParity`; the always-loaded budget guards `internal/config/token_budget_guard_test.go` (`TestAlwaysLoadedTokenBudget`, `TestCodexContractByteCeiling`, `TestCodexNestedTemplateDiscoveryBudget`); `internal/contract/kickoff/activation_test.go:208` (`TestJevAmendmentLinkage`); the catalog hash guards `internal/template/catalog_tier_audit_test.go` (`TestManifestHashFormat`, `TestCatalogHashCoversSkillSubfiles`, `TestCatalogManifestPresent`, `TestAllSkillsInCatalog`) and `embed_catalog_test.go` (`TestLoadEmbeddedCatalog_Success`). `internal/mission/governor_test.go:27,57` reads `manager-todo.md`, which M3 does not touch (M4).

| Guard | Before (HEAD `946896945`) | After (tree of `75ade0ccc`) |
|---|---|---|
| cli doc guards (8 names above) | 7 PASS, 1 SKIP (`TestHeadroomInitSurfaceExport`), `ok 2.633s` | 7 PASS, 1 SKIP, `ok 2.540s`; final re-run 12 names incl. the 4 new tests all PASS, `ok 8.001s` |
| `TestGTDCanonicalSurfaceGolden`, `TestSanitizedPairParity`, `TestRuleTemplateMirrorDrift`, `TestDeclaredRuleMirrorForks`, `TestWorkflowRulePathsPinned`, `TestBacklogJSONDisclosure_*` | PASS | PASS |
| `TestContractModeEmitterSites` | SKIP | SKIP |
| `TestTemplateNoInternalContentLeak` | FAIL — one match, `templates/.claude/rules/moai/workflow/worktree-integration-ops.md` class C1 `SPEC-SESSION-ANCHOR-ATTR-001` | FAIL — the same single match; PRE-EXISTING, not mine, not fixed |
| `TestAlwaysLoadedTokenBudget` | PASS, 64114 tokens (budget 77600) | PASS, 64227 tokens (+113) |
| `TestCodexContractByteCeiling`, `TestCodexNestedTemplateDiscoveryBudget` | PASS | PASS |
| `TestJevAmendmentLinkage` | PASS | PASS |
| catalog hash guards (5 names) | not run before (no template edit yet) | PASS |

**AC matrix (this run, this tree, HEAD `75ade0ccc` content).**

| AC | Command | Observed | Status |
|---|---|---|---|
| AC-TAP-011 | `grep -c "auto-scoped ranking exception" <the four files>` and the same for `"selection order only"` | every file `:1`, exit 0 (both literals, all four files); `go test … -run '^TestAutoRankDoctrineAmendment$' ./internal/cli/` → `--- PASS: TestAutoRankDoctrineAmendment` | PASS |
| AC-TAP-013 | `go test … -run '^(TestAutoRankMirrorParity|TestTodoSkillDocumentsClassification)$' ./internal/cli/`; `… -run '^TestSanitizedPairParity$' ./internal/template/` | both `--- PASS`; `TestSanitizedPairParity` `--- PASS` | PASS |
| AC-TAP-014 | `go test … -run '^TestAutoRankMarkerDisclosure$' ./internal/cli/` | `--- PASS: TestAutoRankMarkerDisclosure` (12 subtests: 3 surfaces × 4 clauses) | PASS |

**E4 greps.** The neutrality expression (SPEC id, requirement token, ISO date, 9+ hex run) over the lines added to the two template files by `git diff 946896945 75ade0ccc` → 0 matches. `grep -n 'AskUserQuestion\|mcp__askuser' internal/cli/todo.go` (comment lines excluded) → exit 1, no match; the new test file has 0. The added template lines contain the word `AskUserQuestion` once, inside the pre-existing `kanban-dispatch.md` promotion paragraph that the edit re-emits as one line — prose, not code.

**E8 — verbatim RED output captured BEFORE the documents were edited** (new test file written first, run on the unamended documents and the unamended `todo.go`; deciding lines):
```
todo_auto_doc_test.go:99: live kanban-dispatch.md does not carry the literal "auto-scoped ranking exception" on a single line
todo_auto_doc_test.go:105: live gtd.md: no single paragraph carries both "auto-scoped ranking exception" and "selection order only"
--- FAIL: TestAutoRankDoctrineAmendment (0.00s)
todo_auto_doc_test.go:212: live passage carries no "auto-scoped ranking exception"
--- FAIL: TestAutoRankMirrorParity (0.00s)
todo_auto_doc_test.go:268: live gtd.md does not state that only a card beginning with the marker is demoted (want the phrase "only a card whose text begins with the [보류 marker is demoted")
todo_auto_doc_test.go:268: --auto flag help does not state that the structural hold is moai todo hold (want the phrase "the structural hold is moai todo hold")
--- FAIL: TestAutoRankMarkerDisclosure (0.00s)
todo_auto_doc_test.go:286: the --auto flag help still asserts the pick order: "process the queue serially: pick one card, … batch approval of the queue in queue order and nothing else"
todo_auto_doc_test.go:299: the refusal still asserts the pick order: "--auto takes no card arguments; the invocation is the operator's batch approval of the queue in queue order, never an admission"
--- FAIL: TestAutoHelpAndRefusalDoNotAssertPickOrder (0.33s)
```
In that run the 4 `literals share one paragraph/…` subtests, the 2 `passage carries the amendment/…` subtests and the 12 disclosure subtests were red. The `prohibition kept/…` subtests and the mirror `live and template agree` / `template passage is neutral` subtests were green on the unamended documents by construction (they pin text the amendment must not change, or compare two byte-identical copies — characterization, not new behaviour); the mutant probes below are what show those subtests bite.

**E10 — mutant probes (each applied to the committed tree `75ade0ccc`, run, restored by `cp` from a saved copy, `cmp` exit 0 after each; final `git status --short` and `git diff --stat` both empty).**
- Bounding literal reworded to `selection order` in live gtd.md → `TestAutoRankDoctrineAmendment` FAIL (`live gtd.md does not carry the literal "selection order only" on a single line`) and `TestAutoRankMirrorParity` FAIL (`live passage carries no "selection order only"`, live ≠ template).
- Literals split across paragraphs (live kanban-dispatch.md) → `TestAutoRankDoctrineAmendment` FAIL (`no single paragraph carries both …`), `TestAutoRankMirrorParity` FAIL (passage differs).
- Exception literal wrapped across a hard line break (live gtd.md) → `TestAutoRankDoctrineAmendment` FAIL (`does not carry the literal "auto-scoped ranking exception" on a single line`) — the audit finding D-N3 shape.
- Template-only wording change (`unchanged` to `untouched`) → `TestAutoRankMirrorParity` FAIL (`the amended passage differs between …`); control: `TestTodoSkillDocumentsClassification` stayed PASS under the same mutant, so only the new test catches a mirror-only drift.
- Prohibition `never reorders by inferred priority,` deleted from live kanban-dispatch.md → `TestAutoRankDoctrineAmendment` FAIL (`live kanban-dispatch.md lost the prohibition "The leader never picks for the operator, never reorders by inferred priority, …"`).
- Disclosure deleted from the flag help only → `TestAutoRankMarkerDisclosure` FAIL on the four `--auto flag help/…` subtests, the six document subtests stayed PASS (each surface is asserted on its own).
- `in queue order` put back into the refusal string → `TestAutoHelpAndRefusalDoNotAssertPickOrder` FAIL (`the refusal still asserts the pick order`).
- A SPEC id written into both gtd.md copies → `TestAutoRankMirrorParity/template_passage_is_neutral` FAIL (`line 26 carries internal content "SPEC-FOO-BAR-001"`) and the existing `TestTodoSkillDocumentsClassification` FAIL.

**Byte measurements (`wc -c`, before at `946896945`, after at `75ade0ccc`).** live `kanban-dispatch.md` 26352 → 26807 (+455); template `kanban-dispatch.md` 26030 → 26485 (+455); live and template `gtd.md` 37010 → 37897 (+887 each); `internal/cli/todo.go` 61641 → 61882 (+241). The always-loaded surface measured by `TestAlwaysLoadedTokenBudget` rose by 113 tokens (64114 → 64227). The single-edit growth of `kanban-dispatch.md` is below the 1,000-byte duty threshold; the commit body states the sizes anyway. The pre-existing live/template drift of `kanban-dispatch.md` (line 177 only, the `moai worktree sweep` sentence) is unchanged: `diff` after the edits still reports `177c177` and nothing else.

**What `make build` changed.** `make build` ran `agents-emit-check`, `templ-generate`, `gen-catalog-hashes --all` and `go build`; the only tracked change was `internal/template/catalog.yaml` (one hash line, the `moai` skill directory), committed in `5ddbd6c44`. `bin/moai` is ignored.

**Builds, vet, lint, format.** `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./internal/cli/ ./internal/kanban/` exit 0; `golangci-lint run --timeout=2m ./internal/cli/` → `0 issues.`; `gofmt -l` flagged the new test file once (fixed with `gofmt -w`, then no output). The M1/M2 baseline set (27 names) re-ran 27 `--- PASS` after the edits.

**Deviations and interpretations (no scope growth).**
- A fourth test, `TestAutoHelpAndRefusalDoNotAssertPickOrder`, covers the "in queue order" rewording. Its name sits outside the `TestAutoRank` prefix so `go test -list 'TestAutoRank'` stays at the planned count (14 names now; `TestAutoRankAgentDoctrine` joins at M4 for 15).
- The unchanged-prohibitions assertion lives as `prohibition kept/…` subtests inside `TestAutoRankDoctrineAmendment` (AC-TAP-011 says that test carries it) rather than as a separately named top-level test.
- `kanban-dispatch.md` `:29` keeps its prohibition text verbatim and gains one trailing sentence pointing at the next paragraph; both pinned literals sit in the `:31` reconciliation paragraph, where the exception is stated. The phrase "named, not excepted" in that paragraph is retained untouched.
- The flag help carries the marker disclosure without backticks (pflag reads a back-quoted word in a usage string as the value placeholder name), so the test compares on backtick-stripped, whitespace-collapsed text.
- `internal/template/catalog.yaml` is part of commit `5ddbd6c44` (a build artifact of the template edit, same-SPEC cascade), not in the M3 file list of the plan.
- `TestContractModeEmitterSites` and `TestHeadroomInitSurfaceExport` SKIP before and after; their coverage of `kanban-dispatch.md` is therefore unobserved here (Gap).

**Gaps (not observed).** No `internal/cli` package-wide test or coverage run and no CI run (CI owns the repository-wide verdict, PENDING at report time). The two skipped guards above did not run. `TestTemplateNoInternalContentLeak` is red on a pre-existing match outside M3. `internal/mission/governor_test.go` (reads `manager-todo.md`) was not run — M4 owns that file. `TestAutoRankAgentDoctrine` and AC-TAP-015 are M4 and unflipped. Raw logs live in the session scratchpad and are machine-local; the deciding lines are quoted above.

**Residual risk.** The literal check is paragraph-scoped and cannot detect a sentence that generalises the exception in other words (spec §G R-6); the Jev-side surfaces still say "display-only" until the follow-up card lands, and the amended wording only narrows that contradiction. The prohibition subtests quote whole sentences, so a legitimate future rewording of those prohibitions fails the test by design and must update the quote deliberately. The mirror-parity passage for `kanban-dispatch.md` is delimited by two heading literals; a later edit that renames either marker makes the test fail loudly rather than pass vacuously (`t.Fatalf` on a missing marker).

### M4 — Agent text, stale comments, build and parity sweep

Run by manager-develop (cycle_type=tdd), card t1400, branch `WT-auto-priority-pick`, starting HEAD `c56fdbf2f`. M4 commits: `232cd8d41` (manager-todo amendment, template mirror edited first, live copy identical; regenerated codex TOML and catalog hash), `fe847add8` (comment-only corrections in `todo_auto.go` and `todo_edit_move.go`), `08903dd0a` (`TestAutoRankAgentDoctrine`). Every Go command ran with the eleven lane variables scrubbed in one compound `unset … && <command>` invocation; the Go toolchain compiled the tree under test directly and no installed `moai` binary was invoked. The post-edit measurements below ran on a clean tree at HEAD `08903dd0a`; this section is the only later edit.

**Charter note.** `manager-develop.md` lists agent files under "Forbidden modifications … out of run-phase scope". The M4 delegation and plan §E M4 name `manager-todo.md` (REQ-TAP-014) as this milestone's deliverable, so the edit follows that instruction; flagged here so the leader can confirm the reading.

**Pre-flight (HEAD `c56fdbf2f`, clean tree).** `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; baseline of the 31 M1-M3 names (`go test -count=1 -v -run '^(…31 names…)$' ./internal/cli/ | grep -c '^--- PASS'` → `31`); `TestManagerTodoJudgmentSubRoleBoundary` and `TestManagerTodoJevBoundaryNamesGrade3` (`internal/mission/governor_test.go:27`, `:57`) both PASS, `ok … 0.363s`; `go test -count=1 ./internal/template/...` → three FAIL (listed in the guard table), `agentemit` and `commandemit` ok. `wc -c` of `manager-todo.md`: live 4328, template 4328 (`cmp` exit 0). `go test -list 'TestAutoRank' ./internal/cli/` → 14 names.

**E8 — verbatim RED of `TestAutoRankAgentDoctrine` on the unamended agent file** (test written first; run before any agent edit; deciding lines, live copy first, the template copy printed the same ten lines):
```
todo_auto_doc_test.go:386: live manager-todo.md serial-cycle contract does not carry the literal "auto-scoped ranking exception" on a single line
todo_auto_doc_test.go:386: live manager-todo.md serial-cycle contract does not carry the literal "selection order only" on a single line
todo_auto_doc_test.go:397: live manager-todo.md serial-cycle contract: no single paragraph carries both "auto-scoped ranking exception" and "selection order only"
todo_auto_doc_test.go:386: live manager-todo.md Jev Decision Boundary does not carry the literal "auto-scoped ranking exception" on a single line
todo_auto_doc_test.go:386: live manager-todo.md Jev Decision Boundary does not carry the literal "selection order only" on a single line
todo_auto_doc_test.go:397: live manager-todo.md Jev Decision Boundary: no single paragraph carries both "auto-scoped ranking exception" and "selection order only"
todo_auto_doc_test.go:406: live manager-todo.md still asserts the pre-amendment wording "strict queue order"
todo_auto_doc_test.go:406: live manager-todo.md still asserts the pre-amendment wording "serial consumption of the queue in queue order and nothing else"
todo_auto_doc_test.go:406: live manager-todo.md still asserts the pre-amendment wording "beyond that order"
todo_auto_doc_test.go:406: live manager-todo.md still asserts the pre-amendment wording "consults the Jev judgment scripts as a display-only signal for dispatch order and priority"
--- FAIL: TestAutoRankAgentDoctrine (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.848s
```
Audit finding D-N4 (named search strings): the four absence assertions quote phrases read from the pre-edit file (Primary Mission line, serial-cycle first bullet twice, frontmatter description) and the run above shows each present before the amendment. The kept-prohibition subtests and the byte-equality subtest were green on the unamended files by construction (characterization); the mutant probes show they bite.

**What the amendment says.** Frontmatter description, Primary Mission line, the serial-cycle first bullet and the Jev Decision Boundary now state the exception; both pinned literals sit on single lines in one paragraph of each of the two sections (`grep -c` output below). The Jev prohibitions are untouched: `Jev output is judgment input for the lead — never authority.` and `It is never the basis of a queue mutation, a completion verdict, a merge approval, or any operator-gate decision.` stay verbatim. Size: live and template `manager-todo.md` 4328 → 5095 bytes each (+767); neither is an always-loaded file.

**Cascade the plan did not name (same-SPEC, required by committed guards).** The codex TOML `internal/template/templates/.codex/agents/moai/manager-todo.toml` embeds the agent body, and `make build` runs a read-only drift check that never regenerates, so `make agents-emit` was run (`ok … agentemit 0.480s`); the TOML diff is 16 insertions, 5 deletions. Editing the template agent file also changed its catalog hash: `TestManifestHashFormat` went red (`CATALOG_HASH_UNSTABLE: manager-todo stored hash=811507e7… computed hash=6cf817ab…`) until `make build` regenerated `internal/template/catalog.yaml`.

**What `make build` changed.** `make build` ran `agents-emit-check`, `commands-emit-check`, `tool-policy-drift-check`, `templ-generate`, `gen-catalog-hashes --all` and `go build`; the only tracked change was `internal/template/catalog.yaml` (one hash line, the `manager-todo` entry, `6cf817ab…`), committed in `232cd8d41`. `bin/moai` is ignored.

**E1 — AC-TAP-015 and the final matrix (this run, this tree, HEAD `08903dd0a`).**

| AC | Command | Observed | Status |
|---|---|---|---|
| AC-TAP-015 | `grep -c "auto-scoped ranking exception" .claude/agents/moai/manager-todo.md internal/template/templates/.claude/agents/moai/manager-todo.md`; the same with `"selection order only"`; `go test -count=1 -v -run '^TestAutoRankAgentDoctrine$' ./internal/cli/` | live `:2`, template `:2` for both literals, exit 0; `--- PASS: TestAutoRankAgentDoctrine` with 4 stale-phrase subtests per copy PASS, 2 region subtests per copy PASS, `live and template are byte-identical` PASS | PASS |
| AC-TAP-001 | `go test -count=1 -v -run '^TestAutoRankSelectionRecord$' ./internal/cli/`; `grep -rl --exclude="*_test.go" "selection: source=" internal/cli` | `--- PASS: TestAutoRankSelectionRecord (0.53s)`; grep stdout `internal/cli/todo_auto_rank.go`, exit 0; control `grep -rl --exclude="*_test.go" "jev: unavailable" internal/cli` → `internal/cli/todo_auto.go`, exit 0 | PASS |
| AC-TAP-002 | `TestAutoRankJevOrdering` | `--- PASS … (0.76s)` | PASS |
| AC-TAP-003 | `TestAutoRankFallbackReasons` | `--- PASS … (1.59s)` | PASS |
| AC-TAP-004 | `TestAutoRankFallbackOrder` | `--- PASS … (0.00s)` | PASS |
| AC-TAP-005 | `TestAutoRankDemotion` | `--- PASS … (0.00s)` | PASS |
| AC-TAP-006 | `TestAutoRankUnmeasuredSignal` | `--- PASS … (0.00s)` | PASS |
| AC-TAP-007 | `TestAutoRankBlockedExcluded` | `--- PASS … (0.44s)` | PASS |
| AC-TAP-008 | `TestAutoRankRescueFirst` | `--- PASS … (0.30s)` | PASS |
| AC-TAP-009 | `TestAutoRankQueueUnchanged`, `TestAutoRankNoQueueWriteGuard`; `grep -n "Mutate\|ArchiveCard" internal/cli/todo_auto_rank.go`; `grep -c '^func ' …` | both `--- PASS` (0.30s, 0.00s); grep no output, exit 1; function count `18`, exit 0 | PASS |
| AC-TAP-010 | `TestAutoRankJevMalformedAnswer` | `--- PASS … (2.05s)` | PASS |
| AC-TAP-011 | `grep -c` of both literals over the four doctrine files; `TestAutoRankDoctrineAmendment` | each of the four files `:1` for both literals, exit 0; `--- PASS: TestAutoRankDoctrineAmendment (0.00s)` | PASS |
| AC-TAP-012 | the ten characterization names | ten `--- PASS`, no `--- FAIL`, `ok … 1.982s` | PASS (see the Definition-of-Done item 3 deviation recorded under M2: `todo_auto_test.go` carries the D-N1 hermetic change; `todo_relation_filter_test.go` is not in `git diff --stat 38b54f29b..HEAD`) |
| AC-TAP-013 | `TestAutoRankMirrorParity`, `TestTodoSkillDocumentsClassification`, `TestSanitizedPairParity` | all three `--- PASS` | PASS |
| AC-TAP-014 | `TestAutoRankMarkerDisclosure`; `grep -c "\[보류" …gtd.md` (live, template) | `--- PASS`; `:2` and `:2`, exit 0 | PASS |

Sweep control: `go test -list 'TestAutoRank' ./internal/cli/` → 15 names then `ok … 0.729s`; the 15-name anchored run printed 15 top-level `--- PASS` and no `--- FAIL` (`ok … 6.606s`). The wider 32-name run (the 31 baseline names plus `TestAutoRankAgentDoctrine`) printed 32 `--- PASS`.

**E2 — builds, vet, lint, format (HEAD `08903dd0a`).** `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./internal/cli/ ./internal/kanban/` exit 0; `golangci-lint run --timeout=2m ./internal/cli/ ./internal/kanban/` → `0 issues.`; `gofmt -l` on the three touched Go files → no output.

**E3 — coverage.** The new test adds helpers inside a `_test.go` file and no production statement; the only production edits are comments, so there is no production coverage figure to move. `TestAutoRankAgentDoctrine` runs 19 subtests (9 per copy plus the byte-equality one). No package-wide figure was taken (CI owns it).

**E4 — greps.** The neutrality expression (`SPEC-[A-Z]`, `REQ-[A-Z]`, `AC-[A-Z]`, ISO date, 9+ hex run) over the 16 lines added to the template `manager-todo.md` (`git diff -U0`) → `0` matches; `AskUserQuestion` over the same lines → `0`; `grep -n 'AskUserQuestion\|mcp__askuser'` over `todo_auto.go`, `todo_edit_move.go`, `todo_auto_rank.go`, `todo.go` → no output. Comment-only proof for the two Go files: `git diff -U0 -- internal/cli/todo_auto.go internal/cli/todo_edit_move.go` filtered to changed lines → 29 lines matching `^[+-]\s*//` and `0` lines not matching it (positive control: the first count is non-zero).

**E9 — guards that read `manager-todo.md`, the template tree or the amended documents (source lines; before = HEAD `c56fdbf2f`, after = HEAD `08903dd0a`).**

| Guard (source) | Before | After |
|---|---|---|
| `TestManagerTodoJudgmentSubRoleBoundary` (`internal/mission/governor_test.go:27`, reads the live agent file) | PASS | PASS |
| `TestManagerTodoJevBoundaryNamesGrade3` (`internal/mission/governor_test.go:57`) | PASS | PASS |
| `TestGoldenCommittedArtifactsMatchEmission`, `TestRealSetBodiesByteEqual` (`internal/template/agentemit/golden_test.go`) | PASS (package ok) | PASS after `make agents-emit` regenerated the TOML (the body edit makes the committed TOML stale until then; not measured red in this run) |
| `TestManifestHashFormat` (`internal/template/catalog_tier_audit_test.go:459`) | PASS | FAIL after the template agent edit (`CATALOG_HASH_UNSTABLE: manager-todo …`), PASS after `make build` regenerated `catalog.yaml` |
| `TestSanitizedPairParity` (`internal/template/sanitized_pair_parity_test.go`) | PASS (not in the package's failure list) | PASS (`-v`) |
| `TestRuleTemplateMirrorDrift`, `TestDeclaredRuleMirrorForks`, `TestGTDCanonicalSurfaceGolden` | PASS (not in the failure list) | PASS (`-v`) |
| `TestTemplateNoInternalContentLeak` (`internal/template/internal_content_leak_test.go:1617`) | FAIL, one match | FAIL, the same single match: `templates/.claude/rules/moai/workflow/worktree-integration-ops.md` class C1 `SPEC-SESSION-ANCHOR-ATTR-001`; no match from any file this card touched. PRE-EXISTING |
| `TestRuleDateProvenance` (`internal/template/rule_date_provenance_audit_test.go:183`) | FAIL — 3 matches `2026-09-29` in `.claude/rules/moai/workflow/worktree-integration-ops.md` lines 234, 252, 253 | FAIL, same three. PRE-EXISTING, not in the delegation's known list, reported not fixed |
| `TestSyncGateCpp_LocalAndTemplateCopiesIdentical` (`internal/template/hook_cpp_gate_behavior_test.go:184`) | FAIL — `sync-phase-quality-gate.sh differs between the local and template copies` | FAIL, same. PRE-EXISTING, reported not fixed |
| `TestJevAmendmentLinkage` (`internal/contract/kickoff/activation_test.go:208`) | not run before the M4 edits (M3 recorded PASS at an earlier HEAD) | PASS |
| `TestTodoSkillDocumentsClassification`, `TestTodoHoldDocumentedOnEverySurface`, `TestTodoListJSONShapeMatchesDoc`, `TestTodoDoctrine_MirrorParityAndStatedColumnCount` | not run before the M4 edits | PASS; `TestHeadroomInitSurfaceExport` SKIP |
| `TestPriorityRank` (`internal/kanban`) | covered by the M1 record | PASS |

The whole `./internal/template/...` package: before three FAIL, after the same three FAIL, and no other. `internal/spec` and `internal/contract` hold no test that reads `manager-todo.md` (grep over `*_test.go` for the agent name found readers only in `internal/mission`, `internal/template` and fixture-based `internal/cli` tests that build their own temp repositories).

**E10 — mutant probes (each applied to the committed tree plus the uncommitted test, run, then restored by `cp` from a saved copy; `git status --short` listed only the uncommitted test file and `git diff --stat` on both agent copies was empty afterwards).**
- Drop `selection order only` from the live Jev paragraph → `live manager-todo.md Jev Decision Boundary does not carry the literal "selection order only" on a single line` and `no single paragraph carries both …`, plus the byte-equality failure.
- Split the two literals across paragraphs in the live serial-cycle contract (blank line inserted between them) → `live manager-todo.md serial-cycle contract: no single paragraph carries both …` (the single-line check stays green, as it should), plus byte-equality.
- Change only the template copy (`nothing else` → `nothing more`, same byte length) → only `manager-todo.md differs between the live file (5095 bytes) and the template mirror (5095 bytes)` fails.
- Restore `strict queue order` in the live copy → `still asserts the pre-amendment wording "strict queue order"`; restored in both copies → exactly the two stale-phrase lines, nothing else.

**E11 — Drift Guard (plan §E "Files changed" versus `git diff --stat 38b54f29b..HEAD`, measured on HEAD `08903dd0a`, before this evidence commit).** Planned: 13 core files plus `todo_edit_move.go` = 14 (`classification.go`, `classification_test.go`, `todo_auto_rank.go`, `todo_auto_rank_test.go`, `todo_auto.go`, `todo.go`, `todo_auto_doc_test.go`, both `kanban-dispatch.md`, both `gtd.md`, both `manager-todo.md`, `todo_edit_move.go`). The diff lists 20 files: the 14 planned, the two SPEC artifacts the run phase owns (`progress.md` evidence, `spec.md` status/`updated:`), and four unplanned:
- `internal/cli/doctor_jev_test.go` (+1 line) — M2: `TestJevCallPath_HasExactlyTheDeclaredConsumers` was red at the start of M2 because `todo_auto_rank.go` imports `internal/jev`; the guard's own comment names the allow-list as the way a consumer arrives.
- `internal/cli/todo_auto_test.go` (+22) — M2: the D-N1 hermetic change to `TestTodoAutoEntryPointFlag`, which otherwise reached the real `gh` once the live seams were wired (observed with a fake `gh`).
- `internal/template/catalog.yaml` (4 changed lines across M3 and M4) — build artifact of the template edits, regenerated by `make build`; same-SPEC cascade.
- `internal/template/templates/.codex/agents/moai/manager-todo.toml` (M4) — regenerated by `make agents-emit`; the committed-TOML drift guard requires it whenever an agent body changes.
The small `todo.go` (+15) and `todo_auto.go` (+25) edits are planned files, not drift. Drift: 4 unplanned of 18 changed non-SPEC files = 22.2% (4 of the 14 planned = 28.6% against the plan's count; counting the two SPEC artifacts as files, 6 of 20 = 30.0%). All three readings are at or below the 30% threshold: warn, no re-planning. The three scope-creep candidates a reviewer might expect are absent: no Jev-side document or code surface was touched, no `internal/jev` change, no SPEC artifact other than `progress.md` and the M1 `spec.md` status line.

**Deviations and interpretations (no scope growth).**
- `TestAutoRankAgentDoctrine` quotes four stale phrases; the frontmatter description clause is stale-checked on the normalized text, so a reflowed copy is caught too.
- Region scoping: the serial-cycle region starts at `Serial-cycle contract (` and the Jev region at `## Jev Decision Boundary`, each ending at the next level-two heading, so the literal pair cannot satisfy the test from an unrelated section; the Primary Mission paragraph carries neither literal.
- The Primary Mission line keeps "queue order" in its general sense and points at the serial-cycle contract; "strict queue order" is gone.
- The Jev Decision Boundary keeps its first sentence ("a permitted display-only signal for dispatch order and priority judgment") because the instruction keeps the Jev ordering-signal prohibition verbatim; the exception is added as its own sentence after it.
- The `todo_edit_move.go` `:99` correction replaces "there are no priority fields" with the actual state (an optional recorded classification applied by the add path; `move` does not read it), read from `kanban/classification.go` in plan §Findings (a)/(b) rather than re-derived.

**Gaps (not observed).** No `internal/cli` package-wide test or coverage run and no CI run (CI owns the repository-wide verdict, PENDING at report time). No live Jev request was made and the score level-index base (plan A-6) remains unobserved. `TestContractModeEmitterSites` and `TestHeadroomInitSurfaceExport` SKIP, so their coverage of the amended documents is unobserved. The agentemit golden guard was not observed red before `make agents-emit` ran. `TestJevAmendmentLinkage` and the cli doc guards were not run before the M4 edits (M4 touches none of their subject files, but that is an inference, not a measurement). The three pre-existing red template guards were reported, not fixed. Raw logs live in the session scratchpad and are machine-local; the deciding lines are quoted above.

**Residual risk.** The stale-phrase absence check is phrase-scoped: a sentence that restates "queue order only" in other words passes it (spec §G R-6). The Jev Decision Boundary still opens with the display-only sentence and the exception follows it, so a reader who stops after the first sentence is not told about the exception until the follow-up card amends the Jev-side surfaces. The kept-prohibition subtests quote whole sentences and fail by design on a legitimate reword. The codex TOML and the catalog hash are derived artifacts: a later body edit that skips `make agents-emit` and `make build` turns two guards red by design.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-10-01T19:46:39Z
card: t1400
cycle_type: tdd
run_commit_sha: pending-backfill
m1_to_mN_commit_strategy: per-milestone commits on branch WT-auto-priority-pick, no push, no amend; M1 96e2c71fe 110918d14; M2 719e6c1b4 4050769f1 (evidence 946896945); M3 5ddbd6c44 75ade0ccc (evidence c56fdbf2f); M4 232cd8d41 fe847add8 08903dd0a (evidence: the commit that adds this section)
ac_pass_count: 15
ac_fail_count: 0
ac_matrix: AC-TAP-001 to AC-TAP-015 PASS, each observed in this run at HEAD 08903dd0a (see the M4 matrix); AC-TAP-012 is a regression guard with the recorded Definition-of-Done item 3 deviation
preserve_list_post_run_count: not-applicable (plan.md names no PRESERVE list)
l44_pre_commit_fetch: not-run (lane worktree, no push; the leader batch-pushes)
l44_post_push_fetch: not-applicable (nothing pushed)
new_warnings_or_lints_introduced: 0 (golangci-lint ./internal/cli/ ./internal/kanban/ -> 0 issues; go vet exit 0)
cross_platform_build:
  host: go build ./... exit 0 (HEAD 08903dd0a)
  windows_amd64: GOOS=windows GOARCH=amd64 go build ./... exit 0 (HEAD 08903dd0a)
total_run_phase_files: 20 changed files in git diff --stat 38b54f29b..HEAD at HEAD 08903dd0a (14 planned, 2 SPEC artifacts, 4 unplanned)
drift_guard: 4 unplanned of 18 non-SPEC files = 22.2% (<= 30%, warn only)
pre_existing_red: TestTemplateNoInternalContentLeak, TestRuleDateProvenance, TestSyncGateCpp_LocalAndTemplateCopiesIdentical (same before and after M4, none from files this card touched)
repository_wide_test_verdict: PENDING (owned by CI on the integration branch)
```
