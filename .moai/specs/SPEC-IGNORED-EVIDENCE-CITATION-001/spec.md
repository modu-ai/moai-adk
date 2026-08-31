---
id: SPEC-IGNORED-EVIDENCE-CITATION-001
title: "Repair tracked citations that name gitignored evidence paths"
version: "0.5.0"
status: in-progress
created: 2026-08-31
updated: 2026-08-31
author: manager-spec
priority: P2
phase: "v3.1.4 target"
module: "internal/cli, internal/hook, .moai/reports"
lifecycle: spec-anchored
tags: "evidence, citation, gitignore, verification-claim-integrity, corpus-cleanup"
related_specs: [SPEC-EVIDENCE-CITATION-CANON-001, SPEC-HIERARCHICAL-TEAM-001]
tier: M
---

# SPEC-IGNORED-EVIDENCE-CITATION-001 — Repair tracked citations that name gitignored evidence paths

## HISTORY

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-08-31 | 0.1.0 | Initial draft. Scope fixed to the 5 class-C3 lines of `.moai/reports/t381/census.md` (measured tree `3f03d9c36`). t375 canon folded in as measured, not assumed. | manager-spec |
| 2026-08-31 | 0.2.0 | plan-audit iter1 (FAIL 0.74) revision. Added two requirements — stale coordinates + behavior preservation, numbered REQ-IEC-010 and REQ-IEC-011 at the time, renumbered to REQ-IEC-009 and REQ-IEC-010 in iter4; corrected the phantom `.codex/…` do-not-touch path; corrected t375 `status:` to `in-progress` with a read date; narrowed the in-scope pattern to `.moai/state/verify` after measuring a second, out-of-scope `.moai/state/` occurrence in the html; showed the `492 − 25 = 467` derivation; relabelled the lane-8 convergence claim as relayed; re-argued the brace-glob exemption against the REQ-ECC-004+005 pair. | manager-spec |
| 2026-08-31 | 0.3.0 | plan-audit iter2 (FAIL 0.78) delta revision, operator-approved iter3. Demoted the probe-boundary requirement (then numbered REQ-IEC-008; that id was reassigned by the iter4 renumber) to a §D constraint (its only coverage had been demoted to a non-gating check, orphaning it). Updated every t375 citation from unmerged-draft to landed (`status: completed`, `origin/develop` `9328a5242`), including the rule-body sentence — now verified at `manager-lead.md:150` rather than anticipated — and upgraded the lane-8 convergence claim from relayed to verified against `manager-lead.md:152`. `acceptance.md` carries the N1/N2 criterion fixes. | manager-spec |
| 2026-08-31 | 0.4.0 | plan-audit iter3 (FAIL on MP-1 firewall; score 0.83 cleared) delta revision, operator-approved iter4, scoped to P1-P5. **P1**: renumbered REQ-IEC-009/010/011 → 008/009/010, closing the numbering gap the iter3 demotion opened — measured 18 occurrence-lines across the 4 SPEC artifacts (the 3 historical audit reports were deliberately NOT rewritten). **P2**: retargeted the two dangling probe-boundary references to the §D constraint by name; the renumber had turned them into a worse defect than dangling, since `REQ-IEC-008` now names Collision avoidance. **P3**: replaced the frozen diff baseline with the three-dot `origin/develop...HEAD` form in AC-IEC-006/007 — the frozen SHA would have failed at integration for t375's landed edits. **P4**: retired the last two "will add" sentences. **P5**: repaired the corrupted plan.md §H. | manager-spec |
| 2026-08-31 | 0.5.0 | Post-run correction (iter5), scoped to two statements the implementation falsified. **(1)** `plan.md` M3 branched on whether the `eb01063e` scratch tree exists; §C.3's test is **identifiability**. Measured: the tree exists in the primary checkout (8 per-category audit files, 244K) but NOT in this worktree, so a tree-existence test both picks the wrong branch AND answers differently per tree — a milestone gated on a gitignored path, this card's own defect class. `plan.md` §B item 3 carried the same false fact and was corrected with it. **(2)** REQ-IEC-005 named one cause for un-identifiability ("the scratch tree no longer exists"); here the tree exists and the cause is that the citation is report-wide over eight per-category audits. The cause is de-specified and the obligation slightly strengthened — the reason must now be recorded, not assumed. Also **P6**: §C.5 re-framed from lane-race avoidance to scope discipline (t375 landed), and an absorb hazard on AC-IEC-007's `ls` half raised but deliberately not repaired. | manager-spec |

---

## §A. Context

### A.1 The defect

Tracked files in this repository cite paths under `.moai/state/` as the provenance for factual
claims. `.moai/state/` is gitignored — the rule is at `.gitignore:298` **in this tree** (measured; a
line number is a moving coordinate and other trees report a different one). Outside the authoring
machine — a fresh clone, CI, another worktree — the cited path does not resolve, so the citation
names evidence that is not there.

Demonstrated by observation, not argument. In this freshly-created worktree:

```
$ git check-ignore -v .moai/state/verify/t225
.gitignore:298:.moai/state/	.moai/state/verify/t225

$ ls .moai/state/verify/
ls: .moai/state/verify/: No such file or directory
```

Five tracked files cite paths beneath a directory that does not exist here.

### A.2 Doctrinal anchor

`.claude/rules/moai/core/verification-claim-integrity.md` §2 (Baseline-Integrity Attribution): a
claim whose cited evidence path no longer resolves is an **unattributed claim**, and must be reported
as a Gap rather than as a Claim. That file's §1.1 surface 1 makes an unobserved completion claim a
defect in its own right.

### A.3 Provenance of scope

Scope is taken from `.moai/reports/t381/census.md` (lead-accepted, measured tree `3f03d9c36`), class
**C3 — 인용 (5)**. This SPEC does not re-derive that census; it cites it.

`.moai/reports/t381/discovery.md` is background only. Its §2-§3 counts are **superseded** by
census.md — they came from an excluding probe.

### A.4 Relation to SPEC-EVIDENCE-CITATION-CANON-001 (card t375) — the reason this card exists

**Citation — landed.** `SPEC-EVIDENCE-CITATION-CANON-001` v0.3.0, `tier: M`, `status: completed`,
requirement block at `spec.md:164-175`. **This is no longer a draft citation**: the SPEC has merged
into `origin/develop` and is read from tracked history rather than from a peer worktree.

Read on **2026-08-31** by:

```bash
git show origin/develop:.moai/specs/SPEC-EVIDENCE-CITATION-CANON-001/spec.md
```

`origin/develop` was `9328a5242` at that read. The command is the citation — the file is tracked in
`develop`, so it is not copied into this card's evidence directory (copying a tracked file into a
report would be the redundant-export shape REQ-ECC-005 warns against).

> **Why this citation pins a read date.** `status:` has held **three** different values during this
> card's lifetime — `draft` (iter1, already stale when written), `in-progress` (iter2, verified), and
> `completed` (iter3, verified above). No reading of it is durable, which is why the value is stated
> with the date and command that produced it and carries no weight in anything this card decides. The
> requirement **line numbers were re-verified unchanged in the landed version** (REQ-ECC-001…007 at
> `:164, :165, :166, :167, :168, :169, :175`; the 004+005 pairing note at `:171`), so every
> consequence drawn below still holds against the merged text.

The four requirements that bear on this card:

| Req | Line | Substance |
|---|---|---|
| REQ-ECC-001 | :164 | A document citing an evidence path as the basis of a verification or verdict shall use a version-tracked path; the canonical location is `.moai/reports/<card-id>/`. |
| REQ-ECC-002 | :165 | **Doctrine-surface** documents (`.claude/rules/`, `.claude/agents/`, `.claude/output-styles/`, `.claude/skills/`) shall name `.moai/state/verify/<session>/` as machine-local scratch and shall not name it as an audit-time citation target. |
| REQ-ECC-003 | :166 | When an artifact is cited as verdict evidence, the actor shall export it to a tracked path **before** citing. Citing an unexported path is an unattributed claim. |
| REQ-ECC-004 | :167 | A citation shall name **a single file** and shall not name a directory. Whole-scratch-directory export is prohibited. |

**Outside the guard, inside the canon.** This is the card's reason for being, and it is a
two-axis statement, not the single-axis "guard is `.md`-only" the earlier framing used:

- **Outside by extension.** REQ-ECC-007 (:175) scopes the guard to `.md` files. Every in-scope
  citation here is `.go`, `.txt`, or `.html`.
- **Outside by directory.** REQ-ECC-007 further scopes the guard to the **doctrine surface**. Every
  in-scope citation here lives in `internal/cli/`, `internal/hook/`, or `.moai/reports/`.
- **Inside the canon.** REQ-ECC-001 and REQ-ECC-003 carry **no doctrine-surface qualifier** — they
  bind any document citing evidence for a verdict. So these five are governed by t375's canon while
  being unreachable by t375's guard. That gap is exactly what this card cleans.

**REQ-ECC-004 reaches further than the coordination summary stated.** The summary named the glob at
`.moai/reports/template-skill-improvement-plan-20260710.html` as the one entry REQ-ECC-004 catches.
Reading the requirement verbatim, it prohibits naming **a directory** as well as requiring a single
file, so it bears on **four** of the five:

| File | REQ-ECC-004 exposure |
|---|---|
| `.moai/reports/template-skill-improvement-plan-20260710.html` | glob (`skill-audit/*.json`) — multi-file |
| `.moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt` | brace-glob (`0{2,3,4,5}-*.txt`) — multi-file |
| `internal/hook/evidence_writer_zeroexec_test.go` | bare directory (`.moai/state/verify/t341/`) |
| `internal/cli/audit_pin_live_test.go` | bare directory (`.moai/state/verify/t225/`) — but see the boundary note below |

The `audit_pin_live_test.go` case is a genuine boundary: REQ-ECC-004 governs a *citation*, and that
line is an **output-location statement** rather than a citation of evidence. Whether 004 reaches it
depends on a reading this card cannot settle unilaterally. §C.4 records it as an open question
routed to the lead rather than decided here.

