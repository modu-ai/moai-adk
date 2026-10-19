# Progress: SPEC-FACTORY-DEFAULT-MODE-001

Card t1737 · Tier M · plan-phase artifact set (spec.md, plan.md, acceptance.md, progress.md,
decision-index.md). This file is lane-owned for §E.2–§E.4; manager-spec populates §E.1 only.

## §E.1 Plan-phase Audit-Ready Signal

- audit_ready: true
- plan_status: audit-ready
- plan_complete_at: 2026-10-19T03:20:00Z (UTC)
- artifacts: spec.md (7 REQ), acceptance.md (11 AC: 8 release-blocking, 3 regression-guard),
  plan.md, decision-index.md (6 rows: 3 DECIDED, 0 POLICY-COVERED, 0 EVIDENCE-NEEDED, 3
  FOUNDER — two implementation-level DEFAULT-APPLIED, one product-level operator-gated)
- lint: `moai spec lint --strict SPEC-FACTORY-DEFAULT-MODE-001` — (pending first run; the
  observed result is recorded here after the run)
- predecessor note: C1 (t1735) skipped by operator exception; safety rationale at plan.md §A.2.

### Plan repair — round-1 audit FAIL (2026-10-19, this session)

Round-1 cold plan-audit: FAIL, overall 0.875, MP-8 must-pass failure (verdict:
`.moai/reports/t1737/plan-audit-cold.md`, audited_sha `f6742080a6`, receipt
`rcpt-c9f6d2f8c1f45eb7e134711c`). Repairs applied by manager-spec:

- **D1 (critical)** — acceptance.md two-cell state re-authored as observed-failure evidence:
  new §D.0 evidence ledger (EV-FDM-01…-15, -A1) carrying command + verbatim stdout + exit
  code + tree SHA per release-blocking AC; AC-FDM-004 demoted release-blocking →
  regression-guard (undecidable disposition — its fail-safe-serial RED cannot be honestly
  observed; none recorded); swept-set guard added (`go test -v` + named `--- PASS:` line
  required; a `[no tests to run]` result is never green); §F DoD updated. Observation
  channels: tree greps + file reads at `f6742080a6`; installed binary `moai todo classify` →
  `Unknown command "classify" for "moai todo"`, exit 1 (build `7705b365e2`, predates HEAD —
  lag stated per tool-provenance attribution); the codex zero-sweep output cited attributed
  to the verdict receipt. PENDING-observation: the three existing parent tests' passing
  outputs (EV-FDM-12/-13/-14) — the Go-restraint window was active at repair time; the lane
  records them at run entry alongside plan §C P2.
- **D2 (major)** — REQ-FDM-004 extended with the serial-slot conflict refusal (in-flight
  target ∧ requested `serial` ∧ another distinct in-flight slot holder → refuse with
  recovery advice, write nothing), grounded in `factorySerialInFlightExcluding`
  (`factory_card.go:324-340`) and the `:943`/`:1406` gates; new AC-FDM-011 +
  `TestTodoClassifySerialSlotConflictRefusal` (M2); decision-index Q6 appended (FOUNDER,
  implementation-level, DEFAULT-APPLIED: refuse).
- **D3 (minor)** — SPEC-FACTORY-CONFLICT-SLOTS-001 disposition stated in-body (spec.md §E;
  plan.md §H): proposed split not executed on this lineage; frontmatter reference kept as
  disposition-only context.
- **D4 (minor)** — plan.md §C P1 expected observation restated as measured (zero hits,
  exit 1).
- **D5 (minor)** — plan.md §F M3 relabeled guard-addition (green-now cells kept green), body
  aligned; M2 coverage line extended with AC-FDM-011.
