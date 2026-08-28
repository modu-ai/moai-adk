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

### M1-M2 — the doctrine section and the exemption marker surface (2026-08-28)

**Baseline (frozen at pre-flight, per plan.md §C step 4 — R2, not a moving ref).**
`BASELINE_SHA = d566ecc7511e1954e3aeb1dff3a60afa5be1089b`. This SPEC's own PRESERVE evidence cites
that literal SHA; it does not cite `origin/develop`, because doing so would make this SPEC an
instance of its own defect. The value was frozen and supplied at dispatch; it was **not** re-resolved
in this run — that omission is recorded as a Gap below rather than papered over.

Worktree HEAD every command below ran against: `a1b8439c36f51f191291cc6db9565cc5b2ec2e25`
(`a1b8439c3`), branch `WT-moving-ref-guard`, in `.claude/worktrees/t342`. Doctrine file under test
throughout: `.claude/rules/moai/core/verification-claim-integrity.md`.

---

#### Claim 1 — the doctrine section exists and carries the predicate (AC-MRG-007)

**Claim.** `### 2.1 Moving-ref attribution — the anchor-or-subject predicate` was added to
`.claude/rules/moai/core/verification-claim-integrity.md`, immediately below §2 (the
baseline-attribution invariant it specializes) and above §3, carrying all four predicate tests, all
four remediation branches R1-R4, all five grounded instances, and the two-separate-steps statement.

**Evidence** (four greps, run against `a1b8439c3`):

```
$ grep -oc 'Test [1-4] — [A-Z][a-z-]*' .claude/rules/moai/core/verification-claim-integrity.md
4
$ grep -oc '\*\*R[1-4]\*\*' .claude/rules/moai/core/verification-claim-integrity.md
4
$ grep -c 'Instance [1-5] —' .claude/rules/moai/core/verification-claim-integrity.md
5
$ grep -c 'two classes and four remedies' .claude/rules/moai/core/verification-claim-integrity.md
1
```

Per-label enumeration, so the counts are not read as four unnamed matches:

```
$ grep -on 'Test [1-4] — [A-Z][a-z-]*' .claude/rules/moai/core/verification-claim-integrity.md
58:Test 1 — Substitution
63:Test 2 — Falsification
68:Test 3 — Re-measurement
72:Test 4 — Read-time
$ grep -o '\*\*R[1-4]\*\*' .claude/rules/moai/core/verification-claim-integrity.md | sort | uniq -c
   1 **R1**
   1 **R2**
   1 **R3**
   1 **R4**
```

**Baseline-attribution.** Commands run in this turn, against this tree, at HEAD `a1b8439c3`, on the
tree frozen from `BASELINE_SHA d566ecc7511e1954e3aeb1dff3a60afa5be1089b`.

**Note on the deciding grep's form.** A looser `grep -c 'Test [1-4] — '` returns **4 on the
Test-2-deleted mutant** as well, because the tie-break paragraph contains the string
`Test 4 — do not resolve to ANCHOR`. The name-bearing form above
(`grep -oc 'Test [1-4] — [A-Z][a-z-]*'`) is the one that actually discriminates, and it is the form
recorded here. The looser form is a criterion that would have passed its own mutation.

---

#### Claim 2 — the criterion is falsifiable: two mutations planted, observed red, reverted

**Mutation A1 — delete the falsification-source test (Test 2).**

```
$ grep -oc 'Test [1-4] — [A-Z][a-z-]*' .claude/rules/moai/core/verification-claim-integrity.md
3
$ grep -c 'Falsification source' .claude/rules/moai/core/verification-claim-integrity.md
0
$ grep -on 'Test [1-4] — [A-Z][a-z-]*' .claude/rules/moai/core/verification-claim-integrity.md
58:Test 1 — Substitution
63:Test 3 — Re-measurement
67:Test 4 — Read-time
```

4 → 3. RED. Reverted; the clean count returns to 4 (Claim 1).

**Mutation A2 — delete Test 4 (read-time action).**

```
$ grep -oc 'Test [1-4] — [A-Z][a-z-]*' .claude/rules/moai/core/verification-claim-integrity.md
3
$ grep -on 'Test [1-4] — [A-Z][a-z-]*' .claude/rules/moai/core/verification-claim-integrity.md
58:Test 1 — Substitution
63:Test 2 — Falsification
68:Test 3 — Re-measurement
```

4 → 3. RED. Reverted.

---

#### Claim 3 — all seven detection limits are stated (AC-MRG-011)

**Claim.** L1 through L7 of `spec.md` §F are stated in the doctrine section, each in the
`L<N> — <statement>` form the criterion greps for, with L7 naming the ANCHOR-branch validation gap.

**Evidence:**

```
$ grep -c 'L[1-7] —' .claude/rules/moai/core/verification-claim-integrity.md
7
$ grep -c 'ANCHOR' .claude/rules/moai/core/verification-claim-integrity.md
10
```

**Baseline-attribution.** This run, this tree, HEAD `a1b8439c3`.

---

#### Claim 4 — three limit mutations planted, observed, reverted; the third is only half-falsifying

**Mutation B1 — delete the L1 paragraph** (the limit AC-MRG-011 names as most likely to be quietly
omitted):

```
$ grep -c 'L[1-7] —' .claude/rules/moai/core/verification-claim-integrity.md
6
```

