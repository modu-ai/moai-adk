# SPEC-AUDIT-EXPORT-CLAUSE-001 — implementation plan

> Ordered by decision-reversibility. §A carries the one irreversible act — the
> `.gitignore` withdrawal — and the decision that scoped it. §B the constraint
> that binds every specified sentence. §C-§G the mechanical application.
> Review effort belongs at the top of this file.

---

## §A The irreversible act — withdrawing the negation

### §A.0 What changed since v0.2.0, and why it is not another iteration

v0.2.0 specified a `git add -f` permission clause: a verb that would place a card
verdict on the remote, declared permitted by the mandate. **The operator ruled
that the 2026-09-14 directive is canonical** — lead-verdict evidence stays on disk
and does not reach the remote — so that clause is withdrawn rather than repaired.
This is a change of basis, not a defect fix, and the consequences run one way
through everything below: §A.4b's deadlock framing dissolves (under the ruling the
plain `git add` refusal is the policy working), §D's Out of Scope entry inverts
(the `.gitignore` is edited, in the narrowing direction), and the paraphrase
surfaces excluded in v0.2.0 are absorbed (§C).

### §A.1 The decision: specification and execution stay in this card

The lead ruled it explicitly and the reasoning is measurable. If the
`.gitignore` edit were deferred, the documents this card repairs would say the
artifact is local while the ignore rules still admitted it — two artifacts saying
different things, which is the state this card exists to remove. The same
reasoning is why §C absorbs the paraphrase surfaces rather than deferring them:
`agent-common-protocol.md` is always-loaded and, since the `develop` absorb, its
wording *describes the negation mechanism directly* (SPEC §A.6). Deferring it
would leave an always-loaded rule asserting a fact this card's own edit falsifies.

### §A.2 Exactly what is withdrawn

Two adjacent regions of the repository `.gitignore`, measured at HEAD `64c7edbf3`
(line numbers re-derived at read time — they moved with the `develop` absorb and
will move again):

| Region | Content | Disposition |
|---|---|---|
| the explanatory block above the negation | the `# Narrow exception (card t1039)` comment and its depth/order rationale | withdrawn — a comment justifying an absent rule is the next reader's trap |
| the three-stage negation idiom | `!.moai/reports/*/` · `.moai/reports/*/*` · `!.moai/reports/*/verdict.md` | withdrawn |

**Not touched:** the directive comment block and its `.moai/reports/*` blanket
(they are the rule being restored to effect), the `plan-audit` carve-out that
follows, and every per-fixture exception below it — four file-level negations
(`t338/ac-count-baseline.txt`, `t528/probe/nondecl-bullets.txt`,
`t530/count-literals.txt`, `t229/live-probe-body.txt`) plus the directory-level
negations that enable them, measured surviving at HEAD `113e487c2`. That
preservation is why REQ-AEC-001 is stated as *verdict artifact* rather than
*artifact*: the broader wording would have been violated by the tree this plan
deliberately leaves alone, while AC-AEC-001's former glob-anchored probe was
structurally unable to see the violation. Requirement and criterion are narrowed
and widened respectively so their reach now matches.

The `.gitignore` comment at
that blanket notes the order between the negation block and the plan-audit
carve-out is load-bearing; removing the upper block only relaxes an ordering
constraint, but the carve-out's own behaviour is re-measured after the edit
(AC-AEC-002) rather than assumed.

**`internal/template/templates/.gitignore` is not touched.** It carries no verdict
negation (`grep -c 'verdict.md'` → 0, SPEC §A.1), so it already agrees with the
directive.

### §A.3 What the withdrawal does NOT do, and the guard against reading it as more

Gitignore does not untrack a tracked file. At the pre-act measurement 12
`verdict.md` files were in the index; 10 were already on `origin/develop`, 2 were
not. The risk is not that someone removes them by accident — it is that a reader
sees the negation gone and concludes the state is clean. That is why the
mechanism is required in the SPEC body (REQ-AEC-002) and why AC-AEC-003 measures
the population rather than merely asserting it was left alone.

**The disposition was subsequently ruled, and the ruling is recorded in SPEC
§A.3a — it is not this plan's to re-decide.** The operator ruled that the 2 files
not yet on `origin/develop` leave the index (files kept on disk, no history
rewrite) and that the 10 already published stay as history. That act is
independent of M1: the withdrawal untracked nothing, and the index removal
withdrew no rule. The discriminating property is **reachability**, not the file
class — a tracked card-report artifact that has not reached `origin/develop` may
be removed because the removal changes only what will be published; one that has
reached it may not, because the removal would not undo the publication.