**The rule-body sentence has LANDED — verified, no longer anticipated.** iter1/iter2 described this
as a sentence t375 "will add". It is now in `origin/develop`, at
`.claude/agents/moai/manager-lead.md:150` (read 2026-08-31 via
`git show origin/develop:.claude/agents/moai/manager-lead.md`):

> Export the named file only, never the scratch directory wholesale; what stays behind, and its loss
> risk, is recorded under Residual-risk.

It is a **rule-body** sentence, not a numbered requirement, so this SPEC still assigns it no
requirement number. It bears directly on `internal/hook/evidence_writer_zeroexec_test.go`: card t341
decided not to export those raw captures, making that entry a **confirmed-unexportable** target rather
than a merely-unexported one. The §C.3 treatment (b) chosen for it — an explicitly non-resolving origin
note that is not offered as the evidence — is what the landed sentence prescribes: the remainder stays
behind, and its loss is recorded rather than disguised.

**Independent convergence — now VERIFIED from the landed text.** iter2 recorded this as relayed from
the lead's coordination message, because t375 was then unmerged and only its author knew what it had
decided. With t375 landed, the decision is readable in `origin/develop` and no longer rests on the
relay. `.claude/agents/moai/manager-lead.md:152` (read 2026-08-31 via `git show origin/develop:…`):

> The cited path MUST resolve at audit time (per `verification-claim-integrity.md` §2 — a cited path
> that no longer resolves is an unattributed claim), and **only the tracked path does**.

That is the same split this card selected as treatment (d) for `internal/cli/audit_pin_live_test.go`:
the "must resolve at audit time" instruction is **kept**, and what is corrected is the *target* it
points at — from the gitignored scratch directory to the tracked path. Lane-8 did not delete the
clause; it re-pointed it, exactly as §C.3 does here.

Two cards reaching the same discriminator independently is evidence the discriminator is real rather
than one author's framing. The argument does not rest on it either way: the discriminator is argued
from the two clauses' differing truth values in §C.3.

---

## §B. Requirements (GEARS)

### REQ-IEC-001 — Non-resolving citation repair

The repository **shall not** carry, in a tracked file, a citation naming a path under `.moai/state/`
as the provenance of a factual claim without an adjacent statement that the path does not resolve
outside the authoring tree.

The in-scope lines are exactly the five files enumerated in §C.1.

### REQ-IEC-002 — False resolvability assertion removal

**When** a tracked comment asserts that a cited path "still resolves at audit time" while that path
is matched by a `.gitignore` rule, the assertion **shall** be removed or corrected, because the
ignore rule contradicts it.

This binds `internal/cli/audit_pin_live_test.go`.

### REQ-IEC-003 — Claim survivability

**Where** a citation's supporting figures already appear in the body of the same comment or document,
the repair **shall** preserve those figures verbatim, so that removing the path costs the reader
nothing.

### REQ-IEC-004 — Origin preservation for path-dependent citations

**Where** a citation is the only pointer to the origin of values that do not appear in the body, the
repair **shall not** delete the pointer outright. It **shall** either re-point to a tracked location
under `.moai/reports/<card-id>/` (the REQ-ECC-001 canonical location), or demote the path to an
explicitly-non-resolving origin note that is **not** presented as the evidence for the claim.

### REQ-IEC-005 — Single-file naming, or a recorded inability

**Where** a repaired citation names a directory or a glob rather than one file, the repair **shall**
name the single file that decided the claim, per REQ-ECC-004. **When** that file cannot be identified
— for any reason — the repair **shall** record the inability, **and its reason**, explicitly at the
citation site, and **shall not** silently retain the glob or directory as though it were a valid
citation.

> **iter5 (correction 2): the cause was over-specified, not the obligation.** The clause previously
> named one cause — "because the scratch tree no longer exists on this machine". The run phase found
> a different one: for
> `.moai/reports/template-skill-improvement-plan-20260710.html` the scratch tree **does** exist, and
> the file is unidentifiable because the citation is report-wide while the eight JSONs are
> per-category batch audits. Writing the requirement's stated cause into that file would have made
> the repaired file assert something false — the defect this card exists to repair — so the
> implementer correctly declined to. The obligation is unchanged and is in fact slightly stronger:
> the reason must now be recorded, rather than assumed to be the one the requirement named.

### REQ-IEC-006 — Carve-out immutability

The repair **shall not** modify any of the 12 lines classified C1 (machine consumer), C2 (fixture),
or C5 (mechanism prose) in census.md §2. A guard widened without classification would false-positive
these: the paths are *defined by* the code, are created at runtime, and absent-before-creation is
correct behavior — telling code that defines a path that the path does not resolve is meaningless.
REQ-ECC-006 (:169) names two of these same consumers (`internal/verify/store.go`,
`internal/web/events.go`) as explicit carve-outs, so the two cards agree on this boundary.

### REQ-IEC-007 — Self-consistency of this card's own evidence

