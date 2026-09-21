---
id: SPEC-AUDIT-EXPORT-CLAUSE-001
title: "documents that promise a remote destination the directive forbids"
version: "0.3.1"
status: draft
created: 2026-09-21
updated: 2026-09-22
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: .claude/agents/moai
lifecycle: spec-anchored
tier: M
tags: "audit-export, gitignore, local-only-evidence, agent-definition, template-mirror, operator-directive"
---

# SPEC-AUDIT-EXPORT-CLAUSE-001 — documents that promise a remote destination the directive forbids

## HISTORY

- 2026-09-22 · v0.3.1 · manager-spec · **Lead review of v0.3.0 — two bounded
  additions, no reopening.** The lead reviewed the draft, independently
  reproduced the `check-ignore` finding, and asked for two things. First, the
  discriminant is promoted from an instrument note in the plan to a stated
  property of this SPEC (§A.2a) with a requirement (REQ-AEC-014) and a criterion
  (AC-AEC-015), because it binds the wording this SPEC **specifies for other
  documents** and not only its own probes — a binding v0.2.0 violated outright
  (§A.2a). Second, §E records the basis for counting 13 files rather than 16 at
  the Tier boundary. Requirements 13 → 14, criteria 14 → 15; see §A.7.

- 2026-09-21 · v0.3.0 · manager-spec · **Direction change, not a defect fix.** The
  operator ruled that the 2026-09-14 directive is canonical: lead-verdict evidence
  stays on disk and does not reach the remote. That inverts the repair. The
  `git add -f` permission clause is **withdrawn**, not repaired; §A.4b's deadlock
  framing is withdrawn with it, because under the ruling the plain `git add`
  refusal is the policy working rather than a trap. The card now also **enacts**
  the ruling: the `.gitignore` negation re-including `verdict.md` is withdrawn in
  this card, together with the mechanism statement that withdrawal binds only
  files created afterwards. The paraphrase surfaces excluded in v0.2.0 are
  **absorbed**, because the ruling converts them from imprecise wording into false
  statements about policy. Requirements were re-derived rather than patched: the
  set is 13 (was 15) and the criteria 14 (was 16), which is how the widened scope
  fits Tier M with headroom — see §A.7.

- 2026-09-21 · v0.2.0 · manager-spec · Plan-audit iteration 2, bounded to the
  returned defect delta. The falsified premise was replaced: an explicit-pathspec
  `git add` on an ignore-matched path refuses loudly (exit 1, naming `-f`), so
  §A.4's silent-failure narrative and every consequence clause derived from it
  were rewritten. REQ-AEC-004 was widened to every specified sentence. AC-AEC-016
  was added as the first direct test of REQ-AEC-012. Six paraphrase surfaces were
  measured and declared out of scope with a named closure.

- 2026-09-21 · v0.1.0 · manager-spec · Initial authoring from card t1059.

---

## §0 Governing principle [HARD]

> **A document that promises a destination the project's policy forbids is not a
> strict document. It is a false one — and the falsehood is silent, because the
> reader who obeys it produces exactly the artifact the policy meant to prevent
> and receives no signal that anything went wrong.**

The subject of this SPEC is not whether evidence should reach the remote. That
question is **settled by the operator** (§A.1) and is out of scope as a decision.
The subject is that four agent-definition copies, one convention document, an
always-loaded rule and its reference, and one coordination agent all tell an actor
that a card-report artifact is tracked, reaches the integration branch, or is a
tracked citation target — and the directive says it is none of those things. One
`.gitignore` negation currently makes a narrow slice of those claims true; this
card withdraws it, which is what makes the wording repair a repair rather than a
regression.

---

## §A Background

### §A.0 Measurement baseline

All measurements in this SPEC were taken in the card worktree
`.claude/worktrees/t1059`, branch `WT-audit-export-clause`, at commit
`64c7edbf3` (local `develop` absorbed), working tree clean
(`git status --porcelain | wc -l` printed `0`).

