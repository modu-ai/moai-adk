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

**Residual risk:** the exemption predicate (`spec.md` §D.1) is a judgment procedure. Its four tests
were validated against five real instances, all of which it classifies correctly — but five is a
small validation set, and **all five are SUBJECT-class** (`spec.md` §D.3 tally). The ANCHOR side of
the predicate rests entirely on the corpus lines of §B.3, which were classified by reading rather
than by applying the tests in anger. A disposition outside the anchor/subject dichotomy may still
exist; the v0.1.0 → v0.2.0 revision is itself evidence that the classification can be incomplete —
instance 3 was classified ANCHOR at v0.1.0 and that reading was wrong. `spec.md` §H carries the open
questions rather than closing them.

**Status transition:** `(none) → draft`, emitted across all four plan-phase artifacts.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
