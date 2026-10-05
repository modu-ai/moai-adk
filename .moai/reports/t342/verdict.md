# t342 — SPEC-MOVING-REF-GUARD-001 run-phase verdict

Card: t342 · Tier M · worktree `.claude/worktrees/t342` · branch `WT-moving-ref-guard`
Lane: lane-8. Authored by the lane orchestrator from evidence read, not from agent claims.

## Baseline attribution

| Anchor | Value | Provenance |
|---|---|---|
| `BASELINE_SHA` | `d566ecc7511e1954e3aeb1dff3a60afa5be1089b` | `git rev-parse origin/develop` at pre-flight (plan.md §C step 4), frozen 2026-08-28 |
| `MERGE_BASELINE_SHA` | `5e194bba27c146d8c2157d92b4a3fb3995919ff0` | `git merge-base HEAD origin/develop`, frozen at the document round (plan.md §C step 5) |
| HEAD at verdict | `3b71281e7` | `git rev-parse --short HEAD`, this tree |

Neither anchor was re-resolved after freezing. `origin/develop` advanced to `5e194bba2` during M1 and
was merged in (`cb90e8a86`); the anchors deliberately did not move with it. Re-resolving an anchor
because mainline advanced is the defect this SPEC exists to detect.

## Claim

M1 through M5 are implemented, and a six-item document round repaired four acceptance criteria. The
deliverable is a `warning`-severity lint rule (`MovingRefUnpinned`) **plus** the anchor-or-subject
predicate that decides whether a finding should be acted on by pinning — the card's [HARD] second
half.

## Evidence

Commits on this branch beyond `MERGE_BASELINE_SHA`:

| Commit | Milestone |
|---|---|
| `91ea212a8` | M1 doctrine + M2 marker surface |
| `1ed8a9997` | M3 detector |
| `81e33c81e` | M4 corpus triage (classification only) |
| `2df13080d` | M5 template mirror, neutralized |
| `3b71281e7` | document round — four criteria repaired, Q0 deferral recorded |

Orchestrator-side verification, re-executed rather than taken from the agents' reports:

| # | Check | Command | Observed |
|---|---|---|---|
| V1 | Tests green, uncached | `go test ./internal/spec/ -run TestMovingRef -count=1 -v \| grep -c '^--- PASS'` | `9` |
| V2 | Rule output over the live corpus | `go run ./cmd/moai spec lint \| grep -c MovingRefUnpinned` | `107` |
| V3 | Doctrine limits | `grep -c 'L[1-7] —' <doctrine>` | `7` |
| V4 | AC-MRG-011's new key | `grep -c 'the ANCHOR branch of the predicate is unvalidated' <doctrine>` | `1` |
| V5 | Mirror completeness | limits / tests / instances in the mirrored copy | `7` / `4` / `5` |
| V6 | Mirror neutrality | SPEC-ID, AC/REQ, 40-hex, date, `CLAUDE.local` scan of the mirror | `0` |
| V7 | Mirror is live | `strings bin/moai \| grep -c 'anchor-or-subject'` | `2` |
| V8 | Nothing added outside this SPEC's dir | `git diff --name-only $MERGE_BASELINE_SHA..HEAD -- .moai/specs \| grep -v SPEC-MOVING-REF-GUARD-001` | empty |
| V9 | Working tree | `git status --short` | empty |

Two mutations were planted by the orchestrator itself, independently of the implementing agents:

- **M-014, the mutant the Definition of Done names** — the claim-marker conjunct replaced by a
  tautology. Observed: `TestMovingRef_NegativeControlOnClaimConjunct` red **alone**, the other eight
  criteria green. That is the DoD's stated property, measured rather than argued. Reverted; green
  restored.
- **The body-only mutant, AC-MRG-009's separation claim** — file iteration restricted to the
  document's own basename. Observed after the document round's fixture re-placement:
  `TestMovingRef_ReadsSiblingArtifacts` red while `TestMovingRef_FiresOnUnpinnedAnchor` stayed
  **green**. Before re-placement the same mutant took seven criteria down together, AC-MRG-001 among
  them. Reverted; green restored.

## What this card actually found

Three acceptance criteria were vacuous or false, and every one of them was true when written.

