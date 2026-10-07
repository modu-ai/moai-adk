---
id: SPEC-AUDIT-CEILING-REPAIR-001
title: "Audit-ceiling engine defect repair — legacy-family latest-verdict resolution and debt-inventory persistence"
version: "0.1.0"
created: 2026-10-07
updated: 2026-10-07
---

# plan.md — SPEC-AUDIT-CEILING-REPAIR-001

## §A Context

- Card t1560, lane-2, run tmhxo0 — in-lane plan→run→sync, review lens
  `--security --deep`. Tree: `.moai/worktrees/t1560`, branch
  `WT-internal-runtime-audit` at origin/main `903ccd028`.
- The engine under repair is the SPEC-AUDIT-CEILING-001 deliverable (card
  t1500); this SPEC is a defect-repair delta against its landed behavior, not
  a redesign. Development mode: tdd (`quality.yaml`
  `constitution.development_mode: tdd`) — reproduction-first is both the mode
  and the card's HARD requirement.
- Defect surface verified in this tree at `903ccd028` (all line numbers read
  from source, not from the intake note):
  - D1: `internal/runtime/audit_ceiling.go:274-277` (latestN via
    `iterationOf` → `<= 0` early return), `:311-323` (`iterationOf` matches
    `conventionFile` only), `:281-300` (prior-round loop WITH the legacy
    fallback), `internal/runtime/audit_counter.go:46` (`conventionFile`),
    `:89,115-116,145-148` (legacy counting; a legacy file can be
    `LatestPath`), `:103-108` in audit_ceiling.go (prevSHA gate on the delta
    checks).
  - D2: `internal/runtime/audit_ceiling.go:461-471` (`persistOutcome` line
    grammar without debts; trail call without debts), `:137-153` (debt-admit
    rung, `Debts: fields.Debts` at `:144`), `internal/auditverdict/verdict.go:34-39`
    (`Debt{ID, DisposeIn, Description}`).
  - D3 (leader scope extension, 2026-10-07): `appendProgressRecord`
    (`internal/runtime/audit_ceiling.go:479-496`, the write at `:495`) —
    unlocked read-modify-write on progress.md, no atomic replace; measured
    by the turn-end codex review gate as 24 parallel §G record calls → 7
    survivors with the pre-existing heading lost (overlay repro on
    `903ccd028`).

## §B Known Issues

1. **D1** — with a legacy-family latest verdict, `previousAuditedSHA` returns
   `""`, so the delta round's baseline is lost and `DeltaEligible`'s
   diff-in-anchors check cannot run → the round becomes a final hit instead
   of a delta round (intake lane-12 measurement; mechanism verified in
   source). Existing coverage gap: `TestPreviousAuditedSHALegacyPriorRound`
   covers legacy-as-PRIOR only (`audit_counter_review_test.go:70`).
2. **D2** — a debt-admit outcome persists to §G and the trail without the
   debt inventory it admitted (intake lane-11 measurement; grammar verified
   at `audit_ceiling.go:462-464` and `:523-524`). Existing coverage gap:
   `TestCeilingPolicyDebtAdmit` (`audit_ceiling_test.go:258`) asserts the
   in-memory outcome's `Debts`, nothing asserts the persisted record.
3. **D3** — concurrent §G record appends race and lose records
   (`appendProgressRecord` read-modify-write, `audit_ceiling.go:479-496`;
   leader-cited gate measurement 24→7 + heading lost). Existing coverage
   gap: no test exercises `appendProgressRecord` under concurrency.
4. **Context (not defects of this card)**: the t1500 seal record notes the
   t1500 branch tip `3d7215b72` was pushed to
   `origin/WT-audit-ceiling-counter` and is NOT an ancestor of origin/main —
   the engine reached main by another landing (verified:
   `git merge-base --is-ancestor 3d7215b72 origin/main` → not ancestor, and
   `audit_ceiling.go` present on `903ccd028`). The t1500 re-review's
   template-neutrality P1 and t1538's mirror-path/overlay gates are separate
   open debt (see spec.md Out of Scope).

## §C Pre-flight

Files affected (3):

| File | Change |
|---|---|
| `internal/runtime/audit_ceiling.go` | D1: dual-family resolution of the latest round number in `previousAuditedSHA`. D2: debt inventory in the two record lines `persistOutcome` writes. D3: atomic §G append in `appendProgressRecord`. |
| `internal/runtime/audit_counter_review_test.go` | D1 reproduction + preservation tests (next to the F4 sibling `TestPreviousAuditedSHALegacyPriorRound`). |
| `internal/runtime/audit_ceiling_test.go` | D2 reproduction + escaping + negative tests (next to `TestCeilingPolicyDebtAdmit`), the D1 engine-level delta test (git-based, pattern of `TestDiffInsideAnchorsHunkScope`), and the D3 concurrency test. |

