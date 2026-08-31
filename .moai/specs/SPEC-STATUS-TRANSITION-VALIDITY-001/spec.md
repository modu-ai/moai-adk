---
id: SPEC-STATUS-TRANSITION-VALIDITY-001
title: "Status-transition validity lint — a transition is judged on the pair, not on who signed it"
version: "0.1.0"
status: in-progress
created: 2026-08-31
updated: 2026-09-01
author: manager-spec
priority: P2
phase: "v3.1.3 target"
module: "internal/spec"
lifecycle: spec-anchored
tier: M
tags: "spec-lint, status-transition, lifecycle, card-t376, internal-spec"
---

# Status-transition validity lint

Card: **t376**. Evidence: `.moai/reports/t376/`.

## Overview

`moai spec lint` today has no rule that judges whether a SPEC's `status:` transition is a legal
edge of the lifecycle DAG. `draft → completed` and `completed → draft` both pass silently, measured.
The one rule that reads a transition at all — `OwnershipTransitionRule` — asks a different question
(*did the right agent sign it?*), answers it only when the commit carries an `Authored-By-Agent:`
trailer, and delegates every matrix-undefined pair to a rule that structurally cannot judge a
transition.

This SPEC adds a rule that judges the **pair itself**, independently of who authored the commit, and
corrects one demotion message that misstates its own cause.

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-08-31 | Initial plan-phase artifacts (card t376) |

## §A — Pre-existing State Survey

### A.0 — Coordinate provenance (the card's line numbers were stale)

The card text and its dispatch carried line numbers read from the primary checkout on `main`
(`48239c7dc`). Those coordinates are **superseded**. Every line number in this SPEC was measured in
this worktree at tree `3f03d9c36` (`WT-status-transition-gap`), and every count below was
re-measured here rather than carried over.

### A.1 — Baseline: 1096 findings, per rule code

Command (output on disk at `.moai/reports/t376/lint-baseline.json`):

```bash
moai spec lint --json
jq -r '.[].code' .moai/reports/t376/lint-baseline.json | sort | uniq -c | sort -rn
```

Observed:

| Code | Count |
|---|---|
| CoverageIncomplete | 846 |
| MovingRefUnpinned | 114 |
| LegacyEARSKeyword | 43 |
| ModalityMalformed | 25 |
| MissingExclusions | 24 |
| StatusGitConsistency | 18 |
| FrontmatterInvalid | 14 |
| InvalidREQID | 6 |
| SyncSHASlotFormat | 5 |
| OwnershipTransitionInvalid | 1 |
| **Total** | **1096** |

`find .moai/specs -maxdepth 1 -mindepth 1 -type d | wc -l` → **716** directories under
`.moai/specs/` at the time of the baseline run, before this SPEC's own directory existed (717 after).
One `OwnershipTransitionInvalid` finding across that corpus is the measurement that decides the
scope below: a fix confined to the ownership matrix would be correct on paper and vacuous in
practice.

### A.2 — Execution probe: 8 cases, real git repos

`go test ./internal/spec/ -run TestT376Probe` (log: `.moai/reports/t376/probe-transition-gap.log`).
Each case builds a real repository, lands two commits carrying one status transition, and lints.

| Case | Transition | 2nd-commit trailer | Era | Transition finding |
|---|---|---|---|---|
| draft_to_completed | draft → completed | manager-develop | V2.x | **none** |
| completed_to_draft | completed → draft | manager-spec | V2.x | **none** |
| implemented_to_completed_wrong_owner | implemented → completed | manager-develop (wrong) | V2.x | OwnershipTransitionInvalid, advisory=true |
| implemented_to_completed_right_owner | implemented → completed | manager-docs | V2.x | none |
| implemented_to_completed_no_trailer | implemented → completed | (absent) | V2.x | none |
| draft_to_inprogress_wrong_owner | draft → in-progress | manager-docs (wrong) | V2.x | OwnershipTransitionInvalid, advisory=true |
| draft_to_inprogress_wrong_owner_modern | draft → in-progress | manager-docs (wrong) | V3R6 | OwnershipTransitionInvalid, **advisory=false** |
| draft_to_completed_modern | draft → completed | manager-develop | V3R6 | **none** |

The `zz_t376_probe_test.go` probe is scratch, not a deliverable. It establishes the four rows this
SPEC treats as ground truth: the two reversal/skip pairs are silent in both eras, the wrong-owner
pair fires, and the same pair without a trailer does not.

### A.3 — The four layers (measured at 3f03d9c36)

