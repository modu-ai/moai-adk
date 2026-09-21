# SPEC-AUDIT-EXPORT-CLAUSE-001 — acceptance criteria

> Every AC names a command and the output that closes it. Commands run from the
> repository root of the tree under test, as a single invocation each — no
> command substitution, no shell-variable exit capture, no writes to the tree
> (plan §E). A zero-count result is admissible only when the AC's paired positive
> control returned non-zero **in the same run** (§D.3).

---

## §A Given-When-Then scenarios

### AC-AEC-001 — the negation and its explanatory block are gone

**Given** the `.gitignore` withdrawal specified in plan §A.2,
**When** the file is searched for any negation re-including a card-report
artifact, and for the comment block that introduced it,
**Then** neither is found.

```bash
grep -nE '^!\.moai/reports/\*/' .gitignore ; echo "negation_exit=$?"
grep -nc 'Narrow exception (card t1039)' .gitignore ; echo "block_exit=$?"
```

Expected: no output from either, `negation_exit=1`, `block_exit=1`.

Paired control — the directive block and its blanket, which this card does **not**
touch, MUST still match, proving the greps reach the file:

```bash
grep -c 'evidence for lead verdicts stays on disk' .gitignore
grep -cE '^\.moai/reports/\*$' .gitignore
```

Expected: both `1`.

Maps REQ-AEC-001

### AC-AEC-002 — the blanket rule is in effect again for a new path

**Given** that the withdrawal restores `.moai/reports/*` to full effect,
**When** two names under a card report directory are tested — the one the
negation used to re-include and one it never did —
**Then** both report as ignored, and the plan-audit carve-out still behaves as
before.

```bash
git check-ignore --no-index .moai/reports/ZZAC002/verdict.md && echo IGNORED || echo NOT-IGNORED
git check-ignore --no-index .moai/reports/ZZAC002/plan-audit.md && echo IGNORED || echo NOT-IGNORED
git check-ignore --no-index .moai/reports/plan-audit/.gitkeep && echo IGNORED || echo NOT-IGNORED
```

Expected, in order: `IGNORED`, `IGNORED`, `NOT-IGNORED`.

Paired control — a path outside the reports tree MUST report `NOT-IGNORED`,
proving the probe can produce that answer at all:

```bash
git check-ignore --no-index README.md && echo IGNORED || echo NOT-IGNORED
```

Expected: `NOT-IGNORED`.

**The plain form is load-bearing; `-v` is a wrong instrument here.** Measured
before the withdrawal, with the negation live: `git check-ignore -v --no-index`
on `verdict.md` exited **0** while the plain form exited **1** — the verbose form
reports that a pattern matched and exits 0 even when the matching pattern is a
negation, i.e. when the path is not ignored. Writing this criterion with `-v`
would pass on a tree where the negation survived. No probe file is created: the
`--no-index` form answers for a path that does not exist, so this AC writes
nothing and has no cleanup step.

Maps REQ-AEC-001, REQ-AEC-002

### AC-AEC-003 — the already-tracked population is untouched

**Given** that withdrawal binds only files created afterwards (SPEC §A.3),
**When** the tracked set and the remote-present set are measured with the
anchored criterion on both sides,
**Then** the tracked set is unchanged from the pre-implementation measurement and
nothing under `.moai/reports/` appears as an index change.

```bash
git ls-files '.moai/reports/*/verdict.md' | wc -l
git ls-tree -r --name-only origin/develop -- .moai/reports | grep -cE '^\.moai/reports/[^/]+/verdict\.md$'
git status --porcelain -- .moai/reports ; echo "status_exit=$?"
```

Expected: `12`; `10`; no `D`, `R`, or staged-deletion line in the third output.

**The two-way difference is the closing evidence, not the counts.** Report both
directions explicitly; a difference of 1 in the counts would have concealed the
true shape, which is 2 tracked-not-remote and 0 remote-only:

```bash
git ls-files '.moai/reports/*/verdict.md' | sort
git ls-tree -r --name-only origin/develop -- .moai/reports | grep -E '^\.moai/reports/[^/]+/verdict\.md$' | sort
```

Expected: the first list contains `.moai/reports/t1039/verdict.md` and
`.moai/reports/t1048/verdict.md`, which the second does not; every other entry of
the first appears in the second; the second contains nothing absent from the
first.

**A substring criterion is a FAIL, not a near-pass.** `grep 'verdict.md'` on the
remote side also matches `.moai/reports/t965/plan-audit-verdict.md` and returns
11. Both sides use the anchored form.

Maps REQ-AEC-002, REQ-AEC-003

### AC-AEC-004 — no surface designates a card-report artifact as tracked

**Given** the scope table in plan §C,
**When** all thirteen surfaces are swept with the markup-insensitive pattern that
located them,
**Then** no match remains.

```bash
grep -rnE 'tracked\*{0,2} (citation target|verdict file|path)' \
  .claude/rules .claude/agents .claude/skills .moai/docs \
  internal/template/templates/.claude internal/template/templates/.moai
echo "exit=$?"
```

Expected: no output, `exit=1`.

Paired control — the replacement phrasing MUST match in the same trees, proving
the sweep reaches the repaired files rather than missing them:

```bash
grep -rlE 'local\*{0,2} record' \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/core/agent-common-protocol-reference.md \
  .claude/agents/moai/manager-lead.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md \
  internal/template/templates/.claude/agents/moai/manager-lead.md | wc -l
```

Expected: `6`.

> **The pattern carries `path` as a third alternative deliberately.** The v0.2.0
> draft measured these surfaces as `tracked** path` and excluded them; the
> `develop` absorb rewrote that phrase out of existence, so a sweep for it now
> returns a clean zero on unrepaired files (SPEC §A.6). Enumerating the retired
> phrasing alongside the current one costs nothing and stops the next absorb from
> producing the same false zero in reverse.

Maps REQ-AEC-004

### AC-AEC-005 — the withdrawn sentence is gone from every auditor copy

**Given** the four agent definitions enumerated in SPEC §A.4,
**When** they are searched for the removed sentence,
**Then** every file reports zero.

```bash
grep -rc 'Never write the verdict to a gitignored location' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md
```

Expected: four lines, each ending `:0`.

Paired control — a string known to remain present MUST report non-zero in every
file, proving the instrument can find text in them at all:

```bash
grep -rc 'Export mandate' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md
```

Expected: four lines, each ending `:1` or higher. `Export mandate` is chosen
because the pattern crosses no emphasis marker: the clause carries `**bold**`
runs, and a pattern spanning a `**` delimiter matches nothing regardless of the
file's content.

Maps REQ-AEC-004, REQ-AEC-008

### AC-AEC-006 — the local-by-design statement is present in every auditor copy

**Given** the SPEC §C.1 replacement tail,
**When** the four files are searched for the statement the repair adds,
**Then** every file reports at least one match.

```bash
grep -rc 'local by design' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md
```

Expected: four lines, each ending `:1` or higher.

Maps REQ-AEC-005

### AC-AEC-007 — no repaired wording names a remote-placing act

**Given** REQ-AEC-006,
**When** every file this card changes is searched for a forced stage or any other
verb that would place a card-report artifact on the remote,
**Then** none is found.

```bash
grep -rnE 'add -f|--force.*reports|push.*verdict' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  .claude/agents/moai/manager-lead.md \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/core/agent-common-protocol-reference.md \
  .moai/docs/audit-artifact-convention.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/manager-lead.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md \
  internal/template/templates/.moai/docs/audit-artifact-convention.md
echo "exit=$?"
```

Expected: no output, `exit=1`.

Paired control — the prohibition the replacement states instead MUST match,
proving the sweep reaches the repaired files:

```bash
grep -rlc 'do not force it into the tree' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md | wc -l
```

Expected: `4`.

> **This criterion exists because the defect it guards dissolved rather than being
> fixed.** The iteration-2 audit found that the permission clause (*"which this
> mandate permits for exactly this destination"*) had no criterion, so a wording
> omitting it would pass every criterion. Under the operator's ruling that clause
> is withdrawn, so there is nothing left to test for presence — and the residual
> risk inverts: the hazard is now that the verb **survives** a partial edit. The
> criterion is therefore a prohibition test over every changed file, not a
> presence test over four.

Maps REQ-AEC-006

### AC-AEC-008 — the FORBIDDEN prohibition survives and rests on a property that still holds

**Given** that each auditor copy carries its own FORBIDDEN wording, and that the
convention's § Where bullet justified the directory by the ignore rules,
**When** each surface is inspected,
**Then** the per-copy prohibition is intact and the convention's justification no
longer rests on being gitignored, while retaining its `.gitignore` pointer.

```bash
grep -c 'FORBIDDEN' .claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md
grep -c 'plan-audit/` is FORBIDDEN' .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md
awk '/^## Where/,/^## When/' .moai/docs/audit-artifact-convention.md | \
  grep -nE 'deliberately.{0,3}\n?.{0,20}gitignored|`.gitignore` comment|never read as a card'
```

Expected: every count non-zero; the third output contains the `.gitignore` comment
pointer and the read-based distinguishing property, and does **not** contain
`deliberately` adjacent to `gitignored`.

Maps REQ-AEC-007

### AC-AEC-009 — the export obligation is not weakened

**Given** REQ-AEC-008,
**When** the four auditor copies are searched,
**Then** the incomplete-audit declaration still stands in each.

```bash
grep -rc 'incomplete audit' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md
```

Expected: four non-zero lines.

Maps REQ-AEC-008

### AC-AEC-010 — the convention document's two false claims are withdrawn and its check still holds

**Given** § Committing and § What makes the convention stick,
**When** both sections are read,
**Then** neither claims the artifact is tracked or reaches the integration branch,
the disposal hazard survives with its mitigation, and the mechanical check names
presence on disk rather than branch reachability.

```bash
awk '/^## Committing/,/^## What makes/' .moai/docs/audit-artifact-convention.md
awk '/^## What makes/,/^## Cross-references/' .moai/docs/audit-artifact-convention.md
```

Expected, read together: the first output contains `is a local file` and
`disposal hazard` and does **not** contain `Audit artifacts are tracked files` or
`reach the integration branch`; the second contains `ls .moai/reports/<card-id>/`
and does **not** name `git ls-files --error-unmatch` or any other
branch-reachability check.

```bash
grep -c 'Audit artifacts are tracked files' .moai/docs/audit-artifact-convention.md
grep -c 'ls-files --error-unmatch' .moai/docs/audit-artifact-convention.md
grep -c 'disposal hazard' .moai/docs/audit-artifact-convention.md
```

Expected: `0`, `0`, non-zero. The third is the paired control: it is a sentence
this card preserves, so a zero there means the greps did not reach the file and
the two zeros above are void.

Maps REQ-AEC-004, REQ-AEC-013

### AC-AEC-011 — the two convention copies are byte-identical

**Given** the local convention doc and its template mirror,
**When** they are compared,
**Then** they are byte-identical.

```bash
cmp internal/template/templates/.moai/docs/audit-artifact-convention.md \
    .moai/docs/audit-artifact-convention.md && echo BYTE-IDENTICAL
```

Expected: `BYTE-IDENTICAL`, exit 0.

Maps REQ-AEC-009

### AC-AEC-012 — the emitted codex layer matches its regenerated source

**Given** that three C2 agent copies were modified,
**When** the emission drift check runs,
**Then** it exits 0 without regenerating anything, and the emitted layer carries
neither withdrawn text.

```bash
make agents-emit-check ; echo "exit=$?"
grep -rc 'Never write the verdict to a gitignored location' \
  internal/template/templates/.codex/agents/moai/plan-auditor.toml \
  internal/template/templates/.codex/agents/moai/sync-auditor.toml
grep -rcE 'tracked\*{0,2} verdict file' \
  internal/template/templates/.codex/agents/moai/manager-lead.toml
```

Expected: `exit=0`; the second command's two lines each ending `:0`; the third
ending `:0`.

Paired control — the three TOMLs MUST still carry text the emission preserves,
proving the greps reached them:

```bash
grep -rc 'Export mandate' \
  internal/template/templates/.codex/agents/moai/plan-auditor.toml \
  internal/template/templates/.codex/agents/moai/sync-auditor.toml
grep -c 'Context-Folding' internal/template/templates/.codex/agents/moai/manager-lead.toml
```

Expected: all non-zero. Measured before implementation, the three probes return
`:1`, `:1`, `:2` — so each has an observed red and a zero here is a real change.

Maps REQ-AEC-011

### AC-AEC-013 — template neutrality on every template-tree edit

**Given** the six files changed under `internal/template/templates/`,
**When** **the lines this card adds** are swept for the forbidden content classes,
**Then** no match is found in any class.

```bash
git diff -U0 develop...HEAD -- internal/template/templates \
  | grep '^+' | grep -vE '^\+\+\+' \
  | grep -E 'SPEC-[A-Z][A-Z0-9-]*-[0-9]{3}|REQ-[A-Z]|CLAUDE\.local|/Users/|\b20[0-9]{2}-[0-9]{2}-[0-9]{2}\b|\b[0-9a-f]{9,40}\b'
echo "exit=$?"
```

Expected: no output, `exit=1`. **This is the sole closing command for this AC.**

Paired control — the same added-line extraction, filtered for a token the repair
introduces, MUST match; otherwise the empty result above means the diff pipeline
reached nothing rather than that the classes are absent:

```bash
git diff -U0 develop...HEAD -- internal/template/templates \
  | grep '^+' | grep -vE '^\+\+\+' | grep -c 'local by design'
```

Expected: non-zero.

> **Revision-pinned, and three-dot deliberately.** A working-tree `git diff` with
> no revision compares against the index, so both the probe and its control return
> empty the moment the change is staged — and under §D.3 the criterion then cannot
> be closed at all. The three-dot form re-derives the merge base at read time, so
> it survives staging, commit, and a later absorb of `develop`. The moving ref is
> kept rather than pinned because the claim is *what this card added relative to
> its branch point*; a literal SHA would falsify it at the first absorb
> (`verification-claim-integrity.md` §2.1, SUBJECT class). `$(git merge-base …)` is
> not used: the worktree guard refuses a git command inside a command
> substitution, which is what made the v0.2.0 form unrunnable where cards are
> worked.
>
> **A whole-file form is not an alternative.** Measured against the
> pre-implementation tree, it returns 19 matching lines in
> `internal/template/templates/.claude/agents/moai/plan-auditor.md` alone —
> rubric placeholders and documentation examples the regex over-matches — so it
> reads as a FAIL that is false by construction.

Maps REQ-AEC-010

### AC-AEC-014 — no introduced sentence asserts this repository's tree state

**Given** REQ-AEC-012,
**When** the lines this card adds across **all thirteen** scope files are swept
for declarative tree-state forms, in every verb form that carries the claim,
**Then** none is found.

```bash
git diff -U0 develop...HEAD -- \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  .claude/agents/moai/manager-lead.md \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/core/agent-common-protocol-reference.md \
  .moai/docs/audit-artifact-convention.md \
  internal/template/templates \
  | grep '^+' | grep -vE '^\+\+\+' \
  | grep -nEi '\b(is|are|was|were|is not|are not|remains?|stays?|becomes?|has been|have been)( not)? (ignore-matched|gitignored|tracked|untracked|on the remote)\b'
echo "exit=$?"
```

Expected: no output, `exit=1`.

Paired control — a hedged or obligation form the repair introduces MUST match,
proving the diff pipeline and the regex both reach the added lines:

```bash
git diff -U0 develop...HEAD -- \
  .claude/agents/moai/plan-auditor.md \
  .claude/rules/moai/core/agent-common-protocol.md \
  .moai/docs/audit-artifact-convention.md \
  internal/template/templates \
  | grep '^+' | grep -vE '^\+\+\+' | grep -cEi 'local by design|is a local file|local\*{0,2} record'
```

Expected: non-zero.

> **The verb set is the repair, and it has an observed red.** The v0.2.0 form
> enumerated only the copula, and the assertion actually present in that draft's
> own specified wording was `remain tracked` — so the criterion written to catch
> the defect returned 0 against the defect while its control returned 1. The
> pattern above returns `1` against that v0.2.0 string (*"artifacts exported
> before this was written remain tracked with no exception recorded for them"*)
> and `0` against the §C.2 wording specified here, which is this criterion's
> RED-now observation and its green path.
>
> **Scope is the added lines, not the files.** Every one of the thirteen files
> legitimately contains declarative tree-state sentences this card did not write;
> REQ-AEC-012 binds sentences *this SPEC introduces*, so the diff is the correct
> subject and a whole-file form would be false by construction.

Maps REQ-AEC-012

### AC-AEC-015 — no repaired document tells its reader to decide ignore status with `-v`

**Given** REQ-AEC-014 and the discriminant in SPEC §A.2a,
**When** the twelve repaired target documents are swept for the verbose
`check-ignore` form,
**Then** none is found.

```bash
grep -rnE 'check-ignore[^`]*-v' \
  .claude/agents/moai/plan-auditor.md \
  .claude/agents/moai/sync-auditor.md \
  .claude/agents/moai/manager-lead.md \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/core/agent-common-protocol-reference.md \
  .moai/docs/audit-artifact-convention.md \
  internal/template/templates/.claude/agents/moai/plan-auditor.md \
  internal/template/templates/.claude/agents/moai/sync-auditor.md \
  internal/template/templates/.claude/agents/moai/manager-lead.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md \
  internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md \
  internal/template/templates/.moai/docs/audit-artifact-convention.md
echo "exit=$?"
```

Expected: no output, `exit=1`.

Paired control (reachability) — the specified check the repair preserves MUST
match in the two convention copies, proving the sweep reaches the repaired files:

```bash
grep -c 'ls .moai/reports/' \
  .moai/docs/audit-artifact-convention.md \
  internal/template/templates/.moai/docs/audit-artifact-convention.md
```

Expected: two lines, each ending `:1` or higher.

> **[HARD] Classification: regression-guard, not release-blocking — and it is
> vacuous against the current tree by construction.** Measured before
> implementation, the probe already returns zero: `check-ignore` appears **0
> times** across all twelve target documents, so a green here proves nothing
> about the repair. Read as a pass it would be an empty-swept-set claim
> (`verification-completeness.md` §1.1). What it actually guards is the repair
> **introducing** the form — and that hazard is real rather than theoretical,
> because the predecessor draft specified exactly this form for four of these
> twelve files (SPEC §A.2a).
>
> **The instrument has an observed red, taken against that draft.** The same
> pattern run over the v0.2.0 specified wording fires, and over the v0.3.1
> wording it does not:
>
> ```
> $ git show HEAD:.moai/specs/SPEC-AUDIT-EXPORT-CLAUSE-001/acceptance.md \
>     | grep -cE 'check-ignore[^`]*-v'
> 3
> $ sed -n '/^## §C Specified wording/,/^## §D Exclusions/p' \
>     .moai/specs/SPEC-AUDIT-EXPORT-CLAUSE-001/spec.md \
>     | grep -cE 'check-ignore[^`]*-v'
> 0
> ```
>
> That pair is this criterion's RED-now cell and its green path: `3` against text
> carrying the defect, `0` against the text specified here. Without it the
> criterion would be indistinguishable from a pattern that matches nothing.
>
> **Scope is the target documents, never this SPEC's own artifacts.** `spec.md`
> §A.2a and `plan.md` §E both quote the verbose form deliberately — that is the
> discriminant being *stated*, not a reader being *instructed*. A sweep widened
> to the SPEC directory would fail on the section that exists to prevent the
> defect.

Maps REQ-AEC-014

---

## §B Edge cases

| Case | Expected handling |
|---|---|
| A copy's FORBIDDEN wording differs from both variants in SPEC §C.1 | Blocker report, not an improvised third variant — the divergence is a finding about the copy |
| `cmp` reports the convention copies already divergent before the edit | Blocker report; the mirror is then edited section-by-section and the pre-existing divergence recorded, never flattened by copying |
| `make agents-emit-check` fails after `make agents-emit` | Emission is non-deterministic or the source layer is malformed; blocker, never resolved by editing a `.toml` |
| Withdrawing the negation appears to change `git status` under `.moai/reports/` | Stop. Tracked files must not move; re-read the status output and report before proceeding (REQ-AEC-003) |
| The plan-audit carve-out's behaviour changes after the withdrawal | Blocker. The `.gitignore` notes the ordering was load-bearing; AC-AEC-002's third probe is what detects it |
| A grep positive control returns zero | Every absence claim in the same run is void. Re-derive the instrument; do not report the zeros |
| A phrase this card searches for has been rewritten upstream | The zero is not absence. SPEC §A.6 is the standing instance; re-locate the claim by meaning and re-measure before reporting |

---

## §C Quality gates

- `internal/template/templates/.gitignore` does not appear in the change set — it
  carries no verdict negation and needs none.
- No file under `internal/template/templates/.codex/` appears in the change set as
  a hand edit; its presence is acceptable only as `make agents-emit` output.
- No `git rm --cached`, and no staged index change, on any path under
  `.moai/reports/`.
- The card's verdict is written to `.moai/reports/t1059/verdict.md` and **left
  local**. Force-staging it would be the card refuting its own direction; the lead
  reads it on disk, and the worktree is not disposed of until that read has
  happened.

---

## §D Definition of Done

### §D.1 Mandatory

1. AC-AEC-001 through AC-AEC-015 all pass, with each absence claim preceded in the
   same run by its positive control. AC-AEC-015 is a regression-guard and is not
   release-blocking: it is vacuous against the current tree by construction, and
   its verdict is read as "the repair introduced nothing", never as evidence the
   repair worked.
2. `make agents-emit-check` exits 0.
3. The two convention copies are byte-identical.
4. The tracked `.moai/reports/` population is measurably unchanged, reported as
   the two-way set difference rather than as counts.
5. The card's verdict artifact exists on disk at `.moai/reports/t1059/verdict.md`
   and has been read by the lead before the worktree is disposed of.

### §D.2 Evidence format

The card's verdict is written in the five-section evidence-bearing format
(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk). Every AC above
appears with its command and its verbatim observed output — a summary line is not
evidence. Because the verdict stays local, the verdict file carries the deciding
lines themselves rather than pointing at scratch that resolves nowhere else.

### §D.3 The zero-result rule

[HARD] No AC above may be closed on a zero-count result unless its paired positive
control returned non-zero **in the same run**. Two hazards make this non-optional
in this card specifically: the clause under repair carries `**bold**` runs, so a
literal grep whose pattern crosses a `**` delimiter matches nothing regardless of
the file's content; and a phrase this card searches for has already been rewritten
once by an upstream absorb, producing a clean zero on unrepaired files (SPEC §A.6).
In both cases a clean zero means either the target is absent or the pattern never
reached the file, and only the control separates the two readings.