1. **AC-MRG-001 could not fail as first implemented.** The `origin/mainx` mutation did not turn it
   red: the §B.3 alternation `origin/(main|develop|HEAD)` has no right boundary, so `origin/mainx`
   matched on the prefix. Fixed by adding `\b`. The boundary's cost was measured rather than assumed
   — 42 findings with it and 42 without, and the same command reproduces §B.3's filter-4 figure of 42
   in this tree. Had the mutation not been planted, the criterion would have passed green while
   guarding nothing.
2. **AC-MRG-011's second conjunct guarded nothing.** With L7 deleted the count half correctly dropped
   7 → 6, while `grep -c 'ANCHOR'` returned 9 — `ANCHOR` is one of the predicate's two class names,
   so it is present in any doctrine that publishes the predicate at all. Replaced with a key on L7's
   distinguishing phrase, and the replacement was re-observed under the same mutation.
3. **AC-MRG-010's decider was anchored to a range spanning mainline.** `BASELINE_SHA..HEAD` returned
   four foreign paths attributable to cards t322 and t303, absorbed by routine merges — a
   wrong-reason red that no work on this branch could clear, and one that cannot tell a real bulk-pin
   from a clean merge. Resolved by freezing `MERGE_BASELINE_SHA` as a second recorded literal at
   pre-flight, rather than by computing a merge-base at read time.

The third is this SPEC reproducing its own subject for a fourth time, and the plan now says so: the
frozen merge-base goes stale on the next mainline absorption exactly as the first anchor did, so
plan.md §C step 5 carries a [HARD] obligation to re-record it with a date before AC-MRG-010 is
decided again. That is limit L6 acting on this SPEC's own evidence, handled the way the doctrine
prescribes — the command is the criterion, the value is a dated reference.

## Corpus triage — classification only

107 findings, 97 outside this SPEC's own directory. Classification: **48 anchor / 56 subject-S1 /
2 subject-S2 / 1 already-frozen / 0 unclassifiable**. Externally: 48 / 48 / 0 / 1.

No SPEC artifact outside this SPEC's own directory was edited, no marker was added anywhere, and no
SHA was pinned anywhere. Bulk-pinning the corpus is the card's named dominant failure mode;
remediating a finding belongs to its owning SPEC.

The two external lines §B.3 named as correctly-unpinned subject claims (`AC-COORD-016`,
`REQ-LB-006`) are both flagged by the rule. That is the expected L2/L4 outcome — the detector reads
shape, never subject — and neither must be pinned. The marker is what they are for.

External S2 occupancy is **0**, corroborating §B.7 from the rule's own output rather than from grep.
The only two S2 lines sit inside this SPEC's directory, both R4-form, exactly as §B.7 predicted: the
class is populated prospectively, by the R4 remediations the doctrine produces.

## Deferred — REQ-MRG-010 and AC-MRG-013 (Q0, option C)

Deferred by operator decision.

**What:** the R4-form lint exclusion, its acceptance criterion, and both counter-mutations (CM-1
positional, CM-2 command-token) — unimplemented and unrun.

**Why:** §B.7 measured R4's reachable class as 0 of 42 candidate lines on two independent probes, and
M4's rule output corroborated it at 0 external occupants. With no occupants the exclusion can only
over-exempt, never under-exempt, so every error available to it is a bypass of the shape the iter-1
audit measured — a fetch-verb-keyed exclusion passing all thirteen criteria while silencing 76 of 117
real unpinned divergence lines.

**Resume condition:** reconsider when the R4 form is actually observed in the corpus. The class has
begun filling from this card's own output. Meanwhile early R4-form lines are silenced with the R3
marker. R4 remains in the finding message and in the doctrine — only its lint exclusion is deferred.

The constraint that survives the deferral: when the exclusion is eventually built it MUST key on
imperative structure, never on a command token. A token key is forgeable by construction.

No follow-up card was issued from this lane; card issuance is the operator's act.

## Gaps — explicitly not observed

- `golangci-lint`, cross-platform `GOOS=... go vet`, and `go test ./...` were not run. Package-scoped
  verification only, per this repository's local-full-suite prohibition; the full-suite verdict is
  CI's on `origin/develop`.
- The rule's per-directory `*.md` re-read has no before/after timing. The corpus run did not finish
  within 120s (compile time included), so its cost is unmeasured rather than known-acceptable.
