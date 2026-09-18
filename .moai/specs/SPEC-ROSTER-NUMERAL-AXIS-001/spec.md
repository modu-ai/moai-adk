---
id: SPEC-ROSTER-NUMERAL-AXIS-001
title: "Numeral-axis layer for the roster guard — count-only claims the enumeration sweep cannot reach"
version: "0.2.0"
status: draft
created: 2026-09-18
updated: 2026-09-18
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/harness/rosterguard"
lifecycle: spec-anchored
tags: "roster-guard, numeral-axis, count-claim, drift-detection, t930"
tier: M
---

# SPEC-ROSTER-NUMERAL-AXIS-001 — Numeral-axis layer for the roster guard

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-18 | manager-spec | Initial plan-phase draft (card t930, branch `WT-numeral-roster-guard`) |
| 0.2.0 | 2026-09-18 | manager-spec | Plan-audit revision: cost arithmetic corrected to the measured residual 46, decision D3 (mirror derivation) added, neutralise-before-count ordering fixed in REQ-RNA-007, REQ-RNA-012 (in-run re-derivation) and REQ-RNA-013 (layer scope) added, AC-RNA-013…016 added, AC-RNA-005/006/011 made verifiable |

## §A Problem Statement

`internal/harness/rosterguard` (card t922) makes a NEW roster listing impossible to add
silently — but only along one axis. `Sweep()` finds a file by ENUMERATION: a file qualifies
when it carries at least `SweepThreshold` (10) distinct agent names. A site that states only a
SIZE — a number — and names few agents or none at all is structurally invisible to that sweep.

Two such sites are live in this tree and were measured in it:

| Path | Line | Text | Names agents |
|---|---|---|---|
| `.moai/project/product.md` | 151 | `… (11 retained agents x 3 model tiers) …` | one, in passing |
| `.moai/project/tech.md` | 17 | `33-cell profile matrix: 11 retained agents x 3 model tiers …` | **zero** |

Both claim 11 where `template.ProfileMatrixAgents()` carries **13**
(`internal/template/profile_matrix.go:234-248`, read in this tree at base `690dfe369`).

[HARD] **This is not a threshold-tuning problem.** `tech.md` names zero agents, so no
enumeration threshold — not 10, not 1 — reaches it. The axis a site ENUMERATES on and the axis
it CLAIMS on are different populations; a sweep keyed to the first cannot see a claim made only
on the second. `registry.go` already records this conclusion in the block comment above the two
hand-registered rows, and names this card as the owner of the general closure.

The two sites themselves are already registered by hand
(`product-md-profile-matrix-size`, `tech-md-profile-matrix-size`), each carrying
`SweepUnreachable` and `KnownStale{DeclaredCount: 11}`. **Repairing those two rows is therefore
NOT this card's deliverable.** The deliverable is closing the general hole: making an
unregistered count-only roster claim impossible to ADD silently, the way the enumeration sweep
already does for listings.

### Reuse target — the shape already exists

`internal/web/docs_tab_contract_test.go` implements exactly this layer for a different subject
(the settings-tab count). Its head comment states the governing principle:

> The settings tab count and the tab name list are ONE fact, owned by `consoleTabs()`.

The roster-side owner of that one fact is `template.ProfileMatrixAgents()`. That file's layer
carries decisions this SPEC adopts rather than re-derives: a noun class with an ASCII word
boundary (`tabNoun`), a measured adjacency window (`adjacency`, 12 chars — a tidier window is
structurally blind to the English sites), a digit axis and a word axis as separate regexps
(`digitNumeralRe` / `wordNumeralRe`), an ordinal neutraliser that removes a numeral that is not
a count (`ordinalPrefixRe`, 第三 in `第三方`), and a content-identified allowlist with a
mandatory reason (`allowRule` / `docsTabAllowRules`).

### Plan-phase measurement (attribution)

The population figures below were measured by the dispatching lead in THIS tree at base
`690dfe369`, walking 5975 files with `sweepSkipPrefixes` plus `CHANGELOG.md`, anchored on the
NOUN with the NEAREST preceding numeral inside a 12-character window:

| Noun class | Digit axis | Word axis |
|---|---|---|
| `{retained agent(s)}` only | 37 files / 54 hits | — |
| `{retained agent(s), agent catalog, agent roster, retained catalog, N-agent catalog}` | **62 files / 75 hits** | 4 files / 4 hits |
| `{agent(s)}` unrestricted | 267 files / 613 hits | — |

[HARD] These are plan-phase inputs, not a run-phase baseline. Run-phase re-derives every figure
in-run against the tree it is changing (REQ-RNA-012); a carried-over number is not a baseline
(`verification-claim-integrity.md` §2).

