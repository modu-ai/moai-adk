# SPEC-MOVING-REF-GUARD-001 — Implementation Plan

> Milestones are ordered by decision-reversibility: the judgment content that review is most likely
> to change comes first, mechanical work last. M1 and M2 are where a reviewer's disagreement is
> cheapest to absorb; M5 is where it is most expensive.

## §A. Context

Card t342. Origin: t241 prediction-ledger adjudication (lane-14) VC-6 — two occurrences, one of
which survived to adoption. The measured instance and its consequence are in `spec.md` §A / §B.

The deliverable is a lint warning **plus** the predicate that decides whether the warning should be
acted on by pinning. The card marks the second half [HARD] and names the dominant failure mode:
ship the warning alone and the next reader mechanically pins everything, destroying the subject-class
claims that are correct as written.

## §B. Known issues carried in

- **42 candidate lines already exist** in `.moai/specs/**` (`spec.md` §B.3). Any severity above
  `warning` reds the corpus on day one and the rational response is bulk suppression — the outcome
  this SPEC prevents. Severity is fixed at `warning` by REQ-MRG-001 and D.5.

  **Corrected at v0.8.0**: `warning` alone is not sufficient, and this line said it was. `--strict`
  escalates any non-advisory warning, and the run-phase build red `develop`'s SPEC Lint job on 110
  `MovingRefUnpinned` findings — 2.6× the 42 estimated here. The finding is now emitted advisory,
  which is what makes the day-one claim above true. See `spec.md` §D.5 and §D.7.
- **The three-dot form is not a safe exemption** (`spec.md` §B.2, measured). A run-phase
  implementer is likely to propose exempting `...`; AC-MRG-004 is the guard against it.
- **The corpus contains both predicate classes and all four remedies** (ANCHOR, SUBJECT/S1,
  SUBJECT/S2; R1-R4 — `spec.md` §D.2/§D.3), so a detector tuned only on the
  true-positive shape will fire on `AC-COORD-016` and `REQ-LB-006`. That is expected and is what the
  marker exists for (L4).

## §C. Pre-flight

Run in the worktree, before the first M3 edit:

1. `git rev-parse --show-toplevel` → the t342 worktree path.
2. `git rev-parse --short HEAD` and `git branch --show-current` → re-read immediately before each
   commit, never carried from an earlier turn.
3. `go build ./internal/spec/...` → rc 0 (establishes the pre-change baseline compiles).
4. `git rev-parse origin/develop` → **record the value in `progress.md` §E.2 at capture time** as
   `BASELINE_SHA`. This SPEC's own PRESERVE evidence uses R2 (freeze at pre-flight), not a moving
   ref. Doing otherwise would make the SPEC an instance of its own defect.
   - Recorded: `BASELINE_SHA = d566ecc7511e1954e3aeb1dff3a60afa5be1089b` (2026-08-28).
5. `git merge-base HEAD origin/develop` → **record the value as a SECOND frozen literal**,
   `MERGE_BASELINE_SHA`. Step 4's literal names a point on mainline *before* this branch diverged,
   so any range anchored there spans mainline as well as this branch — which is correct for
   "what has this work changed" and wrong for "what has this branch added beyond mainline". The
   second question is the one AC-MRG-010 asks, and it needs its own anchor.
   - Recorded: `MERGE_BASELINE_SHA = 5e194bba27c146d8c2157d92b4a3fb3995919ff0` (2026-08-28).
   - Re-recorded after absorbing `origin/develop` at the integration window:
     `MERGE_BASELINE_SHA = c48e8aaa251950e194025ed6a8c4733d61b88cde` (2026-08-29). This is the
     value AC-MRG-010 is decided against from here on. The 2026-08-28 entry is kept rather than
     overwritten: erasing a superseded anchor destroys the record of what was measured when, which
     is the same information loss this SPEC exists to prevent.
   - Re-recorded after the second absorption, at the integration window that merged this card:
     `MERGE_BASELINE_SHA = 51daada00079ee818ae29adb0c2d5c34c5071a29` (2026-08-29). This is the
     value AC-MRG-010 is decided against from here on. Both earlier entries are kept — the record
     of which anchor decided which measurement is the point, and overwriting it would erase exactly
     what this SPEC exists to preserve.
   - **[HARD] Re-record on every mainline absorption.** A merge-base is computed, and this SPEC
     forbids deciding against a computed-at-read-time anchor — so it is frozen here instead. The
     freeze has a cost: the literal goes stale the moment this branch merges mainline again, and
     the criterion it decides then reds on absorbed mainline commits exactly as step 4's literal
     did. Any further absorption therefore obliges a new dated entry above, before AC-MRG-010 is
     decided again. This is limit L6 acting on this SPEC's own evidence, handled the way the
     doctrine says to handle it: the command is the criterion, the value is a dated reference.

