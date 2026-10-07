---
id: SPEC-AUDIT-CEILING-REPAIR-001
title: "Audit-ceiling engine defect repair — acceptance criteria"
version: "0.1.0"
created: 2026-10-07
updated: 2026-10-07
---

# acceptance.md — SPEC-AUDIT-CEILING-REPAIR-001

## §A Purpose and RED-first contract

Every defect AC below is two-celled: the RED-now cell names the command and
the reason it fails on unmodified main `903ccd028`; the green-path cell names
the milestone that flips it (M2 for D1, M3 for D2). The verbatim RED output
and exit code are recorded in `progress.md` §E.2 when run phase executes M1 —
this file pins the expectation, the run records the observation. The RED
baseline tree is `903ccd028` (origin/main tip this branch was cut from).

## §D AC Matrix

- **AC-ACR-001 (D1 RED, release-blocking)** — Given a report directory holding
  a pure legacy stream `<SpecID>-review-1.md` (audited_sha `sha-rev1`),
  `-review-2.md` (audited_sha `sha-rev2`), `-review-3.md` (audited_sha
  `sha-rev3`, the counter-selected latest), When
  `CountAuditRounds` feeds `previousAuditedSHA`, Then it returns `sha-rev2`
  (the largest round strictly below the latest, deduped across families).
  RED: `go test -run '^TestPreviousAuditedSHALatestLegacyRound$' ./internal/runtime/`
  fails on `903ccd028` with got `""` want `sha-rev2` — the `latestN <= 0`
  early exit aborts on the legacy-family latest name. Green at M2. Test:
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
  the fix. Test: `TestPreviousAuditedSHAUnparseableLatestStaysEmpty`.
  Regression-guard (passes before and after; pins what the fix must not
  break).

- **AC-ACR-004 (D1 end-to-end, release-blocking)** — Given a git-initialized
  temp project (pattern of `TestDiffInsideAnchorsHunkScope`) with a committed
  spec.md, two legacy-family rounds (`-review-1.md` audited at the base
  commit, `-review-2.md` latest, carrying a `fix_scope:` anchor line and
  audited at a later commit whose diff touches only `.moai/reports/` paths),
  and harness config with tier ceiling 1 and `auto_delta_rounds: 1`, When
  `EvaluateCeiling` runs, Then it returns no ceiling outcome (nil — the delta
  round is granted). RED: on `903ccd028` the same fixture returns a final-hit
  outcome (non-nil) because `prevSHA` is empty. Green at M2. Test:
  `TestEvaluateCeilingLegacyLatestDeltaGranted` (`audit_ceiling_test.go`).

