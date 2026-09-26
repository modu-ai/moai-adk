# research.md — SPEC-AUTONOMY-DECIDER-MODE-001 (v0.1.0)

This card's research has four parts: what the contract-mode kickoff decider concretely denotes at
the decision layer in this tree (§1), the candidate-population inventory with feasibility
assessment (§2 — enumeration only; the transcript census is M0 run work, per the dispatch), the
power-arithmetic framework with worked arithmetic (§3), and the statistical-method recommendation
with the reason it fits the achievable N (§4). No corpus sweep and no judge call happened in plan
phase.

## §1. What "contract-mode kickoff decider" denotes in this tree

Source: `SPEC-AUTONOMY-CONTRACT-001` (A1) v0.5.2, `status: completed`, present in this tree
(sync close commit `e4ea8eb05` reachable from this branch — `git merge-base --is-ancestor
e4ea8eb05 HEAD` exit 0, re-verified 2026-09-26 in this worktree).

| Element | Where it lives (A1 0.5.2) | What it says |
|---|---|---|
| Decider value set | spec.md §C.8; REQ-CONTRACT-015 | `human \| llm \| llm+jev`. Jev is never a sole decider; a configured `decider: jev` is the configuration error `kickoff_decider_jev_sole` (loading succeeds; only the receipt signing path is refused). Operator decision 2026-09-26, re-decided same day (0.5.0 → 0.5.1): the set is final. |
| Contract default derivation | REQ-CONTRACT-015 | When `decider` is absent or empty, the effective decider derives from the effective mode — `human` under `guided`, **`llm` under `contract`**. The template omits `kickoff.decider` (REQ-CONTRACT-016), so today's contract default IS `llm` 단독. |
| `jev_min_confidence` | REQ-CONTRACT-015 | `workflow.autonomy.kickoff.jev_min_confidence`, range 0.0–1.0, default 0.50. The evidence base for that default was the population t1244 measured — and t1244 concluded that population cannot support ANY judge claim (verdict.md Claim 1). |
| `llm+jev` fallback | §C.8; REQ-CONTRACT-023 | `llm+jev` falls back to `llm` on a Jev-side failure (missing key, `workflow.jev.enabled: false`, low confidence, malformed response), recorded in the receipt (requested/effective decider + fallback reason). |
| Cross-check rules ownership | §C.8 lead decision 2026-09-26 | A1 owns the schema (value set, derivation, receipt fields, structural validator); **A3 (card t1236) owns the rules** — the cross-check agreement (both approve → start; a disagreement → a human) and the fallback decision. |
| Interim rule | §C.8; REQ-CONTRACT-024 | While A3's amendment of the shipped "Jev is display-only" principle has not landed, A1 refuses any receipt whose effective decider is `llm+jev` with `receipt_requires_human`. A3 lifting that rule is a forward note owned by A3. |
| A3 (t1236) | Separate branch — artifacts NOT in this tree | Cited as card context only per the dispatch: A3 is plan-complete on `SPEC-AUTONOMY-GATE-REWIRE-001` (gate rewiring design; from t1244's research record, A3 §A assigns the `jev_min_confidence` evidence measurement to the A5 family and states A3 does not wait for it). This SPEC's verdict is decision EVIDENCE for that design and for the operator's default choice — it implements nothing and amends nothing. |

**The question this card answers, precisely.** Under `mode: contract` with `kickoff.decider`
absent, the effective decider today is `llm` (A1 REQ-CONTRACT-015 derivation). The card asks
whether that default should be `llm` 단독 or `llm+jev` (LLM answer cross-checked by Jev). Changing
the default — config schema default, template, or derivation — is a FOLLOW-UP card (A1 owns the
config; A3 owns the gate rules); t1261 produces the decision evidence only.

**What t1244 left as the starting point.** `SPEC-AUTONOMY-KICKOFF-CALIB-001` (A5, completed)
measured the Jev arm on the kickoff-gate population and closed with: population = 105 operator
kickoff rounds (approve 104 / modify 1 / hold 0), always-approve constant baseline 104/105 =
99.05%, Jev judge agreement 96/105 = 91.43%, qualification bands = 0 because band condition (b)
"accuracy ≥ baseline + 10%p" = 109.05% is mathematically impossible. Its verdict's Claim 5
transferred the llm-vs-llm+jev question to this card, and its Residual-risk states the structural
lesson this SPEC is built on: any population whose constant baseline exceeds 90% makes a
baseline+10%p condition unsatisfiable by design. The same verdict's Evidence section records that
the sweep covered BOTH corpus roots in full — within the kickoff gate there are NO more rounds to
expand to in the snapshot; "라운드 확대" can only mean other decision surfaces or future
accumulation.

## §2. Candidate-population inventory (enumeration; census is M0 work)

Facts available from t1244 artifacts (no new sweep performed in plan phase):

- The same corpus snapshot window held **907 AskUserQuestion tool_use blocks** across both roots,
  of which **148** mentioned kickoff (t1244 research.md §3.2 probe). That leaves roughly **≥ 759
  non-kickoff AskUserQuestion rounds** in the same window — the raw material for candidate C-B.
  Their per-surface label distributions were NOT measured and are unknown until the M0 census.
- The kickoff-mention candidates that FAILED t1244's two-stage gate predicate (X1, 50 items) are
  pre-identified kickoff-ADJACENT non-gate decisions — card issuance, scope/schedule, review-lens
  choices. They are a ready-made M0 census sub-target for C-B (their extraction predicate already
  ran once in t1244's pipeline shape).
- X2 (27 rows) were no-response fallbacks ("No response after 60s"), excluded from the t1244
  population with originals preserved — they are not operator decisions and stay excluded.

### Candidate C-A — kickoff-gate rounds with holds/rejections (라운드 확대)

| Aspect | Assessment |
|---|---|
| Surface | The Implementation Kickoff Approval gate itself. |
| Where recorded | Session transcripts, both corpus roots (roots named in the t1244 verdict Evidence section). |
| Label variance | Measured: approve 104 / modify 1 / hold 0. The sweep is exhaustive over the snapshot — no more rounds exist to add. Future accumulation is possible but reflects the same operator habit (99.05% approve) with no reason to expect variance. |
| Feasibility risk | Fatal as a primary: constant baseline 99.05% > 90% kills any baseline+10%p band (the t1244 structural lesson). Viable only as a future-window supplement if a census ever shows real variance. |
| Disposition | **Rejected as primary.** Kept as context. |

### Candidate C-B — other decision surfaces with recorded operator decisions

| Aspect | Assessment |
|---|---|
| Surface | Non-kickoff AskUserQuestion gates with a recorded operator choice. Concretely enumerable candidates: (B1) plan-audit-failure user-choice rounds (repair / reject / abort), (B2) sync confirmation questions, (B3) blocker re-delegation decision rounds, (B4) escalation/budget decision rounds, (B5) the X1 kickoff-adjacent set above (card issuance, scope/schedule, review lenses), (B6) card pick / promotion rounds. |
| Where recorded | Same session-transcript corpus, same roots; each candidate round is an AskUserQuestion tool_use with a matched tool_result (the t1244 pairing predicate generalizes beyond the kickoff keyword). |
| Label variance | Unknown (M0 census). Known count: ≥ 759 non-kickoff AUQ rounds exist in the snapshot window — volume is not the risk; variance is. Some surfaces are plausibly also approve-biased ("Proceed?" rounds), others (B1, B3) plausibly carry real hold/reject mass. |
| Feasibility risks | (i) Heterogeneous decision semantics — each surface needs its own extraction predicate and its own label space before "accuracy" means anything; a mixed-surface population would be ill-defined. (ii) Ground truth = operator label, so the population inherits operator habit bias — the very failure shape t1244 hit, just at a hopefully lower baseline. (iii) Per-surface N may be small even when the corpus total is large. |
| Disposition | **Census-gated candidate.** Pin as primary only if the M0 census finds a single surface with baseline ≤ 90% AND achievable N (REQ-DM-002 pinning rule). Otherwise available as a supplementary natural batch. |

### Candidate C-C — synthetic control group (defect-injected kickoff-shaped cases)

| Aspect | Assessment |
|---|---|
| Surface | Constructed: take real kickoff-gate question payloads (scrub-passing projection) as base material, inject defects of known classes, pair with untouched controls. |
| Where recorded | Not extracted — constructed by the run-phase instrument under the criteria commit; base material from the transcript corpus; the construction itself is evidence at `.moai/reports/t1261/`. |
| Label variance | **Guaranteed by construction**: ground truth per item is fixed at construction time. Balanced 50:50 defect:clean gives constant always-approve baseline = 50%, restoring a meaningful +10%p band (60%). Per-class ground truths (REQ-DM-004) re-introduce genuine 4-value label variance instead of a binary degenerate set. |
| Feasibility risks | (i) Ecological validity — synthetic defects may be easier or harder than real ones; a judge scoring 95% on synthetic says little about 99%-baseline natural behavior unless difficulty is calibrated. Mitigation: a pre-fixed pilot arm (n=20, difficulty-diagnostic only) with the pre-fixed revision path: if the pilot shows a degenerate ceiling/floor, the construction is revised under a NEW criteria commit BEFORE any main-batch call — never post-hoc. (ii) Constructor bias — defects injected where the constructor expects the judge to look. Mitigation: the defect taxonomy is pre-fixed in the SPEC body (≥ 6 classes) and items are class-balanced. (iii) Payload provenance — base payloads leave the machine, so the scrub gate applies to construction inputs too. |
| Disposition | **Recommended primary** — the only candidate where variance is guaranteed and the power arithmetic is structurally satisfiable. |

### Ranking (research-phase recommendation, to be validated by the M0 census before pinning)

1. **C-C synthetic control** — recommended primary: variance by construction, baseline 50%,
   power satisfiable at achievable N (§3).
2. **C-B best qualifying natural surface** — pin as primary whenever it qualifies (census
   baseline ≤ 90% AND N achievable): a qualifying natural surface beats the synthetic control
   (REQ-DM-002, ecological validity); the lowest baseline breaks ties among qualifying naturals.
   The synthetic control is the fallback when no natural candidate qualifies.
3. **C-A** — context only; future-window supplement at best.

## §3. Power-arithmetic framework (worked arithmetic)

Notation: pinned population N; operator/construction ground-truth labels; constant baselines
p_c(approve) = share of approve labels, p_c(hold) = share of hold labels. Per-arm qualification
band conditions inherit the t1244 shape and are RE-DERIVED against the pinned population's
measured baseline in the criteria commit: (a) n ≥ 20 per band, (b) accuracy ≥ p_c + 10%p,
(c) wrong-automation (ground-truth hold answered approve) ≤ 10%.

**Satisfiability rule (the t1244 lesson, generalized).** Condition (b) is satisfiable iff
p_c ≤ 90%. This single inequality decided t1244 (99.05% → 109.05% impossible) and decides the
candidate ranking here.

**Natural populations at plausible non-approve prevalences** (prevalence q = share of
non-approve labels; p_c = 1 − q):

| q | p_c | band (b) = p_c + 10%p | Satisfiable? | Note |
|---|---|---|---|---|
| 1% (t1244 measured ≈ 0.95%) | 99.05% | 109.05% | NO | The t1244 population; structurally dead. |
| 5% | 95% | 105% | NO | Still dead. |
| 20% | 80% | 90% | YES — barely | Requires the census to find a genuinely 20%-variance surface; corpus evidence (kickoff ≈ 1%) makes this unlikely for gate-like surfaces, plausible only for B1/B3-shaped surfaces. |
| 50% (balanced synthetic) | 50% | 60% | YES — meaningful | The balanced construction target; a judge must merely beat coin-flipping-always-approve by 10 points. |

**Two-arm comparison N (paired design).** Both arms judge the SAME items → per-item paired
binary outcomes (correct/incorrect vs ground truth). Let Δ = 10%p be the accuracy difference the
verdict wants to detect between the arms, π_d the discordant-pair rate (share of items where the
arms disagree in correctness), and π the probability a discordant pair favors the better arm:
π = 0.5 + Δ/(2π_d). Normal-approximation planning envelope (two-sided α = 0.05, power 80%,
z_{α/2} = 1.96, z_β = 0.84): required discordant pairs n_d ≈ (z_{α/2}+z_β)² / (4(π−0.5)²),
total items N ≈ n_d / π_d.

| π_d | π | n_d required | N total |
|---|---|---|---|
| 0.15 | 0.833 | ≈ 18 | ≈ 118 |
| 0.20 | 0.750 | ≈ 32 | ≈ 160 |
| 0.30 | 0.667 | ≈ 71 | ≈ 236 |

All three rows share the stated parameters — two-sided α = 0.05, 80% power, Δ = 10%p (one
parameter set per row set; no row uses a different power level).

**Planning conclusion carried into the SPEC body:** target the pinned population at N = 200
(headroom over the 120–240 envelope); the minimum interpretable batch is N = 120 (balanced 60/60),
at which the minimum detectable difference rises to ≈ 11–12%p at π_d = 0.20 — a number the verdict
must quote when reading a null result at small N. The committed TEST is the exact discordant-pair
binomial (§4); the arithmetic above is the planning envelope only.

**Cost envelope for the arms (order-of-magnitude, from t1244's measured ~2K input tokens per Jev
call and §29 pricing):** N = 200 + controls ≈ 210 items; call_cap = 2 × items per arm-batch. Jev
cross-check ≈ 420 calls × 2K tokens ≈ 840K input tokens ≈ under $0.04 at $42/Btok. LLM-arm cost
depends on the pinned pool (§5) and is re-stated in the criteria commit per pool.

## §4. Statistical method recommendation

**Recommendation: McNemar-style discordant-pair exact binomial test, two-sided α = 0.05, on the
paired per-item outcomes of the two arms.** Reasons it fits this measurement:

1. **The design is paired by construction** — both arms judge identical payloads (REQ-DM-005), so
   the paired comparison is the correct shape and discards between-item variance.
2. **The achievable N is modest** (120–240). At n_d below ~25 the normal approximation to the
   discordant-pair count is unreliable; the exact binomial on discordant pairs stays valid
   throughout the achievable range. (The test is exact; the §3 arithmetic is planning only.)
3. **Per-arm metrics remain simple and comparable with t1244** — agreement accuracy vs ground
   truth, always reported alongside BOTH constant baselines (always-approve, always-hold), and
   wrong-automation ≤ 10% as the safety side-condition. Continuity with the A5 metric family is
   deliberate: the same auditor can read both verdicts side by side.
4. **Confidence-band analysis is secondary here.** t1244's band machinery exists to pick a
   threshold (`jev_min_confidence`); t1261's primary output is a MODE comparison. Confidence
   fields are still recorded and pre-declared (REQ-DM-006) so band views remain possible without
   becoming post-hoc selection: the authoritative Jev confidence source is the response registry's
   choice-label probability field (`probabilities`) — the field REQ-CALIB-008 originally registered
   and the field t1244's sync-audit recomputed with (conclusion-invariant) — and the
   instrument-level `confidence` is auxiliary. t1244's sync-audit D3 found the two fields diverge
   systematically; choosing BEFORE measurement is the fix, and this is it.

## §5. LLM-judge arm pool candidates and recommendation

The `llm` arm needs a named model AND quota pool committed before measurement (dispatch
constraint 4). Candidates:

| Pool | Access | Quota relationship | Production fidelity | Timing |
|---|---|---|---|---|
| **GLM (z.ai)** | `glm_task` MCP tool | Separate from the Anthropic Opus main-session pool | Low-medium: answers "can an LLM judge do this", not "can the production decider model do this" | Immediately executable; fail-open posture documented (`glm_audit`-style: unavailable backend → inconclusive, not error) |
| **Anthropic (Opus)** | Main-session model family | The SAME pool the main session and other lanes live on — usage-limited | **High**: contract mode's `llm` decider in production runs on the main-session model | **Externally gated** — execution timing is the lead's decision (recorded dependency, dispatch constraint 4) |
| **Codex** | `codex_task` MCP tool | Separate pool | Low-medium (different model family from production) | Executable; OPTIONAL/fail-open backend per moai-mcp-tools |

**Recommendation: pin GLM as the committed primary llm arm** — rationale: the measurement must be
executable under the autonomous-kickoff policy (REQ-DM-014) without blocking on an externally
gated quota window, and the GLM pool's quota separation keeps the measurement from competing with
lane traffic. **Pre-declared supplement:** if the lead grants an Anthropic Opus quota window
during the run, ONE additional Opus batch may run under the same protocol as its own batch with
its own caps declared before its first call — never replacing the GLM batch, pools never merged in
analysis; the verdict's mode recommendation is drawn from the committed primary pool and the Opus
batch is reported as production-fidelity supplement. Codex is the named fallback if GLM is
unavailable (same one-batch, own-caps discipline). This keeps pre-registration intact: the
supplement is a pre-declared conditional RULE in the criteria commit, not a post-hoc batch.

## §6. t943 / §30 disjointness (scope hygiene)

- **No re-experiment.** §30 prohibits re-running the t943 premise-triage measurement; its numbers
  (124-card KO accuracy, `premise_dead` precision, 75.0% constant baseline) and its 0.50 gate are
  NOT cited as validated premises for Korean anywhere in this SPEC — they appear only as
  discipline sources with the constant-baseline row alongside (REQ-DM-011).
- **No translation arm.** §30's English control group already measured the language effect
  (~10%p, conclusion-invariant, with the 40-card subsample caveat). t1261 measures Korean
  originals only (REQ-DM-010).
- **New construct, disjoint scope notes.** t943's domain was card premise-triage (verdict as
  decision). t1244's domain was Jev-vs-operator AGREEMENT on the kickoff population (judge
  quality). t1261's domain is DECIDER QUALITY — llm vs llm+jev as the contract-mode kickoff
  default — on a discriminable population. An auditor must not be able to conflate the three; each
  SPEC body states its domain and its neighbors' boundaries (REQ-DM-011).

## §7. Open markers

- **NC-1** — the mode-comparison minimum difference (5%p, REQ-DM-007) and the band increment
  (+10%p) are measurement-before judgment values, chosen from the t1244 family precedent; the
  operator may change them BEFORE the criteria commit only, with a HISTORY record. Default: keep.
- **NC-2** — the pinning rule's natural-over-synthetic precedence (REQ-DM-002) is a judgment value
  grounded in ecological validity. Default: keep. Both are body decisions under the
  autonomy policy (REQ-DM-014), not operator blockers.
