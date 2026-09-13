---
id: SPEC-DECISION-AUTHORITY-001
title: "Human Decision Authority Guardrail: decision-index authority routing at the Implementation Kickoff gate (card t692, issue #1683 item 1)"
version: "0.1.0"
status: draft
created: 2026-09-13
updated: 2026-09-13
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: ".claude/agents/moai, .claude/skills/moai/workflows/plan, .moai/config/sections, internal/config"
lifecycle: spec-anchored
tags: "decision-authority, decision-index, authority-routing, kickoff-approval, issue-1683, t692"
tier: L
issue_number: 1683
depends_on: [SPEC-JUDGMENT-FIRST-MODE-001]
---

# SPEC-DECISION-AUTHORITY-001 — Delegate Reasoning, Not Authority

## HISTORY

| Version | Date | Author | Description |
|---|---|---|---|
| 0.1.0 | 2026-09-13 | manager-spec | Initial Tier L authoring: decision-gate mode axis, decision-index.md plan-phase artifact, four-label authority routing, committed-only authority register, kickoff-gate enrichment, verdict recording, pull-mode composition. Card t692, issue #1683 item #1; composes with landed SPEC-JUDGMENT-FIRST-MODE-001 (item #2). |

## §A Context and Problem

### A.1 Origin

GitHub issue `modu-ai/moai-adk#1683` item #1 ("Human Decision Authority Guardrail for SPEC
Review", external submitter, 2026-08-30, `type:feature`). The issue was split by
SPEC-JUDGMENT-FIRST-MODE-001 §A.1 with a fixed [HARD] adoption order: item #2 (analysis is
pull, not push) landed first and narrowly; item #1 — a decision gate inside SPEC Review with
decision-index authority routing — is this SPEC.

The submitter's compatibility note states the attached design targets pre-v3.1.2 legacy EARS
and is **not** proposed for verbatim adoption; what is adopted is the **Human Decision Authority
model**. Its closing thesis is this SPEC's title: **delegate reasoning, not authority.**

The lane decisions D1-D10 (dispatch prompt, card t692) fix the direction; this SPEC refines
wording but does not change direction.

### A.2 The measured problem

Two decision-authority failures recur when a SPEC resolves questions whose authority no
document carries:

1. **Silent confirmation.** manager-spec resolves a product-level question with a reasonable
   default and bakes the resolution into a REQ. The decision was never made by anyone with the
   authority to make it; the artifact merely records that it looks decided.
2. **Anchored judgment.** The operator's first view of an unresolved question is an option list
   with a `(권장)` label already attached, so their judgment is anchored to a preference they
   never asked for. Item #2 (SPEC-JUDGMENT-FIRST-MODE-001, `recommendation_mode`) removed the
   label on the question channel; it did not create a place where unresolved decisions are
   enumerated **before** the gate that approves the plan.

The proposal's core move separates the two failure classes: questions whose answer **exists in
a committed document** are routed to that document (mechanical), and questions whose answer
**exists nowhere** are routed to the human (founder-level) — instead of being silently resolved
by whoever is writing the artifact.

### A.3 D4 measurement — the authority surfaces, verified this tree

The dispatch's D4 premise — "`.moai/reports/**` is gitignored local material — NOT citable
authority" — was **measured and corrected** on this tree (`62fbd6baf`):

- `git check-ignore -v .moai/reports/t692/issue-1683-proposal.md` exits **1** (not ignored).
  The local `.gitignore` ignores `.moai/reports/*.md` (root-level only) and
  `.moai/reports/plan-audit/*.md`, and its line-107 comment declares card evidence under
  `.moai/reports/**` to be **durable audit evidence** — commit-eligible, not junk.
- The proposal file is currently **untracked** (`?? .moai/reports/t692/` in git status).

The corrected authority criterion is therefore **committed, not "not-ignored"**: an authority
anchor must resolve in the committed tree so any later reader can re-verify it with
`git show <ref>:<path>`. An untracked working-copy document — even a card's own evidence —
carries no citable authority, because nothing in history witnesses that anyone else ever read
the same text (the CLAUDE.local.md §0.2 rule, generalized). All requirements below say
"committed", never "gitignored".

The verified register (each surface read this run):

