# t469 lane recon — AC-HWD-015 vs template neutrality

Lane-session measurements on branch `WT-achwd-strip-exempt`, taken BEFORE manager-spec plan
authoring. Each measurement names its command; all commands ran in this worktree, this session.

## R1 — AC-HWD-015's third file diverges by exactly one stripped token

Command: `diff .claude/rules/moai/core/agent-common-protocol-reference.md internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md`

Observed (single hunk, line 275):

```
275c275
< > Relocated from `agent-common-protocol.md` § Parallel Execution → Attributable diff-check doctrinal switch ... remain inline there (SPEC-SYNC-PARALLEL-DOCS-001 A9).
---
> > Relocated from `agent-common-protocol.md` § Parallel Execution → Attributable diff-check doctrinal switch ... remain inline there.
```

The other two AC-HWD-015 files (`hook-independence.md`, `agent-common-protocol.md`) diff clean
(`diff -q` rc=0 each). AC-HWD-015's "diff reports no difference for every file M3 touches" is
therefore FALSE as written, and the only divergence is a neutrality strip (§25.1 forbidden class:
SPEC ID).

## R2 — pre-existence at a239cf050 (card t449 merge, ancestor of HEAD)

Commands (this tree): `git show a239cf050:<local path>` and
`git show a239cf050:internal/template/templates/<local path>` for all three files, diffed pairwise.

Observed: `agent-common-protocol-reference.md` differs at the SAME line 275 in the SAME
strip shape; `hook-independence.md` and `agent-common-protocol.md` identical at that commit too.
The conflict pre-dates the current neutrality iteration — it was latent from at least a239cf050,
and AC-HWD-015 was authored (t216, 2026-08-24) against a pair that already could not satisfy it.

## R3 — class scope: 24 of 47 differing managed pairs carry forbidden-class tokens

Commands: listed 159 local files under `.claude/{rules,agents,skills,commands,output-styles}/moai/`
(`find … -type f`), diffed each against its `internal/template/templates/` twin; 47 differ; grepped
each differing pair's diff lines for `SPEC-[A-Z0-9]+-[A-Z0-9-]+|REQ-[A-Z0-9]+-[0-9]+|\bt[0-9]{3}\b|[0-9a-f]{40}`.

Observed: 24 pairs carry forbidden-class tokens in their diff lines (full list retained in the
session transcript; sample: `agents/moai/builder-harness.md`, `plan-auditor.md`,
`manager-spec.md`, `core/verification-claim-integrity.md`, `core/settings-management.md`,
`workflow/verification-batch-pattern.md`, `skills/moai/workflows/sync.md` …). The conflict is a
property of the strip-exempt pair shape, not a one-file accident.

## R4 — dependency state of the card branch

- `SPEC-HOOK-WIRING-DRIFT-001` exists on NO develop head: not on `refs/heads/develop`
  (4e4607abe), not on `origin/develop` (d592b0551) — only on `WT-hook-wiring-drift`
  (local tip 3a3f51e83; origin's copy 8aa96bfb1 is its ancestor, i.e. local is ahead).
- Per the dependency-merge rule, merged `WT-hook-wiring-drift` into the card branch:
  `git merge --no-ff WT-hook-wiring-drift` → merge commit a1d7598ac, 21 commits absorbed, zero
  conflicts, working tree clean. If t216's own window merges first, this merge trivializes; if
  this card merges first, t216's content lands through it. Either order is consistent.

## Scope note

t216's `status: completed` premise holds on the branch. No retroactive re-judgment is in scope
(lead [HARD]); this card only amends AC-HWD-015's wording going forward.
