---
id: SPEC-REVIEW-SECRET-SCAN-REFS-001
title: "Review workflow secret scan — coverage of refs not reachable from HEAD"
version: "0.2.0"
status: draft
created: 2026-09-10
updated: 2026-09-10
author: manager-spec (card t629)
priority: P1
phase: "v3.2.0 target"
module: ".claude/skills/moai/workflows"
lifecycle: spec-anchored
tags: "review, security, secrets-scan, git-history, checkpoint, documentation"
tier: M
---

# SPEC — Review workflow secret scan: coverage of refs not reachable from HEAD

## HISTORY

- 2026-09-10: plan-phase authored on card t629 (Class C). Every claim about the defect traces to
  `.moai/reports/t629/reproduction.md`; every cost figure traces to
  `.moai/reports/t629/cost-baseline.md`. Both were measured on tree `feeecc980`; the review
  workflow copies are unchanged between `feeecc980` and HEAD `21e5837dc`
  (`git diff --stat feeecc980 21e5837dc -- <both copies>` printed nothing).
- The choice between the options in §3 is **deliberately left open**. The card reserves any
  narrowing of security-scan coverage for cost to the operator; this SPEC records the options
  and does not choose.
- 2026-09-10 (0.1.1): §3.3 folds in the lead's path-only classification of the full-scan matches
  (lead-reported, not re-measured by the lane). The decision in §3 is still pending.
- 2026-09-10 (0.2.0): the operator decided §3 — Option 2 (per-ref tip-set checkpoint), adopted on
  the condition that the run phase measures it first, with an exact example-value allowlist and no
  path exclusions; answered in the lead session and relayed by the lead. §3.4 adds the adoption
  conditions; §2 states the requirements for Option 2 and adds REQ-009 to REQ-012; §5 aligns its
  unexamined-content wording with §3.3.

## §1 Background and problem statement

The review workflow document (`.claude/skills/moai/workflows/review.md`, with a byte-identical
distributed copy under `internal/template/templates/`) prescribes a secrets scan over git history
in the section `#### Secrets Scan (Incremental with Checkpoint)`:

- Where a checkpoint exists, scan `git log -p <last-sha>..HEAD -G '<regex>'` plus the working
  tree, then move the checkpoint to the current HEAD.
- Only on first run, or when an explicit full-scan flag is passed, scan
  `git log -p --all -G '<regex>'`.
- The section closes by claiming the incremental range plus checkpoint update "produces the same
  coverage over time as the former every-review full scan — no finding class is dropped".

The checkpoint exists **only as prose** in the two document copies: `secrets-scan-checkpoint` has
zero matches in Go source under `internal/`, `pkg/`, `cmd/`. The change surface is therefore the
two document copies, not code.

### §1.1 Measured defect (throwaway fixture repository, outside this repository)

Markers are synthetic PEM-style private-key header lines with two distinct uppercase labels,
`HEADCELL` (committed on the HEAD line) and `SIDECELL` (committed on a side branch not reachable
from HEAD; `git merge-base --is-ancestor side HEAD` exit 1). The scan regex is the one in
`review.md`. Source: `reproduction.md` § Evidence — fixture.

| measurement | incremental scan | `--all` scan |
|---|---|---|
| baseline control, before any marker | — | exit 0, 0 bytes |
| 1 — checkpoint at the clean commit | HEADCELL 1, **SIDECELL 0** | HEADCELL 1, SIDECELL 1 |
| 2 — next review, checkpoint advanced | 0 bytes, **SIDECELL 0** | SIDECELL 1 |
| 3 — after merging the side branch | SIDECELL 1 | — |

Reading: the scanner is alive (HEADCELL 1); a line on an unmerged branch is missed and stays
missed as the checkpoint advances (measurement 2), which refutes the equivalence claim; merging
brings it into range (measurement 3). The blind spot is refs unreachable from HEAD that are never
merged.

### §1.2 Measured on this repository

Source: `cost-baseline.md`. The full `--all` scan matched the regex in **15** commits: **1**
reachable from HEAD, **14** reachable only through other refs. The content of those commits was
deliberately not examined; they may be fixtures, documentation examples, or real leaks. A
HEAD-anchored incremental scan would never reach the 14. For the lead's later path-only
classification of these matches (0 real-leak candidates by path and shape, with the gap that a
path cannot tell a real value inside a documentation or example folder apart), see §3.3.