Pre-flight checks (all observed at plan phase):

- Working tree identity: `git rev-parse --show-toplevel` =
  `.moai/worktrees/t1560`; branch `WT-internal-runtime-audit`; HEAD
  `903ccd028`; `git status --short` clean before SPEC authoring.
- Consistency read (a) — t1500 seal record
  (`.moai/worktrees/t1500/.moai/reports/t1500/lane28-wait-claude-gate.md`,
  §SEAL/§PUSH): the seal froze the engine work at `3d7215b72` (committed,
  pushed, 22/22 AC + card-review repairs). This repair does not contradict
  the seal: it extends the dual-family contract the sealed card-review F4
  test already established (legacy-as-prior → legacy-as-latest), and the
  sealed test surface stays green (AC-ACR-012). The seal's resume points
  (re-review recording, factory stage path, leader push batch) are untouched
  by this card.
- Consistency read (b) — t1538 sealed resume point
  (`origin/WT-t1538-factory-recovery`,
  `.moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/progress.md` §봉인 기록,
  read via `git show` — the local worktree no longer exists): the remaining
  gate inventory is (1) the factory mirror-path P1 (dispatch store + binding
  update in one lock section) and (2) the `^TestReview` overlay reproduction
  family. Both live in factory dispatch / review-gate code, disjoint from
  `internal/runtime/audit_ceiling.go` — this repair does not invalidate the
  remaining-gate inventory.
- `[NEEDS CLARIFICATION]` markers: none — the dispatch settled scope, fix
  directions, record-format constraints, and the consistency obligations.
- No parallel-session dependency: card-dedicated worktree; no other SPEC
  touches `internal/runtime/audit_ceiling.go` on this branch.

## §D Constraints

1. **Reproduction-first (HARD)**: run phase begins with the two reproduction
   tests, each observed failing on unmodified main `903ccd028` for the stated
   reason (record verbatim output + exit code in progress.md §E.2), then the
   minimal fix, then green.
2. **Minimal diff**: no refactor, no renames, no doc-comment sweeps, no
   drive-by cleanups. The default D1 mechanism resolves the latest round
   number at the `previousAuditedSHA` site, leaving `iterationOf`'s
   convention-only contract and signature unchanged (decision-index Q3).
3. **Record contract (D2)**: the debt field is additive, appears only when
   `len(oc.Debts) > 0`, keeps each record a single line, and escapes all
   three debt fields per the quoting discipline of the chosen encoding
   (default: JSON value inside the `debts=` key — decision-index Q2). Other
   outcomes' records stay byte-compatible (REQ-ACR-006).
4. **Best-effort preserved**: a record-write or serialization failure warns
   on stderr and never changes the returned outcome, the override flag, or
   the blocked flag (REQ-ACR-007, AC-ACR-009).
5. **Security lens**: anchored regex on untrusted file names (full-line
   anchors, `regexp.QuoteMeta` on the SPEC id, digit-only capture); debt
   fields are untrusted free text and must be fully escaped; no new error
   path may escape `EvaluateCeiling` from a record failure.
6. **Commit/commit-message discipline**: Conventional Commits with the card
   id and SPEC id (`fix(SPEC-AUDIT-CEILING-REPAIR-001): ... (card t1560)`),
   `Authored-By-Agent:` trailer per the ownership matrix; affected-package
   tests before commit per AGENTS.local.md §4 (no local full-suite run).
7. **D3 fix scope**: the atomicity fix lands in `appendProgressRecord` only —
   in-process serialization of the read-modify-write plus an atomic replace
   (temp file + same-directory rename) as the default direction. Other
   progress.md writers and cross-process flock stay out unless the lane
   finds them required for this fix (spec.md Out of Scope).

## §E Self-Verification (run-phase exit matrix)

- E1 AC matrix — every AC-ACR-001..013 green, with the three RED-first
  evidences recorded verbatim (§E.2).
- E2 Reproduction integrity — each RED evidence shows the test failing on
  unmodified main for the stated reason, then passing after the fix.
- E3 Affected-package measurement — `go test -timeout 30m ./internal/runtime/...`
  green (the owning-package family is the unit of re-measurement); the
  pre-existing ceiling/card-review tests named in AC-ACR-012 all pass
  unchanged.
- E4 Lint/format — `go vet ./internal/runtime/...` and `golangci-lint run`
  clean on the changed package; gofmt clean.