| Authority surface | Verified shape | Governs |
|---|---|---|
| `.moai/project/product.md` | Committed; carries Project Overview / Mission / Vision / Target Audience / Core Features | Product intent — the closest PRD analog this repo has |
| Prior completed SPECs | Committed under `.moai/specs/`; HISTORY tables + `## Amendments` sub-sections record prior verdicts with dates and SHAs (e.g. SPEC-JUDGMENT-FIRST-MODE-001 HISTORY 0.1.0-0.2.4) | Prior founder verdicts (`DECIDED`) |
| `.moai/config/sections/*.yaml` | Committed; 32 section files; operator settings (e.g. `interview.recommendation_mode`, `git_strategy.manual.workflow`) | Explicit operator policy (`POLICY-COVERED`) |
| Project constitution | Committed doctrine: `AGENTS.md` + `.claude/rules/moai/core/moai-constitution.md` | Governing principles |

### A.4 The three adopted conditions (verbatim, binding)

Inherited unchanged from the sibling SPEC §A.1; every requirement below is read through them:

1. **Detect → Explain → Ask, but never decide.**
2. **An LLM "best practice" is not a policy.**
3. **When uncertain, escalate. Never downgrade.**

## §B Solution Shape

A **decision-gate mode axis** activates a plan-phase **decision-index** artifact that routes
every unresolved decision to the authority that actually owns it, presented at the
Implementation Kickoff Approval gate the project already owns.

```yaml
interview:
  decision_gate: off   # off (default = exact status quo) | on
```

**D1 — Gate placement (enrichment, not a new gate).** The proposal's "Human Decision Gate" is
implemented as an **enrichment of the existing Implementation Kickoff Approval gate**
(`.claude/skills/moai/workflows/plan/spec-assembly.md` — the `[HARD] The Implementation
Kickoff Approval AskUserQuestion gate stays MANDATORY` clause at `:217`). MoAI already owns a
mandatory, score-independent human gate at exactly the decision point the proposal targets. A
second gate would be a pipeline-structure change (excluded, §F).

**D2 — Index carrier.** A new plan-phase artifact `decision-index.md` inside the SPEC
directory, authored by manager-spec while the mode is on. It is **stateless** on the status
axis per `.claude/rules/moai/development/spec-frontmatter-schema.md` § Artifact Statelessness
— no `status:` field; the SPEC's lifecycle lives in `spec.md` alone. It is **not** a
plan-auditor report section (excluded, §F).

Index row shape (fixed, so two authors produce the same file):

```
### Q<N>: <the decision, stated as a question>
- Label: DECIDED | POLICY-COVERED | EVIDENCE-NEEDED | FOUNDER
- Authority anchor: <file> §<section>   (DECIDED/POLICY-COVERED only — file + section that exists in the committed tree)
- Why unresolved: <one line — what the documents do not answer>
- Operator verdict: <recorded at kickoff — see REQ-DA-015>
```

**D3 — Labels.** Exactly the proposal's four tokens, verbatim: `DECIDED` / `POLICY-COVERED` /
`EVIDENCE-NEEDED` / `FOUNDER`. No second vocabulary. The proposal's founder-action column
(`NONE`, `NONE_UNLESS_OVERRIDE`, `REQUEST_EVIDENCE`/`TEMPORARY_VERDICT`/`DEFER`,
`DECIDE`/`NEED_ANALYSIS`/`NEED_EVIDENCE`/`DEFER`) is folded into the verdict-action vocabulary
of REQ-DA-015 rather than duplicated as a second per-row column.

**D4 — Authority register.** COMMITTED artifacts only (§A.3): `.moai/project/product.md`,
prior completed SPECs' HISTORY / `## Amendments` rows, `.moai/config/sections/*.yaml` operator
settings, and the project constitution (`AGENTS.md` + `.claude/rules/moai/core/`). Untracked
material — `.moai/reports/**` working copies included — is not citable authority.

**D5 — Flow owner.** The flow change lands in manager-spec's plan phase
(`clarity-interview.md` + `spec-assembly.md`): a decision whose authority no document carries
becomes an index row instead of being silently resolved. **plan-auditor is UNCHANGED in v1** —
the t401 §F report-contract exclusion binds until separately justified; the
auditor-integration question is recorded as an explicit deferral (§E.2), the same way t401
deferred this card.

**D6 — Mode axis.** `interview.decision_gate: on|off`, distributed default `off` = exact status
quo (first-class regression property, AC-DA-004). No automatic or heuristic mode selection
(excluded, §F). The axis lives beside `interview.recommendation_mode` and is **orthogonal** to
it (REQ-DA-018).

