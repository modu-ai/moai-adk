# SPEC-AUDIT-CEILING-REPAIR-001 — progress

status: implemented
card: t1560 (lane-2, run tmhxo0) · tree .moai/worktrees/t1560 · branch WT-internal-runtime-audit @ 903ccd028
evidence: this file (§G carries any ceiling records for THIS spec if the engine ever evaluates it) + .moai/reports/t1560/

## §E.1 Plan-phase Audit-Ready Signal

plan_status: pending-plan-audit
plan_complete_at: 2026-10-07
비고: plan artifacts (spec.md · plan.md · acceptance.md · decision-index.md · references/t1500-seal-excerpt.md) authored 2026-10-07 by manager-spec (card t1560, Tier M). Revision v0.3.0 (2026-10-07): plan-audit-1 repairs integrated (verdict FAIL 0.81, `.moai/reports/t1560/plan-audit-1.md`) + leader ruling #2 (D4 fold-in). `plan_status: audit-ready` is set after the plan-auditor verdict; iteration 2 re-audits delta-scoped to the fix_scope anchors (mapping in plan.md §H).

## §E.2 Run-phase Evidence

### M1 — the eight reproduction REDs (tree `8233644f0`, code-identical to the RED baseline `903ccd028`)

Authored at M1 against the still-pristine engine per the family convention
("RED is a new test — E8 evidence required"); MP-8 basis restated: no repro
test existed at plan phase by design, so the plan-phase `go test -run
'^Test…$'` selectors matched zero tests (`ok … [no tests to run]`, exit 0) —
the expected plan-phase state, never a RED observation and never a pass.
Every RED below ran on `8233644f0` with only test files added (engine files
untouched); `go build`/`go vet` clean and the AC-ACR-012 sealed-surface
selection green before the tests were authored (baseline at 10:26 KST).

