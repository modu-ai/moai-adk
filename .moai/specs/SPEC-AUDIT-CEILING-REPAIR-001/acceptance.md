---
id: SPEC-AUDIT-CEILING-REPAIR-001
title: "Audit-ceiling engine defect repair — acceptance criteria"
version: "0.1.0"
created: 2026-10-07
updated: 2026-10-07
---

# acceptance.md — SPEC-AUDIT-CEILING-REPAIR-001

## §A Classification and RED/GREEN discipline

Severity classes: **RB** = release-blocking, **RG** = regression-guard, **PG**
= process gate. Following the engine family convention (SPEC-AUDIT-CEILING-001
`acceptance.md` §A:11-18): every RB criterion below carries the explicit
per-criterion declaration **"RED is a new test (E8 evidence required)"** — the
repro tests are M1 deliverables, authored at run phase against the still-
pristine `903ccd028` code, and each RED (verbatim stdout + exit code) is
observed and recorded in `progress.md` §E.2 at M1 per
manager-develop-prompt-template §E8. No repro test exists at plan phase by
design, so a plan-phase `go test -run '^Test…$'` that returns
`ok … [no tests to run]` (exit 0) is the expected plan-phase state and is
never a RED observation and never a pass — the plan-audit-1 MP-8 measurement.
RG criteria pass both before and after the fix and pin what the repair must
not break; PG criteria are read-and-note obligations with their own evidence
paths. Document-level RED baseline tree: `903ccd028` (the audited SPEC commit
`e93cbad45` is code-identical to it — plan-audit-1 measured the delta as
artifact-only).

## §D AC Matrix

- **AC-ACR-001 (D1 RED, release-blocking)** — Given a report directory holding
  a pure legacy stream `<SpecID>-review-1.md` (audited_sha `sha-rev1`),
  `-review-2.md` (audited_sha `sha-rev2`), `-review-3.md` (audited_sha
  `sha-rev3`, the counter-selected latest), When
  `CountAuditRounds` feeds `previousAuditedSHA`, Then it returns `sha-rev2`
  (the largest round strictly below the latest, deduped across families).
  RED: `go test -run '^TestPreviousAuditedSHALatestLegacyRound$' ./internal/runtime/`
  fails on `903ccd028` with got `""` want `sha-rev2` — the `latestN <= 0`
  early exit aborts on the legacy-family latest name. **RED is a new test
  (E8 evidence required).** Green at M2. Test:
  `TestPreviousAuditedSHALatestLegacyRound` (`audit_counter_review_test.go`).

- **AC-ACR-002 (D1 mixed families)** — Given a legacy-family latest
  `-review-3.md` and a convention-family prior `plan-audit-iter2.md`
  (audited_sha `sha-iter2`), When `previousAuditedSHA` runs, Then it returns
  `sha-iter2` — the prior-round scan's existing dual-family behavior works
  once the latest number resolves. Green at M2 (same test or sibling).

- **AC-ACR-003 (D1 preservation)** — Given a hand-built `RoundEvidence` whose
  `LatestPath` base name parses under neither family (e.g. `garbage.md`), When
  `previousAuditedSHA` runs, Then it returns `""` — the no-prior-round
  fail-closed semantics are preserved for genuinely unparseable names after
  the fix. **Preserve-behavior check, NOT a RED observation**: it passes on
  unmodified main (`latestN <= 0 → ""` IS the current behavior for
  unparseable names) and must keep passing — M1 runs it as a pre-and-post
  green guard, never in the RED obligation. Test:
  `TestPreviousAuditedSHAUnparseableLatestStaysEmpty`.
  Regression-guard (RG).

- **AC-ACR-004 (D1 end-to-end, release-blocking)** — Given a git-initialized
  temp project (pattern of `TestDiffInsideAnchorsHunkScope`) with a committed
  spec.md, two legacy-family rounds (`-review-1.md` audited at the base
  commit, `-review-2.md` latest, carrying a `fix_scope:` anchor line and
  audited at a later commit whose diff touches only `.moai/reports/` paths),
  and harness config with **tier ceiling 2** and `auto_delta_rounds: 1`, When
  `EvaluateCeiling` runs, Then it returns no ceiling outcome (nil — the delta
  round is granted: count 2 reaches the ceiling 2, and with the D1 fix
  `deltaOK` is true so `2 < 2+1` grants the delta at `audit_ceiling.go:110`).
  RED: `go test -run '^TestEvaluateCeilingLegacyLatestDeltaGranted$'
  ./internal/runtime/` fails on `903ccd028` — the same fixture returns a
  final-hit outcome (non-nil) because `prevSHA` is empty, `deltaOK` is false,
  and `2 < 3 && false` routes to the outcome ladder. (Fixture corrected at
  v0.3.0 per plan-audit-1 D2: the previous ceiling-1 form could not grant a
  delta even after the fix — `2 < 2` at `:110`.) **RED is a new test (E8
  evidence required).** Green at M2. Test:
  `TestEvaluateCeilingLegacyLatestDeltaGranted` (`audit_ceiling_test.go`).