## §2 GEARS requirements

The operator adopted Option 2 with an exact example-value allowlist (§3). The requirements below
are stated for that decision. Where an earlier draft marked a requirement as option-specific, the
readings for Options 1 and 3 are dropped, because those options were not adopted; §3.2 keeps them
as the record of the choice.

- **REQ-001 (Ubiquitous):** The review workflow's secrets-scan procedure — taken as the complete
  sequence of steps the workflow document prescribes across successive reviews — shall report a
  credential-shaped line committed on a ref that is not reachable from HEAD and has not been
  merged. Under the adopted Option 2 the per-review step meets this: the line is reported in the
  first completed scan after its commit becomes reachable from any ref (§3.4 cell ①). Options 1
  and 3 were not adopted.

- **REQ-002 (When):** **When** the checkpoint advances after a completed scan, the procedure shall
  still report a credential-shaped line committed on an unmerged ref not reachable from HEAD that
  no earlier completed step reported, at the latest on the next step whose scope includes that
  ref. Re-reporting a line that an earlier step already reported is not required by this
  requirement; Option 2 scans only commits newly reachable since the last completed scan and is not
  expected to re-report.

- **REQ-003 (Ubiquitous):** The workflow document shall state, for every scan step it prescribes,
  the set of commits that step scans and what the checkpoint records — under Option 2, the tip of
  every ref at the last completed scan — and shall not claim coverage beyond that set.

- **REQ-004 (Where):** **Where** the Option 2 per-review step leaves commits outside its scope —
  including commits reachable from no ref, such as commits reachable only through a reflog — the
  workflow document shall name those commits as uncovered by that step, and shall name the step
  that covers them or state that no prescribed step does. The earlier Option 3 reading of this
  requirement (refs not reachable from HEAD) no longer applies, because Option 3 was not adopted.

- **REQ-005 (Unwanted):** The workflow document shall not state that the incremental scan with a
  checkpoint yields the same coverage as a full-history scan, and shall not state that no finding
  class is dropped.

- **REQ-006 (Ubiquitous):** The local and distributed copies of the review workflow document shall
  be byte-identical, and the distributed copy shall carry no internal SPEC ID, no card ID, and no
  cost figure measured on this repository, and shall name no programming language as primary.

- **REQ-007 (Unwanted):** This change shall not commit to this repository any line matching the
  scan's regex — including in evidence files and fixture scripts. The fixture lives outside the
  repository.

- **REQ-008 (Unwanted):** The run phase shall not edit the secret-scan section of either copy until
  the operator's choice among the options in §3 is recorded in this SPEC.

- **REQ-009 (When):** **When** the run phase begins, it shall take the three adoption measurements
  of §3.4 (cells ①, ②, ③) against a procedure pinned in advance, and shall commit their evidence
  in a commit that is a strict ancestor of every commit touching either copy of the review
  workflow document.

- **REQ-010 (When):** **When** a §3.4 gate cell is untrustworthy by its stated predicate, the run
  phase shall stop before editing either copy, shall record the stop and the failing cell in
  `progress.md`, and shall report back so that the decision returns to the operator.

- **REQ-011 (Ubiquitous):** The secrets-scan procedure shall suppress a match only when the text the
  scan's regex matched equals, exactly, a value on the example-value allowlist; the allowlist shall
  hold only publicly published example values, the procedure shall exclude no path from any scan
  step, and a value that differs from every listed value shall still be reported wherever it sits.

- **REQ-012 (Unwanted):** The allowlist representation shall not place, in either copy of the
  workflow document or in any other file committed to this repository, text that matches the
  scan's regex — a listed value written literally included — and shall still decide suppression by
  exact value.

## §3 Decision — coverage versus cost

**Decision:** Option 2 — per-ref tip-set checkpoint (conditional; see §3.4).

**Handling of known example values:** an exact example-value allowlist; no path exclusions.

Decided by the operator, answered directly in the lead session and relayed by the lead,
2026-09-10. The lane records the decision; it did not make it.

### §3.1 Cost basis common to all options

From `cost-baseline.md`, on this repository at tree `feeecc980`:

| scan | commits in scope | wall time | commits matching |
|---|---|---|---|
| incremental, HEAD-anchored, 55-commit range | 55 | 0.207 s | 0 |
| full, `--all` | 12,157 | 76.084 s | 15 |

**Condition:** measured under heavy contention — load average rose from about 19 to about 34
across the measurement, and other lanes were active. The 76.084 s figure is load-inflated and is
**not** a clean benchmark; the ratio between the two scans is the more robust reading. `--all` on
this repository includes many worktree and card branches; a project with few branches would see a
much smaller ratio (not measured).

### §3.2 The options

| | Option 1 — full `--all` every review | Option 2 — per-ref tip-set checkpoint | Option 3 — HEAD-anchored plus periodic full scan |
|---|---|---|---|
| Per-review scope | every commit reachable from any ref | commits newly reachable from any ref since the last completed scan | commits in `<last-sha>..HEAD` |
| Coverage property | full coverage of ref-reachable history on every review | intended to cover every ref-reachable commit once, when it first becomes reachable | refs not reachable from HEAD are **not** covered per review; covered only when the periodic full scan runs |
| Cost basis | the full-scan row of §3.1 on every review | scales with the number of new commits across all refs; **not measured** | the incremental row of §3.1 per review, plus the full-scan row at the chosen period |
| Checkpoint must store | nothing required for coverage | the tip of **every** ref at the last completed scan — a set of tips, not one SHA | the HEAD SHA of the last completed scan, plus a record of when the last full scan ran |
| Measured on the fixture | yes — the `--all` column of §1.1 reports SIDECELL | **no** — its behaviour on the fixture is not yet measured | per-review step: yes, it misses SIDECELL (§1.1); periodic full-scan step: the `--all` column of §1.1 |

**Option 1 — full `--all` scan on every review.** Removes the checkpoint's role in coverage. Its
cost is the full-scan figure on every review, which on this repository is the load-inflated
76.084 s reading above.

**Option 2 — record the tip of every ref, scan only what became reachable since.** An example
shape is `git log -p --all --not <previous tips> -G '<regex>'`. The checkpoint changes from one
SHA to a set of tips. Cost is expected to scale with new commits rather than total history, but
**no timing and no fixture measurement exists for this option**. Its handling of a previous tip
that no longer exists in the object store is unexamined (see `plan.md` §G); §3.4 cell ② measures
it.

**Option 3 — keep the HEAD-anchored incremental scan, correct the document.** Delete the false
equivalence claim, state plainly which refs the per-review step does not cover, and prescribe a
periodic full `--all` scan. This **narrows per-review coverage relative to Options 1 and 2**; on
this repository, 14 of the 15 matching commits (§1.2) sit outside the per-review scope until the
full scan runs. The period is not specified here and would be part of the operator's choice.

Common to Options 1 and 2 (inferred, not measured): `--all` follows refs, so commits reachable only
from a reflog, or from no ref at all, are outside both.

Options 1 and 3 were not adopted (§3). This subsection keeps all three as the record of the choice.

### §3.3 Design input — full-scan matches on this repository

**Attribution.** The classification below was reported by the lead after this SPEC was first
drafted. It is the lead's classification **by file path only** — the content of the matched
commits was never opened — and this lane has neither re-measured nor verified it.

- The 15 commits matched by the full `--all` scan (§1.2) contain 22 matching file pairs. The lead
  classified all 22 by path as examples or fixtures: a template example file, skill example and
  security-explanation documents, Go test fixtures, and a `.gitignore` whose three matching lines
  equal AWS's published documentation example key (compared without printing the value).
- Real-leak candidates by path and shape: **0**.
- The matches are mostly copies of an old Python tree.
- **Gap (named by the lead):** a path-based judgement cannot, in principle, tell a real value placed
  inside a documentation or example folder apart from an example value. The 0 above is a
  classification by path and shape, not a statement about content.

**Which steps surface these matches.** On the current history, a step whose scope reaches these
commits reports them as matches:

- Option 1 — on every review, since every review runs the full `--all` scan.
- Option 3 — at each periodic full scan.
- Option 2 — on its first completed scan, which has no previous tips to exclude and so reaches the
  full ref-reachable history (**inferred, not measured**; §3.2 records no fixture or timing
  measurement for Option 2).
- The procedure the workflow document prescribes today — its first-run full scan reaches them too.