1. **`internal/spec/lint_ownership.go:78-94`** — `expectedOwnerForTransition(prev, curr)`: the
   `case "draft", "planned"` arm accepts only `in-progress|implemented`, so `draft → completed`
   falls through to `ownerNone`, and `Check` returns nil at the `expected == ownerNone` guard
   (`:404-407`). The `:61` comment delegates matrix-undefined transitions — naming status reversal
   as its example — to `StatusValueEnumRule`.
2. **`internal/spec/lint.go:1105-1110`** — `StatusValueEnumRule.Check(doc *SPECDoc, _ []*SPECDoc)`:
   the signature carries no previous status, so a *transition* cannot be expressed in it at all.
   `completed` is a valid enum member and passes whatever it came from. The delegation in (1) is
   therefore vacuous for the very example its comment cites, which case `completed_to_draft`
   confirms by execution.
3. **`internal/spec/lint.go:1318-1320`** — `StatusGitConsistencyRule.Check` returns early on
   `terminalStatusEnum[fm.Status]`, and `completed: true` sits in that map (`:1290-1295`).
   Separately, `.github/workflows/spec-lint.yml:31` uses `actions/checkout@v7` with no
   `fetch-depth`, so CI lints a shallow clone. That second observation is **config-read only, not
   execution-observed** — recorded as a Gap, not as a verified defect.
4. **`internal/spec/lint.go:239`** — not in the card text; found during this investigation.
   `demote := isGrandfatheredSpecDir(...) || terminalStatusEnum[doc.Frontmatter.Status]`, and
   `applyEraDemotion` (`:296-311`) sets `Advisory = true` on **every** warning of a demoted document,
   not only on `eraDemotableCodes` members. `--strict` escalates only warnings with `!Advisory`
   (`:61`). Writing `status: completed` therefore marks every warning on that SPEC advisory:
   `completed` shelters itself. The probe's last two rows isolate the two disjuncts —
   `draft_to_inprogress_wrong_owner_modern` (`completed` absent, modern era) is the only row in the
   set that produces `advisory=false`.

   From the same path: when demotion fires because of terminal status rather than grandfathering,
   the appended text still reads `[grandfathered era — downgraded to warning]`. The annotation
   misstates its own cause.

### A.4 — Transition census: the exact population the rule will judge

Log: `.moai/reports/t376/transition-census.log`. The census calls
`lookupOwnershipTransitionFromGit` directly, so this is the population the new rule sees, not a
proxy for it. **714** SPEC directories carrying a `spec.md` were scanned; **713** produced a
transition record and **1** produced none.

| N | Edge | N | Edge |
|---|---|---|---|
| 217 | `in-progress → completed` | 125 | `implemented → completed` |
| 104 | `(none) → completed` | 50 | `draft → implemented` |
| 50 | `draft → completed` | 48 | `completed → implemented` |
| 23 | `in-progress → implemented` | 17 | `in-progress → archived` |
| 15 | `(none) → implemented` | 12 | `draft → archived` |
| 11 | `planned → implemented` | 10 | `(none) → draft` |
| 5 | `draft → in-progress` | 4 | `planned → completed` |
| 4 | `(none) → in-progress` | 4 | `draft → superseded` |
| 3 | `Completed → completed` | 2 | `completed → superseded` |
| 2 | `(none) → archived` | 1 | `synced → completed` |
| 1 | `implemented → superseded` | 1 | `Superseded → superseded` |
| 1 | `cancelled → rejected` | 1 | `(none) → "in-progress"` |
| 1 | `approved → completed` | 1 | `archived → superseded` |

Two figures carry the SPEC's own claims:

- **`draft → completed` = 50.** The defect this card targets is not hypothetical; it is 50
  occurrences of a status that skipped both run and sync and landed terminal.
- **`completed → implemented` = 48.** The reversal direction is real too, at comparable scale.

Three further shapes in this table are decided in §A.5, because they are not settled by the DAG
alone.

### A.5 — Decisions taken on the census (all recorded, none silent)

**D1 — `draft → implemented` is VALID (50 occurrences).** `expectedOwnerForTransition`
(`lint_ownership.go:80-82`) already accepts `draft|planned → implemented` and assigns it to
manager-develop. Ruling it invalid would put this rule in direct contradiction with the rule sitting
beside it in the same file, and would emit 50 findings on historical SPECs.