**No milestone in this plan performs a further index change.** §A.3a's post-act
measurement records the tracked-but-not-remote set (`comm -23`) as empty, so the
predicate selects nothing; and the operator's 2026-09-22 disposition rules the
post-removal state final — the two removed files stay on `origin/develop` and
stay intentionally absent from this branch's index, with restoration and
re-staging forbidden. AC-AEC-003 therefore closes on the ruled difference —
every tracked verdict file on the remote, and the remote-only remainder **exactly
the two ruled names** — and any third name in either direction is a finding to
report rather than an act to perform (§B edge cases).

---

## §B The two-state constraint — why no specified sentence asserts the ignore state

A sentence saying "the destination is gitignored" is false in a user project whose
`.gitignore` this repository does not author, and would become false here again
under any future policy move. A sentence saying "the destination is tracked" is
what this card exists to remove. **A sentence stating an obligation or a design
intent is true in both**, which is why §C's replacements are obligation-shaped
rather than fact-shaped, and why REQ-AEC-012 is a prohibition rather than a style
note.

**The prohibition binds the claim, not a surface form — and the criterion has now
been defeated twice for treating them as the same thing.** The v0.2.0 form
enumerated only the copula (`is`/`are`) and missed `remain tracked`, the
assertion live in its own specified wording; the v0.3.1 form added the stative
verbs and was defeated by one interposed adverb (`is **already** tracked`).
AC-AEC-014 now covers both, and — because a regex over surface forms cannot
express REQ-AEC-012's actual discriminator, which is tense and subject — it
closes on its match set equalling a **declared 2-entry exception set** rather
than on a bare zero, and declares the class it still misses. Its control is
required to fire before any result is read.

**REQ-AEC-012's discriminator is stated in the requirement, not left to the
regex.** A generic statement of git behaviour (*"git does not untrack a file that
is already tracked"*) and a past-tense record of a completed act (*"the entries
already in the index stayed there"*) are both permitted: neither goes false under
a policy move, and neither is false in a user project. A present-tense claim
about *this* tree is what the prohibition is for.

---

## §C Scope surfaces — thirteen hand-edited files

| # | File | Change | Copy class |
|---|---|---|---|
| 1 | `.gitignore` | §A.2 withdrawal | repo-only — never ships, and **in AC-AEC-014's sweep** |
| 2 | `.claude/agents/moai/plan-auditor.md` | SPEC §C.1 tail, plan-auditor FORBIDDEN variant | C1 |
| 3 | `.claude/agents/moai/sync-auditor.md` | SPEC §C.1 tail, sync-auditor FORBIDDEN variant | C1 |
| 4 | `internal/template/templates/.claude/agents/moai/plan-auditor.md` | as #2 | C2 |
| 5 | `internal/template/templates/.claude/agents/moai/sync-auditor.md` | as #3 | C2 |
| 6 | `.moai/docs/audit-artifact-convention.md` | SPEC §C.2, §C.3, §C.4 | local |
| 7 | `internal/template/templates/.moai/docs/audit-artifact-convention.md` | byte-identical to #6 | template mirror |
| 8 | `.claude/rules/moai/core/agent-common-protocol.md` | SPEC §C.5 shape, applied to its own § Evidence export bullet | C1 (always-loaded) |
| 9 | `.claude/rules/moai/core/agent-common-protocol-reference.md` | SPEC §C.5 shape | C1 |
| 10 | `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` | as #8 | C2 |
| 11 | `internal/template/templates/.claude/rules/moai/core/agent-common-protocol-reference.md` | as #9 | C2 |
| 12 | `.claude/agents/moai/manager-lead.md` | SPEC §C.5 shape, two lines | C1 |
| 13 | `internal/template/templates/.claude/agents/moai/manager-lead.md` | as #12 | C2 |

Two copy relations behave differently and MUST NOT be conflated:

- **C1 ↔ C2 is hand-maintained and intentionally divergent.** Measured: `cmp -s`
  reports DIVERGENT for all three agent pairs (`plan-auditor`, `sync-auditor`,
  `manager-lead`) and for both rules pairs. Each copy is opened and edited on its
  own; no byte-comparison is asserted between them, and none is verified.
- **#6 ↔ #7 is byte-identical today and must remain so.** Measured `cmp` →
  BYTE-IDENTICAL, 6872 bytes each. `cmp` is the check (AC-AEC-011).

**C3 is never hand-edited.** `internal/template/templates/.codex/agents/moai/*.toml`
is machine-emitted from C2. Three of them carry the affected text — `plan-auditor.toml`
and `sync-auditor.toml` each carry the removed sentence once, and `manager-lead.toml`
carries the tracked-verdict phrase twice — and all three are corrected by
regeneration, never by hand (M4).

**AC-AEC-004's sweep covers rows #2–#13 only, enumerated per file** (operator
disposition 2026-09-22). Row #1, `.gitignore`, is outside that criterion's scope:
its act is the withdrawal (§A.2), measured by AC-AEC-001 and AC-AEC-002, and its
added lines are screened by AC-AEC-014 instead. The criterion does not sweep tree
roots — a whole-tree form reaches files outside this table (measured:
`internal/template/templates/.moai/README.md:77`, pre-existing since
`d97980654`, 2026-07-07).

