---
id: SPEC-AUDIT-EXPORT-CLAUSE-001
title: "the export-mandate clause forbids the destination it mandates"
version: "0.2.0"
status: draft
created: 2026-09-21
updated: 2026-09-21
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: .claude/agents/moai
lifecycle: spec-anchored
tier: M
tags: "audit-export, gitignore, force-add, agent-definition, template-mirror, self-contradicting-clause"
---

# SPEC-AUDIT-EXPORT-CLAUSE-001 — the export-mandate clause forbids the destination it mandates

## HISTORY

- 2026-09-21 · v0.2.0 · manager-spec · Plan-audit iteration 2, bounded to the
  returned defect delta. The falsified premise is replaced: an explicit-pathspec
  `git add` on an ignore-matched path was re-measured and **refuses loudly**
  (exit 1, naming `-f`), so §A.4's silent-failure narrative and every consequence
  clause derived from it are rewritten to the deadlock the measurement actually
  shows (§A.4, §C.1, §C.2). REQ-AEC-004 is widened from the export-mandate clause
  to every sentence this SPEC specifies, which is what let a tree-state assertion
  reach §C.2 — and, found while applying that widening, §C.3 as well. §C.1 is
  shortened from five specified sentences to four. AC-AEC-012's false-FAIL form
  is removed, AC-AEC-013 is deepened until it rejects the v0.1.0 wording,
  AC-AEC-002's mapping is corrected, and AC-AEC-016 is added as the first direct
  test of REQ-AEC-012. Six paraphrase surfaces measured and declared out of scope
  with a named closure. Pattern labels added to REQ-AEC-011/013/014/015.

- 2026-09-21 · v0.1.0 · manager-spec · Initial authoring from card t1059. Every
  measured statement carries the command that produced it; measurements taken in
  the card worktree at the commit named in §A.0. No figure appears here that was
  not observed in this tree.

---

## §0 Governing principle [HARD]

> **An instruction that names a destination and then forbids writing to that
> destination is not a strict rule. It is an unexecutable one, and a reader who
> obeys the first half and the second half cannot obey both.**

The subject of this SPEC is not the ignore policy. The policy is settled and is
explicitly out of scope (§ Out of Scope — the ignore policy). The subject is that
four agent-definition copies and one convention document instruct an actor to
write a file to a path, forbid writing to paths of exactly that kind, and omit
the one verb — a forced stage — that the instruction actually requires in order
to produce the outcome it claims to produce.

---

## §A Background

### §A.0 Measurement baseline

All measurements in this SPEC were taken in the card worktree
`.claude/worktrees/t1059`, branch `WT-audit-export-clause`, at commit
`3dfae918a`, working tree clean (`git status --porcelain` printed nothing).

### §A.1 The clause, and where it lives

A single `[HARD]` declaration — the **export mandate** — exists in four agent
files. Each is a declaration line, not a comment and not a table cell:

| Copy | File | Line |
|---|---|---|
| C1 | `.claude/agents/moai/plan-auditor.md` | 601 |
| C1 | `.claude/agents/moai/sync-auditor.md` | 108 |
| C2 | `internal/template/templates/.claude/agents/moai/plan-auditor.md` | 601 |
| C2 | `internal/template/templates/.claude/agents/moai/sync-auditor.md` | 92 |

Observed by `grep -n "Export mandate" <the four paths>`. The two sync-auditor
copies sit at **different** line numbers (108 and 92); they are not byte-parallel
and MUST be judged separately, per the three-copy doctrine in which C1 and C2 are
independently hand-edited and only C2 → C3 is a generation relation.

Each clause mandates a destination — `.moai/reports/<card-id>/plan-audit.md`,
`plan-audit-iter<N>.md`, `sync-audit.md`, or `.moai/reports/<SPEC-ID>/` — and
then closes with a sentence beginning `Never write the verdict to a gitignored
location`.

### §A.2 The mandated destination is a gitignored location

