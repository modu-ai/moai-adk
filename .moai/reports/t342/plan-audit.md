# SPEC Review Report: SPEC-MOVING-REF-GUARD-001 (card t342)

Iteration: 1/2 (Tier M ceiling)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.82** (Tier M PASS threshold 0.80)

Reasoning context ignored per M1 Context Isolation. The audit reads only the four
artifacts plus the tree they describe.

> **Report location note.** The dispatch asked for this at the primary path
> `/Users/goos/MoAI/moai-adk-go/.moai/reports/t342/plan-audit.md`. The
> worktree-isolation guard refuses a write from this session to the shared
> checkout, so it is written here instead. `.moai/reports/t342/` is **not**
> gitignored (`git check-ignore` → rc 1), so the lead should copy or commit it
> before the worktree is disposed.

## Audit attribution

| Item | Value |
|---|---|
| Tree | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t342` (`git rev-parse --show-toplevel`) |
| Branch | `WT-moving-ref-guard` |
| HEAD at verdict | `43329ec8b`, working tree clean |
| spec.md sha256 | `b31b993ef51b…` |
| acceptance.md sha256 | `ceec0a435b2b…` |

**The artifacts moved under this audit.** At audit entry HEAD was `b3e25945f` and
spec.md hashed `508633629f9f…`. Between 08:03 and 08:11 the lead addendum
(v0.2.0) landed in four separate writes and was then committed as `43329ec8b`.
A snapshot read at 08:08 showed `REQ-MRG-010`, `L6` and `AC-MRG-013` claimed by
the HISTORY row but absent from the body; by 08:10 all three were present. Those
three are **not** reported as defects below — they were mid-write state, and
reporting them would be the stale-measurement defect this SPEC exists to
prohibit. Every finding below was re-verified against `43329ec8b` after the tree
went clean. The instability itself is recorded as P1.

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -c '^- \*\*REQ-MRG-'` → 10;
  IDs `REQ-MRG-001`..`REQ-MRG-010`, sequential, no gaps, no duplicates, uniform
  3-digit padding (spec.md:382-410).
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement
  layer** (`REQ-XXX` in spec.md), not the verification layer. All 10 carry a
  subject + `shall` + response: Ubiquitous (001, 004, 005), State-driven
  ("While the three-dot form is present…", 007), Where-conditional (002, 003,
  006, 008, 009, 010). See MINOR note D8 on the `Where` usage.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present and
  well-typed (spec.md:1-17): `id` matches `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$`,
  `version` quoted semver, `status: draft`, `created`/`updated` ISO dates,
  `priority: P2`, `phase: "v3.2.0 target"` (not a prohibited lifecycle token),
  `lifecycle: spec-anchored`, `tags` comma-string. No rejected snake_case alias.
  Optional `era`/`tier`/`related_specs` valid. The version/HISTORY divergence is
  reported separately as D6 — it is a content defect, not a schema violation.
- **[N/A] MP-4 Section 22 language neutrality** — single-domain SPEC (this repo's
  own Go lint engine + one doctrine section). No multi-language tooling claim.
  Auto-passes per the MP-4 N/A rule.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — 7 referenced SPEC IDs extracted;
  all 7 resolve on disk. Statuses read: `SPEC-GRAPH-FRESHNESS-CADENCE-001`
  in-progress, `SPEC-VERIFICATION-COMPLETENESS-001` completed,
  `SPEC-V3R6-TEST-REFACTOR-001` completed, `SPEC-V3R6-MULTI-SESSION-COORD-001`
  completed, `SPEC-V3R5-LATE-BRANCH-001` completed,
  `SPEC-DESIGN-MOAIWEBV2-002` completed, `SPEC-V3R6-AGENT-FOLDER-SPLIT-001`
  **superseded**. The superseded one is cited evidentially (§B.3, `AC-AFS-012`
  as a corpus shape) and is reconciled by §G "Retrofitting closed or
  grandfather-era SPECs" being out of scope. No BLOCKING finding.