| AC | Test | Command (all with `./internal/runtime/`, `-count=1`) | Verbatim RED output (assertion line) | Exit |
|---|---|---|---|---|
| AC-ACR-001 (D1) | `TestPreviousAuditedSHALatestLegacyRound` | `go test -run '^TestPreviousAuditedSHALatestLegacyRound$' -v` | `audit_counter_review_test.go:240: previous audited SHA "", want sha-rev2 (largest round strictly below the legacy latest)` | 1 |
| AC-ACR-002 (D1) | `TestPreviousAuditedSHAMixedFamilyPrior` | `go test -run '^TestPreviousAuditedSHAMixedFamilyPrior$' -v` | `audit_counter_review_test.go:264: previous audited SHA "", want sha-iter2 (convention prior below the legacy latest)` | 1 |
| AC-ACR-004 (D1 e2e) | `TestEvaluateCeilingLegacyLatestDeltaGranted` | `go test -run '^TestEvaluateCeilingLegacyLatestDeltaGranted$' -v` | `audit_ceiling_test.go:1216: outcome &{Outcome:hold Reasons:[plan-audit ceiling reached (round count 2 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: …] Blocked:true Debts:[]} (override false), want nil — the delta round is granted: count 2 reaches the tier ceiling 2 and the eligibility conditions hold` | 1 |
| AC-ACR-005 (D2) | `TestPersistOutcomeDebtAdmitCarriesDebtInventory` | `go test -run '^TestPersistOutcomeDebtAdmitCarriesDebtInventory$' -v` | `audit_ceiling_test.go:1253: §G record carries no debts= token: - 2026-10-07T10:31:23Z SPEC-ACE-ENG-001 ceiling-outcome outcome=debt-admit reasons="plan-audit ceiling reached (round count 1 >= tier ceiling 1); the verdict fails admission on the label alone and debt-admits — findings recorded as PASS-WITH-DEBT debts, entry admitted without a question (REQ-ACE-004)" evidence=…` | 1 |
| AC-ACR-013 (D3) | `TestAppendProgressRecordConcurrentSurvival` | `go test -run '^TestAppendProgressRecordConcurrentSurvival$' -count=3 -v` | run 1/3: `audit_ceiling_test.go:1303: pre-existing progress content lost:` followed by a progress.md holding only records 18/13/08/20 under the §G heading; runs 2/3 failed the same way (run 2's content additionally interleaved a partial `fusal Record` fragment — torn write) | 1 |
| AC-ACR-014 (D4 a-c) | `TestCountAuditRoundsOverflowOwnRound` | `go test -run '^TestCountAuditRoundsOverflowOwnRound$' -v` | `(a) audit_counter_review_test.go:306: count 1, want 2 (the base and the overflow suffix are two rounds)` · `(b) audit_counter_review_test.go:325: count 1, want 2 (the base never merges with iter1)` · `(c) audit_counter_review_test.go:344: count 1 (sources 3), want 3 (1 normal + 2 unparseable)` | 1 |
| AC-ACR-014 (D4 e) | `TestPreviousAuditedSHABaseRoundBaseline` | `go test -run '^TestPreviousAuditedSHABaseRoundBaseline$' -v` | `audit_counter_review_test.go:410: previous audited SHA "", want sha-base (the base is round 0, the previous audited round)` | 1 |
| AC-ACR-015 (C3) | `TestReqACSetsUnchangedReadsAcceptance` | `go test -run '^TestReqACSetsUnchangedReadsAcceptance$' -v` | `audit_ceiling_test.go:1413: an acceptance-only AC id rename verified as unchanged — the delta must be refused (fail-closed)` | 1 |
| AC-ACR-016 (D3) | `TestAppendProgressRecordInsertsAtSectionEnd` | `go test -run '^TestAppendProgressRecordInsertsAtSectionEnd$' -v` | `audit_ceiling_test.go:1340: record did not land inside the §G block:` (the record appended past the following `## §E.2` heading) | 1 |

**AC-ACR-003 keep-green guard (pre-fix run)**: `go test -run
'^TestPreviousAuditedSHAUnparseableLatestStaysEmpty$' -v ./internal/runtime/`
→ `--- PASS: TestPreviousAuditedSHAUnparseableLatestStaysEmpty (0.00s)`,
exit 0. The preserve arms of `TestCountAuditRoundsOverflowOwnRound`
(`bare_base_alone_counts_1_and_stays_latest`,
`explicit_zero_keeps_own_round_parity`) also passed inside the pre-fix RED
batch — only the (a)/(b)/(c) subtests failed.

**Measurement note (arm (e) count parenthetical)**: acceptance.md's arm (e)
RED cell says "Count=2 but the scan skips the base". The mechanically
measured count for the arm-(e) fixture (base + iter1) at `8233644f0` is
**1**, not 2 — the same base/iter1 collapse arm (b) documents as "RED today:
1" (both files parse to round 1 and dedupe into `seen[1]`). The Count=2
figure could not be reproduced on this tree; the RED was re-anchored to the
stated previous-baseline reason (the test asserts the previous-audited-SHA
behavior before the count assertions). The defect class, the fix, and the
green state are unaffected.

Full verbatim outputs: `.moai/state/verify/t1560/m1-red-batch.txt`,
`m1-red-e2e.txt`, `m1-red-concurrent.txt`, `m1-red-base-baseline.txt`
(machine-local scratch, this run).

### M2 — D1 + D4 + C3 fixes, REDs flip green (commit `c049810a2`)

`go test -count=1 -run '<M2 selectors + sealed counter/delta tests>' -v
./internal/runtime/` → exit 0, all PASS:
`TestPreviousAuditedSHALatestLegacyRound`, `TestPreviousAuditedSHAMixedFamilyPrior`,
`TestPreviousAuditedSHAUnparseableLatestStaysEmpty` (guard stays green),
`TestCountAuditRoundsOverflowOwnRound` (all five subtests),
`TestPreviousAuditedSHABaseRoundBaseline`, `TestReqACSetsUnchangedReadsAcceptance`,
`TestEvaluateCeilingLegacyLatestDeltaGranted`, plus the sealed
`TestCountAuditRounds`, `TestCountAuditRoundsDistinctN`,
`TestCountAuditRoundsLegacyPlanAuditNumbered`,
`TestPreviousAuditedSHALegacyPriorRound`,
`TestCountAuditRoundsExactHeaderAttribution`, `TestDeltaGitHelpers` —
`ok github.com/modu-ai/moai-adk/internal/runtime 9.378s`.

### M3 — D2 + D3 fixes, REDs flip green (commit `38272872c`)

`go test -count=1 -run '<persistence selectors>' -v ./internal/runtime/` →
exit 0, all 17 PASS (the D2/D3 repro REDs, the escaping RED, the two
preserve tests, and the pre-existing ceiling/trail/override family).
Concurrency discipline: `go test -count=5 -race -run
'^TestAppendProgressRecordConcurrentSurvival$|^TestAppendProgressRecordInsertsAtSectionEnd$'`
→ 10/10 `--- PASS`, `ok … 1.334s` (judged over repeated runs, not one
green).

### M4 — consistency notes (AC-ACR-010/011) and re-measurement

**AC-ACR-010 (t1500 seal, read-and-note — non-contradiction)**: the
SPEC-local verbatim excerpt `references/t1500-seal-excerpt.md` (§SEAL/§PUSH,
provenance header retained) re-read at run phase against the post-repair
diff. The seal froze card t1500's engine work at `3d7215b72` (22/22 AC +
card-review repairs). Non-contradiction holds: this repair EXTENDS the
dual-family contract the sealed card-review F4 test
(`TestPreviousAuditedSHALegacyPriorRound`, legacy-as-PRIOR) established to
its uncovered face (legacy-as-LATEST, `TestPreviousAuditedSHALatestLegacyRound`)
— the sealed test itself stays green (M2 run above), the seal's resume
points (re-review recording, factory stage path, leader push batch) are
untouched by this diff, and no sealed behavior was rewritten (the
heading-absent and §G-last record shapes are byte-identical to the sealed
append; only the §G-followed-by-a-section insertion position and the
base/overflow round identity changed, per REQ-ACR-008/009).

