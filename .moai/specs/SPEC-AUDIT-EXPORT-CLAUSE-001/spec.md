---
id: SPEC-AUDIT-EXPORT-CLAUSE-001
title: "documents that promise a remote destination the directive forbids"
version: "0.3.4"
status: completed
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

- 2026-09-22 · v0.3.3 · manager-spec · **Disclosure repair after the plan-audit
  PASS at 0.8125 (reset iteration 2, Tier M ceiling), on the operator ruling
  that the three blocking findings close before run-phase entry.** No
  requirement and no direction changed; all three repairs are in the layer that
  records what this SPEC did rather than in what it specifies. **The
  load-bearing one: v0.3.2 relaxed REQ-AEC-012 after the tree had already
  exercised the relaxation, and said so nowhere.** §A.3b now records that — the
  measured timeline, the sentence the relaxation legalizes, and the reason the
  narrowing is principled on REQ-AEC-012's own rationale rather than
  accommodating. Alongside it: AC-AEC-014's Then-clause is restated as set
  equality, matching its own Expected line and §D.1 item 7, which had stated
  the same closing condition at two incompatible strengths; and
  `acceptance.md`'s preamble no longer asserts a conformance
  (`verification-completeness.md` §2.1's single-invocation form) that ten of the
  fifteen criteria do not meet, with the deviation and its forced cause now
  named in the artifact instead of only outside it. Two optional findings are
  also closed: AC-AEC-003 cites the moving-ref predicate it applied, and
  AC-AEC-014's successor-card sentence is disambiguated. The v0.3.2 entry below
  is left **verbatim and unamended** — it is an incomplete record, not a false
  one, and retrofitting it to match what was later discovered is the move §A.3
  and the preserved §A.3 measurement table exist to refuse. Requirements stay
  14, criteria stay 15 — no budget was drawn.

- 2026-09-22 · v0.3.2 · manager-spec · **Plan-audit reset-iteration-1 repair
  (FAIL 0.75), and the operator ruling that made REQ-AEC-003 stale.** Two
  critical defects drove the FAIL and both trace to one event: after v0.3.1 was
  committed, the operator ruled on the disposition §D had recorded as
  *escalated and unanswered* — `.moai/reports/t1039/verdict.md` and
  `.moai/reports/t1048/verdict.md`, neither of which had reached
  `origin/develop`, were to leave the index while the 10 already on the remote
  were left as history. That ruling was executed without being folded back into
  this SPEC, so v0.3.1 forbade in four places what the tree already contained
  (D1) and AC-AEC-003's expected figures were falsified by it (D2).
  REQ-AEC-003 is rewritten to authorize exactly what was ruled, on the
  reachability predicate rather than on a file list; §D records the resolved
  disposition; AC-AEC-003 closes on set equality against `origin/develop`
  instead of on counts. Six criterion-mechanics defects are repaired alongside:
  REQ-AEC-001 narrowed to the verdict negation it always meant with the
  surviving CI-fixture negations stated as a carve-out (D3), AC-AEC-008 made
  grep-flavour-independent (D4), AC-AEC-015's and §A.2a's RED cells pinned to
  `622e25d22` (D5), AC-AEC-014's sweep corrected to the full thirteen files and
  its regex widened past the adverb gap (D6, D7), AC-AEC-001's stated
  expectation matched to `grep -c`'s actual behaviour (D8), and the pre-commit
  window of the three-dot diff named in both artifacts (D9). REQ-AEC-012 gains
  the tense discriminator the widened regex needs to be principled rather than
  ad hoc. Requirements stay 14, criteria stay 15 — no budget was drawn.

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

The citation is pinned to the literal commit that carries the v0.2.0 draft,
`622e25d22`, not to `HEAD`. `HEAD` is an address that moves: this same command
written against it returned `4` matching lines when it was authored and returns
`1` now, because the v0.3.x rewrites landed in between. Pinning is the ANCHOR
remedy of `verification-claim-integrity.md` §2.1 — the ref here is the address
at which a measurement was taken, not the subject of the claim.

```
$ git show 622e25d22:.moai/specs/SPEC-AUDIT-EXPORT-CLAUSE-001/acceptance.md \
    | grep -n 'check-ignore' | cut -c1-96
90:grep -rc 'check-ignore -v --no-index' \
99:A match of `check-ignore` **without** `--no-index` in these files is a FAIL, not
177:Expected output contains, in this order: `git check-ignore -v --no-index`,
312:git check-ignore -v --no-index "$P"; CHECK=$?; echo "check_exit=$CHECK"
```

exit code 0; four lines, re-measured at HEAD `113e487c2` against the pinned blob.

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

That the population was *later* changed by a separate, separately-ruled act
(§A.3a) does not weaken this: the two acts are independent, and confusing them
is the misreading this section exists to prevent. The withdrawal untracked
nothing; an index removal untracked two files; neither implies the other.

Measured **before the index act of §A.3a**, in this tree at `64c7edbf3`, with an
anchored criterion applied to **both** sides:

| Measurement | Command | Result |
|---|---|---|
| tracked `verdict.md` | `git ls-files '.moai/reports/*/verdict.md' \| wc -l` | **12** |
| present on `origin/develop`, same criterion | `git ls-tree -r --name-only origin/develop -- .moai/reports \| grep -E '^\.moai/reports/[^/]+/verdict\.md$' \| wc -l` | **10** |
| tracked but not on the remote | `comm -23` of the two sorted sets | **2** — `.moai/reports/t1039/verdict.md`, `.moai/reports/t1048/verdict.md` |
| remote-only | `comm -13` of the two sorted sets | **0** (the local set was a proper superset) |

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

**[Corrected 2026-09-22 — see §A.3a.]** The two `exit=128` readings above were
measured against the remote-tracking ref as it stood at `64c7edbf3`, before an
explicit fetch. Re-probed after `git fetch origin develop`, both paths return
exit 0: the files had already landed on `origin/develop` when the index act was
performed. The table and the block above are dated pre-act measurements and are
left as taken; the corrected reachability record lives in §A.3a.

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
two-way set difference was **2 and 0**, and it is the two-way form — not the
counts — that exposed the real shape. Any criterion comparing these sets uses the
anchored pattern on both sides and reports the two differences, never the counts
alone.

### §A.3a [HARD] The ruled index act — what was authorized, and the predicate that bounds it

The v0.3.0 and v0.3.1 drafts recorded the disposition of the 12 tracked files as
**escalated to the operator and unanswered**, and forbade resolving it. The
operator has since ruled, and the ruling was executed. This section records what
was ruled so that the SPEC describes the tree rather than contradicting it.

**What was authorized and performed.** Exactly two paths left the index —
`.moai/reports/t1039/verdict.md` and `.moai/reports/t1048/verdict.md` — by
`git rm --cached`. Both files remain on disk; neither was deleted, and no history
was rewritten.

**The premise under which they were removed, and its correction.** The ruling was
executed on the understanding that the two files had not yet reached
`origin/develop` — the §A.3 measurement recorded `exit=128` for both. That
premise was measured against a stale remote-tracking ref. Re-probed after an
explicit `git fetch origin develop`, both paths are present on `origin/develop`:
`.moai/reports/t1039/verdict.md` landed via `b765153ae` (2026-09-21 17:24) and
`.moai/reports/t1048/verdict.md` via `d72df7454` (2026-09-21 22:03). The
removals therefore acted on paths that were already published.

**The operator's disposition of 2026-09-22 rules the resulting state final.**
The index stands as the removals left it: the two files stay on `origin/develop`,
where their publication is preserved history, and stay **intentionally absent
from this branch's index**. Restoration and re-staging of either file is
forbidden — REQ-AEC-003's first clause (no index alteration on any
already-published path) prohibits it in both directions, so the frozen state is
what the requirement itself enforces. No further index change under
`.moai/reports/` is in scope.