- **[N/A] MP-6 D8 cross-platform discipline** — `grep -c syscall` → 0 across
  spec.md, plan.md, acceptance.md. Auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -c 'NEEDS CLARIFICATION' plan.md`
  → 0. No `research.md` at Tier M. No open marker.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.80 | 0.75 band | Predicate is unusually well-argued (§D.1.1 defends its own shape; §D.3 names its own classification skew). Deducted for D1 — Test 1's evaluation-time is unspecified and the "Test 1 governs" tie-break routes the predicate's founding case away from the remedy it invented. |
| Completeness | 0.82 | 0.75-1.0 boundary | All required sections present; §G carries three `### Out of Scope — <topic>` H3s with specific bullets (spec.md:459/465/472); six limits stated; Q0-Q3 open rather than guessed. Deducted for D5 (Q0 outside DoD), D6 (version), D7 (plan.md residue). |
| Testability | 0.78 | 0.75 band | Every criterion names its falsifying input and its mutation; AC-MRG-013 carries a counter-mutation; DoD requires each mutation actually planted with verbatim red output. Deducted for D2 (no negative control on the claim-marker conjunct), D3 (era-demotion hazard), D4 (AC-MRG-010 decider). |
| Traceability | 0.88 | 1.0 band, one gap | REQ→AC complete in both directions for all 10 REQs; REQ→milestone complete (M1: 005 · M2: 002/003 · M3: 001/004/006-010 · M5: 005-mirror). One orphan AC — D9. |

## Defects Found

**D1 — the predicate cannot reach R4 for the case R4 was invented for — `spec.md`:§D.1 / §D.3 instance 3 — Severity: critical — Class: blocking**

