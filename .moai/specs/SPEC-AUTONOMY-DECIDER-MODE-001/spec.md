---
id: SPEC-AUTONOMY-DECIDER-MODE-001
title: "Contract-mode kickoff decider default measurement — llm vs llm+jev on a discriminable population (card t1261, AUTONOMY-A5b)"
version: "0.2.0"
status: completed
created: 2026-09-26
updated: 2026-09-27
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".moai/specs/SPEC-AUTONOMY-DECIDER-MODE-001"
lifecycle: spec-anchored
tags: "autonomy, autonomy-a5b, decider, llm-judge, llm-jev, jev, kickoff, measurement-design, population-redefinition, power-arithmetic, mcnemar, pre-registration, synthetic-control, korean-original"
tier: M
depends_on: [SPEC-AUTONOMY-KICKOFF-CALIB-001, SPEC-AUTONOMY-CONTRACT-001]
related_specs: [SPEC-AUTONOMY-GATE-REWIRE-001, SPEC-AUTONOMY-TIERS-001]
---

# SPEC-AUTONOMY-DECIDER-MODE-001 — Contract-mode kickoff decider default measurement (card t1261, AUTONOMY-A5b)

## HISTORY

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-09-26 | 0.1.0 | Initial authoring (plan phase — measurement design with first-class population redefinition). Baseline tree: card worktree branch `WT-decider-mode-eval` HEAD `422aa1e20` (= local develop tip at fork; merge-base confirmed 2026-09-26). Card text (verbatim, operator decision 2026-09-26): 「[운영자 결정 09-26 · t1244 분할 · Tier M · 클래스 C] contract 모드 킥오프 판단자 기본값(llm 단독 vs llm+jev) 결정 근거 측정 — t1244 가 운영자 킥오프 기록 105건 중 104건 승인(always-approve 99.05%)이라 이 모집단으로는 어떤 판사도 기준선+10%p 를 넘을 수 없음을 관측(.moai/reports/t1244/). 범위: 판별 가능한 모집단 재정의 먼저(보류·거절이 실제로 있는 라운드 확대, 다른 게이트, 또는 합성 대조군) — 기준선 대비 검정력이 있는지 측정 전에 산술로 보이고, 그 뒤 llm 판사 팔(모델·한도 풀 명시)과 llm+jev 팔을 같은 프로토콜로. §30 t943 재실험 금지 준수.」 No product code, `internal/`, or template changes — protocol-and-analysis card. Population census is run-phase work (M0), deliberately NOT performed at plan phase. | manager-spec |
| 2026-09-27 | 0.2.0 | Operator-approved criteria amendment (operator decision 2026-09-27, lane AskUserQuestion round, option 「기준 재정의 후 재측정 (권장)」) after the first M2 run was voided under the then-binding both-arms judge-control rule — control t1261-005 failed on the llm+jev composite arm, and the 9-control diagnostic (runs/, void_run_diagnostic-tagged) showed the Jev side a CONSTANT hold-responder (9/9 hold, clean 0/3) while GLM was scattered-correct 5/9: judge accuracy, the measurement's dependent variable, was wired into the validity gate (circular). (1) REQ-DM-004 pilot rule reworded — degeneracy binds to CONSTRUCTION-level indicators (population-wide ceiling/floor across arms); an individual arm being a constant responder is a measured verdict finding, not a revision trigger. (2) plan.md M2 judge-control void rule redefined — judge controls are measured and published as DATA per-item for both arms, no longer gate on label accuracy; the void gate binds to PIPELINE validation only (transmission failure / response not parsing to the committed format / composite not computable). Population, comparison predicate, per-arm metrics, confidence fields, band predicate, mapping, and pools UNCHANGED (REQ-DM-006's frozen four untouched). The restarted run re-declares caps with a fresh declared_at before its first call and takes THIS commit as its criteria_commit. Full record: progress.md §E.2 amendment entry. | manager-spec |

## §A. Background

