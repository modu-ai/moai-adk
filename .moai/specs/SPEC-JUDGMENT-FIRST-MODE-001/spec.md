---
id: SPEC-JUDGMENT-FIRST-MODE-001
title: "Judgment-first mode: withhold the recommendation until it is asked for (card t401, issue #1683 item 2)"
version: "0.2.4"
status: completed
created: 2026-09-02
updated: 2026-09-08
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: ".claude/rules/moai/core, .claude/output-styles/moai, internal/config, internal/hook"
lifecycle: spec-anchored
tags: "judgment-first, anchoring, askuser-protocol, recommendation-mode, issue-1683, t401"
tier: L
---

# SPEC-JUDGMENT-FIRST-MODE-001 — Analysis is Pull, Not Push

## HISTORY

| Version | Date | Author | Description |
|---|---|---|---|
| 0.1.0 | 2026-09-02 | manager-spec | Initial Tier L authoring: recommendation-mode axis, six surfaces, runtime observer, vacuity falsifier. |
| 0.2.0 | 2026-09-02 | manager-spec | Iteration-1 plan-audit revision (D1-D12). Verification layer rebuilt as measured two-cell RED-now/green-path adoption; `pull` label-withholding scope widened from decision-type-only to every `AskUserQuestion` (D3 — the runtime cannot distinguish the class); positive-detection requirement REQ-JFM-024 added (D4); S4/S5 coordinates corrected (D9); REQ-JFM-007/008 given concrete shapes; `.github/` template premise corrected (D8). |
| 0.2.1 | 2026-09-02 | manager-spec | Iteration-2 scoped fix pass (D-N1, D-N2, D-N4 blocking; D-N3, D-N5, D-N6 optional). AC-JFM-013's candidate selector made case-insensitive so S1's first-named coordinate `askuser-protocol.md:64` (capital `F` in `**First option label**`) enters the swept set; baseline re-measured 23 → 25 in this tree and `:64` added to the required-`conditioned` row list. design.md's S4/S5 per-surface cells brought onto the 0.2.0 concrete forms. Five mis-cited constraint ids corrected (CONST-3 → REQ-JFM-015 / §B.2; CONST-2 → REQ-JFM-022). Abridged `go test` stdout cells annotated with their line counts; the D3 widening flagged for the kickoff gate in the SPEC body; AC-JFM-021 given an explicit swept-count floor. |
| 0.2.2 | 2026-09-02 | manager-spec | Iteration-3 narrow fix pass (D-N7 blocking; D-N8 optional), operator-authorized beyond the Tier L iteration ceiling to close one finding. The written rule at `acceptance.md` that let a swept candidate be classed `unconditioned-by-design` for absence from §B.1's coordinate table is **removed**: REQ-JFM-016 states a reachability test, not a membership test, and the SPEC's own two admitted precedents (`run.md:137`, `branch-origin-protocol.md:25` — both outside S1-S6) already use the consequence test. `plan/spec-assembly.md:212` (the Implementation Kickoff Approval `(권장)`-first clause — the same clause as `run.md:137`, one file over) and `plan/spec-assembly.md:353` (the sole implementing site of Frozen `CONST-V3R5-035`, whose doctrine site `branch-origin-protocol.md:25` M1 already conditions) added to the required-`conditioned` list with inline reasons; M1 given a `spec-assembly.md` edit-surface row plus its template mirror (verified to exist and be byte-identical to the live file); §E.1 rewritten as a four-coordinate contradiction inventory. Any residual `unconditioned-by-design` judgment is now a blocker report to the orchestrator, not a run-phase classification. AC-JFM-013's RED-now cell re-measured at `ad272be20` with two added coverage controls (`:212`, `:353` — both present in the swept set). AC-JFM-007 given a four-element ordered `Verify` so the S5 sequence is pinned by a criterion (D-N8). |
| 0.2.3 | 2026-09-03 | manager-spec | Provenance amendment per the decision document `.moai/reports/t401/provenance-eligibility-options.md` (Option A + Option C adopted, Option B rejected — the doc's 안 A + 안 C, 안 B 기각): REQ-JFM-025's provenance enumeration extended with `calls_issued` (the asking session's own count of `AskUserQuestion` calls issued during the interval), with AC-JFM-018/023 asserting a four-way `rows_recorded` vs `calls_issued` contrast (observer non-wiring and partial row loss become observable mismatches, never silent passes). `session_start` and a matcher SHA deliberately NOT added — self-proving condition, confirmation stamps gating nothing. REQ count unchanged at 25 (rides REQ-JFM-025's existing enumeration; no new REQ). Affected RED-now cells re-measured and re-pinned. |
| 0.2.4 | 2026-09-07 | manager-spec | Run-phase-escalated repair of two SPEC-body defects (operator-approved, mid-run per D-NEW-1). **(1) AC-JFM-013's sweep window was measuring a different unit than REQ-JFM-016 requires.** The 0.2.2 window — the matched physical line — rested on the claim that "every candidate this tree actually contains is a single-line paragraph or list item (measured)"; the run phase falsified it. `plan/spec-assembly.md:212` is a continuation line of a wrapped `[HARD]` paragraph spanning `:208`-`:214`, so the criterion's only pass route was rewriting that paragraph as one long line to make the token land inside the instrument, and `branch-origin-protocol.md:25` is Frozen, so a window-local condition there is permanently impossible. The window is redefined as **the enclosing markdown block** with a mechanical boundary rule (blank line / heading / code fence / list marker / table row), and the **carrier form** is stated: a row is `conditioned` on a window-local mode reference OR — where this SPEC forbids editing the clause — on a ledger row naming the coordinate the conditioning actually lives at. The window only widens (line ⊆ block), so the obligation is not weakened and the 8 already-`conditioned` ledger rows survive; the 18 `unconditioned-by-design` rows carry a named **run-phase re-sweep task**. `spec-assembly.md:212`'s long line is deliberately left as authored — it is the evidence for this repair. plan.md §F M1's false "measured" claim replaced with the falsification. **Two-class contract kept, third class rejected**: row 4 (`zone-registry.md:869`) was excluded on an explicit-SPEC-exclusion ground the contract does not describe; under the carrier form it is simply `conditioned`, the same shape and carrier chain as `branch-origin-protocol.md:25`, so the differing classes were an artifact of the line window rather than a real distinction. It joins the required-`conditioned` list (six coordinates, was five); no `zone-registry.md` edit is implied, so AC-JFM-012's empty-diff invariant is untouched. **(2) Stale coordinates.** M1's insert renumbered `askuser-protocol.md` below `:102`; §B.1's S1 row `:217` / `:245` / `:253` re-located **by content grep, not arithmetic** to `:265` / `:293` / `:301`. All other SPEC-body coordinates re-verified by content at `82edb9109` and found accurate (S2 `moai.md:451`/`:469`; S3 `:521`/`:537`/`:555`/`:575`; S4 `:353-357`; S5 `:592-597`; S6 `context-window-management.md:84`; `run.md:137`; `spec-assembly.md:212`/`:353`; `zone-registry.md:864`/`:869`; `branch-origin-protocol.md:25`/`:26`). §B.1 marked as an as-of-SHA locating aid rather than an assertion, and AC-JFM-013's required coordinates given an **anchor-text column** so the assertion layer is content-anchored per `verification-completeness.md` §4. |

## §A Context and Problem

### A.1 Origin

GitHub issue `modu-ai/moai-adk#1683` ("Human Decision Authority Guardrail for SPEC Review",
external submitter, 2026-08-30, `type:feature`) proposes a Human Decision Authority model. The
submitter's own compatibility note states the attached design targets pre-v3.1.2 legacy EARS and
is **not** proposed for verbatim adoption; what is proposed is the authority model, and the
submitter asked for maintainer direction before implementation.

The proposal was split, and the adoption order was fixed as [HARD]:

- **Item #2 — "Analysis is Pull, Not Push"** (anchoring removal). Adopted **first and narrowly**.
  This SPEC.
- **Item #1 — a decision gate inside SPEC Review, with decision-index authority routing.** A
  separate card. **Not** this SPEC (see §F).

Three conditions from the issue are adopted verbatim as design constraints and bind every
requirement below:

1. **Detect → Explain → Ask, but never decide.**
2. **An LLM "best practice" is not a policy.**
3. **When uncertain, escalate. Never downgrade.**

### A.2 The measured problem

A read-only survey of `.claude/rules/moai/**`, `.claude/output-styles/moai/*.md`,
`.claude/agents/moai/*.md`, `.claude/skills/moai/**`, `.moai/config/sections/*.yaml`, and
`internal/config/*.go` collected 77 recommendation-emission coordinates, classified A (emission
mandated) / B (permitted) / C (existing anti-anchoring guard). Evidence:
`.moai/reports/t401/lens-recommendation-surfaces.md`.

Two negative findings are load-bearing:

- **No rule anywhere requires the user's independent judgment BEFORE an AI recommendation.** Zero
  hits across the surveyed tree. The four nearest clauses (Report-Before-Ask,
  Requested-Deliverable Primacy, Adaptive strength, Implementation Kickoff Approval) each put
  *evidence* before the recommendation, never *user judgment*. Kickoff Approval is an approval
  gate over options that **already carry the label**.
- **No config key can turn recommendation emission off.** All 33 config sections plus every
  `internal/config/*.go` file were read. The nearest keys control round counts, auto-apply, or
  audit blocking — none control recommendation emission, and most emission clauses are `[HARD]`.

The consequence is anchoring: the user's first act on a decision is choosing among options one of
which is already marked recommended.

### A.3 Correction to the card's premise

The card text asserts that plan-auditor and the orchestrator "always emit observation and
recommendation together". Measurement (`.moai/reports/t401/lens-audit-contract.md`) shows this is
**true on the orchestrator axis and only partly true on the plan-auditor axis**:

| plan-audit path | Does the auditor report reach the user? |
|---|---|
| iter-1 PASS (happy path) | **No** — one log line only; the report body is consumed agent-to-agent |
| iter-1~2 FAIL | **No** — the report routes straight back to manager-spec |
| 3× FAIL escalation | Yes — full review-1~3 text plus a 3-option question |
| run-gate FAIL / INCONCLUSIVE | Yes — Phase 1 blocked, report surfaced |

The auditor's recommendations reach the user on exactly three paths (Kickoff-turn HTML, 3×FAIL
escalation, run-gate block). On the two most common paths they never reach a human at all.