Test 4 is gated: *"Applies only once Tests 1-3 have returned SUBJECT."* Apply the
tests as written to the lead's dispatch base line (`base: origin/develop
44095ddc2`), the instance §D.1 names as R4's motivation:

- **Test 1** says "replace the ref token with the SHA it resolves to *right now*
  … it still says what it meant → ANCHOR." Substituting `ec15ec2cd` at that
  instant *does* preserve the meaning — §D.3 has to add a consideration Test 1's
  own text does not contain ("the substitution decays immediately, because the
  sentence means 'whatever the tip is when you enter'") to get SUBJECT.
- **Test 3** returns SUBJECT ("variance is the point").
- The tie-break then fires: *"where they disagree, Test 1 governs."*

So the procedure as written returns ANCHOR, Test 4 never runs, and R4 is
unreachable for its founding case. This is not hypothetical: §D.3 records that
v0.1.0 classified this instance ANCHOR → R2, and §D.1.1 names that exact routing
as "precisely the error that produced the stale base line." The v0.2.0 addendum
corrected the *worked example* without correcting the *procedure* that produced
the wrong answer.

This is the card's [HARD] deliverable, and it is the undecidable region the
dispatch asked about: Test 1 does not say **at what time** to re-read the
sentence, and write-time vs read-time is exactly what separates ANCHOR from S2.

**Required fix:** state Test 1's evaluation time explicitly — "re-read the
sentence as a later reader will act on it, not as you read it now" — and amend
the tie-break so a Test 1/Test 3 disagreement routes to Test 4 rather than
resolving to ANCHOR. Record the undecidable region in §H as an open question in
its own right; `progress.md` §E.1's "a fifth disposition may exist" covers
coverage, not this tie-break ambiguity.

**D2 — no negative control on the invariant-claim conjunct; an all-matching mutant passes the whole suite — `acceptance.md`:§C — Severity: critical — Class: blocking**

REQ-MRG-001 is a conjunction: moving-ref token **AND** git-command context
**AND** invariant-claim marker **AND** no SHA/baseline-variable. Three of the
four conjuncts have a negative control (AC-MRG-002 the SHA pin, AC-MRG-003(b)
the marker, AC-MRG-013's counter-mutation the R4 form). The **invariant-claim
conjunct has none** — no criterion supplies a line that names a moving ref
inside a git command *without* a claim marker and asserts it is NOT flagged.

A mutant that drops that conjunct — "flag any unpinned line containing
`origin/(main|develop|HEAD)` in a git-command context" — passes AC-MRG-001
(fires), its stated mutation (`origin/mainx` → 0), AC-MRG-002, -004, -005, -006,
-008, -009, and -013. Measured over `.moai/specs/**` at `43329ec8b`, excluding
this SPEC's own directory:

```
grep -rnE 'origin/(main|develop|HEAD)' .moai/specs --include='*.md' \
  | grep -v SPEC-MOVING-REF-GUARD-001 | grep -E 'git [a-z-]+' \
  | grep -cvE '\b[0-9a-f]{7,40}\b'
→ 495
```

495 findings against the ~42 the SPEC sizes its severity argument on — a
~12x over-fire that every criterion passes green. That over-fire is precisely
the bulk-suppression outcome §D.5 exists to prevent, so the mutant does not just
evade detection, it defeats the SPEC's central design decision. This is the
`verification-completeness.md` §2 mutant-probe defect ("an invalid-cases-only
rule passes an all-matching mutant"), and AC-MRG-013's counter-mutation does not
cover it — its fixture (`git diff --name-only origin/main -- internal/
(unchanged)`) carries a claim marker.

**Required fix:** add a MUST criterion — given a fixture line naming a moving ref
in a git-command context with **no** invariant-claim marker (e.g. a plain
`git fetch origin main` instruction), zero `MovingRefUnpinned` findings; mutation
that turns it red = drop the claim-marker conjunct from the rule.

**D3 — AC-MRG-005's `--strict` half is unsatisfiable on an era-grandfathered fixture — `acceptance.md`:77-86 — Severity: major — Class: blocking**

AC-MRG-005 asserts `moai spec lint --strict` exits non-zero on a corpus whose
only findings are `MovingRefUnpinned`. Read against the engine, that holds only
if the fixture SPEC directories classify as V3R6:

- `internal/spec/lint.go:213` — `demote := isGrandfatheredSpecDir(filepath.Dir(doc.Path)) || terminalStatusEnum[...]`
- `applyEraDemotion` (`lint.go:278-283`) — on a demoted doc, `case f.Severity == SeverityWarning: f.Advisory = true`
- `Report.HasErrors` (`lint.go:61`) — `if r.Strict && f.Severity == SeverityWarning && !f.Advisory`

`acceptance.md` §A specifies the fixtures only as "minimal, schema-valid SPEC
directories." A minimal fixture with no `progress.md` classifies V2.x under
era heuristic H-1; one with a `progress.md` lacking `§E.2`/`§E.4` markers
classifies V3R2-R4 under H-2 — both grandfathered, both demoted to `Advisory`,
and `--strict` then does **not** escalate. The criterion's PASS half fails for a
reason that has nothing to do with the rule under test. (The stated mutation is
unaffected: `SeverityError` is not in `eraDemotableCodes`, so it survives
demotion and still turns the criterion red.)

**Required fix:** constrain the fixture in §A — each fixture SPEC directory
carries `era: V3R6` in frontmatter (H-override), or a `progress.md` with `§E.2` +
`§E.4` + a `sync_commit_sha` value. State it as a fixture precondition so the
run-phase implementer cannot satisfy §A while failing AC-MRG-005.

**D4 — AC-MRG-010's decider does not detect a committed corpus edit — `acceptance.md`:147-150 — Severity: major — Class: blocking**

The [HARD] M4 clause is "M4 edits no SPEC artifact." Its enforcement is
`git status --short -- .moai/specs`, which reads the **working tree only**. M4
runs the rule over the corpus and writes a triage report; if it also pins a
finding in another SPEC and commits that edit — the exact failure mode the clause
names — `git status` is clean and the criterion passes green. The stated mutation
("pin a single occurrence in any other SPEC directory") only turns it red while
the edit stays uncommitted.

This also contradicts `acceptance.md` §B, which declares that all criteria are
decided against the frozen `BASELINE_SHA` "never against `origin/develop`
directly." AC-MRG-010 is decided against neither — it is the one criterion with
no baseline at all. The stronger form exists in the corpus this SPEC already
read: `SPEC-DESIGN-MOAIWEBV2-002` AC-MWA-007a asserts **both**
`git diff --name-only "$BASELINE_SHA"..HEAD -- <path>` **and**
`git status --porcelain <path>`, for this precise reason ("per-milestone pushes
must not vacuously empty the diff; committed edits must be caught").

**Required fix:** decide AC-MRG-010 on both surfaces —
`git diff --name-only "$BASELINE_SHA"..HEAD -- .moai/specs | grep -v SPEC-MOVING-REF-GUARD-001`
empty **and** `git status --short -- .moai/specs` empty — and restate the mutation
as a *committed* pin.

**D5 — Q0 is load-bearing for M3 but carries no DoD disposition — `acceptance.md`:221 / `spec.md` §H — Severity: major — Class: optional**

§H Q0 ("R4 form recognition … the run-phase decision most likely to need operator
input") was added with the addendum. The DoD still reads "Open questions
**Q1-Q3** … each carry a recorded disposition." Q0 is the one question M3 cannot
be implemented without answering — AC-MRG-013 itself defers to it ("does not
assert the signature exists yet") — and it is the only one exempt from the
disposition requirement.

**Required fix:** change the DoD line to `Q0-Q3`.

**D6 — frontmatter `version` not bumped with the v0.2.0 HISTORY row — `spec.md`:4 vs :25 — Severity: major — Class: blocking**

`version: "0.1.0"` while HISTORY carries a `| 0.2.0 | 2026-08-28 | Lead addendum
absorbed…` row describing five substantive changes. `updated:` is also unchanged.
A consumer reading the frontmatter gets a version that describes a document
that no longer exists — the documented-state-vs-actual-state divergence this SPEC
is written to prohibit, in its own frontmatter.

**Required fix:** `version: "0.2.0"`; confirm `updated:` reflects the addendum date.

**D7 — plan.md carries two pre-addendum citations — `plan.md`:24, :137 — Severity: minor — Class: blocking**

The addendum swept plan.md §F M1 and M3 but left two lines behind:

- `:24` — "The corpus contains all **three** predicate dispositions". The taxonomy
  is now two classes with an S1/S2 split and four remedies (§D.1, §D.2).
- `:137` — §G anti-patterns: "Guarded by … REQ-MRG-004 (the message names **three**
  branches)". REQ-MRG-004 now requires four, and this is the anti-pattern entry for
  the card's *dominant* failure mode — the one place a run-phase reader looks to
  learn what guards bulk-pinning.

This is the `verification-completeness.md` §3 cross-layer revision sweep defect:
the revision did not end in the file it started in.

**Required fix:** update both lines to the two-class/four-remedy taxonomy.

**D8 — `Where` used as a data conditional rather than a capability gate — `spec.md`: REQ-MRG-002/003/008/009/010 — Severity: minor — Class: optional**

GEARS reframes `Where` as capability gate / feature flag / static config. Five
requirements use it as a plain input-content conditional ("Where a flagged line …
carries a marker"), which is semantically `When` (event-driven) or `While`
(state-driven). This is a legacy-EARS Optional reading, valid inside the
backward-compat window through 2026-11-22, so MP-2 passes — but a GEARS-canonical
authoring would use `When`/`While`.

**Required fix (optional):** relabel to `When`/`While`. No behavioural change.

**D9 — AC-MRG-010 is an orphan acceptance criterion — `acceptance.md`:138 — Severity: minor — Class: optional**

All 10 REQs have at least one AC and all 12 other ACs trace to a REQ.
AC-MRG-010 traces to none — it enforces plan.md M4's [HARD] clause, which was
never lifted into §E. The [HARD] constraint that most directly prevents the
card's dominant failure mode is therefore the only one with no requirement behind
it.

**Required fix (optional):** add a REQ stating that corpus triage classifies
without modifying any SPEC artifact outside this SPEC's own directory, and point
AC-MRG-010 at it.

**D10 — corpus filters 2-4 name no reproducible command — `spec.md`:107-110 / `progress.md` §E.1 — Severity: minor — Class: optional**

Filter 1 is stated verbatim and reproduces exactly: I re-ran it at the pinned rev
and got 1477 (`git grep -nE 'origin/(main|develop|HEAD)' ec15ec2cd --
'.moai/specs/**/*.md' | wc -l` → 1477; the working-tree run excluding this SPEC's
own 26 lines also → 1477). Filters 2-4 are described rather than quoted — "piped
through the claim-marker alternation", "`grep -E 'git [a-z-]+[^\`]*origin/…'`"
with an ellipsis — so 527 / 53 / 42 cannot be reproduced. My own reconstruction
gave 597 / 79 / 53: same order of magnitude, different alternation.