**When** this SPEC's acceptance criteria record a command's verbatim output, the evidence **shall**
be written under `.moai/reports/t381/` (tracked) and **shall not** be written under `.moai/state/`.

A card fixing non-resolving evidence citations while committing the same defect in its own evidence
would be self-refuting.

> **The probe-boundary requirement — `REQ-IEC-008` as numbered before the iter4 renumber — was
> demoted to a §D constraint in iter3 (N3).** It stated a documentation property of
> this SPEC, not a property of the repair, and its only coverage — the former AC-IEC-008 — was demoted
> to a non-gating structural check in iter2 because no milestone can falsify it. That left a
> requirement with no MUST criterion. The demotion resolves the orphan by moving the requirement to
> the layer its subject already belonged to; re-promoting the check instead would have restored
> exactly the vacuity iter2 removed. The obligation is unchanged and still verified — see §D and
> `acceptance.md` §D S-1.
>
> **iter4 (P1): the numbering was closed behind the demotion.** iter3 left a gap at 008 on the stated
> ground that keeping the later ids preserved existing references. That ground did not survive
> measurement — `grep -rln 'REQ-IEC-'` over the whole tree finds the token in exactly four SPEC files
> and three historical audit reports, and `git grep 'REQ-IEC-' origin/develop -- . ':!.moai/specs'`
> returns nothing, so no external consumer existed to protect. The gap protected nothing and cost the
> MP-1 firewall, which admits no exception. Former 009/010/011 are now **008/009/010**.

### REQ-IEC-008 — Collision avoidance

The repair **shall not** modify any file owned by the parallel card t375 (§C.5). Two lanes writing
the same file is a write race, not a merge.

### REQ-IEC-009 — Stale evidence coordinates

A tracked citation **shall not** name a line number without naming the tree that line number was
resolved in.

A line number is a moving coordinate: `.moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt`
cites `.gitignore:284` while this tree measures `:298`, and the census dispatch made the same slip in
the other direction. This is a distinct defect from REQ-IEC-001 — a citation can carry a correct
non-resolution marker and still name a line that has moved. Where the coordinate is not load-bearing,
the repair **shall** prefer removing it over updating it, since an updated coordinate goes stale the
same way.

> **Why this requirement was added in iter2 (D6)**: `extract.txt` was the one in-scope file gated by
> no requirement — REQ-IEC-001 does not reach it, because its header already declares its own
> non-resolution and it therefore satisfies AC-IEC-001 today, repaired or not (measured:
> `grep -ci 'gitignore'` = 1). Milestone M4 was mandated by §C.3 prose alone. This requirement is what
> now binds it.

### REQ-IEC-010 — Behavior preservation

The repair **shall not** alter runtime behavior. Every in-scope edit is confined to a comment, a
header, or a footer; no Go statement, no test assertion, and no runtime path is modified.

---

## §C. Scope

### C.1 In scope — exactly 5 lines (census.md class C3)

Referenced by file path throughout; index numbering is deliberately avoided because it has already
drifted once in coordination traffic.

| File | Line | Cited artifact | Body self-sufficient? |
|---|---|---|---|
| `internal/cli/mcp_glm.go` | 110 | `.moai/state/verify/t225/ac-amp-006-glm-differential-attempt1.md` | **YES** — measured figures (output tokens 3667 vs 3480 under budgets 3072 vs 1024, ratio 1.02) are in the comment body |
| `internal/cli/audit_pin_live_test.go` | 32 | `<repo>/.moai/state/verify/t225/` (directory) | N/A — an output-location statement, plus a **false** resolvability assertion |
| `internal/hook/evidence_writer_zeroexec_test.go` | 10 | `.moai/state/verify/t341/` (directory) | **PARTIAL** — see §C.4 correction 2 |
| `.moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt` | 3 | `.moai/state/verify/t299-sync/0{2,3,4,5}-*.txt` (glob) | **YES (already exported)** — see §C.4 correction 1 |
| `.moai/reports/template-skill-improvement-plan-20260710.html` | 684 | `.moai/state/verify/eb01063e/skill-audit/*.json` (glob) | **NO** — "원시 데이터(JSON)" only |

Three file types: `.go` ×3, `.txt` ×1, `.html` ×1.

**[HARD] Pattern-scope constraint — match `.moai/state/verify`, never the broader `.moai/state/`.**
Measured in this tree (iter2): `.moai/reports/template-skill-improvement-plan-20260710.html` contains
**two** `.moai/state/` occurrences —

```
$ grep -n '\.moai/state/' .moai/reports/template-skill-improvement-plan-20260710.html
529:  … .moai/state/loop-verdict-&lt;id&gt;.json persistence are a rigorous, auditable loop-exit design …
684:  원시 데이터(JSON): <code>.moai/state/verify/eb01063e/skill-audit/*.json</code> …
```

Line 684 is the in-scope citation. **Line 529 is out of scope**: it is a `.moai/state/loop-verdict-`
mechanism/schema reference (census class B2 — `.moai/state/` outside `/verify`, the 467-line
population §C.7 explicitly declines to widen into), and it is mechanism prose in the C5 sense, not a
citation of evidence. Narrowing to `.moai/state/verify` yields exactly `1` per in-scope file:

```
$ git grep -c '\.moai/state/verify' -- <the five files>
… : 1   (each of the five)
```

This was not caught in iter1, whose criteria matched `.moai/state/` and would have swept line 529
into the repair — producing either a false RED or an edit this SPEC forbids. `acceptance.md` §A
carries the same constraint.

### C.2 Repair-direction decision — the citations do not take a uniform treatment

Candidate treatments, evaluated per citation rather than swept:

- **(a) delete** the path reference outright;
- **(b) reword** so the claim stands on what is in the body, with the path demoted to an explicitly
  non-resolving, local-only origin note;
- **(c) re-point** to a tracked location under `.moai/reports/<card-id>/` where the evidence was, or
  could be, exported;
- **(d) correct** a false resolvability assertion while keeping a legitimate output-location
  statement.

The selection, with justification, is §C.3. Every selected treatment is constrained by the landed
rule-body sentence at `manager-lead.md:150` (quoted in §A.4): none of them may leave a
deliberately-unexported artifact standing as verdict evidence.

### C.3 Selected treatment, per file

**`internal/cli/mcp_glm.go` → (a) delete the path, keep the card token.**
The claim's figures are in the comment body (verified by reading the file), so the path carries no
information the reader loses. Deleting it satisfies REQ-ECC-003 trivially — nothing is cited that was
not exported, because nothing needs citing. The card token `t225` is retained as a recoverable
provenance word, which is not a path and makes no resolvability claim.

**`internal/cli/audit_pin_live_test.go` → (d) correct the assertion, keep the output-location
statement.** The two halves of that sentence have different truth values. "Evidence lands in
`<repo>/.moai/state/verify/t225/`" is a true statement about where this test writes. "…so the cited
paths still resolve at audit time" is false, and the ignore rule is what falsifies it. The repair
removes the second and replaces it with the truthful consequence: the directory is gitignored, does
not survive the worktree, and decision-bearing content must therefore be extracted to
`.moai/reports/<card-id>/` before being cited. Lane-8 reached this same split independently (§A.4).

**`internal/hook/evidence_writer_zeroexec_test.go` → (b) demote to an explicitly-non-resolving origin
note.** The runner versions and the runner strings are already in the file (§C.4 correction 2), so
the claim stands without the path; the raw captures behind them are **confirmed unexportable** — t341
decided not to export them and the scratch tree is gone. (c) is therefore unavailable, and (a) would
destroy the only record of where the strings came from. The note must state plainly that the captures
were not exported and do not resolve, which is what keeps it compatible with the landed rule-body
sentence at `manager-lead.md:150`: it is an origin annotation, explicitly not offered as the evidence
for the claim. Under
REQ-IEC-005 the inability to name a single file is recorded at the site rather than papered over.

**`.moai/reports/t299/verify-sync/e2e-lint-4paths.extract.txt` → correct the stale line number only.**
This file is **already** treatment (c) correctly applied (§C.4 correction 1). Its only live defect is
that it cites `.gitignore:284` while this tree measures `:298`. The repair corrects that, or restates
the reference without a line number so it cannot go stale again — the second being preferable, since
a line number is a moving coordinate (REQ-IEC-009).

Its brace-glob is left as-is, and iter2 argues this against the **004+005 pair** rather than against
REQ-ECC-004 alone (D9). The note following t375 `spec.md:171` states that 004 and 005 "form one
ceiling" — 004 constrains the *breadth* of a citation, 005 decides *what to export* within that
breadth — precisely because a broadly-written citation could otherwise satisfy both while exporting
everything. Read against both: REQ-ECC-005 prescribes exporting the command's decision-bearing lines
and leaving the rest in scratch with the loss risk recorded, and that is exactly what this extract
does — the probe's own lines and the totals line are inline, the ~134KB of whole-corpus output stays
unexported, and per-source sha256 plus byte counts are recorded so the unexported remainder is
attributable. The brace-glob names the *sources of that export*, not the evidence for a verdict, so
it is not the broad-citation shape the pairing was written to close. If t375's wording shifts before
this card lands, this exemption is re-confirmed at M1 rather than assumed.

**`.moai/reports/template-skill-improvement-plan-20260710.html` → (c) if the file can be identified,
otherwise (b) with a recorded inability.** This is the only entry where REQ-ECC-004's single-file
rule bites cleanly and cannot be satisfied by inspection alone: `skill-audit/*.json` is a glob, and
the `eb01063e` scratch tree does not exist in this worktree. The run-phase decision rule is stated
rather than guessed: **if** the scratch tree is found on the authoring machine, export the one JSON
that decided the claim to `.moai/reports/` and re-point to it; **if not**, apply (b) and record at the
citation site that the raw data was never exported and cannot now be named per-file. Silently keeping
the glob is prohibited by REQ-IEC-005.

### C.4 Framing corrections found while measuring

Two entries were framed in the dispatch as uniformly path-dependent. Direct reading contradicts that,
and the SPEC records the correction rather than silently adopting the framing.