```
$ git check-ignore -v .moai/reports/t1059/plan-audit.md
.gitignore:227:.moai/reports/*	.moai/reports/t1059/plan-audit.md
$ git check-ignore -v .moai/reports/t1059/verdict.md
.gitignore:227:.moai/reports/*	.moai/reports/t1059/verdict.md
```

`.gitignore:227` is the single line `.moai/reports/*`. It ignore-matches every
new path under `.moai/reports/`, including the exact filenames the clause names.
The clause therefore commands a destination and forbids writing to it, in the
same sentence-pair, today — this is not a state created by any pending card.

### §A.3 The contradiction ships to user projects

`internal/template/templates/.gitignore:267` carries the same `.moai/reports/*`
rule (observed by `grep -n 'moai/reports' internal/template/templates/.gitignore`),
and the convention document has a byte-identical template mirror:

```
$ cmp internal/template/templates/.moai/docs/audit-artifact-convention.md .moai/docs/audit-artifact-convention.md && echo BYTE-IDENTICAL
BYTE-IDENTICAL
$ wc -c .moai/docs/audit-artifact-convention.md internal/template/templates/.moai/docs/audit-artifact-convention.md
    6872 .moai/docs/audit-artifact-convention.md
    6872 internal/template/templates/.moai/docs/audit-artifact-convention.md
```

Any wording this SPEC specifies must therefore be true in a freshly initialized
user project, not only in this repository.

### §A.4 Yet audit artifacts are tracked — the root finding

```
$ find .moai/reports -type f | wc -l
      54
$ git ls-files .moai/reports | wc -l
      54
$ git ls-files .moai/reports | grep -E 'audit'
.moai/reports/plan-audit/.gitkeep
.moai/reports/t675/plan-audit-iter2.md
.moai/reports/t675/plan-audit.md
.moai/reports/t965/plan-audit-verdict-iter2.md
.moai/reports/t965/plan-audit-verdict.md
```

Four audit-family `.md` files are tracked despite matching `.gitignore:227`.
The mechanism is that gitignore does not untrack an already-tracked file; a
**new** file under an ignore-matched path reaches tracked state only through an
explicit `git add -f`.

**What the documents are missing is therefore not an exception. It is a
sanctioned verb.** Neither agent clause nor the convention document mentions the
forced stage that their own instructions require.

### §A.4b What actually happens to an actor who follows the documents

The failure is a **deadlock with no sanctioned exit**, not a silent success. The
distinction decides the repair, so it is measured rather than assumed:

```
$ mkdir -p .moai/reports/ZZPROBE2 && printf 'probe\n' > .moai/reports/ZZPROBE2/plan-audit.md
$ git add .moai/reports/ZZPROBE2/plan-audit.md
The following paths are ignored by one of your .gitignore files:
.moai/reports/ZZPROBE2
hint: Use -f if you really want to add them.
hint: Disable this message with "git config set advice.addIgnoredFile false"
$ echo $?
1
$ git status --porcelain -- .moai/reports/ZZPROBE2/
(no output — nothing staged)
```

An explicit-pathspec `git add` **refuses**: it exits 1, stages nothing, and names
`-f` itself. Suppressing the advice does not change that — the refusal and the
exit status both survive it:

```
$ git -c advice.addIgnoredFile=false add .moai/reports/ZZPROBE2/plan-audit.md
The following paths are ignored by one of your .gitignore files:
.moai/reports/ZZPROBE2
$ echo $?
1
```