**The discriminating property is reachability, not the file class** — for a file
this SPEC never saw. A future reader applying this SPEC to a new file applies
that predicate, not the two names:

> A tracked card-report artifact that has **not** reached `origin/develop` may be
> removed from the index, because doing so changes only what will be published.
> One that **has** reached it may not, because the index removal would not undo
> the publication and the operator ruled that history is preserved rather than
> rewritten.

For the two named files the disposition above governs instead: their state was
ruled final on the actual facts after the premise error was discovered, and
re-opening it is a new operator question, not an application of the predicate.

**What stays forbidden.** The verdict files present on `origin/develop` are
untouched — the operator chose to preserve that history — and history rewriting
of any kind remains out of scope in every case, including for the two above
(REQ-AEC-003, §D).

**Post-act measurement**, re-taken at HEAD `ce6de9407` (2026-09-22, this run),
same anchored criterion on both sides:

| Measurement | Result |
|---|---|
| tracked `verdict.md` | **10** |
| present on `origin/develop`, same criterion | **12** |
| tracked but not on the remote (`comm -23`) | **0** — empty |
| remote-only (`comm -13`) | **2** — exactly `.moai/reports/t1039/verdict.md`, `.moai/reports/t1048/verdict.md` |

The two sets differ by exactly the two ruled removals, which is the state the
disposition rules final and the shape AC-AEC-003 closes on. The counts alone
would not say so: two files leaving the index and two different files arriving on
the remote would also read `10` and `12`, which is why the closing evidence is
the exact two-way difference and not the pair of counts.