### The accounting between the population and the deliverable

The population above is not the deliverable's cost — the RESIDUAL is, and it was measured by the
dispatching lead in this tree at HEAD `6abcc85fa` with its own `registry.go` parse:

| Quantity | Value |
|---|---|
| Hit paths | 63 (62 digit-axis + one word-only, `workflow-specialist.md`) |
| Registry paths carrying a `ClaimCount` row | 20 |
| Hits already discharged by a `ClaimCount` row | 17 |
| **Hits needing a NEW row or an exempt declaration** | **46** |
| Extra paths the rejected any-row rule would free | 1 (`internal/web/agentfm.go`) |

46, not the ~15 the first draft of this SPEC asserted. Decision D3 is the response to that
number; the exact post-derivation figure is a run-phase re-derivation obligation
(REQ-RNA-012), not a number asserted here.

The word-axis expectation belongs here rather than in the requirement text: under the
neutralise-then-count ordering REQ-RNA-007 fixes, the live reported word-axis population is
expected to be **0**, because the single live candidate —
`.claude/agents/harness/workflow-specialist.md:52`, `All four are retained agents.` (read in
this tree) — is a subset predication removed before counting.

Independently re-measured in this worktree while authoring this SPEC:
`go test -count=1 ./internal/harness/rosterguard/...` → `ok … 1.096s`;
`profileMatrixAgentOrder` holds 13 names; `registry.go` carries 30 `Site` rows, 17 of which
mention `ClaimCount`.

## §B Requirements (GEARS)

- **REQ-RNA-001 (the layer)** — **When** a file inside the numeral layer's scope (REQ-RNA-013)
  carries a numeral adjacent to a roster noun, the roster guard shall report that site as an
  undeclared count claim unless the site is discharged under REQ-RNA-004.

- **REQ-RNA-002 (noun class)** — The numeral layer shall anchor on the noun class
  `{retained agent, retained agents, agent catalog, agent roster, retained catalog,
  N-agent catalog}`, with the English forms bounded by ASCII word boundaries so that the noun
  does not match inside a longer word.

- **REQ-RNA-003 (numeral selection)** — **When** more than one numeral precedes a matched noun
  within the adjacency window, the layer shall select the NEAREST preceding numeral. It shall
  not select the leftmost: the leftmost rule mis-reads `CLAUDE.md 4 (13 retained agents)` and
  reports 4.

- **REQ-RNA-004 (discharge rule)** — A numeral hit shall be discharged **only** when its path
  carries a `Registry()` row whose `Claims` has `ClaimCount`, or an explicit numeral-exempt
  declaration for that path. A row carrying `ClaimMembership` alone shall not discharge a
  numeral hit in the same file.

- **REQ-RNA-005 (exempt reason)** — **Where** a numeral-exempt declaration exists, it shall
  carry a non-empty reason. **When** the reason is empty or whitespace-only, the layer shall
  emit a finding reporting the declaration as incomplete and shall not suppress the hit.

- **REQ-RNA-006 (word axis)** — The layer shall implement a spelled-out numeral axis
  (`eleven` / `twelve` / `thirteen` and the locale forms the reuse target already lists)
  alongside the digit axis, and shall carry a positive control proving the regexp fires on
  synthetic input. The axis is a structural hole that shall be correct when it turns on later,
  not a claim that it catches something now.

- **REQ-RNA-007 (selector neutralisation, and the order it runs in)** — The layer shall not
  report a phrase that predicates over a subset rather than stating a size: `one of the 11
  retained agents` (selector) and `four are retained agents` (subset predication) shall produce
  no finding, by the same mechanism `ordinalPrefixRe` uses to neutralise `第三方`.

  The pipeline order shall be **neutralise → match → report → discharge**: neutralisation
  removes selector and ordinal spans from the text BEFORE the axes match, so a neutralised
  phrase is never counted and then discharged. The order is normative because it decides what
  the reported population IS: under it the live word-axis population is 0 by construction, and
  under the reverse order the same string would be counted as 1 and then discharged.

- **REQ-RNA-008 (breadth is printed)** — The layer shall PRINT its hit set, one line per hit,
  anchored at column 0, so the set can be compared against the enumeration rather than merely
  counted.

- **REQ-RNA-009 (anti-vacuity)** — **When** the layer's hit set is empty, the guard shall treat
  that as a measurement failure and fail, not as a clean tree.

- **REQ-RNA-010 (control probe)** — The layer shall carry a control probe that feeds it input
  which MUST produce a violation, and the probe shall be observed to fire.

