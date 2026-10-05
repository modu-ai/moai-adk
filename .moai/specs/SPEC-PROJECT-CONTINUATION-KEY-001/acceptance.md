# SPEC-PROJECT-CONTINUATION-KEY-001 — Acceptance Criteria

Card **t191** · **v0.3.0** · 14 criteria (13 blocking) · every criterion is binary, and **every criterion — blocking and non-blocking alike — has a red reachable by an implementation the plan describes.** No criterion in this file is unfalsifiable; the one that was has been relocated to `plan.md` §D.
Criteria decided by a command name that command; **prose-verified criteria are marked in §D.3 and name the text read instead** — `AC-PCK-004/005/006` and `AC-PCK-014` conjunct 1 are read from the shipped instruction text, which is the artifact the orchestrator consumes.
Baseline tree `2660bcd09`.

**The `AC-PCK-005` / `AC-PCK-006` differential pair.** These two are the mechanical closure of the `pipeline`-is-a-synonym hole, and they work only as a pair: `AC-PCK-005` requires the `card` row's terminal instruction to be `/moai plan` with no run-phase continuation, and `AC-PCK-006` conjunct 1 requires the `pipeline` row to carry past `/moai plan` to gate emission. **No single row text satisfies both.** An implementer who ships `pipeline` as a wording change writes two rows saying the same thing, and whichever thing they say, one criterion goes red — both rows stopping at `/moai plan` reds `AC-PCK-006`; both rows carrying to the gate reds `AC-PCK-005`. Neither criterion carries that guarantee alone; removing or weakening either one reopens the hole.

---

## §D AC Matrix

| AC | REQ | Milestone | Severity |
|---|---|---|---|
| AC-PCK-001 | REQ-PCK-001, REQ-PCK-002 | M2 | blocking |
| AC-PCK-002 | REQ-PCK-002 | M2 | blocking |
| AC-PCK-003 | REQ-PCK-003 | M2 | blocking |
| AC-PCK-004 | REQ-PCK-004 | M1 | blocking |
| AC-PCK-005 | REQ-PCK-005 | M1 | blocking |
| AC-PCK-006 | REQ-PCK-006 | M1 | blocking |
| AC-PCK-007 | REQ-PCK-007 | M1 | blocking |
| AC-PCK-008 | REQ-PCK-008 | M1 | blocking |
| AC-PCK-009 | REQ-PCK-009 | M3 | blocking |
| AC-PCK-010 | REQ-PCK-010 | M4 | blocking |
| AC-PCK-011 | REQ-PCK-011 | M5 | blocking |
| AC-PCK-012 | REQ-PCK-009 (+ §3 D6 mirror parity) | M6 | blocking |
| AC-PCK-013 | REQ-PCK-003, REQ-PCK-004 | M1 | non-blocking |
| AC-PCK-014 | REQ-PCK-012 | M5 | blocking |

The v0.2.0 `AC-PCK-014` (a `git diff --stat` hygiene check over three gate-owning rule files) has been **removed from this file** and relocated to `plan.md` §D Constraints. It had no reachable red by its own admission, and `§D.4` gates on blocking criteria only, so it decided nothing here while falsifying the header's promise. A non-falsifiable prohibition is the normal form for a plan constraint. The former `AC-PCK-015` is renumbered `AC-PCK-014` to keep the sequence gapless.

---

### AC-PCK-001 — the closed set is exactly three tokens

**Given** the config package on the implementation branch,
**When** `go test ./internal/config/ -run TestValidProjectContinuations` runs a case asserting `config.ValidProjectContinuations()`,
**Then** the returned slice is exactly `["none", "card", "pipeline"]` and the test passes.

### AC-PCK-002 — absent resolves to card

**Given** a `workflow.yaml` carrying no `project:` block,
**When** the resolver is called on the loaded config,
**Then** it returns value `card` and an empty unmatched string, and the same holds for a nil `*Config` receiver.

### AC-PCK-003 — an unmatched value resolves to card and is reported

**Given** a `workflow.yaml` carrying `workflow.project.continuation: pipelien`,
**When** the resolver is called,
**Then** it returns value `card` **and** unmatched `pipelien`; the call does not error and does not panic.