### §A.3b [HARD] REQ-AEC-012 was relaxed after the tree had already exercised the relaxation

§A.3a records an index act ruled by the operator and then folded back into this
SPEC. This section records the **second** instance of the same shape, and it is
the one that went unrecorded until the reset-iteration-2 audit measured it. It is
written here because a SPEC whose §0 condemns silent recording failures in other
documents cannot carry one of its own.

**What moved, measured.** Three commits, in this order:

| Commit | Time | What it carried |
|---|---|---|
| `cf45febae` | 03:51:19 | v0.3.1 REQ-AEC-012: *"…shall assert a fact about this repository's tree state, **in any verb form**. Each introduced sentence shall be a statement of obligation, of design intent, or of general git behaviour."* — three permitted classes, no tense qualifier |
| `113e487c2` | 03:52:50 | the `.gitignore` withdrawal, whose added comment carries *"…so the entries already in the index **stayed there and had to be removed deliberately**."* |
| `309900337` | 04:24:03 | v0.3.2 REQ-AEC-012: adds *"in the present tense"*, narrows the object to this repository's **current** tree state, and adds a **fourth** permitted class — *past-tense narration of a completed act* |

**The consequence, stated plainly.** Under the v0.3.1 wording the
`stayed there and had to be removed deliberately` clause was a **violation**: it
asserts a fact about this repository's tree state, it is a verb form, and it is
none of the three classes then permitted. The clause that makes it compliant was
authored 31 minutes after the line landed, by the same agent, in the same card.
**This is a requirement moving to match landed execution, not execution moving to
match a requirement.**

**Why the relaxation nevertheless stands.** REQ-AEC-012's own stated rationale is
that *a present-tense claim about this tree goes false under a policy move or in a
user project whose `.gitignore` this repository does not author*. A past-tense
record of a completed act does not go false under either hazard — the act
happened, and it stays having happened whatever the policy does next. The v0.3.1
wording was therefore **wider than its own purpose**, and narrowing it to the
purpose is a repair of a criterion wider than its intent rather than an
accommodation of a convenient line. The requirement is correct as it now stands
and is not changed by this revision.

**What a later reader may do with this.** Disagree with it. The record exists so
that the relaxation can be re-opened on its merits: if a successor decides that
past-tense narration of this tree's state is also a hazard, the sentence to
re-examine is REQ-AEC-012's fourth permitted class, and the line that would then
become non-compliant again is the `.gitignore` comment quoted above. Nothing here
is load-bearing for the direction; only for the record of how the requirement
reached its current wording.

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

