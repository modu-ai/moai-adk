# plan.md — SPEC-AUTONOMY-DECIDER-MODE-001 (v0.1.0)

This is a measurement-design card. M0 is the run-phase entry milestone and the SPEC's first-class
act: population redefinition with shown power arithmetic BEFORE any judge call. Milestones are
ordered by decision reversibility — the highest-change-likelihood decisions (population pin,
predicate, caps) come first and are committed before any irreversible measurement act; mechanical
steps come last.

## §A. Milestones

### M0 — Population census, power arithmetic, pin (run-phase, FIRST — before any judge call)

- Census each candidate axis from REQ-DM-001 with a per-candidate extraction predicate and
  recorded exclusions: (C-B) surfaces — plan-audit-failure user-choice rounds, sync confirmation,
  blocker re-delegation, escalation/budget, the t1244 X1 kickoff-adjacent set, card-pick rounds —
  each surface measured for label distribution and constant baselines; (C-C) synthetic control —
  construction spec instantiated (defect taxonomy ≥ 6 classes, class-balanced, 50:50) with its
  constant baseline 50% by construction; (C-A) recorded as context (t1244 sweep fact: no more
  kickoff rounds in the snapshot).
- Compute the REQ-DM-003 power arithmetic per candidate: baseline vs +10%p satisfiability
  (baseline ≤ 90%), required N for Δ = 10%p at α = 0.05 / 80% power (envelope ≈ 120–240 at
  plausible discordant rates π_d 0.15–0.30; target N = 200; minimum interpretable batch N = 120).
- Apply the REQ-DM-002 pinning rule (a qualifying natural surface beats synthetic; the lowest
  baseline breaks ties among qualifying naturals) and
  pin ONE population — or declare 판별 불가 with per-candidate evidence (a valid exit).
- Write the criteria — pinned population, predicate (REQ-DM-006), composite rule and mapping
  thresholds (REQ-DM-007), llm arm pool model+quota (REQ-DM-009), bands re-derived from the
  measured baseline, cap numbers concretized — and COMMIT them on this branch; the run record
  carries the hash as `criteria_commit:`. NO judge call happens before this commit.

### M1 — Instruments and scrub gate (run-phase)

- Census/construction tooling refinements from M0 outputs; arm runners for the committed pools.
- Scrub scanner per REQ-DM-012: 4 deny classes + PII lower-bound patterns; POSITIVE CONTROL with
  dummy secrets across all classes observed firing (8/8-class shape per the t1244 precedent)
  BEFORE first transmission; clean control too; fail-closed (hit ⇒ block, never strip-and-send).
- Jev cross-check instrument: TypeSafe `choice` form WITH the criteria map (direct-POST pattern —
  the `ask.sh` criteria-omission defect of 2026-09-26 is known; same endpoint, key, model
  `jev-latest`), raw responses preserved in `runs/`; confidence read from the authoritative
  `probabilities` field (REQ-DM-006).

### M2 — Controls and the two arms (run-phase)

- Pipeline positive control: hand-verified subset of the pinned population re-parsed for
  extraction/label agreement (the REQ-CALIB-004 20/20 shape, scaled to the pinned population
  size and recorded in the criteria commit).
- Judge positive controls: fixed-seed selection with label-variance intent, selected and recorded
  BEFORE the first judge call (the REQ-CALIB-005 shape) — AMENDED 2026-09-27, operator-approved
  (record: progress.md §E.2 amendment entry): controls are measured and recorded per-item for
  BOTH arms as DATA and published in the verdict, but they NO LONGER gate on label accuracy — a
  judge being a constant or weak responder is a measured finding, not an execution defect. The
  VOID gate binds to PIPELINE validation only: a control (or any batch item) where transmission
  failed, the response did not parse to the committed format (label outside the 4-value space,
  non-numeric confidence), or the composite could not be computed is a pipeline failure and still
  voids. The pipeline positive control (extraction/label agreement, REQ-CALIB-004 shape) stays in
  force unchanged.
- llm 단독 arm batch: N items, committed primary pool (GLM / `glm_task`), caps per REQ-DM-013.
- llm+jev arm batch: identical payload_id set (AC-DM-015), llm stage + Jev cross-check stage,
  composite rule per REQ-DM-007. Pre-declared supplement batches (Opus window if granted, or
  Codex fallback) run under their own declared caps, never merged with the primary pool.
- Three caps declared with `declared_at` before the FIRST judge call of the run (REQ-DM-013);
  `criteria_commit` verified before the first call (REQ-DM-014 (a)).

### M3 — Analysis and verdict (run-phase)

- Per-arm accuracy vs ground truth, ALWAYS alongside both constant baselines; wrong-automation
  counts; per-arm band qualification on the re-derived bands.
- REQ-DM-006 exact discordant-pair binomial on the paired outcomes; quote the minimum detectable
  difference when reading a null at N = 120.
- Apply the REQ-DM-007 mapping → mode recommendation (pool-qualified; pools never merged).
- `.moai/reports/t1261/verdict.md` in the 5-section format (Claim / Evidence /
  Baseline-attribution / Gaps / Residual-risk); t1244/t943/t1261 domain disjointness note
  (REQ-DM-011); 판별 불가 path carries per-candidate census + arithmetic when taken.

### M4 — Close (sync-phase)

- Sync with zero product-code change asserted (diff scope check per the A5 pattern); no CHANGELOG
  entry unless the sync-phase check finds a user-facing change (none is planned).
- Verdict exported to the primary checkout via the lead's window (lane protocol — the lane does
  not push; the lead owns develop push).