- **AC-ACR-005 (D2 RED, release-blocking)** — Given the ceiling fixture
  driving a debt-admit outcome with at least one enumerated debt
  (`TestCeilingPolicyDebtAdmit`'s construction: `- debt: D1 dispose_in=run
  ...`), When `EvaluateCeiling` persists the outcome, Then the SPEC's
  `progress.md` §G record line for that outcome carries the admitted debt's
  ID and `dispose_in` (recoverable tokens) and remains a single line. RED:
  `go test -run '^TestPersistOutcomeDebtAdmitCarriesDebtInventory$'
  ./internal/runtime/` fails on `903ccd028` — the record line serializes
  kind/outcome/reasons/evidence only. **RED is a new test (E8 evidence
  required).** Green at M3. Test:
  `TestPersistOutcomeDebtAdmitCarriesDebtInventory` (`audit_ceiling_test.go`).

- **AC-ACR-006 (D2 trail parity)** — Given the same debt-admit evaluation,
  When the persistence call writes the machine-local trail line to
  `.moai/state/audit-enforcement.log`, Then that line also carries the debt
  inventory (same recoverable tokens) and remains a single line. Green at
  M3 (same test or sibling assertion).

- **AC-ACR-007 (D2 escaping, security)** — Given an admitted debt whose
  description (and ID) carry double quotes, semicolons, and a raw newline,
  When the outcome is persisted, Then both record lines remain single lines
  (the newline is escaped, not raw), and the debt's ID, `dispose_in`, and
  description are each recoverable from the record uncorrupted. Green at M3.
  Test: `TestPersistOutcomeDebtFieldsEscapedSingleLine`.

- **AC-ACR-008 (D2 negative)** — Given pass-through, hold, and split
  outcomes (no debts), When persisted, Then their §G and trail lines match
  the pre-repair grammar with no debt field emitted, and
  `RecordRequiredBackendRefusal` / `AcknowledgeRequiredBackend` record shapes
  are unchanged. Green at M3. Test:
  `TestPersistOutcomeNoDebtRecordUnchanged`. Regression-guard.

- **AC-ACR-009 (best-effort invariant)** — Given a record-write failure (e.g.
  an unwritable SpecDir), When a debt-admit evaluation persists, Then the
  returned outcome, override flag, and blocked flag are unchanged (debt-admit
  still admits with override=true) and only a stderr warning is emitted —
  persistence never flips the admission decision. Green at M3 (pin with a
  test arm or reuse the existing warning posture).

- **AC-ACR-010 (consistency read a, process gate)** — The run phase reads the
  t1500 seal record and records in progress.md a non-contradiction note: the
  repair extends the dual-family contract the sealed card-review F4 test
  established and keeps the sealed test surface green. Read-and-note only —
  no re-derivation. Cited source: the SPEC-local verbatim excerpt
  `references/t1500-seal-excerpt.md` (§SEAL/§PUSH, with provenance header —
  re-anchored at v0.3.0 per plan-audit-1 finding D4 because the original
  cross-tree path `.moai/worktrees/t1500/…` does not resolve from the card
  tree and the t1500 worktree is sweep-pending; the absolute primary-checkout
  path `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1500/.moai/reports/t1500/lane28-wait-claude-gate.md`
  is retained in the excerpt's provenance header).

- **AC-ACR-011 (consistency read b, process gate)** — The run phase reads the
  t1538 sealed resume point and records in progress.md a non-contradiction
  note: the remaining-gate inventory (mirror-path P1, `^TestReview` overlay
  family) is disjoint from `internal/runtime/audit_ceiling.go`. Provenance
  note: `SPEC-FACTORY-COMPLETION-RECOVERY-001` lives on
  `origin/WT-t1538-factory-recovery`, not in this tree's `.moai/specs/` —
  read via `git show origin/WT-t1538-factory-recovery:.moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/progress.md`
  (§봉인 기록; verified present on that ref at plan phase).

- **AC-ACR-012 (no-regression suite, release-blocking — no-RED character)** —
  `go test -race -count=1 -timeout 30m ./internal/runtime/...` is green,
  explicitly including the pre-existing sealed-surface tests:
  `TestPreviousAuditedSHALegacyPriorRound`,
  `TestCountAuditRoundsLegacyPlanAuditNumbered`, `TestCeilingPolicyDebtAdmit`,
  `TestCeilingDeltaEligibility`, `TestAuditTrailAppend`,
  `TestDiffInsideAnchorsHunkScope`, `TestEvaluateCeilingUnknownPolicyFailClosed`
  — the t1500 sealed behavior survives the repair unchanged. `go vet` and
  `golangci-lint` clean on the package. **Character (stated per plan-audit-1
  MP-8/D1): this criterion has NO RED cell by design** — it is a
  no-regression gate whose evidence is the full-family green run (`-race
  -count=1`: the D3 fix and the concurrency test are goroutine code); it is
  never satisfied by a RED observation and never recorded from one.

- **AC-ACR-013 (D3 RED, release-blocking — "ProgressAppend atomicity")** —
  Given a specDir whose progress.md already carries content and the §G
  heading, When 24 concurrent `appendProgressRecord` calls run (goroutines
  joined via WaitGroup), Then all 24 record lines are present exactly once,
  the §G heading appears exactly once, and no pre-existing line is lost.
  RED: `go test -run '^TestAppendProgressRecordConcurrentSurvival$'
  -race -count=1 ./internal/runtime/` fails on `903ccd028` — the unlocked
  read-modify-write race loses records; the turn-end codex review gate
  measured 24 parallel calls → 7 survivors with the pre-existing heading
  lost (leader-cited evidence material; the test's own RED run records its
  verbatim output in §E.2; concurrency is judged over repeated runs, never
  one green). **RED is a new test (E8 evidence required).** Green at M3
  (the persistence milestone). Test:
  `TestAppendProgressRecordConcurrentSurvival` (`audit_ceiling_test.go`).

- **AC-ACR-014 (D4 RED, release-blocking)** — **RED modes** (each observed
  failing on unmodified main `903ccd028`; the base report and
  unparseable-suffix files count as their own rounds; parsed-number dedupe
  is unchanged). RED commands: `go test -run
  '^TestCountAuditRoundsOverflowOwnRound$' ./internal/runtime/` (counter
  arms a-c) and `go test -run
  '^TestPreviousAuditedSHABaseRoundBaseline$' ./internal/runtime/` (arm e).
  - (a) base + overflow suffix — `plan-audit.md` +
    `plan-audit-iter99999999999999999999.md` → count 2. RED today: 1 — the
    Atoi range error at `audit_counter.go:111` leaves `n` at its initialized
    1, deduping into `seen[1]`; the overflow file is fail-counted and never
    becomes `LatestPath`.
  - (b) base vs numbered — `plan-audit.md` + `plan-audit-iter1.md` → count
    2. RED today: 1 — the convention branch initializes `n = 1` at
    `audit_counter.go:109` BEFORE the Atoi attempt, so base and numbered
    collapse into `seen[1]`.
  - (c) RED input pinned to OVERFLOW suffixes — one normal + two overflow
    files → 1+N = 3. RED today: 1, sources=3 (the leader's gate measured
    exactly this collapse at `903ccd028`).
  - (e) previous-round baseline (plan-audit-3 B1) — base + iter1 (iter1
    latest, an `audited_sha` line on both files) → `previousAuditedSHA`
    returns the base's audited_sha. RED today: `""` — Count=2 but the scan
    skips the base at `audit_ceiling.go:294` (`n >= latestN`; codex
    overlay-measured at HEAD: Count=2, previous=`""`).
  - **Preserve arms (non-RED — pass today and must keep passing):** the
    sealed parsed-number dedupe — `plan-audit-iter1.md` in two directories
    counts ONE round (pins the sealed `TestCountAuditRounds`: parsed n=1 vs
    parsed n=1 across directories, unchanged); unnumbered NON-report files
    (`iterX`-style) fail-count their own rounds today (sources=3, count=3 —
    codex-measured green via the `conventionUnnumbered` path); a bare
    `plan-audit.md` alone counts 1 and remains the selected `LatestPath`
    evidence.
  - **Semantics note (one line, leader-carried, non-blocking auditor
    note): round-counting semantics change — the base report becomes round
    0 (own counted round, earliest order, eligible as previous-round
    baseline and latest evidence) and overflow suffixes count as their own
    rounds; parsed-number dedupe is unchanged.**
  - **RED is a new test (E8 evidence required).** Green at M2. Tests:
    `TestCountAuditRoundsOverflowOwnRound` and
    `TestPreviousAuditedSHABaseRoundBaseline`
    (`audit_counter_review_test.go`).

## §D.1 Severity classification

| AC | Severity | Rationale |
|---|---|---|
| AC-ACR-001, 004, 005, 013, 014 | Release-blocking (RB) | RED-first defect proofs — each declared "RED is a new test (E8 evidence required)", observed at M1 |
| AC-ACR-012 | Release-blocking (RB, no-RED) | No-regression gate: evidence is the full-family green run (`-race -count=1`), never a RED cell |
| AC-ACR-002, 006, 007, 008, 009 | Normal | Contract completion arms of the fixes |
| AC-ACR-003 | Regression-guard (RG) | Preserve-behavior check — passes on unmodified main and must keep passing |
| AC-ACR-010, 011 | Process gate (PG) | Read-and-note consistency obligations from the dispatch |

## §D.2 Traceability

| REQ | ACs |
|---|---|
| REQ-ACR-001 | AC-ACR-001, AC-ACR-002, AC-ACR-004 |
| REQ-ACR-002 | AC-ACR-003 |
| REQ-ACR-003 | AC-ACR-004 |
| REQ-ACR-004 | AC-ACR-005, AC-ACR-006 |
| REQ-ACR-005 | AC-ACR-007 |
| REQ-ACR-006 | AC-ACR-008 |
| REQ-ACR-007 | AC-ACR-009 |
| REQ-ACR-008 | AC-ACR-013 |
| REQ-ACR-009 | AC-ACR-014 |

Non-REQ criteria (deliberate, repair-SPEC bookkeeping — stated per
plan-audit-1 D7): AC-ACR-010 and AC-ACR-011 are process gates (dispatch
consistency reads — no REQ parent); AC-ACR-012 is the no-regression suite
grounded in spec.md §D's "C5 posture inherited" constraint — it maps to the
repair's constraint layer, not to a single REQ.

## §D.3 Edge cases (covered by the AC set or explicitly observed)

1. Pure legacy stream (rev1/rev2/rev3) — AC-ACR-001.
2. Mixed families with legacy latest — AC-ACR-002.
3. Legacy latest as the ONLY round — no prior exists; `""` before and after
   (the fix changes nothing here; noted, no dedicated AC).
4. Zero-numbered legacy latest (`-review-0.md`) — non-positive number keeps
   the no-prior-round result (the `<= 0` guard is retained).
5. Unparseable latest name — AC-ACR-003.
6. Debt fields carrying `"`, `;`, newline, `@`, `=` — AC-ACR-007.
7. Multiple debts (2+) — inventory enumerates every debt (AC-ACR-005 arm).
8. Record-write failure — AC-ACR-009.
9. No-debt outcomes — AC-ACR-008.
10. Concurrent §G appends (the 24-call repro) — AC-ACR-013; judged over
    repeated runs, not one green (concurrency discipline).
11. §G heading absent at first append under a race — created exactly once,
    never duplicated (AC-ACR-013 arm).
12. Overflow iteration file (`-iter999…9`) never becomes `LatestPath` — the
    fail-counted path skips the best-N update, mirroring the existing
    unnumbered-file handling (AC-ACR-014 arm).

## §D.4 Indirect verification

- AC-ACR-004 verifies the D1 fix through the production entry
  (`EvaluateCeiling`) rather than the helper alone — the defect's
  user-visible effect (final hit instead of delta round) is what flips.
- AC-ACR-005/006 verify D2 through the persisted bytes on disk (progress.md
  and the trail file), not through in-memory structs.

## §D.5 Quality gates (TRUST 5)

- **Tested**: AC-ACR-012 suite green; the five RED-first evidences recorded.
- **Readable**: new tests follow the existing fixture/comment style of their
  host files (`newCeilingFixture`, `writeAuditFixture`, the F-series comment
  convention of `audit_counter_review_test.go`).
- **Unified**: gofmt + golangci-lint clean.
- **Secured**: anchored regexes on untrusted names; debt-field escaping
  (AC-ACR-007); best-effort persistence preserved (AC-ACR-009).
- **Trackable**: Conventional Commits naming the SPEC and card t1560;
  `Authored-By-Agent:` trailer per transition ownership.

## §D.6 Closure gates (Definition of Done)

1. All ACs green with §E.2 evidence; the five RED cells observed and recorded.
2. Consistency notes (AC-ACR-010/011) recorded in progress.md.
3. Affected-package measurement green (`go test -timeout 30m
   ./internal/runtime/...`); vet + lint clean.
4. `spec.md` frontmatter transitions `draft → in-progress` at the first
   run-phase commit (manager-develop, ownership matrix).
5. No further plan-phase artifact commits after the plan-audit verdict
   without a fresh audit — the landed plan commit `e93cbad45` is the audited
   baseline; the lane commits the v0.3.0 revision and iteration 2 re-audits
   the delta (rephrased per plan-audit-1 D9).

## §D.7 Forward-looking checks (non-blocking)

- A future machine parser of §G records would formalize the `debts=` grammar
  (encoding contract then becomes load-bearing — decision-index Q2 records
  today's default).
- Trail-line parity for the required-backend path stays as-is: those records
  carry no debts by construction.