**AC-ACR-011 (t1538 sealed resume point, read-and-note — non-contradiction)**:
`git show origin/WT-t1538-factory-recovery:.moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/progress.md`
§봉인 기록 re-read at run phase: the remaining-gate inventory is (1) the
factory mirror-path P1 (dispatch store + binding update in one lock
section) and (2) the `^TestReview` overlay reproduction family. Both live
in factory dispatch / review-gate code — disjoint from
`internal/runtime/audit_ceiling.go` (E5 scope grep below confirms this
diff's only engine files are the four §C files).

**Re-measurement (this run, this tree, HEAD `38272872c` + M4 docs)**:

- E3 affected-package family: `go test -race -count=1 -timeout 30m
  ./internal/runtime/...` → `ok github.com/modu-ai/moai-adk/internal/runtime
  12.753s` + `ok … internal/runtime/gobin 1.415s` (full package, race on).
  `go test -count=1 -timeout 30m ./internal/auditverdict/...` → `ok … 0.289s`
  (untouched package stays green).
- AC-ACR-012's named sealed tests all pass inside the full-suite run and
  passed an explicit named selection at M2 (list above).
- Consolidated green run of all thirteen repair tests: `go test -count=1
  -run '<13 repair selectors>' -v ./internal/runtime/` → exit 0, 13/13
  `--- PASS` (`.moai/state/verify/t1560/m4-green-all.txt`).
- E4 lint/format: `go vet ./internal/runtime/... ./internal/auditverdict/...`
  clean; `golangci-lint run ./internal/runtime/...` → `0 issues.`; `gofmt -l`
  on both packages → empty.
- E5 scope grep: `git diff --name-only 903ccd028..HEAD` under `internal/`
  → exactly `internal/runtime/audit_ceiling.go`,
  `audit_ceiling_test.go`, `audit_counter.go`,
  `audit_counter_review_test.go` — no `auditverdict`, `DeltaEligible`, or
  JSON-path (`RecordCeilingOutcome`) changes.
- E6 record grammar spot-check: a no-debt outcome's §G line is built by the
  unchanged Sprintf and carries no suffix when the inventory is empty;
  `TestPersistOutcomeNoDebtRecordUnchanged` verifies the pre-repair grammar
  (`- <ts> <spec> ceiling-outcome outcome=… reasons=… evidence=…`, no
  `debts=`) for pass-through/hold, and the refusal/override shapes, green
  pre- and post-fix.
- @MX tag report: no tag changes — no new exported functions, no new
  goroutines or dangerous patterns (the §G mutex is standard in-process
  serialization per plan §D.7), no fan_in changes on tagged functions.
  Existing tags (EvaluateCeiling ANCHOR, CountAuditRounds NOTE) unchanged.

### F2 repair addendum (sync-audit-1, blocking — umask regression in the D3 atomic replace)

RED observed pre-fix at HEAD `8a8c97d2d` (new test,
`TestAppendProgressRecordNewFileModeAppliesUmask`, `//go:build darwin || linux`,
`syscall.Umask(0o077)`):
`audit_ceiling_umask_test.go:31: new progress.md mode 0644, want 0600 (0644
with the umask applied — the pre-repair os.WriteFile semantics)` — matching
the verdict's `/tmp` probe byte-for-byte. Fix at the CALL SITE
(`appendProgressRecord`): a not-yet-existing progress.md is pre-created
empty through `os.WriteFile(path, nil, 0o644)` — the kernel applies the
umask at create time, the exact pre-repair semantics — and the atomic
replace then stat-preserves that mode; `config/atomicfile.Write` itself is
UNTOUCHED (11 other callers keep their verbatim-defaultMode semantics —
zero caller impact). The existing-file arm
(`TestAppendProgressRecordExistingFileModePreserved`: 0600 stays 0600
through the replace) passed pre- and post-fix.

Mode matrix (new-file progress.md):

| umask | pre-repair `os.WriteFile` 0644 | broken (`atomicfile.Write` 0644 verbatim) | repaired (call-site pre-create) |
|---|---|---|---|
| 0022 | 0644 | 0644 | 0644 |
| 0077 | 0600 | 0644 (the regression) | 0600 |

Existing-file progress.md: mode preserved through the replace, both before
and after (0600 stays 0600) — unchanged arm.

Re-measurement: `go test -race -count=1 -timeout 30m ./internal/runtime/...`
→ `ok … 37.711s`; umask + §G pairs `-count=3 -race` → `ok … 2.615s`; vet
clean; `golangci-lint run ./internal/runtime/...` → `0 issues.`. F1/F3
untouched per leader-pending ruling; F3's aside on `persist.go`
`atomicWrite` (new files stay 0600) is observed and NOT acted on this round.

### F1 repair record (sync-audit-1, leader ruling #2 adopted — legacy-branch Atoi clamp)

RED observed pre-fix: `audit_counter_review_test.go:443: count 2, want 3
(two overflow legacy suffixes fail-count their own rounds + one normal)` —
the discarded Atoi range error clamped both overflow legacy files into
`seen[math.MaxInt]` (one merged round) with one of them elected `LatestPath`
over the legit round. Fix (audit_counter.go legacy branch): the Atoi error
routes to the fail-counted path — own round, never `LatestPath` — the same
REQ-ACR-009 semantics the convention-family overflow fix already applies;
legit small legacy numbers unchanged. Regression test:
`TestCountAuditRoundsLegacyOverflowOwnRound` (two distinct overflow legacy
suffixes + one normal → count 3, `LatestPath` = `-review-1.md`). Whole
touched-function family re-run green (18 tests: TestCountAuditRounds*,
TestPreviousAuditedSHA*, TestRoundReportDirs*, TestCountPlanAuditRounds*,
the D1 end-to-end), full package `go test -race -count=1` → `ok … 27.477s`,
vet clean, golangci-lint `0 issues.`. `evidenceRoundOf` already respected
the range error (M2), so `previousAuditedSHA` was never fail-open here —
the counter and the baseline helper now agree.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-10-07
run_commit_sha: 38272872c
run_status: complete
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-applicable (card-dedicated worktree, single-writer lane; branch/HEAD re-read before each commit per the staleness rule)
l44_post_push_fetch: pending (push is the leader's batch — lane does not push)
new_warnings_or_lints_introduced: 0 (golangci-lint 0 issues; go vet clean; gofmt clean)
cross_platform_build: not-run-in-lane (CI matrix owns the darwin/windows verdict; `go vet` compiled both engine packages clean locally)
total_run_phase_files: 4 (the plan §C set: audit_ceiling.go, audit_counter.go, audit_counter_review_test.go, audit_ceiling_test.go) + progress.md/spec.md run-phase records
m1_to_m4_commit_strategy: one commit per milestone (M1 932f1bed0 REDs · M2 c049810a2 D1+D4+C3 · M3 38272872c D2+D3 · M4 this record); CI run on the integrated branch owns the repository-wide test verdict — PENDING at report time
비고: AC-ACR-004's tier-ceiling fixture and AC-ACR-015's git fixture measured
slower under the race detector (2-3 s each) — no flake observed across the
repeated concurrency runs. The arm-(e) Count=2 parenthetical in
acceptance.md could not be reproduced (measured 1, see §E.2 measurement
note) — recorded as an observation, not an AC failure.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop owns this section.>_

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: complete (manager-docs sync-phase deliverables done; spec.md transitions to `implemented` on this sync commit — `completed` held for its own gate per the dispatching leader's instruction, card t1560)
- sync_complete_at: 2026-10-07
- sync_commit_sha: 14a6956a8 (sync commit `docs(SPEC-AUDIT-CEILING-REPAIR-001): sync-phase CHANGELOG entry, implemented transition, §E.4 signal (card t1560)`; backfilled in this follow-up commit — the sync commit could not cite its own hash, per the D3 backfill window, spec-frontmatter-schema § SHA placeholder backfill exemption)
- changelog_entry_position: CHANGELOG.md `## [Unreleased]` > `### Fixed` — first entry (SPEC-AUDIT-CEILING-REPAIR-001; five repair behaviors: D1 legacy-family latest delta round, D2 debt inventory in §G/trail records, D3 atomic §G append + insertion position, D4 round counting base/overflow identity, REQ-ACR-010 both-definition-files delta gate)
- frontmatter_status_transitions:
  - spec.md: in-progress → implemented (this sync commit)
  - plan.md: no `status:` field (Artifact Statelessness) — `updated:` confirmed 2026-10-07, no byte change needed
  - acceptance.md: no `status:` field (Artifact Statelessness) — `updated:` confirmed 2026-10-07, no byte change needed
  - progress.md: status line in-progress → implemented (this sync commit)
- updated_field_refresh: 2026-10-07 (spec.md/plan.md/acceptance.md already dated 2026-10-07 from plan/run phase — confirmed present, no byte change)
- b12_self_test_a (pre-emission grep): `grep -c 'SPEC-AUDIT-CEILING-REPAIR-001' CHANGELOG.md` = 0 (pre-emission, exit 1) → no duplicate entry, emission safe
- b12_self_test_b (AC count match): ac_source=acceptance.md (tier: M); counter live=18 — 16 declared criteria AC-ACR-001..016 (`grep -c '^- \*\*AC-ACR-'` = 16 matrix rows) + 2 prose-example tokens `AC-R-001`/`AC-R-002` at acceptance.md:218 inside AC-ACR-010's fixture prose (example identifier shapes, declare no criterion, ownerless); CHANGELOG entry cites "16 acceptance criteria AC-ACR-001..016" matching the declared set; no AMBIGUOUS halt (both extra tokens uniformly unmarked → live by the mechanical rule, excluded from the cited count by inspection); `[REF]` marker placement is an acceptance.md body edit — manager-spec's surface, recorded as observation, not performed here
- b12_self_test_c (file path verification): paths claimed in the CHANGELOG entry verified — `.moai/specs/SPEC-AUDIT-CEILING-REPAIR-001/spec.md`, `.moai/specs/SPEC-AUDIT-CEILING-REPAIR-001/progress.md` (§E.2 evidence section present)
- canary_compliance_check: N/A (this SPEC defines no forward-looking policy with its own sync tests)
- mx_tag_validation: sync sub-step — no tag changes (§E.2 M4: no new exported functions, no new dangerous patterns; EvaluateCeiling ANCHOR + CountAuditRounds NOTE unchanged)
- 비고: progress.md carries a duplicate empty `## §E.3` placeholder below the filled §E.3 (lines 162-164) — §E.2/§E.3 are manager-develop's surface, left untouched per ownership; era classification reads literal heading presence so the duplicate is inert for lint/audit.