- **REQ-RNA-011 (staleness stays enumerable)** — **Where** the layer reaches a site whose count
  disagrees with `template.ProfileMatrixAgents()`, the site shall be registered with a
  `CountPattern` and a `KnownStale` marker rather than exempted, so the disagreement is
  enumerable and self-expiring.

- **REQ-RNA-012 (in-run re-derivation)** — **When** run-phase begins, it shall re-derive the
  digit-axis and word-axis populations in-run against the tree being changed, and shall use
  those observed values as its baseline. It shall not carry the plan-phase figures of §A
  forward as a measurement (`verification-claim-integrity.md` §2).

- **REQ-RNA-013 (layer scope)** — The numeral layer shall walk the same tree the enumeration
  sweep walks, inheriting `sweepSkipPrefixes` and `sweepSkipFiles`
  (`internal/harness/rosterguard/check.go`), and shall additionally exclude `.moai/research/`
  — a dated-record tree the sweep does not exclude today and in which a live digit-axis
  candidate exists (`.moai/research/anthropic-best-practices-2026-05-24.md:92`,
  `CLAUDE.md §4 agent catalog name mismatch`, read in this tree). The exclusion set shall be
  stated in one place in code and cited by the layer rather than re-listed.

## §C Constraints

- The single owner of the roster fact is `template.ProfileMatrixAgents()`; the layer asserts
  against it and never against a hand-typed list.
- Existing rosterguard behaviour is a PRESERVE surface: `Sweep()`, `CheckSite()`,
  `TestSweepFindsNoUndeclaredRosterListing`, `TestGuardFiresOnDeliberatelyWrongInput` and the
  `SweepUnreachable` semantics keep working unchanged.
- Go file naming `snake_case.go` / `snake_case_test.go`; error wrapping
  `fmt.Errorf("…: %w", err)`; all code, comments and godoc in English.
- Template-First: any edit under `internal/template/templates/` has its local `.claude/`
  counterpart edited in the same commit, and the reverse. The template tree stays neutral
  across the 16 supported programming languages and carries no SPEC ID, internal date, or
  commit SHA.
- Verification is scoped to `./internal/harness/rosterguard/...` and any other package actually
  touched. The full suite is CI's verdict, not a local run.
- Baseline-first ordering: a baseline artifact lands in its OWN commit preceding the
  implementation commit — the commit graph is the only sequencing witness
  (`verification-claim-integrity.md` §2.3).

## §D Decisions

Three decisions in this card shape what the guard reports.

[HARD] **All three are operator-unanswered decisions.** Each was put to the operator and no
answer came inside the window, so each is the lead's / this SPEC's judgment rather than an
operator ruling, and each stays reviewable at the Implementation Kickoff Approval gate. The
project runs `interview.recommendation_mode: pull`; the alternatives are recorded with their
measured cost and no recommendation label. A reviewer who disagrees changes the decision here,
not in the implementation.

### D1 — Noun class

**Adopted**: `{retained agent(s), agent catalog, agent roster, retained catalog,
N-agent catalog}` — 62 files / 75 hits on the digit axis.

| Alternative | Measured cost | Why not adopted |
|---|---|---|
| Narrow: `{retained agent(s)}` | 37 files / 54 hits | Provably misses the live `11-agent catalog` stale sites in `moai-foundation-core` and `moai-foundation-quality`; the class would be blind to the exact wording the drift actually used |
| Unrestricted: `{agent(s)}` | 267 files / 613 hits | The allowlist needed to tame 613 hits becomes the guard's actual subject matter — the exemption list, not the roster, is then what a reader reviews |

### D2 — Discharge rule

**Adopted**: a hit discharges only on a `ClaimCount` row for its path, or an explicit
numeral-exempt declaration carrying a mandatory non-empty reason.

| Alternative | Measured cost | Why not adopted |
|---|---|---|
| Reuse the sweep's any-row-by-path rule | Frees exactly **1** additional path today — `internal/web/agentfm.go`, a row carrying `Claims: 0` (measured by the dispatching lead in this tree at HEAD `6abcc85fa`; the row is `web-agentfm-display-rank` in `registry.go`, read here) | A path registered only for a MEMBERSHIP claim would silently discharge a stale COUNT claim inside it — precisely the failure class this card exists to close; a count and a membership are independent claims, and `ClaimKind` already models them as such. The strict rule therefore buys that closure for one extra authored row, not for a large one |

### D3 — Mirror derivation

**Adopted**: the mirror rows are DERIVED, not authored. A `mirrorOf()` helper builds the
`internal/template/templates/` row from its local `.claude/` or `.moai/` counterpart, so a
mirror pair costs one row a reviewer must read instead of two.

