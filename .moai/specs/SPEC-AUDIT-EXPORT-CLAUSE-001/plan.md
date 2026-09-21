# SPEC-AUDIT-EXPORT-CLAUSE-001 — implementation plan

> Ordered by decision-reversibility. §A carries the one genuine design decision;
> §B the two-state constraint that binds it; §C-§F the mechanical application.
> Review effort belongs at the top of this file.

---

## §A The design decision — what the clause should say instead

### §A.0 The mechanism this argument rests on

v0.1.0 argued its shape from a premise that was measured false: that a plain
`git add` skips an ignore-matched path silently. It refuses loudly — exit 1,
naming `-f` (SPEC §A.4b). The corrected mechanism is a **deadlock**: the compliant
actor receives a correct refusal and a remedy that no sanctioned instruction
permits (`grep -rn 'add -f'` across the rule, agent, skill, and doc trees returns
zero, with a firing positive control — SPEC §A.4b).

This changes what the replacement must carry, and the change is not cosmetic.
Under the false premise the clause had to **raise an alarm git was not raising**.
Under the measured one git already raises it, and the clause's job is narrower
and different: **grant the permission** the reader lacks, and remove the
prohibition that contradicts it. Every sentence below is re-derived against that.

### §A.1 What the four specified sentences buy

| # | Sentence | Bound by | Why it survives the corrected mechanism |
|---|---|---|---|
| 1 | the check | REQ-AEC-002 | The only form true in both ignore-policy states (§B), and the only way the reader knows *when* the permission applies |
| 2 | consequence + `git add -f` + the permission clause | REQ-AEC-001, REQ-AEC-003 | Carries the whole repair. The permission clause is what resolves the deadlock; without it `-f` reads as defiance of the surrounding sentence |
| 3 | the `--no-index` parenthetical | REQ-AEC-002 | The flag inverts the check's answer on a tracked file — measured: without it exit 1, with it exit 0 and the rule printed (SPEC §A.5). A reader free to read it as noise drops it and the instrument lies |
| 4 | the preserved FORBIDDEN prohibition | REQ-AEC-005 | True existing content; dropping it would let the new check be read as licensing that directory |

**Two v0.1.0 sentences are gone.** The separate two-obligation generalization
(*"Writing the file and reaching the branch are separate obligations…"*) restated
what sentence 2's own consequence clause says, and REQ-AEC-003 is rewritten to
require the consequence **inside** that sentence rather than a sentence of its
own. The `--no-index` parenthetical is cut from roughly thirty words to its
operative clause.

### §A.2 Re-argued against a fair minimal alternative

The v0.1.0 rejection compared against a draft that both asserted an ignore state
and carried no instrument — an unfair comparison, since a minimal variant doing
neither is constructible. Stated fairly, it is:

> *"Run `git check-ignore -v --no-index <path>`; on exit 0 a plain `git add`
> refuses the path, so stage it with `git add -f <path>`."*

**This variant is sound, and the specified wording is it plus exactly two things.**
Neither is a preference:

- **The FORBIDDEN prohibition (sentence 4).** Not an addition at all — it is the
  one true clause already in the sentence being replaced. Dropping it would
  silently delete a live prohibition under cover of a wording repair, which
  REQ-AEC-005 forbids.
- **The `--no-index` justification (sentence 3), ~13 words.** The flag's omission
  inverts the answer on exactly the paths where a verdict already exists, and that
  is measured, not asserted.

The honest verdict on length is therefore narrower than v0.1.0 claimed: the
argument does **not** establish that four sentences are required over the minimal
variant's one. It establishes that the minimal variant plus one preserved clause
plus one short parenthetical is the floor, and that the specified wording is at
that floor. The creep the review identified — one restating sentence and an
over-long parenthetical — is removed rather than defended.

**Rejected: rewriting the whole clause.** The mandate's first three sentences
(the destination list, the incomplete-audit declaration, the minimum-content
list) are true, load-bearing, and already carry the convention cross-reference.
Only the final sentence is false. The edit is scoped to that sentence.

---

## §B The two-state constraint — why no sentence may assert the ignore state