7 → 6. RED. Reverted. **Recorded correction to the criterion's own text:** AC-MRG-011 states this
mutation drops the count "to 5". The observed value is **6** — the criterion's expected figure was
written when the limit set was smaller (the count was raised 5 → 6 at v0.2.0 and 6 → 7 at v0.4.0)
and its mutation text was not swept with it. The criterion still goes red, so the falsifiability
holds; the stated number does not.

**Mutation B2 — delete L6** (the limit the R4 exemption creates):

```
$ grep -c 'L[1-7] —' .claude/rules/moai/core/verification-claim-integrity.md
6
```

7 → 6. RED. Reverted.

**Mutation B3 — delete L7** (the ANCHOR-branch validation gap). AC-MRG-011 requires **both** halves
to fail. Observed:

```
$ grep -c 'L[1-7] —' .claude/rules/moai/core/verification-claim-integrity.md
6
$ grep -c 'ANCHOR' .claude/rules/moai/core/verification-claim-integrity.md
9
```

The count half goes red (7 → 6) as specified. **The `ANCHOR` half does not**: the criterion expects
`grep -c 'ANCHOR'` to return 0, and it returns 9. The reason is structural rather than an authoring
slip — `ANCHOR` is one of the predicate's two **class names**, so it necessarily appears throughout
the tests, the tie-break, the remedy table, and the instances. Against a doctrine that publishes the
predicate at all, `grep -c 'ANCHOR' ≥ 1` is satisfied whether or not L7 exists, and therefore guards
nothing. The criterion's `≥ 1` check was written when the token appeared only in L7's own text
(`acceptance.md` v0.4.0, where `grep -n 'ANCHOR' acceptance.md` returned nothing at all).

This is surfaced rather than repaired: rewriting an acceptance criterion mid-run so that it fits the
artifact just produced is the exact shape of a criterion that cannot fail, and `acceptance.md` is
not this milestone's to author. A replacement key that would actually discriminate — e.g. a grep for
L7's distinguishing phrase rather than for the class name — is a decision for the lead or the
operator.

---

#### Claim 5 — the exemption marker surface is fixed (M2, AC-MRG-003 syntax half)

**Claim.** The doctrine section fixes: the syntax `<!-- moving-ref-ok: <reason> -->`; the mandatory
non-empty reason; the line scope (the flagged line or the line immediately above it); and the
incomplete-marker behaviour of REQ-MRG-003 — an empty or whitespace-only reason produces a finding
reporting the marker as incomplete rather than suppressing. `spec.md` §H Q1 is resolved to the
invisible HTML-comment form, with the grounds recorded there and a HISTORY row added (v0.6.0).

**Evidence:**

```
$ grep -c 'moving-ref-ok' .claude/rules/moai/core/verification-claim-integrity.md
1
$ grep -c 'Q1 — Marker syntax. RESOLVED at M2' .moai/specs/SPEC-MOVING-REF-GUARD-001/spec.md
1
```

The marker string appears once, in the fenced syntax block; the four properties are stated as prose
bullets beneath it.

**Baseline-attribution.** This run, this tree, HEAD `a1b8439c3`.

---

#### Gaps — what was explicitly NOT observed

- **M3 (the detector) was not started.** No file under `internal/spec/` was read or written. No
  `MovingRefUnpinned` rule exists, so AC-MRG-001, -002, -004, -005, -006, -008, -009, -013 and -014
  are entirely unobserved by this milestone.
- **M4 (corpus triage) and M5 (template mirror) were not started.** The 42 candidate lines of §B.3
  were not classified, and the doctrine was not mirrored into
  `internal/template/templates/.claude/rules/moai/core/`. AC-MRG-010 and AC-MRG-012 are unobserved.
- **Q0 is unanswered and remains with the operator.** M3 cannot be implemented without it
  (`spec.md` §H).
- **No Go command was run.** `go build`, `go vet`, and `go test` were not executed in this
  delegation — nothing in M1/M2 touches Go, and running them would have measured a tree this work
  did not change. plan.md §C step 3's `go build ./internal/spec/...` pre-flight is therefore
  unobserved.
- **`BASELINE_SHA` was not independently re-resolved.** It was taken as given from the dispatch and
  used as the frozen citation. Its correspondence to `origin/develop` at pre-flight rests on the
  dispatcher's measurement, not on one made here.
- **The doctrine's mirror-neutrality was not checked.** The section as written carries two SPEC-ID /
  AC-ID tokens (`SPEC-GRAPH-FRESHNESS-CADENCE-001`, `AC-COORD-016`) that the template-neutrality
  guard forbids in a mirrored copy. They are legitimate in the local copy and are M5's problem;
  no neutrality check was run here.
- **Only the two AC-deciding greps were run.** The doctrine was not read back end-to-end by a second
  reader, and no check was made that its prose is consistent with `spec.md` §D beyond the greps.

#### Residual-risk — what could still be wrong despite the above

- **The greps verify presence, not correctness.** Every criterion decided here is a token count. A
  doctrine that named all four tests while stating them wrongly would pass every command recorded
  above. The mutations raise the floor (a *missing* element is caught) without touching the ceiling.