**Correction 1 — `e2e-lint-4paths.extract.txt` is not the same defect shape; it is the correct
pattern already applied.** It is a **tracked export**. Its header already states that the source path
"is gitignored … and vanishes with the worktree, so this extract is the citable carrier", carries
per-source sha256 and byte counts, and carries the decision-bearing lines inline (75 lines total). Its
`.moai/state/` reference is an origin note that already declares itself non-resolving — treatment (c)
plus (b), before either was named. Only the stale `.gitignore:284` is live.

**Correction 2 — `evidence_writer_zeroexec_test.go` is more self-sufficient than framed.** The
dispatch says the body has "no values". It does: the comment carries the runner versions explicitly
(`go1.25`, `pytest 8.4.2`, `cargo 1.94.1`, `jest 30.4.2`, `vitest 3.2.7`, `node v22.14.0`), and the
runner strings are the test file's own literals. The `.moai/state/verify/t341/` pointer is provenance
for the **raw** captures, not the sole carrier of the claim. It stays path-dependent for those raw
captures, so REQ-IEC-004 still applies — but "no values in the body" is false.

**Correction 3 — REQ-ECC-004's reach.** Recorded at §A.4: it bears on four files, not one, and its
application to `audit_pin_live_test.go` is an open question (citation vs output-location statement)
routed to the lead rather than decided here.

### C.5 Do-not-touch — files delivered by SPEC-EVIDENCE-CITATION-CANON-001 (t375)

t375 has **landed** (`status: completed` in `origin/develop`), so this is no longer race avoidance
between concurrent lanes — no concurrent lane exists. It is **scope discipline**: these files are
another card's deliverables, and this card's 5-line scope does not reach them. The exclusion is
unchanged; only its reason is (iter5, P6).

- `internal/template/evidence_citation_guard_test.go` — t375 **created** it; its filename is asserted
  literally by t375's own acceptance criterion. Absent in this worktree, present in `origin/develop`
  (measured 2026-08-31: `ls` → `No such file or directory`;
  `git show origin/develop:internal/template/evidence_citation_guard_test.go` → exit 0). See the
  absorb note below.
- `.claude/rules/moai/core/agent-common-protocol.md` **and** its detail companion
  `.claude/rules/moai/core/agent-common-protocol-reference.md`, plus both template originals under
  `internal/template/templates/` — t375 **edited** the summary and detail companions as a pair.
- `.claude/agents/moai/manager-lead.md`, its template original, and the machine-emitted
  `internal/template/templates/.codex/agents/moai/manager-lead.toml`.

> **iter2 correction (D5)**: iter1 listed a top-level `.codex/agents/moai/manager-lead.toml`. That
> path **does not exist in this tree** — `ls` reports `No such file or directory`, and
> `git ls-files '*manager-lead.toml'` returns exactly one path, the
> `internal/template/templates/` one above. A do-not-touch entry naming a non-existent path asserts
> nothing, and the AC diffing it exited 0 with empty output for the wrong reason. Corrected here and
> in `acceptance.md` AC-IEC-007. Every other entry in this list was re-measured and does exist.

> **Absorb hazard — RESOLVED (iter5, operator judgment (a)).** AC-IEC-007's second half used to assert
> that `internal/template/evidence_citation_guard_test.go` is **absent from the tree**, as evidence
> that this card did not create t375's guard file. t375 landed that file into `origin/develop`, so the
> assertion answered "absent" in this worktree and would have answered "present" after absorption —
> failing the criterion for a file this card never touched.
>
> It was first raised here as an open hazard, on the ground that AC-IEC-007 was passing and passing
> criteria should not be rewritten. The operator overturned that premise, and the reasoning is worth
> keeping: **the stop rule protects a criterion passing on its judgment, and this half was not passing
> on judgment — it passed because this worktree happens not to contain the file while
> `origin/develop` does. A criterion whose answer flips when you change trees is not passing; it is
> unjudged.** Correcting it is therefore not a rewrite of a passing criterion but a repair of an
> unjudged one.
>
> The fix folds the guard file into AC-IEC-007's existing three-dot diff as its eighth path, changing
> the question from *is it absent from the tree* to *is it absent from this card's own diff* — a
> property this card controls. Both states were verified without absorbing, using
> `git merge-tree --write-tree` to compute the merged tree; see `acceptance.md` AC-IEC-007 for the
> four measurements. This closes the half of the P3 hazard that P3's three-dot form could not reach,
> because file existence is not a diff.

### C.6 Carve-outs — 12 lines that MUST NOT be modified (REQ-IEC-006)

C1 machine consumer (6): `internal/verify/store.go:15`, `internal/verify/schema.go:44`,
`internal/verify/store_test.go:26,44,45`, `internal/web/events.go:29`.

C2 fixture (4): `internal/goal/evaluate_snapshot_test.go:48,71,166`,
`internal/session/ignored_content_test.go:26`.

C5 mechanism prose (2): `.moai/reports/moai-autonomy-workflow-redesign-20260803.html:388`,
`.moai/reports/model-tier-redesign-20260712.html:57`.