## §B. Key design decisions (recorded per the autonomy policy — recommended option taken, rationale stated)

| Decision | Choice | Rationale |
|---|---|---|
| Population primary | Synthetic control (C-C) recommended primary absent a qualifying natural; a C-B natural surface pins instead whenever it qualifies (census baseline ≤ 90% AND N achievable — a qualifying natural beats synthetic) | Only C-C guarantees label variance and a beatable baseline by construction; natural surfaces inherit operator habit bias — the exact t1244 failure shape. Natural preferred whenever it qualifies, for ecological validity (REQ-DM-002); the lowest baseline breaks ties among qualifying naturals. |
| Comparison test | McNemar-style discordant-pair exact binomial, two-sided α = 0.05 | Paired design by construction; exact test valid at the modest achievable N where the normal approximation is not (research.md §4). |
| Confidence fields | Jev: response registry `probabilities` authoritative; instrument `confidence` auxiliary. LLM arm: field declared in the criteria commit from its response format | t1244 sync-audit D3 found the two Jev fields diverge systematically; declaring BEFORE measurement is the fix (REQ-DM-006). |
| Composite rule | llm label + Jev label: agree → label; disagree → `hold`. Confidence-gating NOT applied inside the composite | Mirrors the A1 §C.8 cross-check shape; `jev_min_confidence` is precisely the value the A5 family lacks evidence for — importing it into the composite would measure a threshold nobody has grounded (REQ-DM-007/008). |
| LLM arm pool | Primary GLM (z.ai, `glm_task`); pre-declared Opus supplement if the lead grants a window; Codex named fallback | Autonomous-kickoff policy requires executability without an externally gated window; GLM quota is separate from lane traffic; the supplement preserves production fidelity when available. Timing dependency on the lead recorded (REQ-DM-009). |
| Mode-comparison δ | 5%p minimum decision-relevant difference (NC-1) | Half the band increment: the band asks "meaningfully beats a constant", the mode question asks "meaningfully beats the other arm"; both are pre-registered judgment values, changeable only before the criteria commit. |

## §C. Run-phase preconditions

1. **Autonomous kickoff** — per §A.2 / REQ-DM-014 (card t1266 policy, lead dispatch 2026-09-26):
   proceeds after plan-audit without a separate operator gate, bound by the bundled conditions
   (criteria_commit precedence, three caps, scrub gate). Dispatch-scoped record, not doctrine.
2. **Predecessor anchors** — `git merge-base --is-ancestor e4ea8eb05 HEAD` (A1 sync close) exit
   0, and t1244 artifacts present in this worktree at `.moai/reports/t1261/reference/` (verdict,
   run-record, extract summary). Both verified 2026-09-26 in this worktree; re-read at run entry.
   A3 (t1236) merge is NOT a precondition — its artifacts are not in this tree and are cited as
   card context only.
3. **Pool availability** — GLM pool key/network present; if absent, the named fallback order
   applies (REQ-DM-009); if no pool is available, a measurement-impossible verdict closes the
   card (the REQ-CALIB-008 fail-open path — degraded execution, never an unscrubbed one).
4. **Jev cross-check instrument** — `scripts/jev/` is untracked local tooling in the primary
   checkout (2026-09-26 `git ls-files` count 0 recorded by t1244); the direct-POST pattern with
   the criteria map is the committed shape. Absence of the key closes via the
   measurement-impossible path; it never loosens the scrub gate.

## §D. Risks and dispositions

| Risk | Disposition |
|---|---|
| Census finds no qualifying candidate | 판별 불가 verdict with per-candidate evidence — a valid exit (REQ-DM-002); the card ends with the population map as its deliverable |
| Synthetic items degenerately easy (ceiling) or impossibly hard (floor) | Difficulty-diagnostic pilot (n = 20); construction revision only via NEW criteria commit BEFORE the main batch (REQ-DM-004) |
| Arms' item sets drift apart (per-arm retries/drops) | Equal payload_id sets machine-checked (AC-DM-015); asymmetric drops recorded unmeasured, never silently dropped |
| Post-hoc threshold temptation after weak results | Criteria frozen at criteria_commit; post-commit changes prohibited (REQ-DM-006); the t943 threshold-raising lesson cited in the verdict |
| Judge pool fail-open propagation | All items unmeasured → measurement-impossible verdict; partial scores never carried forward (REQ-DM-009) |
| Scrub hit on corpus-derived payload | Fail-closed block; item recorded blocked, never stripped-and-sent (REQ-DM-012) |
| Opus supplement never becomes available | Recommendation drawn from the committed primary pool; pool gap named in the verdict's Gaps row (REQ-DM-009) |

## §E. Milestone–REQ mapping

| Milestone | REQs bound |
|---|---|
| M0 | REQ-DM-001, 002, 003 (+ criteria content of 004, 006, 007, 009) |
| M1 | REQ-DM-012 (+ instrument shape of 008), REQ-DM-006 (confidence fields) |
| M2 | REQ-DM-004, 005, 008, 009, 013, 014 |
| M3 | REQ-DM-006, 007, 010, 011, 015 |
| M4 | REQ-DM-015 (+ zero-code-change assertion) |

## §F. Needs-clarification status

None at plan phase. All previously-open either/or decisions are closed by pre-fixed rules in the
SPEC body (population pinning rule, pool choice with gated supplement, pilot revision path,
composite rule) — each recorded with rationale in §B per the autonomy policy. The M0 census can
still SURPRISE (e.g., a natural surface qualifying with a lower baseline than expected); the
response to surprise is the pre-fixed pinning rule, not a new question. No
`[NEEDS CLARIFICATION: ...]` markers are carried.