| Card / source | Owner | Relationship to this SPEC |
|---|---|---|
| t1244 (A5, `SPEC-AUTONOMY-KICKOFF-CALIB-001`, 0.2.1, **completed**) | A5 measurement family | **Direct predecessor (depends_on).** Its verdict (worktree copy at `.moai/reports/t1261/reference/verdict.md`) established: kickoff population 105 rounds (approve 104 / modify 1 / hold 0), always-approve constant baseline 99.05%, Jev judge agreement 91.43% (NOT a constant responder), qualification bands = 0 because band condition (b) = 109.05% is impossible, and — decisive for this card — the sweep covered BOTH corpus roots in full, so no more kickoff rounds exist in the snapshot. Claim 5 transferred the llm-vs-llm+jev question HERE. Its Residual-risk names the structural limit this SPEC is built around: any population whose constant baseline exceeds 90% makes a baseline+10%p condition unsatisfiable by design. Its run discipline (criteria_commit pinning, three caps, scrub gate with positive control, instrument-repair void rules, 5-section verdict) is inherited. |
| t1234 (A1, `SPEC-AUTONOMY-CONTRACT-001`, 0.5.2, **completed**) | Decider schema + config defaults | **Predecessor (depends_on) — defines what is being measured.** In this tree (anchor `e4ea8eb05` reachable, `git merge-base --is-ancestor e4ea8eb05 HEAD` exit 0, re-verified 2026-09-26): decider set `human \| llm \| llm+jev` (§C.8, operator decision 2026-09-26, set final); absent `decider` derives `human` under `guided`, **`llm` under `contract`** (REQ-CONTRACT-015) — so the current contract default IS `llm` 단독; `jev_min_confidence` default 0.50 (the value whose evidence base t1244 found wanting); `llm+jev` falls back to `llm` on Jev-side failure, recorded in the receipt. Changing the default is A1-owned follow-up work; this SPEC produces evidence only. |
| t1236 (A3, `SPEC-AUTONOMY-GATE-REWIRE-001`) | Cross-check rules | **Card context only — its artifacts are NOT in this tree** (plan-complete on a separate branch). From t1244's records: A3 owns the cross-check agreement rules (both approve → start; disagreement → a human) and the fallback decision, and does not wait for the A5-family measurement. This card's verdict is decision EVIDENCE for that design and for the operator's default choice; it implements nothing A3 owns. |
| `CLAUDE.local.md` §29 / §30 | Jev tooling discipline / t943 prohibition | **Discipline sources.** §29: Jev is local-only (`scripts/jev/` untracked), key at `~/.moai/.env.typesafe` separate from the Anthropic quota pool, input-only pricing (~2K tokens/card, <$0.01/batch). §30: the t943 premise-triage measurement is REJECTED — never re-run, never cite its rows without the constant-baseline row, the rejection covers "used as a decision" only. t1261's Jev usage is a NEW construct (decider cross-check quality) — see REQ-DM-011 for the disjointness record. |

### §A.1 The question this card answers

Under `mode: contract` with `workflow.autonomy.kickoff.decider` absent, the effective kickoff
decider today is `llm` — the main-session LLM alone (A1 0.5.2 REQ-CONTRACT-015 derivation). The
card asks whether that default should instead be `llm+jev` (the LLM answer cross-checked by Jev).
t1244 could not answer it: its population (operator kickoff rounds) is 99.05% always-approve, so
no judge can beat baseline+10%p there — a POPULATION power problem, not a judge-quality problem.

This SPEC therefore makes **population redefinition with shown power arithmetic BEFORE
measurement** its first-class milestone: the run MUST measure each candidate population's label
distribution and constant baseline, show the two-arm comparison is discriminable (or declare
판별 불가), pin ONE population as committed criteria — all before the first judge call — and only
then run the two arms (llm 단독 vs llm+jev) on identical payloads under one protocol.

### §A.2 Operator decision — gate disposition (autonomous kickoff, card t1266 policy)