**This correction is what scopes the SPEC.** Rewriting the auditor's output contract would change
text that, on the common paths, no user ever reads — cost without an observable effect. The pull
convention therefore targets **orchestrator user-facing output**, not the auditor report contract.
The auditor's `Required fix:` field and `## Recommendation` section are left exactly as they are.

## §B Solution Shape

Introduce a **recommendation-mode axis** with two values:

```yaml
interview:
  recommendation_mode: push   # push (default, current behavior) | pull (judgment-first)
```

The axis **generalizes the existing Adaptive-strength clause**
(`.claude/rules/moai/core/askuser-protocol.md` § Recommendation Placement Principles, principle 5
— today the only clause in the codebase able to suppress the label, gated on estimated
proficiency). It does **not** author a parallel mechanism.

Under `pull`, the orchestrator withholds **both** the `(권장)` / `(Recommended)` first-option label
— on **every** `AskUserQuestion` call, not only on decision-type ones — **and** the banner
recommendation fields, emitting them only on an explicit user request for analysis or for a
recommendation.

**Why the label rule is not scoped to decision-type questions.** The narrower scope was the
0.1.0 wording and it was withdrawn on measurement. "Decision-type question whose options derive
from investigation results" is a doctrine concept with no runtime carrier: the `AskUserQuestion`
payload carries question text and options and no type tag, and whether a report preceded the call
in the same turn is invisible to a `PreToolUse` hook (design.md §6.3; plan.md §B item 5). A rule
scoped to a class the runtime cannot distinguish cannot be measured, and the criterion asserting it
would have to filter on a field that does not exist — returning an empty sample forever. The label
is a recommendation whatever the question's class, so the obligation widens to match what is
observable rather than the measurement narrowing to match an unobservable obligation. This
widening is a real behavioral scope change relative to 0.1.0, and it is **flagged for operator
review at the Implementation Kickoff Approval gate** — not assumed settled (progress.md §Open
items).