- REQ/AC ids otherwise unchanged (no renumbering, no renaming); all other content kept stable
  for the round-2 delta re-audit (scope: acceptance.md#two-cell-adoption-state,
  spec.md#req-fdm-004, plus the AC-FDM-011 addition).
- **Round-2 repair (delta FAIL 0.9375, receipt `rcpt-ed1e059d6960d148efbfb510`, audited_sha
  `67ac361675`)** — R2-D1: EV-FDM-05/-06/-07/-09/-10/-15 re-observed with complete stdout at
  `67ac361675` (`internal/` byte-identical across all three trees); R2-D2: AC-FDM-001 bullet +
  EV-FDM-12 note reworded to the expected-green/PENDING form; R2-D3: new EV-FDM-A2 (attributed
  overlap reproduction, receipt `rcpt-c9f6d2f8c1f45eb7e134711c`) with AC-FDM-011 citations
  repointed from EV-FDM-A1; R2-D4: EV-FDM-13/-14 recording path corrected to plan §E E2;
  R2-D5: plan §C P4 enumeration extended to Q1–Q6. Single new id EV-FDM-A2; all REQ/AC/entry
  ids otherwise unchanged.

## §E.2 Run-phase Evidence

Run-phase owner: manager-develop (card t1737, worktree `.moai/worktrees/t1737`, branch
`WT-spec-factory-default`). Entry tree `2cd891e5a8`; run commits `1aec6f8986` (M1),
`b8e2f92713` (M2), `82462814da` (M3). Every Go invocation below ran under the board-8c1e
resource admission (1-min load ≤ 32, mem free ≥ 50% re-checked per call; one Go call at a
time) and the compound environment scrub (`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR
MOAI_KANBAN_SETTINGS_INJECTED && TMPDIR=/private/tmp/moai-t1737 <cmd>`; TMPDIR created
first, never under $HOME or in-repo). Raw outputs: `/private/tmp/moai-t1737/` (machine-local
scratch; the deciding lines are quoted verbatim here).

### Baselines (debt FDM-PENDING-BASELINES disposal — recorded BEFORE the flip, tree `2cd891e5a8`)

- **EV-FDM-12** — `go test -v ./internal/factory -run TestEffectiveCardClassificationDerivesAbsentDefaults`,
  exit 0, verbatim: `--- PASS: TestEffectiveCardClassificationDerivesAbsentDefaults (0.00s)` /
  `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/factory	0.408s`.
- **EV-FDM-13/-14** —
  `go test -v ./internal/cli -run 'TestTodoAddDeciderFailureFallsBackWithNotice|TestFactoryNextParallelizableLeasesBesideLiveSerial'`,
  exit 0, verbatim: `--- PASS: TestFactoryNextParallelizableLeasesBesideLiveSerial (2.43s)`
  (both arms PASS) / `--- PASS: TestTodoAddDeciderFailureFallsBackWithNotice (0.87s)` (both
  subtests PASS) / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/cli	4.481s`.
- Recording-path note: acceptance.md §D.0 routes EV-FDM-13/-14 through plan §E E2 (the full
  package suite). The full `./internal/cli` suite run hit go's default 10-minute package
  timeout on this loaded machine (1602 tests already `--- PASS` when `panic: test timed out
  after 10m0s` fired, before reaching these tests); the baseline was recorded via the same
  package binary with the two names as the selector — the verbatim passing output the debt
  requires. The full suite was re-run with `-timeout 30m` for E2 (below). Machine-load
  artifact, not a test defect.
- Baseline-breakage record (plan §C P2): the flip changed the parent assertion
  `TestEffectiveCardClassificationDerivesAbsentDefaults` (absent want serial →
  parallelizable, `internal/factory/classification_test.go`); the parent fail-safe tests
  (`TestTodoAddDeciderFailureFallsBackWithNotice`,
  `TestTodoAddClassificationFileUnavailableFallsBack`,
  `TestTodoAddDefaultDeciderPrintsNoNotice`) kept passing UNCHANGED — the fail-safe value
  they pin did not move (that is AC-FDM-004's separation clause holding).

### RED→GREEN pairs (TDD; verbatim RED captured before each GREEN)

| Milestone | Test | RED evidence (verbatim) | GREEN |
|---|---|---|---|
| M1 | `TestEffectiveClassificationAbsentDefaultsParallelizable` | exit 1: `effective classification of an absent field = {Priority:normal Blocked:false Mode:serial Decider:default ClassifiedAt: Reason:}, want {... Mode:parallelizable ...}` + `absent-field derivation carries no reason` | `--- PASS` (0.00s) |
| M1 | `TestQueueAbsentCardClassificationDefault` | exit 1: `nil record classification mode = "serial", want parallelizable` / `id-absent classification mode = "serial"` / `absent-field item classification = {... Mode:serial ...}` | `--- PASS` (0.18s) |
| M1 | `TestTodoAddDefaultClassificationParallelizable` | exit 1: `default admission classification = &{... Mode:serial ... Reason:deterministic default: no classification judgment supplied}, want parallelizable under decider default` | `--- PASS` (0.58s) |
| M1 | `TestTodoAddDeciderFailureFailsafeSerial` | none — regression guard, green-on-arrival (AC-FDM-004 demoted per §D.0; documented) | `--- PASS` (0.38s first run; 0.48s at E1) |
| M2 | `TestTodoClassifyQueuedCard` | exit 1: `classify: unknown flag: --mode` (verb absent — the unregistered token falls through to the add path; EV-FDM-03/-04's absence, observed) | `--- PASS` (1.49s) |
| M2 | `TestTodoClassifyPickedCard` | exit 1: same verb-absence signature | `--- PASS` (1.52s) |
| M2 | `TestTodoClassifyRefusals` | exit 1: same verb-absence signature (7 subtests) | `--- PASS` (5.69s; all 7 subtests PASS) |
| M2 | `TestTodoClassifySerialSlotConflictRefusal` | exit 1: same verb-absence signature (3 subtests) | `--- PASS` (3.70s; all 3 subtests PASS) |
| M3 | `TestFactorySelectionReadsReclassifiedMode` | none — guard-addition (plan §F M3: green-by-construction) | `--- PASS` (2.68s; 3.12s at E1) |
| M3 | `TestSerialExclusivityAfterDefaultFlip` | none — guard-addition | `--- PASS` (2.81s; 2.17s at E1) |
| M3 | `TestRecordedClassificationNotRewritten` | none — guard-addition | `--- PASS` (0.08s; 0.07s at E1) |

M2 GREEN first attempt failed on three defects of the new TEST code itself (bystander
assertion ignored the add path's own default record; index-based bystander comparison; a
hash helper reading the nonexistent `backlog.json` path) — fixed by verbatim by-ID
comparison (reflect.DeepEqual) and by switching byte-identity reads to the established
`readBacklogBytes` helper (`todo_edit_move_test.go:45`), the same mechanism the hold-verb
refusal tests use. Second attempt: all green.

### E1 — AC PASS/FAIL matrix (observed this run)

Verification: `go test -v ./internal/cli ./internal/factory -run 'Classification|TodoClassify|FactorySelection|SerialExclusivity'`
→ exit 0, 42 `--- PASS` lines, `ok` on both packages (26.010s / 17.111s). Every
release-blocking AC's named `--- PASS:` line was observed (swept-set guard satisfied — no
`[no tests to run]` anywhere):

| AC | Status | Verbatim observed line (this run) |
|---|---|---|
| AC-FDM-001 | PASS | `--- PASS: TestEffectiveClassificationAbsentDefaultsParallelizable (0.00s)` |
| AC-FDM-002 | PASS | `--- PASS: TestQueueAbsentCardClassificationDefault (0.18s)` |
| AC-FDM-003 | PASS | `--- PASS: TestTodoAddDefaultClassificationParallelizable (0.58s)` |
| AC-FDM-004 | PASS (regression-guard) | `--- PASS: TestTodoAddDeciderFailureFailsafeSerial (0.48s)` — observed via its own verification command: the plan §E E1 selector's four alternation branches match none of this test's name segments, so the combined run swept it zero times (the §D.0 swept-set hazard, met in the wild); the AC's own `-run` command from acceptance.md was executed and its `--- PASS` line observed. |
| AC-FDM-005 | PASS | `--- PASS: TestTodoClassifyQueuedCard (0.88s)` |
| AC-FDM-006 | PASS | `--- PASS: TestTodoClassifyPickedCard (1.09s)` |
| AC-FDM-007 | PASS | `--- PASS: TestTodoClassifyRefusals (3.77s)` (7 subtests PASS) |
| AC-FDM-008 | PASS | `--- PASS: TestFactorySelectionReadsReclassifiedMode (3.12s)` |
| AC-FDM-009 | PASS (regression-guard) | `--- PASS: TestSerialExclusivityAfterDefaultFlip (2.17s)` |
| AC-FDM-010 | PASS (regression-guard) | `--- PASS: TestRecordedClassificationNotRewritten (0.07s)` |
| AC-FDM-011 | PASS | `--- PASS: TestTodoClassifySerialSlotConflictRefusal (3.27s)` (3 subtests PASS) |

### E2 — full package suites

First combined run (`go test -timeout 30m ./internal/cli ./internal/factory`, concurrent
packages) observed 4 failures — three baselines the flip moves, updated per plan §C P2 and
committed (`d9ae56471b`; `TestDefaultCardDeciderIsDeterministic`,
`TestFactoryNextStillRefusedBehindADrivenRow`,
`TestFactoryNextRefusesAgainAfterLaterAssign` — full diagnosis in the commit message and the
card progress record), plus one load-jitter timing failure
(`TestFactoryLeaseQueueLockStallBounded`, 8.81s against an 8.75s wall-clock bound under the
concurrent two-package run) that passes in isolation (`--- PASS: TestFactoryLeaseQueueLockStallBounded (19.62s)`,
both subtests). The pre-existing `TestCodexSpecFiles_ExecPrimitivesCodexOnly` failure
(recorded above, present before any change) also appears. The re-run executes the packages
SERIALLY to remove the cross-package contention variable:

- `go test -timeout 20m ./internal/factory` → `ok  	github.com/modu-ai/moai-adk/internal/factory	358.945s`, exit 0.
- `go test -timeout 30m ./internal/cli` → PENDING (running; recorded on completion).

### E3 — coverage on touched packages

### E2 — full package suites

- `go test -timeout 20m ./internal/factory` (serial) → `ok  	github.com/modu-ai/moai-adk/internal/factory	358.945s`, exit 0 — clean.
- `go test -v -timeout 75m ./internal/cli` (serial) → COMPLETED in 2604.562s, exit 1, final
  FAIL list exactly six entries, every one dispositioned below. The verbose run's relevant
  coverage: all 14 flip/verb-family tests carry `--- PASS` lines in that run (full list in
  the card progress record); the earlier 30m and 10m ceilings were total-length overruns
  under measured machine load (1602+ tests green before the first ceiling; the panic dumps
  showed normal alphabetical progress, no hung test) — **no conclusion of
  absence-of-individual-defects is drawn from a timeout alone**; the completed 75m run is
  the authoritative sweep.

The six final FAIL entries and their dispositions (leader-approved; commits
`d9ae56471b`, `a05824c1c6`, `606de2c47b`, `7916bb2859`, `8d302004ed`):

1. **`TestCodexSpecFiles_ExecPrimitivesCodexOnly` — pre-existing defect, attributed to card
   t1869; explicitly NOT counted as a pass.** Verbatim failure:
   `codex_launcher_guards_test.go:255: mcp_codex.go: expected-first-argument row binaryPath
   found no process-start primitive — the table must track the source (a zero-match scan is
   the vacuous-green shape t495's floor guarded)`. Provenance chain (each link measured, not
   assumed): `mcp_codex.go` last changed at `85804b6fe2` (2026-10-18, "fix(codex): drop Orca
   pane-owned env from codex review/audit/task children (card t1869)" — `git log -1` on the
   file, this lane); its guard test last changed in the t500-era merge (`f7f67d04cf`) —
   t1869 changed the source without updating the guard's expectation table; the failure was
   observed in this lane's FIRST full-suite run at entry tree `2cd891e5a8`, before any
   change of this card; this card's diff `git diff --name-only 74ad197380..HEAD` is 15 files
   with zero matches for `mcp_codex|CodexSpecFiles|codex_launcher` (grep exit 1); lane-32
   archaeology and lane-56's empty-diff reproduction (attributed from the lane relay)
   concur. Out of card t1737's scope; belongs to t1869's follow-up.
2. **`TestSD_AC015_LaneQueueAllowlistWalk` — card-scope surface-pin repair, GREEN.** The
   walk derives its set from the command tree; the new verb needed a verbArgs entry matching
   its flag surface (`id, --mode, --reason`), walking as lane-REFUSED (a queue mutation).
   Scoped: `go test -v ./internal/cli -run 'TestSD_AC015_LaneQueueAllowlistWalk|TestGTDAllTodoVerbsParity|TestTodoClassify'`
   → exit 0, `--- PASS: TestSD_AC015_LaneQueueAllowlistWalk (2.14s)` (+ parity and the
   classify family, all PASS). Commit `a05824c1c6`.
3. **`TestGTDAllTodoVerbsParity` — same-class surface-pin repair, GREEN.** `classify` added
   to `todoWant` (gtdWant inherits through the clone; sort kept). Same scoped run →
   `--- PASS: TestGTDAllTodoVerbsParity (0.26s)`. Commit `a05824c1c6`.
4. **`TestLiveReadersUnchangedByHistoryVerb` — the approved M1 flip reaching the golden;
   minimal golden update, GREEN.** The four `classification.mode` fields in
   `testdata/golden/live-readers/list-json.txt` are default-decider records whose answer
   REQ-FDM-002 flips; field values only (serial → parallelizable), no regeneration, no
   unrelated fields; the t1737 provenance block follows the file's t384/t1310/t1332
   convention. Scoped: `go test -v -run TestLiveReadersUnchangedByHistoryVerb` → exit 0,
   `--- PASS: TestLiveReadersUnchangedByHistoryVerb (1.37s)`. Commit `606de2c47b`.
5. **`TestTodoLanded_SHAValidation` — test-harness token collision, whole-token repair,
   GREEN.** The attribution clause's bare `ReplaceAll(s, "t1", ...)` also rewrote path
   tokens sharing the id's prefix (a TMPDIR like `/private/tmp/moai-t1737`), fabricating a
   divergence; the strip is now word-boundary scoped (`\bt1\b`) — a real leak still matches
   and still diverges, so the assertion's reach is unchanged; no TMPDIR change to hide.
   Scoped: `go test -v -run TestTodoLanded_SHAValidation` → exit 0,
   `--- PASS: TestTodoLanded_SHAValidation (2.24s)` across all four subtests. Commit
   `7916bb2859`.
6. **`TestTodoVerbSurfaceZeroDelta` — same-class declaration repair, GREEN.** `classify`
   added to `permittedVerbAdditions` (26 → 27) with the verb's full Use text and contract in
   the declaration comment; the frozen list untouched. Scoped: `go test -v -run
   TestTodoVerbSurfaceZeroDelta` → exit 0, `--- PASS: TestTodoVerbSurfaceZeroDelta (0.00s)`.
   Commit `8d302004ed`.

Also flip-affected and updated earlier in the run (plan §C P2 path, commit `d9ae56471b`):
`TestDefaultCardDeciderIsDeterministic` (want → the absent-read default per REQ-FDM-002),
`TestFactoryNextStillRefusedBehindADrivenRow` and
`TestFactoryNextRefusesAgainAfterLaterAssign` (explicit serial queue item behind the t1453
driven row — a row's serial-ness reads through the queue per REQ-FDM-001). All three
verified `--- PASS` in the completed verbose suite after the isolated verification batch.

### E3 — coverage (scoped, per the coordinator prohibition on further full-suite runs; full-suite coverage deferred to CI per AGENTS.md §4's scope-verification principle)

- Scoped selector over the touched families:
  `go test -cover -run 'Classification|TodoClassify|TodoAdd|FactorySelection|SerialExclusivity|SD_AC015|GTDAllTodoVerbs|LiveReaders|SHAValidation|TodoVerbSurface' ./internal/factory ./internal/cli`
  → exit 0, `ok ... coverage: 10.3% of statements` (factory, 16.951s), `ok ... coverage:
  10.2% of statements` (cli, 36.048s) — package-wide denominators under a scoped selector;
  the meaningful read is per-function on the changed code (coverprofile,
  `/private/tmp/moai-t1737/e3-profile.out`):
  - `internal/factory/classification.go` — `DefaultCardClassification` 100%,
    `AbsentCardClassification` 100%, `ValidateCardClassification` 100%,
    `EffectiveCardClassification` (the seam) 100%.
  - `internal/cli/todo_classify_cmd.go` — the verb RunE body 93.3%, the conflict helper
    76.0%, the recovery-advice helper 75.0%.
  - `internal/cli/factory_card.go` — `factoryQueueClassification` 77.8%.
- The ≥85% package gate is a full-suite measurement and is deferred to the repository CI run
  on the integration branch (that run owns the repository-wide verdict; it is PENDING at
  this report's completion).

### E4 — subagent boundary

- Command: `grep -rn "AskUserQuestion" <all 8 touched files>` → no output, exit 1 (zero hits).

### E5 — spec lint

- Command: `moai spec lint --strict SPEC-FACTORY-DEFAULT-MODE-001` → `0 error(s), 25
  warning(s)`, exit 0. Warnings are advisory `VacuousTestAssertion`-family items on the
  plan-phase artifacts (acceptance.md §D.0 green-path cells), pre-existing from plan close,
  outside run-phase scope.

### E6 — refusal byte-identity, observed by test

- `TestTodoClassifyRefusals` and the conflict subtest hash the queue through
  `readBacklogBytes` before and after every refusal (`string(after) != string(before)`
  fails the test) — observed PASS above for all 8 refusal arms: held, dropped, archived
  (terminal), absent id, empty reason, out-of-set mode, unknown `--decider` flag, and the
  serial-slot conflict. Byte-identity is test-observed, not inspection-asserted.

### E7 — fallback notice count, observed by test

- `TestTodoAddDeciderFailureFailsafeSerial` counts `strings.Count(errOut, "classification
  decider unavailable") == 1`, requires the notice to still name `mode=serial`, and requires
  the recorded classification to be mode serial under decider default — observed PASS above.
  Exactly one notice, fail-safe serial, admission not refused.

### MX tag changes (run phase)

- Added `@MX:ANCHOR` + `@MX:REASON` + `@MX:SPEC` on the new exported
  `AbsentCardClassification` (fan_in 5: the seam, the default decider, factoryQueueClassification's
  two absent branches, the status degrade fallback).
- Refreshed the doc-comment text around the `EffectiveCardClassification` ANCHOR (tag and
  reason unchanged; the comment now names the parallelizable absent-read default and the
  split). No tags removed. classification.go carries 3 ANCHORs — at its per-file limit.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-20T07:50:00Z (UTC)
run_commit_sha: "8d302004ed"
run_status: complete
ac_pass_count: 11
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "not-run (lane worktree, no shared-checkout commit; branch WT-spec-factory-default is card-local)"
l44_post_push_fetch: "not-run (lane does not push — the batch push is the leader's bulk action)"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  go_build_local: "pass (compilation via go vet + every go test run in this report; GOOS=windows matrix owned by CI)"
  ci_matrix: PENDING
total_run_phase_files: 15
m1_to_mn_commit_strategy: "per-milestone commits M1=1aec6f8986 M2=b8e2f92713 M3-guards=82462814da, then leader-approved close-out repairs: flip-affected baselines d9ae56471b, surface walks a05824c1c6, golden 606de2c47b, sha-strip 7916bb2859, surface declaration 8d302004ed"
full_suite_coverage: "deferred to CI (coordinator prohibition on further full-suite runs; AGENTS.md §4 scope-verification) — scoped per-function coverage recorded in §E.2 E3"
red_evidence: "M1/M2 RED verbatim in §E.2 (derivation serial vs parallelizable; verb-absence); AC-FDM-004/-008/-009/-010 are documented regression-guards/guard-additions with no RED claimed per acceptance.md §D.0 and plan §F M3"
known_dispositioned_failures: "6 final FAIL entries in the completed verbose cli suite — all dispositioned in §E.2 E2 (1 pre-existing t1869-attributed, NOT a pass; 5 leader-approved card-scope repairs, all GREEN)"
```

Evidence files (machine-local scratch, quoted verbatim where deciding):
`/private/tmp/moai-t1737/{ev-fdm-12-baseline.txt, ev-fdm-13-14-baseline.txt, m1-red-factory.txt, m1-red-cli.txt, m2-red.txt, m2-green2.txt, m3-guards.txt, e1-ac-matrix.txt, e2-full-suites.txt, e2-factory.txt, e2-cli.txt, e2-cli-v.txt, repair-a-scoped.txt, e3-profile.out}`.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