**D7 — Verdict recording + zero-flag rule.** The kickoff gate presents the index; operator
verdicts are recorded back into decision-index.md rows. Product-level verdicts additionally
require reconciling `.moai/project/product.md` — and the project-doc ownership boundary is
**surfaced at kickoff as a design decision the operator settles** (§E.3), because
`.moai/project/**` scaffolding belongs to manager-docs per the canonical agent catalog, not to
manager-spec. Zero judgment points is not approval: the gate remains a separate affirmative act
(REQ-DA-019).

**D8 — Pull composition.** Index rows carry Detect → Explain → Ask, never an embedded AI
recommendation. In `pull` mode nothing is recommended (the landed convention, unchanged); in
`push` mode the existing recommendation rules are unchanged — the index itself never carries a
preferred answer in either mode.

**D10 — Template surfaces.** Edit surfaces: `.claude/agents/moai/manager-spec.md`,
`.claude/skills/moai/workflows/plan/spec-assembly.md`, `plan/clarity-interview.md`,
`.moai/config/sections/interview.yaml` (template section), plus the index-format reference in
the plan-workflow skill tree. Each carries its mirror under
`internal/template/templates/**` (REQ-DA-020). Measured this tree: `spec-assembly.md` and
`askuser-protocol.md` mirrors are byte-identical; `manager-spec.md`, `clarity-interview.md`,
and `interview.yaml` mirrors already diverge from their local copies (intentional C1↔C2
branching per CLAUDE.local.md §2.0) — so the mirror obligation here is **the same change lands
in both trees**, never byte-identity (AC-DA-018).

## §C Requirements (GEARS)

### C.1 The mode axis

**REQ-DA-001** (Ubiquitous) — The decision-gate axis shall carry exactly two values, `on` and
`off`, exposed as `interview.decision_gate` in `.moai/config/sections/interview.yaml`.

**REQ-DA-002** (Where — capability gate) — Where `interview.decision_gate` is absent, empty,
or `off`, the plan-phase workflow shall behave exactly as it does at base commit `62fbd6baf`,
with no observable behavioral difference: no decision index is authored, no index rows are
collected, and the Implementation Kickoff Approval gate composes unchanged.

**REQ-DA-003** (Where — capability gate) — Where `interview.decision_gate` holds a value that
is neither `on` nor `off`, the config loader shall resolve the axis to `off` and shall record
the unrecognized value, rather than failing the load.

**REQ-DA-004** (Ubiquitous) — The distributed template shall ship the axis default as `off`,
and the template mirror of `interview.yaml` shall remain behaviorally identical to its state
at base commit `62fbd6baf` for every consumer that does not set the key.

### C.2 Index authoring

**REQ-DA-005** (While — state-driven) — While the decision gate is `on`, when manager-spec
executes its plan phase, manager-spec shall author `decision-index.md` inside the SPEC
directory, containing one row per decision encountered during clarification or assembly whose
authority no document carries, in the fixed row shape of §B.

**REQ-DA-006** (Ubiquitous) — The `decision-index.md` artifact shall be stateless on the
status axis: it shall carry no `status:` frontmatter field, per
`.claude/rules/moai/development/spec-frontmatter-schema.md` § Artifact Statelessness.

**REQ-DA-007** (Where — capability gate) — Where the decision gate is `off`, manager-spec
shall not create `decision-index.md`, and no plan-phase artifact, gate, or workflow text shall
reference a decision index.

### C.3 Labels and the authority register

**REQ-DA-008** (Ubiquitous) — The decision index shall route every row using exactly the four
labels `DECIDED`, `POLICY-COVERED`, `EVIDENCE-NEEDED`, and `FOUNDER`, and no second label
vocabulary shall be introduced.

**REQ-DA-009** (Ubiquitous) — Every `DECIDED` or `POLICY-COVERED` row shall cite a concrete
authority anchor — a file plus a section that exists in the committed tree — and an
unverifiable or absent anchor disqualifies the row from those two labels.

**REQ-DA-010** (Ubiquitous) — The authority register shall consist of committed artifacts
only: `.moai/project/product.md`, prior completed SPECs' HISTORY and `## Amendments` rows,
`.moai/config/sections/*.yaml` operator settings, and the project constitution. Untracked
material, including card evidence under `.moai/reports/` that has not been committed, shall
not be cited as authority.