- **AC-ACR-005 (D2 RED, release-blocking)** — Given the ceiling fixture
  driving a debt-admit outcome with at least one enumerated debt
  (`TestCeilingPolicyDebtAdmit`'s construction: `- debt: D1 dispose_in=run
  ...`), When `EvaluateCeiling` persists the outcome, Then the SPEC's
  `progress.md` §G record line for that outcome carries the admitted debt's
  ID and `dispose_in` (recoverable tokens) and remains a single line. RED:
  `go test -run '^TestPersistOutcomeDebtAdmitCarriesDebtInventory$'
  ./internal/runtime/` fails on `903ccd028` — the record line serializes
  kind/outcome/reasons/evidence only. Green at M3. Test:
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
  t1500 seal record
  (`.moai/worktrees/t1500/.moai/reports/t1500/lane28-wait-claude-gate.md`
  §SEAL/§PUSH) and records in progress.md a non-contradiction note: the
  repair extends the dual-family contract the sealed card-review F4 test
  established and keeps the sealed test surface green. Read-and-note only —
  no re-derivation.

- **AC-ACR-011 (consistency read b, process gate)** — The run phase reads the
  t1538 sealed resume point (`origin/WT-t1538-factory-recovery`
  `.moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/progress.md` §봉인 기록)
  and records in progress.md a non-contradiction note: the remaining-gate
  inventory (mirror-path P1, `^TestReview` overlay family) is disjoint from
  `internal/runtime/audit_ceiling.go`.

- **AC-ACR-012 (no-regression suite, release-blocking)** —
  `go test -timeout 30m ./internal/runtime/...` is green, explicitly
  including the pre-existing sealed-surface tests: `TestPreviousAuditedSHALegacyPriorRound`,
  `TestCountAuditRoundsLegacyPlanAuditNumbered`, `TestCeilingPolicyDebtAdmit`,
  `TestCeilingDeltaEligibility`, `TestAuditTrailAppend`,
  `TestDiffInsideAnchorsHunkScope`, `TestEvaluateCeilingUnknownPolicyFailClosed`
  — the t1500 sealed behavior survives the repair unchanged. `go vet` and
  `golangci-lint` clean on the package.

- **AC-ACR-013 (D3 RED, release-blocking — "ProgressAppend atomicity")** —
  Given a specDir whose progress.md already carries content and the §G
  heading, When 24 concurrent `appendProgressRecord` calls run (goroutines
  joined via WaitGroup), Then all 24 record lines are present exactly once,
  the §G heading appears exactly once, and no pre-existing line is lost.
  RED: on `903ccd028` the unlocked read-modify-write race loses records —
  the turn-end codex review gate measured 24 parallel calls → 7 survivors
  with the pre-existing heading lost (leader-cited evidence material; the
  test's own RED run records its verbatim output in §E.2). Green at M3
  (the persistence milestone). Test:
  `TestAppendProgressRecordConcurrentSurvival` (`audit_ceiling_test.go`).

## §D.1 Severity classification

| AC | Severity | Rationale |
|---|---|---|
| AC-ACR-001, 004, 005, 012, 013 | Release-blocking | RED-first defect proofs + sealed-surface no-regression |
| AC-ACR-002, 006, 007, 008, 009 | Normal | Contract completion arms of the two fixes |
| AC-ACR-003 | Regression-guard | Pins preserved fail-closed semantics |
| AC-ACR-010, 011 | Process gate | Read-and-note consistency obligations from the dispatch |

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

## §D.4 Indirect verification

- AC-ACR-004 verifies the D1 fix through the production entry
  (`EvaluateCeiling`) rather than the helper alone — the defect's
  user-visible effect (final hit instead of delta round) is what flips.
- AC-ACR-005/006 verify D2 through the persisted bytes on disk (progress.md
  and the trail file), not through in-memory structs.

## §D.5 Quality gates (TRUST 5)

- **Tested**: AC-ACR-012 suite green; both RED-first evidences recorded.
- **Readable**: new tests follow the existing fixture/comment style of their
  host files (`newCeilingFixture`, `writeAuditFixture`, the F-series comment
  convention of `audit_counter_review_test.go`).
- **Unified**: gofmt + golangci-lint clean.
- **Secured**: anchored regexes on untrusted names; debt-field escaping
  (AC-ACR-007); best-effort persistence preserved (AC-ACR-009).
- **Trackable**: Conventional Commits naming the SPEC and card t1560;
  `Authored-By-Agent:` trailer per transition ownership.

## §D.6 Closure gates (Definition of Done)

1. All ACs green with §E.2 evidence; the three RED cells observed and recorded.
2. Consistency notes (AC-ACR-010/011) recorded in progress.md.
3. Affected-package measurement green (`go test -timeout 30m
   ./internal/runtime/...`); vet + lint clean.
4. `spec.md` frontmatter transitions `draft → in-progress` at the first
   run-phase commit (manager-develop, ownership matrix).
5. No commit on this card's tree from this plan phase — the lane reviews and
   commits the plan artifacts.

## §D.7 Forward-looking checks (non-blocking)

- A future machine parser of §G records would formalize the `debts=` grammar
  (encoding contract then becomes load-bearing — decision-index Q2 records
  today's default).
- Trail-line parity for the required-backend path stays as-is: those records
  carry no debts by construction.