Operator policy (card t1266; lead dispatch 2026-09-26, "킥오프 자율"): this card's measurement run
proceeds autonomously after plan-audit; card-specific condition choices (AC disposition, design
candidates, LIVE budget) take the recommended option WITH rationale recorded in the SPEC body;
anything user-surface or hard-to-reverse escalates through the lead, never directly to the
operator. This mirrors the t1244 record (A5 §A.2) and is recorded FOR THIS DISPATCH ONLY — not a
doctrine generalization (REQ-DM-014).

## §B. Scope

1. **Design layer (this SPEC body)** — candidate population inventory and the pre-fixed pinning
   rule; power-arithmetic requirements; per-arm qualification bands; the llm+jev composite
   decision rule and the mode-recommendation mapping; the confidence-field declaration; the llm
   arm pool with model and quota pool; both-arms-one-protocol discipline; scrub gate; LIVE caps;
   controls; evidence layout. All fixed before measurement.
2. **Execution layer (run-phase, deferred)** — M0 population census and pin (criteria commit),
   instruments and scrub gate, control arms, both measurement arms, analysis, and the verdict at
   `.moai/reports/t1261/verdict.md`. All instruments are disposable tools under the untracked
   report directory; product code, `internal/`, and template mirrors are untouched.
3. **Deferred-by-design** — the transcript-corpus census is run-phase (M0) work; plan phase
   enumerates and assesses candidates from t1244 artifacts only.

### Out of Scope

### Out of Scope — decider implementation and default change

- Changing `workflow.autonomy.kickoff.decider`'s default derivation, the config schema, the
  template default, or `jev_min_confidence` is A1-owned follow-up work (a new card). This SPEC
  measures and recommends; it changes nothing in `internal/contract` or any config surface.
- Implementing the cross-check agreement rules, the fallback decision, or lifting the interim
  `receipt_requires_human` rule is A3's (t1236). The verdict feeds that design; it does not
  perform it.

### Out of Scope — t943 re-experiment and translation arm

- Re-running the t943 premise-triage experiment, re-measuring its numbers, or treating its 0.50
  gate as a validated premise for Korean (§30 prohibition; REQ-DM-011).
- An English-translation arm of the corpus (§30's English control group already measured the
  ~10%p language effect, conclusion-invariant; REQ-DM-010).

### Out of Scope — plan-phase corpus census and whole-history claims

- Sweeping or census-sampling the transcript corpus at plan phase — that is M0 run work under the
  criteria commit. Plan-phase evidence is the t1244 artifact record only.
- Any "whole history" claim: the transcript corpus is a truncating snapshot; the verdict speaks
  for its snapshot date only.

### Out of Scope — other gates' automation

- Automating plan-audit, sync, blocker, or escalation decisions, or extending Jev usage to any
  gate other than the kickoff decider question. A qualifying natural surface from candidate C-B,
  if ever pinned, is measured as kickoff-decider EVIDENCE only — it opens no other automation.

## §C. Requirements (GEARS)

### REQ-DM-001 — Population redefinition precedes any judge call (When)

**When** the measurement run begins, the run shall, **before the first judge call of the entire
measurement**, (a) measure each candidate population's label distribution and constant-response
baseline from the corpus, (b) compute the two-arm power arithmetic for each candidate per
REQ-DM-003, and (c) pin exactly ONE population by the REQ-DM-002 pinning rule — or declare 판별
불가 — committing all three as criteria (REQ-DM-015 criteria_commit) before any judge call. The
candidate axes, with their plan-phase assessments (full inventory: research.md §2): **(C-A)
kickoff-gate rounds with holds/rejections** — rejected as primary: the t1244 sweep covered both
corpus roots in full, so within the kickoff gate no more rounds exist in the snapshot; "라운드
확대" can only mean other decision surfaces or future accumulation, and the measured baseline
99.05% is structurally dead for a +10%p band; **(C-B) other decision surfaces with recorded
operator decisions and real label variance** — census-gated: concretely, plan-audit-failure
user-choice rounds, sync confirmation rounds, blocker re-delegation rounds, escalation/budget
rounds, the t1244 X1 kickoff-adjacent excluded set (card issuance, scope/schedule, review
lenses), and card-pick rounds; the same snapshot window holds ≥ 759 non-kickoff AskUserQuestion
rounds (t1244 probe: 907 total AUQ blocks, 148 kickoff-mentioned); each surface needs its own
extraction predicate and label space before accuracy is defined; **(C-C) synthetic control
group** — defect-injected kickoff-shaped cases with known ground truth, balanced so the constant
baseline is beatable (REQ-DM-004). The census records, per candidate: the decision surface, the
extraction predicate applied, measured label distribution, measured constant baselines, and the
power-arithmetic verdict — undercounts surface as recorded exclusions, never silence
(exclusion lists preserved in the run record, the REQ-CALIB-001 X1/X2 pattern).