Under `push` — the distributed default, and the behavior when the key is absent — nothing changes.

### B.1 The six surfaces in scope

| # | Surface | Coordinates |
|---|---|---|
| S1 | AskUserQuestion first-option `(권장)` / `(Recommended)` label | `askuser-protocol.md:64`, `:265`, `:293`, `:301` |
| S2 | Discovery banner `⏭️ Recommended action:` | `output-styles/moai/moai.md:451`, `:469` |
| S3 | Epic Stats / Epic Status `⏭️ Next:` | `moai.md:521`, `:537`, `:555`, `:575` |
| S4 | Insight banner (What / Why / Alternatives / Implications emitted together, no user-judgment slot) | `moai.md:353-357` |
| S5 | Error Recovery option ordering (`A. Retry as-is` first) | `moai.md:592-597` |
| S6 | The `/clear` natural-language recommendation | `context-window-management.md:84` |

**The line numbers in this table are a locating aid measured at `82edb9109`, not an assertion.**
They decay on the next insert above them: M1's edit to `askuser-protocol.md` renumbered everything
below `:102`, moving S1's `:217` / `:245` / `:253` to `:265` / `:293` / `:301`. Each was re-located
by grepping its content, not by adding the offset. Where a coordinate here disagrees with the file,
the file wins — grep the clause and use the line it is actually on. The coordinates AC-JFM-013
*asserts* (its required-`conditioned` list) carry an anchor-text column for exactly this reason;
this table is enumeration, and per REQ-JFM-016 it is never the scope test.

