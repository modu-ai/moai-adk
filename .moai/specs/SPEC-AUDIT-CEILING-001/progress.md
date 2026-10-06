# progress.md — SPEC-AUDIT-CEILING-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-06
plan_phase_head: 69a085b2d (re-measurement tree for the v0.4.0 re-plan; the v0.4.0 commit lands on top)
plan_phase_branch: WT-audit-ceiling-counter
plan_phase_worktree: .moai/worktrees/t1500
plan_phase_artifacts: 7 (spec.md, plan.md, acceptance.md, design.md, research.md, progress.md, decision-index.md — Tier L set of 5 + progress + decision-index per interview.decision_gate: on)

### Plan-audit v4-iter2 repair — v0.5.1 (2026-10-06, this change)

- v4-iter2 verdict: FAIL 0.94 (0.81 → 0.94, no regression; fresh-series
  iter2) — `.moai/reports/t1500/plan-audit-v4-iter2.md`, audited_sha
  `23fe75465`. iter1's V4-D1/D2/D3 and V4-O1/O2 all verified RESOLVED; one
  new blocking V4-D4 (same mutant-probe class as V4-D3, minted by the
  v0.5.0 repair text: the "removing either producer arm turns the criterion
  red" claim was false in the multi-model direction — a single-model-only
  body greens both v0.5.0 counts). Hunk-limited fix per the verdict's
  reread_hunks (acceptance.md#AC-ACE-008, plan.md#sC-baselines + the
  design §3 phrasing pin); patch version 0.5.1 per the dispatch's
  allowance, recorded in §H.
- V4-D4 closure: AC-ACE-008's claim reworded to THREE asserted counts, both
  mutant directions verified by construction — single-model-only body:
  `convergence_overall` ≥1 (green) + `backend it actually ran` ≥1 (green)
  + `PerBackendVerdicts` = 0 (RED) → criterion red; multi-model-only body:
  `convergence_overall` ≥1 (green) + `PerBackendVerdicts` ≥1 (green) +
  `backend it actually ran` = 0 (RED) → criterion red; receipt-less body:
  all three 0 → red. Third count phrase = `PerBackendVerdicts`, the
  multi-model convergence-result field name plan.md M1's instruction
  carries.
- Deviation from the verdict's example, measured: the prescribed example
  phrase `convergence result` reads **1 / exit 0 @23fe75465** (pre-existing
  unrelated fail-open prose at plan-auditor.md:251) — a RED cell on it
  would be vacuous and its green assertion already satisfied by untouched
  text. `PerBackendVerdicts` measured 0 / exit 1 @23fe75465 and is
  multi-model-only by construction; adopted instead. The rejection and
  measurement are recorded in acceptance.md AC-ACE-008, plan.md §C, and
  design.md §3.