### REQ-DM-002 — Pinning rule and the 판별 불가 exit (Ubiquitous)

The measurement shall apply this pre-fixed pinning rule, committed before measurement: **pin the
best qualifying candidate: a qualifying natural corpus surface always outranks the synthetic
control (ecological validity); among multiple qualifying natural surfaces, the lowest measured
constant baseline is the tie-break; qualify = constant baseline ≤ 90% AND the required N from the
REQ-DM-003 arithmetic is achievable within the REQ-DM-013 caps.** The synthetic control (baseline
50% by construction) is the recommended primary absent a qualifying natural candidate. **If NO candidate qualifies, the card's VALID
termination is a documented 판별 불가 verdict** carrying: the per-candidate census numbers, the
per-candidate arithmetic, and why each fails — a result, not a failure (t1244 ended exactly this
way and its verdict is the format precedent). The 판별 불가 path satisfies this SPEC's acceptance
criteria when its evidence requirements are met (AC-DM-016). The operator may override the
pinning rule's output before the criteria commit only, with rationale recorded (REQ-DM-014
autonomy policy); after the criteria commit the pin is frozen.

### REQ-DM-003 — Power arithmetic committed before measurement (Ubiquitous)

The measurement shall include, in the committed criteria for the pinned population: the measured
label prevalence, both constant baselines (always-approve, always-hold), the paired comparison
method (REQ-DM-006), the required N, and the go/no-go band predicate re-derived so it is NOT
structurally unsatisfiable. The arithmetic seed (worked in research.md §3): a constant baseline
above 90% makes any baseline+10%p condition unsatisfiable — the t1244 lesson; at a balanced 50/50
construction the baseline is 50% and the +10%p band becomes 60% — meaningful again; detecting a
two-arm accuracy difference of Δ = 10%p at two-sided α = 0.05 with 80% power needs roughly N ≈
120–240 items across plausible discordant-pair rates (π_d 0.15–0.30; ≈ 118 at 0.15, ≈ 160 at
0.20, ≈ 236 at 0.30), so the pinned population targets N = 200, with N = 120 the minimum
interpretable batch
(minimum detectable difference ≈ 11–12%p there — quoted whenever a null is read at that size).
The committed test itself is the exact discordant-pair binomial (REQ-DM-006); this arithmetic is
the planning envelope and is committed as such.

### REQ-DM-004 — Synthetic control construction (Where the pinned population is the synthetic control)