A pending sibling card narrows the ignore exception so that one filename under
the card evidence directory becomes un-ignored while the audit-artifact family
stays matched. This SPEC must not assume it has landed.

| State | `.moai/reports/<card>/plan-audit.md` | `.moai/reports/<card>/verdict.md` |
|---|---|---|
| Today | ignore-matched (measured, SPEC §A.2) | ignore-matched (measured, SPEC §A.2) |
| After the sibling card | ignore-matched | not matched — **unverified** |

[UNVERIFIED] The right-hand column's specific claim — *which* filename the
sibling card un-ignores — is derived from the dispatch that opened this card, not
from an artifact readable in this tree: the sibling is named only in prose, with
no SPEC ID or card id to resolve. It is labelled rather than removed because the
conclusion drawn from the table does not depend on it: check-shaped wording is
correct in **both** columns and in a user project whose `.gitignore` this
repository does not author, so REQ-AEC-004 survives whatever the sibling turns
out to do. Only the row's specific design claim is unverified.

Wording that says "the destination is gitignored" is false in the right-hand
column. Wording that says "the destination is tracked" is false in the left.
**Wording that says "run the check" is true in both**, and is additionally true
in a user project whose `.gitignore` this repository does not author. This is the
whole reason the replacement is check-shaped rather than fact-shaped, and it is
also why REQ-AEC-004 is stated as a prohibition rather than as a style note.

---

## §C Scope surfaces — six files, four of them mirrors

| # | File | Change | Copy class |
|---|---|---|---|
| 1 | `.claude/agents/moai/plan-auditor.md` | §C.1 tail, plan-auditor FORBIDDEN variant | C1 |
| 2 | `.claude/agents/moai/sync-auditor.md` | §C.1 tail, sync-auditor FORBIDDEN variant | C1 |
| 3 | `internal/template/templates/.claude/agents/moai/plan-auditor.md` | as #1 | C2 |
| 4 | `internal/template/templates/.claude/agents/moai/sync-auditor.md` | as #2 | C2 |
| 5 | `.moai/docs/audit-artifact-convention.md` | §C.2, §C.3, §C.4 | local |
| 6 | `internal/template/templates/.moai/docs/audit-artifact-convention.md` | byte-identical to #5 | template mirror |

Two copy relations behave differently and MUST NOT be conflated:

- **C1 ↔ C2 is hand-maintained and intentionally divergent.** The sync-auditor
  clause sits at line 108 in C1 and line 92 in C2. Each copy is opened and
  edited on its own; no byte-comparison is asserted between them, and none is
  verified.
- **#5 ↔ #6 is byte-identical today and must remain so.** `cmp` is the check.

**C3 is never hand-edited.** `internal/template/templates/.codex/agents/moai/*.toml`
is machine-emitted from C2. Touching it by hand is silently overwritten at the
next emission.

**Six further files were measured and excluded.** A markup-insensitive sweep for
the paraphrase form (`tracked** path`) returns 8 lines across 6 files that assert
the same falsehood in different words. They are enumerated, measured, and
excluded with a named closure in SPEC § Out of Scope — the paraphrase-class
"tracked path" surfaces. The scope table above is six files **by decision**, not
by oversight; the sweep is recorded there so the exclusion is auditable and the
successor card derivable.

---

## §D Milestones

### M1 — Repair the four agent copies (Priority: High)

Apply the §C.1 replacement tail to files #1-#4, one file at a time, preserving
each copy's existing FORBIDDEN clause verbatim. The tail is identical across all
four; only the FORBIDDEN variant differs, and it differs by auditor role, not by
copy class.

Exit condition: all four copies carry the new tail; no copy carries the string
`Never write the verdict to a gitignored location`.

### M2 — Repair the convention document and its mirror (Priority: High)

Apply §C.2 (§ Committing), §C.3 (§ Where FORBIDDEN justification), and §C.4
(§ What makes the convention stick) to file #5, then apply the identical change
to file #6.

Exit condition: `cmp` reports the two files byte-identical; the § Committing
sentence "They reach the integration branch with the card's evidence commit" is
preceded by the forced-stage instruction rather than standing alone as the whole
procedure.

