# progress.md — SPEC-AUTONOMY-DECIDER-MODE-001

> Plan-phase authoring record (2026-09-26): [PLAN] AC judgments are recorded in the §E.1 block
> below at authoring time as designed-for (grep counts verified during authoring); the formal
> re-execution belongs to the plan-audit gate.

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-26
- spec_version: 0.1.0 (initial plan-phase authoring)
- plan_audit: FAIL (iter-1, 2026-09-26, score 0.91 — Tier M threshold 0.80 met but 2 blocking
  findings: D1 research.md §3 π_d=0.15 row unreproducible from the stated formula [stated
  params yield n_d≈18/N≈118 vs claimed 29/191; 0.20/0.30 rows reproduce], D2 REQ-DM-002
  pinning-rule natural-vs-synthetic precedence reads two ways across spec.md/plan.md/research.md.
  MP-1..7 all PASS; D3 minor blocking: AC-DM-016 lacks the pool-fidelity Gaps-row requirement.
  Repair list + delta-scoped re-audit scope: .moai/reports/t1261/plan-audit-iter1.md)
- iter1_repair_adoption: D4 (AC-DM-002 command coverage over research.md §3 rows) adopted by lane
  per the card t1266 autonomy policy — closes the AC dead zone where a D1-class value error in
  research.md sat invisible to the AC layer; D5–D7 remain operator-discretion items in the
  iter-1 report, not adopted this iteration
- tier: M (spec.md, plan.md, acceptance.md, research.md — measurement-design card, zero code
  deliverables; progress.md present at every tier)
- requirements: 15 (REQ-DM-001..015, continuous) / acceptance criteria: 16 (AC-DM-001..016;
  [PLAN] 11, [RUN] 5 — AC-DM-012..016 are run-record targets, `--- PENDING-RUN` until the run)
- card: t1261 (AUTONOMY-A5b — contract-mode kickoff decider default: llm 단독 vs llm+jev;
  t1244 split, operator decision 2026-09-26, Tier M, Class C)
- intent_clarity: ~9/4 quadrants — interview SKIPPED, rationale: operator-drained card (the
  operator's decision is already verbatim in the card text), factory-lane context (questions
  route through the lead, not the operator), and the card t1266 autonomy policy (card-specific
  condition choices take the recommended option with rationale recorded). The one genuinely
  open either/or — primary LLM-arm pool under an externally gated Opus quota — is closed by a
  pre-fixed conditional rule (GLM primary + pre-declared Opus supplement), not left open.
- measurement_target: the contract-mode kickoff decider DEFAULT decision evidence —
  `workflow.autonomy.kickoff.decider` derives `llm` under `mode: contract` today (A1 0.5.2
  REQ-CONTRACT-015); changing that default is A1-owned follow-up, NOT this card
- population_redefinition: FIRST-CLASS M0 — census per candidate (C-A kickoff rounds:
  sweep-exhausted, rejected; C-B other surfaces: census-gated, ≥ 759 non-kickoff AUQ rounds in
  the t1244 snapshot window; C-C synthetic control: recommended primary, baseline 50% by
  construction) → power arithmetic per candidate → pin ONE by the pre-fixed rule (or 판별 불가,
  a valid exit) → criteria_commit BEFORE any judge call
- power_arithmetic_seed: baseline > 90% kills baseline+10%p (t1244 structural lesson);
  balanced 50/50 → band (b) = 60% meaningful; two-arm Δ = 10%p at α = 0.05 / 80% power needs
  ≈ 120–240 items at plausible discordant rates (π_d 0.15–0.30) — target N = 200, minimum interpretable batch
  N = 120 (MDE ≈ 11–12%p there)
- predicate_preregistered: McNemar-style discordant-pair exact binomial (two-sided α = 0.05);
  per-arm accuracy always alongside both constant baselines; wrong-automation ≤ 10%; band shape
  (a) n ≥ 20 / (b) baseline+10%p / (c) wrong-automation, re-derived from the pinned baseline;
  confidence fields declared BEFORE measurement — Jev authoritative = response registry
  `probabilities` (t1244 sync-audit D3 divergence fix), instrument `confidence` auxiliary
- composite_and_mapping: llm+jev composite = agree → label, disagree → `hold`
  (confidence-gating deliberately outside the composite); mode mapping δ = 5%p — better →
  recommend `llm+jev`; worse or equal → recommend `llm 단독`; pool named on every
  recommendation, pools never merged
- llm_arm_pool: primary GLM (z.ai, `glm_task` — quota separate from the Anthropic lane pool);
  pre-declared Anthropic (Opus) supplement ONLY on a lead-granted quota window (timing
  externally gated — recorded dependency); Codex named fallback; criteria commit names the
  exact model identifier
- preregistration: criteria_commit pattern inherited from t1244 (criteria pinned by a branch
  commit, hash recorded and time-compared against the first judge call); three caps
  (turn 2/item, call 2 × (N + controls) per arm-batch, wall-clock PT8H/batch, batches = 1 per
  arm) declared with `declared_at` before the first call; no cap extension
- scrub_gate: inherited REQ-CALIB-011 shape — 4 deny classes + PII lower-bound set, positive
  control before FIRST transmission, fail-closed (hit ⇒ block, never strip-and-send), applied to
  EVERY external judge (Jev/TypeSafe, GLM, Anthropic, Codex); lower-bound limitation goes to the
  verdict Residual-risk
- korean_and_t943_discipline: no translation arm (§30 measured the ~10%p language effect,
  conclusion-invariant); t943 premise-triage NOT re-run; §30 rows cited only with the
  constant-baseline row; three-domain disjointness (t943 / t1244 / t1261) stated in the SPEC
  body and repeated in the verdict
- predecessor_anchors: A1 sync close `e4ea8eb05` reachable from this branch
  (`git merge-base --is-ancestor e4ea8eb05 HEAD` exit 0, verified 2026-09-26 in this worktree);
  t1244 reference artifacts present at `.moai/reports/t1261/reference/` (verdict, run-record,
  extract summary); both depends_on SPECs `status: completed`; A3 (t1236) artifacts NOT in this
  tree — cited as card context only
- linked_pr: none (card worktree branch `WT-decider-mode-eval`; git-flow lane — local develop
  merge via the lead's window, lane does not push)
- run_preconditions: autonomous kickoff per §A.2 / REQ-DM-014 (card t1266 policy — proceeds
  after plan-audit, bound by bundled conditions); pool availability per REQ-DM-009 fallback
  order; measurement-impossible verdict as the total-unavailability exit; scrub failure blocks
  transmission (fail-closed), a different path from unavailability
- open_clarifications: NONE at plan phase — NC-1 (δ = 5%p and +10%p are pre-registered judgment
  values, changeable only before the criteria commit) and NC-2 (natural-over-synthetic
  tie-break) are body decisions with defaults kept, recorded in research.md §7; no
  `[NEEDS CLARIFICATION: ...]` markers carried
- scope_guard: product code, `internal/`, `internal/template/` untouched — instruments live
  under `.moai/reports/t1261/` (untracked during work); AC-DM-011 checks the commit set
- plan_artifact_hash: 3cabee24a6e5dc2f5b2c09202797f11d893d46bd617747b08c49ae5439b46387
  (sha256 over cat acceptance.md plan.md research.md spec.md, exact bytes; measured at
  plan-audit iter-1 entry, 2026-09-26, tree @ e611a99a1)
- Implementation Kickoff Approval: not requested at plan phase (run entry follows the
  plan-audit gate and the autonomous-kickoff disposition)

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