The precedent is in the same file: `readmeSite()` (`registry.go`) derives the four README locale
rows from one helper, and its doc comment gives the reason — *"four hand-copied rows is the same
forward-only-propagation shape this package guards, one level up"*. A hand-copied mirror row is
that same shape.

The measured need is **46** new rows or exempt declarations (§A). Roughly half the 46 paths are
template mirrors of a local counterpart, so derivation cuts what a reviewer must read and judge
to roughly 26 while leaving the noun class exactly as D1 adopted it — the guard's reach does not
change, only how many rows carry it. [HARD] The exact post-derivation figure is a run-phase
**re-derivation obligation** (REQ-RNA-012, AC-RNA-013): the mirror pairs have not been counted
exhaustively, and "roughly 26" is a projection, not a measurement.

| Alternative | Measured cost | Why not adopted |
|---|---|---|
| Accept 46 authored rows | 46 rows, of which roughly 30 become exemption prose a reviewer must read | Keeps the full benefit, but realises the hazard `plan.md` §F registers as *"Allowlist becomes the subject matter"* — at that size the exemption list, not the roster, is what review is actually about |
| Narrow the noun class to `{retained agent(s)}` | 37 files / 54 hits (−25 files of reach) | Provably loses the five live `11-agent catalog` stale sites in `moai-foundation-core` and `moai-foundation-quality` — the exact drift wording this card exists to reach |
| Split the catalog/roster family into a follow-up card | Residual today drops, but those same five sites stay unregistered until that card is picked | Defers the card's own subject matter; the hole stays open for an unbounded interval |

Empty-reason handling follows the exemption-marker precedent in
`verification-claim-integrity.md` §2.1: an empty or whitespace-only reason produces a finding
reporting the declaration as incomplete rather than a suppression, so "silence the warning"
never becomes cheaper than "declare the reason".

## §E Out of Scope

### Out of Scope — repairing the stale prose

- The newly-reached stale sites (the `11-agent catalog` / `11 retained agents` wording in
  `model-policy.md`, `moai-foundation-core`, `moai-foundation-quality` and their template
  mirrors) are REGISTERED with a `KnownStale` marker, not rewritten. This follows the
  registry's existing stance — repair is deliberately not done by the guard card. The markers
  make the staleness enumerable instead of silent; the prose repair is a separate card.
- `.moai/project/product.md` and `.moai/project/tech.md` are already registered; their prose is
  likewise untouched here.

### Out of Scope — historical citations

- `the then-8-agent catalog` (live in `.claude/agents/moai/manager-docs.md`,
  `.claude/agents/moai/manager-spec.md` and their `internal/template/templates/` mirrors, read
  in this tree) and `17->8 agent catalog` in `internal/template/*_test.go` describe the roster
  AS IT WAS. They are not drift and are not repaired.
- Whether they are excluded by a tense/selector mechanism or by a reasoned per-path exempt
  declaration is a run-phase implementation choice. The OUTCOME is not deferred: AC-RNA-015
  binds it either way — no finding, no repair, and any path handled by exemption carries a
  reason a reviewer can disagree with.
- Dated research records under `.moai/research/` are outside the layer's scope entirely
  (REQ-RNA-013), not merely unrepaired.

### Out of Scope — lowering SweepThreshold

- Lowering the enumeration threshold is explicitly not a substitute for this layer and is not
  attempted. `tech.md` names zero agents; no threshold reaches it.

### Out of Scope — other roster axes

- `AxisDefinitionFiles` and `AxisSubsetByDesign` count claims are handled by the existing
  per-site assertions. This card adds a discovery layer for the retained-roster count claim; it
  does not redefine the axis model.

### Out of Scope — non-roster numeral claims

- Counts about SPECs, skills, rules, locales, or programming languages are outside the noun
  class by construction (D1) and are not brought in.

## §F Success Criteria

- A count-only roster claim added anywhere in the layer's scope, with no registry row and no
  exempt declaration, fails the guard (AC-RNA-001).
- Every hit is accounted for — registered, exempted, or neutralised by mechanism — with none
  left implicit (AC-RNA-013), and the run-phase population is re-derived in-run (AC-RNA-014).
- Every hit is printed one line per hit at column 0, so breadth is comparable (AC-RNA-006).
- The control probe is observed to fire (AC-RNA-007); a zero-hit run fails (AC-RNA-008).
- `CLAUDE.md 4 (13 retained agents)` pins to 13, not 4 (AC-RNA-003).
- The word axis fires on synthetic input while its live population reads 0 (AC-RNA-005).
- Every numeral-exempt row carries a non-empty reason; an empty reason produces a finding
  (AC-RNA-004).
- `go test -count=1 ./internal/harness/rosterguard/...` green in the tree being changed.

🗿 MoAI