**REQ-DA-011** (When — event-detected) — When a row's candidate anchor cannot be verified in
the committed tree, the row shall route to `FOUNDER` — it shall not be downgraded to an
implementation detail, and shall not be relabeled `DECIDED` or `POLICY-COVERED` on the strength
of an unverifiable citation.

### C.4 Flow owner and the kickoff gate

**REQ-DA-012** (While — state-driven) — While the decision gate is `on`, when the plan phase
(clarity interview or SPEC assembly) encounters a decision whose authority no document
carries, manager-spec shall record it as a decision-index row instead of resolving it with a
reasonable default.

**REQ-DA-013** (Where — capability gate) — Where the decision gate is `on`, the Implementation
Kickoff Approval gate shall present the SPEC's decision index as part of its review surface,
as an enrichment of that gate; the gate itself — mandatory, score-independent, its three
canonical options and its existing option structure — shall remain unchanged, and no second
human gate shall be introduced.

**REQ-DA-014** (Ubiquitous) — `plan-auditor`'s report contract shall remain unchanged in v1:
no decision-index section, no judgment-point finding class, and no change to the
`Required fix:` field or `## Recommendation` section. The auditor-integration question is an
explicit deferral, not an oversight.

**REQ-DA-015** (When — event-driven) — When the operator issues a verdict on a decision-index
row at the kickoff gate, the verdict shall be recorded back into that row of
`decision-index.md`, using the proposal's action vocabulary: `DECIDE`, `NEED_ANALYSIS`,
`NEED_EVIDENCE`, or `DEFER`.

**REQ-DA-016** (When — event-driven) — When a recorded verdict is product-level (MVP or phase
scope, tier behavior, UX flow, pricing, privacy or security promise), the kickoff flow shall
surface the reconciliation of `.moai/project/product.md` as a named design decision — flagging
that `.moai/project/**` scaffolding is owned by manager-docs, so the reconcile act is
delegated to manager-docs or explicitly taken by the operator, and never performed silently by
manager-spec. A SPEC-level verdict (retry count, internal algorithm, query shape) shall touch
the decision index and the SPEC only.

### C.5 Pull composition

**REQ-DA-017** (Ubiquitous) — Every decision-index row shall carry Detect → Explain → Ask and
never an embedded AI recommendation: the index shall state what is unresolved and why, and
shall not carry a preferred answer, in either `pull` or `push` recommendation mode.

**REQ-DA-018** (Ubiquitous) — The decision-gate axis and the recommendation-mode axis shall be
orthogonal: `decision_gate` shall not read, write, or condition on
`recommendation_mode`, and vice versa. Under `pull`, the landed withholding convention
(SPEC-JUDGMENT-FIRST-MODE-001) applies unchanged to the kickoff question that presents the
index; under `push`, the existing recommendation rules are unchanged.

### C.6 The zero-flag rule

**REQ-DA-019** (Ubiquitous) — A decision index containing zero rows shall not be treated as
kickoff approval: zero rows means only that this pass found no unresolved decisions, and the
Implementation Kickoff Approval gate remains a separate affirmative act that fires regardless
of the row count.

### C.7 Distribution obligations

**REQ-DA-020** (Ubiquitous) — Every file added or changed under `.claude/` or
`.moai/config/` that has a template counterpart shall have the same change under
`internal/template/templates/<same path>`, and the embedded filesystem shall be regenerated
with `make build`.

**REQ-DA-021** (Ubiquitous) — Text landing under `internal/template/templates/**` shall be
authored neutral from the start: no SPEC IDs, no REQ tokens, no audit citations, no internal
dates, no commit SHAs, no macOS-biased absolute paths, no `CLAUDE.local.md` references, and no
bias toward any one of the 16 supported programming languages.

## §D Constraints

- **CONST-1** — Distributed-template behavior must be behaviorally identical to base commit
  `62fbd6baf` when `decision_gate` is absent or `off`. Enabling the gate is a local config
  act, not a template change.
- **CONST-2** — Landed precedent for the flag shape (default-off, resolution helper, round-trip
  tests): `RecommendationMode` / `ResolvedRecommendationMode()` at
  `internal/config/types.go:1360` and `internal/config/defaults.go` (`RecommendationMode:
  "push"`), with `interview_recommendation_mode_test.go`. The new axis follows the same
  wiring inside the same `InterviewConfig` struct.