Under `verification-claim-integrity.md` §2 an attributed claim names **the
command**; these name a description of one. `progress.md` already flags the right
Gap (grep prevalence, not rule output) but not this one. It matters because 42 is
the sole quantitative support for §D.5's severity decision.

**Required fix:** record the three alternations verbatim in `progress.md` §E.1.

## What I verified and found sound

Stated so the findings above are not read as the whole picture.

- **The R2 precedent is real, not fabricated.** `SPEC-DESIGN-MOAIWEBV2-002`
  `plan.md:36` records `BASELINE_SHA=$(git rev-parse origin/main)` captured before
  the first run-phase commit, and `acceptance.md:14` decides its criteria against
  `$BASELINE_SHA`, with AC-MWA-007a/013 consuming it. Verified verbatim.
- **Instance 2's citation is verbatim.** The quoted
  `SPEC-GRAPH-FRESHNESS-CADENCE-001` v0.2.2 HISTORY text — "the *subject* of audit
  finding N2 rather than addresses into the tree, and refreshing them would erase
  the correction N2 records" — matches the source exactly, and `44095ddc2` is
  indeed the merge that landed t322.
- **The predicate classifies all four corpus instances correctly.** I applied it
  independently: `AC-TST-010` and `AC-AFS-012` (`git diff origin/main …` deciding
  empty/byte-identical) → ANCHOR, substitution preserves meaning. `AC-COORD-016`
  (asserts the literal string `git rev-list --count --left-right
  origin/main...HEAD` is preserved verbatim in a document) → SUBJECT/S1;
  substituting a SHA changes the string being asserted about. `REQ-LB-006`
  (documents `git reset --hard origin/main` as doctrine) → SUBJECT/S1; a pinned
  SHA makes the doctrine wrong. All four land where the SPEC says. All four lines
  read verbatim at `43329ec8b`.