## §D. Constraints

- Severity `warning` only (D.5). Never `error`, never era-demotable.
- The rule reads sibling artifacts (`plan.md`, `acceptance.md`, `progress.md`) via
  `filepath.Dir(doc.Path)`; `SPECDoc.Body` carries `spec.md` only. Precedent for a rule that reaches
  past the single document exists in the engine (`HaikuResidualRule`, which carries its own
  `baseDir`).
- No change to `lint.skip`, to `eraDemotableCodes`, or to any existing rule's severity.
- Template mirror is mandatory for the doctrine file, and must pass the neutrality guard: no SPEC
  IDs, no dates, no commit SHAs, no `CLAUDE.local` references in the mirrored copy.
- No `go test ./...`. Package-scoped only: `go test ./internal/spec/...`.
- Nothing under `e2e/`.
- No push, no PR.

## §E. Self-verification

Each milestone closes by recording, in `progress.md` §E.2: the command run, its verbatim output, the
tree SHA it ran against, and what was **not** observed. A milestone whose evidence cites a moving
ref is rejected on its own SPEC's terms.

## §F. Milestones

### M1 — The predicate and the limits (doctrine)

**Most reversible under review; therefore first.** No code.

Author the doctrine section carrying: the anchor-or-subject predicate and its **four** tests
(`spec.md` §D.1, Test 4 routing within the SUBJECT class); the §D.1 statement that classification
and remedy selection are two separate steps; the **four** remediation branches, R2 the anchor-class
default and R4 for live-state claims (§D.2); the **five** grounded instances (§D.3); the **six**
detection limits (§F). Placement: a section in
`.claude/rules/moai/core/verification-claim-integrity.md`, which already owns the
baseline-attribution invariant this defect violates, rather than a new always-loaded file.

Decides AC-MRG-007, AC-MRG-011.

**Why first:** every later milestone encodes this text. If a reviewer disagrees with the predicate's
shape, changing it here costs a paragraph; changing it after M3 costs the rule, its tests, its
message strings, and the mirror.

### M2 — The exemption marker surface

**Second-most reversible; user-facing and hard to change once written into documents.**

Fix the marker syntax (`<!-- moving-ref-ok: <reason> -->`), the mandatory non-empty reason, the
line-scope rule (flagged line or the line immediately above), and the incomplete-marker behaviour
(REQ-MRG-003). Record the Q1 decision, or carry Q1 to the operator if unresolved.

Decides AC-MRG-003.

**Why before M3:** the marker is the thing authors will have typed into artifacts. Once even a few
exist, changing the syntax means editing every one.

### M3 — The detector

`internal/spec/lint_movingref.go` implementing `Rule` with `Code() == "MovingRefUnpinned"`:
moving-ref token detection in a git-command context, invariant-claim marker matching, SHA-pin and
frozen-baseline-variable exclusion, three-dot non-exemption, marker suppression, divergence-figure
variant (REQ-MRG-006), and a message naming all four remediation branches.

