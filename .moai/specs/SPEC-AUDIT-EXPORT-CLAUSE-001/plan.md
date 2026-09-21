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
follows, and every per-fixture exception below it. The `.gitignore` comment at
that blanket notes the order between the negation block and the plan-audit
carve-out is load-bearing; removing the upper block only relaxes an ordering
constraint, but the carve-out's own behaviour is re-measured after the edit
(AC-AEC-002) rather than assumed.

**`internal/template/templates/.gitignore` is not touched.** It carries no verdict
negation (`grep -c 'verdict.md'` → 0, SPEC §A.1), so it already agrees with the
directive.

### §A.3 What the withdrawal does NOT do, and the guard against reading it as more

Gitignore does not untrack a tracked file. 12 `verdict.md` files stay in the
index; 10 are already on `origin/develop`, 2 are not. **Their disposition is an
open operator question and this card does not touch them** (REQ-AEC-003). The risk
is not that someone removes them by accident — it is that a reader sees the
negation gone and concludes the state is clean. That is why the mechanism is
required in the SPEC body (REQ-AEC-002) and why AC-AEC-003 measures the population
rather than merely asserting it was left alone.

---

## §B The two-state constraint — why no specified sentence asserts the ignore state

A sentence saying "the destination is gitignored" is false in a user project whose
`.gitignore` this repository does not author, and would become false here again
under any future policy move. A sentence saying "the destination is tracked" is
what this card exists to remove. **A sentence stating an obligation or a design
intent is true in both**, which is why §C's replacements are obligation-shaped
rather than fact-shaped, and why REQ-AEC-012 is a prohibition rather than a style
note.

**The prohibition binds verb form, not vocabulary.** The v0.2.0 criterion
enumerated only the copula (`is`/`are`) and therefore missed `remain tracked` —
the assertion that was live in its own specified wording. The criterion here
(AC-AEC-014) covers the stative forms that carry the same claim, and its control
is required to fire before any zero is read.

---

## §C Scope surfaces — thirteen hand-edited files

| # | File | Change | Copy class |
|---|---|---|---|
| 1 | `.gitignore` | §A.2 withdrawal | repo-only |
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

---

## §D Milestones

### M1 — Withdraw the negation (Priority: High)

Apply §A.2 to file #1. Re-derive the line numbers at read time; do not act on the
numbers recorded in SPEC §A.2, which were measured at `64c7edbf3`.

Exit condition: no negation re-including a card-report artifact remains in
`.gitignore`; a new path under a card directory is ignore-matched again
(AC-AEC-001, AC-AEC-002); the tracked population is unchanged (AC-AEC-003).

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
each positive control preceding the absence claim it gates. Write the verdict to
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
| Negation withdrawn | anchored grep over `.gitignore` + control |
| The rule is in effect again | plain `git check-ignore --no-index` on a negated and a non-negated name + an out-of-tree control |
| Tracked population untouched | anchored `git ls-files` set + two-way `comm` against `origin/develop` + `git status --porcelain` on those paths |
| Four auditor copies repaired | anchored grep for the removed sentence (expect 0) + control; grep for the local-by-design statement |
| No remote-placing verb introduced | `git add -f` sweep over every changed file + control |
| Paraphrase surfaces repaired | the SPEC §A.6 markup-insensitive pattern (expect 0) + control |
| Convention mirror parity | `cmp` |
| Emitted codex layer regenerated | `make agents-emit-check` exit status + anchored grep on the three TOMLs |
| Template neutrality | forbidden-content-class sweep over the added lines of the three-dot diff + control |
| No specified sentence asserts tree state | stative-inclusive declarative sweep over the added lines of the three-dot diff + control |
| No repaired document decides ignore status with `-v` | verbose-form sweep over the twelve target documents + reachability control. **Regression-guard, vacuous against the current tree** — its red was observed against the v0.2.0 draft, not against this one (AC-AEC-015) |

---

## §F Anti-patterns for this card

- **Widening `.gitignore`.** This card narrows. If a repaired sentence seems to
  require an ignore exception, the sentence is wrong, not the policy.
- **Untracking the existing files.** `git rm --cached` on any of the 12 tracked
  verdicts is out of scope and is an open operator question (SPEC §D). Withdrawing
  the negation and untracking are two acts; only one is authorized here.
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
