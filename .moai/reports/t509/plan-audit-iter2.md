# SPEC Review Report: SPEC-WEB-CODEX-PANEL-001

Iteration: 2/2 (Tier M ceiling — `.moai/config/sections/harness.yaml:75-78`, `M: 2`)
Verdict: **FAIL**
Overall Score: **0.84** (Tier M PASS threshold 0.80 — `spec-workflow.md` § SPEC Complexity Tier)

Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t509`, branch `WT-codex-model-config`,
HEAD `bd64824ff`. The v0.2.0 artifacts are **uncommitted** (`git status --short` → 4 modified files
under `.moai/specs/SPEC-WEB-CODEX-PANEL-001/`); this audit reads the working tree.

Reasoning context ignored per M1 Context Isolation. The dispatch's five repair claims were used only
to select attack surfaces; every judgement below is against the artifacts and the source in this tree.

**The aggregate clears the threshold and the verdict is still FAIL.** Those two facts are not in
tension by accident — the reason is stated in § Why FAIL at 0.84, and it is the Retry Loop Contract's
regression clause, not a score.

---

## What I verified here vs what I took on the documents' word

**Verified by executing a command in this tree** (every finding below rests on these):

| Claim | Command / source read | Result |
|---|---|---|
| `origin/develop` current value | `git rev-parse origin/develop` | `d4162b368237690caa72f4393005f941a917bf29` — matches the author's report |
| merge-base with `origin/develop` | `git merge-base origin/develop HEAD` | `c068667ad8bd5aa60de13367a94b574f9cfe090e` |
| The three save-path files did **not** move on `origin/develop` since the merge-base | `git diff --name-only c068667ad origin/develop -- internal/web/schemaform.go internal/web/handlers.go internal/settings/sectionapply.go` | empty output |
| `panelHTML`'s end-of-document fallback exists and behaves as the SPEC describes | `internal/web/tab_layout_test.go:76-88` read directly | confirmed verbatim (see F3 below) |
| `assertPanelFields` shape | `tab_layout_test.go:90-103` | confirmed; carries a presence check the AC text omits |
| AC-WCP-012's extract anchor `^func <name>(` match counts | `grep -c "^func $n(" <3 files>` per name | `parseSchemaForm` 1, **`handleSave` 0**, `ApplySchemaEdits` 1 |
| `handleSave` is a method, not a function | `grep -n "^func .*handleSave(" internal/web/handlers.go` | `350:func (a *app) handleSave(w http.ResponseWriter, r *http.Request) {` |
| The pinned 11 codex-token names are exactly the registry's codex population | `grep -rn '"codex' internal/settings/` (5 hits: `schema_sections.go:398,415,416,427,428`) + `grep -n 'codex_' internal/mcp/catalog.go` (6 hits: `:50-55`) | 5 + 6 = **11**, matching spec.md §C.1 name-for-name |
| Those 11 are reachable from `settings.AllFields()` | call chain read: `accessors.go:8 AllFields → schema.go:312 allFields → schema.go:500 sectionExtraFields → schema_sections.go:610 seamSectionFields` (the 5) `+ :614 mcpFields` (the 6) | confirmed — AC-WCP-006's registry-sweep arm is satisfiable, not red-at-arrival |
| `schemaPanelMeta` returns an empty meta for an unregistered panel | `internal/web/schemaform.go:283-290` | confirmed — MU-6's self-disclosure ("the count arm alone would not bite") is honest |
| The rail renders the count unconditionally | `internal/web/shell.templ:156` — `<span class="count">{ itoa(t.Fields) }</span>` | a `Fields` of 0 renders the literal `0`, not an absent element |
| No surviving "green on the pre-change tree" claim | `grep -rn "pre-change tree\|must be green\|green on the" .moai/specs/SPEC-WEB-CODEX-PANEL-001/` | one hit, `acceptance.md:74`, and it is the **corrected** narrative |
| No surviving "≥12" floor as a criterion | `grep -rn "twelve\|≥12\|at least the twelve"` over the SPEC dir | 4 hits, all narrative/history; none is a deciding assertion |
| REQ/AC id sequences | `grep -o` + `sort -u` | REQ-WCP-001…012 sequential, no gaps, no duplicates; AC-WCP-001…014 |

**Explicitly NOT executed, and why:**

- **I did not run the AC-WCP-012 extraction loop.** I tried; the worktree-session guard refused it
  (`git show` inside a compound command). The finding D2-1 below therefore rests on the *measured*
  anchor-match count of **0** plus awk's semantics (`f` is never set, so nothing prints) — a
  deduction from a measurement, not an executed observation. Stated as such; do not read it as a
  reproduction.
- **I ran no Go test and wrote no test file.** Iteration 1 used two throwaway tests; this iteration
  was instructed to change nothing but the report, so the 11-name population and the `AllFields`
  reachability above were established by **reading the call chain**, not by executing it. They
  corroborate iteration 1's runtime measurement; they do not replace it.
- I did not run the `internal/web` suite, did not build, and did not measure the tab-switching
  mechanism (the SPEC declares it unmeasured and out of scope).

**Taken on the documents' word:** `.moai/reports/t509/verdict.md` as an investigation record; the
operator ruling of 2026-09-07; the scope boundary of card t517.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — REQ-WCP-001…012, sequential, no gaps, no duplicates,
  uniform 3-digit padding (measured by `grep -o … | sort -u`). REQ-WCP-012 is the new entry and
  extends the sequence correctly.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`spec.md §C`),
  never against ACs. All 12 match a GEARS pattern; the new REQ-WCP-012 (`spec.md:210`) is Ubiquitous
  ("The rail subnav count for the codex tab shall be zero…"). Given-When-Then in `acceptance.md` is
  the correct verification-layer format and is graded in Group 4 only.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types
  (`spec.md:2-13`); `version: "0.2.0"` quoted, `created`/`updated` ISO, no rejected snake_case alias.
- **[N/A] MP-4 language neutrality** — single-language (Go/templ) SPEC scoped to `internal/web`.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — referenced IDs unchanged from iteration 1; all
  exist with `status: completed`. No BLOCKING.
- **[PASS] MP-6 D8 cross-platform discipline** — no `syscall` in the SPEC body. Auto-pass per D8-4.
- **[PASS] MP-7 clarification gate** — no `[NEEDS CLARIFICATION]` markers in `plan.md`; no
  `research.md` (correct for Tier M).

No must-pass failure. Tier M REQ/AC budget: 12 REQ and 14 AC, both inside the 16/16 ceiling.

---

## Repair-by-repair verdict (the five claims, checked against the tree)

### 1. F1 — AC-WCP-005 rewritten onto the cross-panel axis — **HOLDS**

The greenness claim is **gone**. `acceptance.md:73-78` now says the first draft's assertion was "RED
on the pre-change tree and unimplementable", cites `boolSegment` as the structural cause, and
attributes the 54-of-125 figure to me with the explicit disclaimer "that measurement is the
auditor's, not re-run here, and **nothing below depends on its exact value**". That is the correct
disposition: an unreproducible measurement is cited as provenance, not adopted as a baseline
(`verification-completeness.md` § 2.1, undecidable disposition).

Nothing else depends on it — swept: the only surviving "pre-change tree" string in the SPEC directory
is that corrected sentence itself.

The new axis is the right one and reuses a landed shape: `assertPanelFields`
(`tab_layout_test.go:90-103`) compares `strings.Count(html, marker)` against
`strings.Count(body, marker)` and fails with "a field must live on exactly one tab". Non-vacuity is
supplied by MU-1. **One gap** — see D2-3.

### 2. F2 — AC-WCP-011 split into three observed steps — **HOLDS; the iteration-1 trap is closed**

I tried to defeat the new form and could not, on the axis it was written for:

- **Broken/mistyped/unfetched ref** — closed by step 1. `git rev-parse --verify -q "$BASE"` fails on
  `origin/nosuchref`, so `ref_resolves_rc` is non-zero and the criterion fails before the grep runs.
  The exact trap I reproduced in iteration 1 can no longer produce a pass.
- **Empty diff (mis-pathed tree, BASE == HEAD)** — closed by step 2's `wc -l > 0`.
- **grep error vs no-match** — a malformed pattern yields `rc=2`, not `1`; the criterion demands
  exactly `1`.
- **Stale-but-resolving base** — passes, but in the *conservative* direction: `A...HEAD` is
  three-dot, so a staler base yields a **superset** of changed paths, making the filter more likely
  to match, not less. Not a defect.

The AC's own prose (`acceptance.md:180-185`) records the reproduction and states the reason
correctly: "One exit status cannot carry two facts". Residual, not a finding: the three steps prove
*a* diff exists, not that *this card's* change is in it — a wrong-tree run with unrelated commits
would satisfy all three. Marginal, and out of the axis the repair targeted.

### 3. F3 — AC-WCP-008 panel scoping and the `panelHTML` fallback — **VERIFIED; this is the strongest repair in the set**

The fallback exists and behaves exactly as described. `internal/web/tab_layout_test.go:76-88`:

```go
rest := html[start+len(`data-panel="`+panel+`"`):]
if next := strings.Index(rest, `data-panel="`); next >= 0 {
    rest = rest[:next]
}
return rest
```

There is no `else`. For the **last** panel the slice runs to end-of-document — so a codex panel
placed last would silently turn every panel-scoped assertion into a whole-page one. The author's
claim that this would "widen AC-003 and AC-008 into vacuity at the same time" is correct, and it
would do so **without any test failing**, which is what makes it worth a criterion of its own.

AC-WCP-002's not-last assertion, the "Panel scoping" preamble (`acceptance.md:9-14`), and plan M2's
placement note (`plan.md:72-75`) close it. The guard is real: it protects by making the *suite* red
(the not-last assertion lives in `TestCodexPanel_NoNamedFormElements`, a different test from the ones
it protects), which is adequate — but note it as residual: no single test's own verdict tells you its
region was correctly bounded.

### 4. REQ-WCP-005 / AC-WCP-006 — the exception and the pinned list — **HOLDS**

The pinned 11 are correct against the tree, name for name:
`workflow.audit.gates.codex` (`schema_sections.go:398`), `workflow.audit.codex.model` (`:415`),
`workflow.audit.codex.effort` (`:416`), `workflow.codex.review_gate.enabled` (`:427`),
`workflow.codex.task.allow_write` (`:428`), and the six `mcp.tools.codex_{audit,setup,task,
job_status,job_result,job_cancel}.enabled` derived from `internal/mcp/catalog.go:50-55`.

The registry sweep arm is satisfiable: `AllFields()` (`accessors.go:8`) → `allFields()`
(`schema.go:312`) → `sectionExtraFields()` (`schema.go:500`) → both `seamSectionFields()` and
`mcpFields()` (`schema_sections.go:610,614`). All 11 are reachable. Had `mcpFields` not been in that
chain, the arm would have collected 5 and failed against a pinned 11 — red at arrival for the wrong
reason. It is not.

**Set equality cannot pass on an empty or partial set**, in both directions: a non-empty pinned
literal versus an empty predicate output is unequal, and a partial output is unequal by the missing
member. MU-3 bites through disagreement with the pinned list rather than through a count floor, which
is the correct non-circular carrier. The "≥12" floor is gone as a criterion (swept — the four
surviving mentions are narrative). REQ-WCP-005 now declares `workflow.audit.model` as one named
exception, AC-WCP-014 asserts it separately and asserts the predicate does **not** contain it, and
§C.1 defends why a pinned list in a *test* is not the hand-enumeration the requirement forbids. That
distinction is sound: the prohibition binds the surface where an omission is silent.

### 5. settings_shell.go / D6 — **the author is right, and my iteration-1 framing was wrong**

Stated plainly: **D6's framing was my error.** I wrote that the rail would render "0 beside a panel
showing ~12 rows" as though the invariant were *rail == rendered row count*. It is not. The in-code
contract (`internal/web/settings_shell.go:112-118`) says the rail number "must equal the number **the
panel puts in its own header**", and the `mcp` arm proves the author's reading: `settingsTabFieldCount`
returns `len(settings.SectionFields(settings.SectionMCP))` (`:126`) while `settingsTabFieldNames` for
the same tab returns that list **plus** `workflow.codex.review_gate.enabled`,
`workflow.codex.task.allow_write`, and `glmAPIKeyFormField` (`:145-151`). The count therefore already
excludes three entries that render on the panel, and the comment says why: they are "필드가 아니라
별도 상태 표면" and "패널 머리글도 세지 않는다". Rail-equals-header, with rendered rows deliberately
excluded, is the established behaviour.

So zero is the honest number for a panel that owns no fields and whose header claims none, and §C.2's
reasoning is the code's own reasoning. My finding's *surviving* half is the one the repair addressed:
the coupling was named nowhere and no criterion covered it. That is now closed — `settings_shell.go`
in spec.md §F, §C.2, REQ-WCP-012, AC-WCP-013, MU-6, plan M2's explicit `case`.

MU-6 is honest in a way worth crediting: it discloses that removing the `case` leaves the count at 0,
so the count arm alone would not bite — which I verified (`schemaform.go:283-290` returns
`schemaSectionMeta{PanelID: panelID}`, zero fields). That is why the explicit-case grep is asserted
separately, and the author says so rather than letting the mutant look stronger than it is.

---

## Fresh hunt: did the repairs introduce new vacuity?

Three targets, per the dispatch. Two are clean; one is not.

- **AC-WCP-006's set equality** — cannot pass on an empty or partial set (above). Clean.
- **The three-part AC-WCP-011 form** — cannot pass on a broken ref, an empty diff, or a filter
  error; the one exit code that gates (`filter_rc`) is now gated *by* two prior observations rather
  than carrying their weight. Clean.
- **The panel-scoped sweeps** — AC-WCP-003 explicitly asserts the mirror-row count `> 0` "so a
  slicing bug producing an empty region cannot pass" (`acceptance.md:49-51`), and `panelHTML`
  `t.Fatalf`s when the marker is absent. Clean on the empty-region axis; one unnamed-token gap
  (D2-4).
- **AC-WCP-012's function-body extraction** — **NOT clean.** This is the new finding, and it is the
  same defect class the repair was written to close. See D2-1.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75–1.0 | Requirements unambiguous; §C.1 and §C.2 close the two decisions iteration 1 found open, and the "Panel scoping" preamble (`acceptance.md:9-14`) removes the scoping ambiguity that produced D3. Two minor residues: AC-WCP-013's "shows no field count (zero)" is two different assertions against an unconditional `itoa(t.Fields)` render (D2-2), and AC-WCP-003's "mirror rows > 0" names no countable token (D2-4) |
| Completeness | 0.90 | 0.75–1.0 | All sections present, frontmatter complete, four `### Out of Scope — <topic>` H3s with specific bullets. Iteration 1's sole completeness gap (the `settings_shell.go` coupling) is closed in three places: `spec.md §F:310-311`, `§C.2:235-250`, `plan.md:84-88` |
| Testability | 0.70 | just below the 0.75 anchor | Three of iteration 1's four testability defects are repaired and verified (D1, D2, D3). One remains, and it is worse than the 0.75 anchor's "measurable with minor interpretation": AC-WCP-012 reports `IDENTICAL handleSave` **silently**, having read nothing (D2-1). A silent vacuous pass in a MUST criterion is not an interpretation gap |
| Traceability | 0.95 | 0.75–1.0 | Every REQ-WCP-001…012 has ≥1 AC; every AC-WCP-001…014 maps to an existing REQ (`acceptance.md:279-297`). Iteration 1's D5 (REQ-WCP-009's self-referential carrier) is closed: §D.3 now names the **registry-sweep arm** as the non-circular carrier and explains why the predicate comparison alone cannot be one — and I verified that arm reaches all 11 fields, so the carrier actually carries |

Aggregate (harmonic mean, same method as iteration 1 for comparability): **0.84**.
`4 / (1/0.85 + 1/0.90 + 1/0.70 + 1/0.95) = 4 / 4.769 = 0.839`.

Score trajectory: **0.69 → 0.84**. This is an improvement, so the LEAN score-regression STOP clause
does **not** fire and no scope reduction is proposed.

---

## Why FAIL at 0.84, when 0.84 ≥ 0.80

These two facts have to be reconciled openly rather than resolved by picking whichever supports a
preferred verdict.

The aggregate clears the Tier M threshold. The FAIL comes from a different clause: the Retry Loop
Contract's regression check — *"For each defect listed in the previous iteration's report, verify
whether it was resolved. Unresolved defects from a prior iteration are automatically FAIL regardless
of other scores."*

Iteration 1's D7 said AC-WCP-012's deciding command could not decide its claim. The repair adopted
exactly the right mechanism — body extraction and comparison across revisions — and that mechanism
works for two of its three subjects. For the third, `handleSave`, the instrument reads nothing and
reports success. D7 is therefore **partially resolved**, which under the regression clause is
unresolved.

I want to be explicit that I am not manufacturing this. Two things make it a real block rather than a
scored deduction:

1. The criterion it lives in is the one guarding **REQ-WCP-011** — no save path touched — the
   scope invariant whose violation this whole card is shaped around, and whose failure mode is a
   silent config write.
2. It fails the mutant probe the always-loaded rule prescribes
   (`verification-completeness.md` §2): a mutant that edits `handleSave`'s body satisfies the
   criterion while violating the requirement. By that rule's own test, the criterion is too shallow
   to adopt.

Had the defect been anywhere but a MUST criterion protecting the save path, 0.84 would have been a
PASS with the finding logged as debt.

**This is iteration 2 of a Tier M ceiling of 2.** There is no iteration 3 available under the
contract, so this verdict escalates to the operator with the three standard options. My
recommendation is stated at the end — the fix is one line of a document, and I say so plainly so the
escalation is not read as a scope problem.

---

## Defects Found

**D2-1 — AC-WCP-012's extraction reads nothing for `handleSave` and reports `IDENTICAL`**
`acceptance.md:196-203` — Severity: **critical** — Class: **blocking**
The extraction anchors on `/^func $2\(/`. Measured in this tree:
`grep -c "^func handleSave(" internal/web/handlers.go` → **0**, because the function is a method:
`internal/web/handlers.go:350` reads `func (a *app) handleSave(w http.ResponseWriter, r *http.Request) {`.
The awk program only sets `f` on a matching line, so with zero matches it prints nothing on **both**
sides; `diff` of two empty streams exits 0 and the loop prints `IDENTICAL handleSave`. The criterion
reports success for the one function it never read, and it does so silently — the exact
report-not-verdict shape `verification-completeness.md` §1.1 names, reintroduced by the repair that
was closing it. The other two subjects are fine: `^func parseSchemaForm(` → 1
(`schemaform.go:310`), `^func ApplySchemaEdits(` → 1 (`sectionapply.go:30`).
*Honesty note*: I measured the anchor counts; I did **not** execute the extraction loop (the
worktree guard refused the compound `git show` form). The vacuous-pass consequence follows from awk
semantics, not from an observed run.
*Required fix*: anchor on a form that admits a receiver — e.g. `/^func (\([^)]*\) )?handleSave\(/`,
or more simply `/^func .*[ (]handleSave\(/` — and, because the anchor is now the load-bearing part,
add a self-check to §D.4 that each of the three extractions produced a **non-empty** body before its
`diff` verdict is read. An extraction that yields zero lines is a gap, never an `IDENTICAL`.

**D2-2 — AC-WCP-012 and AC-WCP-011 share `$BASE` across two different semantics**
`acceptance.md:171,197-201` — Severity: **major** — Class: **blocking**
AC-WCP-011 uses `$BASE...HEAD` (three-dot: changes on HEAD since the merge-base — correct, and
immune to upstream movement). AC-WCP-012 reuses the same `$BASE` as `git show "$BASE:$file"`, which
reads the **tip** of `origin/develop`, not the merge-base. `origin/develop` is a moving ref and has
already moved during this card's life (iteration 1 saw `ace1c5440`; it is now `d4162b368`). An
invariance assertion pinned to a moving ref is precisely what `verification-completeness.md` §4
prohibits: once another actor edits one of the three files on develop, the comparison reports
`CHANGED` for a function this card never touched.
*Measured*: currently **latent, not biting** —
`git diff --name-only c068667ad origin/develop -- internal/web/schemaform.go internal/web/handlers.go internal/settings/sectionapply.go`
returns empty output, so none of the three files has moved on develop since the merge-base. The
defect is in the instrument, not yet in the reading.
*Required fix*: set the extraction base to `git merge-base origin/develop HEAD` (capture it once,
name it something other than `$BASE` to keep the two semantics visibly distinct), or pin a commit
SHA per §4. Keep three-dot for AC-WCP-011.

**D2-3 — AC-WCP-005 as written admits a `0 == 0` pass**
`acceptance.md:65-71` — Severity: **minor** — Class: **optional**
The criterion asserts the two counts are equal and says nothing about either being non-zero, so a
mirrored field that renders in *neither* the page nor its owning panel satisfies it. The helper the
AC points at does not have this hole — `assertPanelFields` (`tab_layout_test.go:90-103`) errors with
"panel %q missing control %q" before comparing — so an implementer following the named shape gets
the presence check for free. REQ-WCP-004's other carrier (AC-WCP-009) also asserts the controls still
render. This is a text/instrument mismatch rather than a live gap, which is why it is optional; note
that AC-WCP-003 states its own non-zero guard explicitly and this one does not.
*Required fix*: one clause — "**And** each mirrored field's owning-panel count is `> 0`" — matching
what the cited helper already does.

**D2-4 — AC-WCP-003's "mirror row count > 0" names no countable token**
`acceptance.md:49-51` — Severity: **minor** — Class: **optional**
The non-vacuity guard is stated as a count of "mirror rows in the region" without naming the string
that is counted. `codexMirrorRow` is a Go component name and will not appear in the rendered HTML, so
an implementer must invent a DOM marker; one who picks a page-wide token instead of a region-local
one gets a guard that passes on an empty region — the failure the guard exists to prevent.
*Required fix*: name the marker (e.g. a `data-mirror-row` attribute emitted by `codexMirrorRow`) and
assert its count in the region.

**D2-5 — AC-WCP-013's "shows no field count (zero)" is two different assertions**
`acceptance.md:216-220` — Severity: **minor** — Class: **optional**
The rail renders its count unconditionally: `internal/web/shell.templ:156` is
`<span class="count">{ itoa(t.Fields) }</span>`. So a `Fields` of 0 renders the literal `0`; there is
no path on which the element is absent. An implementer reading "shows no field count" as
*element-absent* writes a test that cannot pass; one reading the parenthetical "(zero)" as
*renders 0* passes. Both readings are available in one sentence.
*Required fix*: say which — "the rail's `count` span for `codex` renders `0`" — and drop the
absent-element reading, or state explicitly that changing the template to omit a zero count is out of
scope.

**D2-6 — REQ-WCP-012 carries implementation detail into the requirement layer**
`spec.md:210-213` — Severity: **minor** — Class: **optional**
The requirement names a Go construct (`an explicit case "codex" in settingsTabFieldNames`), a source
file, and a line range — HOW, where the requirement layer wants WHAT. Checklist item RQ-4. I record
it as optional and with a caveat, because the author's rationale is defensible: §C.2 argues the whole
behavioural content of REQ-WCP-012 is "zero **by decision** rather than by fallback", and that
distinction has no expression except by naming the mechanism, since both branches produce the same
rendered output (verified — `schemaform.go:283-290`). A requirement whose subject genuinely is the
mechanism is a legitimate edge of RQ-4, not a clear violation. Raised so the orchestrator can decide,
not routed as a fix.

---

## Regression Check (iteration 1 defects)

| # | Iteration 1 defect | Status | Evidence |
|---|---|---|---|
| D1 | AC-WCP-005 RED on the pre-change tree; false premise | **RESOLVED** | Rewritten onto the per-panel axis (`acceptance.md:65-83`); greenness claim removed and swept for; `boolSegment` mechanism corrected in `spec.md §B.1:111-121`. Residual D2-3 is a different, minor point |
| D2 | AC-WCP-011 passes on a broken base ref | **RESOLVED** | Three separately-observed steps (`acceptance.md:170-178`); the trap is closed by step 1 and I could not defeat the new form on that axis |
| D3 | AC-WCP-008 satisfiable without the panel rendering | **RESOLVED** | Scoped to `panelHTML(t, html, "codex")`; whole-body arm repurposed to prove the MCP surface undisturbed (`acceptance.md:115-130`); the `panelHTML` fallback the scoping depends on is verified real |
| D4 | "≥12" floor coupled to an open decision | **RESOLVED** | Decision closed in the requirement layer (`spec.md §C.1:215-233`); floor deleted; exception asserted separately by AC-WCP-014 |
| D5 | REQ-WCP-009 not covered by the AC it maps to | **RESOLVED** | AC-WCP-006's registry-sweep arm is an independent oracle; §D.3:294-297 names it as the non-circular carrier; verified the arm reaches all 11 fields |
| D6 | Rail-count coupling named nowhere | **RESOLVED** — and my framing was wrong | Coupling named in three places + REQ-WCP-012 + AC-WCP-013 + MU-6. The *rail == row count* framing is corrected above; the *unnamed coupling* half is genuinely closed |
| D7 | AC-WCP-012's command cannot decide its claim | **PARTIALLY RESOLVED → treated as UNRESOLVED** | Mechanism replaced correctly and works for `parseSchemaForm` and `ApplySchemaEdits`; reads nothing and reports `IDENTICAL` for `handleSave` (D2-1). AC-WCP-009's `--stat` check is genuinely gone, replaced by the same extraction |
| D8 | MU-1/MU-2 rationale asserted a discrimination that does not hold | **RESOLVED** | `acceptance.md:269-275` states the redundancy plainly and says the split is retained deliberately as insurance against a narrower-than-specified guard |
| D9 | `name=` substring sweep + unstated placement dependency | **RESOLVED** | `name="` with the stated reason; control-tag sweep added; the not-last dependency is now AC-WCP-002 with its own preamble |
| D10 | Plan M2's icon step unactionable | **RESOLVED** | `plan.md:78-83` says to hardcode an `iconAt` case inside `fieldsetCodex`, explains why it cannot land anywhere else, and names three cases that exist |

No stagnation: no defect appears unchanged across both iterations. Nine of ten are fully resolved and
the tenth changed shape substantially. The revision is competent work.

---

## Recommendation

**FAIL**, on one blocking defect and one supporting it. The aggregate (0.84) clears Tier M's 0.80;
the verdict is driven by the regression clause on D7, for the reason set out above.

Routing, in order:

1. **D2-1** — change the extract anchor so it admits a method receiver, and add the non-empty-body
   self-check to §D.4. This is a single-line edit to `acceptance.md` plus one closure-gate bullet.
2. **D2-2** — give AC-WCP-012 a merge-base-derived extraction base, distinct from AC-WCP-011's
   three-dot `$BASE`.
3. D2-3, D2-4, D2-5, D2-6 are **optional** — the orchestrator's call, and none of them justifies a
   revision on its own.

**On the ceiling.** This is iteration 2 of 2 for Tier M, so the contract escalates to the operator
with PASS-with-debt / scope-reduction / explicit override. Of those:

- **Scope reduction is not indicated.** The SPEC is not too large; its scope was never the problem in
  either iteration, and the score moved 0.69 → 0.84 on revision alone.
- **PASS-with-debt is the one I would argue against.** The debt would be a silent vacuous pass in the
  criterion guarding the save path — the failure this card's entire read-only design exists to
  prevent — and silent debt is the kind that is never repaid, because nothing goes red to remind
  anyone.
- **What I recommend**: an operator override authorising a third pass scoped to the D2-1 + D2-2
  delta, confirmed by a re-audit of that delta only. The fix is two edits to one document; the cost
  of the extra pass is far below the cost of the escalation ceremony, and well below the cost of
  carrying the debt.

Everything else in this SPEC is ready. §B's re-measured premises, §C.1's closed decision, §C.2's
rail reasoning, the panel-scoping preamble, and the mutant table are all sound work that the fix
should not disturb.

---

## Residual risk

- **D2-1 rests on a deduction, not an execution.** I measured the anchor match count (0) and read
  the awk program; I did not run the loop, because the worktree guard refused the compound `git show`
  form. If awk behaves other than I expect here, the finding is wrong — the cheapest disproof is to
  run the loop and show a non-empty `handleSave` body.
- **The 11-name population and the `AllFields` reachability were established by reading the call
  chain, not by running it.** They agree with iteration 1's runtime measurement, but a build-tag or
  conditional I did not notice could change the set.
- **I ran no test and no build in this tree.** A pre-existing RED in `internal/web` would change the
  meaning of plan §C's baseline step, and I still cannot rule it out — unchanged from iteration 1.
- **The artifacts are uncommitted.** Every line citation above resolves against the working tree at
  HEAD `bd64824ff`; if the files are edited before they are committed, the citations drift.
- **The tree is left unmodified apart from this report.** I wrote no test files and deleted nothing.
