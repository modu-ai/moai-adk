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