### C.7 The census probe and its blind spots (§D probe-boundary constraint)

Probe, measured tree `3f03d9c36`, reproduced in this tree (25 lines):

```bash
git grep -n '\.moai/state/verify' -- . ':!*.md'
```

What this probe structurally cannot see (census.md §1.1):

| # | Blind spot | Scale |
|---|---|---|
| B1 | `.md` files — deliberately excluded (t375's scope) | 531 − 25 = 506 lines |
| B2 | Paths under `.moai/state/` that are not `/verify` (`handoff/`, `session-msg/`, `config-cache.json`, …) | `git grep -n '\.moai/state' -- . ':!*.md'` = 492; 492 − 25 = 467 lines |
| B3 | **Untracked files** — `git grep` sees tracked files only | unmeasured |
| B4 | Ignored paths **outside** `.moai/state/` — `bin/`, `.moai/logs/`, `.moai/cache/`, `.claude/settings.local.json` | unmeasured |
| B5 | **Runtime-assembled paths** — `filepath.Join(root, dir, name)` is invisible to any string probe | unmeasured |
| B6 | **Citations spanning a line boundary** — a line-oriented grep cannot see the relation | unmeasured |

B1's "unverified" caveat in census.md is now **partly resolved**: t375's guard scope is measured
(REQ-ECC-007, `.md` under the doctrine surface — §A.4), so B1's exclusion is confirmed as a real
boundary rather than an assumed one. B2-B6 remain unmeasured.

B2 alone is 18× the card's scope. Widening the probe would make this a different card; it is
disclosed, not widened.

Census.md §1.2 records that this card's own probes excluded part of the target **three times** (a
`| head -8` truncation sold as a census; a primary-checkout measurement standing in for the judgment
tree; a supplementary regex requiring `/t<digits>` that structurally missed `snapshots/abc.json`).
That record is why the §D probe-boundary constraint exists.

### C.8 Exclusions

### Out of Scope — the instruction that produces this pattern (census class C4)

- The 8 class-C4 lines that **instruct** agents to redirect verification output into
  `.moai/state/verify/$MOAI_SESSION_ID/` are NOT repaired here. The lead assigned that axis to t375,
  where REQ-ECC-002 governs it.
- Specifically excluded: `internal/template/templates/.codex/agents/moai/manager-lead.toml` lines
  `54,78,136,139,140,143,150`, and `.moai/config/sections/workflow.yaml:31`. Line `:143` — "The path
  MUST resolve at audit time", pointing at an ignored directory — is the original of the same
  contradiction found in `internal/cli/audit_pin_live_test.go`.
- That `.toml` is machine-emitted from `.claude/agents/moai/manager-lead.md` via `make agents-emit`;
  the `.md` is a t375 edit target, so touching C4 means touching a t375 file (§C.5).
- **Residual risk, stated rather than closed**: cleaning the corpus without changing the instruction
  means the same citations can be re-produced. This card does not close that axis.

### Out of Scope — guard construction

- This SPEC builds **no guard**. t375 owns it (REQ-ECC-007..010). This card cleans a corpus the guard
  cannot reach on either axis — extension and directory (§A.4). Any repair here that reads as guard
  construction is out of scope.

### Out of Scope — probe widening

- Blind spots B2-B6 (§C.7) are disclosed, not investigated. No work is planned against any of them.
- In particular the 467 lines of B2 (`.moai/state/` outside `/verify`) stay untouched.

### Out of Scope — amending a landed SPEC

- `SPEC-HIERARCHICAL-TEAM-001` carries the same instruction (`REQ-FOLD-002` / `AC-FOLD-002`;
  `spec.md:103,161,227` and `acceptance.md:9` — verified present in this tree). Amending a landed SPEC
  is out of scope for **both** t375 and t381; t375 records it as debt.

### Out of Scope — carve-outs and t375-owned files

- The 12 carve-out lines (§C.6) and the t375-owned files (§C.5) are excluded by REQ-IEC-006 and
  REQ-IEC-008 respectively, and are verified untouched by AC-IEC-006 / AC-IEC-007.

---

## §D. Constraints

- **No behavior change.** Restated normatively as REQ-IEC-010, which AC-IEC-011 gates. Every in-scope
  edit is to a comment, a header, or a footer.
- **Probe-boundary disclosure** (demoted in iter3 from the requirement then numbered `REQ-IEC-008`;
  that id was reassigned to Collision avoidance by the iter4 renumber, so this constraint is cited by
  name, never by a REQ id). This SPEC carries, in its own
  body, the census probe command and its stated blind spots (census.md §1.1 B1-B6) — §C.7 — because a
  card about citation completeness that hides its own probe boundary repeats the defect it fixes. It
  is a constraint rather than a requirement because it binds this SPEC's *documentation*, not the
  *repair*: no milestone can falsify it, so no MUST criterion can gate it honestly. Verified by
  `acceptance.md` §D S-1.