**Follow-on decision.** Any option that runs a full-history step therefore needs a way to handle
known example values: an example-value allowlist or path exclusions. That choice is also reserved
to the operator, and is separate from the coverage-versus-cost choice above.

- Path exclusions remove the excluded paths from scan coverage, so they are themselves a coverage
  narrowing — the class of choice card t629 reserves to the operator (**inferred, not measured**).
- An allowlist keyed on exact example values narrows coverage less, because a value that differs
  from the listed examples is still reported wherever it sits (**inferred, not measured**).

This subsection records the design input as it stood before the decision. The operator has since
chosen an exact example-value allowlist with no path exclusions (§3, §3.4).

### §3.4 Adoption conditions

**Source.** The operator attached these conditions to the decision in §3 (lead session, relayed by
the lead, 2026-09-10). Option 2 has no fixture or timing measurement (§3.2), so it is adopted on the
condition that the run phase measures it first.

#### Measure-first gate

The first run-phase milestone takes the three measurements below before any commit touches either
copy of the review workflow document (REQ-009). Before the first gate command runs, the Option 2
procedure under measurement — how and when the tip set is recorded, the scan command, and the
handling of a recorded tip that no longer exists — is pinned verbatim in `progress.md` §E.2, so the
gate measures a procedure fixed in advance. The document edit may prescribe only a procedure the
gate measured.

Every cell records its commands, its output redirected to files outside this repository, and each
exit code read without a pipe, and ends in one verdict — **trustworthy** or **untrustworthy** — by
the predicate stated for it. **If any cell is untrustworthy, the run phase stops:** no further
implementation, no edit to either copy, the stop and the failing cell recorded in `progress.md`,
and a report back so the question returns to the operator (REQ-010).

**Cell ① — the side ref is reported when it first becomes reachable.** On a fixture of the §1.1
shape (outside this repository, markers assembled from fragments at runtime): a first completed
scan records the tip set on a clean history; then a commit carrying `SIDECELL` lands on a side
branch not reachable from HEAD, and a commit carrying `HEADCELL` lands on the HEAD line; then the
next scan runs.

- Trustworthy when all of these hold: the scan of the clean history before any marker prints
  0 bytes; every scan exits 0; `git merge-base --is-ancestor side HEAD` exits 1 immediately before
  the counts are read; and in the scan where the side branch first becomes reachable, the
  `HEADCELL` count is ≥ 1 and the `SIDECELL` count is ≥ 1.
- Untrustworthy otherwise — including `HEADCELL` ≥ 1 with `SIDECELL` 0 (the §1.1 defect
  reproduced), `HEADCELL` 0 (the scanner is not shown alive), a non-empty clean-history scan, or a
  non-zero scan exit.

**Cell ② — a recorded tip that no longer exists.** On the fixture: a completed scan records a tip
set that includes the tip of a branch whose commit is reachable from no other ref; that branch is
deleted, its reflog expired, and unreachable objects pruned; a commit carrying `GONECELL` then lands
on a new branch not reachable from HEAD; then the next scan runs.

- Construction check first: `git cat-file -e <the recorded tip>` exits non-zero. If it exits 0, the
  tip still exists and the cell was not constructed; it is rebuilt, and that reading is a gap, never
  a pass.
- Trustworthy when the missing tip is **detected** — an error or notice naming it is recorded in the
  scan's captured output or error stream — **and** the scan that follows reports `GONECELL` ≥ 1,
  through either acceptable behaviour:
  - (a) falling back to a full `--all` scan; or
  - (b) dropping the missing tip from the exclusion set and scanning with the remaining tips.
- Untrustworthy: an error with no follow-on scan (an unhandled error); exit 0 with `GONECELL` 0 (a
  silent skip); or no recorded detection of the missing tip.

**Cell ③ — wall time on this repository.** Two runs of the pinned procedure: a first run with no
recorded tips, then an incremental run against the tips the first run recorded. The first run
reaches the full ref-reachable history, so — like the full `--all` scan measured in
`cost-baseline.md` — it runs only after the lead approves it, and the approval is recorded before
the run. Scan output stays in files outside this repository; only counts are recorded, and no SHA,
path, or matched value of any matching commit is written into this repository (§5).