### AC-PCK-004 — `none` issues no card and shows the pre-P1 recommended option

**Given** `doc-generation.md` on the implementation branch,
**When** `grep -n "none" -A 6` is read over the Step 4.1.5 resolution block and the Step 4.2 recommended-option table,
**Then** the `none` row instructs skipping issuance entirely; names `Create SPEC` — the pre-P1 wording measured from `e91def4ca` — as the recommended option; enumerates exactly the four pre-P1 options, **omitting `Create SPEC later`**; states that the four-option cap is met without routing any option to `Other`; and no `moai todo add` invocation is reachable on that path.

### AC-PCK-005 — `card` reproduces P1 and stops at `/moai plan`

**Given** the same file,
**When** the `card` row of the Step 4.2 table is read,
**Then** it names `Create the SPEC and start now` as the recommended option; its terminal instruction is `/moai plan` with **no** instruction to proceed to run-phase or to emit the Implementation Kickoff Approval gate; and Step 4.1.5's five standing-source steps (harness-spec read, empty-goal skip, duplicate-suppression read, single `[PROJECT] ` add, id reported) are present and unmodified from their P1 text.

### AC-PCK-006 — `pipeline` carries past `/moai plan`, and carries the gate clause with it

**Given** the same file,
**When** the `pipeline` row of the Step 4.2 table is read,
**Then** all four hold:

1. **The delta.** The row instructs the session to continue past `/moai plan` to **emitting** the Implementation Kickoff Approval gate in this same session. A row whose terminal instruction is `/moai plan` — i.e. a `card`-behaving implementation — fails this conjunct. This is the criterion that distinguishes `pipeline` from `card`; it is red for any implementation that ships `pipeline` as a wording change.
2. **The gate clause.** `grep -c "Implementation Kickoff Approval"` restricted to the `pipeline` row returns **≥ 1**, in the same terms as the `card` option.
3. **No answering.** The row instructs the orchestrator to emit the gate and stop for the operator's answer; it contains no instruction to select, answer, pre-fill, or default that gate.
4. **The clarification precondition.** The row states that the gate is emitted only once the plan phase's `[NEEDS CLARIFICATION]` markers are resolved, and that where markers remain open the row stops at their resolution rather than at the gate. A row that instructs carrying to gate emission unconditionally fails this conjunct.

Conjunct 3 replaces v0.1.0's untestable negative ("contains no instruction to …", which any silent row satisfied) with a positive assertion about what the row must say. Conjunct 4 exists because `plan.md:53` and `plan.md:73` bind a [HARD] ordering — "Implementation Kickoff Approval proceeds only after all clarifications are resolved" — that an implementer writing prose will not supply unprompted; without it the `pipeline` row would have the gate asked before it is permitted to be asked. That is not a bypass (the operator still answers) but it is an ordering breach.

**Conjunct 1 is decided by reading the row's terminal imperative, not by a command** — see §D.3 and the differential-pair note in the header. It is decidable without knowing the author's intent: one reads the row and observes what its last instruction is.

### AC-PCK-007 — the question survives every value

**Given** the same file,
**When** `grep -n "No branch is taken on the operator's behalf" doc-generation.md` and the new value-independence sentence are read,
**Then** both clauses are present, neither is scoped to a subset of values, and the new sentence states **both halves** of `REQ-PCK-007`:

1. **The invariant** — at every value the question is asked, the operator answers, and nothing is skipped, auto-answered, pre-filled, defaulted-on-no-answer, or bypassed.
2. **The three permitted changes** — which option is recommended, **how far that option carries the session**, and how it is worded.

**Why the second half names three things and not two.** v0.2.0 asserted the v0.1.0 presentation-only formula ("only which option is recommended and how it is worded") while `REQ-PCK-006` had already made carry distance the delta. The set was unsatisfiable: an implementer following `plan.md` M1 wrote a sentence failing this criterion, and one satisfying this criterion wrote a sentence contradicting `REQ-PCK-006`. The fix is to permit carry distance explicitly, **not** to weaken the invariant — half 1 is unchanged and gained two prohibited verbs.

### AC-PCK-008 — every run-phase-offering branch carries the kickoff clause