- **The §D.3 skew is named rather than hidden.** "No grounded instance is ANCHOR"
  plus the argument for *why* the collected set is skewed (subject-class defects
  destroy information and become memorable; anchor-class ones are quietly correct
  to pin) is the honest form of a small-validation-set admission. The stated
  residual risk is honest scoping, not an untested predicate dressed as doctrine —
  with the D1 exception, which is a defect in the procedure rather than in the
  validation.
- **The three-dot disproof reproduces.** `git merge-base ec15ec2cd 44095ddc2` →
  `44095ddc2cc1c9fed2b3bd5ac946f48017988aba`, confirming the degeneration
  mechanism §B.2 describes.
- **§F limits are not contradicted by any criterion.** L1 is carried forward
  unresolved and Q3 keeps it open — it is not quietly closed. AC-MRG-001..009 and
  -013 all use `origin/` forms, so none implies L1 coverage. AC-MRG-010 scans
  `.moai/specs` only, consistent with L5. L6 is *created by* the R4 exemption and
  says so; REQ-MRG-010/AC-MRG-013 assert the exemption, not freshness detection —
  consistent, not a coverage claim. AC-MRG-011's `grep -c 'L[1-6] —' → 6` plus the
  manual §C read is a sound decider.
- **The §D.6 side-discipline split reasoning is correct.** A dispatch message is
  not a file the linter reads, so a detection criterion over it genuinely cannot
  fail — that is the right reason, not a convenient one. One consequence is left
  unowned though: §B.5 records a concrete dispatch-format change as an attributed
  decision, and no requirement, criterion, or milestone carries it anywhere. Worth
  deciding deliberately rather than by omission; not raised as a separate defect
  since §B.5 marks it a decision rather than a deliverable.
- **Severity `warning` leaves the guard real teeth.** Verified against the engine:
  `--strict` exists (`internal/cli/spec_lint.go`), and `Report.HasErrors`
  escalates a non-advisory warning under `Strict`, with `spec lint` returning
  exit 1. So `.github/workflows/spec-lint.yml` can gate on it when wanted while
  the default run stays green. The teeth are real — subject to D3.
- **M3's mechanism is feasible as specified.** `SPECDoc.Path` exists
  (`lint.go:433`), so `filepath.Dir(doc.Path)` sibling reading works; `Rule` and
  the `l.rules` slice are as plan.md §H describes. The `HaikuResidualRule`
  precedent is accurately characterised as "reaches past the single document"
  (it scans via `CheckAll` with a project `baseDir` rather than a per-doc sibling
  dir — a different mechanism, but the claim as worded is true).