Every figure carried over from the v0.2.0 draft was **re-measured**, not
inherited. Two did not survive the re-measurement — §A.5 and §A.6 record what
changed and why the change matters.

### §A.1 The directive — the decision this SPEC implements

The repository `.gitignore` carries the operator directive as a comment block
above the blanket rule:

```
$ sed -n '232,235p' .gitignore
# Card/audit reports are local-only artifacts (operator directive 2026-09-14):
# evidence for lead verdicts stays on disk, never on the remote. Only the
# plan-audit scaffold ships, mirroring the template's reports philosophy.
.moai/reports/*
```

Enacted by commit `ba60eb6d5` (2026-09-14, *"chore(repo): untrack .moai/reports —
local-only artifacts"*), an ancestor of HEAD.

The operative words are **"never on the remote"**. The directive is about where an
artifact ends up, not about where it is written. Writing the verdict is still
mandatory; sending it to the remote is what is forbidden.

**The template corroborates the directive independently.** The comment claims to
mirror "the template's reports philosophy", and the template's own ignore file
admits no verdict exception:

```
$ grep -c 'verdict.md' internal/template/templates/.gitignore
0
$ grep -n 'moai/reports' internal/template/templates/.gitignore
257:# Everything under .moai/reports/ — plan-audit verdicts, per-card evidence
262:# Rule order is load-bearing: `.moai/reports/*` excludes the plan-audit
267:.moai/reports/*
268:!.moai/reports/plan-audit/
269:.moai/reports/plan-audit/*.md
270:!.moai/reports/plan-audit/.gitkeep
278:.moai/reports/graph-report.md
```

The zero is attributable: the control on the same file (`wc -l` → 325 lines, and
the five reports rules above) shows the sweep reaches it. So the template tree
needs no ignore change under this direction, and is out of scope as an edit target
for that reason (§D) rather than by oversight.

### §A.2 The negation that currently contradicts the directive

A later block re-includes one filename under every card directory:

```
$ sed -n '257,259p' .gitignore
!.moai/reports/*/
.moai/reports/*/*
!.moai/reports/*/verdict.md
```

Measured effect on a path that does not yet exist:

```
$ git check-ignore --no-index .moai/reports/ZZAC/plan-audit.md && echo IGNORED || echo NOT-IGNORED
.moai/reports/ZZAC/plan-audit.md
IGNORED
$ git check-ignore --no-index .moai/reports/ZZAC/verdict.md && echo IGNORED || echo NOT-IGNORED
NOT-IGNORED
$ git check-ignore --no-index README.md && echo IGNORED || echo NOT-IGNORED     # control
NOT-IGNORED
```

The control distinguishes "not matched" from "the instrument did not run": a path
outside the reports tree reports `NOT-IGNORED` for the ordinary reason, and
`plan-audit.md` reports `IGNORED`, so the two branches are both reachable.

Under the directive this negation is the defect: it opens a supported route for
exactly the class of artifact the directive names. This card withdraws it
(REQ-AEC-001), together with the explanatory block introducing it — a comment
justifying a rule that no longer exists is the next reader's trap.

### §A.2a [HARD] The plain form decides; `-v` answers a different question

The probe above omits `-v`, and the omission is load-bearing rather than
incidental. The two forms answer different questions, and only one of them is
about ignore status:

| Form | The question it answers | What exit 0 means |
|---|---|---|
| `git check-ignore <path>` | **Is this path ignored?** | the path **is** ignored |
| `git check-ignore -v <path>` | **Did any rule match this path?** | some rule matched — **a negation included**, which means the path is **not** ignored |

Measured in this tree at `64c7edbf3`, with the negation still live:

```
$ git check-ignore -v --no-index .moai/reports/ZZPROBE/verdict.md ; echo "exit=$?"
.gitignore:259:!.moai/reports/*/verdict.md	.moai/reports/ZZPROBE/verdict.md
exit=0
$ git check-ignore --no-index .moai/reports/ZZPROBE/verdict.md ; echo "exit=$?"
exit=1
$ git check-ignore --no-index .moai/reports/ZZPROBE/plan-audit.md ; echo "exit=$?"
.moai/reports/ZZPROBE/plan-audit.md
exit=0
```

One path, two exit codes. The verbose output is not even misleading — it names
the negation it matched, on the line it matched it. Only the **exit code** is
misread, and an actor keying on `$?` never reads the line that would have
corrected it.

**The defect is invisible to the obvious probe.** On every **non-negated** path
the two forms agree, exit code included:

```
$ git check-ignore -v --no-index .moai/reports/ZZPROBE/plan-audit.md ; echo "exit=$?"
.gitignore:258:.moai/reports/*/*	.moai/reports/ZZPROBE/plan-audit.md
exit=0
```

So a probe exercising only a name the negation never covered — `plan-audit.md`
being the obvious pick — observes agreement and certifies the wrong instrument.
The divergence surfaces on the **negated** name alone, which is the one name a
reader asking "is `verdict.md` ignored now?" must use and a reader sanity-checking
the tooling is least likely to reach for. Every criterion in this card that
decides ignore status therefore uses the plain form and probes **both** names
(AC-AEC-002).

**The discriminant binds the wording this SPEC specifies for other documents,
not only its own probes.** A specified sentence instructing a reader to run `-v`
and decide on its exit code would ship the defect into doctrine, where every
future actor repairing these files would inherit it. REQ-AEC-014 forbids that
wording and AC-AEC-015 tests for it.

That hazard is not hypothetical, and the evidence is this SPEC's own predecessor.
The v0.2.0 draft specified `git check-ignore -v --no-index` as the check-form
wording for all four auditor copies and made the unflagged form a FAIL:

```
$ git show HEAD:.moai/specs/SPEC-AUDIT-EXPORT-CLAUSE-001/acceptance.md \
    | grep -n 'check-ignore' | cut -c1-96
90:grep -rc 'check-ignore -v --no-index' \
99:A match of `check-ignore` **without** `--no-index` in these files is a FAIL, not
177:Expected output contains, in this order: `git check-ignore -v --no-index`,
312:git check-ignore -v --no-index "$P"; CHECK=$?; echo "check_exit=$CHECK"
```

Line 312 is that draft's own probe keying on the verbose exit code — and the
path it probed was `.moai/reports/ZZAC013/plan-audit.md`, the agreement case, so
the instrument could not have exposed itself. What removed the wording was the
direction change, not a criterion. Nothing in v0.2.0 was watching for it, which
is the reason this section exists rather than a note in the plan.

### §A.3 [HARD] Withdrawal binds only files created afterwards

**Withdrawing the negation does not untrack anything.** Gitignore is consulted for
paths git is not already tracking; a file already in the index stays in the index
and keeps reaching the remote on the next push, whatever the ignore rules say. So
the withdrawal closes the route forward and leaves the existing population
untouched. This SPEC states the mechanism because a reader who withdraws the
negation and stops has repaired half of the state and has no signal telling them
so.

Measured in this tree, with an anchored criterion applied to **both** sides:

| Measurement | Command | Result |
|---|---|---|
| tracked `verdict.md` | `git ls-files '.moai/reports/*/verdict.md' \| wc -l` | **12** |
| present on `origin/develop`, same criterion | `git ls-tree -r --name-only origin/develop -- .moai/reports \| grep -E '^\.moai/reports/[^/]+/verdict\.md$' \| wc -l` | **10** |
| tracked but not on the remote | `comm -23` of the two sorted sets | **2** — `.moai/reports/t1039/verdict.md`, `.moai/reports/t1048/verdict.md` |
| remote-only | `comm -13` of the two sorted sets | **0** (the local set is a proper superset) |

Per-file confirmation, with a firing positive control:

```
$ git cat-file -e origin/develop:.moai/reports/t1039/verdict.md ; echo "exit=$?"
fatal: path '.moai/reports/t1039/verdict.md' exists on disk, but not in 'origin/develop'
exit=128
$ git cat-file -e origin/develop:.moai/reports/t1048/verdict.md ; echo "exit=$?"
fatal: path '.moai/reports/t1048/verdict.md' exists on disk, but not in 'origin/develop'
exit=128
$ git cat-file -e origin/develop:.moai/reports/t965/plan-audit-verdict.md ; echo "exit=$?"   # control
exit=0
```

**[HARD] Instrument warning, carried into the criteria.** The anchor is not
decoration. A substring `grep 'verdict.md'` on the remote side matches
`plan-audit-verdict.md` too and returns **11** rather than 10:

```
$ git ls-tree -r --name-only origin/develop -- .moai/reports | grep 'verdict.md' | wc -l
      11
$ git ls-tree -r --name-only origin/develop -- .moai/reports | grep 'verdict.md' \
    | grep -vE '^\.moai/reports/[^/]+/verdict\.md$'
.moai/reports/t965/plan-audit-verdict.md
```

A count difference of 1 would then have looked like a near-match. It is not: the
two-way set difference is **2 and 0**, and it is the two-way form — not the
counts — that exposed the real shape. Any criterion comparing these sets uses the
anchored pattern on both sides and reports the two differences, never the counts
alone.

**The disposition of those two files is not decided here** (§D).

### §A.4 What the documents currently say

Four agent copies carry a `[HARD]` export mandate whose final sentence is
`Never write the verdict to a gitignored location`:

| Copy | File | Line |
|---|---|---|
| C1 | `.claude/agents/moai/plan-auditor.md` | 601 |
| C1 | `.claude/agents/moai/sync-auditor.md` | 108 |
| C2 | `internal/template/templates/.claude/agents/moai/plan-auditor.md` | 601 |
| C2 | `internal/template/templates/.claude/agents/moai/sync-auditor.md` | 92 |

Observed by `grep -n 'Never write the verdict to a gitignored location'` over the
four paths. The two sync-auditor copies sit at **different** lines (108 and 92);
no copy is byte-parallel with its sibling (`cmp -s` reports DIVERGENT for all
three agent pairs and both rules pairs measured in §A.6), so each is judged on its
own surrounding text.

Under the directive this sentence is not false — the destination *is* meant to be
a location git does not carry. What is false is what it implies next: that an
artifact written there is nevertheless expected to arrive somewhere. The
convention document states that expectation outright:

```
$ sed -n '122,128p' .moai/docs/audit-artifact-convention.md
## Committing

Audit artifacts are tracked files. They reach the integration branch with the
card's evidence commit — never left as uncommitted files in a worktree. A
worktree holding the only copy of a verdict is a disposal hazard: the tree is
removed when the card closes, and the verdict goes with it.
```

Sentence 1 and sentence 2 are both false under the directive, and they are false
in the **opposite direction** from the v0.2.0 reading: the repair is to withdraw
them, not to qualify them with a forced-stage instruction. The third sentence —
the disposal hazard — is true, is the *only* live hazard under this direction, and
survives verbatim.

### §A.5 The § Where bullet loses its distinguishing property either way

```
$ sed -n '49,53p' .moai/docs/audit-artifact-convention.md
- FORBIDDEN: `.moai/reports/plan-audit/`. That directory is deliberately
  gitignored — only its `.gitkeep` is tracked, and the repository's
  `.gitignore` comment marks the directory as local artifacts. A verdict
  written there is disposed of, not exported. Do not repurpose the directory.
```

"Deliberately gitignored" is offered as the property distinguishing the forbidden
directory. Once the negation is withdrawn, **every** card-report path is
ignore-matched again, so the stated property distinguishes nothing — the same
defect the v0.2.0 draft identified, reached by the opposite route. The pointer to
the `.gitignore` comment is true, is not a tree-state assertion (it names a file
the reader can open), and under this direction it is the only pointer in the
convention document to the rationale the whole repair turns on; it is retained
deliberately (REQ-AEC-007).

### §A.6 [HARD] The paraphrase surfaces — re-measured, and the phrase has moved

The v0.2.0 draft excluded six files carrying the phrase `a **tracked** path`.
**That phrase no longer exists in this tree.** The 45 commits absorbed from
`develop` rewrote those sentences:

```
$ grep -rnE 'tracked\*{0,2} path' .claude/rules .claude/agents .claude/skills \
    internal/template/templates/.claude ; echo "exit=$?"
exit=1
$ grep -rlE 'tracked\*{0,2} (citation target|verdict file)' .claude/rules .claude/agents \
    internal/template/templates/.claude | wc -l      # control: the sweep reaches these trees
       6
```

The zero is attributable: the control returns 6 in the same run, so the trees are
reachable and the zero is a true absence of that phrase rather than a broken
instrument. **The file set is unchanged — 6 files, 8 lines — and only the wording
moved:**

```
$ grep -rnE 'tracked\*{0,2} (citation target|verdict file)' .claude/rules .claude/agents \
    internal/template/templates/.claude | cut -d: -f1,2
.claude/rules/moai/core/agent-common-protocol-reference.md:62
.claude/rules/moai/core/agent-common-protocol.md:274
.claude/agents/moai/manager-lead.md:63
.claude/agents/moai/manager-lead.md:152
internal/template/templates/.claude/agents/moai/manager-lead.md:65
internal/template/templates/.claude/agents/moai/manager-lead.md:154
internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md:62
internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md:274
```

**The new wording is coupled to the negation far more tightly than the old.** It
does not merely call the path tracked — it describes the negation mechanism:

> `agent-common-protocol.md:274` — *"The one tracked citation target is the
> **verdict file** — in this repository `.moai/reports/<card-id>/verdict.md`. The
> directory around it is not tracked: the ignore rules re-include that single
> filename and nothing else…"*

Withdrawing the negation falsifies that sentence directly: the ignore rules
re-include nothing, and a verdict file at a new card directory is ignored. The
same falsification reaches `manager-lead.md`'s *"the tracked verdict file"* and
*"the verdict file is the only tracked name under a card directory"*.

`agent-common-protocol.md` is **always-loaded**, so it reaches every session. This
is why the v0.2.0 exclusion does not survive the direction change: it rested on
those sentences making a *different* claim, and under this direction they make the
same claim the `.gitignore` edit in this very card refutes. Splitting them into a
successor card would leave an always-loaded rule asserting a fact this card makes
false, for the duration of that card — which is the state the lead already ruled
against when it required specification and execution to stay together.

### §A.7 Why this still fits Tier M

The widened scope reaches **13 hand-edited files** — four auditor copies, two
convention copies, the repository `.gitignore`, four always-loaded-rule copies,
and two `manager-lead` copies — plus three machine-emitted codex TOMLs that are
not edit targets (§D). Tier M's band is 5-15 files; the changed-line count is
under a hundred, well inside Tier S's LOC guidance. **Files put this at Tier M and
LOC puts it below; the binding constraint was never the tier band but the 16/16
requirement and criterion ceiling.**

The iteration-2 auditor's view was that splitting beats tiering up. **Neither is
needed, and splitting is the worse of the two**: §A.6 measures the coupling that
makes it unsafe. What makes the budget fit is re-deriving the requirement set
rather than patching it — expressing the wording repair as **one act applied
across every surface** (REQ-AEC-004) instead of one requirement per surface. The
result is 14 requirements and 15 criteria against a ceiling of 16 each, with
headroom of 2 and 1 rather than the zero v0.2.0 carried. The v0.3.1 lead-review
additions were absorbed out of that headroom (one requirement, one criterion),
not by widening the budget.

---

## §B Requirements

Requirements use GEARS notation. `<subject>` is generalized per the current
authoring standard.

### §B.1 The enacting act

- **REQ-AEC-001** (Unwanted) — The repository `.gitignore` shall not carry any negation re-including an artifact under a card report directory.
  The explanatory block introducing that negation shall be withdrawn with it, so
  no comment survives justifying a rule the file no longer contains.

- **REQ-AEC-002** (Ubiquitous) — This SPEC shall state that withdrawing the negation binds only files created afterwards.
  It shall state that ignore-matching does not untrack an already-tracked file,
  and shall record the tracked set, the remote-present set, and the two-way
  difference between them, each measured with a criterion anchored to the exact
  path shape rather than to a substring (§A.3).

- **REQ-AEC-003** (Unwanted) — The implementation shall not untrack, stage, or otherwise alter the index state of any file already tracked under `.moai/reports/`.
  The disposition of the already-tracked population is an open operator question
  (§D); this SPEC describes it and does not resolve it.

### §B.2 The wording repair — one act across every surface

- **REQ-AEC-004** (Ubiquitous) — Every sentence in the surfaces enumerated in the implementation plan's scope table that designates a card-report artifact as tracked, or as reaching the integration branch or the remote, shall be withdrawn.
  The prohibition binds the claim, not a phrase: a sentence carrying the claim in
  any verb form is in scope, and the surfaces are the plan's table rather than the
  four export-mandate copies alone.

- **REQ-AEC-005** (Ubiquitous) — The export-mandate tail in each auditor definition shall state that the mandated destination is local by design and that the verdict is read on disk rather than exported to the remote.
  The tail shall state that obligation without asserting the destination's current
  ignore state, so it stays true in a user project whose `.gitignore` this
  repository does not author.

- **REQ-AEC-006** (Unwanted) — No wording this SPEC specifies shall name a forced stage (`git add -f`), nor any other act whose effect is to place a card-report artifact on the remote.

- **REQ-AEC-007** (Ubiquitous) — The export-mandate clause shall retain its per-copy FORBIDDEN prohibition verbatim, and the convention's § Where bullet shall state a property that still distinguishes the forbidden directory once every card-report path is ignore-matched again.
  That bullet shall retain its pointer to the `.gitignore` comment, which is the
  convention document's only route to the directive this repair rests on.

- **REQ-AEC-008** (Unwanted) — The repaired wording shall not weaken the export obligation.
  An audit response without an exported file shall remain an incomplete audit in
  every copy.

### §B.3 Mirrors, emission, neutrality

- **REQ-AEC-009** (Ubiquitous) — Every change to the convention document shall be applied identically to its template mirror at `internal/template/templates/.moai/docs/audit-artifact-convention.md`, and the two copies shall remain byte-identical.

- **REQ-AEC-010** (Where) — Where an edit target lies under the template tree, the wording shall carry no forbidden content class.
  Specifically it shall contain no SPEC identifier, no requirement token, no
  internal date, no commit hash, no platform-specific absolute path, and no
  reference to a local-only development guide.

- **REQ-AEC-011** (Event-driven) — When a C2 agent copy is modified, the actor shall regenerate the emitted codex layer with `make agents-emit`, and shall not hand-edit any file under `internal/template/templates/.codex/`.

### §B.4 Constraints on the wording itself

- **REQ-AEC-012** (Unwanted) — No sentence this SPEC introduces shall assert a fact about this repository's tree state, in any verb form.
  Each introduced sentence shall be a statement of obligation, of design intent,
  or of general git behaviour. The verb-form clause is load-bearing: a stative
  form such as *remain*, *stay*, or *become* carries the same claim a copula
  does, and a criterion enumerating only the copula misses it.

- **REQ-AEC-013** (Ubiquitous) — The convention's § What makes the convention stick mechanical check shall name a check whose obligation still holds once the artifact is local.
  Presence on disk is that check; branch reachability shall not be named, because
  under the directive reaching the branch is no longer the obligation.

- **REQ-AEC-014** (Unwanted) — No check-form wording this SPEC specifies for a target document shall instruct its reader to decide a path's ignore status from the exit code of `git check-ignore -v`.
  A specified check that decides ignore status shall name the plain form, whose
  exit code answers that question (§A.2a). The prohibition binds **specified
  wording** — the sentences this SPEC writes into another document — and is
  therefore distinct from the instrument choice inside this SPEC's own criteria:
  a criterion applying the wrong instrument yields one wrong verdict, whereas
  specified wording carrying it is inherited by every future reader of the
  repaired file. The verbose form remains permissible where a sentence is
  reporting **which rule matched** rather than deciding whether a path is
  ignored.

---

## §C Specified wording

The wording below is normative.

### §C.1 Replacement tail — both auditor copies

Replaces the final sentence of each export-mandate clause, the sentence beginning
`Never write the verdict to a gitignored location`. Everything preceding it is
unchanged.

> This destination is local by design: the verdict stays on disk for the lead to
> read and is not exported to the remote, so do not force it into the tree or
> widen the ignore rules to admit it. The worktree therefore holds the only copy —
> do not dispose of it until the lead has read the verdict. One destination stays
> forbidden regardless: <FORBIDDEN-CLAUSE>.

Three specified sentences, one per obligation: the local-by-design statement with
its two prohibitions (REQ-AEC-005, REQ-AEC-006), the disposal consequence that
follows from it, and the preserved prohibition (REQ-AEC-007). No sentence asserts
the destination's ignore state, so the tail is true in a user project as well as
here (REQ-AEC-012).

`<FORBIDDEN-CLAUSE>` is taken verbatim from the copy being edited, preserving the
existing per-file difference:

- plan-auditor copies: ``the report directory the convention declares FORBIDDEN (`audit-artifact-convention.md` § Where) receives verdicts as disposal, not export``
- sync-auditor copies: ``` `.moai/reports/plan-audit/` is FORBIDDEN — writing there is disposal, not export ```

### §C.2 Replacement — convention § Committing

Replaces the section body.

> An audit artifact is a local file. It is not expected to reach the integration
> branch, and no convention forces it there — the lead reads it on disk, in the
> tree where the card was worked. A worktree holding the only copy of a verdict is
> therefore a disposal hazard rather than an untidiness: the tree is removed when
> the card closes, and the verdict goes with it. Do not dispose of a card's
> worktree until the lead has read its verdict.
>
> Ignore-matching does not untrack a file that is already tracked, so a project
> that tracked audit artifacts before adopting this convention keeps carrying
> those files until it removes them deliberately.

Sentence 1 withdraws both false sentences of the current § Committing rather than
qualifying them (REQ-AEC-004). The disposal-hazard sentence survives because it is
the live hazard under this direction, and gains the mitigation that actually
follows from it. The closing paragraph is general git behaviour plus a conditional
about *a* project, not an assertion about this one — which is what keeps it inside
REQ-AEC-012 while still carrying the explanatory load §A.3 needs.

### §C.3 Replacement — convention § Where, FORBIDDEN bullet justification

> - FORBIDDEN: `.moai/reports/plan-audit/`. Nothing written there is ever read as
>   a card's verdict, and that — not the ignore rules, which cover the sanctioned
>   card destinations too — is what distinguishes the directory; the repository's
>   `.gitignore` comment records the policy behind both. A verdict written there
>   is disposed of, not exported. Do not repurpose the directory.

The distinguishing property is restated as a fact about how the directory is
**read**, which survives the withdrawal of the negation. The ignore-based
justification does not survive it, for the reason measured in §A.5: the ignore
rules cover the sanctioned card destinations as well, so they distinguish
nothing. The `.gitkeep` clause is dropped as a tree-state assertion; the
`.gitignore` pointer is retained deliberately (REQ-AEC-007).

### §C.4 Replacement — convention § What makes the convention stick

Replaces the mechanical-check bullet.

> - **Mechanical check.** `ls .moai/reports/<card-id>/` — a missing verdict file
>   is the detection. Presence on disk is the whole obligation here: the artifact
>   is local by design, so a check for branch reachability would test something
>   the convention does not ask for.

This is the only check command §C specifies for a target document, and it names
no ignore query at all — which is how it satisfies REQ-AEC-014 rather than by
choosing between the two `check-ignore` forms. Should a later revision introduce
a specified sentence that *does* decide ignore status, §A.2a governs its form.

### §C.5 Replacement shape — the paraphrase surfaces

The four always-loaded-rule copies and the two `manager-lead` copies each
designate a card-report artifact as tracked (§A.6). Every such designation is
withdrawn (REQ-AEC-004). The replacement shape, applied per copy against that
copy's own surrounding text rather than as a shared block:

> The verdict file is the **local** record a claim cites — in this repository
> `.moai/reports/<card-id>/verdict.md`. It reaches no clone and no other machine,
> so citing it states where the deciding evidence was written, not where a reader
> elsewhere can fetch it. The obligation is therefore unchanged and its reason is
> narrower: **carry the deciding evidence into the verdict** — the command that
> decided a claim and the lines of output that decided it are written into the
> verdict file, and the claim cites that file.

The export obligation these surfaces carry is **preserved**, not weakened: what
changes is the reason given for it. No sentence in the replacement names a stage,
a branch, or a remote.

---

## §D Exclusions

### Out of Scope — the disposition of the already-tracked verdict files

- Whether the 12 tracked `verdict.md` files, or the 2 of them not yet on
  `origin/develop`, should be untracked, left in place, or removed from history.
  That question is escalated to the operator and is unanswered; §A.3 describes it
  and REQ-AEC-003 forbids resolving it here.
- Any `git rm --cached`, any staging change touching those paths, and any history
  rewrite.

### Out of Scope — the ignore policy beyond the negation

- Widening `.gitignore` to admit any card-report artifact class. The operator
  ruled the opposite direction; this card narrows, and never widens.
- `internal/template/templates/.gitignore`. It carries no verdict negation
  (measured, §A.1), so it already agrees with the directive and needs no edit.
- Whether audit evidence should live under `.moai/reports/` at all, and the
  five-section evidence format the clause already names.

### Out of Scope — adjacent documents

- Any agent definition other than `plan-auditor.md`, `sync-auditor.md`, and
  `manager-lead.md`.
- Any rule file that cross-references the export mandate without designating a
  card-report artifact as tracked.
- The emitted codex layer `internal/template/templates/.codex/agents/moai/*.toml`
  as an *edit* target. It is regenerated, never hand-edited (REQ-AEC-011); it is
  in scope only as a verification surface.

### Out of Scope — behavior changes

- Adding any mechanical enforcement (hook, lint rule, CI guard) for the local-only
  policy. This card is a documentation-truth repair plus the one ignore-rule
  withdrawal that makes it true; a mechanism is a separate decision with its own
  cost.
- Changing where auditors write, when they write, or what they write.

---

## §E Risks

| Risk | Consequence | Mitigation |
|---|---|---|
| The negation is withdrawn and the existing tracked files are read as also withdrawn | A reader believes the state is clean while 12 files still reach the remote | §A.3 states the mechanism, REQ-AEC-002 requires it in the SPEC, and AC-AEC-003 measures the population before and after |
| A future ignore-policy change makes the wording stale again | The same defect recurs | REQ-AEC-012 forbids asserting the ignore state, so no policy change can falsify the specified sentences |
| C1 and C2 drift during the edit | Template ships different wording than the repository dogfoods | Divergence is intentional by doctrine and measured (§A.4); §C specifies the replacement shape, applied per copy against its own surrounding text |
| The paraphrase repair is deferred to a successor card | An always-loaded rule asserts a fact this card makes false | Absorbed rather than excluded — §A.6 measures the coupling; §A.7 shows the budget accommodates it |
| A literal grep for a moved phrase returns a false zero | A surface is silently left unrepaired | §A.6 is the standing counter-example: every absence claim is paired with a control that fired in the same run |
| A reviewer counts 16 files and reads the card as Tier L | The Tier call flips at the band edge and the artifact set is re-derived mid-card | **The basis for 13 is that only hand-edited files are counted.** The three `.codex/*.toml` files are machine-emitted from C2 by `make agents-emit` and are never hand-edited (REQ-AEC-011), so they are an **output** of the change rather than a **surface** of it — no wording is authored in them and no judgement is applied to them. Counting them would count the same authoring act twice. 13 sits inside Tier M's 5-15 band; 16 would not, so the basis is stated here rather than left to the reader to reconstruct |