---

## §D Milestones

### M1 — Withdraw the negation (Priority: High)

Apply §A.2 to file #1. Re-derive the line numbers at read time; do not act on the
numbers recorded in SPEC §A.2, which were measured at `64c7edbf3`.

Exit condition: no negation re-including a **verdict** artifact remains in
`.gitignore` (the four CI-fixture negations survive by design, §A.2); a new path
under a card directory is ignore-matched again (AC-AEC-001, AC-AEC-002); the
tracked verdict set equals the `origin/develop` verdict set, both differences
empty (AC-AEC-003).

**This milestone is ordered first deliberately.** It is the only irreversible act
in the card, and every wording change below is true only once it has landed.
Running it last would mean the documents spend the card's whole middle asserting
something the tree contradicts.

### M2 — Repair the four auditor copies (Priority: High)

Apply the SPEC §C.1 replacement tail to files #2-#5, one file at a time,
preserving each copy's existing FORBIDDEN clause verbatim. The tail is identical
across all four; only the FORBIDDEN variant differs, and it differs by auditor
role, not by copy class.

Exit condition: no copy carries the string `Never write the verdict to a
gitignored location`; every copy carries the local-by-design statement; no copy
names `git add -f` (AC-AEC-005, AC-AEC-006, AC-AEC-007).

### M3 — Repair the convention document and its mirror (Priority: High)

Apply SPEC §C.2, §C.3, and §C.4 to file #6, then apply the identical change to
file #7.

Exit condition: `cmp` reports the two files byte-identical; § Committing no longer
claims the artifact is tracked or reaches the integration branch; the § Where
bullet's distinguishing property no longer rests on the ignore rules and retains
its `.gitignore` pointer; the mechanical check names presence on disk.

**Applied to the mirror as a re-application, not a copy.** Writing #7 by copying
#6 wholesale is acceptable only because they are byte-identical *today* (verified,
SPEC §A.4 / §C table above). If that verification fails at implementation time,
the mirror is edited section-by-section and the divergence is reported as a
blocker rather than flattened.

### M4 — Repair the paraphrase surfaces (Priority: High)

Apply the SPEC §C.5 replacement shape to files #8-#13, each judged against its own
surrounding text. `agent-common-protocol.md` is always-loaded; its bullet is the
highest-reach sentence in the card and is edited first within this milestone so a
failure there surfaces before the lower-reach copies are touched.

Exit condition: no surface in the scope table designates a card-report artifact as
tracked, as reaching the integration branch, or as reaching the remote
(AC-AEC-004); the export obligation each surface carries is still present.

### M5 — Regenerate the emitted codex layer (Priority: High)

Run `make agents-emit`. Mandatory because M2 and M4 touched C2 copies of three
emitted agents; skipping it leaves the TOMLs carrying the withdrawn sentences, and
a binary built from that tree embeds the stale definitions.

Exit condition: `make agents-emit-check` exits 0, and the emitted layer no longer
carries the removed sentence or the tracked-verdict phrase (AC-AEC-012).

### M6 — Verification sweep (Priority: Medium)

Execute the acceptance criteria as a single-turn parallel read-only batch, with
each positive control preceding the absence claim it gates.

**Exit-condition ordering [HARD].** AC-AEC-013 and AC-AEC-014 are evaluated only
after the commits of M1-M5 have landed. Their three-dot diff reads commits, not
the working tree, so running them mid-edit returns empty for probe **and**
control and the zero-result rule refuses the close — an empty result there is
"not yet measurable", never "no violation found". Every other criterion in this
batch reads the tree directly and is order-independent.