- Recorded for each run: the command; the HEAD commit it ran on; load averages immediately before
  and after; wall time; exit code; the number of recorded tips excluded; the number of commits in
  the scan's scope (the same revision arguments listed without `-p` and `-G`, counted by line); and
  the number of matching commits (a count only).
- Recorded immediately before the incremental run: `A`, the commits reachable from the recorded
  tips; `B`, the commits reachable from all refs; and `L`, the commits reachable from the recorded
  tips but from no current ref. `B − A + L` is the number of commits that became reachable since
  the tips were recorded, computed without the procedure's own revision arguments.
- Trustworthy when every field above is present for both runs, both runs exit 0, and the
  incremental run's commits-in-scope count is at most `B − A + L`. If a recorded tip is absent from
  the object store, so that `A` or `L` cannot be computed, the pair is re-taken and is a gap until
  then.
- Untrustworthy: any field missing, a non-zero exit, or an incremental scope above `B − A + L` — the
  shape of a procedure that re-scans history it already covered.
- **Judgement, for plan-audit review:** this predicate decides only whether the timing record is
  complete and internally consistent. It sets no threshold on seconds. Whether the recorded cost is
  acceptable is the operator's reading, and the measurement is taken on a shared, contended machine.

#### Handling known example values — exact allowlist

- A match is suppressed only when the text the scan's regex matched equals a listed value exactly,
  character for character (REQ-011).
- The list holds only publicly published example values — for example, a cloud provider's
  documentation example access key. This repository's own fixture and test values are not listed
  (§5).
- No path is excluded from any scan step.
- A value that differs from every listed value — by one character or entirely — is still reported,
  wherever it sits.

**Constraint — no listed value is committed literally (REQ-012).** The public documentation example
access key matches the scan regex's access-key alternative: the orchestrator assembled that value
from two fragments in a scratch file outside this repository and counted 1 match for that
alternative (2026-09-10). Written literally into either copy, a listed value would commit a
regex-matching line (REQ-007) and make the workflow document itself a match for the scan's
working-tree step. The representation must keep exact-value matching without committing text that
matches the regex. Candidates, recorded here and not chosen:

- a digest of each listed value, compared with the digest of each match;
- assembly of each listed value from fragments at scan time, so that no committed line holds a whole
  listed value.

The choice belongs to the run-phase wording milestone (`plan.md` §F M2) and is bound by this
constraint. No SPEC artifact or evidence file writes a listed value, whole or in fragments.

## §4 Constraints

- The change surface is documentation: the two `review.md` copies. No Go source, no configuration.
- [HARD] Template-First: edit the distributed copy under `internal/template/templates/` first, then
  the local copy. Read both before editing; never overwrite either copy wholesale from the other.
- [HARD] The lane does not run `make build`; the lead runs one build and embed check at batch close.
  The lane verifies the source axis only (distributed copy versus local copy).
- The scan regex itself is unchanged by this SPEC.
- The allowlist representation commits no text that matches the scan regex (§3.4, REQ-012).
- Artifact language: English.

## §5 Out of Scope

### Out of Scope — the 15 matching commits on this repository

- Examining, classifying, or remediating the 15 commits the full scan matched (§1.2). By the lead's
  classification by path and shape (§3.3), the real-leak candidates number 0. Their content was not
  examined, so a real value placed inside a documentation or example path remains a gap; closing
  that gap is outside this card.
- Recording their SHAs in this repository.

### Out of Scope — allowlisting this repository's own values

- Adding this repository's own non-public fixture or test values to the allowlist. The list holds
  publicly published example values only (§3.4).
- Consequence, inferred and not measured: full-history steps on this repository — including
  Option 2's first completed scan — will keep reporting those fixture and test matches. Their count
  is not measured, because the matched content was not opened.

### Out of Scope — code enforcement of the checkpoint

- Implementing the checkpoint, the scan, or a tip-set store in Go. The checkpoint remains a
  documented procedure.
- Observing whether, or how often, any agent actually runs this section.

### Out of Scope — the scan pattern and other scanners

- Changing, widening, or testing the three regex alternatives. Only the PEM-header alternative was
  exercised by the fixture; the coverage defect concerns the commit range, not the pattern.
- Working-tree-only scanners and the dependency vulnerability scan in the same workflow.

### Out of Scope — build and embed verification

- `make build`, the embed check, and any binary-level verification — owned by the lead at batch
  close.