- **No time estimates.** Priority labels only.
- **Repair direction is a decision, not a sweep** (§C.2-§C.3).
- **t375's canon has landed, and its status has moved three times during this card.** Every REQ-ECC
  citation now names a merged document (§A.4). `status:` read `draft`, then `in-progress`, then
  `completed` across the three iterations of this card — which is exactly why the citation pins a read
  date rather than asserting a current value. If t375's wording changes again, §A.4 and the treatments
  in §C.3 are re-read, not assumed.

---

## §E. Residual risk and recorded judgment rules

### E.1 Adjacency is NOT enforced, even after this card lands (P7 — carried, not closed)

**The gap.** REQ-IEC-001 forbids a citation standing "without an **adjacent** statement that the path
does not resolve". AC-IEC-001 does not test adjacency. `grep -L` is **file-granular**: a marker
anywhere in the file satisfies it, so a marker in a header five hundred lines from the citation passes
the gate. The requirement is therefore **broader than its criterion** — under-covered, not circular.

**How the gap was created.** iter1's criterion used a `grep -n -C2` window, which *did* test adjacency
— but decided it by a human reading the window, which was defect D2. The D2 repair removed the prose
judgment **and the adjacency test with it**, and neither iter2 nor iter3 noticed the second loss. This
is a repair that closed one defect and took a neighbour with it; it survived three iterations
unnoticed because every iteration checked that the criterion was mechanical, and none checked that it
still tested what the requirement says.

**Why it is not closed here.** A per-file adjacency window cannot be expressed as one plain invocation
under this tree's execution guard (no `for`, no `$(…)`, no subshell — `acceptance.md` §A). Forcing it
would trade a mechanical criterion back for a prose one, re-opening D2. The two honest resolutions are
to weaken REQ-IEC-001's wording to match what is testable, or to build an adjacency test that survives
the guard; **both touch requirement substance**, which the iter4 delta grant excludes.

**Why the live risk is bounded.** All three §C.3 treatments place the marker at the citation site, so
today the gap is a latent looseness rather than a live error. It becomes live the moment someone
"satisfies" AC-IEC-001 by adding a marker far from the citation — which the criterion would accept.

**What a follow-up card needs** (recorded so the analysis is not re-derived): the requirement text at
REQ-IEC-001; the criterion at `acceptance.md` AC-IEC-001; the guard constraint at `acceptance.md` §A;
the D2 history in this section; and the fact that `grep -c` counts *lines*, so a same-line
citation-plus-marker test is expressible as a single pattern while a ±N-line window is not.

### E.2 Named shape — the half-corrected repair

**This card produced the same shape three times**: a fact is updated in one place and left standing in
the places that depend on it.

| Occurrence | Corrected in | Left standing in |
|---|---|---|
| iter2 → iter3 | §A.4 retired the "will add" framing when t375 landed | §C.2 and §C.3 kept "the sentence t375 M1 will add" (P4) |
| iter3 demotion | §B and §D recorded the demoted requirement | §C.7 and §C.8 kept pointing at its REQ id (P2) |
| iter4 renumber | §B ids renumbered 009/010/011 → 008/009/010 | the two P2 references would have silently re-bound to a *different* live requirement, since `REQ-IEC-008` now names Collision avoidance |

The third is the instructive one: the renumber turned a **dangling** reference into a **wrong** one,
which is strictly worse — a dangling reference announces itself, a re-bound reference does not. The
sweep rule this yields: **when a fact changes, grep for every occurrence of the old fact before
declaring the change done** — and when an identifier is reassigned rather than removed, the grep must
run *before* the reassignment, because afterwards the stale references are indistinguishable from
correct ones.

### E.3 Recorded judgment rules

Two rules the plan-auditor applied to this card's iter3 N4 deviation, preserved verbatim because they
generalize beyond it:

> The test is not "does the reason read well" but **"would deferring it create a fresh inaccuracy"**.

> **Metadata describing already-authorized edits may accompany them; it licenses no new substance.**

The second is stated narrowly on purpose. It admits a version bump and a HISTORY row alongside edits
that were already granted; it does **not** admit a new requirement, a new criterion, or a scope change
under cover of "recording what happened".

---

## §F. Cross-references

- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1, §2 — the doctrinal anchor.
- `.moai/reports/t381/census.md` — accepted scope source (measured tree `3f03d9c36`).
- `.moai/reports/t381/discovery.md` — background; §2-§3 counts SUPERSEDED by census.md.
- `SPEC-EVIDENCE-CITATION-CANON-001` v0.3.0, `status: completed` — **landed in `origin/develop`**
  (read 2026-08-31 via `git show origin/develop:.moai/specs/SPEC-EVIDENCE-CITATION-CANON-001/spec.md`
  at `origin/develop` `9328a5242`) — the canon this corpus falls under; its guard cannot reach this
  corpus (§A.4).
- `SPEC-HIERARCHICAL-TEAM-001` `REQ-FOLD-002` / `AC-FOLD-002` (`spec.md:103,161,227`,
  `acceptance.md:9`) — carries the same instruction; amending a landed SPEC is out of scope for both
  t375 and t381, and t375 records it as debt.