**Applied to the mirror as a re-application, not a copy.** Writing #6 by copying
#5 wholesale is acceptable only because they are byte-identical *today*
(verified in SPEC §A.3). If that verification fails at implementation time, the
mirror is edited section-by-section instead, and the divergence is reported as a
blocker rather than flattened.

### M3 — Regenerate the emitted codex layer (Priority: High)

Run `make agents-emit`. This is mandatory because M1 touched C2; skipping it
leaves the emitted TOMLs carrying the false sentence, and a binary built from
that tree embeds the stale definitions.

Exit condition: `make agents-emit-check` exits 0.

### M4 — Verification sweep (Priority: Medium)

Execute the acceptance criteria as a single-turn parallel read-only batch, with
the positive control preceding every absence claim. Export the verdict per the
repaired § Committing procedure — this SPEC's own evidence commit is the first
exercise of the wording it specifies, and a failure to force-stage it is a
self-refuting close.

---

## §E Verification approach

### The grep hazard, stated before the commands

The clause being searched for carries `**bold**` runs
(`[HARD] **Export mandate — an audit is complete only when its verdict is
exported.**`). A literal grep spanning that markup is markup-sensitive and can
return a clean zero for a file that carries the clause. Two protections:

1. **Search only patterns that cross no emphasis marker.** `Export mandate`
   (inside the bold span, but containing no `*`) and `Never write the verdict to
   a gitignored location` (outside it) are both safe anchors. What is unsafe is
   a pattern spanning a `**` boundary — position relative to the bold run is not
   the criterion; crossing its delimiter is.
2. **Never read a zero without a positive control.** Every absence claim in the
   verification is paired with a grep for a string known to be present in the
   same files. A control that returns zero means the instrument is broken, not
   that the target is absent.

### Verification surfaces

| Surface | Instrument |
|---|---|
| Four agent copies repaired | anchored grep for the old sentence (expect 0 across the `.md` layer) + positive control |
| Emitted codex layer regenerated | `make agents-emit-check` exit status |
| Convention mirror parity | `cmp` |
| Template neutrality | grep sweep for the forbidden content classes over the C2 and mirror diffs |
| Wording truth | the replacement's own check run against a throwaway path, **followed by the plain `git add` whose behaviour the clause predicts** — the check's exit code alone cannot falsify a wrong consequence clause (AC-AEC-013) |
| No specified sentence asserts tree state | declarative-form sweep over the added lines of the diff across all six scope files, with a hedged-form control (AC-AEC-016) |

---

## §F Anti-patterns for this card

- **Editing `.gitignore`.** Out of scope, and the direction the operator
  rejected. If the repaired wording seems to require an ignore change, the
  wording is wrong, not the policy.
- **Hand-editing a `.codex/*.toml`.** Overwritten at the next emission, and the
  edit's existence is erased with it.
- **Asserting byte-parity between C1 and C2.** The line-number divergence
  measured in SPEC §A.1 is the standing counter-example.
- **Reading a grep zero as absence.** The clause's markup makes this the likely
  failure mode; the positive control is not optional.
- **Committing this SPEC's own evidence with a plain `git add`.** The card's
  verdict lands under an ignore-matched path; the plain stage **refuses** (exit 1,
  nothing staged — SPEC §A.4b), so the commit that follows carries no verdict and
  the card closes with no recoverable basis. Read the exit code: the refusal is
  loud but easy to walk past in a batch.
- **Reaching for `git add -A` when the plain stage is refused.** The sweep would
  "work" and is prohibited by `AGENTS.md` § Git, branches, and the shared
  checkout. The sanctioned remedy is `git add -f <the explicit path>`.

---

## §G Cross-references

- `.moai/docs/audit-artifact-convention.md` — the document repaired by M2
- `.claude/agents/moai/plan-auditor.md`, `.claude/agents/moai/sync-auditor.md` —
  the definitions repaired by M1
- `.claude/rules/moai/core/verification-claim-integrity.md` — the
  unattributed-claim invariant the export mandate operationalizes, and the
  reason a verdict that never reaches the branch is a defect rather than an
  untidiness
- `.claude/rules/moai/workflow/kanban-dispatch.md` § Completion is read, never
  trusted — the lead-side reader whose read this repair protects