Only a whole-tree sweep skips quietly, and that staging form is prohibited by the
standing contract (`AGENTS.md` § Git, branches, and the shared checkout — *"Never
sweep-stage. In the primary checkout, never `git add -A`, `git add .`, or `git
commit -a`."*). Silence is therefore reachable only by violating a separate rule
and is not the hazard this SPEC repairs.

The hazard is what the compliant actor meets instead. Having written the file to
the mandated destination, they stage it by explicit pathspec, receive a correct
and loud refusal, and are handed a remedy — `-f` — that **no sanctioned
instruction anywhere permits them to use**:

```
$ grep -rn 'add -f' .claude/rules .claude/agents .claude/skills .moai/docs \
    internal/template/templates/.claude internal/template/templates/.moai
$ echo $?
1
$ grep -rlE 'add -A|add \.' .claude/rules .moai/docs | head -5    # positive control
.claude/rules/moai/core/agent-common-protocol-reference.md
.claude/rules/moai/core/agent-common-protocol.md
.claude/rules/moai/workflow/main-checkout-branch-guard.md
.claude/rules/moai/workflow/kanban-dispatch-detail.md
.moai/docs/mcp-recipes.md
```

Zero occurrences of the forced stage, with a control proving the sweep reaches
those trees. The clause the actor is following has just told them *"Never write
the verdict to a gitignored location"*, so forcing the path reads as defiance of
the instruction rather than compliance with it. Both available exits are defects:
abandon the export and the audit is incomplete by the clause's own declaration,
or force on personal authority against an explicit prohibition.

**So the missing thing is not an alarm — git already rings one. It is a
sanctioned next step, plus removal of the prohibition that contradicts it.**

### §A.5 A second, subtler defect in the reader's own check

```
$ git check-ignore -v .moai/reports/t675/plan-audit.md
$ echo $?
1
$ git check-ignore -v --no-index .moai/reports/t675/plan-audit.md
.gitignore:227:.moai/reports/*	.moai/reports/t675/plan-audit.md
$ echo $?
0
```

Without `--no-index`, `git check-ignore` answers *"is this path already
tracked?"* and returns exit 1 for a tracked file. The question the export
mandate needs answered is *"would a **new** file at this path be ignored?"*, and
only `--no-index` answers it. A repaired clause that names the check without
naming the flag hands the reader an instrument that reports a clean result on
exactly the paths where the hazard is invisible.

### §A.6 The convention document repeats the same falsehood

`.moai/docs/audit-artifact-convention.md` § Committing opens:

> "Audit artifacts are tracked files. They reach the integration branch with the
> card's evidence commit — never left as uncommitted files in a worktree."

The second sentence is false for the same reason: a new artifact does not reach
the branch with an ordinary evidence commit. Two further sentences in the same
document are made false-by-implication or incomplete by the same root:

- § Where, FORBIDDEN bullet: *"That directory is deliberately gitignored"*,
  offered as the property distinguishing the forbidden directory — but the
  sanctioned card-scoped destination is ignore-matched too, so the stated
  property does not distinguish anything.
- § What makes the convention stick, mechanical check: `ls
  .moai/reports/<card-id>/` detects presence on disk, which is the half of the
  obligation that is not at risk.

---

## §B Requirements

Requirements use GEARS notation. `<subject>` is generalized per the current
authoring standard.

### §B.1 The repaired clause

- **REQ-AEC-001** (Ubiquitous) — The export-mandate clause shall name the forced
  stage in each auditor agent definition (`git add -f <path>`) as the sanctioned
  act for the destination it mandates, so that the reader is permitted to use the
  remedy git names on refusal (§A.4b) rather than choosing between abandoning the
  export and acting against the clause.

- **REQ-AEC-002** (Ubiquitous) — The export-mandate clause shall name a check the
  reader can run to decide whether the forced stage is required, and that check
  shall be `git check-ignore -v --no-index <path>`, with the `--no-index` flag and
  its consequence stated, per §A.5.

- **REQ-AEC-003** (Ubiquitous) — The export-mandate clause shall state the
  consequence of an unstaged verdict — that the verdict does not reach the
  integration branch — in the same sentence that names the forced stage, so that
  **where the verdict is written** and **whether it reaches the branch** are
  visibly two obligations rather than one. The clause shall not carry a separate
  generalizing sentence to make that point.

- **REQ-AEC-004** (Ubiquitous) — The wording this SPEC specifies shall not assert that a destination is, or is not, ignore-matched.
  The prohibition binds every specified sentence in every target file enumerated
  in the implementation plan's scope table — not the export-mandate clause alone.
  Each such sentence shall either hedge the state or instruct the reader to run
  the check of REQ-AEC-002.

> REQ-AEC-004 is the two-state requirement, and its scope is every specified
> sentence rather than the export-mandate clause alone — the narrower v0.1.0
> scoping is what let a tree-state assertion reach the convention wording. Before
> the pending narrowing card lands, every name under `.moai/reports/<card-id>/`
> is ignore-matched; after it lands, one filename is un-ignored while the audit
> family remains matched. A sentence asserting either state is false in the
> other. A sentence that instructs a check is true in both, and is also true in a
> user project whose `.gitignore` this repository does not control.

- **REQ-AEC-005** (Ubiquitous) — The export-mandate clause shall retain a
  prohibition on the report directory the convention declares FORBIDDEN, and that
  prohibition shall be stated as unconditional — independent of what the
  REQ-AEC-002 check reports.

- **REQ-AEC-006** (Ubiquitous) — The repaired wording shall be applied to all four
  copies enumerated in §A.1, each judged against the copy's own surrounding text
  rather than against byte-parity with its sibling.

### §B.2 The convention document

- **REQ-AEC-007** (Ubiquitous) — The convention document shall state the forced
  stage and the REQ-AEC-002 check in its § Committing section, and shall state that ignore-matching does not untrack an
  already-tracked file — the fact that explains why existing artifacts stay tracked
  with no exception written for them.

- **REQ-AEC-008** (Ubiquitous) — The § Where FORBIDDEN bullet shall not offer
  "is gitignored" as the property distinguishing the forbidden directory, and
  shall state the property that actually distinguishes it: that nothing written
  there is ever force-staged.

- **REQ-AEC-009** (Ubiquitous) — The convention document shall name, in its
  § What makes the convention stick mechanical check, a command that detects the
  branch-reachability half of the obligation, not only the on-disk half.

- **REQ-AEC-010** (Ubiquitous) — Every change to the convention document shall be
  applied identically to its template mirror at
  `internal/template/templates/.moai/docs/audit-artifact-convention.md`, and the
  two copies shall remain byte-identical after the change.

### §B.3 Constraints on the wording itself

- **REQ-AEC-011** (Where) — Where the edit target lies under the template tree, the wording shall carry no forbidden content class.
  Specifically it shall contain no SPEC identifier, no requirement token, no
  internal date, no commit hash, no platform-specific absolute path, and no
  reference to a local-only development guide.

- **REQ-AEC-012** (Ubiquitous) — Every sentence introduced by this SPEC shall be
  decidable by a command a reader can run, or shall be a statement of obligation.
  No introduced sentence shall assert a fact about this repository's tree state.

- **REQ-AEC-013** (Event-driven) — When the C2 agent copies are modified, the actor shall
  regenerate the emitted codex layer with `make agents-emit` and shall not
  hand-edit any file under `internal/template/templates/.codex/`.

### §B.4 Unwanted behavior

- **REQ-AEC-014** (Unwanted) — The repaired clause shall not instruct the reader to modify
  `.gitignore`, to add an ignore exception, or to widen tracked scope.

- **REQ-AEC-015** (Unwanted) — The repaired clause shall not weaken the existing export
  obligation: an audit response without an exported file shall remain an
  incomplete audit.

---

## §C Specified wording

The wording below is normative. It replaces the final sentence of each
export-mandate clause — the sentence beginning `Never write the verdict to a
gitignored location`. Everything preceding that sentence is unchanged.

### §C.1 Replacement tail — both auditor copies

> The destination above may be ignore-matched; run `git check-ignore -v
> --no-index <path>` rather than assuming either way. On exit 0 a plain `git add`
> **refuses** the path — it exits non-zero and stages nothing, so the verdict
> never reaches the integration branch; stage it with `git add -f <path>` in the
> card's evidence commit, which this mandate permits for exactly this
> destination. (`--no-index` is load-bearing: without it the command answers
> whether the path is already tracked.) One destination stays forbidden whatever
> the check reports: <FORBIDDEN-CLAUSE>.

Four specified sentences, one per obligation: the check (REQ-AEC-002), the
consequence-plus-remedy (REQ-AEC-001 + REQ-AEC-003, deliberately one sentence so
the two obligations are visible without a generalizing sentence to restate them),
the flag justification, and the preserved prohibition (REQ-AEC-005). The
permission clause — *"which this mandate permits for exactly this destination"* —
is the sentence that resolves §A.4b's deadlock: without it the reader still reads
`-f` as defiance of the surrounding clause.

`<FORBIDDEN-CLAUSE>` is taken verbatim from the copy being edited, preserving the
existing per-file difference:

- plan-auditor copies: `the report directory the convention declares FORBIDDEN
  (`audit-artifact-convention.md` § Where) receives verdicts as disposal, not
  export`
- sync-auditor copies: ``.moai/reports/plan-audit/` is FORBIDDEN — writing there
  is disposal, not export`

### §C.2 Replacement — convention § Committing

> An audit artifact becomes a tracked file only when it is deliberately staged,
> and a new one under the report directory may be ignore-matched. Check the
> destination with `git check-ignore -v --no-index <path>`; on exit 0 a plain
> `git add` refuses the path — it exits non-zero and stages nothing — so stage it
> with `git add -f <path>`. Ignore-matching does not untrack a file that is
> already tracked, which is why artifacts exported before this was written remain
> tracked with no exception recorded for them.
>
> The artifact reaches the integration branch with the card's evidence commit —
> never left as an uncommitted file in a worktree. A worktree holding the only
> copy of a verdict is a disposal hazard: the tree is removed when the card
> closes, and the verdict goes with it.

Sentence 1 is recast as an obligation (*"becomes a tracked file only when it is
deliberately staged"*) rather than the v0.1.0 state assertion (*"Audit artifacts
are tracked files"*), which was true of four existing files and false of every
future one. Sentence 2's ignore-state clause is hedged to `may be`, satisfying the
widened REQ-AEC-004; the v0.1.0 form asserted the current ignore state outright
and was false in the right-hand column of the plan's own two-state table.

### §C.3 Replacement — convention § Where, FORBIDDEN bullet justification

> Nothing written there is ever force-staged, and that — not ignore-matching — is
> what distinguishes the directory: the card-scoped destinations above may be
> ignore-matched too (§ Committing). Because no convention forces its contents
> into the tree, a verdict written there is disposed of, not exported. Do not
> repurpose the directory.

Revised under the widened REQ-AEC-004. The v0.1.0 form carried two tree-state
assertions the narrower v0.1.0 scoping did not reach — *"Only its `.gitkeep` is
tracked"* and *"the card-scoped destinations above **are** ignore-matched too"* —
neither of which was flagged in review. The first is dropped (it adds nothing the
distinguishing property does not already carry); the second is hedged.

### §C.4 Addition — convention § What makes the convention stick

Appended to the existing mechanical-check bullet:

> Presence on disk is half the obligation; `git ls-files --error-unmatch
> .moai/reports/<card-id>/<file>` exiting non-zero means the verdict is present
> locally and absent from the branch.

---

## §D Exclusions

### Out of Scope — the ignore policy

- Widening `.gitignore` to admit the audit-verdict artifact family. The operator
  rejected this direction explicitly; this SPEC changes the documents to match
  reality, never the reverse.
- Editing `.gitignore` at all, in either the repository copy or the template
  mirror.
- The pending narrowing card that un-ignores a single filename. This SPEC must be
  correct whether or not that card has landed (REQ-AEC-004) and takes no position
  on its design.

### Out of Scope — the evidence-path policy

- Whether audit evidence should live under `.moai/reports/` at all.
- Whether card evidence should be tracked, exported to a different root, or kept
  local. That question is settled by the operator and is not reopened here.
- The five-section evidence-bearing format itself, and the minimum-content list
  the clause already names.

### Out of Scope — the paraphrase-class "tracked path" surfaces

Six further files assert the same falsehood in different words, calling
`.moai/reports/<card-id>/` *the tracked path*. They are measured, named here, and
deliberately **excluded** — silence about them is the one option not available.

```
$ grep -rnE 'tracked\*{0,2} path' .claude/rules .claude/agents .claude/skills \
    internal/template/templates/.claude | wc -l
       8
$ grep -rlE 'tracked\*{0,2} path' .claude/rules .claude/agents .claude/skills \
    internal/template/templates/.claude
.claude/rules/moai/core/agent-common-protocol-reference.md          (L62)
.claude/rules/moai/core/agent-common-protocol.md                    (L274)
.claude/agents/moai/manager-lead.md                                 (L152, L154)
internal/template/templates/.claude/agents/moai/manager-lead.md     (L154, L156)
internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md   (L62)
internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md             (L274)
```

Instrument note: the markup-insensitive pattern is required. A literal
`grep -rlF 'a tracked path'` over the same trees returns **2** files, not 6 — the
bold run in `a **tracked** path` defeats it. This is the same markup-sensitivity
AC-AEC-002 exists to catch, met in the sweep that scoped this exclusion.

**Why excluded, not absorbed.** Three reasons, in order of weight:

- They make a **different claim**. These sentences designate a citation target
  (*"the citation target is a tracked path"*), not an export instruction. The
  repair they need is a re-designation, not the check-and-force wording this SPEC
  specifies, so absorbing them would put two unrelated repairs under one set of
  requirements.
- **Budget.** This SPEC sits at 15 requirements and 16 acceptance criteria against
  a Tier M ceiling of 16 each (`spec-workflow.md` § SPEC Complexity Tier). Six
  files carrying a distinct repair cannot be added without new requirements and
  criteria, which would exceed the ceiling and force a re-tier to L — the
  over-formalization the tier taxonomy exists to prevent.
- **Reach.** `agent-common-protocol.md` is always-loaded, which makes its wording
  higher-impact than any of the four clause copies here. That argues for its own
  card with its own review, not for a tail appended to this one.

**What closes them:** a successor card carrying the same check-shaped repair to
those six files, scoped to the citation-target designation. Its issuance is an
operator act (the lead is the queue's sole producer — `kanban-dispatch.md` § Entry
into the board is an operator act), so no card id is invented here; the sweep
above is recorded precisely so that card is derivable without re-deriving it. Until
that card lands, the six surfaces keep the falsehood, and this SPEC does not claim
otherwise.

### Out of Scope — adjacent documents

- Any agent definition other than `plan-auditor.md` and `sync-auditor.md`.
- Any rule file that cross-references the export mandate without restating it.
- The emitted codex layer `internal/template/templates/.codex/agents/moai/*.toml`
  as an *edit* target. It is regenerated, never hand-edited (REQ-AEC-013); it is
  in scope only as a verification surface.

### Out of Scope — behavior changes

- Adding any mechanical enforcement (hook, lint rule, CI guard) for the forced
  stage. This SPEC is a documentation-truth repair; a mechanism is a separate
  decision with its own cost.
- Changing where auditors write, when they write, or what they write.

---

## §E Risks

| Risk | Consequence | Mitigation |
|---|---|---|
| The repaired clause is longer than the sentence it replaces | Reader skims past the operative verb | The forced stage appears in the second sentence, ahead of the parenthetical |
| A future ignore-policy change makes the wording stale | The same defect recurs | REQ-AEC-004 forbids asserting the ignore state, so no policy change can falsify the clause |
| C1 and C2 drift during the edit | Template ships different wording than the repository dogfoods | Divergence is intentional by doctrine; §C specifies the tail identically for all four, and each copy's FORBIDDEN clause is preserved verbatim |
| A literal grep for the clause misses a copy | A copy is silently left unrepaired | Acceptance requires a positive control proving the grep fires before any zero is read as absence |
