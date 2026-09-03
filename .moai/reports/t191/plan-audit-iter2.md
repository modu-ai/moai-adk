# SPEC Review Report: SPEC-PROJECT-CONTINUATION-KEY-001 (card t191)

Iteration: 2/2 (Tier M ceiling — `.moai/config/sections/harness.yaml:77` `plan_audit_tier_ceilings.M: 2`; there is no iteration 3 for this tier)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.80** (Tier M threshold 0.80 per `spec-workflow.md:141`) — up from 0.78

Reasoning context ignored per M1 Context Isolation. Audited from `spec.md`, `plan.md`,
`acceptance.md` (Tier M input contract) plus `progress.md` and the source tree, at worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t191`, branch `WT-project-continuation`,
HEAD `3b863da7b`, working tree clean before this report was written.

**PASS-WITH-DEBT clears by the narrowest possible margin and is not a clean bill.** Four blocking
defects are enumerated in §Defects Found. Because the Tier M iteration ceiling is exhausted, they
route as a **scoped delta fix before Implementation Kickoff Approval**, not as another audit round.

The revision is substantively good. All three iteration-1 blocking defects are genuinely resolved
rather than argued away, and the author was **right to reject my D5** and **right to correct my D2
error count** — both concessions are recorded below. The new blocking defects are of a different
and milder class than iteration 1's: three are places where an AC names evidence that cannot
actually fire, and one is an internal contradiction the author created by updating `REQ-PCK-006`
and `plan.md` M1 without updating `REQ-PCK-007` and `AC-PCK-007` alongside them.

---

## Regression Check — iteration-1 defects

| # | Iteration-1 defect | Status | Evidence |
|---|---|---|---|
| D1 | `pipeline` behaviourally indistinguishable from `card`; new option text not obliged to carry the kickoff clause | **RESOLVED** | `REQ-PCK-006` (`spec.md:84`) rewritten to define `pipeline` positively by carry distance and to require the kickoff clause; `AC-PCK-006` gains three conjuncts; `spec.md` §3 D1.1 states the delta against the shipped `card` text. The delta is real — see Hunt 1. **Introduced iter2-D1** (REQ-PCK-007 not updated with it). |
| D2 | Wizard-localization gap factually false; blessed fallback breaks an existing test | **RESOLVED in full** | `spec.md` §5 entry withdrawn and replaced with the measured correction; `plan.md` §B item 1 marked `[WITHDRAWN]` and redirected to the `audit_model` precedent; `acceptance.md` §D.5 folding branch struck; `AC-PCK-010` rebuilt as three conjuncts naming `TestWizardQuestionTranslationCompleteness` as enforcer and forbidding an `optionTranslationExemptIDs` entry. **My error-count sub-claim was wrong — conceded below.** |
| D3 | `AC-PCK-008`'s blocking conjunct had no reachable red | **RESOLVED** | Replaced with the per-branch kickoff-clause count against the measured pre-change value `1`; diff-stat re-filed as non-blocking `AC-PCK-014` and explicitly struck from `REQ-PCK-008`'s verification path (`acceptance.md` §D.2). **Introduced iter2-D4** (minor). |
| D4 | `REQ-PCK-011` "descriptions" resolved two ways | **RESOLVED** | Resolved toward the stronger reading: `REQ-PCK-011` now names a per-option description explicitly, `plan.md` M5 wraps the row in `withOptionDesc`, `AC-PCK-011` raises 20 → 32 entries. **Introduced iter2-D2** (the AC's stated red mechanism is wrong). |
| D5 | Inventory row cited at 2922-2924, "actually 2921-2923" | **REJECTED — correctly. My defect, not the author's.** | `grep -n "workflow.worktree.auto_cleanup" internal/config/testdata/shipped_key_inventory.yaml` → `2922:- path: "workflow.worktree.auto_cleanup"`. The row spans 2922-2924 as v0.1.0 stated. My iteration-1 figure came from counting positions inside a `sed -n '2918,2926p'` window instead of reading absolute line numbers. The citation was correct throughout and the substance was never in dispute. |
| D6 | `REQ-PCK-004` silent on the two P1-added options under `none` | **RESOLVED** | `REQ-PCK-004` (`spec.md:82`) now enumerates the four pre-P1 options inline, states the omission of `Create SPEC later` with its reason, and states the four-option cap is met without `Other` routing; `AC-PCK-004` asserts all three. |

**No stagnation.** No defect appears unchanged across both iterations. Five resolved, one correctly
rejected as auditor error.

### Concession — my iteration-1 D2 error count was wrong

I claimed the withdrawn folding fallback fails `TestWizardQuestionTranslationCompleteness` with
**9** errors; I flagged it as derived from loop bounds rather than observed. The author says **3**.
**The author is right**, and the control flow settles it without executing anything:

```
$ awk 'NR>=110 && NR<=134 {print NR": "$0}' internal/cli/wizard/translations_completeness_test.go
117: 			if q.Type != QuestionTypeSelect || optionTranslationExemptIDs[q.ID] {
118: 				continue
119: 			}
120: 			if len(trans.Options) != len(q.Options) {
121: 				t.Errorf("locale %q: question %q has %d option translations, want %d",
122: 					locale, q.ID, len(trans.Options), len(q.Options))
123: 				continue
124: 			}
125: 			for j, opt := range trans.Options {
```

Under the folding fallback `trans.Options` is empty and `q.Options` has length 3, so `:120` is
true, one `t.Errorf` fires at `:121`, and `:123` `continue`s **before** the per-option loop at
`:125` is ever entered. With `localizableLocales = []string{"ko","ja","zh"}`
(`translations_completeness_test.go:7`) that is exactly **one error per locale, three in total**.
My 9 assumed the loop ran; it does not.

Two precisions worth keeping, because the count is scenario-dependent and the author's revision
now states the 3 figure as though it were unconditional:

- **3** is the count for the *pure folding fallback* (no `Options` slice at all) — the scenario
  named in both documents.
- **9** is the count for a *different* scenario: an `Options` slice of the correct length 3 whose
  `Desc` fields are empty. Then `:120` is false, the loop runs, and `:129-130` fires three times
  per locale. That scenario is a plausible half-implementation of M4 and is not covered by the
  "3, not 9" statement.

**Neither count has been executed** — the question does not yet exist, so the test cannot be run
against it. Both figures are read from control flow. The count does not affect the conclusion
either party reached: the fallback is withdrawn, which the author accepted in full.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-PCK-001` … `REQ-PCK-012`, sequential, no gap, no
  duplicate, uniform 3-digit padding: `grep -o "REQ-PCK-[0-9]\{3\}" spec.md | sort -u` returns
  exactly `001 002 003 004 005 006 007 008 009 010 011 012`. The one new requirement extends the
  sequence rather than inserting into it.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`spec.md`
  §2), never against the ACs. The new `REQ-PCK-012` ("The option-label keys shall use the `.opt.`
  prefix …") is Ubiquitous and matches "The `<subject>` shall `<response>`". The two rewritten
  requirements keep their patterns: `REQ-PCK-005` and `REQ-PCK-006` remain State-driven ("While
  the resolved value is `<token>`, the Phase 14 workflow shall …"); `REQ-PCK-004` remains
  State-driven despite its added enumeration. No informal "should"/"may" in normative text.
- **[PASS] MP-3 YAML frontmatter validity** — 12 canonical fields present, `head -20 … | grep -cE`
  over the canonical field names returns `12`; `version: "0.2.0"` remains a quoted semver string;
  `created`/`updated` remain ISO `2026-09-02`. No rejected snake_case alias.
- **[N/A] MP-4 Section 22 language neutrality** — unchanged from iteration 1: the SPEC is scoped
  to this repository's own `internal/**` and its distributed template and names no multi-language
  tooling matrix. N/A auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — the referenced-SPEC set is unchanged
  (`SPEC-SYNC-STRATEGY-KEY-001`, `SPEC-TODO-ENABLE-FLAG-001`, `SPEC-V3R3-PROJECT-HARNESS-001`);
  all three exist and all three carry `status: completed`, none in
  `{retired, superseded, archived}`. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → `0`. Auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn "NEEDS CLARIFICATION" .moai/specs/SPEC-PROJECT-CONTINUATION-KEY-001/`
  → no match, `rc=1`. Note that iter2-D5 concerns this marker convention as a *property the
  shipped `pipeline` prose must respect*, not as an open marker in this SPEC's own artifacts.

All seven clear.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.68 | between 0.50 and 0.75 | `REQ-PCK-007` (`spec.md:85`) directly contradicts `REQ-PCK-006` (`spec.md:84`) — iter2-D1. `REQ-PCK-006` omits the clarification-marker precondition on gate emission — iter2-D5. The other ten requirements are crisp, and `pipeline` itself is now unambiguous, which is a large gain over v0.1.0. A contradiction is worse than the "minor ambiguity" the 0.75 band describes, but ten of twelve requirements are clean. |
| Completeness | 0.90 | 0.75-1.0 | All required sections present; three `### Out of Scope — <topic>` H3s retained. §5 Gaps materially improved — the false entry withdrawn with its measurement, and two new honest entries added (settings write layer untraced, `origin/develop` advanced 35 commits). The D5 rejection is argued in-document with its evidence rather than silently ignored. |
| Testability | 0.72 | between 0.50 and 0.75 | The `AC-PCK-005` / `AC-PCK-006` differential pair is a genuine advance (Hunt 2). Against it: `AC-PCK-011` and `AC-PCK-015` both name enforcing tests that cannot fire on the omission they describe (iter2-D2, iter2-D3), `AC-PCK-014` has no red by its own admission, and `acceptance.md:3`'s blanket claim is therefore false. More than one AC is imprecise, which puts it below the 0.75 band. |
| Traceability | 1.00 | 1.0 | `acceptance.md:10-26` maps all 15 ACs; every `REQ-PCK-001..012` appears in at least one **blocking** AC row — verified requirement by requirement. `AC-PCK-014`'s demotion is declared in §D.1 and §D.2 rather than left implicit, and `REQ-PCK-008`'s carrier is explicitly reassigned to `AC-PCK-008`. |

Aggregate: harmonic mean 4 / (1/0.68 + 1/0.90 + 1/0.72 + 1/1.00) = **0.8047 → 0.80**.

The score clears the Tier M threshold of 0.80 at the boundary. I report it as PASS-WITH-DEBT
rather than PASS because four blocking defects remain and the iteration ceiling is exhausted, so
there is no further audit round in which they would otherwise be caught.

---

## Hunt 1 — the carry-distance definition, attacked

**The delta is real this time. It does not collapse back into wording.** But the requirement set
did not fully absorb it, and two holes remain.

### The decisive measurement: `/moai plan` does not emit the gate

The whole claim rests on there being a gap between "`/moai plan` returns" and "the kickoff gate is
emitted". If `/moai plan` already ended by emitting the gate, `pipeline` would describe the status
quo and the synonym defect would simply have relocated. I measured it:

```
$ grep -n "Implementation Kickoff Approval" .claude/skills/moai/workflows/plan.md
53:**[NEEDS CLARIFICATION: <topic>]** markers identify unresolved questions that MUST be settled before Implementation Kickoff Approval (plan→run HUMAN GATE).
73:5. Implementation Kickoff Approval proceeds only after all clarifications are resolved
```

Both occurrences are *references to a downstream gate*, not an instruction to emit one. The
workflow's own terminal signal confirms the handoff: "This signal marks the plan artifacts as
finalized and enables the Plan Audit Gate at `/moai run` Phase 1." So the plan workflow ends at
artifact finalization and the gate belongs to the plan→run boundary that `/moai run` crosses.

The gap `pipeline` fills therefore exists. `card`'s shipped option terminates at `/moai plan`
(`doc-generation.md:350` — its run-phase sentence is a *disclaimer* about a later step, exactly as
`spec.md` §3 D1.1 argues), and nothing carries the session further. Under `pipeline` the session
continues to gate emission. These are different orchestrator action sequences after the same
operator answer. **`REQ-PCK-006` is no longer a synonym of `REQ-PCK-005`.**

### Does "emit the gate" become "and proceed once emitted"?

**No, and this is properly bounded.** `AC-PCK-006` conjunct 3 requires the row to "instruct the
orchestrator to emit the gate and stop for the operator's answer" and to contain "no instruction
to select, answer, pre-fill, or default that gate". `REQ-PCK-008` is unchanged and still forbids
any value being read as pre-authorizing run-phase entry. The author's "asked more, never fewer"
argument survives inspection: the operator is asked one additional question and answers it.

### Does carrying further pre-authorize anything by momentum?

**Not mechanically, and I looked for a vector rather than accepting the shape of the argument.**
The gate is an `AskUserQuestion` the operator answers; nothing in this SPEC touches which option
carries the recommendation on it, and `askuser-protocol.md` § Recommendation Placement Principles
makes that placement adaptive to observed majority rather than fixed, so there is no static
recommendation for `pipeline` to have tilted. The progression-mode axis is likewise untouched
(`orchestration-mode-selection.md:16`: "the axis selects only what happens after the gate
passes"). I found no mechanical momentum vector.

**But there is a non-momentum ordering hole, and it is real.** `plan.md:73` binds:
"Implementation Kickoff Approval proceeds only after all clarifications are resolved", and
`plan.md:53` says `[NEEDS CLARIFICATION]` markers "MUST be settled before Implementation Kickoff
Approval". `REQ-PCK-006` instructs the session to carry through to gate emission and says nothing
about this precondition. An orchestrator following the `pipeline` row literally, on a plan phase
that emitted clarification markers, would emit the gate with markers open — breaching a [HARD]
ordering clause. This is not a gate bypass (the operator still answers), but it is the gate being
asked before it is allowed to be asked. Filed as iter2-D5, blocking.

### Is the delta observable by the criterion that claims to measure it?

Yes — see Hunt 2. This is the strongest part of the revision.

### One honest qualification the SPEC does not make

The gap between `card` and `pipeline` is widened from **both** ends. `REQ-PCK-005` now adds "as
far as `/moai plan` and no further, leaving run-phase entry to a separately-initiated operator
action" — a constraint P1 never stated. P1's option text terminates at `/moai plan` but is silent
about what may follow, so `REQ-PCK-005` pins something P1 left open, and `REQ-PCK-002` claims
`card` "shall reproduce the behaviour PR #1601 established". The pin is faithful to the shipped
text and I do not think it is wrong, but the SPEC presents the delta as `pipeline` reaching
further when it is partly `card` being held back. Filed as iter2-D6, non-blocking.

---

## Hunt 2 — reachable-red re-probe on the new and changed criteria

### `AC-PCK-006` conjunct 1 — reachable red, established, and stronger than the AC claims

The conjunct reads: "A row whose terminal instruction is `/moai plan` — i.e. a `card`-behaving
implementation — fails this conjunct." Probed the same way I exposed `AC-PCK-008` in iteration 1
— by asking what implementation makes it go red rather than by reading its intent.

**It has a reachable red, and the mechanism is a differential pair the author did not name.**
`AC-PCK-005` now requires the `card` row's "terminal instruction is `/moai plan` with **no**
instruction to proceed to run-phase or to emit the Implementation Kickoff Approval gate".
`AC-PCK-006` conjunct 1 requires the `pipeline` row to continue past `/moai plan` to gate
emission. **These two are mutually exclusive over a single row text.** An implementer who ships
`pipeline` as a wording change writes two rows that say the same thing, and whichever thing they
say, one of the two criteria goes red:

- Both rows terminate at `/moai plan` → `AC-PCK-006` conjunct 1 red.
- Both rows carry to gate emission → `AC-PCK-005` red.

There is no row text satisfying both. That closes the synonym hole mechanically, which is exactly
what iteration-1 D1 asked for.

**The qualification**: the conjunct is decided by *reading prose*, not by a command exit code. It
is decidable without knowing the author's intent — one reads the row and observes what its last
imperative is — so it is not the "judgeable only by reading intent" failure I was asked to test
for. But `acceptance.md:3` claims every criterion "names the command that decides it", and this
one names none. `§D.3` already declares `REQ-PCK-004/005/006` as indirect prose verification,
which covers the substance; the header overstates. Rolled into iter2-D4.

### `AC-PCK-015` — red is reachable, but the AC names tests that cannot fire on this field

The `.opt.` guard exists exactly as the author describes. `internal/web/assets/app.js:273`:

```js
      var str = (key.indexOf(".opt.") >= 0 ? enDict : dict)[key];
```

and its tripwires:

```
$ awk 'NR>=132 && NR<=140 {print NR": "$0}' internal/web/audit_option_desc_test.go
135: func TestOptionLabelsStayEnglishStillGuarded(t *testing.T) {
136: 	js := readEmbeddedAsset(t, "app.js")
137: 	if !strings.Contains(js, `".opt."`) {
138: 		t.Fatal(`app.js lost the ".opt." guard — enum option labels would follow the active locale, reversing the G1-2 decision this card upheld`)
```

**But the AC's cited red for its second conjunct is wrong.** It cites `audit_option_desc_test.go:78`
as failing "if a description key contains `.opt.`". That test iterates `auditOptionDescFields()`,
which is a hardcoded map of exactly four audit fields:

```
$ grep -n "func auditOptionDescFields" -A 8 internal/web/audit_option_desc_test.go
23-	return map[string]string{
24-		"workflow.audit.model":        "f.workflow.audit.model.option.",
25-		"workflow.audit.gates.claude": "f.workflow.audit.gate.option.",
26-		"workflow.audit.gates.codex":  "f.workflow.audit.gate.option.",
27-		"workflow.audit.gates.glm":    "f.workflow.audit.gate.option.",
```

`workflow.project.continuation` is not in that map and never will be, so the named test cannot
fire on this field. The test that **would** catch it is the all-sections sweep the AC does not
name — `internal/web/option_desc_test.go:50-58` `TestEveryOptionDescKeyAvoidsOptGuard`, driven by
`allOptionDescFields` which walks `settings.AllSections()` (`:24-25`) and carries its own
non-vacuity floor (`:36-44`). So the red **is** reachable; the AC points at the wrong evidence.
Filed as iter2-D3.

Conjunct 1 (labels use the `.opt.` prefix) has **no** mechanical red: the prefix is a literal
string argument to `closedSeam`, and a wrong prefix produces locale-following labels that no test
forbids for this field. It is decided by reading the schema row.

### `AC-PCK-011`'s raised bar — the `withOptionDesc` half has no mechanical red at all

The 32-entry count is well-founded for the keys that *are* declared:
`internal/web/option_desc_test.go:72` requires `strings.Count(dict, key) != 4` to fail, so each
`.option.` key must appear in exactly four locale maps.

**But the AC's claim that a bare `closedSeam(...)` "fails the count" is wrong.** If the
implementer omits the `withOptionDesc` wrapper, the field carries no `OptionDesc` at all, so
`allOptionDescFields` skips it (`option_desc_test.go:27` — `if opt.OptionDesc != ""`) and every
description test passes vacuously on this field. And the two coverage tests the AC names compare
the locale maps **to each other**, not the schema to the dictionary:

```
$ awk 'NR>=408 && NR<=418 {print NR": "$0}' internal/web/i18n_governance_test.go
408: func TestI18nKeyCoverageForward(t *testing.T) {
409: 	cat := shippedI18nCatalogue(t)
410: 	en := cat["en"]
411: 	for _, loc := range i18nNonEnLocales {
412: 		for k := range en {
413: 			if _, ok := cat[loc][k]; !ok {
```

A bare `closedSeam` declares five keys, all five get written to all four maps, and coverage
passes. **A `REQ-PCK-011` under-delivery therefore ships silently green.** Filed as iter2-D2.

The repository's own convention shows the fix: `TestCrossSessionI18nKeysInAllLocales`
(`internal/web/crosssession_test.go:100`) and `TestFeedbackAutoSubmitI18nKeysInAllLocales`
(`internal/web/feedback_panel_test.go:33`) are per-SPEC tests added by the SPEC that introduced
the keys. This SPEC needs one too.

### `AC-PCK-010` — red reachable, correctly cited

Conjunct 3 names `TestWizardQuestionTranslationCompleteness` and forbids an
`optionTranslationExemptIDs` entry. Both are accurate against source (`:95-134`, `:13-15`). No
defect.

---

## Hunt 3 — `AC-PCK-014`, the self-declared no-red criterion

**Honest, not a hiding place — but misfiled, and it makes the document contradict itself.**

Honest, because it is labelled in three places rather than buried: its own body says it "carries
no reachable red and MUST NOT be presented as the verification of `REQ-PCK-008`", `§D.1` repeats
it, and `§D.2` explicitly strikes it from `REQ-PCK-008`'s verification path. That is the opposite
of hiding — an auditor cannot mistake it for verification. The author also corrected its base
from `origin/develop` to the merge-base, which is right given the 35-commit advance.

**But a criterion that cannot fail decides nothing, and an acceptance set is the set of things
that decide closure.** `§D.4`'s closure gate reads "All thirteen blocking ACs pass" — `AC-PCK-014`
is not among them, so it gates nothing at all. Its actual content is a *constraint on the
implementer* ("do not touch these three files"), and the plan already has two homes for exactly
that: `plan.md` §D Constraints and §G Anti-patterns, where a non-falsifiable prohibition is normal
and costs nothing.

Keeping it in `acceptance.md` has a measurable cost: `acceptance.md:3` now claims every criterion
"has a red reachable by an implementation the plan describes", and `AC-PCK-014`'s own body denies
it eleven lines later. Filed as iter2-D4.

---

## Hunt 4 — `REQ-PCK-012` / `AC-PCK-015` scope creep

**The guard exists exactly as described** (quoted under Hunt 2), and
`internal/web/console_ux_fix_test.go:87-88` is the second tripwire the author cites:

```
87: 	if !strings.Contains(fn, `".opt."`) {
88: 		t.Error(`applyI18n has no ".opt." guard — enum option labels would follow the active locale (G1-2)`)
```

**Judgment: mostly in scope, with one half that is not.**

*In scope*: `REQ-PCK-011` obliges this SPEC to ship per-option description keys, so this SPEC must
choose their names, and that choice is load-bearing — name them
`f.workflow.project.continuation.opt.desc.<token>` and `app.js:273` freezes them to English in
every locale, silently defeating `REQ-PCK-011`. A requirement constraining a naming decision this
SPEC is forced to make is not a second concern; it is the first concern's fine print. The
`.option.` choice is correct, and the author's reasoning for why it escapes the substring test
("the character after `opt` is `i`") is right.

*Out of scope*: `REQ-PCK-012`'s trailing clause "the guard shall not be modified", and
`AC-PCK-015`'s third conjunct "`applyI18n`'s `.opt.` guard in `app.js` is unmodified". `app.js`
appears in no M1-M6 file list, the prohibition is already enforced by two existing tripwires that
this SPEC does not own, and it protects a prior decision (G1-2) from a hypothetical implementer.
This is the same shape as `AC-PCK-014` — a no-red guard on a file nothing touches — and it is
mildly duplicative rather than harmful. I am **not** filing it as a defect: unlike `AC-PCK-014` it
rides on a conjunct that does have a red, it costs one clause, and the naming hazard it sits
beside is genuine. Recorded as an observation only.

---

## Hunt 5 — the author's second leg of the progression-mode rejection

**The evidence is accurate. The inference is slightly overstated, and the conclusion survives —
for a better reason than the author gave.**

The measurement reproduces:

```
$ grep -n "Recommended" .claude/skills/moai/workflows/goal.md
rc=1
```

No match. So `goal.md` genuinely specifies no recommended option on the progression-mode axis, and
the author's "no baseline to pin `card` against" is a fair reading of that file.

**The overstatement**: a recommendation position does exist in the general rule —
`askuser-protocol.md` § Socratic Interview Structure item 3 makes the first option's
`(Recommended)` / `(권장)` suffix mandatory on every `AskUserQuestion`, so the progression-mode
question has a recommended option even though `goal.md` does not say which. "No baseline" is
therefore not quite right; "the baseline is not fixed in `goal.md`" is.

**Why the conclusion nonetheless holds, and holds more strongly**: `askuser-protocol.md`
§ Recommendation Placement Principles makes the recommended option the *observed-majority* choice
and adaptive to inferred proficiency — deliberately not a static value. A requirement cannot pin
`card` to a recommendation that is by design a moving, behaviour-derived target, and attempting it
would put this SPEC in conflict with that rule. So the rejection is sound; its stated ground
(unspecified) is weaker than its real ground (specified as adaptive, therefore unpinnable).

The first leg, which the coordinator verified independently, I did not re-derive.
`goal.md:112-113` reads verbatim: "The selected mode is persisted in goal state as
`progression_mode` (default `autonomous` when the user declines to choose)." Confirmed in passing
while reading the surrounding block; it makes the proposed reading describe the status quo, which
is the synonym defect again.

---

## Hunt 6 — the declared-unresolved items

**Declared honestly; none is being used to park a defect. And the item the coordinator cares most
about is not merely likely — it is settled, in the plan's favour.**

### The three-segment settings write path — RESOLVED by this audit

`spec.md` §5 calls this "likely but not measured". It is measurable, and it holds:

```
$ awk 'NR>=380 && NR<=391 {print NR": "$0}' internal/settings/schema_sections.go
380: 		withOptionDesc(closedSeam(SectionWorkflow, "workflow", "f.workflow.audit.model.opt.",
381: 			config.ValidAuditModels(), "", "", "workflow", "audit", "model"),
382: 			"f.workflow.audit.model.option."),
383: 		withOptionDesc(closedSeam(SectionWorkflow, "workflow", "f.workflow.audit.gate.opt.",
384: 			config.ValidAuditGates(), "", "", "workflow", "audit", "gates", "claude"),
```

A **three**-segment path (`workflow, audit, model`) and a **four**-segment path
(`workflow, audit, gates, claude`) already ship in the very section this SPEC writes into. The
mechanism is depth-agnostic end to end: `seamField` stores `Path: path` verbatim
(`schema_sections.go:95`); `sectionapply.go:52-54` feeds `f.Persist.Path` straight into
`yamlpatch.KeyEdit`; and `KeyEdit`'s own doc comment gives a five-segment example —

```
$ grep -n "type KeyEdit" -A 7 internal/settings/yamlpatch/yamlpatch.go
29-	// Path는 문서 루트 매핑부터의 키 경로다.
30-	// 예: ["workflow", "team", "role_profiles", "implementer", "model"].
31-	Path []string
```

— with `yamlpatch.go:93-94` iterating the whole path and `PatchFile` upserting missing mappings
(`:37-41`). Both the M4 (`yamlpatch.KeyEdit`) and M5 (`FieldDef` path) assertions are sound. The
risk that M5 fails on path depth does not exist.

### The remaining four

- **Zero tests executed** — declared, and true of my audit as well. The four tests both documents
  reason about were read in source; none was run. This is the largest shared gap and it is stated
  in both directions.
- **`make build` not run** — declared; `AC-PCK-012` is probed only by re-running its `cmp` matrix
  on the current tree, which is what iteration 1 did.
- **`context_folding` triage class unchecked** — declared, and correctly scoped: the SPEC cites it
  only as evidence that prose-read workflow keys exist, and makes no claim about its class.
- **#1600/#1601 CHANGELOG collision** — still carried from the card brief, still declared as such.
  It underwrites only the `plan.md` M6 instruction to add the CHANGELOG entry additively, which is
  good practice regardless of whether the specific collision is reproduced.

None of these is doing load-bearing work for a claim made elsewhere in the document. The one that
was — the wizard gap in v0.1.0 — is exactly the one that turned out to be false, and it has been
withdrawn.

---

## Defects Found

**iter2-D1 — `REQ-PCK-007` contradicts `REQ-PCK-006`, and `AC-PCK-007` will go red against the
plan's own M1 instruction** — `spec.md:85`, `spec.md:84`, `acceptance.md` §AC-PCK-007, `plan.md`
M1 — Severity: **critical** — Class: **blocking** — The revision changed what the key does but
not the requirement that says what the key may do. `REQ-PCK-007` still reads "the key shall change
**only** which option is recommended and how that option is worded" — the v0.1.0 presentation-only
formula. `REQ-PCK-006` now makes carry distance the delta, which is neither of those two things.
`plan.md` M1 was updated and now instructs the implementer to write "the key changes only which
option is recommended, **how far it carries**, and how it is worded". `AC-PCK-007` was **not**
updated and still asserts "the new sentence states that the key changes only which option is
recommended and how it is worded". So an implementer who follows `plan.md` M1 exactly writes a
sentence that fails `AC-PCK-007`, and one who satisfies `AC-PCK-007` writes a sentence that
contradicts `REQ-PCK-006`. The set is unsatisfiable as written.
**Required fix**: change `REQ-PCK-007` to "…the key shall change only which option is recommended,
how far that option carries the session, and how it is worded", and change `AC-PCK-007`'s Then
clause to assert that three-part sentence. Both edits are one line each and must land together.

**iter2-D2 — `AC-PCK-011`'s `withOptionDesc` requirement has no mechanical red, and the AC names
tests that cannot supply one** — `acceptance.md` §AC-PCK-011 "Reachable red" note — Severity:
**major** — Class: **blocking** — A bare `closedSeam(...)` without the `withOptionDesc` wrapper
produces a field with no `OptionDesc`, which `allOptionDescFields` skips at
`internal/web/option_desc_test.go:27`, so every per-option-description test passes vacuously on
this field. The two tests the AC names — `TestI18nKeyCoverageForward` / `Reverse`
(`i18n_governance_test.go:408,423`) — compare the four locale maps to each other, not the schema
to the dictionary, so the five keys a bare `closedSeam` does declare are present in all four maps
and coverage passes. A `REQ-PCK-011` under-delivery therefore ships silently green, which is the
same class of defect as iteration-1 D3.
**Required fix**: add to `plan.md` M5 a new test
`TestProjectContinuationI18nKeysInAllLocales` in `internal/web`, following
`crosssession_test.go:100` and `feedback_panel_test.go:33` — the repository's established
per-SPEC pattern — asserting all eight keys resolve in each of the four maps; and rewrite
`AC-PCK-011`'s "Reachable red" note to name that new test as the enforcer instead of the two
coverage tests.

**iter2-D3 — `AC-PCK-015` cites audit-scoped tests that cannot fire on this field** —
`acceptance.md` §AC-PCK-015 "Reachable red" note — Severity: **major** — Class: **blocking** —
The note cites `internal/web/audit_option_desc_test.go:78` as the test that fails "if a
description key contains `.opt.`". That test iterates `auditOptionDescFields()`, a hardcoded
four-entry map (`:23-28`) containing only `workflow.audit.model` and the three
`workflow.audit.gates.*` fields. `workflow.project.continuation` is not in it, so the cited test
is structurally incapable of firing on this SPEC's field. An implementer who names the
description keys `…continuation.opt.desc.<token>` and relies on the cited test to catch it gets a
green run and English-frozen descriptions in ko/ja/zh.
**Required fix**: replace the citation with `internal/web/option_desc_test.go:50-58`
`TestEveryOptionDescKeyAvoidsOptGuard`, the all-sections sweep driven by
`settings.AllSections()` (`:24-25`), which does cover the new field. Retain the
`audit_option_desc_test.go:137-138` citation — that one is correct, since it asserts on `app.js`
itself rather than on a field set.

**iter2-D5 — `REQ-PCK-006` omits the clarification-marker precondition on gate emission** —
`spec.md:84`, against `plan.md:53` and `plan.md:73` — Severity: **major** — Class: **blocking** —
`REQ-PCK-006` instructs the session to carry "past `/moai plan` to the emission of the
Implementation Kickoff Approval gate" with no precondition attached. The plan workflow binds one:
"[NEEDS CLARIFICATION: <topic>] markers identify unresolved questions that MUST be settled before
Implementation Kickoff Approval (plan→run HUMAN GATE)" (`plan.md:53`) and "Implementation Kickoff
Approval proceeds only after all clarifications are resolved" (`plan.md:73`). An orchestrator
following the `pipeline` row on a plan phase that emitted markers would emit the gate before it is
permitted to be emitted. This does not bypass the gate — the operator still answers — but it
breaches a [HARD] ordering clause, and it is precisely the kind of precondition an implementer
writing prose will not supply unprompted.
**Required fix**: extend `REQ-PCK-006` with "…and shall emit that gate only once the plan phase's
`[NEEDS CLARIFICATION]` markers are resolved, per `plan.md:53,73`; where markers remain open the
row shall stop at their resolution rather than at the gate." Add a fourth conjunct to
`AC-PCK-006` asserting the `pipeline` row states that precondition, and add the ordering to
`plan.md` M1's `pipeline` bullet.

**iter2-D4 — `acceptance.md:3` claims every criterion has a reachable red; `AC-PCK-014` denies it,
and `AC-PCK-006` conjunct 1 names no command** — `acceptance.md:3`, `acceptance.md` §AC-PCK-014 —
Severity: **minor** — Class: **non-blocking** — The header asserts "every criterion is binary,
names the command that decides it, and has a red reachable by an implementation the plan
describes". `AC-PCK-014`'s own body states it "carries no reachable red"; `AC-PCK-006` conjunct 1
and `AC-PCK-015` conjunct 1 are prose reads that name no command. The header's strengthening was
the right instinct — it is the promise iteration-1 D3 asked for — but it is now falsified inside
the same document.
**Required fix**: qualify the header to "every **blocking** criterion has a red reachable by an
implementation the plan describes; prose-verified criteria are marked in §D.3 and name the text
read rather than a command", and move `AC-PCK-014` out of `acceptance.md` into `plan.md` §D
Constraints, where a non-falsifiable prohibition is the normal form and gates nothing.

**iter2-D6 — the `card`/`pipeline` gap is widened partly by tightening `card` beyond P1** —
`spec.md:83` (`REQ-PCK-005`), against `spec.md:80` (`REQ-PCK-002`) — Severity: **minor** — Class:
**non-blocking** — `REQ-PCK-005` adds "as far as `/moai plan` and no further, leaving run-phase
entry to a separately-initiated operator action". P1's option text terminates at `/moai plan` but
says nothing about what may follow, so this pins a question P1 left open, while `REQ-PCK-002`
claims `card` "shall reproduce the behaviour PR #1601 established". The pin is faithful to the
shipped text and I judge it correct, but §3 D1.1 presents the delta as `pipeline` reaching further
when it is in part `card` being held back, and a reader comparing `REQ-PCK-002` against
`REQ-PCK-005` will notice the addition without being told it was deliberate.
**Required fix**: add one sentence to §3 D1.1: "`REQ-PCK-005`'s 'and no further' makes explicit a
boundary P1 left unstated — P1's option terminates at `/moai plan` — so the delta is partly a
clarification of `card` and not solely an extension of `pipeline`. This is a decision, not a
change to P1 behaviour."

Blocking: **4** (iter2-D1, D2, D3, D5). Non-blocking: **2** (iter2-D4, D6).

---

## What is right, briefly

- The carry-distance definition is a real fix to a real defect, and it is measured against the
  shipped `card` text rather than against the card brief. §3 D1.1's admission that "a third enum
  value whose complete observable effect is one string is not a value" is the correct reading of
  iteration-1 D1.
- The `AC-PCK-005` / `AC-PCK-006` differential pair closes the synonym hole mechanically. That is
  a better answer than the one iteration-1 asked for.
- §3 D1.3 rejects a proposed alternative **in writing, with its measurement**, rather than
  silently declining it — including the state-mechanism confirmation that keeps the rejection
  about the default's position rather than the mechanism's existence.
- The D5 rejection and the D2 error-count correction are both right, and both are argued from
  source rather than asserted. An author who pushes back correctly on an auditor is doing the job.
- §5 Gaps grew two honest entries under audit pressure (settings write layer, `origin/develop`
  advance) rather than shrinking.

---

## Gaps — what this audit did not observe

- **No test was executed.** `TestWizardQuestionTranslationCompleteness`,
  `TestEveryOptionDescKeyAvoidsOptGuard`, `TestEveryOptionDescKeyTranslatedInAllLocales`,
  `TestI18nKeyCoverageForward` / `Reverse`, `TestOptionLabelsStayEnglishStillGuarded`, and
  `TestShippedConfigKeysHaveReaders` were all read in source and their firing conditions traced by
  reading assertion bodies and loop scopes. None was run. Every "this test would/would not fire"
  statement above is a control-flow reading, not an observed run — including my concession on the
  3-versus-9 count.
- **The `AC-PCK-011` / `AC-PCK-015` red-mechanism findings rest on scope reads, not on a
  demonstration.** I established that `allOptionDescFields` skips fields with no `OptionDesc`
  (`option_desc_test.go:27`) and that `auditOptionDescFields` is a four-entry literal map
  (`:23-28`). I did not construct a `FieldDef` and run the suite against it.
- **`make build` was not run**, and `AC-PCK-012`'s `cmp` matrix was not re-run this iteration —
  iteration 1 measured it as `0/0/0/1` and nothing in `3b863da7b` touches those four files
  (`git show --stat` lists only the five SPEC/report artifacts).
- **The §1.1 RED-now baselines were not re-measured this iteration.** Iteration 1 reproduced all
  six at tree `2660bcd09`, and `3b863da7b` changes no source file, so they stand by attribution to
  that measurement rather than by fresh observation.
- **`internal/core/project/initializer_expansion.go` was read only around the
  `writeWorkflowTodoYAML` call site (`:50-71`).** I confirmed the yamlpatch rationale in its
  comments but did not read the `writeWorkflowTodoYAML` body itself, so the M4 precedent is
  verified at the mechanism level (`KeyEdit` depth, `PatchFile` upsert) rather than by reading the
  precedent function end to end.
- **`/moai run` Phase 1 was not read.** My Hunt 1 conclusion that the kickoff gate lives at the
  plan→run boundary rests on `plan.md`'s two references plus its terminal handoff signal and on
  `orchestration-mode-selection.md:18`; I did not open `run.md` to confirm where the gate is
  emitted from.
- **The `progression_mode` Go mechanism** (`internal/goal`, `internal/hook/handoff_inject.go:241`)
  cited in §3 D1.3 was not verified; the rejection does not depend on it, since it turns on the
  default's position rather than the mechanism's existence.
- **`git status --short` was not captured at report time**; the working tree contained this report
  file when written.

Nothing else was left unobserved.

---

## Recommendation

PASS-WITH-DEBT at 0.80 against a Tier M threshold of 0.80, all seven must-pass criteria clear,
four blocking defects outstanding. The Tier M ceiling is exhausted, so these route as a **scoped
delta fix before Implementation Kickoff Approval**, not as a third audit round. All four fixes are
localized edits to existing lines; none reopens a design question.

Order:

1. **iter2-D1 first** — it is the only unsatisfiable-as-written defect, and both edits are one
   line. Until it lands, `AC-PCK-007` and `plan.md` M1 instruct opposite things.
2. **iter2-D5 second** — it adds a precondition to `REQ-PCK-006` and a conjunct to `AC-PCK-006`.
   Doing it after D1 avoids touching the same requirement twice.
3. **iter2-D2 and iter2-D3 together** — both are "the AC names evidence that cannot fire", both
   are fixed in the acceptance file, and D2 additionally adds one test to `plan.md` M5.
4. **iter2-D4 and iter2-D6** — one-line edits, no dependencies.

iter2-D2 and iter2-D3 share a root cause worth naming for the author: when an AC cites an existing
test as its red, the citation must be checked against that test's **scope**, not only its name and
message. Both cited tests say exactly the right thing in their failure strings; both iterate a
field set that this SPEC's field will never enter. That is the same shape as iteration-1 D3 — a
criterion that reads as verification but cannot fire — relocated from the criterion's command to
the criterion's cited enforcer.