- **AC-MRG-009 is well-designed.** Its mutation ("restrict the rule to
  `doc.Body`; this fails while AC-MRG-001 still passes") is exactly why it is a
  separate criterion, and it is correct — `doc.Body` carries `spec.md` only.
- **AC-MRG-013 carries a counter-mutation.** Requiring that the R4 exclusion not
  become a blanket bypass, and the DoD refusing an R4 exclusion accepted on
  AC-MRG-013 alone, is the mutant-probe discipline applied without being asked.
  D2 is the one conjunct that did not get this treatment.
- **Tier M is correct.** One new Go rule file + one test file + fixtures + one
  doctrine section + one template mirror ≈ 5-15 files, well under 1000 LOC.
  10 REQ / 13 AC sit inside the Tier M 16/16 budget.

## Process finding

**P1 — the SPEC was rewritten and committed while under audit — Severity: major — Class: blocking (process, not artifact)**

Four writes to three artifacts between 08:03 and 08:11, then a commit
(`b3e25945f` → `43329ec8b`), all inside the audit window. This violates
`agent-common-protocol.md` § Background Agent Execution: *"While a worktree is
being actively audited, it has exactly one writer… Observing an unexpected HEAD
move or a foreign commit on an actively audited worktree is a process defect:
report it to the lead immediately and record it in the progress record — never
continue quietly."*

The concrete cost here was a near miss rather than an actual one: a verdict
written from the 08:08 snapshot would have reported three fabricated defects
(`REQ-MRG-010`, `L6`, `AC-MRG-013` "claimed by HISTORY but absent") that were
mid-write state. That is this card's own subject — a measurement whose validity
expired being served as current — occurring for the third time in its own
lifecycle (§B.4 the dispatch, §B.5 the format change, and now the audit).

**Required fix:** re-audit iterations on this card wait for a committed, quiescent
tree, and the lead confirms the writer has stopped before dispatching the audit.

## Recommendation

The SPEC is genuinely strong: the predicate is argued rather than asserted, the
addendum improved it under review (instance 3's reclassification is the correct
call and records its own prior error), the R2 precedent and all four corpus
citations are real and verified, the falsifiability contract is above the
standard of this lane, and the detection limits are stated honestly including one
the SPEC's own exemption creates. No must-pass criterion failed and the aggregate
clears the Tier M threshold. It is not a FAIL.

It is also not a clean PASS. Four blocking defects sit in the surfaces the card
marked [HARD], and two of them cannot be carried as debt because the milestones
that encode them come first:

1. **D1 before M1.** M1 authors the doctrine text. Shipping the predicate with a
   tie-break that routes its founding case away from R4 publishes a procedure
   that cannot produce the remedy the same document invented — and M1 is the
   cheapest possible place to fix it (plan.md §F says so itself).
2. **D2 before M3.** M3 writes the rule. Without a claim-marker negative control,
   an implementation that over-fires ~12x passes the entire suite green, and that
   over-fire defeats §D.5's severity argument — the SPEC's central design
   decision.
3. **D3 and D4** are one-line criterion repairs; make them in the same pass.
4. **D6 and D7** are mechanical (version bump, two stale citations).
5. **D5, D8, D9, D10** are optional — surfaced for the orchestrator's discretion,
   not routed into a revision by this report.

Suggested route: fix D1-D4, D6, D7 as a targeted edit rather than a plan-phase
re-entry, then run a scoped iteration-2 re-audit over that enumerated delta only
(Tier M ceiling permits exactly one more iteration). Wait for a quiescent
committed tree first, per P1.

---

# Addendum — targeted re-audit at `43329ec8b` (lead follow-up)

Verified independently before citing anything below: `git log --oneline -2` →
`43329ec8b` (v0.2.0 addendum) with parent `b3e25945f` (the dispatched SHA). The
tree is clean apart from this report's own untracked directory, and both audited
artifacts hash identically to the verdict above (`b31b993ef51b…`,
`ceec0a435b2b…`) — so the body of this report already judged `43329ec8b`, and
nothing in it needs re-basing.

Counts verified mechanically rather than accepted from the tally:
`grep -c '^\*\*Test [0-9]'` → 4 · `grep -c '^| \*\*R[0-9]\*\*'` → 4 ·
`grep -c '^- \*\*REQ-MRG-'` → 10 · `grep -c '^- \*\*L[0-9]'` → 6 ·
`grep -c '^### AC-MRG-'` → 13 · `grep -c '^\*\*Instance '` → 5. All match.

**Revised score: 0.80** (was 0.82) — exactly at the Tier M threshold. Clarity
0.80→0.78 (D12), Testability 0.78→0.72 (D11). Verdict unchanged at
**PASS-WITH-DEBT**, but the margin is now zero: one further blocking finding
takes it under. Read the verdict as "fix the blocking list before M1/M3", not as
"proceed".

## D11 — an R4 exclusion keyed on the fetch verb passes all 13 criteria, the counter-mutation included, and silences a third of the true-positive class — `acceptance.md`:AC-MRG-013 — Severity: critical — Class: blocking

You asked me to try to write an R4 exclusion that passes AC-MRG-013 and still
switches the guard off. I could not do it against AC-MRG-001 — and that is a
credit to the counter-mutation, reported here as a negative result. But I could
do it against the rest of the suite.

**First, what the counter-mutation does successfully block.** Its fixture
`git diff --name-only origin/main -- internal/ (unchanged)` differs from
AC-MRG-001's fixture only in table-cell wrapping. Any exclusion loose enough to
exempt one exempts the other, so an exclusion that keeps the counter-mutation red
necessarily keeps AC-MRG-001 red. The positional form I tried first — "a git verb
before the ref and a parenthesized value after it" — exempts the counter-mutation
fixture and is caught immediately. The counter-mutation is doing real work.

**The bypass it does not block.** Key the exclusion on the fetch verb — the most
salient token in R4's *only* exemplar (`measure at entry with git fetch origin
develop (dispatch-time reference value: <sha>)`), and the natural thing for a
run-phase implementer to reach for while Q0 leaves the signature ungeneralized:

| Criterion | Its fixture | Under a fetch-verb exclusion |
|---|---|---|
| AC-MRG-001 | `git diff --name-only origin/main -- internal/` | no fetch verb → still flagged → **passes** |
| AC-MRG-006 | `git rev-list --count --left-right origin/main...HEAD → 0 0` | no fetch verb → still flagged → **passes** |
| AC-MRG-013 fixture | `… git fetch origin develop (dispatch-time reference value: …)` | exempt → **passes** |
| AC-MRG-013 counter-mutation | `git diff --name-only origin/main -- internal/ (unchanged)` | no fetch verb → still flagged → **passes** |

Every criterion green. Now the corpus reach, measured at `43329ec8b` over
`.moai/specs/**` excluding this SPEC's own directory:

```
# candidate lines (moving ref + git-command context + claim marker, no SHA)
… | grep -cvE '\b[0-9a-f]{7,40}\b'                → 100
# of those, lines also carrying the fetch verb (exempted by the bypass)
… | grep -c 'git fetch'                           → 36
# corpus lines pairing the fetch verb with 'rev-list --count --left-right'
grep -rn 'rev-list --count --left-right' … | grep -c 'git fetch'  → 81
```

**36 of 100** candidates silenced, and **81** corpus lines pair the fetch verb
with the divergence form — the REQ-MRG-006 class, whose own criterion still
passes because its fixture omits the fetch. The pairing is not incidental: the
Pre-Spawn / Pre-Edit Sync Check block in `agent-common-protocol.md` is a fetch
followed by `git rev-list --count --left-right origin/main...HEAD`, so every
progress record quoting that block carries both tokens on adjacent or identical
lines. The bypass lands squarely on the most common real instance of the defect.

So the answer to the question as posed: **the counter-mutation protects
AC-MRG-001's shape and nothing else.** The hazard the author flagged is real, the
guard written for it is correctly aimed, and it is one fixture too narrow.

**Required fix:** extend AC-MRG-013's counter-mutation set with a second fixture
carrying a command token *and* a claim — a fetch chained to a
`rev-list --count --left-right` reading `0 0` — which MUST still be flagged. And
state in Q0 that the R4 signature must key on the *imperative structure* (an
instruction to measure, with the value syntactically demoted) rather than on any
command token, because a command-token key is forgeable by construction.

## D12 — §D.4's incentive argument was never restated for four remedies; R4 is an unpriced exemption — `spec.md`:§D.4 vs §D.2 — Severity: major — Class: blocking

`grep` over §D.4 returns no mention of R4 or of four remedies. Its argument is
still the v0.1.0 two-way one: *"A bare marker would make 'silence the warning'
cheaper than 'pin the SHA' … With a reason required, declaring and pinning cost
about the same, and the author picks on the merits."*

That pricing was computed against a two-door choice — marker-with-reason, or pin.
R4 opens a third door, and it carries **no reason requirement at all**: an author
who wants a line to stop being flagged can rephrase it into R4 shape and is done.
Whatever signature Q0 settles on, satisfying a shape is cheaper than writing a
justification. §D.4's own criterion — "declaring and pinning cost about the
same" — is therefore false as of v0.2.0, in the section that states it.

This is the incentive half of the residual L6 half-states. L6 accepts that the
detector cannot judge a reference value's *freshness*, which is sound on its own
terms and correctly reasoned (flagging the recommended form would teach readers
to avoid it — that acceptance I agree with). What L6 does not say is that the same
shape-blindness makes R4 the cheapest available silencer. D11 is that gap
exploited by an implementer; D12 is the same gap exploited by an author.

**Required fix:** restate §D.4's incentive paragraph against four remedies, and
decide deliberately whether R4 needs a cost — the natural one being that the R4
form must name the command a reader is to run, which is already R4's definition
and is not free to fake convincingly. Extend L6 to state the incentive residual
alongside the freshness residual.

## Answers to the four judgment calls

**1. Is the instance-3 reclassification honest about its limits, or under-specified?**
Both, and they are separable. The *recording* is honest — §D.3 states the v0.1.0
verdict was ANCHOR → R2, calls it "wrong in a way worth recording", and gives the
sharp reason (an anchor's value must be fixed; an S2 claim's must not be —
opposite requirements on the same property). §D.1.1's refusal to promote S2 to a
third top-level class is also correct: S1 and S2 genuinely answer the top-level
question the same way. That reasoning holds.

The *procedure* is under-specified, and D1 is the proof: the worked example was
corrected without correcting the tests that produced the wrong answer. Test 1
still says "substitute the SHA it resolves to right now … still says what it
meant → ANCHOR"; §D.3 has to smuggle in "but the substitution decays immediately"
to reach SUBJECT, and the tie-break ("Test 1 governs") actively routes the case
back to ANCHOR. A predicate that misclassified one of its own five grounded
instances is not disqualified by that — first applications are where procedures
get debugged — but it is disqualified from shipping while the debugging stopped
at the example.

**2. Does L6's acceptance hollow out R4?** Not on freshness — that acceptance is
sound, and the stated mitigation (command first, so a reader who follows the line
re-measures regardless of a stale value) is real. It hollows out R4 on
*incentives*, which L6 does not mention: see D12.

**3. All five instances SUBJECT, none ANCHOR — sharp observation or validation
gap?** Both, and the SPEC presents only the first. The selection-bias argument is
correct and well-put: subject-class defects destroy information and become
memorable, anchor-class ones are quietly correct to pin and generate no incident,
so a set built from escalated cases is structurally skewed toward the exemption
class. I agree with the observation.

What §D.3 does not draw is the consequence. It calls the skew "the strongest
available argument for shipping the predicate" — but the same fact says the
ANCHOR half rests entirely on seven corpus lines the author classified by reading,
with no instance that was independently escalated and adjudicated. The SUBJECT
half has five adjudicated cases; the ANCHOR half has zero. So yes: **the ANCHOR
branch has not been tested in the sense the SUBJECT branch has.**

And the two halves of this report converge on it. D1 shows the specific way the
ANCHOR branch misfires — Test 1 over-returns ANCHOR, which is exactly what an
unvalidated branch would be expected to do. The skew and D1 are one fact seen
from two directions: the ANCHOR branch is both the unvalidated one and the one
that demonstrably misclassifies. §D.3 should state the skew as a limitation as
well as a strength, and that statement belongs in the doctrine M1 publishes.

**4. Was the dispatch a fifth grounded case?** Verified, and yes — but the SPEC
already carries that occurrence as instances 3 and 5, and adding the audit
dispatch as a sixth would be padding rather than evidence. The audit-window
instance is recorded as P1 in the body above, where it belongs: it is a process
violation of the one-writer rule, not a new shape of the defect. Its value is
that the defect recurred a third time in this card's own lifecycle while the card
was being written to prevent it — worth one sentence in §B, not a new instance
slot.

## Revised recommendation

The blocking list is now six: **D1** (predicate cannot reach R4 for its founding
case) and **D12** (R4 unpriced) before **M1**, which publishes the doctrine text;
**D2** (no claim-marker negative control) and **D11** (fetch-verb bypass) before
**M3**, which writes the rule; **D3** and **D4** are one-line criterion repairs;
**D6** and **D7** are mechanical. D5, D8, D9, D10 remain optional.

D2 and D11 are the same defect class arriving from opposite directions — an
over-broad rule that over-fires ~12x, and an over-broad exclusion that under-fires
by a third — and neither is caught by the current suite. Both close with one
fixture each. That is the right work to do before M3, and it is substantially
cheaper than discovering either after the rule ships at `warning` severity, where
an under-firing guard reads exactly like a clean corpus.