**Given** `doc-generation.md` on the implementation branch,
**When** the Step 4.2 recommended-option table is read for all three values,
**Then** each recommended-option text that offers run-phase entry carries the Implementation Kickoff Approval sentence verbatim, and

```bash
grep -c "Implementation Kickoff Approval" .claude/skills/moai/workflows/project/doc-generation.md
```

returns a count **no lower than the number of such options** — pre-change value `1`, measured on `2660bcd09`; expected `≥ 2` after M1, since `card` and `pipeline` both offer run-phase entry and `none` does not.

**Reachable red, established.** The v0.1.0 criterion was a `git diff --stat origin/develop...HEAD` over three gate-owning files; run verbatim at plan time it returned empty with `rc=0`, so it was already satisfied before the SPEC existed, and none of the three files appears in any M1-M6 file list. This replacement goes red on the leak it is meant to catch: M1 rewrites the Step 4.2 option block containing the sole occurrence, so an implementer who writes the `pipeline` row without the clause drops the count below the option count.

### AC-PCK-009 — the shipped key passes the anti-rot guard

**Given** the template `workflow.yaml` carrying `continuation: card` and the inventory carrying the new row,
**When** `go test ./internal/config/ -run TestShippedConfigKeysHaveReaders` runs,
**Then** the test passes, and the same test fails when the inventory row alone is removed (the mutant establishes the guard is not vacuous here).

### AC-PCK-010 — the wizard offers the closed set in four locales

**Given** the wizard package on the implementation branch,
**When** `go test ./internal/cli/wizard/...` runs,
**Then** all three hold:

1. The `project_continuation` question's `Type` is `QuestionTypeSelect`, its `Default` is `card`, and its option values equal `config.ValidProjectContinuations()`.
2. For each of `ko` / `ja` / `zh`, the `translations` entry carries a non-empty `Title`, a non-empty `Description`, and an `Options` slice of length **3**, every element having a non-empty `Label` **and** a non-empty `Desc` — the `audit_model` shape (`translations.go:137`).
3. **`TestWizardQuestionTranslationCompleteness` passes.** That existing test (`translations_completeness_test.go:95-134`) is the enforcer for conjunct 2, and `project_continuation` must NOT be added to `optionTranslationExemptIDs` to satisfy it.

**Reachable red.** `grep -rn "project_continuation" internal/cli/wizard/ | wc -l` → `0` today, so the question does not exist. Two distinct half-implementations both fail `TestWizardQuestionTranslationCompleteness`, with **different counts** — the count is scenario-dependent, so neither figure is the criterion:

| Half-implementation | Errors | Mechanism |
|---|---|---|
| No `Options` slice at all (the withdrawn folding fallback) | **3** — one per locale | `len(trans.Options)=0 != 3` at `:120`, `t.Errorf` at `:121`, `continue` at `:123` **before** the per-option loop at `:125` |
| `Options` slice of length 3 with empty `Desc` fields | **9** — three per locale | `:120` false, loop at `:125` runs, the empty-`Desc` branch fires per option |

Both are plausible outcomes of M4 and both are red, which is what the criterion needs. **Neither count has been executed** — the question does not yet exist, so the test cannot be run against it; both are control-flow readings of `:117-130`.

### AC-PCK-011 — the console field is closed-set and fully localized

**Given** the settings schema on the implementation branch,
**When** `go test ./internal/settings/... ./internal/web/...` runs,
**Then** a `FieldDef` exists for path `workflow → project → continuation` whose option values derive from `config.ValidProjectContinuations()` and which is wrapped in `withOptionDesc` (`schema_sections.go:143`), and `internal/web/assets/i18n.js` carries, in **each** of the four locale maps, all eight of: `f.workflow.project.continuation.title`, `.desc`, three `.opt.<token>` labels, and three `.option.<token>` descriptions — **8 keys × 4 locales = 32 entries, none empty**.

**Reachable red — supplied by a NEW test this SPEC must add, because no existing test can fire on the omission.**

The v0.2.0 red mechanism was wrong and is withdrawn. A bare `closedSeam(...)` without the `withOptionDesc` wrapper produces a field carrying **no** `OptionDesc`, and `allOptionDescFields` skips exactly those fields:

```
$ awk 'NR>=25 && NR<=31 {print NR": "$0}' internal/web/option_desc_test.go
27: 			for _, opt := range f.Options {
28: 				if opt.OptionDesc != "" {
```

So every per-option-description test passes **vacuously** on this field. And `TestI18nKeyCoverageForward` / `Reverse` (`i18n_governance_test.go:408,423`) compare the four locale maps **to each other**, not the schema to the dictionary — a bare `closedSeam` declares five keys, all five land in all four maps, coverage passes. A `REQ-PCK-011` under-delivery would therefore ship silently green.

**The enforcer**: M5 adds `TestProjectContinuationI18nKeysInAllLocales` in `internal/web`, following the repository's established per-SPEC pattern (`crosssession_test.go:100`, `feedback_panel_test.go:33`, both using the `i18nKeyInAllLocales` helper at `schema_label_test.go:49-55`). It asserts (a) the `workflow.project.continuation` `FieldDef` carries a non-empty `OptionDesc` on all three options — which fires precisely on the missing `withOptionDesc` wrapper — and (b) all eight keys resolve in all four locale maps.

Once the field does carry `OptionDesc`s, the three existing all-sections sweeps also engage: `TestEveryOptionDescKeyAvoidsOptGuard` (`option_desc_test.go:50`), `TestEveryOptionDescKeyTranslatedInAllLocales` (`:64`, requiring `strings.Count(dict, key) == 4`), and `TestEveryOptionDescRendersOnConsolePage` (`:80`). They cannot catch the wrapper's absence, only its contents.

### AC-PCK-012 — mirror parity holds after `make build`

**Given** a completed `make build`,
**When** `cmp` is run over the three byte-identical pairs (`doc-generation.md`, `todo.md`, `tab_schema.json`) and over `workflow.yaml`,
**Then** the three return rc=0 and `workflow.yaml` returns rc=1, with `diff` showing only the intended repo-local content (`branch_guard`, `agent_stop_guard`, populated `audit` pins, `context_folding`) plus the new key on both sides.

**Reachable red.** Iteration 1 measured the matrix as `0/0/0/1` on the current tree. A template-only edit without `make build` and the local-mirror update flips one of the three zeros to 1; copying the template `workflow.yaml` verbatim over the local one flips the fourth from 1 to 0 and destroys the repo-local values.

### AC-PCK-013 — the unmatched value reaches the operator (non-blocking)

**Given** the Step 4.2 report contract in `doc-generation.md`,
**When** the report template is read,
**Then** it carries a line that prints the offending value together with the canonical domain when the resolver returned a non-empty unmatched string, and prints nothing when it did not.

### AC-PCK-014 — the `.opt.` English-label guard is preserved

**Given** the implementation branch,
**When** `go test ./internal/web/...` runs,
**Then** the option-label keys use the `.opt.` prefix, the per-option-description keys use a prefix containing no `.opt.` substring (`.option.` satisfies this — the character after `opt` is `i`), and `applyI18n`'s `".opt."` guard in `app.js` is unmodified.

**Reachable red — the description-key enforcer is corrected.**

The v0.2.0 citation of `audit_option_desc_test.go:78` was wrong: that test iterates `auditOptionDescFields()`, a hardcoded four-entry map covering only `workflow.audit.model` and the three `workflow.audit.gates.*` fields, so it is structurally incapable of firing on `workflow.project.continuation`.

The enforcer that **does** fire is the all-sections sweep:

```
$ awk 'NR>=50 && NR<=57 {print NR": "$0}' internal/web/option_desc_test.go
50: func TestEveryOptionDescKeyAvoidsOptGuard(t *testing.T) {
51: 	for name, f := range allOptionDescFields(t) {
52: 		for _, opt := range f.Options {
53: 			if strings.Contains(opt.OptionDesc, ".opt.") {
54: 				t.Errorf("field %q option %q OptionDesc %q contains \".opt.\" …
```

`allOptionDescFields` walks `settings.AllSections()` (`option_desc_test.go:24-25`) and carries its own non-vacuity floor (`:36-44`), so it reaches this SPEC's field once the field carries `OptionDesc`s. An implementer who names the description keys `…continuation.opt.desc.<token>` trips it.

