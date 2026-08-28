# SPEC-MOVING-REF-GUARD-001 — Progress

Card: t342 · Tier M · Worktree `.claude/worktrees/t342`, branch `WT-moving-ref-guard`.

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored 2026-08-28 in worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t342` at HEAD `ec15ec2cd`, working tree clean at
authoring time.

**Measurements taken during authoring** (each re-run in this tree, not carried over):

| Figure | Command | Output |
|---|---|---|
| Tree identity | `git rev-parse --show-toplevel` | `…/.claude/worktrees/t342` |
| Branch | `git branch --show-current` | `WT-moving-ref-guard` |
| HEAD | `git rev-parse --short HEAD` | `ec15ec2cd` |
| Branch provenance | `git reflog show WT-moving-ref-guard` | created from `origin/develop` at `ec15ec2cd` |
| Merge-base degeneration | `git merge-base ec15ec2cd 44095ddc2` | `44095ddc2cc1c9fed2b3bd5ac946f48017988aba` |
| Two-dot, stale anchor | `git diff --stat 44095ddc2 -- internal/hook` | `3 files changed, 288 insertions(+)` |
| Three-dot, stale anchor | `git diff --stat 44095ddc2...HEAD -- internal/hook` | `3 files changed, 288 insertions(+)` |
| Two-dot, current anchor | `git diff --stat ec15ec2cd -- internal/hook` | *(empty)* |
| Corpus filter 1 | moving-ref mentions in `.moai/specs/**/*.md` | `1477` |
| Corpus filter 2 | + git-command context | `527` |
| Corpus filter 3 | + invariant-claim marker | `53` |
| Corpus filter 4 | − lines carrying a 7-40 hex SHA | `42` |
| SPEC ID validity | Bash ERE check against `internal/spec/lint.go` `specIDPattern` | `PASS` |
| Parent of the created tip (v0.2.0, re-verified) | `git rev-parse ec15ec2cd^` | `44095ddc2cc1c9fed2b3bd5ac946f48017988aba` |

The filter-4 set of 42 was enumerated in full and read, not sampled. Its three-way classification is
recorded in `spec.md` §B.3.

**Filter commands, verbatim (audit D10).** v0.1.0 and v0.2.0 described filters 2-4 rather than
quoting them, so the figures could not be reproduced — the iter-1 auditor's reconstruction gave
597 / 79 / 53 against the recorded 527 / 53 / 42, same order of magnitude, different alternation.
Under `verification-claim-integrity.md` §2 an attributed claim names *the command*, not a
description of one. The three alternations are therefore recorded here, and re-run at `43329ec8b`
with this SPEC's own directory excluded, where they reproduce the original figures exactly:

```bash
# Filter 2 — git-command context (→ 527)
grep -rnE 'git [a-z-]+[^`]*origin/(main|develop|HEAD)' .moai/specs --include='*.md' \
  | grep -vc SPEC-MOVING-REF-GUARD-001

# Filter 3 — + invariant-claim marker (→ 53)
… | grep -v SPEC-MOVING-REF-GUARD-001 \
  | grep -ciE 'byte-unchanged|byte unchanged|unchanged|preserv|보존|no diff|empty|0 files|부재|absent|변경 ?없|그대로'