S6 lives **outside** the AskUserQuestion channel, so none of that channel's neutrality guards
(neutral-description rule, Recommendation Placement Principles, Report-Before-Ask) currently reach
it. It is the one surface where the mode has to be stated locally rather than inherited.

### B.2 Frozen-zone handling

`CONST-V3R5-035` in `.claude/rules/moai/core/zone-registry.md` (`zone: Frozen`,
`canary_gate: true`) mandates the `(권장)`-first structure, anchored at
`.claude/rules/moai/development/branch-origin-protocol.md#hard-rules`.

The decision is to make that clause **conditional on the mode, not to rewrite it**. The Frozen
clause keeps its current text verbatim as the **push-mode branch**; a pull-mode branch is added
beside it. The registry entry's `clause:` string is therefore unchanged, which keeps the canary
gate satisfiable.

### B.3 Regression evidence — both axes required

The card warns explicitly that a convention nobody follows is vacuous. Both axes are required:

- **Runtime axis** — a `PreToolUse` observer on `AskUserQuestion` recording, per call, whether a
  `(권장)` / `(Recommended)` label was present, the active mode, and the question type, appended as
  JSONL under `.moai/logs/`. This is an **observer, never a deny**. Feasibility is pre-verified:
  `.claude/settings.json` already registers three `PreToolUse` matcher blocks all routing to
  `handle-pre-tool.sh`, and `internal/hook/pre_tool.go` already carries sibling observation
  branches for `Agent|Task` (agent-model) and `SendMessage` (stop-guard) that parse
  `input.ToolInput`. An `AskUserQuestion` branch is the same shape.
- **Static axis** — a CI grep guard asserting the doctrine text and the §8 banner templates stay
  consistent with the pull convention, shaped like the existing
  `.github/workflows/template-neutrality-check.yaml`.

## §C Requirements (GEARS)

### C.1 The mode axis

**REQ-JFM-001** (Ubiquitous) — The recommendation-mode axis shall carry exactly two values,
`push` and `pull`, exposed as `interview.recommendation_mode` in
`.moai/config/sections/interview.yaml`.

**REQ-JFM-002** (Where — capability gate) — Where `interview.recommendation_mode` is absent,
empty, or `push`, the orchestrator shall emit recommendation signals on all six surfaces exactly
as it does at base commit `ad272be20`, with no observable behavioral difference.

**REQ-JFM-003** (Where — capability gate) — Where `interview.recommendation_mode` holds a value
that is neither `push` nor `pull`, the config loader shall resolve the mode to `push` and shall
record the unrecognized value, rather than failing the load.

**REQ-JFM-004** (Ubiquitous) — The distributed template shall ship the axis default as `push`, and
the template mirror of `interview.yaml` shall be behaviorally identical to its state at base
commit `ad272be20` for every consumer that does not set the key.

### C.2 Pull-mode output convention