**Where** the pinned population is the synthetic control, the run shall construct it as follows,
all pre-fixed in the criteria commit: base payloads are real kickoff-gate question payloads from
the corpus, carried as the scrub-passing projection (REQ-DM-012) with the operator response
removed; the defect taxonomy has at least six pre-fixed classes — (i) mandatory-section omission,
(ii) scope violation, (iii) violated precondition, (iv) fabricated-evidence claim, (v) AC-count
mismatch, (vi) internal contradiction between acceptance criteria — and items are class-balanced;
the defect:clean ratio is 50:50, so the always-approve constant baseline is 50%; each defect
class pre-fixes its ground-truth label from the 4-value space (clean → `approve`; defect classes
→ the label a correct gate would return, e.g. missing mandatory section → `hold`, scope violation
→ `modify`), so the label space keeps genuine 4-value variance; **the pilot rule**: a
difficulty-diagnostic pilot arm (n = 20) may run before the main batch — CONSTRUCTION-level
degeneracy is what triggers revision, meaning a population-wide ceiling or floor ACROSS arms
(both arms ≥ 90% or ≤ 10% overall, or every defect class trivially caught — or trivially
missed — by both); the construction is then revised under a NEW criteria commit BEFORE any
main-batch call, and pilot results are recorded but never merged into main-batch statistics. An
INDIVIDUAL arm being a constant or weak responder on the pilot (e.g. one judge answering hold
throughout) is a measured finding for the verdict — NOT construction degeneracy and NOT a
revision trigger (operator-approved amendment 2026-09-27; ground: the voided first M2 run's
diagnostic already showed construction signal via the other arm — clean 3/3, defects caught
2/6, neither ceiling nor floor).

### REQ-DM-005 — Both arms, one protocol (Ubiquitous)

The measurement shall run the llm 단독 arm and the llm+jev arm on **the same pinned population,
the same predicate, the same scrub gate, the same caps, the same pre-registration, and identical
payloads** — the llm+jev arm adds only the pre-fixed Jev cross-check stage (REQ-DM-008). The
arms' item sets MUST be equal as payload_id sets (machine-checkable, AC-DM-015); no item is
judged by one arm only. Where a pool produces a malformed or failed response, the item is
re-tried within the per-item turn cap and otherwise recorded unmeasured — never silently dropped
from one arm only (an asymmetry there breaks the pairing the REQ-DM-006 test stands on).

### REQ-DM-006 — 판정식 사전 고정 (predicate pre-fixed; Ubiquitous)

The measurement shall fix ALL of the following in the committed criteria before the first judge
call, with no post-hoc metric selection: **(1) comparison predicate** — the McNemar-style
discordant-pair exact binomial test, two-sided α = 0.05, on the paired per-item
correct/incorrect outcomes of the two arms (rationale: the design is paired by construction and
the exact test stays valid at the achievable N; the REQ-DM-003 arithmetic is the planning
envelope only); **(2) per-arm metrics** — agreement accuracy against ground truth (operator label
on a natural population, construction label on the synthetic control), reported ALWAYS alongside
both constant baselines (REQ-DM-006 inherits the REQ-CALIB-006 never-report-alone rule), and
wrong-automation (ground-truth `hold` answered `approve`) ≤ 10%; **(3) confidence fields,
declared before measurement** — for the Jev cross-check, the authoritative confidence source is
the response registry's choice-label probability field (`probabilities`); the instrument-level
`confidence` field is recorded as auxiliary ONLY — t1244's sync-audit D3 found the two diverge
systematically, and the fix is declaring the authoritative field BEFORE measurement; for the llm
arm, the authoritative confidence field is declared from its committed response format in the
criteria commit; **(4) band predicate** — the per-arm qualification band re-derived from the
pinned population's measured baseline in the criteria commit: (a) band n ≥ 20, (b) accuracy ≥
measured baseline + 10%p, (c) wrong-automation ≤ 10% (t1244 band shape, re-based). Changing any
of these after the first judge call is prohibited — the threshold-raising trap t943 already
reproduced (raising a gate reduces adoption, does not return accuracy).

### REQ-DM-007 — Mode recommendation predicate (When)