# Filter 4 — − SHA pin (→ 42)
… | grep -cvE '\b[0-9a-f]{7,40}\b'
```

The claim-marker alternation in filter 3 is the load-bearing one and is the piece the auditor could
not reconstruct: `byte-unchanged|byte unchanged|unchanged|preserv|보존|no diff|empty|0 files|부재|absent|변경 ?없|그대로`, applied case-insensitively.

**Explicitly NOT observed at plan-phase** (Gaps):

- The detector does not exist, so the 42-line candidate set is a *grep prevalence measurement*, not
  a rule output. The M4 triage count may differ once the real pattern is implemented, and no
  criterion here assumes the two agree.
- The corpus classification in `spec.md` §B.3 names seven anchor-class and two subject-class lines
  by identifier; the remaining candidates were read but are not individually classified in the
  plan-phase artifact. Full classification is M4's deliverable.
- No `go test` was run — plan-phase authored no Go code.
- The claim that three-dot is stable under *diverging* upstream advance (as opposed to the absorbed
  advance measured above) was reasoned, not measured; fabricating a diverging branch pair was
  judged out of proportion for a plan-phase figure. `spec.md` §B.2 states the mechanism and cites
  only the measurement actually taken.

**v0.2.0 additions — what was and was not observed for the lead's addendum:**

- **Observed.** Grounded instance 3's attribution was re-verified in this turn before being written
  into the SPEC: `git rev-parse ec15ec2cd^` → `44095ddc2…` and `git reflog show WT-moving-ref-guard`
  → created from `origin/develop` at `ec15ec2cd`. The lead's dispatched value was therefore exactly
  one integration stale, established independently of the dispatch that reported it.
- **NOT observed — the dispatch-format change (`spec.md` §B.5) is an attributed decision, not an
  observed practice.** It is recorded on the lead's authority and dated 2026-08-28. No dispatch
  other than the message announcing it has been seen in the new form, so the SPEC cites it as a
  design decision and a source of remedy R4 — never as evidence that the format is in use. Whether
  it holds is a run-phase observation, not a plan-phase one.
- **NOT observed — R4's recognizable signature.** REQ-MRG-010 requires the R4 form to be exempt, but
  the lead's line is a single instantiation rather than a grammar, and no attempt was made here to
  generalize it. `spec.md` §H Q0 carries it as the least-settled open question; AC-MRG-013 fixes the
  required *behaviour* without asserting the signature exists.

**v0.3.0 — plan audit iter-1 remediation, re-measured rather than carried.**

Verdict: PASS-WITH-DEBT 0.82, revised 0.80 after the targeted re-audit
(`.moai/reports/t342/plan-audit.md`, in this worktree — the isolation guard refused the primary
path, and the lead is handling recovery; this worktree's copy is left untouched per instruction).

Every load-bearing figure was re-measured here at `43329ec8b` before being written into the SPEC.
Two of the auditor's D11 figures did not reproduce, one reproduced exactly, and D2's did:

| Figure | Audit | Re-measured here | Verdict |
|---|---|---|---|
| D11 candidate lines (diff-shaped class) | 100 | **52** | did not reproduce — different alternation (D10) |
| D11 of those with the fetch verb | 36 | **6** | did not reproduce — same cause |
| D11 fetch ∧ `rev-list --count --left-right` corpus lines | 81 | **81** | reproduces exactly |
| D2 all-matching-mutant over-fire | 495 | **495** | reproduces exactly |

The two non-reproducing figures are D10's consequence and not a dispute about the finding. Measuring
the class the bypass actually lands on — unpinned divergence lines, the REQ-MRG-006 class — gives
**76 of 117 silenced (65%)**, which is *worse* than the auditor's estimate. D11 is therefore
confirmed and strengthened, and `spec.md` §B.6 records the re-measurement rather than the audit's
numbers.

Sample of the silenced class (`grep -rn 'rev-list --count --left-right' … | grep 'git fetch'`, no
SHA on the line):

```
SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001/plan.md:63
  `git fetch origin && git rev-list --count --left-right origin/main...HEAD` → expect `0 0`
SPEC-AGENT-PARALLEL-OPT-001/plan.md:92
  `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` | 병렬 세션 레이스 부재 확인