**REQ-JFM-005** (While — state-driven) — While the mode is `pull`, when the orchestrator composes
any `AskUserQuestion` call, the orchestrator shall omit the `(권장)` / `(Recommended)` suffix from
every option label (S1). The obligation is deliberately unscoped by question class: the runtime
carries no question-type field, so a class-scoped obligation would be unmeasurable (§B).

**REQ-JFM-006** (While) — While the mode is `pull`, the orchestrator shall withhold the Discovery
banner `⏭️ Recommended action:` field (S2) and the Epic Stats / Epic Status `⏭️ Next:` field (S3),
rendering the remaining banner fields unchanged.

**REQ-JFM-007** (While) — While the mode is `pull`, when the orchestrator renders the Insight
banner, it shall emit `What` / `Why` / `Alternatives` / `Implications` **without** collapsing them
into a directive, and shall carry an explicit user-judgment slot in place of an implied
conclusion (S4). The slot's shape is fixed: one additional final field, literally
`Your call: [what the reader decides]`, placed after `Implications:` and inside the same banner
frame, carrying no orchestrator preference.

**REQ-JFM-008** (While) — While the mode is `pull`, when the orchestrator renders Error Recovery
options, it shall order the options by **increasing cost of the action to the user** — least
destructive first, most destructive last — rather than by expected desirability, and shall not
place `Retry as-is` first by convention (S5). The rule is stated as an applicable ordering, not as
"an order that does not signal a preference", so that two readers produce the same order twice:
`Pause` (no state change) → `Retry as-is` → `Alt approach` → `Abort+preserve`.

**REQ-JFM-009** (While) — While the mode is `pull`, when accumulated context crosses the
model-specific handoff threshold, the orchestrator shall state the measured threshold crossing and
the available actions **without** phrasing a `/clear` directive as a recommendation (S6).

**REQ-JFM-010** (When — event-driven) — When the user explicitly requests an analysis, a
preference, or a recommendation, the orchestrator shall emit the withheld recommendation on the
requested surface, in the same form it would carry under `push`.

### C.3 The three adopted conditions

**REQ-JFM-011** (Ubiquitous) — The pull-mode convention shall follow **Detect → Explain → Ask,
never decide**: withholding a recommendation shall never withhold the underlying observation, the
evidence, or the enumerated options.

**REQ-JFM-012** (Ubiquitous) — The convention shall treat an LLM-derived "best practice" as
**not** a policy: pull-mode text shall not present a model-inferred default as an established
project rule.

**REQ-JFM-013** (When) — When the orchestrator is uncertain which option holds under `pull`, it
shall escalate the decision to the user rather than downgrading it to a silently-applied default.

**REQ-JFM-014** (Ubiquitous) — Pull mode shall not weaken any existing evidence obligation. The
Report-Before-Ask gate, Requested-Deliverable Primacy, the neutral-description rule, and the
Implementation Kickoff Approval human gate shall remain binding and unmodified in both modes.

### C.4 Frozen-zone and doctrine consistency

**REQ-JFM-015** (Where) — Where the mode is `push`, the `CONST-V3R5-035` clause text shall apply
verbatim as written at base commit `ad272be20`; the pull-mode branch shall be added beside it
without altering the registry `clause:` string.

**REQ-JFM-016** (Ubiquitous) — No doctrine file shall carry a `[HARD]` clause requiring a
`(Recommended)`-first option that is unconditioned on the mode. Any such clause reachable from the
six surfaces shall carry a mode reference. (The test is **reachability**, not membership in §B.1's
coordinate table — that table enumerates S1's known instances, and the AC-JFM-013 sweep exists to
find the ones it missed. Four such clauses are currently identified: `run.md:137` and
`plan/spec-assembly.md:212`, both mandating `(Recommended)` / `(권장)`-first at the Implementation
Kickoff Approval gate; `plan/spec-assembly.md:353`, the sole implementing site of the Frozen
`CONST-V3R5-035` BODP-gate clause; and that clause's doctrine site
`branch-origin-protocol.md:25`. All four are downstream consumers of S1, not further surfaces. See
§E.1.)

### C.5 Regression evidence

**REQ-JFM-017** (When) — When an `AskUserQuestion` tool call is issued, the `PreToolUse` handler
shall append one JSONL row under `.moai/logs/` recording at minimum: the timestamp, the session
id, the resolved recommendation mode, whether a `(권장)` / `(Recommended)` label was present on any
option label, and the observed question type.