- **REQ-AEC-001** (Unwanted) — The repository `.gitignore` shall not carry any negation re-including a **verdict artifact** under a card report directory.
  The explanatory block introducing that negation shall be withdrawn with it, so
  no comment survives justifying a rule the file no longer contains.
  **Carve-out — the CI-fixture negations survive, and are not in scope.** Four
  negations re-include a named test-guard fixture under a card report directory
  (`.moai/reports/t338/ac-count-baseline.txt`,
  `.moai/reports/t528/probe/nondecl-bullets.txt`,
  `.moai/reports/t530/count-literals.txt`,
  `.moai/reports/t229/live-probe-body.txt`), together with the directory-level
  negations that enable them. They are operator-approved and CI reads them, as
  `.gitignore` already records at the head of that block (*"test guard fixtures
  under reports stay tracked — CI reads them"*). This requirement is stated as
  *verdict artifact* rather than *artifact* precisely so that it and AC-AEC-001
  have the same reach: a requirement forbidding every negation would be violated
  by a tree this card deliberately preserves, and a criterion blind to that
  violation would report green over it.

- **REQ-AEC-002** (Ubiquitous) — This SPEC shall state that withdrawing the negation binds only files created afterwards.
  It shall state that ignore-matching does not untrack an already-tracked file,
  and shall record the tracked set, the remote-present set, and the two-way
  difference between them, each measured with a criterion anchored to the exact
  path shape rather than to a substring (§A.3).

- **REQ-AEC-003** (Unwanted) — The implementation shall not alter the index state of any file under `.moai/reports/` that has already reached `origin/develop`, and shall not rewrite history for any such file in any case.
  **Where** a tracked card-report artifact has not reached `origin/develop`, its
  removal from the index is permitted by the operator ruling of §A.3a, the files
  shall remain on disk, and no history rewrite accompanies it. Exactly two
  paths were removed under that ruling — `.moai/reports/t1039/verdict.md` and
  `.moai/reports/t1048/verdict.md` — and the operator's disposition of
  2026-09-22 rules the resulting state final: both paths are present on
  `origin/develop` and intentionally absent from this branch's index, and the
  first clause above forbids restoring or re-staging either one. No further
  index change under `.moai/reports/` is authorized by this SPEC without a new
  ruling.

### §B.2 The wording repair — one act across every surface

- **REQ-AEC-004** (Ubiquitous) — Every sentence in the wording surfaces enumerated in the implementation plan's scope table (rows #2–#13) that designates a card-report artifact as tracked, or as reaching the integration branch or the remote, shall be withdrawn.
  The prohibition binds the claim, not a phrase: a sentence carrying the claim in
  any verb form is in scope, and the surfaces are the plan's table rather than the
  four export-mandate copies alone. The repository `.gitignore` (scope row #1) is
  outside this requirement's reach: its act is the withdrawal (REQ-AEC-001),
  measured by AC-AEC-001 and AC-AEC-002, so pre-existing prose in it is not swept
  here — its added lines are screened by AC-AEC-014 instead.

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

- **REQ-AEC-012** (Unwanted) — No sentence this SPEC introduces shall assert, in the present tense, a fact about **this repository's current** tree state, in any verb form.
  Each introduced sentence shall be a statement of obligation, of design intent,
  of general git behaviour, or of past-tense narration of a completed act.
  Three clauses are load-bearing and each was added because a criterion missed
  the class it names:
  - **Verb form.** A stative form such as *remain*, *stay*, or *become* carries
    the same claim a copula does, and a criterion enumerating only the copula
    misses it.
  - **Adverbial interposition.** *is already tracked* carries the same claim as
    *is tracked*; a criterion requiring the verb and the participle to be
    adjacent is defeated by one word.
  - **Tense and subject.** The prohibition exists because a present-tense claim
    about this tree goes false under a policy move or in a user project whose
    `.gitignore` this repository does not author. Neither hazard reaches a
    generic statement of how git behaves (*"git does not untrack a file that is
    already tracked"*) nor a past-tense record of what was done (*"the entries
    already in the index stayed there and had to be removed deliberately"*).
    Both are permitted, and AC-AEC-014 carries them as a declared, enumerated
    exception set rather than as silent regex misses. **The tense clause and the
    past-tense permission were added in v0.3.2, after a line exercising them had
    already landed; §A.3b records that timeline and the reason the narrowing
    stands.**

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

### Out of Scope — the already-remote verdict files and every history rewrite

- The `verdict.md` files present on `origin/develop` (12 at the 2026-09-22
  re-measurement). Their disposition was ruled and then re-ruled final at the
  2026-09-22 operator disposition: they stay as they are, because an index
  removal would not undo the publication and the operator chose to preserve that
  history rather than rewrite it (§A.3a). Re-opening it is a new operator
  question, not this card's.
- Any history rewrite, on any path under `.moai/reports/`, including the two
  files whose index removal §A.3a records. The ruling permitted an index
  change and nothing else; the files stay on disk and the commits that carried
  them stay in the graph.
- Restoration or re-staging of `.moai/reports/t1039/verdict.md` or
  `.moai/reports/t1048/verdict.md`. The 2026-09-22 disposition rules the
  post-removal state final: both are on `origin/develop` and intentionally
  absent from this branch's index, and REQ-AEC-003's first clause forbids any
  index alteration on an already-published path in both directions.
- Any further index change under `.moai/reports/`. §A.3a's post-act measurement
  records the tracked-but-not-remote set (`comm -23`) as empty, so the predicate
  that permitted the two removals now selects nothing; a further change needs a
  new ruling and is not covered here (REQ-AEC-003).

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
| The negation is withdrawn and the existing tracked files are read as also withdrawn | A reader believes the state is clean while tracked files still reach the remote | §A.3 states the mechanism, REQ-AEC-002 requires it in the SPEC, and AC-AEC-003 measures the population before and after |
| The §A.3a ruling is read as a general licence to untrack card-report artifacts | A later card removes a file already published on the remote, which the ruling forbids and which an index removal cannot undo | §A.3a states the predicate (reachability), not a file list; REQ-AEC-003 forbids the already-remote case and every history rewrite; AC-AEC-003 closes on the ruled difference (`comm -23` empty, `comm -13` exactly the two ruled names), so a change on either side breaks it immediately |
| A future ignore-policy change makes the wording stale again | The same defect recurs | REQ-AEC-012 forbids asserting the current ignore state, so no policy change can falsify the specified sentences |
| A criterion enumerating surface forms is read as complete | A sentence carrying the forbidden claim in an unenumerated form passes | REQ-AEC-012 names the three classes that have already defeated a criterion here (verb form, adverbial interposition, tense/subject) and AC-AEC-014 carries a declared exception set, so a permitted match is recorded rather than silently tolerated |
| C1 and C2 drift during the edit | Template ships different wording than the repository dogfoods | Divergence is intentional by doctrine and measured (§A.4); §C specifies the replacement shape, applied per copy against its own surrounding text |
| The paraphrase repair is deferred to a successor card | An always-loaded rule asserts a fact this card makes false | Absorbed rather than excluded — §A.6 measures the coupling; §A.7 shows the budget accommodates it |
| A literal grep for a moved phrase returns a false zero | A surface is silently left unrepaired | §A.6 is the standing counter-example: every absence claim is paired with a control that fired in the same run |
| A reviewer counts 16 files and reads the card as Tier L | The Tier call flips at the band edge and the artifact set is re-derived mid-card | **The basis for 13 is that only hand-edited files are counted.** The three `.codex/*.toml` files are machine-emitted from C2 by `make agents-emit` and are never hand-edited — the property REQ-AEC-011 requires and AC-AEC-012 measures — so they are an **output** of the change rather than a **surface** of it — no wording is authored in them and no judgement is applied to them. Counting them would count the same authoring act twice. 13 sits inside Tier M's 5-15 band; 16 would not, so the basis is stated here rather than left to the reader to reconstruct |