```

**NOT observed at v0.3.0 (Gaps):**

- No Go code exists, so AC-MRG-013 CM-2 and AC-MRG-014 are *specified* falsifiable, not *observed*
  falsifiable. Both mutations are run-phase obligations under the DoD; neither has been planted.
- The fetch-verb bypass was reasoned from fixture inspection, not executed against a built rule.
  The 76/117 reach is a corpus measurement of what *would* be silenced, not an observed silencing.
- Whether keying on imperative structure is *implementable* was not tested — §H Q0 keeps the
  recognition method open, and this revision constrains only what it must not key on.

**P1 — the tree moved under the auditor, and the recurrence is recorded rather than tidied away.**
The v0.2.0 addendum was written and committed (`b3e25945f` → `43329ec8b`) inside the audit window,
violating the one-writer rule in `agent-common-protocol.md` § Background Agent Execution. The
auditor caught it and re-verified against the settled tree; a verdict written from its 08:08
snapshot would have reported three fabricated defects from mid-write state. That is this card's own
subject — a measurement whose validity expired being served as current — occurring for the **third**
time in its own lifecycle, after §B.4 (the dispatch base line) and §B.5 (the format change). It is
recorded here as a process finding rather than as a sixth grounded instance, per the auditor's
judgment that it is a violation of the writer rule and not a new shape of the defect.

**v0.4.0 — plan audit iter-2 (PASS-WITH-DEBT 0.86, +0.06 monotonic), targeted pre-run edit.**

Iter-2 verified all eight iter-1 blocking defects closed and raised D13/D14 blocking, D15/D16
optional. Applied as a targeted edit rather than a plan-phase re-entry — iter-2 was the Tier M
ceiling.

**D13 — the old AC-MRG-013 fixture was vacuous, re-measured independently rather than accepted.**
The lead measured two of the three axes; all three were re-run here, plus the pipeline check that
settles it:

| Axis | Old fixture (lead's dispatch line) | New fixture |
|---|---|---|
| `grep -cE 'origin/[a-z]'` | **0** — `git fetch origin develop` is two args, no slash token (an L1 blind-spot line) | **1** |
| `grep -cE '\b[0-9a-f]{7,40}\b'` | **1** — REQ-MRG-008 already exempts it | **0** |
| claim marker (filter-3 alternation, `-i`) | **0** | **1** |
| git-command context (filter 2) | — | **1** |
| `grep -cE '\$[A-Z_]*BASELINE[A-Z_]*'` | — | **0** |
| **full REQ-MRG-001 pipeline (filters 2-4)** | **0** — unflaggable, so removing the R4 exclusion changed nothing | **1** |

The last row is the defect: with the old fixture the exemption could have been entirely absent and
every criterion would still have passed. The new fixture is
`` - verify `internal/hook` is unchanged by this work: run `git diff --name-only origin/develop -- internal/hook` at read time (reference reading 2026-08-28: empty) ``.

**Residual-scope verification (the thing the old fixture was hiding).** REQ-MRG-010's exclusion does
real work on exactly one class, established by measuring the two shapes that need no exclusion:

- *SHA-valued R4 reference* — the old fixture itself: `\b[0-9a-f]{7,40}\b` → **1**, so REQ-MRG-008
  exempts it without REQ-MRG-010.
- *Divergence-class R4 reference* — `- run `git rev-list --count --left-right origin/main...HEAD` at
  entry (reference reading 2026-08-28: 0 0)`: divergence class → **1**, carries a date → **1**, so
  REQ-MRG-006's date conjunct exempts it without REQ-MRG-010.
- *The residual* — the new fixture: divergence class → **0**, hex → **0**, so neither of the above
  applies and the R4 exclusion is the only thing that can exempt it.

**D14** — L7 added, and the mechanism matters more than the text: before this, `grep -n 'ANCHOR'
acceptance.md` returned nothing, so the ANCHOR-branch limitation lived only in §D.3 prose and M1
could have published the doctrine without it while passing every criterion. As L7 it is inside
AC-MRG-011's count (6→7) and its own `grep -c 'ANCHOR'` check.

**v0.5.0 — D13's second condition closed: R4's scope measured, not asserted.**

The v0.4.0 fixture swap was necessary and not sufficient; leaving REQ-MRG-010's reach
uncharacterized would have seated a different vacuous criterion in the same chair. Measured (§B.7):
**0 of 42** candidate lines fall in R4's class, on two independent probes — a narrow one keyed on
imperative measure-instruction shapes and a deliberately broader one keyed on any value-demotion
marker. R4 is one day old, so no corpus line is written in a form that did not exist when it was
written. Reported as empty rather than talked up: per the lead's instruction, a dead clause honestly
labelled beats a live-looking one nobody can exercise. §H Q0 now carries the scope decision, with
option C (defer the exclusion, use the R3 marker meanwhile) recommended but not taken.

**D13's provenance — the transferable part, recorded as such.** The vacuous fixture was the lead's
dispatch line, and it became a fixture *because* the lead had flagged that line as an instance of
the defect. The prose intuition was right; the mechanical question — can this card's own conjuncts
actually see this line? — was never asked of it. An instance handed down by a lead and adopted as a
fixture without verification becomes a vacuous control by construction, and the authority of the
source is precisely what suppresses the check. This is not a fixture-selection slip to be logged
anonymously; it is a named failure mode, and the guard against it is that **a fixture's properties
are measured on adoption, whatever its provenance**.

**The asymmetry D13 exposes, in this card's own terms.** CM-1 and CM-2 both test that the exclusion
is not too **wide**. Nothing tested that it works **at all**. Those are the two directions of a
single question, and this card exists to insist both are checked — §D.1's own predicate is built on
the symmetric claim that an unobserved success and an unobserved failure are both gaps. The suite
had one direction covered three times over and the other not at all, on the card's [HARD]
deliverable. The ordering note now in AC-MRG-013 (run the removal mutation *first*) is the operative
fix; this record is the reason.

**Count reconciliation — three figures, one pass, because the commands were written down.** The
auditor's `100 / 36`, my `117 / 76`, and the `129 / 81` pair reconcile to a single difference:
whether this SPEC's own directory is excluded, and at which revision. Re-measured at this revision:
divergence total `130` unexcluded vs `117` self-excluded; fetch ∧ divergence `86` unexcluded vs
`76` self-excluded-and-hex-filtered. The earlier `81` is now `86` because *this SPEC* added lines
quoting those commands. None of these is a correction of another; they are different measurements
that were mistaken for the same one while the commands stayed unwritten. That is D10's whole point,
and the reconciliation was only possible because v0.3.0 recorded the alternations verbatim.

**NOT observed at v0.5.0 (Gaps):**

- The two §B.7 probes are keyed on English imperative markers. An R4-form line written in Korean, or
  in a phrasing neither alternation anticipates, would not be counted — so "0 of 42" is a floor on
  emptiness, not a proof of it. Given R4's age the true count is very unlikely to exceed zero, but
  "unlikely" is not "measured" and the distinction is kept.
- Whether option C is *implementable* — whether the R3 marker reads naturally on an R4 line — was
  not tested. It is a recommendation, not a validated design.

**NOT observed at v0.4.0 (Gaps):**

- Still no Go code, so every fixture property above is a measurement of the fixture *text* against
  the specified filters — not an observation of a built rule flagging or exempting it. The claim
  "removing the R4 exclusion makes the finding reappear" is a *prediction* from the pipeline
  measurement, and becomes evidence only when M3 plants the mutation.
- The residual-scope characterization was verified against three constructed shapes, not against the
  live corpus. Whether R4-form lines of the residual class actually occur in `.moai/specs/**` was not
  measured — the form is new, so the expected count is low, but "low" was not established.

**Residual risk:** the exemption predicate (`spec.md` §D.1) is a judgment procedure. Its four tests
were validated against five real instances, all of which it classifies correctly — but five is a
small validation set, and **all five are SUBJECT-class** (`spec.md` §D.3 tally). The ANCHOR side of
the predicate rests entirely on the corpus lines of §B.3, which were classified by reading rather
than by applying the tests in anger. A disposition outside the anchor/subject dichotomy may still
exist; the v0.1.0 → v0.2.0 revision is itself evidence that the classification can be incomplete —
instance 3 was classified ANCHOR at v0.1.0 and that reading was wrong. `spec.md` §H carries the open
questions rather than closing them.

The v0.3.0 audit sharpened this: the SUBJECT branch has five adjudicated instances, the **ANCHOR
branch has zero** — it rests on seven corpus lines classified by the author alone. D1 (Test 1
over-returning ANCHOR for want of an evaluation time) is exactly the failure an unvalidated branch
would be expected to have, and it was found by an auditor rather than by the author. The skew and D1
are one fact from two directions, now stated as a limitation in `spec.md` §D.3 and not only as a
strength.

**Status transition:** `(none) → draft`, emitted across all four plan-phase artifacts.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