> **Accepted debt, stated rather than implied.** The SSOT matrix in
> `spec-frontmatter-schema.md` § Status Transition Ownership Matrix carries **no
> `draft → implemented` row**. Code and SSOT therefore disagree, and this card knowingly adopts the
> code's reading without repairing the divergence. Nothing here should be read as the SSOT saying
> `draft → implemented` is canonical — it does not say so. Closing the divergence (either adding the
> row or narrowing the Go arm) is a later card.

**D2 — the finding GATES; no emission-site `Advisory` flag.** The decision stands on this: an
emission-site `Advisory` would make the rule unable to gate anything, anywhere, which would close the
gap in appearance only. A rule that cannot fail is the defect class this card exists to remove, and
shipping one as the fix is not available.

> **Corrected rationale — the original reason was false against this SPEC's own census.** D2 was
> first justified by the claim that "the overwhelming majority of what this rule would flag ends in
> `completed`", and therefore sits under the layer-4 demotion anyway. That premise came from the
> instruction handed to this SPEC and was corrected by plan-audit iter-1 (D1). The census says
> otherwise: of the ~98 projected findings, **50** end in `completed` and **48** end in
> `implemented`, which is not in `terminalStatusEnum` (`internal/spec/lint.go:1290-1295`:
> superseded / archived / rejected / completed). So it is **roughly half**, not an overwhelming
> majority. Whether those 48 demote at all depends on the era classification of their documents,
> which **this SPEC does not measure** — the gating population is therefore unknown until M2
> measures it (AC-STV-019). The decision is unchanged; only its stated reason was wrong, and it is
> recorded here so a later reader does not re-derive the false one.

**D3 — an unrecognized status token gets its OWN finding code.** The census found status values
outside the canonical 8-value enum in real history: `synced`, `approved`, `cancelled`, the case
variants `Completed` / `Superseded`, and one quote-wrapped `"in-progress"`. "This transition is
wrong" and "I do not recognize this status value" are different facts with different remedies, and
reporting the second as the first would make the message misstate its cause — the same defect class
as the demotion-message misattribution this card already fixes. Passing an unrecognized token
silently is explicitly rejected: passing what it does not recognize is exactly how layers 1 and 2
fail.

**D4 — `in-progress → completed` is VALID (217 occurrences, the largest bucket).** Derived from the
SSOT, not invented: matrix row 3 reads `in-progress → implemented → completed` with the note that
the `completed` transition **is merged into the single sync commit**. A commit that moves
`in-progress` straight to `completed` is therefore the canonical 3-phase close as the matrix
describes it, not a skipped phase. Reading row 3 as requiring a separate `implemented` write would
flag the single most common close shape in the corpus.

**D5 — `(none) → X` is UNJUDGEABLE and is skipped (136 occurrences across six rows).** `(none)` does
not mean "the SPEC was created at status X". It means the extractor found an added `status:` line
with no removed counterpart inside its lookback window — the shape a squash merge produces, and the
shape a truncated history produces. The extractor cannot distinguish "created at X" from "earlier
history not visible here", so the pair is not evidence of a transition at all. Judging it would
convert a measurement limitation into 136 findings, of which `(none) → completed` alone is 104.
`(none) → draft` is skipped by the same rule rather than by a special case — the skip is about what
the extractor knows, not about which target is canonical.

> D4 and D5 were **not** in the decision set handed to this SPEC; they were forced by the census and
> are recorded here for the same reason D1's debt is: a decision nobody wrote down is a decision
> nobody can overrule.

### A.6 — Projected corpus impact (derived from A.4, NOT measured)

Applying §A.7's set to the §A.4 census by hand yields roughly **98** `StatusTransitionInvalid`
(50 `draft → completed` + 48 `completed → implemented`) and **7** `StatusTokenUnrecognized`
(`Completed` 3, plus `synced`, `approved`, `Superseded`, `cancelled` at 1 each = 3+1+1+1+1 — the
quote-wrapped `"in-progress"` sits behind a `(none)` skip and never reaches the token test, which is
itself a consequence of check ordering worth confirming).

These are **projections computed from a table, not observations of the rule running.** They exist so
that M2's measured numbers have something to be compared against and so a large divergence is
noticed. A projection that matches is weak evidence; a projection that misses is a finding.

### A.7 — The canonical transition set (derived, not invented)