**REQ-JFM-018** (Ubiquitous) — The `AskUserQuestion` observer shall never deny, block, or delay a
call. It shall fail open on every uncertainty — an unparseable payload, an absent config, an
unwritable log path — and shall never displace an established permission decision.

**REQ-JFM-019** (When) — When a pull request or push touches the doctrine files or the §8 banner
templates carrying the pull convention, the static CI guard shall assert that the doctrine text
and the banner templates remain mutually consistent, and shall fail the check on divergence.

**REQ-JFM-020** (Ubiquitous) — The acceptance criteria shall include at least one
release-blocking criterion that can FAIL while the convention text is present — that is, a
criterion measuring adherence at runtime rather than the existence of the rule.

### C.6 Distribution obligations

**REQ-JFM-021** (Ubiquitous) — Every file added or changed under `.claude/`, `.moai/config/`, or
`.github/` that has a template counterpart shall have a matching change under
`internal/template/templates/<same path>`, and the embedded filesystem shall be regenerated with
`make build`.

**REQ-JFM-022** (Ubiquitous) — Text landing under `internal/template/templates/**` shall be
authored neutral from the start: no SPEC IDs, no REQ tokens, no audit citations, no internal
dates, no commit SHAs, no macOS-biased absolute paths, no `CLAUDE.local.md` references, and no
bias toward any one of the 16 supported programming languages.

**REQ-JFM-023** (When) — When a hook wrapper under `.claude/hooks/moai/` is changed, the change
shall land in its `.sh.tmpl` sibling in the same milestone, so `moai update` cannot silently
revert it.

### C.7 Falsifier integrity

**REQ-JFM-024** (Ubiquitous) — The observer's label detector shall be demonstrated **firing**
before the vacuity falsifier is read, on two independent surfaces: a unit assertion that a payload
carrying a `(권장)` / `(Recommended)` option label yields `label_present: true` and one without it
yields `false`; and a live pre-landing window, recorded while this repository is still at
`recommendation_mode: push` and the doctrine is unamended, containing at least one row with
`label_present: true`. A falsifier read from a detector never observed firing asserts nothing —
an observer whose detector is inert satisfies "zero violations" on every row while the convention
is entirely unfollowed.

**REQ-JFM-025** (Where — capability gate) — Where the recorded-window evidence is collected by a
session other than the one implementing the SPEC, the exported artifact shall carry a provenance
record naming the source absolute path, the asking session's `session_id`, the collection interval,
the row count and the `label_present: true` count measured at export time, the asking session's
own count of `AskUserQuestion` calls issued during the interval (`calls_issued` — a value the
asking session knows without the observer), and the export command; and the acceptance criteria
shall assert both the artifact and its provenance, including the contrast between
`calls_issued` and the recorded row count. The window is
collected by the session that actually asks and whose observer is actually wired — under the kanban
division of labour the lead session, which owns the operator channel, and never the card's lane,
which issues no `AskUserQuestion` calls at all; and a session already running when the
`AskUserQuestion` matcher was added does not pick the matcher up, so the collecting session is one
started after the matcher landed — and its rows land under that session's `CLAUDE_PROJECT_DIR`
(its cwd when unset), which is generally not the card worktree. An exported copy whose origin is unrecorded is an unattributed
claim, and a criterion read from it asserts nothing. A `session_start` timestamp and a matcher
SHA are deliberately **not** part of this record: the exported window's own existence and row
counts already prove the wired-session condition, so those fields would be confirmation stamps
gating nothing — and fields that gate nothing leave the impression verification finished when it
did not (decision: `.moai/reports/t401/provenance-eligibility-options.md`, Option B rejection —
the doc adopts Option A + Option C).
The residual sample-bias hole — an unwired session leaves no rows at all, which `session_start`
cannot reach in principle — is handled by `calls_issued`, which makes observer non-wiring and
partial row loss observable rather than silently reading as "no violations"; it does not
eliminate them.

## §D Constraints

- **CONST-1** — Distributed-template behavior must be byte-identical to base commit `ad272be20`
  when `recommendation_mode` is absent or `push`. Dogfooding this repository at `pull` is a local
  config act, not a template change.
