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
  - D4 (leader ruling #2 fold-in, 2026-10-07):
    `internal/runtime/audit_counter.go:107-114` — a convention-family file
    whose numeric suffix overflows the integer range fails Atoi at `:111`
    and `n` keeps its initialized 1, so `plan-audit.md` +
    `plan-audit-iter999…9.md` dedupe into ONE round (ceiling undercount).

## §B Known Issues

1. **D1** — with a legacy-family latest verdict, `previousAuditedSHA` returns
   `""`, so the delta round's baseline is lost and `DeltaEligible`'s
   diff-in-anchors check cannot run → the round becomes a final hit instead
   of a delta round (intake lane-12 measurement; mechanism verified in
   source). Existing coverage gap: `TestPreviousAuditedSHALegacyPriorRound`
   covers legacy-as-PRIOR only (`audit_counter_review_test.go:70`). Same
   family, leader ruling #2 fold-in — **D4**: an out-of-int-range iteration
   suffix collapses to round 1 at `audit_counter.go:111` (undercount; see
   §A). Existing coverage gap: no test exercises an overflow suffix.
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

Files affected (4):

| File | Change |
|---|---|
| `internal/runtime/audit_ceiling.go` | D1: dual-family resolution of the latest round number in `previousAuditedSHA`. D2: debt inventory in the two record lines `persistOutcome` writes. D3: atomic §G append in `appendProgressRecord`. |
| `internal/runtime/audit_counter.go` | D4: Atoi range error at `:111` counts as its own (fail-counted) round instead of collapsing to 1. |
| `internal/runtime/audit_counter_review_test.go` | D1 reproduction + preservation tests and the D4 overflow + base-report parity test (next to the F4 sibling `TestPreviousAuditedSHALegacyPriorRound`). |
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

1. **Reproduction-first (HARD)**: run phase begins with the five reproduction
   tests (the five RB RED-first criteria's tests), each observed failing on
   unmodified main `903ccd028` for the stated reason (record verbatim output
   + exit code in progress.md §E.2), then the minimal fix, then green.
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
8. **D4 fix scope**: the Atoi range error at `audit_counter.go:111` routes to
   the counter's existing fail-counted path (own round, never `LatestPath`),
   and the base report's round identity stops flowing through the `n = 1`
   initialization at `:109` (own round; still eligible as `LatestPath`;
   parsed-number dedupe within and across families unchanged) — leader
   ruling #3, both modes of the same n=1-init mechanism. Round ordering and
   baseline selection (plan-audit-3 B1): the base report orders EARLIEST —
   round 0, per the engine's own plan-round convention (`planAuditRoundFile`,
   `audit_ceiling.go:560-565`, where `plan-audit.md` already ranks as
   iteration 0) — preceding every numbered round; and the previous-round
   scan in `previousAuditedSHA` becomes base-aware (scan-local dual-family
   parse, per §D.2's D1 default mechanism and §G's sanctioned exception) so
   a numbered latest resolves the base as its previous audited round when no
   numbered round orders between them — today the scan skips the base
   (`n >= latestN`, `audit_ceiling.go:294`) and returns `""` (codex
   overlay-measured at HEAD: Count=2, previous=`""`). The leader-required
   one-line semantics note is carried in AC-ACR-014's wording.

## §E Self-Verification (run-phase exit matrix)

- E1 AC matrix — every AC-ACR-001..014 green, with the RED-first evidences
  (AC-ACR-001/004/005/013/014, each declared "RED is a new test") recorded
  verbatim in §E.2, plus AC-ACR-003's pre-and-post keep-green guard run.
- E2 Reproduction integrity — each RED evidence shows the test failing on
  unmodified main for the stated reason, then passing after the fix.
- E3 Affected-package measurement — `go test -race -count=1 -timeout 30m
  ./internal/runtime/...` green (the owning-package family is the unit of
  re-measurement; `-race -count=1` bound per plan-audit-1 D8 — the D3 fix
  and the concurrency test are goroutine code); the pre-existing
  ceiling/card-review tests named in AC-ACR-012 all pass unchanged.
- E4 Lint/format — `go vet ./internal/runtime/...` and `golangci-lint run`
  clean on the changed package; gofmt clean.
- E5 Scope grep — the diff touches exactly the four files of §C
  (`audit_ceiling.go`, `audit_counter.go`, `audit_counter_review_test.go`,
  `audit_ceiling_test.go`); no `auditverdict`, `DeltaEligible`, or JSON-path
  (`RecordCeilingOutcome`) changes.
- E6 Record grammar spot-check — §G record for a no-debt outcome is
  byte-identical to the pre-repair grammar (AC-ACR-008).
- E7 Consistency notes — the t1500/t1538 read-and-note records present in
  progress.md (AC-ACR-010/011).

## §F Milestones (priority-ordered; no time estimates)

- **M1 (Priority High) — RED: the five RED-first reproduction tests.** Author
  `TestPreviousAuditedSHALatestLegacyRound` and
  `TestEvaluateCeilingLegacyLatestDeltaGranted` (D1),
  `TestPersistOutcomeDebtAdmitCarriesDebtInventory` (D2),
  `TestAppendProgressRecordConcurrentSurvival` (D3), and
  `TestCountAuditRoundsOverflowOwnRound` (D4). Observe each failing on
  unmodified main `903ccd028` for its stated reason (D1: `""` returned
  because `iterationOf` parses only the convention family / nil not
  returned at the delta gate; D2: record line carries no debt tokens; D3:
  concurrent §G appends lose records — the gate measured 24→7 survivors +
  heading lost; D4: two files count 1, want 2). Each is declared "RED is a
  new test (E8 evidence required)" in acceptance.md §A — no repro test
  exists at plan phase by design; record the verbatim RED outputs in
  progress.md §E.2. Separately, run
  `TestPreviousAuditedSHAUnparseableLatestStaysEmpty` as a **pre-and-post
  green guard** (AC-ACR-003, preserve-behavior — it passes on unmodified
  main and must keep passing; NOT part of the RED obligation, per
  plan-audit-1 D3).
- **M2 (Priority High) — D1 + D4 minimal fix + green.** Resolve the latest round
  number through dual-family parsing (convention OR legacy, same regex shape
  the prior-round loop uses: `^<QuoteMeta(SpecID)>-review-([0-9]+)\.md$`),
  keeping `latestN <= 0 → no prior round` for genuinely unparseable names;
  and route the `audit_counter.go:111` Atoi range error to the existing
  fail-counted path (own round, never `LatestPath`). Green: AC-ACR-001..004,
  AC-ACR-014.
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
- Do NOT rewrite `iterationOf`'s exported contract or other helpers — with
  the ONE sanctioned exception (plan-audit-3 B1): the previous-round scan in
  `previousAuditedSHA` may carry a scan-local dual-family, base-aware parse
  (base report = round 0, earliest order) so a numbered latest resolves the
  base as its previous audited round; `iterationOf`'s own convention-only
  contract and signature stay untouched, and no broader helper rewrite is
  in scope.
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

- `spec.md` REQ-ACR-001..009 · `acceptance.md` AC-ACR-001..014 ·
  `decision-index.md` Q1-Q6 · `progress.md` §E
- fix_scope anchor → repair location (v0.3.0, for the iteration-2 delta
  re-audit; anchors from `.moai/reports/t1560/plan-audit-1.md`):
  - `acceptance.md#§A` — §A rewritten: classification + per-criterion
    "RED is a new test" declaration (verdict MP-8/D1)
  - `acceptance.md#AC-ACR-004` — fixture tier ceiling 2; RED command +
    re-derived expectation (verdict D2)
  - `acceptance.md#AC-ACR-010` — SPEC-local excerpt
    `references/t1500-seal-excerpt.md` cited (verdict D4)
  - `acceptance.md#AC-ACR-012` — no-RED no-regression character stated;
    `-race -count=1` bound (MP-8/D1, D8)
  - `acceptance.md#AC-ACR-013` — RED command line added (MP-8/D1)
  - `acceptance.md#§D.2` — REQ-ACR-009 mapping + non-REQ rows (D7)
  - `plan.md#§F-M1` — RED obligation scoped to the five RED-first tests;
    AC-ACR-003 keep-green guard (verdict D3)
  - `plan.md#§E3` — `-race -count=1` bound (D8)
  - `plan.md#§H` — ranges updated (REQ 001..009, AC 001..014, Q1-Q5) + this
    mapping (D6)
  - `spec.md#frontmatter-version` — `version: "0.3.0"` (D5)
- v0.3.1 (leader rulings #3; iter-2 verdict `.moai/reports/t1560/plan-audit-2.md`,
  FAIL 0.96, receipt rcpt-2146b6cf92f676039aae9490 — fix_scope:
  `plan.md#§C-heading`, `plan.md#§E5`, `plan.md#§H-mapping`,
  `spec.md#§A-opening`, `acceptance.md#§D.6-DoD1`, `acceptance.md#§D.5-Tested`):
  - `plan.md#§C-heading` — "Files affected (4)" (R1 token 1)
  - `plan.md#§E5` — explicit four-path list (R1 token 2)
  - `spec.md#§A-opening` — four defect sites D1-D4, both engine files named (R1 token 3)
  - `acceptance.md#§D.6-DoD1` — five RED cells (R1 token 4)
  - `acceptance.md#§D.5-Tested` — the five RED-first evidences (R1 token 5)
  - `acceptance.md#AC-ACR-014` + `spec.md#REQ-ACR-009` — base-report parity
    extension (leader ruling #3 R2)
  - iteration-2 audit targets named per R2: REQ-ACR-009 / AC-ACR-014
    (audited in full by iteration 2 per its delta-scope note)
- v0.3.2 (plan-audit-3 blocking repairs; verdict FAIL 0.75, iteration 3/3 —
  ceiling final hit, receipt rcpt-a6fb7c9d03b341c573bf2fb4 — fix_scope:
  `spec.md#REQ-ACR-009`, `acceptance.md#AC-ACR-014`, `plan.md#§D.8`,
  `spec.md#§D-C5`, `spec.md#frontmatter-version`, `plan.md#§D.1-constraint-1`,
  `plan.md#§H-Q-range`):
  - B1 (base-round ordering + baseline selection) → `spec.md` REQ-ACR-009
    (round-0 ordering + previous-baseline sentences), `plan.md` §D.8 (two
    fix sites + round-0 precedent), `plan.md` §G (sanctioned scan-local
    base-aware parse exception), `spec.md` §D C5 sentence (round-count /
    ceiling-timing scope corrected), `acceptance.md` AC-ACR-014 arm (e)
  - B2 (headline re-scope) → `acceptance.md` AC-ACR-014: headline scoped to
    the two fail-counted RED modes; (c) RED input pinned to overflow
    suffixes; sealed parsed-number dedupe, iterX fail-counting, and
    base-stays-LatestPath reclassified as preserve arms
  - B3 (token class) → `spec.md` frontmatter `version: "0.3.2"`,
    `plan.md` §D.1 constraint 1 ("five reproduction tests"), §H header
    Q1-Q6, plus the class-level sweep guard below

**Count/version sweep guard (class-level, plan-audit-3 B3 — mechanical,
run before the lane commits):** (1) `spec.md` frontmatter `version:` MUST
equal the top `## HISTORY` entry's version; (2) the §H header ranges line
MUST equal the live counts (REQ count, AC count, decision-index Q count);
(3) grep the count-bearing phrase inventory — `two defects|three files|
three RED|both RED|two reproduction|both evidences|(3)|Q1-Q5` — over the
SPEC directory and resolve every hit against the final artifact set (a
historical mapping row is exempt only where its revision block dates it).
- SPEC-AUDIT-CEILING-001 (`REQ-ACE-003/004/007/012` — the postures this
  repair restores), SPEC-AUDIT-CEILING-002 (JSON path, untouched)
- Card t1560 intake: `.moai/reports/t1560/intake.md`
- plan-audit-1 verdict: `.moai/reports/t1560/plan-audit-1.md` (FAIL 0.81,
  receipt rcpt-136dc402f67a6b1d93bf69bf)
- Leader rulings: scope extension (2026-10-07, decision-index Q4), D4
  fold-in (2026-10-07, decision-index Q5), R1 sweep + base-report fold
  (2026-10-07, decision-index Q6) — all on the 주제당 한 장 basis
- t1500 seal: excerpt at `references/t1500-seal-excerpt.md` (source:
  primary checkout
  `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1500/.moai/reports/t1500/lane28-wait-claude-gate.md`)
- t1538 seal: `origin/WT-t1538-factory-recovery:.moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/progress.md`
