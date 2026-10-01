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