Write the verdict to
`.moai/reports/t1059/verdict.md` and **leave it there** — under this card's own
direction that file is local evidence the lead reads on disk, and forcing it into
the tree would be the card refuting itself on its first exercise.

---

## §E Verification approach

### The instruments, and what is wrong with the obvious ones

Three instrument hazards were measured in this tree. Each shaped a criterion, and
each is recorded so the next author does not re-introduce the convenient form.

**1. `git check-ignore -v` returns exit 0 on a negation match.** The property,
its measurement, and the reason an ordinary probe cannot see it are stated once
in **SPEC §A.2a** — that section is the SSOT and is not restated here. It was
promoted out of this plan into the SPEC body at v0.3.1 precisely because it binds
more than this card's instruments: it also governs any check-form wording the
SPEC specifies for a target document (REQ-AEC-014, AC-AEC-015).

The operational consequences for this plan:

- **AC-AEC-002 uses the plain form** and probes both a negated and a non-negated
  name. A criterion keyed on `-v` exit 0 meaning "ignore-matched" is a wrong
  instrument, and it is wrong *silently* — the two forms agree on every
  non-negated path, so a probe against `plan-audit.md` alone certifies it.
- **Nothing in §C may specify the verbose form to a reader.** §C.4 is the only
  specified check command and it names no ignore query at all.

**2. A substring match on `verdict.md` over-counts.** It also matches
`plan-audit-verdict.md`, inflating the remote set from 10 to 11 (SPEC §A.3). Both
sides of the comparison use the anchored criterion, and the criterion reports the
**two-way set difference** rather than the counts — the counts differed by 1 and
the true shape was 2 and 0.

**3. A literal grep for a phrase that has moved returns a clean zero.** The
v0.2.0 exclusion rested on `tracked** path`, which the `develop` absorb rewrote out
of existence; the phrase now returns 0 across every tree (SPEC §A.6). Every
absence claim in this card is paired with a control that fired **in the same run**,
and a control returning zero voids every zero beside it.

### Two further shape constraints on the criteria

- **Revision-pinned, not working-tree-scoped.** Criteria that sweep the lines this
  card adds use `git diff develop...HEAD -- <files>`. The working-tree form
  (`git diff -- <files>`, no revision) returns empty the moment the change is
  staged, so both the probe and its control go to zero and the criterion cannot be
  closed under its own zero-result rule. The three-dot form re-derives the merge
  base at read time, so it survives staging, commit, and a later absorb of
  `develop`. `develop` is kept as a moving ref here deliberately (`verification-claim-integrity.md`
  §2.1, SUBJECT class): the claim is *what this card changed relative to its branch
  point*, and pinning a literal SHA would falsify it the first time the card
  absorbs `develop`.

  **[HARD] It is nevertheless empty before the change is committed, and that
  window is named here rather than left to be rediscovered.** `develop...HEAD`
  compares the merge base to the **commit** `HEAD`, so an uncommitted
  working-tree edit is invisible to it whether staged or not. During M2-M4, before
  the milestone's commit lands, AC-AEC-013 and AC-AEC-014 return empty for both
  probe and control, and §D.3's zero-result rule correctly refuses the close.
  Measured at HEAD `113e487c2` with only M1 committed: the AC-AEC-014 pathspec
  yields `0` added lines without `.gitignore` and `39` with it. **These two
  criteria close only after their milestone's commit lands** — see M6's exit
  condition. The earlier statement that the form *"survives staging, commit, and a
  later absorb of `develop`"* is true and incomplete: it names what the form
  survives and not what precedes it, and the consequence is a blocked close rather
  than a false pass, which is why the ordering is stated rather than mechanised.
- **One compound invocation, no command substitution, no tree writes.** The
  worktree guard refuses a git command it cannot statically verify, which is what
  made the v0.2.0 AC-AEC-013 unrunnable in the environment where cards are worked.
  `$(git merge-base …)` is refused for that reason; the three-dot form is the
  substitution-free equivalent and was confirmed to run here. The ignore probe was
  additionally reduced so it **writes nothing**: `git check-ignore --no-index`
  answers for a path that does not exist, so the criterion needs no `mkdir`, no
  file, and no cleanup — which removes the `rm -rf` on a shell-computed path
  entirely rather than making it safer.

### Verification surfaces