- E5 Scope grep — the diff touches exactly the three files of §C; no
  `auditverdict`, `DeltaEligible`, or JSON-path (`RecordCeilingOutcome`)
  changes.
- E6 Record grammar spot-check — §G record for a no-debt outcome is
  byte-identical to the pre-repair grammar (AC-ACR-008).
- E7 Consistency notes — the t1500/t1538 read-and-note records present in
  progress.md (AC-ACR-010/011).

## §F Milestones (priority-ordered; no time estimates)

- **M1 (Priority High) — RED: the three reproduction tests.** Author
  `TestPreviousAuditedSHALatestLegacyRound` +
  `TestPreviousAuditedSHAUnparseableLatestStaysEmpty`
  (`audit_counter_review_test.go`),
  `TestPersistOutcomeDebtAdmitCarriesDebtInventory`
  (`audit_ceiling_test.go`), and
  `TestAppendProgressRecordConcurrentSurvival`
  (`audit_ceiling_test.go`). Observe each failing on unmodified main
  `903ccd028` for its stated reason (D1: `""` returned because
  `iterationOf` parses only the convention family; D2: record line carries
  no debt tokens; D3: concurrent §G appends lose records — the gate
  measured 24→7 survivors + heading lost). Record verbatim outputs in
  progress.md §E.2.
- **M2 (Priority High) — D1 minimal fix + green.** Resolve the latest round
  number through dual-family parsing (convention OR legacy, same regex shape
  the prior-round loop uses: `^<QuoteMeta(SpecID)>-review-([0-9]+)\.md$`),
  keeping `latestN <= 0 → no prior round` for genuinely unparseable names.
  Green: AC-ACR-001..004.
- **M3 (Priority High) — D2 + D3 minimal fix + green (persistence milestone).**
  Extend the two record lines `persistOutcome` writes with the debt inventory
  when non-empty (default encoding: a `debts=` field carrying a JSON array of
  `{id, dispose_in, description}` — single line by construction, complete
  escaping via the stdlib encoder; decision-index Q2), and make the §G
  append atomic and complete under concurrency (decision-index Q4 — default
  direction: serialize the read-modify-write in-process + atomic
  temp-file-and-rename; broader hardening excluded per spec.md Out of
  Scope). Green: AC-ACR-005..009, AC-ACR-013.
  The D2 record-format decision is the highest-change-likelihood decision in
  this plan — if the lane adjusts the encoding, only M3's tests and the
  decision-index row move; the ACs are format-agnostic (tokens recoverable).
- **M4 (Priority Medium) — consistency notes + re-measurement.** Record the
  t1500/t1538 read-and-note confirmations in progress.md (they were read at
  plan phase; run phase re-confirms against the post-repair diff), run the
  §E matrix (E3-E5), and write the @MX tag report if any tag changed.

## §G Anti-Patterns

- Do NOT change `DeltaEligible`, the diff-in-anchors check, or any ladder
  rung's admission semantics.
- Do NOT make persistence strict — no error return, no panic, no admission
  flip on record failure.
- Do NOT rewrite or normalize historical §G/trail lines (append-only record).
- Do NOT extend `iterationOf`'s contract or rename helpers (default
  mechanism keeps it untouched; both its callers live in
  `previousAuditedSHA`).
- Do NOT add a second recording path — `RecordCeilingOutcome` stays the ONE
  JSON-path writer; `persistOutcome` stays the §G/trail writer.
- Do NOT touch `internal/auditverdict` (predicate, `Debt` struct, parsing).
- Do NOT extend the concurrency fix beyond `appendProgressRecord` — no locks
  on other progress.md writers (audit reporter, agent §E edits), no
  cross-process flock unless required for this fix.
- Do NOT leak SPEC ids or internal state into `internal/template/templates/`
  (the t1500 re-review P1 class) — no template files are involved in this
  repair at all.

## §H Cross-References

- `spec.md` REQ-ACR-001..007 · `acceptance.md` AC-ACR-001..012 ·
  `decision-index.md` Q1-Q3 · `progress.md` §E
- SPEC-AUDIT-CEILING-001 (`REQ-ACE-003/004/007/012` — the postures this
  repair restores), SPEC-AUDIT-CEILING-002 (JSON path, untouched)
- Card t1560 intake: `.moai/reports/t1560/intake.md`
- Leader scope-extension ruling (2026-10-07): D3 adopted on the 주제당 한 장
  basis — decision-index.md Q4 records it
- t1500 seal: `.moai/worktrees/t1500/.moai/reports/t1500/lane28-wait-claude-gate.md`
- t1538 seal: `origin/WT-t1538-factory-recovery:.moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/progress.md`