- **The doctrine section adds ~14.5 KB to an always-loaded file.** Measured, not estimated:
  `git show HEAD:<doctrine> | wc -c` → `8224`; `wc -c <doctrine>` → `23020`; growth 14,796 bytes.
  Every session pays that on every turn and again after each `/clear`,
  including sessions that never write a moving-ref claim. `spec.md` §C fixes the placement as one
  always-loaded doctrine section and REQ-MRG-005 requires it, so this is a design decision rather
  than a defect — but a reviewer preferring a `paths:`-scoped companion would be disagreeing with
  the SPEC, not with the implementation. The `rule-authoring.md` statement duty is discharged in the
  commit body.
- **`plan.md` §F M1 still says "the **six** detection limits (§F)".** `spec.md` §F carries seven and
  AC-MRG-011 greps for seven; the plan text was not swept when L7 was added at v0.4.0. Seven were
  authored. The stale plan line was left in place rather than corrected, because sweeping it is not
  in M1/M2's scope — recorded here so it is not read later as evidence that six were intended.
- **The five grounded instances remain a small validation set**, and the ANCHOR branch still has
  zero adjudicated instances (L7). The doctrine now publishes that gap, which does not close it.

---

### M3 — the detector (2026-08-28)

**Scope.** `internal/spec/lint_movingref.go` (new), its registration in `internal/spec/lint.go`,
`internal/spec/lint_movingref_test.go` (new), and eight fixture SPEC directories under
`internal/spec/testdata/movingref/`. No SPEC artifact outside this SPEC's own directory was touched,
and within it only this file.

**Q0 disposition — resolved as option C, by operator decision.** REQ-MRG-010 (the R4-form exclusion)
and AC-MRG-013 (with both counter-mutations CM-1 and CM-2) are **deferred out of this scope**. No
R4-form exclusion, imperative-structure recognizer, or command-token exclusion exists in the rule;
no AC-MRG-013 fixture was written and CM-1 / CM-2 were not run. Grounds, per `spec.md` §B.7: R4's
reachable class measured **0 of 42** candidate lines on two independent probes, so an exclusion for
it can only over-exempt, never under-exempt — every error available to it is a D11-shaped bypass.
R4 remains in the finding **message** (AC-MRG-008 asserts it); only its lint exclusion is deferred.
Recording the deferral in `spec.md` / `plan.md` / `acceptance.md` is manager-spec's artifact
ownership and was deliberately NOT done here.

#### Claim 1 — the suite is green, and the selector swept something

The case count is reported rather than `ok` alone: a test-name regex matching zero tests prints `ok`
and asserts nothing.

**Evidence.**

```
$ go test ./internal/spec/ -run 'TestMovingRef' -v 2>&1 | tail -25
=== RUN   TestMovingRef_FiresOnUnpinnedAnchor
--- PASS: TestMovingRef_FiresOnUnpinnedAnchor (0.38s)
=== RUN   TestMovingRef_PinnedClaimNotFlagged
--- PASS: TestMovingRef_PinnedClaimNotFlagged (0.23s)
=== RUN   TestMovingRef_MarkerSuppressesOnlyWithReason
--- PASS: TestMovingRef_MarkerSuppressesOnlyWithReason (0.67s)
=== RUN   TestMovingRef_ThreeDotNotExempt
--- PASS: TestMovingRef_ThreeDotNotExempt (0.22s)
=== RUN   TestMovingRef_SeverityWarningAndExitCode
--- PASS: TestMovingRef_SeverityWarningAndExitCode (0.48s)
=== RUN   TestMovingRef_DivergenceFigureVariant
--- PASS: TestMovingRef_DivergenceFigureVariant (0.24s)
=== RUN   TestMovingRef_MessageNamesAllFourBranches
--- PASS: TestMovingRef_MessageNamesAllFourBranches (0.24s)
=== RUN   TestMovingRef_ReadsSiblingArtifacts
--- PASS: TestMovingRef_ReadsSiblingArtifacts (0.22s)
=== RUN   TestMovingRef_NegativeControlOnClaimConjunct
--- PASS: TestMovingRef_NegativeControlOnClaimConjunct (0.23s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/spec	3.101s
```

Nine cases selected, nine PASS — the selector is not vacuous.

Whole-package, with the case count measured rather than inferred:

```
$ go test ./internal/spec/... 2>&1 | tail -20
ok  	github.com/modu-ai/moai-adk/internal/spec	28.595s

$ go test ./internal/spec/... -v 2>&1 | grep -c '^--- PASS\|^    --- PASS'
572

$ go test ./internal/spec/... -v 2>&1 | grep -cE '^(--- FAIL|    --- FAIL|--- SKIP|    --- SKIP)'
0

$ go vet ./internal/spec/... && echo VET_RC_0
VET_RC_0

$ go build ./... && echo BUILD_RC_0
BUILD_RC_0
```

572 cases pass; 0 fail and 0 skip. `gofmt -l` reports neither new file.

#### Claim 2 — every stated mutation was planted, observed red, and reverted

Ten mutations. Each is recorded with its verbatim failing output; each was reverted and the suite
re-run green afterwards. A criterion asserted falsifiable without the mutation having been run is
not evidence (`verification-claim-integrity.md` §1).

**M-001 (AC-MRG-001) — `origin/main` → `origin/mainx` in the fixture. The first attempt did NOT
turn it red, and that is the finding.**