- **CONST-2** — Landed precedent for the flag shape (default-off, local dogfood on):
  `Workflow.BranchGuard.Enabled`, `workflow.integration_lock.enabled`,
  `workflow.agent_stop_guard.enabled`. The new key follows the same wiring, but lives under
  `interview` rather than `workflow`, because it configures interview output rather than a guard.
- **CONST-3** — `moai update` wipes `.claude/rules/moai`, `.claude/skills/moai*`, and
  `.moai/config` wholesale before redeploying from the embedded template. Every file this SPEC
  touches inside those roots must have a template mirror, or it is destroyed on the next update.
- **CONST-4** — `plan-auditor`'s report contract, the `Required fix:` defect field, and the
  `## Recommendation` section are out of scope and must not be edited (§A.3).
- **CONST-5** — Naming taxonomy is Epic / SPEC / Milestone / Constitution. Sprint / Round / Wave /
  cohort are retired.
- **CONST-6** — No time estimates in any artifact. Priority labels and phase ordering only.
- **CONST-7** — The observer is a PreToolUse observation branch only. No deny path is authored in
  this SPEC, in either mode.

## §E Boundary notes

### E.1 Four inherited coordinates — consumers of S1, not further surfaces

REQ-JFM-016's test is **reachability from the six surfaces**, not membership in §B.1's coordinate
table. The table enumerates S1's known instances; the AC-JFM-013 sweep exists precisely to find
instances the table missed, so using the table as the scope test would be circular. Four
coordinates are admitted on the consequence test — each carries a `[HARD]` `(Recommended)` /
`(권장)`-first mandate that, left unconditioned while S1 says "withhold under pull", would place
two `[HARD]` clauses in direct contradiction.

| Coordinate | Why it is a contradiction if left unconditioned |
|---|---|
| `.claude/skills/moai/workflows/run.md:137` | Requires the Implementation Kickoff Approval question's first option to be marked `(Recommended)`. A downstream consumer of S1. |
| `.claude/skills/moai/workflows/plan/spec-assembly.md:212` | The **same** Kickoff-gate clause as `run.md:137`, one file over: `[HARD]` … "does NOT relax its three canonical options … or the `(권장)` first-option label". Conditioning `run.md:137` alone leaves the contradiction standing verbatim at the same gate. |
| `.claude/skills/moai/workflows/plan/spec-assembly.md:353` | The `(권장)`-first line of the Phase 13 BODP Gate (`spec-assembly.md:330-356`) — the **sole implementing site** of Frozen `CONST-V3R5-035`, whose clause subject is literally "Skill body BODP gate" (`zone-registry.md:864-870`, `canary_gate: true`). |
| `.claude/rules/moai/development/branch-origin-protocol.md:25` | The **doctrine site** of that same Frozen clause (§B.2). |

The last two are one clause seen from two ends. Conditioning the doctrine site while leaving its
only implementation untouched would leave the change **half-applied**: after M1,
`branch-origin-protocol.md` would say the BODP gate withholds the label under `pull` while
`spec-assembly.md:353` says the first option carries `(권장)`, unconditionally — a contradiction at
a `zone: Frozen`, `canary_gate: true` surface, which is the exact shape this section exists to
prevent.

REQ-JFM-016 therefore requires a **one-line mode reference** at each coordinate — not a rewrite of
any gate, and not a change to the Kickoff gate's mandatory, score-independent nature. Milestone M1
carries all four (plus their template mirrors, per REQ-JFM-021), and flags them for operator review
as the coordinates outside the nominated six.

### E.2 What pull mode is, and is not

Pull mode is an **output convention**. It withholds a recommendation until asked. It references no
authority source, changes no pipeline structure, adds no stage, and moves no gate.

### E.3 A mid-session mode flip takes effect at the next composition

The mode is resolved at output-composition time, per surface (design.md §2) — it is not latched at
session start. A flip therefore takes effect on the next composed output and never rewrites, recalls,
or re-renders a question or banner already emitted. An `AskUserQuestion` round already on screen when
the config changes keeps the labels it was composed with; the next round carries the new mode. No
migration, drain, or in-flight handling is required, and none is authored.

### E.4 Artifact language — English, deliberately

