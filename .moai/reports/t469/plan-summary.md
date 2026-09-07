# t469 Plan Summary — SPEC-HOOK-WIRING-DRIFT-001 v0.4.0 amendment (AC-HWD-015 strip-aware mirror identity)

Date: 2026-09-03 · Tree: worktree `.claude/worktrees/t469`, branch `WT-achwd-strip-exempt`, HEAD `a1d7598ac`
Phase: plan (manager-spec) · Deliverable: plan-phase artifacts for the in-place amendment; run phase applies

## Claim

1. AC-HWD-015's "diff reports no difference" is FALSE as written for
   `agent-common-protocol-reference.md`, and cannot be made true without
   violating REQ-HWD-014 (template-side neutrality stripping).
2. The divergence is a single line (275), wholly the neutrality strip of
   `(SPEC-SYNC-PARALLEL-DOCS-001 A9)`.
3. The divergence pre-exists at `a239cf050` (ancestor of HEAD) — it coexists
   with this SPEC's authoring rather than post-dating it.
4. The conflict is a class property: 47 of the managed-root local/template
   pairs differ; 24 of those 47 carry forbidden-class tokens in their diff
   lines.
5. Amendment shape: in-place v0.4.0, scoped to the three M3 files, with ONE
   machine-checkable command (single plain `perl` invocation, exit 0/1) that
   normalizes forbidden-class tokens on both sides before comparing.

## Evidence (commands run in this plan phase, verbatim outputs)

All commands run from `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t469` at HEAD `a1d7598ac`.

**E1 — per-file mirror identity (items 1-2):**

```
$ for f in .claude/rules/moai/development/hook-independence.md \
>   .claude/rules/moai/core/agent-common-protocol.md \
>   .claude/rules/moai/core/agent-common-protocol-reference.md; do
>   diff -q "$f" "internal/template/templates/$f"
> done
Only in outputs (diff -q is silent on identity):
  IDENTICAL: hook-independence.md
  IDENTICAL: agent-common-protocol.md
  (agent-common-protocol-reference.md: differs — full diff below)
```

The full diff for the third file:

```
$ diff .claude/rules/moai/core/agent-common-protocol-reference.md \
>      internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md
275c275
< > Relocated from `agent-common-protocol.md` § Parallel Execution → Attributable diff-check doctrinal switch to keep the always-loaded file within its size budget. The switch rule, the three match conditions, the four mismatch names, and the never-silent-skip boundary remain inline there (SPEC-SYNC-PARALLEL-DOCS-001 A9).
---
> > Relocated from `agent-common-protocol.md` § Parallel Execution → Attributable diff-check doctrinal switch to keep the always-loaded file within its size budget. The switch rule, the three match conditions, the four mismatch names, and the never-silent-skip boundary remain inline there.
rc=1
```

**E2 — pre-existence at `a239cf050` (item 3):** worktree-guard-safe split form
(blobs written to `/tmp`, then diffed; compound/process-substitution form
rejected by the session guard):

```
$ git show a239cf050:.claude/rules/moai/core/agent-common-protocol-reference.md > /tmp/t469-local-ref.md
$ git show a239cf050:internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md > /tmp/t469-tmpl-ref.md
$ git merge-base --is-ancestor a239cf050 HEAD && echo "ANCESTOR-OK"
ANCESTOR-OK
$ diff /tmp/t469-local-ref.md /tmp/t469-tmpl-ref.md
275c275
< > Relocated from `agent-common-protocol.md` § Parallel Execution → ... remain inline there (SPEC-SYNC-PARALLEL-DOCS-001 A9).
---
> > Relocated from `agent-common-protocol.md` § Parallel Execution → ... remain inline there.
rc=1
```

Same `275c275` line pair at both revisions.

**E3 — class scope (item 4):** script `.moai/state/verify/t469/pairs.sh` +
`.moai/state/verify/t469/classes.sh` (gitignored scratch, re-runnable; outputs
`pairs.txt`, `classes.txt` alongside):

```
$ bash .moai/state/verify/t469/pairs.sh
== DIFF count: 47
== NOTEMPLATE count: 15
$ bash .moai/state/verify/t469/classes.sh
== pairs with forbidden classes:       24
```