The `audit_option_desc_test.go:137-138` citation is **retained** and is correct — `TestOptionLabelsStayEnglishStillGuarded` asserts on `app.js` itself rather than on a field set, so it fires on any removal of the guard regardless of which fields exist. `console_ux_fix_test.go:87-88` is the second such tripwire.

**Conjunct 1 (labels use the `.opt.` prefix) has no mechanical red** — the prefix is a literal string argument to `closedSeam`, and a wrong prefix produces locale-following labels no test forbids for this field. It is decided by reading the schema row, and is marked as prose-verified in §D.3.

---

## §D.1 Severity

Thirteen blocking, one non-blocking. `AC-PCK-013` is non-blocking because it is a report-wording obligation whose absence degrades diagnosis without changing behaviour (`AC-PCK-003` guarantees the resolution itself); it nonetheless carries a reachable red, since the report line does not exist today.

No criterion in this file is unfalsifiable. The v0.2.0 `AC-PCK-014` was, by its own admission, and has been relocated to `plan.md` §D Constraints — `§D.4` gates on blocking criteria only, so an unfalsifiable non-blocking criterion decided nothing here while contradicting the header.

## §D.2 Traceability

Every `REQ-PCK-001..012` appears in at least one **blocking** AC row of §D. `REQ-PCK-007` and `REQ-PCK-008` are prohibitions verified positively rather than negatively: `AC-PCK-007` by clause presence and scope, `AC-PCK-008` by a per-branch kickoff-clause count against a measured pre-change value. The v0.1.0 diff-stat check remains **struck from `REQ-PCK-008`'s verification path**; it now lives in `plan.md` §D as an implementer constraint and carries no verification weight anywhere.

`REQ-PCK-006` is carried by `AC-PCK-006`'s four conjuncts **and** by its differential pairing with `AC-PCK-005` (header note) — the pair, not either criterion alone, is what makes a synonym implementation mechanically impossible.

## §D.3 Indirect verification

`REQ-PCK-004`, `REQ-PCK-005` and `REQ-PCK-006` govern orchestrator prose, which has no runtime harness. They are verified by reading the shipped instruction text, which is the artifact the orchestrator actually consumes. This is indirect and is recorded as such.

Two further conjuncts are prose-verified rather than command-verified, and are named here so the header's promise stays true:

- **`AC-PCK-006` conjunct 1** — decided by reading the `pipeline` row's terminal imperative. Its red is supplied mechanically by the differential pair with `AC-PCK-005`.
- **`AC-PCK-014` conjunct 1** — decided by reading the schema row's label prefix. Conjuncts 2 and 3 of that criterion are command-verified.

## §D.4 Closure gates

- All thirteen blocking ACs pass.
- `go test ./internal/config/... ./internal/settings/... ./internal/cli/wizard/... ./internal/core/project/... ./internal/web/...` is green, with each package's output cited.
- `golangci-lint run` on the touched packages is clean.
- `git status --short` shows no unintended file.

## §D.5 Forward-looking checks

- **The folding fallback is withdrawn.** v0.1.0 blessed resolving `REQ-PCK-010` by folding option descriptions into the question body; that path fails `TestWizardQuestionTranslationCompleteness` and is now forbidden by `AC-PCK-010` conjunct 3. `GetLocalizedQuestion` needs no extension — it already carries option translations (`translations.go:571-586`).
- If a later SPEC adds a fourth continuation token, `AC-PCK-001` must be updated deliberately rather than relaxed to a length check, and `AC-PCK-011`'s entry count rises from 32 to 40.
- If a later SPEC gives `pipeline` a mechanical carrier (for example a `progression_mode` default), `AC-PCK-006` conjunct 1 must be re-derived against that carrier rather than against the option text — see `spec.md` §3 D1.3 for why the progression-mode reading was rejected at this iteration.
- **When citing an existing test as an AC's red, check that test's field SCOPE, not only its name and failure message.** Both v0.2.0 miscitations (`AC-PCK-011`, `AC-PCK-014`) named tests whose failure strings said exactly the right thing but which iterate a field set this SPEC's field never enters. That is the iteration-1 "criterion that reads as verification but cannot fire" defect relocated from the criterion's command to its cited enforcer.