- Cross-layer sweep: design §3 pins the arm-distinctive phrasings
  (multi-model `PerBackendVerdicts` / single-model "backend it actually
  ran"); plan §C gains the multi-model baseline row; plan M1's multi-model
  bullet names the pinned token. No REQ id changed (15); AC count 21.
- Note carried from the verdict: the claude required-backend gate remains
  unmet (inconclusive, both iterations) — no PASS can be yielded regardless
  of score until a claude verdict is obtained; measure with `moai verify
  audit-plan` after the audit_multi call before dispatching the next round.

### Plan-audit v4-iter1 repair — v0.5.0 (2026-10-06, this change)

- v4-iter1 verdict: FAIL 0.81 (Tier L threshold 0.85; fresh v0.4.0-series
  iter1, NOT the exhausted v0.3.0 series' iter4) —
  `.moai/reports/t1500/plan-audit-v4-iter1.md`, audited_sha `a4c5b9594`.
  All 9 must-pass green; 3 blocking (V4-D1/D2/D3) + 2 optional (V4-O1/O2).
  This change closes the full delta, delta-scoped to the verdict's
  fix_scope (spec.md#REQ-ACE-003/#REQ-ACE-008/#REQ-ACE-013,
  acceptance.md#AC-ACE-003/#AC-ACE-008/#AC-ACE-015,
  design.md#s2-ladder/#s3-receipt-schema); the lane runs the v4-iter2
  delta re-audit.
- spec.md version → 0.5.0; §H Amendments carries the per-defect record.
- Blocking closures: V4-D1 (REQ-ACE-003's trigger scoped to verdicts that
  FAIL the shared predicate's admission — the pass-through arm precedes and
  the refusal REQ never claims an admission-clean verdict; REQ-ACE-013
  extended to any ceiling state — the effective ceiling AND the
  tier-ceiling final hit — closing the second face; AC-ACE-003 and
  AC-ACE-015 carry the pass-through exclusion arm), V4-D2 (the receipt
  projection rule defined for every label of the auditor's own verdict
  enum — {PASS, PASS-WITH-DEBT} → pass, {FAIL, FAIL_WARNED} → fail,
  {INCONCLUSIVE} → inconclusive on the backend line / fail on
  convergence_overall; raw label preserved in the verdict body, no silent
  upgrade; AC-ACE-008's PASS-WITH-DEBT end-to-end arm
  `TestParseReceiptPassWithDebtProjection`), V4-D3 (AC-ACE-008's
  export-path verification split per-arm — the multi-model
  `convergence_overall` grep AND a new single-model `backend it actually
  ran` grep, each with its own RED-now cell, both measured 0 / exit 1 at
  a4c5b9594 before this change — so removing either producer arm turns the
  criterion red).
- Optionals folded: V4-O1 (design §1's "i.e. at max-N" gloss qualified
  "under contiguous numbering"), V4-O2 (the Interlock note's stale Q2-Q6
  pointer → Q2-Q5, noting Q0 decided / Q6 fell away).
- Cross-layer sweep: REQ-ACE-008's projection → design §3 + plan M1;
  REQ-ACE-013's boundary extension → design §2 rung 0 + §G risk 5 + plan
  M3 + AC-ACE-003/013/015 arms; plan §C gains the single-model grep
  baseline. No REQ id changed (15 REQs); the AC count stayed 21
  (extensions, not additions).
- Note (from the verdict's Operational Notes): the required-backend claude
  gate was inconclusive this round (subscription unavailable) — even a
  score-PASS repair round cannot leave the audit PASS until a claude
  required-backend verdict is obtained; measure with `moai verify
  audit-plan` after repairs, before dispatching the next audit round.

### v0.4.0 re-plan under operator decision D9 (2026-10-06, this tree, HEAD 69a085b2d)

- Operator decision D9 (card t1500): resume with NARROWED scope — (1) a
  CLI-side per-SPEC audit-iteration counter, (2) a single path that records
  the policy outcome when the cap is reached (record debt then proceed /
  split scope / hold — no questions asked), (3) codify that a required-
  backend fail blocks run entry. Extra rules surfaced by iter3 (D33
  cross-card re-audit dedupe; the run-gate doc reconciliation) are SPLIT
  OFF to follow-up work. The narrowed scope supersedes v0.3.0's scope; the
  iter3 verdict and tree are reference only.
- Scope-cut cross-layer sweep completed: former REQ-ACE-013/014 (doc
  reconciliation) deleted with their AC-ACE-013/014, §D.2 row list, and
  LEDGER-ACE-013-A/014-A/014-B; former REQ-ACE-015/016 renumbered to
  REQ-ACE-014/015; new REQ-ACE-013 = the D31 pass-through arm. No
  requirement cites a deleted surface (spec.md §E carries the two new
  Out-of-Scope topics; acceptance.md §D.1 carries the id-history map).
- D31-D36 disposition table recorded in decision-index.md; Q0 = D9 row
  marked OPERATOR-DECIDED; Q6 marked FELL AWAY.
- Tier judgment: STAYS Tier L — the narrowed SPEC remains multi-subsystem
  (runtime + auditverdict + config + homestate/contract wiring + 2 deployed
  docs + mirrors + state) and gate-semantics (constitutional-adjacent);
  REQ 15/25 and AC 21/21→21/25 within the L ceilings. Re-tiering to M would
  churn the plan-audit threshold (0.85 → 0.80) and demote existing
  design.md/research.md for no benefit — recorded, not taken.
- Baseline re-measurements RE-EXECUTED in this tree at 69a085b2d (not
  carried over from iter1/iter2): every value in plan.md §C and the
  grep-class RED cells of acceptance.md reproduced (no counter 0/exit 1;
  receipt grep 0/exit 1; orphan note 2/exit 0; bare symmetry selector
  0/exit 1; `go test -list` ok-line-only/exit 0; GateConfig 0 non-test;
  plan-auditor.md convergence_overall 0/exit 1; mirrors plan-auditor.md
  DIFF + convention doc SAME; ceilings S1/M2/L3 + auto_delta_rounds 1 +
  hold-and-split; resolver fail-open re-read at mcp_worktree_root.go:123-131).
  The phase-execution.md / auto-semantics.md baseline greps were RETIRED
  with the deleted surfaces (those files are no longer edit targets).
- Fresh audit series: the narrowed v0.4.0 re-plan starts a NEW plan-audit
  iteration series (iter1 of the new series) under the card's own ceiling
  accounting — the prior series (0.69 → 0.83 → 0.81 + STOP) is closed with
  the D9 decision; this re-plan is not an iter4 of the exhausted series.

### Baseline measurements (v0.1.0 original run, this tree, HEAD 2f492df19, 2026-10-04 — superseded by the v0.4.0 re-measurement above, retained for history)

| # | Measurement | Command | Observed |
|---|---|---|---|
| 1 | Conflict text present | `grep -c "Override and proceed" .claude/skills/moai/workflows/run/phase-execution.md` | 1 |
| 2 | No counter in runtime | `grep -rn "audit.round\|AuditRound\|iteration.count" internal/runtime/*.go` (non-test) | 0 hits |
| 3 | No receipt parsing | `grep -c "convergence\|receipt" internal/auditverdict/verdict.go` | 0 |
| 4 | §9 inventory size | pipe-line count over auto-semantics.md :190-201 | 10 disposition rows |
| 5 | Config keys are orphans | internal/config/loader.go:345-346 | "no Go reader" on both ceiling keys |
| 6 | Pre-existing mirror drift | `diff -q` deployed vs template | harness.yaml DIFF, phase-execution.md DIFF, plan-auditor.md DIFF (added in the iter1 correction — research.md §3); auto-semantics.md SAME, convention doc SAME |
| 7 | Tier ceilings verified | harness.yaml :75-78 | S:1 M:2 L:3; policy :82-84 auto_delta_rounds=1, on_final_hit=hold-and-split |

### Plan-phase decisions recorded

1. Tier L — multi-subsystem (Go runtime + auditverdict + config + 2 docs +
   mirrors + state), ~13-15 files, gate-semantics (constitutional-adjacent)
   scope; artifact set of 5 + progress + decision-index.
2. SPEC ID `SPEC-AUDIT-CEILING-001` — regex PASS (Bash-verified), unique in
   `.moai/specs/` (0 prior matches).
3. No new config keys — the outcome ladder is REQ-encoded; config gains only
   a Go reader (research.md §4, simplicity ladder).
4. Receipt-absence disposition defaults to refuse (fail-closed) — flagged as
   decision-index Q4 for operator confirmation.
5. decision-index.md authored (6 rows: 1 POLICY-COVERED, 5 FOUNDER) per
   `interview.decision_gate: on`.
6. The card's own plan audit follows the ceiling policy this SPEC encodes:
   authored once, then repair rounds per the policy (the iter1 and iter2
   repairs recorded below, inside the card's own 3+1 iteration ceiling) —
   the prose policy the CLI will enforce has governed this card's own audit
   loop.

### Interlock notes for later phases

- manager-develop owns §E.2/§E.3; manager-docs owns §E.4. This file's later
  edits by those agents are expected and permitted.
- decision-index.md rows carry empty Operator verdicts; the kickoff
  conversation should surface Q2-Q5 before or during run entry (Q4 gates
  M1's receipt-absence branch; Q2/Q3/Q5 gate M3). Q0 is OPERATOR-DECIDED
  (the D9 resume posture) and Q6 FELL AWAY with the D9 scope cut — neither
  is surfaced (V4-O2).

### Plan-audit iter1 repair (2026-10-04, this change)

- iter1 verdict: FAIL 0.69 (Tier L threshold 0.85) —
  `.moai/reports/t1500/plan-audit-iter1.md`; 18 findings, 13 blocking. This
  change closes the defect delta; the lane runs iter2 (delta-scoped
  re-audit).
- spec.md version → 0.2.0; §H Amendments carries the per-defect record.
- Blocking closures: D1 (E8 new-test declarations on every RB new-test row;
  bare `TestStructYAMLSymmetry` harness added to M2; exit codes recorded in
  grep-class RED cells), D2 (debt-admit re-gated to label-only failures;
  hash mismatch and no-findings verdicts hold; AC-ACE-017), D3 (receipt
  projects every required backend's verdict; REQ-ACE-008 trigger moved to
  required-configured trees; AC-ACE-018/019), D4 (refusal rerouted to the
  live seams decide.go/homestate; `GateConfig.Invoke` measured
  production-dead — 0 non-test refs; AC-ACE-022), D5 (identity = SPEC id +
  iteration number; AC-ACE-020), D6 (delta eligibility encoded;
  AC-ACE-021), D7 (at-ceiling condition on REQ-ACE-004..006), D8 (one
  negative-verdict outcome vocabulary naming the override, `EnvSkipAudit`,
  and FAIL_WARNED), D9 (AC-ACE-015 rescoped region-scoped with a named
  known-FAIL list; research §3 corrected — plan-auditor.md mirror DIFF),
  D10 (override record → progress.md §G, outside the hash set), D11 (M1
  override input removed), D12 (AC-ACE-014 replace arithmetic 10−1+11=20 +
  per-named-row verification), D13 (`.moi/` → `.moai/`, ×3), D14
  (AC-ACE-016 → REQ-ACE-007 remap; AC-ACE-022 added for REQ-ACE-016; edge
  links restated), D15 (M3 gated on Q2/Q3/Q5, defaults kickoff-amendable),
  D16 (§D.3 → §D.2), D17 (`on_final_hit` validated; M2 harness.yaml mirror
  step dropped; plan_audit_global orphan out of scope), D18 (per-key
  receipt repeat rule stated).
- Held coherence passes: the Ceiling-field choice (design §7), receipt
  projection from real `ConvergenceResult` fields, the §9 vocabulary match,
  spec lint --strict 0.
- No REQ id changed (16 REQs); the AC set extended 16 → 22 (Tier L ceiling
  25).

### Plan-audit iter2 repair (2026-10-04, this change)

- iter2 verdict: FAIL 0.83 (Tier L threshold 0.85; iter1 0.69 → 0.83, no
  regression) — `.moai/reports/t1500/plan-audit-iter2.md`; all 18 iter1
  findings verified RESOLVED; new blocking set D19-D24 + optional D25-D30.
  This change closes the full delta (all 6 blocking + all 6 optional); the
  lane runs iter3 (the last numbered round under the card's own 3+1
  ceiling).
- spec.md version → 0.3.0; §H Amendments carries the per-defect record.
- Blocking closures: D19 (receipt producer assigned — plan.md M1 names the
  plan-auditor agent body's export step as the writer, sourced from the
  `audit_multi` convergence result, mirror included; export-path RED-now
  grep + green arm added to AC-ACE-008, reclassified RB), D20 (REQ-ACE-004's
  never-debt-admit list extended with the REQ-ACE-009/010 receipt refusals;
  consuming-seam admission semantics stated — the seam re-runs the full
  predicate with one label conversion, design.md §2; AC-ACE-004 negative arm
  + `TestCeilingPolicyReceiptHold`), D21 (config-error disposition named
  fail-closed — REQ-ACE-010 third trigger arm, design.md §4 resolution
  contract, AC-ACE-010 invalid-config arm + `TestAdmitConfigErrorRefused`),
  D22 (dedupe identity resolved within one audited state — REQ-ACE-001,
  design.md §1 cross-card rule; cross-card renumbering cannot stall the
  count; AC-ACE-001 cross-card arm + §C edge 9), D23 (REQ-ACE-003 trigger
  reworded to the admission-seam event the design delivers; AC-ACE-003 Then
  reworded; AC-ACE-022 extended to one arm per LIVE seam including the
  kickoff evaluator + `TestKickoffEvaluatorCeilingRefusal`), D24
  (decision-index Q5 rewritten to the label-only predicate).
- Optionals folded (none deferred): D25 (AC-ACE-013 per-member vocabulary
  greps LEDGER-ACE-013-A; AC-ACE-014 whole-file total count LEDGER-ACE-014-B
  33 → 43 — plain single-command form), D26 (EnvSkipAudit disposition
  pointer → spec.md §D.2 row 2; M4 + design §9 restatements scoped to "the
  question branches"), D27 (five stale cross-references: plan §H AC range,
  plan §B/progress baseline mirror-drift lists, progress decision 6,
  acceptance header 1:1 claim, plan §C `^| ` decomposition label), D28
  (package-wide `go test -list` corroboration on AC-ACE-002; §A exception
  note), D29 (AC-ACE-008/011 reclassified RB with E8 declarations;
  AC-ACE-005/012 RG rationale stated in §A), D30 (hold-release path in
  REQ-ACE-006 + design §8; C5 wording; REQ-ACE-007 extended to
  required-backend refusals; REQ-ACE-012 "durable" corrected to the §G
  durable carrier).
- New RED-now claims re-executed in this tree before commit: plan-auditor.md
  `convergence_overall` grep = 0 exit 1 (AC-ACE-008 export arm) and
  `go test -list '^TestStructYAMLSymmetry$' ./internal/config` lists no test
  (AC-ACE-002 D28 corroboration) — both at a991e9bbb, recorded in the
  criteria.
- Design decisions taken (recorded per the verdict's Residual-risk note):
  D23 resolved by rewording to admission semantics, not pre-spawn wiring —
  the prose loop self-governs below the ceiling and the machine seam is
  where the CLI can refuse (consistent with iter2's rejection of codex
  P2-7); D20 resolved by seam-reruns-full-Admit with one label conversion.
- Held coherence passes: the Ceiling-field design (§7), receipt projection
  from real `ConvergenceResult` fields, the §9 vocabulary match, 16 REQ, and
  the 22-AC count (extensions, not additions — Tier L ceiling 25 not
  approached).

### Plan-audit iter3 — ceiling hit + STOP (2026-10-04, recorded by lane-26)

- iter3 verdict: **FAIL 0.81 with STOP signal** (score regression 0.83 →
  0.81; iter1 0.69 → iter2 0.83 → iter3 0.81) —
  `.moai/reports/t1500/plan-audit-iter3.md`, audited_sha `9dd4d5c74`,
  receipt `rcpt-daab786d53b574d2bbe55847`. All 12 iter2 defects (D19-D24
  blocking + D25-D30 optional) verified RESOLVED; regression sweep fully
  green (must-pass 9/9, RED cells verbatim, 16 REQ ↔ 22 AC, lint strict
  0/0). New blocking: D31 (ceiling ladder has no arm for a fully-passing
  verdict at/over the ceiling — healthy resumes fall to hold by omission),
  D32 (receipt producer fed only from multi-model ConvergenceResult —
  single-model audits on required-resolving trees have no writer and block
  permanently), D33 (same-SHA dedupe vs "new card re-audit is a new round"
  cannot both hold for no-repair cross-card re-audits — ceiling blind to
  that churn), D34 (§G risk 2 contradicts v0.3.0's own REQ-ACE-001). Optional
  D35-D36. Cross-model gate UNMET this round: claude required-but-
  inconclusive (capacity), glm inconclusive, codex fail — this is the exact
  situation the card's work item 2 exists to block run entry on.
- Per the leader's dispatch discipline ("상한에 닿으면 수리를 이어 가지
  말고 보고") and the plan-auditor STOP clause (score regression — no
  unconditional iter4), the lane STOPS here: 3/3 numbered rounds consumed,
  delta round NOT taken automatically. Operator choices per the verdict:
  scope reduction / accept-with-debt / explicit override (or a leader
  decision on any of these). Tree held at HEAD `9dd4d5c74` (v0.3.0),
  unpushed; all three verdict files preserved.

## §E.2 Run-phase Evidence

_pending run-phase — owned by manager-develop_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase — owned by manager-develop_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase — owned by manager-docs_