**When** both arms' results are in, the measurement shall map them to the default recommendation
by this pre-fixed rule, committed before the first judge call. **Composite definition (llm+jev
arm):** per item, the llm label and the Jev label are combined — agreement → that label;
disagreement → `hold` (defer to a human). This mirrors the cross-check shape recorded in A1 §C.8
and is deliberately confidence-free: `jev_min_confidence` gating is NOT applied inside the
composite (that threshold is exactly what the A5 family lacks evidence for; Jev confidence is
recorded per REQ-DM-006 for supplementary band views only). **Recommendation mapping** — with
δ = 5%p as the minimum decision-relevant difference (measurement-before judgment value, NC-1):
**"llm+jev better"** = composite accuracy ≥ llm accuracy + δ AND the REQ-DM-006 test significant
AND composite wrong-automation ≤ llm wrong-automation → recommend default `llm+jev`;
**"llm+jev worse"** = llm accuracy ≥ composite accuracy + δ with the test significant, OR
composite wrong-automation > 10% → recommend default `llm 단독`; **"equal"** = neither → recommend
default `llm 단독` (simpler, one fewer dependency, and the burden of proof for adding a
cross-check lies with the cross-check). Any recommendation names the pool it is drawn from
(REQ-DM-009); pools are never merged in analysis.

### REQ-DM-008 — The Jev cross-check stage (Ubiquitous)

The measurement shall run the llm+jev arm's Jev stage on the same scrub-passing payload
projection as the llm arm, using the t1244-repaired instrument pattern: the TypeSafe `choice`
form WITH the criteria map (the `ask.sh` criteria-omission defect of 2026-09-26 is known — the
direct-POST pattern with the criteria map is the committed instrument shape), model
`jev-latest`, key at `~/.moai/.env.typesafe` (separate from the Anthropic quota pool), raw
responses preserved unprocessed under `runs/`. The Jev stage receives the operator response of
the base corpus item ONLY where the construction intends it — on the synthetic control the Jev
stage receives the same constructed payload as the llm arm and no ground-truth hint. The 4-value
label space (`approve | hold | modify | other`) and the M1–M4 mapping pre-resolution rules are
inherited from REQ-CALIB-002 so both arms' outputs and the t1244 record stay comparable.

### REQ-DM-009 — LLM arm pool: model and quota pool named (Where the llm arm runs)

**Where** the llm 단독 arm runs, the run shall use the pool committed in the criteria commit:
**primary pool = GLM (z.ai), via the `glm_task` MCP tool** — chosen because its quota is separate
from the Anthropic Opus main-session pool (the measurement does not compete with lane traffic),
it is immediately executable under the autonomous-kickoff policy, and its fail-open posture
(unavailable backend → inconclusive, not error) is documented; **pre-declared supplement = one
Anthropic (Opus) batch, run only if the lead grants a quota window** — the production-fidelity
pool (contract mode's `llm` decider runs on the main-session model), whose execution timing is
EXTERNALLY GATED by usage limits, a recorded dependency of this card; the supplement batch is
declared with its own caps before its own first call, never replaces the primary batch, and pools
are never merged in analysis; **named fallback = Codex (`codex_task`)** if GLM is unavailable,
under the same own-batch own-caps discipline. The criteria commit names the primary pool's exact
model identifier and quota pool. A fail-open line from the judge tool records that item
unmeasured; all items unmeasured closes the card with a measurement-impossible verdict (the
REQ-CALIB-008 pattern).

### REQ-DM-010 — Korean original discipline (Ubiquitous)

The measurement shall not translate any population item, payload, or judge instruction at any
stage — corpus items and constructed payloads stay Korean verbatim UTF-8. Grounds: §29's 0.50
gate came from a 16-card sample measured on English translations of Korean cards, and §30's
English control group already measured the language effect (~10%p, conclusion-invariant, with
the 40-card subsample caveats) — a translation arm would repeat a rejected experiment (REQ-DM-011).

### REQ-DM-011 — t943 re-experiment prohibition and scope disjointness (Ubiquitous)

The measurement shall NOT re-run the t943 premise-triage experiment and shall NOT cite its
numbers or its 0.50 gate as a validated premise for Korean; §30 citations always carry the
constant-baseline row alongside (e.g. 2-class 58.9% only next to the constant 75.0% row). The
three same-tool domains stay disjoint in every artifact: **t943** = card premise-triage verdict
quality (rejected for decision use); **t1244** = Jev-vs-operator agreement on the kickoff
population (judge quality; population power-fatal); **t1261** = llm vs llm+jev DECIDER quality
on a discriminable population. Scope notes in the verdict name this card's construct and its
neighbors' boundaries so an auditor cannot conflate them. This card's positive result, if any,
applies to exactly one surface: the contract-mode kickoff decider default evidence — it opens no
Jev sole decision and no other gate's automation.

### REQ-DM-012 — Outbound scrub gate before every external transmission (Ubiquitous)

The measurement shall pass every payload bound for ANY external judge — Jev/TypeSafe, GLM,
Anthropic, Codex — through the mechanical scrub scanner before transmission, fail-closed: (a)
key-shaped strings (`sk-`/`ghp_`/`gho_`/`AKIA`/PEM blocks/`.env` content; this repo's key files
are `~/.moai/.env.typesafe` and `~/.moai/.env.glm`); (b) `.claude/settings.local.json`-family
values (tmux pane ids, token-shaped values); (c) absolute paths; (d) customer/PII lower-bound
patterns (resident-registration, phone, email, card-number shapes — the D12 lower-bound set;
pattern additions found during the run are scrub strengthening, not criteria changes, and are
recorded). The scanner's POSITIVE CONTROL (dummy secrets planted, all classes observed firing —
the §30 three-step ③ shape) runs before the FIRST transmission; a scanner that has never fired
proves nothing. Every payload carries a `payload_id` and a scan row (`clean`/`hit`) in the run
record; a hit BLOCKS transmission — never strip-and-send. The lower-bound limitation of the
pattern set is recorded in the verdict's Residual-risk. Data minimization: the transmission
whitelist is the payload projection only (question text, option labels/descriptions, task
instruction); session dumps and transcript context are never sent. Judge-tool fail-open (missing
key/network) never loosens this gate — no scrub, no transmission.