Derived from `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum and
§ Status Transition Ownership Matrix. This SPEC introduces no parallel definition.

| Edge | Source |
|---|---|
| `draft → in-progress` | matrix row 2 |
| `in-progress → implemented` | matrix row 3 (`in-progress → implemented → completed`) |
| `implemented → completed` | matrix row 3 |
| `in-progress → completed` | matrix row 3 + its single-sync-commit note — §A.5 D4 |
| `completed → in-progress` | matrix row 7 — declared amendment, legitimate |
| `draft → implemented` | §A.5 D1 — adopted from `lint_ownership.go:80-82`; **the SSOT has no such row** |
| `* → superseded` | matrix row 4 |
| `* → archived` | matrix row 5 |
| `* → rejected` | matrix row 6 |
| edges touching `planned` | § Status Enum — legacy-optional, no active-flow owner; tolerated, never flagged |

Two classes sit outside the set without being invalid:

- `(none) → X` — not judged at all (§A.5 D5). The extractor cannot tell creation from truncation.
  This subsumes the former `(none) → draft` row, which is skipped rather than accepted.
- a pair naming a token outside the canonical 8-value enum — reported under a **different** code
  (§A.5 D3), because the fact being reported is different.

Everything else is invalid. `draft → completed` and `completed → draft` are the two the probe
measured as silent; `completed → implemented` is the same reversal class the census counted 48 times.

## §B — Requirements (GEARS)

### REQ-STV-001 (Ubiquitous)

The lint engine shall validate the `(previousStatus, currentStatus)` pair of a SPEC's most recent
`status:` transition, read from git history, against the canonical transition set of §A.7, and shall
emit a `StatusTransitionInvalid` finding for any pair outside that set.

### REQ-STV-002 (Ubiquitous)

The lint engine shall reach its `StatusTransitionInvalid` verdict without reading the
`Authored-By-Agent:` commit trailer, so that a commit carrying no trailer is judged identically to
one that does.

### REQ-STV-003 (Event-driven)

**When** a SPEC's most recent status transition is `draft → completed`, the lint engine shall emit a
`StatusTransitionInvalid` finding naming both statuses and the commit SHA that performed the
transition.

### REQ-STV-004 (Event-driven)

**When** a SPEC's most recent status transition is `completed → draft`, the lint engine shall emit a
`StatusTransitionInvalid` finding naming both statuses and the commit SHA that performed the
transition.

### REQ-STV-005 (Unwanted)

The lint engine shall not emit `StatusTransitionInvalid` for any edge in the canonical set of §A.7,
including the declared `completed → in-progress` amendment edge, the single-sync-commit
`in-progress → completed` close, and the `draft → implemented` edge adopted in §A.5 D1.

### REQ-STV-006 (Capability gate)

**Where** a transition edge names the legacy-optional `planned` status on either side, the lint
engine shall accept the edge without emitting a finding, because §A.7's source declares `planned` to
have no active-flow owner and grandfathered history carries it.

### REQ-STV-007 (Capability gate)

**Where** git history is unreachable or carries no status transition for the SPEC, the lint engine
shall skip the check without emitting any finding at error severity.

### REQ-STV-008 (Ubiquitous)

The demotion annotation appended by `applyEraDemotion` shall name the cause that actually fired —
grandfathered era, or terminal lifecycle status — rather than naming grandfathered era in both
cases.

### REQ-STV-009 (Ubiquitous)

The new rule shall emit at warning severity **without** setting the emission-site `Advisory` flag, so
that a finding on a modern-era, non-terminal SPEC is escalated by `--strict` (§A.5 D2).

### REQ-STV-012 (Unwanted)

The new rule shall observe only: it shall not mutate any SPEC file, and it shall not alter
`terminalStatusEnum`, `eraDemotableCodes`, the `StatusGitConsistencyRule` early return, or the
demotion decision at `lint.go:239`.

### REQ-STV-010 (State-driven)

**While** reporting a post-change `moai spec lint` baseline, the run phase shall report the finding
count **per rule code** and compare it against the §A.1 per-code table, attributing each movement to
its own code rather than to a single aggregate delta.

### REQ-STV-011 (Ubiquitous)

The regression test set shall include, in one execution, at least one case that MUST fire, so that
an all-silent run is distinguishable from a harness that stopped working.

### REQ-STV-013 (Event-driven)

**When** either side of the transition pair names a status token outside the canonical 8-value enum,
the lint engine shall emit a `StatusTokenUnrecognized` finding naming the unrecognized token, and
shall not report that document under `StatusTransitionInvalid` for the same pair.

### REQ-STV-014 (Capability gate)

**Where** the previous status of the pair is absent — the `(none) → X` shape the extractor produces
for an added `status:` line with no removed counterpart — the lint engine shall skip both the
transition-validity check and the token check for that document, because the pair is not evidence of
a transition (§A.5 D5).

### REQ-STV-015 (Unwanted)

The `StatusTokenUnrecognized` finding shall not duplicate what `StatusValueEnumRule` already reports
for the same document: the two shall be shown, by measurement rather than by assumption, to describe
disjoint facts — the frontmatter's current value versus a token seen in git history.

## §C — Exclusions (What NOT to Build)

Each item below is a non-goal of this card. None of them is closed by this card, and none of them
closes this card.

### Out of Scope — StatusGitConsistencyRule terminal early return

- The `terminalStatusEnum[fm.Status]` early return at `internal/spec/lint.go:1318-1320` (layer 3)
  stays exactly as it is. This card adds a rule beside it; it does not reopen it.

### Out of Scope — the era / terminal-status demotion path

- The demotion decision at `internal/spec/lint.go:239` and the blanket
  `Advisory = true` on every warning of a demoted document (layer 4) stay as they are. Only the
  *message* is corrected (REQ-STV-008).
- **Stated, accepted limitation.** Because that path is untouched, a `StatusTransitionInvalid`
  finding on a demoted document is marked `Advisory = true` and is therefore not escalated by
  `--strict`. Two distinct mechanisms demote, and both shelter findings this card exists to catch.
  The first is **terminal lifecycle status**: a SPEC whose frontmatter says `status: completed` is
  demoted by the terminal-status branch, so `draft → completed` is reported but does not gate,
  because the very status it transitioned into shelters it. The second is the **grandfathered-era
  exemption**, a policy carve-out rather than a lifecycle property: a document on a grandfathered-era
  directory is demoted regardless of its frontmatter status, including statuses such as `implemented`
  that are not in `terminalStatusEnum`. Measured on this corpus at commit `b1bcce4f4`, the 97
  `StatusTransitionInvalid` findings split **49** sheltered by terminal status and **48** sheltered
  *solely* by the era exemption. That split is a state of one tree at one moment, not a permanent
  property of the design: card t382 (`SPEC-ERA-H3-NARROWING-001`) is narrowing the era exemption, and
  if it narrows, the era-sheltered 48 are the population that begins gating. Every such finding stays
  visible in `--json` output and in the per-code baseline; it is not a gate. This is a known
  consequence of the scope boundary, recorded here rather than left as a silent gap.

### Out of Scope — spec-lint.yml fetch-depth

- `.github/workflows/spec-lint.yml` shallow-clone depth is **card t371**, not this card.
- **t371 landing `fetch-depth: 0` does not close this card.** Layer 3's terminal early return fires
  before clone depth is ever consulted, and layers 1-2 are unrelated to clone depth. Two cards that
  appear to cover each other get neither one fixed.

### Out of Scope — ownership matrix and trailer semantics

- `expectedOwnerForTransition`, `trailerAgentOwnerKind`, and the `Authored-By-Agent:` convention are
  unchanged. The new rule sits beside `OwnershipTransitionRule`, answering a different question; the
  existing rule keeps its own trailer-gated behavior and its silent skip.

### Out of Scope — the `draft → implemented` SSOT divergence

- §A.5 D1 adopts the Go arm's reading of `draft → implemented` and leaves the SSOT matrix, which has
  no such row, untouched. This card does not decide which of the two is right; it records that they
  disagree and picks the one that does not contradict the rule beside it. Reconciling them — adding
  the matrix row, or narrowing `expectedOwnerForTransition` — is a later card.

### Out of Scope — unrecognized status tokens in the corpus

- The rule reports `StatusTokenUnrecognized` where it finds one (§A.5 D3). Repairing the ~7 SPECs
  whose history carries `synced`, `approved`, `cancelled`, or a case variant is not part of this
  card.

### Out of Scope — corpus remediation

- Existing SPECs whose recorded history carries an invalid transition are not repaired by this card.
  The card ships the detector and reports what it finds; remediation of whatever it surfaces is
  separate work.

## §D — Acceptance Criteria

The binding AC set is `acceptance.md` in this directory (Tier M). It covers REQ-STV-001 through
REQ-STV-015.

## §E — Cross-references

- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum, § Status Transition Ownership Matrix — the DAG SSOT §A.4 derives from
- `internal/spec/CLAUDE.md` — module conventions (lint rule pattern, observation-only rules, no-false-positive-on-closed-SPECs obligation)
- `.moai/reports/t376/probe-transition-gap.log` — the 8-case execution probe
- `.moai/reports/t376/transition-census.log` — the 714-directory transition census (§A.4)
- `.moai/reports/t376/lint-baseline.json` — the 1096-finding baseline
- Card t371 — `spec-lint.yml` fetch-depth (adjacent, non-overlapping)