**The R4-form exclusion (REQ-MRG-010) is NOT built here — deferred by operator decision on
`spec.md` §H Q0, option C.** It was the delicate one: too loose and it is a blanket bypass that
silently disables the rule, too tight and it flags the recommended remedy — and with R4's reachable
class measured empty (§B.7, corroborated by M4's external-occupancy 0), only the first error was
available to it. Deferring removes the bypass rather than trying to aim around it. AC-MRG-013 and
its two counter-mutations (CM-1 positional, CM-2 command-token) defer with it, unrun; early R4-form
lines are silenced with the R3 marker instead. R4 stays in the finding message — the doctrine
carries the remedy; only its lint exclusion is deferred.

[HARD] **When that exclusion is eventually built, it MUST key on imperative structure, never on a
command token.** This is settled, not open, and it survives the deferral. The iter-1 audit
demonstrated a fetch-verb-keyed exclusion passing all thirteen criteria then in force while
silencing 76 of 117 unpinned divergence lines (`spec.md` §B.6). A token key is forgeable by
construction.

Registered in the `l.rules` slice in `internal/spec/lint.go`.

Tests in `internal/spec/lint_movingref_test.go`, fixture-driven under `internal/spec/testdata/`.
Every test states the mutation that makes it fail (see `acceptance.md` §D).

Decides AC-MRG-001, -002, -004, -005, -006, -008, -009, -013, -014.

### M4 — Corpus triage record (classification only)

Run the rule over `.moai/specs/**`, record the finding count, and classify each finding into anchor
/ subject / already-frozen with a one-line reason. Written to `.moai/reports/t342/corpus-triage.md`.

**[HARD] M4 edits no SPEC artifact.** Remediating a finding belongs to that finding's owning SPEC.
A milestone that bulk-pins the corpus would be this card enacting the failure mode it was written
to prevent, and the triage record exists precisely so the classification survives without the edit.

Decides AC-MRG-010.

### M5 — Template mirror and build

Mirror the M1 doctrine into `internal/template/templates/.claude/rules/moai/core/`, neutralized;
run `make build`; confirm the neutrality guard passes.

Decides AC-MRG-012.

**Last because it is purely mechanical** — it encodes M1's final text and has no decision content of
its own.

## §G. Anti-patterns

- **Bulk-pinning the corpus.** The card's named dominant failure mode. Guarded by M4's [HARD] clause
  and by REQ-MRG-004 (the message names four branches), each of which §D.4 prices so that none is
  a free silencer.
- **Exempting the three-dot form.** Disproved by measurement in `spec.md` §B.2; guarded by
  AC-MRG-004.
- **A bare marker with no reason.** Would make silencing cheaper than fixing; guarded by AC-MRG-003,
  which requires the empty-reason case to still produce a finding.
- **Believing the class names the remedy.** Two classes, four remedies (`spec.md` §D.1). A reader
  who collapses the two steps has only as many remedies as classes and takes the first that fits —
  a second, subtler route to the same "pin everything" outcome.
- **An R4 exclusion wide enough to be a bypass.** Passes AC-MRG-013 while silently disabling
  AC-MRG-001; guarded by that criterion's mandatory counter-mutation. Not reachable in this SPEC
  after the Q0 option-C deferral — no exclusion is built here — but the bullet is kept because the
  hazard travels with REQ-MRG-010 to whatever work implements it.
- **Testing only one direction of an exemption.** CM-1 and CM-2 both ask whether the exclusion is
  too *wide*; until iter-2 nothing asked whether it works *at all*, and the fixture was in fact
  unflaggable. Both directions or neither — an exemption needs a criterion that fails when the
  exemption is absent, not only ones that fail when it is too broad.
- **Adopting a lead-supplied instance as a fixture without measuring it.** The vacuous AC-MRG-013
  fixture was the lead's dispatch line, promoted because the lead had identified it as an instance of
  the defect. The prose intuition was correct and the mechanical question was never asked; the
  authority of the source is what suppressed the check. Measure a fixture's properties on adoption,
  whatever its provenance (`progress.md` §E.1).
- **Keeping a clause alive by manufacturing a scope for it.** §B.7 measured R4's reachable scope as
  empty and says so, rather than describing the class in a way that implies occupants. An exemption
  with no scope is a dead clause; label it and let the operator decide (§H Q0).
- **An acceptance criterion that cannot fail.** Every criterion in `acceptance.md` names its
  falsifying input. A guard whose criterion cannot fail is indistinguishable from a guard that is
  switched off — and this SPEC's deliverable *is* a guard, so the hazard is doubled here.
- **Verifying this SPEC's own PRESERVE claims against a moving ref.** Guarded by §C step 4.

## §H. Cross-references

- `.claude/rules/moai/core/verification-claim-integrity.md` §2 — the baseline-attribution invariant
  this defect violates; M1's placement target.
- `internal/spec/lint.go` — the `Rule` interface, the `l.rules` registration slice, the
  `Severity` / `Advisory` / `Strict` exit-code semantics D.5 relies on.
- `SPEC-GRAPH-FRESHNESS-CADENCE-001` v0.2.2 HISTORY — grounded instance 2, and the source of the
  predicate's generality beyond git refs.