### REQ-DM-013 — LIVE caps (When)

**When** the measurement runs, it shall declare, in the run record with a `declared_at`
timestamp per cap BEFORE the first judge call: `turn_cap` = 2 calls per item (initial + one
malformed-retry);
`call_cap` = 2 × (pinned N + controls) per arm-batch (Jev cross-check stage counted inside the
llm+jev arm's cap); `wall_clock_cap` = PT8H per batch; `batches` = 1 primary batch per arm (+1
per pre-declared supplement batch, each with its own declared caps). Reaching any cap: write the
verdict with what was measured — NO cap extension, NO re-measurement (the REQ-CALIB-008 rule).
Instrument repair follows the t1244 rule: repair, restart, void prior outputs, preserve the void
rows separately.

### REQ-DM-014 — Autonomous kickoff and bundled conditions (When)

**When** the measurement run executes, it shall run under the operator policy recorded in §A.2
(card t1266; lead dispatch 2026-09-26, "킥오프 자율"): the run proceeds autonomously after
plan-audit, bound by the bundled conditions — **(a) criteria_commit**: the REQ-DM-001/002/003/006/
007/009 criteria are fixed by a commit on this branch whose hash the run record records as
`criteria_commit:`, verified (`git cat-file -e`) and time-compared against the first judge call;
**(b) three caps** declared per REQ-DM-013 before the first call; **(c) scrub gate** passing per
REQ-DM-012 before every transmission. Card-specific condition choices take the recommended
option with rationale recorded; anything user-surface or hard-to-reverse escalates through the
lead, never directly to the operator. This record binds THIS DISPATCH ONLY — not a doctrine
generalization.

### REQ-DM-015 — Evidence layout and the tracked criteria (Where)

**Where** evidence is written, the run shall place it as follows: measurement instruments, raw
outputs, the census, and the run record live under `.moai/reports/t1261/` (worktree-local;
untracked during work — the t1244 precedent); the SPEC artifacts AND the pinned criteria are
tracked and committed on this branch (`criteria_commit` names a real commit on this branch —
verified with `git cat-file -e`, time-compared against the first judge call); the completion
verdict at `.moai/reports/t1261/verdict.md` follows the 5-section evidence-bearing format
(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) — every number in it attributed
to a command run in this run against this tree, constant baselines always alongside judge
numbers; the verdict is what the lead reads, exported to the primary checkout at close. Plan-phase
evidence lives in this SPEC's progress.md and research.md.