(15 NOTEMPLATE = local-only files under the roots with no template twin —
out of scope, not mirror pairs. The 24 classified pairs list, headed by
`agent-common-protocol-reference.md => SPEC-ID`, is in `classes.txt`.)

**E4 — amended check passes the real tree:**

```
$ perl -e 'local $/; my $rc=0; for my $f (@ARGV){ open my $L,"<",$f or die "open $f: $!"; open my $T,"<","internal/template/templates/$f" or die "open tmpl $f: $!"; my ($a,$b)=(<$L>,<$T>); for ($a,$b){ s/\s*\((?:SPEC-[A-Z][A-Z0-9]*(?:-[A-Z0-9]+)*|REQ-[A-Z][A-Z0-9]*)-[0-9]{3}(?:\s+[A-Z][0-9]{1,2})?\)//g; s/\s*\b(?:SPEC-[A-Z][A-Z0-9]*(?:-[A-Z0-9]+)*|REQ-[A-Z][A-Z0-9]*)-[0-9]{3}\b//g; s/\s*\bt[0-9]{2,3}\b//g; s/\s*\b[0-9a-f]{40}\b//g; } if ($a ne $b){ print "MISMATCH $f\n"; $rc=1; } } exit $rc' \
  .claude/rules/moai/development/hook-independence.md \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/core/agent-common-protocol-reference.md
(no output)
exit=0
```

**E5 — mutants constructed and executed** (fixtures under
`.moai/state/verify/t469/fixture/`, removed after measurement):

- Mutant (i) — non-token editorial text inserted on the token-bearing line of
  the LOCAL copy → `MISMATCH .claude/rules/moai/core/agent-common-protocol-reference.md`, `mutantB-exit=1`
- Mutant (ii) — forbidden token inserted into the TEMPLATE copy → check exits 0
  (absorbed by design, direction-agnostic on token runs); the compensating
  control observed: `grep -cE 'SPEC-[A-Z]' <mutant template copy>` → `7` (the
  neutrality scan AC-HWD-016 catches what the mirror check absorbs)
- Mutant (iii) — mirror-only edit → exit 1 (same mechanism as (i), the unamended
  AC's own mutant class, retained)

**E6 — real template copy neutrality baseline (AC-HWD-016 precondition):**

```
$ grep -cE 'SPEC-[A-Z]|REQ-[A-Z]|\bt[0-9]{2,3}\b' internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md
0
grep-exit=1 (1 = zero hits = clean)
```

## Baseline-attribution

- All measurements: this plan phase, this worktree (`t469`), HEAD `a1d7598ac`
  (= `origin/develop` `d592b0551` + `WT-hook-wiring-drift` `3a3f51e83`),
  2026-09-03.
- The lane's four measurements were independently re-derived here — none is
  carried over. Values match the lane's report in all four cases.
- Scripts under `.moai/state/verify/t469/` are gitignored scratch; the commands
  in this file are the durable record and are re-runnable as written.

## Gaps (explicitly not observed)

- The `internal date` forbidden class is not normalized by the amended check —
  it has no agreed regex (AC-HWD-016 scans it only in prose terms). A future
  date-class strip would produce a neutrality-mandated MISMATCH.
- The mutant space is three constructed mutants, not exhaustion — stated as
  such in the amended AC's `Mutant:` note.
- `moai spec lint` was not run against a v0.4.0-spec frontmatter (the frontmatter
  bump is run-phase work); the plan assumes the existing schema continues to
  validate — to be confirmed by the plan-audit and/or run phase.
- Normalization over-absorption risk on token-bearing lines (mutant ii
  generalized) is accepted at three-file scope, not eliminated — see plan.md
  §I.5.

## Residual-risk (focus for the plan-auditor)

1. Whether the amended AC's normalization regex (parenthetical-first, then bare
   tokens) is tight enough — a local-side parenthetical differing from a
   template-side parenthetical in non-token content is invisible to the check.
2. Whether REQ-HWD-013's reworded "identical after normalization" leaves any
   residual contradiction with REQ-HWD-014 for classes beyond the four regex
   classes (internal date).
3. Whether the amendment's Out of Scope entry (fleet-wide strip-aware
   invariant) correctly closes the 24-pair class finding or whether the auditor
   requires a follow-up card reference.