```
$ go test ./internal/spec/ -run 'TestMovingRef_FiresOnUnpinnedAnchor' 2>&1 | tail -20
ok  	github.com/modu-ai/moai-adk/internal/spec	0.430s
```

The criterion was **vacuous as first implemented**: the §B.3 filter-2 alternation
`origin/(main|develop|HEAD)` has no right boundary, so `origin/mainx` matches on the `origin/main`
prefix and the mutation could not falsify anything. A trailing `\b` was added to
`movingRefGitContextPattern` and `movingRefTokenPattern` — a deliberate tightening of the recorded
filter, priced below — after which the mutation is falsifying:

```
$ go test ./internal/spec/ -run 'TestMovingRef_FiresOnUnpinnedAnchor' 2>&1 | tail -20
--- FAIL: TestMovingRef_FiresOnUnpinnedAnchor (0.24s)
    lint_movingref_test.go:90: expected exactly 1 MovingRefUnpinned finding, got 0: []
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/spec	0.605s
FAIL
```

Reverted; suite green (`ok ... 2.975s`).

**M-002 (AC-MRG-002) — delete the hex-SHA exclusion branch.**

```
--- FAIL: TestMovingRef_PinnedClaimNotFlagged (0.23s)
    lint_movingref_test.go:114: expected 0 findings on a SHA-pinned claim, got 1: [{testdata/movingref/pinned-sha/acceptance.md 15 warning MovingRefUnpinned `origin/main` decides an invariant claim on this line with no SHA pin and no frozen-baseline variable, so the measurement expires the moment mainline advances. …}]
FAIL	github.com/modu-ai/moai-adk/internal/spec	3.144s
```

**M-003 (AC-MRG-003) — remove the non-empty-reason check (any marker suppresses).**

```
--- FAIL: TestMovingRef_MarkerSuppressesOnlyWithReason (0.64s)
    lint_movingref_test.go:136: (c) bare marker: expected 1 finding, got 0: []
FAIL	github.com/modu-ai/moai-adk/internal/spec	3.027s
```

**M-004 (AC-MRG-004) — add a `strings.Contains(line, "...")` early return.**

```
--- FAIL: TestMovingRef_ThreeDotNotExempt (0.23s)
    lint_movingref_test.go:154: expected 1 finding on the three-dot form, got 0: []
--- FAIL: TestMovingRef_DivergenceFigureVariant (0.23s)
    lint_movingref_test.go:213: expected 1 finding on the divergence-figure form, got 0: []
FAIL	github.com/modu-ai/moai-adk/internal/spec	3.081s
```

AC-MRG-006 also goes red because the divergence fixture uses the `...` form; the criterion the
mutation names is red, and the collateral is consistent rather than a separate signal.

**M-005 (AC-MRG-005) — emit at `SeverityError`.**

```
--- FAIL: TestMovingRef_SeverityWarningAndExitCode (0.45s)
    lint_movingref_test.go:185: expected severity "warning", got "error"
    lint_movingref_test.go:191: non-strict lint must exit 0 when the only findings are MovingRefUnpinned warnings
FAIL	github.com/modu-ai/moai-adk/internal/spec	0.813s
```

Both halves red — the severity assertion **and** the exit-code assertion, which is the half the
criterion is actually about.

**M-006 (AC-MRG-006) — append a resolved SHA to the divergence fixture line.**

```
--- FAIL: TestMovingRef_DivergenceFigureVariant (0.27s)
    lint_movingref_test.go:213: expected 1 finding on the divergence-figure form, got 0: []
FAIL	github.com/modu-ai/moai-adk/internal/spec	2.969s
```

The rule keys on the missing pin, not on the `rev-list` verb alone.

**M-008a (AC-MRG-008) — shorten the message to name only pinning.**

```
--- FAIL: TestMovingRef_MessageNamesAllFourBranches (0.21s)
    lint_movingref_test.go:249: message omits R1 (pin): expected substring "pin the resolved SHA" in: … Pin the SHA.
    lint_movingref_test.go:249: message omits R2 (freeze at pre-flight): …
    lint_movingref_test.go:249: message omits R3 (declare the exemption): …
    lint_movingref_test.go:249: message omits R4 (state the command): …
    lint_movingref_test.go:254: message must not present pinning as the sole remedy; got: …
    lint_movingref_test.go:258: message must state that it is a question rather than a verdict; got: …
FAIL	github.com/modu-ai/moai-adk/internal/spec	0.576s
```

**M-008b (AC-MRG-008, second mutation) — drop ONLY R4, leaving three branches.**

```
--- FAIL: TestMovingRef_MessageNamesAllFourBranches (0.21s)
    lint_movingref_test.go:249: message omits R4 (state the command): expected substring "state the measuring command" in: `origin/main` decides an invariant claim on this line … then choose among four remedies — R1 pin the resolved SHA; R2 freeze at pre-flight (`BASELINE_SHA=$(git rev-parse origin/main)`) and decide the criterion against the frozen value; R3 keep the moving ref and declare the exemption with a non-empty reason, `<!-- moving-ref-ok: <reason> -->`; Pinning is one branch of four, not the default. …
FAIL	github.com/modu-ai/moai-adk/internal/spec	0.574s
```

The assertion still fails with three branches present, and fails on the R4 row alone — one assertion
per branch, so no single branch can be dropped silently.