`conversation_language` is `ko`, and the repository's recent SPEC bodies are Korean. These five
artifacts are authored in **English by deliberate exception**, recorded here rather than left
unstated. The reason is Template-First: the doctrine text these artifacts quote, restate, and diff
against — `askuser-protocol.md`, `moai.md` §8, `zone-registry.md`, `run.md` — is English and lands
under `internal/template/templates/**`, where REQ-JFM-022 requires neutral English. Authoring the SPEC
in Korean would put every quoted clause through a translation round-trip on the exact strings the
acceptance criteria `grep` for, which is where a drift would be least visible. Operator-reversible:
translating the prose is safe provided identifiers, paths, commands, flags, and quoted doctrine
strings stay verbatim.

## §F Exclusions

Everything below was considered and is deliberately not built. Each entry names why, so the item
#1 card and the issue reporter both inherit a stated boundary rather than a silence.

### Out of Scope — proposal item #1 (decision gate)

- The Stage 1 Mechanical Verification / Stage 2 Adversarial Interrogation split. It is a
  pipeline-structure change; the lead fixed item #2 as the first and narrow experiment.
- The decision index and its authority routing labels (`DECIDED`, `POLICY-COVERED`,
  `EVIDENCE-NEEDED`, `FOUNDER`). Item #1 scope.
- A Human Decision Gate inside SPEC Review. Item #1 scope.
- A first-class "judgment point" section in the plan-auditor report template. Measurement confirms
  no such construct exists today and every finding folds into a defect with a `Required fix:`;
  introducing one is a report-contract change, which §A.3 and CONST-4 exclude.

### Out of Scope — authority-source definition (carry-forward)

- Defining what `DECIDED` and `POLICY-COVERED` would reference, and whether that authority lives
  in the `product.md` / `structure.md` / `tech.md` family or in a new layer. **Reason:** those two
  labels belong to the proposal's decision-index authority routing, which is item #1. Pull mode
  does not reference an authority source at all — it is an output convention that withholds a
  recommendation until asked. Recorded here as an explicit deferral so the item #1 card inherits
  it cleanly rather than rediscovering it.

### Out of Scope — auditor report contract

- `plan-auditor`'s `## Recommendation` section, its `Required fix:` defect field, and
  `sync-auditor`'s `Required fix:` / `### Recommendations` sections. **Reason:** §A.3 — on the two
  most common plan-audit paths the report never reaches a user, so editing it buys text change
  without observable effect. The pull convention targets orchestrator user-facing output.
- `super-advisor`'s non-binding `Recommendation` output contract. It is explicitly non-binding and
  is consulted on request, which is already the pull shape.

### Out of Scope — verdict-vocabulary changes

- The `zero-flag ≠ pass` rule as a new clause. Measurement shows the concern is already
  structurally satisfied: PASS requires all must-pass PASS/N-A **and** score ≥ tier threshold
  **and** evidence citation on every PASS; zero defects satisfies none of those. Adding a clause
  restating an already-binding property is text without effect.
- Any change to the `PASS | FAIL` verdict vocabulary, to `UNVERIFIED` handling, or to
  `PASS-WITH-DEBT` (which is an orchestrator-offered user option, not an auditor token).

### Out of Scope — PRD and legacy-EARS propagation

- Propagating the authority model into a PRD template or any product-requirement surface. Not
  requested by the lead's split and not measurable from the six surfaces.
- Adopting the issue's attached design verbatim. The submitter's own compatibility note states it
  targets pre-v3.1.2 legacy EARS and is not proposed for verbatim adoption.

### Out of Scope — deny paths and mode auto-selection

- Any deny or block behavior on `AskUserQuestion`. CONST-7: the runtime axis is an observer.
- Automatic or heuristic selection of the mode (for example, inferring `pull` from user
  proficiency). The existing Adaptive-strength principle already covers proficiency; the mode is
  an explicit operator setting, and conflating the two would reintroduce a system-chosen default
  in the exact place this SPEC removes one.

## §G References

- Issue: `modu-ai/moai-adk#1683`
- Evidence: `.moai/reports/t401/preflight.md`, `.moai/reports/t401/lens-recommendation-surfaces.md`,
  `.moai/reports/t401/lens-audit-contract.md`
- Base commit: `ad272be20` (= `origin/develop`), worktree `.claude/worktrees/t401`, branch
  `WT-analysis-pull`