- **CONST-3** — `moai update` wipes `.claude/rules/moai`, `.claude/skills/moai*`, and
  `.moai/config` wholesale before redeploying from the embedded template. Every file this
  SPEC touches inside those roots must carry its change into the template mirror, or it is
  destroyed on the next update.
- **CONST-4** — `plan-auditor`'s report contract is out of scope (t401 §F / SPEC-JFM §A.3,
  carried forward by REQ-DA-014).
- **CONST-5** — Naming taxonomy is Epic / SPEC / Milestone / Constitution. No Sprint / Round /
  Wave / cohort.
- **CONST-6** — No time estimates in any artifact. Priority labels and phase ordering only.
- **CONST-7** — No deny or block paths anywhere in this SPEC: the decision gate enriches an
  approval gate, it never blocks a tool call. AskUserQuestion remains the sole question
  channel, orchestrator-only.
- **CONST-8** — C1↔C2 mirror branching is intentional (CLAUDE.local.md §2.0): local
  `manager-spec.md` / `clarity-interview.md` / `interview.yaml` already differ from their
  mirrors. The mirror obligation is the same change in both trees, verified by token
  presence in both — never by byte-identity.

## §E Boundary notes

### E.1 What the decision gate is, and is not

It is an **enumeration-and-routing surface**. It enumerates what no document has decided and
routes each item to the authority that owns it. It references no pipeline stage split, adds no
gate, moves no gate, and changes no verdict vocabulary.

### E.2 The auditor deferral (explicit, like t401's)

Integrating the decision index into plan-auditor — a judgment-point finding class, an
auditor-authored index section — would be a report-contract change. SPEC-JFM §A.3 measured
that on the two most common plan-audit paths the report never reaches a user, so editing the
report buys text change without observable effect. That measurement binds here unchanged. The
deferral is recorded so a future card inherits a stated boundary: **the auditor never sees or
produces the index in v1** (REQ-DA-014).

### E.3 Operator-facing decisions (named for kickoff review)

Three decisions in this SPEC are genuinely the operator's, not the author's. They are surfaced
here with recommendations and MUST be presented at the Implementation Kickoff Approval gate:

| # | Decision | Recommendation | Why it is the operator's |
|---|---|---|---|
| OD-1 (D6) | Distributed default for `decision_gate` | **`off`** — the proposal itself scopes the guardrail as optionally applicable ("선택적으로 적용할 수 있는"), and a first release that changes every user's plan phase contradicts the regression property the axis exists to protect. This repo dogfoods `on`. | Distributed-default posture is a product decision |
| OD-2 (D7) | Owner of the product-level reconcile act | **Orchestrator offers reconciliation as a post-kickoff step; the edit routes to manager-docs** (which owns `.moai/project/**` scaffolding per the canonical agent catalog) or is taken by the operator by hand — never manager-spec. | `.moai/project/**` ownership crosses the agent catalog boundary |
| OD-3 (D5) | Auditor integration | **Defer** (REQ-DA-014) — v1 lands the flow without touching the auditor report contract. | Report-contract change has no observable effect on the common paths (SPEC-JFM §A.3) |

### E.4 Artifact language — English, deliberately

Same exception as SPEC-JFM §E.4: these artifacts quote and diff against English doctrine
(`spec-assembly.md`, `askuser-protocol.md`, `manager-spec.md`) that lands under
`internal/template/templates/**`. Korean prose would put every quoted clause through a
translation round-trip on the exact strings the acceptance criteria `grep` for. Operator
reversible: translating the prose is safe provided identifiers, paths, commands, flags, and
quoted doctrine strings stay verbatim.

### E.5 Composability with the sibling SPEC