- The 107 classifications are one reader's judgment — limit L2 acting on the triage rather than on
  the detector. The 11 rows where a second reader could reasonably differ are named in the triage
  record §6; `0 unclassifiable` means every row has a stated reading, not that no row is doubtful.
- The anchor branch carries 48 author-classified rows on top of L7's existing validation gap, and
  should be weighted less confidently than the subject rows.
- `.moai/reports/**` is not scanned (limit L5) — the carrier the R4-motivating instance lives on.
- A template file's correctness for a downstream user cannot be observed from this repository.
- One rule-behaviour candidate, unaddressed and outside the milestones' scope:
  `frozenBaselinePattern` matches a variable **reference** but not an **assignment**, so a line that
  *teaches* R2 is flagged while the sibling line that *applies* it is correctly exempt. Reproduced
  at HEAD `a77f29a6b`, both directions:

  ```
  $ sed -n '36p' .moai/specs/SPEC-DESIGN-MOAIWEBV2-002/plan.md
  - [ ] Record `BASELINE_SHA=$(git rev-parse origin/main)` BEFORE the first run-phase commit/push …
  $ sed -n '36p' .moai/specs/SPEC-DESIGN-MOAIWEBV2-002/plan.md | grep -cE '\$\{?[A-Z_]*BASELINE[A-Z_]*\}?'
  0

  $ sed -n '14p' .moai/specs/SPEC-DESIGN-MOAIWEBV2-002/acceptance.md
  `$BASELINE_SHA` below is the pre-flight-recorded origin/main SHA (`BASELINE_SHA=$(git rev-parse origin/main)`, …
  $ sed -n '14p' .moai/specs/SPEC-DESIGN-MOAIWEBV2-002/acceptance.md | grep -cE '\$\{?[A-Z_]*BASELINE[A-Z_]*\}?'
  1
  ```

  And the rule's own output over the live corpus confirms which sibling it flags:

  ```
  $ go run ./cmd/moai spec lint | grep MovingRefUnpinned | grep MOAIWEBV2-002
  WARNING   MovingRefUnpinned   …/SPEC-DESIGN-MOAIWEBV2-002/plan.md       35
  WARNING   MovingRefUnpinned   …/SPEC-DESIGN-MOAIWEBV2-002/plan.md       36
  WARNING   MovingRefUnpinned   …/SPEC-DESIGN-MOAIWEBV2-002/progress.md   58
  ```

  `acceptance.md:14` — the line that actually applies R2 — is absent from the output, exempted by its
  `$BASELINE_SHA` reference. `plan.md:36`, which records the R2 capture command, carries `BASELINE`
  with no `$` before it (`$(git …` puts a `(` after the `$`), so the pattern misses it. The
  consequence is that the guard warns on the sentence teaching its own recommended remedy. Whether
  the pattern should also match the assignment form is a rule-behaviour decision, not a triage call,
  and is left to the operator.

## Residual risk

- The eight fixtures are synthetic and were authored by the same actor as the rule, so a shared
  misreading of REQ-MRG-001 would stay invisible until the corpus disagrees.
- `shaPinPattern` is lexical: any hex-shaped word on the line exempts it, so a claim can be exempted
  by an anchor that anchors nothing.
- The divergence branch's figure heuristic accepts `:` and `=` as citation markers behind a single
  negative control.
- The triage's most contestable call: instruction lines — a command plus a pass condition, 35
  external rows — were read as S1, on the ground that an instruction does not *assert* a current
  state. A reviewer reading them as S2 moves 35 rows from R3 to R4 and reopens the R4 scope question.
- The doctrine adds roughly 14.8 KB to an always-loaded file, paid by every session including those
  that never write a moving-ref claim. spec.md §C and REQ-MRG-005 specify exactly that placement, so
  it is a design decision rather than a defect — but a reviewer preferring a `paths:`-scoped
  companion would be disagreeing with the SPEC, not with the implementation.

## Process note

One `git merge` failed with `fatal: Unable to write index` and succeeded on retry, after the working
tree was confirmed clean and `MERGE_HEAD` absent. Cause established by the lead: a stale zero-byte
`index.lock` in the primary checkout, since removed.