| Surface | Instrument |
|---|---|
| Verdict negation withdrawn | verdict-name-anchored grep over `.gitignore` (`^!\.moai/reports/[^/]*/.*verdict`) + control; red pinned at `269fb89c1` |
| The rule is in effect again | plain `git check-ignore --no-index` on a negated and a non-negated name + an out-of-tree control + a two-form instrument self-check on a surviving fixture negation |
| Tracked set vs remote set | anchored `git ls-files` set + two-way `comm` against `origin/develop`, `comm -23` **empty** and `comm -13` **exactly the two ruled removals** + `git status --porcelain` on those paths |
| Four auditor copies repaired | anchored grep for the removed sentence (expect 0) + control; grep for the local-by-design statement |
| No remote-placing verb introduced | `git add -f` sweep over every changed file + control |
| Paraphrase surfaces repaired | the SPEC §A.6 markup-insensitive pattern (expect 0) + control, swept over the **twelve wording surfaces** (scope rows #2–#13, enumerated per file); RED re-derived against the pinned pre-repair blobs at `269fb89c1` |
| Convention mirror parity | `cmp` |
| Emitted codex layer regenerated | `make agents-emit-check` exit status + anchored grep on the three TOMLs |
| Template neutrality | forbidden-content-class sweep over the added lines of the three-dot diff + control |
| No specified sentence asserts tree state | stative- and adverbial-inclusive declarative sweep over the added lines of the three-dot diff across **all thirteen** files (`.gitignore` included) + control, closing on the match set equalling the AC's declared 2-entry exception set; own mutant probe re-run at close |
| No repaired document decides ignore status with `-v` | verbose-form sweep over the twelve target documents + reachability control. **Regression-guard, vacuous against the current tree** — its red was observed against the v0.2.0 draft, not against this one (AC-AEC-015) |

---

## §F Anti-patterns for this card

- **Widening `.gitignore`.** This card narrows. If a repaired sentence seems to
  require an ignore exception, the sentence is wrong, not the policy.
- **Untracking a verdict file that is already on `origin/develop`.** Forbidden by
  REQ-AEC-003 in every case: an index removal does not undo a publication, so it
  buys nothing and loses the record. The operator ruled that published history is
  preserved rather than rewritten (SPEC §A.3a).
- **Reading §A.3a as a general licence to untrack.** It authorized exactly the
  tracked-but-not-remote set, which its own post-act measurement records as now
  empty. Withdrawing the negation and removing from the index remain two
  independent acts; neither implies the other, and neither is a reason to perform
  the other.
- **Rewriting history on any `.moai/reports/` path**, including the two files
  §A.3a authorized removing from the index. The ruling permitted an index change
  and nothing else; the files stay on disk and the commits that carried them stay
  in the graph.
- **Reading the withdrawal as having cleaned the state.** It binds files created
  afterwards. The mechanism is stated in SPEC §A.3 precisely because this is the
  silent misreading.
- **Deferring the always-loaded surface.** `agent-common-protocol.md` reaches
  every session and now describes the negation mechanism directly; leaving it for
  a successor card means an always-loaded rule asserts a fact this card's own edit
  falsified.
- **Hand-editing a `.codex/*.toml`.** Overwritten at the next emission, and the
  edit's existence is erased with it.
- **Asserting byte-parity between C1 and C2.** Measured DIVERGENT for all five
  pairs in the scope table; the divergence is intentional by doctrine.
- **Reading a grep zero as absence.** SPEC §A.6 is the standing counter-example in
  this very card — a phrase that had simply moved.
- **Force-staging this card's own verdict.** The card's direction says the verdict
  is local. Exporting it to the remote as the card's first exercise of its own
  wording would refute it.

---

## §G Cross-references

- `.gitignore` — the directive comment block (the decision this card implements)
  and the negation withdrawn by M1
- `.moai/docs/audit-artifact-convention.md` — the document repaired by M3
- `.claude/agents/moai/plan-auditor.md`, `.claude/agents/moai/sync-auditor.md` —
  the definitions repaired by M2
- `.claude/rules/moai/core/agent-common-protocol.md` — the always-loaded surface
  repaired by M4, and the § Evidence export obligation whose *reason* changes
  while the obligation itself does not
- `.claude/rules/moai/core/verification-claim-integrity.md` — §1 (no unobserved
  claim), §2.1 (the moving-ref ANCHOR/SUBJECT predicate applied in §E)
- `.claude/rules/moai/development/verification-completeness.md` — §1.1 (an empty
  swept set asserts nothing), §2 (the mutant probe)
- `.claude/rules/moai/workflow/kanban-dispatch.md` § Completion is read, never
  trusted — the lead-side reader this repair serves, and the reason a local
  verdict is sufficient for the obligation it carries