**M-009 (AC-MRG-009) — restrict the rule to `doc.Body`.**

```
--- FAIL: TestMovingRef_FiresOnUnpinnedAnchor (0.23s)
    lint_movingref_test.go:90: expected exactly 1 MovingRefUnpinned finding, got 0: []
--- FAIL: TestMovingRef_MarkerSuppressesOnlyWithReason (0.69s)
    lint_movingref_test.go:129: (a) no marker: expected 1 finding, got 0: []
    lint_movingref_test.go:136: (c) bare marker: expected 1 finding, got 0: []
--- FAIL: TestMovingRef_ThreeDotNotExempt (0.23s)
    lint_movingref_test.go:154: expected 1 finding on the three-dot form, got 0: []
--- FAIL: TestMovingRef_SeverityWarningAndExitCode (0.22s)
    lint_movingref_test.go:182: expected 1 MovingRefUnpinned finding, got 0
--- FAIL: TestMovingRef_DivergenceFigureVariant (0.22s)
    lint_movingref_test.go:213: expected 1 finding on the divergence-figure form, got 0: []
--- FAIL: TestMovingRef_MessageNamesAllFourBranches (0.22s)
    lint_movingref_test.go:236: expected 1 finding, got 0
--- FAIL: TestMovingRef_ReadsSiblingArtifacts (0.23s)
    lint_movingref_test.go:274: expected 1 finding from the sibling progress.md, got 0: []
FAIL	github.com/modu-ai/moai-adk/internal/spec	2.838s
```

AC-MRG-009 goes red as required. **Its stated separation clause does not hold, and the reason is in
`acceptance.md` rather than in the implementation.** AC-MRG-009 says the body-only mutant "fails
while AC-MRG-001 still passes — which is exactly why it is a separate criterion". But AC-MRG-001's
own fixture is specified as *"a fixture SPEC whose `acceptance.md` carries the row"*, and
`SPECDoc.Body` carries `spec.md` alone — so a body-only mutant necessarily takes AC-MRG-001 down
with it. The two clauses are inconsistent with each other; the fixture placement was followed as
written and the inconsistency is reported rather than resolved here (it is `acceptance.md` body
content, which run-phase does not own).

**M-014 (AC-MRG-014) — delete the claim-marker conjunct. The mutant the Definition of Done singles
out.**

```
--- FAIL: TestMovingRef_NegativeControlOnClaimConjunct (0.22s)
    lint_movingref_test.go:296: expected 0 findings on a claim-free instruction, got 1: [{testdata/movingref/no-claim/progress.md 14 warning MovingRefUnpinned `origin/main` decides an invariant claim on this line with no SHA pin and no frozen-baseline variable, …}]
FAIL	github.com/modu-ai/moai-adk/internal/spec	3.126s
```

**AC-MRG-014 is the ONLY red.** Every other criterion — AC-MRG-001, -002, -003, -004, -005, -006,
-008, -009 — stays green under the all-matching mutant, which is the property the DoD asserts and
which is here measured rather than argued.

#### Claim 3 — the over-fire the claim-marker conjunct prevents, measured at this tree

`acceptance.md` AC-MRG-014 cites **495** findings for the mutant at `43329ec8b`. Re-measured here
rather than carried over, using the rule's own alternations as a grep equivalent (a grep prevalence
measurement, not rule output — the same caveat §E.1 records for the 42):

```
$ grep -rnE 'git [a-z-]+[^`]*origin/(main|develop|HEAD)\b' .moai/specs --include='*.md' \
    | grep -v SPEC-MOVING-REF-GUARD-001 | grep -cvE '\b[0-9a-f]{7,40}\b'
451

$ grep -rnE 'git [a-z-]+[^`]*origin/(main|develop|HEAD)\b' .moai/specs --include='*.md' \
    | grep -v SPEC-MOVING-REF-GUARD-001 \
    | grep -iE 'byte-unchanged|byte unchanged|unchanged|preserv|보존|no diff|empty|0 files|부재|absent|변경 ?없|그대로' \
    | grep -cvE '\b[0-9a-f]{7,40}\b'
42
```

**451 vs 42 — a 10.7x over-fire at this tree**, against the 495 vs 42 (11.8x) recorded at
`43329ec8b`. The figures differ because the trees differ; the conclusion does not. Bulk suppression
is the rational response to 451 warnings, which defeats §D.5 directly.

The second command also reproduces `spec.md` §B.3's filter-4 figure of **42 exactly**, at this tree.

#### Claim 4 — the `\b` tightening is priced, and costs nothing on the live corpus

The tightening M-001 forced is a deviation from the alternation recorded verbatim in §E.1, so its
effect is measured rather than assumed. Same filter chain, `\b` removed:

```
$ grep -rnE 'git [a-z-]+[^`]*origin/(main|develop|HEAD)' .moai/specs --include='*.md' \
    | grep -v SPEC-MOVING-REF-GUARD-001 \
    | grep -iE 'byte-unchanged|byte unchanged|unchanged|preserv|보존|no diff|empty|0 files|부재|absent|변경 ?없|그대로' \
    | grep -cvE '\b[0-9a-f]{7,40}\b'