This SPEC **depends on** SPEC-JUDGMENT-FIRST-MODE-001 (frontmatter `depends_on`; status
`completed` — the Depends_on pre-flight fulfillment definition is satisfied). Composition
points: (1) REQ-DA-017 inherits the pull convention for the kickoff question that presents the
index; (2) the `spec-assembly.md:212` clause that SPEC-JFM 0.2.2 conditioned ("the `(권장)`
first-option label withheld under `recommendation_mode: pull`") is the same clause this SPEC
enriches — the decision-index presentation rides inside that already-mode-conditioned gate and
touches no part of it.

## §F Exclusions

Everything below was considered and is deliberately not built. Each entry names why, so the
submitter and future cards inherit a stated boundary rather than a silence. Each carries
forward or re-justifies a SPEC-JFM §F exclusion per the dispatch (D9).

### Out of Scope — proposal structure elements

- The Stage 1 Mechanical Verification / Stage 2 Adversarial Interrogation split. It is a
  pipeline-structure change; MoAI's separation of concerns already lives in the flow split
  between mechanical lint (spec-lint) and human judgment (the kickoff gate). D1 folds the
  proposal's gate into the gate that exists. (Carried from SPEC-JFM §F.)
- A standalone Human Decision Gate as a new workflow phase. D1: enrichment of the existing
  Implementation Kickoff Approval gate only.
- A first-class "judgment point" section, or a decision-index section, in the plan-auditor
  report template. Report-contract change; §A.3 of the sibling SPEC measured no observable
  effect on the common paths. The auditor is unchanged in v1 (REQ-DA-014, §E.2). (Carried
  from SPEC-JFM §F.)

### Out of Scope — vocabulary and propagation

- Any change to the `PASS | FAIL` verdict vocabulary, to `UNVERIFIED` handling, or to
  `PASS-WITH-DEBT`. The index's row labels (`DECIDED`…) and verdict actions (`DECIDE`…) are a
  decision vocabulary, not an audit verdict vocabulary; conflating the two would put
  decision authority in an audit artifact. (Carried from SPEC-JFM §F.)
- A PRD template, or any new product-requirement surface. `product.md` reconcile only
  (REQ-DA-016); this repo's authority register is the §A.3 table, not a new document class.
- Adopting the issue's attached design verbatim, including its legacy EARS artifact layout
  (`docs/review/{SPEC-ID}`, `answer.md`, `analysis/{Qn}.md`), its `spec-interrogator` role,
  and its three-stage orchestration. The submitter's compatibility note excludes verbatim
  adoption; GEARS is the current notation. (Carried from SPEC-JFM §F.)

### Out of Scope — behavior changes to the existing channel

- Any deny or block behavior on `AskUserQuestion`, or any new deny path in the plan phase.
  The decision gate enriches an approval gate; it never blocks a tool call. (Carried from
  SPEC-JFM §F CONST-7.)
- Automatic or heuristic selection of the decision-gate mode (inferring `on` from task
  complexity, SPEC tier, or user proficiency). The mode is an explicit operator setting;
  a system-chosen default in this exact place would recreate the failure class this SPEC
  removes. (Carried from SPEC-JFM §F.)
- Embedding AI analysis or a preferred answer in decision-index rows. Analysis stays pull
  (the operator may request it per row at kickoff); the index itself never recommends
  (REQ-DA-017). (Composes with SPEC-JFM rather than extending it.)

### Out of Scope — v1 implementation boundaries

- The `go` runtime changes beyond `internal/config` (the `InterviewConfig` field, default,
  and resolver). The gate presentation is doctrine text executed by the orchestrator and
  manager-spec, not a hook or a Go pipeline stage; no observer is authored (contrast
  SPEC-JFM's PreToolUse observer — the index is authored by an agent, so a
  file-existence/static check verifies it; a runtime observer would gate nothing).
- Manager-docs-side reconciliation tooling for `product.md`. The reconcile act is defined
  (REQ-DA-016) and its ownership decision is surfaced (§E.3 OD-2); automating it is a
  follow-up once the operator settles the owner.
- `sync-auditor` changes of any kind. The decision index is a plan-phase artifact; the sync
  audit scores implementation against acceptance criteria and never sees it.

## §G References

- Issue: `modu-ai/moai-adk#1683` (item #1; item #2 = SPEC-JUDGMENT-FIRST-MODE-001, completed)
- Proposal (verbatim, untracked evidence): `.moai/reports/t692/issue-1683-proposal.md`
- Sibling SPEC: `.moai/specs/SPEC-JUDGMENT-FIRST-MODE-001/spec.md` (§A.1 split, §F exclusions
  carried forward, §A.3 auditor measurement)
- Lane decisions D1-D10: card t692 dispatch prompt (2026-09-13)
- Base commit: `62fbd6baf` (= local develop head at authoring), worktree
  `.claude/worktrees/t692`, branch `WT-judgment-authority`
