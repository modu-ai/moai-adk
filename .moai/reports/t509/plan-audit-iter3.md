# SPEC Review Report: SPEC-WEB-CODEX-PANEL-001

Iteration: 3 (delta re-audit)
Verdict: **PASS**
Overall Score: **0.89** (Tier M PASS threshold 0.80)

Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t509`, branch `WT-codex-model-config`,
HEAD `ba7d93411`.

**Provenance correction — read this before the findings.** The SPEC directory is **not** clean at
HEAD. `git status --short` shows three files staged-but-uncommitted at the time of this audit:
`acceptance.md`, `progress.md`, and `.moai/reports/t509/verdict.md`. I did not create them; they
were already staged when I arrived, and I changed nothing but this report. **This audit therefore
reads the working tree, not HEAD `ba7d93411`.**

One of those staged changes is material to item 2. `git diff --cached acceptance.md` is a single
line — the MU-8 row — and the staged version is *stronger* than HEAD's: it adds "that was *run
rather than reasoned*", the explicit `base 0 / head 0 / diff exit 0` figures, and the instruction to
"apply it to each of the three targets, including the two whose current shape the anchor already
matches". My item-2 verdict below is against the **staged** row. Against HEAD's row the same verdict
holds on substance — the mutant and its GREEN→RED transition are identical — but the "run rather
than reasoned" attribution and the all-three-targets generalization are not there. If this SPEC is
consumed from HEAD rather than from the working tree, that difference is real.

I am also recording the concurrency itself: a second actor staged work in this tree inside the audit
window. That is a process observation, not a finding against the SPEC, but an audited worktree
should have one writer.

Reasoning context ignored per M1 Context Isolation. The dispatch's five items were used only to
select attack surfaces and to supply calibration numbers I was told to reproduce rather than cite.
Every judgement below is against the artifacts and the source in this tree.

**Iteration-count note.** Tier M resolves to a ceiling of 2
(`.moai/config/sections/harness.yaml` → `plan_audit_tier_ceilings: M: 2`). This is iteration 3,
which exists only because the operator took the explicit-override branch of the Retry Loop
Contract's post-ceiling escalation. Recorded so the overrun is visible rather than silent.

---

## What I verified here vs what I took on the documents' word

**Verified by executing a command in this tree.** Every finding and every non-finding below rests
on these; nothing in the "Executed" column is quoted from the author or from my prior reports.

| Claim | Command | Result |
|---|---|---|
| merge-base with `origin/develop` | `git merge-base origin/develop HEAD` | `c068667ad8bd5aa60de13367a94b574f9cfe090e` |
| `handleSave` base-side extraction | `git show c068667ad…:internal/web/handlers.go \| awk '/^func (\([^)]*\) )?handleSave\(/{f=1} f{print} f&&/^}$/{exit}' \| wc -l` | **209** |
| `handleSave` head-side extraction | same awk, on `internal/web/handlers.go` | **209** |
| `parseSchemaForm` base / head | same pair on `internal/web/schemaform.go` | **67 / 67** |
| `ApplySchemaEdits` base / head | same pair on `internal/settings/sectionapply.go` | **53 / 53** |
| Naive anchor `^func handleSave\(`, base side | `git show c068667ad…:…handlers.go \| awk '/^func handleSave\(/…' \| wc -l` | **0** |
| Naive anchor, head side | same awk on the working file | **0** |
| `diff` of two empty streams | `diff a.txt b.txt; echo rc=$?` on two empty files | **rc=0** |
| method-form vs plain-form anchor counts | six plain `grep -cE` invocations | `handleSave` 1/0, `parseSchemaForm` 0/1, `ApplySchemaEdits` 0/1 |
| receiver-tolerant anchor on the merge-base copy (control) | `grep -cE "^func (\([^)]*\) )?handleSave\(" <base copy>` | **1** |
| naive anchor on the merge-base copy (control) | `grep -cE "^func handleSave\(" <base copy>` | **0** |
| `handleSave`'s true span | `grep -n "^func " internal/web/handlers.go` + `sed -n '553,564p'` | 350…558 = 209 lines; line 558 is the closing `}` |
| `parseSchemaForm` / `ApplySchemaEdits` true spans | `grep -n "^func "` on each file | 310…376 = 67; 30…82 = 53 |
| AC-WCP-009's two functions are plain funcs | `grep -rnE "^func (\([^)]*\) )?partitionWorkflowFields\("` and same for `isCodexToggleFieldName` | `schemaform.go:185` and `:177`, both plain |
| No raw-string `}` at column 0 in the three targets | `grep -c '\`'` on all three, then `grep -n '\`' handlers.go` | 9 hits in `handlers.go`, all comments or one single-line raw string at `:658-660`, outside `handleSave`; 0 in the other two |
| No conflicting duplicate recipe in `plan.md` / `spec.md` | `grep -n 'AC-WCP-012\|git show\|--stat\|awk'` over both | only two `Decides:` references; no second recipe |
| MP-1 id sequences | `grep -o … \| sort -u` | REQ-WCP-001…012, AC-WCP-001…014 — sequential, no gaps, no duplicates |
| MP-7 clarification gate | `grep -rn 'NEEDS CLARIFICATION'` over the SPEC dir | rc=1, no match |
| MP-6 D8 | `grep -c 'syscall' spec.md` | 0 — auto-pass per D8-4 |

**Explicitly NOT executed:**

- **I ran no Go test, no `go build`, no `make templ-generate`, and no `go vet`.** AC-WCP-012's
  build-and-test arm is therefore unverified by me; I judged only its extraction arm.
- **I did not re-audit the criteria that passed iteration 2.** Per the dispatch, scope is the delta.
  I re-ran only the four cheap must-pass probes above, because a repair can break them.
- **I did not audit HEAD.** Three files are staged-but-uncommitted (see the provenance correction
  above); I read the working tree. The one material difference is the MU-8 row.
- **I could not isolate the repair delta in git.** Iteration 2 audited an *uncommitted* working tree
  at `bd64824ff`; the repair folded v0.2.0 and the repair itself into one commit (`ba7d93411`), so
  `git diff bd64824ff..HEAD` returns 323 changed lines of `acceptance.md` — the union, not the
  delta. I bounded the delta by reading the current artifacts against iteration 2's verbatim quotes
  instead. Stated so it is not read as a git-verified delta.

**Taken on the documents' word:** the operator ruling of 2026-09-07; the scope boundary of card
t517; `verdict.md` as an investigation record.

---

## The five dispatch items

### 1. Does the AC-WCP-012 rewrite close the vacuous pass? — **Yes, and layer 2 is what closes it**

The criterion's Then-clause now reads: *"each extraction yields more than zero lines on **both**
sides, **And** each pair is identical"* — the non-zero condition is normative, ordered first, and
carries an explicit failure name (`EXTRACTION_EMPTY`) plus the sentence that forbids the escape
hatch: *"a **failure**, never a pass — because two empty extractions `diff` clean and report
`IDENTICAL` while having read nothing at all."*

The question posed was whether layer 2 (the assertion) or layer 1 (the anchor) is doing the work.
It is layer 2, and the reasoning is not a preference — it is forced by the failure mode:

- The anchor answers *"does this pattern match the shape present today?"* I measured that it does:
  receiver-tolerant anchor → 1 match on each of the three targets, and the three
  method/plain rows sum to exactly 1 each (1/0, 0/1, 0/1), so no target is both forms and none is
  neither. That measurement is real, and it is also **exhausted the moment a signature changes**.
- The assertion answers *"did the instrument read anything?"* — a property of the run, not of the
  pattern. It holds under any future shape change, including the one that already happened to
  `handleSave` (a plain func that became a method, which is what produced the original defect).

Concretely: a corrected anchor cannot defend against a mis-anchoring, because under a mis-anchoring
the corrected anchor is precisely the thing that is gone. Only a count of what was read survives.
The criterion states this itself — *"The corrected anchor answers 'does it match now?'; the count
answers 'will a changed shape be caught?' — and only the second survives the next signature
change"* — and I am recording that the statement is correct rather than merely well-phrased.

The criterion does **not** lean on the anchor being right. It applies the count to all three
targets, including the two whose current shape the anchor matches only incidentally, and says so
in those words. That is the correct generalization, and it is the opposite of the minimal fix
(repair the one observed-broken anchor) that would have deferred the defect.

### 2. Is MU-8 real? — **Yes. It bites, and the transition is exactly as claimed**

I ran both halves rather than reasoning about them.

- **GREEN under the pre-repair form.** Naive anchor `^func handleSave\(`: base side `0` lines, head
  side `0` lines (both measured above). `diff` of two empty streams exits `0` (measured). The
  pre-repair criterion therefore prints `IDENTICAL handleSave` for the function it never read.
  Confirmed end to end.
- **RED under the repaired form.** The same mis-anchoring yields `0` on both sides, and the Then-
  clause requires `> 0` on both sides. `EXTRACTION_EMPTY` → failure. Confirmed.

The transition is real and it establishes what the mutant claims to establish: the assertion, not
the anchor, is the fix. MU-8 is the only mutant in §D.2 that tests the *instrument* rather than the
*implementation*, and it is the one this defect class needed.

One qualification, verified and worth recording because it is a trap for whoever runs the mutant.
MU-8 names two mechanisms — *"drop the `(\([^)]*\) )?` receiver group, or misspell the function
name"*. The first mechanism is **inert on two of the three targets**: with the receiver group
dropped, `parseSchemaForm` still extracts 67 lines and `ApplySchemaEdits` still extracts 53
(measured). Only the misspell mechanism produces an empty extraction there. The mutant's text is
effect-defined — *"so that target's extraction matches nothing"* — so an implementer who reads the
whole clause picks the right mechanism per target; one who takes the first-named mechanism and
applies it uniformly will observe GREEN on two targets and may conclude the mutant does not bite.
Recorded as D3-2 below, minor and optional.

### 3. The AC-WCP-009 clause — **blast radius, not scope creep**

The adjudication turns on one question: was AC-WCP-009 inside the scope of the defect that drove
the iteration-2 FAIL? It was, and the record is explicit on both sides:

- Iteration 1's D7 covered it. Iteration 1's recommendation #6 reads: *"Replace AC-WCP-012's token
  grep **and AC-WCP-009's `--stat` check** with function-body comparisons (D7)"*
  (`plan-audit.md:292`).
- Iteration 2's regression row for D7 covers it: *"AC-WCP-009's `--stat` check is genuinely gone,
  replaced by the same extraction"* (`plan-audit-iter2.md:355`).

So AC-WCP-009 is not new territory the author wandered into; it is the second half of the same
defect, and it was named as such twice before the repair. The two-item repair list I was given was
narrower than D7's actual footprint — which is a property of the list, not of the author's scope
discipline.

The author's argument is also correct on its merits, and I checked it rather than accepting it.
AC-WCP-009 delegates its verification to AC-WCP-012's recipe (*"the function-body comparison of
AC-WCP-012 applied to these two functions"*). A delegation left unqualified inherits whatever the
recipe becomes — including the repaired recipe — so the clause is arguably belt-and-braces rather
than load-bearing. But the delegated-to text now carries a `[HARD]` assertion whose applicability
to a *different pair of functions* is a judgement an implementer would have to make unaided, and
the clause removes that judgement for the cost of one sentence. Under M6 that is a repair of the
document's internal consistency, which is the blocking class, not the optional one.

The clause's self-description is accurate: I verified both `partitionWorkflowFields`
(`schemaform.go:185`) and `isCodexToggleFieldName` (`schemaform.go:177`) are plain funcs today,
matching the AC's own parenthetical *"both are plain funcs today"*.

The disclosure was unprompted, and the change adds no requirement, touches no other AC, and alters
no milestone. **Not creep.**

### 4. Did the repairs introduce new vacuity? — **No new vacuity. Three residues, all minor**

This was the item with the highest prior probability of a finding, and I attacked it on five
vectors rather than one. Four came back clean; the residues are recorded as D3-1…D3-3.

- **Early-termination / prefix-read.** The extraction terminates on the first column-0 `}` after the
  anchor. A raw string literal containing a line `}` at column 0 would truncate the body silently,
  and the non-zero assertion would *not* catch it (the count stays non-zero). This is the strongest
  vector I found, and it does not bite: I verified each extraction reaches the function's true end —
  `handleSave` 350…558 (line 558 is the closing brace, next func at 564), `parseSchemaForm` 310…376,
  `ApplySchemaEdits` 30…82 — and that the only backticks in the three files are in comments plus one
  single-line raw string at `handlers.go:658-660`, outside all three targets. Latent, not live;
  recorded as D3-1, optional.
- **Multiple anchor matches.** If two same-named methods existed in one file, awk would compare only
  the first. Measured: exactly 1 match per target. Not live.
- **Anchor over-match on a name prefix.** The anchor requires the literal `(` after the name, so
  `handleSaveDraft(` cannot match. Clean.
- **Awk portability.** The ERE optional-group form runs correctly under this machine's awk — I ran
  it six times. Clean.
- **A second, stale copy of the recipe.** `plan.md` and `spec.md` carry no competing recipe (grep
  above); M6's gate says *"Confirm the zero-new-key invariant by diff, not by memory"* and delegates
  the rest. The recipe has a single home. Clean.

Two smaller residues did come from the repair itself, and I record them rather than waving them
through: the deciding `diff` step has no literal invocation while the non-deciding counts do
(D3-3), and AC-WCP-009's normative Then-clause omits the non-zero condition its own instrument line
now enforces (D3-4). Both are minor, and neither reintroduces a silent pass.

The prior-defect count on this card (two vacuous passes, the second introduced by the repair of the
first) justified looking hard. It does not justify manufacturing a third. I looked, and the third
is not there.

### 5. Independent reproduction of the extraction counts

Run by me, in this tree, as six plain commands with literal paths — the decomposed shape the
criterion specifies, which is indeed runnable here where the loop form is not:

| target | base side (`c068667ad`) | head side | agrees with the author |
|---|---|---|---|
| `handleSave` (`internal/web/handlers.go`) | **209** | **209** | yes |
| `parseSchemaForm` (`internal/web/schemaform.go`) | **67** | **67** | yes |
| `ApplySchemaEdits` (`internal/settings/sectionapply.go`) | **53** | **53** | yes |

Mis-anchored control: naive `^func handleSave\(` → **0 / 0**, with `diff` exit **0**. The
receiver-tolerant anchor on the same merge-base copy → **1**; the naive anchor on the same copy →
**0**. Every number in the dispatch's calibration block reproduced exactly, and every number in the
criterion's own table reproduced exactly.

The reproduction is what makes the criterion's reported figures an attributed baseline rather than
a self-report: the repair side measured its own output, and an independent run in the same tree
returned the same values.

---

## Must-Pass Results

Re-probed only where a repair could have broken them; the rest stand from iteration 2.

- **[PASS] MP-1 REQ number consistency** — REQ-WCP-001…012, AC-WCP-001…014, sequential, no gaps, no
  duplicates, uniform padding (measured this iteration).
- **[PASS] MP-2 GEARS format compliance** — judged against the requirement layer (`spec.md §C`)
  only. Unchanged by the delta; the repair touched the verification layer. Given-When-Then in
  `acceptance.md` is the correct verification-layer format and is graded in Group 4.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types;
  `version: "0.2.0"` is a quoted semver string, `created`/`updated` ISO, no rejected snake_case
  alias. (The *staleness* of that version value is D3-5 below; MP-3 tests presence and type, and
  both hold.)
- **[N/A] MP-4 language neutrality** — single-language (Go/templ) SPEC scoped to `internal/web`.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — referenced IDs unchanged by the delta; no BLOCKING.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall' spec.md` → 0. Auto-pass, D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION'` over the SPEC dir → rc=1,
  no match.

No must-pass failure.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75–1.0 | AC-WCP-012 gained an explicit merge-base derivation with its rationale, a measured three-row shape table, and a named failure state — all three reduce interpretation. Residues: AC-WCP-013's "shows no field count (zero)" still carries two readings (iteration 2's D2-5, unrepaired, optional), and the merge-base is called `$BASE` in prose but `<BASE>` in the command sample (D3-6) |
| Completeness | 0.90 | 0.75–1.0 | All sections present; four `### Out of Scope — <topic>` H3s with specific bullets; frontmatter complete. The delta added the missing baseline derivation and the missing failure state. Two small gaps: the deciding `diff` step has no literal invocation (D3-3) and HISTORY does not record this iteration (D3-5) |
| Testability | 0.85 | 0.75–1.0 | Iteration 2's sole Testability blocker is closed and I reproduced the closure end to end: the MUST criterion guarding the save path now cannot pass on an unread function, and MU-8 demonstrates the transition GREEN→RED. Every remaining residue is "measurable with minor interpretation" — the 0.75 anchor's own language — and none of them sits on a criterion that can pass silently. Not 1.0: AC-WCP-003's countable token is still unnamed (D2-4, unrepaired, optional) and AC-WCP-009's Then-clause under-states its instrument (D3-4) |
| Traceability | 0.95 | 0.75–1.0 | Every REQ-WCP-001…012 has ≥1 AC; every AC-WCP-001…014 maps to an existing REQ (§D.3 table). Unchanged by the delta except favourably: MU-8 gives AC-WCP-012 the mutant coverage it lacked. AC-WCP-009's delegated assertion has no mutant of its own (D3-2's sibling), which is the only reason this is not 1.0 |

Aggregate (harmonic mean, same method as iterations 1 and 2 for comparability): **0.89**.
`4 / (1/0.85 + 1/0.90 + 1/0.85 + 1/0.95) = 4 / 4.5167 = 0.8856`.

Score trajectory: **0.69 → 0.84 → 0.89**. Improving, so the LEAN score-regression STOP clause does
not fire and no scope reduction is proposed.

---

## Why this is a PASS, stated plainly

The verdict rests on two things and I want neither of them to look like a judgement call.

**The must-pass firewall is clean** — seven criteria, no failure, four of them re-probed this
iteration rather than inherited.

**The one blocking defect that forced iteration 2's FAIL is resolved, and I verified the resolution
by running it rather than by reading the repair.** Iteration 2 failed a 0.84 because
`verification-completeness.md`'s mutant test showed AC-WCP-012 was too shallow to adopt: a mutant
editing `handleSave` satisfied the criterion while violating REQ-WCP-011. That mutant now fails the
criterion — `EXTRACTION_EMPTY` on a zero read — and the mutant that proves it (MU-8) is written
into §D.2 with its pre-repair GREEN observed rather than asserted. Iteration 2's other blocking
defect (D2-2, the moving-ref base) is likewise resolved: the base is now `git merge-base
origin/develop HEAD`, computed explicitly, with the moving-ref hazard named and rejected in the
text.

**On the regression clause.** The Retry Loop Contract says unresolved prior-iteration defects are
automatically FAIL. Read literally against a defect list that includes optional findings, that
clause never releases a SPEC — which is precisely the over-engineering brake M6 exists to prevent
("a long list of optional findings does not by itself justify a FAIL, and it must not be used to
manufacture one"). I read the clause as binding the **blocking** class, and I say so explicitly
rather than letting the reconciliation happen silently. Both iteration-2 blocking defects are
resolved. The four optional ones (D2-3, D2-4, D2-5, D2-6) are unrepaired and remain the
orchestrator's discretion; none of them can produce a silent pass, and three of them are text/
instrument mismatches where the named instrument already does the right thing.

The six findings below are all minor. Not one of them is a criterion that can pass without
observing what it claims to observe, which is the property this card has failed twice on and now
holds.

---

## Defects Found

**D3-1 — the extraction can silently read a prefix, and the non-zero count does not catch that**
`acceptance.md` § AC-WCP-012, the awk terminator — Severity: **minor** — Class: **optional**
The extraction stops at the first column-0 `}` after the anchor. A raw string literal containing a
line `}` at column 0 inside one of the three bodies would truncate the extraction on both sides
symmetrically: the count stays non-zero (so `EXTRACTION_EMPTY` does not fire), the `diff` is clean,
and an edit past the truncation point passes. Verified **not live**: each extraction reaches its
function's true end (`handleSave` 350…558, `parseSchemaForm` 310…376, `ApplySchemaEdits` 30…82,
each matching its measured line count), and the only backticks in the three files are comments plus
one single-line raw string outside all three targets.
*Why optional*: a hypothetical the SPEC does not claim to cover, on a construct absent from all
three targets. Routing it would add a guard for a case that does not exist.
*If the orchestrator wants it closed*: assert the extraction's last line is `}` and that the count
equals `<next top-level func line> − <anchor line>` — or simply record the three expected counts as
pinned values, which the criterion already does in prose.

**D3-2 — MU-8's first-named mechanism is inert on two of its three targets**
`acceptance.md` §D.2, MU-8 row — Severity: **minor** — Class: **optional**
MU-8 offers two mechanisms: drop the receiver group, or misspell the name. Measured: with the
receiver group dropped, `parseSchemaForm` still extracts **67** lines and `ApplySchemaEdits` still
extracts **53** — the mutation is inert there, because both are plain funcs. Only the misspell
mechanism empties those two. The mutant is effect-defined ("so that target's extraction matches
nothing"), so a careful reader picks correctly; a reader who applies the first-named mechanism
uniformly observes GREEN on two of three and may record that the mutant does not bite — the inverse
of what the mutant is for.
*Required fix (one clause)*: "on a plain-func target use the misspell mechanism; the receiver-group
drop empties only a method target."

**D3-3 — the deciding `diff` step has no literal invocation, while the non-deciding counts do**
`acceptance.md` § AC-WCP-012, the six-command block — Severity: **minor** — Class: **blocking**
The block shows six `… | wc -l` commands verbatim, then describes the comparison in prose: "then a
`diff` of the two extractions per target". The count is diagnostic-turned-normative; the `diff` is
the identity verdict — and it is the one with no runnable form. An implementer's obvious
construction is a process substitution (`diff <(git show …|awk …) <(awk …)`), which is a compound
form the same worktree guard that refused the loop is likely to refuse, pushing them back to a
shape the criterion has not sanctioned — after the criterion went to the trouble of explaining
exactly which shapes are runnable here.
*Why blocking rather than optional*: the criterion states an obligation (`each pair is identical`)
whose instrument it does not supply, in a criterion whose whole subject is instruments that do not
do what they appear to do. It is one command to fix.
*Required fix*: show the two-file form — redirect each side to a file, then `diff base.txt
head.txt; echo "diff_rc=$?"` — matching the "plain commands, literal paths" discipline the block
already establishes for the counts.

**D3-4 — AC-WCP-009's Then-clause under-states the instrument its own "Decided by" line enforces**
`acceptance.md` § AC-WCP-009 — Severity: **minor** — Class: **optional**
The Given-When-Then asserts only that the two bodies "are byte-identical to their pre-change
forms". The non-zero-extraction obligation appears one paragraph later, in the "Decided by" line.
AC-WCP-012 puts it in the Then-clause, where it is normative. So the two ACs that share one
instrument state it at different strengths, and AC-WCP-009's normative sentence is satisfiable by
the exact vacuous pass the delegated instrument forbids.
*Note*: this is a real asymmetry introduced by the delta, and it is also the smallest possible one —
the delegation is unambiguous and the instrument line is explicit. It does not create a passable
gap for anyone reading the AC through.
*Required fix*: lift "**Then** each extraction yields more than zero lines on both sides, **And**
the bodies are byte-identical" into the Then-clause.

**D3-5 — the delta is not recorded in HISTORY, and `version:` still reads 0.2.0**
`spec.md` § HISTORY / frontmatter — Severity: **minor** — Class: **blocking**
`version: "0.2.0"` and the HISTORY row for 0.2.0 describe the **iteration-1** repairs. The
iteration-2 repairs — a substantive rewrite of AC-WCP-012, a new MU-8, and the AC-WCP-009 clause —
are in the artifacts with no row of their own; the commit message itself says "v0.2.x", so the
author knew a version had moved. The SPEC's own change record therefore describes a document that
no longer exists.
*Why blocking*: internal consistency between the artifact and its self-description — the M6
blocking class — not a preference. It does not affect any criterion's ability to decide, hence
minor severity.
*Required fix*: bump to `0.2.1` and add one HISTORY row naming the iteration-2 repairs and this
report.

**D3-6 — `$BASE` in prose, `<BASE>` in the command sample**
`acceptance.md` § AC-WCP-012 — Severity: **minor** — Class: **optional**
The text says "Its output is `$BASE` below"; the sample command reads `git show
<BASE>:internal/web/handlers.go`. The angle-bracket form is the better one here — it signals
substitute-a-literal and keeps this base visibly distinct from AC-WCP-011's `BASE=origin/develop`,
which is a different value with the same name. The prose sentence undoes that distinction in one
word.
*Required fix*: say "Its output is the `<BASE>` substituted below."

---

## Regression Check (iteration 2 defects)

| # | Defect | Status | Evidence |
|---|---|---|---|
| D2-1 | AC-WCP-012's extraction reads nothing for `handleSave` and reports `IDENTICAL` (critical, blocking) | **RESOLVED** | Both layers verified by execution. Anchor: receiver-tolerant form extracts 209/209 for `handleSave` (I ran all six commands). Assertion: the Then-clause requires `> 0` on both sides and names a zero `EXTRACTION_EMPTY`, a failure. MU-8 demonstrates the transition — pre-repair 0/0 with `diff` rc=0 → `IDENTICAL` (GREEN, reproduced), post-repair the same mis-anchoring fails on the count (RED) |
| D2-2 | AC-WCP-012 and AC-WCP-011 share `$BASE` across two semantics; extraction base was a moving ref (major, blocking) | **RESOLVED** | The base is now `git merge-base origin/develop HEAD`, computed explicitly, with the moving-ref hazard named and rejected in the criterion's own text. AC-WCP-011 keeps three-dot `$BASE...HEAD`. Residual naming slip only: D3-6, optional |
| D2-3 | AC-WCP-005 admits a `0 == 0` pass (minor, **optional**) | UNRESOLVED | AC-WCP-005 text unchanged. Optional; the named helper `assertPanelFields` still carries the presence check the AC text omits |
| D2-4 | AC-WCP-003's "mirror row count > 0" names no countable token (minor, **optional**) | UNRESOLVED | AC-WCP-003 text unchanged |
| D2-5 | AC-WCP-013's "shows no field count (zero)" is two assertions (minor, **optional**) | UNRESOLVED | AC-WCP-013 text unchanged |
| D2-6 | REQ-WCP-012 carries implementation detail into the requirement layer (minor, **optional**) | UNRESOLVED | Raised in iteration 2 for orchestrator decision, not routed as a fix; unchanged, and the author's rationale (§C.2) remains the defensible edge of RQ-4 it was |

Both blocking defects are resolved. The four unresolved ones are all optional-class and all
minor-severity; per M6 they are surfaced for the orchestrator's discretion and do not carry the
verdict.

**Stagnation check**: no defect appears unchanged across all three iterations *in the blocking
class*. D2-4 and D2-5 originate in iteration 2 and this is their second appearance, both optional.
No stagnation flag.

---

## Recommendation

**PASS.** No hedging: the must-pass firewall is clean, the aggregate is 0.89 against a Tier M
threshold of 0.80, and the single blocking defect that forced iteration 2's FAIL is closed on the
axis that matters — I reproduced the vacuous pass under the old form and the failure under the new
one, in this tree, rather than accepting the repair's account of itself.

The repair is also better than the minimum. Fixing the one broken anchor would have satisfied the
literal finding and left the class open; asserting a non-zero read on all three targets closes the
class, and the author says why in the criterion instead of leaving the reader to work it out. The
AC-WCP-009 clause is the same instinct applied one AC over, and it was inside D7's scope from
iteration 1 onward.

Three iterations of finding defects did not oblige a fourth, and the sixth-vector hunt in item 4
came back without one.

For the orchestrator, in priority order:

1. **D3-5** — bump `version:` to `0.2.1` and add the HISTORY row. One row; the artifact currently
   misdescribes itself.
2. **D3-3** — show the `diff` invocation. One command, in the criterion whose subject is
   instruments that do not do what they appear to do.
3. **D3-4** — lift the non-zero clause into AC-WCP-009's Then. One sentence.
4. **D3-2, D3-6** — one clause each, at the orchestrator's discretion.
5. **D3-1** and iteration 2's D2-3…D2-6 — surfaced, not routed. Routing them would add guards for
   cases these documents do not claim to cover.

None of these blocks run-phase entry. The Implementation Kickoff Approval human gate is unaffected
by this verdict and remains required.

---

## Residual risk

- **AC-WCP-012's build-and-test arm is unverified by me.** I ran no `go build`, no `go vet`, no
  `make templ-generate`, and no test. The criterion's extraction arm is verified; its green-build
  arm is not.
- **The repair delta is not git-isolable.** Iteration 2 audited an uncommitted tree; the repair
  commit folds v0.2.0 and the repair together. I bounded the delta by reading against iteration 2's
  verbatim quotes, which is weaker than a diff. If a change landed inside the v0.2.0 surface that
  iteration 2 had already cleared, and it is not in AC-WCP-012 / AC-WCP-009 / §D.2, this audit did
  not look at it.
- **Everything outside the delta is inherited.** Criteria that passed iteration 2 were not re-read
  except for the four must-pass probes. A repair-induced break outside AC-WCP-012, AC-WCP-009, and
  §D.2 would not have been seen.
- **`origin/develop` moves.** All measurements are pinned to merge-base `c068667ad`, which is
  stable by construction — but the merge-base itself changes if this branch is rebased, and the
  three reported line counts would then need re-measuring. The criterion is correct about this; the
  numbers in it are a snapshot.
- **The extraction's prefix-read vector (D3-1) is verified absent today, not structurally
  impossible.** A future raw string literal in any of the three bodies reopens it silently.
- **The verdict is against an uncommitted tree.** Three files are staged and not committed. If they
  are amended further, or committed differently, this audit's subject and the eventual commit are
  not the same artifact. The staged MU-8 row in particular is what item 2 was judged against.