42
```

**42 with the boundary and 42 without it** — the tightening is a no-op on the live corpus and
changes no §B.3 figure. It buys AC-MRG-001's falsifiability at zero measured cost.

#### Claim 5 — AC-MRG-002's fixture deviates from its literal wording, deliberately

AC-MRG-002 says "the AC-MRG-001 fixture with `origin/main` **replaced by** a 40-character
hexadecimal SHA". Taken literally the moving-ref token disappears, the line then fails conjuncts 1+2
rather than being exempted by REQ-MRG-008, and the criterion's own stated mutation (delete the
hex-SHA exclusion branch) **cannot turn it red** — a vacuous criterion of exactly the shape §A
exists to prevent.

The fixture therefore **retains** the ref and records the resolved SHA on the same line — the shape
filter 4 of §B.3 actually removes (a line in a git-command context carrying a 7-40 hex SHA), and the
shape REQ-MRG-008 is worded for ("While a **line's moving ref** carries a resolved SHA pin"). M-002's
red output above confirms the criterion is falsifiable under this fixture.

#### Claim 6 — AC-MRG-006 was KEPT, not withdrawn

REQ-MRG-006 is SHOULD-tier and `spec.md` §H Q2 permits withdrawal at run-phase. It was kept. The
measurement that decided it: the divergence branch requires a **cited figure** (a result pair such
as `→ 0 0`), not merely the `rev-list --count --left-right` verb, so it does not fire on
AC-MRG-014's negative-control line, which names the same command and asserts nothing. That fixture
passing green (0 findings) with AC-MRG-006 passing green (1 finding) is the evidence the two shapes
are separated. No over-fire was observed on any fixture.

#### Baseline-attribution

- Frozen pre-flight baseline: `BASELINE_SHA = d566ecc7511e1954e3aeb1dff3a60afa5be1089b`, taken as
  given from the dispatch and cited as a literal SHA, never re-resolved from `origin/develop`.
  Re-resolving it would make this SPEC an instance of its own defect.
- Every command above ran against worktree `WT-moving-ref-guard` at HEAD **`cb90e8a86`** (`git
  rev-parse --short HEAD`, re-read immediately before this record was written), with the M3 changes
  present in the working tree and not yet committed.
- `git status --short` at that moment: ` M internal/spec/lint.go`, `?? internal/spec/lint_movingref.go`,
  `?? internal/spec/lint_movingref_test.go`, `?? internal/spec/testdata/movingref/` — no file outside
  this SPEC's directory and `internal/spec/` was modified.

#### Gaps — what was explicitly NOT observed

- **M4 (corpus triage) and M5 (template mirror) were not started.** AC-MRG-010 and AC-MRG-012 are
  entirely unobserved. The 451 / 42 figures above are a grep prevalence measurement for the severity
  argument only — they are **not** the M4 triage, which must classify each finding as anchor /
  subject / already-frozen with a reason, and they are not rule output.
- **REQ-MRG-010 and AC-MRG-013 are unimplemented and untested** (Q0 option C). CM-1 and CM-2 were
  not run. No claim is made here about whether an R4 exclusion would work.
- **The rule was never run over the live `.moai/specs` corpus.** Every figure in Claim 3 and Claim 4
  is grep, not `MovingRefUnpinned` output. The two may disagree, and no criterion here assumes they
  agree — the same caveat §E.1 records for the plan-phase 42.
- **`AC-MRG-007` and `AC-MRG-011` were not re-verified.** They are M1's criteria; nothing in M3
  re-read the doctrine file.
- **`BASELINE_SHA` was not independently re-resolved** — same Gap as M1/M2. Its correspondence to
  `origin/develop` at pre-flight rests on the dispatcher's measurement.
- **No cross-platform build was run.** Only host `go build ./...`; no `GOOS=windows` / `GOOS=linux`
  vet.
- **No `go test ./...`.** Package-scoped only, per plan.md §D and the card's constraint.
- **`golangci-lint` was not run.** Only `gofmt -l` (clean on both new files) and `go vet`.
- **The five §B.3-named corpus lines were not checked against the rule.** Whether the rule actually
  fires on `AC-TST-010` / `AC-AFS-012` and stays silent on `AC-COORD-016` / `REQ-LB-006` is M4's
  question and is unobserved.

#### Residual-risk — what could still be wrong despite the above

- **The eight fixtures are synthetic and were authored by the same actor as the rule.** A shared
  misreading of REQ-MRG-001 would produce a fixture and a rule that agree with each other and both
  disagree with the corpus. Only M4 tests that.
- **Limit L2 is unchanged and now has teeth.** The detector reads shape, never subject, so on the
  live corpus it will fire on `REQ-LB-006` and any other line where the command is quoted subject
  matter. That is expected (L4) and is what the marker is for — but until M4 classifies the corpus,
  the true-positive rate is unmeasured, and a reader meeting a wall of findings may bulk-suppress
  regardless of what the message says.
- **The divergence branch's figure-citation heuristic is the least-grounded piece.** It accepts
  `→ | -> | => | : | =` followed by two integers, or a backtick-wrapped integer pair. `:` and `=`
  are broad, and no corpus measurement was taken of how often they precede an unrelated integer pair
  on a `rev-list` line. AC-MRG-014's fixture is the only negative control it has.
- **`shaPinPattern` is line-scoped and lexical.** Any 7-40 character lowercase-hex word on the line
  exempts it, including an English word that happens to be hex-shaped (`defaced`, `decade` + a
  digit) and an unrelated SHA. This is inherited from §B.3 filter 4 rather than introduced here, but
  it means a claim can be exempted by an anchor that anchors nothing — a rotted-value blind spot of
  the same family as L6.
- **The rule re-reads every `*.md` in each SPEC directory on every `Lint()` call.** Whole-package
  test time is 28.6s and no regression was observed, but no before/after timing was measured, so the
  cost on a large catalog is unquantified.

### M4 — corpus triage record, classification only (2026-08-28)

**Deliverable**: `.moai/reports/t342/corpus-triage.md`. Classification only — `plan.md` §F M4's
[HARD] no-edit clause held: no marker was added, no SHA was pinned, and no SPEC artifact outside
this SPEC's own directory was touched.

#### Claim 1 — the rule emits 107 findings over `.moai/specs/**`

```
$ go run ./cmd/moai spec lint --json > <out>
rc=0

$ jq -r '.[].code' <out> | sort | uniq -c | sort -rn
 107 MovingRefUnpinned
  24 MissingExclusions
  18 StatusGitConsistency
  14 FrontmatterInvalid
   7 LegacyEARSKeyword
   1 OwnershipTransitionInvalid

$ jq -r '.[] | select(.code=="MovingRefUnpinned") | "\(.file):\(.line)"' <out> | wc -l ; … | sort -u | wc -l
     107
     107
```

107 distinct `file:line` pairs, no duplicates despite `Check` being invoked once per `SPECDoc` and
re-scanning the whole SPEC directory each time. Zero findings are the REQ-MRG-003 incomplete-marker
variant — no `moving-ref-ok` marker exists anywhere in the corpus yet.

#### Claim 2 — 107 vs the §B.3 prevalence figure of 42, fully attributed

`§E.1` records that the plan-phase 42 is a grep prevalence measurement, not rule output, that the
two may differ, and that no criterion assumes they agree. They differ, by **+55** on the
comparable population (97 external findings vs 42), and neither figure was adjusted to meet the
other.

```
rule findings, whole catalogue                                   107
rule findings, this SPEC's own directory excluded                 97
§B.3 filter-4 grep, re-run at HEAD 1ed8a9997, own dir excluded     42
```

Decomposition, measured:

```
$ cut -f2 <97 external findings> | grep -ciE 'byte-unchanged|byte unchanged|unchanged|preserv|보존|no diff|empty|0 files|부재|absent|변경 ?없|그대로'
41
$ cut -f2 <97 external findings> | grep -cE 'rev-list\s+--count\s+--left-right'
61
```

- **+56** from the REQ-MRG-006 divergence-figure branch, which filter 3 never contained (61 lines
  carry the divergence command; 56 of them do not match the claim alternation). 41 + 61 − 97 = 5
  lines satisfy both branches.
- **−1** from the REQ-MRG-008 frozen-baseline exclusion, which filter 4 does not have:

```
$ comm -23 <grep-42 locations> <rule-97 locations>
.moai/specs/SPEC-DESIGN-MOAIWEBV2-002/acceptance.md:14
```

  That line uses `$BASELINE_SHA` — the R2 remedy in use, correctly exempted by the rule and blindly
  counted by the grep. 41 + 56 = 97. The §B.3 figure reproduces at this HEAD unchanged (`42`).

#### Claim 3 — classification tally, and the S2 count corroborates §B.7

```
class            all   external   own dir
anchor            48         48         0
subject / S1      56         48         8
subject / S2       2          0         2
already-frozen     1          1         0
unclassifiable     0          0         0
total            107         97        10
```

Both §B.7 probes, re-run against the rule's 97 external findings rather than the grep 42, return
**1** hit each; the hit (`SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/plan.md:116`) is a directive with no
demoted dated reference value, so it is a probe false positive. **External R4 occupancy is 0**,
unchanged from §B.7's measurement. Both S2 findings sit in this SPEC's own directory and are the R4
form this SPEC's own doctrine created — exactly what §B.7 predicted ("every R4 remediation M1
recommends produces exactly one"). They are flagged because the REQ-MRG-010 exclusion is deferred
(Q0 option C).

All three §B.3-named classes reproduce under the rule, including both subject-class lines
(`AC-COORD-016` finding 82, `REQ-LB-006` finding 53), which the rule flags as expected under L2/L4
and which **must not be pinned**.

#### Claim 4 — AC-MRG-010's two decider forms, both run, both recorded

Form A, the criterion as literally written in `acceptance.md:161`:

```
$ git diff --name-only "$BASELINE_SHA"..HEAD -- .moai/specs | grep -v SPEC-MOVING-REF-GUARD-001
.moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/progress.md
.moai/specs/SPEC-GRAPH-FRESHNESS-CADENCE-001/spec.md
.moai/specs/SPEC-SYNC-STRATEGY-KEY-001/progress.md
.moai/specs/SPEC-SYNC-STRATEGY-KEY-001/spec.md
```

**NON-EMPTY**, and attributed rather than remediated:

```
$ git log --format='%h %s' --no-merges "$BASELINE_SHA"..HEAD -- <those two directories>
f9c827217 docs(SPEC-GRAPH-FRESHNESS-CADENCE-001): correct two overstated claims in the sync signal (t322)
28f16d030 docs(SPEC-GRAPH-FRESHNESS-CADENCE-001): backfill sync_commit_sha (t322)
bc66c30b7 docs(SPEC-GRAPH-FRESHNESS-CADENCE-001): sync-phase close (3-phase plan->run->sync) (t322)
ed68889e3 docs(SPEC-SYNC-STRATEGY-KEY-001): terminal transition implemented -> completed (t303)
```

All four are cards t322 / t303, landed on mainline and absorbed by this branch's two merge commits;
each confirmed an ancestor of `origin/develop` (re-fetched immediately before the judgment,
`origin/develop` = `77b2bcae6` at that moment) with `merge-base --is-ancestor` rc=0 on all four.
None was authored on `WT-moving-ref-guard`.

Form B, narrowed to what this branch adds beyond mainline:

```
$ git diff --name-only 5e194bba27c146d8c2157d92b4a3fb3995919ff0..HEAD -- .moai/specs | grep -v SPEC-MOVING-REF-GUARD-001
(empty)
```

Second command of the pair, both forms:

```
$ git status --short -- .moai/specs
(empty)
```

`acceptance.md` was NOT edited to change the decider — its body content is manager-spec's. The
report's §4.1 states a view on which form should decide, as an input to the orchestrator.

#### Baseline-attribution

- Frozen pre-flight baseline: `BASELINE_SHA = d566ecc7511e1954e3aeb1dff3a60afa5be1089b`, cited as a
  literal and never re-resolved from `origin/develop`, per `plan.md` §C step 4.
- Every command above ran in worktree `WT-moving-ref-guard` at HEAD **`1ed8a9997`**
  (`git rev-parse --short HEAD`, re-read immediately before this record was written), working tree
  clean at the time of the lint run.
- `origin/develop` was re-fetched immediately before the ancestry judgment and resolved to
  `77b2bcae6` at that moment; the four commits' ancestry is attributed to that reading, not to a
  ref value read earlier in the session.

#### Gaps — what was explicitly NOT observed

- **M5 (template mirror) was not started.** AC-MRG-012 is entirely unobserved.
- **REQ-MRG-010 and AC-MRG-013 remain unimplemented and untested** (Q0 option C). CM-1 and CM-2 were
  not run. Findings 27 and 32 are the deferral's live consequence; no claim is made about whether an
  exclusion would classify them correctly.
- **Every classification is one reader's judgment** — limit L2 acting on the triage rather than on
  the detector. The predicate was applied by hand by the same actor that ran the rule, with no
  independent adjudication. L7 sharpens this for the anchor rows: the ANCHOR branch has zero
  adjudicated instances, and this milestone adds 48 author-classified anchors to the seven §F L7
  already counts.
- **Rows where a second reader could reasonably differ are named in the report §6** (findings 11,
  22, 42, 65, 67, 68, 80, 81, 91, 98, 103) rather than forced into a bucket. `unclassifiable` is 0
  because each of those has a stated reading, not because no doubt exists.
- **No claim that 107 is the complete defect set.** Limits L1, L3, and L5 are unchanged and
  unmeasured here; `.moai/reports/**`, where the R4-motivating instance lives, was not scanned.
- **No test was run and no build was performed.** `go run ./cmd/moai spec lint` compiled the tree as
  a side effect (rc 0) and is the only compilation evidence this milestone produces. No
  `go test`, no `gofmt`, no `golangci-lint`, no cross-platform vet.
- **The rule was run against this worktree at this HEAD only.** The figures are not claimed to hold
  on `origin/develop` or any other branch.

#### Residual-risk

- **A misreading of §D.1 would reproduce itself uniformly** across 107 rows rather than surfacing as
  an inconsistency, because the shape buckets were derived from the corpus by the same reader who
  then applied them. The external S2 count of 0 is the only figure with an independent check against
  it (§B.7, measured before this milestone existed), and it agrees.
- **The `DIR` bucket is 35 external rows and carries the most doctrinal weight.** If a directive
  counts as S2 — that is, if Test 4's "a reader will act on it" is read to include executing a
  procedure step — then 35 findings move from R3 to R4 and §B.7's empty-scope conclusion moves with
  them. The narrower reading was taken (Test 4 asks whether the claim *asserts the current state of
  a moving thing*, which a directive does not), and a reviewer who disagrees should expect the R4
  scope question to reopen.
- **107 findings is a large enough wall to provoke the response this SPEC exists to prevent.** 56
  are subject-class lines correct as written that will keep firing until 56 markers are written; the
  incentive to write one blanket suppression instead is real, and only the reason field's cost
  opposes it — a price enforced by review, not by the detector (L6's second residual).
- **The §2 decomposition rests on grep re-derivations of the rule's own branches**, not on
  instrumented rule output: the rule does not report which conjunct fired, so 41 / 61 / 56 were
  reconstructed by applying its published patterns to the flagged lines. A divergence between the
  reconstruction and the rule's actual internal path would not be visible.
- **`SPEC-DESIGN-MOAIWEBV2-002/plan.md:36` (finding 7) shows the rule flagging the line that teaches
  R2** while correctly exempting the line that applies it, because the frozen-baseline pattern
  matches a variable reference and not the assignment that creates it. Recorded as an observation
  only; changing it is a Go source change M4 does not make.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