## §D. Acceptance criteria and REQ↔AC mapping

All acceptance criteria live in `acceptance.md` (AC-DM-001..016; 15 REQ / 16 AC — AC-DM-001
covers REQ-DM-001+002, AC-DM-004 covers REQ-DM-007+008, AC-DM-008 covers REQ-DM-012+013,
AC-DM-010 covers REQ-DM-014+015, AC-DM-011 covers REQ-DM-010+011 attribution integrity). The
[PLAN]/[RUN] two-layer split follows the A5 pattern: [PLAN] ACs are decidable at plan phase
(design-artifact presence and content); [RUN] ACs are decidable only against run-phase records
and read `--- PENDING-RUN` until then — which is not a FAIL.

| AC | Layer | Covers |
|---|---|---|
| AC-DM-001 | [PLAN] | REQ-DM-001, REQ-DM-002 |
| AC-DM-002 | [PLAN] | REQ-DM-003 |
| AC-DM-003 | [PLAN] | REQ-DM-006 |
| AC-DM-004 | [PLAN] | REQ-DM-007, REQ-DM-008 |
| AC-DM-005 | [PLAN] | REQ-DM-005 |
| AC-DM-006 | [PLAN] | REQ-DM-009 |
| AC-DM-007 | [PLAN] | REQ-DM-004 |
| AC-DM-008 | [PLAN] | REQ-DM-012, REQ-DM-013 |
| AC-DM-009 | [PLAN] | REQ-DM-010, REQ-DM-011 |
| AC-DM-010 | [PLAN] | REQ-DM-014, REQ-DM-015 |
| AC-DM-011 | [PLAN] | t1244 attribution integrity + scope guard (REQ-DM-011, REQ-DM-015 context) |
| AC-DM-012 | [RUN] | REQ-DM-001, REQ-DM-002, REQ-DM-003, REQ-DM-014 |
| AC-DM-013 | [RUN] | REQ-DM-013, REQ-DM-014 |
| AC-DM-014 | [RUN] | REQ-DM-012 |
| AC-DM-015 | [RUN] | REQ-DM-005 |
| AC-DM-016 | [RUN] | REQ-DM-002, REQ-DM-006, REQ-DM-007, REQ-DM-015 |

## §E. Residual risks and open markers

- **Judgment values are pre-registered, not proven.** The +10%p band increment and the δ = 5%p
  mode-comparison minimum are measurement-before judgment values from the A5 family precedent
  (NC-1); the natural-over-synthetic precedence is an ecological-validity judgment (NC-2). Both
  are changeable only before the criteria commit, with a HISTORY record.
- **Synthetic-vs-natural validity gap.** A judge's measured quality on constructed defects bounds
  its quality on real gate decisions only weakly; the verdict's Residual-risk must state this
  whenever the pinned population is the synthetic control, and any default recommendation is
  evidence for the operator, not a self-executing change (the change itself is A1-owned
  follow-up).
- **Judge-pool transfer.** The committed primary pool (GLM) differs from the production decider
  model (main-session Anthropic). The pre-declared Opus supplement narrows this gap only when the
  lead grants a window; otherwise the recommendation is pool-qualified and the gap is a named
  Gaps row.
- **Census undercount.** The M0 census can only see surfaces whose decisions are recorded as
  machine-pairable AskUserQuestion rounds; decisions recorded in other shapes are invisible to it
  and belong in the census's recorded exclusions, never in silence.
- **Corpus truncation.** The transcript corpus is a truncating snapshot (t1244 observed block
  drift between same-day runs); the snapshot date and sweep command are mandatory run-record
  fields, and whole-history claims are out of scope.
- **PII lower-bound set.** The scrub scanner's deny patterns are a lower bound; pattern classes
  outside the set can pass. Recorded in Residual-risk every time (REQ-DM-012).
